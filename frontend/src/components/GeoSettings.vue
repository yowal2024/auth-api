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
    const names = { dbip: "本地 DB-IP", tencent: "腾讯地图", amap: "高德地图" };
    const apiKey = computed({
      get: () => keys.value[provider.value] || "",
      set: (value) => {
        keys.value[provider.value] = value;
      },
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
      saving.value = true;
      try {
        await request("/admin/system/geo", {
          method: "POST",
          data: {
            provider: provider.value,
            key: provider.value === "dbip" ? "" : apiKey.value.trim(),
          },
        });
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
    <p class="geo-description">
      所有应用共用定位服务。绑定 IP 区域开关和距离阈值仍在各应用中设置。
    </p>
    <Alert v-if="legacyApps.length" type="warning" class="geo-notice">
      现有应用使用不同的定位配置，目前继续沿用各自配置。选择并保存后，所有应用将使用统一的服务。
    </Alert>
    <Space v-if="failed" class="geo-notice">
      <span>读取配置失败</span>
      <Button :loading="loading" @click="load">重试加载</Button>
    </Space>
    <Form layout="vertical" :disabled="!ready || saving">
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
      <FormItem label="定位服务">
        <Select v-model="provider" placeholder="请选择定位服务">
          <Option value="dbip">本地 DB-IP</Option>
          <Option value="tencent">腾讯地图</Option>
          <Option value="amap">高德地图</Option>
        </Select>
      </FormItem>
      <FormItem
        v-if="provider && provider !== 'dbip'"
        label="服务 KEY"
        extra="填写所选平台的应用 KEY，所有应用共用。"
      >
        <InputPassword
          v-model="apiKey"
          placeholder="对应平台的应用 KEY"
          :max-length="128"
          allow-clear
        />
      </FormItem>
      <FormItem
        v-if="provider === 'dbip'"
        label="本地 IP 库"
        extra="无需 KEY。安装完成后保存，即可供所有应用使用。"
      >
        <LocalGeo class="system-local-geo" />
      </FormItem>
      <Button type="primary" :loading="saving" :disabled="!ready" @click="save"
        >保存定位服务</Button
      >
    </Form>
  </section>
</template>

<style scoped>
.geo-settings {
  max-width: 640px;
  padding: 8px 0 24px;
}
.geo-description {
  color: var(--color-text-2);
  line-height: 1.7;
  margin: 0 0 24px;
}
.geo-notice {
  margin-bottom: 20px;
}
.system-local-geo {
  width: 100%;
}
</style>
