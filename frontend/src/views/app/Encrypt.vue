<!-- Reconstructed editable Vue source; see README for recovery limits. -->
<script>
import { reportError } from "@/utils/errors.js";
import { useAppStore } from "@/stores/app.js";
import {
  listEncryptions,
  deleteEncryption,
  saveEncryption,
  generateRSAKey,
} from "@/api/app.js";
import { watch, onMounted, reactive, ref, computed } from "vue";
import {
  Button,
  Input,
  Spin,
  Pagination,
  Modal,
  Dropdown,
  Doption,
  Tag,
  Switch,
  Empty,
  Form,
  FormItem,
  Select,
  Option,
  Textarea,
  InputNumber,
  Message,
} from "@arco-design/web-vue";
import {
  IconPlus,
  IconSearch,
  IconMoreVertical,
} from "@arco-design/web-vue/es/icon";
export default {
  name: "Encrypt",
  components: {
    Button,
    IconPlus,
    Input,
    IconSearch,
    Spin,
    Dropdown,
    Doption,
    IconMoreVertical,
    Tag,
    Switch,
    Empty,
    Pagination,
    Modal,
    Form,
    FormItem,
    Select,
    Option,
    Textarea,
    InputNumber,
  },
  setup(Xe) {
    const appStore = useAppStore();
    const x = ref([]);
    const loading = ref(false);
    const i = reactive({
      keyword: "",
      page: 1,
      size: 12,
    });
    const visible = ref(false);
    const formRef = ref();
    const pageState = ref(false);
    const A = () => ({
      id: 0,
      name: "",
      algorithm: "aes",
      clientPrivateKey: "",
      clientPublicKey: "",
      serverPrivateKey: "",
      serverPublicKey: "",
      encoding: "base64",
      symKey: "",
      signMode: "verify",
      timeSkew: 60,
    });
    const formData = reactive(A());
    const w = computed(() => {
      if (!i.keyword) return x.value;
      const s = i.keyword.toLowerCase();
      return x.value.filter((e) => (e.name || "").toLowerCase().includes(s));
    });
    const G = computed(() => Math.max(1, Math.ceil(w.value.length / i.size)));
    const D = computed(() => {
      const s = (i.page - 1) * i.size;
      return w.value.slice(s, s + i.size);
    });
    const handleListEncryptions = async () => {
      loading.value = true;
      try {
        x.value = await listEncryptions({
          appid: appStore.currentId,
        });
      } finally {
        loading.value = false;
      }
    };
    const handlePressEnter = () => {
      i.page = 1;
    };
    const handleClick = () => {
      pageState.value = true;
      Object.assign(formData, A());
      visible.value = true;
      Promise.resolve().then(() => {
        pageState.value = false;
      });
    };
    const Y = (s) => {
      pageState.value = true;
      Object.assign(formData, A(), s);
      visible.value = true;
      Promise.resolve().then(() => {
        pageState.value = false;
      });
    };
    const handleBeforeOk = async () => {
      if (await formRef.value?.validate()) return false;
      try {
        await handleSaveEncryption();
      } catch {
        return false;
      }
      return false;
    };
    const handleSaveEncryption = async () => {
      await saveEncryption({
        ...formData,
        appid: appStore.currentId,
      });
      Message.success(formData.id ? "已保存" : "已添加");
      visible.value = false;
      handleListEncryptions();
    };
    const handleDeleteEncryption = (s) => {
      Modal.warning({
        title: "删除加密方案",
        content: `确定删除「${s.name}」?`,
        hideCancel: false,
        onOk: async () => {
          await deleteEncryption(s.id);
          Message.success("已删除");
          handleListEncryptions();
        },
      });
    };
    const handleSaveEncryption2 = async (s, e) => {
      const p = e ? "verify" : "none";
      await saveEncryption({
        ...s,
        signMode: p,
      });
      s.signMode = p;
    };
    const J = {
      aes: 16,
      des: 8,
      rc4: 16,
    };
    const handleClick3 = () => {
      const s = J[formData.algorithm] || 16;
      const e =
        "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
      const p = crypto.getRandomValues(new Uint8Array(s));
      formData.symKey = Array.from(p, (y) => e[y % e.length]).join("");
    };
    const F = (s, e) => {
      const y = btoa(String.fromCharCode(...new Uint8Array(s))).match(
        /.{1,64}/g,
      ).join(`
`);
      return `-----BEGIN ${e}-----
${y}
-----END ${e}-----`;
    };
    const handleGenerateRSAKey = async () => {
      if (window.isSecureContext && window.crypto?.subtle)
        try {
          const s = await crypto.subtle.generateKey(
            {
              name: "RSA-OAEP",
              modulusLength: 2048,
              publicExponent: new Uint8Array([1, 0, 1]),
              hash: "SHA-256",
            },
            true,
            ["encrypt", "decrypt"],
          );
          const e = await crypto.subtle.exportKey("spki", s.publicKey);
          const p = await crypto.subtle.exportKey("pkcs8", s.privateKey);
          return {
            publicKey: F(e, "PUBLIC KEY"),
            privateKey: F(p, "PRIVATE KEY"),
          };
        } catch {}
      return await generateRSAKey();
    };
    const loadingState = ref(false);
    const handleClick2 = async () => {
      loadingState.value = true;
      try {
        if (formData.algorithm === "session") {
          const s = await handleGenerateRSAKey();
          formData.serverPrivateKey = s.privateKey;
          formData.serverPublicKey = s.publicKey;
          Message.success("已生成服务端 RSA 2048 密钥对");
        } else {
          const [s, e] = await Promise.all([
            handleGenerateRSAKey(),
            handleGenerateRSAKey(),
          ]);
          formData.clientPrivateKey = s.privateKey;
          formData.clientPublicKey = s.publicKey;
          formData.serverPrivateKey = e.privateKey;
          formData.serverPublicKey = e.publicKey;
          Message.success("已生成两对 RSA 2048 密钥");
        }
      } catch (s) {
        reportError(s, "生成失败");
      } finally {
        loadingState.value = false;
      }
    };
    watch(
      () => formData.algorithm,
      (s, e) => {
        !visible.value ||
          s === e ||
          pageState.value ||
          ((formData.symKey = ""),
          (formData.clientPrivateKey = ""),
          (formData.clientPublicKey = ""),
          (formData.serverPrivateKey = ""),
          (formData.serverPublicKey = ""));
      },
    );
    watch(() => appStore.currentId, handleListEncryptions);
    onMounted(handleListEncryptions);
    return {
      appStore,
      x,
      i,
      pageState,
      A,
      w,
      G,
      D,
      handleListEncryptions,
      handlePressEnter,
      handleClick,
      Y,
      handleBeforeOk,
      handleSaveEncryption,
      handleDeleteEncryption,
      handleSaveEncryption2,
      J,
      handleClick3,
      F,
      handleGenerateRSAKey,
      loadingState,
      handleClick2,
      loading,
      visible,
      formRef,
      formData,
    };
  },
};
</script>
<template>
  <div class="encrypt-page">
    <div class="toolbar">
      <Button type="primary" class="add-btn" @click="handleClick">
        <template #icon><IconPlus></IconPlus></template>
        <template #default>新建加密方案</template>
      </Button>
      <div class="grow"></div>
      <Input
        v-model="i.keyword"
        placeholder="关键词搜索..."
        allow-clear=""
        class="search-input"
        @pressEnter="handlePressEnter"
      >
        <template #suffix>
          <IconSearch
            :style="{
              cursor: 'pointer',
            }"
            @click="handlePressEnter"
          ></IconSearch>
        </template>
      </Input>
    </div>
    <Spin
      :loading="loading"
      :style="{
        width: '100%',
      }"
    >
      <template #default>
        <template v-if="D.length">
          <div :key="0" class="card-grid">
            <template v-for="o in D" :key="o.id">
              <div class="encrypt-card">
                <div class="card-head">
                  <div class="card-name">{{ o.name }} <Tag v-if="appStore.current?.encryption === o.id" color="green" size="small">使用中</Tag></div>
                  <Dropdown trigger="click" position="br">
                    <template #content>
                      <Doption @click="(R) => Y(o)"
                        ><template #default>编辑</template></Doption
                      >
                      <Doption
                        :disabled="appStore.current?.encryption === o.id"
                        @click="(R) => handleDeleteEncryption(o)"
                        class="danger"
                      >
                        <template #default>删除</template>
                      </Doption>
                    </template>
                    <template #default>
                      <button class="more-btn">
                        <IconMoreVertical></IconMoreVertical>
                      </button>
                    </template>
                  </Dropdown>
                </div>
                <div class="card-row">
                  <span class="label">加密方式</span>
                  <Tag color="arcoblue" size="small">
                    <template #default>{{ o.algorithm }}</template>
                  </Tag>
                </div>
                <div class="card-row">
                  <span class="label">时差校验</span>
                  <span class="value">{{ o.timeSkew + "秒" }}</span>
                </div>
                <div class="card-row">
                  <span class="label">数据签名</span>
                  <Switch
                    size="small"
                    :model-value="o.signMode === 'verify'"
                    @change="(R) => handleSaveEncryption2(o, R)"
                  ></Switch>
                </div>
              </div>
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
          " 当前第 " + i.page + " 页 共 " + G + " 页 " + w.length + " 条结果 "
        }}
      </span>
      <Pagination
        v-model:current="i.page"
        :total="w.length"
        :page-size="i.size"
      ></Pagination>
    </div>
    <Modal
      v-model:visible="visible"
      :title="formData.id ? '编辑方案' : '添加方案'"
      :width="420"
      ok-text="提交"
      cancel-text="取消"
      :on-before-ok="handleBeforeOk"
    >
      <template #default>
        <Form ref="formRef" :model="formData" size="small" auto-label-width="">
          <template #default>
            <FormItem
              field="name"
              label="方案名称"
              :rules="[
                {
                  required: !0,
                  message: '请填写方案名称',
                },
              ]"
            >
              <template #default>
                <Input
                  v-model="formData.name"
                  placeholder="如: 1.0安卓版RSA加密"
                ></Input>
              </template>
            </FormItem>
            <FormItem
              label="加密类型"
              :extra-style="{
                marginTop: 0,
              }"
            >
              <template #default>
                <Select
                  v-model="formData.algorithm"
                  :style="{
                    flex: '1',
                  }"
                >
                  <template #default>
                    <Option value="aes"
                      ><template #default>AES-128-CBC</template></Option
                    >
                    <Option value="des"
                      ><template #default>DES加密</template></Option
                    >
                    <Option value="rc4"
                      ><template #default>RC4加密</template></Option
                    >
                    <Option value="rsa"
                      ><template #default>RSA加密</template></Option
                    >
                    <Option value="session">
                      <template #default>会话密钥(AES 协商)</template>
                    </Option>
                  </template>
                </Select>
                <template
                  v-if="
                    formData.algorithm === 'rsa' ||
                    formData.algorithm === 'session'
                  "
                >
                  <Button
                    :key="0"
                    type="outline"
                    :loading="loadingState"
                    :style="{
                      'margin-left': '8px',
                    }"
                    @click="handleClick2"
                  >
                    <template #default>一键生成密钥</template>
                  </Button>
                </template>
                <template v-else>
                  <Button
                    :key="1"
                    type="outline"
                    :style="{
                      'margin-left': '8px',
                    }"
                    @click="handleClick3"
                  >
                    <template #default>随机生成密钥</template>
                  </Button>
                </template>
              </template>
              <template #extra>
                <template v-if="formData.algorithm === 'aes'">
                  <span>
                    AES-128-CBC：IV 加密时随机生成 16
                    字节，与密文拼接后输出；解密时先取 IV 再使用 KEY+IV 解密。
                  </span>
                </template>
                <template v-if="formData.algorithm === 'session'">
                  <span>
                    会话密钥：客户端启动时 handshake 用服务端公钥加密一段随机
                    AES-256 key 上传，服务端 Redis 存 session(24h
                    滑动续期)，后续业务请求/响应都走 AES
                    信封。客户端零长期私钥，session
                    失效自动重协商。仅需填服务端公私钥(handshake 用)。
                  </span>
                </template>
              </template>
            </FormItem>
            <template
              v-if="
                formData.algorithm === 'rsa' || formData.algorithm === 'session'
              "
            >
              <template v-if="formData.algorithm === 'rsa'">
                <FormItem :key="0" label="客户端私钥">
                  <template #default>
                    <Textarea
                      v-model="formData.clientPrivateKey"
                      placeholder="（可空）此私钥仅用于客户端在获取到服务端返回的数据时进行解密"
                      class="key-textarea"
                      :rows="3"
                    ></Textarea>
                  </template>
                </FormItem>
              </template>
              <template v-else></template>
              <template v-if="formData.algorithm === 'rsa'">
                <FormItem :key="1" label="客户端公钥">
                  <template #default>
                    <Textarea
                      v-model="formData.clientPublicKey"
                      placeholder="此公钥用于服务端返回数据时进行加密"
                      class="key-textarea"
                      :rows="3"
                    ></Textarea>
                  </template>
                </FormItem>
              </template>
              <template v-else></template>
              <FormItem label="服务端私钥">
                <template #default>
                  <Textarea
                    v-model="formData.serverPrivateKey"
                    placeholder="此私钥保留在服务端，用于解密客户端提交过来的参数"
                    class="key-textarea"
                    :rows="3"
                  ></Textarea>
                </template>
              </FormItem>
              <FormItem label="服务端公钥">
                <template #default>
                  <Textarea
                    v-model="formData.serverPublicKey"
                    placeholder="此公钥保存在客户端，用加密请求服务端时的参数"
                    class="key-textarea"
                    :rows="3"
                  ></Textarea>
                </template>
              </FormItem>
              <template v-if="formData.algorithm === 'session'">
                <FormItem :key="2" label="编码类型">
                  <template #default>
                    <Select v-model="formData.encoding">
                      <template #default>
                        <Option value="base64"
                          ><template #default>Base64</template></Option
                        >
                        <Option value="hex"
                          ><template #default>Hex</template></Option
                        >
                      </template>
                    </Select>
                  </template>
                </FormItem>
              </template>
              <template v-else></template>
            </template>
            <template v-else>
              <FormItem label="编码类型">
                <template #default>
                  <Select v-model="formData.encoding">
                    <template #default>
                      <Option value="base64"
                        ><template #default>Base64</template></Option
                      >
                      <Option value="hex"
                        ><template #default>Hex</template></Option
                      >
                    </template>
                  </Select>
                </template>
              </FormItem>
              <FormItem label="密钥">
                <template #default>
                  <Input
                    v-model="formData.symKey"
                    placeholder="请输入密钥"
                  ></Input>
                </template>
              </FormItem>
            </template>
            <FormItem label="数据签名">
              <template #default>
                <Select v-model="formData.signMode">
                  <template #default>
                    <Option value="verify"
                      ><template #default>验证</template></Option
                    >
                    <Option value="none"
                      ><template #default>不验证</template></Option
                    >
                  </template>
                </Select>
              </template>
            </FormItem>
            <FormItem label="时差校验">
              <template #default>
                <InputNumber
                  v-model="formData.timeSkew"
                  :min="0"
                  :precision="0"
                >
                  <template #suffix>秒</template>
                </InputNumber>
              </template>
            </FormItem>
          </template>
        </Form>
      </template>
    </Modal>
  </div>
