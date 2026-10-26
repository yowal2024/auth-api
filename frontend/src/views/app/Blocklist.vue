<!-- Reconstructed editable Vue source; see README for recovery limits. -->
<script>
import { keepModalOpen } from "@/utils/modal.js";
import { useAppStore } from "@/stores/app.js";
import { listBlocks, deleteBlock, createBlock } from "@/api/app.js";
import { formatDateTime } from "@/utils/format.js";
import { watch, onMounted, reactive, computed, ref } from "vue";
import {
  Button,
  Select,
  Input,
  Table,
  Modal,
  Option,
  Popconfirm,
  Link,
  Tag,
  Form,
  FormItem,
  Textarea,
  Message,
} from "@arco-design/web-vue";
import { IconPlus, IconSearch } from "@arco-design/web-vue/es/icon";
export default {
  name: "Blocklist",
  components: {
    Button,
    IconPlus,
    Select,
    Option,
    Input,
    IconSearch,
    Table,
    Tag,
    Popconfirm,
    Link,
    Modal,
    Form,
    FormItem,
    Textarea,
  },
  setup(me) {
    const appStore = useAppStore();
    const r = reactive({
      type: void 0,
      keyword: "",
    });
    const rows = ref([]);
    const loading = ref(false);
    const visible = ref(false);
    const formData = reactive({
      type: "ip",
      value: "",
      reason: "",
    });
    const pagination = computed(() => ({
      total: rows.value.length,
      showTotal: true,
      showPageSize: true,
      pageSizeOptions: [10, 20, 50, 100],
    }));
    const tableColumns = [
      {
        title: "应用内编号",
        dataIndex: "appRecordId",
        width: 110,
      },
      {
        title: "类型",
        slotName: "type",
        width: 100,
      },
      {
        title: "值",
        slotName: "value",
        width: 240,
        ellipsis: true,
        tooltip: true,
      },
      {
        title: "原因",
        dataIndex: "reason",
        width: 280,
        ellipsis: true,
        tooltip: true,
      },
      {
        title: "封禁时间",
        slotName: "createTime",
        width: 200,
      },
      {
        title: "操作",
        slotName: "operations",
        fixed: "right",
        width: 100,
      },
    ];
    const T = computed(
      () =>
        ({
          ip: "192.168.1.1 或 192.168.0.0/16",
          mi: "机器码 hash",
          region: "广东省 / 美国",
        })[formData.type],
    );
    const handleListBlocks = async () => {
      loading.value = true;
      try {
        rows.value = await listBlocks({
          appid: appStore.currentId,
          ...r,
        });
      } finally {
        loading.value = false;
      }
    };
    const handleClick = () => {
      Object.assign(formData, {
        type: "ip",
        value: "",
        reason: "",
      });
      visible.value = true;
    };
    const handleCreateBlock = async () => {
      await createBlock({
        ...formData,
        appid: appStore.currentId,
      });
      Message.success("已封禁");
      visible.value = false;
      handleListBlocks();
    };
    const handleDeleteBlock = async (y) => {
      await deleteBlock(y.id);
      Message.success("已解封");
      handleListBlocks();
    };
    watch(() => appStore.currentId, handleListBlocks);
    onMounted(handleListBlocks);
    return {
      keepModalOpen,
      appStore,
      r,
      tableColumns,
      T,
      handleListBlocks,
      handleClick,
      handleCreateBlock,
      handleDeleteBlock,
      rows,
      loading,
      pagination,
      visible,
      formData,
      formatDateTime,
    };
  },
};
</script>
<template>
  <div class="page">
    <div class="page-card">
      <div class="toolbar">
        <Button type="primary" @click="handleClick">
          <template #icon><IconPlus></IconPlus></template>
          <template #default>新增封禁</template>
        </Button>
        <div class="grow"></div>
        <Select
          v-model="r.type"
          placeholder="类型"
          allow-clear=""
          :style="{
            width: '140px',
          }"
          @change="handleListBlocks"
        >
          <template #default>
            <Option value="ip"><template #default>IP</template></Option>
            <Option value="mi"><template #default>机器码</template></Option>
            <Option value="region"><template #default>区域</template></Option>
          </template>
        </Select>
        <Input
          v-model="r.keyword"
          placeholder="按值搜索"
          allow-clear=""
          :style="{
            width: '220px',
          }"
          @pressEnter="handleListBlocks"
          @clear="handleListBlocks"
        >
          <template #suffix>
            <IconSearch
              :style="{
                cursor: 'pointer',
              }"
              @click="handleListBlocks"
            ></IconSearch>
          </template>
        </Input>
      </div>
      <Table
        :data="rows"
        :columns="tableColumns"
        :loading="loading"
        :pagination="pagination"
        row-key="id"
      >
        <template #type="{ record: a }">
          <Tag
            :color="
              {
                ip: 'orange',
                mi: 'red',
                region: 'purple',
              }[a.type]
            "
          >
            <template #default>
              {{
                {
                  ip: "IP",
                  mi: "机器码",
                  region: "区域",
                }[a.type]
              }}
            </template>
          </Tag>
        </template>
        <template #value="{ record: a }">{{ a.value }}</template>
        <template #createTime="{ record: a }">{{
          formatDateTime(a.createTime)
        }}</template>
        <template #operations="{ record: a }">
          <Popconfirm content="确定解封？" @ok="(ce) => handleDeleteBlock(a)">
            <template #default>
              <Link status="danger"><template #default>解封</template></Link>
            </template>
          </Popconfirm>
        </template>
      </Table>
    </div>
    <Modal
      v-model:visible="visible"
      title="新增封禁"
      :on-before-ok="() => keepModalOpen(handleCreateBlock)"
      :width="400"
    >
      <template #default>
        <Form :model="formData" layout="vertical">
          <template #default>
            <FormItem label="类型">
              <template #default>
                <Select v-model="formData.type">
                  <template #default>
                    <Option value="ip"
                      ><template #default>IP 地址</template></Option
                    >
                    <Option value="mi"
                      ><template #default>机器码</template></Option
                    >
                    <Option value="region">
                      <template #default>IP 区域（省份/国家）</template>
                    </Option>
                  </template>
                </Select>
              </template>
            </FormItem>
            <FormItem label="值">
              <template #default>
                <Input v-model="formData.value" :placeholder="T"></Input>
              </template>
            </FormItem>
            <FormItem label="原因">
              <template #default>
                <Textarea
                  v-model="formData.reason"
                  :auto-size="{
                    minRows: 2,
                    maxRows: 5,
                  }"
                ></Textarea>
              </template>
            </FormItem>
          </template>
        </Form>
      </template>
    </Modal>
  </div>
</template>
