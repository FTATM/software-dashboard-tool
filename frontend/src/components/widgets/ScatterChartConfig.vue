<template>
  <div class="flex flex-col gap-5">

    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">{{ $t('lineChart.config.dataOptions') }}</h4>

      <div class="grid grid-cols-2 gap-4 mb-3">
        <label class="form-control w-full">
          <div class="label pb-1"><span class="label-text font-semibold">{{ $t('common.historyRange') }}</span></div>
          <select v-model="localConfig.historyRange" class="select select-bordered select-sm w-full">
            <option value="0">{{ $t('common.timeRanges.live') }}</option>
            <option value="15m">{{ $t('common.timeRanges.m15') }}</option>
            <option value="30m">{{ $t('common.timeRanges.m30') }}</option>
            <option value="1h">{{ $t('common.timeRanges.h1') }}</option>
            <option value="3h">{{ $t('common.timeRanges.h3') }}</option>
            <option value="6h">{{ $t('common.timeRanges.h6') }}</option>
            <option value="24h">{{ $t('common.timeRanges.h24') }}</option>
            <option value="7d">{{ $t('common.timeRanges.d7') }}</option>
            <option value="custom">{{ $t('common.timeRanges.custom') }}</option>
          </select>
        </label>

        <label class="form-control w-full">
          <div class="label pb-1"><span class="label-text font-semibold">{{ $t('common.maxDataPoints') }}</span></div>
          <input type="number" v-model.number="localConfig.maxPoints" class="input input-bordered input-sm w-full"
            placeholder="100" />
        </label>
      </div>

      <div v-if="localConfig.historyRange === 'custom'"
        class="grid grid-cols-2 gap-4 p-3 mt-2 bg-base-100 rounded-lg border border-base-300">
        <label class="form-control w-full">
          <div class="label pb-1"><span class="label-text font-semibold text-primary">{{ $t('common.from') }}</span>
          </div>
          <VueDatePicker v-model="localConfig.customFrom" :is-24="true" auto-apply :preset-dates="presetDates"
            :locale="dateFnsLocale" :format-locale="dateFnsLocale" :dark="themeStore.isDarkTheme"
            :formats="{ input: 'dd/MM/yyyy HH:mm' }" :action-row="{
              selectBtnLabel: $t('common.select'),
              cancelBtnLabel: $t('common.cancel')
            }" teleport-center>
            <template #input-icon>
              <Icon icon="lucide:calendar-clock" class="w-5 h-5 ml-3 text-base-content/50" />
            </template>
          </VueDatePicker>
        </label>

        <label class="form-control w-full">
          <div class="label pb-1"><span class="label-text font-semibold text-primary">{{ $t('common.to') }}</span></div>
          <VueDatePicker v-model="localConfig.customTo" :min-date="localConfig.customFrom" :is-24="true" auto-apply
            :preset-dates="presetDates" :locale="dateFnsLocale" :format-locale="dateFnsLocale"
            :dark="themeStore.isDarkTheme" :formats="{ input: 'dd/MM/yyyy HH:mm' }" :action-row="{
              selectBtnLabel: $t('common.select'),
              cancelBtnLabel: $t('common.cancel')
            }" teleport-center>
            <template #input-icon>
              <Icon icon="lucide:calendar-clock" class="w-5 h-5 ml-3 text-base-content/50" />
            </template>
          </VueDatePicker>
        </label>
      </div>
    </div>

    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">{{ $t('scatterChart.config.regressionAnalysis') }}</h4>

      <div class="flex flex-col gap-3 mb-4">
        <label class="cursor-pointer label justify-start gap-4">
          <input type="checkbox" v-model="localConfig.showRegression" class="toggle toggle-primary" />
          <span class="label-text font-semibold">{{ $t('scatterChart.config.showRegression') }}</span>
        </label>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mt-2">
        <label class="form-control w-full">
          <div class="label pb-1">
            <span class="label-text font-semibold">{{ $t('scatterChart.config.xAxisLabel') }}</span>
          </div>
          <input type="text" v-model="localConfig.xAxisName" class="input input-bordered input-sm w-full"
            :placeholder="$t('scatterChart.config.timePlaceholder')" />
        </label>

        <label class="form-control w-full">
          <div class="label pb-1">
            <span class="label-text font-semibold">{{ $t('scatterChart.config.yAxisLabel') }}</span>
          </div>
          <input type="text" v-model="localConfig.yAxisName" class="input input-bordered input-sm w-full"
            :placeholder="$t('scatterChart.config.valuePlaceholder')" />
        </label>
      </div>
    </div>

    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <div class="flex justify-between items-center mb-4">
        <h4 class="font-bold text-sm text-base-content m-0">{{ $t('common.deviceColors') }}</h4>
      </div>

      <div class="flex flex-col gap-2">
        <div v-for="device in activeDevices" :key="device.deviceId"
          class="flex items-center gap-3 p-2 bg-base-100 border border-base-300 rounded-lg">
          <input type="color" v-model="localConfig.deviceColors[device.deviceId]"
            class="h-8 w-12 cursor-pointer rounded border border-base-300 p-0 shrink-0" />
          <span class="text-sm font-semibold flex-1">{{ device.deviceName }}</span>
        </div>
        <div v-if="activeDevices.length === 0" class="text-sm text-base-content/50 py-2 text-center">
          {{ $t('common.noDevice') }}
        </div>
      </div>
    </div>

  </div>
