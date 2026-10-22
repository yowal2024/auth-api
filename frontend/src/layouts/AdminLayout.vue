<!-- Reconstructed editable Vue source; see README for recovery limits. -->
<script>
import { onMounted, onUnmounted, ref, Transition, computed } from "vue";
import {
  Layout,
  LayoutHeader,
  Tag,
  LayoutSider,
  Menu,
  MenuItem,
  SubMenu,
  LayoutContent,
} from "@arco-design/web-vue";
import { useRoute, useRouter } from "vue-router";
import {
  IconMenuFold,
  IconMenuUnfold,
  IconDashboard,
  IconClose,
  IconSafe,
  IconApps,
  IconBook,
  IconList,
  IconTool,
  IconFile,
  IconUserGroup,
  IconCode,
  IconCommon,
  IconStop,
  IconStorage,
  IconSettings,
  IconNotification,
  IconUser,
} from "@arco-design/web-vue/es/icon";
import { useUserStore } from "@/stores/user.js";
import { useAppStore } from "@/stores/app.js";

import HeaderActions from "@/components/HeaderActions.vue";
import { defaultAvatar } from "@/utils/avatar.js";
import { useTheme } from "@/composables/useTheme.js";
import { entryPath } from "@/utils/route.js";
export default {
  name: "AdminLayout",
  components: {
    Layout,
    LayoutHeader,
    IconMenuFold,
    IconMenuUnfold,
    Tag,
    HeaderActions,
    LayoutSider,
    Menu,
    MenuItem,
    IconDashboard,
    SubMenu,
    IconClose,
    LayoutContent,
    Transition,
  },
  setup(Ze) {
    const buildInfo = __BUILD_INFO__;
    const buildTime = new Intl.DateTimeFormat("zh-CN", {
      timeZone: "Asia/Shanghai",
      year: "numeric",
      month: "2-digit",
      day: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
      hourCycle: "h23",
    }).format(new Date(buildInfo.builtAt)).replaceAll("/", "-");
    const Z = new Date().getFullYear();
    const route = useRoute();
    const router = useRouter();
    const userStore = useUserStore();
    const appStore = useAppStore();
    const { dark: F } = useTheme();
    const f = ref(false);
    const v = ref(false);
    const k = ref(false);
    const E = () => {
      f.value = window.innerWidth < 992;
      f.value || (v.value = false);
    };
    const derivedState = computed(() => appStore.current);
    const q = {
      user: "用户版",
      kami: "卡密版",
    };
    const O = (I) => q[I] || "用户版";
    const pageState = {
      IconUser: IconUser,
      IconNotification: IconNotification,
      IconSettings: IconSettings,
      IconStorage: IconStorage,
      IconStop: IconStop,
      IconCommon: IconCommon,
      IconCode: IconCode,
      IconUserGroup: IconUserGroup,
      IconFile: IconFile,
      IconTool: IconTool,
      IconList: IconList,
      IconBook: IconBook,
      IconApps: IconApps,
      IconDashboard: IconDashboard,
      IconSafe: IconSafe,
    };
    const W = computed(() => {
      const l = (
        router
          .getRoutes()
          .find((o) => o.path === "/:entry" && o.children?.length)?.children ||
        []
      ).filter((o) => o.path && !o.meta?.hidden);
      const C = {
        user: {
          key: "user",
          label: "USER",
          items: [],
        },
        app: {
          key: "app",
          label: "APP",
          items: [],
        },
        sys: {
          key: "sys",
          label: "系统",
          items: [],
        },
      };
      const N = derivedState.value?.type;
      const h = {};
      return (
        l.forEach((o) => {
          const w = o.meta?.group;
          if (
            !w ||
            o.meta?.hidden ||
            (o.meta?.superOnly && userStore.profile?.role !== "super") ||
            o.path === "dashboard" ||
            o.meta?.hideForTypes?.includes(N)
          )
            return;
          const M = {
            path: entryPath("/" + o.path),
            title: o.meta.title,
            icon: pageState[o.meta.icon] || IconList,
          };
          const y = o.meta?.submenu;
          if (y) {
            let b = h[y.key];
            b ||
              ((b = {
                key: y.key,
                title: y.title,
                icon: pageState[y.icon] || IconList,
                children: [],
              }),
              (h[y.key] = b),
              C[w]?.items.push(b));
            b.children.push(M);
          } else C[w]?.items.push(M);
        }),
        Object.values(C).filter((o) => o.items.length)
      );
    });
    const handleMenuItemClick = (I) => {
      I.startsWith("/") && router.push(I);
      J();
    };
    const handleClick = () => router.push(entryPath("/app"));
    const J = () => {
      f.value && (v.value = false);
    };
    onMounted(async () => {
      E();
      window.addEventListener("resize", E);
      userStore.profile || (await userStore.fetchMe().catch(() => {}));
      appStore.apps.length || (await appStore.fetchApps().catch(() => {}));
    });
    onUnmounted(() => window.removeEventListener("resize", E));
    return {
      buildInfo,
      buildTime,
      Z,
      route,
      router,
      userStore,
      appStore,
      F,
      f,
      v,
      k,
      E,
      derivedState,
      q,
      O,
      pageState,
      W,
      handleMenuItemClick,
      handleClick,
      J,
      defaultAvatar,
      entryPath,
    };
  },
};
</script>
<template>
  <Layout class="layout">
    <template #default>
      <LayoutHeader class="header">
        <template #default>
          <div class="header-left">
            <template v-if="f">
              <div :key="0" class="collapse-btn" @click="(d) => (v = !v)">
                <IconMenuUnfold></IconMenuUnfold>
              </div>
            </template>
            <template v-else></template>
            <div class="logo" @click="handleClick">
              <div class="logo-icon">
                <img :src="derivedState?.logo || defaultAvatar" alt="logo" />
              </div>
              <template v-if="f"></template>
              <template v-else>
                <div :key="0" class="logo-text">
                  <span class="name">{{ derivedState?.name || "Auth" }}</span>
                  <template v-if="derivedState">
                    <Tag :key="0" size="small" class="ver-tag">
                      <template #default>{{ O(derivedState.type) }}</template>
                    </Tag>
                  </template>
                  <template v-else></template>
                </div>
              </template>
            </div>
          </div>
          <div class="header-right"><HeaderActions></HeaderActions></div>
        </template>
      </LayoutHeader>
      <Layout class="body">
        <template #default>
          <template v-if="f &amp;&amp; v">
            <div :key="0" class="sider-mask" @click="(d) => (v = !1)"></div>
          </template>
          <template v-else></template>
          <template v-if="f"></template>
          <template v-else>
            <LayoutSider
              :key="1"
              :collapsed="k"
              :collapsed-width="60"
              :width="220"
              :resize-directions="[]"
              :class="[
                'sider',
                {
                  'is-collapsed': k,
                },
              ]"
            >
              <template #default>
                <Menu
                  class="sider-menu"
                  :collapsed="k"
                  :selected-keys="[route.path]"
                  :style="{
                    width: '100%',
                    background: 'transparent',
                  }"
                  :theme="F ? 'dark' : 'light'"
                  @menuItemClick="handleMenuItemClick"
                >
                  <template #default>
                    <MenuItem :key="entryPath('/dashboard')">
                      <template #icon><IconDashboard></IconDashboard></template>
                      <template #default>总览</template>
                    </MenuItem>
                    <template v-for="d in W">
                      <div
                        :class="[
                          'menu-group-title',
                          {
                            collapsed: k,
                          },
                        ]"
                      >
                        {{ d.label }}
                      </div>
                      <template v-for="s in d.items">
                        <template v-if="s.children">
                          <SubMenu :key="s.key">
                            <template #icon
                              ><component :is="s.icon"></component
                            ></template>
                            <template #title>{{ s.title }}</template>
                            <template #default>
                              <template v-for="S in s.children" :key="S.path">
                                <MenuItem>
                                  <template #default>{{ S.title }}</template>
                                </MenuItem>
                              </template>
                            </template>
                          </SubMenu>
                        </template>
                        <template v-else>
                          <MenuItem :key="s.path">
                            <template #icon
                              ><component :is="s.icon"></component
                            ></template>
                            <template #default>{{ " " + s.title }}</template>
                          </MenuItem>
                        </template>
                      </template>
                    </template>
                  </template>
                </Menu>
                <div class="sider-foot">
                  <component
                    :is="k ? 'icon-menu-unfold' : 'icon-menu-fold'"
                    class="fold-btn"
                    @click="(d) => (k = !k)"
                  ></component>
                </div>
              </template>
            </LayoutSider>
          </template>
          <template v-if="f">
            <aside
              :key="2"
              :class="[
                'mobile-drawer',
                {
                  open: v,
                },
              ]"
            >
              <div class="mobile-head">
                <div class="logo" @click="handleClick">
                  <div class="logo-icon">
                    <img
                      :src="derivedState?.logo || defaultAvatar"
                      alt="logo"
                    />
                  </div>
                  <div class="logo-text">
                    <span class="name">{{ derivedState?.name || "Auth" }}</span>
                    <template v-if="derivedState">
                      <Tag :key="0" size="small" class="ver-tag">
                        <template #default>{{ O(derivedState.type) }}</template>
                      </Tag>
                    </template>
                    <template v-else></template>
                  </div>
                </div>
                <div class="close-btn" @click="(d) => (v = !1)">
                  <IconClose></IconClose>
                </div>
              </div>
              <Menu
                :selected-keys="[route.path]"
                :style="{
                  width: '100%',
                  background: 'transparent',
                  border: 'none',
                }"
                :theme="F ? 'dark' : 'light'"
                @menuItemClick="handleMenuItemClick"
              >
                <template #default>
                  <MenuItem :key="entryPath('/dashboard')">
                    <template #icon><IconDashboard></IconDashboard></template>
                    <template #default>总览</template>
                  </MenuItem>
                  <template v-for="d in W">
                    <template v-for="s in d.items">
                      <template v-if="s.children">
                        <SubMenu :key="s.key">
                          <template #icon
                            ><component :is="s.icon"></component
                          ></template>
                          <template #title>{{ s.title }}</template>
                          <template #default>
                            <template v-for="S in s.children" :key="S.path">
                              <MenuItem>
                                <template #default>{{ S.title }}</template>
                              </MenuItem>
                            </template>
                          </template>
                        </SubMenu>
                      </template>
                      <template v-else>
                        <MenuItem :key="s.path">
                          <template #icon
                            ><component :is="s.icon"></component
                          ></template>
                          <template #default>{{ " " + s.title }}</template>
                        </MenuItem>
                      </template>
                    </template>
                  </template>
                </template>
              </Menu>
            </aside>
          </template>
          <template v-else></template>
          <Layout class="content-col">
            <template #default>
              <LayoutContent class="main">
                <template #default>
                  <router-view>
                    <template #default="{ Component: d }">
                      <Transition name="fade" mode="out-in">
                        <template #default
                          ><component
                            :is="d"
                            :key="`${appStore.currentId}:${route.path}`"
                          ></component
                        ></template>
                      </Transition>
                    </template>
                  </router-view>
                </template>
              </LayoutContent>
              <footer class="app-footer">
                <div class="footer-left">
                  {{ " Copyright © " + Z + " Auth 管理系统 " }}
                </div>
                <div class="footer-right">
                  <span class="meta">构建信息</span>
                  <span class="ver">v{{ buildInfo.version }}</span>
                  <time class="ver" :datetime="buildInfo.builtAt">{{ buildTime }} (UTC+8)</time>
                </div>
              </footer>
            </template>
          </Layout>
        </template>
      </Layout>
    </template>
  </Layout>
