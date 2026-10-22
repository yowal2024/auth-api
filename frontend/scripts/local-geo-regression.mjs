import { chromium } from "playwright";
import { createServer } from "vite";
import assert from "node:assert/strict";
import { mockRequest } from "./fixtures/original-mock.js";

const server = await createServer({
  server: { host: "127.0.0.1", port: 18189, strictPort: true },
});
await server.listen();
let browser;
const errors = [],
  checks = [];
try {
  browser = await chromium.launch({
    executablePath:
      process.env.CHROME_PATH ||
      "C:/Program Files/Google/Chrome/Application/chrome.exe",
    headless: true,
  });
  const context = await browser.newContext();
  await context.addInitScript(() => {
    localStorage.setItem("auth_token", "test");
    localStorage.setItem("auth_current_appid", "1");
  });
  let ready = false,
    fail = false,
    role = "super",
    saved;
  let releaseDownload,
    updating = false;
  const status = () => ({
    ready,
    updating,
    stage: updating ? "downloading" : "complete",
    downloadedBytes: updating ? 50 : 100,
    totalBytes: 100,
    buildTime: ready ? 1790812800 : 0,
    updatedAt: ready ? 1791850000 : 0,
    lastError: "",
  });
  await context.route("**/api/**", async (route) => {
    const req = route.request(),
      url = new URL(req.url()),
      path = url.pathname.replace("/api/private", "/admin");
    if (!url.pathname.startsWith("/api/")) return route.continue();
    let out = { code: 0, data: {} };
    if (url.pathname === "/api/install/status")
      out.data = { installed: true, adminEntry: "private" };
    else if (path === "/admin/me")
      out.data = { id: 1, username: "admin", role };
    else if (path === "/admin/app/list")
      out.data = [
        {
          id: 1,
          name: "Local geo test",
          type: "user",
          mode: "free",
          status: 1,
          userTokenTtl: 86400,
          bindRegion: true,
          geoProvider: "dbip",
          ipDistanceMax: 50,
        },
      ];
    else if (path === "/admin/app/1" && req.method() === "PUT") {
      saved = req.postDataJSON();
      out.data = saved;
    } else if (path === "/admin/geo/local") out.data = status();
    else if (path === "/admin/geo/local/update") {
      if (fail) out = { code: 500, msg: "测试更新失败" };
      else {
        updating = true;
        await new Promise((resolve) => {
          releaseDownload = resolve;
        });
        updating = false;
        ready = true;
        out.data = status();
      }
    } else {
      try {
        out.data = await mockRequest(path, {
          method: req.method(),
          params: Object.fromEntries(url.searchParams),
        });
      } catch {
        out.data = [];
      }
    }
    await route.fulfill({
      contentType: "application/json",
      body: JSON.stringify(out),
    });
  });
  const page = await context.newPage();
  page.on("pageerror", (e) => errors.push(e.message));
  await page.goto("http://127.0.0.1:18189/private/app-info");
  const panel = page.locator(".local-geo");
  await panel.getByText("未安装", { exact: true }).waitFor();
  assert.equal(await page.getByPlaceholder("对应平台的应用 KEY").count(), 0);
  await panel.getByRole("button", { name: "安装", exact: true }).click();
  await panel.getByText("下载中 50%", { exact: true }).waitFor();
  assert.equal(await panel.getByRole("button").count(), 0);
  if (process.env.TEST_SCREENSHOT_DIR)
    await panel.screenshot({
      path: process.env.TEST_SCREENSHOT_DIR + "/geo-progress.png",
    });
  releaseDownload();
  await panel.getByText("已安装", { exact: true }).waitFor();
  assert.match(
    await panel.locator(".local-geo-time").innerText(),
    /^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$/,
  );
  await panel.locator(".local-geo-status").hover();
  await page.getByText("库版本：2026-10", { exact: true }).waitFor();
  await page.mouse.move(0, 0);
  assert.equal(await panel.locator("input").count(), 0);
  assert.equal(await page.locator(".geo-row > .local-geo.geo-key").count(), 1);
  checks.push(
    "inline KEY-slot status; real polled progress; installed month; no IP test input",
  );
  if (process.env.TEST_SCREENSHOT_DIR) {
    await panel.screenshot({
      path: process.env.TEST_SCREENSHOT_DIR + "/geo-installed.png",
    });
    await panel.locator(".local-geo-status").hover();
    await page.getByText("库版本：2026-10", { exact: true }).waitFor();
    await page.screenshot({
      path: process.env.TEST_SCREENSHOT_DIR + "/geo-tooltip.png",
    });
    await page.mouse.move(0, 0);
  }
  fail = true;
  await panel.getByRole("button", { name: "更新", exact: true }).click();
  await page.getByText("测试更新失败", { exact: true }).waitFor();
  await panel.getByText("已安装", { exact: true }).waitFor();
  await page.getByRole("button", { name: "保存修改", exact: true }).click();
  await page.getByText("已保存", { exact: true }).waitFor();
  assert.equal(saved.geoProvider, "dbip");
  role = "admin";
  await page.reload();
  await panel.getByText("已安装", { exact: true }).waitFor();
  assert.equal(
    await panel.getByRole("button", { name: "更新", exact: true }).count(),
    0,
  );
  await page.setViewportSize({ width: 390, height: 844 });
  assert.ok(
    await page.evaluate(
      () => document.documentElement.scrollWidth <= innerWidth + 1,
    ),
  );
  checks.push(
    "failed update preserves ready state; provider persists; only super admins can update; mobile fits",
  );
  assert.deepEqual(errors, []);
  console.log(JSON.stringify({ checks, errors }, null, 2));
} finally {
  await browser?.close();
  await server.close();
}
