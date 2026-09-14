<template>
  <div class="flex flex-col gap-5">

    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">{{ $t('scoreCard.config.layoutVisuals') }}</h4>

      <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">

        <label class="form-control w-full">
          <div class="label pb-1"><span class="label-text font-semibold">{{ $t('scoreCard.config.leftVisual') }}</span></div>
          <select v-model="localConfig.visualType" class="select select-bordered select-sm w-full">
            <option value="line">{{ $t('scoreCard.config.visuals.line') }}</option>
            <option value="bar">{{ $t('scoreCard.config.visuals.bar') }}</option>
            <option value="icon">{{ $t('scoreCard.config.visuals.icon') }}</option>
            <option value="none">{{ $t('scoreCard.config.visuals.none') }}</option>
          </select>
        </label>

        <label v-if="localConfig.visualType === 'icon'" class="form-control w-full">
          <div class="label pb-1"><span class="label-text font-semibold">{{ $t('scoreCard.config.selectIcon') }}</span></div>
          <select v-model="localConfig.icon" class="select select-bordered select-sm w-full">
            <option value="lucide:activity">{{ $t('scoreCard.config.icons.activity') }}</option>
            <option value="lucide:trending-up">{{ $t('scoreCard.config.icons.trendingUp') }}</option>
            <option value="lucide:trending-down">{{ $t('scoreCard.config.icons.trendingDown') }}</option>
            <option value="lucide:dollar-sign">{{ $t('scoreCard.config.icons.dollar') }}</option>
            <option value="lucide:users">{{ $t('scoreCard.config.icons.users') }}</option>
            <option value="lucide:box">{{ $t('scoreCard.config.icons.box') }}</option>
            <option value="lucide:cpu">{{ $t('scoreCard.config.icons.cpu') }}</option>
            <option value="lucide:zap">{{ $t('scoreCard.config.icons.zap') }}</option>
            <option value="lucide:thermometer">{{ $t('scoreCard.config.icons.thermo') }}</option>
            <option value="lucide:droplets">{{ $t('scoreCard.config.icons.droplets') }}</option>
          </select>
        </label>

        <label class="form-control w-full sm:col-span-2">
          <div class="label pb-1">
            <span class="label-text font-semibold">{{ $t('scoreCard.config.customSubtext') }}</span>
            <span class="label-text-alt text-base-content/60">{{ $t('scoreCard.config.customSubtextHint') }}</span>
          </div>
          <input type="text" v-model="localConfig.customSubText" class="input input-bordered input-sm w-full"
            :placeholder="$t('scoreCard.config.customSubtextPlaceholder')" />
        </label>

      </div>
    </div>

    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">{{ $t('scoreCard.config.dataAggregation') }}</h4>

      <label class="form-control w-full">
        <div class="label pb-1"><span class="label-text font-semibold">{{ $t('scoreCard.config.calcStrategy') }}</span></div>
        <select v-model="localConfig.aggregationMode"
          class="select select-bordered select-sm w-full font-semibold text-primary">
          <optgroup :label="$t('scoreCard.config.liveGroup')">
            <option value="live_single">{{ $t('scoreCard.config.liveModes.single') }}</option>
            <option value="live_sum">{{ $t('scoreCard.config.liveModes.sum') }}</option>
            <option value="live_avg">{{ $t('scoreCard.config.liveModes.avg') }}</option>
            <option value="live_count">{{ $t('scoreCard.config.liveModes.count') }}</option>
          </optgroup>
          <optgroup :label="$t('scoreCard.config.historyGroup')">
            <option value="history_avg">{{ $t('scoreCard.config.historyModes.avg') }}</option>
            <option value="history_min">{{ $t('scoreCard.config.historyModes.min') }}</option>
            <option value="history_max">{{ $t('scoreCard.config.historyModes.max') }}</option>
          </optgroup>
        </select>
      </label>
    </div>

    <div v-if="localConfig.aggregationMode?.startsWith('history_')"
      class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">{{ $t('scoreCard.config.historyRangeSettings') }}</h4>

      <div class="grid grid-cols-2 gap-4">
        <label class="form-control w-full">
          <div class="label pb-1"><span class="label-text font-semibold">{{ $t('common.timeRange') }}</span></div>
          <select v-model="localConfig.historyRange" class="select select-bordered select-sm w-full">
            <option value="15m">{{ $t('common.timeRanges.m15') }}</option>
            <option value="1h">{{ $t('common.timeRanges.h1') }}</option>
            <option value="24h">{{ $t('common.timeRanges.h24') }}</option>
            <option value="7d">{{ $t('common.timeRanges.d7') }}</option>
            <option value="30d">{{ $t('common.timeRanges.d30') }}</option>
            <option value="3m">{{ $t('common.timeRanges.M3') }}</option>
            <option value="6m">{{ $t('common.timeRanges.M6') }}</option>
            <option value="1y">{{ $t('common.timeRanges.y1') }}</option>
          </select>
        </label>
        <label class="form-control w-full">
          <div class="label pb-1"><span class="label-text font-semibold">{{ $t('common.maxDataPoints') }}</span></div>
          <input type="number" v-model.number="localConfig.maxPoints" class="input input-bordered input-sm w-full" />
        </label>
      </div>
    </div>

    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">{{ $t('scoreCard.config.valueFormatting') }}</h4>

      <div class="grid grid-cols-2 gap-4">
        <label class="form-control w-full">
          <div class="label pb-1"><span class="label-text font-semibold">{{ $t('common.prefix') }}</span></div>
          <input type="text" v-model="localConfig.prefix" class="input input-bordered input-sm w-full"
            :placeholder="$t('common.prefixPlaceholder')" />
        </label>

        <label class="form-control w-full">
          <div class="label pb-1"><span class="label-text font-semibold">{{ $t('common.unitSuffix') }}</span></div>
          <input type="text" v-model="localConfig.unit" class="input input-bordered input-sm w-full"
            :placeholder="$t('common.unitPlaceholder')" />
        </label>

        <label class="form-control w-full sm:col-span-2">
          <div class="label pb-1">
            <span class="label-text font-semibold">{{ $t('common.decimalPlaces') }}</span>
            <span class="label-text-alt text-base-content/60">{{ $t('scoreCard.config.decimalHint') }}</span>
          </div>
          <input type="number" v-model="localConfig.decimalPlaces" min="0" max="4"
            class="input input-bordered input-sm w-full" />
        </label>
      </div>
    </div>

  </div>
</template>

<script setup>
import { ref, watch, onMounted } from 'vue';

const props = defineProps({
  modelValue: { type: Object, default: () => ({}) },
  selectedDeviceIds: { type: Array, default: () => [] },
  allDevices: { type: Array, default: () => [] }
});

const emit = defineEmits(['update:modelValue']);

const localConfig = ref({
  visualType: props.modelValue.visualType || 'line',
  icon: props.modelValue.icon || 'lucide:activity',
  customSubText: props.modelValue.customSubText || '',
  aggregationMode: props.modelValue.aggregationMode || 'live_single',
  historyRange: props.modelValue.historyRange || '1h',
  maxPoints: props.modelValue.maxPoints || 1000,
  prefix: props.modelValue.prefix || '',
  unit: props.modelValue.unit || '',
  decimalPlaces: props.modelValue.decimalPlaces !== undefined ? props.modelValue.decimalPlaces : 0
});

watch(localConfig, (newVal) => {
  emit('update:modelValue', JSON.parse(JSON.stringify(newVal)));
}, { deep: true });

onMounted(() => {
  emit('update:modelValue', JSON.parse(JSON.stringify(localConfig.value)));
});
</script>