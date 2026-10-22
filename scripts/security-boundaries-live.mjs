import assert from "node:assert/strict";
import { createHmac, randomBytes } from "node:crypto";

const base = process.env.TEST_BASE_URL || "http://127.0.0.1:18080";
const entry = process.env.TEST_ADMIN_ENTRY || "admin";
assert.ok(process.env.TEST_ADMIN_PASSWORD, "Set TEST_ADMIN_PASSWORD");
async function call(path, method = "GET", data, headers = {}) {
  return (await fetch(base + path, {
    method, headers: { "Content-Type": "application/json", ...headers },
    body: data === undefined ? undefined : JSON.stringify(data),
    signal: AbortSignal.timeout(15000),
  })).json();
}
function ok(out) { assert.equal(out.code, 0, out.msg); return out.data; }
const token = ok(await call(`/api/${entry}/login`, "POST", {
  username: process.env.TEST_ADMIN_USER || "admin", password: process.env.TEST_ADMIN_PASSWORD,
})).token;
const headers = { Authorization: "Bearer " + token };
const admin = (path, method = "GET", data, app) => call(`/api/${entry}` + path, method, data, {
  ...headers, "X-App-Id": String(app?.id || 0),
});
async function sdk(app, path, data) {
  const ts = String(Math.floor(Date.now() / 1000)), nonce = randomBytes(16).toString("hex");
  const body = JSON.stringify(data);
  return call("/api/sdk/" + path, "POST", data, {
    "X-App-Id": String(app.id), "X-Ts": ts, "X-Nonce": nonce,
    "X-Sign": createHmac("sha256", app.tokenKey).update(ts + "\n" + nonce + "\n" + body).digest("hex"),
  });
}
const apps = [];
let operator;
try {
  assert.equal(ok(await call("/api/install/status")).adminEntry, undefined);
  assert.equal(ok(await call("/api/install/status", "GET", undefined, { "X-Admin-Entry": "wrong-entry" })).adminEntry, undefined);
  assert.equal(ok(await call("/api/install/status", "GET", undefined, { "X-Admin-Entry": entry })).adminEntry, entry);
  console.log("PASS public status does not disclose the admin entry");

  const app = ok(await admin("/app", "POST", { name: "Security regression " + Date.now(), type: "user", mode: "free", bruteForceEnabled: true }));
  apps.push(app);
  assert.equal(app.bruteForceEnabled, true);
  ok(await sdk(app, "register", { username: "guarduser", password: "test-pass-123" }));
  for (const name of ["guarduser", "GUARDUSER", "GuardUser", "guarduser ", "gUArdUsEr"]) {
    assert.equal((await sdk(app, "logon", { username: name, password: "wrong-pass" })).code, 1001);
  }
  assert.equal((await sdk(app, "logon", { username: "GUARDuser", password: "wrong-pass" })).code, 1017);
  console.log("PASS case and trailing-space aliases share the login limit");

  ok(await sdk(app, "register", { username: "raceguard", password: "test-pass-123" }));
  const attempts = await Promise.all(Array.from({ length: 20 }, () => sdk(app, "logon", { username: "raceguard", password: "wrong-pass" })));
  assert.equal(attempts.filter(x => x.code === 1001).length, 5);
  assert.equal(attempts.filter(x => x.code === 1017).length, 15);
  console.log("PASS parallel guesses cannot exceed five password checks");

  const username = "sec" + randomBytes(4).toString("hex");
  operator = ok(await admin("/admin", "POST", { username, password: "test-pass-123", role: "admin" }));
  const lowToken = ok(await call(`/api/${entry}/login`, "POST", { username, password: "test-pass-123" })).token;
  const lowHeaders = { Authorization: "Bearer " + lowToken, "X-App-Id": String(app.id) };
  for (const [method, path] of [["GET", "/function/list"], ["GET", "/function/1"], ["POST", "/function"], ["POST", "/function/0/run"], ["POST", "/function/1/run"], ["DELETE", "/function/1"]]) {
    assert.equal((await call(`/api/${entry}` + path, method, method === "GET" ? undefined : {}, lowHeaders)).code, 403, path);
  }
  const result = ok(await admin("/function/0/run", "POST", { code: "function handler(){return {ok:true};}" }, app));
  assert.equal(result.ok, true);
  console.log("PASS ordinary admins cannot access scripts; super-admin execution works");

  ok(await admin("/admin", "POST", { id: operator.id, role: "admin", password: "changed-pass-123" }));
  assert.equal((await call(`/api/${entry}/me`, "GET", undefined, lowHeaders)).code, 401);
  ok(await call(`/api/${entry}/login`, "POST", { username, password: "changed-pass-123" }));
  console.log("PASS password reset rejects the old admin token");
} finally {
  if (operator) ok(await admin(`/admin/${operator.id}`, "DELETE"));
  for (const app of apps.reverse()) ok(await admin(`/app/${app.id}`, "DELETE"));
  console.log("Temporary accounts and apps removed");
}
