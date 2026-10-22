<script>
import { ref, onMounted } from "vue";
import {
  Table,
  Button,
  Input,
  Select,
  Option,
  Space,
} from "@arco-design/web-vue";
import { request } from "@/utils/request.js";
import { formatDateTime } from "@/utils/format.js";
export default {
  name: "SecurityEvents",
  components: { Table, Button, Input, Select, Option, Space },
  setup() {
    const rows = ref([]),
      total = ref(0),
      page = ref(1),
      kind = ref(""),
      ip = ref(""),
      loading = ref(false);
    const load = async (p = 1) => {
      loading.value = true;
      try {
        const v = await request("/admin/system/security/events", {
          params: { page: p, size: 20, kind: kind.value, ip: ip.value },
        });
        rows.value = v.list;
        total.value = v.total;
        page.value = p;
      } catch {
      } finally {
        loading.value = false;
      }
    };
    onMounted(() => load());
    const kinds = {
      rate_limit: "频率限制",
      login_failed: "登录失败",
      card_rejected: "卡密验证失败",
      signature_rejected: "签名或加密异常",
      blocked: "封禁拦截",
    };
    const columns = [
      { title: "时间", slotName: "time", width: 180 },
      { title: "应用 ID", dataIndex: "appid", width: 110 },
      { title: "事件", dataIndex: "summary" },
      { title: "来源 IP", dataIndex: "ip", width: 170 },
      { title: "接口", dataIndex: "route", ellipsis: true, tooltip: true },
    ];
    return {
      rows,
      total,
      page,
      kind,
      ip,
      loading,
      load,
      kinds,
      columns,
      formatDateTime,
    };
  },
};
</script>
<template>
  <div>
    <Space wrap class="security-tools"
      ><Select
        v-model="kind"
        placeholder="事件类型"
        allow-clear
        style="width: 180px"
        ><Option v-for="(label, value) in kinds" :key="value" :value="value">{{
          label
        }}</Option></Select
      ><Input
        v-model="ip"
        placeholder="来源 IP"
        style="width: 200px"
        allow-clear
        @press-enter="load()"
      /><Button type="primary" :loading="loading" @click="load()"
        >查询</Button
      ></Space
    >
    <p class="security-hint">
      保留最近 30
      天；相同应用、IP、接口和事件类型每分钟记录一次，避免异常流量刷满日志。
    </p>
    <Table
      :data="rows"
      :columns="columns"
      row-key="id"
      :loading="loading"
      :scroll="{ x: 850 }"
      :pagination="{ total, current: page, pageSize: 20 }"
      @page-change="load"
      ><template #time="{ record }">{{
        formatDateTime(record.createTime)
      }}</template></Table
    >
  </div>
</template>
<style scoped>
.security-tools {
  margin-bottom: 8px;
}
.security-hint {
  color: var(--color-text-3);
  font-size: 12px;
  margin-bottom: 16px;
}
</style>
