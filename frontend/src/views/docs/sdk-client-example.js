// Node.js 18+ 示例：明文或 AES-CBC 模式。保存为 sdk-client.mjs。
// tokenKey 用原始字符串；aesKey 传实际解码后的 16/24/32 字节 Buffer。
import {
  createHmac,
  randomBytes,
  createCipheriv,
  createDecipheriv,
} from "node:crypto";

export function createSDKClient({
  baseURL,
  appId,
  tokenKey = "",
  mi = "",
  aesKey,
  encoding = "base64",
}) {
  let token = "";
  const base = new URL(baseURL);
  function encode(payload) {
    if (!aesKey) return payload;
    const iv = randomBytes(16);
    const cipher = createCipheriv(`aes-${aesKey.length * 8}-cbc`, aesKey, iv);
    const ciphertext = Buffer.concat([
      iv,
      cipher.update(JSON.stringify(payload), "utf8"),
      cipher.final(),
    ]);
    return { data: ciphertext.toString(encoding) };
  }
  function decode(data) {
    if (!aesKey || typeof data !== "string") return data;
    const bytes = Buffer.from(data, encoding);
    const decipher = createDecipheriv(
      `aes-${aesKey.length * 8}-cbc`,
      aesKey,
      bytes.subarray(0, 16),
    );
    return JSON.parse(
      Buffer.concat([
        decipher.update(bytes.subarray(16)),
        decipher.final(),
      ]).toString("utf8"),
    );
  }
  async function post(path, payload = {}) {
    const url = new URL(path, base);
    if (url.origin !== base.origin) throw new Error("接口地址必须与服务端同源");
    const body = JSON.stringify(encode(payload)); // 只序列化一次：签名与发送复用它。
    const headers = {
      "Content-Type": "application/json",
      "X-App-Id": String(appId),
    };
    if (token) headers["X-Token"] = token;
    if (mi) headers["X-Mi"] = mi;
    if (tokenKey) {
      const ts = String(Math.floor(Date.now() / 1000));
      const nonce = randomBytes(16).toString("hex");
      Object.assign(headers, {
        "X-Ts": ts,
        "X-Nonce": nonce,
        "X-Sign": createHmac("sha256", tokenKey)
          .update(ts + "\n" + nonce + "\n" + body, "utf8")
          .digest("hex"),
      });
    }
    const response = await fetch(url, {
      method: "POST",
      headers,
      body,
      redirect: "error",
      signal: AbortSignal.timeout(15000),
    });
    if (!response.ok) throw new Error(`HTTP ${response.status}`);
    const envelope = await response.json();
    if (envelope.code !== 0) {
      const error = new Error(envelope.msg || "请求失败");
      error.code = envelope.code;
      if (error.code === 1006) token = "";
      throw error; // 不自动重试充值、注册、云函数等可能有副作用的调用。
    }
    return decode(envelope.data);
  }
  return {
    post,
    setToken(value) {
      token = value || "";
    },
  };
}

// 使用方式（替换成当前应用配置；后台自定义接口后也替换下面路径）：
// const client = createSDKClient({ baseURL: "http://127.0.0.1:18080", appId: 1000,
//   tokenKey: "填写应用 tokenKey", mi: "稳定的设备标识" });
// const ini = await client.post("/api/sdk/ini", { versionKey: "android", versionNo: "1.0.0" });
// const login = await client.post("/api/sdk/logon", { username: "alice", password: "你的密码", mi: "稳定的设备标识" });
// client.setToken(login.token);
// const info = await client.post("/api/sdk/info");
// await client.post("/api/sdk/logout");
// client.setToken("");
