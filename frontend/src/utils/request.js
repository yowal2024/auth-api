import axios from "axios";
import { Message } from "@arco-design/web-vue";
import router from "@/router/index.js";
import { getAdminEntry } from "./install-status.js";
import { getToken, clearToken, getCurrentAppId } from "./storage.js";
import { useUserStore } from "@/stores/user.js";
import { reportError, silenceError } from "./errors.js";

export function rewriteAdminPath(url) {
  const entry = getAdminEntry();
  if (entry === "admin") return url;
  if (url === "/admin") return `/${entry}`;
  return url.startsWith("/admin/") ? `/${entry}${url.slice(6)}` : url;
}

export function rawRequest(url, options = {}) {
  const headers = { ...options.headers };
  const token = getToken();
  const appId = getCurrentAppId();
  if (token && !headers.Authorization)
    headers.Authorization = `Bearer ${token}`;
  if (appId && !headers["X-App-Id"]) headers["X-App-Id"] = appId;
  return fetch(`/api${rewriteAdminPath(url)}`, { ...options, headers });
}

const client = axios.create({ baseURL: "/api", timeout: 15000 });
client.interceptors.request.use((config) => {
  const token = getToken();
  if (token) config.headers.Authorization = `Bearer ${token}`;
  const appId = getCurrentAppId();
  if (appId) config.headers["X-App-Id"] = appId;
  if (config.url) config.url = rewriteAdminPath(config.url);
  return config;
});
let redirecting = false;
function expireLogin(config) {
  // A response from a previous session must not log out the current session.
  if (!getToken() || config?.headers?.Authorization !== `Bearer ${getToken()}`)
    return;
  clearToken();
  useUserStore().resetAuth();
  if (
    redirecting ||
    router.currentRoute.value.path === `/${getAdminEntry()}/login`
  )
    return;
  redirecting = true;
  Message.error("登录已过期，请重新登录");
  router.replace(`/${getAdminEntry()}/login`).finally(() => {
    redirecting = false;
  });
}

async function responseEnvelope(response) {
  if (
    response.config.responseType === "blob" ||
    response.config.responseType === "arraybuffer"
  ) {
    // An attachment may legitimately contain JSON; only decode API error envelopes.
    if (
      response.headers["content-disposition"] ||
      !response.headers["content-type"]?.includes("application/json")
    )
      return null;
    const text =
      response.data instanceof Blob
        ? await response.data.text()
        : new TextDecoder().decode(response.data);
    try {
      return JSON.parse(text);
    } catch {
      return null;
    }
  }
  return response.data;
}
client.interceptors.response.use(
  async (response) => {
    const envelope = await responseEnvelope(response);
    if (envelope === null) return response;
    const { code, msg, data } = envelope || {};
    if (code === 0 || code === 200) return data;
    if (code === 401) {
      expireLogin(response.config);
      return Promise.reject(silenceError(new Error("unauthorized")));
    }
    const error = new Error(msg || "请求失败");
    error.code = code;
    error.data = data;
    reportError(error);
    return Promise.reject(error);
  },
  async (error) => {
    if (axios.isCancel(error)) return Promise.reject(silenceError(error));
    const envelope = error.response
      ? await responseEnvelope(error.response)
      : null;
    const unauthorized =
      error.response?.status === 401 || envelope?.code === 401;
    if (error.code === "ECONNABORTED") error.message = "请求超时，请重试";
    if (envelope) {
      error.data = envelope.data;
      error.code = envelope.code;
      error.message = envelope.msg || error.message;
    }
    if (unauthorized) {
      expireLogin(error.config);
      silenceError(error);
    } else reportError(error, "网络错误");
    return Promise.reject(error);
  },
);
export const request = (url, options = {}) => client({ url, ...options });
