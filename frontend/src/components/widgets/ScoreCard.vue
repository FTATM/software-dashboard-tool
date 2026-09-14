<template>
  <div class="flex flex-col h-full w-full overflow-hidden rounded-box" :style="backgroundStyle">
    
    <div class="w-full px-4 py-2 shadow-sm shrink-0">
      <h3 class="m-0 text-sm sm:text-base font-bold tracking-wide text-left" :style="{ color: widgetData.widgetStyle?.textHex || '#ffffff' }">
        {{ widgetData?.widgetLabel }}
      </h3>
    </div>

    <div class="flex-1 w-full relative flex items-center justify-between px-6 py-4">
      
      <div v-if="!hasDevices" class="absolute inset-0 flex items-center justify-center text-sm italic text-center p-4" :style="{ color: widgetData.widgetStyle?.textHex || '#ffffff' }">
        {{ $t('common.noDevice') }}
      </div>

      <div v-else-if="isLoadingHistory" class="absolute inset-0 flex flex-col items-center justify-center text-sm gap-3" :style="{ color: widgetData.widgetStyle?.textHex || '#ffffff' }">
        <span class="loading loading-spinner loading-md"></span>
      </div>

      <template v-else-if="isReady">
        
        <div class="w-1/3 h-full flex items-center justify-start shrink-0">
          <Icon v-if="chartData.visualType === 'icon'" :icon="chartData.icon || 'lucide:activity'" 
                class="w-16 h-16 opacity-80" 
                :style="{ color: widgetData.widgetStyle?.chartHex || '#ffffff' }" />
          
          <v-chart v-else-if="chartData.visualType === 'line' || chartData.visualType === 'bar'" 
                   class="w-full h-16 opacity-80" :option="sparklineOption" autoresize />
        </div>
        
        <div class="flex-1 flex flex-col items-end justify-center text-right z-10 pl-2">
          
          <div class="flex items-baseline gap-1">
            <span v-if="chartData.prefix" class="text-2xl sm:text-3xl font-medium opacity-80" :style="{ color: widgetData.widgetStyle?.textHex || '#ffffff' }">
              {{ chartData.prefix }}
            </span>
            
            <span class="text-5xl sm:text-6xl font-extrabold tracking-tight drop-shadow-sm" :style="{ color: widgetData.widgetStyle?.textHex || '#ffffff' }">
              {{ displayValue }}
            </span>
            
            <span v-if="chartData.unit" class="text-xl sm:text-2xl font-semibold opacity-80 ml-1" :style="{ color: widgetData.widgetStyle?.textHex || '#ffffff' }">
              {{ chartData.unit }}
            </span>
          </div>
          
          <div class="mt-1 text-xs sm:text-sm font-semibold opacity-70 tracking-wide uppercase" :style="{ color: widgetData.widgetStyle?.textHex || '#ffffff' }">
            {{ chartData.customSubText || subTextLabel }}
          </div>

        </div>
      </template>

    </div>
  </div>
</template>

<script setup>
import { computed, ref, shallowRef, onMounted, onUnmounted, nextTick, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useLiveStreamStore } from '@/stores/useLiveStreamStore';
import { useFetch } from '@/composables/useFetch';
import { Icon } from '@iconify/vue';

import VChart from 'vue-echarts';
import { use } from 'echarts/core';
import { CanvasRenderer } from 'echarts/renderers';
import { LineChart, BarChart } from 'echarts/charts';
import { GridComponent } from 'echarts/components';

use([CanvasRenderer, LineChart, BarChart, GridComponent]);

const { t } = useI18n();

const props = defineProps({
  widgetData: { type: Object, default: () => ({}) }
});

const isReady = ref(false);
const isLoadingHistory = ref(false);
const historyValues = shallowRef([]); 
const liveDataHistory = shallowRef([]);

const liveStreamStore = useLiveStreamStore();
const { data: historyData, error: historyError, execute: fetchHistoryApi } = useFetch();

