<template>
  <div class="flex flex-col h-full w-full p-4 overflow-hidden" :style="backgroundStyle">

    <!-- Glassmorphism Header -->
    <div class="backdrop-blur-md px-4 py-3 shadow-sm z-10 flex justify-between items-center rounded-t-lg">
      <h3 class="m-0 text-base font-extrabold tracking-wide truncate pr-2"
        :style="{ color: widgetData.widgetStyle?.textHex || '#334155' }">
        {{ widgetData?.widgetLabel }}
      </h3>

      <span v-if="hasData" class="badge font-bold shrink-0 text-white shadow-sm"
        :style="{ backgroundColor: criticalCount > 0 ? config.criticalColor : config.resolvedColor, borderColor: criticalCount > 0 ? config.criticalColor : config.resolvedColor }">
        {{ criticalCount }} {{ config.critName }}
      </span>
    </div>

    <!-- Main Content Area -->
    <div
      class="flex-1 w-full relative overflow-y-auto backdrop-blur-sm border-x border-b border-base-200/50 p-2 rounded-b-lg">

      <div v-if="!hasDevices" class="absolute inset-0 flex items-center justify-center text-sm italic text-center p-4"
        :style="{ color: widgetData.widgetStyle?.textHex || '#64748b' }">
        {{ $t('common.noDevice') }}
      </div>

      <div v-else-if="!hasData" class="absolute inset-0 flex flex-col items-center justify-center text-sm gap-3"
        :style="{ color: widgetData.widgetStyle?.textHex || '#64748b' }">
        <span class="loading loading-spinner loading-md text-primary"></span>
        {{ $t('common.waitingData') }}
      </div>

      <!-- Actual Alerts List -->
      <ul v-else class="flex flex-col gap-2">
        <li v-if="filteredAlerts.length === 0" class="text-center py-6 font-medium text-sm"
          :style="{ color: widgetData.widgetStyle?.textHex || '#64748b' }">
          {{ $t('alertWidget.noAlerts') }}
        </li>

        <li v-for="alert in filteredAlerts" :key="alert.id"
          class="flex items-start gap-3 p-3 rounded-lg border-l-4 shadow-sm hover:bg-base-200/20 transition-colors backdrop-blur-md"
          :style="{ borderLeftColor: getStateColor(alert.state) }">

          <div class="mt-0.5 shrink-0">
            <span class="flex h-3 w-3 rounded-full shadow-sm"
              :style="{ backgroundColor: getStateColor(alert.state) }"></span>
          </div>

          <div class="flex-1 min-w-0">
            <div class="flex justify-between items-start gap-2">
              <div class="flex items-center gap-2 truncate">
                <h4 class="font-bold text-sm truncate" :style="{ color: widgetData.widgetStyle?.textHex || '#334155' }">
                  {{ alert.title }}
                </h4>

                <span class="text-[10px] font-bold uppercase tracking-wider px-1.5 py-0.5 rounded text-white shadow-sm"
                  :style="{ backgroundColor: getStateColor(alert.state) }">
                  {{ getStateName(alert.state) }}
                </span>
              </div>
              <div class="flex flex-col items-end text-xs font-semibold whitespace-nowrap gap-0.5"
                :style="{ color: widgetData.widgetStyle?.textHex || '#64748b' }">
                <!-- Current Time -->
                <span>{{ alert.time.split(' ')[0] }}</span>
                <span class="opacity-75">{{ alert.time.split(' ')[1] }}</span>

                <!-- First Critical Hit Time -->
                <div class="mt-1 pt-1 border-t border-base-300/50 flex flex-col items-end w-full text-[10px]">
                  <span v-if="alert.firstCritTime" :style="{ color: config.criticalColor }" class="font-bold">
                    {{ $t('alertWidget.hitTime') }} {{ alert.firstCritTime.split(' ')[1] }}
                  </span>
                  <span v-else class="opacity-40 italic">
                    {{ $t('alertWidget.noHit') }}
                  </span>
                </div>
              </div>
            </div>

            <p v-if="!config.compactMode" class="text-xs mt-1 line-clamp-2 leading-relaxed font-mono opacity-80"
              :style="{ color: widgetData.widgetStyle?.textHex || '#334155' }">
              {{ alert.desc }}
            </p>
          </div>
        </li>
      </ul>

    </div>
  </div>
