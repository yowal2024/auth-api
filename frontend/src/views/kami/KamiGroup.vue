<!-- Reconstructed editable Vue source; see README for recovery limits. -->
<script>
import { keepModalOpen } from "@/utils/modal.js";
import { listKamiGroups, deleteKamiGroup, saveKamiGroup } from "@/api/kami.js";
import { useAppStore } from "@/stores/app.js";

import { formatDuration } from "@/utils/format.js";
import { watch, onMounted, reactive, computed, ref } from "vue";
import {
  Button,
  Input,
  Table,
  Modal,
  Space,
  Link,
  Popconfirm,
  Tag,
  Form,
  FormItem,
  InputNumber,
  Select,
  Option,
  Message,
} from "@arco-design/web-vue";
import { IconPlus, IconSearch } from "@arco-design/web-vue/es/icon";
const k = 9999999999;
export default {
  name: "KamiGroup",
  components: {
    Button,
    IconPlus,
    Input,
    IconSearch,
    Table,
    Tag,
    Space,
    Link,
    Popconfirm,
    Modal,
    Form,
    FormItem,
    InputNumber,
    Select,
    Option,
  },
  setup(It) {
    const S = {
      second: 1,
      minute: 60,
      hour: 3600,
      day: 86400,
    };
    const C = (a) =>
      !a || a >= k
        ? {
            amount: 1,
            unit: "forever",
          }
        : a % 86400 === 0
          ? {
              amount: a / 86400,
              unit: "day",
            }
          : a % 3600 === 0
            ? {
                amount: a / 3600,
                unit: "hour",
              }
            : a % 60 === 0
              ? {
                  amount: a / 60,
                  unit: "minute",
                }
              : {
                  amount: a,
                  unit: "second",
                };
    const appStore = useAppStore();
    const p = ref([]);
    const loading = ref(false);
    const visible = ref(false);
    const formData = reactive({
      id: 0,
      name: "",
      type: "vip",
      amount: 1,
      unit: "day",
      remark: "",
    });
    const d = reactive({
      keyword: "",
    });
    const O = computed(() =>
      formData.unit === "forever"
        ? k
        : Math.max(1, Math.floor(formData.amount || 0)) *
          (S[formData.unit] || 1),
    );
    const rows = computed(() => {
      if (!d.keyword) return p.value;
      const a = d.keyword.toLowerCase();
      return p.value.filter(
        (t) =>
          (t.name || "").toLowerCase().includes(a) ||
          (t.remark || "").toLowerCase().includes(a),
      );
    });
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
        title: "分组名称",
        dataIndex: "name",
        width: 200,
        ellipsis: true,
        tooltip: true,
      },
      {
        title: "类型",
        slotName: "type",
        width: 120,
      },
      {
        title: "面值",
        slotName: "value",
        width: 180,
      },
      {
        title: "备注",
        dataIndex: "remark",
        width: 240,
        ellipsis: true,
        tooltip: true,
      },
      {
        title: "操作",
        slotName: "operations",
        fixed: "right",
        width: 140,
      },
    ];
    const handleListKamiGroups = async () => {
      loading.value = true;
      try {
        p.value = await listKamiGroups();
      } finally {
        loading.value = false;
      }
    };
    const handleClick = () => {
      Object.assign(formData, {
        id: 0,
        name: "",
        type: "vip",
        amount: 1,
        unit: "day",
        remark: "",
      });
      visible.value = true;
    };
    const M = (a) => {
      const { amount: t, unit: f } = C(a.value);
      Object.assign(formData, {
        id: a.id,
        name: a.name,
        type: a.type || "vip",
        amount: t,
        unit: f,
        remark: a.remark || "",
      });
      visible.value = true;
    };
    const handleSaveKamiGroup = async () => {
      await saveKamiGroup({
        id: formData.id,
        name: formData.name,
        type: formData.type,
        value: O.value,
        remark: formData.remark,
        appid: appStore.currentId,
      });
      Message.success("已保存");
      visible.value = false;
      handleListKamiGroups();
    };
    const handleDeleteKamiGroup = async (a) => {
      await deleteKamiGroup(a.id);
      Message.success("已删除");
      handleListKamiGroups();
    };
    watch(() => appStore.currentId, handleListKamiGroups);
    onMounted(handleListKamiGroups);
    return {
      keepModalOpen,
      S,
      C,
      appStore,
      p,
      d,
      O,
      tableColumns,
      handleListKamiGroups,
      handleClick,
      M,
      handleSaveKamiGroup,
      handleDeleteKamiGroup,
      rows,
      loading,
      pagination,
      visible,
      formData,
      k,
      formatDuration,
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
          <template #default>新增分组</template>
        </Button>
        <div class="grow"></div>
        <Input
          v-model="d.keyword"
          placeholder="名称 / 备注搜索"
          allow-clear=""
          :style="{
            width: '220px',
          }"
        >
          <template #suffix>
            <IconSearch
              :style="{
                cursor: 'pointer',
              }"
            ></IconSearch>
          </template>
        </Input>
      </div>
      <Table
        :data="rows"
        :columns="tableColumns"
        :loading="loading"
        :pagination="pagination"
        row-key="id"
      >
        <template #type>
          <Tag color="gold"><template #default>VIP 时长</template></Tag>
        </template>
        <template #value="{ record: l }">
          <span>{{ l.value >= k ? "永久" : formatDuration(l.value) }}</span>
        </template>
        <template #operations="{ record: l }">
          <Space>
            <template #default>
              <Link @click="($) => M(l)"
                ><template #default>编辑</template></Link
              >
              <Popconfirm
                content="删除分组会同时删除组内全部卡密及关联记录，已登录客户端也会失效，确认？"
                @ok="($) => handleDeleteKamiGroup(l)"
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
    <Modal
      v-model:visible="visible"
      :title="formData.id ? '编辑分组' : '新增分组'"
      :on-before-ok="() => keepModalOpen(handleSaveKamiGroup)"
      width="auto"
    >
      <template #default>
        <div class="md-w-80">
          <Form :model="formData" size="small" auto-label-width="">
            <template #default>
              <FormItem label="分组名称">
                <template #default
                  ><Input v-model="formData.name"></Input
                ></template>
              </FormItem>
              <FormItem label="时长">
                <template #default>
                  <InputNumber
                    v-model="formData.amount"
                    :min="1"
                    :disabled="formData.unit === 'forever'"
                  >
                    <template #append>
                      <Select
                        v-model="formData.unit"
                        :style="{
                          width: '70px',
                        }"
                        :bordered="!1"
                      >
                        <template #default>
                          <Option value="second"
                            ><template #default>秒</template></Option
                          >
                          <Option value="minute"
                            ><template #default>分</template></Option
                          >
                          <Option value="hour"
                            ><template #default>时</template></Option
                          >
                          <Option value="day"
                            ><template #default>天</template></Option
                          >
                          <Option value="forever"
                            ><template #default>永久</template></Option
                          >
                        </template>
                      </Select>
                    </template>
                  </InputNumber>
                </template>
              </FormItem>
              <FormItem label="备注">
                <template #default
                  ><Input v-model="formData.remark"></Input
                ></template>
              </FormItem>
            </template>
          </Form>
        </div>
      </template>
    </Modal>
  </div>
</template>
<style scoped>
.md-w-80 {
  width: 20rem;
}
</style>
