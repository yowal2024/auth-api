<!-- Reconstructed editable Vue source; see README for recovery limits. -->
<script>
import { computed, ref, reactive } from "vue";
import {
  Card,
  Form,
  FormItem,
  Input,
  InputPassword,
  Link,
  Button,
  Message,
} from "@arco-design/web-vue";
import { useRouter } from "vue-router";
import { IconSafe, IconUser, IconLock } from "@arco-design/web-vue/es/icon";
import { useUserStore } from "@/stores/user.js";
import { entryPath } from "@/utils/route.js";
import { getCaptcha } from "@/api/auth.js";
import SliderBar from "@/components/SliderBar.vue";
export default {
  name: "Login",
  components: {
    Card,
    IconSafe,
    Form,
    FormItem,
    Input,
    IconUser,
    InputPassword,
    IconLock,
    SliderBar,
    Link,
    Button,
  },
  setup(y) {
    const router = useRouter();
    const userStore = useUserStore();
    const formRef = ref();
    const loading = ref(false);
    const formData = reactive({
      username: "",
      password: "",
      captchaCode: "",
    });
    const needCaptcha = ref(false);
    const captchaImage = ref("");
    const captchaId = ref("");
    const captchaType = ref("");
    const captchaPrompt = ref("");
    const sliderPassed = ref(false);
    const clickPoints = ref([]);
    const captchaPlaceholder = computed(() =>
      captchaType.value === "math"
        ? "请输入计算结果"
        : captchaType.value === "chinese"
          ? "请输入图中汉字"
          : "请输入验证码",
    );
    const refreshCaptcha = async () => {
      try {
        const n = await getCaptcha();
        captchaId.value = n.id;
        captchaImage.value = n.image;
        captchaType.value = n.type || "";
        captchaPrompt.value = n.prompt || "";
        formData.captchaCode = "";
        clickPoints.value = [];
        sliderPassed.value = false;
      } catch {}
    };
    const handleCaptchaClick = (n) => {
      if (clickPoints.value.length >= 3) return;
      const e = n.currentTarget.getBoundingClientRect();
      const k = 300 / e.width;
      const S = 150 / e.height;
      const L = Math.round((n.clientX - e.left) * k);
      const V = Math.round((n.clientY - e.top) * S);
      clickPoints.value.push({
        x: L,
        y: V,
      });
      clickPoints.value.length === 3 &&
        (formData.captchaCode = clickPoints.value
          .map((I) => `${I.x},${I.y}`)
          .join(";"));
    };
    const submitLogin = async () => {
      if (loading.value) return;
      if (
        needCaptcha.value &&
        captchaType.value === "slider" &&
        !sliderPassed.value
      ) {
        Message.warning("请先完成滑块验证");
        return;
      }
      if (
        needCaptcha.value &&
        captchaType.value === "click" &&
        clickPoints.value.length < 3
      ) {
        Message.warning("请依次点击图中 3 个汉字");
        return;
      }
      if (!(await formRef.value?.validate())) {
        loading.value = true;
        try {
          await userStore.login(
            formData.username,
            formData.password,
            captchaId.value,
            formData.captchaCode,
          );
          Message.success("登录成功");
          await router.push(entryPath("/app"));
        } catch (e) {
          if (e?.data?.needCaptcha) needCaptcha.value = true;
          needCaptcha.value &&
            captchaType.value !== "slider" &&
            (await refreshCaptcha());
        } finally {
          loading.value = false;
        }
      }
    };
    return {
      router,
      userStore,
      needCaptcha,
      captchaImage,
      captchaId,
      captchaType,
      captchaPrompt,
      sliderPassed,
      clickPoints,
      captchaPlaceholder,
      refreshCaptcha,
      handleCaptchaClick,
      submitLogin,
      formRef,
      formData,
      loading,
    };
  },
};
</script>
<template>
  <div class="login-page">
    <div class="login-bg"></div>
    <Card class="login-card" :bordered="!1">
      <template #default>
        <div class="brand">
          <IconSafe
            :size="36"
            :style="{
              color: '#165dff',
            }"
          ></IconSafe>
          <h2>Auth 管理后台</h2>
          <p class="sub">用户认证 / 多应用 / 云函数管理</p>
        </div>
        <Form ref="formRef" :model="formData" layout="vertical">
          <template #default>
            <FormItem
              field="username"
              :rules="[
                {
                  required: !0,
                  message: '请输入账号',
                },
              ]"
              hide-label=""
            >
              <template #default>
                <Input
                  v-model="formData.username"
                  placeholder="账号"
                  size="large"
                  allow-clear=""
                  @pressEnter="submitLogin"
                >
                  <template #prefix><IconUser></IconUser></template>
                </Input>
              </template>
            </FormItem>
            <FormItem
              field="password"
              :rules="[
                {
                  required: !0,
                  message: '请输入密码',
                },
              ]"
              hide-label=""
            >
              <template #default>
                <InputPassword
                  v-model="formData.password"
                  placeholder="密码"
                  size="large"
                  @pressEnter="submitLogin"
                >
                  <template #prefix><IconLock></IconLock></template>
                </InputPassword>
              </template>
            </FormItem>
            <template v-if="needCaptcha">
              <template v-if="captchaType === 'slider'">
                <div :key="0" class="captcha-block">
                  <SliderBar
                    :passed="sliderPassed"
                    @pass="(m) => (sliderPassed = !0)"
                    @reset="(m) => (sliderPassed = !1)"
                  ></SliderBar>
                </div>
              </template>
              <template v-else>
                <template v-if="captchaType === 'click'">
                  <div :key="1" class="captcha-block">
                    <div class="click-prompt">{{ captchaPrompt }}</div>
                    <div class="click-img-wrap" @click="handleCaptchaClick">
                      <img
                        :src="captchaImage"
                        class="click-img"
                        draggable="false"
                      />
                      <template v-for="(m, X) in clickPoints" :key="X">
                        <span
                          class="click-dot"
                          :style="{
                            left: m.x - 12 + 'px',
                            top: m.y - 12 + 'px',
                          }"
                        >
                          {{ X + 1 }}
                        </span>
                      </template>
                    </div>
                    <div class="click-actions">
                      <Link @click="refreshCaptcha"
                        ><template #default>换一张</template></Link
                      >
                      <span class="click-hint">{{
                        clickPoints.length + "/3 已点击"
                      }}</span>
                    </div>
                  </div>
                </template>
                <template v-else>
                  <div :key="2" class="captcha-block">
                    <div class="captcha-row">
                      <Input
                        v-model="formData.captchaCode"
                        :placeholder="captchaPlaceholder"
                        size="large"
                        allow-clear=""
                        @pressEnter="submitLogin"
                      >
                        <template #prefix><IconSafe></IconSafe></template>
                      </Input>
                      <template v-if="captchaImage">
                        <img
                          :key="0"
                          :src="captchaImage"
                          class="captcha-img"
                          title="点击刷新"
                          @click="refreshCaptcha"
                        />
                      </template>
                      <template v-else></template>
                    </div>
                  </div>
                </template>
              </template>
            </template>
            <template v-else></template>
            <Button
              type="primary"
              size="large"
              long=""
              :loading="loading"
              :disabled='needCaptcha &amp;&amp; captchaType === "slider" &amp;&amp; !sliderPassed'
              @click="submitLogin"
            >
              <template #default>登 录</template>
            </Button>
          </template>
        </Form>
      </template>
    </Card>
  </div>
