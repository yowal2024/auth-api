<!-- Reconstructed editable Vue source; see README for recovery limits. -->
<script>
import { onMounted, ref, reactive } from "vue";
import {
  Notification,
  Form,
  FormItem,
  RadioGroup,
  Radio,
  Input,
  InputNumber,
  InputPassword,
  Select,
  Option,
  Alert,
  Spin,
  Button,
  Message,
} from "@arco-design/web-vue";
import { useRouter } from "vue-router";
import { IconSafe } from "@arco-design/web-vue/es/icon";
import {
  testMySQL,
  testRedis,
  saveInstallConfig,
  migrateDatabase,
  createAdmin,
  finalizeInstall,
} from "@/api/install.js";
import { clearInstallStatus, refreshInstallStatus } from "@/utils/install-status.js";
export default {
  name: "Install",
  components: {
    IconSafe,
    Form,
    FormItem,
    RadioGroup,
    Radio,
    Input,
    InputNumber,
    InputPassword,
    Select,
    Option,
    Alert,
    Spin,
    Button,
  },
  setup(ye) {
    const router = useRouter();
    onMounted(() => {
      Notification.warning({
        title: "程序未安装",
        content: "请填写下方信息完成系统初始化。",
        duration: 3e3,
        closable: true,
      });
    });
    const loading = ref(false);
    const m = ref("");
    const u = ref(null);
    const w = ref(false);
    const f = ref("fresh");
    const formData = reactive({
      mysql: {
        host: "127.0.0.1",
        port: 3306,
        user: "root",
        password: "",
        database: "auth",
        tablePrefix: "a_",
      },
      redis: {
        host: "127.0.0.1",
        port: 6379,
        db: 0,
        password: "",
      },
      admin: {
        entry: "admin",
        username: "admin",
        password: "",
        confirm: "",
      },
    });
    const x = () => ({
      addr: `${formData.redis.host}:${formData.redis.port}`,
      db: formData.redis.db,
      password: formData.redis.password,
    });
    const handleTestMySQL = async () => {
      if (((u.value = null), !formData.mysql.database)) {
        u.value = {
          code: 0,
          msg: "请填写数据库名称",
        };
        return;
      }
      if (f.value === "fresh") {
        if (!formData.admin.entry) {
          u.value = {
            code: 0,
            msg: "请填写管理员入口",
          };
          return;
        }
        const d = formData.admin.username;
        if (d.length < 5 || d.length > 12) {
          u.value = {
            code: 0,
            msg: "管理员账号需 5~12 位",
          };
          return;
        }
        const r = formData.admin.password;
        if (r.length < 6 || r.length > 18) {
          u.value = {
            code: 0,
            msg: "管理员密码需 6~18 位",
          };
          return;
        }
        if (r !== formData.admin.confirm) {
          u.value = {
            code: 0,
            msg: "两次密码不一致",
          };
          return;
        }
      }
      loading.value = true;
      const p = (d, r) => {
        u.value = {
          code: d,
          msg: r,
        };
        loading.value = false;
      };
      const e = x();
      m.value = "测试 MySQL 连接...";
      let t = await testMySQL({
        ...formData.mysql,
      });
      if (
        t.code !== 0 ||
        ((m.value = "测试 Redis 连接..."), (t = await testRedis(e)), t.code !== 0) ||
        ((m.value = "写入配置文件..."),
        (t = await saveInstallConfig({
          mysql: {
            ...formData.mysql,
          },
          redis: e,
        })),
        t.code !== 0) ||
        ((m.value = "执行数据库迁移..."), (t = await migrateDatabase()), t.code !== 0) ||
        (f.value === "fresh" &&
          ((m.value = "创建超级管理员..."),
          (t = await createAdmin({
            username: formData.admin.username,
            password: formData.admin.password,
            adminEntry: formData.admin.entry,
          })),
          t.code !== 0)) ||
        ((m.value = "完成安装..."), (t = await finalizeInstall()), t.code !== 0)
      )
        return p(t.code, t.msg);
      if (((w.value = true), clearInstallStatus(), t.data?.restartRequired)) {
        loading.value = false;
        m.value = "";
        u.value = {
          code: 0,
          msg: "系统已写入配置,但热切换失败,请在服务器手动重启 auth-api 进程后再登录",
        };
        return;
      }
      const v = await refreshInstallStatus(formData.admin.entry || "admin");
      const U = (v && v.adminEntry) || formData.admin.entry || "admin";
      Message.success("安装完成");
      router.push(`/${U}/login`);
    };
    return {
      router,
      m,
      u,
      w,
      f,
      x,
      handleTestMySQL,
      formData,
      loading,
    };
  },
};
</script>
<template>
  <div class="install-page">
    <div class="install-bg"></div>
    <div class="install-wrap">
      <div class="brand">
        <IconSafe :size="34" class="logo"></IconSafe>
        <h2>系统安装</h2>
        <p class="sub">欢迎使用 Auth 用户管理系统</p>
      </div>
      <Form :model="formData" layout="horizontal" auto-label-width="" label-align="right">
        <template #default>
          <h3 class="sec-title">
            <span class="bar"></span>
            安装方式
          </h3>
          <FormItem
            label="安装方式"
            tooltip="全新安装会建库迁移并创建超级管理员;升级安装仅补齐新增的数据库结构,不重建管理员"
          >
            <template #default>
              <RadioGroup v-model="f" type="button">
                <template #default>
                  <Radio value="fresh"><template #default>全新安装</template></Radio>
                  <Radio value="upgrade"><template #default>升级安装</template></Radio>
                </template>
              </RadioGroup>
            </template>
          </FormItem>
          <h3 class="sec-title">
            <span class="bar"></span>
            MySQL 配置
          </h3>
          <div class="grid-2">
            <FormItem label="数据库地址" required="">
              <template #default>
                <Input v-model="formData.mysql.host" placeholder="127.0.0.1"></Input>
              </template>
            </FormItem>
            <FormItem label="数据库端口" required="">
              <template #default>
                <InputNumber v-model="formData.mysql.port" :min="1" :max="65535"></InputNumber>
              </template>
            </FormItem>
            <FormItem label="数据库名称" required="">
              <template #default>
                <Input v-model="formData.mysql.database" placeholder="请输入数据库名称"></Input>
              </template>
            </FormItem>
            <FormItem label="数据库账号" required="">
              <template #default>
                <Input v-model="formData.mysql.user" placeholder="root"></Input>
              </template>
            </FormItem>
            <FormItem label="数据库密码">
              <template #default>
                <InputPassword
                  v-model="formData.mysql.password"
                  placeholder="留空表示无密码"
                ></InputPassword>
              </template>
            </FormItem>
            <FormItem
              label="表前缀"
              tooltip="所有业务表名会带这个前缀,用于跟其它系统共用 DB 时做隔离。默认 a_,留空表示不加前缀。安装后请勿再改。"
            >
              <template #default>
                <Input v-model="formData.mysql.tablePrefix" placeholder="a_"></Input>
              </template>
            </FormItem>
          </div>
          <h3 class="sec-title">
            <span class="bar"></span>
            Redis 配置
          </h3>
          <div class="grid-2">
            <FormItem label="redis 地址" required="">
              <template #default>
                <Input v-model="formData.redis.host" placeholder="127.0.0.1"></Input>
              </template>
            </FormItem>
            <FormItem label="redis 端口" required="">
              <template #default>
                <InputNumber v-model="formData.redis.port" :min="1" :max="65535"></InputNumber>
              </template>
            </FormItem>
            <FormItem label="redis 库">
              <template #default>
                <Select v-model="formData.redis.db" placeholder="0">
                  <template #default>
                    <template v-for="s in 16" :key="s - 1">
                      <Option :value="s - 1">
                        <template #default>{{ s - 1 }}</template>
                      </Option>
                    </template>
                  </template>
                </Select>
              </template>
            </FormItem>
            <FormItem label="redis 密码">
              <template #default>
                <InputPassword
                  v-model="formData.redis.password"
                  placeholder="默认空"
                ></InputPassword>
              </template>
            </FormItem>
          </div>
          <template v-if="f === 'fresh'">
            <h3 :key="0" class="sec-title">
              <span class="bar"></span>
              管理员账号
            </h3>
          </template>
          <template v-else></template>
          <template v-if="f === 'fresh'">
            <div :key="1" class="grid-2">
              <FormItem
                label="管理员入口"
                required=""
                tooltip="自定义后台登录入口名,可防止后台被他人扫描或攻击"
              >
                <template #default>
                  <Input v-model="formData.admin.entry" placeholder="admin"></Input>
                </template>
              </FormItem>
              <FormItem label="管理员账号" required="">
                <template #default>
                  <Input v-model="formData.admin.username" placeholder="管理员账号 5~12 位"></Input>
                </template>
              </FormItem>
              <FormItem label="登录密码" required="">
                <template #default>
                  <InputPassword
                    v-model="formData.admin.password"
                    placeholder="管理员密码 6~18 位"
                  ></InputPassword>
                </template>
              </FormItem>
              <FormItem label="确认密码" required="">
                <template #default>
                  <InputPassword
                    v-model="formData.admin.confirm"
                    placeholder="再次输入密码"
                  ></InputPassword>
                </template>
              </FormItem>
            </div>
          </template>
          <template v-else></template>
          <template v-if="u">
            <Alert :key="2" type="error" class="err-alert" :title="`错误码 ${u.code}`">
              <template #default>{{ u.msg }}</template>
            </Alert>
          </template>
          <template v-else></template>
          <template v-if="loading">
            <div :key="3" class="phase">
              <Spin :loading="!0" :size="14"></Spin>
              <span>{{ m }}</span>
            </div>
          </template>
          <template v-else></template>
          <Button
            type="primary"
            size="large"
            long=""
            class="submit"
            :loading="loading"
            :disabled="w"
            @click="handleTestMySQL"
          >
            <template #default>{{ w ? "已安装" : "确认安装" }}</template>
          </Button>
        </template>
      </Form>
    </div>
  </div>
