<!-- Reconstructed editable Vue source; see README for recovery limits. -->
<script>
import BackupSettings from "@/components/BackupSettings.vue";
import SecurityEvents from "@/components/SecurityEvents.vue";
import { reportError } from "@/utils/errors.js";
import { keepModalOpen } from "@/utils/modal.js";
import { usePagedList } from "@/composables/usePagedList.js";
import {
  getSystemSettings,
  saveSystemSettings,
  generateHTTPSCertificate,
  saveHTTPSSettings,
  randomizeAPIRoutes,
  saveAPIRoutes,
  randomizeAPICodes,
  saveAPICodes,
  listLogs,
  clearLogs,
  getHTTPSSettings,
  listAPIRoutes,
  listAPICodes,
} from "@/api/system.js";
import { formatDateTime } from "@/utils/format.js";
import { watch, onMounted, ref, reactive, computed } from "vue";
import {
  Breadcrumb,
  Tabs,
  BreadcrumbItem,
  TabPane,
  Form,
  Row,
  Col,
  FormItem,
  Input,
  InputNumber,
  RadioGroup,
  Radio,
  Select,
  Option,
  Button,
  Tag,
  Collapse,
  CollapseItem,
  Textarea,
  Popconfirm,
  InputSearch,
  Table,
  Link,
  Switch,
  Modal,
  Message,
} from "@arco-design/web-vue";
import { useRouter } from "vue-router";
import {
  IconSave,
  IconSafe,
  IconSync,
  IconEdit,
  IconClose,
  IconDelete,
} from "@arco-design/web-vue/es/icon";
import { refreshInstallStatus } from "@/utils/install-status.js";
import { entryPath } from "@/utils/route.js";
export default {
  name: "SystemSettings",
  components: {
BackupSettings,SecurityEvents,
    Breadcrumb,
    BreadcrumbItem,
    Tabs,
    TabPane,
    Form,
    Row,
    Col,
    FormItem,
    Input,
    InputNumber,
    RadioGroup,
    Radio,
    Select,
    Option,
    Button,
    IconSave,
    Tag,
    IconSafe,
    Collapse,
    CollapseItem,
    Textarea,
    IconClose,
    Popconfirm,
    IconSync,
    IconEdit,
    InputSearch,
    Table,
    Switch,
    Link,
    IconDelete,
    Modal,
  },
  setup(Ea) {
    const router = useRouter();
    const D = ref("settings");
    const formData = reactive({
      webUrl: "",
      uploadLimit: 10,
      adminTokenExpire: 604800,
      adminEntry: "admin",
      mysqlMaxConn: 50,
      mysqlMaxIdle: 10,
      cacheEnabled: false,
      userLogEnabled: true,
      adminLogEnabled: true,
      logLevel: "all",
    });
    const loading = ref(false);
    const settingsReady = ref(false);
    const handleGetSystemSettings = async () => {
      try {
        const l = await getSystemSettings();
        Object.assign(formData, l, {
          cacheEnabled: l.cacheEnabled === true || l.cacheEnabled === 1,
          userLogEnabled: l.userLogEnabled === true || l.userLogEnabled === 1,
          adminLogEnabled: l.adminLogEnabled === true || l.adminLogEnabled === 1,
          logLevel: ["all", "warning", "error"].includes(l.logLevel) ? l.logLevel : "all",
        });
        settingsReady.value = true;
      } catch {}
    };
    const handleSaveSystemSettings = async () => {
      if (!settingsReady.value || loading.value) return;
      loading.value = true;
      try {
        const l = router.currentRoute.value.params.entry;
        const e = await saveSystemSettings(formData);
        if (e?.hotSwap === false) Message.warning("设置已保存，但部分配置未即时生效，请重试保存");
        else Message.success("已保存");
        if (e && e.adminEntry && e.adminEntry !== l) {
          await refreshInstallStatus(e.adminEntry);
          const u = router.currentRoute.value.fullPath.replace(`/${l}/`, `/${e.adminEntry}/`);
          router.replace(u);
        }
      } catch (error) {
        reportError(error);
      } finally {
        loading.value = false;
      }
    };
    const s = reactive({
      enabled: false,
      hasCert: false,
      cert: "",
      key: "",
      fingerprint: "",
    });
    const loadingState2 = ref(false);
    const loadingState = ref(false);
    let re = false;
    const handleGetHTTPSSettings = async () => {
      try {
        const l = await getHTTPSSettings();
        s.enabled = !!l.enabled;
        s.hasCert = !!l.hasCert;
        s.fingerprint = l.fingerprint || "";
        s.cert = l.cert || "";
        s.key = "";
        re = true;
      } catch {}
    };
    const handleGenerateHTTPSCertificate = async () => {
      loadingState.value = true;
      try {
        const l = (formData.webUrl || "")
          .replace(/^https?:\/\//, "")
          .split("/")[0]
          .split(":")[0];
        const e = await generateHTTPSCertificate({
          host: l,
        });
        s.hasCert = true;
        s.fingerprint = e.fingerprint || "";
        s.cert = "";
        s.key = "";
        Message.success("已生成自签证书");
      } finally {
        loadingState.value = false;
      }
    };
    const handleSaveHTTPSSettings = async () => {
      if (s.enabled && !s.hasCert && !(s.cert && s.key)) {
        Message.warning("请先生成或上传服务端证书再开启 HTTPS");
        return;
      }
      loadingState2.value = true;
      try {
        const l = {
          enabled: s.enabled,
        };
        s.cert && s.key && ((l.cert = s.cert), (l.key = s.key));
        const e = await saveHTTPSSettings(l);
        e && e.hotSwap === false
          ? Message.error("保存成功,但热切换失败:" + (e.hotSwapError || ""))
          : Message.success("已保存,新连接生效(进程未重启)");
        handleGetHTTPSSettings();
      } finally {
        loadingState2.value = false;
      }
    };
    const j = reactive({
      keyword: "",
    });
    const w = ref([]);
    const loadingState4 = ref(false);
    const loadingState3 = ref(false);
    const x = ref(false);
    const tableColumns = [
      {
        title: "API",
        slotName: "apiKey",
        width: 180,
        ellipsis: true,
        tooltip: true,
      },
      {
        title: "名称",
        dataIndex: "name",
        width: 240,
        ellipsis: true,
        tooltip: true,
      },
      {
        slotName: "rewritten",
        titleSlotName: "rewrittenHeader",
        width: 140,
        align: "center",
      },
      {
        title: "重写后路径",
        slotName: "overridePath",
      },
    ];
    const handleListAPIRoutes = async () => {
      loadingState4.value = true;
      try {
        const l = await listAPIRoutes({});
        w.value = (l || []).map((e) => ({
          ...e,
          rewritten: !!e.rewritten,
          overridePath: e.overridePath || "",
        }));
      } finally {
        loadingState4.value = false;
      }
    };
    const handleClick2 = () => {
      x.value = true;
    };
    const handleClick = () => {
      x.value = false;
      handleListAPIRoutes();
    };
    const handleSaveAPIRoutes = async (l) => {
      if (!x.value)
        try {
          l.rewritten && !l.overridePath && (l.overridePath = J());
          await saveAPIRoutes([
            {
              id: l.id,
              overridePath: l.overridePath || "",
              rewritten: !!l.rewritten,
            },
          ]);
          Message.success(l.rewritten ? "已启用重定义" : "已停用重定义");
        } catch {
          l.rewritten = !l.rewritten;
        }
    };
    const rows = computed(() => {
      const l = (j.keyword || "").trim().toLowerCase();
      return l
        ? w.value.filter(
            (e) =>
              (e.apiKey || "").toLowerCase().includes(l) ||
              (e.name || "").toLowerCase().includes(l) ||
              (e.overridePath || "").toLowerCase().includes(l),
          )
        : w.value;
    });
    const Le = computed(() => w.value.length > 0 && w.value.every((l) => l.rewritten));
    const handleSaveAPIRoutes2 = async (l) => {
      if (loadingState3.value) return;
      const previous = w.value.map((row) => ({ ...row }));
      w.value.forEach((e) => {
        e.rewritten = !!l;
        l && !e.overridePath && (e.overridePath = J());
      });
      if (x.value) return;
      loadingState3.value = true;
      try {
        await saveAPIRoutes(
          w.value.map((e) => ({
            id: e.id,
            overridePath: e.overridePath || "",
            rewritten: !!e.rewritten,
          })),
        );
        Message.success(l ? "已全部启用" : "已全部停用");
      } catch (error) {
        w.value = previous;
        reportError(error);
      } finally {
        loadingState3.value = false;
      }
    };
    const handleSaveAPIRoutes3 = async () => {
      loadingState3.value = true;
      try {
        const l = w.value.map((u) => ({
          id: u.id,
          overridePath: (u.overridePath || "").trim(),
          rewritten: !!u.rewritten,
        }));
        if (l.find((u) => u.rewritten && !u.overridePath)) {
          Message.warning("启用了重定义的行必须填写重写后路径");
          return;
        }
        await saveAPIRoutes(l);
        Message.success("已保存");
        x.value = false;
        handleListAPIRoutes();
      } finally {
        loadingState3.value = false;
      }
    };
    const handleRandomizeAPIRoutes = async () => {
      const l = await randomizeAPIRoutes();
      Message.success(`已为 ${l?.count ?? 0} 条路由生成新路径`);
      handleListAPIRoutes();
    };
    const J = () => {
      const l = crypto.getRandomValues(new Uint8Array(4));
      return Array.from(l, (e) => e.toString(16).padStart(2, "0")).join("");
    };
    const X = reactive({
      keyword: "",
    });
    const C = ref([]);
    const loadingState6 = ref(false);
    const loadingState5 = ref(false);
    const P = ref(false);
    const tableColumns2 = [
      {
        title: "默认 code",
        slotName: "code",
        width: 120,
      },
      {
        title: "消息",
        dataIndex: "message",
        width: 240,
        ellipsis: true,
        tooltip: true,
      },
      {
        slotName: "rewritten",
        titleSlotName: "rewrittenHeader",
        width: 140,
        align: "center",
      },
      {
        title: "重写后 code",
        slotName: "overrideCode",
      },
    ];
    const handleListAPICodes = async () => {
      loadingState6.value = true;
      try {
        const l = await listAPICodes({});
        C.value = (l || []).map((e) => ({
          ...e,
          rewritten: !!e.rewritten,
          overrideCodeText: e.overrideCode ? String(e.overrideCode) : "",
        }));
      } finally {
        loadingState6.value = false;
      }
    };
    const handleClick4 = () => {
      P.value = true;
    };
    const handleClick3 = () => {
      P.value = false;
      handleListAPICodes();
    };
    const handleSaveAPICodes = async (l) => {
      if (!P.value)
        try {
          l.rewritten && !l.overrideCodeText && (l.overrideCodeText = String(te()));
          const e = parseInt(l.overrideCodeText, 10) || 0;
          await saveAPICodes([
            {
              id: l.id,
              overrideCode: e > 0 ? e : 0,
              rewritten: !!l.rewritten,
            },
          ]);
          Message.success(l.rewritten ? "已启用重定义" : "已停用重定义");
        } catch {
          l.rewritten = !l.rewritten;
        }
    };
    const tableRows = computed(() => {
      const l = (X.keyword || "").trim().toLowerCase();
      return l
        ? C.value.filter(
            (e) =>
              String(e.code).includes(l) ||
              (e.overrideCodeText || "").includes(l) ||
              (e.message || "").toLowerCase().includes(l),
          )
        : C.value;
    });
    const Ke = computed(() => C.value.length > 0 && C.value.every((l) => l.rewritten));
    const handleSaveAPICodes2 = async (l) => {
      if (loadingState5.value) return;
      const previous = C.value.map((row) => ({ ...row }));
      C.value.forEach((e) => {
        e.rewritten = !!l;
        l && !e.overrideCodeText && (e.overrideCodeText = String(te()));
      });
      if (P.value) return;
      loadingState5.value = true;
      try {
        await saveAPICodes(
          C.value.map((e) => ({
            id: e.id,
            overrideCode: parseInt(e.overrideCodeText, 10) || 0,
            rewritten: !!e.rewritten,
          })),
        );
        Message.success(l ? "已全部启用" : "已全部停用");
      } catch (error) {
        C.value = previous;
        reportError(error);
      } finally {
        loadingState5.value = false;
      }
    };
    const handleSaveAPICodes3 = async () => {
      loadingState5.value = true;
      try {
        const l = [];
        for (const e of C.value) {
          const raw = e.overrideCodeText.trim();
          const u = Number(raw);
          if ((raw && (!/^\d+$/.test(raw) || !Number.isSafeInteger(u) || u < 1 || u > 2147483647)) || (e.rewritten && !raw)) {
            Message.warning(`code=${e.code} 启用了重定义,请填写正整数`);
            return;
          }
          l.push({
            id: e.id,
            overrideCode: u > 0 ? u : 0,
            rewritten: !!e.rewritten,
          });
        }
        await saveAPICodes(l);
        Message.success("已保存");
        P.value = false;
        handleListAPICodes();
      } catch (error) {
        reportError(error);
      } finally {
        loadingState5.value = false;
      }
    };
    const handleRandomizeAPICodes = async () => {
      const l = await randomizeAPICodes();
      Message.success(`已为 ${l?.count ?? 0} 个状态码生成新码`);
      handleListAPICodes();
    };
    const te = () => 1e3 + Math.floor(Math.random() * 9e3);
    const pageState = reactive({
      type: "system",
      keyword: "",
      page: 1,
      size: 10,
    });
    const tableRows2 = ref([]);
    const z = ref(0);
    const loadingState7 = ref(false);
    const T = ref([]);
    const tableColumns3 = [
      {
        title: "操作类型",
        slotName: "action",
        width: 220,
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
      current: pageState.page,
      pageSize: pageState.size,
      total: z.value,
      showTotal: false,
      showPageSize: true,
      pageSizeOptions: [10, 20, 50, 100],
    }));
    const Qe = computed(() => Math.max(1, Math.ceil(z.value / pageState.size)));
    const handleListLogs = usePagedList(listLogs, pageState, tableRows2, z, loadingState7, T);
    const handlePageChange = (l) => {
      pageState.page = l;
      handleListLogs();
    };
    const handlePageSizeChange = (l) => {
      pageState.size = l;
      pageState.page = 1;
      handleListLogs();
    };
    const visible = ref(false);
    const h = ref(0);
    const handleClick5 = () => {
      h.value = 0;
      visible.value = true;
    };
    const handleClearLogs = async () => {
      if (T.value.length)
        await clearLogs({
          ids: T.value,
          type: "system",
        });
      else {
        const l = h.value > 0 ? Math.floor(Date.now() / 1e3) - h.value * 86400 : 0;
        await clearLogs({
          type: "system",
          before: l,
        });
      }
      Message.success("已清理");
      T.value = [];
      visible.value = false;
      handleListLogs();
    };
    watch(D, (l) => {
      l === "https" && !re && handleGetHTTPSSettings();
      l === "routes" && !w.value.length && handleListAPIRoutes();
      l === "codes" && !C.value.length && handleListAPICodes();
      l === "logs" && handleListLogs();
    });
    onMounted(handleGetSystemSettings);
    return {
      keepModalOpen,
      router,
      D,
      handleGetSystemSettings,
      handleSaveSystemSettings,
      s,
      loadingState2,
      loadingState,
      re,
      handleGetHTTPSSettings,
      handleGenerateHTTPSCertificate,
      handleSaveHTTPSSettings,
      j,
      w,
      loadingState4,
      loadingState3,
      x,
      tableColumns,
      handleListAPIRoutes,
      handleClick2,
      handleClick,
      handleSaveAPIRoutes,
      Le,
      handleSaveAPIRoutes2,
      handleSaveAPIRoutes3,
      handleRandomizeAPIRoutes,
      J,
      X,
      C,
      loadingState6,
      loadingState5,
      P,
      tableColumns2,
      handleListAPICodes,
      handleClick4,
      handleClick3,
      handleSaveAPICodes,
      tableRows,
      Ke,
      handleSaveAPICodes2,
      handleSaveAPICodes3,
      handleRandomizeAPICodes,
      te,
      pageState,
      tableRows2,
      z,
      loadingState7,
      T,
      tableColumns3,
      Qe,
      handleListLogs,
      handlePageChange,
      handlePageSizeChange,
      h,
      handleClick5,
      handleClearLogs,
      formData,
      settingsReady,
      loading,
      rows,
      visible,
      pagination,
      entryPath,
      formatDateTime,
    };
  },
};
</script>
<template>
  <div class="sys-page">
    <Breadcrumb class="crumb">
      <template #default>
        <BreadcrumbItem>
          <template #default>
            <a @click="(o) => $router.push(entryPath('/app'))">应用列表</a>
          </template>
        </BreadcrumbItem>
        <BreadcrumbItem><template #default>系统设置</template></BreadcrumbItem>
      </template>
    </Breadcrumb>
    <div class="page-card">
      <Tabs v-model:active-key="D" type="line" size="large">
        <template #default>
          <TabPane key="settings" title="系统设置">
            <template #default>
              <Form :model="formData" :disabled="!settingsReady || loading" layout="vertical" class="sys-form">
                <template #default>
                  <Row :gutter="24">
                    <template #default>
                      <Col :xs="24" :md="12">
                        <template #default>
                          <FormItem label="域名/IP" extra="用于生成 HTTPS 自签证书的域名，不会修改服务监听地址">
                            <template #default>
                              <Input
                                v-model="formData.webUrl"
                                placeholder="https://user.yowal.cn"
                              ></Input>
                            </template>
                          </FormItem>
                        </template>
                      </Col>
                      <Col :xs="24" :md="12">
                        <template #default>
                          <FormItem label="文件上传限制">
                            <template #default>
                              <InputNumber
                                v-model="formData.uploadLimit"
                                :min="1"
                                :max="200"
                                :step="1"
                              >
                                <template #suffix>MB</template>
                              </InputNumber>
                            </template>
                          </FormItem>
                        </template>
                      </Col>
                      <Col :xs="24" :md="12">
                        <template #default>
                          <FormItem label="管理员Token有效期">
                            <template #default>
                              <InputNumber v-model="formData.adminTokenExpire" :min="60" :step="60">
                                <template #suffix>秒</template>
                              </InputNumber>
                            </template>
                          </FormItem>
                        </template>
                      </Col>
                      <Col :xs="24" :md="12">
                        <template #default>
                          <FormItem label="管理员入口">
                            <template #default>
                              <Input v-model="formData.adminEntry" placeholder="admin"></Input>
                            </template>
                          </FormItem>
                        </template>
                      </Col>
                      <Col :xs="24" :md="12">
                        <template #default>
                          <FormItem label="mysql最大连接数">
                            <template #default>
                              <InputNumber
                                v-model="formData.mysqlMaxConn"
                                :min="1"
                                :max="1e4"
                              ></InputNumber>
                            </template>
                          </FormItem>
                        </template>
                      </Col>
                      <Col :xs="24" :md="12">
                        <template #default>
                          <FormItem label="mysql最大空闲连接数">
                            <template #default>
                              <InputNumber
                                v-model="formData.mysqlMaxIdle"
                                :min="0"
                                :max="1e4"
                              ></InputNumber>
                            </template>
                          </FormItem>
                        </template>
                      </Col>
                      <Col :xs="24" :md="12">
                        <template #default>
                          <FormItem label="系统日志级别" extra="独立控制系统配置操作日志，不受管理员日志开关影响">
                            <template #default>
                              <Select v-model="formData.logLevel">
                                <template #default>
                                  <Option value="all">
                                    <template #default>全部操作(默认)</template>
                                  </Option>
                                  <Option value="warning">
                                    <template #default>仅失败操作</template>
                                  </Option>
                                  <Option value="error">
                                    <template #default>仅服务器错误(5xx)</template>
                                  </Option>
                                </template>
                              </Select>
                            </template>
                          </FormItem>
                        </template>
                      </Col>
                      <Col :xs="24" :md="12">
                        <template #default>
                          <FormItem label="用户日志开关">
                            <template #default>
                              <RadioGroup v-model="formData.userLogEnabled" type="button">
                                <template #default>
                                  <Radio :value="!0"><template #default>开启</template></Radio>
                                  <Radio :value="!1"><template #default>关闭</template></Radio>
                                </template>
                              </RadioGroup>
                            </template>
                          </FormItem>
                        </template>
                      </Col>
                      <Col :xs="24" :md="12">
                        <template #default>
                          <FormItem label="管理员日志开关">
                            <template #default>
                              <RadioGroup v-model="formData.adminLogEnabled" type="button">
                                <template #default>
                                  <Radio :value="!0"><template #default>开启</template></Radio>
                                  <Radio :value="!1"><template #default>关闭</template></Radio>
                                </template>
                              </RadioGroup>
                            </template>
                          </FormItem>
                        </template>
                      </Col>
                      <Col :xs="24" :md="12">
                        <template #default>
                          <FormItem label="数据缓存" extra="缓存在线 IP 定位结果 24 小时；本地 DB-IP 查询不受此开关影响">
                            <template #default>
                              <RadioGroup v-model="formData.cacheEnabled" type="button">
                                <template #default>
                                  <Radio :value="!0"><template #default>开启</template></Radio>
                                  <Radio :value="!1"><template #default>关闭</template></Radio>
                                </template>
                              </RadioGroup>
                            </template>
                          </FormItem>
                        </template>
                      </Col>
                    </template>
                  </Row>
                  <div class="form-actions">
                    <Button type="primary" :loading="loading" :disabled="!settingsReady" @click="handleSaveSystemSettings">
                      <template #icon><IconSave></IconSave></template>
                      <template #default>保存设置</template>
                    </Button>
                  </div>
                </template>
              </Form>
            </template>
          </TabPane>
          <TabPane key="backup" title="自动备份"><BackupSettings v-if="D === 'backup'" /></TabPane>
