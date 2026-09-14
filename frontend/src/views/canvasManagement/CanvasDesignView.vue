<template>
  <NoAccess v-if="!hasPermission(mainMenuName, 'Display')" />

  <div v-else class="flex h-screen w-full bg-base-200 relative overflow-hidden transition-colors duration-500"
    :style="activeCanvasStyle">

    <button v-if="!isSidebarOpen"
      class="btn btn-primary shadow-lg absolute left-5 top-[100px] z-40 rounded-full transition-transform hover:scale-105"
      @click="isSidebarOpen = true">
      <Icon icon="lucide:layout-grid" class="w-5 h-5 mr-1" /> {{ $t('canvasDesign.widgetsTitle') }}
    </button>

    <aside
      class="absolute left-0 top-0 bottom-0 w-[250px] bg-base-100 border-r border-base-300 p-5 flex flex-col z-50 transition-transform duration-300 ease-[cubic-bezier(0.4,0,0.2,1)]"
      :class="[isSidebarOpen ? 'translate-x-0' : '-translate-x-full', isDraggingTool ? '!-translate-x-full' : '']">
      <div class="flex justify-between items-start mb-5 border-b-2 border-base-200 pb-3">
        <div>
          <h3 class="m-0 text-[1.1rem] font-bold text-base-content">{{ $t('canvasDesign.availableWidgets') }}</h3>
          <p class="mt-1 mb-0 text-[0.85rem] text-base-content/60">{{ $t('canvasDesign.dragToCanvas') }}</p>
        </div>
        <button class="btn btn-ghost btn-sm btn-square text-base-content/50 hover:text-error"
          @click="isSidebarOpen = false">
          <Icon icon="lucide:x" class="w-5 h-5" />
        </button>
      </div>

      <div class="flex flex-col gap-3 overflow-y-auto pr-2">
        <div v-for="tool in availableTools" :key="tool.type"
          class="flex items-center gap-3 p-3 bg-base-200/50 border border-dashed border-base-300 rounded-box cursor-grab transition-all duration-200 hover:bg-base-200 hover:border-primary hover:shadow-sm active:cursor-grabbing"
          draggable="true" @dragstart="onDragStart($event, tool.type)" @dragend="onDragEnd">
          <Icon :icon="tool.icon" :rotate="tool.rotate || ''" class="w-6 h-6 text-primary shrink-0" />
          <span class="text-sm font-medium text-base-content">{{ tool.name }}</span>
        </div>
      </div>
    </aside>

    <div class="flex-1 w-full p-6 overflow-y-auto">

      <div
        class="flex flex-wrap md:flex-nowrap justify-between items-center gap-4 mb-6 bg-base-100/90 backdrop-blur-md p-4 rounded-box shadow-sm border border-base-200">
        <div class="flex items-center gap-3 font-medium text-base-content overflow-x-auto">
          <label for="canvas-select" class="whitespace-nowrap">{{ $t('canvasDesign.editingCanvas') }}</label>
          <select id="canvas-select" v-model="activeCanvasId" @change="loadCurrentCanvas()"
            class="select select-bordered select-sm min-w-[200px]">
            <option v-for="[id, canvas] in allUserCanvasesMap" :key="id" :value="id">
              {{ canvas.name }}
            </option>
          </select>
          <span
            class="badge badge-warning badge-sm font-bold uppercase tracking-wider animate-pulse py-3 px-3 whitespace-nowrap shrink-0">
            {{ $t('canvasDesign.editorMode') }}
          </span>
        </div>

        <div class="flex gap-3 shrink-0">

          <button @click="showGrid = !showGrid" class="btn btn-outline btn-sm md:btn-md bg-base-100"
            :class="{ 'bg-base-200': showGrid }">
            <Icon :icon="showGrid ? 'lucide:eye-off' : 'lucide:eye'" class="w-4 h-4 mr-2" />
            {{ showGrid ? $t('canvasDesign.hideGrid') : $t('canvasDesign.showGrid') }}
          </button>
          <button @click="toggleLock" class="btn btn-outline btn-sm md:btn-md bg-base-100"
            :class="{ 'bg-error/10 text-error border-error': currentCanvasSettings.lock }">
            <Icon :icon="currentCanvasSettings.lock ? 'lucide:lock' : 'lucide:unlock'" class="w-4 h-4 mr-2" />
            {{ currentCanvasSettings.lock ? $t('canvasDesign.locked') : $t('canvasDesign.unlocked') }}
          </button>
          <button @click="openCanvasConfig" class="btn btn-outline btn-sm md:btn-md bg-base-100">
            <Icon icon="lucide:palette" class="w-4 h-4 mr-2" /> {{ $t('canvasDesign.canvasStyle') }}
          </button>

          <button type="button" class="btn btn-warning btn-sm md:btn-md relative overflow-hidden select-none"
            @pointerdown="startHold" @pointerup="cancelHold" @pointerleave="cancelHold" @contextmenu.prevent>
            <!-- Progress fill effect -->
            <span class="absolute inset-0 bg-black/20 pointer-events-none transition-[width] ease-linear" :style="{
              width: isHolding ? '100%' : '0%',
              transitionDuration: isHolding ? '500ms' : '150ms'
            }" />

            <!-- Grid container stacks both labels in the same slot to lock width -->
            <span class="relative z-10 inline-grid grid-cols-1 grid-rows-1 place-items-center">
              <span class="col-start-1 row-start-1 transition-opacity duration-150"
                :class="isHolding ? 'opacity-100' : 'opacity-0 pointer-events-none'">
                {{ $t('common.onHold') }}
              </span>
              <span class="col-start-1 row-start-1 transition-opacity duration-150"
                :class="!isHolding ? 'opacity-100' : 'opacity-0 pointer-events-none'">
                {{ $t('canvasDesign.resetLayout') }}
              </span>
            </span>
          </button>
          <button @click="saveCurrentCanvas" class="btn btn-success text-white btn-sm md:btn-md">
            <Icon icon="lucide:save" class="w-4 h-4 mr-2" /> {{ $t('common.save') }}
          </button>
        </div>
      </div>

      <div class="flex-1 min-h-[500px] w-full relative" :class="{ 'show-grid-overlay': showGrid }"
        @dragover.prevent="onDragOver" @drop="onDrop">
        <GridLayout v-model:layout="activeLayout" :col-num="12" :row-height="30" :is-draggable="true"
          :is-resizable="true" :vertical-compact="false" :prevent-collision="currentCanvasSettings.lock">
          <GridItem v-for="item in activeLayout" :key="item.i" :x="item.x" :y="item.y" :w="item.w" :h="item.h"
            :i="item.i">
            <div :class="[
              'w-full h-full bg-base-100 border border-base-200 rounded-box shadow-sm overflow-hidden flex flex-col',
              '!cursor-grab border-2 border-dashed !border-indigo-500/100 active:!cursor-grabbing is-editable',
              { 'opacity-60 border-2 border-dashed !border-primary pointer-events-none scale-[0.98]': item.i === 'drop-placeholder' }
            ]" @contextmenu.prevent="onRightClick($event, item.i)">
              <component :is="widgetMap[item.widgetTypeName]" :widget-data="item" />
            </div>
          </GridItem>
        </GridLayout>
      </div>
    </div>
  </div>

  <template v-if="hasPermission('Canvas Design', 'Display')">
    <div v-if="contextMenu.show" class="fixed inset-0 z-[99]" @click="closeContextMenu"
      @contextmenu.prevent="closeContextMenu"></div>
    <div v-if="contextMenu.show" class="fixed z-[100]"
      :style="{ top: contextMenu.y + 'px', left: contextMenu.x + 'px' }">
      <ul class="menu bg-base-100 w-56 rounded-box shadow-xl border border-base-200">
        <li>
          <a class="hover:bg-base-200" @click="promptConfig">
            <Icon icon="lucide:settings" class="w-4 h-4 text-base-content/70" /> {{ $t('canvasDesign.configureWidget')
            }}
          </a>
        </li>
        <li>
          <a class="text-error hover:bg-error/10 hover:text-error" @click="promptDelete">
            <Icon icon="lucide:trash-2" class="w-4 h-4" /> {{ $t('canvasDesign.deleteWidget') }}
          </a>
        </li>
      </ul>
    </div>

    <dialog ref="canvasConfigModal" class="modal z-[200]">
      <div class="modal-box">
        <h3 class="font-bold text-lg mb-4 flex items-center gap-2">
          <Icon icon="lucide:palette" class="w-5 h-5 text-primary" /> {{ $t('canvasDesign.canvasBackground') }}
        </h3>

        <div class="form-control w-full mt-2">
          <div class="label pb-1 flex justify-between items-center">
            <span class="label-text font-bold">{{ $t('canvasDesign.backgroundStyle') }}</span>
            <label class="cursor-pointer label py-0 px-0 gap-2">
              <span class="label-text-alt font-medium">{{ $t('canvasDesign.solid') }}</span>
              <input type="checkbox" v-model="canvasConfigForm.useGradient" class="toggle toggle-primary toggle-sm" />
              <span class="label-text-alt font-medium">{{ $t('canvasDesign.gradient') }}</span>
            </label>
          </div>

          <input v-if="!canvasConfigForm.useGradient" type="color" v-model="canvasConfigForm.bgHex"
            class="h-10 w-full cursor-pointer rounded border border-base-300 p-0 shadow-sm" />

          <div v-else
            class="flex flex-col sm:flex-row items-center gap-3 bg-base-200/40 p-3 rounded-xl border border-base-200">
            <div class="flex-1 flex flex-col gap-1.5 w-full">
              <span class="text-[10px] text-base-content/60 font-bold uppercase tracking-wider">{{
                $t('canvasDesign.start') }}</span>
              <input type="color" v-model="canvasConfigForm.bgHex"
                class="h-9 w-full cursor-pointer rounded border border-base-300 p-0 shadow-sm" />
            </div>
            <Icon icon="lucide:arrow-right" class="w-4 h-4 text-base-content/30 mt-5 hidden sm:block" />
            <div class="flex-1 flex flex-col gap-1.5 w-full">
              <span class="text-[10px] text-base-content/60 font-bold uppercase tracking-wider">{{
                $t('canvasDesign.end') }}</span>
              <input type="color" v-model="canvasConfigForm.bgHex2"
                class="h-9 w-full cursor-pointer rounded border border-base-300 p-0 shadow-sm" />
            </div>
            <div class="flex-1 flex flex-col gap-1.5 w-full">
              <span class="text-[10px] text-base-content/60 font-bold uppercase tracking-wider">{{
                $t('canvasDesign.angle') }}</span>
              <select v-model="canvasConfigForm.bgGradientDir" class="select select-bordered select-sm h-9 w-full">
                <option value="135deg">{{ $t('canvasDesign.directions.diagonal1') }}</option>
                <option value="to right">{{ $t('canvasDesign.directions.horizontal') }}</option>
                <option value="to bottom">{{ $t('canvasDesign.directions.vertical') }}</option>
                <option value="45deg">{{ $t('canvasDesign.directions.diagonal2') }}</option>
              </select>
            </div>
          </div>
        </div>

        <div class="modal-action mt-6">
          <button type="button" @click="closeCanvasConfig" class="btn btn-ghost">{{ $t('common.cancel') }}</button>
          <button type="button" @click="saveCanvasConfig" class="btn btn-primary text-white">{{
            $t('canvasDesign.applyStyle') }}</button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop"><button @click="closeCanvasConfig">close</button></form>
    </dialog>

    <dialog ref="deleteModal" class="modal z-[200]">
      <div class="modal-box modal-animate">
        <h3 class="font-bold text-lg text-base-content">{{ $t('canvasDesign.deleteWidget') }}</h3>
        <p class="py-4 text-base-content/70">{{ $t('canvasDesign.deleteWarning') }}</p>
        <div class="modal-action">
          <button type="button" @click="closeDeleteModal" class="btn btn-ghost">{{ $t('common.noCancel') }}</button>
          <button type="button" @click="confirmDelete" class="btn btn-error text-white">{{ $t('common.yesDelete')
          }}</button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop"><button @click="closeDeleteModal">close</button></form>
    </dialog>

    <dialog ref="configModal" class="modal z-[200]">
      <div class="modal-box max-w-2xl min-h-[450px] max-h-[85vh] flex flex-col p-0">
        <div class="px-6 pt-6 pb-2 relative">
          <button
            class="btn btn-sm btn-circle btn-ghost absolute right-4 top-4 text-base-content/50 hover:text-error hover:bg-error/10"
            @click="closeConfigModal">
            <Icon icon="lucide:x" class="w-4 h-4" />
          </button>
          <h3 class="font-bold text-lg text-base-content flex items-center gap-2 pr-8">
            <Icon icon="lucide:settings-2" class="w-5 h-5 text-primary" /> {{ $t('canvasDesign.widgetConfig') }}
            <span class="badge badge-neutral ml-auto">{{ configForm.widgetTypeName }}</span>
          </h3>
          <div role="tablist" class="tabs tabs-bordered mt-4 w-full">
            <a role="tab" class="tab h-10" :class="{ 'tab-active': activeTab === 'general' }"
              @click="activeTab = 'general'">
              <span class="font-semibold">{{ $t('canvasDesign.generalSettings') }}</span>
            </a>
            <a v-if="widgetConfigComponentMap[configForm.widgetTypeName]" role="tab" class="tab h-10"
              :class="{ 'tab-active': activeTab === 'chart' }" @click="activeTab = 'chart'">
              <span class="font-semibold">{{ $t('canvasDesign.chartConfig') }}</span>
            </a>
          </div>
        </div>

        <div class="p-6 flex-1 overflow-y-auto bg-base-100/50">
          <div v-show="activeTab === 'general'" class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div class="flex flex-col gap-3">
              <label class="form-control w-full">
                <div class="label pb-1"><span class="label-text font-bold">{{ $t('canvasDesign.widgetLabel') }}</span>
                </div>
                <input type="text" v-model="configForm.widgetLabel" class="input input-bordered input-sm w-full" />
              </label>

              <div v-if="deviceSelectionMode !== 'none'" class="form-control w-full">
                <div class="label pb-1 flex justify-between items-end">
                  <div>
                    <span class="label-text font-bold">{{ $t('canvasDesign.dataSource') }}</span>
                    <span v-if="deviceSelectionMode === 'single'"
                      class="label-text-alt text-error font-semibold ml-2">{{ $t('canvasDesign.selectOneDevice')
                      }}</span>
                  </div>

                  <div v-if="deviceSelectionMode === 'multiple'"
                    class="join bg-base-200 p-0.5 rounded-lg border border-base-300">
                    <button type="button" class="btn btn-xs join-item border-none"
                      :class="configForm.dataSourceType === 'device' ? 'bg-base-100 shadow-sm text-primary hover:bg-base-100' : 'btn-ghost text-base-content/60'"
                      @click="configForm.dataSourceType = 'device'">
                      {{ $t('common.devices') }}
                    </button>
                    <button type="button" class="btn btn-xs join-item border-none"
                      :class="configForm.dataSourceType === 'group' ? 'bg-base-100 shadow-sm text-primary hover:bg-base-100' : 'btn-ghost text-base-content/60'"
                      @click="configForm.dataSourceType = 'group'">
                      {{ $t('common.group') }}
                    </button>
                  </div>
                </div>

                <!-- Device Dropdown -->
                <div v-if="configForm.dataSourceType === 'device'">
                  <SearchableDropdown v-model="syncDeviceIds" :multiple="deviceSelectionMode === 'multiple'"
                    :options="Array.from(deviceMasterMap.values())" label-key="deviceName" value-key="deviceId"
                    :placeholder="$t('common.searchDevice')" />
                </div>

                <!-- Group Dropdown -->
                <div v-else-if="configForm.dataSourceType === 'group'" class="flex flex-col gap-2">
                  <SearchableDropdown v-model="configForm.deviceGroupId" :multiple="false"
                    :options="Array.from(groupMasterMap.values())" label-key="groupName" value-key="groupId"
                    :placeholder="$t('common.searchGroup')" />

                  <div v-if="configForm.deviceGroupId"
                    class="flex flex-wrap gap-1.5 p-2 bg-base-200/50 rounded-lg border border-base-200 min-h-[2.5rem]">
                    <template v-if="getDevicesForGroup(configForm.deviceGroupId).length > 0">
                      <span v-for="dev in getDevicesForGroup(configForm.deviceGroupId)" :key="dev.deviceId"
                        class="badge badge-primary badge-sm font-semibold z-0">
                        {{ dev.deviceName }}
                      </span>
                    </template>
                    <span v-else class="text-xs text-base-content/50 font-medium my-auto italic">{{
                      $t('canvasDesign.noDevicesInGroup')
                    }}</span>
                  </div>
                </div>
              </div>
              <div v-else
                class="form-control w-full p-4 bg-base-200/50 rounded-lg border border-base-200 text-center text-sm text-base-content/60">
                {{ $t('canvasDesign.noSourceRequired') }}
              </div>

              <label class="form-control w-full mt-2">
                <div class="label pb-1"><span class="label-text font-bold">{{ $t('common.textColor') }}</span></div>
                <input type="color" v-model="configForm.widgetStyle.textHex"
                  class="h-10 w-full cursor-pointer rounded border border-base-300 p-0 shadow-sm" />
              </label>

              <div class="form-control w-full mt-2">
                <div class="label pb-1 flex justify-between items-center">
                  <span class="label-text font-bold">{{ $t('canvasDesign.backgroundStyle') }}</span>
                  <label class="cursor-pointer label py-0 px-0 gap-2">
                    <span class="label-text-alt font-medium">{{ $t('canvasDesign.solid') }}</span>
                    <input type="checkbox" v-model="configForm.widgetStyle.useGradient"
                      class="toggle toggle-primary toggle-sm" />
                    <span class="label-text-alt font-medium">{{ $t('canvasDesign.gradient') }}</span>
                  </label>
                </div>

                <input v-if="!configForm.widgetStyle.useGradient" type="color" v-model="configForm.widgetStyle.bgHex"
                  class="h-10 w-full cursor-pointer rounded border border-base-300 p-0 shadow-sm" />
                <div v-else
                  class="flex flex-col sm:flex-row items-center gap-3 bg-base-200/40 p-3 rounded-xl border border-base-200">
                  <div class="flex-1 flex flex-col gap-1.5 w-full">
                    <span class="text-[10px] text-base-content/60 font-bold uppercase tracking-wider">{{
                      $t('canvasDesign.start')
                    }}</span>
                    <input type="color" v-model="configForm.widgetStyle.bgHex"
                      class="h-9 w-full cursor-pointer rounded border border-base-300 p-0 shadow-sm" />
                  </div>
                  <Icon icon="lucide:arrow-right" class="w-4 h-4 text-base-content/30 mt-5 hidden sm:block" />
                  <div class="flex-1 flex flex-col gap-1.5 w-full">
                    <span class="text-[10px] text-base-content/60 font-bold uppercase tracking-wider">{{
                      $t('canvasDesign.end')
                    }}</span>
                    <input type="color" v-model="configForm.widgetStyle.bgHex2"
                      class="h-9 w-full cursor-pointer rounded border border-base-300 p-0 shadow-sm" />
                  </div>
                  <div class="flex-1 flex flex-col gap-1.5 w-full">
                    <span class="text-[10px] text-base-content/60 font-bold uppercase tracking-wider">{{
                      $t('canvasDesign.angle')
                    }}</span>
                    <select v-model="configForm.widgetStyle.bgGradientDir"
                      class="select select-bordered select-sm h-9 w-full">
                      <option value="135deg">{{ $t('canvasDesign.directions.diagonal1') }}</option>
                      <option value="to right">{{ $t('canvasDesign.directions.horizontal') }}</option>
                      <option value="to bottom">{{ $t('canvasDesign.directions.vertical') }}</option>
                      <option value="45deg">{{ $t('canvasDesign.directions.diagonal2') }}</option>
                    </select>
                  </div>
                </div>
              </div>

            </div>
            <div class="flex flex-col gap-3">
              <div class="label pb-0"><span class="label-text font-bold">{{ $t('canvasDesign.gridPosition') }}</span>
              </div>
              <div class="grid grid-cols-4 gap-2 items-end">
                <label class="form-control">
                  <span class="label-text-alt mb-1 text-center font-semibold leading-tight">
                    {{ $t('canvasDesign.xCol') }}
                  </span>
                  <input type="number" v-model="configForm.layoutData.x"
                    class="input input-bordered input-sm w-full text-center" />
                </label>

                <label class="form-control">
                  <span class="label-text-alt mb-1 text-center font-semibold leading-tight">
                    {{ $t('canvasDesign.yRow') }}
                  </span>
                  <input type="number" v-model="configForm.layoutData.y"
                    class="input input-bordered input-sm w-full text-center" />
                </label>

                <label class="form-control">
                  <span class="label-text-alt mb-1 text-center font-semibold leading-tight">
                    {{ $t('canvasDesign.width') }}
                  </span>
                  <input type="number" v-model="configForm.layoutData.w"
                    class="input input-bordered input-sm w-full text-center" />
                </label>

                <label class="form-control">
                  <span class="label-text-alt mb-1 text-center font-semibold leading-tight">
                    {{ $t('canvasDesign.height') }}
                  </span>
                  <input type="number" v-model="configForm.layoutData.h"
                    class="input input-bordered input-sm w-full text-center" />
                </label>
              </div>
            </div>
          </div>
          <div v-if="widgetConfigComponentMap[configForm.widgetTypeName]" v-show="activeTab === 'chart'">
            <component :is="widgetConfigComponentMap[configForm.widgetTypeName]"
              :modelValue="configForm.customChartData" :key="widgetToConfig" :selected-device-ids="configForm.deviceIds"
              :all-devices="Array.from(deviceMasterMap.values())"
              @update:modelValue="(val) => configForm.customChartData = val" />
          </div>
        </div>
        <div class="px-6 py-4 border-t border-base-200 bg-base-100 flex justify-end gap-3 rounded-b-box">
          <button type="button" @click="closeConfigModal" class="btn btn-ghost">{{ $t('common.cancel') }}</button>
          <button type="button" @click="saveConfig" class="btn btn-primary text-white px-8"
            :disabled="configForm.customChartData?._isInvalid">
            {{ $t('canvasDesign.saveSettings') }}
          </button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop"><button @click="closeConfigModal">close</button></form>
    </dialog>
  </template>
