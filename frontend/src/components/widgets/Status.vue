<template>
  <div class="flex flex-col h-full w-full p-4 overflow-hidden" :style="backgroundStyle">
    
    <div class="backdrop-blur-md px-4 py-3 shadow-sm z-10 flex justify-between items-center shrink-0 rounded-lg">
      <h3 class="m-0 text-base font-extrabold tracking-wide truncate pr-2" :style="{ color: widgetData.widgetStyle?.textHex || '#334155' }">
        {{ widgetData?.widgetLabel }}
      </h3>
      <span v-if="hasData && liveDeviceName && showDeviceName"
        class="badge badge-neutral badge-lg py-4 px-4 text-sm font-bold shrink-0 shadow-sm">
        {{ liveDeviceName }}
      </span>
    </div>

    <div class="flex-1 w-full min-h-0 relative p-4 overflow-y-auto flex flex-col mt-2">
      
      <div v-if="!hasDevices" class="m-auto text-sm italic text-center" :style="{ color: widgetData.widgetStyle?.textHex || '#64748b' }">
        {{ $t('common.noDevice') }}
      </div>

      <div v-else-if="!hasData" class="m-auto flex flex-col items-center text-sm gap-3" :style="{ color: widgetData.widgetStyle?.textHex || '#64748b' }">
        <span class="loading loading-spinner loading-md text-primary"></span>
        {{ $t('common.waitingData') }}
      </div>
      
      <div v-else class="m-auto flex flex-col items-center justify-center text-center w-full">
        <div class="flex items-baseline justify-center text-center">
          
          <span v-if="config.prefix" class="font-semibold mr-2 text-2xl" :style="{ color: config.prefixColor }">
            {{ config.prefix }}
          </span>
          
          <span class="font-extrabold leading-none tracking-tight drop-shadow-sm transition-colors duration-300" 
                :style="{ color: evaluatedStatus.color, fontSize: config.fontSize + 'px' }">
            {{ evaluatedStatus.text }}
          </span>
          
          <span v-if="config.unit" class="font-bold ml-2 text-2xl" :style="{ color: config.unitColor }">
            {{ config.unit }}
          </span>
        </div>

        <div v-if="config.subtext" class="mt-2 font-medium text-sm text-center whitespace-pre-line w-full" :style="{ color: config.subtextColor }">
          {{ config.subtext }}
        </div>
      </div>

    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, onUnmounted, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useLiveStreamStore } from '@/stores/useLiveStreamStore';

const { t } = useI18n();

const props = defineProps({
  widgetData: {
    type: Object,
    default: () => ({})
  }
});

const liveStreamStore = useLiveStreamStore();
const widgetInstanceId = `status-${props.widgetData?.id || Math.random().toString(36).substring(2, 9)}`;

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
const liveValue = computed(() => (hasData.value ? deviceObj.value.value : 0));
const liveDeviceName = computed(() => deviceObj.value?.name || '');

const showDeviceName = computed(() => {
  const custom = props.widgetData?.customChartData || {};
  return custom.showDeviceName !== undefined ? custom.showDeviceName : true;
});

const config = computed(() => {
  const customData = props.widgetData?.customChartData || {};
  return {
    valueColor: customData.valueColor || '#10b981', 
    fontSize: customData.fontSize || 56, 
    prefix: customData.prefix || '',
    prefixColor: customData.prefixColor || '#64748b',
    unit: customData.unit || '',
    unitColor: customData.unitColor || '#64748b',      
    subtext: customData.subtext || '',
    subtextColor: customData.subtextColor || '#64748b',
    enableConditions: customData.enableConditions || false,
    conditions: customData.conditions || []
  };
});

onMounted(() => {
  if (props.widgetData?.deviceIds?.length) {
    liveStreamStore.registerDevices(widgetInstanceId, props.widgetData.deviceIds);
  }
});

onUnmounted(() => {
  liveStreamStore.unregisterDevices(widgetInstanceId);
});

watch(() => props.widgetData?.deviceIds, (newIds) => {
  liveStreamStore.registerDevices(widgetInstanceId, newIds || []);
});

const evaluatedStatus = computed(() => {
  const currentValue = liveValue.value !== null && liveValue.value !== undefined ? liveValue.value : '0';

  if (!config.value.enableConditions || config.value.conditions.length === 0) {
    return { text: currentValue, color: config.value.valueColor };
  }

  for (const cond of config.value.conditions) {
    let match = false;
    
    const numCurrent = Number(currentValue);
    const numCompare = Number(cond.compareValue);
    const isNumeric = !isNaN(numCurrent) && !isNaN(numCompare) && cond.compareValue !== '';

    switch (cond.operator) {
      case '==': match = currentValue == cond.compareValue; break;
      case '!=': match = currentValue != cond.compareValue; break;
      case '>': match = isNumeric ? numCurrent > numCompare : currentValue > cond.compareValue; break;
      case '<': match = isNumeric ? numCurrent < numCompare : currentValue < cond.compareValue; break;
      case '>=': match = isNumeric ? numCurrent >= numCompare : currentValue >= cond.compareValue; break;
      case '<=': match = isNumeric ? numCurrent <= numCompare : currentValue <= cond.compareValue; break;
    }

    if (match) {
      return { 
        text: cond.displayText, 
        color: cond.color || config.value.valueColor 
      };
    }
  }

  return { text: currentValue, color: config.value.valueColor };
});
</script>