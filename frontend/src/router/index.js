import { createRouter, createWebHistory } from "vue-router";
import { routes } from "./routes.js";
import {
  refreshInstallStatus,
  getCachedInstallStatus,
  getAdminEntry,
} from "@/utils/install-status.js";
import { getToken } from "@/utils/storage.js";
import { Message } from "@arco-design/web-vue";

const router = createRouter({ history: createWebHistory(), routes });
router.beforeEach(async (to) => {
  const candidate = typeof to.params.entry === "string" ? to.params.entry : undefined;
  const status = (await refreshInstallStatus(candidate)) || getCachedInstallStatus();
  if (!status) {
    Message.error("无法连接服务器，请检查网络后刷新重试");
    return to.path === "/home" ? true : "/home";
  }
  if (!status?.installed) return to.path === "/install" ? true : "/install";
  if (to.path === "/install") return "/home";
  const entry = getAdminEntry();
  if (to.params.entry !== undefined && to.params.entry !== status.adminEntry)
    return "/home";
  if (
    to.matched.some((record) => record.path === "/:entry/login") ||
    to.meta.public
  )
    return true;
  if (!getToken()) return `/${entry}/login`;
  if (to.meta.superOnly) {
    const { useUserStore } = await import("@/stores/user.js");
    const user = useUserStore();
    try {
      await user.fetchMe();
    } catch {
      return false;
    }
    if (user.profile?.role !== "super") {
      Message.warning("此页面仅限超级管理员");
      return `/${entry}/app`;
    }
  }
  if (to.matched.some((record) => record.path === "/:entry")) {
    const { useAppStore } = await import("@/stores/app.js");
    const apps = useAppStore();
    if (!apps.apps.length) {
      try {
        await apps.fetchApps();
      } catch {
        return false;
      }
    }
    if (!apps.current) return `/${entry}/app`;
    if (to.meta.hideForTypes?.includes(apps.current.type))
      return `/${entry}/dashboard`;
  }
  return true;
});
router.afterEach((to) => {
  document.title = `${to.meta.title || "Auth"} · 管理后台`;
});
export default router;
