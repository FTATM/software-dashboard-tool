<template>
  <div class="flex flex-col gap-5">
    
    <div class="grid grid-cols-1 gap-4">
      <label class="form-control w-full">
        <div class="label pb-1"><span class="label-text font-bold">{{ $t('statusWidget.config.defaultValueColor') }}</span></div>
        <input type="color" v-model="localConfig.valueColor"
          class="h-10 w-full cursor-pointer rounded border border-base-300 p-0 shadow-sm" />
      </label>
    </div>

    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">{{ $t('statusWidget.config.textFormatting') }}</h4>
      
      <div class="grid grid-cols-2 gap-4 mb-4">
        <label class="form-control w-full">
          <div class="label pb-1">
            <span class="label-text font-semibold">{{ $t('common.prefix') }}</span>
          </div>
          <div class="flex items-center gap-2">
            <input type="text" v-model="localConfig.prefix"
              class="input input-bordered input-sm w-full" :placeholder="$t('common.prefixPlaceholder')" />
            <input type="color" v-model="localConfig.prefixColor" 
              class="h-8 w-10 shrink-0 cursor-pointer rounded border border-base-300 p-0 shadow-sm" :title="$t('common.prefix')" />
          </div>
        </label>
        
        <label class="form-control w-full">
          <div class="label pb-1">
            <span class="label-text font-semibold">{{ $t('common.unitSuffix') }}</span>
          </div>
          <div class="flex items-center gap-2">
            <input type="text" v-model="localConfig.unit"
              class="input input-bordered input-sm w-full" :placeholder="$t('common.unitPlaceholder')" />
            <input type="color" v-model="localConfig.unitColor" 
              class="h-8 w-10 shrink-0 cursor-pointer rounded border border-base-300 p-0 shadow-sm" :title="$t('common.unitSuffix')" />
          </div>
        </label>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <label class="form-control w-full">
          <div class="label pb-1">
            <span class="label-text font-semibold">{{ $t('common.fontSize') }}</span>
          </div>
          <input type="number" v-model.number="localConfig.fontSize"
            class="input input-bordered input-sm w-full" placeholder="56" />
        </label>

        <label class="form-control w-full">
          <div class="label pb-1 flex justify-between items-center">
            <span class="label-text font-semibold">{{ $t('common.subtextDescription') }}</span>
            <input type="color" v-model="localConfig.subtextColor" 
              class="h-5 w-8 cursor-pointer rounded border border-base-300 p-0 shadow-sm" :title="$t('common.subtextDescription')" />
          </div>
          <textarea v-model="localConfig.subtext"
            class="textarea textarea-bordered w-full h-20" :placeholder="$t('common.subtextPlaceholder')"></textarea>
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

    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <div class="flex justify-between items-center mb-4">
        <h4 class="font-bold text-sm text-base-content m-0">{{ $t('statusWidget.config.conditionalDisplay') }}</h4>
        <input type="checkbox" v-model="localConfig.enableConditions" class="toggle toggle-primary toggle-sm" />
      </div>

      <div v-if="localConfig.enableConditions" class="flex flex-col gap-3">
        <div class="flex justify-end">
          <button @click="addCondition" class="btn btn-xs btn-primary btn-outline">
            {{ $t('statusWidget.config.addCondition') }}
          </button>
        </div>

        <div v-for="(cond, index) in localConfig.conditions" :key="index" 
             class="flex flex-wrap items-center gap-2 p-3 bg-base-100 border border-base-300 rounded-lg shadow-sm">
          
          <div class="flex items-center gap-2 w-full md:w-auto">
            <span class="text-xs font-semibold whitespace-nowrap">{{ $t('statusWidget.config.ifValue') }}</span>
            <select v-model="cond.operator" class="select select-bordered select-xs w-16">
              <option value="==">==</option>
              <option value="!=">!=</option>
              <option value=">">&gt;</option>
              <option value=">=">&gt;=</option>
              <option value="<">&lt;</option>
              <option value="<=">&lt;=</option>
            </select>
            <input type="text" v-model="cond.compareValue" class="input input-bordered input-xs w-20" placeholder="Value" />
          </div>

          <div class="flex items-center gap-2 w-full md:w-auto flex-1">
            <span class="text-xs font-semibold whitespace-nowrap">{{ $t('common.show') }}</span>
            <input type="text" v-model="cond.displayText" class="input input-bordered input-xs flex-1" :placeholder="$t('statusWidget.config.textToDisplay')" />
            <input type="color" v-model="cond.color" class="h-6 w-8 cursor-pointer rounded border border-base-300 p-0" :title="$t('statusWidget.config.textColorIfMet')" />
          </div>

          <button @click="removeCondition(index)" class="btn btn-xs btn-ghost text-error hover:bg-error/10 px-2" :title="$t('common.delete')">
            <Icon icon="lucide:x" class="w-4 h-4" />
          </button>
        </div>

        <div v-if="localConfig.conditions.length === 0" class="text-xs text-base-content/50 italic text-center py-2">
          {{ $t('statusWidget.config.noConditions') }}
        </div>
        <p class="text-[10px] text-base-content/60 mt-1 leading-tight">
          {{ $t('statusWidget.config.evaluationRule') }}
        </p>
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

const defaultConditions = computed(() => [
  { operator: '==', compareValue: '1', displayText: t('statusWidget.config.defaults.online'), color: '#10b981' },
  { operator: '==', compareValue: '0', displayText: t('statusWidget.config.defaults.offline'), color: '#ef4444' }
]);

const localConfig = ref({
  showDeviceName: props.modelValue.showDeviceName !== undefined ? props.modelValue.showDeviceName : true,
  valueColor: props.modelValue.valueColor || '#10b981',
  fontSize: props.modelValue.fontSize !== undefined ? props.modelValue.fontSize : 56,
  prefix: props.modelValue.prefix || '',
  prefixColor: props.modelValue.prefixColor || '#64748b',
  unit: props.modelValue.unit || '',
  unitColor: props.modelValue.unitColor || '#64748b',      
  subtext: props.modelValue.subtext || '',
  subtextColor: props.modelValue.subtextColor || '#64748b',
  enableConditions: props.modelValue.enableConditions || false,
  conditions: props.modelValue.conditions ? JSON.parse(JSON.stringify(props.modelValue.conditions)) : defaultConditions.value
});

const addCondition = () => {
  localConfig.value.conditions.push({
    operator: '==',
    compareValue: '',
    displayText: t('statusWidget.config.statusMet'),
    color: localConfig.value.valueColor 
  });
};

const removeCondition = (index) => {
  localConfig.value.conditions.splice(index, 1);
};

watch(localConfig, (newVal) => {
  emit('update:modelValue', JSON.parse(JSON.stringify(newVal)));
}, { deep: true });

onMounted(() => {
  emit('update:modelValue', JSON.parse(JSON.stringify(localConfig.value)));
});
</script>