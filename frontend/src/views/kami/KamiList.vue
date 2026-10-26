<!-- Reconstructed editable Vue source; see README for recovery limits. -->
<script>
import AccountSessions from "@/components/AccountSessions.vue";
import { usePagedList } from "@/composables/usePagedList.js";
import { reportError } from "@/utils/errors.js";
import { keepModalOpen } from "@/utils/modal.js";
import {
  listKami,
  listKamiGroups,
  cleanUsedKami,
  deleteKami,
  updateKami,
  generateKami,
} from "@/api/kami.js";
import { useAppStore } from "@/stores/app.js";

import { formatDateTime, formatDuration } from "@/utils/format.js";
import { watch, onMounted, reactive, ref, computed } from "vue";
import {
  Button,
  Popconfirm,
  Select,
  Input,
  Table,
  Drawer,
  Modal,
  Option,
  Link,
  Tag,
  Tooltip,
  Tabs,
  TabPane,
  Form,
  FormItem,
  InputNumber,
  Message,
} from "@arco-design/web-vue";
import {
  IconPlus,
  IconDelete,
  IconSearch,
  IconQuestionCircle,
} from "@arco-design/web-vue/es/icon";
export default {
  name: "KamiList",
  components: {
AccountSessions,
    Button,
    IconPlus,
    Popconfirm,
    IconDelete,
    Select,
    Option,
    Input,
    IconSearch,
    Table,
    Tooltip,
    Tag,
    Link,
    Drawer,
    Tabs,
    TabPane,
    Form,
    FormItem,
    IconQuestionCircle,
    InputNumber,
    Modal,
  },
  setup(mt) {
    const appStore = useAppStore();
    const S = computed(() => appStore.current?.type === "kami");
    const Y = computed(() => S.value && !!appStore.current?.bindDevice);
    const Z = computed(() => S.value && !!appStore.current?.bindRegion);
    const ee = computed(() => appStore.current?.ipDistanceMax || 50);
    const m = reactive({
      keyword: "",
      groupId: void 0,
      status: void 0,
      page: 1,
      size: 20,
    });
    const rows = ref([]);
    const A = ref(0);
    const loading = ref(false);
    const F = ref([]);
    const x = ref(false);
    const P = ref(false);
    const i = reactive({
      groupId: null,
      count: 10,
      prefix: "AUTH-",
      length: 16,
      remark: "",
      exportFmt: "none",
    });
    const visible = ref(false);
    const pageState = ref("basic");
    const loadingState = ref(false);
    const formData = reactive({
      id: 0,
      card: "",
      status: 0,
      remark: "",
      groupName: "",
      creator: "",
      createTime: 0,
      createIp: "",
      bindLimit: 0,
      regionLimit: 0,
    });
    const te = (o) => {
      formData.id = o.id;
      formData.card = o.card;
      formData.status = o.status;
      formData.remark = o.remark || "";
      formData.groupName = o.groupName || "";
      formData.creator = o.creator || "";
      formData.createTime = o.createTime || 0;
      formData.createIp = o.createIp || "";
      formData.bindLimit = o.bindLimit || 0;
      formData.regionLimit = o.regionLimit || 0;
      pageState.value = "basic";
      visible.value = true;
    };
    const handleUpdateKami = async () => {
      if (loadingState.value) return false;
      loadingState.value = true;
      try {
        await updateKami(formData.id, {
          remark: formData.remark,
          bindLimit: formData.bindLimit || 0,
          regionLimit: formData.regionLimit || 0,
        });
        Message.success("已保存");
        visible.value = false;
        handleListKami();
      } catch (error) {
        reportError(error);
      } finally {
        loadingState.value = false;
      }
      return false;
    };
    const freezing = ref(false);
    const toggleFrozen = async (row) => {
      if (freezing.value) return;
      freezing.value = true;
      try {
        await updateKami(row.id, { frozen: row.status !== 2 });
        Message.success(row.status === 2 ? "已解冻" : "已冻结");
        await handleListKami();
      } catch (error) {
        reportError(error);
      } finally {
        freezing.value = false;
      }
    };
    const tableColumns = [
      {
        title: "应用内编号",
        dataIndex: "appRecordId",
        width: 110,
      },
      {
        title: "卡密",
        slotName: "card",
        width: 240,
        ellipsis: true,
        tooltip: true,
      },
      {
        title: "分组",
        dataIndex: "groupName",
        width: 160,
        ellipsis: true,
        tooltip: true,
      },
      {
        title: "状态",
        slotName: "status",
        width: 100,
      },
      {
        title: "备注",
        slotName: "remark",
        width: 160,
        ellipsis: true,
        tooltip: true,
      },
      {
        title: "使用用户",
        slotName: "useUser",
        width: 160,
        ellipsis: true,
        tooltip: true,
      },
      {
        title: "生成时间",
        slotName: "createTime",
        width: 200,
      },
      {
        title: "使用时间",
        slotName: "useTime",
        width: 200,
      },
      {
        title: "最后活跃时间",
        slotName: "lastActiveTime",
        width: 180,
      },
      {
        title: "操作",
        slotName: "operations",
        fixed: "right",
        width: 180,
      },
    ];
    const pagination = computed(() => ({
      current: m.page,
      pageSize: m.size,
      total: A.value,
      showTotal: true,
      showPageSize: true,
      pageSizeOptions: [20, 50, 100, 200],
    }));
    const handleListKami = usePagedList(listKami, m, rows, A, loading);
    const handlePageChange = (o) => {
      m.page = o;
      handleListKami();
    };
    const handlePageSizeChange = (o) => {
      m.size = o;
      m.page = 1;
      handleListKami();
    };
    const handleListKamiGroups = async () => {
      try {
        F.value = await listKamiGroups();
      } catch (error) {
        reportError(error);
      }
    };
    const handleClick = () => {
      Object.assign(i, {
        groupId: null,
        count: 10,
        prefix: "AUTH-",
        length: 16,
        remark: "",
        exportFmt: "none",
      });
      x.value = true;
    };
    const re = (o, e) => {
      if (!o?.length) return;
      const b = Date.now();
      if (e === "csv") {
        const u = `card\r
`;
        const B = o.map((c) => `"${String(c).replace(/"/g, '""')}"`).join(`\r
`);
        const U = new Blob([u + B], {
          type: "text/csv;charset=utf-8",
        });
        M(U, `kamis_${b}.csv`);
      } else if (e === "txt") {
        const u = new Blob(
          [
            o.join(`\r
`),
          ],
          {
            type: "text/plain;charset=utf-8",
          },
        );
        M(u, `kamis_${b}.txt`);
      }
    };
    const M = (o, e) => {
      const b = URL.createObjectURL(o);
      const u = document.createElement("a");
      u.href = b;
      u.download = e;
      document.body.appendChild(u);
      u.click();
      u.remove();
      URL.revokeObjectURL(b);
    };
    const handleGenerateKami = async () => {
      if (P.value) return false;
      if (!i.groupId) return (Message.warning("请选择分组"), false);
      P.value = true;
      try {
        const o = await generateKami({
          groupId: i.groupId,
          count: i.count,
          prefix: i.prefix,
          length: i.length,
          remark: i.remark,
          appid: appStore.currentId,
        });
        Message.success(`已生成 ${o?.count ?? i.count} 个卡密`);
        i.exportFmt !== "none" && o?.cards?.length && re(o.cards, i.exportFmt);
        x.value = false;
        handleListKami();
      } finally {
        P.value = false;
      }
    };
    const handleDeleteKami = async (o) => {
      await deleteKami(o.id);
      Message.success("已删除");
      handleListKami();
    };
    const handleOk = async () => {
      const e = (await cleanUsedKami())?.count ?? 0;
      e > 0
        ? Message.success(`已清理 ${e} 条已使用卡密`)
        : Message.info("没有可清理的已使用卡密");
      handleListKami();
    };
    const pe = async (o) => {
      try {
        await navigator.clipboard.writeText(o);
        Message.success("已复制卡密");
      } catch {
        Message.error("复制失败");
      }
    };
    watch(
      () => appStore.currentId,
      () => {
        handleListKami();
        handleListKamiGroups();
      },
    );
    onMounted(() => {
      handleListKami();
      handleListKamiGroups();
    });
    return {
      keepModalOpen,
      appStore,
      S,
      Y,
      Z,
      ee,
      m,
      A,
      F,
      x,
      P,
      i,
      pageState,
      loadingState,
      te,
      handleUpdateKami,
      toggleFrozen,
      freezing,
      tableColumns,
      handleListKami,
      handlePageChange,
      handlePageSizeChange,
      handleListKamiGroups,
      handleClick,
      re,
      M,
      handleGenerateKami,
      handleDeleteKami,
      handleOk,
      pe,
      rows,
      pagination,
      loading,
      visible,
      formData,
      formatDateTime,
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
          <template #default>创建卡密</template>
        </Button>
        <Popconfirm
          content="确定清理所有已使用的卡密？此操作不可恢复"
          @ok="handleOk"
        >
          <template #default>
            <Button status="warning">
              <template #icon><IconDelete></IconDelete></template>
              <template #default>清理卡密</template>
            </Button>
          </template>
        </Popconfirm>
        <Select
          v-model="m.groupId"
          placeholder="分组"
          allow-clear=""
          :style="{
            width: '180px',
          }"
          @change="handleListKami"
        >
          <template #default>
            <template v-for="t in F" :key="t.id">
              <Option :value="t.id">
                <template #default>{{ t.name }}</template>
              </Option>
            </template>
          </template>
        </Select>
        <Select
          v-model="m.status"
          placeholder="状态"
          allow-clear=""
          :style="{
            width: '120px',
          }"
          @change="handleListKami"
        >
          <template #default>
            <Option :value="0"><template #default>未使用</template></Option>
            <Option :value="1"><template #default>已使用</template></Option>
            <Option :value="2"><template #default>已冻结</template></Option>
          </template>
        </Select>
        <Input
          v-model="m.keyword"
          placeholder="按卡号搜索"
          allow-clear=""
          :style="{
            width: '200px',
          }"
          @pressEnter="handleListKami"
        >
          <template #suffix>
            <IconSearch
              :style="{
                cursor: 'pointer',
              }"
              @click="handleListKami"
            ></IconSearch>
          </template>
        </Input>
      </div>
      <Table
        :data="rows"
        :columns="tableColumns"
        :pagination="pagination"
        :loading="loading"
        row-key="id"
        @pageChange="handlePageChange"
        @pageSizeChange="handlePageSizeChange"
      >
        <template #card="{ record: t }">
          <Tooltip :content="t.card">
            <template #default>
              <span class="card-cell" @click="(Q) => pe(t.card)">{{
                t.card
              }}</span>
            </template>
          </Tooltip>
        </template>
        <template #status="{ record: t }">
          <Tag :color="t.status === 2 ? 'red' : t.status === 1 ? 'gray' : 'green'">
            <template #default>{{
              t.status === 2 ? "已冻结" : t.status === 1 ? "已使用" : "未使用"
            }}</template>
          </Tag>
        </template>
        <template #createTime="{ record: t }">{{
          formatDateTime(t.createTime)
        }}</template>
        <template #useTime="{ record: t }">
          {{ t.useTime ? formatDateTime(t.useTime) : "-" }}
        </template>
        <template #useUser="{ record: t }">{{ t.useUser || "-" }}</template>
        <template #remark="{ record: t }">{{ t.remark || "-" }}</template>
        <template #lastActiveTime="{ record }">{{ record.lastActiveTime ? formatDateTime(record.lastActiveTime) : "—" }}</template>