</template>
<style scoped>
.layout {
  height: 100vh;
  display: flex;
  flex-direction: column;
}
.header {
  background: var(--color-bg-2);
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 24px 0 0;
  height: var(--layout-header-h);
  flex-shrink: 0;
}
.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
  height: 100%;
  min-width: 0;
}
.header-left .logo {
  height: 100%;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 18px;
  min-width: 0;
  color: var(--color-text-1);
  cursor: pointer;
}
.header-left .logo .logo-icon {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.header-left .logo .logo-icon img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}
.header-left .logo .logo-text {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}
.header-left .logo .logo-text .name {
  font-size: 15px;
  font-weight: 600;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.header-left .logo .logo-text .ver-tag {
  background: #165dff1f;
  color: rgb(var(--primary-6));
  border: none;
}
.collapse-btn {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 22px;
  cursor: pointer;
  color: var(--color-text-2);
  margin-left: 8px;
  flex-shrink: 0;
  border-radius: 4px;
}
.collapse-btn:hover {
  color: rgb(var(--primary-6));
  background: var(--color-fill-2);
}
.header-right {
  display: flex;
  align-items: center;
  flex-shrink: 0;
}
.body {
  flex: 1;
  min-height: 0;
}
.sider {
  background: var(--color-bg-2, #fff);
}
.sider :deep(.arco-layout-sider-children) {
  display: flex;
  flex-direction: column;
  height: 100%;
}
.sider .sider-menu {
  flex: 1;
  overflow-y: auto;
}
.sider .sider-foot {
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  padding: 0 14px;
  flex-shrink: 0;
}
.sider .sider-foot .fold-btn {
  font-size: 18px;
  color: var(--color-text-3);
  cursor: pointer;
}
.sider .sider-foot .fold-btn:hover {
  color: rgb(var(--primary-6));
}
.sider :deep(.arco-menu-light),
.sider :deep(.arco-menu-dark) {
  background: transparent;
}
.sider.is-collapsed :deep(.arco-menu-collapsed) {
  width: 60px !important;
}
.sider.is-collapsed :deep(.arco-menu-collapsed .arco-menu-inner) {
  padding: 4px 0 !important;
}
.sider.is-collapsed :deep(.arco-menu-collapsed .arco-menu-item),
.sider.is-collapsed :deep(.arco-menu-collapsed .arco-menu-pop-header),
.sider.is-collapsed :deep(.arco-menu-collapsed .arco-menu-inline-header) {
  position: relative !important;
  padding: 0 !important;
  margin: 4px 0 !important;
  overflow: hidden !important;
  text-indent: -9999px;
}
.sider.is-collapsed :deep(.arco-menu-collapsed .arco-menu-item .arco-icon),
.sider.is-collapsed
  :deep(.arco-menu-collapsed .arco-menu-pop-header .arco-icon),
.sider.is-collapsed
  :deep(.arco-menu-collapsed .arco-menu-inline-header .arco-icon) {
  position: absolute !important;
  top: 50% !important;
  left: 50% !important;
  transform: translate(-50%, -50%) !important;
  margin: 0 !important;
  font-size: 18px !important;
  text-indent: 0;
}
.sider .menu-group-title {
  padding: 14px 20px 6px;
  font-size: 11px;
  font-weight: 500;
  color: var(--color-text-3);
  letter-spacing: 0.5px;
}
.sider .menu-group-title.collapsed {
  padding: 12px 0 4px;
  text-align: center;
  letter-spacing: 0;
}
.content-col {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}
.main {
  flex: 1;
  min-height: 0;
  background: var(--content-bg, #f5f7fa);
  overflow-y: auto;
  padding: 0;
}
.app-footer {
  flex: none;
  display: flex;
  flex-wrap: wrap;
  gap: 6px 16px;
  justify-content: space-between;
  align-items: center;
  padding: 10px 20px;
  font-size: 12px;
  color: var(--color-text-3, #86909c);
  background: var(--color-bg-2, #ffffff);
  border-top: 1px solid var(--color-border-2, #e5e6eb);
  user-select: none;
}
.app-footer .footer-left {
  display: flex;
  align-items: center;
  gap: 6px;
}
.app-footer .footer-right {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 4px 10px;
}
.app-footer .footer-right .meta {
  opacity: 0.7;
}
.app-footer .footer-right .ver {
  color: var(--color-text-2, #4e5969);
  font-weight: 500;
}
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.15s;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
.sider-mask {
  position: fixed;
  inset: 0;
  background: #0006;
  z-index: 99;
}
.mobile-drawer {
  position: fixed;
  left: 0;
  top: 0;
  bottom: 0;
  width: 78vw;
  max-width: 320px;
  background: var(--color-bg-2);
  z-index: 100;
  transition: transform 0.25s;
  transform: translate(-100%);
  display: flex;
  flex-direction: column;
  overflow-y: auto;
}
.mobile-drawer.open {
  transform: translate(0);
}
.mobile-drawer .mobile-head {
  height: var(--layout-header-h);
  padding: 0 14px 0 18px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-shrink: 0;
}
.mobile-drawer .mobile-head .logo {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  min-width: 0;
}
.mobile-drawer .mobile-head .logo .logo-icon {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.mobile-drawer .mobile-head .logo .logo-icon img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}
.mobile-drawer .mobile-head .logo .logo-text {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}
.mobile-drawer .mobile-head .logo .logo-text .name {
  font-size: 16px;
  font-weight: 600;
  color: var(--color-text-1);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.mobile-drawer .mobile-head .logo .logo-text .ver-tag {
  flex-shrink: 0;
  background: #165dff1f;
  color: rgb(var(--primary-6));
  border: none;
}
.mobile-drawer .mobile-head .close-btn {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 22px;
  color: var(--color-text-3);
  cursor: pointer;
  flex-shrink: 0;
  border-radius: 4px;
}
.mobile-drawer .mobile-head .close-btn:hover {
  color: rgb(var(--primary-6));
  background: var(--color-fill-2);
}
.mobile-drawer :deep(.arco-menu) {
  padding: 6px 0;
}
@media (max-width: 991px) {
  .header {
    padding: 0 12px 0 0;
  }
  .header-left {
    gap: 8px;
  }
  .header-left .logo {
    padding: 0 8px;
  }
}
</style>
