<!-- Reconstructed editable Vue source; see README for recovery limits. -->
<script>
import { reportError } from "@/utils/errors.js";
import { onMounted, onUnmounted, reactive, ref, computed } from "vue";
import {
  Button,
  Select,
  Input,
  Spin,
  Pagination,
  Modal,
  Option,
  Card,
  Popconfirm,
  Space,
  Tag,
  Empty,
  Form,
  FormItem,
  Upload,
  Switch,
  InputGroup,
  Message,
} from "@arco-design/web-vue";
import { useRouter } from "vue-router";
import {
  IconEye,
  IconEyeInvisible,
  IconPlus,
  IconSearch,
  IconCopy,
  IconDelete,
} from "@arco-design/web-vue/es/icon";
import { useAppStore } from "@/stores/app.js";
import { useUserStore } from "@/stores/user.js";
import { deleteApp, updateApp, createApp } from "@/api/app.js";
import { uploadAppLogo } from "@/api/auth.js";
import { defaultAvatar } from "@/utils/avatar.js";
import { entryPath } from "@/utils/route.js";
export default {
  name: "AppList",
  components: {
    IconEye,
    IconEyeInvisible,
    Button,
    IconPlus,
    Select,
    Option,
    Input,
    IconSearch,
    Spin,
    Card,
    Space,
    Tag,
    Popconfirm,
    IconDelete,
    IconCopy,
    Empty,
    Pagination,
    Modal,
    Form,
    FormItem,
    Upload,
    Switch,
    InputGroup,
  },
  setup(wt, { expose: N }) {
    const router = useRouter();
    const appStore = useAppStore();
    const userStore = useUserStore();
    const isSuper = computed(() => userStore.profile?.role === "super");
    const L = ref([]);
    const loading = ref(false);
    const revealedKeys = reactive({});
    const isMobile = ref(false);
    const updateMobile = () => {
      isMobile.value = window.innerWidth < 992;
    };
    const keyDialog = reactive({
      visible: false,
      name: "",
      key: "",
    });
    const filters = reactive({
      keyword: "",
      type: void 0,
      page: 1,
      size: 12,
    });
    const visible = ref(false);
    const saving = ref(false);
    const formRef = ref();
    const formData = reactive({
      id: 0,
      name: "",
      type: "user",
      inheritFrom: void 0,
      logo: "",
      pkg: "",
      status: 1,
      tokenKey: "",
    });
    const appTypeLabels = {
      user: "用户版",
      kami: "卡密版",
    };
    const appTypeLabel = (s) => appTypeLabels[s] || "用户版";
    const filteredApps = computed(() => {
      let s = L.value;
      if (
        (filters.type && (s = s.filter((e) => e.type === filters.type)),
        filters.keyword)
      ) {
        const e = filters.keyword.toLowerCase();
        s = s.filter(
          (f) =>
            String(f.id).includes(e) ||
            (f.name || "").toLowerCase().includes(e),
        );
      }
      return s;
    });
    const pageCount = computed(() =>
      Math.max(1, Math.ceil(filteredApps.value.length / filters.size)),
    );
    const pageApps = computed(() => {
      const s = (filters.page - 1) * filters.size;
      return filteredApps.value.slice(s, s + filters.size);
    });
    const loadApps = async () => {
      loading.value = true;
      try {
        L.value = await appStore.fetchApps();
        filters.page = Math.min(filters.page, pageCount.value);
      } finally {
        loading.value = false;
      }
    };
    const resetPage = () => {
      filters.page = 1;
    };
    const openCreate = () => {
      Object.keys(formData).forEach((key) => delete formData[key]);
      Object.assign(formData, {
        id: 0,
        name: "",
        type: "user",
        inheritFrom: void 0,
        logo: "",
        pkg: "",
        status: 1,
        tokenKey: "",
      });
      visible.value = true;
    };
    const openEdit = (s) => {
      Object.keys(formData).forEach((key) => delete formData[key]);
      Object.assign(
        formData,
        {
          id: 0,
          name: "",
          type: "user",
          inheritFrom: void 0,
          logo: "",
          pkg: "",
          status: 1,
          tokenKey: "",
        },
        s,
      );
      visible.value = true;
    };
    const saveApp = async () => {
      if (saving.value) return;
      if (!(await formRef.value?.validate())) {
        saving.value = true;
        try {
          formData.id
            ? await updateApp(formData.id, formData)
            : await createApp(formData);
          Message.success(formData.id ? "已保存" : "已添加");
          visible.value = false;
          loadApps();
        } finally {
          saving.value = false;
        }
      }
    };
    const removeApp = async (s) => {
      await deleteApp(s.id);
      Message.success("已删除");
      loadApps();
    };
    const openApp = (s) => {
      appStore.setCurrent(s.id);
      router.push(entryPath("/dashboard"));
    };
    const toggleKey = (s) => {
      revealedKeys[s] = !revealedKeys[s];
    };
    const showKey = (s) => {
      isMobile.value
        ? ((keyDialog.name = s.name),
          (keyDialog.key = s.tokenKey),
          (keyDialog.visible = true))
        : toggleKey(s.id);
    };
    const copyKey = async (s) => {
      if (s)
        try {
          if (navigator.clipboard && window.isSecureContext)
            await navigator.clipboard.writeText(s);
          else {
            const e = document.createElement("textarea");
            e.value = s;
            e.style.cssText = "position:fixed;opacity:0";
            document.body.appendChild(e);
            e.select();
            document.execCommand("copy");
            document.body.removeChild(e);
          }
          Message.success("已复制 Key");
        } catch {
          Message.error("复制失败");
        }
    };
    const uploadLogo = async (s, e) => {
      const f = e?.file;
      if (f)
        try {
          const { url: u } = await uploadAppLogo(f);
          formData.logo = u;
          Message.success("logo 上传成功,保存后生效");
        } catch (u) {
          reportError(u, "上传失败");
        }
    };
    const generateTokenKey = () => {
      formData.tokenKey = Array.from(crypto.getRandomValues(new Uint8Array(16)))
        .map((s) => s.toString(16).padStart(2, "0"))
        .join("");
    };
    N({
      onEdit: openEdit,
    });
    onMounted(() => {
      loadApps();
      updateMobile();
      window.addEventListener("resize", updateMobile);
    });
    onUnmounted(() => window.removeEventListener("resize", updateMobile));
    return {
      isSuper,
      N,
      router,
      appStore,
      L,
      revealedKeys,
      isMobile,
      updateMobile,
      keyDialog,
      filters,
      saving,
      appTypeLabels,
      appTypeLabel,
      filteredApps,
      pageCount,
      pageApps,
      loadApps,
      resetPage,
      openCreate,
      openEdit,
      saveApp,
      removeApp,
      openApp,
      toggleKey,
      showKey,
      copyKey,
      uploadLogo,
      generateTokenKey,
      loading,
      visible,
      formRef,
      formData,
      defaultAvatar,
    };
  },
};
</script>
<template>
  <div class="app-list-page">
    <div class="toolbar">
      <Button v-if="isSuper" type="primary" class="add-btn" @click="openCreate">
        <template #icon><IconPlus></IconPlus></template>
        <template #default>添加应用</template>
      </Button>
      <div class="form-row">
        <Select
          v-model="filters.type"
          placeholder="应用类型"
          allow-clear=""
          class="type-sel"
          @change="resetPage"
        >
          <template #default>
            <Option value="user"><template #default>用户版</template></Option>
            <Option value="kami"><template #default>卡密版</template></Option>
          </template>
        </Select>
        <Input
          v-model="filters.keyword"
          placeholder="搜索应用名称或APPID"
          allow-clear=""
          class="search-input"
          @pressEnter="resetPage"
        >
          <template #suffix>
            <IconSearch
              :style="{
                cursor: 'pointer',
              }"
              @click="resetPage"
            ></IconSearch>
          </template>
        </Input>
      </div>
    </div>
    <Spin
      :loading="loading"
      :style="{
        width: '100%',
      }"
    >
      <template #default>
        <template v-if="pageApps.length">
          <div :key="0" class="card-grid">
            <template v-for="t in pageApps" :key="t.id">
              <Card class="app-card" :bordered="!0" size="medium">
                <template #title>
                  <Space size="small" align="center">
                    <template #default>
                      <span
                        :class="['status-dot', t.status === 1 ? 'on' : 'off']"
                      ></span>
                      <span class="app-name">{{ t.name }}</span>
                      <Tag checkable="" :default-checked="!0">
                        <template #default>{{ appTypeLabel(t.type) }}</template>
                      </Tag>
                    </template>
                  </Space>
                </template>
                <template v-if="isSuper" #extra>
                  <Popconfirm
                    :content="`确认删除应用：${t.name} ？`"
                    @ok="(I) => removeApp(t)"
                  >
                    <template #default>
                      <button class="del-btn">
                        <IconDelete></IconDelete>
                      </button>
                    </template>
                  </Popconfirm>
                </template>
                <template #default>
                  <div class="card-body-row">
                    <div class="card-body-left">
                      <div class="logo-box">
                        <img :src="t.logo || defaultAvatar" alt="logo" />
                      </div>
                      <div class="info">
                        <p class="info-line">{{ "APPID:" + t.id }}</p>
                        <div v-if="isSuper" class="info-line key-line">
                          <span class="key-text">
                            {{
                              "Key:" +
                              (revealedKeys[t.id]
                                ? t.tokenKey
                                : "****************")
                            }}
                          </span>
                          <button class="eye-btn" @click="(I) => showKey(t)">
                            <component
                              :is='!isMobile &amp;&amp; revealedKeys[t.id] ? "icon-eye-invisible" : "icon-eye"'
                            ></component>
                          </button>
                          <template v-if="isMobile"></template>
                          <template v-else>
                            <button
                              :key="0"
                              class="eye-btn"
                              @click="(I) => copyKey(t.tokenKey)"
                            >
                              <IconCopy></IconCopy>
                            </button>
                          </template>
                        </div>
                      </div>
                    </div>
                    <Button
                      type="primary"
                      size="small"
                      class="enter-btn"
                      @click="(I) => openApp(t)"
                    >
                      <template #default>进入</template>
                    </Button>
                  </div>
                </template>
              </Card>
            </template>
          </div>
        </template>
        <template v-else>
          <Empty
            :key="1"
            :style="{
              padding: '80px 0',
            }"
          ></Empty>
        </template>
      </template>
    </Spin>
    <div class="footer-bar">
      <span class="page-info">
        {{
          " 当前第 " +
          filters.page +
          " 页 共 " +
          pageCount +
          " 页 " +
          filteredApps.length +
          " 条结果 "
        }}
      </span>
      <Pagination
        v-model:current="filters.page"
        :total="filteredApps.length"
        :page-size="filters.size"
        @change="(t) => (filters.page = t)"
      ></Pagination>
    </div>
    <Modal
      v-model:visible="visible"
      :title="formData.id ? '编辑应用' : '添加应用'"
      width="auto"
      :mask-closable="!1"
      modal-class="app-modal"
    >
      <template #footer>
        <Button @click="(t) => (visible = !1)"
          ><template #default>取消</template></Button
        >
        <Button type="primary" :loading="saving" @click="saveApp">
          <template #default>{{ formData.id ? "保存" : "确定添加" }}</template>
        </Button>
      </template>
      <template #default>
        <Form ref="formRef" :model="formData" auto-label-width="">
          <template #default>
            <FormItem
              field="name"
              label="应用名称"
              :rules="[
                {
                  required: !0,
                  message: '请填写应用名称',
                },
              ]"
              asterisk-position="start"
            >
              <template #default>
                <Input
                  v-model="formData.name"
                  placeholder="请填写应用名称"
                ></Input>
              </template>
            </FormItem>
            <FormItem
              field="type"
              label="应用类型"
              :rules="[
                {
                  required: !0,
                  message: '请选择应用类型',
                },
              ]"
              asterisk-position="start"
            >
              <template #default>
                <Select v-model="formData.type">
                  <template #default>
                    <Option value="user"
                      ><template #default>用户版</template></Option
                    >
                    <Option value="kami"
                      ><template #default>卡密版</template></Option
                    >
                  </template>
                </Select>
              </template>
            </FormItem>
            <template v-if="formData.id"></template>
            <template v-else>
              <FormItem :key="0" field="inheritFrom" label="配置继承">
                <template #default>
                  <Select
                    v-model="formData.inheritFrom"
                    placeholder="可选"
                    allow-clear=""
                  >
                    <template #default>
                      <template v-for="t in L" :key="t.id">
                        <Option :value="t.id">
                          <template #default>{{ t.name }}</template>
                        </Option>
                      </template>
                    </template>
                  </Select>
                </template>
              </FormItem>
            </template>
            <FormItem label="LOGO">
              <template #default>
                <Upload
                  :show-file-list="!1"
                  :auto-upload="!1"
                  list-type="picture-card"
                  accept="image/*"
                  @change="uploadLogo"
                >
                  <template #upload-button>
                    <div class="logo-upload">
                      <template v-if="formData.logo">
                        <img :key="0" :src="formData.logo" />
                      </template>
                      <template v-else><IconPlus :key="1"></IconPlus></template>
                    </div>
                  </template>
                </Upload>
              </template>
            </FormItem>
            <template v-if="formData.id">
              <FormItem :key="1" field="pkg" label="包名">
                <template #default>
                  <Input
                    v-model="formData.pkg"
                    placeholder="com.example.app"
                  ></Input>
                </template>
              </FormItem>
            </template>
            <template v-else></template>
            <template v-if="formData.id">
              <FormItem :key="2" label="状态">
                <template #default>
                  <Switch
                    v-model="formData.status"
                    :checked-value="1"
                    :unchecked-value="0"
                  ></Switch>
                </template>
              </FormItem>
            </template>
            <template v-else></template>
            <template v-if="formData.id">
              <FormItem :key="3" label="Key">
                <template #default>
                  <InputGroup>
                    <template #default>
                      <Input v-model="formData.tokenKey" readonly=""></Input>
                      <Button @click="generateTokenKey"
                        ><template #default>重置</template></Button
                      >
                    </template>
                  </InputGroup>
                </template>
              </FormItem>
            </template>
            <template v-else></template>
          </template>
        </Form>
      </template>
    </Modal>
    <Modal
      v-model:visible="keyDialog.visible"
      :title="`${keyDialog.name} · Key`"
      :footer="!1"
      modal-class="key-modal"
    >
      <template #default>
        <div class="key-full">{{ keyDialog.key }}</div>
        <Button
          type="primary"
          long=""
          class="key-copy-btn"
          @click="(t) => copyKey(keyDialog.key)"
        >
          <template #icon><IconCopy></IconCopy></template>
          <template #default>复制 Key</template>
        </Button>
      </template>
    </Modal>
  </div>
