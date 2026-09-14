<template>
  <div v-if="hasPermission(mainMenuName, 'Display')"
    class="flex flex-col h-screen w-full relative overflow-hidden transition-colors duration-500"
    :style="activeCanvasStyle">

    <!-- Top Control Bar -->
    <div class="w-full px-6 pt-6 pb-2 shrink-0 flex items-center justify-between z-10">

      <!-- Left side: Canvas Selector -->
      <div
        class="inline-flex items-center gap-2 bg-base-100/80 backdrop-blur-md px-3 py-1.5 rounded-xl shadow-sm border border-base-200/50 hover:bg-base-100 transition-colors">
        <Icon icon="lucide:layout-dashboard" class="w-4 h-4 text-primary shrink-0" />
        <select id="canvas-select" v-model="activeCanvasId" @change="loadCurrentCanvas()"
          class="select select-sm border-none bg-transparent shadow-none focus:ring-0 focus:outline-none min-w-[180px] font-bold text-base-content p-0 h-auto min-h-0 pl-1">
          <option v-for="[id, canvas] in allUserCanvasesMap" :key="id" :value="id">
            {{ canvas.name }}
          </option>
        </select>
      </div>

      <!-- Right side: Data Query Button -->
      <button @click="openQueryModal" class="btn btn-sm btn-primary backdrop-blur-md shadow-sm hover:border-primary">
        <Icon icon="lucide:terminal-square" class="w-4 h-4 mr-1" />
        {{ $t('dashboard.dataQuery') }}
      </button>

    </div>

    <div class="flex-1 w-full px-6 pb-6 overflow-y-auto">
      <div class="flex-1 min-h-[500px] w-full">
        <GridLayout v-model:layout="activeLayout" :col-num="12" :row-height="30" :is-draggable="false"
          :is-resizable="false" :vertical-compact="false">
          <GridItem v-for="item in activeLayout" :key="item.i" :x="item.x" :y="item.y" :w="item.w" :h="item.h"
            :i="item.i">
            <div
              class="w-full h-full rounded-box shadow-sm overflow-hidden cursor-default widget-container flex flex-col">
              <component :is="widgetMap[item.widgetTypeName]" :widget-data="item" />
            </div>
          </GridItem>
        </GridLayout>
      </div>
    </div>

    <!-- Query Execution Modal -->
    <dialog ref="queryModal" class="modal">
      <div class="modal-box w-11/12 max-w-5xl p-0 overflow-hidden shadow-2xl flex flex-col max-h-[85vh]">

        <!-- Modal Header -->
        <div class="px-6 py-5 border-b border-base-200 bg-base-100 flex justify-between items-center shrink-0">
          <h3 class="m-0 text-xl font-extrabold text-base-content flex items-center gap-2">
            <div class="p-2 bg-primary/10 text-primary rounded-lg flex items-center justify-center">
              <Icon icon="lucide:database" class="w-5 h-5" />
            </div>
            {{ $t('dashboard.executeDataQuery') }}
          </h3>
          <button class="btn btn-sm btn-circle btn-ghost" @click="closeQueryModal">
            <Icon icon="lucide:x" class="w-4 h-4" />
          </button>
        </div>

        <div class="p-6 bg-base-100 flex-1 overflow-y-auto flex flex-col gap-5">

          <!-- Query Input Section -->
          <div class="flex flex-col gap-2">
            <label class="font-semibold text-sm text-base-content">{{ $t('dashboard.queryString') }}</label>
            <div class="flex gap-4 items-start">
              <textarea v-model="rawQuery" class="textarea textarea-bordered textarea-primary w-full font-mono text-sm"
                :placeholder="$t('dashboard.queryPlaceholder')" rows="3"></textarea>
              <button @click="runQuery" :disabled="!rawQuery || isQuerying"
                class="btn btn-primary text-white shrink-0 h-[84px] px-6">
                <span v-if="isQuerying" class="loading loading-spinner loading-md"></span>
                <div v-else class="flex flex-col items-center">
                  <Icon icon="lucide:play" class="w-5 h-5 mb-1" /> {{ $t('dashboard.run') }}
                </div>
              </button>
            </div>
            <div v-show="queryError" class="font-semibold text-error">query: <{{ queryError?.message || "Error" }}>
            </div>
          </div>

          <!-- Dynamic Results Table -->
          <div class="flex flex-col flex-1 min-h-[300px]">
            <h4 class="font-bold text-sm text-base-content mb-2 border-b border-base-200 pb-2">{{
              $t('dashboard.resultOutput') }}</h4>

            <div v-if="queryResults.length > 0"
              class="overflow-x-auto overflow-y-auto border border-base-200 rounded-box max-h-[400px]">
              <table class="table table-sm table-zebra w-full">
                <thead class="bg-base-200/80 sticky top-0 z-10 backdrop-blur-sm">
                  <tr>
                    <th v-for="col in queryColumns" :key="col" class="uppercase tracking-wider font-bold text-xs">{{ col
                    }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(row, idx) in queryResults" :key="idx">
                    <td v-for="col in queryColumns" :key="col">{{ row[col] }}</td>
                  </tr>
                </tbody>
              </table>
            </div>

            <!-- Empty States -->
            <div v-else-if="hasQueried && !isQuerying"
              class="flex-1 flex flex-col items-center justify-center text-base-content/50 border border-dashed border-base-300 rounded-box bg-base-200/20">
              <Icon icon="lucide:search-x" class="w-8 h-8 mb-2 opacity-50" />
              <p>{{ $t('dashboard.noResults') }}</p>
            </div>
            <div v-else
              class="flex-1 flex flex-col items-center justify-center text-base-content/30 border border-dashed border-base-300 rounded-box">
              <Icon icon="lucide:terminal" class="w-8 h-8 mb-2 opacity-50" />
              <p>{{ $t('dashboard.enterQueryPrompt') }}</p>
            </div>
          </div>

        </div>

        <!-- Footer -->
        <div class="border-t border-base-200 p-5 flex justify-end gap-3 shrink-0 bg-base-50">
          <button type="button" class="btn btn-ghost" @click="closeQueryModal">{{ $t('common.closeWindow') }}</button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop"><button @click="closeQueryModal">close</button></form>
    </dialog>

  </div>
  <NoAccess v-else />
</template>

<script setup>
import { ref, onMounted, watch, onUnmounted, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { GridLayout, GridItem } from 'grid-layout-plus';
import { useFetch } from '@/composables/useFetch';
import { useMutation } from '@/composables/useMutation';
import { toast } from 'vue3-toastify';
import { Icon } from '@iconify/vue';
import { useLiveStreamStore } from '@/stores/useLiveStreamStore';
import { usePermissionStore } from '@/stores/usePermissionStore';
import NoAccess from '@/components/NoAccess.vue';
import { useErrorHandler } from '@/composables/useErrorHandler';
const { handleError } = useErrorHandler();

import BarChart from '@/components/widgets/BarChart.vue';
import BulletChart from '@/components/widgets/BulletChart.vue';
import AnalogueGauge from '@/components/widgets/AnalogueGauge.vue';
import LineChart from '@/components/widgets/LineChart.vue';
import PieChart from '@/components/widgets/PieChart.vue'
import ScatterChart from '@/components/widgets/ScatterChart.vue'
import BarProcess from '@/components/widgets/BarProcess.vue'
import Status from '@/components/widgets/Status.vue'
import Table from '@/components/widgets/Table.vue'
import Alert from '@/components/widgets/Alert.vue'
import Text from '@/components/widgets/Text.vue'
import ScoreCard from '@/components/widgets/ScoreCard.vue';
import RowChart from '@/components/widgets/RowChart.vue';
import InputAction from '@/components/widgets/InputAction.vue';
import DigitalGauge from '@/components/widgets/DigitalGauge.vue';
import BarLineChart from '@/components/widgets/BarLineChart.vue';
import Visualization from '@/components/widgets/Visualization.vue';

const { t } = useI18n();
const mainMenuName = 'Dashboard';

const widgetMap = {
  'BarChart': BarChart, 'BulletChart': BulletChart, 'AnalogueGauge': AnalogueGauge, 'LineChart': LineChart,
  'PieChart': PieChart, 'ScatterChart': ScatterChart, 'BarProcess': BarProcess, 'Status': Status,
  'Table': Table, 'Alert': Alert, 'Text': Text, 'ScoreCard': ScoreCard, 'RowChart': RowChart,
  'InputAction': InputAction, 'DigitalGauge': DigitalGauge, 'BarLineChart': BarLineChart, 'Visualization': Visualization,
};

const liveStreamStore = useLiveStreamStore();
const permissionStore = usePermissionStore();
const { hasPermission } = permissionStore;

const allUserCanvasesMap = ref(new Map());
const activeCanvasId = ref(null);
const activeLayout = ref([]);
const widgetTypeMasterMap = new Map();

const queryModal = ref(null);
const rawQuery = ref('');
const queryResults = ref([]);
const queryColumns = ref([]);
const hasQueried = ref(false);

const { data: widgetTypeMasterData, error: widgetTypeMasterError, execute: widgetTypeMasterApi } = useFetch();
const { data: userAllCanvasFetchData, error: userAllCanvasFetchError, execute: userAllCanvasFetchApi } = useFetch();
const { data: queryData, error: queryError, isLoading: isQuerying, execute: executeQueryApi } = useMutation();

const activeCanvasStyle = computed(() => {
  const canvas = allUserCanvasesMap.value.get(activeCanvasId.value);
  if (!canvas || !canvas.canvasStyle) return {};

  const style = canvas.canvasStyle;

  if (style.useGradient) {
    const c1 = style.bgHex || '#ffffff';
    const c2 = style.bgHex2 || '#f1f5f9';
    const angle = style.bgGradientDir || '135deg';
    return { background: `linear-gradient(${angle}, ${c1}, ${c2})` };
  }

  if (style.bgHex) {
    return { background: style.bgHex };
  }

  return {};
});

const openQueryModal = () => {
  queryError.value = "";
  queryModal.value.showModal();
};

const closeQueryModal = () => {
  queryModal.value.close();
};

const runQuery = async () => {
  if (!rawQuery.value.trim()) return;

  hasQueried.value = false;
  queryResults.value = [];
  queryColumns.value = [];

  await executeQueryApi('/canvas/data/query', { query: rawQuery.value }, 'POST');

  if (!queryError.value && queryData.value?.data) {
    const results = queryData.value.data;

    if (Array.isArray(results) && results.length > 0) {
      queryResults.value = results;
      queryColumns.value = Object.keys(results[0]);
    }

    hasQueried.value = true;
  } else {
    toast.error(handleError(queryError, 'dashboard.messages.queryFailed'));
  }
};

const loadUserCanvas = async () => {
  await userAllCanvasFetchApi('/canvas/getalldetailbyuser');
  if (!userAllCanvasFetchError.value && userAllCanvasFetchData.value) {
    const newMap = new Map();
    userAllCanvasFetchData.value.data.forEach(canvas => {
      const formattedLayout = (canvas.widgets || []).map((widget, index) => {
        const typeInfo = widgetTypeMasterMap.get(widget.widgetTypeId);
        return {
          x: widget.layoutData.x, y: widget.layoutData.y, w: widget.layoutData.w, h: widget.layoutData.h,
          i: widget.widgetId ? widget.widgetId.toString() : index.toString(),
          widgetTypeName: typeInfo ? typeInfo.widgetTypeName : '',

          widgetId: widget.widgetId, widgetTypeId: widget.widgetTypeId, deviceIds: widget.deviceIds || [],
          dataSourceType: widget.deviceGroupId ? 'group' : 'device',
          deviceGroupId: widget.deviceGroupId || null,

          widgetLabel: widget.widgetLabel || '',
          widgetStyle: widget.widgetStyle || { bgHex: '#ffffff', bgHex2: '#f1f5f9', bgGradientDir: '135deg', textHex: '#334155', chartHex: '#3b82f6', useGradient: false },
          customChartData: widget.customChartData || {}
        };
      });
      const stringId = canvas.canvasId.toString();
      newMap.set(stringId, {
        id: stringId,
        name: canvas.canvasName,
        layout: formattedLayout,
        canvasStyle: {
          bgHex: canvas.canvasStyle?.bgHex || '#ffffff',
          bgHex2: canvas.canvasStyle?.bgHex2 || '#f1f5f9',
          bgGradientDir: canvas.canvasStyle?.bgGradientDir || '135deg',
          useGradient: canvas.canvasStyle?.useGradient || false,
          verticalCompact: canvas.canvasStyle?.verticalCompact ?? true
        }
      });
    });
    allUserCanvasesMap.value = newMap;
    if (newMap.size > 0) {
      activeCanvasId.value = newMap.keys().next().value;
      loadCurrentCanvas();
    }
  } else if (userAllCanvasFetchError.value) {
    toast.error(t('common.messages.loadFailed', { item: "User Canvas" }));
  }
};

const loadCurrentCanvas = () => {
  const selectedCanvas = allUserCanvasesMap.value.get(activeCanvasId.value);
  if (selectedCanvas) {
    activeLayout.value = JSON.parse(JSON.stringify(selectedCanvas.layout));
  }
};

const setupData = async () => {
  await widgetTypeMasterApi('/widgettype/getall');
  if (!widgetTypeMasterError.value && widgetTypeMasterData.value) {
    for (let i of widgetTypeMasterData.value.data) widgetTypeMasterMap.set(i.widgetTypeId, i);
  } else if (widgetTypeMasterError.value) {
    toast.error(t('common.messages.loadFailed', { item: "Widget Type" }));
  }
};

onMounted(async () => {
  if (!hasPermission(mainMenuName, 'Display')) return;
  await setupData();
  await loadUserCanvas();
});

onUnmounted(() => {
  liveStreamStore.disconnect();
});


</script>