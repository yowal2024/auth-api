// Real HTTP SDK lifecycle checks. Creates and removes dedicated test applications.
import assert from "node:assert/strict";
import { createHmac, randomBytes } from "node:crypto";
const base = process.env.TEST_BASE_URL || "http://127.0.0.1:18080";
assert.ok(process.env.TEST_ADMIN_PASSWORD, "Set TEST_ADMIN_PASSWORD");
const checks = [], failures = [], apps = [];
async function http(path, method = "GET", payload, headers = {}) {
  const response = await fetch(base + path, { method, headers: { "Content-Type": "application/json", ...headers }, body: payload === undefined ? undefined : JSON.stringify(payload), signal: AbortSignal.timeout(15000) });
  return response.json();
}
const auth = await http("/api/admin/login", "POST", { username: process.env.TEST_ADMIN_USER || "admin", password: process.env.TEST_ADMIN_PASSWORD });
assert.equal(auth.code, 0, auth.msg);
async function admin(app, path, method = "GET", payload, expected = 0) {
  const out = await http("/api/admin" + path, method, payload, { Authorization: "Bearer " + auth.data.token, "X-App-Id": String(app?.id || 0) });
  assert.equal(out.code, expected, `${method} ${path}: ${out.msg}`); return out.data;
}
async function sdk(app, path, payload = {}, token = "", mi = "device-A", expected = 0, extra = {}) {
  const body = JSON.stringify(payload), ts = String(Math.floor(Date.now()/1000)), nonce = randomBytes(16).toString("hex");
  const out = await http("/api/sdk/" + path, "POST", payload, { "X-App-Id": String(app.id), "X-Token": token, "X-Mi": mi, "X-Ts": ts, "X-Nonce": nonce, "X-Sign": createHmac("sha256", app.tokenKey).update(ts+"\n"+nonce+"\n"+body).digest("hex"), ...extra });
  assert.equal(out.code, expected, `${app.type} ${path}: expected ${expected}, got ${out.code} ${out.msg}`); return out.data;
}
async function rawSDK(app, path, method, token, expected) {
  const body=method==="GET"?"":"{}",ts=String(Math.floor(Date.now()/1000)),nonce=randomBytes(16).toString("hex");
  const out=await http("/api/sdk/"+path,method,method==="GET"?undefined:{},{"X-App-Id":String(app.id),"X-Token":token,"X-Mi":"device-A","X-Ts":ts,"X-Nonce":nonce,"X-Sign":createHmac("sha256",app.tokenKey).update(ts+"\n"+nonce+"\n"+body).digest("hex")});
  assert.equal(out.code,expected,`${path}: ${out.msg}`);
}
async function check(name, work) { try { await work(); checks.push(name); console.log("PASS", name); } catch(e) { failures.push({name,error:e.message});console.log("FAIL",name,e.message); } }
async function createApp(type) { const app = await admin(null,"/app","POST",{name:`SDK lifecycle ${type} ${Date.now()}`,type,mode:"free",bindDevice:true,bruteForce:false});apps.push(app);return app; }
try {
 for(const type of ["user","kami"]){
  const app=await createApp(type);let login,id,a;
  if(type==="user"){a=await sdk(app,"register",{username:"multitest",password:"test-pass-123",mi:"device-A"});id=a.user.id;login=mi=>sdk(app,"logon",{username:"multitest",password:"test-pass-123",mi},"",mi);}
  else{const g=await admin(app,"/kami/group","POST",{name:"multi",type:"vip",value:3600});const card=(await admin(app,"/kami/generate","POST",{groupId:g.id,count:1,length:16})).cards[0];login=mi=>sdk(app,"logon",{card,mi},"",mi);a=await login("device-A");id=a.user.id;}
  const path=`/${type}/${id}`;await admin(app,path,"PUT",{bindLimit:1});let b=await login("device-B");
  const online=async expected=>assert.equal((await admin(app,"/dashboard")).onlineNow,expected);
  await check(type+"：一台退出不应让另一台的在线状态消失",async()=>{await sdk(app,"logout",{},a.token);await online(1);await sdk(app,"heartbeat",{},b.token,"device-B");});
  await check(type+"：最后一个客户端退出后在线数为零",async()=>{await sdk(app,"logout",{},b.token,"device-B");await online(0);});
  await check(type+"：同一设备两次登录，一个退出后另一个仍在线",async()=>{const first=await login("device-A"),second=await login("device-A");await sdk(app,"logout",{},first.token);await online(1);await sdk(app,"heartbeat",{},second.token);await sdk(app,"logout",{},second.token);await online(0);});
  await check(type+"：两个客户端同时退出后统计为零",async()=>{const one=await login("device-A"),two=await login("device-B");await Promise.all([sdk(app,"logout",{},one.token),sdk(app,"logout",{},two.token,"device-B")]);await online(0);});
  a=await login("device-A");b=await login("device-B");
  await check(type+"：解绑一台保留另一台在线，全部解绑立即清零",async()=>{const device=(await admin(app,path+"/devices")).find(x=>x.mi==="device-A");await admin(app,path+`/devices/${device.id}`,"DELETE");await online(1);await sdk(app,"info",{},a.token,"device-A",1006);await sdk(app,"heartbeat",{},b.token,"device-B");await admin(app,path+"/devices","DELETE");await online(0);await sdk(app,"info",{},b.token,"device-B",1006);});
  a=await login("device-A");
  await check(type+"：永久封禁立即清理在线记录",async()=>{await admin(app,path,"PUT",{banExpire:-1});try{await online(0);await sdk(app,"heartbeat",{},a.token,"device-A",1002);}finally{await admin(app,path,"PUT",{banExpire:0});}});
  a=await login("device-A");
  if(type==="user")await check("修改密码立即清理在线状态",async()=>{await admin(app,path+"/password","POST",{password:"new-pass-123"});await online(0);await sdk(app,"info",{},a.token,"device-A",1006);});
  await check(type+"：短令牌到期后在线记录同步过期",async()=>{await admin(app,path+"/devices","DELETE");await admin(app,`/app/${app.id}`,"PUT",{name:app.name,userTokenTtl:1});let short;if(type==="user"){short=await sdk(app,"logon",{username:"multitest",password:"new-pass-123",mi:"device-A"});}else{short=await login("device-A");}await new Promise(r=>setTimeout(r,3000));await sdk(app,"info",{},short.token,"device-A",1006);await online(0);});
 }
}finally{for(const app of apps.reverse())await admin(null,`/app/${app.id}`,"DELETE");console.log(JSON.stringify({passed:checks.length,failures},null,2));}
if(failures.length)process.exitCode=1;
