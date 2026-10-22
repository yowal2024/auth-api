import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
const base = process.env.TEST_BASE_URL || "http://127.0.0.1:18080";
assert.ok(process.env.TEST_ADMIN_PASSWORD, "Set TEST_ADMIN_PASSWORD");
const ok = out => { assert.equal(out.code, 0, out.msg); return out.data; };
const token = ok(await (await fetch(base+"/api/admin/login", {method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({username:"admin",password:process.env.TEST_ADMIN_PASSWORD})})).json()).token;
async function admin(app,path,method="GET",data) {
  return (await fetch(base+"/api/admin"+path,{method,headers:{Authorization:"Bearer "+token,"X-App-Id":String(app?.id||0),"Content-Type":"application/json"},body:data===undefined?undefined:JSON.stringify(data)})).json();
}
async function upload(app,path,name,bytes,extra=false) {
  const form=new FormData();form.append("file",new Blob([bytes]),name);
  if(extra)form.append("ignored",new Blob([Buffer.alloc(7<<20)]),"ignored.bin");
  return (await fetch(base+"/api/admin"+path,{method:"POST",headers:{Authorization:"Bearer "+token,"X-App-Id":String(app.id)},body:form})).json();
}
const apps=[];let passed=0;
function pass(message){passed++;console.log("PASS",message);}
try {
  for(const name of ["Download security A","Download security B"]) apps.push(ok(await admin(null,"/app","POST",{name,type:"user",mode:"free"})));
  const [a,b]=apps;
  const bytes=Buffer.from('{\n "code": 0, "data": "下载校验", "large": 9007199254740993\n}\n');
  const file=ok(await upload(a,"/extend-file/upload","配置.json",bytes));
  const urls=[file.downloadUrl,`/api/sdk/extend-file/${file.id}/download`,`/api/admin/extend-file/${file.id}/download`];
  for(const path of urls){
    const headers={Authorization:"Bearer "+token,"X-App-Id":String(a.id)};
    const r=await fetch(base+path,{headers});assert.equal(r.status,200);assert.deepEqual(Buffer.from(await r.arrayBuffer()),bytes);
    assert.match(r.headers.get("content-disposition"),/^attachment;/);assert.equal(r.headers.get("x-content-type-options"),"nosniff");assert.equal(r.headers.get("content-type"),"application/octet-stream");
    const range=await fetch(base+path,{headers:{...headers,Range:"bytes=0-9"}});assert.equal(range.status,206);assert.deepEqual(Buffer.from(await range.arrayBuffer()),bytes.subarray(0,10));
  }pass("静态、SDK、后台下载 JSON 字节与响应头一致，支持 Range");
  const head=await fetch(base+file.downloadUrl,{method:"HEAD"});assert.equal(head.status,200);assert.equal(head.headers.get("content-length"),String(bytes.length));pass("HEAD 长度正确且无正文");
  assert.equal((await admin(b,`/extend-file/${file.id}/download`)).code,404);
  for(const method of ["PUT","DELETE"])assert.equal((await admin(b,`/extend-file/${file.id}`,method,method==="PUT"?{name:"foreign"}:undefined)).code,404);
  const list=ok(await(await fetch(base+"/api/sdk/extend-file/list",{headers:{"X-App-Id":String(a.id)}})).json());assert.ok(list.every(x=>!x.filepath));pass("跨应用下载、修改、删除被拒绝，SDK 隐藏磁盘路径");
  ok(await admin(null,`/app/${a.id}`,"PUT",{name:a.name,status:0}));assert.equal((await fetch(base+file.downloadUrl)).status,404);
  ok(await admin(null,`/app/${a.id}`,"PUT",{name:a.name,status:1}));assert.equal((await fetch(base+file.downloadUrl)).status,200);pass("应用停用拒绝旧链接，启用后恢复");
  ok(await admin(a,`/extend-file/${file.id}`,"DELETE"));assert.equal((await fetch(base+file.downloadUrl)).status,404);pass("文件删除后旧链接失效");
  const png=Buffer.from("89504e470d0a1a0a0000000d49484452000000010000000108060000001f15c4890000000b49444154789c636000020000050001a5f645400000000049454e44ae426082","hex");
  const avatar=ok(await upload(a,"/upload/avatar","avatar.png",png));const image=await fetch(base+avatar.url);assert.equal(image.headers.get("content-type"),"image/png");assert.deepEqual(Buffer.from(await image.arrayBuffer()),png);
  assert.equal((await upload(a,"/upload/avatar","active.svg",Buffer.from('<svg><script>alert(1)</script></svg>'))).code,400);
  assert.equal((await upload(a,"/upload/avatar","avatar.png",png,true)).code,400);pass("图片正常显示，SVG 和夹带超大额外文件被拒绝");
  for(const path of ["/uploads/avatars/","/uploads/app_logos/","/uploads/extend_files/","/uploads/avatars/%2e%2e%2fconfig.yaml","/uploads/extend_files/%2e%2e%5cconfig.yaml"]){assert.equal((await fetch(base+path)).status,404,path);}pass("目录枚举与编码路径穿越被拒绝");
  const evil=ok(await upload(a,"/extend-file/upload","active.html",Buffer.from('<script>alert(1)</script>')));const html=await fetch(base+evil.downloadUrl);assert.equal(html.headers.get("content-type"),"application/octet-stream");assert.match(html.headers.get("content-disposition"),/^attachment;/);ok(await admin(a,`/extend-file/${evil.id}`,"DELETE"));pass("HTML 扩展文件强制附件下载");
  assert.equal((await upload(a,"/ver/upload-apk","fake.apk",Buffer.from("not an APK"))).code,400);pass("伪造 APK 被拒绝");
  assert.ok(process.env.TEST_APK_PATH,"Set TEST_APK_PATH to a compiled fixture APK");
  const apk=await readFile(process.env.TEST_APK_PATH);const uploaded=ok(await upload(a,"/ver/upload-apk","fixture.apk",apk));
  assert.equal((await fetch(base+uploaded.url)).status,404);
  const data={versionName:"security",versionKey:"security",versionNo:"1",isCanonical:1,canonicalPath:uploaded.path,apkUrl:uploaded.url,status:1};
  assert.equal((await admin(b,"/ver","POST",data)).code,400);
  const ver=ok(await admin(a,"/ver","POST",data));const published=await fetch(base+uploaded.url);assert.equal(published.status,200);assert.deepEqual(Buffer.from(await published.arrayBuffer()),apk);
  ok(await admin(a,"/ver","POST",{...data,id:ver.id,status:0}));assert.equal((await fetch(base+uploaded.url)).status,404);
  ok(await admin(a,"/ver","POST",{...data,id:ver.id,status:1}));ok(await admin(a,`/ver/${ver.id}`,"DELETE"));assert.equal((await fetch(base+uploaded.url)).status,404);pass("APK 跨应用引用拒绝，发布后可下载，下架和删除后失效");
  const orphan=ok(await upload(b,"/extend-file/upload","orphan.txt",Buffer.from("orphan")));ok(await admin(null,`/app/${b.id}`,"DELETE"));apps.pop();assert.equal((await fetch(base+orphan.downloadUrl)).status,404);pass("删除应用后遗留文件链接失效");
} finally { for(const app of apps.reverse())ok(await admin(null,`/app/${app.id}`,"DELETE"));console.log(`Passed ${passed} groups; temporary apps removed`); }
