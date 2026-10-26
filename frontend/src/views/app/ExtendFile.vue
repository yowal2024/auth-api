<!-- Reconstructed editable Vue source; see README for recovery limits. -->
<script>
import { reportError } from "@/utils/errors.js";
import { keepModalOpen } from "@/utils/modal.js";
import { useAppStore } from "@/stores/app.js";
import {
  listExtendFiles,
  listFileCategories,
  deleteExtendFile,
  uploadExtendFile,
  updateExtendFile,
} from "@/api/app.js";
import { request, rewriteAdminPath } from "@/utils/request.js";
import { formatDateTime } from "@/utils/format.js";
import {
  watch,
  onMounted,
  onBeforeUnmount,
  reactive,
  computed,
  ref,
} from "vue";
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
  Badge,
  Tag,
  Form,
  FormItem,
  Upload,
  Progress,
  Textarea,
  Message,
} from "@arco-design/web-vue";
import {
  IconUpload,
  IconSearch,
  IconDownload,
  IconCopy,
  IconFile,
} from "@arco-design/web-vue/es/icon";
export default {
  name: "ExtendFile",
  components: {
    Button,
    IconUpload,
    Select,
    Option,
    Input,
    IconSearch,
    Table,
    IconFile,
    Tag,
    Badge,
    Space,
    Link,
    IconDownload,
    IconCopy,
    Popconfirm,
    Modal,
    Form,
    FormItem,
    Upload,
    Progress,
    Textarea,
  },
  setup(ao) {
    const appStore = useAppStore();
    const x = reactive({
      category: "",
      keyword: "",
    });
    const rows = ref([]);
    const N = ref([]);
    const loading = ref(false);
    const visible = ref(false);
    const U = ref(false);
    const w = ref(false);
    const saving = ref(false);
    const u = ref(0);
    const g = ref([]);
    const formData = reactive({
      name: "",
      category: "",
      version: "",
      remark: "",
    });
    const s = reactive({
      id: 0,
      name: "",
      category: "",
      version: "",
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
        title: "应用内编号",
        dataIndex: "appRecordId",
        width: 110,
      },
      {
        title: "文件名",
        slotName: "name",
        width: 250,
        ellipsis: true,
        tooltip: true,
      },
      {
        title: "大小",
        slotName: "filesize",
        width: 100,
      },
      {
        title: "分类",
        slotName: "category",
        width: 120,
      },
      {
        title: "版本",
        slotName: "version",
        width: 100,
      },
      {
        title: "下载次数",
        slotName: "downloadCount",
        width: 100,
      },
      {
        title: "上传时间",
        slotName: "createdAt",
        width: 180,
      },
      {
        title: "操作",
        slotName: "operations",
        fixed: "right",
        width: 240,
      },
    ];
    let listSequence = 0;
    onBeforeUnmount(() => listSequence++);
    const handleListExtendFiles = async () => {
      const sequence = ++listSequence;
      loading.value = true;
      try {
        const result = await listExtendFiles({
          appid: appStore.currentId,
          ...x,
        });
        if (sequence === listSequence) rows.value = result;
      } catch (error) {
        reportError(error);
      } finally {
        if (sequence === listSequence) loading.value = false;
      }
    };
    const handleListFileCategories = async () => {
      try {
        N.value = await listFileCategories({
          appid: appStore.currentId,
        });
      } catch (l) {
        console.error("加载分类失败:", l);
      }
    };
    const O = (l) => {
      if (!l || l === 0) return "0 B";
      const e = 1024;
      const c = ["B", "KB", "MB", "GB"];
      const i = Math.floor(Math.log(l) / Math.log(e));
      return Math.round((l / Math.pow(e, i)) * 100) / 100 + " " + c[i];
    };
    const j = (l) => {
      const e = l.toLowerCase();
      return e.includes("zip") || e.includes("rar") || e.includes("7z")
        ? {
            color: "#f7ba1e",
          }
        : e.includes("image") || e.includes("png") || e.includes("jpg")
          ? {
              color: "#00d4aa",
            }
          : e.includes("data") || e.includes("dat") || e.includes("bin")
            ? {
                color: "#3370ff",
              }
            : {
                color: "#86909c",
              };
    };
    const handleClick = () => {
      Object.assign(formData, {
        name: "",
        category: "",
        version: "",
        remark: "",
      });
      g.value = [];
      u.value = 0;
      visible.value = true;
    };
    const handleChange = (l) => {
      g.value = l;
    };
    const handleUploadExtendFile = async () => {
      if (w.value) return;
      if (!g.value[0]?.file) {
        Message.warning("请选择文件");
        return;
      }
      const l = new FormData();
      l.append("file", g.value[0].file);
      l.append("name", formData.name);
      l.append("category", formData.category);
      l.append("version", formData.version);
      l.append("remark", formData.remark);
      w.value = true;
      u.value = 0;
      try {
        await uploadExtendFile(
          {
            appid: appStore.currentId,
          },
          l,
          (e) => {
            if (e.total > 0)
              u.value = Math.min(100, Math.round((e.loaded * 100) / e.total));
          },
        );
        u.value = 100;
        Message.success("上传成功");
        visible.value = false;
        handleListExtendFiles();
        handleListFileCategories();
      } catch (e) {
        reportError(e, "上传失败");
      } finally {
        w.value = false;
        u.value = 0;
      }
    };
    const G = (l) => {
      Object.assign(s, {
        id: l.id,
        name: l.name,
        category: l.category,
        version: l.version,
        remark: l.remark,
      });
      U.value = true;
    };
    const handleUpdateExtendFile = async () => {
      if (saving.value) return;
      saving.value = true;
      try {
        await updateExtendFile(s.id, s);
        Message.success("保存成功");
        U.value = false;
        handleListExtendFiles();
        handleListFileCategories();
      } catch (l) {
        reportError(l, "保存失败");
      } finally {
        saving.value = false;
      }
    };
    const handleDeleteExtendFile = async (l) => {
      try {
        await deleteExtendFile(l.id);
        Message.success("删除成功");
        handleListExtendFiles();
        handleListFileCategories();
      } catch (e) {
        reportError(e, "删除失败");
      }
    };
    const W = async (l) => {
      try {
        const { data: e } = await request(
          `/admin/extend-file/${l.id}/download`,
          {
            responseType: "blob",
            timeout: 300000,
          },
        );
        const c = window.URL.createObjectURL(new Blob([e]));
        const i = document.createElement("a");
        i.href = c;
        i.setAttribute("download", l.filename);
        document.body.appendChild(i);
        i.click();
        document.body.removeChild(i);
        window.URL.revokeObjectURL(c);
      } catch (e) {
        reportError(e, "下载失败");
      }
    };
    const X = (l) => {
      const c = `${window.location.origin}/api${rewriteAdminPath(`/admin/extend-file/${l.id}/download`)}`;
      navigator.clipboard
        .writeText(c)
        .then(() => {
          Message.success("API下载链接已复制（需要登录权限）");
        })
        .catch(() => {
          Message.error("复制失败");
        });
    };
    watch(
      () => appStore.currentId,
      () => {
        x.category = "";
        handleListExtendFiles();
        handleListFileCategories();
      },
    );
    onMounted(() => {
      handleListExtendFiles();
      handleListFileCategories();
    });
    return {
      saving,
      keepModalOpen,
      appStore,
      x,
      N,
      U,
      w,
      u,
      g,
      s,
      tableColumns,
      handleListExtendFiles,
      handleListFileCategories,
      O,
      j,
      handleClick,
      handleChange,
      handleUploadExtendFile,
      G,
      handleUpdateExtendFile,
      handleDeleteExtendFile,
      W,
      X,
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
          <template #icon><IconUpload></IconUpload></template>
          <template #default>上传文件</template>
        </Button>
        <div class="grow"></div>
        <Select
          v-model="x.category"
          placeholder="全部分类"
          allow-clear=""
          :style="{
            width: '150px',
            marginRight: '12px',
          }"
          @change="handleListExtendFiles"
        >
          <template #default>
            <template v-for="o in N" :key="o">
              <Option :value="o">
                <template #default>{{ o }}</template>
              </Option>
            </template>
          </template>
        </Select>
        <Input
          v-model="x.keyword"
          placeholder="关键词搜索..."
          allow-clear=""
          :style="{
            width: '260px',
          }"
          @pressEnter="handleListExtendFiles"
        >
          <template #suffix>
            <IconSearch
              :style="{
                cursor: 'pointer',
              }"
              @click="handleListExtendFiles"
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
        <template #name="{ record: o }">
          <div
            :style="{
              display: 'flex',
              'align-items': 'center',
              gap: '8px',
            }"
          >
            <IconFile
              :style="{
                fontSize: '20px',
                color: j(o.filetype).color,
              }"
            ></IconFile>
            <span>{{ o.name }}</span>
          </div>
        </template>
        <template #filesize="{ record: o }">{{ O(o.filesize) }}</template>
        <template #category="{ record: o }">
          <template v-if="o.category">
            <Tag :key="0" color="arcoblue">
              <template #default>{{ o.category }}</template>
            </Tag>
          </template>
          <template v-else>
            <span
              :key="1"
              :style="{
                color: 'var(--color-text-3)',
              }"
            >
              -
            </span>
          </template>
        </template>
        <template #version="{ record: o }">
          <template v-if="o.version">
            <Tag :key="0" color="green">
              <template #default>{{ o.version }}</template>
            </Tag>
          </template>
          <template v-else>
            <span
              :key="1"
              :style="{
                color: 'var(--color-text-3)',
              }"
            >
              -
            </span>
          </template>
        </template>
        <template #downloadCount="{ record: o }">
          <Badge :count="o.downloadCount" :max-count="999"></Badge>
        </template>
        <template #createdAt="{ record: o }">{{
          formatDateTime(o.createdAt)
        }}</template>
        <template #operations="{ record: o }">
          <Space>
            <template #default>
              <Link @click="(B) => W(o)">
                <template #default>
                  <IconDownload></IconDownload>
                  下载
                </template>
              </Link>
              <Link @click="(B) => G(o)"
                ><template #default>编辑</template></Link
              >
              <Link @click="(B) => X(o)">
                <template #default>
                  <IconCopy></IconCopy>
                  复制链接
                </template>
              </Link>
              <Popconfirm
                content="确定删除？文件将被永久删除"
                @ok="(B) => handleDeleteExtendFile(o)"
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
      title="上传扩展文件"
      ok-text="上传"
      cancel-text="取消"
      :on-before-ok="() => keepModalOpen(handleUploadExtendFile)"
      :width="500"
      :ok-loading="w"
      :on-before-cancel="() => !w"
      :closable="!w"
      :cancel-button-props="{ disabled: w }"
      :mask-closable="!1"
      :esc-to-close="!1"
    >
      <template #default>
        <Form :model="formData" layout="vertical" :disabled="w">
          <template #default>
            <FormItem label="文件">
              <template #default>
                <Upload
                  :auto-upload="!1"
                  :limit="1"
                  :file-list="g"
                  @change="handleChange"
                  accept="*"
                  :disabled="w"
                >
                  <template #upload-button>
                    <Button type="outline" :disabled="w">
                      <template #default>
                        <IconUpload></IconUpload>
                        选择文件
                      </template>
                    </Button>
                  </template>
                </Upload>
                <template v-if="g.length > 0">
                  <div
                    :key="0"
                    :style="{
                      'margin-top': '8px',
                      color: 'var(--color-text-2)',
                    }"
                  >
                    {{ " 大小: " + O(g[0].file?.size || 0) }}
                  </div>
                </template>
                <template v-else></template>
                <template v-if="w">
                  <div
                    :key="1"
                    :style="{
                      'margin-top': '12px',
                    }"
                  >
                    <Progress
                      :percent="u / 100"
                      :status="u === 100 ? 'success' : 'normal'"
                    ></Progress>
                    <div
                      :style="{
                        'margin-top': '4px',
                        'font-size': '12px',
                        color: 'var(--color-text-3)',
                      }"
                    >
                      {{ u === 100 ? "处理中，请稍候..." : `正在上传 ${u}%` }}
                    </div>
                  </div>
                </template>
                <template v-else></template>
              </template>
            </FormItem>
            <FormItem label="显示名称">
              <template #default>
                <Input
                  v-model="formData.name"
                  placeholder="留空则使用文件名"
                ></Input>
              </template>
            </FormItem>
            <FormItem label="分类">
              <template #default>
                <Input
                  v-model="formData.category"
                  placeholder="如: 资源包、配置文件、更新包等"
                ></Input>
              </template>
            </FormItem>
            <FormItem label="版本号">
              <template #default>
                <Input
                  v-model="formData.version"
                  placeholder="如: 1.0.0"
                ></Input>
              </template>
            </FormItem>
            <FormItem label="备注">
              <template #default>
                <Textarea
                  v-model="formData.remark"
                  placeholder="描述此文件的用途"
                  :auto-size="{
                    minRows: 2,
                  }"
                ></Textarea>
              </template>
            </FormItem>
          </template>
        </Form>
      </template>
    </Modal>
    <Modal
      v-model:visible="U"
      title="编辑文件信息"
      ok-text="保存"
      cancel-text="取消"
      :on-before-ok="() => keepModalOpen(handleUpdateExtendFile)"
      :ok-loading="saving"
      :on-before-cancel="() => !saving"
      :closable="!saving"
      :mask-closable="!saving"
      :esc-to-close="!saving"
      :cancel-button-props="{ disabled: saving }"
      :width="500"
    >
      <template #default>
        <Form :model="s" layout="vertical" :disabled="saving">
          <template #default>
            <FormItem label="显示名称">
              <template #default>
                <Input v-model="s.name" placeholder="文件显示名称"></Input>
              </template>
            </FormItem>
            <FormItem label="分类">
              <template #default>
                <Input
                  v-model="s.category"
                  placeholder="如: 资源包、配置文件等"
                ></Input>
              </template>
            </FormItem>
            <FormItem label="版本号">
              <template #default>
                <Input v-model="s.version" placeholder="如: 1.0.0"></Input>
              </template>
            </FormItem>
            <FormItem label="备注">
              <template #default>
                <Textarea
                  v-model="s.remark"
                  placeholder="描述此文件的用途"
                  :auto-size="{
                    minRows: 2,
                  }"
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
.page {
  padding: 20px;
}
.page-card {
  background: var(--color-bg-2);
  border-radius: 4px;
  padding: 20px;
}
.toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}
.grow {
  flex: 1;
}
</style>
