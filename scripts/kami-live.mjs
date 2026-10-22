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
  const app=await createApp("kami"), other=await createApp("kami");
  await check("两种模式独立：卡密模式拒绝注册，用户模式不能直接用卡密登录",async()=>{await sdk(app,"register",{username:"testuser",password:"test-pass-123",mi:"device-A"},"","device-A",400);const userApp=await createApp("user");const g=await admin(userApp,"/kami/group","POST",{name:"user recharge",type:"vip",value:60});const card=(await admin(userApp,"/kami/generate","POST",{groupId:g.id,count:1,length:16})).cards[0];await sdk(userApp,"logon",{card,mi:"device-A"},"","device-A",400);const u=await sdk(userApp,"register",{username:"testuser",password:"test-pass-123",mi:"device-A"});await sdk(userApp,"kamiTopup",{card},u.token);assert.equal((await sdk(userApp,"info",{},u.token)).user.vipActive,true);await sdk(userApp,"kamiTopup",{card},u.token,"device-A",2002);await admin(userApp,`/user/${u.user.id}`,"DELETE");});
  const group=await admin(app,"/kami/group","POST",{name:"card audit",type:"vip",value:3600});
  const generate=async(g=group,count=1,prefix="")=>(await admin(app,"/kami/generate","POST",{groupId:g.id,count,length:16,prefix})).cards;
  const record=async(card)=>(await admin(app,"/kami/list?keyword="+encodeURIComponent(card))).list.find(x=>x.card===card);
  const login=async(card,mi="device-A",code=0)=>sdk(app,"logon",{card,mi},"",mi,code);
  const [card,source,unused]=await generate(group,3,"AUDIT-");
  let first=await login(card), ct=first.token, cid=first.user.id;
  await check("生成卡号唯一、长度、前缀和生成记录",async()=>{const list=await generate(group,20,"TEST-");assert.equal(new Set(list).size,20);assert.ok(list.every(x=>x.startsWith("TEST-")&&x.length===21));const row=await record(list[0]);assert.ok(row.creator&&row.createIp&&row.createTime);});
  await check("非法生成参数和跨应用分组被拒绝",async()=>{for(const v of [{count:0},{count:1001},{length:7},{length:65}])await admin(app,"/kami/generate","POST",{groupId:group.id,count:1,length:16,...v},400);await admin(other,"/kami/generate","POST",{groupId:group.id,count:1,length:16},400);});
  await check("跨应用卡号、令牌不能使用",async()=>{await sdk(other,"logon",{card,mi:"device-A"},"","device-A",2001);await sdk(other,"info",{},ct,"device-A",1006);await sdk(other,"kamiTopup",{card:source},ct,"device-A",1006);});
  await check("同卡并发登录只激活一次，VIP 不重复累加",async()=>{const [v]=await generate();const results=await Promise.all(Array.from({length:6},()=>login(v)));assert.equal(new Set(results.map(x=>x.user.vip)).size,1);const row=await record(v);assert.equal(row.status,1);assert.equal(row.useUser,v);assert.equal((await admin(app,`/kami/${row.id}/devices`)).length,1);});
  await check("并发充值仅成功一次，VIP 精确增加一份",async()=>{const old=(await sdk(app,"info",{},ct)).user.vip;const body={card:source};const results=await Promise.all(Array.from({length:6},async()=>{try{return await sdk(app,"kamiTopup",body,ct);}catch(e){assert.match(e.message,/got 2002/);return null;}}));assert.equal(results.filter(Boolean).length,1);assert.equal((await sdk(app,"info",{},ct)).user.vip,old+3600);assert.equal((await record(source)).useUser,card);});
  await check("已作为充值券消费的卡号不能再登录",async()=>{await login(source,"device-A",2002);});
  await check("冻结未使用卡不能激活或充值，解冻不改变使用状态",async()=>{const row=await record(unused);await admin(app,`/kami/${row.id}`,"PUT",{frozen:true});assert.equal((await record(unused)).status,2);await login(unused,"device-A",2003);await sdk(app,"kamiTopup",{card:unused},ct,"device-A",2002);await admin(app,`/kami/${row.id}`,"PUT",{frozen:false});assert.equal((await record(unused)).status,0);});
  await check("冻结已激活账号拦截旧令牌，解冻不重置 VIP",async()=>{const vip=(await sdk(app,"info",{},ct)).user.vip;await admin(app,`/kami/${cid}`,"PUT",{frozen:true});await sdk(app,"heartbeat",{},ct,"device-A",2003);await login(card,"device-A",2003);await admin(app,`/kami/${cid}`,"PUT",{frozen:false});assert.equal((await login(card)).user.vip,vip);});
  await check("已消费充值券冻结解冻后不能重复使用",async()=>{const row=await record(source);await admin(app,`/kami/${row.id}`,"PUT",{frozen:true});await admin(app,`/kami/${row.id}`,"PUT",{frozen:false});assert.equal((await record(source)).status,1);await sdk(app,"kamiTopup",{card:source},ct,"device-A",2002);});
  await check("无效分组类型、零面值和超限面值被拒绝",async()=>{for(const value of [{type:"anything",value:60},{type:"vip",value:0},{type:"vip",value:315360000001}])await admin(app,"/kami/group","POST",{name:"invalid",...value},400);});
  await check("永久卡能够激活并充值，数值不溢出",async()=>{const permanent=await admin(app,"/kami/group","POST",{name:"permanent",type:"vip",value:315360000000});const values=await generate(permanent,2);const active=await login(values[0]);assert.ok(active.user.vip>315360000000);await sdk(app,"kamiTopup",{card:values[1]},active.token);assert.ok((await sdk(app,"info",{},active.token)).user.vip>630720000000);});
  await check("奖励时长精确增加，负奖励被拒绝",async()=>{const vip=(await sdk(app,"info",{},ct)).user.vip;await admin(app,`/kami/${cid}/reward`,"POST",{seconds:60});assert.equal((await sdk(app,"info",{},ct)).user.vip,vip+60);await admin(app,`/kami/${cid}/reward`,"POST",{seconds:-1},400);});
  await check("VIP 到期拦截，现有令牌可以充值续期",async()=>{await admin(app,`/app/${app.id}`,"PUT",{name:app.name,mode:"vip"});await admin(app,`/kami/${cid}`,"PUT",{vip:1});await sdk(app,"info",{},ct,"device-A",1018);await login(card,"device-A",1018);const [v]=await generate();await sdk(app,"kamiTopup",{card:v},ct);assert.equal((await sdk(app,"info",{},ct)).user.vipActive,true);});
  await check("后台卡密数据不能跨应用查看或修改",async()=>{await admin(other,`/kami/${cid}`,"GET",undefined,404);await admin(other,`/kami/${cid}`,"PUT",{frozen:true},404);await admin(other,`/kami/${cid}/devices`,"GET",undefined,404);await sdk(app,"heartbeat",{},ct);});
  await check("卡密资料与扩展变量存取",async()=>{await sdk(app,"updateInfo",{nickname:"卡密测试"},ct);assert.equal((await admin(app,`/kami/${cid}`)).nickname,"卡密测试");await sdk(app,"userExtends",{set:{audit:"first"}},ct);await sdk(app,"userExtends",{set:{audit:"second"}},ct);assert.equal((await sdk(app,"userExtends",{},ct)).extends.audit,"second");});
  await check("冻结充值券不能充值，解冻后恰好消费一次",async()=>{const [v]=await generate();const row=await record(v);await admin(app,`/kami/${row.id}`,"PUT",{frozen:true});await sdk(app,"kamiTopup",{card:v},ct,"device-A",2002);await admin(app,`/kami/${row.id}`,"PUT",{frozen:false});await sdk(app,"kamiTopup",{card:v},ct);await sdk(app,"kamiTopup",{card:v},ct,"device-A",2002);});
  await check("删除卡密分组后旧令牌失效且在线统计清理",async()=>{const g=await admin(other,"/kami/group","POST",{name:"delete group",type:"vip",value:60});const v=(await admin(other,"/kami/generate","POST",{groupId:g.id,count:1,length:16})).cards[0];const active=await sdk(other,"logon",{card:v,mi:"device-A"});await admin(other,`/kami/group/${g.id}`,"DELETE");await sdk(other,"info",{},active.token,"device-A",1006);assert.equal((await admin(other,"/dashboard")).onlineNow,0);});
  await check("删除整个应用也立即清理在线账号",async()=>{const extra=await createApp("kami");const g=await admin(extra,"/kami/group","POST",{name:"delete app",type:"vip",value:60});const v=(await admin(extra,"/kami/generate","POST",{groupId:g.id,count:1,length:16})).cards[0];const active=await sdk(extra,"logon",{card:v,mi:"device-A"});const before=(await admin(null,"/dashboard")).onlineNow;await admin(null,`/app/${extra.id}`,"DELETE");apps.splice(apps.indexOf(extra),1);await sdk(extra,"info",{},active.token,"device-A",403);assert.equal((await admin(null,"/dashboard")).onlineNow,before-1);});
} finally {
  for(const app of apps.reverse()) await admin(null,`/app/${app.id}`,"DELETE");
  console.log(JSON.stringify({passed:checks.length,failures,temporaryAppsDeleted:apps.map(a=>a.id)},null,2));
}
if(failures.length)process.exitCode=1;
