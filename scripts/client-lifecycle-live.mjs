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
  const app=await createApp("user"), pass="Sdk-test-123", name="liveuser";
  const reg=await sdk(app,"register",{username:name,password:pass,mi:"device-A"});const uid=reg.user.id;
  let token=reg.token;
  const login=async (mi="device-A",password=pass,expected=0)=>sdk(app,"logon",{username:name,password,mi},"",mi,expected);
  await check("注册、登录、心跳、用户信息",async()=>{token=(await login()).token;assert.equal((await sdk(app,"info",{},token)).user.id,uid);assert.equal((await sdk(app,"heartbeat",{},token)).ok,true);});
  await check("设备/IP/注册和登录记录",async()=>{const user=await admin(app,`/user/${uid}`), devices=await admin(app,`/user/${uid}/devices`), logs=(await admin(app,`/user/${uid}/logs`)).list;assert.equal(devices.length,1);assert.equal(devices[0].mi,"device-A");assert.ok(user.registerIp && user.lastIp && devices[0].ip);assert.ok(user.lastLoginTime>0 && devices[0].lastTime>0);assert.ok(logs.some(x=>x.action==="register") && logs.some(x=>x.action==="login"));console.log("observed IP",user.lastIp);});
  await check("错误密码、缺机器码、伪造设备头被拒绝",async()=>{await login("device-A","wrong-pass",1001);await login("",pass,1011);await sdk(app,"info",{},token,"",1006);await sdk(app,"heartbeat",{},token,"device-B",1006);});
  await check("X-Forwarded-For 不能伪造记录 IP",async()=>{await sdk(app,"logon",{username:name,password:pass,mi:"device-A"},"","device-A",0,{"X-Forwarded-For":"8.8.8.8","X-Real-IP":"8.8.8.8"});assert.notEqual((await admin(app,`/user/${uid}`)).lastIp,"8.8.8.8");});
  let tokenB;
  await check("设备上限、增加名额、两台设备独立令牌",async()=>{await login("device-B",pass,1008);await admin(app,`/user/${uid}`,"PUT",{bindLimit:1});tokenB=(await login("device-B")).token;assert.equal((await admin(app,`/user/${uid}/devices`)).length,2);await sdk(app,"heartbeat",{},token);await sdk(app,"heartbeat",{},tokenB,"device-B");await login("device-C",pass,1008);});
  await check("移除单台设备后其令牌失效，另一台保持登录",async()=>{const dev=(await admin(app,`/user/${uid}/devices`)).find(x=>x.mi==="device-B");await admin(app,`/user/${uid}/devices/${dev.id}`,"DELETE");await sdk(app,"heartbeat",{},tokenB,"device-B",1006);await sdk(app,"heartbeat",{},token);});
  await check("设备重新绑定后旧令牌不能复活（包含解绑期间未请求）",async()=>{const old=(await login("device-B")).token;const dev=(await admin(app,`/user/${uid}/devices`)).find(x=>x.mi==="device-B");await admin(app,`/user/${uid}/devices/${dev.id}`,"DELETE");tokenB=(await login("device-B")).token;await sdk(app,"heartbeat",{},old,"device-B",1006);await sdk(app,"heartbeat",{},tokenB,"device-B");});
  await check("封禁后旧令牌不能访问、不能登录",async()=>{await admin(app,`/user/${uid}`,"PUT",{banExpire:-1});await sdk(app,"info",{},token,"device-A",1002);await login("device-A",pass,1002);await admin(app,`/user/${uid}`,"PUT",{banExpire:0});token=(await login()).token;});
  await check("机器码和 IP 黑名单",async()=>{for(const [type,value,code] of [["mi","device-A",1021],["ip",(await admin(app,`/user/${uid}`)).lastIp,1003]]){const block=await admin(app,"/block","POST",{type,value,reason:"lifecycle test"});try{await sdk(app,"info",{},token,"device-A",code);await login("device-A",pass,code);}finally{await admin(app,`/block/${block.id}`,"DELETE");}}});
  await check("资料和扩展信息前后端一致",async()=>{await sdk(app,"updateInfo",{nickname:"测试昵称"},token);await sdk(app,"userExtends",{set:{test_key:"test_value"}},token);assert.equal((await admin(app,`/user/${uid}`)).nickname,"测试昵称");assert.ok((await admin(app,`/user/${uid}/extends`)).some(x=>x.key==="test_key"&&x.value==="test_value"));});
  await check("用户 VIP 到期拦截、充值恢复且不能重复消费",async()=>{await admin(app,`/app/${app.id}`,"PUT",{name:app.name,mode:"vip"});await sdk(app,"heartbeat",{},token,"device-A",1018);await login("device-A",pass,1018);const group=await admin(app,"/kami/group","POST",{name:"vip test",type:"vip",value:3600});const card=(await admin(app,"/kami/generate","POST",{groupId:group.id,count:1,length:16})).cards[0];await sdk(app,"kamiTopup",{card},token);await sdk(app,"kamiTopup",{card},token,"device-A",2002);assert.equal((await sdk(app,"info",{},token)).user.vipActive,true);assert.ok((await admin(app,`/user/${uid}/logs`)).list.some(x=>x.action==="kamiTopup"));await admin(app,`/app/${app.id}`,"PUT",{name:app.name,mode:"free"});});
  await check("修改密码后所有旧令牌失效",async()=>{await admin(app,`/user/${uid}/password`,"POST",{password:"Sdk-new-123"});await sdk(app,"info",{},token,"device-A",1006);await login("device-A",pass,1001);token=(await login("device-A","Sdk-new-123")).token;});
  await check("修改密码同时废止第二台设备令牌",async()=>{await sdk(app,"heartbeat",{},tokenB,"device-B",1006);});
  await check("退出后令牌失效且登出日志存在",async()=>{await sdk(app,"logout",{},token);await sdk(app,"info",{},token,"device-A",1006);const logs=(await admin(app,`/user/${uid}/logs`)).list;assert.ok(logs.some(x=>x.action==="logout"),"missing logout log");});
  token=(await login("device-A","Sdk-new-123")).token;
  tokenB=(await login("device-B","Sdk-new-123")).token;
  await check("删除用户后全部受保护接口拒绝旧令牌",async()=>{await admin(app,`/user/${uid}`,"DELETE");for(const path of ["info","heartbeat","logout","updateInfo","userExtends","kamiTopup","cloudFunction"])await sdk(app,path,{},token,"device-A",1006);await login("device-A","Sdk-new-123",1001);});
  await check("删除后立即从在线统计移除",async()=>assert.equal((await admin(app,"/dashboard")).onlineNow,0));
  await check("删除账号后头像、签名和克隆接口也拒绝旧令牌",async()=>{for(const [path,method] of [["uploadAvatar","POST"],["regen-sign","POST"],["clone","GET"]])await rawSDK(app,path,method,token,1006);});
  await check("删除用户同时使第二台设备下线",async()=>{await sdk(app,"heartbeat",{},tokenB,"device-B",1006);});
  await check("同名重新注册不能使用旧令牌，批量删除立即下线",async()=>{const fresh=await sdk(app,"register",{username:name,password:pass,mi:"device-A"});assert.notEqual(fresh.user.id,uid);await sdk(app,"info",{},token,"device-A",1006);await admin(app,"/user/batch-delete","POST",{ids:[fresh.user.id]});await sdk(app,"heartbeat",{},fresh.token,"device-A",1006);assert.equal((await admin(app,"/dashboard")).onlineNow,0);});
  const cardApp=await createApp("kami");const group=await admin(cardApp,"/kami/group","POST",{name:"test",type:"vip",value:3600});const cards=(await admin(cardApp,"/kami/generate","POST",{groupId:group.id,count:2,length:16})).cards;
  let ct,cid;
  await check("卡密首次激活、重复登录、设备/IP与日志",async()=>{const first=await sdk(cardApp,"logon",{card:cards[0],mi:"device-A"});ct=first.token;cid=first.user.id;const second=await sdk(cardApp,"logon",{card:cards[0],mi:"device-A"});assert.equal(second.user.vip,first.user.vip);assert.equal((await admin(cardApp,`/kami/${cid}/devices`)).length,1);assert.ok((await admin(cardApp,`/kami/${cid}`)).lastIp);const logs=(await admin(cardApp,"/log/list?category=user")).list;assert.ok(logs.some(x=>x.username===cards[0]&&x.action==="login"));});
  await check("卡密设备上限、解绑后旧令牌失效",async()=>{await sdk(cardApp,"logon",{card:cards[0],mi:"device-B"},"","device-B",1008);await admin(cardApp,`/kami/${cid}/unbind-mi`,"POST",{});await sdk(cardApp,"info",{},ct,"device-A",1006);});
  await check("卡密重新绑定后旧令牌不能复活（包含解绑期间未请求）",async()=>{const old=(await sdk(cardApp,"logon",{card:cards[0],mi:"device-A"})).token;await admin(cardApp,`/kami/${cid}/unbind-mi`,"POST",{});await sdk(cardApp,"logon",{card:cards[0],mi:"device-A"});await sdk(cardApp,"info",{},old,"device-A",1006);});
  ct=(await sdk(cardApp,"logon",{card:cards[0],mi:"device-A"})).token;
  await check("卡密永久封禁和解封",async()=>{await admin(cardApp,`/kami/${cid}`,"PUT",{banExpire:-1});await sdk(cardApp,"heartbeat",{},ct,"device-A",1002);await sdk(cardApp,"logon",{card:cards[0],mi:"device-A"},"","device-A",1002);await admin(cardApp,`/kami/${cid}`,"PUT",{banExpire:0});await sdk(cardApp,"heartbeat",{},ct);});
  await check("卡密充值仅能使用一次",async()=>{await sdk(cardApp,"kamiTopup",{card:cards[1]},ct);await sdk(cardApp,"kamiTopup",{card:cards[1]},ct,"device-A",2002);});
  await check("删除卡密后旧令牌失效且在线数清零",async()=>{await admin(cardApp,`/kami/${cid}`,"DELETE");await sdk(cardApp,"heartbeat",{},ct,"device-A",1006);assert.equal((await admin(cardApp,"/dashboard")).onlineNow,0);});
  await check("批量清理已用卡密也立即下线",async()=>{const card=(await admin(cardApp,"/kami/generate","POST",{groupId:group.id,count:1,length:16})).cards[0];const active=await sdk(cardApp,"logon",{card,mi:"device-A"});await admin(cardApp,"/kami/clean-used","POST",{});await sdk(cardApp,"info",{},active.token,"device-A",1006);assert.equal((await admin(cardApp,"/dashboard")).onlineNow,0);});
  await check("短期令牌到期后被拒绝",async()=>{assert.equal((await admin(app,`/app/${app.id}`,"PUT",{name:app.name,userTokenTtl:1})).userTokenTtl,1);const fresh=await sdk(app,"register",{username:"expireuser",password:pass,mi:"device-A"});await new Promise(resolve=>setTimeout(resolve,3000));await sdk(app,"heartbeat",{},fresh.token,"device-A",1006);await admin(app,`/user/${fresh.user.id}`,"DELETE");});
} finally {
  for(const app of apps.reverse())await admin(null,`/app/${app.id}`,"DELETE");
  console.log(JSON.stringify({passed:checks.length,failures,temporaryAppsDeleted:apps.map(a=>a.id)},null,2));
}
if(failures.length)process.exitCode=1;
