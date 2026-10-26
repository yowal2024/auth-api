<!-- Reconstructed editable Vue source; see README for recovery limits. -->
<script>
import { reportError } from "@/utils/errors.js";
import { useAppStore } from "@/stores/app.js";
import { useRouter } from "vue-router";
import { entryPath } from "@/utils/route.js";
import {
  listEncryptions,
  getMTLS,
  issueClientCertificate,
  updateApp,
  generateCA,
} from "@/api/app.js";
import { uploadAppLogo } from "@/api/auth.js";
import { defaultAvatar } from "@/utils/avatar.js";
import { watch, onMounted, reactive, ref } from "vue";
import {
  Upload,
  Tag,
  Form,
  InputNumber,
  Select,
  Switch,
  Message,
  Popconfirm,
  Button,
  Modal,
  FormItem,
  Input,
  Tooltip,
  Option,
  Alert,
  Textarea,
} from "@arco-design/web-vue";
import {
  IconCamera,
  IconRefresh,
  IconCopy,
  IconSafe,
  IconFile,
  IconSave,
} from "@arco-design/web-vue/es/icon";
export default {
  name: "AppInfo",
  components: {
    Upload,
    IconCamera,
    Tag,
    Form,
    FormItem,
    Input,
    Tooltip,
    IconRefresh,
    IconCopy,
    Select,
    Option,
    InputNumber,
    Switch,
    Popconfirm,
    Button,
    IconSafe,
    IconFile,
    Modal,
    Alert,
    Textarea,
    IconSave,
  },
  setup(Et) {
    const router = useRouter();
    const openGeoSettings = () =>
      router.push({ path: entryPath("/system"), query: { tab: "geo" } });
    const appStore = useAppStore();
    const loadingState2 = ref(false);
    const P = ref([]);
    const formData = reactive({
      id: 0,
      name: "",
      tokenKey: "",
      mode: "vip",
      status: 1,
      logo: "",
      bindDevice: false,
      bindRegion: false,
      diffUpdate: false,
      randomPackage: false,
      encryption: null,
      tokenTtlValue: 7,
      tokenTtlUnit: "d",
      ipDistanceMax: 50,
      bruteForceEnabled: false,
      mtlsEnabled: false,
    });
    const O = (a) => {
      const e = Number(a) || 0;
      return e <= 0
        ? {
            value: 7,
            unit: "d",
          }
        : e % 86400 === 0
          ? {
              value: e / 86400,
              unit: "d",
            }
          : e % 3600 === 0
            ? {
                value: e / 3600,
                unit: "h",
              }
            : e % 60 === 0
              ? {
                  value: e / 60,
                  unit: "m",
                }
              : {
                  value: e,
                  unit: "s",
                };
    };
    const j = (a, e) => {
      const c = Number(a) || 0;
      return Math.max(
        1,
        c *
          ({
            s: 1,
            m: 60,
            h: 3600,
            d: 86400,
          }[e] || 1),
      );
    };
    const G = {
      free: "免费",
      vip: "付费",
    };
    const Y = (a) => G[a] || "免费";
    const H = {
      aes: "AES",
      des: "DES",
      rc4: "RC4",
      rsa: "RSA",
    };
    const W = (a) => H[a] || (a || "").toUpperCase();
    const K = () => {
      const a = appStore.current;
      if (!a) return;
      const e = O(a.userTokenTtl);
      Object.assign(formData, {
        id: a.id,
        name: a.name || "",
        tokenKey: a.tokenKey || "",
        mode: a.mode || "free",
        status: a.status ?? 1,
        logo: a.logo || "",
        bindDevice: !!a.bindDevice,
        bindRegion: !!a.bindRegion,
        diffUpdate: !!a.diffUpdate,
        randomPackage: !!a.randomPackage,
        encryption: a.encryption || null,
        tokenTtlValue: e.value,
        tokenTtlUnit: e.unit,
        ipDistanceMax: a.ipDistanceMax || 50,
        bruteForceEnabled: !!a.bruteForceEnabled,
        mtlsEnabled: !!a.mtlsEnabled,
      });
    };
    const handleListEncryptions = async () => {
      try {
        P.value = await listEncryptions({
          appid: appStore.currentId,
        });
      } catch {
        P.value = [];
      }
    };
    const r = reactive({
      hasCA: false,
      fingerprint: "",
    });
    const loading = ref(false);
    const loadingState = ref(false);
    const visible = ref(false);
    const v = reactive({
      cert: "",
      key: "",
      fingerprint: "",
    });
    const handleGetMTLS = async () => {
      if (formData.id)
        try {
          const a = await getMTLS(formData.id);
          r.hasCA = !!a.hasCA;
          r.fingerprint = a.fingerprint || "";
        } catch {
          r.hasCA = false;
          r.fingerprint = "";
        }
    };
    const handleGenerateCA = async () => {
      loading.value = true;
      try {
        const a = await generateCA(formData.id);
        r.hasCA = true;
        r.fingerprint = a.fingerprint || "";
        Message.success("已生成 CA");
      } finally {
        loading.value = false;
      }
    };
    const handleClick4 = async () => {
      loadingState.value = true;
      try {
        const a = await issueClientCertificate(formData.id);
        v.cert = a.cert || "";
        v.key = a.key || "";
        v.fingerprint = a.fingerprint || "";
        visible.value = true;
      } finally {
        loadingState.value = false;
      }
    };
    const h = (a, e) => {
      const c = new Blob([a], {
        type: "application/x-pem-file",
      });
      const u = URL.createObjectURL(c);
      const b = document.createElement("a");
      b.href = u;
      b.download = e;
      b.click();
      URL.revokeObjectURL(u);
    };
    const handleClick5 = () => h(v.cert, `app-${formData.id}-cert.pem`);
    const handleClick6 = () => h(v.key, `app-${formData.id}-key.pem`);
    const handleClick3 = () => {
      formData.bindRegion ||
        Message.info('需要先开启"绑定IP区域"才能设置距离阈值');
    };
    const handleUploadAppLogo = async (a, e) => {
      const c = e?.file;
      if (c)
        try {
          const { url: u } = await uploadAppLogo(c);
          formData.logo = u;
          Message.success("logo 上传成功,保存后生效");
        } catch (u) {
          reportError(u, "上传失败");
        }
    };
    const handleClick = () => {
      formData.tokenKey = Array.from(crypto.getRandomValues(new Uint8Array(16)))
        .map((a) => a.toString(16).padStart(2, "0"))
        .join("");
      Message.info("已生成新 KEY,保存后生效");
    };
    const handleClick2 = async () => {
      try {
        await navigator.clipboard.writeText(formData.tokenKey);
        Message.success("已复制");
      } catch {
        Message.error("复制失败");
      }
    };
    const handleUpdateApp = async () => {
      loadingState2.value = true;
      try {
        await updateApp(formData.id, {
          name: formData.name,
          tokenKey: formData.tokenKey,
          mode: formData.mode,
          status: formData.status,
          logo: formData.logo,
          bindDevice: formData.bindDevice,
          bindRegion: formData.bindRegion,
          diffUpdate: formData.diffUpdate,
          randomPackage: formData.randomPackage,
          encryption: formData.encryption,
          userTokenTtl: j(formData.tokenTtlValue, formData.tokenTtlUnit),
          ipDistanceMax: formData.ipDistanceMax,
          bruteForceEnabled: formData.bruteForceEnabled,
          mtlsEnabled: formData.mtlsEnabled,
        });
        Message.success("已保存");
        await appStore.fetchApps();
        K();
      } finally {
        loadingState2.value = false;
      }
    };
    watch(
      () => appStore.currentId,
      () => {
        K();
        handleListEncryptions();
        handleGetMTLS();
      },
    );
    onMounted(async () => {
      appStore.apps.length || (await appStore.fetchApps());
      K();
      handleListEncryptions();
      handleGetMTLS();
    });
    return {
      openGeoSettings,
      appStore,
      loadingState2,
      P,
      O,
      j,
      G,
      Y,
      H,
      W,
      K,
      handleListEncryptions,
      r,
      loadingState,
      v,
      handleGetMTLS,
      handleGenerateCA,
      handleClick4,
      h,
      handleClick5,
      handleClick6,
      handleClick3,
      handleUploadAppLogo,
      handleClick,
      handleClick2,
      handleUpdateApp,
      formData,
      loading,
      visible,
      defaultAvatar,
    };
  },
};
</script>
<template>
  <div class="page info-page">
    <div class="banner">
      <Upload
        :show-file-list="!1"
        :auto-upload="!1"
        list-type="picture-card"
        accept="image/*"
        @change="handleUploadAppLogo"
      >
        <template #upload-button>
          <div class="banner-logo">
            <template v-if="formData.logo"
              ><img :key="0" :src="formData.logo"
            /></template>
            <template v-else><img :key="1" :src="defaultAvatar" /></template>
            <div class="banner-logo-mask"><IconCamera></IconCamera></div>
          </div>
        </template>
      </Upload>
      <div class="banner-meta">
        <div class="banner-name">{{ formData.name || "未命名应用" }}</div>
        <div class="banner-sub">
          <span>{{ "APPID: " + formData.id }}</span>
          <span class="dot"></span>
          <Tag :color="formData.status === 1 ? 'green' : 'gray'">
            <template #default>{{
              formData.status === 1 ? "运行中" : "已禁用"
            }}</template>
          </Tag>
          <span class="dot"></span>
          <Tag color="arcoblue">
            <template #default>{{ Y(formData.mode) }}</template>
          </Tag>
        </div>
      </div>
    </div>
    <div class="info-grid">
      <div class="info-card">
        <div class="card-title">基础信息</div>
        <Form :model="formData" auto-label-width="" class="info-form">
          <template #default>
            <FormItem label="应用名称">
              <template #default>
                <Input
                  v-model="formData.name"
                  placeholder="请输入应用名称"
                ></Input>
              </template>
            </FormItem>
            <FormItem label="APPKEY">
              <template #default>
                <div class="key-row">
                  <Input
                    v-model="formData.tokenKey"
                    readonly=""
                    class="key-input"
                  ></Input>
                  <Tooltip content="重新生成">
                    <template #default>
                      <button class="key-btn" @click="handleClick">
                        <IconRefresh></IconRefresh>
                      </button>
                    </template>
                  </Tooltip>
                  <Tooltip content="复制">
                    <template #default>
                      <button class="key-btn" @click="handleClick2">
                        <IconCopy></IconCopy>
                      </button>
                    </template>
                  </Tooltip>
                </div>
              </template>
            </FormItem>
            <FormItem label="运营模式">
              <template #default>
                <Select v-model="formData.mode">
                  <template #default>
                    <Option value="free"
                      ><template #default>免费</template></Option
                    >
                    <Option value="vip"
                      ><template #default>付费</template></Option
                    >
                  </template>
                </Select>
              </template>
            </FormItem>
            <FormItem label="应用状态">
              <template #default>
                <Select v-model="formData.status">
                  <template #default>
                    <Option :value="1"
                      ><template #default>正常</template></Option
                    >
                    <Option :value="0"
                      ><template #default>禁用</template></Option
                    >
                  </template>
                </Select>
              </template>
            </FormItem>
          </template>
        </Form>
      </div>
      <div class="info-card">
        <div class="card-title">安全限制</div>
        <div class="switch-list">
          <div class="switch-item">
            <div class="switch-meta">
              <div class="switch-title">Token 过期</div>
              <div class="switch-desc">
                SDK 用户登录颁发的 token 过期时间,每次访问会滑动续期
              </div>
            </div>
            <div class="ttl-row">
              <InputNumber
                v-model="formData.tokenTtlValue"
                :min="1"
                :step="1"
                :precision="0"
                class="ttl-input"
              ></InputNumber>
              <Select v-model="formData.tokenTtlUnit" class="ttl-unit">
                <template #default>
                  <Option value="s"><template #default>秒</template></Option>
                  <Option value="m"><template #default>分</template></Option>
                  <Option value="h"><template #default>小时</template></Option>
                  <Option value="d"><template #default>天</template></Option>
                </template>
              </Select>
            </div>
          </div>
          <div class="switch-item">
            <div class="switch-meta">
              <div class="switch-title">绑定设备</div>
              <div class="switch-desc">
                开启后用户登录会绑定设备,切换设备需校验
              </div>
            </div>
            <Switch v-model="formData.bindDevice"></Switch>
          </div>
          <div class="switch-item">
            <div class="switch-meta">
              <div class="switch-title">绑定IP区域</div>
              <div class="switch-desc">
                开启后用户登录会按 IP 经纬度记录地点,新地点超出上限时拒登
              </div>
            </div>
            <Switch v-model="formData.bindRegion"></Switch>
          </div>
          <div
            :class="[
              'switch-item ip-geo-item',
              {
                'is-disabled': !formData.bindRegion,
              },
            ]"
            @click="handleClick3"
          >
            <div class="switch-meta">
              <div class="switch-title">IP 地点距离阈值</div>
              <div class="switch-desc">
                小于此距离视为同一地点，仅在绑定 IP 区域开启时生效。
                <Button type="text" size="mini" @click.stop="openGeoSettings">
                  配置系统定位服务
                </Button>
              </div>
            </div>
            <div
              :class="[
                'geo-row',
                {
                  'is-locked': !formData.bindRegion,
                },
              ]"
            >
              <InputNumber
                v-model="formData.ipDistanceMax"
                :min="1"
                :step="1"
                :precision="0"
                class="geo-distance"
                placeholder="km"
                :disabled="!formData.bindRegion"
              >
                <template #suffix>km</template>
              </InputNumber>
            </div>
          </div>
            <div class="switch-item">
              <div class="switch-meta">
                <div class="switch-title">登录防爆破</div>
              <div class="switch-desc">
                同一账号 5 分钟内连续登录失败 5 次,锁定 5 分钟禁止再次尝试
              </div>
            </div>
            <Switch v-model="formData.bruteForceEnabled"></Switch>
          </div>
          <div class="switch-item">
            <div class="switch-meta">
              <div class="switch-title">差异更新</div>
              <div class="switch-desc">
                开启后客户端按版本号增量下载 Bspatch
                补丁,下载地址由服务端动态生成
              </div>
            </div>
            <Switch v-model="formData.diffUpdate"></Switch>
          </div>
          <div class="switch-item">
            <div class="switch-meta">
              <div class="switch-title">随机包名</div>
              <div class="switch-desc">
                开启后每次下载会动态重打包出独立包名,降低被识别和封禁的概率.但是只支持
                Android
              </div>
            </div>
            <Switch v-model="formData.randomPackage"></Switch>
          </div>
          <div class="switch-item">
            <div class="switch-meta">
              <div class="switch-title">加密方案</div>
              <div class="switch-desc">客户端通信使用的加密算法,空则不加密</div>
            </div>
            <Select
              v-model="formData.encryption"
              placeholder="不加密"
              allow-clear=""
              :style="{
                width: '260px',
              }"
            >
              <template #default>
                <template v-for="n in P" :key="n.id">
                  <Option :value="n.id">
                    <template #default>{{
                      n.name + "（" + W(n.algorithm) + "） "
                    }}</template>
                  </Option>
                </template>
              </template>
            </Select>
          </div>
        </div>
      </div>
      <div class="info-card">
        <div class="card-title">双向认证(mTLS)</div>
        <div class="switch-list">
          <div class="switch-item">
            <div class="switch-meta">
              <div class="switch-title">SDK 双向认证</div>
              <div class="switch-desc">
                开启后该应用的 SDK 接口强制校验客户端证书(CN 自动绑定
                appid)。请先生成 CA 并把签发的客户端证书集成进
                SDK,再开启此开关并保存。需服务端已开启 HTTPS。
              </div>
            </div>
            <Switch v-model="formData.mtlsEnabled"></Switch>
          </div>
          <div class="switch-item mtls-ca-item">
            <div class="switch-meta">
              <div class="switch-title">应用 CA</div>
              <div class="switch-desc">
                <template v-if="r.hasCA">
                  已生成 · 指纹
                  <span class="fp">{{ r.fingerprint }}</span>
                </template>
                <template v-else>尚未生成 CA,无法签发客户端证书</template>
              </div>
            </div>
            <div class="mtls-btns">
              <template v-if="r.hasCA">
                <Popconfirm
                  :key="0"
                  content="重新生成 CA 会使该应用已签发的所有客户端证书立即失效,确定继续?"
                  @ok="handleGenerateCA"
                >
                  <template #default>
                    <Button :loading="loading"
                      ><template #default>重新生成 CA</template></Button
                    >
                  </template>
                </Popconfirm>
              </template>
              <template v-else>
                <Button
                  :key="1"
                  type="primary"
                  :loading="loading"
                  @click="handleGenerateCA"
                >
                  <template #icon><IconSafe></IconSafe></template>
                  <template #default>生成 CA</template>
                </Button>
              </template>
              <Button
                :disabled="!r.hasCA"
                :loading="loadingState"
                @click="handleClick4"
              >
                <template #icon><IconFile></IconFile></template>
                <template #default>签发客户端证书</template>
              </Button>
            </div>
          </div>
        </div>
      </div>
    </div>
    <Modal
      v-model:visible="visible"
      title="客户端证书已签发"
      :width="640"
      :footer="!1"
    >
      <template #default>
        <Alert
          type="warning"
          :style="{
            'margin-bottom': '12px',
          }"
        >
          <template #default>
            {{
              " 私钥仅本次展示,请立即下载保存。证书 CN = app-" +
              formData.id +
              ",SDK 集成时用作 mTLS 客户端证书。 "
            }}
          </template>
        </Alert>
        <div class="pem-block">
          <div class="pem-head">
            <span>证书 cert.pem</span>
            <Button size="mini" @click="handleClick5"
              ><template #default>下载</template></Button
            >
          </div>
          <Textarea
            :model-value="v.cert"
            readonly=""
            :auto-size="{
              minRows: 4,
              maxRows: 8,
            }"
          ></Textarea>
        </div>
        <div class="pem-block">
          <div class="pem-head">
            <span>私钥 key.pem</span>
            <Button size="mini" @click="handleClick6"
              ><template #default>下载</template></Button
            >
          </div>
          <Textarea
            :model-value="v.key"
            readonly=""
            :auto-size="{
              minRows: 4,
              maxRows: 8,
            }"
          ></Textarea>
        </div>
      </template>
    </Modal>
    <div class="footer-actions">
      <Button
        type="primary"
        size="large"
        :loading="loadingState2"
        @click="handleUpdateApp"
      >
        <template #icon><IconSave></IconSave></template>
        <template #default>保存修改</template>
      </Button>
    </div>
  </div>
