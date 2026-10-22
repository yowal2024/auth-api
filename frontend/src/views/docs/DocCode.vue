<script>
import { ref, onBeforeUnmount } from "vue";
import { Button } from "@arco-design/web-vue";
import { IconCopy } from "@arco-design/web-vue/es/icon";
export default {
  name: "DocCode",
  components: { Button, IconCopy },
  setup() {
    const block = ref(null);
    const status = ref("复制");
    let timer;
    async function copy() {
      try {
        await navigator.clipboard.writeText(block.value.textContent);
        status.value = "已复制";
      } catch {
        status.value = "复制失败，请手动选择";
      }
      clearTimeout(timer);
      timer = setTimeout(() => {
        status.value = "复制";
      }, 2000);
    }
    onBeforeUnmount(() => clearTimeout(timer));
    return { block, status, copy };
  },
};
</script>
<template>
  <div class="doc-code">
    <div class="code-tools">
      <span
        ><span class="code-symbol" aria-hidden="true">&lt;/&gt;</span>
        示例代码</span
      ><Button type="secondary" size="mini" @click="copy" aria-live="polite">
        <template #icon><IconCopy /></template>
        {{ status }}
      </Button>
    </div>
    <pre ref="block"><slot /></pre>
  </div>
</template>
<style scoped>
.doc-code {
  margin: 14px 0 20px;
  overflow: hidden;
  border: 1px solid #26364e;
  border-radius: 10px;
  background: #101d32;
  color: #dfe9fa;
}
.code-tools {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 15px;
  border-bottom: 1px solid #ffffff14;
  color: #9dacc4;
  font-size: 11px;
  letter-spacing: 0.3px;
  background: #ffffff03;
}
.code-symbol {
  color: #79a7ff;
  margin-right: 7px;
  font-family: monospace;
}

pre {
  overflow: auto;
  margin: 0;
  padding: 16px 18px 20px;
  max-height: 560px;
  tab-size: 2;
  line-height: 1.85;
}
pre :slotted(code) {
  font-family: Consolas, "SFMono-Regular", monospace;
  font-size: 12px;
  color: #dfe9fa;
  background: transparent;
  padding: 0;
  white-space: pre;
  overflow-wrap: normal;
}

.code-tools :deep(.arco-btn) {
  color: #dfe9fa;
  background: #ffffff12;
}
</style>