</template>

<script setup>
import { ref, watch, onMounted, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { Icon } from '@iconify/vue';
import { VueDatePicker } from '@vuepic/vue-datepicker';
import { useThemeStore } from '@/stores/useThemeStore';
import thLocale from 'date-fns/locale/th';
import enLocale from 'date-fns/locale/en-US';

const themeStore = useThemeStore();
const { t, locale } = useI18n();

const resolveLocale = (loc) => {
  if (!loc) return undefined;
  if (loc.localize) return loc;
  if (loc.default?.localize) return loc.default;
  if (loc.th?.localize) return loc.th;
  if (loc.enUS?.localize) return loc.enUS;
  return loc.default || loc;
};

const dateFnsLocale = computed(() => {
  const isThai = locale.value?.startsWith('th');
  return resolveLocale(isThai ? thLocale : enLocale);
});

const props = defineProps({
  modelValue: { type: Object, default: () => ({}) },
  selectedDeviceIds: { type: Array, default: () => [] },
  allDevices: { type: Array, default: () => [] }
});

const emit = defineEmits(['update:modelValue']);

const presetDates = ref([{ label: t('common.today'), value: new Date() }]);

const activeDevices = computed(() => {
  if (!props.selectedDeviceIds) return [];
  return props.selectedDeviceIds
    .map(id => props.allDevices.find(device => device.deviceId === id))
    .filter(Boolean);
});

const defaultColorPalette = [
  '#3b82f6', // 1. Blue
  '#10b981', // 2. Emerald
  '#f59e0b', // 3. Amber
  '#ef4444', // 4. Red
  '#8b5cf6', // 5. Violet
  '#06b6d4', // 6. Cyan
  '#84cc16', // 7. Lime
  '#ec4899', // 8. Pink
  '#f97316', // 9. Orange
  '#14b8a6', // 10. Teal
  '#6366f1', // 11. Indigo
  '#eab308', // 12. Yellow
  '#d946ef', // 13. Fuchsia
  '#0ea5e9', // 14. Light Blue
  '#22c55e', // 15. Green
  '#f43f5e'  // 16. Rose
];

const localConfig = ref({
  historyRange: props.modelValue.historyRange || '1h',
  customFrom: props.modelValue.customFrom || '',
  customTo: props.modelValue.customTo || '',
  maxPoints: props.modelValue.maxPoints || 1000,
  showRegression: props.modelValue.showRegression !== undefined ? props.modelValue.showRegression : true,
  xAxisName: props.modelValue.xAxisName || '',
  yAxisName: props.modelValue.yAxisName || '',
  deviceColors: props.modelValue.deviceColors || {}
});

watch(() => props.selectedDeviceIds, (newIds) => {
  newIds.forEach((id, index) => {
    if (!localConfig.value.deviceColors[id]) {
      localConfig.value.deviceColors[id] = defaultColorPalette[index % defaultColorPalette.length];
    }
  });
}, { immediate: true, deep: true });

watch(localConfig, (newVal) => {
  emit('update:modelValue', JSON.parse(JSON.stringify(newVal)));
}, { deep: true });

onMounted(() => {
  emit('update:modelValue', JSON.parse(JSON.stringify(localConfig.value)));
});
</script>