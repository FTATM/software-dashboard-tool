<template>
  <div v-if="hasPermission(mainMenuName, 'Display')" class="w-full mx-auto p-4 flex flex-col h-full gap-4">

    <!-- Header & Tabs -->
    <div
      class="bg-base-100 shadow-sm rounded-box border border-base-200 p-6 flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
      <div class="flex items-center gap-4">
        <div class="p-3 bg-primary/10 text-primary rounded-xl flex items-center justify-center">
          <Icon icon="lucide:file-text" class="w-7 h-7" />
        </div>
        <div>
          <h2 class="m-0 text-2xl font-extrabold text-base-content tracking-tight">{{ $t('logReport.title') }}</h2>
          <p class="mt-1 mb-0 text-base-content/60 text-sm font-medium">{{ $t('logReport.subtitle') }}</p>
        </div>
      </div>

      <div class="inline-flex p-1 bg-base-200/60 rounded-xl border border-base-content/10">
        <button type="button"
          class="flex items-center px-3 py-1.5 rounded-lg text-sm font-semibold transition-all duration-200"
          :class="activeTab === 'system' ? 'bg-base-100 text-primary shadow-sm' : 'text-base-content/70 hover:text-base-content'"
          @click="switchTab('system')">
          <Icon icon="lucide:server" class="w-4 h-4 mr-2" />
          {{ $t('logReport.tabSystem') }}
        </button>

        <button type="button"
          class="flex items-center px-3 py-1.5 rounded-lg text-sm font-semibold transition-all duration-200"
          :class="activeTab === 'device' ? 'bg-base-100 text-primary shadow-sm' : 'text-base-content/70 hover:text-base-content'"
          @click="switchTab('device')">
          <Icon icon="lucide:cpu" class="w-4 h-4 mr-2" />
          {{ $t('logReport.tabDevice') }}
        </button>
      </div>
    </div>

    <!-- Server-Side Filter Menu -->
    <div class="bg-base-100 shadow-sm rounded-box border border-base-200 p-4">
      <div class="grid grid-cols-1 md:grid-cols-5 gap-4 items-end">

        <label class="form-control w-full">
          <div class="label pb-1"><span class="label-text font-semibold">{{ $t('logReport.selectStartDate') }}</span>
          </div>
          <VueDatePicker v-model="filters.from" :is-24="true" auto-apply :preset-dates="presetDates"
            :locale="dateFnsLocale" :format-locale="dateFnsLocale" :dark="themeStore.isDarkTheme"
            :formats="{ input: 'dd/MM/yyyy HH:mm' }" :action-row="{
              selectBtnLabel: $t('common.select'),
              cancelBtnLabel: $t('common.cancel')
            }" :placeholder="$t('logReport.selectStartDate')" teleport-center>
            <template #input-icon>
              <Icon icon="lucide:calendar-clock" class="w-5 h-5 ml-3 text-base-content/50" />
            </template>
          </VueDatePicker>
        </label>

        <label class="form-control w-full">
          <div class="label pb-1"><span class="label-text font-semibold">{{ $t('logReport.selectEndDate') }}</span>
          </div>
          <VueDatePicker v-model="filters.to" :is-24="true" auto-apply :preset-dates="presetDates"
            :locale="dateFnsLocale" :format-locale="dateFnsLocale" :dark="themeStore.isDarkTheme"
            :formats="{ input: 'dd/MM/yyyy HH:mm' }" :action-row="{
              selectBtnLabel: $t('common.select'),
              cancelBtnLabel: $t('common.cancel')
            }" :placeholder="$t('logReport.selectEndDate')" teleport-center>
            <template #input-icon>
              <Icon icon="lucide:calendar-clock" class="w-5 h-5 ml-3 text-base-content/50" />
            </template>
          </VueDatePicker>
        </label>

        <!-- Menu Types Dropdown (replaces Entity Types) -->
        <label v-if="activeTab === 'system'" class="form-control w-full">
          <div class="label pb-1"><span class="label-text font-semibold">{{ $t('logReport.menuTypes') }}</span></div>
          <SearchableDropdown v-model="filters.menuTypes" :options="formattedMenuTypes" labelKey="name" valueKey="id"
            :placeholder="$t('logReport.selectMenus')" multiple />
        </label>

        <label class="form-control w-full" :class="{ 'md:col-span-2': activeTab !== 'system' }">
          <div class="label pb-1"><span class="label-text font-semibold">{{ $t('logReport.keywordSearch') }}</span>
          </div>
          <input type="text" v-model="filters.keyword" :placeholder="$t('logReport.searchPlaceholder')"
            class="input input-bordered input-sm w-full h-[3rem]" />
        </label>

        <button @click="fetchLogs(1)" class="btn btn-primary h-[3rem] w-full text-white">
          <Icon icon="lucide:filter" class="w-4 h-4 mr-1" />
          {{ $t('logReport.applyFilters') }}
        </button>
      </div>
    </div>

    <!-- Data Table -->
    <div class="flex-1 flex flex-col min-h-0">
      <TableData :data="logTableData" :columns="currentColumns" :is-loading="isLoading" :server-side="true"
        :total-row-count="totalServerRecords" :page-count="computedPageCount" :pagination="tablePagination"
        @update:pagination="handlePaginationChange" @update:sorting="handleSortingChange">
        <template #toolbar-actions>
          <div class="dropdown dropdown-end">
            <div tabindex="0" role="button" class="btn btn-outline btn-secondary shadow-sm transition-all">
              <span v-if="isExporting" class="loading loading-spinner loading-sm"></span>
              <Icon v-else icon="lucide:download" class="w-5 h-5 mr-1" />
              {{ $t('common.export') }}
            </div>
            <ul tabindex="0" class="dropdown-content z-[1] menu p-2 shadow bg-base-100 rounded-box w-32">
              <li><a @click="exportData('json')">JSON</a></li>
              <li><a @click="exportData('csv')">CSV</a></li>
              <li><a @click="exportData('excel')">Excel</a></li>
            </ul>
          </div>
        </template>

        <template #cell-action="{ value }">
          <div class="badge badge-sm font-bold uppercase" :class="{
            'badge-success text-white': value === 'CREATE',
            'badge-info text-white': value === 'UPDATE',
            'badge-error text-white': value === 'DELETE',
            'badge-secondary text-white': value === 'QUERY'
          }">
            {{ value }}
          </div>
        </template>

        <template #cell-createdAt="{ value }">
          <span class="text-sm font-mono text-base-content/70">{{ formatTime(value) }}</span>
        </template>

        <template #cell-receivedAt="{ value }">
          <span class="text-sm font-mono text-base-content/70">{{ formatTime(value) }}</span>
        </template>
      </TableData>
    </div>
  </div>
  <NoAccess v-else />
