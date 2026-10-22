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
  app = await api("/app", { name: `Card browser ${Date.now()}`, type: "kami", mode: "free" });
  const group = await api("/kami/group", { name: "UI test", type: "vip", value: 60 });
  const card = (await api("/kami/generate", { groupId: group.id, count: 1, length: 16 })).cards[0];
  browser = await chromium.launch({ executablePath: process.env.CHROME_PATH || "C:/Program Files/Google/Chrome/Application/chrome.exe", headless: true });
  const context = await browser.newContext();
  await context.addInitScript(({ token, id }) => { localStorage.setItem("auth_token", token); localStorage.setItem("auth_current_appid", String(id)); }, { token, id: app.id });
  const page = await context.newPage();
  page.on("pageerror", e => errors.push(e.message));
  await page.goto(base + "/admin/kami");
  const row = page.locator("tr").filter({ hasText: card });
  await row.getByText("未使用", { exact: true }).waitFor();
  for (const [action, status, label] of [["冻结", 2, "已冻结"], ["解冻", 0, "未使用"]]) {
    await row.getByText(action, { exact: true }).click();
    const saved = page.waitForResponse(r => r.url().includes("/api/admin/kami/") && r.request().method() === "PUT");
    await page.locator(".arco-popconfirm:visible").getByRole("button", { name: "确定", exact: true }).click();
    assert.equal((await (await saved).json()).code, 0);
    await row.getByText(label, { exact: true }).waitFor();
    assert.equal((await api("/kami/list")).list[0].status, status);
  }
  await row.getByText("编辑", { exact: true }).click();
  await page.locator(".arco-drawer:visible").getByText("未使用", { exact: true }).waitFor();
  assert.deepEqual(errors, []);
  console.log("PASS 浏览器冻结、解冻、状态显示和编辑弹窗与真实数据库一致");
} finally {
  await browser?.close();
  if (app) await api(`/app/${app.id}`, undefined, "DELETE");
}
