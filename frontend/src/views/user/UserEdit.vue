<!-- Reconstructed editable Vue source; see README for recovery limits. -->
<script>
import AccountSessions from "@/components/AccountSessions.vue";
import { reportError } from "@/utils/errors.js";
import { onMounted, reactive, ref, computed } from "vue";
import {
  Breadcrumb,
  Upload,
  Tabs,
  BreadcrumbItem,
  TabPane,
  Form,
  Row,
  Col,
  FormItem,
  Input,
  Select,
  Option,
  DatePicker,
  InputNumber,
  Tooltip,
  InputPassword,
  Switch,
  Button,
  Table,
  Textarea,
  Popconfirm,
  Link,
  Empty,
  Tag,
  Message,
} from "@arco-design/web-vue";
import { useRoute, useRouter } from "vue-router";
import {
  IconCamera,
  IconQuestionCircle,
  IconDelete,
  IconPlus,
} from "@arco-design/web-vue/es/icon";
import {
  listUserGroups,
  getUser,
  listUserExtends,
  listUserDevices,
  listUserLogs,
  updateUser,
  saveUserExtends,
  deleteUserExtend,
  deleteAllUserDevices,
  deleteUserDevice,
} from "@/api/user.js";
import { uploadAvatar } from "@/api/auth.js";
import { useAppStore } from "@/stores/app.js";