</template>

<script setup>
import { computed, ref, shallowRef, watch, onMounted, onUnmounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { useLiveStreamStore } from '@/stores/useLiveStreamStore';
import { useFormatter } from '@/composables/useFormatter';

const { formatTime } = useFormatter();
const { t } = useI18n();

const props = defineProps({
  widgetData: {
    type: Object,
    default: () => ({})
  }
});

const widgetInstanceId = `alert-${props.widgetData?.id || Math.random().toString(36).substring(2, 9)}`;

const backgroundStyle = computed(() => {
  const colorObj = props.widgetData.widgetStyle || {};
  const c1 = colorObj.bgHex || '#ffffff';
  if (!colorObj.useGradient) return { backgroundColor: c1 };
  const c2 = colorObj.bgHex2 || c1;
  const angle = colorObj.bgGradientDir || '135deg';
  return { background: `linear-gradient(${angle}, ${c1}, ${c2})` };
});

const liveStreamStore = useLiveStreamStore();
const deviceAlerts = shallowRef({});
const lastProcessedTimestamps = new Map();

const hasDevices = computed(() => props.widgetData?.deviceIds && props.widgetData.deviceIds.length > 0);
const hasData = computed(() => {
  const ids = props.widgetData?.deviceIds || [];
  return ids.some(id => {
    const d = liveStreamStore.liveData[String(id)];
    return d && d.value !== null && d.value !== undefined;
  });
});

const config = computed(() => {
  const customData = props.widgetData?.customChartData || {};
  return {
    maxAlerts: customData.maxAlerts !== undefined ? customData.maxAlerts : 5,
    showResolved: customData.showResolved !== undefined ? customData.showResolved : true,
    compactMode: customData.compactMode !== undefined ? customData.compactMode : false,

    critName: customData.critName || t('alertWidget.config.defaultCritName'),
    critOp: customData.critOp || '>',
    critVal: customData.critVal !== undefined ? customData.critVal : 90,

    warnName: customData.warnName || t('alertWidget.config.defaultWarnName'),
    warnOp: customData.warnOp || '>',
    warnVal: customData.warnVal !== undefined ? customData.warnVal : 70,

    resolvedName: customData.resolvedName || t('alertWidget.config.defaultResName'),

    criticalColor: customData.criticalColor || '#ef4444',
    warningColor: customData.warningColor || '#f59e0b',
    resolvedColor: customData.resolvedColor || '#10b981',
    resetTrigger: customData.resetTrigger || 0,
  };
});

onMounted(() => {
  if (props.widgetData?.deviceIds?.length) {
    liveStreamStore.registerDevices(widgetInstanceId, props.widgetData.deviceIds);
  }
});

onUnmounted(() => {
  liveStreamStore.unregisterDevices(widgetInstanceId);
  lastProcessedTimestamps.clear();
});

watch(() => props.widgetData?.deviceIds, (newIds) => {
  deviceAlerts.value = {};
  lastProcessedTimestamps.clear();
  liveStreamStore.registerDevices(widgetInstanceId, newIds || []);
}, { deep: false });

watch(() => config.value.resetTrigger, () => {
  const updatedAlerts = { ...deviceAlerts.value };
  Object.keys(updatedAlerts).forEach(key => {
    updatedAlerts[key] = { ...updatedAlerts[key], firstCritTime: null };
  });
  deviceAlerts.value = updatedAlerts;
});

const getStateColor = (state) => {
  if (state === 'critical') return config.value.criticalColor;
  if (state === 'warning') return config.value.warningColor;
  return config.value.resolvedColor;
};

const getStateName = (state) => {
  if (state === 'critical') return config.value.critName;
  if (state === 'warning') return config.value.warnName;
  return config.value.resolvedName;
};

const checkCondition = (val, op, target) => {
  if (target === undefined || target === null || target === '') return false;
  const numVal = Number(val);
  const numTarget = Number(target);
  if (isNaN(numVal) || isNaN(numTarget)) return false;

  switch (op) {
    case '>': return numVal > numTarget;
    case '>=': return numVal >= numTarget;
    case '<': return numVal < numTarget;
    case '<=': return numVal <= numTarget;
    case '==': return numVal == numTarget;
    case '!=': return numVal != numTarget;
    default: return false;
  }
};

const evaluateState = (val) => {
  if (checkCondition(val, config.value.critOp, config.value.critVal)) return 'critical';
  if (checkCondition(val, config.value.warnOp, config.value.warnVal)) return 'warning';
  return 'resolved';
};

watch(() => liveStreamStore.liveData, (newData) => {
  const rawDeviceIds = props.widgetData?.deviceIds || [];
  if (rawDeviceIds.length === 0) return;

  // ⚡ FIX: Early exit if no devices on THIS alert list received new telemetry
  const hasRelevantUpdate = rawDeviceIds.some(rawId => {
    const id = String(rawId);
    const dev = newData[id];
    if (!dev?.updatedValueAt || dev.value === null || dev.value === undefined) return false;
    const last = lastProcessedTimestamps.get(id);
    return !last || new Date(dev.updatedValueAt).getTime() > new Date(last).getTime();
  });

  if (!hasRelevantUpdate) return;

  let hasModifications = false;
  const nextAlerts = { ...deviceAlerts.value };

  rawDeviceIds.forEach(rawId => {
    const id = String(rawId);
    const device = newData[id];

    if (!device || device.value === null || device.value === undefined) return;

    const timeToken = device.updatedValueAt || null;
    const lastTime = lastProcessedTimestamps.get(id);

    if (timeToken && lastTime && new Date(timeToken).getTime() <= new Date(lastTime).getTime()) {
      return;
    }
    if (timeToken) {
      lastProcessedTimestamps.set(id, timeToken);
    }

    const state = evaluateState(device.value);
    const existingAlert = nextAlerts[id];
    const alertDate = timeToken ? new Date(timeToken) : new Date();
    const timeStr = formatTime(alertDate);

    let capturedTime = existingAlert ? existingAlert.firstCritTime : null;
    if (state === 'critical' && !capturedTime) {
      capturedTime = timeStr;
    } else if (state !== 'critical') {
      capturedTime = null;
    }

    nextAlerts[id] = {
      id: id,
      state: state,
      title: device.name || `${t('common.device')} ${id}`,
      desc: `${t('alertWidget.latestReading')} ${device.value}`,
      time: timeStr,
      timestamp: alertDate.getTime(),
      firstCritTime: capturedTime
    };

    hasModifications = true;
  });

  if (hasModifications) {
    deviceAlerts.value = nextAlerts;
  }
});

const criticalCount = computed(() => {
  return Object.values(deviceAlerts.value).filter(a => a.state === 'critical').length;
});

const filteredAlerts = computed(() => {
  let alerts = Object.values(deviceAlerts.value);

  if (!config.value.showResolved) {
    alerts = alerts.filter(a => a.state !== 'resolved');
  }

  const severityRank = { 'critical': 3, 'warning': 2, 'resolved': 1 };
  alerts.sort((a, b) => {
    if (severityRank[b.state] !== severityRank[a.state]) {
      return severityRank[b.state] - severityRank[a.state];
    }
    return b.timestamp - a.timestamp;
  });

  return alerts.slice(0, config.value.maxAlerts);
});
</script>