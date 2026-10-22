import { chromium } from "playwright";
import assert from "node:assert/strict";
const base = process.env.TEST_BASE_URL || "http://127.0.0.1:18189";
const browser = await chromium.launch({ executablePath: process.env.CHROME_PATH || "C:/Program Files/Google/Chrome/Application/chrome.exe", headless: true });
try {
  const page = await browser.newPage({ viewport: { width: 1440, height: 1000 } });
  const errors = [];
  page.on("pageerror", e => errors.push(e.message));
  await page.goto(base + "/home");
  await page.locator(".hero-actions .arco-btn-primary").waitFor();
  assert.equal(await page.locator(".feature-grid .arco-card").count(), 3);
  assert.equal(await page.locator(".architecture-flow .arco-card").count(), 4);
  assert.ok(await page.locator(".feature-tags .arco-tag").count());
  for (const selector of [".feature-grid .feature-tags", ".architecture-flow .feature-tags"]) {
    const tops = await page.locator(selector).evaluateAll(els => els.map(el => el.getBoundingClientRect().top));
    assert.ok(Math.max(...tops) - Math.min(...tops) < 2, "card tag rows must align: " + selector);
  }
  assert.equal(await page.locator('a[href*="/admin"], a[href*="/docs"], a[href*="/login"]').count(), 0);
  await page.screenshot({ path: process.env.TEMP + "/auth-home-arco.png", fullPage: true });
  await page.getByRole("link", { name: "探索产品能力" }).click();
  await page.waitForURL("**#capabilities");
  await page.getByRole("link", { name: "了解平台架构" }).click();
  await page.waitForURL("**#platform");
  for (const width of [360, 390, 768, 1024, 1440]) {
    await page.setViewportSize({ width, height: 900 });
    assert.ok(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth), "page overflow at " + width);
    assert.ok(await page.locator(".feature,.architecture-card").evaluateAll(els => els.every(el => el.scrollWidth <= el.clientWidth)), "card content overflow at " + width);
  }
  await page.evaluate(() => { document.body.setAttribute("arco-theme", "dark"); window.scrollTo(0, 0); });
  await page.waitForTimeout(150);
  assert.ok((await page.locator(".platform-visual").evaluate(el => getComputedStyle(el).backgroundImage)).includes("28, 41, 68"));
  await page.screenshot({ path: process.env.TEMP + "/auth-home-arco-dark.png", fullPage: true });
  assert.deepEqual(errors, []);
  console.log("PASS Arco buttons/cards/tags, section navigation, hidden entry links, mobile overflow and page errors");
} finally { await browser.close(); }
