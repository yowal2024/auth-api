<script>
import { computed, onMounted, ref } from "vue";
import {
  Alert,
  Button,
  Form,
  FormItem,
  InputPassword,
  Message,
  Option,
  Select,
  Space,
} from "@arco-design/web-vue";
import {
  IconInfoCircle,
  IconLocation,
  IconLock,
  IconSave,
  IconSettings,
  IconStorage,
  IconSync,
} from "@arco-design/web-vue/es/icon";
import SettingsHeader from "@/components/SettingsHeader.vue";
import LocalGeo from "@/components/LocalGeo.vue";
import { request } from "@/utils/request.js";
import { reportError } from "@/utils/errors.js";

export default {
  components: {
    Alert,
    Button,
    Form,
    FormItem,
    InputPassword,
    Option,
    Select,
    Space,
    LocalGeo,
    SettingsHeader,
    IconInfoCircle,
    IconLocation,
    IconLock,
    IconSave,
    IconSettings,
    IconStorage,
    IconSync,
  },
  setup() {
    const provider = ref("");
    const keys = ref({});
    const ready = ref(false);
    const loading = ref(false);
    const saving = ref(false);
    const failed = ref(false);
    const legacyApps = ref([]);
    const sourceApp = ref();
    const savedConfig = ref(null);
    const names = { dbip: "本地 DB-IP", tencent: "腾讯地图", amap: "高德地图" };
    const services = [
      { id: "dbip", hint: "本地查询 · 无需 KEY" },
      { id: "tencent", hint: "在线查询 · 需要 KEY" },
      { id: "amap", hint: "在线查询 · 需要 KEY" },
    ];
    const apiKey = computed({
      get: () => keys.value[provider.value] || "",
      set: (value) => {
        keys.value[provider.value] = value;
      },
    });
    const dirty = computed(
      () =>
        ready.value &&
        (provider.value !== savedConfig.value?.provider ||
          (provider.value !== "dbip" &&
            apiKey.value.trim() !== savedConfig.value?.key)),
    );
    const saveState = computed(() => {
      if (failed.value) return { label: "读取失败", tone: "error" };
      if (!ready.value || loading.value)
        return { label: "读取中", tone: "neutral" };
      if (saving.value) return { label: "保存中", tone: "neutral" };
      if (dirty.value) return { label: "待保存", tone: "pending" };
      if (!savedConfig.value?.provider)
        return { label: "待统一配置", tone: "pending" };
      return { label: "已保存", tone: "saved" };
    });
    const savedProvider = computed(() => {
      if (failed.value) return "暂时无法读取";
      if (!ready.value) return "正在读取配置…";
      return names[savedConfig.value?.provider] || "沿用各应用配置";
    });

    async function load() {
      loading.value = true;
      failed.value = false;
      try {
        const result = await request("/admin/system/geo");
        provider.value = result.provider || "";
        keys.value = result.provider
          ? { [result.provider]: result.key || "" }
          : {};
        legacyApps.value = result.legacyApps || [];
        savedConfig.value = {
          provider: provider.value,
          key: provider.value === "dbip" ? "" : apiKey.value.trim(),
        };
        ready.value = true;
      } catch (error) {
        failed.value = true;
        reportError(error);
      } finally {
        loading.value = false;
      }
    }
    function loadLegacy(id) {
      const app = legacyApps.value.find((app) => app.id === id);
      if (app) {
        provider.value = app.provider;
        keys.value[app.provider] = app.key || "";
      }
    }
    async function save() {
      if (!ready.value || saving.value) return;
      if (!provider.value) return Message.warning("请选择 IP 定位服务");
      if (provider.value !== "dbip" && !apiKey.value.trim())
        return Message.warning("请填写服务 KEY");
      const data = {
        provider: provider.value,
        key: provider.value === "dbip" ? "" : apiKey.value.trim(),
      };
      saving.value = true;
      try {
        await request("/admin/system/geo", {
          method: "POST",
          data,
        });
        savedConfig.value = data;
        keys.value[data.provider] = data.key;
        legacyApps.value = [];
        Message.success("已保存，所有应用立即生效");
      } catch (error) {
        reportError(error);
      } finally {
        saving.value = false;
      }
    }
    onMounted(load);
    return {
      provider,
      ready,
      loading,
      saving,
      failed,
      legacyApps,
      sourceApp,
      names,
      services,
      saveState,
      savedProvider,
      savedConfig,
      apiKey,
      load,
      loadLegacy,
      save,
    };
  },
};
</script>

