<template>
  <div v-if="hasPermission(mainMenuName, 'Display')" class="w-full h-full overflow-y-auto p-4 sm:p-6 space-y-5">

    <!-- Anchored Page Header Card with Background & Tabs -->
    <div class="bg-base-100 border border-base-300 rounded-2xl p-5 shadow-xs flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
      <div class="flex items-start sm:items-center gap-3.5">
        <div class="p-3 bg-primary/10 text-primary rounded-xl flex items-center justify-center shrink-0">
          <Icon icon="lucide:file-text" class="w-6 h-6" />
        </div>
        <div>
          <!-- Breadcrumbs -->
          <div class="flex items-center gap-1.5 text-xs font-semibold text-base-content/50 uppercase tracking-wider mb-0.5">
            <span class="text-primary">{{ $t('logReport.title') }}</span>
          </div>

          <!-- Title & Subtitle -->
          <div class="flex items-center gap-2.5">
            <h1 class="m-0 text-xl sm:text-2xl font-black text-base-content tracking-tight">
              {{ $t('logReport.title') }}
            </h1>
          </div>
          <p class="mt-0.5 mb-0 text-base-content/60 text-xs font-medium">
            {{ $t('logReport.subtitle') }}
          </p>
        </div>
      </div>

      <!-- Segmented Log Mode Tabs Switcher -->
      <div class="inline-flex p-1 bg-base-200/80 rounded-xl border border-base-300 shrink-0 self-stretch sm:self-auto">
        <button 
          type="button"
          class="flex-1 sm:flex-initial flex items-center justify-center px-3.5 py-1.5 rounded-lg text-xs sm:text-sm font-bold transition-all duration-200"
          :class="activeTab === 'system' ? 'bg-base-100 text-primary shadow-xs' : 'text-base-content/60 hover:text-base-content'"
          @click="switchTab('system')">
          <Icon icon="lucide:server" class="w-4 h-4 mr-1.5" />
          {{ $t('logReport.tabSystem') }}
        </button>

        <button 
          type="button"
          class="flex-1 sm:flex-initial flex items-center justify-center px-3.5 py-1.5 rounded-lg text-xs sm:text-sm font-bold transition-all duration-200"
          :class="activeTab === 'device' ? 'bg-base-100 text-primary shadow-xs' : 'text-base-content/60 hover:text-base-content'"
          @click="switchTab('device')">
          <Icon icon="lucide:cpu" class="w-4 h-4 mr-1.5" />
          {{ $t('logReport.tabDevice') }}
        </button>
      </div>
    </div>

    <!-- Reusable KPI Summary Status Cards -->
    <StatCardGroup :items="statCardsData" />

    <!-- Server-Side Filter Toolbar Card -->
    <div class="bg-base-100 border border-base-300 rounded-2xl p-4 sm:p-5 shadow-xs">
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-12 gap-3 items-end">

        <!-- Start Date -->
        <label class="form-control w-full lg:col-span-3">
          <div class="label pb-1">
            <span class="label-text font-semibold text-xs uppercase tracking-wider text-base-content/70">{{ $t('logReport.selectStartDate') }}</span>
          </div>
          <VueDatePicker 
            v-model="filters.from" 
            :is-24="true" 
            auto-apply 
            :preset-dates="presetDates"
            :locale="dateFnsLocale" 
            :format-locale="dateFnsLocale" 
            :dark="themeStore.isDarkTheme"
            :formats="{ input: 'dd/MM/yyyy HH:mm' }" 
            :action-row="{ selectBtnLabel: $t('common.select'), cancelBtnLabel:$t('common.cancel') }" 
            :placeholder="$t('logReport.selectStartDate')" 
            teleport-center>
            <template #input-icon>
              <Icon icon="lucide:calendar-clock" class="w-4 h-4 ml-3 text-base-content/50" />
            </template>
          </VueDatePicker>
        </label>

        <!-- End Date -->
        <label class="form-control w-full lg:col-span-3">
          <div class="label pb-1">
            <span class="label-text font-semibold text-xs uppercase tracking-wider text-base-content/70">{{ $t('logReport.selectEndDate') }}</span>
          </div>
          <VueDatePicker 
            v-model="filters.to" 
            :is-24="true" 
            auto-apply 
            :preset-dates="presetDates"
            :locale="dateFnsLocale" 
            :format-locale="dateFnsLocale" 
            :dark="themeStore.isDarkTheme"
            :formats="{ input: 'dd/MM/yyyy HH:mm' }" 
            :action-row="{ selectBtnLabel: $t('common.select'), cancelBtnLabel:$t('common.cancel') }" 
            :placeholder="$t('logReport.selectEndDate')" 
            teleport-center>
            <template #input-icon>
              <Icon icon="lucide:calendar-clock" class="w-4 h-4 ml-3 text-base-content/50" />
            </template>
          </VueDatePicker>
        </label>

        <!-- Menu Types Dropdown (System Tab Only) -->
        <label v-if="activeTab === 'system'" class="form-control w-full lg:col-span-2">
          <div class="label pb-1">
            <span class="label-text font-semibold text-xs uppercase tracking-wider text-base-content/70">{{ $t('logReport.menuTypes') }}</span>
          </div>
          <SearchableDropdown 
            v-model="filters.menuTypes" 
            :options="formattedMenuTypes" 
            labelKey="name" 
            valueKey="id"
            :placeholder="$t('logReport.selectMenus')" 
            multiple />
        </label>

        <!-- Keyword Search -->
        <label class="form-control w-full" :class="activeTab === 'system' ? 'lg:col-span-3' : 'lg:col-span-4'">
          <div class="label pb-1">
            <span class="label-text font-semibold text-xs uppercase tracking-wider text-base-content/70">{{ $t('logReport.keywordSearch') }}</span>
          </div>
          <div class="relative">
            <div class="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none text-base-content/40">
              <Icon icon="lucide:search" class="w-4 h-4" />
            </div>
            <input 
              type="text" 
              v-model="filters.keyword" 
              :placeholder="$t('logReport.searchPlaceholder')"
              @keyup.enter="fetchLogs(1)"
              class="input input-sm h-10 input-bordered w-full pl-9 rounded-xl text-xs" />
          </div>
        </label>

        <!-- Apply Button -->
        <div class="w-full" :class="activeTab === 'system' ? 'lg:col-span-1' : 'lg:col-span-2'">
          <button 
            @click="fetchLogs(1)" 
            class="btn btn-sm h-10 btn-primary w-full rounded-xl font-semibold shadow-xs hover:shadow-md transition-all gap-1 text-white">
            <Icon icon="lucide:filter" class="w-4 h-4" />
            <span>{{ $t('logReport.applyFilters') }}</span>
          </button>
        </div>

      </div>
    </div>

    <!-- Data Table Card -->
    <div class="bg-base-100 border border-base-300 rounded-2xl p-4 sm:p-5 shadow-xs">
      <TableData 
        :data="logTableData" 
        :columns="currentColumns" 
        :is-loading="isLoading" 
        :server-side="true"
        :total-row-count="totalServerRecords" 
        :page-count="computedPageCount" 
        :pagination="tablePagination"
        @update:pagination="handlePaginationChange" 
        @update:sorting="handleSortingChange">
        
        <!-- Toolbar Actions: Colored Export Button -->
        <template #toolbar-actions>
          <div class="dropdown dropdown-end">
            <div 
              tabindex="0" 
              role="button" 
              class="btn btn-sm btn-outline btn-secondary rounded-xl shadow-xs hover:shadow-sm transition-all font-medium">
              <span v-if="isExporting" class="loading loading-spinner loading-xs"></span>
              <Icon v-else icon="lucide:download" class="w-4 h-4 mr-1" />
              {{ $t('common.export') }}
            </div>
            <ul tabindex="0" class="dropdown-content z-50 menu p-1.5 shadow-xl bg-base-100 rounded-xl border border-base-300 w-36 mt-1 text-xs font-medium">
              <li><a @click="exportData('json')" class="rounded-lg py-1.5">JSON</a></li>
              <li><a @click="exportData('csv')" class="rounded-lg py-1.5">CSV</a></li>
              <li><a @click="exportData('excel')" class="rounded-lg py-1.5">Excel</a></li>
            </ul>
          </div>
        </template>

        <!-- Action Badge -->
        <template #cell-action="{ value }">
          <span 
            class="inline-flex items-center px-2 py-0.5 rounded-md text-[10px] font-black uppercase tracking-wider"
            :class="{
              'bg-success/10 text-success border border-success/20': value === 'CREATE',
              'bg-info/10 text-info border border-info/20': value === 'UPDATE',
              'bg-error/10 text-error border border-error/20': value === 'DELETE',
              'bg-secondary/10 text-secondary border border-secondary/20': value === 'QUERY'
            }">
            {{ value }}
          </span>
        </template>

        <!-- Created At Timestamp -->
        <template #cell-createdAt="{ value }">
          <span class="font-mono text-xs text-base-content/70">{{ formatTime(value) }}</span>
        </template>

        <!-- Received At Timestamp -->
        <template #cell-receivedAt="{ value }">
          <span class="font-mono text-xs text-base-content/70">{{ formatTime(value) }}</span>
        </template>

        <!-- Entity ID Cell -->
        <template #cell-entityId="{ value }">
          <span v-if="value" class="font-mono text-xs font-bold text-base-content/50">#{{ value }}</span>
          <span v-else class="text-base-content/30 text-xs font-mono">-</span>
        </template>

        <!-- Device ID Cell -->
        <template #cell-deviceId="{ value }">
          <span class="font-mono text-xs font-bold text-base-content/50">#{{ value }}</span>
        </template>

        <!-- Username Cell -->
        <template #cell-username="{ value }">
          <span v-if="value" class="font-mono text-xs font-medium text-base-content/80">@{{ value }}</span>
          <span v-else class="text-base-content/30 text-xs font-mono">-</span>
        </template>

        <!-- Value Data Cell -->
        <template #cell-valueData="{ value }">
          <span class="font-mono font-semibold text-xs text-base-content">{{ value ?? '-' }}</span>
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
import StatCardGroup from '@/components/StatCardGroup.vue';
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

