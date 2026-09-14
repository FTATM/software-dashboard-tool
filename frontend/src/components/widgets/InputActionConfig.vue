<template>
  <div class="flex flex-col gap-5">
    
    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">{{ $t('inputAction.config.actionMode') }}</h4>
      
      <label class="form-control w-full">
        <select v-model="localConfig.actionMode" class="select select-bordered w-full font-bold text-primary">
          <option value="preset">{{ $t('inputAction.config.systemPreset') }}</option>
          <option value="custom">{{ $t('inputAction.config.customWebhook') }}</option>
        </select>
      </label>
    </div>

    <div v-if="localConfig.actionMode === 'preset'" class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">{{ $t('inputAction.config.presetConfig') }}</h4>
      
      <div class="grid grid-cols-1 gap-4">
        <label class="form-control w-full">
          <div class="label pb-1"><span class="label-text font-semibold">{{ $t('inputAction.config.selectPreset') }}</span></div>
          
          <select v-model="localConfig.presetType" class="select select-bordered select-sm w-full">
            <option v-for="preset in presetList" :key="preset.id" :value="preset.id">
              {{ preset.name }} ({{ preset.url }})
            </option>
          </select>
        </label>

        <label class="form-control w-full">
          <div class="label pb-1">
            <span class="label-text font-semibold">{{ $t('inputAction.config.commandKey') }} / {{ $t('common.baseCommandGroup') }}</span>
            <span class="label-text-alt text-base-content/60">{{ $t('inputAction.config.sentToBackend') }}</span>
          </div>
          <input type="text" v-model="localConfig.command" class="input input-bordered input-sm w-full font-mono" :placeholder="$t('inputAction.config.commandKeyPlaceholder')" />
        </label>
      </div>

      <!-- ⚡ Localized Overrides Section -->
      <div class="mt-4 border-t border-base-200/50 pt-4">
        <div class="flex justify-between items-center mb-2">
          <span class="label-text font-semibold text-primary">{{ $t('common.enableOverrides') }}</span>
          <input type="checkbox" v-model="localConfig.enableOverrides" @change="handleOverrideToggle" class="toggle toggle-primary toggle-sm" />
        </div>

        <div v-if="localConfig.enableOverrides" class="flex flex-col gap-2 max-h-48 overflow-y-auto pr-2 mt-3">
          <div v-for="device in activeDevices" :key="device.deviceId" 
               class="flex items-center gap-3 p-2 bg-base-100 rounded-lg border border-base-200 shadow-sm">
            <span class="flex-1 text-sm font-semibold truncate" :title="device.deviceName">{{ device.deviceName }}</span>
            
            <input type="text" v-model="localConfig.deviceOverrides[device.deviceId]" 
                   class="input input-bordered input-sm w-32 font-mono" :placeholder="$t('common.default')" />
          </div>

          <div v-if="activeDevices.length === 0" class="text-sm italic text-base-content/50 text-center py-2">
            {{ $t('common.noDevicesSelected') }}
          </div>
        </div>
      </div>
    </div>

    <div v-if="localConfig.actionMode === 'custom'" class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">{{ $t('inputAction.config.customApiConfig') }}</h4>
      
      <div class="grid grid-cols-1 gap-4">
        
        <div class="flex gap-2">
          <label class="form-control w-1/3">
            <div class="label pb-1"><span class="label-text font-semibold">{{ $t('inputAction.config.method') }}</span></div>
            <select v-model="localConfig.customMethod" class="select select-bordered select-sm w-full font-mono">
              <option value="GET">GET</option>
              <option value="POST">POST</option>
              <option value="PUT">PUT</option>
              <option value="DELETE">DELETE</option>
            </select>
          </label>

          <label class="form-control w-2/3">
            <div class="label pb-1"><span class="label-text font-semibold">{{ $t('inputAction.config.endpointUrl') }}</span></div>
            <input type="text" v-model="localConfig.customUrl" class="input input-bordered input-sm w-full font-mono" placeholder="https://api.example.com/webhook" />
          </label>
        </div>

        <div class="form-control w-full">
          <div class="label pb-1">
            <span class="label-text font-semibold">{{ $t('inputAction.config.headersJson') }}</span>
          </div>
          <textarea v-model="localConfig.customHeaders" class="textarea textarea-bordered w-full h-24 font-mono text-sm leading-tight" placeholder='{
  "Content-Type": "application/json"
}'></textarea>
        </div>

        <div class="form-control w-full" v-if="localConfig.customMethod !== 'GET'">
          <div class="label pb-1">
            <span class="label-text font-semibold">{{ $t('inputAction.config.bodyJson') }}</span>
          </div>
          <textarea v-model="localConfig.customBody" class="textarea textarea-bordered w-full h-24 font-mono text-sm leading-tight" placeholder='{
  "action": "turn_on"
}'></textarea>
        </div>

      </div>
    </div>

    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">{{ $t('inputAction.config.buttonAppearance') }}</h4>
      
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <label class="form-control w-full">
          <div class="label pb-1"><span class="label-text font-semibold">{{ $t('inputAction.config.buttonLabel') }}</span></div>
          <input type="text" v-model="localConfig.buttonText" class="input input-bordered input-sm w-full" :placeholder="$t('inputAction.config.buttonLabelPlaceholder')" />
        </label>
        
        <label class="form-control w-full">
          <div class="label pb-1"><span class="label-text font-semibold">{{ $t('inputAction.config.buttonIcon') }}</span></div>
          <select v-model="localConfig.icon" class="select select-bordered select-sm w-full">
            <option value="lucide:zap">{{ $t('inputAction.config.icons.zap') }}</option>
            <option value="lucide:power">{{ $t('inputAction.config.icons.power') }}</option>
            <option value="lucide:refresh-cw">{{ $t('inputAction.config.icons.refresh') }}</option>
            <option value="lucide:play">{{ $t('inputAction.config.icons.play') }}</option>
            <option value="lucide:square">{{ $t('inputAction.config.icons.stop') }}</option>
            <option value="lucide:check-circle">{{ $t('inputAction.config.icons.check') }}</option>
            <option value="lucide:globe">{{ $t('inputAction.config.icons.globe') }}</option>
          </select>
        </label>
      </div>
    </div>

  </div>