<template>
  <section class="geo-settings">
    <SettingsHeader
      title="IP 定位服务"
      description="统一管理定位来源，为应用的 IP 区域校验提供支持。"
      note="系统级配置"
      ><IconLocation
    /></SettingsHeader>
    <Alert v-if="legacyApps.length" type="warning" class="geo-notice">
      现有应用使用不同的定位配置，目前继续沿用各自配置。选择并保存后，所有应用将使用统一的服务。
    </Alert>
    <Space v-if="failed" class="geo-notice">
      <span>读取配置失败</span>
      <Button :loading="loading" @click="load">重试加载</Button>
    </Space>
    <div class="geo-layout">
      <section class="geo-config-panel" aria-labelledby="geo-config-title">
        <div class="geo-panel-heading">
          <div>
            <h3 id="geo-config-title">服务配置</h3>
            <p>选择一种定位服务，并完成对应配置。</p>
          </div>
          <span class="geo-state" :class="saveState.tone" role="status">
            <i aria-hidden="true"></i>{{ saveState.label }}
          </span>
        </div>
        <Form class="geo-form" layout="vertical" :disabled="!ready || saving">
          <div class="geo-form-fields">
            <FormItem
              v-if="legacyApps.length"
              label="载入已有应用配置"
              extra="载入后可修改，点击保存才会统一生效。"
            >
              <Select
                v-model="sourceApp"
                placeholder="选择已有配置"
                @change="loadLegacy"
              >
                <Option v-for="app in legacyApps" :key="app.id" :value="app.id"
                  >{{ app.name }} · {{ names[app.provider] }}</Option
                >
              </Select>
            </FormItem>
            <fieldset class="geo-provider-field" :disabled="!ready || saving">
              <legend>定位服务</legend>
              <div class="geo-providers">
                <label
                  v-for="service in services"
                  :key="service.id"
                  class="geo-provider-card"
                  :class="{ selected: provider === service.id }"
                >
                  <input
                    v-model="provider"
                    class="geo-provider-radio"
                    type="radio"
                    name="geo-provider"
                    :value="service.id"
                    :aria-label="names[service.id]"
                  />
                  <span class="geo-provider-icon">
                    <IconStorage
                      v-if="service.id === 'dbip'"
                      aria-hidden="true"
                    />
                    <IconLocation v-else aria-hidden="true" />
                  </span>
                  <strong>{{ names[service.id] }}</strong>
                  <span class="geo-provider-hint">{{ service.hint }}</span>
                </label>
              </div>
            </fieldset>
            <FormItem
              v-if="provider && provider !== 'dbip'"
              label="服务 KEY"
              :extra="`填写${names[provider]}的应用 KEY，保存后供所有应用共用。`"
            >
              <InputPassword
                v-model="apiKey"
                placeholder="对应平台的应用 KEY"
                :max-length="128"
                allow-clear
              >
                <template #prefix><IconLock aria-hidden="true" /></template>
              </InputPassword>
            </FormItem>
            <FormItem
              v-else-if="provider === 'dbip'"
              label="本地 IP 库"
              extra="无需 KEY。安装完成后保存，即可供所有应用使用。"
            >
              <LocalGeo class="system-local-geo" />
            </FormItem>
            <div v-else class="geo-empty-hint">
              <IconInfoCircle aria-hidden="true" />选择上方服务，继续完成配置。
            </div>
          </div>
          <div class="geo-form-actions">
            <Button
              type="primary"
              :loading="saving"
              :disabled="!ready"
              @click="save"
            >
              <template #icon><IconSave /></template>保存定位服务
            </Button>
            <span>修改配置后，记得保存生效。</span>
          </div>
        </Form>
      </section>
      <aside class="geo-guide" aria-labelledby="geo-guide-title">
        <h3 id="geo-guide-title">配置概览</h3>
        <div class="geo-current">
          <span>当前已保存的服务</span>
          <strong class="geo-current-name">{{ savedProvider }}</strong>
        </div>
        <dl class="geo-summary">
          <div>
            <dt>适用范围</dt>
            <dd>
              {{
                !ready
                  ? "—"
                  : savedConfig.provider
                    ? "所有应用"
                    : "各应用原有配置"
              }}
            </dd>
          </div>
          <div>
            <dt>服务凭证</dt>
            <dd>
              {{
                !ready
                  ? "—"
                  : !savedConfig.provider
                    ? "沿用原配置"
                    : savedConfig.provider === "dbip"
                      ? "无需 KEY"
                      : savedConfig.key
                        ? "已配置 KEY"
                        : "未配置 KEY"
              }}
            </dd>
          </div>
        </dl>
        <div class="geo-guide-items">
          <div class="geo-guide-item">
            <IconSettings aria-hidden="true" />
            <div>
              <h4>绑定规则独立设置</h4>
              <p>
                绑定 IP 区域开关和距离阈值，仍在各应用的「应用信息」中设置。
              </p>
            </div>
          </div>
          <div class="geo-guide-item">
            <IconSync aria-hidden="true" />
            <div>
              <h4>保存后自动应用</h4>
              <p>新的服务配置将在后续区域校验时使用，无需重启服务。</p>
            </div>
          </div>
        </div>
      </aside>
    </div>
  </section>
