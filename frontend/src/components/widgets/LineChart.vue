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
import { use } from 'echarts/core';
import { CanvasRenderer } from 'echarts/renderers';
import { LineChart } from 'echarts/charts';
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components';
import { useLiveStreamStore } from '@/stores/useLiveStreamStore';
import { useFetch } from '@/composables/useFetch';
import { useFormatter } from '@/composables/useFormatter';

const { formatTime } = useFormatter();
use([CanvasRenderer, LineChart, GridComponent, TooltipComponent, LegendComponent]);

const { t } = useI18n();

const props = defineProps({
  widgetData: { type: Object, default: () => ({}) }
});

const widgetInstanceId = `linechart2-${props.widgetData?.id || Math.random().toString(36).substring(2, 9)}`;

const backgroundStyle = computed(() => {
  const colorObj = props.widgetData.widgetStyle || {};
  const c1 = colorObj.bgHex || '#ffffff';

  if (!colorObj.useGradient) {
    return { backgroundColor: c1 };
  }

  const c2 = colorObj.bgHex2 || c1;
  const angle = colorObj.bgGradientDir || '135deg';

  return {
    background: `linear-gradient(${angle}, ${c1}, ${c2})`
  };
});

const isReady = ref(false);
const isLoadingHistory = ref(false);
const liveStreamStore = useLiveStreamStore();
const { data: historyData, error: historyError, execute: fetchHistoryApi } = useFetch();

const deviceSeries = shallowRef({});
const lastProcessedTimestamps = new Map();

const hasDevices = computed(() => props.widgetData?.deviceIds && props.widgetData.deviceIds.length > 0);
const hasData = computed(() => Object.values(deviceSeries.value).some(s => s.data && s.data.length > 0));

const config = computed(() => {
  const customData = props.widgetData?.customChartData || {};
  return {
    historyRange: customData.historyRange || '1h',
    customFrom: customData.customFrom || '',
    customTo: customData.customTo || '',
    maxPoints: customData.maxPoints || 100,
    isSmooth: customData.isSmooth !== undefined ? customData.isSmooth : true,
    showArea: customData.showArea !== undefined ? customData.showArea : true,
    isStacked: customData.isStacked !== undefined ? customData.isStacked : false,
    yAxisName: customData.yAxisName || '',
    deviceColorsMap: customData.deviceColors || {}
  };
});

// Binary search to find the nearest recorded point in time
const findClosestPoint = (data, targetTime) => {
  if (!data || data.length === 0) return null;
  let low = 0;
  let high = data.length - 1;

  while (low <= high) {
    const mid = (low + high) >> 1;
    if (data[mid][0] === targetTime) return data[mid];
    if (data[mid][0] < targetTime) low = mid + 1;
    else high = mid - 1;
  }

  if (low >= data.length) return data[data.length - 1];
  if (high < 0) return data[0];

  const d1 = Math.abs(data[low][0] - targetTime);
  const d2 = Math.abs(data[high][0] - targetTime);
  return d1 < d2 ? data[low] : data[high];
};

// Dynamically calculate acceptable gap based on total view span & point density
const getToleranceMs = () => {
  let totalSpanMs = 0;
  const range = config.value.historyRange;

  if (range === 'custom' && config.value.customFrom) {
    const fromTime = new Date(config.value.customFrom).getTime();
    const toTime = config.value.customTo ? new Date(config.value.customTo).getTime() : Date.now();
    totalSpanMs = Math.max(toTime - fromTime, 0);
  } else {
    switch (range) {
      case '15m': totalSpanMs = 15 * 60 * 1000; break;
      case '30m': totalSpanMs = 30 * 60 * 1000; break;
      case '1h':  totalSpanMs = 60 * 60 * 1000; break;
      case '3h':  totalSpanMs = 3 * 60 * 60 * 1000; break;
      case '6h':  totalSpanMs = 6 * 60 * 60 * 1000; break;
      case '24h': totalSpanMs = 24 * 60 * 60 * 1000; break;
      case '7d':  totalSpanMs = 7 * 24 * 60 * 60 * 1000; break;
      default:    totalSpanMs = 60 * 60 * 1000;
    }
  }

  const maxPoints = config.value.maxPoints || 100;
  // Calculate average interval between points
  const avgBucketInterval = totalSpanMs / maxPoints;

  // Allow up to 1.5x the bucket interval, with a minimum floor of 30 seconds for live streaming
  return Math.max(avgBucketInterval * 1.5, 30 * 1000);
};