import { formatDateTime } from "@/utils/format.js";
import { exportA } from "@/utils/avatar.js";
import { entryPath } from "@/utils/route.js";
const H = 9999999999;
export default {
  name: "UserEdit",
  components: {
AccountSessions,
    Breadcrumb,
    BreadcrumbItem,
    Upload,
    IconCamera,
    Tabs,
    TabPane,
    Form,
    Row,
    Col,
    FormItem,
    Input,
    Select,
    Option,
    DatePicker,
    Tooltip,
    IconQuestionCircle,
    InputNumber,
    InputPassword,
    Switch,
    Button,
    Table,
    IconPlus,
    Textarea,
    IconDelete,
    Popconfirm,
    Link,
    Empty,
    Tag,
  },
  setup(ia) {
    const route = useRoute();
    const router = useRouter();
    const appStore = useAppStore();
    const formRef = ref();
    const G = ref("basic");
    const loading = ref(false);
    const ready = ref(false);
    const U = computed(() => !!appStore.current?.bindDevice);
    const J = computed(() => !!appStore.current?.bindRegion);
    const K = computed(() => appStore.current?.ipDistanceMax || 50);
    const formData = reactive({
      id: 0,
      username: "",
      nickname: "",
      avatar: "",
      registerIp: "",
      createTime: 0,
      groupId: void 0,
      vipMode: "date",
      vipDate: null,
      bindLimit: 0,
      regionLimit: 0,
      ipList: "",
      password: "",
      banEnabled: false,
      banPermanent: false,
      banDate: null,
    });
    const tableRows = computed(() => {
      try {
        return (JSON.parse(formData.ipList || "[]") || []).map((e, r) => ({
          ...e,
          _k: `ip${r}`,
        }));
      } catch {
        return [];
      }
    });
    const R = ref([]);
    const handleListUserGroups = async () => {
      try {
        R.value = (await listUserGroups()) || [];
      } catch {
        R.value = [];
      }
    };
    const handleUploadAvatar = async (l, e) => {
      const r = e?.file;
      if (r)
        try {
          const { url: p } = await uploadAvatar(r);
          formData.avatar = p;
          Message.success("头像上传成功,保存后生效");
        } catch (p) {
          reportError(p, "上传失败");
        }
    };
    const handleChange = (l) => {
      if (!l) { formData.banDate = null; formData.banPermanent = false; }
    };
    const handleClick = () => router.push(entryPath("/user"));
    const handleGetUser = async () => {
      const l = Number(route.params.id);
      const e = await getUser(l);
      e &&
        (Object.assign(formData, {
          id: e.id,
          username: e.username,
          nickname: e.nickname || "",
          avatar: e.avatar || "",
          registerIp: e.registerIp || e.lastIp || "",
          createTime: e.createTime,
          groupId: e.groupId || void 0,
          vipMode: e.vip >= H ? "forever" : "date",
          vipDate: e.vip && e.vip < H ? e.vip * 1e3 : null,
          bindLimit: e.bindLimit || 0,
          regionLimit: e.regionLimit || 0,
          ipList: e.ipList || "",
          password: "",
          banEnabled: !!e.banExpire,
          banPermanent: e.banExpire < 0,
          banDate: e.banExpire > 0 ? e.banExpire * 1e3 : null,
        }),
        handleListUserExtends(),
        U.value && handleListUserDevices(),
        handleListUserLogs());
      ready.value = !!e;
    };
    const handleUpdateUser = async () => {
      if (!ready.value || loading.value) return;
      if (formData.banEnabled && !formData.banPermanent && !formData.banDate) {
        Message.error("请选择禁用截止时间或永久禁用");
        return;
      }
      loading.value = true;
      try {
        const l = {
          vip:
            formData.vipMode === "forever"
              ? H
              : formData.vipDate
                ? Math.floor(formData.vipDate / 1e3)
                : 0,
          bindLimit: formData.bindLimit,
          regionLimit: formData.regionLimit,
          banExpire: !formData.banEnabled ? 0 : formData.banPermanent ? -1 : Math.floor(formData.banDate / 1e3),
          nickname: formData.nickname,
          avatar: formData.avatar,
        };
        formData.groupId && (l.groupId = formData.groupId);
        formData.password && (l.password = formData.password);
        await updateUser(formData.id, l);
        formData.password = "";
        Message.success("已保存");
      } catch (error) {
        reportError(error);
      } finally {
        loading.value = false;
      }
    };
    const handleUpdateUser2 = async () => {
      await updateUser(formData.id, {
        clearIpList: true,
      });
      formData.ipList = "";
      Message.success("已清空登录地点记录");
    };
    const tableColumns2 = [
      {
        title: "IP",
        slotName: "ip",
        width: 160,
        ellipsis: true,
      },
      {
        title: "省/市",
        slotName: "region",
        width: 200,
        ellipsis: true,
        tooltip: true,
      },
      {
        title: "经纬度",
        slotName: "latlng",
        width: 200,
      },
      {
        title: "最近登录",
        slotName: "time",
        align: "right",
        width: 200,
      },
      {
        title: "操作",
        slotName: "operations",
        fixed: "right",
        width: 80,
      },
    ];
    const ge = (l, e) =>
      `${Number(l || 0).toFixed(4)}, ${Number(e || 0).toFixed(4)}`;
    const be = (l, e) => {
      const r = (l || "").trim();
      const p = (e || "").trim();
      return r && p && r !== p ? `${r} ${p}` : r || p || "-";
    };
    const handleUpdateUser3 = async (l) => {
      const e = tableRows.value
        .filter((r, p) => p !== l)
        .map(({ _k: r, ...p }) => p);
      await updateUser(formData.id, {
        ipList: JSON.stringify(e),
      });
      formData.ipList = JSON.stringify(e);
      Message.success("已移除");
    };
    const rows = ref([]);
    const loadingState = ref(false);
    const loadingState2 = ref(false);
    let ke = 0;
    const tableColumns = [
      {
        title: "变量名",
        slotName: "key",
        width: 220,
      },
      {
        title: "变量值",
        slotName: "value",
      },
      {
        title: "",
        slotName: "operations",
        titleSlotName: "title-op",
        align: "right",
        width: 60,
      },
    ];
    const handleListUserExtends = async () => {
      loadingState.value = true;
      try {
        const l = await listUserExtends(formData.id);
        rows.value = l.map((e) => ({
          ...e,
          _k: `s${e.id}`,
        }));
      } catch (error) {
        reportError(error);
      } finally {
        loadingState.value = false;
      }
    };
    const handleClick2 = () => {
      if (!ready.value || loadingState.value || loadingState2.value) return;
      rows.value.push({
        id: 0,
        key: "",
        value: "",
        _k: `n${++ke}`,
      });
    };
    const he = async (l) => {
      if (loadingState2.value || loadingState.value) return;
      const row = rows.value[l];
      if (!row) return;
      loadingState2.value = true;
      try {
        if (row.id) await deleteUserExtend(formData.id, row.id);
        rows.value.splice(l, 1);
      } catch (error) {
        reportError(error);
      } finally {
        loadingState2.value = false;
      }
    };
    const handleSaveUserExtends = async () => {
      if (!ready.value) return;
      if (loadingState2.value || loadingState.value) return;
      const keys = rows.value.map((row) => row.key.trim());
      if (keys.some((key) => !key) || new Set(keys).size !== keys.length) {
        Message.warning("变量名不能为空或重复");
        return;
      }
      loadingState2.value = true;
      try {
        await saveUserExtends(formData.id, {
          list: rows.value
            .filter((l) => l.key)
            .map(({ id: l, key: e, value: r }) => ({
              id: l || 0,
              key: e.trim(),
              value: r,
            })),
        });
        Message.success("已保存");
        await handleListUserExtends();
      } catch (error) {
        reportError(error);
      } finally {
        loadingState2.value = false;
      }
    };
    const tableRows2 = ref([]);
    const loadingState3 = ref(false);
    const tableColumns3 = [
      {
        title: "机器码",
        slotName: "mi",
        width: 280,
        ellipsis: true,
      },
      {
        title: "IP / 区域",
        slotName: "ip",
        align: "right",
        width: 200,
        ellipsis: true,
        tooltip: true,
      },
      {
        title: "绑定时间",
        slotName: "bindTime",
        align: "right",
        width: 200,
      },
      {
        title: "最近活跃",
        slotName: "lastTime",
        align: "right",
        width: 200,
      },
      {
        title: "操作",
        slotName: "operations",
        fixed: "right",
        width: 80,
      },
    ];
    const handleListUserDevices = async () => {
      loadingState3.value = true;
      try {
        tableRows2.value = await listUserDevices(formData.id);
      } finally {
        loadingState3.value = false;
      }
    };
    const handleDeleteUserDevice = async (l) => {
      await deleteUserDevice(formData.id, l.id);
      Message.success("已解绑");
      handleListUserDevices();
    };
    const handleDeleteAllUserDevices = async () => {
      const { affected: l = 0 } =
        (await deleteAllUserDevices(formData.id)) || {};
      Message.success(`已解绑 ${l} 台设备`);
      handleListUserDevices();
    };
    const k = reactive({
      page: 1,
      size: 10,
    });
    const tableRows3 = ref([]);
    const j = ref(0);
    const loadingState4 = ref(false);
    const tableColumns4 = [
      {
        title: "类型",
        slotName: "type",
        width: 120,
      },
      {
        title: "操作",
        dataIndex: "action",
        ellipsis: true,
        tooltip: true,
      },
    ];
    const pagination = computed(() => ({
      current: k.page,
      pageSize: k.size,
      total: j.value,
      showTotal: true,
      showPageSize: true,
      pageSizeOptions: [10, 20, 50, 100],
    }));
    const handleListUserLogs = async () => {
      loadingState4.value = true;
      try {
        const l = await listUserLogs(formData.id, k);
        tableRows3.value = l.list;
        j.value = l.total;
      } finally {
        loadingState4.value = false;
      }
    };
    const Te = (l) =>
      ({
        login: "登录",
        kami: "卡密",
        func: "云函数",
        user: "账户",
      })[l] || l;
    const Ee = (l) =>
      ({
        login: "blue",
        kami: "orange",
        func: "purple",
        user: "green",
      })[l] || "gray";
    const q = async (l) => {
      if (l)
        try {
          await navigator.clipboard.writeText(l);
          Message.success("已复制");
        } catch {
          Message.error("复制失败");
        }
    };
    onMounted(() => {
      handleListUserGroups();
      handleGetUser().catch(reportError);
    });
    return {
      ready,
      route,
      router,
      appStore,
      G,
      U,
      J,
      K,
      tableRows,
      R,
      handleListUserGroups,
      handleUploadAvatar,
      handleChange,
      handleClick,
      handleGetUser,
      handleUpdateUser,
      handleUpdateUser2,
      tableColumns2,
      ge,
      be,
      handleUpdateUser3,
      loadingState,
      loadingState2,
      ke,
      tableColumns,
      handleListUserExtends,
      handleClick2,
      he,
      handleSaveUserExtends,
      tableRows2,
      loadingState3,
      tableColumns3,
      handleListUserDevices,
      handleDeleteUserDevice,
      handleDeleteAllUserDevices,
      k,
      tableRows3,
      j,
      loadingState4,
      tableColumns4,
      handleListUserLogs,
      Te,
      Ee,
      q,
      formRef,
      formData,
      loading,
      rows,
      pagination,
      exportA,
      formatDateTime,
    };
  },
};
</script>
<template>
  <div class="user-edit-page">
    <Breadcrumb class="crumb">
      <template #default>
        <BreadcrumbItem>
          <template #default><a @click="handleClick">用户列表</a></template>
        </BreadcrumbItem>
        <BreadcrumbItem><template #default>编辑用户</template></BreadcrumbItem>
      </template>
    </Breadcrumb>
    <div class="profile-card">
      <Upload
        :show-file-list="!1"
        :auto-upload="!1"
        accept="image/*"
        @change="handleUploadAvatar"
      >
        <template #upload-button>
          <div class="avatar-upload">
            <img :src="formData.avatar || exportA" alt="avatar" />
            <div class="avatar-mask"><IconCamera></IconCamera></div>
          </div>
        </template>
      </Upload>
      <div class="profile-meta">
        <div class="profile-name">{{ formData.nickname || "匿名用户" }}</div>
        <div class="profile-row">
          <span class="item">{{ "账号：" + formData.username }}</span>
          <span class="sep">|</span>
          <span class="item">{{
            "注册IP：" + (formData.registerIp || "-")
          }}</span>
          <span class="sep">|</span>
          <span class="item">{{
            "注册时间：" + formatDateTime(formData.createTime)
          }}</span>
        </div>
      </div>
    </div>
    <Tabs v-model:active-key="G" class="edit-tabs">
      <template #default>
        <TabPane key="basic" title="基本信息">
          <template #default>
            <Form
              ref="formRef"
              :model="formData"
              :disabled="!ready || loading"
              layout="vertical"
              class="edit-form"
            >
              <template #default>
                <Row :gutter="16">
                  <template #default>
                    <Col :xs="24" :md="12">
                      <template #default>
                        <FormItem field="nickname" label="昵称">
                          <template #default>
                            <Input
                              v-model="formData.nickname"
                              placeholder="请输入昵称"
                              allow-clear=""
                              :max-length="32"
                            ></Input>
                          </template>
                        </FormItem>
                      </template>
                    </Col>
                    <Col :xs="24" :md="12">
                      <template #default>
                        <FormItem field="groupId" label="用户分组">
                          <template #default>
                            <Select
                              v-model="formData.groupId"
                              placeholder="选择分组"
                            >
                              <template #default>
                                <template v-for="t in R" :key="t.id">
                                  <Option :value="t.id">
                                    <template #default>
                                      {{ t.name + " (Lv." + t.level + ") " }}
                                    </template>
                                  </Option>
                                </template>
                              </template>
                            </Select>
                          </template>
                        </FormItem>
                      </template>
                    </Col>
                    <Col :xs="24" :md="12">
                      <template #default>
                        <FormItem field="vipDate" label="会员到期时间">
                          <template #default>
                            <div class="vip-row">
                              <Select
                                v-model="formData.vipMode"
                                :style="{
                                  width: '80px',
                                  flexShrink: 0,
                                }"
                              >
                                <template #default>
                                  <Option value="date"
                                    ><template #default>限时</template></Option
                                  >
                                  <Option value="forever">
                                    <template #default>永久</template>
                                  </Option>
                                </template>
                              </Select>
                              <DatePicker
                                v-model="formData.vipDate"
                                show-time=""
                                format="YYYY-MM-DD HH:mm:ss"
                                value-format="timestamp"
                                placeholder="选择到期日期"
                                :disabled="formData.vipMode === 'forever'"
                                :style="{
                                  flex: 1,
                                  minWidth: 0,
                                }"
                              ></DatePicker>
                            </div>
                          </template>
                        </FormItem>
                      </template>
                    </Col>
                    <template v-if="U">
                      <Col :key="0" :xs="24" :md="12">
                        <template #default>
                          <FormItem field="bindLimit">
                            <template #label>
                              额外绑定设备数量
                              <Tooltip
                                :content="`系统基线允许 1 个设备(首次登录设备),此处填的是额外配额。共允许 ${1 + (formData.bindLimit || 0)} 个设备`"
                              >
                                <template #default>
                                  <IconQuestionCircle
                                    class="hint-icon"
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
                                :style="{
                                  width: '100%',
                                }"
                              ></InputNumber>
                            </template>
                          </FormItem>
                        </template>
                      </Col>
                    </template>
                    <template v-else></template>
                    <template v-if="J">
                      <Col :key="1" :xs="24" :md="12">
                        <template #default>
                          <FormItem field="regionLimit">
                            <template #label>
                              额外绑定IP区域数量
                              <Tooltip
                                :content="`系统基线允许 1 个地点(首次登录地),此处填的是额外配额。共允许 ${1 + (formData.regionLimit || 0)} 个地点;距离阈值 ${K}km`"
                              >
                                <template #default>
                                  <IconQuestionCircle
                                    class="hint-icon"
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
                                :style="{
                                  width: '100%',
                                }"
                              ></InputNumber>
                            </template>
                          </FormItem>
                        </template>
                      </Col>
                    </template>
                    <template v-else></template>
                    <Col :xs="24" :md="12">
                      <template #default>
                        <FormItem field="password" label="密码">
                          <template #default>
                            <InputPassword
                              v-model="formData.password"
                              placeholder="空则不修改密码"
                              allow-clear=""
                            ></InputPassword>
                          </template>
                        </FormItem>
                      </template>
                    </Col>
                    <Col :xs="24" :md="12">
                      <template #default>
                        <FormItem>
                          <template #label>
                            <Switch
                              v-model="formData.banEnabled"
                              size="small"
                              @change="handleChange"
                            ></Switch>
                            <span class="ban-label">禁用用户期限</span>
                          </template>
                          <template #default>
                            <div class="ban-controls">
                            <div class="ban-date">
                            <DatePicker
                              :style="{ width: '100%' }"
                              v-model="formData.banDate"
                              show-time=""
                              format="YYYY-MM-DD HH:mm:ss"
                              value-format="timestamp"
                              placeholder="请选择日期时间"
                              :disabled="!formData.banEnabled || formData.banPermanent"
                            ></DatePicker>
                            </div>
                            <div class="ban-permanent">
                              <Switch v-model="formData.banPermanent" :disabled="!formData.banEnabled" size="small" aria-label="永久禁用" />
                              <span class="ban-label">永久禁用</span>
                            </div>
                            </div>
                          </template>
                        </FormItem>
                      </template>
                    </Col>
                  </template>
                </Row>
                <Button
                  type="primary"
                  long=""
                  :loading="loading"
                  :disabled="!ready"
                  @click="handleUpdateUser"
                >
                  <template #default>提交</template>
                </Button>
              </template>
            </Form>
          </template>
        </TabPane>
        <TabPane key="extend" title="扩展信息">
          <template #default>
            <Table
              :data="rows"
              :columns="tableColumns"
              :loading="loadingState"
              :pagination="!1"
              :bordered="!1"
              row-key="_k"
              :class="[
                'extend-table',
                {
                  'is-empty': !rows.length,
                },
              ]"
            >
              <template #title-op>
                <Tooltip content="新增变量">
                  <template #default>
                    <IconPlus class="add-icon" @click="handleClick2"></IconPlus>
                  </template>
                </Tooltip>
              </template>
              <template #key="{ record: t }">
                <Input v-model="t.key" placeholder="如: name" :disabled="loadingState2"></Input>
              </template>
              <template #value="{ record: t }">
                <Textarea v-model="t.value" placeholder="如: 张三" :disabled="loadingState2"></Textarea>
              </template>
              <template #operations="{ rowIndex: t }">
                <Button type="text" status="danger" :disabled="loadingState2" aria-label="删除变量" @click="(f) => he(t)">
                  <template #icon><IconDelete></IconDelete></template>
                </Button>
              </template>
              <template #empty><span></span></template>
            </Table>
            <template v-if="rows.length">
              <Button
                :key="0"
                type="primary"
                long=""
                :loading="loadingState2"
                class="extend-submit"
                @click="handleSaveUserExtends"
              >
                <template #default>提交保存</template>
              </Button>
            </template>
            <template v-else></template>
          </template>
        </TabPane>
        <template v-if="J">
          <TabPane key="ipregion" title="IP区域">
            <template #default>
              <div class="iplist-toolbar">
                <div class="iplist-info">
                  {{
                    " 已记录 " +
                    tableRows.length +
                    " / " +
                    (1 + (formData.regionLimit || 0)) +
                    " 个地点 "
                  }}
                  <span class="muted">
                    {{
                      "(系统基线 1 + 额外 " +
                      (formData.regionLimit || 0) +
                      ";距离阈值 " +
                      K +
                      "km)"
                    }}
                  </span>
                </div>
                <Popconfirm
                  content="确定解绑该用户名下的全部 IP 地点?"
                  @ok="handleUpdateUser2"
                >
                  <template #default>
                    <Button
                      size="small"
                      status="danger"
                      :disabled="!tableRows.length"
                    >
                      <template #default>解绑全部</template>
                    </Button>
                  </template>
                </Popconfirm>
              </div>
              <template v-if="tableRows.length">
                <Table
                  :key="0"
                  :data="tableRows"
                  :columns="tableColumns2"
                  :pagination="!1"
                  row-key="_k"
                >
                  <template #ip="{ record: t }">
                    <Tooltip :content="t.ip">
                      <template #default>
                        <span class="copy-cell" @click="(f) => q(t.ip)">{{
                          t.ip || "-"
                        }}</span>
                      </template>
                    </Tooltip>
                  </template>
                  <template #region="{ record: t }">
                    <template v-if="t.province || t.city">
                      <span :key="0">{{ be(t.province, t.city) }}</span>
                    </template>
                    <template v-else
                      ><span :key="1" class="muted">未解析</span></template
                    >
                  </template>
                  <template #latlng="{ record: t }">
                    <template v-if="t.lat || t.lng">
                      <span :key="0">{{ ge(t.lat, t.lng) }}</span>
                    </template>
                    <template v-else
                      ><span :key="1" class="muted">未解析</span></template
                    >
                  </template>
                  <template #time="{ record: t }">{{
                    formatDateTime(t.time)
                  }}</template>
                  <template #operations="{ rowIndex: t }">
                    <Popconfirm
                      content="移除该地点?"
                      @ok="(f) => handleUpdateUser3(t)"
                    >
                      <template #default>
                        <Link status="danger"
                          ><template #default>移除</template></Link
                        >
                      </template>
                    </Popconfirm>
                  </template>
                </Table>
              </template>
              <template v-else>
                <Empty
                  :key="1"
                  description="暂无 IP 地点记录"
                  class="tab-empty"
                ></Empty>
              </template>
            </template>
          </TabPane>
        </template>
        <template v-else></template>
        <template v-if="U">
          <TabPane key="device" title="绑定设备">
            <template #default>
              <div class="iplist-toolbar">
                <div class="iplist-info">
                  {{
                    " 已记录 " +
                    tableRows2.length +
                    " / " +
                    (1 + (formData.bindLimit || 0)) +
                    " 个设备 "
                  }}
                  <span class="muted">
                    {{
                      "(系统基线 1 + 额外 " + (formData.bindLimit || 0) + ")"
                    }}
                  </span>
                </div>
                <Popconfirm
                  content="确定解绑该用户名下的全部设备?"
                  @ok="handleDeleteAllUserDevices"
                >
                  <template #default>
                    <Button
                      size="small"
                      status="danger"
                      :disabled="!tableRows2.length"
                    >
                      <template #default>解绑全部</template>
                    </Button>
                  </template>
                </Popconfirm>
              </div>
              <template v-if="tableRows2.length">
                <Table
                  :key="0"
                  :data="tableRows2"
                  :columns="tableColumns3"
                  :loading="loadingState3"
                  :pagination="!1"
                  row-key="id"
                >
                  <template #mi="{ record: t }">
                    <Tooltip :content="t.mi">
                      <template #default>
                        <span class="copy-cell" @click="(f) => q(t.mi)">{{
                          t.mi
                        }}</span>
                      </template>
                    </Tooltip>
                  </template>
                  <template #ip="{ record: t }">
                    <div class="dev-cell dev-cell-right">
                      <div class="dev-title">{{ t.region || "-" }}</div>
                      <div class="dev-sub">{{ t.ip }}</div>
                    </div>
                  </template>
                  <template #bindTime="{ record: t }">{{
                    formatDateTime(t.bindTime)
                  }}</template>
                  <template #lastTime="{ record: t }">{{
                    formatDateTime(t.lastTime)
                  }}</template>
                  <template #operations="{ record: t }">
                    <Popconfirm
                      content="确定解绑该设备？"
                      @ok="(f) => handleDeleteUserDevice(t)"
                    >
                      <template #default>
                        <Link status="danger"
                          ><template #default>解绑</template></Link
                        >
                      </template>
                    </Popconfirm>
                  </template>
                </Table>
              </template>
              <template v-else>
                <Empty
                  :key="1"
                  description="暂无绑定设备"
                  class="tab-empty"
                ></Empty>
              </template>
            </template>
          </TabPane>
        </template>
        <template v-else></template>
        <TabPane key="sessions" title="会话管理"><AccountSessions v-if="formData.id && G === 'sessions'" :account-id="formData.id" /></TabPane>
