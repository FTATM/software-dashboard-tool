<template>
  <div class="flex flex-col h-full w-full p-4 overflow-hidden" :style="backgroundStyle">

    <div class="backdrop-blur-md px-4 py-3 shadow-sm z-10 flex justify-between items-center rounded-t-lg">
      <h3 class="m-0 text-base font-extrabold tracking-wide" :style="{ color: widgetData.widgetStyle?.textHex || '#334155' }">
        {{ widgetData?.widgetLabel }}
      </h3>
      <span class="badge badge-neutral badge-sm font-semibold" v-if="hasData && config.showRowCount">
        {{ displayData.length }} / {{ config.maxRows }} {{ $t('common.rows') }}
      </span>
    </div>

    <div class="flex-1 w-full relative overflow-auto border-x border-b border-base-200/50 bg-base-100/30 backdrop-blur-sm rounded-b-lg flex flex-col justify-center">

      <div v-if="!hasDevices" class="absolute inset-0 flex items-center justify-center text-sm italic text-center p-4" :style="{ color: widgetData.widgetStyle?.textHex || '#64748b' }">
        {{ $t('common.noDevicesConfig') }}
      </div>

      <div v-else-if="!hasData" class="absolute inset-0 flex flex-col items-center justify-center text-sm gap-3" :style="{ color: widgetData.widgetStyle?.textHex || '#64748b' }">
        <span class="loading loading-spinner loading-md text-primary"></span>
        {{ $t('common.waitingData') }}
      </div>

      <table v-else class="table w-full text-left relative" :class="{ 'table-zebra': config.isStriped, 'table-sm': config.isDense }">
        <thead class="sticky top-0 z-10 shadow-sm" :style="{ backgroundColor: config.headerColor, color: config.headerTextColor }">
          <tr>
            <th v-for="(col, index) in displayColumns" :key="index" class="font-bold text-sm tracking-wide whitespace-nowrap">
              {{ col }}
            </th>
          </tr>
        </thead>
        <tbody :style="{ color: widgetData.widgetStyle?.textHex || '#334155' }">
          <tr v-for="(row, index) in displayData" :key="index" class="hover:bg-base-200/30 transition-colors">
            <td v-for="(col, colIndex) in displayColumns" :key="colIndex" class="border-b border-base-200/50 whitespace-nowrap font-medium">
              {{ row[col] !== undefined ? row[col] : '-' }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
import { computed, shallowRef, onMounted, onUnmounted, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useLiveStreamStore } from '@/stores/useLiveStreamStore';
import { useFormatter } from '@/composables/useFormatter';

const { formatTime } = useFormatter();
const { t } = useI18n();

const props = defineProps({
  widgetData: { type: Object, default: () => ({}) }
});

const widgetInstanceId = `table-${props.widgetData?.id || Math.random().toString(36).substring(2, 9)}`;

const backgroundStyle = computed(() => {
  const colorObj = props.widgetData.widgetStyle || {};
  const c1 = colorObj.bgHex || '#ffffff';
  if (!colorObj.useGradient) return { backgroundColor: c1 };
  const c2 = colorObj.bgHex2 || c1; 
  const angle = colorObj.bgGradientDir || '135deg';
  return { background: `linear-gradient(${angle}, ${c1}, ${c2})` };
});

const config = computed(() => {
  const customData = props.widgetData?.customChartData || {};
  return {
    isStriped: customData.isStriped !== undefined ? customData.isStriped : true,
    isDense: customData.isDense !== undefined ? customData.isDense : false,
    showRowCount: customData.showRowCount !== undefined ? customData.showRowCount : true,
    maxRows: customData.maxRows !== undefined ? customData.maxRows : 10,
    headerColor: customData.headerColor || '#f8fafc',
    headerTextColor: customData.headerTextColor || '#334155',
    showTimeColumn: customData.showTimeColumn !== undefined ? customData.showTimeColumn : true 
  };
});

const liveStreamStore = useLiveStreamStore();
const liveTableRows = shallowRef([]);
const lastProcessedTimestamps = new Map();

const hasDevices = computed(() => props.widgetData?.deviceIds && props.widgetData.deviceIds.length > 0);
const hasData = computed(() => liveTableRows.value.length > 0);

const displayColumns = computed(() => {
  const cols = [];
  if (config.value.showTimeColumn) cols.push(t('common.time'));

  const rawDeviceIds = props.widgetData?.deviceIds || [];
  rawDeviceIds.forEach(rawId => {
    const id = String(rawId);
    const device = liveStreamStore.liveData[id];
    cols.push(device ? device.name : `${t('common.device')} ${id}`);
  });
  return cols;
});

onMounted(() => {
  if (props.widgetData?.deviceIds?.length) {
    liveStreamStore.registerDevices(widgetInstanceId, props.widgetData.deviceIds);
  }
});

onUnmounted(() => {
  liveStreamStore.unregisterDevices(widgetInstanceId);
  lastProcessedTimestamps.clear();
});

watch(() => liveStreamStore.liveData, (newData) => {
  const rawDeviceIds = props.widgetData?.deviceIds || [];
  if (rawDeviceIds.length === 0) return;

  // Check if at least one device has fresh telemetry
  let hasNewTelemetry = false;
  let latestEventTime = null;

  for (const rawId of rawDeviceIds) {
    const id = String(rawId);
    const dev = newData[id];

    if (!dev || dev.value === null || dev.value === undefined) continue;
    if (!dev.updatedValueAt) continue;

    const eventTime = new Date(dev.updatedValueAt).getTime();
    if (isNaN(eventTime)) continue;

    const lastTime = lastProcessedTimestamps.get(id);
    if (!lastTime || eventTime > lastTime) {
      hasNewTelemetry = true;
      lastProcessedTimestamps.set(id, eventTime);
      if (!latestEventTime || eventTime > latestEventTime) {
        latestEventTime = eventTime;
      }
    }
  }

  // Deduplication guard: ignore repeating database polls
  if (!hasNewTelemetry) return;

  const timeStr = formatTime(latestEventTime ? new Date(latestEventTime) : new Date());
  const newRow = {};

  if (config.value.showTimeColumn) {
    newRow[t('common.time')] = timeStr;
  }

  rawDeviceIds.forEach(rawId => {
    const id = String(rawId);
    const device = newData[id];
    const colName = device ? device.name : `${t('common.device')} ${id}`;
    newRow[colName] = (device && device.value !== null && device.value !== undefined) ? device.value : '-';
  });

  const nextRows = [newRow, ...liveTableRows.value];
  if (nextRows.length > config.value.maxRows) {
    nextRows.length = config.value.maxRows;
  }
  liveTableRows.value = nextRows;
});

watch(() => props.widgetData?.deviceIds, (newIds) => {
  liveTableRows.value = [];
  lastProcessedTimestamps.clear();
  liveStreamStore.registerDevices(widgetInstanceId, newIds || []);
}, { deep: false });

const displayData = computed(() => liveTableRows.value);
</script>