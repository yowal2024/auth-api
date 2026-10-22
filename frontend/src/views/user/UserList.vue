<!-- Reconstructed editable Vue source; see README for recovery limits. -->
<script>
import { keepModalOpen } from "@/utils/modal.js";
import { usePagedList } from "@/composables/usePagedList.js";
import { reportError } from "@/utils/errors.js";
import { watch, onMounted, ref, reactive, computed } from "vue";
import {
  Button,
  Popconfirm,
  Select,
  InputGroup,
  Table,
  Modal,
  Option,
  Input,
  Space,
  Link,
  Tag,
  Avatar,
  Form,
  FormItem,
  InputPassword,
  Tooltip,
  InputNumber,
  Message,
} from "@arco-design/web-vue";
import { useRouter } from "vue-router";
import {
  IconPlus,
  IconGift,
  IconDownload,
  IconDelete,
  IconSearch,
  IconUser,
  IconRefresh,
  IconLock,
} from "@arco-design/web-vue/es/icon";
import {
  listUserGroups,
  listUsers,
  batchDeleteUsers,
  deleteUser,
  createUser,
  exportUsers,
  rewardUsers,
} from "@/api/user.js";
import { useAppStore } from "@/stores/app.js";

import { formatDateTime, formatVIPExpiry } from "@/utils/format.js";
import { exportA } from "@/utils/avatar.js";
import { entryPath } from "@/utils/route.js";
export default {
  name: "UserList",
  components: {
    Button,
    IconPlus,
    IconGift,
    IconDownload,
    Popconfirm,
    IconDelete,
    Select,
    Option,
    InputGroup,
    Input,
    IconSearch,
    Table,
    Avatar,
    Tag,
    Space,
    Link,
    Modal,
    Form,
    FormItem,
    IconUser,
    InputPassword,
    IconLock,
    Tooltip,
    IconRefresh,
    InputNumber,
  },
  setup(Pt) {
    const router = useRouter();
    const u = reactive({
      keyword: "",
      status: void 0,
      userType: void 0,
      groupId: void 0,
      searchField: "username",
      page: 1,
      size: 20,
    });
    const H = computed(() =>
      u.searchField === "fuzzy"
        ? "账号 / 昵称 / ID 模糊匹配"
        : `请输入用户${
            {
              username: "账号",
              nickname: "昵称",
              id: "ID",
            }[u.searchField] || "账号"
          }`,
    );
    const rows = ref([]);
    const q = ref(0);
    const loading = ref(false);
    const w = ref([]);
    const Y = Math.floor(Date.now() / 1e3);
    const L = ref(false);
    const rewarding = ref(false);
    const g = reactive({
      value: 1,
      unit: "day",
      target: "all",
    });
    const z = ref(false);
    const P = ref(false);
    const i = reactive({
      scope: "all",
      groupId: void 0,
      format: "csv",
    });
    const handleClick3 = () => {
      Object.assign(i, {
        scope: "all",
        groupId: void 0,
        format: "csv",
      });
      z.value = true;
    };
    const handleExportUsers = async () => {
      if (P.value) return;
      if (i.scope === "group" && !i.groupId) {
        Message.warning("请选择分组");
        return;
      }
      P.value = true;
      try {
        const s = {
          scope: i.scope,
          format: i.format,
        };
        i.scope === "group" && (s.groupId = i.groupId);
        const e = await exportUsers(s);
        const c = `users_${Date.now()}.${i.format}`;
        const b = URL.createObjectURL(e);
        const V = document.createElement("a");
        V.href = b;
        V.download = c;
        document.body.appendChild(V);
        V.click();
        document.body.removeChild(V);
        URL.revokeObjectURL(b);
        Message.success("已导出");
        z.value = false;
      } catch (s) {
        reportError(s, "导出失败");
      } finally {
        P.value = false;
      }
    };
    const visible = ref(false);
    const loadingState = ref(false);
    const createFormRef = ref();
    const formData = reactive({
      username: "",
      password: "",
      groupId: void 0,
    });
    const appStore = useAppStore();
    const I = ref([]);
    const handleListUserGroups = async () => {
      try {
        I.value = (await listUserGroups()) || [];
      } catch {
        I.value = [];
      }
    };
    const handleClick = () => {
      Object.assign(formData, {
        username: "",
        password: "",
        groupId: I.value[0]?.id,
      });
      visible.value = true;
    };
    const handleClick4 = () => {
      const s = "abcdefghjkmnpqrstuvwxyzABCDEFGHJKMNPQRSTUVWXYZ23456789";
      let C = "";
      const c = new Uint32Array(12);
      crypto.getRandomValues(c);
      for (let b = 0; b < 12; b++) C += s[c[b] % s.length];
      formData.password = C;
    };
    const handleCreateUser = async () => {
      if (loadingState.value) return;
      loadingState.value = true;
      try {
        if (!(await createFormRef.value?.validate())) {
          await createUser({
            ...formData,
          });
          Message.success("已添加");
          visible.value = false;
          handleListUsers();
        }
      } catch (error) {
        reportError(error);
      } finally {
        loadingState.value = false;
      }
    };
    const handleClick2 = () => {
      Object.assign(g, {
        value: 1,
        unit: "day",
        target: "all",
      });
      L.value = true;
    };
    const handleRewardUsers = async () => {
      if (rewarding.value) return;
      if (!Number.isInteger(g.value) || g.value < 1 || g.value > 1000000) {
        Message.warning("奖励数值需为 1 到 1000000 的整数");
        return;
      }
      rewarding.value = true;
      try {
        await rewardUsers({ ...g });
        Message.success("已发放");
        L.value = false;
        handleListUsers();
      } finally {
        rewarding.value = false;
      }
    };
    const tableColumns = [
      {
        title: "ID",
        dataIndex: "id",
        width: 70,
      },
      {
        title: "账号",
        slotName: "username",
        width: 220,
        ellipsis: true,
        tooltip: true,
      },
      {
        title: "VIP",
        slotName: "vip",
        width: 140,
      },
      {
        title: "状态",
        slotName: "status",
        width: 80,
      },
      {
        title: "分组",
        slotName: "group",
        width: 140,
      },
      {
        title: "在线状态",
        slotName: "online",
        width: 100,
      },
      {
        title: "最近活跃IP",
        slotName: "lastIp",
        width: 200,
        ellipsis: true,
        tooltip: true,
      },
      {
        title: "最后活跃时间",
        slotName: "lastActiveTime",
        width: 200,
      },
      {
        title: "注册时间",
        slotName: "createTime",
        width: 200,
      },
      {
        title: "操作",
        slotName: "operations",
        fixed: "right",
        width: 140,
      },
    ];
    const pagination = computed(() => ({
      current: u.page,
      pageSize: u.size,
      total: q.value,
      showTotal: true,
      showPageSize: true,
      pageSizeOptions: [10, 20, 50, 100],
    }));
    const handleListUsers = usePagedList(listUsers, u, rows, q, loading, w);
    const handlePageChange = (s) => {
      u.page = s;
      handleListUsers();
    };
    const handlePageSizeChange = (s) => {
      u.size = s;
      u.page = 1;
      handleListUsers();
    };
    const pe = (s) => router.push(entryPath(`/user/edit/${s.id}`));
    const handleDeleteUser = async (s) => {
      await deleteUser(s.id);
      Message.success("已删除");
      handleListUsers();
    };
    const handleOk = async () => {
      w.value.length &&
        (await batchDeleteUsers(w.value),
        Message.success(`已删除 ${w.value.length} 个用户`),
        (w.value = []),
        handleListUsers());
    };
    watch(
      () => appStore.currentId,
      () => {
        handleListUserGroups();
        handleListUsers();
      },
    );
    onMounted(() => {
      handleListUserGroups();
      handleListUsers();
    });
    return {
      keepModalOpen,
      rewarding,
      router,
      u,
      H,
      q,
      w,
      Y,
      L,
      g,
      z,
      P,
      i,
      handleClick3,
      handleExportUsers,
      loadingState,
      appStore,
      I,
      handleListUserGroups,
      handleClick,
      handleClick4,
      handleCreateUser,
      handleClick2,
      handleRewardUsers,
      tableColumns,
      handleListUsers,
      handlePageChange,
      handlePageSizeChange,
      pe,
      handleDeleteUser,
      handleOk,
      rows,
      pagination,
      loading,
      visible,
      createFormRef,
      formData,
      exportA,
      formatVIPExpiry,
      formatDateTime,
    };
  },
};
</script>
<template>
  <div class="page">
    <div class="page-card">
      <div class="topbar">
        <div class="toolbar">
          <Button type="primary" @click="handleClick">
            <template #icon><IconPlus></IconPlus></template>
            <template #default>添加用户</template>
          </Button>
          <Button @click="handleClick2">
            <template #icon><IconGift></IconGift></template>
            <template #default>发送奖励</template>
          </Button>
          <Button @click="handleClick3">
            <template #icon><IconDownload></IconDownload></template>
            <template #default>导出数据</template>
          </Button>
          <Popconfirm
            :content="`确定删除选中的 ${w.length} 个用户?`"
            @ok="handleOk"
          >
            <template #default>
              <Button status="danger" :disabled="!w.length">
                <template #icon><IconDelete></IconDelete></template>
                <template #default>批量删除</template>
              </Button>
            </template>
          </Popconfirm>
        </div>
        <div class="filters">
          <div class="filter-row filter-status">
            <Select
              v-model="u.status"
              placeholder="全部状态"
              allow-clear=""
              @change="handleListUsers"
            >
              <template #default>
                <Option :value="1"><template #default>正常</template></Option>
                <Option :value="0"><template #default>冻结</template></Option>
              </template>
            </Select>
          </div>
          <div class="filter-row filter-user">
            <Select
              v-model="u.userType"
              placeholder="全部用户"
              allow-clear=""
              @change="handleListUsers"
            >
              <template #default>
                <Option value="vip"
                  ><template #default>VIP 用户</template></Option
                >
                <Option value="free"
                  ><template #default>普通用户</template></Option
                >
              </template>
            </Select>
          </div>
          <div class="filter-row filter-group">
            <Select
              v-model="u.groupId"
              placeholder="全部分组"
              allow-clear=""
              @change="handleListUsers"
            >
              <template #default>
                <template v-for="a in I" :key="a.id">
                  <Option :value="a.id">
                    <template #default>{{
                      a.name + " (Lv." + a.level + ") "
                    }}</template>
                  </Option>
                </template>
              </template>
            </Select>
          </div>
          <div class="filter-row filter-search">
            <InputGroup class="search-group">
              <template #default>
                <Select
                  v-model="u.searchField"
                  :style="{
                    width: '110px',
                  }"
                >
                  <template #default>
                    <Option value="fuzzy"
                      ><template #default>模糊搜索</template></Option
                    >
                    <Option value="username"
                      ><template #default>账号</template></Option
                    >
                    <Option value="nickname"
                      ><template #default>昵称</template></Option
                    >
                    <Option value="id"><template #default>ID</template></Option>
                  </template>
                </Select>
                <Input
                  v-model="u.keyword"
                  :placeholder="H"
                  allow-clear=""
                  @pressEnter="handleListUsers"
                  @clear="handleListUsers"
                >
                  <template #suffix>
                    <IconSearch
                      :style="{
                        cursor: 'pointer',
                      }"
                      @click="handleListUsers"
                    ></IconSearch>
                  </template>
                </Input>
              </template>
            </InputGroup>
          </div>
        </div>
      </div>
      <Table
        :data="rows"
        :columns="tableColumns"
        :pagination="pagination"
        :loading="loading"
        :row-selection="{
          type: 'checkbox',
          showCheckedAll: !0,
        }"
        v-model:selectedKeys="w"
        row-key="id"
        @pageChange="handlePageChange"
        @pageSizeChange="handlePageSizeChange"
      >
        <template #username="{ record: a }">
          <div class="user-cell">
            <Avatar :size="36" class="user-cell-avatar">
              <template #default
                ><img :src="a.avatar || exportA" alt="avatar"
              /></template>
            </Avatar>
            <div class="user-cell-text">
              <div class="user-cell-name">{{ a.username }}</div>
              <div class="user-cell-sub">{{ a.nickname || "匿名用户" }}</div>
            </div>
          </div>
        </template>
        <template #vip="{ record: a }">
          <Tag :color="a.vip > Y ? 'green' : 'gray'">
            <template #default>{{ formatVIPExpiry(a.vip) }}</template>
          </Tag>
        </template>
        <template #status="{ record: a }">
          <Tag :color="a.status === 1 ? 'green' : 'red'">
            <template #default>{{ a.status === 1 ? "正常" : "冻结" }}</template>
          </Tag>
        </template>
        <template #group="{ record: a }">
          <template v-if="a.groupName">
            <Tag :key="0" color="arcoblue">
              <template #default>{{
                a.groupName + " Lv." + a.groupLevel
              }}</template>
            </Tag>
          </template>
          <template v-else
            ><span :key="1" class="user-cell-sub">-</span></template
          >
        </template>
        <template #online="{ record: a }">
          <Tag :color="a.online ? 'green' : 'gray'">
            <template #default>
              <span
                :class="[
                  'dot',
                  {
                    on: a.online,
                  },
                ]"
              ></span>
              {{ " " + (a.online ? "在线" : "离线") }}
            </template>
          </Tag>
        </template>
        <template #lastIp="{ record: a }">
          <div class="user-cell-text">
            <div class="user-cell-name">{{ a.lastRegion || "-" }}</div>
            <div class="user-cell-sub">{{ a.lastIp || "-" }}</div>
          </div>
        </template>
        <template #lastActiveTime="{ record: a }">
          {{ a.lastActiveTime ? formatDateTime(a.lastActiveTime) : "-" }}
        </template>
        <template #createTime="{ record: a }">{{
          formatDateTime(a.createTime)
        }}</template>
        <template #operations="{ record: a }">
          <Space>
            <template #default>
              <Link @click="(Ue) => pe(a)"
                ><template #default>编辑</template></Link
              >
              <Popconfirm
                content="确定删除？"
                @ok="(Ue) => handleDeleteUser(a)"
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
      title="添加用户"
      width="auto"
      modal-class="create-user-modal"
      :on-before-cancel="() => !loadingState"
      :closable="!loadingState"
      :mask-closable="!loadingState"
      :esc-to-close="!loadingState"
    >
      <template #footer>
        <Button
          type="primary"
          long=""
          :loading="loadingState"
          @click="handleCreateUser"
        >
          <template #default>提交</template>
        </Button>
      </template>
      <template #default>
        <Form
          ref="createFormRef"
          :model="formData"
          :disabled="loadingState"
          layout="vertical"
          hide-label=""
        >
          <template #default>
            <FormItem
              field="username"
              hide-asterisk=""
              :rules="[
                {
                  required: !0,
                  message: '请设置账号',
                },
                {
                  match: /^.{5,12}$/,
                  message: '账号长度为 5~12 位',
                },
              ]"
            >
              <template #default>
                <Input
                  v-model="formData.username"
                  placeholder="请设置账号5~12位"
                >
                  <template #prefix><IconUser></IconUser></template>
                </Input>
              </template>
            </FormItem>
            <FormItem
              field="password"
              hide-asterisk=""
              :rules="[
                {
                  required: !0,
                  message: '请输入密码',
                },
                {
                  match: /^.{6,18}$/,
                  message: '密码长度为 6~18 位',
                },
              ]"
            >
              <template #default>
                <InputPassword
                  v-model="formData.password"
                  placeholder="请输入密码6~18"
                >
                  <template #prefix><IconLock></IconLock></template>
                  <template #suffix>
                    <Tooltip content="随机密码">
                      <template #default>
                        <IconRefresh
                          class="rand-pwd-btn"
                          @click="handleClick4"
                        ></IconRefresh>
                      </template>
                    </Tooltip>
                  </template>
                </InputPassword>
              </template>
            </FormItem>
            <FormItem
              field="groupId"
              hide-asterisk=""
              :rules="[
                {
                  required: !0,
                  message: '请选择分组',
                },
              ]"
            >
              <template #default>
                <Select v-model="formData.groupId" placeholder="选择分组">
                  <template #default>
                    <template v-for="a in I" :key="a.id">
                      <Option :value="a.id">
                        <template #default>{{
                          a.name + " (Lv." + a.level + ") "
                        }}</template>
                      </Option>
                    </template>
                  </template>
                </Select>
              </template>
            </FormItem>
          </template>
        </Form>
      </template>
    </Modal>
    <Modal
      v-model:visible="z"
      title="导出账号数据"
      :width="360"
      ok-text="导出"
      :ok-loading="P"
      :on-before-ok="() => keepModalOpen(handleExportUsers)"
      :on-before-cancel="() => !P"
      :closable="!P"
      :mask-closable="!P"
      :esc-to-close="!P"
      :cancel-button-props="{ disabled: P }"
    >
      <template #default>
        <Form :model="i" auto-label-width="" :disabled="P">
          <template #default>
            <FormItem label="筛选范围">
              <template #default>
                <Select v-model="i.scope">
                  <template #default>
                    <Option value="all"
                      ><template #default>全部用户</template></Option
                    >
                    <Option value="group"
                      ><template #default>按分组</template></Option
                    >
                    <Option value="vip"
                      ><template #default>会员用户</template></Option
                    >
                    <Option value="banned"
                      ><template #default>冻结用户</template></Option
                    >
                  </template>
                </Select>
              </template>
            </FormItem>
            <template v-if="i.scope === 'group'">
              <FormItem :key="0" label="选择分组">
                <template #default>
                  <Select v-model="i.groupId" placeholder="请选择分组">
                    <template #default>
                      <template v-for="a in I" :key="a.id">
                        <Option :value="a.id">
                          <template #default>{{
                            a.name + " (Lv." + a.level + ") "
                          }}</template>
                        </Option>
                      </template>
                    </template>
                  </Select>
                </template>
              </FormItem>
            </template>
            <template v-else></template>
            <FormItem label="数据格式">
              <template #default>
                <Select v-model="i.format">
                  <template #default>
                    <Option value="csv"
                      ><template #default>CSV (Excel 可读)</template></Option
                    >
                    <Option value="txt"
                      ><template #default>TXT (Tab 分隔)</template></Option
                    >
                  </template>
                </Select>
              </template>
            </FormItem>
          </template>
        </Form>
      </template>
    </Modal>
    <Modal
      v-model:visible="L"
      title="发送奖励"
      :width="320"
      ok-text="提交"
      :on-before-ok="() => keepModalOpen(handleRewardUsers)"
      :ok-loading="rewarding"
      :on-before-cancel="() => !rewarding"
      :closable="!rewarding"
      :mask-closable="!rewarding"
      :esc-to-close="!rewarding"
      :cancel-button-props="{ disabled: rewarding }"
    >
      <template #default>
        <Form :model="g" size="small" auto-label-width="" :disabled="rewarding">
          <template #default>
            <FormItem label="会员值">
              <template #default>
                <InputNumber
                  v-model="g.value"
                  :min="1"
                  :max="1000000"
                  :precision="0"
                >
                  <template #append>
                    <Select
                      v-model="g.unit"
                      :style="{
                        width: '70px',
                      }"
                      :bordered="!1"
                    >
                      <template #default>
                        <Option value="day"
                          ><template #default>天</template></Option
                        >
                        <Option value="month"
                          ><template #default>月</template></Option
                        >
                        <Option value="year"
                          ><template #default>年</template></Option
                        >
                      </template>
                    </Select>
                  </template>
                </InputNumber>
              </template>
            </FormItem>
            <FormItem label="奖励对象">
              <template #default>
                <Select v-model="g.target">
                  <template #default>
                    <Option value="all"
                      ><template #default>所有人</template></Option
                    >
                    <Option value="vip"
                      ><template #default>仅会员</template></Option
                    >
                  </template>
                </Select>
              </template>
            </FormItem>
          </template>
        </Form>
      </template>
    </Modal>
  </div>
