import { chromium } from "playwright";
import assert from "node:assert/strict";
const base = process.env.TEST_BASE_URL || "http://127.0.0.1:18189";
const api = process.env.TEST_API_URL || "http://127.0.0.1:18080";
assert.ok(process.env.TEST_ADMIN_PASSWORD, "Set TEST_ADMIN_PASSWORD");
const login = await (
  await fetch(api + "/api/admin/login", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      username: "admin",
      password: process.env.TEST_ADMIN_PASSWORD,
    }),
  })
).json();
assert.equal(login.code, 0);
const browser = await chromium.launch({
  executablePath: "C:/Program Files/Google/Chrome/Application/chrome.exe",
  headless: true,
});
try {
  const context = await browser.newContext({
    viewport: { width: 1480, height: 980 },
    permissions: ["clipboard-read", "clipboard-write"],
  });
  await context.addInitScript(
    (token) => localStorage.setItem("auth_token", token),
    login.data.token,
  );
  const page = await context.newPage();
  const errors = [];
  page.on("pageerror", (e) => errors.push(e.message));
  await page.goto(base + "/admin/docs");
  await page.locator(".overview-title").waitFor();
  assert.ok(await page.locator(".doc-pagination button").first().evaluate(el => {
    const label = el.querySelector("small").getBoundingClientRect();
    const title = el.querySelector("strong").getBoundingClientRect();
    return title.top >= label.bottom;
  }), "pagination label and title must be on separate lines");
  assert.ok(await page.locator(".docs-nav .arco-menu").count());
  assert.ok((await page.locator(".nav-item").first().boundingBox()).height < 50);
  assert.ok(await page.locator(".nav-search .arco-input-wrapper").count());
  assert.ok(await page.locator(".docs-card.arco-card").count());
  assert.ok(await page.locator(".reading-toolbar .arco-breadcrumb").count());
  await page.getByRole("textbox", { name: "搜索文档" }).fill("/cloudFunction");
  assert.equal(await page.locator(".nav-item").count(), 1);
  await page.locator(".nav-item").click();
  await page.waitForURL("**section=cloudFunction");
  await page.getByRole("button", { name: "清空搜索", exact: true }).click();
  assert.ok((await page.locator(".nav-item").count()) > 17);
  await page
    .getByRole("textbox", { name: "搜索文档" })
    .fill("not-a-real-endpoint");
  await page.locator(".nav-empty").waitFor();
  await page.locator(".nav-empty button").click();
  await page.goto(base + "/admin/docs?section=logon");
  await page.locator(".endpoint-strip").waitFor();
  assert.ok(await page.locator(".doc-table .arco-table-container").count());
  assert.ok((await page.locator(".docs-outline .arco-menu-item").count()) >= 3);
  await page.locator(".docs-outline .arco-menu-item").last().click();
  assert.ok(
    (await page.locator(".docs-content").evaluate((el) => el.scrollTop)) > 0,
  );
  const block = page.locator(".doc-code:visible").first();
  await block.getByRole("button", { name: "复制", exact: true }).click();
  assert.equal(
    await page.evaluate(() => navigator.clipboard.readText()),
    await block.locator("pre").textContent(),
  );
  await block.getByText("已复制", { exact: true }).waitFor();
  await page.locator(".docs-content").evaluate((el) => (el.scrollTop = 0));
  await page.screenshot({ path: process.env.TEMP + "/auth-doc-ui-api.png" });
  await page.locator(".nav-item").filter({ hasText: "云函数开发指南" }).click();
  await page.waitForURL("**section=cloud");
  await page.goBack();
  await page.locator(".endpoint-strip").waitFor();
  assert.ok(page.url().includes("section=logon"));
  await page.setViewportSize({ width: 390, height: 844 });
  await page.reload();
  await page.locator(".panel-title").waitFor();
  assert.equal(await page.locator(".api-facts tr").count(), 4, "mobile API facts must use one pair per row");
  assert.ok(await page.locator(".doc-table").first().evaluate(el => {
    const scroller = el.querySelector(".arco-table-content");
    scroller.scrollLeft = 200;
    return scroller.scrollLeft > 0;
  }), "mobile parameter table must scroll horizontally");
  assert.equal(await page.locator(".nav-scroll").isVisible(), false);
  await page.getByRole("button", { name: "展开目录 +" }).click();
  await page.locator(".nav-item").filter({ hasText: "云函数开发指南" }).click();
  await page.locator("#cloud-doc-title").waitFor();
  assert.equal(await page.locator(".nav-scroll").isVisible(), false);
  await page.screenshot({ path: process.env.TEMP + "/auth-doc-ui-mobile.png" });
  assert.ok(
    await page.evaluate(
      () => document.documentElement.scrollWidth <= innerWidth,
    ),
  );
  await page.setViewportSize({ width: 1480, height: 980 });
  await page.evaluate(() => document.body.setAttribute("arco-theme", "dark"));
  await page.waitForTimeout(150);
  await page.screenshot({ path: process.env.TEMP + "/auth-doc-ui-dark.png" });
  await page.goto(base + "/admin/docs?section=client");
  await page.locator(".sdk-guide .arco-alert").waitFor();
  assert.ok(await page.locator(".sdk-guide .arco-table").count());
  await page.goto(base + "/admin/docs?section=profile");
  await page.locator(".doc-table").waitFor();
  await page.goto(base + "/admin/docs?section=invalid-section");
  await page.locator(".overview-title").waitFor();
  assert.deepEqual(errors, []);
  console.log(
    "PASS search, empty state, outline navigation, exact clipboard copy, back history, mobile menu, overflow, page errors",
  );
} finally {
  await browser.close();
}