</template>
<style scoped>
.info-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.banner {
  background: linear-gradient(135deg, #165dff14, #722ed10f);
  border-radius: 8px;
  padding: 20px 24px;
  display: flex;
  align-items: center;
  gap: 18px;
  box-shadow: 0 1px 2px #0000000a;
}
.banner-logo {
  position: relative;
  width: 72px;
  height: 72px;
  border-radius: 14px;
  overflow: hidden;
  background: var(--color-bg-2);
  box-shadow: 0 2px 8px #0000000f;
  cursor: pointer;
}
.banner-logo img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}
.banner-logo .banner-logo-mask {
  position: absolute;
  inset: 0;
  background: #00000073;
  color: #fff;
  font-size: 22px;
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: opacity 0.2s;
}
.banner-logo:hover .banner-logo-mask {
  opacity: 1;
}
.banner-meta {
  min-width: 0;
  flex: 1;
}
.banner-name {
  font-size: 20px;
  font-weight: 600;
  color: var(--color-text-1);
  margin-bottom: 8px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.banner-sub {
  display: flex;
  align-items: center;
  gap: 10px;
  color: var(--color-text-3);
  font-size: 13px;
  flex-wrap: wrap;
}
.banner-sub .dot {
  width: 3px;
  height: 3px;
  border-radius: 50%;
  background: var(--color-text-4);
  opacity: 0.5;
}
.info-card {
  background: var(--color-bg-2);
  border-radius: 8px;
  padding: 18px 24px 6px;
  box-shadow: 0 1px 2px #0000000a;
}
.card-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text-1);
  margin-bottom: 16px;
  padding-left: 8px;
  border-left: 3px solid rgb(var(--primary-6));
  line-height: 1;
}
.info-grid {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.info-grid .info-card {
  margin: 0;
}
.info-form :deep(.arco-form-item) {
  margin-bottom: 18px;
}
.info-form :deep(.arco-form-item-label) {
  color: var(--color-text-2);
}
.ttl-row {
  display: flex;
  gap: 6px;
  flex: 0 0 auto;
}
.ttl-row .ttl-input {
  width: 120px;
}
.ttl-row .ttl-unit {
  width: 80px;
  flex: 0 0 80px;
}
.geo-row {
  display: flex;
  gap: 6px;
  flex: 0 0 auto;
}
.geo-row .geo-distance {
  width: 120px;
  flex: 0 0 120px;
}
.geo-row.is-locked {
  opacity: 0.5;
}
.ip-geo-item .switch-meta {
  flex: 1 1 220px;
  min-width: 200px;
}
.ip-geo-item.is-disabled {
  cursor: not-allowed;
}
.key-row {
  display: flex;
  align-items: center;
  width: 100%;
  border: 1px solid var(--color-neutral-3);
  border-radius: 4px;
  background: var(--color-fill-2);
  transition: border-color 0.15s;
}
.key-row:focus-within {
  border-color: rgb(var(--primary-6));
  background: var(--color-bg-2);
}
.key-row .key-input {
  flex: 1;
}
.key-row .key-input :deep(.arco-input-wrapper) {
  border: none !important;
  background: transparent !important;
  box-shadow: none !important;
}
.key-row .key-btn {
  width: 36px;
  height: 32px;
  background: transparent;
  border: none;
  border-left: 1px solid var(--color-neutral-3);
  color: var(--color-text-3);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 15px;
  transition: all 0.15s;
}
.key-row .key-btn:hover {
  color: rgb(var(--primary-6));
  background: var(--color-fill-3);
}
.switch-list {
  display: flex;
  flex-direction: column;
}
.switch-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 0;
  gap: 16px;
}
.switch-item + .switch-item {
  border-top: 1px solid var(--color-neutral-2);
}
.switch-item .switch-meta {
  min-width: 0;
}
.switch-item .switch-title {
  font-size: 14px;
  color: var(--color-text-1);
  font-weight: 500;
}
.switch-item .switch-desc {
  font-size: 12px;
  color: var(--color-text-3);
  margin-top: 4px;
}
.switch-item .mtls-btns {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}
.switch-item .fp {
  word-break: break-all;
}
.footer-actions {
  display: flex;
  justify-content: flex-end;
  padding: 4px 4px 8px;
}
@media (max-width: 768px) {
  .info-page {
    padding: 12px;
  }
  .banner {
    padding: 16px;
    gap: 14px;
  }
  .banner-logo {
    width: 56px;
    height: 56px;
    border-radius: 12px;
  }
  .banner-name {
    font-size: 17px;
  }
  .info-card {
    padding: 16px 16px 4px;
  }
  .info-form :deep(.arco-form-item-label-col) {
    flex: 0 0 80px !important;
    max-width: 80px !important;
  }
  .footer-actions {
    justify-content: stretch;
  }
  .footer-actions .arco-btn {
    flex: 1;
  }
  .ip-geo-item {
    flex-direction: column;
    align-items: stretch;
  }
  .ip-geo-item .switch-meta {
    flex: 1 1 auto;
  }
  .mtls-ca-item {
    flex-direction: column;
    align-items: stretch;
  }
  .mtls-ca-item .switch-meta {
    flex: 1 1 auto;
  }
  .mtls-ca-item .mtls-btns {
    width: 100%;
  }
  .mtls-ca-item .mtls-btns .arco-btn {
    flex: 1;
  }
  .geo-row {
    width: 100%;
    flex-wrap: wrap;
  }
  .geo-row .geo-distance {
    flex: 0 0 100px;
    width: 100px;
  }
}
</style>
