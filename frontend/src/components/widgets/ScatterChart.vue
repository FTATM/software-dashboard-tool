<template>
  <div class="flex flex-col h-full w-full p-4 overflow-hidden" :style="backgroundStyle">

    <div class="backdrop-blur-md px-4 py-3 shadow-sm z-10 flex justify-between items-center rounded-lg">
      <h3 class="m-0 text-base font-extrabold tracking-wide"
        :style="{ color: widgetData.widgetStyle?.textHex || '#334155' }">
        {{ widgetData?.widgetLabel }}
      </h3>
    </div>

    <div class="flex-1 w-full min-h-0 relative flex flex-col justify-center mt-2">

      <div v-if="!hasDevices" class="absolute inset-0 flex items-center justify-center text-sm italic text-center p-4"
        :style="{ color: widgetData.widgetStyle?.textHex || '#64748b' }">
        {{ $t('common.noDevicesConfig') }}
      </div>

      <div v-else-if="isLoadingHistory" class="absolute inset-0 flex flex-col items-center justify-center text-sm gap-3"
        :style="{ color: widgetData.widgetStyle?.textHex || '#64748b' }">
        <span class="loading loading-spinner loading-md text-primary"></span>
        {{ $t('common.fetchingHistory') }}
      </div>

      <div v-else-if="!hasData" class="absolute inset-0 flex flex-col items-center justify-center text-sm gap-3"
        :style="{ color: widgetData.widgetStyle?.textHex || '#64748b' }">
        <span class="loading loading-spinner loading-md text-primary"></span>
        {{ $t('common.waitingData') }}
      </div>

      <v-chart v-else-if="isReady" class="absolute inset-0 w-full h-full" :option="chartOption" autoresize />

    </div>
  </div>
</template>