</template>

<script setup>
import { useLiveStreamStore } from '@/stores/useLiveStreamStore';
import { ref, onMounted, computed, watch, onUnmounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { GridLayout, GridItem } from 'grid-layout-plus';
import { useFetch } from '@/composables/useFetch';
import { useMutation } from '@/composables/useMutation';
import { toast } from 'vue3-toastify';
import { usePermissionStore } from '@/stores/usePermissionStore';
import { Icon } from '@iconify/vue';

import NoAccess from '@/components/NoAccess.vue';
import SearchableDropdown from '@/components/SearchableDropdown.vue';
import { useErrorHandler } from '@/composables/useErrorHandler';
const { handleError } = useErrorHandler();

import BarChart from '@/components/widgets/BarChart.vue';
import BarChartConfig from '@/components/widgets/BarChartConfig.vue';
import BulletChart from '@/components/widgets/BulletChart.vue';
import BulletChartConfig from '@/components/widgets/BulletChartConfig.vue';
import AnalogueGauge from '@/components/widgets/AnalogueGauge.vue';
import AnalogueGaugeConfig from '@/components/widgets/AnalogueGaugeConfig.vue';
import LineChart from '@/components/widgets/LineChart.vue';
import LineChartConfig from '@/components/widgets/LineChartConfig.vue';
import PieChart from '@/components/widgets/PieChart.vue';
import PieChartConfig from '@/components/widgets/PieChartConfig.vue';
import ScatterChart from '@/components/widgets/ScatterChart.vue';
import ScatterChartConfig from '@/components/widgets/ScatterChartConfig.vue';
import BarProcess from '@/components/widgets/BarProcess.vue';
import BarProcessConfig from '@/components/widgets/BarProcessConfig.vue';
import Status from '@/components/widgets/Status.vue';
import StatusConfig from '@/components/widgets/StatusConfig.vue';
import Table from '@/components/widgets/Table.vue';
import TableConfig from '@/components/widgets/TableConfig.vue';
import Alert from '@/components/widgets/Alert.vue';
import AlertConfig from '@/components/widgets/AlertConfig.vue';
import Text from '@/components/widgets/Text.vue';
import TextConfig from '@/components/widgets/TextConfig.vue';
import ScoreCard from '@/components/widgets/ScoreCard.vue';
import ScoreCardConfig from '@/components/widgets/ScoreCardConfig.vue';
import RowChart from '@/components/widgets/RowChart.vue';
import RowChartConfig from '@/components/widgets/RowChartConfig.vue';
import InputAction from '@/components/widgets/InputAction.vue';
import InputActionConfig from '@/components/widgets/InputActionConfig.vue';
import DigitalGauge from '@/components/widgets/DigitalGauge.vue';
import DigitalGaugeConfig from '@/components/widgets/DigitalGaugeConfig.vue';
import BarLineChart from '@/components/widgets/BarLineChart.vue';
import BarLineChartConfig from '@/components/widgets/BarLineChartConfig.vue';
import Visualization from '@/components/widgets/Visualization.vue';
import VisualizationConfig from '@/components/widgets/VisualizationConfig.vue';

const { t } = useI18n();
const mainMenuName = 'Canvas Design';

const widgetMap = {
  'BarChart': BarChart, 'BulletChart': BulletChart, 'AnalogueGauge': AnalogueGauge, 'LineChart': LineChart,
  'PieChart': PieChart, 'ScatterChart': ScatterChart, 'BarProcess': BarProcess, 'Status': Status,
  'Table': Table, 'Alert': Alert, 'Text': Text, 'ScoreCard': ScoreCard, 'RowChart': RowChart,
  'InputAction': InputAction, 'DigitalGauge': DigitalGauge, 'BarLineChart': BarLineChart, 'Visualization': Visualization,
};

const widgetConfigComponentMap = {
  'BarChart': BarChartConfig, 'AnalogueGauge': AnalogueGaugeConfig, 'BulletChart': BulletChartConfig, 'LineChart': LineChartConfig,
  'PieChart': PieChartConfig, 'ScatterChart': ScatterChartConfig, 'BarProcess': BarProcessConfig, 'Status': StatusConfig,
  'Table': TableConfig, 'Alert': AlertConfig, 'Text': TextConfig, 'ScoreCard': ScoreCardConfig,
  'RowChart': RowChartConfig, 'InputAction': InputActionConfig, 'DigitalGauge': DigitalGaugeConfig,
  'BarLineChart': BarLineChartConfig, 'Visualization': VisualizationConfig,
};

const permissionStore = usePermissionStore();
const { hasPermission } = permissionStore;
const liveStreamStore = useLiveStreamStore();

const availableTools = computed(() => [
  { type: 'BarChart', name: t('canvasDesign.widgets.barChart'), icon: 'lucide:bar-chart-3' },
  { type: 'RowChart', name: t('canvasDesign.widgets.rowChart'), icon: 'lucide:bar-chart-horizontal' },
  { type: 'BulletChart', name: t('canvasDesign.widgets.bulletChart'), icon: 'lucide:crosshair' },
  { type: 'AnalogueGauge', name: t('canvasDesign.widgets.analogueGauge'), icon: 'lucide:gauge-circle' },
  { type: 'DigitalGauge', name: t('canvasDesign.widgets.digitalGauge'), icon: 'lucide:disc' },
  { type: 'LineChart', name: t('canvasDesign.widgets.lineChart'), icon: 'lucide:line-chart' },
  { type: 'PieChart', name: t('canvasDesign.widgets.pieChart'), icon: 'lucide:pie-chart' },
  { type: 'ScatterChart', name: t('canvasDesign.widgets.scatterChart'), icon: 'lucide:scatter-chart' },
  { type: 'BarProcess', name: t('canvasDesign.widgets.barProcess'), icon: 'lucide:line-dot-right-horizontal' },
  { type: 'BarLineChart', name: t('canvasDesign.widgets.barLineChart'), icon: 'lucide:chart-column-big' },
  { type: 'Status', name: t('canvasDesign.widgets.status'), icon: 'lucide:activity' },
  { type: 'Table', name: t('canvasDesign.widgets.table'), icon: 'lucide:table' },
  { type: 'Alert', name: t('canvasDesign.widgets.alert'), icon: 'lucide:bell' },
  { type: 'Text', name: t('canvasDesign.widgets.text'), icon: 'lucide:type' },
  { type: 'ScoreCard', name: t('canvasDesign.widgets.scoreCard'), icon: 'lucide:captions' },
  { type: 'InputAction', name: t('canvasDesign.widgets.inputAction'), icon: 'lucide:mouse-pointer-click' },
  { type: 'Visualization', name: t('canvasDesign.widgets.visualization'), icon: 'lucide:paintbrush-vertical' },
]);

const activeTab = ref('general');
const isSidebarOpen = ref(false);
const isDraggingTool = ref(false);
const draggedWidgetTypeName = ref(null);

const contextMenu = ref({ show: false, x: 0, y: 0, widgetId: null });
const deleteModal = ref(null);
const configModal = ref(null);
const showGrid = ref(true);

const canvasConfigModal = ref(null);
const canvasConfigForm = ref({ bgHex: '#ffffff', bgHex2: '#f1f5f9', bgGradientDir: '135deg', useGradient: false });

const configForm = ref({
  widgetLabel: '', deviceIds: [],
  dataSourceType: 'device', deviceGroupId: null,
  widgetStyle: { bgHex: '', bgHex2: '', bgGradientDir: '135deg', textHex: '', chartHex: '', useGradient: false },
  layoutData: { x: 0, y: 0, w: 4, h: 8 }, customChartData: {}
});

const widgetToDelete = ref(null);
const widgetToConfig = ref(null);

const allUserCanvasesMap = ref(new Map());
const activeCanvasId = ref(null);
const activeLayout = ref([]);
const widgetTypeMasterMap = new Map();
const deviceMasterMap = new Map();
const groupMasterMap = new Map();

const { data: widgetTypeMasterData, error: widgetTypeMasterError, execute: widgetTypeMasterApi } = useFetch();
const { data: userAllCanvasFetchData, error: userAllCanvasFetchError, execute: userAllCanvasFetchApi } = useFetch();
const { data: deviceAllNameData, error: deviceAllNameFetchError, execute: deviceAllNameFetchApi } = useFetch();
const { data: groupAllFetch, error: groupAllFetchError, execute: groupAllFetchApi } = useFetch();
const { res: upsertRes, error: upsertError, execute: upsertWidgetApi } = useMutation();

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

const deviceSelectionMode = computed(() => {
  const type = configForm.value.widgetTypeName;
  if (['Text'].includes(type)) return 'none';
  if (['AnalogueGauge', 'DigitalGauge', 'BulletChart', 'Status', 'BarLineChart'].includes(type)) return 'single';
  return 'multiple';
});

const syncDeviceIds = computed({
  get() {
    if (deviceSelectionMode.value === 'single') {
      return configForm.value.deviceIds?.length > 0 ? configForm.value.deviceIds[0] : null;
    }
    return configForm.value.deviceIds || [];
  },
  set(newValue) {
    if (deviceSelectionMode.value === 'single') {
      configForm.value.deviceIds = newValue ? [newValue] : [];
    } else {
      configForm.value.deviceIds = newValue || [];
    }
  }
});

const currentCanvasSettings = computed(() => {
  const canvas = allUserCanvasesMap.value.get(activeCanvasId.value);
  return canvas?.canvasStyle || { lock: false };
});

const toggleLock = () => {
  const canvas = allUserCanvasesMap.value.get(activeCanvasId.value);
  if (canvas && canvas.canvasStyle) {
    canvas.canvasStyle.lock = !canvas.canvasStyle.lock;
  }
};

const getDevicesForGroup = (groupId) => {
  const group = groupMasterMap.get(groupId);
  if (!group || !group.deviceIds) return [];
  return group.deviceIds.map(id => deviceMasterMap.get(id)).filter(Boolean);
};

const loadUserCanvas = async () => {
  await userAllCanvasFetchApi('/canvas/getalldetailbyuser');
  if (!userAllCanvasFetchError.value && userAllCanvasFetchData.value) {
    const newMap = new Map();
    userAllCanvasFetchData.value.data.forEach(canvas => {
      const formattedLayout = (canvas.widgets || []).map((widget, index) => {
        const typeInfo = widgetTypeMasterMap.get(widget.widgetTypeId);

        let dsType = 'device';
        let dGroupId = widget.deviceGroupId || null;
        let resolvedDeviceIds = widget.deviceIds || [];

        if (dGroupId !== null) {
          dsType = 'group';
          const group = groupMasterMap.get(dGroupId);
          if (group && group.deviceIds) {
            resolvedDeviceIds = [...group.deviceIds];
          } else {
            resolvedDeviceIds = [];
          }
        }

        return {
          x: widget.layoutData.x, y: widget.layoutData.y, w: widget.layoutData.w, h: widget.layoutData.h,
          i: widget.widgetId ? widget.widgetId.toString() : index.toString(),
          widgetTypeName: typeInfo ? typeInfo.widgetTypeName : '',
          widgetId: widget.widgetId, widgetTypeId: widget.widgetTypeId,

          deviceIds: resolvedDeviceIds,
          dataSourceType: dsType,
          deviceGroupId: dGroupId,

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
          lock: canvas.canvasStyle?.lock ?? false,
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
    for (let i of widgetTypeMasterData.value.data) {
      widgetTypeMasterMap.set(i.widgetTypeId, i);
    }
  } else if (widgetTypeMasterError.value) {
    toast.error(t('common.messages.loadFailed', { item: "widget type" }));
  }

  await deviceAllNameFetchApi('/device/getalldevicename');
  if (!deviceAllNameFetchError.value && deviceAllNameData.value) {
    for (let i of deviceAllNameData.value.data) {
      deviceMasterMap.set(i.deviceId, i);
    }
  } else if (deviceAllNameFetchError.value) {
    toast.error(t('common.messages.loadFailed', { item: "device" }));
  }

  await groupAllFetchApi('/device/group/getalldetail');
  if (!groupAllFetchError.value && groupAllFetch.value) {
    for (let i of groupAllFetch.value.data) {
      groupMasterMap.set(i.groupId, i);
    }
  } else if (groupAllFetchError.value) {
    toast.error(t('common.messages.loadFailed', { item: "group device" }));
  }
};

const saveCurrentCanvas = async () => {
  const canvas = allUserCanvasesMap.value.get(activeCanvasId.value);
  const widgetsList = activeLayout.value.map(item => ({
    widgetId: item.widgetId || 0, widgetTypeId: item.widgetTypeId || 0, deviceIds: item.deviceIds || [],
    deviceGroupId: item.dataSourceType === 'group' ? (item.deviceGroupId || null) : null,
    widgetLabel: item.widgetLabel || t('canvasDesign.newWidget', { type: item.widgetTypeName }),
    layoutData: { x: item.x, y: item.y, w: item.w, h: item.h },
    widgetStyle: item.widgetStyle, customChartData: item.customChartData
  }));

  const payload = {
    canvasId: Number(activeCanvasId.value),
    canvasStyle: canvas ? canvas.canvasStyle : null,
    UpsertWidgets: widgetsList
  };

  await upsertWidgetApi('/widget/upsert', payload, 'POST');

  if (!upsertError.value && upsertRes.value?.ok) {
    if (canvas) {
      canvas.layout = JSON.parse(JSON.stringify(activeLayout.value));
      allUserCanvasesMap.value.set(activeCanvasId.value, canvas);
    }
    toast.success(t('common.messages.saved'));
  } else {
    toast.error(handleError(upsertError, 'common.messages.saveError'));
  }
};

const isSpaceOccupied = (layout, newX, newY, newW, newH, ignoreId) => {
  return layout.some(item => {
    if (item.i === ignoreId) return false;
    return (newX < item.x + item.w && newX + newW > item.x && newY < item.y + item.h && newY + newH > item.y);
  });
};

const onDragStart = (event, widgetTypeName) => {
  event.dataTransfer.setData('widgetTypeName', widgetTypeName);
  event.dataTransfer.effectAllowed = 'copy';
  draggedWidgetTypeName.value = widgetTypeName;
  setTimeout(() => isDraggingTool.value = true, 0);
};

const onDragOver = (event) => {
  if (!isDraggingTool.value || !draggedWidgetTypeName.value) return;
  const dropZone = event.currentTarget.getBoundingClientRect();
  const mouseX = event.clientX - dropZone.left;
  const mouseY = event.clientY - dropZone.top;
  const colNum = 12, rowHeight = 30, margin = 10, newWidgetWidth = 4;
  const colWidth = dropZone.width / colNum;

  let gridX = Math.floor(mouseX / colWidth);
  let gridY = Math.floor(mouseY / (rowHeight + margin));
  if (gridX + newWidgetWidth > colNum) gridX = colNum - newWidgetWidth;
  if (gridX < 0) gridX = 0;
  if (gridY < 0) gridY = 0;

  if (currentCanvasSettings.value.lock) {
    if (isSpaceOccupied(activeLayout.value, gridX, gridY, newWidgetWidth, 8, 'drop-placeholder')) {
      return;
    }
  }

  const placeholderIndex = activeLayout.value.findIndex(item => item.i === 'drop-placeholder');
  if (placeholderIndex !== -1) {
    const ghost = activeLayout.value[placeholderIndex];
    if (ghost.x !== gridX || ghost.y !== gridY) {
      ghost.x = gridX;
      ghost.y = gridY;
    }
  } else {
    activeLayout.value.push({ x: gridX, y: gridY, w: newWidgetWidth, h: 8, i: 'drop-placeholder', widgetTypeName: draggedWidgetTypeName.value });
  }
};

const onDragEnd = () => {
  isDraggingTool.value = false;
  draggedWidgetTypeName.value = null;
  activeLayout.value = activeLayout.value.filter(item => item.i !== 'drop-placeholder');
};

const onDrop = () => {
  const placeholderIndex = activeLayout.value.findIndex(item => item.i === 'drop-placeholder');
  if (placeholderIndex !== -1) {
    const newItem = activeLayout.value[placeholderIndex];
    newItem.i = 'new-' + Date.now().toString();

    let foundTypeId = 0;
    for (let [id, typeObj] of widgetTypeMasterMap.entries()) {
      if (typeObj.widgetTypeName === newItem.widgetTypeName) { foundTypeId = id; break; }
    }

    newItem.widgetTypeId = foundTypeId; newItem.widgetId = 0;
    newItem.deviceIds = [];
    newItem.deviceGroupId = null;
    newItem.dataSourceType = 'device';

    const translatedName = t(`canvasDesign.widgets.${newItem.widgetTypeName.charAt(0).toLowerCase() + newItem.widgetTypeName.slice(1)}`);
    newItem.widgetLabel = t('canvasDesign.newWidget', { type: translatedName !== `canvasDesign.widgets.${newItem.widgetTypeName.charAt(0).toLowerCase() + newItem.widgetTypeName.slice(1)}` ? translatedName : newItem.widgetTypeName });

    newItem.widgetStyle = { bgHex: '#ffffff', bgHex2: '#f1f5f9', bgGradientDir: '135deg', textHex: '#334155', chartHex: '#3b82f6', useGradient: false };
    newItem.customChartData = {};
  }
  draggedWidgetTypeName.value = null;
  isDraggingTool.value = false;
};

const openCanvasConfig = () => {
  const canvas = allUserCanvasesMap.value.get(activeCanvasId.value);
  if (canvas && canvas.canvasStyle) {
    canvasConfigForm.value = {
      bgHex: canvas.canvasStyle.bgHex || '#ffffff',
      bgHex2: canvas.canvasStyle.bgHex2 || '#f1f5f9',
      bgGradientDir: canvas.canvasStyle.bgGradientDir || '135deg',
      useGradient: canvas.canvasStyle.useGradient || false
    };
  } else {
    canvasConfigForm.value = { bgHex: '#ffffff', bgHex2: '#f1f5f9', bgGradientDir: '135deg', useGradient: false };
  }
  canvasConfigModal.value.showModal();
};

const closeCanvasConfig = () => {
  canvasConfigModal.value.close();
};

const saveCanvasConfig = () => {
  const canvas = allUserCanvasesMap.value.get(activeCanvasId.value);
  if (canvas) {
    canvas.canvasStyle = JSON.parse(JSON.stringify(canvasConfigForm.value));
    allUserCanvasesMap.value.set(activeCanvasId.value, canvas);
  }
  closeCanvasConfig();
};

const onRightClick = (event, widgetId) => {
  contextMenu.value = { show: true, x: event.clientX, y: event.clientY, widgetId: widgetId };
};

const isHolding = ref(false);
let holdTimer = null;

function startHold() {
  isHolding.value = true;
  holdTimer = setTimeout(() => {
    isHolding.value = false;
    loadCurrentCanvas();
  }, 500);
}

function cancelHold() {
  if (holdTimer) {
    clearTimeout(holdTimer);
    holdTimer = null;
  }
  isHolding.value = false;
}

const closeContextMenu = () => { contextMenu.value.show = false; };
const closeDeleteModal = () => { deleteModal.value.close(); widgetToDelete.value = null; };
const promptDelete = () => { widgetToDelete.value = contextMenu.value.widgetId; closeContextMenu(); deleteModal.value.showModal(); };
const confirmDelete = () => { activeLayout.value = activeLayout.value.filter(item => item.i !== widgetToDelete.value); closeDeleteModal(); };

const promptConfig = () => {
  widgetToConfig.value = contextMenu.value.widgetId;
  closeContextMenu();
  const widget = activeLayout.value.find(item => item.i === widgetToConfig.value);

  if (widget) {
    configForm.value = {
      widgetLabel: widget.widgetLabel || '',
      deviceIds: [...(widget.deviceIds || [])],
      widgetTypeId: widget.widgetTypeId || 0,
      widgetTypeName: widget.widgetTypeName,

      dataSourceType: widget.dataSourceType || 'device',
      deviceGroupId: widget.deviceGroupId || null,

      widgetStyle: JSON.parse(JSON.stringify({
        bgHex: widget.widgetStyle?.bgHex || '#ffffff',
        bgHex2: widget.widgetStyle?.bgHex2 || widget.widgetStyle?.bgHex || '#ffffff',
        bgGradientDir: widget.widgetStyle?.bgGradientDir || '135deg',
        textHex: widget.widgetStyle?.textHex || '#334155',
        chartHex: widget.widgetStyle?.chartHex || '#3b82f6',
        useGradient: widget.widgetStyle?.useGradient || false
      })),

      layoutData: { x: widget.x, y: widget.y, w: widget.w, h: widget.h },
      customChartData: JSON.parse(JSON.stringify(widget.customChartData || {}))
    };
  }

  activeTab.value = 'general';
  configModal.value.showModal();
};

const closeConfigModal = () => { configModal.value.close(); widgetToConfig.value = null; };

const saveConfig = () => {
  const widgetIndex = activeLayout.value.findIndex(item => item.i === widgetToConfig.value);

  if (widgetIndex !== -1) {
    let finalDeviceIds = [...configForm.value.deviceIds];
    let finalGroupId = null;

    if (configForm.value.dataSourceType === 'group' && configForm.value.deviceGroupId) {
      finalGroupId = configForm.value.deviceGroupId;
      const selectedGroup = groupMasterMap.get(finalGroupId);
      if (selectedGroup && selectedGroup.deviceIds) {
        finalDeviceIds = [...selectedGroup.deviceIds];
      } else {
        finalDeviceIds = [];
      }
    }

    const updatedItem = {
      ...activeLayout.value[widgetIndex],
      widgetLabel: configForm.value.widgetLabel,

      deviceIds: finalDeviceIds,
      deviceGroupId: finalGroupId,
      dataSourceType: configForm.value.dataSourceType,

      widgetStyle: { ...configForm.value.widgetStyle },
      customChartData: JSON.parse(JSON.stringify(configForm.value.customChartData)),
    };

    updatedItem.x = Number(configForm.value.layoutData.x);
    updatedItem.y = Number(configForm.value.layoutData.y);
    updatedItem.w = Number(configForm.value.layoutData.w);
    updatedItem.h = Number(configForm.value.layoutData.h);

    const newLayout = [...activeLayout.value];
    newLayout[widgetIndex] = updatedItem;
    activeLayout.value = newLayout;
  }

  closeConfigModal();
};

watch(
  [() => configForm.value.deviceGroupId, () => configForm.value.dataSourceType],
  ([newGroupId, newType]) => {
    if (newType === 'group') {
      if (newGroupId) {
        const group = groupMasterMap.get(newGroupId);
        configForm.value.deviceIds = group && group.deviceIds ? [...group.deviceIds] : [];
      } else {
        configForm.value.deviceIds = [];
      }
    }
  }
);

onMounted(async () => {
  if (!hasPermission(mainMenuName, 'Display')) return;
  await setupData();
  await loadUserCanvas();
});

onUnmounted(() => {
  liveStreamStore.disconnect();
});
</script>

<style scoped>
@keyframes slideDown {
  from {
    opacity: 0;
    transform: translateY(-20px) scale(0.95);
  }

  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

.modal-animate {
  animation: slideDown 0.25s ease-out;
}

:deep(.vgl-item__resizer) {
  width: 35px !important;
  height: 35px !important;
  right: 2px !important;
  bottom: 2px !important;
  cursor: se-resize !important;
  z-index: 10 !important;
}

.is-editable:hover+ :deep(.vgl-item__resizer)::after {
  content: '';
  position: absolute;
  right: 4px;
  bottom: 4px;
  width: 12px;
  height: 12px;
  border-right: 3px solid #3b82f6;
  border-bottom: 3px solid #3b82f6;
  border-radius: 2px;
}

.show-grid-overlay :deep(.vgl-layout)::before {
  position: absolute;
  width: calc(100% - 5px);
  height: calc(100% - 5px);
  margin: 5px;
  content: '';
  background-image:
    repeating-linear-gradient(to right, lightgrey 0px, lightgrey 1px, transparent 1px, transparent calc(100% / 12)),
    repeating-linear-gradient(to bottom, lightgrey 0px, lightgrey 1px, transparent 1px, transparent 40px);
  pointer-events: none;
  z-index: 0;
}
</style>