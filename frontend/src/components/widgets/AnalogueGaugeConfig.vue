<template>
  <div class="flex flex-col gap-5">
    
    <!-- Gauge Scale Configuration -->
    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">{{ $t('analogueGauge.config.scaleSettings') }}</h4>
      
      <div class="grid grid-cols-2 gap-4 mb-3">
        <label class="form-control w-full">
          <div class="label pb-1"><span class="label-text font-semibold">{{ $t('analogueGauge.config.minValue') }}</span></div>
          <input type="number" v-model.number="localConfig.min"
            class="input input-bordered input-sm w-full transition-colors"
            :class="{ 'input-error text-error font-bold': localConfig.min >= localConfig.max }"
            placeholder="0" />
        </label>
        
        <label class="form-control w-full">
          <div class="label pb-1"><span class="label-text font-semibold">{{ $t('common.maxValue') }}</span></div>
          <input type="number" v-model.number="localConfig.max"
            class="input input-bordered input-sm w-full transition-colors" 
            :class="{ 'input-error text-error font-bold': localConfig.min >= localConfig.max }"
            placeholder="100" />
        </label>
      </div>

      <label class="form-control w-full">
        <div class="label pb-1">
          <span class="label-text font-semibold">{{ $t('common.unitLabel') }}</span>
        </div>
        <input type="text" v-model="localConfig.unit"
          class="input input-bordered input-sm w-full" placeholder="%" />
      </label>
    </div>
    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">{{ $t('common.displayOptions') }}</h4>
      <label class="cursor-pointer label justify-start gap-4">
        <input type="checkbox" v-model="localConfig.showDeviceName" class="toggle toggle-primary toggle-sm" />
        <span class="label-text font-semibold">{{ $t('common.showDeviceName') }}</span>
      </label>
    </div>

    <!-- Toggle for Grade Zones -->
    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">{{ $t('analogueGauge.config.colorZones') }}</h4>
      <label class="cursor-pointer label justify-start gap-4">
        <input type="checkbox" v-model="localConfig.useGrades" class="toggle toggle-primary toggle-sm" />
        <span class="label-text font-semibold">{{ $t('analogueGauge.config.enableZones') }}</span>
      </label>
    </div>

    <!-- Grade Text Display -->
    <div v-if="localConfig.useGrades" class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">{{ $t('analogueGauge.config.gradeLabels') }}</h4>
      
      <div class="flex flex-col gap-3">
        <label class="cursor-pointer label justify-start gap-4">
          <input type="checkbox" v-model="localConfig.showGradeText" class="toggle toggle-primary toggle-sm" />
          <span class="label-text font-semibold">{{ $t('analogueGauge.config.showGradeText') }}</span>
        </label>
        
        <label class="form-control w-full">
          <div class="label pb-1"><span class="label-text font-semibold">{{ $t('analogueGauge.config.textSize') }}</span></div>
          <input type="number" v-model.number="localConfig.gradeTextSize"
            class="input input-bordered input-sm w-full" :disabled="!localConfig.showGradeText" placeholder="14" />
        </label>
      </div>
    </div>

    <!-- Dynamic Grade Zoning -->
    <div v-if="localConfig.useGrades" class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <div class="flex justify-between items-center mb-4">
        <h4 class="font-bold text-sm text-base-content m-0">{{ $t('analogueGauge.config.gradeZones') }}</h4>
        <button @click="addGrade" class="btn btn-primary btn-xs">
          {{ $t('common.addZone') }}
        </button>
      </div>

      <!-- Warnings -->
      <div v-if="hasMaxMismatch" class="alert alert-error shadow-sm mb-4 py-2 px-3">
        <Icon icon="lucide:circle-x" class="w-5 h-5"/>
        <span class="text-sm"><strong>Error:</strong> {{ $t('analogueGauge.config.errors.maxMismatch', { max: localConfig.max }) }}</span>
      </div>

      <div v-if="hasMinMismatch" class="alert alert-error shadow-sm mb-4 py-2 px-3">
        <Icon icon="lucide:circle-x" class="w-5 h-5"/>
        <span class="text-sm"><strong>Error:</strong> {{ $t('analogueGauge.config.errors.minMismatch', { min: localConfig.min }) }}</span>
      </div>

      <div v-if="localConfig.min >= localConfig.max" class="alert alert-error shadow-sm mb-4 py-2 px-3">
        <Icon icon="lucide:circle-x" class="w-5 h-5"/>
        <span class="text-sm"><strong>Error:</strong> {{ $t('analogueGauge.config.errors.minMaxInvalid') }}</span>
      </div>
      
      <div class="flex flex-col gap-3">
        <div v-for="(grade, index) in localConfig.grades" :key="index" class="flex flex-wrap md:flex-nowrap items-center gap-2 bg-base-100 p-2 rounded border border-base-200">
          
          <input type="color" v-model="grade.color" class="h-8 w-10 cursor-pointer rounded border border-base-300 p-0 shrink-0" />
          
          <input type="text" v-model="grade.name" class="input input-bordered input-sm w-full md:w-1/2" :placeholder="$t('analogueGauge.config.gradeName')" />
          
          <div class="flex items-center gap-2 w-full md:w-auto">
            <span class="text-xs font-semibold whitespace-nowrap">{{ $t('analogueGauge.config.upTo') }}</span>
            <input type="number" v-model.number="grade.limit" 
              class="input input-bordered input-sm w-full md:w-24 transition-colors" 
              :class="{ 'input-error text-error font-bold': grade.limit <= localConfig.min || grade.limit > localConfig.max }" />
          </div>

          <button @click="removeGrade(index)" class="btn btn-ghost btn-sm btn-square text-error shrink-0" title="Remove Grade">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        <div v-if="localConfig.grades.length === 0" class="text-xs text-base-content/50 italic text-center py-2">
          {{ $t('analogueGauge.config.noZones') }}
        </div>
      </div>
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

