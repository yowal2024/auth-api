import { chromium } from "playwright";
import { createServer } from "vite";
import assert from "node:assert/strict";
import { mockRequest } from "./fixtures/original-mock.js";

const server = await createServer({
  server: { host: "127.0.0.1", port: 18182, strictPort: true },
});
await server.listen();
let browser;
const errors = [],
  checks = [];
async function waitForMock(ready) {
  const deadline = Date.now() + 5000;
  while (!ready()) {
    assert.ok(Date.now() < deadline, "mock request did not arrive");
    await new Promise((resolve) => setTimeout(resolve, 10));
  }
}
try {
  browser = await chromium.launch({
    executablePath:
      process.env.CHROME_PATH ||
      "C:/Program Files/Google/Chrome/Application/chrome.exe",
    headless: true,
  });
  const context = await browser.newContext();
  let failVersion = true,
    versionPosts = 0;
  let releaseSlowLogs, releaseOldSession;
  let clearLogBody;
  let adminRole = "super";
  let restrictedRequests = 0;
  await context.addInitScript(() => {
    localStorage.setItem("auth_token", "test-token");
    localStorage.setItem("auth_current_appid", "999999");
  });
  await context.route("**/api/**", async (route) => {
    const req = route.request(),
      url = new URL(req.url());
    if (!url.pathname.startsWith("/api/")) return route.continue();
    let data;
    if (url.pathname === "/api/install/status")
      data = { code: 0, data: { installed: true, adminEntry: "private" } };
    else if (url.pathname === "/api/private/me")
      data = { code: 0, data: { id: 1, username: "tester", role: adminRole } };
    else if (adminRole === "admin" && /^\/api\/private\/(system|admin|function|encrypt|extend|extend-file|ver)(\/|$)/.test(url.pathname)) {
      restrictedRequests++;
      data = { code: 403, msg: "restricted" };
    }
    else if (url.pathname === "/api/private/encrypt/list")
      data = { code: 0, data: [{ id: 7, name: "Active policy", algorithm: "session", timeSkew: 60, signMode: "verify" }] };
    else if (url.pathname === "/api/private/test-old-session") {
      await new Promise((resolve) => {
        releaseOldSession = resolve;
      });
      data = { code: 401, msg: "old session expired" };
    } else if (url.pathname === "/api/private/log/list") {
      const keyword = url.searchParams.get("keyword") || "initial";
      if (keyword === "slow")
        await new Promise((resolve) => {
          releaseSlowLogs = resolve;
        });
      data = {
        code: 0,
        data: {
          list: [
            {
              id: 77,
              username: keyword,
              action: "logon",
              data: keyword,
              createTime: 100,
            },
          ],
          total: 1,
        },
      };
    } else if (url.pathname === "/api/private/log/clear") {
      clearLogBody = req.postDataJSON();
      data = { code: 0, data: { affected: 1 } };
    } else if (url.pathname === "/api/private/test-401")
      return route.fulfill({
        status: 401,
        contentType: "application/json",
        body: JSON.stringify({ code: 401, msg: "expired" }),
      });
    else if (url.pathname === "/api/private/test-blob")
      data = { code: 404, msg: "missing download" };
    else if (
      url.pathname === "/api/private/test-json-file" ||
      url.pathname === "/api/private/user/export"
    )
      return route.fulfill({
        contentType: "application/json",
        headers: { "Content-Disposition": 'attachment; filename="data.json"' },
        body: '{"code":123,"value":"real file"}',
      });
    else if (url.pathname === "/api/private/app/list")
      data = {
        code: 0,
        data: [
          {
            id: 1,
            name: "Test application",
            type: "user",
            mode: "free",
            status: 1,
            tokenKey: "test-key",
            encryption: 7,
            userTokenTtl: 86400,
          },
        ],
      };
    else if (url.pathname === "/api/private/ver" && req.method() === "POST") {
      versionPosts++;
      data = failVersion
        ? { code: 400, msg: "version rejected" }
        : { code: 0, data: { id: 1 } };
    } else {
      let body = {};
      try {
        body = req.postDataJSON() || {};
      } catch {}
      try {
        data = {
          code: 0,
          data: await mockRequest(
            url.pathname.replace("/api/private", "/admin"),
            {
              method: req.method(),
              params: Object.fromEntries(url.searchParams),
              data: body,
            },
          ),
        };
      } catch {
        data = { code: 0, data: [] };
      }
    }
    await route.fulfill({
      contentType: "application/json",
      body: JSON.stringify(data),
    });
  });
  const page = await context.newPage();
  page.on("pageerror", (e) => errors.push(e.message));
  await page.goto("http://127.0.0.1:18182/private/version");
  await page.getByRole("button", { name: /创建版本/ }).waitFor();
  assert.equal(
    await page.evaluate(() => localStorage.getItem("auth_current_appid")),
    "1",
  );
  assert.match(await page.title(), /版本控制/);
  checks.push("stale application selection repaired before page requests");
  await page.getByRole("button", { name: /创建版本/ }).click();
  const modal = page.locator(".arco-modal:visible");
  await modal.getByRole("button", { name: "提交", exact: true }).click();
  assert.equal(versionPosts, 0);
  assert.equal(await modal.count(), 1);
  await modal.getByPlaceholder("如: 安卓版").fill("Android");
  await modal.getByPlaceholder("如: android").fill("stable");
  await modal.getByRole("button", { name: "提交", exact: true }).click();
  await page.getByText("version rejected", { exact: true }).waitFor();
  assert.equal(
    await modal.getByPlaceholder("如: android").inputValue(),
    "stable",
  );
  failVersion = false;
  await modal.getByRole("button", { name: "提交", exact: true }).click();
  await modal.waitFor({ state: "hidden" });
  assert.equal(versionPosts, 2);
  checks.push(
    "validation and failed save retain version form; successful retry closes it",
  );
  const downloads = await page.evaluate(async () => {
    const { request } = await import("/src/utils/request.js");
    let rejected = false;
    try {
      await request("/admin/test-blob", { responseType: "blob" });
    } catch {
      rejected = true;
    }
    const file = await request("/admin/test-json-file", {
      responseType: "blob",
    });
    const { exportUsers } = await import("/src/api/user.js");
    const exported = await exportUsers({ format: "json" });
    return {
      rejected,
      content: await file.data.text(),
      exported: await exported.text(),
    };
  });
  assert.equal(downloads.rejected, true);
  assert.match(downloads.content, /real file/);
  assert.match(downloads.exported, /real file/);
  checks.push("download errors rejected; real JSON attachment preserved");
  const links = await page.evaluate(async () => {
    const { safeWebURL } = await import("/src/utils/url.js");
    return {
      script: safeWebURL("javascript:alert(1)"),
      data: safeWebURL("data:text/html,test"),
      valid: safeWebURL("/uploads/app.apk"),
    };
  });
  assert.equal(links.script, undefined);
  assert.equal(links.data, undefined);
  assert.equal(links.valid, "http://127.0.0.1:18182/uploads/app.apk");
  checks.push("version links reject executable schemes");
  const editorModels = await page.evaluate(async () => {
    const { editor } = await import("/src/utils/monaco.js");
    return editor.getModels().length;
  });
  await page.evaluate(async () => {
    const { default: router } = await import("/src/router/index.js");
    await router.push("/private/function/edit/0");
  });
  await page.locator(".monaco-editor").waitFor();
  await page.evaluate(async () => {
    const { default: router } = await import("/src/router/index.js");
    await router.push("/private/dashboard");
  });
  await page.locator(".chart-body canvas").waitFor();
  assert.equal(
    await page.evaluate(async () => {
      const { editor } = await import("/src/utils/monaco.js");
      return editor.getModels().length;
    }),
    editorModels,
  );
  checks.push(
    "editor models disposed after navigation; reduced chart bundle renders",
  );
  await page.locator(".fold-btn.arco-icon-menu-fold").waitFor();
  await page.locator(".fold-btn").click();
  await page.locator(".fold-btn.arco-icon-menu-unfold").waitFor();
  assert.equal(
    await page.locator(".arco-icon-sun-fill,.arco-icon-moon-fill").count(),
    1,
  );
  checks.push(
    "locally registered dynamic icons render after removing global plugins",
  );

  await page
    .getByText("missing download", { exact: true })
    .waitFor({ state: "hidden" });
  await page.evaluate(async () => {
    const { request } = await import("/src/utils/request.js");
    const { reportError } = await import("/src/utils/errors.js");
    try {
      await request("/admin/test-blob", { responseType: "blob" });
    } catch (error) {
      reportError(error, "download failed");
    }
  });
  assert.equal(
    await page.getByText("missing download", { exact: true }).count(),
    1,
  );
  checks.push("API error and page catch produce one notification");

  await page.evaluate(async () => {
    const { default: router } = await import("/src/router/index.js");
    await router.push("/private/logs/user");
  });
  await page.locator(".log-title").filter({ hasText: "initial" }).waitFor();
  const search = page.getByPlaceholder("用户搜索...");
  await search.fill("slow");
  await search.press("Enter");
  // Wait for the mock to hold the first response before issuing the new query.
  await waitForMock(() => releaseSlowLogs);
  await search.fill("fast");
  await search.press("Enter");
  await page.locator(".log-title").filter({ hasText: "fast" }).waitFor();
  const slowResponse = page.waitForResponse((res) =>
    res.url().includes("keyword=slow"),
  );
  releaseSlowLogs();
  await slowResponse;
  await page.waitForTimeout(100);
  assert.equal(
    await page.locator(".log-title").filter({ hasText: "slow" }).count(),
    0,
  );
  await page.locator(".arco-table-tr .arco-checkbox").last().click();
  await page.getByRole("button", { name: "清理日志", exact: true }).click();
  await page
    .locator(".arco-popconfirm")
    .getByRole("button", { name: "确定", exact: true })
    .click();
  await page.getByText("已清理", { exact: true }).waitFor();
  assert.deepEqual(clearLogBody, { ids: [77], category: "user" });
  checks.push(
    "latest log query wins; selected deletion explicitly targets user logs",
  );

  await page.evaluate(async () => {
    const { default: router } = await import("/src/router/index.js");
    await router.push("/private/encrypt");
  });
  const activePolicy = page.locator(".encrypt-card").filter({ hasText: "Active policy" });
  await activePolicy.getByText("使用中", { exact: true }).waitFor();
  await activePolicy.locator(".more-btn").click();
  assert.match(await page.locator(".arco-dropdown-option").filter({ hasText: "删除" }).getAttribute("class"), /disabled/);
  await page.keyboard.press("Escape");
  checks.push("selected encryption policy is marked in use and cannot be deleted from the menu");

  adminRole = "admin";
  await page.evaluate(async () => {
    const { useUserStore } = await import("/src/stores/user.js");
    await useUserStore().fetchMe();
  });
  assert.equal(await page.locator(".header-actions .arco-icon-settings").count(), 0);
  for (const path of ["/private/system", "/private/admin", "/private/function", "/private/app-info", "/private/extend", "/private/extend-file", "/private/version", "/private/encrypt", "/private/logs/admin"]) {
    await page.evaluate(async (path) => {
      const { default: router } = await import("/src/router/index.js");
      await router.push(path);
    }, path);
    await page.waitForURL("**/private/app");
  }
  assert.equal(await page.getByRole("button", {name:"添加应用",exact:true}).count(), 0);
  assert.equal(await page.locator(".del-btn,.key-line").count(), 0);
  await page.evaluate(async () => {
    const { default: router } = await import("/src/router/index.js");
    await router.push("/private/logs/user");
  });
  assert.equal(await page.getByRole("button", {name:"清理日志",exact:true}).count(), 0);
  assert.equal(restrictedRequests, 0);
  checks.push("ordinary admins cannot enter super-only pages or trigger their APIs");
  adminRole = "super";
  await page.evaluate(async () => {
    const { default: router } = await import("/src/router/index.js");
    const { useUserStore } = await import("/src/stores/user.js");
    await useUserStore().fetchMe();
    await router.push("/private/logs/user");
  });

  await page.evaluate(async () => {
    const { request } = await import("/src/utils/request.js");
    window.oldSessionRequest = request("/admin/test-old-session").catch(
      () => {},
    );
  });
  await waitForMock(() => releaseOldSession);
  await page.evaluate(async () => {
    const { useUserStore } = await import("/src/stores/user.js");
    const { setToken } = await import("/src/utils/storage.js");
    useUserStore().token = "new-session";
    setToken("new-session");
  });
  releaseOldSession();
  await page.evaluate(() => window.oldSessionRequest);
  assert.equal(
    await page.evaluate(() => localStorage.getItem("auth_token")),
    "new-session",
  );
  assert.ok(page.url().endsWith("/private/logs/user"));
  checks.push("late 401 from the previous login preserves the new session");
  await page.evaluate(async () => {
    const { request } = await import("/src/utils/request.js");
    try {
      await request("/admin/test-401");
    } catch {}
  });
  await page.waitForURL("**/private/login");
  assert.equal(
    await page.evaluate(() => localStorage.getItem("auth_token")),
    null,
  );
  checks.push("HTTP 401 clears login and redirects to custom entry");
  assert.deepEqual(errors, []);
  console.log(JSON.stringify({ checks, errors }, null, 2));
} finally {
  await browser?.close();
  await server.close();
}