const widgetInstanceId = `scorecard-${props.widgetData?.id || Math.random().toString(36).substring(2, 9)}`;
const chartData = computed(() => props.widgetData?.customChartData || {});

const backgroundStyle = computed(() => {
  const colorObj = props.widgetData.widgetStyle || {};
  const c1 = colorObj.bgHex || '#ffffff';
  if (!colorObj.useGradient) return { backgroundColor: c1 };
  const c2 = colorObj.bgHex2 || c1; 
  const angle = colorObj.bgGradientDir || '135deg';
  return { background: `linear-gradient(${angle}, ${c1}, ${c2})` };
});

const hasDevices = computed(() => props.widgetData?.deviceIds && props.widgetData.deviceIds.length > 0);
const hasData = computed(() => {
  if (!hasDevices.value) return false;
  const mode = chartData.value.aggregationMode || 'live_single';
  if (mode.startsWith('history_')) return historyValues.value.length > 0;
  if (mode === 'live_count') return true; 
  return props.widgetData.deviceIds.some(id => {
    const item = liveStreamStore.liveData[String(id)];
    return item?.value !== undefined && item?.value !== null;
  });
});

const displayValue = computed(() => {
  const mode = chartData.value.aggregationMode || 'live_single';
  const decimals = chartData.value.decimalPlaces !== undefined ? chartData.value.decimalPlaces : 0;
  const ids = props.widgetData.deviceIds || [];
  if (ids.length === 0) return '--';
  let result = 0;

  if (mode.startsWith('live_')) {
    if (mode === 'live_count') return ids.length.toString();
    const activeVals = ids
      .map(id => liveStreamStore.liveData[String(id)]?.value)
      .filter(v => v !== undefined && v !== null)
      .map(Number);
    if (activeVals.length === 0) return '--';

    if (mode === 'live_single') result = activeVals[0];
    else if (mode === 'live_sum') result = activeVals.reduce((a, b) => a + b, 0);
    else if (mode === 'live_avg') result = activeVals.reduce((a, b) => a + b, 0) / activeVals.length;

  } else if (mode.startsWith('history_')) {
    if (historyValues.value.length === 0) return '--';
    if (mode === 'history_min') result = Math.min(...historyValues.value);
    else if (mode === 'history_max') result = Math.max(...historyValues.value);
    else if (mode === 'history_avg') result = historyValues.value.reduce((a, b) => a + b, 0) / historyValues.value.length;
  }
  return isNaN(result) ? '--' : result.toFixed(decimals);
});

watch(displayValue, (val) => {
  if (val !== '--' && !chartData.value.aggregationMode?.startsWith('history_')) {
    const newHistory = [...liveDataHistory.value, Number(val)];
    if (newHistory.length > 20) newHistory.shift(); 
    liveDataHistory.value = newHistory;
  }
});

const sparklineOption = computed(() => {
  const type = chartData.value.visualType === 'bar' ? 'bar' : 'line';
  const color = props.widgetData.widgetStyle?.chartHex || '#ffffff';
  
  const data = chartData.value.aggregationMode?.startsWith('history_') 
    ? historyValues.value 
    : liveDataHistory.value;

  return {
    grid: { left: 0, right: 0, top: 5, bottom: 5 },
    xAxis: { type: 'category', show: false, boundaryGap: type === 'bar' },
    yAxis: { type: 'value', show: false, scale: true },
    series: [{
      type: type, 
      data: data,
      itemStyle: { color: color, borderRadius: [2, 2, 0, 0] },
      lineStyle: { color: color, width: 3 },
      showSymbol: false,
      animation: false,
      sampling: type === 'line' ? 'lttb' : undefined,
      large: type === 'bar',
      largeThreshold: 100
    }]
  };
});