</template>

<style scoped>
.geo-notice {
  margin-bottom: 20px;
}
.geo-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 320px;
  gap: 20px;
}
.geo-config-panel,
.geo-guide {
  min-width: 0;
  border: 1px solid var(--color-border-2);
  border-radius: 0;
}
.geo-config-panel {
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
.geo-panel-heading {
  display: flex;
  align-items: center;
  gap: 14px;
  justify-content: space-between;
  padding: 22px 24px 0;
}
.geo-panel-heading p {
  margin: 5px 0 0;
  color: var(--color-text-3);
  font-size: 13px;
  line-height: 1.6;
}
.geo-settings h3 {
  margin: 0;
  color: var(--color-text-1);
  font-size: 15px;
  font-weight: 600;
}
.geo-state {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
  padding: 4px 9px;
  border-radius: 0;
  background: var(--color-fill-2);
  color: var(--color-text-3);
  font-size: 12px;
  line-height: 1.5;
}
.geo-state i {
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: currentColor;
}
.geo-state.saved {
  color: rgb(var(--success-7));
  background: var(--color-success-light-1);
}
.geo-state.pending {
  color: rgb(var(--warning-7));
  background: var(--color-warning-light-1);
}
.geo-state.error {
  color: rgb(var(--danger-7));
  background: var(--color-danger-light-1);
}
.geo-form {
  display: flex;
  flex: 1;
  flex-direction: column;
}
.geo-form-fields {
  flex: 1;
  padding: 24px;
}
.geo-provider-field {
  min-width: 0;
  margin: 0 0 24px;
  padding: 0;
  border: 0;
}
.geo-provider-field legend {
  margin-bottom: 12px;
  padding: 0;
  font-size: 14px;
  color: var(--color-text-1);
}
.geo-providers {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}
.geo-provider-card {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-width: 0;
  padding: 16px;
  border: 1px solid var(--color-border-2);
  border-radius: 0;
  cursor: pointer;
  transition:
    border-color 0.15s,
    background-color 0.15s;
}
.geo-provider-card:hover {
  border-color: rgb(var(--primary-5));
}
.geo-provider-card.selected {
  border-color: rgb(var(--primary-6));
  background: var(--color-primary-light-1);
}
.geo-provider-field:disabled .geo-provider-card {
  cursor: not-allowed;
  opacity: 0.6;
}
.geo-provider-radio {
  position: absolute;
  top: 16px;
  right: 14px;
  width: 15px;
  height: 15px;
  margin: 0;
  accent-color: rgb(var(--primary-6));
  cursor: inherit;
}
.geo-provider-radio:focus-visible {
  outline: 2px solid rgb(var(--primary-6));
  outline-offset: 3px;
}
:global(body[arco-theme="dark"]) .geo-provider-radio {
  color-scheme: dark;
}
.geo-provider-icon {
  display: grid;
  place-items: center;
  width: 32px;
  height: 32px;
  margin-bottom: 4px;
  border-radius: 0;
  background: var(--color-fill-2);
  color: var(--color-text-2);
  font-size: 20px;
}
.selected .geo-provider-icon {
  background: var(--color-primary-light-2);
  color: rgb(var(--primary-6));
}
.geo-provider-card strong {
  color: var(--color-text-1);
  font-size: 14px;
  font-weight: 500;
  line-height: 1.5;
}
.geo-provider-hint {
  color: var(--color-text-3);
  font-size: 12px;
  line-height: 1.5;
}
.geo-form-fields :deep(.arco-form-item:last-child) {
  margin-bottom: 0;
}
.geo-form-fields :deep(.arco-input-wrapper) {
  min-height: 40px;
  border-radius: 0;
}
.geo-form-actions {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px 16px;
  padding: 18px 24px;
  border-top: 1px solid var(--color-border-2);
  background: var(--color-fill-1);
}
.geo-form-actions :deep(.arco-btn) {
  flex-shrink: 0;
  height: 36px;
  border-radius: 0;
}
.geo-form-actions > span,
.geo-empty-hint {
  color: var(--color-text-3);
  font-size: 12px;
  line-height: 1.6;
}
.geo-empty-hint {
  display: flex;
  align-items: center;
  gap: 8px;
  min-height: 64px;
}
.system-local-geo {
  width: 100%;
}
.geo-guide {
  padding: 24px;
  background: var(--color-fill-1);
}
.geo-current {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-top: 24px;
}
.geo-current > span {
  color: var(--color-text-3);
  font-size: 12px;
}
.geo-current-name {
  color: var(--color-text-1);
  font-size: 20px;
  font-weight: 600;
  line-height: 1.5;
}
.geo-summary {
  display: grid;
  gap: 14px;
  margin: 20px 0 24px;
  font-size: 13px;
  line-height: 1.5;
}
.geo-summary > div {
  display: flex;
  justify-content: space-between;
  gap: 12px;
}
.geo-summary dt {
  color: var(--color-text-3);
}
.geo-summary dd {
  margin: 0;
  color: var(--color-text-2);
  text-align: right;
}
.geo-guide-items {
  display: grid;
  gap: 20px;
  padding-top: 22px;
  border-top: 1px solid var(--color-border-2);
}
.geo-guide-item {
  display: flex;
  align-items: flex-start;
  gap: 10px;
}
.geo-guide-item > svg {
  flex-shrink: 0;
  margin-top: 2px;
  color: var(--color-text-3);
  font-size: 17px;
}
.geo-guide-item h4 {
  margin: 0;
  color: var(--color-text-2);
  font-size: 13px;
  font-weight: 500;
  line-height: 1.6;
}
.geo-guide-item p {
  margin: 6px 0 0;
  color: var(--color-text-3);
  font-size: 12px;
  line-height: 1.8;
}
@media (max-width: 1180px) {
  .geo-layout {
    grid-template-columns: minmax(0, 1fr);
  }
  .geo-guide-items {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 24px;
  }
}
@media (max-width: 600px) {
  .geo-panel-heading,
  .geo-form-fields,
  .geo-form-actions,
  .geo-guide {
    padding: 16px;
  }
  .geo-panel-heading {
    align-items: flex-start;
    padding-bottom: 0;
  }
  .geo-panel-heading p {
    font-size: 12px;
  }
  .geo-providers {
    grid-template-columns: minmax(0, 1fr);
    gap: 10px;
  }
  .geo-provider-card {
    display: grid;
    grid-template-columns: 32px minmax(0, 1fr);
    gap: 4px 12px;
    padding: 14px 36px 14px 14px;
  }
  .geo-provider-icon {
    grid-row: 1 / 3;
    align-self: center;
    margin: 0;
  }
  .geo-provider-radio {
    top: calc(50% - 7.5px);
  }
  .geo-form-actions > span {
    flex-basis: 100%;
  }
  .geo-guide-items {
    grid-template-columns: minmax(0, 1fr);
    gap: 20px;
  }
}
@media (prefers-reduced-motion: reduce) {
  .geo-provider-card {
    transition: none;
  }
}
</style>