// Dynamic bilingual summary stats for StatCardGroup
const statCardsData = computed(() => {
  const isSystem = activeTab.value === 'system';
  const totalLogs = totalServerRecords.value;
  const currentCount = logTableData.value.length;

  if (isSystem) {
    return [
      {
        label: t('logReport.stats.totalLogs'),
        value: totalLogs.toLocaleString(),
        icon: 'lucide:file-text',
        color: 'primary'
      },
      {
        label: t('logReport.stats.pageLogs'),
        value: currentCount,
        icon: 'lucide:list',
        color: 'info',
        valueClass: 'text-info'
      },
      {
        label: t('logReport.stats.systemMenus'),
        value: menuTypesList.value.length,
        icon: 'lucide:layers',
        color: 'accent'
      },
      {
        label: t('logReport.stats.mode'),
        value: t('logReport.stats.systemScope'),
        icon: 'lucide:shield-alert',
        color: 'success',
        valueClass: 'text-success text-base sm:text-lg'
      }
    ];
  } else {
    const uniqueDevices = new Set(logTableData.value.map(d => d.deviceId)).size;

    return [
      {
        label: t('logReport.stats.totalLogs'),
        value: totalLogs.toLocaleString(),
        icon: 'lucide:file-text',
        color: 'primary'
      },
      {
        label: t('logReport.stats.pageLogs'),
        value: currentCount,
        icon: 'lucide:list',
        color: 'info',
        valueClass: 'text-info'
      },
      {
        label: t('logReport.stats.uniqueDevices'),
        value: uniqueDevices,
        icon: 'lucide:cpu',
        color: 'accent'
      },
      {
        label: t('logReport.stats.mode'),
        value: t('logReport.stats.deviceScope'),
        icon: 'lucide:radio',
        color: 'success',
        valueClass: 'text-success text-base sm:text-lg'
      }
    ];
  }
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