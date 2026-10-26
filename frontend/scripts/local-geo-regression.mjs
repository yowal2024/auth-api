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
  let systemSaved,
    failSave = false,
    systemReads = 0;
  let systemConfig = { provider: "dbip", key: "", legacyApps: [] };
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
    } else if (path === "/admin/system/geo" && req.method() === "GET") {
      systemReads++;
      out.data = systemConfig;
    } else if (path === "/admin/system/geo" && req.method() === "POST") {
      if (failSave) out = { code: 500, msg: "测试保存失败" };
      else {
        systemSaved = req.postDataJSON();
        systemConfig = { ...systemSaved, legacyApps: [] };
        out.data = systemSaved;
      }
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
  await page.getByText("IP 地点距离阈值", { exact: true }).waitFor();
  assert.equal(await page.locator(".local-geo").count(), 0);
  assert.equal(await page.getByPlaceholder("对应平台的应用 KEY").count(), 0);
  assert.equal(await page.locator(".geo-distance input").inputValue(), "50");
  await page
    .getByRole("button", { name: "配置系统定位服务", exact: true })
    .click();
  await page.waitForURL("**/private/system?tab=geo");
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
  assert.equal(
    await page.locator(".geo-settings .system-local-geo").count(),
    1,
  );
  checks.push(
    "app retains its distance rule and links to system settings; local database has polled progress and installed month",
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
  await page.getByRole("button", { name: "保存定位服务", exact: true }).click();
  await page.getByText("已保存，所有应用立即生效", { exact: true }).waitFor();
  assert.deepEqual(systemSaved, { provider: "dbip", key: "" });
  await page.locator(".geo-settings .arco-select").click();
  await page
    .locator(".arco-select-option")
    .getByText("腾讯地图", { exact: true })
    .click();
  const keyInput = page.getByPlaceholder("对应平台的应用 KEY");
  await keyInput.fill("new-system-key");
  failSave = true;
  await page.getByRole("button", { name: "保存定位服务", exact: true }).click();
  await page.getByText("测试保存失败", { exact: true }).waitFor();
  assert.equal(await keyInput.inputValue(), "new-system-key");
  assert.equal(systemConfig.provider, "dbip");
  failSave = false;
  const saveTencent = page.waitForResponse(
    (response) =>
      response.url().endsWith("/api/private/system/geo") &&
      response.request().method() === "POST",
  );
  await page.getByRole("button", { name: "保存定位服务", exact: true }).click();
  await saveTencent;
  assert.deepEqual(systemSaved, { provider: "tencent", key: "new-system-key" });
  await page.reload();
  await keyInput.waitFor();
  assert.equal(await keyInput.inputValue(), "new-system-key");
  if (process.env.TEST_SCREENSHOT_DIR)
    await page.screenshot({
      path: process.env.TEST_SCREENSHOT_DIR + "/system-geo-settings.png",
      fullPage: true,
    });
  checks.push(
    "global provider and KEY persist; failed saves keep the form and active configuration",
  );

  systemConfig = {
    provider: "",
    key: "",
    legacyApps: [
      {
        id: 1,
        name: "Legacy Tencent",
        provider: "tencent",
        key: "preserved-key",
      },
      { id: 2, name: "Legacy Amap", provider: "amap", key: "other-key" },
    ],
  };
  systemSaved = undefined;
  await page.reload();
  await page.getByText(/现有应用使用不同的定位配置/).waitFor();
  await page.locator(".geo-settings .arco-select").first().click();
  await page
    .locator(".arco-select-option")
    .getByText("Legacy Tencent · 腾讯地图", { exact: true })
    .click();
  assert.equal(await keyInput.inputValue(), "preserved-key");
  assert.equal(systemSaved, undefined);
  const saveLegacy = page.waitForResponse(
    (response) =>
      response.url().endsWith("/api/private/system/geo") &&
      response.request().method() === "POST",
  );
  await page.getByRole("button", { name: "保存定位服务", exact: true }).click();
  await saveLegacy;
  assert.deepEqual(systemSaved, { provider: "tencent", key: "preserved-key" });
  checks.push(
    "conflicting old configurations require an explicit save and their keys can be reused",
  );

  await page.setViewportSize({ width: 390, height: 844 });
  assert.ok(
    await page.evaluate(
      () => document.documentElement.scrollWidth <= innerWidth + 1,
    ),
  );
  await page.goto("http://127.0.0.1:18189/private/app-info");
  await page.getByText("IP 地点距离阈值", { exact: true }).waitFor();
  await page.locator(".geo-distance input").fill("100");
  await page.getByRole("button", { name: "保存修改", exact: true }).click();
  await page.getByText("已保存", { exact: true }).waitFor();
  assert.equal(saved.ipDistanceMax, 100);
  assert.equal(Object.hasOwn(saved, "geoProvider"), false);
  assert.equal(Object.hasOwn(saved, "geoKey"), false);
  assert.deepEqual(systemSaved, { provider: "tencent", key: "preserved-key" });
  role = "admin";
  const before = systemReads;
  await page.goto("http://127.0.0.1:18189/private/system?tab=geo");
  await page.waitForFunction(() => !location.pathname.endsWith("/system"));
  assert.equal(await page.locator(".geo-settings").count(), 0);
  assert.equal(systemReads, before);
  checks.push(
    "mobile fits; app saves preserve global configuration; ordinary admins cannot read the system KEY",
  );
  assert.deepEqual(errors, []);
  console.log(JSON.stringify({ checks, errors }, null, 2));
} finally {
  await browser?.close();
  await server.close();
}
