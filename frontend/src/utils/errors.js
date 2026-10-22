import { Message } from "@arco-design/web-vue";

const reported = new WeakSet();
export function silenceError(error) {
  if (error && typeof error === "object") reported.add(error);
  return error;
}

export function reportError(error, fallback = "请求失败") {
  if (error && typeof error === "object" && reported.has(error)) return;
  silenceError(error);
  Message.error(error?.message || fallback);
}