<template #operations="{ record: t }">
          <Link @click="(Q) => te(t)"><template #default>编辑</template></Link>
          <Popconfirm :content="t.status === 2 ? '确认解冻？解冻会保留原使用状态。' : '冻结后该卡密将无法登录或充值，确认？'" @ok="() => toggleFrozen(t)">
            <Link :disabled="freezing" :style="{ marginLeft: '8px' }">{{ t.status === 2 ? '解冻' : '冻结' }}</Link>
          </Popconfirm>
          <Popconfirm content="确定删除？" @ok="(Q) => handleDeleteKami(t)">
            <template #default>
              <Link
                status="danger"
                :disabled="!S &amp;&amp; t.status === 1"
                :style="{
                  'margin-left': '8px',
                }"
              >
                <template #default>删除</template>
              </Link>
            </template>
          </Popconfirm>
        </template>
      </Table>
    </div>
    <Drawer
      v-model:visible="visible"
      title="编辑卡密"
      :on-before-cancel="() => !loadingState"
      :closable="!loadingState"
      :mask-closable="!loadingState"
      :esc-to-close="!loadingState"
      :width="480"
      :footer="!1"
      unmount-on-close=""
    >
      <template #default>
        <div class="kami-edit">
          <div class="id-card">
            <div class="id-card-top">
              <div class="card-no" :title="formData.card">
                {{ formData.card }}
              </div>
              <Tag class="card-tag">
                <template #default>{{
                  formData.status === 2 ? "已冻结" : formData.status === 1 ? "已使用" : "未使用"
                }}</template>
              </Tag>
            </div>
            <div class="id-card-label">卡号</div>
            <div class="id-row">
              <span>卡密组</span>
              <span>{{ formData.groupName || "-" }}</span>
            </div>
            <div class="id-row">
              <span>创建者</span>
              <span>{{ formData.creator || "-" }}</span>
            </div>
            <div class="id-row">
              <span>创建时间</span>
              <span>{{
                formData.createTime ? formatDateTime(formData.createTime) : "-"
              }}</span>
            </div>
            <div class="id-row">
              <span>创建IP</span>
              <span>{{ formData.createIp || "-" }}</span>
            </div>
          </div>
          <Tabs v-model:active-key="pageState" type="line" class="edit-tabs">
            <template #default>
              <TabPane v-if="S" key="sessions" title="会话管理"><AccountSessions v-if="visible && pageState === 'sessions'" :account-id="formData.id" kind="kami" /></TabPane>
