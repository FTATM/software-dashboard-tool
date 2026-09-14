<template>
  <div class="flex flex-col gap-4 w-full">

    <div class="flex justify-between items-center w-full">

      <div v-if="!serverSide" class="flex items-center gap-1 w-full max-w-md">
        <div class="relative w-full flex-1">
          <div class="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none text-base-content/50">
            <Icon icon="lucide:search" class="w-4 h-4" />
          </div>
          <input type="text" v-model="globalFilterText"
            class="input input-bordered w-full pl-10 focus:border-primary transition-colors shadow-sm"
            :placeholder="$t('tableData.searchPlaceholder')" />
        </div>

        <div class="join shadow-sm h-[3rem] flex items-center bg-base-100">
          <div class="tooltip tooltip-bottom" :data-tip="$t('tableData.matchCase')">
            <button class="join-item btn btn-outline border-base-300 h-[3rem]"
              :class="{ 'bg-primary text-primary-content border-primary hover:bg-primary/90': matchCase, 'hover:bg-base-200': !matchCase }"
              @click="matchCase = !matchCase">
              <Icon icon="lucide:case-sensitive" class="w-5 h-5" />
            </button>
          </div>
          <div class="tooltip tooltip-bottom" :data-tip="$t('tableData.wholeWord')">
            <button class="join-item btn btn-outline border-base-300 h-[3rem]"
              :class="{ 'bg-primary text-primary-content border-primary hover:bg-primary/90': matchWholeWord, 'hover:bg-base-200': !matchWholeWord }"
              @click="matchWholeWord = !matchWholeWord">
              <Icon icon="lucide:whole-word" class="w-5 h-5" />
            </button>
          </div>
        </div>
      </div>

      <div v-else></div>

      <div>
        <slot name="toolbar-actions"></slot>
      </div>
    </div>

    <div class="bg-base-100 rounded-box shadow-sm border border-base-200 overflow-x-auto flex flex-col">
      <table class="table w-full">
        <thead class="bg-base-200 text-base-content text-sm">
          <tr v-for="headerGroup in table.getHeaderGroups()" :key="headerGroup.id">

            <th v-for="header in headerGroup.headers" :key="header.id" :class="[
              header.column.columnDef.meta?.headerClass,
              header.column.getCanSort() ? 'cursor-pointer select-none hover:bg-base-300/50 transition-colors' : ''
            ]" @click="header.column.getToggleSortingHandler()?.($event)">
              <div class="flex items-center gap-1"
                :class="header.column.columnDef.meta?.headerClass?.includes('text-right') ? 'justify-end' : ''">
                <FlexRender v-if="!header.isPlaceholder" :header="header" />
                <span v-if="header.column.getCanSort()"
                  class="w-4 h-4 flex items-center justify-center text-base-content/50">
                  <Icon v-if="header.column.getIsSorted() === 'asc'" icon="lucide:arrow-up"
                    class="w-3 h-3 text-primary" />
                  <Icon v-else-if="header.column.getIsSorted() === 'desc'" icon="lucide:arrow-down"
                    class="w-3 h-3 text-primary" />
                  <Icon v-else icon="lucide:arrow-up-down" class="w-3 h-3 opacity-30" />
                </span>
              </div>
            </th>

          </tr>
        </thead>

        <tbody>
          <tr v-if="isLoading">
            <td :colspan="table.getAllColumns().length" class="text-center py-12">
              <span class="loading loading-spinner loading-lg text-primary"></span>
            </td>
          </tr>

          <tr v-if="table.getRowModel().rows.length === 0 && !isLoading">
            <td :colspan="table.getAllColumns().length" class="text-center py-12 text-base-content/50">
              {{ $t('tableData.noRecords') }}
            </td>
          </tr>

          <tr v-for="row in table.getRowModel().rows" :key="row.id" class="hover:bg-base-200/30 transition-colors">
            <td v-for="cell in row.getAllCells()" :key="cell.id" :class="cell.column.columnDef.meta?.cellClass">
              <slot :name="'cell-' + cell.column.id" :row="row.original" :value="cell.getValue()">
                <FlexRender :cell="cell" />
              </slot>
            </td>
          </tr>
        </tbody>
      </table>

      <div v-if="!isLoading && table.getPageCount() > 0"
        class="flex flex-col lg:flex-row items-center justify-between p-4 border-t border-base-200 bg-base-100 gap-4 mt-auto">

        <div class="flex flex-col sm:flex-row items-center gap-4 w-full lg:w-auto">
          <div class="flex items-center gap-2">
            <span class="text-sm font-medium text-base-content/70">{{ $t('tableData.rowsPerPage') }}</span>
            <select class="select select-bordered select-sm w-20" :value="paginationState.pageSize"
              @change="table.setPageSize(Number($event.target.value))">
              <option :value="10">10</option>
              <option :value="20">20</option>
              <option :value="50">50</option>
              <option :value="100">100</option>
            </select>
          </div>

          <span class="text-sm font-medium text-base-content/70 hidden sm:block sm:border-l sm:border-base-300 sm:pl-4">
            {{ $t('tableData.showing') }} <span class="font-bold text-base-content">{{ startRecord }}</span>
            {{ $t('tableData.to') }} <span class="font-bold text-base-content">{{ endRecord }}</span>
            <template v-if="totalRecords !== null">
              {{ $t('tableData.of') }} <span class="font-bold text-base-content">{{ totalRecords }}</span>
            </template>
            {{ $t('tableData.records') }}
          </span>
        </div>

        <div class="flex items-center gap-4">
          <span class="text-sm font-medium text-base-content/70">
            {{ $t('tableData.page', { current: paginationState.pageIndex + 1, total: table.getPageCount() }) }}
          </span>
          <div class="join">
            <button class="join-item btn btn-sm" @click="table.setPageIndex(0)"
              :disabled="!table.getCanPreviousPage()">«</button>
            <button class="join-item btn btn-sm" @click="table.previousPage()"
              :disabled="!table.getCanPreviousPage()">‹</button>
            <button class="join-item btn btn-sm" @click="table.nextPage()"
              :disabled="!table.getCanNextPage()">›</button>
            <button v-if="!serverSide" class="join-item btn btn-sm"
              @click="table.setPageIndex(table.getPageCount() - 1)" :disabled="!table.getCanNextPage()">»</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { Icon } from '@iconify/vue';
