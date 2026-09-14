<template>
  <div class="flex flex-col h-full w-full overflow-hidden rounded-box relative" :style="backgroundStyle">
    
    <div class="absolute inset-0 z-0 bg-base-200">
      <img v-if="chartConfig.imageUrl" 
           :src="chartConfig.imageUrl" 
           :style="{ objectFit: chartConfig.imageFit || 'contain' }" 
           class="w-full h-full" 
           alt="Custom Visualization" />
           
      <div v-else class="w-full h-full flex flex-col items-center justify-center text-base-content/40">
        <Icon icon="lucide:image" class="w-12 h-12 mb-2 opacity-50" />
        <span class="text-sm font-semibold italic">{{ $t('visualizationWidget.noImageUrl') }}</span>
      </div>
    </div>

    <div v-if="chartConfig.showHeader !== false" class="backdrop-blur-md bg-base-100/60 px-4 py-3 shadow-sm z-10 flex justify-between items-center rounded-t-box border-b border-base-200/50">
      <h3 class="m-0 text-base font-extrabold tracking-wide drop-shadow-sm" :style="{ color: widgetData.widgetStyle?.textHex || '#334155' }">
        {{ widgetData?.widgetLabel }}
      </h3>
    </div>

    <div v-if="chartConfig.showOverlay && hasDevices" class="absolute z-10 pointer-events-none" :class="overlayPositionClasses">
      
      <div class="backdrop-blur-md shadow-xl px-4 py-3 rounded-xl border flex flex-col gap-3 pointer-events-auto transition-all hover:scale-105 min-w-[150px] max-w-[250px]"
           :style="{ 
             backgroundColor: (chartConfig.overlayBgColor || '#ffffff') + 'CC', 
             borderColor: (chartConfig.overlayTextColor || '#334155') + '33' 
           }">
        
        <div v-for="dev in deviceList" :key="dev.id" 
             class="flex flex-col border-b last:border-b-0 pb-2 last:pb-0"
             :style="{ borderColor: (chartConfig.overlayTextColor || '#334155') + '33' }">
          
          <div class="text-[10px] font-bold opacity-70 uppercase tracking-wider truncate w-full" :style="{ color: chartConfig.overlayTextColor || '#334155' }">
            {{ dev.name }}
          </div>
          
          <div v-if="!dev.hasData" class="flex items-center gap-2 text-sm font-semibold opacity-60 mt-1" :style="{ color: chartConfig.overlayTextColor || '#334155' }">
            <span class="loading loading-spinner loading-xs"></span>
          </div>
          
          <div v-else class="text-xl font-extrabold leading-none mt-1 truncate" :style="{ color: widgetData.widgetStyle?.chartHex || '#3b82f6' }">
            {{ dev.displayValue }}
            <span class="text-sm font-semibold opacity-70 ml-0.5" :style="{ color: chartConfig.overlayTextColor || '#334155' }">
              {{ chartConfig.unit || '' }}
            </span>
          </div>

        </div>
        
      </div>
    </div>

  </div>
</template>

<script setup>
import { computed, onMounted, onUnmounted, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { Icon } from '@iconify/vue';
import { useLiveStreamStore } from '@/stores/useLiveStreamStore';

const { t } = useI18n();

const props = defineProps({
  widgetData: { type: Object, default: () => ({}) }
});

const liveStreamStore = useLiveStreamStore();
const widgetInstanceId = `visualization-${props.widgetData?.id || Math.random().toString(36).substring(2, 9)}`;

const chartConfig = computed(() => props.widgetData?.customChartData || {});

const backgroundStyle = computed(() => {
  if (chartConfig.value.imageUrl) return {}; 
  const colorObj = props.widgetData.widgetStyle || {};
  const c1 = colorObj.bgHex || '#ffffff';
  if (!colorObj.useGradient) return { backgroundColor: c1 };
  const c2 = colorObj.bgHex2 || c1; 
  return { background: `linear-gradient(${colorObj.bgGradientDir || '135deg'}, ${c1}, ${c2})` };
});

const hasDevices = computed(() => props.widgetData?.deviceIds && props.widgetData.deviceIds.length > 0);

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

const deviceList = computed(() => {
  if (!hasDevices.value) return [];
  
  return props.widgetData.deviceIds.map(rawId => {
    const id = String(rawId);
    const stream = liveStreamStore.liveData[id];
    const hasData = stream?.value !== undefined && stream?.value !== null;
    
    let displayValue = '--';
    let name = `${t('common.device')} ${id}`;
    
    if (hasData) {
      name = stream.name || name;
      const val = Number(stream.value);
      const decimals = chartConfig.value.decimalPlaces !== undefined ? chartConfig.value.decimalPlaces : 1;
      displayValue = isNaN(val) ? stream.value : val.toFixed(decimals);
    }
    
    return {
      id: id,
      name: name,
      hasData: hasData,
      displayValue: displayValue
    };
  });
});

const overlayPositionClasses = computed(() => {
  const pos = chartConfig.value.overlayPosition || 'bottom-right';
  switch (pos) {
    case 'top-left': return 'top-14 left-4'; 
    case 'top-right': return 'top-14 right-4';
    case 'bottom-left': return 'bottom-4 left-4';
    case 'bottom-right': default: return 'bottom-4 right-4';
  }
});
</script>