<script setup>
import { computed, ref, shallowRef, onMounted, onUnmounted, nextTick, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import VChart from 'vue-echarts';
import { useLiveStreamStore } from '@/stores/useLiveStreamStore';
import { useFetch } from '@/composables/useFetch';
import { use } from 'echarts/core';
import { CanvasRenderer } from 'echarts/renderers';
import { ScatterChart, LineChart } from 'echarts/charts';
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components';
import { useFormatter } from '@/composables/useFormatter';

const { formatTime } = useFormatter();

use([CanvasRenderer, ScatterChart, LineChart, GridComponent, TooltipComponent, LegendComponent]);

const { t } = useI18n();

const props = defineProps({
  widgetData: { type: Object, default: () => ({}) }
});

const isReady = ref(false);
const isLoadingHistory = ref(false);
const liveStreamStore = useLiveStreamStore();
const { data: historyData, error: historyError, execute: fetchHistoryApi } = useFetch();
const deviceSeries = shallowRef({});
const lastProcessedTimestamps = new Map();

const widgetInstanceId = `scatter-${props.widgetData?.id || Math.random().toString(36).substring(2, 9)}`;

const backgroundStyle = computed(() => {
  const colorObj = props.widgetData.widgetStyle || {};
  const c1 = colorObj.bgHex || '#ffffff';
  if (!colorObj.useGradient) return { backgroundColor: c1 };
  const c2 = colorObj.bgHex2 || c1;
  const angle = colorObj.bgGradientDir || '135deg';
  return { background: `linear-gradient(${angle}, ${c1}, ${c2})` };
});

const hasDevices = computed(() => props.widgetData?.deviceIds && props.widgetData.deviceIds.length > 0);
const hasData = computed(() => Object.values(deviceSeries.value).some(s => s.data && s.data.length > 0));

const config = computed(() => {
  const customData = props.widgetData?.customChartData || {};
  return {
    historyRange: customData.historyRange || '1h',
    customFrom: customData.customFrom || '',
    customTo: customData.customTo || '',
    maxPoints: customData.maxPoints || 100,
    showRegression: customData.showRegression !== undefined ? customData.showRegression : true,
    xAxisName: customData.xAxisName || '',
    yAxisName: customData.yAxisName || '',
    deviceColorsMap: customData.deviceColors || {}
  };
});

const calculateRegressionLine = (data) => {
  const n = data.length;
  if (n < 2) return [];
  const minX = data[0][0];
  const maxX = data[n - 1][0];
  let sumX = 0, sumY = 0, sumXY = 0, sumXX = 0;
  data.forEach(point => {
    const normX = point[0] - minX;
    sumX += normX;
    sumY += point[1];
    sumXY += normX * point[1];
    sumXX += normX * normX;
  });
  const denominator = (n * sumXX - sumX * sumX);
  if (denominator === 0) return [];
  const slope = (n * sumXY - sumX * sumY) / denominator;
  const intercept = (sumY - slope * sumX) / n;
  return [[minX, intercept], [maxX, slope * (maxX - minX) + intercept]];
};

const initializeHistory = async () => {
  const rawDeviceIds = props.widgetData?.deviceIds || [];
  if (rawDeviceIds.length === 0 || config.value.historyRange === '0') return;

  if (config.value.historyRange === 'custom' && !config.value.customFrom) return;

  isLoadingHistory.value = true;
  const idQuery = rawDeviceIds.join(',');

  let utcFrom, utcTo;
  if (config.value.historyRange === 'custom') {
    utcFrom = new Date(config.value.customFrom).toISOString();
    utcTo = config.value.customTo ? new Date(config.value.customTo).toISOString() : new Date().toISOString();
  } else {
    const now = new Date();
    const past = new Date();
    switch (config.value.historyRange) {
      case '15m': past.setMinutes(now.getMinutes() - 15); break;
      case '30m': past.setMinutes(now.getMinutes() - 30); break;
      case '1h': past.setHours(now.getHours() - 1); break;
      case '3h': past.setHours(now.getHours() - 3); break;
      case '6h': past.setHours(now.getHours() - 6); break;
      case '24h': past.setHours(now.getHours() - 24); break;
      case '7d': past.setDate(now.getDate() - 7); break;
    }
    utcFrom = past.toISOString();
    utcTo = now.toISOString();
  }

  const fromQuery = encodeURIComponent(utcFrom);
  const toQuery = encodeURIComponent(utcTo);
  const apiUrl = `/device/charthistory?deviceIds=${idQuery}&from=${fromQuery}&to=${toQuery}&maxPoints=${config.value.maxPoints}`;

  await fetchHistoryApi(apiUrl);
  if (!historyError.value && historyData.value) {
    const newSeries = {};
    Object.entries(historyData.value.data).forEach(([id, pointsArr]) => {
      const deviceIdStr = String(id);
      newSeries[deviceIdStr] = { name: `${t('common.device')} ${id}`, data: pointsArr || [] };

      if (pointsArr && pointsArr.length > 0) {
        const lastPoint = pointsArr[pointsArr.length - 1];
        lastProcessedTimestamps.set(deviceIdStr, Number(lastPoint[0]));
      }
    });
    deviceSeries.value = newSeries;
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
  lastProcessedTimestamps.clear();
});

watch(
  () => liveStreamStore.liveData,
  (newData) => {
    if (config.value.historyRange === 'custom' && config.value.customFrom && config.value.customTo) {
      return;
    }
    const rawDeviceIds = props.widgetData?.deviceIds || [];
    if (rawDeviceIds.length === 0) return;

    // ⚡ FIX: Early exit if no devices on THIS scatter chart received newer data
    const hasRelevantUpdate = rawDeviceIds.some(rawId => {
      const id = String(rawId);
      const dev = newData[id];
      if (!dev?.updatedValueAt || dev.value === null || dev.value === undefined) return false;
      const last = lastProcessedTimestamps.get(id);
      return !last || new Date(dev.updatedValueAt).getTime() > last;
    });

    if (!hasRelevantUpdate) return;

    let hasModifications = false;
    const nextSeries = { ...deviceSeries.value };

    rawDeviceIds.forEach(rawId => {
      const id = String(rawId);
      const device = newData[id];

      if (!device || device.value === null || device.value === undefined) return;
      if (!device.updatedValueAt) return;

      const eventTime = new Date(device.updatedValueAt).getTime();
      if (isNaN(eventTime)) return;

      const lastTime = lastProcessedTimestamps.get(id);
      if (lastTime && eventTime <= lastTime) return;
      lastProcessedTimestamps.set(id, eventTime);

      if (!nextSeries[id]) {
        nextSeries[id] = { name: device.name || `${t('common.device')} ${id}`, data: [] };
      }
      if (device.name) {
        nextSeries[id].name = device.name;
      }

      // ⚡ FIX: Mutate array in-place without spreading
      const currentData = nextSeries[id].data;
      currentData.push([eventTime, Number(device.value)]);

      if (currentData.length > config.value.maxPoints) {
        const overflow = currentData.length - config.value.maxPoints;
        currentData.splice(0, overflow);
      }

      hasModifications = true;
    });

    if (hasModifications) {
      deviceSeries.value = nextSeries;
    }
  }
);

watch(() => props.widgetData?.deviceIds, (newIds) => {
  deviceSeries.value = {};
  lastProcessedTimestamps.clear();
  liveStreamStore.registerDevices(widgetInstanceId, newIds || []);
  initializeHistory();
}, { deep: false });

watch(
  () => [config.value.historyRange, config.value.maxPoints, config.value.customFrom, config.value.customTo],
  (newVals, oldVals) => {
    if (JSON.stringify(newVals) === JSON.stringify(oldVals)) return;
    if (newVals[0] === 'custom' && !newVals[2]) return;

    deviceSeries.value = {};
    lastProcessedTimestamps.clear();
    initializeHistory();
  },
  { deep: true }
);

const chartOption = computed(() => {
  const fallbackColors = [
    '#3b82f6', '#10b981', '#f59e0b', '#ef4444',
    '#8b5cf6', '#06b6d4', '#84cc16', '#ec4899',
    '#f97316', '#14b8a6', '#6366f1', '#eab308',
    '#d946ef', '#0ea5e9', '#22c55e', '#f43f5e'
  ];
  const chartTextColor = props.widgetData.widgetStyle?.textHex || '#334155';
  const rawDeviceIds = props.widgetData?.deviceIds || [];

  const dynamicSeries = [];
  const legendNames = [];

  rawDeviceIds.forEach((rawId, index) => {
    const id = String(rawId);
    const dataObj = deviceSeries.value[id];
    if (!dataObj) return;

    const color = config.value.deviceColorsMap[id] || fallbackColors[index % fallbackColors.length];
    const actualName = dataObj.name;
    const rawData = dataObj.data;
    legendNames.push(actualName);
    dynamicSeries.push({
      name: actualName,
      type: 'scatter',
      itemStyle: { color: color, opacity: 0.8 },
      symbolSize: 12,
      large: true,
      largeThreshold: 500,
      animation: true,
      data: rawData
    });

    if (config.value.showRegression) {
      const trendName = `${actualName} (${t('scatterChart.trend')})`;
      legendNames.push(trendName);
      dynamicSeries.push({
        name: trendName,
        type: 'line',
        showSymbol: false,
        smooth: true,
        sampling: 'lttb',
        animation: true,
        itemStyle: { color: color },
        lineStyle: { width: 2, type: 'dashed' },
        data: calculateRegressionLine(rawData)
      });
    }
  });

  return {
    textStyle: { color: chartTextColor },
    tooltip: {
      trigger: 'item',
      formatter: (params) => {
        const timeStr = formatTime(params.value[0]);
        const yValue = params.value[1] !== undefined && params.value[1] !== null ? Number(params.value[1]).toFixed(2) : '-';
        return `<strong>${params.seriesName}</strong><br/>${timeStr}<br/>${t('scatterChart.value')}: <span style="color:${chartTextColor}">${yValue}</span>`;
      }
    },
    legend: {
      type: 'scroll',
      show: true,
      bottom: 5,
      left: 10,
      right: 70,
      data: legendNames,
      textStyle: { color: chartTextColor },
    },
    grid: { left: '3%', right: '8%', bottom: '12%', top: '15%', containLabel: true },
    xAxis: {
      type: 'time',
      name: config.value.xAxisName,
      nameTextStyle: { fontWeight: 'bold', color: chartTextColor },
      boundaryGap: false,
      splitNumber: 4,
      axisLabel: {
        color: chartTextColor,
        hideOverlap: true,
        formatter: (value) => {
          const fullDateTime = formatTime(value);
          return fullDateTime.replace(' ', '\n');
        }
      }
    },
    yAxis: {
      type: 'value',
      name: config.value.yAxisName,
      nameTextStyle: { fontWeight: 'bold', color: chartTextColor },
      axisLabel: { color: chartTextColor },
      scale: true
    },
    series: dynamicSeries
  };
});
</script>