import {
  useTable,
  tableFeatures,
  globalFilteringFeature,
  createFilteredRowModel,
  rowSortingFeature,
  createSortedRowModel,
  sortFn_alphanumeric,
  sortFn_text,
  rowPaginationFeature,
  createPaginatedRowModel,
  FlexRender
} from '@tanstack/vue-table';

const { t } = useI18n();

const props = defineProps({
  data: { type: Array, required: true },
  columns: { type: Array, required: true },
  initialSorting: { type: Array, default: () => [] },
  isLoading: { type: Boolean, default: false },

  serverSide: { type: Boolean, default: false },
  pageCount: { type: Number, default: -1 },
  totalRowCount: { type: Number, default: 0 },
  pagination: {
    type: Object,
    default: () => ({ pageIndex: 0, pageSize: 10 })
  }
});

const emit = defineEmits(['update:pagination', 'update:sorting']);

const globalFilterText = ref('');
const matchCase = ref(false);
const matchWholeWord = ref(false);
const sorting = ref(props.initialSorting);

const combinedFilterState = computed(() => ({
  text: globalFilterText.value,
  matchCase: matchCase.value,
  matchWholeWord: matchWholeWord.value
}));

const paginationState = ref({ ...props.pagination });

watch(() => props.pagination, (newVal) => {
  if (newVal) {
    const isDifferent = newVal.pageIndex !== paginationState.value.pageIndex || newVal.pageSize !== paginationState.value.pageSize;
    if (isDifferent) {
      paginationState.value = { ...newVal };
    }
  }
}, { deep: true, immediate: true });

watch([globalFilterText, matchCase, matchWholeWord], () => {
  if (paginationState.value.pageIndex > 0) {
    paginationState.value.pageIndex = 0;
    emit('update:pagination', paginationState.value);
  }
});

const customFilterFn = (row, columnId, filterValue) => {
  if (!filterValue || !filterValue.text) return true;

  const cellValue = row.getValue(columnId);
  if (cellValue == null) return false;

  let text = String(cellValue);
  let search = String(filterValue.text);

  if (!filterValue.matchCase) {
    text = text.toLowerCase();
    search = search.toLowerCase();
  }

  if (filterValue.matchWholeWord) {
    try {
      const escapedSearch = search.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
      const regex = new RegExp(`\\b${escapedSearch}\\b`, filterValue.matchCase ? '' : 'i');
      return regex.test(text);
    } catch (e) {
      return false;
    }
  }

  return text.includes(search);
};

const features = tableFeatures({
  globalFilteringFeature,
  rowSortingFeature,
  rowPaginationFeature,
  filteredRowModel: createFilteredRowModel(),
  sortedRowModel: createSortedRowModel(),
  paginatedRowModel: createPaginatedRowModel(),
  filterFns: { customSearch: customFilterFn },
  sortFns: { alphanumeric: sortFn_alphanumeric, text: sortFn_text }
});

const table = useTable({
  features,
  get data() { return props.data; },
  get columns() { return props.columns; },

  get manualPagination() { return props.serverSide; },
  get pageCount() { return props.serverSide ? props.pageCount : undefined; },
  get manualSorting() { return props.serverSide; },

  state: {
    get globalFilter() { return combinedFilterState.value; },
    get sorting() { return sorting.value; },
    get pagination() { return paginationState.value; }
  },

  onSortingChange: (updater) => {
    sorting.value = typeof updater === 'function' ? updater(sorting.value) : updater;
    emit('update:sorting', sorting.value);
  },

  onPaginationChange: (updaterOrValue) => {
    paginationState.value = typeof updaterOrValue === 'function'
      ? updaterOrValue(paginationState.value)
      : updaterOrValue;

    emit('update:pagination', paginationState.value);
  },

  globalFilterFn: 'customSearch',
  enableSortingRemoval: false
});

const startRecord = computed(() => {
  if (props.serverSide) {
    if (props.data.length === 0) return 0;
  } else {
    if (table.getFilteredRowModel().rows.length === 0) return 0;
  }
  return (paginationState.value.pageIndex * paginationState.value.pageSize) + 1;
});

const endRecord = computed(() => {
  if (props.serverSide) {
    return (paginationState.value.pageIndex * paginationState.value.pageSize) + props.data.length;
  } else {
    const totalFiltered = table.getFilteredRowModel().rows.length;
    return Math.min((paginationState.value.pageIndex + 1) * paginationState.value.pageSize, totalFiltered);
  }
});

const totalRecords = computed(() => {
  if (props.serverSide) {
    return props.totalRowCount;
  }
  return table.getFilteredRowModel().rows.length;
});
</script>