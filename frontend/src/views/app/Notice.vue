<!-- Reconstructed editable Vue source; see README for recovery limits. -->
<script>
import { useAppStore } from "@/stores/app.js";
import { listNotices, saveNotice, deleteNotice } from "@/api/app.js";
import { formatDateTime } from "@/utils/format.js";
import { watch, onMounted, reactive, ref } from "vue";
import { Input, Textarea, Button, Spin, Empty, Popconfirm, Message } from "@arco-design/web-vue";
import { IconDelete, IconUser } from "@arco-design/web-vue/es/icon";
export default {
  name: "Notice",
  components: {
    Input,
    Textarea,
    Button,
    Spin,
    Popconfirm,
    IconDelete,
    IconUser,
    Empty,
  },
  setup(rt) {
    const appStore = useAppStore();
    const d = ref([]);
    const loadingState = ref(false);
    const loading = ref(false);
    const e = reactive({
      title: "",
      content: "",
    });
    const handleListNotices = async () => {
      loadingState.value = true;
      try {
        d.value = await listNotices({
          appid: appStore.currentId,
        });
      } finally {
        loadingState.value = false;
      }
    };
    const handleSaveNotice = async () => {
      if (!e.content.trim() && !e.title.trim()) {
        Message.warning("请填写公告内容");
        return;
      }
      loading.value = true;
      try {
        await saveNotice({
          title: e.title.trim(),
          content: e.content.trim(),
          appid: appStore.currentId,
        });
        Message.success("已发布");
        e.title = "";
        e.content = "";
        handleListNotices();
      } finally {
        loading.value = false;
      }
    };
    const handleDeleteNotice = async (v) => {
      await deleteNotice(v.id);
      Message.success("已删除");
      handleListNotices();
    };
    watch(() => appStore.currentId, handleListNotices);
    onMounted(handleListNotices);
    return {
      appStore,
      d,
      loadingState,
      e,
      handleListNotices,
      handleSaveNotice,
      handleDeleteNotice,
      loading,
      formatDateTime,
    };
  },
};
</script>
<template>
  <div class="page notice-page">
    <div class="publisher">
      <Input v-model="e.title" placeholder="公告标题" class="title-input" :max-length="60"></Input>
      <Textarea
        v-model="e.content"
        placeholder="发布公告内容...."
        :auto-size="{
          minRows: 4,
          maxRows: 10,
        }"
        class="content-area"
      ></Textarea>
      <div class="publisher-actions">
        <div class="grow"></div>
        <Button type="primary" :loading="loading" @click="handleSaveNotice">
          <template #default>发布</template>
        </Button>
      </div>
    </div>
    <Spin :loading="loadingState" class="feed-spin">
      <template #default>
        <template v-if="d.length">
          <div :key="1" class="notice-feed">
            <template v-for="t in d" :key="t.id">
              <div class="notice-item">
                <div class="notice-header">
                  <div class="notice-title">{{ t.title || t.content?.slice(0, 30) }}</div>
                  <Popconfirm
                    content="确定删除该公告？"
                    type="warning"
                    @ok="(dt) => handleDeleteNotice(t)"
                  >
                    <template #default>
                      <button class="del-btn"><IconDelete></IconDelete></button>
                    </template>
                  </Popconfirm>
                </div>
                <template v-if="t.title &amp;&amp; t.content">
                  <div :key="0" class="notice-content">{{ t.content }}</div>
                </template>
                <template v-else></template>
                <div class="notice-meta">
                  <span class="meta-item">
                    <IconUser></IconUser>
                    {{ " " + (t.author || "admin") }}
                  </span>
                  <span class="meta-item">{{ formatDateTime(t.updateTime) }}</span>
                  <span class="meta-item">{{ (t.views || 0) + "次浏览" }}</span>
                </div>
              </div>
            </template>
          </div>
        </template>
        <template v-else>
          <Empty
            :key="0"
            :style="{
              padding: '60px 0',
            }"
          ></Empty>
        </template>
      </template>
    </Spin>
  </div>
</template>
<style scoped>
.notice-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.publisher {
  background: var(--color-bg-2);
  border-radius: 6px;
  padding: 16px;
  box-shadow: 0 1px 2px #0000000a;
}
.publisher .title-input {
  margin-bottom: 10px;
}
.publisher .title-input :deep(.arco-input-wrapper) {
  background: transparent;
  border: none;
  padding: 0;
  font-size: 16px;
  font-weight: 600;
}
.publisher .content-area :deep(.arco-textarea-wrapper) {
  background: var(--color-fill-2);
  border-color: transparent;
}
.publisher .publisher-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 14px;
}
.publisher .publisher-actions .grow {
  flex: 1;
}
.feed-spin {
  width: 100%;
}
.notice-feed {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.notice-item {
  background: var(--color-bg-2);
  border-radius: 6px;
  box-shadow: 0 1px 2px #0000000a;
  padding: 18px 20px;
}
.notice-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}
.notice-title {
  font-size: 15px;
  font-weight: 500;
  color: var(--color-text-1);
  line-height: 1.5;
  word-break: break-all;
}
.del-btn {
  background: transparent;
  border: none;
  cursor: pointer;
  color: var(--color-text-3);
  padding: 2px 4px;
  display: inline-flex;
  font-size: 16px;
}
.del-btn:hover {
  color: rgb(var(--danger-6));
}
.notice-content {
  margin-top: 6px;
  font-size: 13px;
  color: var(--color-text-2);
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-all;
}
.notice-meta {
  margin-top: 14px;
  padding-top: 12px;
  border-top: 1px solid var(--color-neutral-3);
  display: flex;
  flex-wrap: wrap;
  gap: 14px;
  font-size: 12px;
  color: var(--color-text-3);
}
.notice-meta .meta-item {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}
</style>
