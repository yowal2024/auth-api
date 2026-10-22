<script>
import { ref, onMounted, onBeforeUnmount, computed } from "vue";
import { Button, Progress, Tooltip, Message } from "@arco-design/web-vue";
import { request } from "@/utils/request.js";
import { reportError } from "@/utils/errors.js";
import { useUserStore } from "@/stores/user.js";

export default {
  components: { Button, Progress, Tooltip },
  setup() {
    const user = useUserStore();
    const status = ref(null);
    const loading = ref(true);
    const updating = ref(false);
    const failed = ref(false);
    let active = true,
      timer,
      sequence = 0;
    const busy = computed(() => updating.value || status.value?.updating);
    const canUpdate = computed(() => user.profile?.role === "super");
    const version = computed(() =>
      status.value?.buildTime
        ? new Date(status.value.buildTime * 1000).toISOString().slice(0, 7)
        : "",
    );
    const updatedAt = computed(() => {
      if (!status.value?.updatedAt) return "";
      const d = new Date(status.value.updatedAt * 1000);
      const pad = (v) => String(v).padStart(2, "0");
      return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
    });
    const percent = computed(() =>
      status.value?.totalBytes > 0
        ? Math.min(1, status.value.downloadedBytes / status.value.totalBytes)
        : 0,
    );
    const progressText = computed(() => {
      if (status.value?.stage === "verifying") return "校验中";
      if (status.value?.stage !== "downloading") return "连接中";
      if (status.value.totalBytes > 0)
        return `下载中 ${Math.floor(percent.value * 100)}%`;
      return `已下载 ${((status.value.downloadedBytes || 0) / 1048576).toFixed(1)} MB`;
    });
    function poll() {
      clearTimeout(timer);
      if (active && busy.value) timer = setTimeout(() => refresh(true), 700);
    }
    async function refresh(quiet = false) {
      const current = ++sequence;
      try {
        const value = await request("/admin/geo/local");
        if (active && sequence === current) {
          status.value = value;
          failed.value = false;
        }
      } catch (error) {
        if (active && sequence === current) failed.value = true;
        if (!quiet) reportError(error);
      } finally {
        loading.value = false;
        poll();
      }
    }
    async function update() {
      if (busy.value) return;
      updating.value = true;
      status.value = {
        ...status.value,
        stage: "connecting",
        downloadedBytes: 0,
        totalBytes: 0,
      };
      poll();
      try {
        const value = await request("/admin/geo/local/update", {
          method: "POST",
          timeout: 240000,
        });
        if (active) {
          ++sequence;
          status.value = value;
          failed.value = false;
          Message.success("本地 IP 库已更新并生效");
        }
      } catch (error) {
        reportError(error);
        if (active) await refresh();
      } finally {
        updating.value = false;
        poll();
      }
    }
    onMounted(async () => {
      await refresh();
      if (!user.profile) await user.fetchMe().catch(() => {});
    });
    onBeforeUnmount(() => {
      active = false;
      ++sequence;
      clearTimeout(timer);
    });
    return {
      status,
      loading,
      updating,
      failed,
      busy,
      canUpdate,
      version,
      updatedAt,
      percent,
      progressText,
      refresh,
      update,
    };
  },
};
</script>

<template>
  <div class="local-geo" :class="{ 'is-busy': busy }">
    <div v-if="busy" class="local-geo-progress" aria-live="polite">
      <span class="local-geo-progress-label">{{ progressText }}</span>
      <Progress
        :percent="percent"
        :show-text="false"
        size="small"
        :status="status?.stage === 'verifying' ? 'normal' : 'active'"
      />
    </div>
    <Tooltip v-else position="top">
      <template #content>
        <div class="local-geo-tip">
          <div v-if="status?.ready">库版本：{{ version }}</div>
          <div v-if="updatedAt">最近更新：{{ updatedAt }}</div>
          <div class="local-geo-credit">
            <a
              href="https://db-ip.com/"
              target="_blank"
              rel="noopener noreferrer"
              style="
                color: inherit;
                text-decoration: underline;
                text-underline-offset: 3px;
              "
              >IP Geolocation by DB-IP</a
            >
            · CC BY 4.0
          </div>
          <div v-if="!canUpdate">安装更新需要超级管理员</div>
        </div>
      </template>
      <span class="local-geo-status">
        <template v-if="loading">读取中</template>
        <template v-else-if="failed">状态读取失败</template>
        <template v-else-if="status?.ready">
          <span class="local-geo-installed"><i></i>已安装</span>
          <span class="local-geo-time">{{ updatedAt || version }}</span>
        </template>
        <template v-else>未安装</template>
      </span>
    </Tooltip>
    <Button
      v-if="failed && !busy"
      class="local-geo-action"
      size="mini"
      type="text"
      @click="refresh()"
      >重试</Button
    >
    <Button
      v-else-if="canUpdate && !busy"
      class="local-geo-action"
      size="mini"
      type="text"
      :disabled="loading"
      @click="update"
      >{{ status?.ready ? "更新" : "安装" }}</Button
    >
  </div>
</template>

<style scoped>
.local-geo {
  display: flex;
  align-items: center;
  gap: 8px;
  height: 32px;
  padding: 0 8px 0 12px;
  box-sizing: border-box;
  background: var(--color-fill-2);
  border: 1px solid transparent;
  border-radius: var(--border-radius-small);
}
.local-geo-status {
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--color-text-2);
  font-size: 12px;
  white-space: nowrap;
}
.local-geo-installed {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  flex-shrink: 0;
}
.local-geo-installed i {
  width: 5px;
  height: 5px;
  background: rgb(var(--success-6));
  border-radius: 50%;
}
.local-geo-time {
  color: var(--color-text-2);
  font-variant-numeric: tabular-nums;
  font-size: 12px;
}
.local-geo.is-busy {
  background: var(--color-primary-light-1);
}
.local-geo-progress {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  min-width: 0;
}
.local-geo-progress-label {
  flex-shrink: 0;
  color: rgb(var(--primary-6));
  font-size: 12px;
  font-variant-numeric: tabular-nums;
}
.local-geo-progress :deep(.arco-progress) {
  flex: 1;
  min-width: 0;
}
.local-geo :deep(.local-geo-action) {
  flex-shrink: 0;
  padding: 0 4px;
  height: 22px;
  font-size: 12px;
}
.local-geo-tip {
  line-height: 1.8;
  font-size: 12px;
}
.local-geo-credit {
  opacity: 0.85;
}
</style>
