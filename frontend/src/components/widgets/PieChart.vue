<template>
  <div class="flex flex-col h-full w-full p-4 overflow-hidden" :style="backgroundStyle">

    <div class="backdrop-blur-md px-4 py-3 shadow-sm z-10 flex justify-between items-center rounded-lg">
      <h3 class="m-0 text-base font-extrabold tracking-wide"
        :style="{ color: widgetData.widgetStyle?.textHex || '#334155' }">
        {{ widgetData?.widgetLabel }}
      </h3>
    </div>

    <div class="flex-1 w-full min-h-[150px] relative mt-2 flex flex-col justify-center">

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
import { PieChart } from 'echarts/charts';
import { TooltipComponent, LegendComponent } from 'echarts/components';

use([CanvasRenderer, PieChart, TooltipComponent, LegendComponent]);

const { t } = useI18n();

const props = defineProps({
  widgetData: { type: Object, default: () => ({}) }
});

const isReady = ref(false);
const liveStreamStore = useLiveStreamStore();
const widgetInstanceId = `piechart-${props.widgetData?.id || Math.random().toString(36).substring(2, 9)}`;

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

  const showLegend = customData.showLegend !== undefined ? customData.showLegend : true;
  const innerRadius = customData.innerRadius || '40%';
  const outerRadius = customData.outerRadius || '70%';
  const borderRadius = customData.borderRadius !== undefined ? customData.borderRadius : 5;
  const deviceColorsMap = customData.deviceColors || {};
  const fallbackColors = [
    '#3b82f6', '#10b981', '#f59e0b', '#ef4444',
    '#8b5cf6', '#06b6d4', '#84cc16', '#ec4899',
    '#f97316', '#14b8a6', '#6366f1', '#eab308',
    '#d946ef', '#0ea5e9', '#22c55e', '#f43f5e'
  ];
  const rawDeviceIds = props.widgetData?.deviceIds || [];
  const dynamicData = [];

  rawDeviceIds.forEach((rawId, index) => {
    const id = String(rawId);
    const dataObj = liveStreamStore.liveData[id];
    dynamicData.push({
      name: dataObj ? dataObj.name : t('common.loading'),
      value: (dataObj && dataObj.value !== null) ? Number(dataObj.value) : 0,
      itemStyle: { color: deviceColorsMap[id] || fallbackColors[index % fallbackColors.length] }
    });
  });

  return {
    textStyle: { color: chartTextColor },
    tooltip: { trigger: 'item' },
    legend: {
      type: 'scroll',
      show: showLegend,
      bottom: 5,
      left: 10,
      right: 70,
      textStyle: { color: chartTextColor }
    },
    series: [
      {
        name: 'Data',
        type: 'pie',
        radius: [innerRadius, outerRadius],
        itemStyle: { borderRadius: borderRadius, borderColor: '#fff', borderWidth: 2 },
        label: { show: false, position: 'center' },
        emphasis: { label: { show: true, fontSize: 16, fontWeight: 'bold' } },
        labelLine: { show: false },
        data: dynamicData
      }
    ]
  };
});
</script>