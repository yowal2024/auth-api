import { chromium } from 'playwright';
import { createServer } from 'vite';
import assert from 'node:assert/strict';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
const server = await createServer({server:{host:'127.0.0.1',port:18187,strictPort:true}});
await server.listen();
let browser;
const errors=[], checks=[];
try {
 browser=await chromium.launch({executablePath:process.env.CHROME_PATH || 'C:/Program Files/Google/Chrome/Application/chrome.exe',headless:true});
 const context=await browser.newContext();
 await context.addInitScript(()=>localStorage.setItem('auth_token','test-token'));
 await context.route('**/api/**',async route=>{
  const path=new URL(route.request().url()).pathname;
  if(!path.startsWith('/api/'))return route.continue();
  let data={};
  if(path==='/api/install/status')data={installed:true,adminEntry:'admin'};
  else if(path==='/api/admin/me')data={id:1,username:'admin',role:'super'};
  else if(path==='/api/admin/app/list')data=[{id:1000,name:'Demo',type:'user'}];
  else if(path==='/api/admin/dashboard') {
   const days=Number(new URL(route.request().url()).searchParams.get('range') || 7);
   data={users:1,usersToday:1,kamiAvail:0,kamiUsedToday:0,onlineNow:0,onlinePeakToday:0,recent:[],trend:{dates:Array.from({length:days},(_,i)=>new Date(Date.UTC(2026,9,12-days+1+i)).toISOString().slice(5,10)),logins:Array(days).fill(0),regs:Array.from({length:days},(_,i)=>i===days-1?1:0)}};
  }
  await route.fulfill({contentType:'application/json',body:JSON.stringify({code:0,msg:'ok',data})});
 });
 const page=await context.newPage();page.on('pageerror',e=>errors.push(e.message));
 for(const dark of [false,true]) for(const width of [1484,768,390]) {
  await page.setViewportSize({width,height:900});
  await page.goto('http://127.0.0.1:18187/admin/dashboard');
  await page.locator('.chart-body canvas').waitFor();
  await page.evaluate(dark=>document.body.setAttribute('arco-theme',dark?'dark':'light'),dark);
  await page.waitForTimeout(400);
  const bounds=await page.evaluate(async()=>{
   const url=performance.getEntriesByType('resource').find(x=>x.name.includes('/.vite/deps/echarts_core.js')).name; const {getInstanceByDom}=await import(url);
   const chart=getInstanceByDom(document.querySelector('.chart-body'));
   const legend=chart.getViewOfComponentModel(chart.getModel().getComponent('legend')).group;
   const rect=legend.getBoundingRect().clone();rect.applyTransform(legend.getComputedTransform());
   const grid=chart.getModel().getComponent('grid').coordinateSystem.getRect();
   const ticks=chart.getModel().getComponent('yAxis').axis.scale.getTicks().map(x=>x.value);
   return {legend:{x:rect.x,y:rect.y,width:rect.width,height:rect.height},grid:{x:grid.x,y:grid.y},ticks,width:chart.getWidth()};
  });
  assert.ok(bounds.legend.y+bounds.legend.height<bounds.grid.y,JSON.stringify(bounds));
  assert.ok(bounds.legend.x>=0 && bounds.legend.x+bounds.legend.width<=bounds.width,JSON.stringify(bounds));
  assert.ok(bounds.ticks.every(Number.isInteger),JSON.stringify(bounds));
  await page.locator('.chart-body').screenshot({path:join(tmpdir(),`auth-dashboard-${width}-${dark ? "dark" : "light"}.png`)});
  checks.push({viewportWidth:width,dark,...bounds});
  for(const days of [30,7]) {
   const response=page.waitForResponse(r=>new URL(r.url()).pathname==='/api/admin/dashboard' && new URL(r.url()).searchParams.get('range')===String(days));
   await page.getByText(`${days} 天`,{exact:true}).click();
   const result=await (await response).json();assert.equal(result.data.trend.dates.length,days);
   await page.waitForFunction(days=>document.querySelector('.arco-card-header-title')?.textContent.includes(`近 ${days} 天登录与注册趋势`),days);
   const lengths=await page.evaluate(async()=>{const url=performance.getEntriesByType('resource').find(x=>x.name.includes('/.vite/deps/echarts_core.js')).name;const {getInstanceByDom}=await import(url);const option=getInstanceByDom(document.querySelector('.chart-body')).getOption();return [option.xAxis[0].data.length,...option.series.map(x=>x.data.length)];});
   assert.deepEqual(lengths,[days,days,days]);
   assert.ok(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth));
  }

 }
 assert.deepEqual(errors,[]);console.log(JSON.stringify({checks,errors},null,2));
} finally {await browser?.close();await server.close();}


