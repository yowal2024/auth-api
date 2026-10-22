import { chromium } from 'playwright';
import { createServer } from 'vite';
import assert from 'node:assert/strict';
const server = await createServer({server:{host:'127.0.0.1',port:18185,strictPort:true}});
await server.listen();
let browser;
const errors=[], checks=[];
try {
  browser=await chromium.launch({executablePath:process.env.CHROME_PATH || 'C:/Program Files/Google/Chrome/Application/chrome.exe',headless:true});
  const context=await browser.newContext();
  await context.addInitScript(()=>localStorage.setItem('auth_token','test-token'));
  let codePosts=0, routePosts=0, lastCodes, settingsSaved;
  let settings={webUrl:'',uploadLimit:10,adminTokenExpire:3600,adminEntry:'admin',mysqlMaxConn:20,mysqlMaxIdle:0,cacheEnabled:1,userLogEnabled:0,adminLogEnabled:1,logLevel:'info'};
  await context.route('**/api/**',async route=>{
    const req=route.request(), path=new URL(req.url()).pathname;
    if(!path.startsWith('/api/'))return route.continue();
    let data={}; let code=0, msg='ok';
    if(path==='/api/install/status')data={installed:true,adminEntry:'admin'};
    else if(path==='/api/admin/system/settings'){
      if(req.method()==='POST'){settingsSaved=req.postDataJSON();settings={...settingsSaved};data={adminEntry:'admin',hotSwap:true};}else data=settings;
    }
    else if(path==='/api/admin/me')data={id:1,username:'admin',role:'super'};
    else if(path==='/api/admin/system/api-routes')data=[{id:1,apiKey:'info',name:'Info',overridePath:'',rewritten:false}];
    else if(path==='/api/admin/system/api-codes')data=[{id:1,code:1006,message:'login required',overrideCode:0,rewritten:false}];
    else if(path==='/api/admin/system/api-routes/save'){routePosts++;code=400;msg='routes rejected';}
    else if(path==='/api/admin/system/api-codes/save'){codePosts++;lastCodes=req.postDataJSON();code=400;msg='codes rejected';}
    else if(path==='/api/admin/log/list')data={list:[],total:0};
    await route.fulfill({contentType:'application/json',body:JSON.stringify({code,msg,data})});
  });
  const page=await context.newPage();page.on('pageerror',e=>errors.push(e.message));
  await page.goto('http://127.0.0.1:18185/admin/system');
  const panel=()=>page.locator('.arco-tabs-content-item-active > .arco-tabs-pane');
  const field=(label)=>page.locator('.arco-form-item').filter({has:page.locator('.arco-form-item-label').filter({hasText:label})});
  await field('数据缓存').locator('.arco-radio-checked').getByText('开启',{exact:true}).waitFor();
  await field('用户日志开关').locator('.arco-radio-checked').getByText('关闭',{exact:true}).waitFor();
  await field('管理员日志开关').locator('.arco-radio-checked').getByText('开启',{exact:true}).waitFor();
  await field('系统日志级别').getByText('全部操作(默认)',{exact:true}).waitFor();
  await field('数据缓存').getByText('关闭',{exact:true}).click();
  await field('用户日志开关').getByText('开启',{exact:true}).click();
  await page.getByRole('button',{name:'保存设置',exact:true}).click();
  await page.getByText('已保存',{exact:true}).waitFor();
  assert.equal(settingsSaved.cacheEnabled,false);assert.equal(settingsSaved.userLogEnabled,true);assert.equal(settingsSaved.adminLogEnabled,true);assert.equal(settingsSaved.logLevel,'all');
  await page.reload();await field('数据缓存').locator('.arco-radio-checked').getByText('关闭',{exact:true}).waitFor();
  await field('用户日志开关').locator('.arco-radio-checked').getByText('开启',{exact:true}).waitFor();
  checks.push('numeric and boolean settings render; untouched flags save as booleans; reload retains values; legacy info maps to all');

  for(const [title,error] of [['API路由','routes rejected'],['API状态码','codes rejected']]){
    await page.locator('.arco-tabs-tab').filter({has:page.getByText(title,{exact:true})}).click();
    const switches=panel().getByRole('switch');
    await switches.first().click();
    await page.getByText(error,{exact:true}).waitFor();
    assert.equal(await switches.first().getAttribute('aria-checked'),'false');
    assert.equal(await switches.last().getAttribute('aria-checked'),'false');
    checks.push(`${title}: rejected bulk toggle restores all switches`);
  }
  await panel().getByRole('button',{name:'自定义重写',exact:true}).click();
  const codeInput=panel().getByPlaceholder('可自定义或一键重定义');
  await codeInput.fill('123abc');
  await panel().getByRole('switch').last().click();
  await panel().getByRole('button',{name:'保存',exact:true}).click();
  await page.getByText(/请填写正整数/).waitFor();
  assert.equal(codePosts,1);
  await codeInput.fill('23456');
  const saved=page.waitForResponse(r=>r.url().endsWith('/api-codes/save'));
  await panel().getByRole('button',{name:'保存',exact:true}).click();
  await saved;
  assert.equal(lastCodes.items[0].overrideCode,23456);
  assert.equal(await codeInput.inputValue(),'23456');
  assert.equal(routePosts,1);
  checks.push('invalid code text makes no request; failed valid save retains the edit');
  assert.deepEqual(errors,[]);console.log(JSON.stringify({checks,errors},null,2));
} finally {await browser?.close();await server.close();}
