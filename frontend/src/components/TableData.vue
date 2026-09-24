<template>
  <div class="flex flex-col gap-3 w-full">

    <!-- Table Header Toolbar -->
    <div class="flex flex-col sm:flex-row justify-between items-stretch sm:items-center gap-3 w-full">

      <!-- Search & Filters -->
      <div v-if="!serverSide" class="flex items-center gap-2 w-full sm:max-w-md">
        <div class="relative w-full">
          <div class="absolute inset-y-0 left-0 flex items-center pl-3.5 pointer-events-none text-base-content/40">
            <Icon icon="lucide:search" class="w-4 h-4" />
          </div>
          <input 
            type="text" 
            v-model="globalFilterText"
            class="input input-sm h-10 w-full pl-10 pr-9 rounded-xl bg-base-100 border-base-300 focus:border-primary focus:ring-2 focus:ring-primary/10 transition-all text-sm shadow-xs"
            :placeholder="$t('tableData.searchPlaceholder')" />
          <button 
            v-if="globalFilterText" 
            @click="globalFilterText = ''"
            type="button" 
            class="absolute inset-y-0 right-0 flex items-center pr-3 text-base-content/40 hover:text-base-content transition-colors">
            <Icon icon="lucide:x" class="w-3.5 h-3.5" />
          </button>
        </div>

        <!-- Filter Modifiers Toggle Pill -->
        <div class="flex items-center bg-base-200/80 p-1 rounded-xl border border-base-300 shrink-0 h-10 shadow-xs">
          <div class="tooltip tooltip-bottom" :data-tip="$t('tableData.matchCase')">
            <button 
              type="button"
              class="btn btn-xs btn-ghost h-8 w-8 p-0 rounded-lg transition-all"
              :class="matchCase ? 'bg-base-100 text-primary shadow-xs font-bold' : 'text-base-content/50 hover:text-base-content'"
              @click="matchCase = !matchCase">
              <Icon icon="lucide:case-sensitive" class="w-4 h-4" />
            </button>
          </div>
          <div class="tooltip tooltip-bottom" :data-tip="$t('tableData.wholeWord')">
            <button 
              type="button"
              class="btn btn-xs btn-ghost h-8 w-8 p-0 rounded-lg transition-all"
              :class="matchWholeWord ? 'bg-base-100 text-primary shadow-xs font-bold' : 'text-base-content/50 hover:text-base-content'"
              @click="matchWholeWord = !matchWholeWord">
              <Icon icon="lucide:whole-word" class="w-4 h-4" />
            </button>
          </div>
        </div>
      </div>

      <div v-else></div>

      <!-- Action Buttons Slot -->
      <div class="flex items-center gap-2 justify-end">
        <slot name="toolbar-actions"></slot>
      </div>
    </div>

    <!-- Table Container Card -->
    <div class="bg-base-100 rounded-2xl border border-base-300 shadow-xs overflow-hidden flex flex-col transition-all">
      <div class="overflow-x-auto w-full">
        <table class="table w-full border-collapse">
          <!-- Table Header -->
          <thead class="bg-base-200/60 border-b border-base-300 text-base-content/70">
            <tr v-for="headerGroup in table.getHeaderGroups()" :key="headerGroup.id">
              <th 
                v-for="header in headerGroup.headers" 
                :key="header.id" 
                :class="[
                  header.column.columnDef.meta?.headerClass,
                  header.column.getCanSort() ? 'cursor-pointer select-none hover:bg-base-300/50 transition-colors' : '',
                  'py-3.5 px-4 text-[11px] font-bold uppercase tracking-wider'
                ]" 
                @click="header.column.getToggleSortingHandler()?.($event)">
                <div 
                  class="flex items-center gap-1.5"
                  :class="header.column.columnDef.meta?.headerClass?.includes('text-right') ? 'justify-end' : ''">
                  <FlexRender v-if="!header.isPlaceholder" :header="header" />
                  <span v-if="header.column.getCanSort()" class="w-4 h-4 flex items-center justify-center shrink-0">
                    <Icon v-if="header.column.getIsSorted() === 'asc'" icon="lucide:chevron-up" class="w-3.5 h-3.5 text-primary" />
                    <Icon v-else-if="header.column.getIsSorted() === 'desc'" icon="lucide:chevron-down" class="w-3.5 h-3.5 text-primary" />
                    <Icon v-else icon="lucide:chevrons-up-down" class="w-3.5 h-3.5 opacity-30" />
                  </span>
                </div>
              </th>
            </tr>
          </thead>

          <!-- Table Body -->
          <tbody class="divide-y divide-base-200 text-sm">
            <!-- Loading State -->
            <tr v-if="isLoading">
              <td :colspan="table.getAllColumns().length" class="text-center py-16">
                <div class="flex flex-col items-center justify-center gap-2">
                  <span class="loading loading-spinner loading-md text-primary"></span>
                  <span class="text-xs font-medium text-base-content/50 tracking-wide">Loading data...</span>
                </div>
              </td>
            </tr>

            <!-- Empty State -->
            <tr v-else-if="table.getRowModel().rows.length === 0">
              <td :colspan="table.getAllColumns().length" class="text-center py-16">
                <div class="flex flex-col items-center justify-center gap-2 text-base-content/40">
                  <div class="p-3 bg-base-200 rounded-full">
                    <Icon icon="lucide:inbox" class="w-6 h-6" />
                  </div>
                  <span class="text-sm font-medium text-base-content/60">{{ $t('tableData.noRecords') }}</span>
                </div>
              </td>
            </tr>

            <!-- Data Rows -->
            <tr 
              v-else
              v-for="row in table.getRowModel().rows" 
              :key="row.id" 
              class="hover:bg-base-200/40 transition-colors">
              <td 
                v-for="cell in row.getAllCells()" 
                :key="cell.id" 
                :class="[cell.column.columnDef.meta?.cellClass, 'py-3.5 px-4 text-base-content/85 align-middle']">
                <slot :name="'cell-' + cell.column.id" :row="row.original" :value="cell.getValue()">
                  <FlexRender :cell="cell" />
                </slot>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Footer Pagination -->
      <div 
        v-if="!isLoading && table.getPageCount() > 0"
        class="flex flex-col sm:flex-row items-center justify-between px-5 py-3.5 border-t border-base-200 bg-base-100 gap-4 mt-auto">

        <!-- Left: Page Size & Records Count -->
        <div class="flex items-center gap-4 w-full sm:w-auto justify-between sm:justify-start text-xs text-base-content/70">
          <div class="flex items-center gap-2">
            <span class="font-medium">{{ $t('tableData.rowsPerPage') }}</span>
            <select 
              class="select select-bordered select-xs h-8 px-2 rounded-lg bg-base-100 font-medium" 
              :value="paginationState.pageSize"
              @change="table.setPageSize(Number($event.target.value))">
              <option :value="10">10</option>
              <option :value="20">20</option>
              <option :value="50">50</option>
              <option :value="100">100</option>
            </select>
          </div>

          <div class="h-3.5 w-px bg-base-300 hidden sm:block"></div>

          <span class="font-medium text-base-content/60">
            {{ $t('tableData.showing') }} <span class="font-semibold text-base-content">{{ startRecord }}</span>–<span class="font-semibold text-base-content">{{ endRecord }}</span>
            <template v-if="totalRecords !== null">
              {{ $t('tableData.of') }} <span class="font-semibold text-base-content">{{ totalRecords }}</span>
            </template>
          </span>
        </div>

        <!-- Right: Pagination Buttons -->
        <div class="flex items-center gap-3">
          <span class="text-xs font-medium text-base-content/60">
            {{ $t('tableData.page', { current: paginationState.pageIndex + 1, total: table.getPageCount() }) }}
          </span>

          <div class="flex items-center gap-1">
            <button 
              class="btn btn-ghost btn-xs btn-square rounded-lg border border-base-300 hover:bg-base-200 disabled:opacity-30 disabled:border-transparent" 
              @click="table.setPageIndex(0)"
              :disabled="!table.getCanPreviousPage()"
              title="First Page">
              <Icon icon="lucide:chevrons-left" class="w-3.5 h-3.5" />
            </button>
            <button 
              class="btn btn-ghost btn-xs btn-square rounded-lg border border-base-300 hover:bg-base-200 disabled:opacity-30 disabled:border-transparent" 
              @click="table.previousPage()"
              :disabled="!table.getCanPreviousPage()"
              title="Previous Page">
              <Icon icon="lucide:chevron-left" class="w-3.5 h-3.5" />
            </button>
            <button 
              class="btn btn-ghost btn-xs btn-square rounded-lg border border-base-300 hover:bg-base-200 disabled:opacity-30 disabled:border-transparent" 
              @click="table.nextPage()"
              :disabled="!table.getCanNextPage()"
              title="Next Page">
              <Icon icon="lucide:chevron-right" class="w-3.5 h-3.5" />
            </button>
            <button 
              v-if="!serverSide" 
              class="btn btn-ghost btn-xs btn-square rounded-lg border border-base-300 hover:bg-base-200 disabled:opacity-30 disabled:border-transparent"
              @click="table.setPageIndex(table.getPageCount() - 1)" 
              :disabled="!table.getCanNextPage()"
              title="Last Page">
              <Icon icon="lucide:chevrons-right" class="w-3.5 h-3.5" />
            </button>
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