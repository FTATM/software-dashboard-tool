import { defineStore } from 'pinia';
import { shallowRef, ref } from 'vue';

export const useLiveStreamStore = defineStore('liveStream', () => {
  const liveData = shallowRef({});
  const activeDeviceIds = ref([]);

  // Multi-subscriber registry: Map<subscriberId: string, deviceIds: string[]>
  const subscriberRegistry = new Map();

  let eventSource = null;
  let restartTimer = null;

  const getConsolidatedDeviceIds = () => {
    const allIds = new Set();
    for (const ids of subscriberRegistry.values()) {
      for (const id of ids) {
        if (id !== null && id !== undefined && id !== '') {
          allIds.add(String(id));
        }
      }
    }
    return [...allIds].sort();
  };

  const scheduleSSERestart = () => {
    if (restartTimer) clearTimeout(restartTimer);
    restartTimer = setTimeout(() => {
      const nextIds = getConsolidatedDeviceIds();
      const currentIds = [...activeDeviceIds.value].sort();

      if (JSON.stringify(nextIds) === JSON.stringify(currentIds)) {
        return;
      }

      activeDeviceIds.value = nextIds;
      restartSSE();
    }, 40);
  };

  const registerDevices = (subscriberId, deviceIds = []) => {
    if (!subscriberId) subscriberId = 'default';
    const sanitized = Array.isArray(deviceIds)
      ? [...new Set(deviceIds.map(String).filter(Boolean))]
      : [];
    subscriberRegistry.set(String(subscriberId), sanitized);
    scheduleSSERestart();
  };

  const unregisterDevices = (subscriberId) => {
    if (!subscriberId) subscriberId = 'default';
    if (subscriberRegistry.has(String(subscriberId))) {
      subscriberRegistry.delete(String(subscriberId));
      scheduleSSERestart();
    }
  };

  const restartSSE = () => {
    if (eventSource) {
      eventSource.close();
      eventSource = null;
    }

    if (activeDeviceIds.value.length === 0) {
      liveData.value = {};
      return;
    }

    const deviceIdQuery = activeDeviceIds.value.join(',');
    const baseUrl = import.meta.env.VITE_API_BASE_URL;

    eventSource = new EventSource(`${baseUrl}/device/chartstream?deviceId=${deviceIdQuery}`, {
      withCredentials: true
    });

    eventSource.onmessage = (event) => {
      try {
        const payload = JSON.parse(event.data);
        const deviceData = payload.deviceData || {};

        const nextMap = { ...liveData.value };

        for (const [key, device] of Object.entries(deviceData)) {
          const id = String(device.deviceId || key);

          nextMap[id] = {
            deviceId: id,
            name: device.deviceName || nextMap[id]?.name || `Device ${id}`,
            value: device.valueData ?? null,
            updatedValueAt: device.updatedValueAt || null
          };
        }

        liveData.value = nextMap;
      } catch (err) {
        console.error('[useLiveStreamStore] Failed to parse SSE payload:', err);
      }
    };

    eventSource.onerror = () => {
      if (eventSource && eventSource.readyState === EventSource.CLOSED) {
        console.warn('[useLiveStreamStore] SSE closed. Awaiting reconnection.');
      }
    };
  };

  const disconnect = () => {
    if (restartTimer) clearTimeout(restartTimer);
    if (eventSource) {
      eventSource.close();
      eventSource = null;
    }
    subscriberRegistry.clear();
    activeDeviceIds.value = [];
    liveData.value = {};
  };

  return {
    liveData,
    activeDeviceIds,
    registerDevices,
    unregisterDevices,
    disconnect
  };
});