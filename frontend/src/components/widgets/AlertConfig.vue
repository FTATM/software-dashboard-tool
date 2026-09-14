<template>
  <div class="flex flex-col gap-5">

    <!-- Display Toggles -->
    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">{{ $t('common.displayOptions') }}</h4>

      <div class="grid grid-cols-1 gap-4 mb-4">
        <label class="form-control w-full">
          <div class="label pb-1">
            <span class="label-text font-semibold">{{ $t('alertWidget.config.maxAlerts') }}</span>
          </div>
          <input type="number" v-model.number="localConfig.maxAlerts" class="input input-bordered input-sm w-full"
            placeholder="5" />
        </label>
      </div>

      <div class="flex flex-col gap-3">
        <label class="cursor-pointer label justify-start gap-4">
          <input type="checkbox" v-model="localConfig.showResolved" class="toggle toggle-primary toggle-sm" />
          <span class="label-text font-semibold">{{ $t('alertWidget.config.includeResolved') }}</span>
        </label>

        <label class="cursor-pointer label justify-start gap-4">
          <input type="checkbox" v-model="localConfig.compactMode" class="toggle toggle-primary toggle-sm" />
          <span class="label-text font-semibold">{{ $t('alertWidget.config.compactMode') }}</span>
        </label>

      </div>
    </div>

    <!-- Live Alert Threshold Logic with Custom Labels -->
    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">{{ $t('alertWidget.config.thresholdsTitle') }}</h4>

      <!-- Critical Rule -->
      <div class="flex items-center flex-wrap gap-2 mb-3 bg-base-100 p-2 rounded-lg border border-base-200 shadow-sm">
        <input type="color" v-model="localConfig.criticalColor"
          class="h-8 w-10 shrink-0 cursor-pointer rounded border border-base-300 p-0 shadow-sm"
          title="Critical Color" />
        <input type="text" v-model="localConfig.critName" class="input input-bordered input-sm w-24 font-bold"
          :placeholder="$t('alertWidget.config.defaultCritName')" />

        <span class="text-sm font-semibold px-1">{{ $t('alertWidget.config.ifValue') }}</span>
        <select v-model="localConfig.critOp" class="select select-bordered select-sm w-16">
          <option value=">">&gt;</option>
          <option value=">=">&gt;=</option>
          <option value="<">&lt;</option>
          <option value="<=">&lt;=</option>
          <option value="==">==</option>
          <option value="!=">!=</option>
        </select>
        <input type="number" v-model.number="localConfig.critVal" class="input input-bordered input-sm w-24"
          placeholder="90" />
        <button @click="triggerReset" class="btn btn-outline btn-error btn-xs w-full mt-2">
          {{ $t('alertWidget.config.resetHits') }}
        </button>
      </div>

      <!-- Warning Rule -->
      <div class="flex items-center flex-wrap gap-2 mb-3 bg-base-100 p-2 rounded-lg border border-base-200 shadow-sm">
        <input type="color" v-model="localConfig.warningColor"
          class="h-8 w-10 shrink-0 cursor-pointer rounded border border-base-300 p-0 shadow-sm" title="Warning Color" />
        <input type="text" v-model="localConfig.warnName" class="input input-bordered input-sm w-24 font-bold"
          :placeholder="$t('alertWidget.config.defaultWarnName')" />

        <span class="text-sm font-semibold px-1">{{ $t('alertWidget.config.ifValue') }}</span>
        <select v-model="localConfig.warnOp" class="select select-bordered select-sm w-16">
          <option value=">">&gt;</option>
          <option value=">=">&gt;=</option>
          <option value="<">&lt;</option>
          <option value="<=">&lt;=</option>
          <option value="==">==</option>
          <option value="!=">!=</option>
        </select>
        <input type="number" v-model.number="localConfig.warnVal" class="input input-bordered input-sm w-24"
          placeholder="70" />
      </div>

      <!-- Resolved/Normal Rule (Fallback) -->
      <div class="flex items-center flex-wrap gap-2 bg-base-100 p-2 rounded-lg border border-base-200 shadow-sm">
        <input type="color" v-model="localConfig.resolvedColor"
          class="h-8 w-10 shrink-0 cursor-pointer rounded border border-base-300 p-0 shadow-sm"
          title="Resolved Color" />
        <input type="text" v-model="localConfig.resolvedName" class="input input-bordered input-sm w-24 font-bold"
          :placeholder="$t('alertWidget.config.defaultResName')" />
        <span class="text-sm font-semibold text-base-content/50 italic px-2">{{ $t('alertWidget.config.fallbackState')
        }}</span>
      </div>

      <p class="text-[10px] text-base-content/50 mt-3 leading-tight">
        {{ $t('alertWidget.config.evaluationRule') }}
      </p>
    </div>

  </div>
</template>

<script setup>
import { ref, watch, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

const props = defineProps({
  modelValue: {
    type: Object,
    default: () => ({})
  }
});

const emit = defineEmits(['update:modelValue']);

const localConfig = ref({
  maxAlerts: props.modelValue.maxAlerts !== undefined ? props.modelValue.maxAlerts : 5,
  showResolved: props.modelValue.showResolved !== undefined ? props.modelValue.showResolved : true,
  compactMode: props.modelValue.compactMode !== undefined ? props.modelValue.compactMode : false,

  critName: props.modelValue.critName || t('alertWidget.config.defaultCritName'),
  critOp: props.modelValue.critOp || '>',
  critVal: props.modelValue.critVal !== undefined ? props.modelValue.critVal : 90,

  warnName: props.modelValue.warnName || t('alertWidget.config.defaultWarnName'),
  warnOp: props.modelValue.warnOp || '>',
  warnVal: props.modelValue.warnVal !== undefined ? props.modelValue.warnVal : 70,

  resolvedName: props.modelValue.resolvedName || t('alertWidget.config.defaultResName'),

  criticalColor: props.modelValue.criticalColor || '#ef4444',
  warningColor: props.modelValue.warningColor || '#f59e0b',
  resolvedColor: props.modelValue.resolvedColor || '#10b981',
  resetTrigger: props.modelValue.resetTrigger || 0,
});

const triggerReset = () => {
  localConfig.value.resetTrigger = Date.now();
};

watch(localConfig, (newVal) => {
  emit('update:modelValue', JSON.parse(JSON.stringify(newVal)));
}, { deep: true });

onMounted(() => {
  emit('update:modelValue', JSON.parse(JSON.stringify(localConfig.value)));
});
</script>