import assert from "node:assert/strict";
import { createServer } from "node:http";
import { once } from "node:events";
import {
  createHmac,
  createCipheriv,
  createDecipheriv,
  randomBytes,
} from "node:crypto";
import { createSDKClient } from "../src/views/docs/sdk-client-example.js";
import { sdkDetails } from "../src/views/docs/sdk-reference.js";

// Exercise the exact code shown to users against a protocol-checking HTTP server.
const tokenKey = "docs-test-key";
const aesKey = randomBytes(32);
const nonces = new Set();
const requests = [];
let serverFailure;
const server = createServer(async (req, res) => {
  try {
    let body = "";
    for await (const chunk of req) body += chunk;
    const { "x-ts": ts, "x-nonce": nonce, "x-sign": sign } = req.headers;
    assert.equal(
      sign,
      createHmac("sha256", tokenKey)
        .update(`${ts}\n${nonce}\n${body}`)
        .digest("hex"),
    );
    assert.ok(!nonces.has(nonce), "Each request needs a fresh nonce");
    nonces.add(nonce);
    assert.equal(req.headers["x-app-id"], "1000");
    assert.equal(req.headers["x-mi"], "device-test");
    let payload = JSON.parse(body);
    const encrypted = req.url.startsWith("/aes");
    if (encrypted) {
      const bytes = Buffer.from(payload.data, "base64");
      const decrypt = createDecipheriv(
        "aes-256-cbc",
        aesKey,
        bytes.subarray(0, 16),
      );
      payload = JSON.parse(
        Buffer.concat([decrypt.update(bytes.subarray(16)), decrypt.final()]),
      );
    }
    requests.push({ path: req.url, payload, token: req.headers["x-token"] });
    let data = { payload, message: "中文响应" };
    if (encrypted) {
      const iv = randomBytes(16);
      const encrypt = createCipheriv("aes-256-cbc", aesKey, iv);
      data = Buffer.concat([
        iv,
        encrypt.update(JSON.stringify(data)),
        encrypt.final(),
      ]).toString("base64");
    }
    res.setHeader("Content-Type", "application/json");
    res.end(
      JSON.stringify(
        req.url.endsWith("/expired")
          ? { code: 1006, msg: "令牌过期" }
          : { code: 0, msg: "ok", data },
      ),
    );
  } catch (error) {
    serverFailure = error;
    res.writeHead(500).end();
  }
});
server.listen(0, "127.0.0.1");
await once(server, "listening");
try {
  const baseURL = `http://127.0.0.1:${server.address().port}`;
  for (const encrypted of [false, true]) {
    const client = createSDKClient({
      baseURL,
      appId: 1000,
      tokenKey,
      mi: "device-test",
      aesKey: encrypted ? aesKey : undefined,
    });
    const path = encrypted ? "/aes" : "/plain";
    client.setToken("test-user-token");
    const input = { username: "中文", escaped: '"\\\n' };
    assert.deepEqual((await client.post(path, input)).payload, input);
    assert.equal(requests.at(-1).token, "test-user-token");
    await assert.rejects(client.post(`${path}/expired`), { code: 1006 });
    assert.deepEqual((await client.post(path)).payload, {});
    assert.equal(
      requests.at(-1).token,
      undefined,
      "Expired token should be cleared",
    );
    await assert.rejects(client.post("https://other.example/api"), /同源/);
  }
  assert.equal(
    requests.length,
    6,
    "Mutating requests must not be automatically retried",
  );
  for (const example of sdkDetails.userExtends.reqExamples)
    JSON.parse(example.code);
  if (serverFailure) throw serverFailure;
  console.log(
    "PASS: documented plain/AES requests, exact HMAC bytes, unique nonces, Unicode, empty JSON, expired token and no retries",
  );
} finally {
  server.close();
  server.closeAllConnections();
}
