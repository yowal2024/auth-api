<!-- Reconstructed editable Vue source; see README for recovery limits. -->
<script>
import { usePagedList } from "@/composables/usePagedList.js";
import { listLogs, listLogActions, clearLogs } from "@/api/system.js";
import { formatDateTime } from "@/utils/format.js";
import { onMounted, ref, reactive, computed } from "vue";
import {
  Popconfirm,
  Select,
  RangePicker,
  InputSearch,
  Table,
  Button,
  Option,
  Tooltip,
  Message,
} from "@arco-design/web-vue";
import { IconDelete } from "@arco-design/web-vue/es/icon";
export default {
  name: "AdminLogs",
  components: {
    Popconfirm,
    Button,
    IconDelete,
    Select,
    Option,
    RangePicker,
    InputSearch,
    Table,
    Tooltip,
  },
  setup(ge) {
    const a = reactive({
      category: "admin",
      action: void 0,
      keyword: "",
      start: void 0,
      end: void 0,
      page: 1,
      size: 10,
    });
    const pageState = ref([]);
    const rows = ref([]);
    const p = ref(0);
    const loading = ref(false);
    const c = ref([]);
    const f = ref([]);
    const tableColumns = [
      {
        title: "操作类型",
        slotName: "action",
        width: 200,
        ellipsis: true,
        tooltip: true,
      },
      {
        title: "操作数据",
        slotName: "data",
        width: 300,
        ellipsis: true,
        tooltip: true,
      },
      {
        title: "操作IP",
        slotName: "ip",
        align: "right",
        width: 200,
        ellipsis: true,
        tooltip: true,
      },
      {
        title: "操作时间",
        slotName: "createTime",
        align: "right",
        width: 200,
      },
    ];
    const pagination = computed(() => ({
      current: a.page,
      pageSize: a.size,
      total: p.value,
      showTotal: false,
      showPageSize: true,
      pageSizeOptions: [10, 20, 50, 100],
    }));
    const V = computed(() => Math.max(1, Math.ceil(p.value / a.size)));
    const handleChange = (n) => {
      a.start = n?.[0]
        ? Math.floor(new Date(n[0] + " 00:00:00").getTime() / 1e3)
        : void 0;
      a.end = n?.[1]
        ? Math.floor(new Date(n[1] + " 23:59:59").getTime() / 1e3)
        : void 0;
      handleListLogs();
    };
    const handleListLogs = usePagedList(listLogs, a, rows, p, loading, c);
    const handleListLogActions = async () => {
      try {
        f.value = await listLogActions({
          category: "admin",
        });
      } catch {}
    };
    const handleClearLogs = async () => {
      await clearLogs(
        c.value.length
          ? {
              ids: c.value,
              category: "admin",
            }
          : {
              category: "admin",
            },
      );
      Message.success("已清理");
      c.value = [];
      handleListLogs();
      handleListLogActions();
    };
    onMounted(() => {
      handleListLogs();
      handleListLogActions();
    });
    return {
      a,
      pageState,
      p,
      c,
      f,
      tableColumns,
      V,
      handleChange,
      handleListLogs,
      handleListLogActions,
      handleClearLogs,
      rows,
      loading,
      pagination,
      formatDateTime,
    };
  },
};
</script>
<template>
  <div class="page">
    <div class="page-card">
      <div class="toolbar">
        <Popconfirm
          :content="
            c.length
              ? `确定清理选中的 ${c.length} 条日志?`
              : '确定清理当前应用的全部管理员日志?'
          "
          @ok="handleClearLogs"
        >
          <template #default>
            <Button type="primary">
              <template #icon><IconDelete></IconDelete></template>
              <template #default>清理日志</template>
            </Button>
          </template>
        </Popconfirm>
        <div class="grow"></div>
        <Select
          v-model="a.action"
          placeholder="全部日志分类"
          allow-clear=""
          :style="{
            width: '180px',
          }"
          @change="handleListLogs"
        >
          <template #default>
            <template v-for="e in f" :key="e">
              <Option :value="e">
                <template #default>{{ e }}</template>
              </Option>
            </template>
          </template>
        </Select>
        <RangePicker
          v-model="pageState"
          :style="{
            width: '260px',
          }"
          value-format="YYYY-MM-DD"
          @change="handleChange"
        ></RangePicker>
        <InputSearch
          v-model="a.keyword"
          placeholder="管理员搜索..."
          allow-clear=""
          :style="{
            width: '220px',
          }"
          @search="handleListLogs"
          @pressEnter="handleListLogs"
          @clear="handleListLogs"
        ></InputSearch>
      </div>
      <Table
        :data="rows"
        :columns="tableColumns"
        :loading="loading"
        :pagination="pagination"
        :row-selection="{
          type: 'checkbox',
          showCheckedAll: !0,
        }"
        v-model:selectedKeys="c"
        row-key="id"
        @pageChange="
          (e) => {
            ((a.page = e), handleListLogs());
          }
        "
        @pageSizeChange="
          (e) => {
            ((a.size = e), (a.page = 1), handleListLogs());
          }
        "
      >
        <template #action="{ record: e }">
          <div class="log-cell">
            <div class="log-title">{{ e.summary || e.action }}</div>
            <div class="log-sub">{{ e.operator }}</div>
          </div>
        </template>
        <template #data="{ record: e }">
          <Tooltip :content="e.data" :disabled="!e.data">
            <template #default>
              <div class="log-data">{{ e.data || "-" }}</div>
            </template>
          </Tooltip>
        </template>
        <template #ip="{ record: e }">
          <div class="log-cell log-cell-right">
            <div class="log-title">{{ e.region || "-" }}</div>
            <div class="log-sub">{{ e.ip }}</div>
          </div>
        </template>
        <template #createTime="{ record: e }">{{
          formatDateTime(e.createTime)
        }}</template>
      </Table>
      <div class="log-footer">
        <span>{{
          "当前第 " + a.page + " 页 共 " + V + " 页 " + p + " 条结果"
        }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.log-cell {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.log-cell-right {
  align-items: flex-end;
}
.log-cell .log-title {
  color: var(--color-text-1);
  font-size: 14px;
}
.log-cell .log-sub {
  color: var(--color-text-3);
  font-size: 12px;
}
.log-data {
  color: var(--color-text-2);
  font-size: 13px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.log-footer {
  margin-top: 12px;
  color: var(--color-text-3);
  font-size: 13px;
}
</style>