<TabPane key="security" title="安全事件"><SecurityEvents v-if="D === 'security'" /></TabPane>
<TabPane key="https" title="HTTPS / 双向认证">
            <template #default>
              <Form :model="s" layout="vertical" class="sys-form">
                <template #default>
                  <Row :gutter="24">
                    <template #default>
                      <Col :xs="24" :md="12">
                        <template #default>
                          <FormItem label="HTTPS 开关" extra="开启前请先生成或上传服务端证书。保存后在当前端口切换 HTTP/HTTPS，请使用对应协议重新打开页面。">
                            <template #default>
                              <RadioGroup v-model="s.enabled" type="button">
                                <template #default>
                                  <Radio :value="!0"><template #default>开启</template></Radio>
                                  <Radio :value="!1"><template #default>关闭</template></Radio>
                                </template>
                              </RadioGroup>
                            </template>
                          </FormItem>
                        </template>
                      </Col>
                      <Col :span="24">
                        <template #default>
                          <FormItem label="服务端证书">
                            <template #default>
                              <div class="cert-box">
                                <div class="cert-status">
                                  <template v-if="s.hasCert">
                                    <Tag :key="0" color="green">
                                      <template #default>已配置</template>
                                    </Tag>
                                  </template>
                                  <template v-else>
                                    <Tag :key="1" color="red">
                                      <template #default>未配置</template>
                                    </Tag>
                                  </template>
                                  <template v-if="s.fingerprint">
                                    <span :key="2" class="fp">{{ "指纹 " + s.fingerprint }}</span>
                                  </template>
                                  <template v-else></template>
                                </div>
                                <Button
                                  :loading="loadingState"
                                  @click="handleGenerateHTTPSCertificate"
                                >
                                  <template #icon><IconSafe></IconSafe></template>
                                  <template #default>生成自签证书</template>
                                </Button>
                              </div>
                            </template>
                          </FormItem>
                        </template>
                      </Col>
                      <Col :span="24">
                        <template #default>
                          <Collapse
                            :bordered="!1"
                            class="cert-collapse"
                            expand-icon-position="right"
                          >
                            <template #default>
                              <CollapseItem key="upload" header="上传证书 PEM(可选,替代自签)">
                                <template #default>
                                  <FormItem label="证书 PEM(cert)">
                                    <template #default>
                                      <Textarea
                                        v-model="s.cert"
                                        :auto-size="{
                                          minRows: 4,
                                          maxRows: 8,
                                        }"
                                        placeholder="-----BEGIN CERTIFICATE-----"
                                      ></Textarea>
                                    </template>
                                  </FormItem>
                                  <FormItem label="私钥 PEM(key,仅上传时填写,不回显)">
                                    <template #default>
                                      <Textarea
                                        v-model="s.key"
                                        :auto-size="{
                                          minRows: 4,
                                          maxRows: 8,
                                        }"
                                        placeholder="-----BEGIN PRIVATE KEY-----"
                                      ></Textarea>
                                    </template>
                                  </FormItem>
                                </template>
                              </CollapseItem>
                            </template>
                          </Collapse>
                        </template>
                      </Col>
                    </template>
                  </Row>
                  <div class="form-actions">
                    <Button
                      type="primary"
                      :loading="loadingState2"
                      @click="handleSaveHTTPSSettings"
                    >
                      <template #icon><IconSave></IconSave></template>
                      <template #default>保存</template>
                    </Button>
                  </div>
                </template>
              </Form>
            </template>
          </TabPane>
          <TabPane key="routes" title="API路由">
            <template #default>
              <div class="toolbar">
                <div class="btn-group">
                  <template v-if="x">
                    <Button type="primary" :loading="loadingState3" @click="handleSaveAPIRoutes3">
                      <template #icon><IconSave></IconSave></template>
                      <template #default>保存</template>
                    </Button>
                    <Button @click="handleClick">
                      <template #icon><IconClose></IconClose></template>
                      <template #default>取消</template>
                    </Button>
                  </template>
                  <template v-else>
                    <Popconfirm
                      content="将为所有路由重新生成随机路径并启用,确定继续?"
                      @ok="handleRandomizeAPIRoutes"
                    >
                      <template #default>
                        <Button type="primary">
                          <template #icon><IconSync></IconSync></template>
                          <template #default>一键重定义</template>
                        </Button>
                      </template>
                    </Popconfirm>
                    <Button @click="handleClick2">
                      <template #icon><IconEdit></IconEdit></template>
                      <template #default>自定义重写</template>
                    </Button>
                  </template>
                </div>
                <div class="grow"></div>
                <InputSearch
                  v-model="j.keyword"
                  placeholder="关键词搜索..."
                  allow-clear=""
                  :style="{
                    width: '260px',
                  }"
                ></InputSearch>
              </div>
              <Table
                :data="rows"
                :columns="tableColumns"
                :loading="loadingState4"
                :pagination="!1"
                row-key="id"
                :scroll="{
                  y: 'calc(100vh - 320px)',
                }"
              >
                <template #apiKey="{ record: o }">
                  <span class="api-key">{{ o.apiKey }}</span>
                </template>
                <template #rewrittenHeader>
                  <span class="head-toggle">
                    重定义
                    <Switch :model-value="Le" :disabled="loadingState3" size="small" @change="handleSaveAPIRoutes2"></Switch>
                  </span>
                </template>
                <template #rewritten="{ record: o }">
                  <Switch
                    v-model="o.rewritten"
                    :disabled="loadingState3"
                    size="small"
                    @change="(y) => handleSaveAPIRoutes(o)"
                  ></Switch>
                </template>
                <template #overridePath="{ record: o }">
                  <template v-if="x">
                    <Input
                      :key="0"
                      v-model="o.overridePath"
                      placeholder="可自定义或一键重定义"
                      allow-clear=""
                    >
                      <template #append>
                        <Link @click="(y) => (o.overridePath = J())">
                          <template #default>随机</template>
                        </Link>
                      </template>
                    </Input>
                  </template>
                  <template v-else>
                    <template v-if="o.overridePath">
                      <span :key="1" class="path-text">{{ o.overridePath }}</span>
                    </template>
                    <template v-else></template>
                  </template>
                </template>
              </Table>
            </template>
          </TabPane>
          <TabPane key="codes" title="API状态码">
            <template #default>
              <div class="toolbar">
                <div class="btn-group">
                  <template v-if="P">
                    <Button type="primary" :loading="loadingState5" @click="handleSaveAPICodes3">
                      <template #icon><IconSave></IconSave></template>
                      <template #default>保存</template>
                    </Button>
                    <Button @click="handleClick3">
                      <template #icon><IconClose></IconClose></template>
                      <template #default>取消</template>
                    </Button>
                  </template>
                  <template v-else>
                    <Popconfirm
                      content="将为所有状态码重新生成 4 位随机码并启用,确定继续?"
                      @ok="handleRandomizeAPICodes"
                    >
                      <template #default>
                        <Button type="primary">
                          <template #icon><IconSync></IconSync></template>
                          <template #default>一键重定义</template>
                        </Button>
                      </template>
                    </Popconfirm>
                    <Button @click="handleClick4">
                      <template #icon><IconEdit></IconEdit></template>
                      <template #default>自定义重写</template>
                    </Button>
                  </template>
                </div>
                <div class="grow"></div>
                <InputSearch
                  v-model="X.keyword"
                  placeholder="关键词搜索..."
                  allow-clear=""
                  :style="{
                    width: '260px',
                  }"
                ></InputSearch>
              </div>
              <Table
                :data="tableRows"
                :columns="tableColumns2"
                :loading="loadingState6"
                :pagination="!1"
                row-key="id"
                :scroll="{
                  y: 'calc(100vh - 320px)',
                }"
              >
                <template #code="{ record: o }">
                  <span>{{ o.code }}</span>
                </template>
                <template #rewrittenHeader>
                  <span class="head-toggle">
                    重定义
                    <Switch :model-value="Ke" :disabled="loadingState5" size="small" @change="handleSaveAPICodes2"></Switch>
                  </span>
                </template>
                <template #rewritten="{ record: o }">
                  <Switch
                    v-model="o.rewritten"
                    size="small"
                    @change="(y) => handleSaveAPICodes(o)"
                    :disabled="loadingState5"
                  ></Switch>
                </template>
                <template #overrideCode="{ record: o }">
                  <template v-if="P">
                    <Input
                      :key="0"
                      v-model="o.overrideCodeText"
                      placeholder="可自定义或一键重定义"
                      allow-clear=""
                    >
                      <template #append>
                        <Link @click="(y) => (o.overrideCodeText = String(te()))">
                          <template #default>随机</template>
                        </Link>
                      </template>
                    </Input>
                  </template>
                  <template v-else>
                    <template v-if="o.overrideCodeText">
                      <span :key="1" class="path-text">{{ o.overrideCodeText }}</span>
                    </template>
                    <template v-else></template>
                  </template>
                </template>
              </Table>
            </template>
          </TabPane>
          <TabPane key="logs" title="系统日志">
            <template #default>
              <div class="toolbar">
                <Button type="primary" @click="handleClick5">
                  <template #icon><IconDelete></IconDelete></template>
                  <template #default>清理日志</template>
                </Button>
                <div class="grow"></div>
                <InputSearch
                  v-model="pageState.keyword"
                  placeholder="类型搜索..."
                  allow-clear=""
                  :style="{
                    width: '260px',
                  }"
                  @search="handleListLogs"
                  @pressEnter="handleListLogs"
                ></InputSearch>
              </div>
              <Modal
                v-model:visible="visible"
                title="清理日志"
                :on-before-ok="() => keepModalOpen(handleClearLogs)"
                :width="360"
                modal-class="clear-modal"
              >
                <template #default>
                  <div class="clear-row">
                    <span class="clear-label">时间</span>
                    <Select
                      v-model="h"
                      placeholder="全部"
                      :style="{
                        flex: '1',
                      }"
                    >
                      <template #default>
                        <Option :value="0"><template #default>全部</template></Option>
                        <Option :value="7"><template #default>7天前</template></Option>
                        <Option :value="15"><template #default>15天前</template></Option>
                        <Option :value="30"><template #default>30天前</template></Option>
                        <Option :value="90"><template #default>90天前</template></Option>
                      </template>
                    </Select>
                  </div>
                  <template v-if="T.length">
                    <div
                      :key="0"
                      class="hint"
                      :style="{
                        'margin-top': '8px',
                      }"
                    >
                      {{ " 注:已勾选 " + T.length + " 条,将忽略时间范围,直接清理选中项。 " }}
                    </div>
                  </template>
                  <template v-else></template>
                </template>
              </Modal>
              <Table
                :data="tableRows2"
                :columns="tableColumns3"
                :loading="loadingState7"
                :pagination="pagination"
                :row-selection="{
                  type: 'checkbox',
                  showCheckedAll: !0,
                }"
                v-model:selectedKeys="T"
                row-key="id"
                @pageChange="handlePageChange"
                @pageSizeChange="handlePageSizeChange"
              >
                <template #action="{ record: o }">
                  <div class="log-cell">
                    <div class="log-title">{{ o.summary || o.action }}</div>
                    <div class="log-sub">{{ o.operator }}</div>
                  </div>
                </template>
                <template #ip="{ record: o }">
                  <div class="log-cell log-cell-right">
                    <div class="log-title">{{ o.region || "-" }}</div>
                    <div class="log-sub">{{ o.ip }}</div>
                  </div>
                </template>
                <template #createTime="{ record: o }">{{ formatDateTime(o.createTime) }}</template>
              </Table>
              <div class="log-footer">
                <span class="page-info">
                  {{ " 当前第 " + pageState.page + " 页 共 " + Qe + " 页 " + z + " 条结果 " }}
                </span>
              </div>
            </template>
          </TabPane>
        </template>
      </Tabs>
    </div>
  </div>