</template>

<script setup>
// ... existing script setup block (no changes needed) ...
import { ref, watch, onMounted, computed } from 'vue';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

const presetList = computed(() => [
  { id: 'default_command', name: t('inputAction.presets.defaultCommand'), url: '/device/command' },
]);

const props = defineProps({
  modelValue: { type: Object, default: () => ({}) },
  selectedDeviceIds: { type: Array, default: () => [] },
  allDevices: { type: Array, default: () => [] }
});

const emit = defineEmits(['update:modelValue']);

const activeDevices = computed(() => {
  if (!props.selectedDeviceIds) return [];
  return props.selectedDeviceIds
    .map(id => props.allDevices.find(device => device.deviceId === id))
    .filter(Boolean);
});

const localConfig = ref({
  actionMode: props.modelValue.actionMode || 'preset',
  presetType: props.modelValue.presetType || 'default_command',
  command: props.modelValue.command || props.modelValue.commandKey || 'action', 
  
  enableOverrides: Object.keys(props.modelValue.deviceOverrides || {}).length > 0,
  
  deviceOverrides: props.modelValue.deviceOverrides || {},
  customUrl: props.modelValue.customUrl || '',
  customMethod: props.modelValue.customMethod || 'POST',
  customHeaders: props.modelValue.customHeaders || '{\n  "Content-Type": "application/json"\n}',
  customBody: props.modelValue.customBody || '{\n  "status": "active"\n}',
  icon: props.modelValue.icon || 'lucide:zap',
  buttonText: props.modelValue.buttonText || t('inputAction.execute')
});

const handleOverrideToggle = () => {
  if (!localConfig.value.enableOverrides) {
    localConfig.value.deviceOverrides = {};
  }
};

watch(() => props.selectedDeviceIds, (newIds) => {
  if (!localConfig.value.enableOverrides) return;

  const cleanOverrides = {};
  newIds.forEach(id => {
    if (localConfig.value.deviceOverrides[id]) {
      cleanOverrides[id] = localConfig.value.deviceOverrides[id];
    }
  });
  localConfig.value.deviceOverrides = cleanOverrides;
}, { immediate: true, deep: true });

watch(localConfig, (newVal) => {
  const payload = JSON.parse(JSON.stringify(newVal));
  if (!payload.enableOverrides) {
    payload.deviceOverrides = {};
  }
  emit('update:modelValue', payload);
}, { deep: true });

onMounted(() => {
  emit('update:modelValue', JSON.parse(JSON.stringify(localConfig.value)));
});
</script>