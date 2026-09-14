<template>
  <div class="flex flex-col gap-5">
    
    <!-- ECharts Colors for Actual & Target -->
    <div class="grid grid-cols-2 gap-4">
      <label class="form-control w-full">
        <div class="label pb-1"><span class="label-text font-bold">{{ $t('bulletChart.config.performanceBar') }}</span></div>
        <input type="color" v-model="localConfig.barColor"
          class="h-10 w-full cursor-pointer rounded border border-base-300 p-0" />
      </label>
      <label class="form-control w-full">
        <div class="label pb-1"><span class="label-text font-bold">{{ $t('bulletChart.config.targetMarker') }}</span></div>
        <input type="color" v-model="localConfig.targetColor"
          class="h-10 w-full cursor-pointer rounded border border-base-300 p-0" />
      </label>
    </div>

    <!-- Scale & Target Settings -->
    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">{{ $t('bulletChart.config.scaleTarget') }}</h4>
      
      <div class="grid grid-cols-2 gap-4 mb-3">
        <label class="form-control w-full">
          <div class="label pb-1"><span class="label-text font-semibold text-primary">{{ $t('bulletChart.config.targetGoal') }}</span></div>
          <input type="number" v-model.number="localConfig.targetValue"
            class="input input-bordered input-sm w-full border-primary" placeholder="85" />
        </label>
        
        <label class="form-control w-full">
          <div class="label pb-1">
            <span class="label-text font-semibold">{{ $t('bulletChart.config.xAxisMax') }}</span>
            <span class="label-text-alt ml-2 text-base-content/60">{{ $t('bulletChart.config.leaveEmptyAuto') }}</span>
          </div>
          <input type="number" v-model.number="localConfig.xAxisMax"
            class="input input-bordered input-sm w-full" :placeholder="$t('bulletChart.config.auto')" />
        </label>
      </div>

      <label class="form-control w-full">
        <div class="label pb-1">
          <span class="label-text font-semibold">{{ $t('common.unitLabel') }}</span>
        </div>
        <input type="text" v-model="localConfig.unit"
          class="input input-bordered input-sm w-full" placeholder="e.g., GB, °C, %" />
      </label>
    </div>

    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">{{ $t('common.displayOptions') }}</h4>
      <label class="cursor-pointer label justify-start gap-4">
        <input type="checkbox" v-model="localConfig.showDeviceName" class="toggle toggle-primary toggle-sm" />
        <span class="label-text font-semibold">{{ $t('common.showDeviceName') }}</span>
      </label>
    </div>

    <!-- Dynamic Threshold Zones -->
    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <div class="flex justify-between items-center mb-4">
        <h4 class="font-bold text-sm text-base-content m-0">{{ $t('bulletChart.config.bgThresholdZones') }}</h4>
        <button @click="addThreshold" class="btn btn-xs btn-primary btn-outline">
          {{ $t('common.addZone') }}
        </button>
      </div>
      
      <div class="flex flex-col gap-3">
        <div v-for="(zone, index) in localConfig.thresholds" :key="zone.id" 
             class="flex items-end gap-2 p-3 bg-base-100 border border-base-300 rounded-lg">
          
          <label class="form-control flex-1">
            <span class="label-text-alt mb-1 font-semibold">{{ $t('common.name') }}</span>
            <input type="text" v-model="zone.name" class="input input-bordered input-xs w-full" />
          </label>
          
          <label class="form-control w-24">
            <span class="label-text-alt mb-1 font-semibold">{{ $t('common.maxValue') }}</span>
            <input type="number" v-model.number="zone.value" class="input input-bordered input-xs w-full" />
          </label>
          
          <label class="form-control w-12">
            <span class="label-text-alt mb-1 font-semibold">{{ $t('common.color') }}</span>
            <input type="color" v-model="zone.color" class="h-6 w-full cursor-pointer rounded border border-base-300 p-0" />
          </label>

          <button @click="removeThreshold(index)" class="btn btn-xs btn-ghost text-error hover:bg-error/10 px-2" :title="$t('common.delete')">
            <Icon icon="lucide:x" class="w-5 h-5" />
          </button>
        </div>

        <div v-if="localConfig.thresholds.length === 0" class="text-center text-sm text-base-content/50 py-2">
          {{ $t('bulletChart.config.noZones') }}
        </div>
      </div>
      <p class="text-xs text-base-content/60 mt-3">
        {{ $t('bulletChart.config.echartsWarning') }}
      </p>
    </div>

  </div>
</template>

<script setup>
import { ref, watch, onMounted, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { Icon } from '@iconify/vue';

const { t } = useI18n();

const props = defineProps({
  modelValue: {
    type: Object,
    default: () => ({})
  }
});

const emit = defineEmits(['update:modelValue']);

const defaultThresholds = computed(() => [
  { id: Date.now() + 1, name: t('bulletChart.config.defaults.excellent'), value: 120, color: '#e2e8f0' },
  { id: Date.now() + 2, name: t('bulletChart.config.defaults.poor'), value: 60, color: '#94a3b8' }
]);

const localConfig = ref({
  showDeviceName: props.modelValue.showDeviceName !== undefined ? props.modelValue.showDeviceName : true,
  barColor: props.modelValue.barColor || '#1A5FB4',
  targetColor: props.modelValue.targetColor || '#26A269',
  targetValue: props.modelValue.targetValue !== undefined ? props.modelValue.targetValue : 85,
  xAxisMax: props.modelValue.xAxisMax !== undefined ? props.modelValue.xAxisMax : null,
  unit: props.modelValue.unit || '',
  thresholds: props.modelValue.thresholds ? JSON.parse(JSON.stringify(props.modelValue.thresholds)) : defaultThresholds.value
});

const addThreshold = () => {
  localConfig.value.thresholds.push({
    id: Date.now(),
    name: t('bulletChart.config.newZone'),
    value: 0,
    color: '#e5e7eb'
  });
};

const removeThreshold = (index) => {
  localConfig.value.thresholds.splice(index, 1);
};

watch(localConfig, (newVal) => {
  emit('update:modelValue', JSON.parse(JSON.stringify(newVal)));
}, { deep: true });

onMounted(() => {
  emit('update:modelValue', JSON.parse(JSON.stringify(localConfig.value)));
});
</script>