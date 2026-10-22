<!-- Reconstructed editable Vue source; see README for recovery limits. -->
<script>
import { reportError } from "@/utils/errors.js";
import { onMounted, computed, ref, reactive } from "vue";
import {
  Breadcrumb,
  Card,
  Modal,
  BreadcrumbItem,
  Avatar,
  Input,
  Form,
  FormItem,
  InputPassword,
  Message,
} from "@arco-design/web-vue";
import { useRouter } from "vue-router";
import { IconCamera, IconEdit } from "@arco-design/web-vue/es/icon";
import { getAdminEntry } from "@/utils/install-status.js";
import { useUserStore } from "@/stores/user.js";
import { uploadAvatar, updatePassword } from "@/api/auth.js";
import { entryPath } from "@/utils/route.js";
export default {
  name: "Profile",
  components: {
    Breadcrumb,
    BreadcrumbItem,
    Card,
    Avatar,
    IconCamera,
    Input,
    IconEdit,
    Modal,
    Form,
    FormItem,
    InputPassword,
  },
  setup(xe) {
    const router = useRouter();
    const userStore = useUserStore();
    const r = computed(() => userStore.profile);
    const A = computed(() =>
      (r.value?.username || "A").slice(0, 1).toUpperCase(),
    );
    const E = computed(() =>
      r.value?.avatar
        ? {}
        : {
            backgroundColor: "#165dff",
          },
    );
    const v = reactive({
      nickname: false,
      username: false,
    });
    const u = reactive({
      nickname: "",
      username: "",
    });
    const savingProfile = ref(false);
    const savingPassword = ref(false);
    const x = (a) => {
      if (savingProfile.value || !r.value) return;
      u[a] = r.value?.[a] || "";
      v[a] = true;
    };
    const f = async (a) => {
      if (savingProfile.value || !v[a] || !r.value) return;
      const e = (u[a] || "").trim();
      const i = r.value?.[a] || "";
      if (e !== i) {
        if (
          a === "username" &&
          (new TextEncoder().encode(e).length < 5 ||
            new TextEncoder().encode(e).length > 12)
        ) {
          Message.error("账号长度需在 5-12 字节");
          return;
        }
        savingProfile.value = true;
        try {
          await userStore.updateProfile({
            username: a === "username" ? e : r.value.username,
            nickname: a === "nickname" ? e : r.value.nickname || "",
            avatar: r.value.avatar || "",
          });
          Message.success("已保存");
          v[a] = false;
        } catch (error) {
          reportError(error);
        } finally {
          savingProfile.value = false;
        }
      } else v[a] = false;
    };
    const fileRef = ref(null);
    const handleClick = () => {
      if (!savingProfile.value) fileRef.value?.click();
    };
    const handleUploadAvatar = async (a) => {
      const e = a.target.files?.[0];
      if (savingProfile.value) return;
      if (((a.target.value = ""), !!e)) {
        if (e.size > 5 * 1024 * 1024) {
          Message.error("图片不能超过 5MB");
          return;
        }
        savingProfile.value = true;
        const token = userStore.token;
        try {
          const { url: i } = await uploadAvatar(e);
          if (userStore.token !== token || !r.value) return;
          await userStore.updateProfile({
            username: r.value.username,
            nickname: r.value.nickname || "",
            avatar: i,
          });
          Message.success("头像已更新");
        } catch (i) {
          reportError(i, "上传失败");
        } finally {
          savingProfile.value = false;
        }
      }
    };
    const visible = ref(false);
    const pwdFormRef = ref(null);
    const formData = reactive({
      oldPassword: "",
      newPassword: "",
      confirmPassword: "",
    });
    const S = (a, e) => {
      a !== formData.newPassword ? e("两次密码不一致") : e();
    };
    const handleClose = () => {
      formData.oldPassword = "";
      formData.newPassword = "";
      formData.confirmPassword = "";
    };
    const handleUpdatePassword = async () => {
      if (savingPassword.value) return false;
      savingPassword.value = true;
      try {
        if (await pwdFormRef.value?.validate()) return false;
        const length = new TextEncoder().encode(formData.newPassword).length;
        if (length < 6 || length > 18) {
          Message.error("密码长度需在 6-18 字节");
          return false;
        }
        const token = userStore.token;
        await updatePassword({
          oldPassword: formData.oldPassword,
          newPassword: formData.newPassword,
        });
        handleClose();
        visible.value = false;
        if (userStore.token !== token) return false;
        Message.success("密码已修改,请重新登录");
        userStore.resetAuth();
        router.push(`/${getAdminEntry()}/login`);
      } catch (error) {
        reportError(error);
      } finally {
        savingPassword.value = false;
      }
      return false;
    };
    onMounted(async () => {
      r.value || (await userStore.fetchMe().catch(() => {}));
    });
    return {
      savingProfile,
      savingPassword,
      router,
      userStore,
      r,
      A,
      E,
      v,
      u,
      x,
      f,
      handleClick,
      handleUploadAvatar,
      S,
      handleClose,
      handleUpdatePassword,
      fileRef,
      visible,
      pwdFormRef,
      formData,
      entryPath,
    };
  },
};
</script>
<template>
  <div class="profile-page">
    <Breadcrumb class="crumb">
      <template #default>
        <BreadcrumbItem>
          <template #default>
            <a @click="(s) => $router.push(entryPath('/app'))">应用列表</a>
          </template>
        </BreadcrumbItem>
        <BreadcrumbItem><template #default>个人中心</template></BreadcrumbItem>
      </template>
    </Breadcrumb>
    <Card class="profile-card" :bordered="!1">
      <template #default>
        <div class="rows">
          <div class="row">
            <div class="label">头像</div>
            <div class="value">
              <div class="avatar-wrap" @click="handleClick">
                <Avatar :size="48" :image-url="r?.avatar || ''" :style="E">
                  <template #default>
                    <template v-if="r?.avatar"></template>
                    <template v-else>{{ A }}</template>
                  </template>
                </Avatar>
                <div class="avatar-hover"><IconCamera></IconCamera></div>
              </div>
              <input
                ref="fileRef"
                type="file"
                accept="image/jpeg,image/png,image/gif,image/webp"
                :style="{
                  display: 'none',
                }"
                @change="handleUploadAvatar"
              />
            </div>
          </div>
          <div class="row">
            <div class="label">昵称</div>
            <div class="value">
              <template v-if="v.nickname">
                <Input
                  :key="0"
                  v-model="u.nickname"
                  :disabled="savingProfile"
                  size="small"
                  :style="{
                    width: '240px',
                  }"
                  :max-length="32"
                  @pressEnter="(s) => f('nickname')"
                  @blur="(s) => f('nickname')"
                  allow-clear=""
                  autofocus=""
                ></Input>
              </template>
              <template v-else>
                <span :key="1" class="editable" @click="(s) => x('nickname')">
                  {{ (r?.nickname || "点击设置昵称") + " " }}
                  <IconEdit></IconEdit>
                </span>
              </template>
            </div>
          </div>
          <div class="row">
            <div class="label">账号</div>
            <div class="value">
              <template v-if="v.username">
                <Input
                  :key="0"
                  v-model="u.username"
                  :disabled="savingProfile"
                  size="small"
                  :style="{
                    width: '240px',
                  }"
                  :max-length="32"
                  @pressEnter="(s) => f('username')"
                  @blur="(s) => f('username')"
                  autofocus=""
                ></Input>
              </template>
              <template v-else>
                <span :key="1" class="editable" @click="(s) => x('username')">
                  {{ r?.username + " " }}
                  <IconEdit></IconEdit>
                </span>
              </template>
            </div>
          </div>
          <div class="row">
            <div class="label">密码</div>
            <div class="value">
              <span class="editable" @click="(s) => (visible = !0)">
                ********
                <IconEdit></IconEdit>
              </span>
            </div>
          </div>
        </div>
      </template>
    </Card>
    <Modal
      v-model:visible="visible"
      title="修改密码"
      :on-before-ok="handleUpdatePassword"
      :ok-loading="savingPassword"
      :on-before-cancel="() => !savingPassword"
      :closable="!savingPassword"
      :mask-closable="!savingPassword"
      :esc-to-close="!savingPassword"
      :cancel-button-props="{ disabled: savingPassword }"
      @close="handleClose"
      unmount-on-close=""
    >
      <template #default>
        <Form
          :model="formData"
          ref="pwdFormRef"
          layout="vertical"
          :disabled="savingPassword"
        >
          <template #default>
            <FormItem
              field="oldPassword"
              label="原密码"
              :rules="[
                {
                  required: !0,
                  message: '请输入原密码',
                },
              ]"
            >
              <template #default>
                <InputPassword
                  v-model="formData.oldPassword"
                  placeholder="请输入原密码"
                  allow-clear=""
                ></InputPassword>
              </template>
            </FormItem>
            <FormItem
              field="newPassword"
              label="新密码"
              :rules="[
                {
                  required: !0,
                  message: '请输入新密码',
                },
                {
                  minLength: 6,
                  message: '至少 6 位',
                },
                {
                  maxLength: 18,
                  message: '最多 18 位',
                },
              ]"
            >
              <template #default>
                <InputPassword
                  v-model="formData.newPassword"
                  placeholder="请输入新密码"
                  allow-clear=""
                ></InputPassword>
              </template>
            </FormItem>
            <FormItem
              field="confirmPassword"
              label="确认新密码"
              :rules="[
                {
                  required: !0,
                  message: '请再次输入新密码',
                },
                {
                  validator: S,
                },
              ]"
            >
              <template #default>
                <InputPassword
                  v-model="formData.confirmPassword"
                  placeholder="请再次输入新密码"
                  allow-clear=""
                ></InputPassword>
              </template>
            </FormItem>
          </template>
        </Form>
      </template>
    </Modal>
  </div>