<TabPane key="basic" title="基本信息">
                <template #default>
                  <Form
                    :model="formData"
                    :disabled="loadingState"
                    layout="vertical"
                    size="medium"
                    class="basic-form"
                  >
                    <template #default>
                      <FormItem label="卡密备注">
                        <template #default>
                          <Input
                            v-model="formData.remark"
                            placeholder="如：活动卡密(可空)"
                            allow-clear=""
                          ></Input>
                        </template>
                      </FormItem>
                      <template v-if="Y">
                        <FormItem :key="0">
                          <template #label>
                            额外绑定设备数量
                            <Tooltip
                              :content="`系统基线允许 1 个设备(首次登录设备),此处填的是额外配额。共允许 ${1 + (formData.bindLimit || 0)} 个设备`"
                            >
                              <template #default>
                                <IconQuestionCircle
                                  class="help-ico"
                                ></IconQuestionCircle>
                              </template>
                            </Tooltip>
                          </template>
                          <template #default>
                            <InputNumber
                              v-model="formData.bindLimit"
                              :min="0"
                              :precision="0"
                              placeholder="0"
                            ></InputNumber>
                          </template>
                        </FormItem>
                      </template>
                      <template v-else></template>
                      <template v-if="Z">
                        <FormItem :key="1">
                          <template #label>
                            额外绑定IP区域数量
                            <Tooltip
                              :content="`系统基线允许 1 个地点(首次登录地),此处填的是额外配额。共允许 ${1 + (formData.regionLimit || 0)} 个地点;距离阈值 ${ee}km`"
                            >
                              <template #default>
                                <IconQuestionCircle
                                  class="help-ico"
                                ></IconQuestionCircle>
                              </template>
                            </Tooltip>
                          </template>
                          <template #default>
                            <InputNumber
                              v-model="formData.regionLimit"
                              :min="0"
                              :precision="0"
                              placeholder="0"
                            ></InputNumber>
                          </template>
                        </FormItem>
                      </template>
                      <template v-else></template>
                    </template>
                  </Form>
                </template>
              </TabPane>
            </template>
          </Tabs>
          <div class="submit-bar">
            <Button
              type="primary"
              long=""
              :loading="loadingState"
              @click="handleUpdateKami"
            >
              <template #default>提交</template>
            </Button>
          </div>
        </div>
      </template>
    </Drawer>
    <Modal
      v-model:visible="x"
      title="创建卡密"
      :on-before-ok="() => keepModalOpen(handleGenerateKami)"
      width="auto"
      :ok-loading="P"
      :on-before-cancel="() => !P"
      :closable="!P"
      :mask-closable="!P"
      :esc-to-close="!P"
      :cancel-button-props="{ disabled: P }"
      modal-class="gen-modal"
    >
      <template #default>
        <div class="md-w-80">
          <Form
            :model="i"
            size="small"
            auto-label-width=""
            class="gen-form"
            :disabled="P"
          >
            <template #default>
              <FormItem label="卡密组">
                <template #default>
                  <Select v-model="i.groupId" placeholder="选择分组">
                    <template #default>
                      <template v-for="t in F" :key="t.id">
                        <Option :value="t.id">
                          <template #default>
                            {{
                              t.name +
                              "（" +
                              (t.value >= 9999999999
                                ? "永久"
                                : formatDuration(t.value)) +
                              "） "
                            }}
                          </template>
                        </Option>
                      </template>
                    </template>
                  </Select>
                </template>
              </FormItem>
              <FormItem label="备注">
                <template #default>
                  <Input
                    v-model="i.remark"
                    placeholder="如: 活动卡密 (可空)"
                    allow-clear=""
                  ></Input>
                </template>
              </FormItem>
              <FormItem>
                <template #label>
                  <span>长度</span>
                  <Tooltip content="卡密随机部分长度 (不含前缀),范围 8-32">
                    <template #default>
                      <IconQuestionCircle class="help-ico"></IconQuestionCircle>
                    </template>
                  </Tooltip>
                </template>
                <template #default>
                  <InputNumber
                    v-model="i.length"
                    :min="8"
                    :max="32"
                    :style="{
                      width: '100%',
                    }"
                  ></InputNumber>
                </template>
              </FormItem>
              <FormItem>
                <template #label>
                  <span>前缀</span>
                  <Tooltip content="可选,会拼接到卡号最前面">
                    <template #default>
                      <IconQuestionCircle class="help-ico"></IconQuestionCircle>
                    </template>
                  </Tooltip>
                </template>
                <template #default>
                  <Input
                    v-model="i.prefix"
                    placeholder="如: TK- (可空)"
                    allow-clear=""
                  ></Input>
                </template>
              </FormItem>
              <FormItem>
                <template #label>
                  <span>数量</span>
                  <Tooltip content="单次生成数量,最多 10000">
                    <template #default>
                      <IconQuestionCircle class="help-ico"></IconQuestionCircle>
                    </template>
                  </Tooltip>
                </template>
                <template #default>
                  <InputNumber
                    v-model="i.count"
                    :min="1"
                    :max="1e4"
                    :style="{
                      width: '100%',
                    }"
                  >
                    <template #suffix>张</template>
                  </InputNumber>
                </template>
              </FormItem>
              <FormItem label="导出">
                <template #default>
                  <Select v-model="i.exportFmt">
                    <template #default>
                      <Option value="none"
                        ><template #default>不导出</template></Option
                      >
                      <Option value="txt"
                        ><template #default>txt</template></Option
                      >
                      <Option value="csv"
                        ><template #default>csv</template></Option
                      >
                    </template>
                  </Select>
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
.card-cell {
  cursor: pointer;
  color: var(--color-text-1);
}
.card-cell:hover {
  color: rgb(var(--primary-6));
}
.help-ico {
  color: var(--color-text-3);
  margin-left: 4px;
  cursor: help;
}
.kami-edit {
  display: flex;
  flex-direction: column;
  height: 100%;
}
.kami-edit .id-card {
  background: linear-gradient(
    135deg,
    rgb(var(--primary-5)),
    rgb(var(--primary-7))
  );
  padding: 16px 18px;
  color: #fff;
  margin-bottom: 12px;
}
.kami-edit .id-card .id-card-top {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 8px;
}
.kami-edit .id-card .id-card-top .card-no {
  font-size: 18px;
  font-weight: 600;
  letter-spacing: 0.5px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.kami-edit .id-card .id-card-top .card-tag {
  background: #ffffff2e;
  color: #fff;
  border: none;
  flex-shrink: 0;
}
.kami-edit .id-card .id-card-label {
  font-size: 12px;
  opacity: 0.75;
  margin-bottom: 14px;
}
.kami-edit .id-card .id-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 13px;
  padding: 6px 0;
  border-top: 1px solid rgba(255, 255, 255, 0.12);
}
.kami-edit .id-card .id-row span:first-child {
  opacity: 0.75;
}
.kami-edit .id-card .id-row span:last-child {
  font-weight: 500;
  max-width: 60%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.kami-edit .edit-tabs {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
}
.kami-edit .edit-tabs :deep(.arco-tabs-content) {
  padding-top: 12px;
}
.kami-edit .basic-form :deep(.arco-form-item) {
  margin-bottom: 14px;
}
.kami-edit .basic-form :deep(.arco-input-number) {
  width: 100%;
}
.kami-edit .submit-bar {
  padding: 14px 0 0;
  border-top: 1px solid var(--color-border-2);
  flex-shrink: 0;
}
.gen-form :deep(.arco-form-item) {
  margin-bottom: 12px;
}
.gen-form :deep(.arco-form-item:last-child) {
  margin-bottom: 0;
}
.md-w-80 {
  width: 20rem;
}
</style>
