<template>
  <div class="flex flex-col h-full w-full p-4 overflow-hidden" :style="backgroundStyle">
    
    <div class="backdrop-blur-md px-4 py-3 shadow-sm z-10 flex justify-between items-center rounded-lg">
      <h3 class="m-0 text-base font-extrabold tracking-wide truncate pr-2" :style="{ color: widgetData.widgetStyle?.textHex || '#334155' }">
        {{ widgetData?.widgetLabel }}
      </h3>
      <span v-if="hasData && liveDeviceName && showDeviceName"
        class="badge badge-neutral badge-lg py-4 px-4 text-sm font-bold shrink-0 shadow-sm">
        {{ liveDeviceName }}
      </span>
    </div>
    
    <div class="flex-1 w-full min-h-[150px] relative flex flex-col justify-center mt-2">
      
      <div v-if="!hasDevices" class="absolute inset-0 flex items-center justify-center text-sm italic text-center p-4" :style="{ color: widgetData.widgetStyle?.textHex || '#64748b' }">
        {{ $t('common.noDevicesConfig') }}
      </div>

      <div v-else-if="!hasData" class="absolute inset-0 flex flex-col items-center justify-center text-sm gap-3" :style="{ color: widgetData.widgetStyle?.textHex || '#64748b' }">
        <span class="loading loading-spinner loading-md text-primary"></span>
        {{ $t('common.waitingData') }}
      </div>

      <v-chart v-else-if="isReady" class="absolute inset-0 w-full h-full" :option="chartOption" autoresize />

    </div>
  </div>
</template>

<script setup>
import { computed, ref, onMounted, onUnmounted, nextTick, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import VChart from 'vue-echarts';
import { useLiveStreamStore } from '@/stores/useLiveStreamStore';
import { use } from 'echarts/core';
import { CanvasRenderer } from 'echarts/renderers';
import { BarChart, ScatterChart } from 'echarts/charts';
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components';

use([CanvasRenderer, BarChart, ScatterChart, GridComponent, TooltipComponent, LegendComponent]);

const { t } = useI18n();

const props = defineProps({
  widgetData: { type: Object, default: () => ({}) }
});

const isReady = ref(false);
const liveStreamStore = useLiveStreamStore();
const widgetInstanceId = `bullet-${props.widgetData?.id || Math.random().toString(36).substring(2, 9)}`;

const backgroundStyle = computed(() => {
  const colorObj = props.widgetData.widgetStyle || {};
  const c1 = colorObj.bgHex || '#ffffff';
  if (!colorObj.useGradient) return { backgroundColor: c1 };
  const c2 = colorObj.bgHex2 || c1; 
  const angle = colorObj.bgGradientDir || '135deg';
  return { background: `linear-gradient(${angle}, ${c1}, ${c2})` };
});

const hasDevices = computed(() => props.widgetData?.deviceIds && props.widgetData.deviceIds.length > 0);
const deviceId = computed(() => hasDevices.value ? String(props.widgetData.deviceIds[0]) : null);

const deviceObj = computed(() => (deviceId.value ? liveStreamStore.liveData[deviceId.value] : null));
const hasData = computed(() => deviceObj.value?.value !== undefined && deviceObj.value?.value !== null);
const liveValue = computed(() => (hasData.value ? Number(deviceObj.value.value) : 0));
const liveDeviceName = computed(() => deviceObj.value?.name || '');

const showDeviceName = computed(() => {
  const custom = props.widgetData?.customChartData || {};
  return custom.showDeviceName !== undefined ? custom.showDeviceName : true;
});

onMounted(async () => {
  if (props.widgetData?.deviceIds?.length) {
    liveStreamStore.registerDevices(widgetInstanceId, props.widgetData.deviceIds);
  }
  await nextTick();
  setTimeout(() => { isReady.value = true; }, 50);
});

onUnmounted(() => {
  liveStreamStore.unregisterDevices(widgetInstanceId);
});

watch(() => props.widgetData?.deviceIds, (newIds) => {
  liveStreamStore.registerDevices(widgetInstanceId, newIds || []);
});

const chartOption = computed(() => {
  const customData = props.widgetData?.customChartData || {};
  const chartTextColor = props.widgetData.widgetStyle?.textHex || '#334155';
  
  const barColor = customData.barColor || '#1A5FB4';
  const targetColor = customData.targetColor || '#26A269';
  const target = customData.targetValue !== undefined ? customData.targetValue : 85;
  const unit = customData.unit || '';
  const xAxisMax = customData.xAxisMax || null; 

  const thresholds = customData.thresholds || [
    { name: t('bulletChart.config.defaults.excellent'), value: 120, color: '#e2e8f0' },
    { name: t('bulletChart.config.defaults.poor'), value: 60, color: '#94a3b8' }
  ];

  const sortedThresholds = [...thresholds].sort((a, b) => b.value - a.value);

  const dynamicSeries = sortedThresholds.map((zone, index) => {
    return {
      name: zone.name,
      type: 'bar',
      barWidth: '50%',
      barGap: index === 0 ? '0%' : '-100%',
      data: [zone.value],
      itemStyle: { color: zone.color },
      animation: false,
      tooltip: { valueFormatter: (value) => `${value}${unit}` }
    };
  });

  dynamicSeries.push({
    name: t('bulletChart.actualValue'),
    type: 'bar',
    barWidth: '20%',
    barGap: '-100%',
    data: [liveValue.value],
    itemStyle: { color: barColor },
    z: 3,
    tooltip: { valueFormatter: (value) => `${value}${unit}` }
  });

  dynamicSeries.push({
    name: t('bulletChart.target'),
    type: 'scatter',
    symbol: 'rect',
    symbolSize: [4, 40],
    data: [target],
    itemStyle: { color: targetColor },
    z: 4,
    tooltip: { valueFormatter: (value) => `${value}${unit}` }
  });

  return {
    textStyle: { color: chartTextColor },
    tooltip: { trigger: 'item', axisPointer: { type: 'shadow' } },
    legend: { bottom: 0, icon: 'circle', itemWidth: 10, textStyle: { color: chartTextColor } },
    grid: { left: '2%', right: '5%', bottom: '15%', top: '10%', containLabel: true },
    xAxis: { type: 'value', max: xAxisMax, splitLine: { show: false }, axisLabel: { color: chartTextColor, formatter: `{value}${unit}` } },
    yAxis: { type: 'category', data: ['Metric'], axisLine: { show: false }, axisTick: { show: false }, axisLabel: { show: false } },
    series: dynamicSeries
  };
});
</script>