</template>
<style scoped>
.profile-page {
  padding: 16px 24px 24px;
}
.crumb {
  margin-bottom: 12px;
}
.profile-card {
  max-width: 900px;
  background: var(--color-bg-2);
  border-radius: 8px;
}
.rows {
  display: flex;
  flex-direction: column;
}
.row {
  display: flex;
  align-items: center;
  padding: 20px 8px;
}
.label {
  width: 80px;
  color: var(--color-text-3);
  font-size: 14px;
  flex-shrink: 0;
}
.value {
  flex: 1;
  min-width: 0;
  font-size: 14px;
  color: var(--color-text-1);
}
.editable {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
  border-radius: 4px;
  padding: 4px 8px;
  margin-left: -8px;
  transition: background 0.15s;
}
.editable .arco-icon {
  color: var(--color-text-3);
  font-size: 14px;
  opacity: 0;
  transition: opacity 0.15s;
}
.editable:hover {
  background: var(--color-fill-2);
}
.editable:hover .arco-icon {
  opacity: 1;
}
.avatar-wrap {
  position: relative;
  display: inline-block;
  cursor: pointer;
  border-radius: 50%;
  overflow: hidden;
}
.avatar-wrap .avatar-hover {
  position: absolute;
  inset: 0;
  background: #00000080;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  opacity: 0;
  transition: opacity 0.15s;
}
.avatar-wrap:hover .avatar-hover {
  opacity: 1;
}
</style>
