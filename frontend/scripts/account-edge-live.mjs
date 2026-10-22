import { chromium } from "playwright";
import assert from "node:assert/strict";
const base = process.env.TEST_BASE_URL || "http://127.0.0.1:18080";
assert.ok(process.env.TEST_ADMIN_PASSWORD, "Set TEST_ADMIN_PASSWORD");
let token, app, browser;
const errors = [];
async function api(path, body, method = body === undefined ? "GET" : "POST") {
  const response = await fetch(base + "/api/admin" + path, { method, headers: { "Content-Type": "application/json", Authorization: "Bearer " + (token || ""), "X-App-Id": String(app?.id || 0) }, body: body === undefined ? undefined : JSON.stringify(body) });
  const out = await response.json(); assert.equal(out.code, 0, path + ": " + out.msg); return out.data;
}
try {
  token = (await api("/login", { username: "admin", password: process.env.TEST_ADMIN_PASSWORD })).token;
  app = await api("/app", { name: `Ban browser ${Date.now()}`, type: "user", mode: "free" });
  const group = await api("/user/group", { name: "UI test", level: 1 });
  const user = await api("/user", { username: "uitestban", password: "test-pass-123", groupId: group.id });
  await api(`/user/${user.id}`, { banExpire: -1 }, "PUT");
  browser = await chromium.launch({ executablePath: process.env.CHROME_PATH || "C:/Program Files/Google/Chrome/Application/chrome.exe", headless: true });
  const context = await browser.newContext();
  await context.addInitScript(({ token, id }) => { localStorage.setItem("auth_token", token); localStorage.setItem("auth_current_appid", String(id)); }, { token, id: app.id });
  const page = await context.newPage();
  page.on("pageerror", e => errors.push(e.message));
  await page.goto(base + `/admin/user/edit/${user.id}`);
  const permanent = page.getByRole("switch", { name: "永久禁用", exact: true });
  await permanent.waitFor();
  assert.equal(await permanent.getAttribute("aria-checked"), "true");
  await page.getByPlaceholder("请输入昵称").fill("ban preserved");
  async function save() {
    const saved = page.waitForResponse(r => r.url().endsWith(`/api/admin/user/${user.id}`) && r.request().method() === "PUT");
    await page.getByRole("button", { name: "提交", exact: true }).click();
    assert.equal((await (await saved).json()).code, 0);
  }
  await save();
  assert.equal((await api(`/user/${user.id}`)).banExpire,-1);
  await page.reload();
  await permanent.waitFor();
  assert.equal(await permanent.getAttribute("aria-checked"), "true");
  const field = page.locator(".arco-form-item").filter({ hasText: "禁用用户期限" });
  await field.getByRole("switch").first().click();
  await save();
  assert.equal((await api(`/user/${user.id}`)).banExpire,0);
  await field.getByRole("switch").first().click();
  await page.getByRole("button", { name: "提交", exact: true }).click();
  await page.getByText("请选择禁用截止时间或永久禁用", { exact: true }).waitFor();
  assert.equal((await api(`/user/${user.id}`)).banExpire,0);
  await permanent.click();
  await save();
  assert.equal((await api(`/user/${user.id}`)).banExpire,-1);
  assert.deepEqual(errors, []);
  console.log("PASS 永久封禁正确回显、保存保持封禁、取消封禁和空日期拦截");
} finally {
  await browser?.close();
  if (app) await api(`/app/${app.id}`, undefined, "DELETE");
}
