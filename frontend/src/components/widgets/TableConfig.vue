<template>
  <div class="flex flex-col gap-5">

    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">{{ $t('common.displayOptions') }}</h4>

      <div class="flex flex-col gap-3">
        <label class="cursor-pointer label justify-start gap-4">
          <input type="checkbox" v-model="localConfig.isDense" class="toggle toggle-primary toggle-sm" />
          <span class="label-text font-semibold">{{ $t('tableWidget.config.denseSpacing') }}</span>
        </label>

        <label class="cursor-pointer label justify-start gap-4">
          <input type="checkbox" v-model="localConfig.isStriped" class="toggle toggle-primary toggle-sm" />
          <span class="label-text font-semibold">{{ $t('tableWidget.config.stripedRows') }}</span>
        </label>

        <label class="cursor-pointer label justify-start gap-4">
          <input type="checkbox" v-model="localConfig.showRowCount" class="toggle toggle-primary toggle-sm" />
          <span class="label-text font-semibold">{{ $t('tableWidget.config.showRowCount') }}</span>
        </label>

        <label class="form-control w-full mt-2">
          <div class="label pb-1">
            <span class="label-text font-semibold">{{ $t('tableWidget.config.maxRows') }}</span>
          </div>
          <input type="number" v-model.number="localConfig.maxRows" class="input input-bordered input-sm w-full"
            placeholder="10" min="1" max="100" />
        </label>

        <label class="cursor-pointer label justify-start gap-4">
          <input type="checkbox" v-model="localConfig.showTimeColumn" class="toggle toggle-secondary toggle-sm" />
          <span class="label-text font-semibold">{{ $t('tableWidget.config.showTimeColumn') }}</span>
        </label>
      </div>
    </div>

    <div class="grid grid-cols-2 gap-4">
      <label class="form-control w-full">
        <div class="label pb-1"><span class="label-text font-bold">{{ $t('tableWidget.config.headerBg') }}</span></div>
        <input type="color" v-model="localConfig.headerColor"
          class="h-10 w-full cursor-pointer rounded border border-base-300 p-0" />
      </label>
      <label class="form-control w-full">
        <div class="label pb-1"><span class="label-text font-bold">{{ $t('tableWidget.config.headerTextColor') }}</span></div>
        <input type="color" v-model="localConfig.headerTextColor"
          class="h-10 w-full cursor-pointer rounded border border-base-300 p-0" />
      </label>
    </div>

  </div>
</template>

<script setup>
import { ref, watch, onMounted } from 'vue';

const props = defineProps({
  modelValue: {
    type: Object,
    default: () => ({})
  }
});

const emit = defineEmits(['update:modelValue']);

const localConfig = ref({
  isStriped: props.modelValue.isStriped !== undefined ? props.modelValue.isStriped : true,
  isDense: props.modelValue.isDense !== undefined ? props.modelValue.isDense : false,
  showRowCount: props.modelValue.showRowCount !== undefined ? props.modelValue.showRowCount : true,
  maxRows: props.modelValue.maxRows !== undefined ? props.modelValue.maxRows : 10,
  headerColor: props.modelValue.headerColor || '#f8fafc',
  headerTextColor: props.modelValue.headerTextColor || '#334155',
  showTimeColumn: props.modelValue.showTimeColumn !== undefined ? props.modelValue.showTimeColumn : true
});

watch(localConfig, (newVal) => {
  emit('update:modelValue', { ...newVal });
}, { deep: true });

onMounted(() => {
  emit('update:modelValue', { ...localConfig.value });
});
</script>