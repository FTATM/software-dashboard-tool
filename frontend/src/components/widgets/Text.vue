<template>
  <div class="flex flex-col h-full w-full p-4 overflow-hidden" :style="backgroundStyle">
    
    <div v-if="config.showHeader" class="backdrop-blur-md px-4 py-3 shadow-sm z-10 flex justify-between items-center rounded-t-lg">
      <h3 class="m-0 text-base font-extrabold tracking-wide" :style="{ color: widgetData.widgetStyle?.textHex || '#334155' }">
        {{ widgetData?.widgetLabel }}
      </h3>
    </div>

    <div class="flex-1 w-full relative overflow-y-auto"
         :class="{ 
           'bg-base-100/30 backdrop-blur-sm border-x border-b border-base-200/50 rounded-b-lg p-4': config.showHeader, 
           'p-2': !config.showHeader 
         }">
      
      <div class="w-full break-words rich-text-container"
           v-html="config.content || $t('textWidget.defaultContent')"
           :style="{ 
             textAlign: config.textAlign, 
             fontSize: config.fontSize + 'px', 
             color: config.textColor 
           }">
      </div>
      
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

const props = defineProps({
  widgetData: {
    type: Object,
    default: () => ({})
  }
});

const backgroundStyle = computed(() => {
  const colorObj = props.widgetData.widgetStyle || {};
  const c1 = colorObj.bgHex || '#ffffff';
  
  if (!colorObj.useGradient) {
    return { backgroundColor: c1 };
  }

  const c2 = colorObj.bgHex2 || c1; 
  const angle = colorObj.bgGradientDir || '135deg';
  
  return {
    background: `linear-gradient(${angle}, ${c1}, ${c2})`
  };
});

const config = computed(() => {
  const customData = props.widgetData?.customChartData || {};
  return {
    content: customData.content || '',
    showHeader: customData.showHeader !== undefined ? customData.showHeader : true,
    textAlign: customData.textAlign || 'left',
    fontSize: customData.fontSize !== undefined ? customData.fontSize : 14,
    textColor: customData.textColor || '#3584e4' 
  };
});
</script>

<style scoped>
.rich-text-container :deep(b),
.rich-text-container :deep(strong) {
  font-weight: 700;
}

.rich-text-container :deep(i),
.rich-text-container :deep(em) {
  font-style: italic;
}

.rich-text-container :deep(u) {
  text-decoration: underline;
}

.rich-text-container :deep(ul) {
  list-style-type: disc;
  padding-left: 1.5rem;
  margin-top: 0.5rem;
  margin-bottom: 0.5rem;
}

.rich-text-container :deep(ol) {
  list-style-type: decimal;
  padding-left: 1.5rem;
  margin-top: 0.5rem;
  margin-bottom: 0.5rem;
}

.rich-text-container :deep(li) {
  margin-bottom: 0.25rem;
}
</style>