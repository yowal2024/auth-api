import { chromium } from 'playwright';
import { createServer } from 'vite';
import fs from 'node:fs';
import path from 'node:path';
import assert from 'node:assert/strict';

assert.ok(process.env.API_PROXY_TARGET || process.env.TEST_BASE_URL, 'requires isolated integration backend');
fs.mkdirSync('recovery', {recursive:true});
if (process.env.TEST_REPORT_PATH) fs.mkdirSync(path.dirname(process.env.TEST_REPORT_PATH), {recursive:true});
const server=process.env.TEST_BASE_URL ? null : await createServer({server:{host:'127.0.0.1',port:18183,strictPort:true}});
if (server) await server.listen();
const base=process.env.TEST_BASE_URL || 'http://127.0.0.1:18183';
const entry=process.env.TEST_ADMIN_ENTRY || 'private';
const entryPath=value=>value.replace(/^\/private(?=\/|$)/,`/${entry}`);
let browser;const errors=[];const apiErrors=[];const pages=[];const pending=[];
try {
 browser=await chromium.launch({executablePath:process.env.CHROME_PATH||'C:/Program Files/Google/Chrome/Application/chrome.exe',headless:true});
 const context=await browser.newContext({viewport:{width:1440,height:1000}});
 const page=await context.newPage();
 page.on('pageerror',e=>errors.push(e.message));
 page.on('response',res=>{if(new URL(res.url()).pathname.startsWith('/api/')) pending.push((async()=>{try{const v=await res.json();if(v.code!==0)apiErrors.push({url:res.url(),code:v.code,msg:v.msg})}catch{}})())});
 await page.goto(base+entryPath('/private/login'));
 await page.getByPlaceholder('账号',{exact:true}).fill(process.env.TEST_ADMIN_USER || 'admin');await page.getByPlaceholder('密码',{exact:true}).fill(process.env.TEST_ADMIN_PASSWORD || 'recovery-pass-123');
 await page.getByRole('button',{name:/登\s*录/}).click();await page.waitForURL(`**/${entry}/app`);
 await page.evaluate(id=>localStorage.setItem('auth_current_appid',id),process.env.TEST_APP_ID);
 for(const path of ['/private/app','/private/dashboard','/private/user','/private/user/group','/private/kami','/private/kami/group','/private/notice','/private/extend','/private/extend-file','/private/version','/private/encrypt','/private/blocklist','/private/function','/private/function/edit/0','/private/logs/user','/private/logs/admin','/private/system','/private/admin','/private/profile','/private/docs']){
  await page.goto(base+entryPath(path));await page.waitForLoadState('networkidle');
  const text=await page.locator('body').innerText();assert.ok(text.length>40,'blank page '+path);pages.push({path,characters:text.length});
 }
 await page.goto(base+entryPath('/private/dashboard'));await page.waitForLoadState('networkidle');
 await page.screenshot({path:'recovery/live-backend-dashboard.png',fullPage:true});
 await Promise.allSettled(pending);
} finally {if(browser)await browser.close();if(server)await server.close();fs.writeFileSync(process.env.TEST_REPORT_PATH || 'recovery/live-backend-check.json',JSON.stringify({pages,errors,apiErrors},null,2));}
console.log(JSON.stringify({pages:pages.length,errors,apiErrors}));assert.equal(errors.length,0);assert.equal(apiErrors.length,0);
