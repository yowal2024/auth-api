import { chromium } from "playwright";
import assert from "node:assert/strict";
import os from "node:os";
import path from "node:path";

const base = process.env.TEST_BASE_URL;
assert.ok(base, "TEST_BASE_URL must point to the authorized test backend");
const entry = process.env.TEST_ADMIN_ENTRY || "admin";
const password = process.env.TEST_ADMIN_PASSWORD;
assert.ok(password, "TEST_ADMIN_PASSWORD is required");
const browser = await chromium.launch({
  executablePath: process.env.CHROME_PATH || "C:/Program Files/Google/Chrome/Application/chrome.exe",
  headless: true,
});
const errors = [], checks = [];
let adminToken, createdID;
const api = async (endpoint, options = {}) => {
  const response = await fetch(`${base}/api/${entry}${endpoint}`, {
    ...options,
    headers: { Authorization: `Bearer ${adminToken}`, "Content-Type": "application/json", ...options.headers },
  });
  const value = await response.json();
  assert.equal(value.code, 0, value.msg);
  return value.data;
};
try {
  const admin = await browser.newPage({ viewport: { width: 1440, height: 1000 } });
  admin.on("pageerror", error => errors.push(error.message));
  await admin.goto(`${base}/${entry}/login`);
  await admin.getByPlaceholder("账号", { exact: true }).fill(process.env.TEST_ADMIN_USER || "admin");
  await admin.getByPlaceholder("密码", { exact: true }).fill(password);
  await admin.getByRole("button", { name: /登\s*录/ }).click();
  await admin.waitForURL(`**/${entry}/app`);
  adminToken = await admin.evaluate(() => localStorage.getItem("auth_token"));
  const apps = await api("/app/list");
  assert.ok(apps.length, "one existing test application is required");
  const app = apps.find(item => item.type === "user") || apps[0];
  await admin.goto(`${base}/${entry}/admin`);
  await admin.getByRole("button", { name: "新增管理员", exact: true }).click();
  const modal = admin.locator(".arco-modal:visible");
  const username = `op${Date.now().toString().slice(-9)}`;
  await modal.locator("input").first().fill(username);
  await modal.locator('input[type="password"]').fill("role-test-123");
  const saved = admin.waitForResponse(r => r.url().endsWith(`/api/${entry}/admin`) && r.request().method() === "POST");
  await modal.getByRole("button", { name: "确定", exact: true }).click();
  const response = await saved;
  assert.equal(response.request().postDataJSON().role, "admin");
  const result = await response.json();
  assert.equal(result.code, 0, result.msg);
  createdID = result.data.id;
  await modal.waitFor({ state: "hidden" });
  checks.push("create ordinary administrator through the real form");

  const context = await browser.newContext({ viewport: { width: 1440, height: 1000 } });
  const page = await context.newPage();
  page.on("pageerror", error => errors.push(error.message));
  await page.goto(`${base}/${entry}/login`);
  await page.getByPlaceholder("账号", { exact: true }).fill(username);
  await page.getByPlaceholder("密码", { exact: true }).fill("role-test-123");
  await page.getByRole("button", { name: /登\s*录/ }).click();
  await page.waitForURL(`**/${entry}/app`);
  await page.waitForLoadState("networkidle");
  assert.equal(await page.getByRole("button", { name: "添加应用", exact: true }).count(), 0);
  assert.equal(await page.locator(".del-btn,.key-line").count(), 0);
  await page.evaluate(id => localStorage.setItem("auth_current_appid", String(id)), app.id);
  for (const route of ["dashboard", app.type === "user" ? "user" : "kami", "kami", "blocklist", "notice", "logs/user"]) {
    if (route === "logs/user" && app.type === "kami") continue;
    await page.goto(`${base}/${entry}/${route}`);
    await page.waitForLoadState("networkidle");
    assert.ok(page.url().endsWith(`/${route}`));
    if (route === "logs/user") assert.equal(await page.getByRole("button", { name: "清理日志", exact: true }).count(), 0);
  }
  for (const title of ["应用信息", "扩展配置", "扩展文件", "版本控制", "加密控制", "云端函数", "管理员日志"]) {
    assert.equal(await page.locator(".arco-menu").getByText(title, { exact: true }).count(), 0, title);
  }
  await page.screenshot({ path: path.join(os.tmpdir(), "ordinary-admin-desktop.png"), fullPage: true });
  await page.setViewportSize({ width: 390, height: 844 });
  await page.waitForTimeout(200);
  assert.ok(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth));
  await page.screenshot({ path: path.join(os.tmpdir(), "ordinary-admin-mobile.png"), fullPage: true });
  checks.push("ordinary administrator sees operational pages and cannot see configuration controls or log deletion");
  for (const route of ["system", "app-info", "extend", "extend-file", "version", "encrypt", "function", "admin", "logs/admin"]) {
    await page.goto(`${base}/${entry}/${route}`);
    await page.waitForURL(`**/${entry}/app`);
  }
  checks.push("direct restricted page URLs redirect to application selection");
  assert.deepEqual(errors, []);
  console.log(JSON.stringify({ checks, errors }, null, 2));
} finally {
  if (createdID) await api(`/admin/${createdID}`, { method: "DELETE" });
  await browser.close();
}
