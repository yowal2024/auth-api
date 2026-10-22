import { chromium } from "playwright";
import { createServer } from "vite";
import assert from "node:assert/strict";
import { mockRequest } from "./fixtures/original-mock.js";

const server = await createServer({
  server: { host: "127.0.0.1", port: 18184, strictPort: true },
});
await server.listen();
let browser;
const errors = [],
  checks = [];
const calls = {
  password: 0,
  reward: 0,
  create: 0,
  export: 0,
  profile: 0,
  upload: 0,
};
const fail = {
  password: true,
  reward: true,
  create: true,
  export: true,
  profile: true,
};
let releaseReward, releaseUpload;
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
  const context = await browser.newContext({ acceptDownloads: true });
  await context.addInitScript(() => {
    localStorage.setItem("auth_token", "test-token");
    localStorage.setItem("auth_current_appid", "1");
  });
  await context.route("**/api/**", async (route) => {
    const req = route.request(),
      url = new URL(req.url());
    if (!url.pathname.startsWith("/api/")) return route.continue();
    const path = url.pathname.replace("/api", "");
    let data,
      body = {};
    try {
      body = req.postDataJSON() || {};
    } catch {}
    const action = {
      "PUT /admin/me/password": "password",
      "PUT /admin/me/profile": "profile",
      "POST /admin/user/reward": "reward",
      "POST /admin/user": "create",
      "GET /admin/user/export": "export",
      "POST /admin/extend-file/upload": "upload",
    }[`${req.method()} ${path}`];
    if (path === "/install/status")
      data = { code: 0, data: { installed: true, adminEntry: "admin" } };
    else if (
      path === "/admin/extend-file/list" ||
      path === "/admin/extend-file/categories"
    )
      data = { code: 0, data: [] };
    else if (action) {
      calls[action]++;
      if (action === "reward")
        await new Promise((resolve) => {
          releaseReward = resolve;
        });
      if (action === "upload")
        await new Promise((resolve) => {
          releaseUpload = resolve;
        });
      if (action === "export" && !fail.export)
        return route.fulfill({
          contentType: "text/csv",
          headers: {
            "Content-Disposition": 'attachment; filename="users.csv"',
          },
          body: "id,username\n1,tester",
        });
      data = fail[action]
        ? { code: 400, msg: `${action} rejected` }
        : { code: 0, data: action === "profile" ? body : {} };
    } else {
      data = {
        code: 0,
        data: await mockRequest(path, {
          method: req.method(),
          params: Object.fromEntries(url.searchParams),
          data: body,
        }),
      };
    }
    await route.fulfill({
      contentType: "application/json",
      body: JSON.stringify(data),
    });
  });
  const page = await context.newPage();
  page.on("pageerror", (error) => errors.push(error.message));
  const modal = page.locator(".arco-modal:visible");
  await page.goto("http://127.0.0.1:18184/admin/user");
  await page.getByRole("button", { name: "添加用户", exact: true }).click();
  await modal.getByPlaceholder("请设置账号5~12位").fill("tester2");
  await modal.getByPlaceholder("请输入密码6~18").fill("secret123");
  await modal.locator(".arco-select-view").click();
  await page.locator(".arco-select-option:visible").first().click();
  await modal.getByRole("button", { name: "提交", exact: true }).click();
  await page.getByText("create rejected", { exact: true }).waitFor();
  assert.equal(
    await modal.getByPlaceholder("请设置账号5~12位").inputValue(),
    "tester2",
  );
  fail.create = false;
  await modal.getByRole("button", { name: "提交", exact: true }).click();
  await modal.waitFor({ state: "hidden" });
  assert.equal(calls.create, 2);
  checks.push(
    "create-user failure retains input without unhandled rejection; retry succeeds",
  );

  await page.getByRole("button", { name: /导出/ }).click();
  await modal.getByRole("button", { name: "导出", exact: true }).click();
  await page.getByText("export rejected", { exact: true }).waitFor();
  assert.equal(await modal.count(), 1);
  fail.export = false;
  const downloaded = page.waitForEvent("download");
  await modal.getByRole("button", { name: "导出", exact: true }).click();
  assert.match((await downloaded).suggestedFilename(), /\.csv$/);
  await modal.waitFor({ state: "hidden" });
  checks.push("export failure retains dialog; retry downloads CSV and closes");

  await page.getByRole("button", { name: /发送奖励/ }).click();
  await modal.locator("input").first().fill("3");
  await modal.getByRole("button", { name: "提交", exact: true }).click();
  await waitForMock(() => releaseReward);
  assert.equal(
    await modal.getByRole("button", { name: "取消", exact: true }).isDisabled(),
    true,
  );
  await page.keyboard.press("Escape");
  assert.equal(await modal.count(), 1);
  releaseReward();
  await page.getByText("reward rejected", { exact: true }).waitFor();
  assert.equal(await modal.locator("input").first().inputValue(), "3");
  fail.reward = false;
  releaseReward = undefined;
  await modal.getByRole("button", { name: "提交", exact: true }).click();
  await waitForMock(() => releaseReward);
  assert.equal(calls.reward, 2);
  releaseReward();
  await modal.waitFor({ state: "hidden" });
  checks.push(
    "reward submission blocks cancellation and retains failed input for retry",
  );

  await page.goto("http://127.0.0.1:18184/admin/extend-file");
  await page.getByRole("button", { name: /上传文件/ }).click();
  await modal
    .locator('input[type="file"]')
    .setInputFiles({
      name: "config.txt",
      mimeType: "text/plain",
      buffer: Buffer.from("test"),
    });
  await modal.getByRole("button", { name: "上传", exact: true }).click();
  await waitForMock(() => releaseUpload);
  assert.equal(
    await modal.getByRole("button", { name: "取消", exact: true }).isDisabled(),
    true,
  );
  await page.keyboard.press("Escape");
  assert.equal(await modal.count(), 1);
  // Drive an intermediate progress value to verify Arco's 0..1 API contract.
  await page.evaluate(() => {
    let instance = document.querySelector(".page").__vueParentComponent;
    while (instance && instance.type.name !== "ExtendFile")
      instance = instance.parent;
    if (!instance) throw Error("ExtendFile component missing");
    instance.proxy.u = 50;
  });
  await page.getByText("正在上传 50%", { exact: true }).waitFor();
  assert.equal(
    await modal.locator(".arco-progress-line-text").innerText(),
    "50%",
  );
  releaseUpload();
  await modal.waitFor({ state: "hidden" });
  assert.equal(calls.upload, 1);
  checks.push(
    "upload remains open until completion and renders 50 percent correctly",
  );

  await page.goto("http://127.0.0.1:18184/admin/profile");
  await page
    .locator(".row")
    .filter({ has: page.locator(".label", { hasText: "账号" }) })
    .getByText("admin", { exact: true })
    .waitFor();
  const nicknameRow = page
    .locator(".row")
    .filter({ has: page.locator(".label", { hasText: "昵称" }) });
  await nicknameRow.locator(".editable").click();
  await nicknameRow.locator("input").fill("updated-name");
  await nicknameRow.locator("input").press("Enter");
  await page.getByText("profile rejected", { exact: true }).waitFor();
  assert.equal(await nicknameRow.locator("input").inputValue(), "updated-name");
  fail.profile = false;
  await nicknameRow.locator("input").press("Enter");
  await nicknameRow
    .locator(".editable")
    .filter({ hasText: "updated-name" })
    .waitFor();
  assert.equal(calls.profile, 2);
  checks.push(
    "profile failure retains editable value; Enter and blur do not duplicate saves",
  );

  await page
    .locator(".row")
    .filter({ has: page.locator(".label", { hasText: "密码" }) })
    .locator(".editable")
    .click();
  await modal.getByRole("button", { name: "确定", exact: true }).click();
  assert.equal(calls.password, 0);
  await modal
    .getByPlaceholder("请输入原密码", { exact: true })
    .fill("old-secret");
  await modal
    .getByPlaceholder("请输入新密码", { exact: true })
    .fill("new-secret");
  await modal
    .getByPlaceholder("请再次输入新密码", { exact: true })
    .fill("wrong-secret");
  await modal.getByRole("button", { name: "确定", exact: true }).click();
  assert.equal(calls.password, 0);
  await modal
    .getByPlaceholder("请再次输入新密码", { exact: true })
    .fill("new-secret");
  await modal.getByRole("button", { name: "确定", exact: true }).click();
  await page.getByText("password rejected", { exact: true }).waitFor();
  assert.equal(
    await modal.getByPlaceholder("请输入新密码", { exact: true }).inputValue(),
    "new-secret",
  );
  fail.password = false;
  await modal.getByRole("button", { name: "确定", exact: true }).click();
  await page.waitForURL("**/admin/login");
  assert.equal(calls.password, 2);
  assert.equal(
    await page.evaluate(() => localStorage.getItem("auth_token")),
    null,
  );
  checks.push(
    "password validates, submits, retains failed input and logs out after success",
  );
  assert.deepEqual(errors, []);
  console.log(JSON.stringify({ checks, errors }, null, 2));
} finally {
  await browser?.close();
  await server.close();
}
