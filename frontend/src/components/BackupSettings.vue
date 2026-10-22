<script>
import { ref, reactive, onMounted, onBeforeUnmount } from "vue";
import {
  Form,
  FormItem,
  Switch,
  Checkbox,
  InputNumber,
  Button,
  Space,
  Table,
  Tag,
  Message,
  Popconfirm,
  Tooltip,
} from "@arco-design/web-vue";
import { request } from "@/utils/request.js";
import { formatDateTime } from "@/utils/format.js";
export default {
  name: "BackupSettings",
  components: {
    Form,
    FormItem,
    Switch,
    Checkbox,
    InputNumber,
    Button,
    Space,
    Table,
    Tag,
    Popconfirm,
    Tooltip,
  },
  setup() {
    const form = reactive({
      enabled: false,
      intervalHours: 24,
      keepCount: 7,
      includeFiles: true,
    });
    const rows = ref([]),
      loaded = ref(false),
      saving = ref(false),
      starting = ref(false),
      fetching = ref(false),
      downloading = ref("");
    let timer,
      disposed = false;
    const load = async (settings = false) => {
      if (fetching.value) return;
      fetching.value = true;
      try {
        const v = await request("/admin/system/backup");
        if (disposed) return;
        rows.value = v.runs || [];
        if (settings || !loaded.value) Object.assign(form, v.settings);
        loaded.value = true;
      } catch {
      } finally {
        fetching.value = false;
      }
    };
    const save = async () => {
      saving.value = true;
      try {
        await request("/admin/system/backup/settings", {
          method: "POST",
          data: { ...form },
        });
        Message.success("备份设置已保存");
      } catch {
      } finally {
        saving.value = false;
      }
    };
    const run = async () => {
      starting.value = true;
      try {
        await request("/admin/system/backup/run", { method: "POST" });
        Message.success("备份任务已开始");
        await load();
      } catch {
      } finally {
        starting.value = false;
      }
    };
    const download = async (id) => {
      downloading.value = id;
      try {
        const response = await request(`/admin/system/backup/${id}/download`, {
          responseType: "blob",
          timeout: 0,
        });
        const url = URL.createObjectURL(response.data),
          a = document.createElement("a");
        a.href = url;
        a.download = `${id}.tar.gz`;
        a.click();
        setTimeout(() => URL.revokeObjectURL(url), 1000);
      } catch {
      } finally {
        downloading.value = "";
      }
    };
    onMounted(() => {
      load(true);
      timer = setInterval(() => {
        if (!document.hidden) load();
      }, 5000);
    });
    onBeforeUnmount(() => {
      disposed = true;
      clearInterval(timer);
    });
    const labels = {
      running: "备份中",
      complete: "已完成",
      failed: "失败",
      expired: "已按保留策略清理",
    };
    const columns = [
      { title: "开始时间", slotName: "startedAt", minWidth: 190 },
      { title: "完成时间", slotName: "finishedAt", minWidth: 190 },
      { title: "状态", slotName: "status", width: 160 },
      { title: "大小", slotName: "size", width: 110 },

      { title: "操作", slotName: "actions", width: 90 },
    ];
    return {
      form,
      rows,
      loaded,
      saving,
      starting,
      fetching,
      downloading,
      save,
      run,
      load,
      download,
      labels,
      columns,
      formatDateTime,
    };
  },
};
</script>
<template>
  <div class="backup-settings">
    <section class="backup-policy">
      <div class="policy-heading">
        <div class="policy-title">
          <span class="section-title">自动备份</span>
          <Switch
            v-model="form.enabled"
            :disabled="!loaded"
            size="small"
            aria-label="自动备份"
          />
          <span class="policy-caption">开启后按设定间隔执行，保存后生效</span>
        </div>
        <Button
          type="primary"
          :loading="saving"
          :disabled="!loaded"
          @click="save"
          >保存备份设置</Button
        >
      </div>
      <Form :model="form" layout="vertical" class="backup-form">
        <FormItem label="备份间隔"
          ><InputNumber
            v-model="form.intervalHours"
            :min="1"
            :max="720"
            :precision="0"
            :disabled="!loaded"
            ><template #suffix>小时</template></InputNumber
          ></FormItem
        >
        <FormItem label="保留最近成功备份"
          ><InputNumber
            v-model="form.keepCount"
            :min="1"
            :max="30"
            :precision="0"
            :disabled="!loaded"
            ><template #suffix>份</template></InputNumber
          ></FormItem
        >
        <FormItem label="备份内容" class="content-field"
          ><div class="backup-content">
            <Tooltip content="每份备份始终包含数据库、运行配置和已生成的主密钥">
              <Checkbox :model-value="true" disabled>数据库与配置</Checkbox>
            </Tooltip>
            <Checkbox
              v-model="form.includeFiles"
              :disabled="!loaded"
              aria-label="同时备份上传文件"
              >上传文件</Checkbox
            >
          </div></FormItem
        >
      </Form>
      <div class="policy-footer">
        <span class="backup-dot"></span
        ><span>备份仅保存在本机，建议定期下载到其他设备。</span
        ><Tooltip
          content="始终包含数据库、运行配置和已生成的主密钥；上传文件包含头像、扩展文件和原始 APK，不包含差异补丁和 IP 库。备份含有凭据，请妥善保管，恢复步骤见项目 README。"
          :content-style="{ maxWidth: '340px', lineHeight: '1.7' }"
          ><Button type="text" size="mini">备份说明</Button></Tooltip
        >
      </div>
    </section>
    <div class="records-heading">
      <div class="records-title">
        <span class="section-title">备份记录</span
        ><span class="policy-caption">按已保存的配置执行</span>
      </div>
      <Space :size="8"
        ><Button :loading="fetching" @click="load()">刷新</Button
        ><Popconfirm content="按已保存的配置立即备份？" @ok="run"
          ><Button
            type="outline"
            :loading="starting"
            :disabled="!loaded || rows.some((r) => r.status === 'running')"
            >立即备份</Button
          ></Popconfirm
        ></Space
      >
    </div>
    <Table
      :data="rows"
      :columns="columns"
      row-key="id"
      :scroll="{ x: 760 }"
      :bordered="false"
      size="medium"
      :pagination="rows.length > 10 ? { pageSize: 10 } : false"
    >
      <template #startedAt="{ record }">{{
        formatDateTime(record.startedAt)
      }}</template>
      <template #finishedAt="{ record }">{{
        record.finishedAt ? formatDateTime(record.finishedAt) : "—"
      }}</template>
      <template #status="{ record }"
        ><Tag
          :color="
            record.status === 'complete'
              ? 'green'
              : record.status === 'failed'
                ? 'red'
                : 'blue'
          "
          >{{ labels[record.status] || record.status }}</Tag
        ><Tooltip v-if="record.error" :content="record.error"
          ><Button type="text" size="mini">原因</Button></Tooltip
        ></template
      >
      <template #size="{ record }">{{
        record.size ? (record.size / 1048576).toFixed(2) + " MB" : "—"
      }}</template>
      <template #actions="{ record }"
        ><Button
          v-if="record.status === 'complete'"
          type="text"
          :loading="downloading === record.id"
          :disabled="!!downloading"
          @click="download(record.id)"
          >下载</Button
        ></template
      >
    </Table>
  </div>