</template>
<style scoped>
.user-cell {
  display: flex;
  align-items: center;
  gap: 10px;
}
.user-cell .user-cell-avatar {
  background: transparent;
  flex-shrink: 0;
}
.user-cell .user-cell-avatar :deep(img) {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.user-cell .user-cell-text {
  min-width: 0;
  line-height: 1.4;
}
.user-cell .user-cell-name {
  font-size: 14px;
  font-weight: 500;
  color: var(--color-text-1);
}
.user-cell .user-cell-sub {
  font-size: 12px;
  color: var(--color-text-3);
}
.rand-pwd-btn {
  cursor: pointer;
  color: var(--color-text-3);
}
.rand-pwd-btn:hover {
  color: rgb(var(--primary-6));
}
.dot {
  display: inline-block;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--color-text-3);
  margin-right: 4px;
  vertical-align: middle;
}
.dot.on {
  background: rgb(var(--success-6));
}
.topbar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}
.topbar .toolbar {
  margin-bottom: 0;
  flex-shrink: 0;
}
.topbar .filters {
  flex: 1;
  min-width: 0;
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  margin-bottom: 0;
}
.topbar .filter-row {
  min-width: 0;
}
.topbar .filter-row :deep(.arco-select),
.topbar .filter-row :deep(.arco-input-group),
.topbar .filter-row :deep(.arco-input-wrapper) {
  width: 100%;
}
.topbar .filter-status,
.topbar .filter-user,
.topbar .filter-group {
  width: 180px;
  flex-shrink: 0;
}
.topbar .filter-search {
  flex: 1;
  min-width: 0;
  max-width: 420px;
}
.search-group {
  display: flex;
}
.search-group :deep(.arco-input-wrapper) {
  flex: 1;
}
@media (max-width: 768px) {
  .topbar {
    flex-direction: column;
    align-items: stretch;
  }
  .topbar .toolbar {
    flex-direction: row !important;
    align-items: center !important;
    flex-wrap: wrap;
  }
  .topbar .filters {
    flex-direction: column;
  }
  .topbar .filters .filter-status,
  .topbar .filters .filter-user,
  .topbar .filters .filter-group,
  .topbar .filters .filter-search {
    width: 100%;
    flex: none;
    min-width: 0;
    max-width: none;
  }
}
</style>
