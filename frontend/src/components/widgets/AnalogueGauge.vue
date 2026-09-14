<template>
  <div class="flex flex-col h-full w-full p-4 overflow-hidden" :style="backgroundStyle">

    <div class="backdrop-blur-md px-4 py-3 shadow-sm z-10 flex justify-between items-center rounded-lg">
      <h3 class="m-0 text-base font-extrabold tracking-wide truncate pr-2"
        :style="{ color: widgetData.widgetStyle?.textHex || '#334155' }">
        {{ widgetData?.widgetLabel }}
      </h3>

      <span v-if="hasData && liveDeviceName && showDeviceName"
        class="badge badge-neutral badge-lg py-4 px-4 text-sm font-bold shrink-0 shadow-sm">
        {{ liveDeviceName }}
      </span>
    </div>

    <div class="flex-1 w-full min-h-0 relative mt-4 flex flex-col justify-center">

      <div v-if="!hasDevices" class="absolute inset-0 flex items-center justify-center text-sm italic text-center p-4"
        :style="{ color: widgetData.widgetStyle?.textHex || '#64748b' }">
        {{ $t('common.noDevice') }}
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
import { use } from 'echarts/core';
import { CanvasRenderer } from 'echarts/renderers';
import { GaugeChart } from 'echarts/charts';
import { TooltipComponent } from 'echarts/components';
import { useLiveStreamStore } from '@/stores/useLiveStreamStore';

use([CanvasRenderer, GaugeChart, TooltipComponent]);

const { t } = useI18n();

const props = defineProps({
  widgetData: { type: Object, default: () => ({}) }
});

const isReady = ref(false);
const liveStreamStore = useLiveStreamStore();
const widgetInstanceId = `gauge-${props.widgetData?.id || Math.random().toString(36).substring(2, 9)}`;

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

  const min = customData.min !== undefined ? customData.min : 0;
  const max = customData.max !== undefined ? customData.max : 100;
  const unit = customData.unit || '';
  const useGrades = customData.useGrades || false;
  const showGradeText = useGrades ? (customData.showGradeText !== undefined ? customData.showGradeText : true) : false;
  const gradeTextSize = customData.gradeTextSize !== undefined ? customData.gradeTextSize : 14;

  let colorStops = [[1, '#3b82f6']];
  let sortedGrades = [];

  const splitNum = 5;
  const tickValues = [];
  for (let i = 0; i <= splitNum; i++) {
    tickValues.push(min + (max - min) * (i / splitNum));
  }
  const tickLabels = {};

  if (useGrades) {
    const rawGrades = customData.grades || [];
    sortedGrades = [...rawGrades].sort((a, b) => a.limit - b.limit);
    const range = max - min;

    colorStops = sortedGrades.map(grade => {
      const clampedLimit = Math.min(grade.limit, max);
      let pct = (clampedLimit - min) / range;
      if (pct > 1) pct = 1;
      if (pct < 0) pct = 0;
      return [pct, grade.color];
    });
    if (colorStops.length === 0) colorStops.push([1, '#e2e8f0']);

    if (showGradeText) {
      sortedGrades.forEach((grade, idx) => {
        const prevLimit = idx === 0 ? min : sortedGrades[idx - 1].limit;
        const currentLimit = Math.min(grade.limit, max);
        const midPoint = prevLimit + (currentLimit - prevLimit) / 2;

        const closestTick = tickValues.reduce((prev, curr) =>
          Math.abs(curr - midPoint) < Math.abs(prev - midPoint) ? curr : prev
        );
        tickLabels[closestTick] = grade.name;
      });
    }
  }

  return {
    textStyle: { color: chartTextColor },
    series: [
      {
        type: 'gauge',
        startAngle: 225,
        endAngle: -45,
        center: ['50%', '55%'],
        radius: '90%',
        min: min,
        max: max,
        splitNumber: splitNum,
        progress: {
          show: !useGrades,
          width: 16
        },
        axisLine: {
          lineStyle: { width: 16, color: colorStops }
        },
        pointer: {
          itemStyle: { color: 'auto' }
        },
        axisTick: { show: false },
        splitLine: {
          length: 12,
          lineStyle: { color: chartTextColor, width: 2, opacity: 0.4 }
        },
        axisLabel: {
          distance: 22,
          show: true,
          color: chartTextColor,
          fontSize: gradeTextSize,
          formatter: function (value) {
            if (useGrades && showGradeText) {
              const match = Object.keys(tickLabels).find(t => Math.abs(Number(t) - value) < 0.001);
              return match ? tickLabels[match] : '';
            }
            return Math.round(value);
          }
        },
        anchor: {
          show: true,
          showAbove: true,
          size: 20,
          itemStyle: {
            borderWidth: 6,
            borderColor: chartTextColor
          }
        },
        title: { show: false },
        detail: {
          fontSize: 23,
          offsetCenter: [0, '80%'],
          valueAnimation: true,
          color: chartTextColor,
          formatter: function (value) {
            return (Math.round(value * 10) / 10) + `\n{unitText| ${unit}}`;
          },
          rich: {
            unitText: {
              fontSize: 18,
              color: chartTextColor,
              opacity: 0.7,
              padding: [0, 0, 0, 0]
            },
          }
        },
        data: [{ value: liveValue.value, name: '' }]
      }
    ]
  };
});
</script>