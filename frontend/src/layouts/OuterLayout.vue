<!-- Reconstructed editable Vue source; see README for recovery limits. -->
<script>
import { useUserStore } from "@/stores/user.js";
import HeaderActions from "@/components/HeaderActions.vue";
import { onMounted } from "vue";
import { IconSafe } from "@arco-design/web-vue/es/icon";
export default {
  name: "OuterLayout",
  components: {
    IconSafe,
    HeaderActions,
  },
  setup(y) {
    const userStore = useUserStore();
    onMounted(async () => {
      userStore.profile || (await userStore.fetchMe().catch(() => {}));
    });
    return {
      userStore,
    };
  },
};
</script>
<template>
  <div class="outer-layout">
    <header class="outer-header">
      <div class="brand">
        <IconSafe
          :size="22"
          :style="{
            color: 'rgb(var(--primary-6))',
          }"
        ></IconSafe>
        <span class="title">Auth 管理后台</span>
      </div>
      <div class="grow"></div>
      <HeaderActions></HeaderActions>
    </header>
    <main class="outer-main"><router-view></router-view></main>
  </div>
</template>
<style scoped>
.outer-layout {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: var(--content-bg);
}
.outer-header {
  height: 56px;
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 0 24px;
  background: var(--color-bg-2);
  border-bottom: 1px solid var(--color-border);
}
.outer-header .brand {
  display: flex;
  align-items: center;
  gap: 10px;
}
.outer-header .brand .title {
  white-space: nowrap;
  font-size: 16px;
  font-weight: 600;
  color: var(--color-text-1);
}
.outer-header .grow {
  flex: 1;
}
.outer-main {
  flex: 1;
  overflow-y: auto;
}
@media (max-width: 600px) {
  .outer-header {
    padding: 0 12px;
    gap: 8px;
  }
  .outer-header .brand {
    flex-shrink: 0;
  }
  .outer-header .brand .title {
    display: none;
  }
}
</style>