</template>
<style scoped>
.backup-settings {
  color: var(--color-text-1);
}
.backup-policy {
  border: 1px solid var(--color-border-2);
  border-radius: 8px;
  overflow: hidden;
  margin-bottom: 24px;
}
.policy-heading {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  padding: 16px 20px;
  background: var(--color-fill-1);
}
.policy-title,
.records-title {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}
.section-title {
  font-size: 15px;
  font-weight: 600;
  white-space: nowrap;
}
.policy-caption {
  font-size: 12px;
  color: var(--color-text-3);
}
.backup-form {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) minmax(300px, 1.3fr);
  gap: 24px;
  padding: 20px 20px 16px;
}
.backup-form :deep(.arco-form-item) {
  margin-bottom: 0;
}
.backup-form :deep(.arco-input-number) {
  width: 100%;
}
.backup-content {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  flex-wrap: wrap;
  gap: 8px 24px;
  min-height: 32px;
  width: 100%;
  box-sizing: border-box;
  padding: 5px 12px;
  background: var(--color-fill-2);
  border-radius: var(--border-radius-small);
  font-size: 13px;
}
.backup-content :deep(.arco-checkbox) {
  margin-right: 0;
  white-space: nowrap;
}
.backup-content :deep(.arco-checkbox-disabled .arco-checkbox-label) {
  color: var(--color-text-2);
}
.policy-footer {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
  padding: 0 20px 14px;
  color: var(--color-text-3);
  font-size: 12px;
}
.backup-dot {
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: rgb(var(--primary-6));
  margin-right: 3px;
  flex-shrink: 0;
}
.records-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 14px;
}
.backup-settings :deep(.arco-table-th),
.backup-settings :deep(.arco-table-td) {
  font-variant-numeric: tabular-nums;
}
@media (max-width: 1100px) {
  .backup-form {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 16px 24px;
  }
  .content-field {
    grid-column: 1 / -1;
  }
  .backup-content {
    justify-content: flex-start;
    gap: 24px;
  }
}
@media (max-width: 640px) {
  .policy-heading {
    padding: 14px;
    align-items: flex-start;
    flex-wrap: wrap;
  }
  .policy-title {
    gap: 10px;
  }
  .policy-title .policy-caption {
    flex-basis: 100%;
  }
  .backup-form {
    padding: 16px 14px;
    gap: 16px 12px;
  }
  .policy-footer {
    padding: 0 14px 12px;
  }
  .backup-dot {
    display: none;
  }
  .records-heading {
    flex-wrap: wrap;
  }
  .records-title .policy-caption {
    display: none;
  }
}
</style>
