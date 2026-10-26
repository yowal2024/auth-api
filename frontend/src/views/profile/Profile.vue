<!-- Reconstructed editable Vue source; see README for recovery limits. -->
<script>
import { reportError } from "@/utils/errors.js";
import { onMounted, computed, ref, reactive } from "vue";
import {
  Breadcrumb,
  Button,
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
import {
  IconCamera,
  IconEdit,
  IconUser,
  IconSafe,
  IconLock,
} from "@arco-design/web-vue/es/icon";
import SettingsHeader from "@/components/SettingsHeader.vue";
import { getAdminEntry } from "@/utils/install-status.js";
import { useUserStore } from "@/stores/user.js";
import { uploadAvatar, updatePassword } from "@/api/auth.js";
import { entryPath } from "@/utils/route.js";
export default {
  name: "Profile",
  components: {
    Breadcrumb,
    BreadcrumbItem,
    Button,
    SettingsHeader,
    IconUser,
    IconSafe,
    IconLock,
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
    const roleName = computed(() =>
      !r.value ? "读取中" : r.value.role === "super" ? "超级管理员" : "管理员",
    );
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
    const cancelEdit = (field) => {
      if (!savingProfile.value) v[field] = false;
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
      roleName,
      cancelEdit,
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
  <div class="profile-page settings-surface">
    <Breadcrumb class="crumb"
      ><BreadcrumbItem
        ><a @click="$router.push(entryPath('/app'))"
          >应用列表</a
        ></BreadcrumbItem
      ><BreadcrumbItem>个人中心</BreadcrumbItem></Breadcrumb
    >
    <div class="profile-shell">
      <SettingsHeader
        title="个人中心"
        description="管理你的个人资料、登录账号与密码。"
        note="当前账号"
        ><IconUser
      /></SettingsHeader>
      <div class="profile-layout">
        <aside class="profile-identity">
          <button
            type="button"
            class="avatar-wrap"
            aria-label="选择头像"
            :disabled="savingProfile || !r"
            @click="handleClick"
          >
            <Avatar
              :size="88"
              shape="square"
              :image-url="r?.avatar || ''"
              :style="E"
              ><template v-if="!r?.avatar">{{ A }}</template></Avatar
            >
            <span class="avatar-hover"><IconCamera /></span>
          </button>
          <h3>{{ r?.nickname || r?.username || "正在读取…" }}</h3>
          <span class="profile-role">{{ roleName }}</span>
          <div class="profile-account">
            <span>登录账号</span><strong>{{ r?.username || "—" }}</strong>
          </div>
          <Button :disabled="savingProfile || !r" @click="handleClick"
            ><template #icon><IconCamera /></template>更换头像</Button
          >
          <p class="profile-avatar-hint">
            支持 JPG、PNG、GIF、WebP<br />图片大小不超过 5 MB
          </p>
          <input
            ref="fileRef"
            type="file"
            accept="image/jpeg,image/png,image/gif,image/webp"
            hidden
            @change="handleUploadAvatar"
          />
        </aside>
        <div class="profile-details">
          <section class="settings-panel">
            <div class="settings-panel-heading">
              <IconUser />
              <h3>基本资料</h3>
              <span>管理展示名称与登录账号</span>
            </div>
            <div class="profile-rows">
              <div class="row">
                <div class="label">昵称</div>
                <div class="value">
                  <div v-if="v.nickname" class="profile-editor">
                    <Input
                      v-model="u.nickname"
                      aria-label="昵称"
                      :disabled="savingProfile"
                      :max-length="32"
                      @press-enter="f('nickname')"
                      @keydown.esc="cancelEdit('nickname')"
                      allow-clear
                      autofocus
                    />
                    <div class="profile-edit-actions">
                      <Button
                        type="primary"
                        :loading="savingProfile"
                        @click="f('nickname')"
                        >保存</Button
                      ><Button
                        :disabled="savingProfile"
                        @click="cancelEdit('nickname')"
                        >取消</Button
                      >
                    </div>
                  </div>
                  <button
                    v-else
                    type="button"
                    class="editable"
                    aria-label="修改昵称"
                    :disabled="savingProfile || !r"
                    @click="x('nickname')"
                  >
                    <span class="field-text">{{
                      r?.nickname || "点击设置昵称"
                    }}</span
                    ><span class="edit-action"><IconEdit />修改</span>
                  </button>
                  <p class="profile-field-hint">
                    用于后台展示，最多 32 个字符。
                  </p>
                </div>
              </div>
              <div class="row">
                <div class="label">账号</div>
                <div class="value">
                  <div v-if="v.username" class="profile-editor">
                    <Input
                      v-model="u.username"
                      aria-label="账号"
                      :disabled="savingProfile"
                      :max-length="32"
                      @press-enter="f('username')"
                      @keydown.esc="cancelEdit('username')"
                      autofocus
                    />
                    <div class="profile-edit-actions">
                      <Button
                        type="primary"
                        :loading="savingProfile"
                        @click="f('username')"
                        >保存</Button
                      ><Button
                        :disabled="savingProfile"
                        @click="cancelEdit('username')"
                        >取消</Button
                      >
                    </div>
                  </div>
                  <button
                    v-else
                    type="button"
                    class="editable"
                    aria-label="修改账号"
                    :disabled="savingProfile || !r"
                    @click="x('username')"
                  >
                    <span class="field-text">{{ r?.username || "—" }}</span
                    ><span class="edit-action"><IconEdit />修改</span>
                  </button>
                  <p class="profile-field-hint">
                    用于登录后台，长度为 5–12 字节。
                  </p>
                </div>
              </div>
            </div>
            <p class="settings-note">
              修改后点击保存，或按 Enter 提交；取消可放弃本次编辑。
            </p>
          </section>
          <section class="settings-panel">
            <div class="settings-panel-heading">
              <IconSafe />
              <h3>账号安全</h3>
              <span>管理登录密码</span>
            </div>
            <div class="profile-rows">
              <div class="row">
                <div class="label">密码</div>
                <div class="value">
                  <button
                    type="button"
                    class="editable"
                    aria-label="修改密码"
                    :disabled="!r || savingPassword"
                    @click="visible = true"
                  >
                    <span class="password-mask">********</span
                    ><span class="edit-action"><IconLock />修改密码</span>
                  </button>
                  <p class="profile-field-hint">
                    修改时需验证原密码，新密码长度为 6–18 字节。
                  </p>
                </div>
              </div>
            </div>
            <p class="settings-note">
              <IconSafe />密码修改成功后会退出当前会话，请使用新密码重新登录。
            </p>
          </section>
        </div>
      </div>
    </div>
    <Modal
      v-model:visible="visible"
      title="修改密码"
      modal-class="settings-dialog"
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
                  autocomplete="current-password"
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
                  autocomplete="new-password"
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
                  autocomplete="new-password"
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
.profile-shell {
  padding: 16px 24px 24px;
  background: var(--color-bg-2);
  border: 1px solid var(--color-border-1);
}
.profile-layout {
  display: grid;
  grid-template-columns: 300px minmax(0, 1fr);
  gap: 24px;
  align-items: stretch;
}
.profile-identity {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 32px 24px;
  border: 1px solid var(--color-border-2);
  background: var(--color-fill-1);
  min-width: 0;
}
.avatar-wrap {
  position: relative;
  display: block;
  padding: 0;
  border: 0;
  background: transparent;
  cursor: pointer;
}
.avatar-wrap :deep(.arco-avatar) {
  border-radius: 0;
}
.avatar-hover {
  position: absolute;
  inset: 0;
  display: grid;
  place-items: center;
  background: #00000080;
  color: #fff;
  font-size: 24px;
  opacity: 0;
  transition: opacity 0.15s;
}
.avatar-wrap:hover .avatar-hover,
.avatar-wrap:focus-visible .avatar-hover {
  opacity: 1;
}
.avatar-wrap:focus-visible,
.editable:focus-visible {
  outline: 2px solid rgb(var(--primary-6));
  outline-offset: 3px;
}
.profile-identity h3 {
  max-width: 100%;
  margin: 20px 0 10px;
  overflow-wrap: anywhere;
  text-align: center;
  color: var(--color-text-1);
  font-size: 20px;
  line-height: 1.5;
  font-weight: 600;
}
.profile-role {
  padding: 4px 10px;
  color: rgb(var(--primary-6));
  background: var(--color-primary-light-1);
  font-size: 12px;
  line-height: 1.5;
}
.profile-account {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  flex-wrap: wrap;
  gap: 8px;
  width: 100%;
  margin: 26px 0;
  padding: 18px 0;
  border-top: 1px solid var(--color-border-2);
  border-bottom: 1px solid var(--color-border-2);
  font-size: 13px;
}
.profile-account span {
  color: var(--color-text-3);
}
.profile-account strong {
  color: var(--color-text-2);
  font-weight: 500;
  overflow-wrap: anywhere;
}
.profile-avatar-hint {
  margin: 14px 0 0;
  text-align: center;
  color: var(--color-text-3);
  font-size: 12px;
  line-height: 1.9;
}
.profile-details {
  display: grid;
  align-content: start;
  gap: 20px;
  min-width: 0;
}
.profile-rows {
  padding: 0 22px;
}
.row {
  display: flex;
  align-items: flex-start;
  gap: 20px;
  padding: 22px 0;
}
.row + .row {
  border-top: 1px solid var(--color-border-1);
}
.label {
  width: 80px;
  flex-shrink: 0;
  color: var(--color-text-3);
  font-size: 14px;
  line-height: 36px;
}
.value {
  flex: 1;
  min-width: 0;
  font-size: 14px;
  color: var(--color-text-1);
}
.editable {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  width: 100%;
  min-height: 36px;
  padding: 0;
  border: 0;
  background: transparent;
  color: inherit;
  font: inherit;
  text-align: left;
  cursor: pointer;
}
.editable:disabled,
.avatar-wrap:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}
.edit-action {
  display: inline-flex;
  align-items: center;
  flex-shrink: 0;
  gap: 6px;
  color: rgb(var(--primary-6));
  font-size: 13px;
}
.editable:hover .edit-action {
  text-decoration: underline;
  text-underline-offset: 4px;
}
.field-text {
  min-width: 0;
  overflow-wrap: anywhere;
}
.password-mask {
  letter-spacing: 3px;
  color: var(--color-text-2);
}
.profile-field-hint {
  margin: 5px 0 0;
  color: var(--color-text-3);
  font-size: 12px;
  line-height: 1.7;
}
.profile-editor {
  display: flex;
  align-items: flex-start;
  flex-wrap: wrap;
  gap: 10px;
}
.profile-editor :deep(.arco-input-wrapper) {
  flex: 1 1 180px;
  width: auto;
  max-width: 420px;
}
.profile-edit-actions {
  display: flex;
  gap: 8px;
}
@media (max-width: 1100px) {
  .profile-layout {
    grid-template-columns: 240px minmax(0, 1fr);
    gap: 20px;
  }
  .profile-identity {
    padding: 28px 20px;
  }
  .row {
    gap: 12px;
  }
  .label {
    width: 50px;
  }
}
@media (max-width: 760px) {
  .profile-page {
    padding: 12px;
  }
  .profile-shell {
    padding: 12px;
  }
  .profile-layout {
    grid-template-columns: minmax(0, 1fr);
  }
  .profile-identity {
    padding: 24px;
  }
  .profile-account {
    max-width: 280px;
    margin: 18px 0;
  }
  .profile-rows {
    padding: 0 16px;
  }
  .row {
    flex-direction: column;
    gap: 4px;
    padding: 18px 0;
  }
  .label {
    width: auto;
    line-height: 1.5;
  }
  .value {
    width: 100%;
  }
  .profile-editor :deep(.arco-input-wrapper) {
    max-width: none;
    flex-basis: 100%;
  }
}
@media (prefers-reduced-motion: reduce) {
  .avatar-hover {
    transition: none;
  }
}
</style>
