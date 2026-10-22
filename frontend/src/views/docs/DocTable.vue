<script>
import { computed } from "vue";
import { Table, TableColumn } from "@arco-design/web-vue";
export default {
  components: { Table, TableColumn },
  props: {
    headers: { type: Array, required: true },
    rows: { type: Array, required: true },
    codeColumn: { type: Number, default: 0 },
  },
  setup(props) {
    const data = computed(() =>
      props.rows.map((cells, index) => ({ key: index, cells })),
    );
    return { data };
  },
};
</script>
<template>
  <Table
    class="doc-table"
    :data="data"
    :pagination="false"
    size="small"
    :bordered="{ cell: true }"
    :scroll="{ x: headers.length > 3 ? 720 : headers.length === 3 ? 600 : 460 }"
    :scrollbar="false"
  >
    <template #columns>
      <TableColumn
        v-for="(title, index) in headers"
        :key="index"
        :title="title"
        :width="
          index === headers.length - 1 ? undefined : index === 0 ? 180 : 110
        "
      >
        <template #cell="{ record }">
          <code v-if="index === codeColumn">{{ record.cells[index] }}</code>
          <span v-else>{{ record.cells[index] }}</span>
        </template>
      </TableColumn>
    </template>
  </Table>
</template>
<style scoped>
.doc-table {
  margin: 14px 0 20px;
}
.doc-table :deep(.arco-table-cell) {
  overflow-wrap: anywhere;
  line-height: 1.75;
}
.doc-table :deep(.arco-table-td) {
  vertical-align: top;
}
code {
  color: rgb(var(--primary-6));
  font-family: Consolas, monospace;
  font-size: 12px;
}
</style>
