<!-- Reconstructed editable Vue source; see README for recovery limits. -->
<script>
import { keepModalOpen } from "@/utils/modal.js";
import { listAdmins, deleteAdmin, saveAdmin } from "@/api/system.js";
import { formatDateTime } from "@/utils/format.js";
import { useUserStore } from "@/stores/user.js";
import { entryPath } from "@/utils/route.js";
import { onMounted, computed, ref, reactive } from "vue";
import {
  Breadcrumb,
  Button,
  Table,
  Modal,
  BreadcrumbItem,
  Space,
  Link,
  Popconfirm,
  Tag,
  Form,
  FormItem,
  Input,
  InputPassword,
  RadioGroup,
  Radio,
  Message,
} from "@arco-design/web-vue";
import { IconPlus } from "@arco-design/web-vue/es/icon";
export default {
  name: "AdminList",
  components: {
    Breadcrumb,
    BreadcrumbItem,
    Button,
    IconPlus,
    Table,
    Tag,
    Space,
    Link,
    Popconfirm,
    Modal,
    Form,
    FormItem,
    Input,
    InputPassword,
    RadioGroup,
    Radio,
  },
  setup(St) {
    const userStore = useUserStore();
    const y = computed(() => Number(userStore.profile?.id || 0));
    const r = computed(() => userStore.profile?.role === "super");
    const rows = ref([]);
    const loading = ref(false);
    const visible = ref(false);
    const formData = reactive({
      id: 0,
      username: "",
      password: "",
      role: "admin",
    });
    const V = () => r.value;
    const P = (n) => r.value && n.id !== y.value;
    const R = computed(() => !r.value);
    const I = computed(() => formData.role === "super"
      ? "拥有全部管理权限；系统必须保留至少一位超级管理员。"
      : "可管理用户、卡密、设备会话、黑名单和公告，查看用户日志；不能修改应用或系统配置。");
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
        title: "账号",
        dataIndex: "username",
        width: 200,
        ellipsis: true,
        tooltip: true,
      },
      {
        title: "角色",
        slotName: "role",
        width: 130,
      },
      {
        title: "最近登录",
        slotName: "lastLogin",
        width: 200,
      },
      {
        title: "登录 IP",
        dataIndex: "lastIp",
        width: 160,
      },
      {
        title: "操作",
        slotName: "operations",
        fixed: "right",
        width: 140,
      },
    ];
    const handleListAdmins = async () => {
      loading.value = true;
      try {
        rows.value = await listAdmins();
      } finally {
        loading.value = false;
      }
    };
    const handleClick = () => {
      Object.assign(formData, {
        id: 0,
        username: "",
        password: "",
        role: "admin",
      });
      visible.value = true;
    };
    const M = (n) => {
      Object.assign(formData, {
        ...n,
        password: "",
      });
      visible.value = true;
    };
    const handleSaveAdmin = async () => {
      await saveAdmin(formData);
      Message.success("已保存");
      visible.value = false;
      handleListAdmins();
    };
    const handleDeleteAdmin = async (n) => {
      await deleteAdmin(n.id);
      Message.success("已删除");
      handleListAdmins();
    };
    onMounted(async () => {
      userStore.profile || (await userStore.fetchMe());
      handleListAdmins();
    });
    return {
      keepModalOpen,
      userStore,
      y,
      r,
      V,
      P,
      R,
      I,
      tableColumns,
      handleListAdmins,
      handleClick,
      M,
      handleSaveAdmin,
      handleDeleteAdmin,
      rows,
      loading,
      pagination,
      visible,
      formData,
      entryPath,
      formatDateTime,
    };
  },
};
</script>
<template>
  <div class="page">
    <Breadcrumb class="crumb">
      <template #default>
        <BreadcrumbItem>
          <template #default>
            <a @click="(a) => $router.push(entryPath('/app'))">应用列表</a>
          </template>
        </BreadcrumbItem>
        <BreadcrumbItem><template #default>多管理员</template></BreadcrumbItem>
      </template>
    </Breadcrumb>
    <div class="page-card">
      <div class="toolbar">
        <template v-if="r">
          <Button :key="0" type="primary" @click="handleClick">
            <template #icon><IconPlus></IconPlus></template>
            <template #default>新增管理员</template>
          </Button>
        </template>
        <template v-else></template>
        <div class="grow"></div>
      </div>
      <Table
        :data="rows"
        :columns="tableColumns"
        :loading="loading"
        :pagination="pagination"
        row-key="id"
      >
        <template #role="{ record: a }">
          <Tag :color="a.role === 'super' ? 'red' : 'blue'">
            {{ a.role === "super" ? "超级管理员" : "普通管理员" }}
          </Tag>
        </template>
        <template #lastLogin="{ record: a }">
          {{ a.lastLogin ? formatDateTime(a.lastLogin) : "-" }}
        </template>
        <template #operations="{ record: a }">
          <Space>
            <template #default>
              <Link :disabled="!V(a)" @click="X => V(a) &amp;&amp; M(a)">
                <template #default>编辑</template>
              </Link>
              <template v-if="P(a)">
                <Popconfirm
                  :key="0"
                  content="确定删除？"
                  @ok="(X) => handleDeleteAdmin(a)"
                >
                  <template #default>
                    <Link status="danger"
                      ><template #default>删除</template></Link
                    >
                  </template>
                </Popconfirm>
              </template>
              <template v-else>
                <Link :key="1" status="danger" disabled=""
                  ><template #default>删除</template></Link
                >
              </template>
            </template>
          </Space>
        </template>
      </Table>
    </div>
    <Modal
      v-model:visible="visible"
      :title="formData.id ? '编辑管理员' : '新增管理员'"
      :on-before-ok="() => keepModalOpen(handleSaveAdmin)"
      :width="480"
    >
      <template #default>
        <Form :model="formData" layout="vertical">
          <template #default>
            <FormItem label="账号">
              <template #default>
                <Input
                  v-model="formData.username"
                  :disabled="!!formData.id"
                ></Input>
              </template>
            </FormItem>
            <FormItem :label="formData.id ? '新密码（不改请留空）' : '密码'">
              <template #default>
                <InputPassword
                  v-model="formData.password"
                  placeholder="至少 6 位"
                ></InputPassword>
              </template>
            </FormItem>
            <FormItem label="角色">
              <template #default>
                <RadioGroup v-model="formData.role" :disabled="R">
                  <template #default>
                    <Radio value="super"
                      ><template #default>超管</template></Radio
                    >
                    <Radio value="admin"
                      ><template #default>普通</template></Radio
                    >
                  </template>
                </RadioGroup>
                <template v-if="I">
                  <div :key="0" class="role-tip">{{ I }}</div>
                </template>
                <template v-else></template>
              </template>
            </FormItem>
          </template>
        </Form>
      </template>
    </Modal>
  </div>
</template>
<style scoped>
.page {
  padding: 16px 24px 24px;
}
.crumb {
  margin-bottom: 12px;
}
.role-tip {
  margin-top: 4px;
  font-size: 12px;
  color: var(--color-text-3);
}
</style>
