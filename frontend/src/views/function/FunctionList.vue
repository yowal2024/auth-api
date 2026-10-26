<!-- Reconstructed editable Vue source; see README for recovery limits. -->
<script>
import { reportError } from "@/utils/errors.js";
import { watch, onMounted, reactive, computed, ref } from "vue";
import {
  Button,
  Select,
  Input,
  Alert,
  Table,
  Option,
  Link,
  Space,
  Popconfirm,
  Switch,
  Tag,
  Message,
} from "@arco-design/web-vue";
import { useRouter } from "vue-router";
import { IconPlus, IconSearch } from "@arco-design/web-vue/es/icon";
import {
  listFunctions,
  runFunction,
  deleteFunction,
  saveFunction,
} from "@/api/function.js";
import { useAppStore } from "@/stores/app.js";

import { formatDateTime } from "@/utils/format.js";
import { entryPath } from "@/utils/route.js";
export default {
  name: "FunctionList",
  components: {
    Button,
    IconPlus,
    Select,
    Option,
    Input,
    IconSearch,
    Alert,
    Link,
    Table,
    Tag,
    Switch,
    Space,
    Popconfirm,
  },
  setup(yt) {
    const router = useRouter();
    const appStore = useAppStore();
    const i = reactive({
      keyword: "",
      trigger: void 0,
    });
    const rows = ref([]);
    const loading = ref(false);
    const w = [
      {
        value: "before_login",
        label: "登录前",
      },
      {
        value: "after_login",
        label: "登录后",
      },
      {
        value: "before_register",
        label: "注册前",
      },
      {
        value: "after_register",
        label: "注册后",
      },
      {
        value: "kami_use",
        label: "卡密使用",
      },
      {
        value: "manual",
        label: "手动调用",
      },
    ];
    const x = (n) => w.find((t) => t.value === n)?.label || n;
    const pagination = computed(() => ({
      total: rows.value.length,
      showTotal: true,
      showPageSize: true,
      pageSizeOptions: [10, 20, 50, 100],
    }));
    const tableColumns = [
      {
        title: "应用内编号",
        dataIndex: "appRecordId",
        width: 110,
      },
      {
        title: "函数名",
        dataIndex: "name",
        width: 200,
        ellipsis: true,
        tooltip: true,
      },
      {
        title: "触发器",
        slotName: "trigger",
        width: 140,
      },
      {
        title: "描述",
        dataIndex: "remark",
        width: 280,
        ellipsis: true,
        tooltip: true,
      },
      {
        title: "启用",
        slotName: "enabled",
        width: 80,
      },
      {
        title: "最近执行",
        slotName: "lastRun",
        width: 200,
      },
      {
        title: "操作",
        slotName: "operations",
        fixed: "right",
        width: 200,
      },
    ];
    const handleListFunctions = async () => {
      loading.value = true;
      try {
        rows.value = await listFunctions({
          appid: appStore.currentId,
          ...i,
        });
      } finally {
        loading.value = false;
      }
    };
    const handleClick = () => router.push(entryPath("/function/edit/0"));
    const I = (n) => router.push(entryPath(`/function/edit/${n.id}`));
    const handleSaveFunction = async (n, t) => {
      await saveFunction({
        ...n,
        enabled: t,
      });
      Message.success(t ? "已启用" : "已停用");
      handleListFunctions();
    };
    const handleDeleteFunction = async (n) => {
      await deleteFunction(n.id);
      Message.success("已删除");
      handleListFunctions();
    };
    const handleRunFunction = async (n) => {
      try {
        const t = await runFunction(n.id, {});
        if (t?.ok === false) {
          Message.error(t.error || "函数执行失败");
          return;
        }
        Message.success(`返回: ${JSON.stringify(t?.result ?? t)}`);
        await handleListFunctions();
      } catch (error) {
        reportError(error);
      }
    };
    watch(() => appStore.currentId, handleListFunctions);
    onMounted(handleListFunctions);
    return {
      entryPath,
      router,
      appStore,
      i,
      w,
      x,
      tableColumns,
      handleListFunctions,
      handleClick,
      I,
      handleSaveFunction,
      handleDeleteFunction,
      handleRunFunction,
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
        <Button type="primary" @click="handleClick">
          <template #icon><IconPlus></IconPlus></template>
          <template #default>新增云函数</template>
        </Button>
        <div class="grow"></div>
        <Select
          v-model="i.trigger"
          placeholder="触发器"
          allow-clear=""
          :style="{
            width: '160px',
          }"
          @change="handleListFunctions"
        >
          <template #default>
            <template v-for="e in w" :key="e.value">
              <Option :value="e.value">
                <template #default>{{ e.label }}</template>
              </Option>
            </template>
          </template>
        </Select>
        <Input
          v-model="i.keyword"
          placeholder="函数名搜索"
          allow-clear=""
          :style="{
            width: '220px',
          }"
          @pressEnter="handleListFunctions"
          @clear="handleListFunctions"
        >
          <template #suffix>
            <IconSearch
              :style="{
                cursor: 'pointer',
              }"
              @click="handleListFunctions"
            ></IconSearch>
          </template>
        </Input>
      </div>
      <Alert
        type="info"
        closable=""
        :style="{
          'margin-bottom': '12px',
        }"
      >
        <template #default>
          云函数使用 Goja 引擎执行 JavaScript，支持注入
          <code>db / http / crypto / app / user</code>
          内置对象。完整 API 见
          <Link
            @click="
              router.push({
                path: entryPath('/docs'),
                query: { section: 'cloud' },
              })
            "
            ><template #default>云函数文档</template></Link
          >
        </template>
      </Alert>
      <Table
        :data="rows"
        :columns="tableColumns"
        :loading="loading"
        :pagination="pagination"
        row-key="id"
      >
        <template #trigger="{ record: e }">
          <Tag>
            <template #default>{{ x(e.trigger) }}</template>
          </Tag>
        </template>
        <template #enabled="{ record: e }">
          <Switch
            :model-value="!!e.enabled"
            @change="(u) => handleSaveFunction(e, u)"
            size="small"
          ></Switch>
        </template>
        <template #lastRun="{ record: e }">
          <template v-if="e.lastRun">
            <span :key="0">{{ formatDateTime(e.lastRun) }}</span>
          </template>
          <template v-else>
            <span
              :key="1"
              :style="{
                color: 'var(--color-text-3)',
              }"
            >
              从未执行
            </span>
          </template>
        </template>
        <template #operations="{ record: e }">
          <Space>
            <template #default>
              <Link @click="(u) => I(e)"
                ><template #default>编辑</template></Link
              >
              <Link @click="(u) => handleRunFunction(e)"
                ><template #default>试运行</template></Link
              >
              <Popconfirm
                content="确定删除？"
                @ok="(u) => handleDeleteFunction(e)"
              >
                <template #default>
                  <Link status="danger"
                    ><template #default>删除</template></Link
                  >
                </template>
              </Popconfirm>
            </template>
          </Space>
        </template>
      </Table>
    </div>
  </div>
</template>