</template>
<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  overflow: hidden;
}
.login-bg {
  position: absolute;
  inset: 0;
  background:
    radial-gradient(
      circle at 15% 25%,
      rgba(22, 93, 255, 0.07),
      transparent 50%
    ),
    radial-gradient(
      circle at 85% 80%,
      rgba(22, 93, 255, 0.05),
      transparent 45%
    ),
    #f5f7fa;
  z-index: 0;
}
.login-card {
  position: relative;
  z-index: 1;
  width: 380px;
  padding: 12px 18px;
  border-radius: 12px;
  background: #fff;
  border: 1px solid var(--color-border-2);
  box-shadow: 0 8px 32px #0f172a0f;
}
.brand {
  text-align: center;
  margin-bottom: 24px;
}
.brand h2 {
  margin: 8px 0 4px;
  font-size: 22px;
  color: #1f2d3d;
}
.brand .sub {
  margin: 0;
  color: var(--color-text-3);
  font-size: 12px;
}
.captcha-block {
  margin-bottom: 16px;
}
.captcha-row {
  display: flex;
  gap: 10px;
  align-items: center;
  width: 100%;
}
.captcha-img {
  height: 40px;
  width: 120px;
  border-radius: 4px;
  border: 1px solid var(--color-border-2);
  cursor: pointer;
  flex-shrink: 0;
}
.click-prompt {
  font-size: 13px;
  color: var(--color-text-2);
  margin-bottom: 8px;
  font-weight: 500;
}
.click-img-wrap {
  position: relative;
  display: inline-block;
  width: 100%;
  cursor: crosshair;
  border-radius: 6px;
  overflow: hidden;
  border: 1px solid var(--color-border-2);
  user-select: none;
}
.click-img {
  display: block;
  width: 100%;
  height: auto;
  pointer-events: none;
}
.click-dot {
  position: absolute;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: #165dffd9;
  color: #fff;
  font-size: 12px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  pointer-events: none;
  box-shadow: 0 1px 4px #0000004d;
}
.click-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 6px;
  font-size: 12px;
}
.click-hint {
  color: var(--color-text-3);
}
</style>
