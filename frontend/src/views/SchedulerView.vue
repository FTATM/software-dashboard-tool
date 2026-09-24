<template>
  <div v-if="hasPermission(mainMenuName, 'Display')" class="w-full h-full overflow-y-auto p-4">

    <!-- Page Header Card -->
    <div
      class="bg-base-100 shadow-sm rounded-box border border-base-200 p-6 mb-6 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
      <div class="flex items-center gap-4">
        <div class="p-3 bg-primary/10 text-primary rounded-xl flex items-center justify-center">
          <Icon icon="lucide:calendar-clock" class="w-7 h-7" />
        </div>

        <div>
          <h2 class="m-0 text-2xl font-extrabold text-base-content tracking-tight">{{ $t('scheduler.title') }}</h2>
          <p class="mt-1 mb-0 text-base-content/60 text-sm font-medium">{{ $t('scheduler.subtitle') }}</p>
        </div>
      </div>
    </div>

    <!-- Data Table -->
    <TableData :data="schedules" :columns="tableColumns" :is-loading="isSchedulesLoading"
      :initial-sorting="[{ id: 'status', desc: false }]">
      <template #toolbar-actions>
        <button class="btn btn-primary shadow-sm hover:shadow-md transition-all" @click="openCreateModal">
          <Icon icon="lucide:plus" class="w-5 h-5 stroke-[3]" />
          {{ $t('scheduler.addSchedule') }}
        </button>
      </template>

      <template #cell-scheduleId="{ value }">
        <div class="tooltip tooltip-right" :class="copiedId === value ? 'tooltip-success' : ''"
          :data-tip="copiedId === value ? 'Copied!' : 'Copy ID'">
          <div @click="copyToClipboard(value)"
            class="max-w-[100px] md:max-w-[130px] truncate font-mono text-[11px] text-base-content/50 cursor-pointer hover:text-primary transition-colors">
            {{ value }}
          </div>

        </div>
      </template>

      <template #cell-target="{ row }">
        <div v-if="row.deviceGroupId" class="flex items-center gap-2">
          <div class="p-1.5 bg-secondary/10 rounded-md">
            <Icon icon="lucide:layers" class="w-4 h-4 text-secondary" />
          </div>
          <span class="font-semibold">{{ getGroupName(row.deviceGroupId) }}</span>
        </div>
        <div v-else-if="row.deviceId" class="flex items-center gap-2">
          <div class="p-1.5 bg-primary/10 rounded-md">
            <Icon icon="lucide:cpu" class="w-4 h-4 text-primary" />
          </div>
          <span class="font-semibold">{{ getDeviceName(row.deviceId) }}</span>
        </div>
        <div v-else class="badge badge-error badge-sm text-white font-bold">
          {{ $t('scheduler.noTarget') }}
        </div>
      </template>

      <template #cell-taskAction="{ value, row }">
        <div v-if="row.deviceGroupId" class="flex flex-col gap-1">
          <span class="text-sm font-medium whitespace-nowrap">
            {{ $t('common.command') }}: <span
              class="font-mono bg-base-200 px-1.5 py-0.5 rounded border border-base-300">{{ value?.command || '-'
              }}</span>
          </span>
          <span v-if="getActiveOverridesCount(row) > 0"
            class="text-[10px] text-base-content/60 font-bold uppercase tracking-wide">
            {{ $t('common.overridesCount', { count: getActiveOverridesCount(row) }) }}
          </span>
        </div>
        <div v-else>
          <span class="font-mono bg-base-200 px-2 py-1 rounded text-sm border border-base-300">{{ value?.command ||
            value || '-' }}</span>
        </div>
      </template>

      <template #cell-scheduleType="{ value }">
        <span class="badge badge-outline badge-sm uppercase text-[10px] font-bold">
          {{ value === 'one_time' ? $t('scheduler.typeOneTime') : $t('scheduler.typeRecurring') }}
        </span>
      </template>

      <template #cell-startTime="{ value }">
        <div class="flex items-center gap-2">
          <Icon icon="lucide:clock" class="w-4 h-4 text-base-content/50" />
          {{ value }}
        </div>
      </template>

      <template #cell-status="{ row }">
        <!-- 1. Check for missing protocol (Highest Priority Warning) -->
        <div v-if="hasNoProtocol(row)"
          class="badge badge-sm badge-warning text-white font-bold uppercase tracking-wider text-[10px]">
          {{ $t('scheduler.status.noProtocol') }}
        </div>

        <!-- ⚡ 2. NEW: Check if the schedule is Overdue (Error/Danger) -->
        <div v-else-if="isOverdue(row)"
          class="badge badge-sm badge-error text-white font-bold uppercase tracking-wider text-[10px] shadow-sm">
          {{ $t('scheduler.status.overdue') }}
        </div>

        <!-- 3. Standard Database Statuses -->
        <div v-else class="badge badge-sm font-bold uppercase tracking-wider text-[10px]" :class="{
          'badge-success text-white': row.status === 'completed',
          'badge-info text-white': row.status === 'active',
          'badge-error text-white': row.status === 'cancelled'
        }">
          {{ $t(`scheduler.status.${row.status}`) }}
        </div>
      </template>

      <template #cell-actions="{ row }">
        <div class="flex justify-end gap-2">
          <button class="btn btn-sm btn-primary" @click="openEditModal(row)">
            <Icon icon="lucide:pencil" class="w-5 h-5" />
          </button>
          <button class="btn btn-sm btn-error text-white shadow-sm hover:shadow-md transition-all"
            @click="openDeleteModal(row)">
            <Icon icon="lucide:trash-2" class="w-5 h-5" />
          </button>
        </div>
      </template>
    </TableData>

    <!-- Create / Edit Modal -->
    <dialog class="modal modal-bottom sm:modal-middle" :class="{ 'modal-open': isModalOpen }">
      <div class="modal-box sm:max-w-lg p-0 flex flex-col max-h-[90vh]">

        <div class="px-6 py-5 border-b border-base-200 bg-base-100 flex justify-between items-center shrink-0">
          <h3 class="m-0 font-bold text-lg flex items-center gap-2">
            <Icon :icon="modalMode === 'create' ? 'lucide:calendar-plus' : 'lucide:calendar-check'"
              class="w-5 h-5 text-primary" />
            {{ modalMode === 'create' ? $t('scheduler.createSchedule') : $t('scheduler.editSchedule') }}
          </h3>
          <button type="button" class="btn btn-sm btn-circle btn-ghost" @click="closeModal">
            <Icon icon="lucide:x" class="w-4 h-4" />
          </button>
        </div>

        <form @submit.prevent="saveSchedule" class="flex flex-col flex-1 min-h-0">
          <div class="p-6 flex flex-col gap-2 flex-1 overflow-y-auto min-h-0">

            <div class="form-control w-full">
              <div class="label pb-2">
                <span class="label-text font-semibold">{{ $t('scheduler.targetType') }}</span>
              </div>
              <div class="flex gap-4">
                <label
                  class="label cursor-pointer justify-start gap-2 bg-base-200/50 p-2 rounded-lg border border-base-200 flex-1 hover:bg-base-200 transition-colors">
                  <input type="radio" value="device" class="radio radio-primary radio-sm" v-model="form.targetType"
                    @change="v$.$reset()" />
                  <span class="label-text font-medium flex items-center gap-2">
                    <Icon icon="lucide:cpu" class="w-4 h-4" /> {{ $t('common.device') }}
                  </span>
                </label>
                <label
                  class="label cursor-pointer justify-start gap-2 bg-base-200/50 p-2 rounded-lg border border-base-200 flex-1 hover:bg-base-200 transition-colors">
                  <input type="radio" value="group" class="radio radio-primary radio-sm" v-model="form.targetType"
                    @change="v$.$reset()" />
                  <span class="label-text font-medium flex items-center gap-2">
                    <Icon icon="lucide:layers" class="w-4 h-4" /> {{ $t('common.group') }}
                  </span>
                </label>
              </div>
            </div>

            <div class="form-control w-full mt-2">
              <div class="label pb-1">
                <span class="label-text font-semibold">{{ form.targetType === 'device' ?
                  $t('scheduler.selectTargetDevice')
                  : $t('scheduler.selectTargetGroup') }}</span>
              </div>

              <template v-if="form.targetType === 'device'">
                <SearchableDropdown v-model="form.deviceId" :options="realDevices" label-key="deviceName"
                  value-key="deviceId" :placeholder="$t('common.searchDevice')" />
                <div class="label px-1 py-0 h-5">
                  <span v-if="v$.deviceId.$error" class="label-text-alt text-error font-medium">{{
                    v$.deviceId.$errors[0].$message }}</span>
                </div>
              </template>
              <template v-else>
                <SearchableDropdown v-model="form.deviceGroupId" :options="groups" label-key="groupName"
                  value-key="groupId" :placeholder="$t('common.searchGroup')" />
                <div class="label px-1 py-0 h-5">
                  <span v-if="v$.deviceGroupId.$error" class="label-text-alt text-error font-medium">{{
                    v$.deviceGroupId.$errors[0].$message }}</span>
                </div>
              </template>
            </div>

            <!-- Single Device Command -->
            <template v-if="form.targetType === 'device'">
              <label class="form-control w-full">
                <div class="label pb-1">
                  <span class="label-text font-semibold">{{ $t('scheduler.taskCommand') }}</span>
                </div>
                <input type="text" v-model="form.taskActionPayload.command"
                  @blur="v$.taskActionPayload.command.$touch()"
                  :class="['input input-bordered w-full', { 'input-error': v$.taskActionPayload.command.$error }]"
                  :placeholder="$t('scheduler.taskCommandPlaceholder')" />
                <div class="label px-1 py-0 h-5">
                  <span v-if="v$.taskActionPayload.command.$error" class="label-text-alt text-error font-medium">{{
                    v$.taskActionPayload.command.$errors[0].$message }}</span>
                </div>
              </label>
            </template>

            <!-- Group Device Command -->
            <template v-if="form.targetType === 'group' && form.deviceGroupId">
              <div class="p-4 bg-base-200/50 rounded-box border border-base-200 mt-2">
                <h4 class="font-bold text-sm mb-3">{{ $t('common.groupCommandSettings') }}</h4>

                <label class="form-control w-full">
                  <div class="label pb-1"><span class="label-text font-semibold text-primary">{{
                    $t('common.baseCommandGroup') }}</span></div>
                  <input type="text" v-model="form.taskActionPayload.command"
                    @blur="v$.taskActionPayload.command.$touch()"
                    :class="['input input-bordered w-full border-primary', { 'input-error': v$.taskActionPayload.command.$error }]"
                    placeholder="e.g., 0" />
                  <div class="label px-1 py-0 h-5">
                    <span v-if="v$.taskActionPayload.command.$error" class="label-text-alt text-error font-medium">{{
                      v$.taskActionPayload.command.$errors[0].$message }}</span>
                  </div>
                </label>

                <div class="flex justify-between items-center mt-2 border-t border-base-200/50 pt-4 mb-2">
                  <span class="label-text font-semibold text-primary">{{ $t('common.enableOverrides') }}</span>
                  <input type="checkbox" v-model="form.enableOverrides" @change="handleOverrideToggle"
                    class="toggle toggle-primary toggle-sm" />
                </div>

                <div v-if="form.enableOverrides" class="flex flex-col gap-2 max-h-48 overflow-y-auto pr-2">
                  <div v-for="device in activeGroupDevices" :key="device.deviceId"
                    class="flex items-center gap-3 p-2 bg-base-100 rounded-lg border border-base-200">
                    <span class="flex-1 text-sm font-semibold truncate" :title="device.deviceName">{{ device.deviceName
                      }}</span>

                    <input type="text" v-model="form.taskActionPayload.deviceOverrides[device.deviceId]"
                      class="input input-bordered input-sm w-32" :placeholder="$t('common.default')" />
                  </div>
                  <div v-if="activeGroupDevices.length === 0"
                    class="text-sm italic text-base-content/50 text-center py-4">
                    {{ $t('common.noDevicesAssigned') }}
                  </div>
                </div>
              </div>
            </template>

            <div class="form-control w-full mt-2">
              <div class="label pb-2">
                <span class="label-text font-semibold">{{ $t('scheduler.scheduleType') }}</span>
              </div>
              <div class="flex gap-4">
                <label
                  class="label cursor-pointer justify-start gap-2 bg-base-200/50 p-2 rounded-lg border border-base-200 flex-1 hover:bg-base-200 transition-colors">
                  <input type="radio" value="one_time" class="radio radio-primary radio-sm"
                    v-model="form.scheduleType" />
                  <span class="label-text font-medium">{{ $t('scheduler.oneTimeEvent') }}</span>
                </label>
                <label
                  class="label cursor-pointer justify-start gap-2 bg-base-200/50 p-2 rounded-lg border border-base-200 flex-1 hover:bg-base-200 transition-colors">
                  <input type="radio" value="recurring" class="radio radio-primary radio-sm"
                    v-model="form.scheduleType" />
                  <span class="label-text font-medium">{{ $t('scheduler.recurringCron') }}</span>
                </label>
              </div>
            </div>

            <div v-if="form.scheduleType === 'recurring'"
              class="grid grid-cols-1 sm:grid-cols-2 gap-x-4 gap-y-1 mt-2 p-4 bg-base-200/50 border border-base-200 rounded-box">
              <label class="form-control w-full sm:col-span-2">
                <div class="label pb-1">
                  <span class="label-text font-semibold">{{ $t('scheduler.runFrequency') }}</span>
                </div>
                <select v-model="cronPreset" @change="handleCronPresetChange" class="select select-bordered w-full">
                  <option value="*/5 * * * *">{{ $t('scheduler.freq.min5') }}</option>
                  <option value="*/15 * * * *">{{ $t('scheduler.freq.min15') }}</option>
                  <option value="0 * * * *">{{ $t('scheduler.freq.hour1') }}</option>
                  <option value="0 0 * * *">{{ $t('scheduler.freq.midnight') }}</option>
                  <option value="0 8 * * *">{{ $t('scheduler.freq.am8') }}</option>
                  <option value="custom">{{ $t('scheduler.freq.custom') }}</option>
                </select>
              </label>

              <!-- Custom Cron Section with Simple & Advanced Toggle -->
              <div v-if="cronPreset === 'custom'"
                class="form-control w-full sm:col-span-2 mt-2 p-3 bg-base-100 rounded-lg border border-base-200 shadow-sm">

                <!-- Header & Toggle -->
                <div class="flex justify-between items-center mb-4">
                  <div>
                    <span class="label-text font-semibold">{{ $t('scheduler.customTime') }}</span>
                    <div class="text-xs text-base-content/60">
                      {{ isAdvancedCron ? $t('scheduler.advancedMode') : $t('scheduler.customTimeHint') }}
                    </div>
                  </div>
                  <label class="cursor-pointer label flex gap-2 p-0">
                    <span class="label-text text-xs font-bold text-base-content/70">{{ $t('scheduler.advanced')
                      }}</span>
                    <input type="checkbox" class="toggle toggle-primary toggle-sm" v-model="isAdvancedCron"
                      @change="handleAdvancedToggle" />
                  </label>
                </div>

                <!-- SIMPLE MODE: Time Picker & Days -->
                <div v-if="!isAdvancedCron" class="flex flex-col gap-3">
                  <VueDatePicker v-model="customTime" time-picker format="HH:mm" :dark="themeStore.isDarkTheme"
                    :locale="dateFnsLocale" :format-locale="dateFnsLocale" placeholder="Select Time" teleport-center
                    :action-row="{
                      selectBtnLabel: $t('common.select'),
                      cancelBtnLabel: $t('common.cancel')
                    }" @update:model-value="handleTimeChange" @closed="v$.cronExpression.$touch()">
                    <template #input-icon>
                      <Icon icon="lucide:clock" class="w-5 h-5 ml-3 text-base-content/50" />
                    </template>
                  </VueDatePicker>

                  <div>
                    <div class="flex flex-wrap gap-2">
                      <button v-for="day in weekDays" :key="day.value" type="button" class="btn btn-sm"
                        :class="customDays.includes(day.value) ? 'btn-primary text-white border-primary' : 'btn-outline border-base-content/20 text-base-content/70'"
                        @click="toggleDay(day.value)">
                        {{ $t(day.key) }}
                      </button>
                    </div>
                  </div>
                </div>

                <!-- ADVANCED MODE: 5-Input Grid -->
                <div v-else class="flex gap-2 w-full">
                  <label class="form-control flex-1">
                    <div class="label pb-1 px-1"><span class="label-text text-[10px] font-bold">{{
                      $t('scheduler.cronParts.minute') }}</span></div>
                    <input type="text" v-model="cronParts.minute" @input="updateFromCronParts"
                      class="input input-sm input-bordered w-full text-center font-mono" />
                  </label>
                  <label class="form-control flex-1">
                    <div class="label pb-1 px-1"><span class="label-text text-[10px] font-bold">{{
                      $t('scheduler.cronParts.hour') }}</span></div>
                    <input type="text" v-model="cronParts.hour" @input="updateFromCronParts"
                      class="input input-sm input-bordered w-full text-center font-mono" />
                  </label>
                  <label class="form-control flex-1">
                    <div class="label pb-1 px-1"><span class="label-text text-[10px] font-bold">{{
                      $t('scheduler.cronParts.day') }}</span></div>
                    <input type="text" v-model="cronParts.day" @input="updateFromCronParts"
                      class="input input-sm input-bordered w-full text-center font-mono" />
                  </label>
                  <label class="form-control flex-1">
                    <div class="label pb-1 px-1"><span class="label-text text-[10px] font-bold">{{
                      $t('scheduler.cronParts.month') }}</span></div>
                    <input type="text" v-model="cronParts.month" @input="updateFromCronParts"
                      class="input input-sm input-bordered w-full text-center font-mono" />
                  </label>
                  <label class="form-control flex-1">
                    <div class="label pb-1 px-1"><span class="label-text text-[10px] font-bold">{{
                      $t('scheduler.cronParts.week') }}</span></div>
                    <input type="text" v-model="cronParts.week" @input="updateFromCronParts"
                      class="input input-sm input-bordered w-full text-center font-mono" />
                  </label>
                </div>

                <div class="label px-1 py-0 mt-1 h-5">
                  <span v-if="v$.cronExpression.$error" class="label-text-alt text-error font-medium">{{
                    v$.cronExpression.$errors[0].$message }}</span>
                </div>
              </div>

              <label class="form-control w-full sm:col-span-2 mt-2">
                <div class="label pb-1">
                  <span class="label-text font-semibold">{{ $t('scheduler.endTime') }}</span>
                </div>
                <VueDatePicker v-model="form.endTime" :is-24="true" :enable-time-picker="true" auto-apply
                  :preset-dates="presetDates" :locale="dateFnsLocale" :format-locale="dateFnsLocale"
                  :dark="themeStore.isDarkTheme" :formats="{ input: 'dd/MM/yyyy HH:mm' }" :action-row="{
                    selectBtnLabel: $t('common.select'),
                    cancelBtnLabel: $t('common.cancel')
                  }" :placeholder="$t('scheduler.selectEndTime')" teleport-center>
                  <template #input-icon>
                    <Icon icon="lucide:calendar-clock" class="w-5 h-5 ml-3 text-base-content/50" />
                  </template>
                </VueDatePicker>
              </label>
            </div>

            <label class="form-control w-full mt-2">
              <div class="label pb-1">
                <span class="label-text font-semibold">{{ $t('scheduler.startTime') }}</span>
              </div>
              <VueDatePicker v-model="form.startTime" :is-24="true" :enable-time-picker="true" auto-apply
                :preset-dates="presetDates" :locale="dateFnsLocale" :format-locale="dateFnsLocale"
                :dark="themeStore.isDarkTheme" :formats="{ input: 'dd/MM/yyyy HH:mm' }" :action-row="{
                  selectBtnLabel: $t('common.select'),
                  cancelBtnLabel: $t('common.cancel'),
                }" :placeholder="$t('scheduler.selectStartTime')" teleport-center @closed="v$.startTime.$touch()">
                <template #input-icon>
                  <Icon icon="lucide:calendar-clock" class="w-5 h-5 ml-3 text-base-content/50" />
                </template>
              </VueDatePicker>
              <div class="label px-1 py-0 h-5">
                <span v-if="v$.startTime.$error" class="label-text-alt text-error font-medium">{{
                  v$.startTime.$errors[0].$message }}</span>
              </div>
            </label>

            <label v-if="modalMode === 'edit'" class="form-control w-full">
              <div class="label pb-1">
                <span class="label-text font-semibold">{{ $t('common.status') }}</span>
              </div>
              <select v-model="form.status" class="select select-bordered w-full">
                <option value="active">{{ $t('scheduler.status.active') }}</option>
                <option value="completed">{{ $t('scheduler.status.completed') }}</option>
                <option value="cancelled">{{ $t('scheduler.status.cancelled') }}</option>
              </select>
            </label>
          </div>

          <div class="px-6 py-4 border-t border-base-200 bg-base-100 flex justify-end gap-3 shrink-0">
            <button type="button" class="btn btn-ghost" @click="closeModal" :disabled="isSaving">{{ $t('common.cancel')
              }}</button>
            <button type="submit" class="btn btn-primary text-white px-8" :disabled="isSaving">
              <span v-if="isSaving" class="loading loading-spinner loading-sm"></span>
              {{ isSaving ? $t('scheduler.saving') : $t('scheduler.saveSchedule') }}
            </button>
          </div>
        </form>
      </div>
      <form method="dialog" class="modal-backdrop" @click="closeModal"><button>close</button></form>
    </dialog>

    <dialog ref="deleteModal" class="modal z-[200]">
      <div class="modal-box">
        <h3 class="font-bold text-lg text-error flex items-center gap-2">
          <Icon icon="lucide:alert-triangle" class="w-6 h-6" /> {{ $t('common.confirmDelete') || 'Confirm Deletion' }}
        </h3>
        <p class="py-4 text-base-content/80">
          {{ $t('scheduler.deleteWarning')
            || 'Are you sure you want to delete this schedule? This action cannot be undone.' }}
        </p>
        <div class="modal-action">
          <button type="button" @click="closeDeleteModal" class="btn btn-ghost" :disabled="isDeleting">
            {{ $t('common.noCancel') || 'Cancel' }}
          </button>
          <button type="button" @click="confirmDelete" class="btn btn-error text-white" :disabled="isDeleting">
            <span v-if="isDeleting" class="loading loading-spinner loading-sm"></span>
            {{ $t('common.yesDelete') || 'Delete' }}
          </button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop"><button @click="closeDeleteModal">close</button></form>
    </dialog>
  </div>
  <NoAccess v-else />