const subTextLabel = computed(() => {
  const mode = chartData.value.aggregationMode || 'live_single';
  const ids = props.widgetData.deviceIds || [];
  if (mode === 'live_single') {
    const firstId = ids[0] ? String(ids[0]) : null;
    return (firstId && liveStreamStore.liveData[firstId]?.name) || `${t('scoreCard.devicePrefix')} ${firstId || '?'}`;
  }
  if (mode === 'live_count') return `${ids.length} ${t('common.devices')}`;

  const range = chartData.value.historyRange || '1h';
  const rangeText = mode.startsWith('history_') ? ` (${range})` : ` ${t('scoreCard.live')}`;
  let prefix = '';
  if (mode.includes('sum')) prefix = t('scoreCard.total');
  if (mode.includes('avg')) prefix = t('scoreCard.average');
  if (mode.includes('min')) prefix = t('scoreCard.min');
  if (mode.includes('max')) prefix = t('scoreCard.max');
  return `${prefix} ${t('scoreCard.of')} ${ids.length} ${t('scoreCard.devs')}${rangeText}`;
});

const initializeHistory = async () => {
  const mode = chartData.value.aggregationMode || 'live_single';
  const rawDeviceIds = props.widgetData?.deviceIds || [];
  if (!mode.startsWith('history_') || rawDeviceIds.length === 0) return;
  
  const range = chartData.value.historyRange || '1h';
  let utcFrom, utcTo;
  const now = new Date();
  const past = new Date();
  if (range === 'custom') {
    utcFrom = new Date(chartData.value.customFrom).toISOString();
    utcTo = new Date(chartData.value.customTo).toISOString();
  } else {
    switch (range) {
      case '15m': past.setMinutes(now.getMinutes() - 15); break;
      case '30m': past.setMinutes(now.getMinutes() - 30); break;
      case '1h': past.setHours(now.getHours() - 1); break;
      case '3h': past.setHours(now.getHours() - 3); break;
      case '6h': past.setHours(now.getHours() - 6); break;
      case '24h': past.setHours(now.getHours() - 24); break;
      case '7d': past.setDate(now.getDate() - 7); break;
      case '30d': past.setDate(now.getDate() - 30); break; 
      case '3m': past.setMonth(now.getMonth() - 3); break;
      case '6m': past.setMonth(now.getMonth() - 6); break;
      case '1y': past.setFullYear(now.getFullYear() - 1); break;
    }
    utcFrom = past.toISOString();
    utcTo = now.toISOString();
  }

  isLoadingHistory.value = true;
  await fetchHistoryApi(`/device/charthistory?deviceIds=${rawDeviceIds.join(',')}&from=${encodeURIComponent(utcFrom)}&to=${encodeURIComponent(utcTo)}&maxPoints=${chartData.value.maxPoints || 1000}`);
  
  if (!historyError.value && historyData.value) {
    const flatValues = [];
    Object.values(historyData.value.data).forEach(pointsArr => {
      pointsArr.forEach(p => { 
        if (p[1] !== null && p[1] !== undefined) flatValues.push(Number(p[1])); 
      });
    });
    historyValues.value = flatValues;
  }
  isLoadingHistory.value = false;
};

onMounted(async () => {
  if (props.widgetData?.deviceIds?.length) {
    liveStreamStore.registerDevices(widgetInstanceId, props.widgetData.deviceIds);
  }
  await initializeHistory();
  await nextTick();
  setTimeout(() => { isReady.value = true; }, 50);
});

onUnmounted(() => {
  liveStreamStore.unregisterDevices(widgetInstanceId);
});

watch(() => props.widgetData?.deviceIds, (newIds) => {
  liveStreamStore.registerDevices(widgetInstanceId, newIds || []);
});

watch(
  () => [props.widgetData?.deviceIds, chartData.value.aggregationMode, chartData.value.historyRange], 
  (newVals, oldVals) => {
    if (JSON.stringify(newVals) === JSON.stringify(oldVals)) return;
    historyValues.value = [];
    liveDataHistory.value = [];
    initializeHistory();
  }, 
  { deep: true }
);
</script>