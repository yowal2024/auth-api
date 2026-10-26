<script>
import { ref, onMounted, onBeforeUnmount, computed } from "vue";
import { Button, Progress, Popover, Message } from "@arco-design/web-vue";
import { IconInfoCircle } from "@arco-design/web-vue/es/icon";
import { request } from "@/utils/request.js";
import { reportError } from "@/utils/errors.js";
import { useUserStore } from "@/stores/user.js";

export default {
  components: { Button, Progress, Popover, IconInfoCircle },
  setup() {
    const user = useUserStore();
    const status = ref(null);
    const loading = ref(true);
    const updating = ref(false);
    const failed = ref(false);
    const infoVisible = ref(false);
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
      infoVisible,
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
    <div v-else class="local-geo-details">
      <div class="local-geo-status">
        <span v-if="loading">读取中</span>
        <span v-else-if="failed">状态读取失败</span>
        <template v-else-if="status?.ready">
          <span class="local-geo-installed"
            ><i aria-hidden="true"></i>已安装</span
          >
          <span class="local-geo-version">库版本：{{ version }}</span>
        </template>
        <span v-else>未安装</span>
        <Popover
          v-model:popup-visible="infoVisible"
          position="bl"
          trigger="click"
          title="本地 IP 库说明"
          :content-style="{
            maxWidth: 'min(300px, calc(100vw - 32px))',
            borderRadius: '0',
          }"
        >
          <Button
            class="local-geo-info"
            type="text"
            size="mini"
            aria-label="本地 IP 库说明"
            :aria-expanded="infoVisible"
            @keydown.esc.stop="infoVisible = false"
            ><template #icon><IconInfoCircle /></template
          ></Button>
          <template #content>
            <div class="local-geo-tip">
              <p>使用本地 IP 库查询，无需在线服务 KEY。</p>
              <p class="local-geo-credit">
                <a
                  href="https://db-ip.com/"
                  target="_blank"
                  rel="noopener noreferrer"
                  >IP Geolocation by DB-IP</a
                ><br />CC BY 4.0
              </p>
              <p v-if="!canUpdate">安装更新需要超级管理员。</p>
            </div>
          </template>
        </Popover>
      </div>
      <div v-if="updatedAt && !loading && !failed" class="local-geo-updated">
        最近更新 <time class="local-geo-time">{{ updatedAt }}</time>
      </div>
    </div>
    <Button
      v-if="failed && !busy"
      class="local-geo-action"
      size="small"
      @click="refresh()"
      >重试</Button
    >
    <Button
      v-else-if="canUpdate && !busy"
      class="local-geo-action"
      size="small"
      :disabled="loading"
      @click="update"
      >{{ status?.ready ? "更新" : "安装" }}</Button
    >
  </div>
</template>
<style scoped>
.local-geo {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: center;
  gap: 16px;
  min-height: 72px;
  padding: 14px 16px;
  background: var(--color-fill-1);
  border: 1px solid var(--color-border-2);
}
.local-geo-details {
  min-width: 0;
}
.local-geo-status {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px 12px;
  color: var(--color-text-2);
  font-size: 13px;
}
.local-geo-installed {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
}
.local-geo-installed i {
  width: 5px;
  height: 5px;
  background: rgb(var(--success-6));
  border-radius: 50%;
}
.local-geo-version,
.local-geo-updated {
  color: var(--color-text-3);
  font-size: 12px;
  line-height: 1.7;
}
.local-geo-updated {
  display: flex;
  align-items: baseline;
  flex-wrap: wrap;
  gap: 0 8px;
  margin-top: 5px;
}
.local-geo-time {
  font-variant-numeric: tabular-nums;
}
.local-geo :deep(.local-geo-info) {
  flex-shrink: 0;
  padding: 0;
  width: 22px;
  height: 22px;
  color: var(--color-text-3);
}
.local-geo :deep(.local-geo-info:focus-visible) {
  outline: 2px solid rgb(var(--primary-6));
  outline-offset: 2px;
}
.local-geo.is-busy {
  background: var(--color-primary-light-1);
}
.local-geo-progress {
  grid-column: 1 / -1;
  display: flex;
  align-items: center;
  gap: 12px;
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
  border-radius: 0;
}
.local-geo-tip {
  font-size: 12px;
  line-height: 1.8;
}
.local-geo-tip p {
  margin: 0;
}
.local-geo-credit {
  margin-top: 8px !important;
  color: var(--color-text-3);
}
.local-geo-credit a {
  color: rgb(var(--primary-6));
  text-decoration: underline;
  text-underline-offset: 3px;
}
@media (max-width: 600px) {
  .local-geo {
    padding: 12px;
    gap: 10px;
  }
  .local-geo-status {
    gap: 6px 8px;
  }
}
</style>