</template>
<style scoped>
.install-page {
  min-height: 100vh;
  position: relative;
  padding: 64px 24px 48px;
  overflow: hidden;
}
.install-bg {
  position: absolute;
  inset: 0;
  background:
    radial-gradient(circle at 0% 0%, rgba(22, 93, 255, 0.06), transparent 40%),
    radial-gradient(circle at 100% 100%, rgba(22, 93, 255, 0.05), transparent 40%), #f5f7fa;
  z-index: 0;
}
.install-wrap {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 1000px;
  margin: 0 auto;
  background: #fff;
  border-radius: 12px;
  border: 1px solid var(--color-border-2);
  box-shadow: 0 8px 32px #0f172a0f;
  padding: 28px 32px 32px;
}
.brand {
  text-align: center;
  margin: 0 0 24px;
}
.brand .logo {
  color: rgb(var(--primary-6));
}
.brand h2 {
  margin: 6px 0 4px;
  font-size: 22px;
  color: #1f2d3d;
}
.brand .sub {
  margin: 0;
  color: var(--color-text-3);
  font-size: 13px;
}
.sec-title {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 18px 0 16px;
  padding: 0;
  font-size: 15px;
  color: var(--color-text-1);
  font-weight: 600;
}
.sec-title .bar {
  width: 3px;
  height: 14px;
  background: rgb(var(--primary-6));
  border-radius: 2px;
}
.err-alert {
  margin: 4px 0 12px;
}
.phase {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0 0 12px;
  color: var(--color-text-2);
  font-size: 13px;
}
.submit {
  margin-top: 8px;
}
.tip {
  margin: 6px 0 0;
  color: var(--color-text-3);
  font-size: 12px;
}
:deep(.arco-input-number) {
  width: 100%;
}
.grid-2 {
  display: grid;
  grid-template-columns: 1fr;
  column-gap: 20px;
}
@media (min-width: 768px) {
  .grid-2 {
    grid-template-columns: 1fr 1fr;
  }
}
</style>
