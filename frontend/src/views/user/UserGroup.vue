<!-- Reconstructed editable Vue source; see README for recovery limits. -->
<script>
import { keepModalOpen } from "@/utils/modal.js";
import { listUserGroups, deleteUserGroup, saveUserGroup } from "@/api/user.js";
import { useAppStore } from "@/stores/app.js";

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
  Tag,
  Form,
  FormItem,
  InputNumber,
  Message,
} from "@arco-design/web-vue";
import { IconPlus, IconSearch } from "@arco-design/web-vue/es/icon";
export default {
  name: "UserGroup",
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
  },
  setup(ke) {
    const appStore = useAppStore();
    const p = ref([]);
    const loading = ref(false);
    const visible = ref(false);
    const formRef = ref();
    const handleAction = () => ({
      id: 0,
      name: "",
      level: 0,
      remark: "",
    });
    const formData = reactive(handleAction());
    const i = reactive({
      keyword: "",
    });
    const rows = computed(() => {
      if (!i.keyword) return p.value;
      const s = i.keyword.toLowerCase();
      return p.value.filter(
        (e) =>
          (e.name || "").toLowerCase().includes(s) ||
          (e.remark || "").toLowerCase().includes(s),
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
        title: "等级",
        slotName: "level",
        width: 120,
      },
      {
        title: "备注",
        dataIndex: "remark",
        width: 280,
        ellipsis: true,
        tooltip: true,
      },
      {
        title: "创建时间",
        slotName: "createTime",
        width: 180,
      },
      {
        title: "操作",
        slotName: "operations",
        fixed: "right",
        width: 140,
      },
    ];
    const handleListUserGroups = async () => {
      loading.value = true;
      try {
        p.value = await listUserGroups();
      } finally {
        loading.value = false;
      }
    };
    const handleClick = () => {
      Object.assign(formData, handleAction());
      visible.value = true;
    };
    const T = (s) => {
      Object.assign(formData, handleAction(), {
        id: s.id,
        name: s.name,
        level: s.level,
        remark: s.remark || "",
      });
      visible.value = true;
    };
    const handleSaveUserGroup = async () => {
      (await formRef.value?.validate()) ||
        (await saveUserGroup({
          id: formData.id,
          name: formData.name,
          level: formData.level,
          remark: formData.remark,
        }),
        Message.success("已保存"),
        (visible.value = false),
        handleListUserGroups());
    };
    const handleDeleteUserGroup = async (s) => {
      await deleteUserGroup(s.id);
      Message.success("已删除");
      handleListUserGroups();
    };
    watch(() => appStore.currentId, handleListUserGroups);
    onMounted(handleListUserGroups);
    return {
      keepModalOpen,
      appStore,
      p,
      handleAction,
      i,
      tableColumns,
      handleListUserGroups,
      handleClick,
      T,
      handleSaveUserGroup,
      handleDeleteUserGroup,
      rows,
      loading,
      pagination,
      visible,
      formRef,
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
          <template #default>新增分组</template>
        </Button>
        <div class="grow"></div>
        <Input
          v-model="i.keyword"
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
        <template #level="{ record: o }">
          <Tag color="arcoblue">
            <template #default>{{ "Lv." + o.level }}</template>
          </Tag>
        </template>
        <template #createTime="{ record: o }">{{
          formatDateTime(o.createTime)
        }}</template>
        <template #operations="{ record: o }">
          <Space>
            <template #default>
              <Link @click="(M) => T(o)"
                ><template #default>编辑</template></Link
              >
              <Popconfirm
                content="该分组下若有用户将无法删除,确认？"
                @ok="(M) => handleDeleteUserGroup(o)"
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
      :on-before-ok="() => keepModalOpen(handleSaveUserGroup)"
      width="auto"
    >
      <template #default>
        <div class="md-w-80">
          <Form
            ref="formRef"
            :model="formData"
            size="small"
            auto-label-width=""
          >
            <template #default>
              <FormItem
                field="name"
                label="分组名称"
                :rules="[
                  {
                    required: !0,
                    message: '请填写分组名称',
                  },
                ]"
              >
                <template #default>
                  <Input
                    v-model="formData.name"
                    placeholder="如: 黄金会员"
                  ></Input>
                </template>
              </FormItem>
              <FormItem label="等级">
                <template #default>
                  <InputNumber
                    v-model="formData.level"
                    :min="0"
                    :precision="0"
                    placeholder="数字,越大权限越高"
                  ></InputNumber>
                </template>
              </FormItem>
              <FormItem label="备注">
                <template #default>
                  <Input v-model="formData.remark" placeholder="可选"></Input>
                </template>
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
