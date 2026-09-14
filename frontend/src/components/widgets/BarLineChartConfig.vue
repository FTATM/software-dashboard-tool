<template>
  <div class="flex flex-col gap-5">

    <!-- Data Grouping & History -->
    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">{{ $t('barLineChart.config.dataGrouping') }}</h4>

      <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-3">
        <label class="form-control w-full">
          <div class="label pb-1"><span class="label-text font-semibold">{{ $t('common.timeRange') }}</span></div>
          <select v-model="localConfig.historyRange" class="select select-bordered select-sm w-full">
            <option value="24h">{{ $t('common.timeRanges.h24') }}</option>
            <option value="7d">{{ $t('common.timeRanges.d7') }}</option>
            <option value="30d">{{ $t('common.timeRanges.d30') }}</option>
            <option value="3m">{{ $t('common.timeRanges.M3') }}</option>
            <option value="6m">{{ $t('common.timeRanges.M6') }}</option>
            <option value="1y">{{ $t('common.timeRanges.y1') }}</option>
          </select>
        </label>

        <label class="form-control w-full">
          <div class="label pb-1"><span class="label-text font-semibold">{{ $t('barLineChart.config.bucketInterval')
          }}</span></div>
          <select v-model="localConfig.bucketInterval"
            class="select select-bordered select-sm w-full font-bold text-primary">

            <option value="hour" :disabled="['30d', '3m', '6m', '1y'].includes(localConfig.historyRange)">
              {{ $t('barLineChart.config.intervals.hour') }}
            </option>

            <option value="day">{{ $t('barLineChart.config.intervals.day') }}</option>
            <option value="month">{{ $t('barLineChart.config.intervals.month') }}</option>
          </select>
        </label>

        <label class="form-control w-full">
          <div class="label pb-1"><span class="label-text font-semibold">{{ $t('barLineChart.config.aggregation')
          }}</span></div>
          <select v-model="localConfig.aggregationMode"
            class="select select-bordered select-sm w-full font-bold text-secondary">
            <option value="sum">{{ $t('barLineChart.config.aggregations.sum') }}</option>
            <option value="avg">{{ $t('barLineChart.config.aggregations.avg') }}</option>
            <option value="max">{{ $t('barLineChart.config.aggregations.max') }}</option>
            <option value="min">{{ $t('barLineChart.config.aggregations.min') }}</option>
          </select>
        </label>
      </div>
    </div>

    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">{{ $t('common.displayOptions') }}</h4>
      <label class="cursor-pointer label justify-start gap-4">
        <input type="checkbox" v-model="localConfig.showDeviceName" class="toggle toggle-primary toggle-sm" />
        <span class="label-text font-semibold">{{ $t('common.showDeviceName') }}</span>
      </label>
    </div>

    <!-- Chart Appearance -->
    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">{{ $t('barLineChart.config.visualsTitle') }}</h4>

      <div class="grid grid-cols-2 gap-4 mb-4">
        <label class="form-control w-full">
          <div class="label pb-1"><span class="label-text font-semibold">{{ $t('barLineChart.config.barColor') }}</span>
          </div>
          <input type="color" v-model="localConfig.barColor"
            class="h-10 w-full cursor-pointer rounded border border-base-300 p-0" />
        </label>

        <label class="form-control w-full">
          <div class="label pb-1"><span class="label-text font-semibold">{{ $t('barLineChart.config.lineColor')
          }}</span></div>
          <input type="color" v-model="localConfig.lineColor"
            class="h-10 w-full cursor-pointer rounded border border-base-300 p-0" />
        </label>
      </div>

      <div class="grid grid-cols-2 gap-4 mb-4">
        <label class="form-control w-full">
          <div class="label pb-1"><span class="label-text font-semibold">{{ $t('barLineChart.config.barLegendName')
          }}</span></div>
          <input type="text" v-model="localConfig.barName" class="input input-bordered input-sm w-full"
            :placeholder="$t('barLineChart.config.barLegendPlaceholder')" />
        </label>
      </div>

      <div class="flex flex-col gap-4">
        <label class="cursor-pointer label justify-start gap-4 w-max">
          <input type="checkbox" v-model="localConfig.isSmooth" class="toggle toggle-primary toggle-sm" />
          <span class="label-text font-semibold">{{ $t('common.smoothCurve') }}</span>
        </label>

        <label class="form-control w-full">
          <div class="label pb-1">
            <span class="label-text font-semibold">{{ $t('barLineChart.config.yAxisLabel') }}</span>
          </div>
          <input type="text" v-model="localConfig.yAxisName" class="input input-bordered input-sm w-full"
            :placeholder="$t('barLineChart.config.yAxisPlaceholder')" />
        </label>
      </div>
    </div>

  </div>
</template>

<script setup>
import { ref, watch, onMounted } from 'vue';

const props = defineProps({
  modelValue: { type: Object, default: () => ({}) }
});

const emit = defineEmits(['update:modelValue']);

const localConfig = ref({
  showDeviceName: props.modelValue.showDeviceName !== undefined ? props.modelValue.showDeviceName : true,
  historyRange: props.modelValue.historyRange || '7d',
  bucketInterval: props.modelValue.bucketInterval || 'day',
  aggregationMode: props.modelValue.aggregationMode || 'sum',

  barColor: props.modelValue.barColor || '#4472c4',
  lineColor: props.modelValue.lineColor || '#ed7d31',

  barName: props.modelValue.barName || 'Units',
  lineName: props.modelValue.lineName || 'Average Trend',

  yAxisName: props.modelValue.yAxisName || '',
  isSmooth: props.modelValue.isSmooth !== undefined ? props.modelValue.isSmooth : false
});

watch(localConfig, (newVal) => {
  // ⚡ NEW: Auto-correct the interval if they pick a long range while 'hour' is selected
  const longRanges = ['30d', '3m', '6m', '1y'];
  if (longRanges.includes(newVal.historyRange) && newVal.bucketInterval === 'hour') {
    newVal.bucketInterval = 'day'; // Automatically fall back to daily buckets
  }

  emit('update:modelValue', JSON.parse(JSON.stringify(newVal)));
}, { deep: true });

onMounted(() => {
  emit('update:modelValue', JSON.parse(JSON.stringify(localConfig.value)));
});
</script>