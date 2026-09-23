<template>
  <div class="flex flex-col h-full w-full p-4 overflow-hidden rounded-box" :style="backgroundStyle">
    <div class="backdrop-blur-md px-4 py-3 shadow-sm z-10 flex justify-between items-center rounded-lg">
      <h3 class="m-0 text-base font-extrabold tracking-wide"
        :style="{ color: widgetData.widgetStyle?.textHex || '#334155' }">
        {{ widgetData?.widgetLabel }}
      </h3>
      <span v-if="liveDeviceName && showDeviceName"
        class="badge badge-neutral badge-lg py-4 px-4 text-sm font-bold shrink-0 shadow-sm">
        {{ liveDeviceName }}
      </span>
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

      <div v-else-if="aggregatedCategories.length === 0"
        class="absolute inset-0 flex items-center justify-center text-sm italic text-center p-4"
        :style="{ color: widgetData.widgetStyle?.textHex || '#64748b' }">
        {{ $t('common.noDataAvailable') || 'No data available' }}
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
import { BarChart, LineChart } from 'echarts/charts';
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components';
import { useFetch } from '@/composables/useFetch';
import { useLiveStreamStore } from '@/stores/useLiveStreamStore';

use([CanvasRenderer, BarChart, LineChart, GridComponent, TooltipComponent, LegendComponent]);

const { t, locale } = useI18n();

const props = defineProps({
  widgetData: { type: Object, default: () => ({}) }
});

const isReady = ref(false);
const isLoadingHistory = ref(false);
const aggregatedCategories = shallowRef([]);
const aggregatedValues = shallowRef([]);
const aggregatedLineValues = shallowRef([]);
const lastProcessedTimestamp = ref(null);
const historyDeviceName = ref('');

const widgetInstanceId = `barline2-${props.widgetData?.id || Math.random().toString(36).substring(2, 9)}`;

const { data: historyData, error: historyError, execute: fetchHistoryApi } = useFetch();

const liveStreamStore = useLiveStreamStore();
const chartConfig = computed(() => props.widgetData?.customChartData || {});
const hasDevices = computed(() => props.widgetData?.deviceIds && props.widgetData.deviceIds.length > 0);
const deviceId = computed(() => hasDevices.value ? String(props.widgetData.deviceIds[0]) : null);
const liveDeviceName = computed(() => {
  if (!deviceId.value) return '';
  return liveStreamStore.liveData?.[deviceId.value]?.name || historyDeviceName.value || '';
});
const showDeviceName = computed(() => chartConfig.value.showDeviceName !== undefined ? chartConfig.value.showDeviceName : true);

const backgroundStyle = computed(() => {
  const colorObj = props.widgetData.widgetStyle || {};
  const c1 = colorObj.bgHex || '#ffffff';
  if (!colorObj.useGradient) return { backgroundColor: c1 };
  const c2 = colorObj.bgHex2 || c1;
  const angle = colorObj.bgGradientDir || '135deg';
  return { background: `linear-gradient(${angle}, ${c1}, ${c2})` };
});

