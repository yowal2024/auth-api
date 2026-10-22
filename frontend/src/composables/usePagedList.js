import { onBeforeUnmount } from "vue";
import { reportError } from "@/utils/errors.js";

// Only the latest query owns the rows and loading state.
export function usePagedList(
  fetchPage,
  filters,
  rows,
  total,
  loading,
  selected,
) {
  let sequence = 0;
  let previousFilter;
  onBeforeUnmount(() => sequence++);
  return async () => {
    const { page, size, ...query } = filters;
    const filterKey = JSON.stringify(query);
    if (previousFilter !== undefined && previousFilter !== filterKey)
      filters.page = 1;
    previousFilter = filterKey;
    const current = ++sequence;
    if (selected) selected.value = [];
    loading.value = true;
    try {
      let result = await fetchPage({ ...filters });
      if (current !== sequence) return;
      const lastPage = Math.max(1, Math.ceil(result.total / filters.size));
      if (filters.page > lastPage) {
        filters.page = lastPage;
        result = await fetchPage({ ...filters });
        if (current !== sequence) return;
      }
      rows.value = result.list;
      total.value = result.total;
    } catch (error) {
      if (current === sequence) reportError(error);
    } finally {
      if (current === sequence) loading.value = false;
    }
  };
}