</template>

<script setup>
import { ref, onMounted, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { Icon } from '@iconify/vue';
import { toast } from 'vue3-toastify';
import { useFetch } from '@/composables/useFetch';
import { useDownload } from '@/composables/useDownload';
import { usePermissionStore } from '@/stores/usePermissionStore';
import TableData from '@/components/TableData.vue';
import NoAccess from '@/components/NoAccess.vue';
import SearchableDropdown from '@/components/SearchableDropdown.vue';
import { useErrorHandler } from '@/composables/useErrorHandler';
const { handleError } = useErrorHandler();
import { useFormatter } from '@/composables/useFormatter';
const { formatTime } = useFormatter();

import { VueDatePicker } from '@vuepic/vue-datepicker';
import thLocale from 'date-fns/locale/th';
import enLocale from 'date-fns/locale/en-US';
import { useThemeStore } from '@/stores/useThemeStore';
const themeStore = useThemeStore();
const { t, locale } = useI18n();
const mainMenuName = 'Log Report';

const { data: fetchResult, isLoading, error: fetchError, execute: executeFetch } = useFetch();
const { isDownloading: isExporting, error: exportError, executeDownload } = useDownload();
const permissionStore = usePermissionStore();
const { hasPermission } = permissionStore;

const activeTab = ref('system');
const logTableData = ref([]);
const menuTypesList = ref([]);
const totalServerRecords = ref(0);
const tableSorting = ref([]);
const presetDates = ref([
  { label: t('common.today'), value: new Date() }
]);

const filters = ref({
  from: null,
  to: null,
  keyword: '',
  menuTypes: [],
  page: 1,
  limit: 50
});

const tablePagination = computed(() => ({
  pageIndex: filters.value.page - 1,
  pageSize: filters.value.limit
}));

const computedPageCount = computed(() => {
  if (filters.value.limit === 0) return 0;
  return Math.ceil(totalServerRecords.value / filters.value.limit);
});

const resolveLocale = (loc) => {
  if (!loc) return undefined;
  if (loc.localize) return loc;
  if (loc.default?.localize) return loc.default;
  if (loc.th?.localize) return loc.th;
  if (loc.enUS?.localize) return loc.enUS;
  return loc.default || loc;
};

const dateFnsLocale = computed(() => {
  const isThai = locale.value?.startsWith('th');
  return resolveLocale(isThai ? thLocale : enLocale);
});

const handlePaginationChange = async (newPagination) => {
  const limitChanged = filters.value.limit !== newPagination.pageSize;
  const pageChanged = (filters.value.page - 1) !== newPagination.pageIndex;

  if (limitChanged || pageChanged) {
    filters.value.limit = newPagination.pageSize;
    await fetchLogs(newPagination.pageIndex + 1);
  }
};

const formattedMenuTypes = computed(() => {
  return menuTypesList.value.map(type => ({
    id: type,
    name: type.charAt(0).toUpperCase() + type.slice(1)
  }));
});

const systemColumns = computed(() => [
  { header: t('logReport.table.timestamp'), accessorKey: 'createdAt', meta: { headerClass: 'w-44' } },
  { header: t('logReport.table.action'), accessorKey: 'action', meta: { headerClass: 'w-24' } },
  { header: t('logReport.table.menu'), accessorKey: 'menuType', meta: { headerClass: 'w-32 font-semibold capitalize' } },
  { header: t('logReport.table.entity'), accessorKey: 'entityType', meta: { headerClass: 'w-32 capitalize' } },
  { header: t('logReport.table.entityId'), accessorKey: 'entityId', meta: { headerClass: 'w-24' } },
  { header: t('common.user'), accessorKey: 'username', meta: { headerClass: 'w-32' } }
]);

const deviceColumns = computed(() => [
  { header: t('logReport.table.timestamp'), accessorKey: 'receivedAt', meta: { headerClass: 'w-48' } },
  { header: t('common.deviceId'), accessorKey: 'deviceId', meta: { headerClass: 'w-32 font-bold' } },
  { header: t('common.deviceName'), accessorKey: 'deviceName', meta: { headerClass: 'w-32' } },
  { header: t('logReport.table.value'), accessorKey: 'valueData', meta: { headerClass: 'w-24' } }
]);

const currentColumns = computed(() => activeTab.value === 'system' ? systemColumns.value : deviceColumns.value);

const switchTab = async (tabName) => {
  if (activeTab.value === tabName) return;
  activeTab.value = tabName;
  filters.value.page = 1;
  logTableData.value = [];
  await fetchLogs(1);
};

const fetchLogs = async (targetPage) => {
  filters.value.page = targetPage;

  const params = new URLSearchParams({
    tab: activeTab.value,
    page: filters.value.page,
    limit: filters.value.limit,
    keyword: filters.value.keyword
  });

  if (filters.value.from) params.append('from', new Date(filters.value.from).toISOString());
  if (filters.value.to) params.append('to', new Date(filters.value.to).toISOString());

  if (filters.value.menuTypes.length > 0) {
    params.append('menuTypes', filters.value.menuTypes.join(','));
  }

  if (tableSorting.value.length > 0) {
    const sort = tableSorting.value[0];
    params.append('sortBy', sort.id);
    params.append('sortDesc', sort.desc);
  }

  await executeFetch(`/logreport/searchlogs?${params.toString()}`);

  if (!fetchError.value && fetchResult.value) {
    logTableData.value = fetchResult.value.data.logs || [];
    totalServerRecords.value = fetchResult.value.data.totalCount || 0;
  } else {
    toast.error(fetchError.value?.message || t('common.messages.loadError'));
  }
};

const handleSortingChange = async (newSorting) => {
  tableSorting.value = newSorting;
  await fetchLogs(1);
};

const exportData = async (format) => {
  const extension = format === 'excel' ? 'xlsx' : format;

  const now = new Date();
  const timestamp = `${now.getFullYear()}${String(now.getMonth() + 1).padStart(2, '0')}${String(now.getDate()).padStart(2, '0')}_${String(now.getHours()).padStart(2, '0')}${String(now.getMinutes()).padStart(2, '0')}`;
  const filename = `${activeTab.value}_logs_${timestamp}.${extension}`;

  const params = new URLSearchParams({
    tab: activeTab.value,
    keyword: filters.value.keyword,
    export: 'true'
  });

  if (filters.value.from) params.append('from', new Date(filters.value.from).toISOString());
  if (filters.value.to) params.append('to', new Date(filters.value.to).toISOString());

  if (filters.value.menuTypes.length > 0) {
    params.append('menuTypes', filters.value.menuTypes.join(','));
  }

  const success = await executeDownload(`/logreport/export/logs?format=${format}&${params.toString()}`, filename);

  if (success) {
    toast.success(`${format.toUpperCase()} file downloaded successfully!`);
  } else {
    toast.error(handleError(exportError, 'device.messages.exportFailed', { format: format.toUpperCase() }));
  }
};

const loadMenuTypes = async () => {
  await executeFetch(`/logreport/getmenutypes`);
  if (!fetchError.value && fetchResult.value) {
    menuTypesList.value = fetchResult.value.data || [];
  }
};

onMounted(async () => {
  if (!hasPermission(mainMenuName, 'Display')) return;
  const now = new Date();
  const yesterday = new Date(now);
  yesterday.setHours(yesterday.getHours() - 24);
  filters.value.from = yesterday;
  const endOfToday = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 23, 59, 59);
  filters.value.to = endOfToday;

  await loadMenuTypes();
  fetchLogs(1);
});
</script>