import { chromium } from "playwright";
import assert from "node:assert/strict";

const base = process.env.TEST_BASE_URL;
const entry = process.env.TEST_ADMIN_ENTRY || "admin";
assert.ok(base && process.env.TEST_ADMIN_PASSWORD, "Set TEST_BASE_URL and TEST_ADMIN_PASSWORD");
const login = await (await fetch(`${base}/api/${entry}/login`, {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({ username: process.env.TEST_ADMIN_USER || "admin", password: process.env.TEST_ADMIN_PASSWORD }),
})).json();
assert.equal(login.code, 0, login.msg);
const browser = await chromium.launch({
  executablePath: process.env.CHROME_PATH || "C:/Program Files/Google/Chrome/Application/chrome.exe",
  headless: true,
});
try {
  const context = await browser.newContext({ viewport: { width: 1484, height: 950 } });
  await context.addInitScript(token => localStorage.setItem("auth_token", token), login.data.token);
  const page = await context.newPage();
  const errors = [];
  page.on("pageerror", error => errors.push(error.message));
  const response = page.waitForResponse(r => new URL(r.url()).pathname === `/api/${entry}/dashboard`);
  await page.goto(`${base}/${entry}/dashboard`);
  const result = await (await response).json();
  assert.equal(result.code, 0);
  const recent = result.data.recent;
  assert.ok(recent.length > 0 && recent.length <= 8, "Use an instance with existing activity logs");
  await page.locator(".recent-timeline .arco-timeline-item").first().waitFor();
  const rows = page.locator(".recent-timeline .arco-timeline-item");
  assert.equal(await rows.count(), recent.length);
  for (let i = 0; i < recent.length; i++) {
    assert.ok(await rows.nth(i).getByText(recent[i].text, { exact: true }).isVisible());
    assert.ok(await rows.nth(i).getByText(recent[i].time, { exact: true }).isVisible());
    assert.ok(["primary", "success"].includes(recent[i].type));
  }
  assert.equal(await page.locator(".recent-empty").count(), 0);
  assert.deepEqual(errors, []);
  console.log(JSON.stringify({ recentRows: recent.length, renderedInAPIOrder: true, errors }));
} finally {
  await browser.close();
}