</template>
<style scoped>
.encrypt-page {
  padding: 24px 32px;
  min-height: 100%;
  background: var(--content-bg, #f5f7fa);
}
.toolbar {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}
.toolbar .add-btn {
  flex-shrink: 0;
}
.toolbar .grow {
  flex: 1;
}
.toolbar .search-input {
  width: 260px;
  max-width: 100%;
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
@media (min-width: 1280px) {
  .card-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}
.encrypt-card {
  background: var(--color-bg-2);
  border-radius: 6px;
  padding: 14px 18px;
  box-shadow: 0 1px 2px #0000000a;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--color-neutral-3);
  margin-bottom: 6px;
}
.card-name {
  font-size: 15px;
  font-weight: 600;
  color: var(--color-text-1);
}
.more-btn {
  background: transparent;
  border: none;
  cursor: pointer;
  color: var(--color-text-3);
  padding: 2px 4px;
  display: inline-flex;
  font-size: 16px;
}
.more-btn:hover {
  color: rgb(var(--primary-6));
}
.card-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 13px;
  min-height: 28px;
}
.card-row .label {
  color: var(--color-text-3);
}
.card-row .value {
  color: var(--color-text-1);
}
.footer-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 24px;
  flex-wrap: wrap;
  gap: 12px;
}
.footer-bar .page-info {
  color: var(--color-text-3);
  font-size: 13px;
}
:deep(.arco-dropdown-option.danger) {
  color: rgb(var(--danger-6));
}
:deep(.key-textarea textarea) {
  resize: vertical;
  min-height: 60px;
  font-family: ui-monospace, Consolas, monospace;
  font-size: 12px;
}
</style>
