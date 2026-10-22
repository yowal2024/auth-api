<!-- Reconstructed editable Vue source; see README for recovery limits. -->
<script>
import { onMounted, onUnmounted, computed, reactive, ref } from "vue";
import {
  Button,
  Row,
  Col,
  Card,
  Form,
  FormItem,
  Input,
  Select,
  Option,
  InputNumber,
  Textarea,
  Switch,
  Message,
} from "@arco-design/web-vue";
import { useRoute, useRouter } from "vue-router";
import {
  IconLeft,
  IconPlayArrow,
  IconSave,
} from "@arco-design/web-vue/es/icon";
import { getFunction, runFunction, saveFunction } from "@/api/function.js";
import { entryPath } from "@/utils/route.js";
import { editor as xq } from "@/utils/monaco.js";
export default {
  name: "FunctionEditor",
  components: {
    Button,
    IconLeft,
    IconPlayArrow,
    IconSave,
    Row,
    Col,
    Card,
    Form,
    FormItem,
    Input,
    Select,
    Option,
    InputNumber,
    Textarea,
    Switch,
  },
  setup(o) {
    const route = useRoute();
    const router = useRouter();
    const functionId = computed(() => Number(route.params.id || 0));
    const editorRef = ref();
    let editorInstance = null;
    let disposed = false;
    const ready = ref(false);
    const formData = reactive({
      id: 0,
      name: "",
      trigger: "manual",
      code: `// 在这里编写你的云函数
// 内置 API: db, http, crypto, app, user, ctx

async function handler(ctx) {
  // ctx.payload 包含触发时传入的数据
  return { ok: true, msg: 'hello from cloud function' }
}`,
      remark: "",
      timeout: 1e3,
      enabled: true,
    });
    const triggerOptions = [
      {
        value: "before_login",
        label: "登录前",
      },
      {
        value: "after_login",
        label: "登录后",
      },
      {
        value: "before_register",
        label: "注册前",
      },
      {
        value: "after_register",
        label: "注册后",
      },
      {
        value: "kami_use",
        label: "卡密使用",
      },
      {
        value: "manual",
        label: "手动调用",
      },
    ];
    const saving = ref(false);
    const loading = ref(false);
    const runOutput = ref("");
    const goBack = () => router.push(entryPath("/function"));
    const createEditor = () => {
      editorInstance = xq.create(editorRef.value, {
        value: formData.code,
        language: "javascript",
        theme: "vs-dark",
        fontSize: 13,
        minimap: {
          enabled: false,
        },
        automaticLayout: true,
        tabSize: 2,
        scrollBeyondLastLine: false,
      });
    };
    const loadFunction = async () => {
      if (functionId.value === 0) return;
      const _ = await getFunction(functionId.value);
      if (disposed) return;
      Object.assign(formData, _);
      formData.enabled = !!_.enabled;
      editorInstance?.setValue(formData.code || "");
    };
    const saveFunctionEntry = async () => {
      if (saving.value) return;
      if (!formData.name.trim()) {
        Message.warning("请填写函数名称");
        return;
      }
      formData.code = editorInstance?.getValue() || "";
      saving.value = true;
      try {
        await saveFunction(formData);
        Message.success("已保存");
        goBack();
      } catch {
        // Preserve the editor contents; the request layer displays the error.
      } finally {
        saving.value = false;
      }
    };
    const runFunctionCode = async () => {
      if (loading.value) return;
      formData.code = editorInstance?.getValue() || "";
      loading.value = true;
      try {
        const _ = await runFunction(
          functionId.value || 0,
          {},
          {
            code: formData.code,
            timeout: formData.timeout,
          },
        );
        runOutput.value = JSON.stringify(_, null, 2);
      } catch (_) {
        runOutput.value = `Error: ${_.message}`;
      } finally {
        loading.value = false;
      }
    };
    onMounted(async () => {
      createEditor();
      try {
        await loadFunction();
        ready.value = true;
      } catch {
        /* Keep saving disabled if loading fails. */
      }
    });
    onUnmounted(() => {
      disposed = true;
      const model = editorInstance?.getModel();
      editorInstance?.dispose();
      model?.dispose();
    });
    return {
      ready,
      route,
      router,
      functionId,
      editorInstance,
      triggerOptions,
      saving,
      runOutput,
      goBack,
      createEditor,
      loadFunction,
      saveFunctionEntry,
      runFunctionCode,
      loading,
      editorRef,
      formData,
    };
  },
};
</script>
<template>
  <div class="page">
    <div class="page-card">
      <div class="toolbar">
        <Button @click="goBack">
          <template #icon><IconLeft></IconLeft></template>
          <template #default>返回</template>
        </Button>
        <h3
          :style="{
            margin: '0',
          }"
        >
          {{ functionId === 0 ? "新建云函数" : `编辑：${formData.name || ""}` }}
        </h3>
        <div class="grow"></div>
        <Button @click="runFunctionCode" :loading="loading" :disabled="!ready">
          <template #icon><IconPlayArrow></IconPlayArrow></template>
          <template #default>试运行</template>
        </Button>
        <Button
          type="primary"
          @click="saveFunctionEntry"
          :loading="saving"
          :disabled="!ready"
        >
          <template #icon><IconSave></IconSave></template>
          <template #default>保存</template>
        </Button>
      </div>
      <Row :gutter="16">
        <template #default>
          <Col :span="14">
            <template #default>
              <Card
                title="代码（JavaScript / Goja）"
                :bordered="!1"
                :body-style="{
                  padding: '0',
                }"
              >
                <template #default>
                  <div
                    ref="editorRef"
                    :style="{
                      height: '480px',
                      width: '100%',
                    }"
                  ></div>
                </template>
              </Card>
            </template>
          </Col>
          <Col :span="10">
            <template #default>
              <Card title="基本信息" :bordered="!1">
                <template #default>
                  <Form :model="formData" layout="vertical">
                    <template #default>
                      <FormItem label="函数名">
                        <template #default
                          ><Input v-model="formData.name"></Input
                        ></template>
                      </FormItem>
                      <FormItem label="触发器">
                        <template #default>
                          <Select v-model="formData.trigger">
                            <template #default>
                              <template
                                v-for="ce in triggerOptions"
                                :key="ce.value"
                              >
                                <Option :value="ce.value">
                                  <template #default>{{ ce.label }}</template>
                                </Option>
                              </template>
                            </template>
                          </Select>
                        </template>
                      </FormItem>
                      <FormItem label="超时(ms)">
                        <template #default>
                          <InputNumber
                            v-model="formData.timeout"
                            :min="100"
                            :max="1e4"
                          ></InputNumber>
                        </template>
                      </FormItem>
                      <FormItem label="描述">
                        <template #default>
                          <Textarea
                            v-model="formData.remark"
                            :auto-size="{
                              minRows: 2,
                              maxRows: 4,
                            }"
                          ></Textarea>
                        </template>
                      </FormItem>
                      <FormItem label="启用">
                        <template #default
                          ><Switch v-model="formData.enabled"></Switch
                        ></template>
                      </FormItem>
                    </template>
                  </Form>
                </template>
              </Card>
              <Card
                title="试运行输出"
                :bordered="!1"
                :style="{
                  'margin-top': '12px',
                }"
              >
                <template #default>
                  <pre class="run-output">{{
                    runOutput || "点击「试运行」执行代码，输出将显示在这里"
                  }}</pre>
                </template>
              </Card>
            </template>
          </Col>
        </template>
      </Row>
    </div>
  </div>
</template>

<style scoped>
.toolbar h3 {
  margin-left: 8px;
}
.run-output {
  background: #1e1e1e;
  color: #d4d4d4;
  padding: 12px;
  border-radius: 4px;
  font-family:
    JetBrains Mono,
    Courier New,
    monospace;
  font-size: 12px;
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 240px;
  overflow: auto;
}
</style>