</template>
<style scoped>
.sys-page {
  padding: 16px 24px;
}
.crumb {
  margin-bottom: 12px;
}
.page-card {
  background: var(--color-bg-2);
  border-radius: 6px;
  padding: 4px 20px 20px;
  box-shadow: 0 1px 2px #0000000a;
}
.sys-form {
  padding-top: 12px;
}
.sys-form :deep(.arco-form-item-label) {
  font-weight: 500;
}
.sys-form :deep(.arco-input-number) {
  width: 100%;
}
.api-key {
  font-weight: 500;
  color: var(--color-text-1);
}
.head-toggle {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}
.form-actions {
  margin-top: 8px;
}
.cert-box {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
}
.cert-status {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}
.cert-status .fp {
  color: var(--color-text-3);
  font-size: 12px;
  word-break: break-all;
  min-width: 0;
  flex: 1 1 240px;
}
.cert-collapse :deep(.arco-collapse-item-header) {
  padding-left: 0;
}
.toolbar {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
  flex-wrap: wrap;
  align-items: center;
  padding-top: 12px;
}
.toolbar .grow {
  flex: 1;
}
.toolbar .btn-group {
  display: flex;
  gap: 8px;
  flex-wrap: nowrap;
  flex-shrink: 0;
}
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
.log-cell .log-sub,
.hint {
  color: var(--color-text-3);
  font-size: 12px;
}
.clear-row {
  display: flex;
  align-items: center;
  gap: 12px;
}
.clear-label {
  color: var(--color-text-1);
  font-size: 14px;
  white-space: nowrap;
}
.log-footer {
  margin-top: 12px;
  color: var(--color-text-3);
  font-size: 13px;
}
.log-footer .page-info {
  display: inline-block;
}
@media (max-width: 768px) {
  .sys-page {
    padding: 12px;
  }
  .page-card {
    padding: 4px 12px 12px;
  }
  .toolbar .grow {
    display: none;
  }
  .toolbar .btn-group {
    width: 100%;
  }
  .toolbar .arco-input-search {
    flex: 1 1 100%;
    width: 100% !important;
  }
  :deep(.arco-table-container) {
    overflow-x: auto;
  }
  :deep(.arco-table-element) {
    min-width: 640px;
  }
}
</style>