</template>
<style scoped>
.app-list-page {
  padding: 24px 32px;
  min-height: 100%;
  background: var(--content-bg, #f5f7fa);
}
.toolbar {
  display: flex;
  flex-direction: column;
  gap: 16px;
  margin-bottom: 20px;
}
.toolbar .add-btn {
  align-self: flex-start;
  flex-shrink: 0;
  width: auto;
}
.toolbar .form-row {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.toolbar .form-row > * {
  width: 100%;
}
@media (min-width: 768px) {
  .toolbar {
    flex-direction: row;
    align-items: center;
    flex-wrap: wrap;
  }
  .toolbar .form-row {
    flex: 1;
    flex-direction: row;
    justify-content: flex-end;
    width: auto;
    gap: 12px;
  }
  .toolbar .form-row > * {
    width: auto;
  }
  .toolbar .form-row .type-sel {
    flex: 1 1 0;
    max-width: 200px;
    min-width: 140px;
  }
  .toolbar .form-row .search-input {
    flex: 2 1 0;
    max-width: 420px;
    min-width: 200px;
  }
}
.card-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 14px;
}
@media (min-width: 768px) {
  .card-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
@media (min-width: 1024px) {
  .card-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}
.app-card :deep(.arco-card-body) {
  padding: 14px 16px;
}
.app-card :deep(.arco-card-header) {
  height: auto;
  padding: 10px 16px;
  border-bottom: 1px solid var(--color-neutral-3);
}
.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  display: inline-block;
}
.status-dot.on {
  background: #00b42a;
}
.status-dot.off {
  background: #c9cdd4;
}
.app-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text-1);
}
.del-btn {
  background: transparent;
  border: none;
  cursor: pointer;
  color: var(--color-text-2);
  padding: 2px 4px;
}
.del-btn:hover {
  color: rgb(var(--danger-6));
}
.card-body-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
.enter-btn {
  flex-shrink: 0;
}
.card-body-left {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}
.logo-box {
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.logo-box img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}
.info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}
.info .info-line {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-1);
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
}
.info .key-line .key-text {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.info .eye-btn {
  flex-shrink: 0;
  background: transparent;
  border: none;
  cursor: pointer;
  padding: 0;
  color: var(--color-text-3);
  display: inline-flex;
  align-items: center;
}
.info .eye-btn:hover {
  color: rgb(var(--primary-6));
}
.footer-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 24px;
}
.footer-bar .page-info {
  color: var(--color-text-3);
  font-size: 13px;
}
</style>
