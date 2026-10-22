export const cloudExamples = {
  basic: `function handler(ctx) {
  const input = ctx.payload || {};
  console.log("trigger:", ctx.trigger);
  return { message: "Hello, " + (input.name || "visitor"), appId: ctx.app.id };
}`,
  reject: `// 配置触发器：before_register
function handler(ctx) {
  const input = ctx.payload || {};
  // 当前注册触发器字段为 Username；兼容后台 JSON 试运行的 username。
  const name = String(input.Username || input.username || "");
  if (name.toLowerCase().startsWith("admin")) {
    return { allow: false, msg: "该账号名称不可注册" };
  }
  return { allow: true };
}`,
  database: `function handler(ctx) {
  const input = ctx.payload || {};
  // 使用 ? 占位符；参数逐个传入，不要传数组或拼接用户输入。
  const rows = db.query("SELECT ? AS value", String(input.value || "hello"));
  return { value: rows[0].value };
}
// 业务查询示例（把 a_ 换成实际表前缀）：
// db.query("SELECT id, nickname FROM a_users WHERE appid = ? AND id = ? LIMIT 1",
//          ctx.app.id, ctx.user.id);`,
  http: `function handler(ctx) {
  // 替换为你实际使用的公网 HTTPS 地址。
  const response = http.get("https://example.com/");
  if (!response.ok) {
    throw new Error("业务错误:远端服务暂不可用");
  }
  return { status: response.status, contentLength: response.body.length };
}
// POST 会将第二个参数编码为 JSON：
// const response = http.post("https://your-domain.example/api", { appId: ctx.app.id });
// 仅当远端返回 JSON 时使用 JSON.parse(response.body)。`,
  crypto: `function handler(ctx) {
  const text = String((ctx.payload || {}).text || "abc");
  return {
    sha256: crypto.sha256(text),
    encoded: crypto.base64Encode(text),
    decoded: crypto.base64Decode(crypto.base64Encode(text))
  };
}`,
  error: `function handler(ctx) {
  if (!ctx.user) {
    throw new Error("业务错误:请在客户端登录后调用");
  }
  return { username: ctx.user.username };
}`,
  request: `POST /api/sdk/cloudFunction
Content-Type: application/json
X-App-Id: <应用 ID>
X-Token: <用户登录令牌>
X-Ts: <Unix 秒时间戳>
X-Nonce: <随机串>
X-Sign: <按应用 tokenKey 计算的签名>

{ "name": "hello", "param": { "name": "Alice" } }`,
  result: `// 后台试运行的响应 data（外层 code=0 仅代表试运行请求已处理）
{
  "ok": true,
  "result": { "message": "Hello, Alice", "appId": 1000 },
  "logs": ["[log] trigger: manual"],
  "duration": 1
}
// 脚本执行失败：data.ok=false，data.error 给出错误；空字段可能省略。

// SDK 调用成功：返回值直接放在外层 data
{ "code": 0, "msg": "ok", "data": { "message": "Hello, Alice", "appId": 1000 } }
// SDK 抛出“业务错误:”前缀的异常：默认错误码 1099
{ "code": 1099, "msg": "请在客户端登录后调用" }`,
};