const initializeHistory = async () => {
  const rawDeviceIds = props.widgetData?.deviceIds || [];
  if (rawDeviceIds.length === 0 || config.value.historyRange === '0') return;

  if (config.value.historyRange === 'custom' && !config.value.customFrom) {
    return;
  }

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

  if (!historyError.value && historyData.value?.data) {
    const newSeries = {};
    Object.entries(historyData.value.data).forEach(([id, item]) => {
      const deviceIdStr = String(id);
      const pointsArr = Array.isArray(item) ? item : (item?.data || []);
      const deviceName = item?.name || `${t('common.device')} ${id}`;

      newSeries[deviceIdStr] = { 
        name: deviceName, 
        data: pointsArr 
      };

      // Seed latest timestamp to prevent duplicate live additions
      if (pointsArr && pointsArr.length > 0) {
        const lastPoint = pointsArr[pointsArr.length - 1];
        lastProcessedTimestamps.set(deviceIdStr, Number(lastPoint[0]));
      }
    });
    deviceSeries.value = newSeries;
  } else {
    console.error(historyError.value?.message || t('lineChart.messages.fetchHistoryFailed'));
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
      if (lastTime && eventTime <= lastTime) {
        return;
      }
      lastProcessedTimestamps.set(id, eventTime);

      if (!nextSeries[id]) {
        nextSeries[id] = { name: device.name || `${t('common.device')} ${id}`, data: [] };
      }
      if (device.name) {
        nextSeries[id].name = device.name;
      }

      const currentData = nextSeries[id].data;
      if (currentData.length > 0) {
        const lastPoint = currentData[currentData.length - 1];
        if (lastPoint[0] === eventTime) {
          return;
        }
      }
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
  const rawDeviceIds = props.widgetData?.deviceIds || [];
  const chartTextColor = props.widgetData.widgetStyle?.textHex || '#334155';

  const legendNames = [];
  const dynamicSeries = [];

  rawDeviceIds.forEach((rawId, index) => {
    const id = String(rawId);
    const dataObj = deviceSeries.value[id];
    if (!dataObj) return;

    const color = config.value.deviceColorsMap[id] || fallbackColors[index % fallbackColors.length];
    legendNames.push(dataObj.name);

    dynamicSeries.push({
      name: dataObj.name,
      type: 'line',
      smooth: config.value.isSmooth,
      stack: config.value.isStacked ? t('lineChart.total') : null,
      itemStyle: { color: color },
      areaStyle: config.value.showArea ? { color: color, opacity: 0.2 } : null,
      showSymbol: false,
      sampling: 'lttb',
      animation: true,
      data: dataObj.data
    });
  });

  return {
    textStyle: { color: chartTextColor },
    tooltip: {
      trigger: 'axis',
      formatter: (params) => {
        if (!params || !params.length) return '';
        const hoverTime = params[0].value[0];
        const timeStr = formatTime(hoverTime);
        const toleranceMs = getToleranceMs();

        let tipHtml = `<strong>${timeStr}</strong><br/>`;

        // Map series that match the hovered tick exactly
        const matchedMap = new Map();
        params.forEach(p => {
          matchedMap.set(p.seriesName, p);
        });

        // Check each line series
        dynamicSeries.forEach(s => {
          const matched = matchedMap.get(s.name);
          let valStr = '-';

          if (matched && matched.value[1] !== undefined && matched.value[1] !== null) {
            valStr = Number(matched.value[1]).toFixed(2);
          } else if (s.data && s.data.length > 0) {
            const closest = findClosestPoint(s.data, hoverTime);
            // Only show value if within the dynamic tolerance window
            if (closest && Math.abs(closest[0] - hoverTime) <= toleranceMs) {
              if (closest[1] !== undefined && closest[1] !== null) {
                valStr = Number(closest[1]).toFixed(2);
              }
            }
          }

          const marker = `<span style="display:inline-block;margin-right:4px;border-radius:10px;width:10px;height:10px;background-color:${s.itemStyle.color};"></span>`;
          tipHtml += `${marker} <span style="color:${chartTextColor}">${s.name}: <b>${valStr}</b></span><br/>`;
        });

        return tipHtml;
      }
    },
    legend: {
      type: 'scroll',
      data: legendNames,
      bottom: 5,
      left: 10,
      right: 70,
      textStyle: { color: chartTextColor }
    },
    grid: { left: '3%', right: '4%', bottom: '12%', top: '15%', containLabel: true },
    xAxis: {
      type: 'time',
      boundaryGap: false,
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
      nameTextStyle: { color: chartTextColor },
      axisLabel: { color: chartTextColor },
      scale: true
    },
    series: dynamicSeries
  };
});
</script>