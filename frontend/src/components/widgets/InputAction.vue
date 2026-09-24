<template>
  <div class="flex flex-col h-full w-full p-4 overflow-hidden rounded-box" :style="backgroundStyle">
    
    <div class="backdrop-blur-md px-4 py-3 shadow-sm z-10 flex justify-between items-center rounded-lg">
      <h3 class="m-0 text-base font-extrabold tracking-wide" :style="{ color: widgetData.widgetStyle?.textHex || '#334155' }">
        {{ widgetData?.widgetLabel }}
      </h3>
    </div>

    <div class="flex-1 w-full relative flex flex-col items-center justify-center mt-2">
      <div v-if="actionConfig.actionMode === 'preset' && !hasDevices" class="text-sm italic text-center p-4" :style="{ color: widgetData.widgetStyle?.textHex || '#64748b' }">
        {{ $t('inputAction.noDevicePreset') }}
      </div>

      <div v-else class="w-full flex flex-col items-center justify-center gap-4">
        <button class="btn border-none shadow-md hover:shadow-lg transition-all w-3/4 max-w-[200px]"
                :style="{ backgroundColor: widgetData.widgetStyle?.chartHex || '#3b82f6', color: '#ffffff' }"
                :disabled="isCurrentlySending"
                @click="fireAction">
          <span v-if="isCurrentlySending" class="loading loading-spinner loading-sm"></span>
          <Icon v-else :icon="actionConfig.icon || 'lucide:zap'" class="w-5 h-5 mr-1" />
          {{ actionConfig.buttonText || $t('inputAction.execute') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { Icon } from '@iconify/vue';
import { toast } from 'vue3-toastify';
import { useMutation } from '@/composables/useMutation';

const { t } = useI18n();

const presetList = computed(() => [
  { id: 'default_command', name: t('inputAction.presets.defaultCommand'), url: '/device/triggercommand' },
]);

const props = defineProps({
  widgetData: { type: Object, default: () => ({}) }
});

const actionConfig = computed(() => props.widgetData?.customChartData || {});
const hasDevices = computed(() => props.widgetData?.deviceIds && props.widgetData.deviceIds.length > 0);

const { error: commandError, isLoading: isSendingInternal, execute: sendCommandApi } = useMutation();
const isSendingExternal = ref(false);
const isCurrentlySending = computed(() => isSendingInternal.value || isSendingExternal.value);

const backgroundStyle = computed(() => {
  const colorObj = props.widgetData.widgetStyle || {};
  const c1 = colorObj.bgHex || '#ffffff';
  if (!colorObj.useGradient) return { backgroundColor: c1 };
  const c2 = colorObj.bgHex2 || c1; 
  return { background: `linear-gradient(${colorObj.bgGradientDir || '135deg'}, ${c1}, ${c2})` };
});

const fireAction = async () => {
  if (isCurrentlySending.value) return;

  if (actionConfig.value.actionMode === 'custom') {
    await executeCustomWebhook();
  } else {
    await executePresetCommand();
  }
};

const executePresetCommand = async () => {
  if (!hasDevices.value) return;

  const selectedPreset = presetList.value.find(p => p.id === actionConfig.value.presetType) || presetList.value[0];

  const activeOverrides = {};
  const overrides = actionConfig.value.deviceOverrides || {};
  props.widgetData.deviceIds.forEach(id => {
    if (overrides[id] && overrides[id].trim() !== '') {
      activeOverrides[id] = overrides[id];
    }
  });

  // ⚡ SIMPLIFIED: Removed defaultCommand, only passing command and deviceOverrides
  const baseCommand = actionConfig.value.command || actionConfig.value.commandKey || '';
  const taskActionPayload = {
    command: baseCommand,
    deviceOverrides: activeOverrides
  };
  
  const isGroupTarget = props.widgetData.dataSourceType === 'group';
  const requestBody = {
    deviceIds: props.widgetData.deviceIds,
    isGroup: isGroupTarget,
    taskAction: taskActionPayload
  };

  await sendCommandApi(selectedPreset.url, requestBody, 'POST');

  if (!commandError.value) {
    toast.success(t('inputAction.messages.commandSent', { count: props.widgetData.deviceIds.length }));
  } else {
    toast.error(commandError.value?.message || t('inputAction.messages.commandFailed'));
  }
};

const executeCustomWebhook = async () => {
  if (!actionConfig.value.customUrl) {
    toast.error(t('inputAction.messages.noCustomUrl'));
    return;
  }
  isSendingExternal.value = true;
  try {
    let headersObj = {};
    if (actionConfig.value.customHeaders && actionConfig.value.customHeaders.trim() !== '') {
      try { headersObj = JSON.parse(actionConfig.value.customHeaders); } 
      catch (err) { throw new Error(t('inputAction.messages.invalidHeaders')); }
    }
    let bodyData = undefined;
    const method = actionConfig.value.customMethod || 'POST';
    if (method !== 'GET' && actionConfig.value.customBody && actionConfig.value.customBody.trim() !== '') {
      try {
        JSON.parse(actionConfig.value.customBody);
        bodyData = actionConfig.value.customBody; 
      } catch (err) { throw new Error(t('inputAction.messages.invalidBody')); }
    }
    const response = await fetch(actionConfig.value.customUrl, { method, headers: headersObj, body: bodyData });
    if (!response.ok) throw new Error(`HTTP ${response.status} ${response.statusText}`);
    toast.success(t('inputAction.messages.webhookSuccess'));
  } catch (error) {
    toast.error(t('inputAction.messages.requestFailed', { error: error.message }));
  } finally {
    isSendingExternal.value = false;
  }
};
</script>