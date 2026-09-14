<template>
  <div class="flex flex-col gap-5">
    
    <!-- Chart Appearance -->
    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">{{ $t('common.chartAppearance') }}</h4>
      
      <div class="grid grid-cols-2 gap-4">
        <label class="form-control w-full">
          <div class="label pb-1"><span class="label-text font-semibold">{{ $t('common.textColor') }}</span></div>
          <input type="color" v-model="localConfig.textColor"
            class="h-10 w-full cursor-pointer rounded border border-base-300 p-0" />
        </label>
        
        <label class="form-control w-full">
          <div class="label pb-1">
            <span class="label-text font-semibold">{{ $t('barChart.config.yAxisLabel') }}</span>
          </div>
          <input type="text" v-model="localConfig.yAxisName"
            class="input input-bordered input-sm w-full" :placeholder="$t('barChart.config.yAxisPlaceholder')" />
        </label>
      </div>
    </div>

    <!-- Device Specific Colors -->
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
          {{ $t('common.noDevicesSelected') }}
        </div>
      </div>
    </div>

  </div>
</template>

<script setup>
import { ref, watch, onMounted, computed } from 'vue';

const props = defineProps({
  modelValue: {
    type: Object,
    default: () => ({})
  },
  selectedDeviceIds: {
    type: Array,
    default: () => []
  },
  allDevices: {
    type: Array,
    default: () => []
  }
});

const emit = defineEmits(['update:modelValue']);

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
  textColor: props.modelValue.textColor || '#334155',
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