import { chromium } from "playwright";
import assert from "node:assert/strict";
import { createHmac, randomUUID } from "node:crypto";

const base = process.env.TEST_BASE_URL;
const entry = process.env.TEST_ADMIN_ENTRY || "admin";
assert.ok(base && process.env.TEST_ADMIN_PASSWORD, "Set TEST_BASE_URL and TEST_ADMIN_PASSWORD");
let token, app, browser;
const errors = [], checks = [];
async function api(path, body, method = body === undefined ? "GET" : "POST", sdk = false) {
  const raw = body === undefined ? undefined : JSON.stringify(body);
  const headers = { "Content-Type": "application/json" };
  if (app) headers["X-App-Id"] = String(app.id);
  if (token && !sdk) headers.Authorization = `Bearer ${token}`;
  if (sdk) {
    const ts = String(Math.floor(Date.now() / 1000)), nonce = randomUUID();
    headers["X-Ts"] = ts; headers["X-Nonce"] = nonce; headers["X-Mi"] = body.mi;
    headers["X-Sign"] = createHmac("sha256", app.tokenKey).update(`${ts}\n${nonce}\n${raw || ""}`).digest("hex");
  }
  const response = await fetch(`${base}/api${path}`, { method, headers, body: raw });
  const result = await response.json();
  assert.equal(result.code, 0, `${path}: ${result.msg}`);
  return result.data;
}
try {
  token = (await api(`/${entry}/login`, { username: process.env.TEST_ADMIN_USER || "admin", password: process.env.TEST_ADMIN_PASSWORD })).token;
  app = await api(`/${entry}/app`, { name: `UI verification ${Date.now()}`, type: "user", mode: "free", bindDevice: true });
  const group = await api(`/${entry}/user/group`, { name: "Verification", level: 1 });
  const user = await api(`/${entry}/user`, { username: "uitester", password: "test-pass-123", groupId: group.id });
  const userPath = `/${entry}/user/${user.id}`;
  await api(userPath, { bindLimit: 2 }, "PUT");
  for (const mi of ["test-device-a", "test-device-b"]) await api("/sdk/logon", { username: "uitester", password: "test-pass-123", mi }, "POST", true);
  await api(`/${entry}/app/${app.id}`, { name: app.name, bindRegion: true }, "PUT");
  await api(userPath, { ipList: JSON.stringify([{ ip: "192.0.2.1", time: 1700000000 }, { ip: "192.0.2.2", time: 1700000001 }]) }, "PUT");
  browser = await chromium.launch({ executablePath: process.env.CHROME_PATH || "C:/Program Files/Google/Chrome/Application/chrome.exe", headless: true });
  const context = await browser.newContext();
  await context.addInitScript(({ token, id }) => {
    localStorage.setItem("auth_token", token);
    localStorage.setItem("auth_current_appid", String(id));
  }, { token, id: app.id });
  const page = await context.newPage();
  page.on("pageerror", error => errors.push(error.message));
  const panel = () => page.locator('.arco-tabs-content-item-active > .arco-tabs-pane');
  async function tab(name) {
    const selected = page.locator('.arco-tabs-tab').filter({ has: page.getByText(name, { exact: true }) });
    await selected.click();
    await page.waitForFunction(name => [...document.querySelectorAll('.arco-tabs-tab-active')].some(el => el.textContent.trim() === name), name);
  }
  async function confirm() {
    const popup = page.locator('.arco-popconfirm:visible');
    await popup.getByRole('button', { name: '确定', exact: true }).click();
    await popup.waitFor({ state: 'hidden' });
  }
  await page.goto(`${base}/${entry}/user/edit/${user.id}`);
  await page.getByPlaceholder("请输入昵称").fill("UI verified");
  const saved = page.waitForResponse(r => r.url().endsWith(`/api${userPath}`) && r.request().method() === 'PUT');
  await panel().getByRole('button', { name: '提交', exact: true }).click();
  await saved;
  assert.equal((await api(userPath)).nickname, "UI verified");
  checks.push('basic user information saves to the real database');

  await tab('扩展信息');
  await panel().locator('.add-icon').click();
  await panel().getByPlaceholder('如: name').fill('test_variable');
  await panel().getByPlaceholder('如: 张三').fill('stored value');
  await panel().getByRole('button', { name: '提交保存', exact: true }).click();
  await page.getByText('已保存', { exact: true }).last().waitFor();
  await page.waitForLoadState('networkidle');
  assert.equal((await api(`${userPath}/extends`))[0].value, 'stored value');
  await panel().getByPlaceholder('如: 张三').fill('changed value');
  const extensionSave = page.waitForResponse(r => r.url().endsWith(`${userPath}/extends`) && r.request().method() === 'POST');
  await panel().getByRole('button', { name: '提交保存', exact: true }).click();
  const extensionResponse = await extensionSave;
  assert.equal((await extensionResponse.json()).code, 0);
  await page.waitForLoadState('networkidle');
  assert.equal((await api(`${userPath}/extends`))[0].value, 'changed value');
  await panel().getByRole('button', { name: '删除变量', exact: true }).click();
  await panel().getByPlaceholder('如: name').waitFor({ state: 'hidden' });
  assert.equal((await api(`${userPath}/extends`)).length, 0);
  await page.reload();
  await tab('扩展信息');
  assert.equal(await panel().getByPlaceholder('如: name').count(), 0);
  checks.push('extension variables create, edit and delete persist after refresh');

  await tab('绑定设备');
  await panel().getByText('test-device-a', { exact: true }).waitFor();
  await panel().getByText('解绑', { exact: true }).first().click();
  await confirm();
  await page.waitForLoadState('networkidle');
  assert.equal((await api(`${userPath}/devices`)).length, 1);
  await panel().getByRole('button', { name: '解绑全部', exact: true }).click();
  await confirm();
  await panel().getByText('暂无绑定设备', { exact: true }).waitFor();
  assert.equal((await api(`${userPath}/devices`)).length, 0);
  checks.push('single and all device unbinding reach the backend');

  await tab('IP区域');
  await panel().getByText('移除', { exact: true }).first().click();
  await confirm();
  await page.waitForLoadState('networkidle');
  assert.equal(JSON.parse((await api(userPath)).ipList).length, 1);
  await panel().getByRole('button', { name: '解绑全部', exact: true }).click();
  await confirm();
  await panel().getByText('暂无 IP 地点记录', { exact: true }).waitFor();
  assert.equal(JSON.parse((await api(userPath)).ipList).length, 0);
  await tab('用户日志');
  await panel().locator('.arco-table-tr').first().waitFor();
  assert.ok((await api(`${userPath}/logs`)).total > 0);
  await tab('基本信息');
  assert.equal(await page.getByPlaceholder('请输入昵称').inputValue(), 'UI verified');
  checks.push('IP removal, user logs and returning to basic information work');

  // Exercise other app-scoped actions using this disposable application.
  const encryption = await api(`/${entry}/encrypt`, { name: 'UI encryption', algorithm: 'aes', encoding: 'base64', symKey: '1234567890abcdef', timeSkew: 60, signMode: 'verify' });
  await page.goto(`${base}/${entry}/encrypt`);
  const encryptionCard = page.locator('.encrypt-card').filter({ hasText: 'UI encryption' });
  const encryptionSaved = page.waitForResponse(r => r.url().endsWith('/encrypt') && r.request().method() === 'POST');
  await encryptionCard.getByRole('switch').click();
  assert.equal((await (await encryptionSaved).json()).code, 0);
  assert.equal((await api(`/${entry}/encrypt/list`)).find(e => e.id === encryption.id).signMode, 'none');
  await encryptionCard.locator('.more-btn').click();
  await page.getByText('删除', { exact: true }).last().click();
  await page.locator('.arco-modal:visible').getByRole('button', { name: '确定', exact: true }).click();
  await encryptionCard.waitFor({ state: 'hidden' });
  assert.equal((await api(`/${entry}/encrypt/list`)).length, 0);
  checks.push('encryption signature toggle and deletion persist');

  const fn = await api(`/${entry}/function`, { name: 'ui_manual', trigger: 'manual', code: 'function handler(ctx) { return {answer:42}; }', timeout: 3000, enabled: true });
  await page.goto(`${base}/${entry}/function`);
  const fnRow = page.locator('.arco-table-tr').filter({ hasText: 'ui_manual' });
  const fnSaved = page.waitForResponse(r => r.url().endsWith('/function') && r.request().method() === 'POST');
  await fnRow.getByRole('switch').click();
  assert.equal((await (await fnSaved).json()).code, 0);
  const disabledFn = await api(`/${entry}/function/${fn.id}`);
  assert.equal(disabledFn.enabled, 0);
  assert.equal(disabledFn.code, fn.code);
  await fnRow.getByText('试运行', { exact: true }).click();
  await page.getByText('返回: {"answer":42}', { exact: true }).waitFor();
  await api(`/${entry}/function`, { ...disabledFn, code: 'throw new Error("UI intentional failure")' });
  await page.reload();
  await fnRow.getByText('试运行', { exact: true }).click();
  await page.locator('.arco-message-error').filter({ hasText: 'UI intentional failure' }).waitFor();
  await fnRow.getByText('删除', { exact: true }).click();
  await confirm();
  await fnRow.waitFor({ state: 'hidden' });
  checks.push('cloud function toggle preserves code; run distinguishes success from script failure; delete works');

  await page.goto(`${base}/${entry}/extend-file`);
  await page.getByRole('button', { name: '上传文件', exact: true }).click();
  const uploadModal = page.locator('.arco-modal:visible');
  await uploadModal.locator('input[type="file"]').setInputFiles({ name: 'verification.txt', mimeType: 'text/plain', buffer: Buffer.from('real backend download') });
  await uploadModal.getByRole('button', { name: '上传', exact: true }).click();
  await uploadModal.waitFor({ state: 'hidden' });
  const fileRow = page.locator('.arco-table-tr').filter({ hasText: 'verification.txt' });
  await fileRow.waitFor();
  const download = page.waitForEvent('download');
  await fileRow.getByText('下载', { exact: true }).click();
  const stream = await (await download).createReadStream();
  const chunks = []; for await (const chunk of stream) chunks.push(chunk);
  assert.equal(Buffer.concat(chunks).toString(), 'real backend download');
  await fileRow.getByText('删除', { exact: true }).click();
  await confirm();
  await fileRow.waitFor({ state: 'hidden' });
  assert.equal((await api(`/${entry}/extend-file/list`)).length, 0);
  checks.push('extension file uploads, downloads identical bytes and deletes');

  await page.goto(`${base}/${entry}/system`);
  for (const [name, endpoint] of [['HTTPS / 双向认证', 'https'], ['API路由', 'api-routes'], ['API状态码', 'api-codes'], ['系统日志', null]]) {
    const response = page.waitForResponse(r => r.url().includes(endpoint ? `/system/${endpoint}` : '/log/list'));
    await tab(name);
    assert.equal((await (await response).json()).code, 0);
    assert.ok((await panel().innerText()).length > 10);
  }
  await tab('系统设置');
  checks.push('all system settings tabs switch and load their real APIs');
  assert.deepEqual(errors, []);
  console.log(JSON.stringify({ checks, errors }, null, 2));
} finally {
  await browser?.close();
  if (app) await api(`/${entry}/app/${app.id}`, undefined, 'DELETE');
}