<TabPane key="logs" title="用户日志">
          <template #default>
            <Table
              :data="tableRows3"
              :columns="tableColumns4"
              :loading="loadingState4"
              :pagination="pagination"
              row-key="id"
              @pageChange="
                (t) => {
                  ((k.page = t), handleListUserLogs());
                }
              "
              @pageSizeChange="
                (t) => {
                  ((k.size = t), (k.page = 1), handleListUserLogs());
                }
              "
            >
              <template #type="{ record: t }">
                <Tag :color="Ee(t.type)">
                  <template #default>{{ Te(t.type) }}</template>
                </Tag>
              </template>
            </Table>
          </template>
        </TabPane>
      </template>
    </Tabs>
  </div>
</template>
<style scoped>
.user-edit-page {
  padding: 16px 20px;
}
.crumb {
  font-size: 13px;
  margin-bottom: 12px;
}
.crumb a {
  color: var(--color-text-3);
  cursor: pointer;
}
.crumb a:hover {
  color: rgb(var(--primary-6));
}
.profile-card {
  background: var(--color-bg-2);
  border-radius: 6px;
  padding: 18px 20px;
  display: flex;
  align-items: center;
  gap: 16px;
  box-shadow: 0 1px 2px #0000000a;
  margin-bottom: 16px;
}
.profile-card .avatar-upload {
  position: relative;
  width: 56px;
  height: 56px;
  border-radius: 50%;
  overflow: hidden;
  background: var(--color-fill-2);
  cursor: pointer;
  flex-shrink: 0;
}
.profile-card .avatar-upload img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}
.profile-card .avatar-upload .avatar-mask {
  position: absolute;
  inset: 0;
  background: #00000073;
  color: #fff;
  font-size: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: opacity 0.2s;
}
.profile-card .avatar-upload:hover .avatar-mask {
  opacity: 1;
}
.profile-card .profile-meta {
  min-width: 0;
  flex: 1;
}
.profile-card .profile-name {
  font-size: 15px;
  font-weight: 600;
  color: var(--color-text-1);
  margin-bottom: 4px;
}
.profile-card .profile-row {
  color: var(--color-text-3);
  font-size: 13px;
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}
.profile-card .profile-row .sep {
  opacity: 0.4;
}
.edit-tabs {
  background: var(--color-bg-2);
  border-radius: 6px;
  padding: 0 20px 20px;
  box-shadow: 0 1px 2px #0000000a;
}
.edit-form {
  max-width: 100%;
}
.edit-form :deep(.arco-form-item-label) {
  color: var(--color-text-2);
  font-size: 13px;
}
.edit-form .hint-icon {
  margin-left: 4px;
  color: var(--color-text-3);
  cursor: help;
}
.edit-form .ban-label {
  margin-left: 8px;
}
.ban-controls {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px 16px;
  width: 100%;
  min-width: 0;
}
.ban-controls .ban-date {
  flex: 1 1 220px;
  width: auto;
  min-width: 0;
  max-width: 100%;
}
.ban-permanent {
  display: inline-flex;
  align-items: center;
  flex: 0 0 auto;
  min-height: 32px;
  white-space: nowrap;
}
.vip-row {
  display: flex;
  gap: 8px;
  align-items: center;
  width: 100%;
}
.iplist-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}
.iplist-toolbar .iplist-info {
  color: var(--color-text-2);
  font-size: 13px;
}
.iplist-toolbar .iplist-info .muted {
  color: var(--color-text-3);
  margin-left: 6px;
}
.muted {
  color: var(--color-text-3);
}
.dev-cell {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.dev-cell-right {
  align-items: flex-end;
}
.dev-cell .dev-title {
  color: var(--color-text-1);
  font-size: 14px;
}
.dev-cell .dev-sub {
  color: var(--color-text-3);
  font-size: 12px;
}
.copy-cell {
  cursor: pointer;
  color: var(--color-text-1);
}
.copy-cell:hover {
  color: rgb(var(--primary-6));
}
.log-data {
  color: var(--color-text-2);
  font-size: 13px;
}
.add-icon {
  cursor: pointer;
  color: var(--color-text-2);
  font-size: 16px;
}
.add-icon:hover {
  color: rgb(var(--primary-6));
}
.extend-submit {
  margin-top: 16px;
}
.extend-table.is-empty :deep(.arco-table-body) {
  display: none;
}
</style>
