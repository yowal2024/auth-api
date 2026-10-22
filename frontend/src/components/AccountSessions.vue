<script>
import { ref, watch, onBeforeUnmount } from "vue";
import {
  Table,
  Button,
  Popconfirm,
  Space,
  Tag,
  Message,
} from "@arco-design/web-vue";
import { request } from "@/utils/request.js";
import { formatDateTime } from "@/utils/format.js";
export default {
  name: "AccountSessions",
  components: { Table, Button, Popconfirm, Space, Tag },
  props: {
    accountId: { type: Number, required: true },
    kind: { type: String, default: "user" },
  },
  setup(props) {
    const rows = ref([]),
      loading = ref(false),
      busy = ref(false);
    let timer,
      generation = 0,
      disposed = false;
    const base = () => `/admin/${props.kind}/${props.accountId}/sessions`;
    const refresh = async () => {
      const current = ++generation;
      if (!props.accountId) return;
      loading.value = true;
      try {
        const data = await request(base());
        if (current === generation && !disposed) rows.value = data || [];
      } catch {
      } finally {
        if (current === generation) loading.value = false;
      }
    };
    const kick = async (id) => {
      if (busy.value) return;
      busy.value = true;
      try {
        await request(base() + (id ? `/${id}` : ""), { method: "DELETE" });
        Message.success(id ? "该会话已下线" : "全部会话已下线");
        await refresh();
      } catch {
      } finally {
        busy.value = false;
      }
    };
    watch(
      () => [props.accountId, props.kind],
      () => {
        rows.value = [];
        refresh();
      },
      { immediate: true },
    );
    timer = setInterval(() => {
      if (!document.hidden && !loading.value && !busy.value) refresh();
    }, 15000);
    onBeforeUnmount(() => {
      disposed = true;
      generation++;
      clearInterval(timer);
    });
    const columns = [
      {
        title: "设备 / 机器码",
        dataIndex: "mi",
        width: 180,
        ellipsis: true,
        tooltip: true,
      },
      { title: "登录 IP", dataIndex: "loginIp", width: 140 },
      { title: "最近 IP", dataIndex: "lastIp", width: 140 },
      { title: "登录时间", slotName: "issuedAt", width: 190 },
      { title: "最近活跃", slotName: "lastActiveAt", width: 190 },
      { title: "最近心跳", slotName: "lastHeartbeatAt", width: 190 },
      { title: "操作", slotName: "actions", width: 100, fixed: "right" },
    ];
    return { rows, loading, busy, refresh, kick, columns, formatDateTime };
  },
};
</script>
<template>
  <div class="account-sessions">
    <Space class="session-tools" wrap>
      <Button :loading="loading" @click="refresh">刷新</Button>
      <Popconfirm
        content="下线该账号的全部会话？客户端需要重新登录。"
        @ok="kick()"
      >
        <Button status="danger" :disabled="!rows.length || busy"
          >全部下线</Button
        >
      </Popconfirm>
      <span class="session-hint"
        >每 15
        秒刷新；会话有效不代表设备持续在线，最近心跳以服务端接收时间为准。</span
      >
    </Space>
    <Table
      :data="rows"
      :columns="columns"
      row-key="id"
      :loading="loading"
      :scroll="{ x: 1230 }"
      :pagination="{ pageSize: 10 }"
    >
      <template #issuedAt="{ record }">{{
        record.issuedAt ? formatDateTime(record.issuedAt) : "—"
      }}</template>
      <template #lastActiveAt="{ record }">{{
        record.lastActiveAt ? formatDateTime(record.lastActiveAt) : "—"
      }}</template>
      <template #lastHeartbeatAt="{ record }">{{
        record.lastHeartbeatAt
          ? formatDateTime(record.lastHeartbeatAt)
          : "尚未收到"
      }}</template>
      <template #actions="{ record }"
        ><Popconfirm content="下线这个会话？" @ok="kick(record.id)"
          ><Button type="text" status="danger" :disabled="busy"
            >下线</Button
          ></Popconfirm
        ></template
      >
    </Table>
  </div>
</template>
<style scoped>
.account-sessions {
  min-width: 0;
}
.session-tools {
  margin-bottom: 16px;
}
.session-hint {
  color: var(--color-text-3);
  font-size: 12px;
}
</style>
