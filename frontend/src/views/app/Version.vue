<!-- Reconstructed editable Vue source; see README for recovery limits. -->
<script>
import { reportError } from "@/utils/errors.js";
import { safeWebURL } from "@/utils/url.js";
import { useAppStore } from "@/stores/app.js";
import {
  listVersions,
  deleteVersion,
  uploadAPK,
  saveVersion,
} from "@/api/app.js";
import { formatDateTime } from "@/utils/format.js";
import { watch, onMounted, reactive, computed, ref } from "vue";
import {
  Button,
  Select,
  Input,
  Table,
  Modal,
  Option,
  Space,
  Link,
  Popconfirm,
  Tag,
  Form,
  FormItem,
  AutoComplete,
  InputNumber,
  Textarea,
  Message,
} from "@arco-design/web-vue";
import {
  IconPlus,
  IconSearch,
  IconUpload,
  IconDelete,
} from "@arco-design/web-vue/es/icon";
export default {
  name: "Version",
  components: {
    Button,
    IconPlus,
    Select,
    Option,
    Input,
    IconSearch,
    Table,
    Tag,
    Link,
    Space,
    Popconfirm,
    Modal,
    Form,
    FormItem,
    AutoComplete,
    InputNumber,
    IconUpload,
    IconDelete,
    Textarea,
  },
  setup(Fe) {
    const appStore = useAppStore();
    const d = reactive({
      keyword: "",
      versionName: "",
    });
    const P = ref([]);
    const loading = ref(false);
    const visible = ref(false);
    const formRef = ref();
    const C = () => ({
      id: 0,
      versionName: "",
      versionKey: "",
      major: 1,
      minor: 0,
      patch: 0,
      status: 1,
      url: "",
      changelog: "",
      apkName: "",
      apkSize: 0,
      apkUrl: "",
      isCanonical: 0,
      canonicalPath: "",
      cloneDiffEnable: 0,
      originalPackage: "",
    });
    const formData = reactive(C());
    const apkInputRef = ref(null);
    const S = computed(
      () => !!(appStore.current?.diffUpdate || appStore.current?.randomPackage),
    );
    const F = computed(() => {
      const n = [];
      return (
        appStore.current?.diffUpdate && n.push("差异更新"),
        appStore.current?.randomPackage && n.push("随机包名"),
        `已开启 ${n.join(" / ")},下载地址会由服务端动态生成`
      );
    });
    const loadingState = ref(false);
    const handleUploadAPK = async (n) => {
      const e = n.target.files?.[0];
      if (((n.target.value = ""), !e)) return;
      if (loadingState.value) return;
      if (
        !e.name.toLowerCase().endsWith(".apk") ||
        e.size > 200 * 1024 * 1024
      ) {
        Message.warning("请选择不超过 200 MB 的 APK 文件");
        return;
      }
      loadingState.value = true;
      const u = Message.loading({
        content: "正在上传 APK...",
        duration: 0,
      });
      try {
        const i = await uploadAPK(e);
        formData.apkName = e.name;
        formData.apkSize = e.size;
        formData.apkUrl = i.url || "";
        formData.canonicalPath = i.path || "";
        formData.originalPackage = i.originalPackage || "";
        formData.isCanonical = 1;
        formData.cloneDiffEnable = 1;
        u.close();
        Message.success(`APK 已上传:${i.originalPackage}`);
      } catch (i) {
        console.error("[apk-upload] failed", i);
        u.close();
        reportError(i, "APK 上传失败");
      } finally {
        loadingState.value = false;
      }
    };
    const handleClick2 = () => {
      formData.apkName = "";
      formData.apkSize = 0;
      formData.apkUrl = "";
      formData.canonicalPath = "";
      formData.originalPackage = "";
      formData.isCanonical = 0;
      formData.cloneDiffEnable = 0;
    };
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
        title: "版本名",
        dataIndex: "versionName",
        width: 140,
        ellipsis: true,
        tooltip: true,
      },
      {
        title: "版本索引",
        dataIndex: "versionKey",
        width: 140,
        ellipsis: true,
        tooltip: true,
      },
      {
        title: "版本号",
        dataIndex: "versionNo",
        width: 120,
      },
      {
        title: "状态",
        slotName: "status",
        width: 90,
      },
      {
        title: "下载地址",
        slotName: "url",
        width: 300,
        ellipsis: true,
        tooltip: true,
      },
      {
        title: "发布时间",
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
    const rows = computed(() => {
      let n = P.value;
      if (
        (d.versionName &&
          (n = n.filter((e) => e.versionName === d.versionName)),
        d.keyword)
      ) {
        const e = d.keyword.toLowerCase();
        n = n.filter(
          (u) =>
            (u.versionName || "").toLowerCase().includes(e) ||
            (u.versionKey || "").toLowerCase().includes(e),
        );
      }
      return n;
    });
    const j = computed(() => {
      const n = new Set(P.value.map((e) => e.versionName).filter(Boolean));
      return Array.from(n);
    });
    const tableRows = j;
    const handleListVersions = async () => {
      loading.value = true;
      try {
        P.value = await listVersions({
          appid: appStore.currentId,
        });
      } finally {
        loading.value = false;
      }
    };
    const handleClick = () => {
      Object.assign(formData, C());
      visible.value = true;
    };
    const W = (n) => {
      Object.assign(formData, C(), n);
      const [e = 1, u = 0, i = 0] = String(n.versionNo || "")
        .split(".")
        .map((w) => Number(w) || 0);
      formData.major = e;
      formData.minor = u;
      formData.patch = i;
      visible.value = true;
    };
    const handleBeforeOk = async () => {
      if (loadingState.value) {
        Message.warning("请等待 APK 上传完成");
        return false;
      }
      if (await formRef.value?.validate()) return false;
      try {
        return await handleSaveVersion();
      } catch {
        return false;
      }
    };
    const handleSaveVersion = async () => {
      if (S.value && !formData.apkName) {
        Message.warning("请上传 APK 文件");
        return false;
      }
      const n = {
        ...formData,
        versionNo: `${formData.major}.${formData.minor}.${formData.patch}`,
        appid: appStore.currentId,
      };
      S.value
        ? (n.url = "")
        : Object.assign(n, {
            apkName: "",
            apkSize: 0,
            apkUrl: "",
            canonicalPath: "",
            originalPackage: "",
            isCanonical: 0,
            cloneDiffEnable: 0,
          });
      delete n.major;
      delete n.minor;
      delete n.patch;
      await saveVersion(n);
      Message.success(formData.id ? "已保存" : "已添加");
      visible.value = false;
      handleListVersions();
      return true;
    };
    const handleDeleteVersion = async (n) => {
      await deleteVersion(n.id);
      Message.success("已删除");
      handleListVersions();
    };
    watch(() => appStore.currentId, handleListVersions);
    onMounted(handleListVersions);
    return {
      safeWebURL,
      appStore,
      d,
      P,
      C,
      S,
      F,
      loadingState,
      handleUploadAPK,
      handleClick2,
      tableColumns,
      j,
      tableRows,
      handleListVersions,
      handleClick,
      W,
      handleBeforeOk,
      handleSaveVersion,
      handleDeleteVersion,
      rows,
      loading,
      pagination,
      visible,
      formRef,
      formData,
      apkInputRef,
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
          <template #default>创建版本</template>
        </Button>
        <div class="grow"></div>
        <Select
          v-model="d.versionName"
          placeholder="全部版本"
          allow-clear=""
          :style="{
            width: '160px',
          }"
        >
          <template #default>
            <template v-for="t in j" :key="t">
              <Option :value="t">
                <template #default>{{ t }}</template>
              </Option>
            </template>
          </template>
        </Select>
        <Input
          v-model="d.keyword"
          placeholder="关键词搜索..."
          allow-clear=""
          :style="{
            width: '220px',
          }"
          @pressEnter="handleListVersions"
          @clear="handleListVersions"
        >
          <template #suffix>
            <IconSearch
              :style="{
                cursor: 'pointer',
              }"
              @click="handleListVersions"
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
        <template #status="{ record: t }">
          <Tag :color="t.status === 1 ? 'green' : 'gray'">
            <template #default>{{ t.status === 1 ? "正常" : "停用" }}</template>
          </Tag>
        </template>
        <template #url="{ record: t }">
          <template v-if="t.isCanonical">
            <Tag
              :key="0"
              color="blue"
              size="small"
              :style="{
                'margin-right': '6px',
              }"
            >
              <template #default>基线</template>
            </Tag>
          </template>
          <template v-else></template>
          <template v-if="t.cloneDiffEnable">
            <Tag
              :key="1"
              color="green"
              size="small"
              :style="{
                'margin-right': '6px',
              }"
            >
              <template #default>差异</template>
            </Tag>
          </template>
          <template v-else></template>
          <template v-if="t.apkName">
            <Tag
              color="arcoblue"
              size="small"
              :style="{
                'margin-right': '6px',
              }"
            >
              <template #default>动态</template>
            </Tag>
            <template v-if="t.originalPackage">
              <span :key="0">{{ t.originalPackage }}</span>
            </template>
            <template v-else></template>
          </template>
          <template v-else>
            <Link
              :key="3"
              :href="safeWebURL(t.url)"
              :disabled="!safeWebURL(t.url)"
              target="_blank"
              rel="noopener noreferrer"
            >
              <template #default>{{ t.url }}</template>
            </Link>
          </template>
        </template>
        <template #createTime="{ record: t }">{{
          formatDateTime(t.createTime)
        }}</template>
        <template #operations="{ record: t }">
          <Space>
            <template #default>
              <Link @click="(re) => W(t)"
                ><template #default>编辑</template></Link
              >
              <Popconfirm
                content="确定删除？"
                @ok="(re) => handleDeleteVersion(t)"
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
      :title="formData.id ? '编辑版本' : '添加版本'"
      :width="460"
      modal-class="ver-modal"
      ok-text="提交"
      cancel-text="取消"
      :on-before-ok="handleBeforeOk"
      :mask-closable="!loadingState"
      :closable="!loadingState"
      :esc-to-close="!loadingState"
      :cancel-button-props="{ disabled: loadingState }"
    >
      <template #default>
        <Form ref="formRef" :model="formData" auto-label-width="">
          <template #default>
            <FormItem
              field="versionName"
              label="版本名"
              :rules="[
                {
                  required: !0,
                  message: '请填写版本名',
                },
              ]"
              asterisk-position="start"
              extra="创建输入新版本名时需选中它或按回车生效"
            >
              <template #default>
                <AutoComplete
                  v-model="formData.versionName"
                  :data="tableRows"
                  placeholder="如: 安卓版"
                  allow-clear=""
                ></AutoComplete>
              </template>
            </FormItem>
            <FormItem
              field="versionKey"
              label="版本索引"
              :rules="[
                {
                  required: !0,
                  message: '请填写版本索引',
                },
              ]"
              asterisk-position="start"
            >
              <template #default>
                <Input
                  v-model="formData.versionKey"
                  placeholder="如: android"
                ></Input>
              </template>
            </FormItem>
            <FormItem label="版本号" asterisk-position="start">
              <template #default>
                <div class="ver-no-row">
                  <InputNumber
                    v-model="formData.major"
                    :min="0"
                    :precision="0"
                    hide-button=""
                  ></InputNumber>
                  <span class="dot">.</span>
                  <InputNumber
                    v-model="formData.minor"
                    :min="0"
                    :precision="0"
                    hide-button=""
                  ></InputNumber>
                  <span class="dot">.</span>
                  <InputNumber
                    v-model="formData.patch"
                    :min="0"
                    :precision="0"
                    hide-button=""
                  ></InputNumber>
                </div>
              </template>
            </FormItem>
            <FormItem label="版本状态">
              <template #default>
                <Select v-model="formData.status">
                  <template #default>
                    <Option :value="1"
                      ><template #default>正常</template></Option
                    >
                    <Option :value="0"
                      ><template #default>停用</template></Option
                    >
                  </template>
                </Select>
              </template>
            </FormItem>
            <template v-if="S">
              <FormItem :key="1" label="上传 APK" :extra="F">
                <template #default>
                  <input
                    ref="apkInputRef"
                    type="file"
                    accept=".apk"
                    :style="{
                      display: 'none',
                    }"
                    @change="handleUploadAPK"
                  />
                  <Space>
                    <template #default>
                      <Button
                        :loading="loadingState"
                        @click="(t) => apkInputRef?.click()"
                      >
                        <template #icon><IconUpload></IconUpload></template>
                        <template #default>
                          {{
                            " " +
                            (formData.apkName ? "重新选择" : "选择 APK 文件")
                          }}
                        </template>
                      </Button>
                      <template v-if="formData.apkName">
                        <Button :key="0" status="danger" @click="handleClick2">
                          <template #icon><IconDelete></IconDelete></template>
                          <template #default>移除</template>
                        </Button>
                      </template>
                      <template v-else></template>
                    </template>
                  </Space>
                </template>
              </FormItem>
            </template>
            <template v-else>
              <FormItem :key="0" label="下载地址">
                <template #default>
                  <Input
                    v-model="formData.url"
                    placeholder="如: http://www.uverif.com/1.0.apk"
                  ></Input>
                </template>
              </FormItem>
            </template>
            <FormItem label="更新提示">
              <template #default>
                <Textarea
                  v-model="formData.changelog"
                  :auto-size="{
                    minRows: 4,
                    maxRows: 10,
                  }"
                  :placeholder="`如: 1、更新了界面
2、优化了用户体验
3、修复了已知BUG`"
                ></Textarea>
              </template>
            </FormItem>
          </template>
        </Form>
      </template>
    </Modal>
  </div>
</template>
