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
 const app=await createApp("user");const user=await sdk(app,"register",{username:"audittest",password:"test-pass-123",mi:"device-A"});const uid=user.user.id,token=user.token;
 await check("后台拒绝重复扩展键，客户端读写一致",async()=>{await admin(app,`/user/${uid}/extends`,"POST",{key:"theme",value:"first"});await admin(app,`/user/${uid}/extends`,"POST",{key:"theme",value:"second"},400);await sdk(app,"userExtends",{set:{theme:"third"}},token);assert.equal((await sdk(app,"userExtends",{},token)).extends.theme,"third");});
 await check("批量保存重复扩展键时整批回滚",async()=>{await admin(app,`/user/${uid}/extends`,"POST",{list:[{key:"dupe",value:"1"},{key:"DUPE",value:"2"}]},400);assert.ok(!(await admin(app,`/user/${uid}/extends`)).some(x=>x.key.toLowerCase()==="dupe"));});
 await check("不存在的扩展变量 ID 不能伪报保存成功",async()=>{await admin(app,`/user/${uid}/extends`,"POST",{id:999999999,key:"missing",value:"x"},404);});
 await check("扩展键长度和空白名称返回参数错误",async()=>{await admin(app,`/user/${uid}/extends`,"POST",{key:"x".repeat(65),value:"x"},400);await admin(app,`/user/${uid}/extends`,"POST",{key:"   ",value:"x"},400);});
 await check("无效奖励范围与空选择被拒绝",async()=>{for(const v of [{target:"unknown"},{target:"selected",userIds:[]}])await admin(app,"/user/reward","POST",{value:1,unit:"day",...v},400);});
 await check("并发创建同名变量仅一个成功",async()=>{const results=await Promise.all(Array.from({length:5},()=>http(`/api/admin/user/${uid}/extends`,"POST",{key:"concurrent",value:"x"},{Authorization:"Bearer "+auth.data.token,"X-App-Id":String(app.id)})));assert.equal(results.filter(x=>x.code===0).length,1);assert.ok(results.every(x=>x.code===0||x.code===400));});
 await check("普通管理员不能越权访问系统和管理员管理，重置密码后旧会话失效",async()=>{const name="edge"+randomBytes(3).toString("hex"), created=await admin(null,"/admin","POST",{username:name,password:"test-pass-123",role:"admin"});try{const login=await http("/api/admin/login","POST",{username:name,password:"test-pass-123"});assert.equal(login.code,0);const headers={Authorization:"Bearer "+login.data.token};for(const path of ["/admin/list","/system/settings","/system/api-routes"])assert.equal((await http("/api/admin"+path,"GET",undefined,headers)).code,403);assert.equal((await http("/api/admin/me","GET",undefined,headers)).code,0);await admin(null,"/admin","POST",{id:created.id,role:"admin",password:"new-pass-123"});assert.equal((await http("/api/admin/me","GET",undefined,headers)).code,401);}finally{await admin(null,`/admin/${created.id}`,"DELETE");}});
 const ca=await createApp("kami"),g=await admin(ca,"/kami/group","POST",{name:"audit",type:"vip",value:60});const cards=(await admin(ca,"/kami/generate","POST",{groupId:g.id,count:2,length:16})).cards;const active=await sdk(ca,"logon",{card:cards[0],mi:"device-A"});const pool=(await admin(ca,"/kami/list")).list;const unused=pool.find(x=>x.card===cards[1]);
 await check("清理已用卡包含冻结的已用卡，但保留冻结的未用卡",async()=>{for(const id of [active.user.id,unused.id])await admin(ca,`/kami/${id}`,"PUT",{frozen:true});const out=await admin(ca,"/kami/clean-used","POST",{});assert.equal(out.affected,1);const remaining=(await admin(ca,"/kami/list")).list;assert.deepEqual(remaining.map(x=>x.id),[unused.id]);await sdk(ca,"info",{},active.token,"device-A",1006);});
} finally {for(const app of apps.reverse())await admin(null,`/app/${app.id}`,"DELETE");console.log(JSON.stringify({passed:checks.length,failures},null,2));}
if(failures.length)process.exitCode=1;