const getBucketKey = (dateStr, interval, currentLocale) => {
  const d = new Date(dateStr);
  if (interval === 'hour') {
    return d.toLocaleString(currentLocale, { calendar: 'gregory', year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', hour12: false }) + ':00';
  } else if (interval === 'day') {
    return d.toLocaleString(currentLocale, { calendar: 'gregory', year: 'numeric', month: 'short', day: 'numeric' });
  } else if (interval === 'month') {
    return d.toLocaleString(currentLocale, { calendar: 'gregory', year: 'numeric', month: 'short' });
  }
  return '';
};

const calculateGlobalAverageLine = (categories, values) => {
  const totalAggregatedSum = values.reduce((sum, val) => sum + val, 0);
  const globalAverage = values.length > 0 ? totalAggregatedSum / values.length : 0;
  return categories.map(() => globalAverage);
};

const initializeAndBucketHistory = async () => {
  if (!deviceId.value) return;

  isLoadingHistory.value = true;

  const range = chartConfig.value.historyRange || '7d';
  const now = new Date();
  const past = new Date();

  switch (range) {
    case '24h': past.setHours(now.getHours() - 24); break;
    case '7d': past.setDate(now.getDate() - 7); break;
    case '30d': past.setDate(now.getDate() - 30); break;
    case '3m': past.setMonth(now.getMonth() - 3); break;
    case '6m': past.setMonth(now.getMonth() - 6); break;
    case '1y': past.setFullYear(now.getFullYear() - 1); break;
  }

  const utcFrom = encodeURIComponent(past.toISOString());
  const utcTo = encodeURIComponent(now.toISOString());

  await fetchHistoryApi(`/device/charthistory?deviceIds=${deviceId.value}&from=${utcFrom}&to=${utcTo}&maxPoints=5000`);

  const deviceHistory = historyData.value?.data?.[deviceId.value];
  if (!historyError.value && deviceHistory) {
    // Unpack data array and save device name
    const rawPoints = Array.isArray(deviceHistory) ? deviceHistory : (deviceHistory.data || []);
    if (deviceHistory.name) {
      historyDeviceName.value = deviceHistory.name;
    }

    const interval = chartConfig.value.bucketInterval || 'day';
    const aggregation = chartConfig.value.aggregationMode || 'sum';
    const buckets = {};
    const currentLocale = locale.value === 'th' ? 'th-TH' : 'en-GB';

    rawPoints.forEach(p => {
      if (p[1] === null || p[1] === undefined) return;
      const key = getBucketKey(p[0], interval, currentLocale);
      if (!buckets[key]) buckets[key] = [];
      buckets[key].push(Number(p[1]));
    });

    const categories = Object.keys(buckets);
    const values = categories.map(key => {
      const arr = buckets[key];
      if (!arr || arr.length === 0) return 0;
      if (aggregation === 'sum') return arr.reduce((a, b) => a + b, 0);
      if (aggregation === 'max') return arr.reduce((max, val) => Math.max(max, val), -Infinity);
      if (aggregation === 'min') return arr.reduce((min, val) => Math.min(min, val), Infinity);
      return arr.reduce((a, b) => a + b, 0) / arr.length;
    });

    aggregatedCategories.value = categories;
    aggregatedValues.value = values;
    aggregatedLineValues.value = calculateGlobalAverageLine(categories, values);

    if (rawPoints.length > 0) {
      lastProcessedTimestamp.value = rawPoints[rawPoints.length - 1][0];
    } else {
      lastProcessedTimestamp.value = null;
    }
  }

  isLoadingHistory.value = false;
};

// Real-time update logic
watch(
  () => liveStreamStore.liveData[deviceId.value],
  (newData) => {
    if (!newData || !isReady.value) return;

    // 1. Verify numeric value
    const rawVal = newData.value;
    if (rawVal === undefined || rawVal === null) return;

    // 2. Telemetry validation: ignore if updatedValueAt is NULL
    const streamTime = newData.updatedValueAt;
    if (!streamTime) return;

    // 3. Deduplicate repeating backend polls
    if (lastProcessedTimestamp.value && new Date(streamTime) <= new Date(lastProcessedTimestamp.value)) {
      return;
    }
    lastProcessedTimestamp.value = streamTime;

    // 4. Update buckets
    const interval = chartConfig.value.bucketInterval || 'day';
    const currentLocale = locale.value === 'th' ? 'th-TH' : 'en-GB';
    const currentKey = getBucketKey(streamTime, interval, currentLocale);

    const categories = [...aggregatedCategories.value];
    const values = [...aggregatedValues.value];
    const aggregation = chartConfig.value.aggregationMode || 'sum';
    const numericVal = Number(rawVal);

    const categoryIndex = categories.indexOf(currentKey);

    if (categoryIndex !== -1) {
      if (aggregation === 'sum') {
        values[categoryIndex] += numericVal;
      } else if (aggregation === 'max') {
        values[categoryIndex] = Math.max(values[categoryIndex], numericVal);
      } else if (aggregation === 'min') {
        values[categoryIndex] = Math.min(values[categoryIndex], numericVal);
      }
    } else {
      categories.push(currentKey);
      values.push(numericVal);
    }

    aggregatedCategories.value = categories;
    aggregatedValues.value = values;
    aggregatedLineValues.value = calculateGlobalAverageLine(categories, values);
  },
  { deep: true }
);

onMounted(async () => {
  if (props.widgetData?.deviceIds?.length) {
    liveStreamStore.registerDevices(widgetInstanceId, props.widgetData.deviceIds);
  }
  await initializeAndBucketHistory();
  await nextTick();
  setTimeout(() => { isReady.value = true; }, 50);
});

onUnmounted(() => {
  liveStreamStore.unregisterDevices(widgetInstanceId);
});

watch(
  () => props.widgetData?.deviceIds,
  (newIds) => {
    liveStreamStore.registerDevices(widgetInstanceId, newIds || []);
    initializeAndBucketHistory();
  }
);

watch(
  () => [
    chartConfig.value.historyRange,
    chartConfig.value.bucketInterval,
    chartConfig.value.aggregationMode,
    locale.value,
  ],
  () => {
    initializeAndBucketHistory();
  }
);

const chartOption = computed(() => {
  const textColor = props.widgetData.widgetStyle?.textHex || '#334155';
  const barColor = chartConfig.value.barColor || '#3b82f6';
  const lineColor = chartConfig.value.lineColor || '#f97316';
  const barName = chartConfig.value.barName || 'Bar Value';
  const lineName = t('barLineChart.config.lineGlobalAvg');
  const yAxisName = chartConfig.value.yAxisName || '';

  return {
    textStyle: { color: textColor },
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'cross' },
      valueFormatter: (value) => (value !== undefined && value !== null ? Number(value).toFixed(2) : '-')
    },
    legend: {
      data: [barName, lineName],
      bottom: 0,
      textStyle: { color: textColor }
    },
    grid: { left: '3%', right: '4%', bottom: '12%', top: '15%', containLabel: true },
    xAxis: {
      type: 'category',
      data: aggregatedCategories.value,
      axisPointer: { type: 'shadow' },
      axisLabel: {
        color: textColor,
        hideOverlap: true,
        formatter: (value) => {
          if (value.includes(':')) {
            const lastSpaceIndex = value.lastIndexOf(' ');
            return value.substring(0, lastSpaceIndex) + '\n' + value.substring(lastSpaceIndex + 1);
          }
          return value;
        }
      }
    },
    yAxis: {
      type: 'value',
      name: yAxisName,
      nameTextStyle: { color: textColor },
      axisLabel: { color: textColor },
      splitLine: { lineStyle: { color: textColor, opacity: 0.1 } }
    },
    series: [
      {
        name: barName,
        type: 'bar',
        barWidth: '40%',
        itemStyle: { color: barColor, borderRadius: [4, 4, 0, 0] },
        large: true,
        largeThreshold: 400,
        animation: true,
        data: aggregatedValues.value
      },
      {
        name: lineName,
        type: 'line',
        smooth: chartConfig.value.isSmooth ?? true,
        itemStyle: { color: lineColor },
        lineStyle: { width: 3, color: lineColor },
        showSymbol: false,
        animation: true,
        data: aggregatedLineValues.value
      }
    ]
  };
});
</script>