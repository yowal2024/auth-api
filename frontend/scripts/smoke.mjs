import { chromium } from 'playwright';
import { createServer } from 'vite';
import fs from 'node:fs';
import assert from 'node:assert/strict';
import { mockRequest } from './fixtures/original-mock.js';
fs.mkdirSync('recovery', {recursive:true});

const server=await createServer({server:{host:'127.0.0.1',port:18181,strictPort:true}});
await server.listen();
let browser;
const results=[];const errors=[];const warnings=[];
const interactions=[];
try {
 browser=await chromium.launch({executablePath:process.env.CHROME_PATH||'C:/Program Files/Google/Chrome/Application/chrome.exe',headless:true});
 const context=await browser.newContext({viewport:{width:1440,height:1000}});
 let installed=true;
 await context.route('**/api/**',async route=>{
  const req=route.request(),url=new URL(req.url());if(!url.pathname.startsWith('/api/'))return route.continue();let data;
  if(url.pathname==='/api/install/status')data={code:0,data:{installed,adminEntry:'admin'}};
  else if(url.pathname.startsWith('/api/install/'))data={code:0,data:{ok:true}};
  else if(url.pathname==='/api/admin/system/https')data={code:0,data:{enabled:false,cert:'',key:'',httpsAddr:':8443'}};
  else if(url.pathname==='/api/admin/extend-file/list')data={code:0,data:[]};
  else if(url.pathname==='/api/admin/app/1/mtls')data={code:0,data:{enabled:false}};
  else if(url.pathname==='/api/admin/extend-file/categories')data={code:0,data:[]};
  else try {let body={};try{body=req.postDataJSON()||{}}catch{}data={code:0,data:await mockRequest(url.pathname.slice(4),{method:req.method(),params:Object.fromEntries(url.searchParams),data:body})}}catch(e){data={code:0,data:[]};warnings.push({path:url.pathname,error:e.message})}
  await route.fulfill({contentType:'application/json',body:JSON.stringify(data)});
 });
 const page=await context.newPage();
 let current='';page.on('pageerror',e=>errors.push({page:current,error:e.message}));page.on('console',m=>{if(m.type()==='warning'&&m.text().includes('[Vue warn]'))warnings.push({page:current,message:m.text().slice(0,600)})});
 await page.addInitScript(()=>{localStorage.setItem('auth_token','local-test-token');localStorage.setItem('auth_current_appid','1')});
 const paths=['/home','/admin/login','/admin/app','/admin/system','/admin/docs','/admin/admin','/admin/profile','/admin/dashboard','/admin/user','/admin/user/group','/admin/user/edit/1','/admin/app-info','/admin/notice','/admin/extend','/admin/extend-file','/admin/version','/admin/encrypt','/admin/blocklist','/admin/kami','/admin/kami/group','/admin/function','/admin/function/edit/0','/admin/logs/user','/admin/logs/admin'];
 for(const pathname of paths){current=pathname;const before=errors.length;await page.goto('http://127.0.0.1:18181'+pathname);await page.waitForTimeout(650);results.push({path:pathname,title:await page.title(),bodyLength:(await page.locator('body').innerText()).length,errors:errors.length-before});assert.ok(results.at(-1).bodyLength>20,'blank page '+pathname);console.log(pathname,results.at(-1).bodyLength,results.at(-1).errors)}
 current='/admin/login interaction';await page.goto('http://127.0.0.1:18181/admin/login');await page.getByPlaceholder('账号',{exact:true}).fill('admin');await page.getByPlaceholder('密码',{exact:true}).fill('123456');await page.getByRole('button',{name:/登\s*录/}).click();await page.waitForURL('**/admin/app');interactions.push('login redirects to app list');
 await page.getByRole('button',{name:'添加应用',exact:true}).click();const modal=page.locator('.arco-modal:visible');await modal.waitFor();assert.ok((await modal.innerText()).includes('应用名称'));await modal.locator('input').first().fill('本地恢复验证应用');await modal.getByRole('button',{name:'确定添加',exact:true}).click();await modal.waitFor({state:'hidden'});await page.locator('.app-name').filter({hasText:'本地恢复验证应用'}).waitFor();interactions.push('create application submits and refreshes');
 for(const pathname of ['/admin/user','/admin/user/group','/admin/kami','/admin/kami/group','/admin/extend','/admin/encrypt','/admin/blocklist']){
  current=pathname+' modal';await page.goto('http://127.0.0.1:18181'+pathname);await page.waitForTimeout(400);
  const trigger=page.getByRole('button',{name:/新增|添加|新建|创建|生成卡密/}).first();if(await trigger.count()){await trigger.click();await page.locator('.arco-modal:visible,.arco-drawer:visible').first().waitFor();await page.waitForTimeout(100);interactions.push('opens form '+pathname)}
 }
 current='/admin/function/edit/0 interaction';await page.goto('http://127.0.0.1:18181/admin/function/edit/0');await page.locator('.monaco-editor').waitFor();assert.ok(await page.locator('.monaco-editor').count());await page.getByRole('button',{name:'试运行',exact:true}).click();await page.waitForTimeout(500);interactions.push('editor mounts and run request completes');
 await page.setViewportSize({width:390,height:844});await page.goto('http://127.0.0.1:18181/admin/app');await page.waitForTimeout(400);await page.screenshot({path:'recovery/mobile-apps.png',fullPage:true});interactions.push('mobile app list renders');
 installed=false;current='/install';await page.goto('http://127.0.0.1:18181/install');await page.waitForTimeout(400);results.push({path:'/install',bodyLength:(await page.locator('body').innerText()).length});
 await page.screenshot({path:'recovery/install-screenshot.png',fullPage:true});
} finally {if(browser)await browser.close();await server.close();fs.writeFileSync('recovery/browser-check.json',JSON.stringify({results,interactions,errors,warnings},null,2));}
console.log(JSON.stringify({errors,warnings},null,2));if(errors.length)process.exitCode=1;
