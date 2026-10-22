import axios from "axios";
const client = axios.create({ baseURL: "/api/install", timeout: 15000 });
const request = (url, options = {}) =>
  client({ url, ...options })
    .then((response) => response.data)
    .catch((error) => ({ code: -1, msg: error?.message || "网络错误", data: null }));
export const getInstallStatus = (entry = "") => request("/status", {
  method: "GET",
  headers: entry ? { "X-Admin-Entry": entry } : {},
});
export const testMySQL = (data) => request("/test-mysql", { method: "POST", data });
export const testRedis = (data) => request("/test-redis", { method: "POST", data });
export const saveInstallConfig = (data) => request("/save-config", { method: "POST", data });
export const migrateDatabase = () => request("/migrate", { method: "POST" });
export const createAdmin = (data) => request("/create-admin", { method: "POST", data });
export const finalizeInstall = () => request("/finalize", { method: "POST" });