// Using computed to allow dynamic translations if the locale changes
const defaultGrades = computed(() => [
  { name: 'Grade D', limit: 25, color: '#FF6E76' },
  { name: 'Garde C', limit: 50, color: '#FDDD60' },
  { name: 'Garde B', limit: 75, color: '#58D9F9' },
  { name: 'Garde A', limit: 100, color: '#7CFFB2' }
]);

const localConfig = ref({
  showDeviceName: props.modelValue.showDeviceName !== undefined ? props.modelValue.showDeviceName : true,
  min: props.modelValue.min !== undefined ? props.modelValue.min : 0,
  max: props.modelValue.max !== undefined ? props.modelValue.max : 100,
  unit: props.modelValue.unit || '',
  useGrades: props.modelValue.useGrades !== undefined ? props.modelValue.useGrades : false, 
  showGradeText: props.modelValue.showGradeText !== undefined ? props.modelValue.showGradeText : true,
  gradeTextSize: props.modelValue.gradeTextSize !== undefined ? props.modelValue.gradeTextSize : 14,
  grades: props.modelValue.grades ? JSON.parse(JSON.stringify(props.modelValue.grades)) : defaultGrades.value
});

const highestLimit = computed(() => {
  if (localConfig.value.grades.length === 0) return null;
  return Math.max(...localConfig.value.grades.map(g => g.limit));
});

const hasMaxMismatch = computed(() => {
  return localConfig.value.useGrades && localConfig.value.grades.length > 0 && highestLimit.value !== localConfig.value.max;
});

const hasMinMismatch = computed(() => {
  return localConfig.value.useGrades && localConfig.value.grades.some(grade => grade.limit <= localConfig.value.min);
});

const isInvalid = computed(() => {
  if (localConfig.value.min >= localConfig.value.max) return true;
  if (!localConfig.value.useGrades) return false; 
  return hasMaxMismatch.value || hasMinMismatch.value || localConfig.value.grades.length === 0;
});

const addGrade = () => {
  localConfig.value.grades.push({ 
    name: t('analogueGauge.config.gradeName'), 
    limit: localConfig.value.max, 
    color: '#3b82f6' 
  });
};

const removeGrade = (index) => {
  localConfig.value.grades.splice(index, 1);
};

watch(localConfig, (newVal) => {
  emit('update:modelValue', { ...newVal, _isInvalid: isInvalid.value });
}, { deep: true });

onMounted(() => {
  emit('update:modelValue', { ...localConfig.value, _isInvalid: isInvalid.value });
});
</script>