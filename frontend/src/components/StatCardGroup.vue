<template>
  <div class="grid grid-cols-2 sm:grid-cols-4 gap-3 sm:gap-4 w-full">
    <div 
      v-for="(item, index) in items" 
      :key="index"
      class="bg-base-100 border border-base-300 rounded-2xl p-4 shadow-xs flex items-center gap-3.5 transition-transform hover:-translate-y-0.5">
      
      <!-- Icon Container -->
      <div 
        class="w-10 h-10 rounded-xl flex items-center justify-center shrink-0"
        :class="getIconColorClass(item.color)">
        <Icon :icon="item.icon" class="w-5 h-5" />
      </div>

      <!-- Metric Label & Value -->
      <div class="min-w-0 flex-1">
        <span class="text-[11px] uppercase tracking-wider font-bold text-base-content/50 block truncate">
          {{ item.label }}
        </span>
        <span 
          class="text-xl sm:text-2xl font-black tracking-tight"
          :class="item.valueClass || 'text-base-content'">
          {{ item.value }}
        </span>
      </div>

    </div>
  </div>
</template>

<script setup>
import { Icon } from '@iconify/vue';

defineProps({
  items: {
    type: Array,
    required: true
  }
});

const getIconColorClass = (color = 'primary') => {
  const colorMap = {
    primary: 'bg-primary/10 text-primary',
    success: 'bg-success/10 text-success',
    accent: 'bg-accent/10 text-accent',
    info: 'bg-info/10 text-info',
    warning: 'bg-warning/10 text-warning',
    error: 'bg-error/10 text-error',
    ghost: 'bg-base-200 text-base-content/60',
    line: 'bg-[#06c755]/10 text-[#06c755]' // Added LINE green brand color
  };
  return colorMap[color] || colorMap.primary;
};
</script>