</template>

<script setup>
import { ref, onMounted, computed, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { Icon } from '@iconify/vue';
import { toast } from 'vue3-toastify';
import { useFetch } from '@/composables/useFetch';
import { useMutation } from '@/composables/useMutation';

import { useVuelidate } from '@vuelidate/core';
import { required, requiredIf, helpers } from '@vuelidate/validators';

import { usePermissionStore } from '@/stores/usePermissionStore';
import NoAccess from '@/components/NoAccess.vue';
import SearchableDropdown from '@/components/SearchableDropdown.vue';
import TableData from '@/components/TableData.vue';
import { useErrorHandler } from '@/composables/useErrorHandler';
const { handleError } = useErrorHandler();
import { useFormatter } from '@/composables/useFormatter';
const { formatTime } = useFormatter();
import { VueDatePicker } from '@vuepic/vue-datepicker';
import { useThemeStore } from '@/stores/useThemeStore';
import thLocale from 'date-fns/locale/th';
import enLocale from 'date-fns/locale/en-US';



const themeStore = useThemeStore();
const { t, locale } = useI18n();
const mainMenuName = 'Scheduler'

const { data: devicesData, error: deviceDataError, execute: fetchDevices } = useFetch();
const { data: groupsData, error: groupDataError, execute: fetchGroups } = useFetch();
const { error: deleteError, isLoading: isDeleting, execute: deleteScheduleApi } = useMutation();

const { data: schedulesData, error: schedulesDataError, isLoading: isSchedulesLoading, execute: fetchSchedules } = useFetch();
const { error: createError, execute: createScheduleApi } = useMutation();
const { error: updateError, execute: updateScheduleApi } = useMutation();

const permissionStore = usePermissionStore();
const { hasPermission } = permissionStore;

const devices = ref([]);
const groups = ref([]);
const schedules = ref([]);
const isModalOpen = ref(false);
const modalMode = ref('create');
const isSaving = ref(false);
const presetDates = ref([{ label: t('common.today'), value: new Date() }]);
const cronPreset = ref('*/5 * * * *');
const deleteModal = ref(null);
const scheduleToDelete = ref(null);

// --- Custom Cron State ---
const customTime = ref({ hours: 12, minutes: 0 });
const customDays = ref([]);
const isAdvancedCron = ref(false);
const cronParts = ref({ minute: '*', hour: '*', day: '*', month: '*', week: '*' });

const weekDays = [
  { key: 'scheduler.days.sun', value: '0' },
  { key: 'scheduler.days.mon', value: '1' },
  { key: 'scheduler.days.tue', value: '2' },
  { key: 'scheduler.days.wed', value: '3' },
  { key: 'scheduler.days.thu', value: '4' },
  { key: 'scheduler.days.fri', value: '5' },
  { key: 'scheduler.days.sat', value: '6' }
];

const tableColumns = computed(() => [
  { header: t('scheduler.table.id'), accessorKey: 'scheduleId' },
  {
    header: t('scheduler.table.target'),
    id: 'target',
    accessorFn: (row) => {
      if (row.deviceGroupId) return getGroupName(row.deviceGroupId);
      if (row.deviceId) return getDeviceName(row.deviceId);
      return t('scheduler.noTarget');
    },
    enableSorting: false
  },
  { header: t('scheduler.table.taskAction'), accessorKey: 'taskAction' },
  {
    header: t('scheduler.table.type'),
    id: 'scheduleType',
    accessorFn: (row) => row.scheduleType === 'one_time' ? t('scheduler.typeOneTime') : t('scheduler.typeRecurring')
  },
  {
    header: t('scheduler.table.startTime'),
    id: 'startTime',
    accessorFn: (row) => row.startTime ? formatTime(row.startTime) : '-'
  },
  {
    header: t('common.status'),
    id: 'status',
    accessorFn: (row) => {
      if (hasNoProtocol(row)) return t('scheduler.status.noProtocol');
      if (isOverdue(row)) return t('scheduler.status.overdue');
      return t(`scheduler.status.${row.status}`);
    },
    sortingFn: (rowA, rowB) => {
      const priority = { 'active': 1, 'completed': 2, 'cancelled': 3 };
      const weightA = priority[rowA.original.status] || 99;
      const weightB = priority[rowB.original.status] || 99;
      return weightA - weightB;
    }
  },
  { header: t('common.actions'), id: 'actions', enableSorting: false, meta: { headerClass: 'text-right', cellClass: 'text-right' } }
]);

const form = ref({
  scheduleId: null,
  targetType: 'device',
  deviceId: null,
  deviceGroupId: null,
  enableOverrides: false,
  taskActionPayload: {
    command: '',
    deviceOverrides: {}
  },
  scheduleType: 'one_time',
  status: 'active',
  startTime: null,
  endTime: null,
  cronExpression: '*/5 * * * *'
});

// Robust resolver to handle CJS, ESM, or nested { default: ... } structures
const resolveLocale = (loc) => {
  if (!loc) return undefined;
  if (loc.localize) return loc;
  if (loc.default?.localize) return loc.default;
  if (loc.th?.localize) return loc.th;
  if (loc.enUS?.localize) return loc.enUS;
  return loc.default || loc;
};

const dateFnsLocale = computed(() => {
  // Use startsWith so both 'th' and 'th-TH' match
  const isThai = locale.value?.startsWith('th');
  return resolveLocale(isThai ? thLocale : enLocale);
});

const realDevices = computed(() => {
  return devices.value.filter(d => !d.refDeviceId);
});

// --- Custom Cron Generators ---
const updateCustomCronString = () => {
  const mins = customTime.value.minutes;
  const hrs = customTime.value.hours;
  const days = customDays.value.length > 0 ? customDays.value.join(',') : '*';
  form.value.cronExpression = `${mins} ${hrs} * * ${days}`;
};

const handleTimeChange = (timeObj) => {
  if (timeObj) {
    customTime.value = timeObj;
    updateCustomCronString();
  }
};

const toggleDay = (dayValue) => {
  const index = customDays.value.indexOf(dayValue);
  if (index === -1) {
    customDays.value.push(dayValue);
  } else {
    customDays.value.splice(index, 1);
  }
  customDays.value.sort();
  updateCustomCronString();
};

const updateFromCronParts = () => {
  form.value.cronExpression = `${cronParts.value.minute} ${cronParts.value.hour} ${cronParts.value.day} ${cronParts.value.month} ${cronParts.value.week}`;
};

const parseIntoCronParts = (cronStr) => {
  const parts = (cronStr || '* * * * *').split(' ');
  cronParts.value = {
    minute: parts[0] || '*',
    hour: parts[1] || '*',
    day: parts[2] || '*',
    month: parts[3] || '*',
    week: parts[4] || '*'
  };
};

const handleAdvancedToggle = () => {
  if (isAdvancedCron.value) {
    parseIntoCronParts(form.value.cronExpression);
  } else {
    updateCustomCronString();
  }
};

const rules = computed(() => ({
  deviceId: {
    requiredIfDevice: helpers.withMessage(t('scheduler.validation.selectDevice'), requiredIf(() => form.value.targetType === 'device'))
  },
  deviceGroupId: {
    requiredIfGroup: helpers.withMessage(t('scheduler.validation.selectGroup'), requiredIf(() => form.value.targetType === 'group'))
  },
  taskActionPayload: {
    command: {
      required: helpers.withMessage(t('scheduler.validation.taskCommandRequired'), required)
    }
  },
  startTime: {
    required: helpers.withMessage(t('scheduler.validation.startTimeRequired'), required)
  },
  cronExpression: {
    requiredIfRecurring: helpers.withMessage(t('scheduler.validation.cronRequired'), requiredIf(() => form.value.scheduleType === 'recurring' && cronPreset.value === 'custom'))
  }
}));

const v$ = useVuelidate(rules, form);

const activeGroupDevices = computed(() => {
  if (form.value.targetType !== 'group' || !form.value.deviceGroupId) return [];
  const group = groups.value.find(g => g.groupId === form.value.deviceGroupId);

  let groupDevs = [];
  if (group && group.devices) {
    groupDevs = group.devices;
  } else if (group && group.deviceIds) {
    groupDevs = group.deviceIds
      .map(id => devices.value.find(d => d.deviceId === id))
      .filter(Boolean);
  } else {
    groupDevs = devices.value.filter(d => d.deviceGroupId === form.value.deviceGroupId);
  }

  // ⚡ Exclude virtual devices from the command override list
  return groupDevs.filter(d => !d.refDeviceId);
});

const hasNoProtocol = (row) => {
  if (row.deviceId) {
    const device = devicesData.value?.data?.find(d => d.deviceId === row.deviceId);
    // ⚡ Flag warning if device has no protocol OR is a virtual sensor
    if (device && (!device.protocol || device.protocol === 'none' || device.refDeviceId)) {
      return true;
    }
  }
  return false;
};

const isOverdue = (row) => {
  if (row.status === 'active') {
    const nowMs = new Date().getTime();

    if (row.scheduleType === 'one_time' && row.startTime) {
      const startTimeMs = new Date(row.startTime).getTime();
      if (startTimeMs < nowMs) {
        return true;
      }
    }

    if (row.scheduleType === 'recurring' && row.endTime) {
      const endTimeMs = new Date(row.endTime).getTime();
      if (endTimeMs < nowMs) {
        return true;
      }
    }
  }

  return false;
};

const copiedId = ref(null);

const copyToClipboard = async (text) => {
  try {
    await navigator.clipboard.writeText(text);

    copiedId.value = text;

    // reset after 2s
    setTimeout(() => {
      if (copiedId.value === text) {
        copiedId.value = null;
      }
    }, 2000);

  } catch (err) {
    toast.error("Failed to copy ID.");
  }
};

const handleOverrideToggle = () => {
  if (!form.value.enableOverrides) {
    form.value.taskActionPayload.deviceOverrides = {};
  }
};


const loadData = async () => {
  await fetchDevices('/device/getalldetail');
  if (deviceDataError.value) toast.error(deviceDataError.value.message || t('common.messages.loadFailed', { item: "device" }));
  if (devicesData.value?.data) devices.value = devicesData.value.data;
  await fetchGroups('/device/group/getalldetail');
  if (groupDataError.value) toast.error(groupDataError.value.message || t('common.messages.loadFailed'), { item: "group device" });
  if (groupsData.value?.data) groups.value = groupsData.value.data;
  await fetchSchedules('/schedule/getalldetail');
  if (schedulesDataError.value) toast.error(schedulesDataError.value.message || t('common.messages.loadFailed', { item: "schedule" }));
  if (schedulesData.value?.data) schedules.value = schedulesData.value.data;
};

const getDeviceName = (id) => {
  const device = devices.value.find(d => d.deviceId === id);
  return device ? device.deviceName : t('scheduler.unknownDevice');
};

const getGroupName = (id) => {
  const group = groups.value.find(g => g.groupId === id);
  return group ? group.groupName : t('scheduler.unknownGroup');
};

const getActiveOverridesCount = (row) => {
  if (!row.deviceGroupId || !row.taskAction?.deviceOverrides) return 0;

  const group = groups.value.find(g => g.groupId === row.deviceGroupId);
  let activeIds = [];

  if (group && group.devices) {
    activeIds = group.devices.map(d => Number(d.deviceId));
  } else if (group && group.deviceIds) {
    activeIds = group.deviceIds.map(id => Number(id));
  } else {
    activeIds = devices.value.filter(d => d.deviceGroupId === row.deviceGroupId).map(d => Number(d.deviceId));
  }

  let count = 0;
  for (const idStr of Object.keys(row.taskAction.deviceOverrides)) {
    if (activeIds.includes(Number(idStr))) {
      count++;
    }
  }

  return count;
};

const standardPresets = ['*/5 * * * *', '*/15 * * * *', '0 * * * *', '0 0 * * *', '0 8 * * *'];

// 1. Clean reset for custom cron inputs
const resetCustomCronState = () => {
  customTime.value = { hours: 12, minutes: 0 };
  customDays.value = [];
  isAdvancedCron.value = false;
  cronParts.value = { minute: '*', hour: '*', day: '*', month: '*', week: '*' };
};

// 2. Safely populate or clear cron states based on incoming data
const setupCronFromExpression = (cronExpression, scheduleType) => {
  resetCustomCronState();

  // If one-time or empty, fall back to default standard preset
  if (scheduleType !== 'recurring' || !cronExpression) {
    cronPreset.value = '*/5 * * * *';
    form.value.cronExpression = '*/5 * * * *';
    return;
  }

  // If it matches a standard dropdown option
  if (standardPresets.includes(cronExpression)) {
    cronPreset.value = cronExpression;
    return;
  }

  // It is a custom cron expression:
  cronPreset.value = 'custom';
  parseIntoCronParts(cronExpression);

  const parts = cronExpression.trim().split(/\s+/);
  if (parts.length >= 5) {

    // ⚡ UPDATED: Stricter logic to determine if the cron must use Advanced Mode
    const isComplex =
      !/^\d+$/.test(parts[0]) || // Minute is not a pure number (contains *, /, etc.)
      !/^\d+$/.test(parts[1]) || // Hour is not a pure number (contains *, /, etc.)
      parts[2] !== '*' ||        // Specific day of month is used
      parts[3] !== '*';          // Specific month is used

    if (isComplex) {
      isAdvancedCron.value = true;
    } else {
      isAdvancedCron.value = false;
      customTime.value = {
        minutes: parseInt(parts[0], 10) || 0,
        hours: parseInt(parts[1], 10) || 0
      };

      if (parts[4] !== '*' && parts[4] !== '?') {
        customDays.value = parts[4].split(',');
      } else {
        customDays.value = [];
      }
    }
  } else {
    isAdvancedCron.value = true;
  }
};

const openCreateModal = () => {
  modalMode.value = 'create';
  form.value = {
    scheduleId: null,
    targetType: 'device',
    deviceId: null,
    deviceGroupId: null,
    enableOverrides: false,
    taskActionPayload: { command: '', deviceOverrides: {} },
    scheduleType: 'one_time',
    status: 'active',
    startTime: null,
    endTime: null,
    cronExpression: '*/5 * * * *',
  };

  // Always reset to clean state on create
  setupCronFromExpression('*/5 * * * *', 'one_time');

  v$.value.$reset();
  isModalOpen.value = true;
};

const openEditModal = (schedule) => {
  modalMode.value = 'edit';
  const isGroup = schedule.deviceGroupId !== null && schedule.deviceGroupId !== undefined;

  let parsedAction = { command: '', deviceOverrides: {} };

  if (typeof schedule.taskAction === 'object' && schedule.taskAction !== null) {
    parsedAction = { ...parsedAction, ...schedule.taskAction };
  } else if (typeof schedule.taskAction === 'string') {
    parsedAction.command = schedule.taskAction;
  }

  if (!parsedAction.deviceOverrides) parsedAction.deviceOverrides = {};

  form.value = {
    scheduleId: schedule.scheduleId,
    targetType: isGroup ? 'group' : 'device',
    deviceId: schedule.deviceId || null,
    deviceGroupId: schedule.deviceGroupId || null,
    enableOverrides: Object.keys(parsedAction.deviceOverrides).length > 0,
    taskActionPayload: parsedAction,
    scheduleType: schedule.scheduleType,
    status: schedule.status,
    startTime: schedule.startTime ? new Date(schedule.startTime) : null,
    endTime: schedule.endTime ? new Date(schedule.endTime) : null,
    cronExpression: schedule.cronExpression || ''
  };

  if (isGroup && form.value.deviceGroupId) {
    const cleanOverrides = {};
    activeGroupDevices.value.forEach(d => {
      if (parsedAction.deviceOverrides[d.deviceId]) {
        cleanOverrides[d.deviceId] = parsedAction.deviceOverrides[d.deviceId];
      }
    });
    form.value.taskActionPayload.deviceOverrides = cleanOverrides;
    form.value.enableOverrides = Object.keys(cleanOverrides).length > 0;
  }

  // Populate from real data or reset cleanly if one_time / standard preset
  setupCronFromExpression(schedule.cronExpression, schedule.scheduleType);

  v$.value.$reset();
  isModalOpen.value = true;
};

const closeModal = () => {
  isModalOpen.value = false;
  resetCustomCronState();
};

const saveSchedule = async () => {
  const isFormValid = await v$.value.$validate();
  if (!isFormValid) return;

  if (!form.value.enableOverrides) {
    form.value.taskActionPayload.deviceOverrides = {};
  }

  isSaving.value = true;

  const payload = {
    deviceId: form.value.targetType === 'device' ? Number(form.value.deviceId) : null,
    deviceGroupId: form.value.targetType === 'group' ? Number(form.value.deviceGroupId) : null,
    taskAction: form.value.taskActionPayload,
    scheduleType: form.value.scheduleType,
    status: form.value.status,
    startTime: new Date(form.value.startTime).toISOString(),
  };

  if (form.value.scheduleType === 'recurring') {
    payload.cronExpression = form.value.cronExpression;
    if (form.value.endTime) {
      payload.endTime = new Date(form.value.endTime).toISOString();
    }
  }

  if (modalMode.value === 'create') {
    await createScheduleApi('/schedule/create', payload, 'POST');

    if (!createError.value) {
      toast.success(t('common.messages.created'));
      await loadData();
      closeModal();
    } else {
      toast.error(handleError(createError, 'common.messages.createError'));
    }
  } else {
    payload.scheduleId = form.value.scheduleId;
    await updateScheduleApi('/schedule/update', payload, 'PUT');

    if (!updateError.value) {
      toast.success(t('common.messages.updated'));
      await loadData();
      closeModal();
    } else {
      toast.error(handleError(updateError, 'common.messages.updateError'));
    }
  }

  isSaving.value = false;
};

const handleCronPresetChange = () => {
  if (cronPreset.value !== 'custom') {
    form.value.cronExpression = cronPreset.value;
  } else {
    if (isAdvancedCron.value) {
      updateFromCronParts();
    } else {
      updateCustomCronString();
    }
  }
};

const openDeleteModal = (schedule) => {
  scheduleToDelete.value = schedule;
  deleteModal.value.showModal();
};

const closeDeleteModal = () => {
  deleteModal.value.close();
  scheduleToDelete.value = null;
};

const confirmDelete = async () => {
  if (!scheduleToDelete.value) return;

  await deleteScheduleApi(`/schedule/delete/${scheduleToDelete.value.scheduleId}`, null, 'DELETE');

  if (!deleteError.value) {
    toast.success(t('common.messages.deleted') || 'Schedule deleted successfully');
    await loadData();
    closeDeleteModal();
  } else {
    toast.error(handleError(deleteError, 'common.messages.deleteError'));
  }
};

watch(() => form.value.scheduleType, (newType) => {
  if (newType === 'recurring' && !form.value.cronExpression) {
    setupCronFromExpression('*/5 * * * *', 'recurring');
  }
});

onMounted(async () => {
  if (!hasPermission(mainMenuName, 'Display')) return;
  await loadData();
});
</script>