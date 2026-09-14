<template>
  <div class="flex flex-col h-full w-full p-4 overflow-hidden" :style="backgroundStyle">

    <div class="backdrop-blur-md px-4 py-3 shadow-sm z-10 flex justify-between items-center rounded-lg">
      <h3 class="m-0 text-base font-extrabold tracking-wide"
        :style="{ color: widgetData.widgetStyle?.textHex || '#334155' }">
        {{ widgetData?.widgetLabel }}
      </h3>
    </div>

    <div class="flex-1 w-full min-h-[150px] relative flex flex-col justify-center mt-2">

      <div v-if="!hasDevices" class="absolute inset-0 flex items-center justify-center text-sm italic text-center p-4"
        :style="{ color: widgetData.widgetStyle?.textHex || '#64748b' }">
        {{ $t('common.noDevicesConfig') }}
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
import { computed, ref, onMounted, onUnmounted, nextTick, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import VChart from 'vue-echarts';
import { useLiveStreamStore } from '@/stores/useLiveStreamStore';
import { use } from 'echarts/core';
import { CanvasRenderer } from 'echarts/renderers';
import { BarChart } from 'echarts/charts';
import { GridComponent, TooltipComponent, TitleComponent } from 'echarts/components';

use([CanvasRenderer, BarChart, GridComponent, TooltipComponent, TitleComponent]);

const { t } = useI18n();

const props = defineProps({
  widgetData: { type: Object, default: () => ({}) }
});

const isReady = ref(false);
const liveStreamStore = useLiveStreamStore();
const widgetInstanceId = `barprocess-${props.widgetData?.id || Math.random().toString(36).substring(2, 9)}`;

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
  const ids = props.widgetData?.deviceIds || [];
  return ids.some(id => {
    const item = liveStreamStore.liveData[String(id)];
    return item?.value !== undefined && item?.value !== null;
  });
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

  const trackColor = customData.trackColor || '#e2e8f0';
  const maxValue = customData.maxValue && customData.maxValue > 0 ? customData.maxValue : 100;
  const unit = customData.unit || '';
  const barThickness = customData.barThickness || 16;
  const borderRadius = customData.borderRadius !== undefined ? customData.borderRadius : 8;
  const showTextLabel = customData.showTextLabel !== undefined ? customData.showTextLabel : true;
  const deviceColorsMap = customData.deviceColors || {};
  const fallbackColors = [
    '#3b82f6', '#10b981', '#f59e0b', '#ef4444',
    '#8b5cf6', '#06b6d4', '#84cc16', '#ec4899',
    '#f97316', '#14b8a6', '#6366f1', '#eab308',
    '#d946ef', '#0ea5e9', '#22c55e', '#f43f5e'
  ];

  const rawDeviceIds = props.widgetData?.deviceIds || [];
  const deviceNames = [];
  const deviceValues = [];

  rawDeviceIds.forEach(rawId => {
    const id = String(rawId);
    const data = liveStreamStore.liveData[id];
    deviceNames.push(data ? data.name : t('common.loading'));
    deviceValues.push((data && data.value !== null) ? data.value : 0);
  });

  const backgroundValues = deviceNames.map(() => maxValue);

  return {
    textStyle: { color: chartTextColor },
    title: {
      text: `${t('barProcess.config.maxGoal')}: ${maxValue} ${unit}`,
      right: '15%',
      top: 0,
      textStyle: {
        color: chartTextColor,
        fontSize: 13,
        fontWeight: 'bold',
        opacity: 0.8
      }
    },
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'none' },
      formatter: function (params) {
        const p = params.find(param => param.seriesName === 'Progress');
        if (!p) return '';
        const pct = Math.round((p.value / maxValue) * 100);
        return `<strong>${p.name}</strong><br/>Value: ${p.value}${unit} (${pct}%)`;
      }
    },
    grid: {
      left: '2%',
      right: '15%',
      top: 30,
      bottom: 10,
      containLabel: true
    },
    xAxis: { type: 'value', max: maxValue, show: false },
    yAxis: { type: 'category', data: deviceNames, show: false, inverse: true },
    series: [
      {
        name: 'Background',
        type: 'bar',
        itemStyle: { color: trackColor, borderRadius: borderRadius },
        silent: true,
        barWidth: barThickness,
        barGap: '-100%',
        barCategoryGap: '40%',
        data: backgroundValues,
        label: {
          show: true,
          position: ['0%', '-20px'],
          formatter: '{b}',
          color: chartTextColor,
          fontSize: 13,
          fontWeight: 'bold'
        }
      },
      {
        name: 'Progress',
        type: 'bar',
        itemStyle: {
          borderRadius: borderRadius,
          color: function (params) {
            const devId = String(rawDeviceIds[params.dataIndex]);
            return deviceColorsMap[devId] || fallbackColors[params.dataIndex % fallbackColors.length];
          }
        },
        barWidth: barThickness,
        z: 3,
        label: {
          show: showTextLabel,
          position: 'right',
          formatter: function (params) {
            const pct = Math.round((params.value / maxValue) * 100);
            return `{pctStyle|${pct}%}`;
          },
          rich: { pctStyle: { color: chartTextColor, fontWeight: 'bold', fontSize: 12 } }
        },
        data: deviceValues
      }
    ]
  };
});
</script>