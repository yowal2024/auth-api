<!-- Reconstructed editable Vue source; see README for recovery limits. -->
<script>
import { keepModalOpen } from "@/utils/modal.js";
import { useAppStore } from "@/stores/app.js";
import { listExtends, deleteExtend, saveExtend } from "@/api/app.js";
import { formatDateTime } from "@/utils/format.js";
import { watch, onMounted, reactive, computed, ref } from "vue";
import {
  Button,
  Input,
  Table,
  Modal,
  Space,
  Link,
  Popconfirm,
  Tooltip,
  Form,
  FormItem,
  Textarea,
  Message,
} from "@arco-design/web-vue";
import { IconPlus, IconSearch } from "@arco-design/web-vue/es/icon";
export default {
  name: "Extend",
  components: {
    Button,
    IconPlus,
    Input,
    IconSearch,
    Table,
    Tooltip,
    Space,
    Link,
    Popconfirm,
    Modal,
    Form,
    FormItem,
    Textarea,
  },
  setup(ve) {
    const appStore = useAppStore();
    const m = reactive({
      keyword: "",
    });
    const rows = ref([]);
    const loading = ref(false);
    const visible = ref(false);
    const formData = reactive({
      id: 0,
      key: "",
      value: "",
      remark: "",
    });
    const pagination = computed(() => ({
      total: rows.value.length,
      showTotal: true,
      showPageSize: true,
      pageSizeOptions: [10, 20, 50, 100],
    }));
    const tableColumns = [
      {
        title: "ID",
        dataIndex: "id",
        width: 70,
      },
      {
        title: "扩展名",
        dataIndex: "remark",
        width: 180,
        ellipsis: true,
        tooltip: true,
      },
      {
        title: "扩展键",
        slotName: "key_",
        width: 180,
        ellipsis: true,
        tooltip: true,
      },
      {
        title: "扩展值",
        slotName: "value",
        width: 280,
        ellipsis: true,
      },
      {
        title: "更新时间",
        slotName: "updateTime",
        width: 200,
      },
      {
        title: "操作",
        slotName: "operations",
        fixed: "right",
        width: 140,
      },
    ];
    const handleListExtends = async () => {
      loading.value = true;
      try {
        rows.value = await listExtends({
          appid: appStore.currentId,
          ...m,
        });
      } finally {
        loading.value = false;
      }
    };
    const handleClick = () => {
      Object.assign(formData, {
        id: 0,
        key: "",
        value: "",
        remark: "",
      });
      visible.value = true;
    };
    const T = (s) => {
      Object.assign(
        formData,
        {
          id: 0,
          key: "",
          value: "",
          remark: "",
        },
        s,
      );
      visible.value = true;
    };
    const handleSaveExtend = async () => {
      await saveExtend({
        ...formData,
        appid: appStore.currentId,
      });
      Message.success("已保存");
      visible.value = false;
      handleListExtends();
    };
    const handleDeleteExtend = async (s) => {
      await deleteExtend(s.id);
      Message.success("已删除");
      handleListExtends();
    };
    const C = async (s) => {
      try {
        await navigator.clipboard.writeText(s);
        Message.success("已复制");
      } catch {
        Message.error("复制失败");
      }
    };
    watch(() => appStore.currentId, handleListExtends);
    onMounted(handleListExtends);
    return {
      keepModalOpen,
      appStore,
      m,
      tableColumns,
      handleListExtends,
      handleClick,
      T,
      handleSaveExtend,
      handleDeleteExtend,
      C,
      rows,
      loading,
      pagination,
      visible,
      formData,
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
          <template #default>添加扩展</template>
        </Button>
        <div class="grow"></div>
        <Input
          v-model="m.keyword"
          placeholder="关键词搜索..."
          allow-clear=""
          :style="{
            width: '260px',
          }"
          @pressEnter="handleListExtends"
        >
          <template #suffix>
            <IconSearch
              :style="{
                cursor: 'pointer',
              }"
              @click="handleListExtends"
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
        <template #key_="{ record: e }">{{ e.key }}</template>
        <template #value="{ record: e }">
          <Tooltip :content="e.value">
            <template #default>
              <span class="copy-cell" @click="(b) => C(e.value)">{{
                e.value
              }}</span>
            </template>
          </Tooltip>
        </template>
        <template #updateTime="{ record: e }">{{
          formatDateTime(e.updateTime)
        }}</template>
        <template #operations="{ record: e }">
          <Space>
            <template #default>
              <Link @click="(b) => T(e)"
                ><template #default>编辑</template></Link
              >
              <Popconfirm
                content="确定删除？"
                @ok="(b) => handleDeleteExtend(e)"
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
      :title="formData.id ? '编辑扩展' : '添加扩展'"
      ok-text="提交"
      cancel-text="取消"
      :on-before-ok="() => keepModalOpen(handleSaveExtend)"
      :width="400"
    >
      <template #default>
        <Form :model="formData" layout="vertical">
          <template #default>
            <FormItem label="扩展名">
              <template #default>
                <Input
                  v-model="formData.remark"
                  placeholder="如: 客服QQ"
                ></Input>
              </template>
            </FormItem>
            <FormItem label="扩展键">
              <template #default>
                <Input v-model="formData.key" placeholder="如: qq"></Input>
              </template>
            </FormItem>
            <FormItem label="扩展值">
              <template #default>
                <Textarea
                  v-model="formData.value"
                  placeholder="如: 51154393"
                ></Textarea>
              </template>
            </FormItem>
          </template>
        </Form>
      </template>
    </Modal>
  </div>
</template>
<style scoped>
.copy-cell {
  cursor: pointer;
  color: var(--color-text-1);
}
.copy-cell:hover {
  color: rgb(var(--primary-6));
}
</style>
