<!-- Reconstructed editable Vue source; see README for recovery limits. -->
<script>
import { useRouter, useRoute } from "vue-router";
import { onMounted, computed } from "vue";
import { Tooltip, Avatar, Popover, Message } from "@arco-design/web-vue";
import {
  IconMoonFill,
  IconSunFill,
  IconSettings,
  IconSync,
  IconBook,
  IconApps,
  IconUser,
  IconUserGroup,
  IconExport,
} from "@arco-design/web-vue/es/icon";
import { getAdminEntry } from "@/utils/install-status.js";
import { useUserStore } from "@/stores/user.js";
import { useTheme } from "@/composables/useTheme.js";
import { entryPath } from "@/utils/route.js";
export default {
  name: "HeaderActions",
  components: {
    IconMoonFill,
    IconSunFill,
    Tooltip,
    IconSettings,
    IconSync,
    IconBook,
    IconApps,
    Popover,
    IconUser,
    IconUserGroup,
    IconExport,
    Avatar,
  },
  setup(dt) {
    const router = useRouter();
    const route = useRoute();
    const userStore = useUserStore();
    const { dark: dark } = useTheme();
    const isAppList = computed(() => route.path === entryPath("/app"));
    const profile = computed(() => userStore.profile);
    const avatarInitial = computed(() =>
      (profile.value?.username || "A").slice(0, 1).toUpperCase(),
    );
    const avatarStyle = computed(() =>
      profile.value?.avatar
        ? {}
        : {
            backgroundColor: "#165dff",
          },
    );
    const applyTheme = (n) => {
      n
        ? document.body.setAttribute("arco-theme", "dark")
        : document.body.removeAttribute("arco-theme");
    };
    const toggleTheme = () => {
      const n = !dark.value;
      localStorage.setItem("auth_theme", n ? "dark" : "light");
      applyTheme(n);
    };
    const navigateTo = (n) => router.push(entryPath(n));
    const openDocs = () => {
      router.push(entryPath("/docs"));
    };
    const clearCache = () => {
      const n = localStorage.getItem("auth_token");
      localStorage.removeItem("auth_current_appid");
      n && localStorage.setItem("auth_token", n);
      sessionStorage.removeItem("install_status_v1");
      Message.success("已清理,即将刷新");
      setTimeout(() => location.reload(), 500);
    };
    const handleGetAdminEntry = async () => {
      await userStore.logout();
      router.push(`/${getAdminEntry()}/login`);
    };
    onMounted(() => {
      applyTheme(localStorage.getItem("auth_theme") === "dark");
    });
    return {
      router,
      route,
      userStore,
      dark,
      isAppList,
      profile,
      avatarInitial,
      avatarStyle,
      applyTheme,
      toggleTheme,
      navigateTo,
      openDocs,
      clearCache,
      handleGetAdminEntry,
    };
  },
};
</script>
<template>
  <div class="header-actions">
    <Tooltip :content="dark ? '切换浅色' : '切换暗色'">
      <template #default>
        <button class="action-btn" @click="toggleTheme">
          <component
            :is="dark ? 'icon-moon-fill' : 'icon-sun-fill'"
          ></component>
        </button>
      </template>
    </Tooltip>
    <Tooltip v-if="profile?.role === 'super'" content="系统设置">
      <template #default>
        <button class="action-btn" @click="(m) => navigateTo('/system')">
          <IconSettings></IconSettings>
        </button>
      </template>
    </Tooltip>
    <Tooltip content="清理缓存">
      <template #default>
        <button class="action-btn" @click="clearCache">
          <IconSync></IconSync>
        </button>
      </template>
    </Tooltip>
    <Tooltip content="使用文档">
      <template #default>
        <button class="action-btn" @click="openDocs">
          <IconBook></IconBook>
        </button>
      </template>
    </Tooltip>
    <template v-if="isAppList"></template>
    <template v-else>
      <Tooltip :key="0" content="返回应用列表">
        <template #default>
          <button class="action-btn" @click="(m) => navigateTo('/app')">
            <IconApps></IconApps>
          </button>
        </template>
      </Tooltip>
    </template>
    <Popover
      trigger="click"
      position="br"
      :content-style="{
        padding: 0,
      }"
    >
      <template #content>
        <div class="user-popover">
          <div class="user-info">
            <div class="info-row">
              <span class="label">昵称:</span>
              <span>{{
                profile?.nickname || profile?.username || "admin"
              }}</span>
            </div>
            <div class="info-row">
              <span class="label">账号:</span>
              <span>{{ profile?.username || "admin" }}</span>
            </div>
          </div>
          <div class="menu-item" @click="(m) => navigateTo('/profile')">
            <IconUser></IconUser>
            <span>个人中心</span>
          </div>
          <div v-if="profile?.role === 'super'" class="menu-item" @click="(m) => navigateTo('/admin')">
            <IconUserGroup></IconUserGroup>
            <span>多管理员</span>
          </div>
          <div class="menu-item" @click="handleGetAdminEntry">
            <IconExport></IconExport>
            <span>退出登录</span>
          </div>
        </div>
      </template>
      <template #default>
        <button class="action-btn avatar-btn">
          <Avatar
            :size="28"
            :image-url="profile?.avatar || ''"
            :style="avatarStyle"
          >
            <template #default>
              <template v-if="profile?.avatar"></template>
              <template v-else>{{ avatarInitial }}</template>
            </template>
          </Avatar>
        </button>
      </template>
    </Popover>
  </div>
</template>
<style scoped>
.header-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}
.action-btn {
  width: 34px;
  height: 34px;
  border: none;
  border-radius: 50%;
  background: var(--color-fill-2);
  color: var(--color-text-2);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  font-size: 16px;
  transition: all 0.2s;
  padding: 0;
}
.action-btn:hover {
  background: var(--color-fill-3);
  color: rgb(var(--primary-6));
}
.avatar-btn {
  background: transparent;
}
.avatar-btn:hover {
  background: var(--color-fill-2);
}
.user-popover {
  min-width: 180px;
  padding: 4px 0;
}
.user-info {
  padding: 12px 16px 8px;
}
.user-info .info-row {
  font-size: 14px;
  line-height: 28px;
  color: var(--color-text-1);
  display: flex;
  gap: 6px;
}
.user-info .info-row .label {
  color: var(--color-text-3);
}
.menu-item {
  padding: 8px 16px;
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 14px;
  color: var(--color-text-1);
  cursor: pointer;
  transition: background 0.15s;
}
.menu-item:hover {
  background: var(--color-fill-2);
}
</style>
