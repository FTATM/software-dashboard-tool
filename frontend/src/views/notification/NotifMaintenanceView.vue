<template>
  <div v-if="hasPermission(mainMenuName, 'Display')" class="w-full h-full overflow-y-auto p-4">

    <!-- Page Header Card -->
    <div
      class="bg-base-100 shadow-sm rounded-box border border-base-200 p-6 mb-6 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
      <div class="flex items-center gap-4">
        <div class="p-3 bg-warning/15 text-warning rounded-xl flex items-center justify-center">
          <Icon icon="lucide:wrench" class="w-7 h-7" />
        </div>
        <div>
          <h2 class="m-0 text-2xl font-extrabold text-base-content tracking-tight">{{ $t('notifMaintenance.title') }}</h2>
          <p class="mt-1 mb-0 text-base-content/60 text-sm font-medium">{{ $t('notifMaintenance.subtitle') }}</p>
        </div>
      </div>
    </div>

    <!-- Data Table -->
    <div class="flex-1 flex flex-col min-h-0">
      <TableData :data="maintenanceTableData" :columns="tableColumns" :is-loading="isLoading"
        :initial-sorting="[{ id: 'maintenanceId', desc: false }]">
        <template #toolbar-actions>
          <button class="btn btn-primary shadow-sm hover:shadow-md transition-all" @click="openCreateModal">
            <Icon icon="lucide:plus" class="w-5 h-5" />
            {{ $t('notifMaintenance.addMaintenance') }}
          </button>
        </template>

        <template #cell-maintenanceId="{ value }">
          <span class="font-medium text-base-content/50">{{ value }}</span>
        </template>

        <!-- Target (Device or Group) -->
        <template #cell-target="{ row }">
          <div v-if="row.deviceGroupId" class="flex items-center gap-2">
            <div class="p-1.5 bg-secondary/10 rounded-md">
              <Icon icon="lucide:layers" class="w-4 h-4 text-secondary" />
            </div>
            <span :class="[
              'font-semibold',
              isUnknownGroup(row.deviceGroupId) ? 'text-base-content/50 italic' : 'text-base-content'
            ]">
              {{ getGroupName(row.deviceGroupId) }}
            </span>
          </div>
          <div v-else-if="row.deviceId" class="flex items-center gap-2">
            <div class="p-1.5 bg-primary/10 rounded-md">
              <Icon icon="lucide:cpu" class="w-4 h-4 text-primary" />
            </div>
            <span :class="[
              'font-bold',
              isUnknownDevice(row.deviceId) ? 'text-base-content/50 italic font-medium' : 'text-primary'
            ]">
              {{ getDeviceName(row.deviceId) }}
            </span>
          </div>
          <div v-else class="badge badge-error badge-sm text-white font-bold">
            {{ $t('scheduler.noTarget') }}
          </div>
        </template>

        <!-- Start Date (Hour precision: dd/MM/yyyy, HH:00) -->
        <template #cell-startDate="{ value }">
          <div class="flex items-center gap-1.5 text-sm font-mono">
            <Icon icon="lucide:calendar" class="w-4 h-4 text-base-content/50" />
            <span>{{ formatDateHour(value) }}</span>
          </div>
        </template>

        <!-- Maintenance Due Date (Hour precision: dd/MM/yyyy, HH:00) -->
        <template #cell-maintenanceDate="{ value, row }">
          <div class="flex flex-col gap-0.5">
            <div class="flex items-center gap-1.5 text-sm font-mono"
              :class="{
                'text-error font-bold': getStatusKey(row) === 'late',
                'text-info font-bold': getStatusKey(row) === 'maintenance',
                'text-warning font-semibold': getStatusKey(row) === 'near',
                'text-base-content': !['late', 'maintenance', 'near'].includes(getStatusKey(row))
              }">
              <Icon icon="lucide:calendar-clock" class="w-4 h-4"
                :class="{
                  'text-error': getStatusKey(row) === 'late',
                  'text-info': getStatusKey(row) === 'maintenance',
                  'text-warning': getStatusKey(row) === 'near',
                  'text-base-content/50': !['late', 'maintenance', 'near'].includes(getStatusKey(row))
                }" />
              <span>{{ formatDateHour(value) }}</span>
            </div>
            <span v-if="row.intervalDays" class="text-[11px] text-primary font-mono pl-5">
              ({{ $t('notifMaintenance.daysCount', { days: row.intervalDays }) }})
            </span>
          </div>
        </template>

        <!-- Daily Operating Count (Days & Hours) -->
        <template #cell-dayCount="{ value, row }">
          <div class="flex flex-col gap-0.5">
            <div class="flex items-center gap-1.5 font-mono bg-base-200 px-2.5 py-1 rounded-md w-fit">
              <Icon icon="lucide:timer" class="w-3.5 h-3.5 text-primary" />
              <span class="font-bold text-base-content">
                {{ $t('notifMaintenance.daysCount', { days: value ?? 0 }) }}
              </span>
              <span class="text-xs text-base-content/60">
                ({{ $t('notifMaintenance.hoursCount', { hours: (value ?? 0) * 24 }) }})
              </span>
            </div>
            <span class="text-[11px] text-base-content/50 font-mono pl-1">
              {{ $t('notifMaintenance.targetDaysLabel', { total: getTotalTargetDays(row) }) }}
            </span>
          </div>
        </template>

        <!-- Reason / Note -->
        <template #cell-reason="{ value }">
          <div class="truncate max-w-[200px] text-sm text-base-content/80" :title="value">
            {{ value || $t('notifMaintenance.noReason') }}
          </div>
        </template>

        <!-- Status Badge: Counting (Green), Near Maintenance (Yellow), Maintenance / Time Met (Blue), Late (Red), Disabled -->
        <template #cell-status="{ row }">
          <span v-if="getStatusKey(row) === 'disabled'"
            class="badge badge-sm font-semibold badge-ghost text-base-content/40">
            {{ $t('common.disabled') }}
          </span>
          <span v-else-if="getStatusKey(row) === 'late'"
            class="badge badge-sm font-bold badge-error text-white shadow-sm">
            {{ $t('notifMaintenance.statusLate') }}
          </span>
          <span v-else-if="getStatusKey(row) === 'maintenance'"
            class="badge badge-sm font-bold badge-info text-white shadow-sm">
            {{ $t('notifMaintenance.statusMaintenance') }}
          </span>
          <span v-else-if="getStatusKey(row) === 'near'"
            class="badge badge-sm font-semibold badge-warning text-white shadow-sm">
            {{ $t('notifMaintenance.statusNearMaintenance') }}
          </span>
          <span v-else class="badge badge-sm font-semibold badge-success text-white">
            {{ $t('notifMaintenance.statusCounting') }}
          </span>
        </template>

        <!-- Actions -->
        <template #cell-actions="{ row }">
          <div class="flex justify-end gap-2">
            <button @click="openEditModal(row)" class="btn btn-sm btn-primary">
              <Icon icon="lucide:pencil" class="w-5 h-5" />
            </button>
            <button @click="openDeleteModal(row)" class="btn btn-sm btn-error text-white">
              <Icon icon="lucide:trash-2" class="w-5 h-5" />
            </button>
          </div>
        </template>
      </TableData>
    </div>

    <!-- Create / Edit Modal -->
    <dialog ref="maintenanceModal" class="modal modal-bottom sm:modal-middle">
      <div class="modal-box sm:w-11/12 sm:max-w-xl p-0 flex flex-col max-h-[90vh] shadow-2xl">
        <div class="px-6 py-5 border-b border-base-200 bg-base-100 flex justify-between items-center shrink-0">
          <h3 class="m-0 text-xl font-extrabold text-base-content flex items-center gap-2">
            <Icon :icon="isEditing ? 'lucide:calendar-cog' : 'lucide:calendar-plus'" class="w-5 h-5 text-primary" />
            {{ isEditing ? $t('notifMaintenance.editMaintenance') : $t('notifMaintenance.createMaintenance') }}
          </h3>
          <button type="button" class="btn btn-sm btn-circle btn-ghost" @click="closeModal">
            <Icon icon="lucide:x" class="w-4 h-4" />
          </button>
        </div>

        <form @submit.prevent="submitForm" autocomplete="off" class="flex flex-col flex-1 min-h-0">
          <div class="p-6 bg-base-100 flex flex-col gap-4 flex-1 overflow-y-auto min-h-0">

            <!-- Target Type Radio (Device or Group) -->
            <div class="form-control w-full">
              <div class="label pb-1.5">
                <span class="label-text font-semibold">{{ $t('scheduler.targetType') }}</span>
              </div>
              <div class="flex gap-4">
                <label
                  class="label cursor-pointer justify-start gap-2 bg-base-200/50 p-2.5 rounded-lg border border-base-200 flex-1 hover:bg-base-200 transition-colors">
                  <input type="radio" value="device" class="radio radio-primary radio-sm" v-model="form.targetType"
                    @change="handleTargetTypeChange" />
                  <span class="label-text font-medium flex items-center gap-2">
                    <Icon icon="lucide:cpu" class="w-4 h-4" /> {{ $t('common.device') }}
                  </span>
                </label>
                <label
                  class="label cursor-pointer justify-start gap-2 bg-base-200/50 p-2.5 rounded-lg border border-base-200 flex-1 hover:bg-base-200 transition-colors">
                  <input type="radio" value="group" class="radio radio-primary radio-sm" v-model="form.targetType"
                    @change="handleTargetTypeChange" />
                  <span class="label-text font-medium flex items-center gap-2">
                    <Icon icon="lucide:layers" class="w-4 h-4" /> {{ $t('common.group') }}
                  </span>
                </label>
              </div>
            </div>

            <!-- Target Device or Group Dropdown -->
            <div class="form-control w-full relative">
              <div class="label pb-1">
                <span class="label-text font-semibold">
                  {{ form.targetType === 'device' ? $t('scheduler.selectTargetDevice') : $t('scheduler.selectTargetGroup') }}
                </span>
                <span class="label-text-alt text-error">*</span>
              </div>

              <template v-if="form.targetType === 'device'">
                <SearchableDropdown v-model="form.deviceId" :options="realDeviceList" labelKey="deviceName"
                  valueKey="deviceId" :placeholder="$t('common.searchDevice')" :error="v$.deviceId.$error"
                  @blur="v$.deviceId.$touch()" />
                <div v-if="v$.deviceId.$error" class="px-1 pt-1">
                  <span class="text-xs text-error font-medium">
                    {{ v$.deviceId.$errors[0].$message }}
                  </span>
                </div>
              </template>

              <template v-else>
                <SearchableDropdown v-model="form.deviceGroupId" :options="groupList" labelKey="groupName"
                  valueKey="groupId" :placeholder="$t('common.searchGroup')" :error="v$.deviceGroupId.$error"
                  @blur="v$.deviceGroupId.$touch()" />
                <div v-if="v$.deviceGroupId.$error" class="px-1 pt-1">
                  <span class="text-xs text-error font-medium">
                    {{ v$.deviceGroupId.$errors[0].$message }}
                  </span>
                </div>
              </template>
            </div>

            <!-- Maintenance Schedule Mode: Select by Date vs Select by Days -->
            <div class="form-control w-full">
              <div class="label pb-1.5">
                <span class="label-text font-semibold">{{ $t('notifMaintenance.dateSelectionMode') }}</span>
              </div>
              <div class="flex gap-4">
                <label
                  class="label cursor-pointer justify-start gap-2 bg-base-200/50 p-2.5 rounded-lg border border-base-200 flex-1 hover:bg-base-200 transition-colors">
                  <input type="radio" value="date" class="radio radio-primary radio-sm" v-model="form.dateMode"
                    @change="handleDateModeChange" />
                  <span class="label-text font-medium flex items-center gap-2">
                    <Icon icon="lucide:calendar" class="w-4 h-4" /> {{ $t('notifMaintenance.modeByDate') }}
                  </span>
                </label>
                <label
                  class="label cursor-pointer justify-start gap-2 bg-base-200/50 p-2.5 rounded-lg border border-base-200 flex-1 hover:bg-base-200 transition-colors">
                  <input type="radio" value="days" class="radio radio-primary radio-sm" v-model="form.dateMode"
                    @change="handleDateModeChange" />
                  <span class="label-text font-medium flex items-center gap-2">
                    <Icon icon="lucide:timer" class="w-4 h-4" /> {{ $t('notifMaintenance.modeByDays') }}
                  </span>
                </label>
              </div>
            </div>

            <!-- Start Date (Hour-only precision: dd/MM/yyyy HH:00, defaults to current hour) -->
            <div class="form-control w-full">
              <div class="label pb-1">
                <span class="label-text font-semibold">{{ $t('notifMaintenance.startDate') }}</span>
                <span class="label-text-alt text-error">*</span>
              </div>
              <VueDatePicker v-model="form.startDate" :is-24="true" :enable-time-picker="true"
                :enable-minutes="false" :no-minutes-overlay="true" :minutes-increment="60" :minutes-grid-increment="60"
                auto-apply :preset-dates="presetDates" :locale="dateFnsLocale" :format-locale="dateFnsLocale"
                :dark="themeStore.isDarkTheme" :formats="{ input: 'dd/MM/yyyy HH:00' }" :action-row="{
                  selectBtnLabel: $t('common.select'),
                  cancelBtnLabel: $t('common.cancel')
                }" :placeholder="$t('notifMaintenance.selectStartDate')" teleport-center
                @update:model-value="handleStartDateChange" @closed="v$.startDate.$touch()">
                <template #input-icon>
                  <Icon icon="lucide:calendar" class="w-5 h-5 ml-3 text-base-content/50" />
                </template>
              </VueDatePicker>
              <div v-if="v$.startDate.$error" class="px-1 pt-1">
                <span class="text-xs text-error font-medium">
                  {{ v$.startDate.$errors[0].$message }}
                </span>
              </div>
            </div>

            <!-- OPTION A: Select by Date (Hour-only precision: dd/MM/yyyy HH:00) -->
            <div v-if="form.dateMode === 'date'" class="form-control w-full">
              <div class="label pb-1">
                <span class="label-text font-semibold">{{ $t('notifMaintenance.maintenanceDate') }}</span>
                <span class="label-text-alt text-error">*</span>
              </div>
              <VueDatePicker v-model="form.maintenanceDate" :is-24="true" :enable-time-picker="true"
                :enable-minutes="false" :no-minutes-overlay="true" :minutes-increment="60" :minutes-grid-increment="60"
                auto-apply :min-date="form.startDate || undefined" :locale="dateFnsLocale" :format-locale="dateFnsLocale"
                :dark="themeStore.isDarkTheme" :formats="{ input: 'dd/MM/yyyy HH:00' }" :action-row="{
                  selectBtnLabel: $t('common.select'),
                  cancelBtnLabel: $t('common.cancel')
                }" :placeholder="$t('notifMaintenance.selectMaintenanceDate')" teleport-center
                @update:model-value="handleMaintenanceDateChange" @closed="v$.maintenanceDate.$touch()">
                <template #input-icon>
                  <Icon icon="lucide:calendar-clock" class="w-5 h-5 ml-3 text-base-content/50" />
                </template>
              </VueDatePicker>
              <div v-if="v$.maintenanceDate.$error" class="px-1 pt-1">
                <span class="text-xs text-error font-medium">
                  {{ v$.maintenanceDate.$errors[0].$message }}
                </span>
              </div>
            </div>

            <!-- OPTION B: Select by Days (e.g. in 180 days at selected hour HH:00) -->
            <div v-else class="p-4 bg-base-200/50 rounded-xl border border-base-200 flex flex-col gap-3">
              <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
                <label class="form-control w-full">
                  <div class="label pb-1">
                    <span class="label-text font-semibold">{{ $t('notifMaintenance.intervalDays') }}</span>
                    <span class="label-text-alt text-error">*</span>
                  </div>
                  <input type="number" min="1" step="1" v-model.number="form.intervalDays"
                    @input="recalculateMaintenanceDateFromDays" @blur="v$.intervalDays.$touch()"
                    :placeholder="$t('notifMaintenance.intervalDaysPlaceholder')"
                    :class="['input input-bordered w-full font-mono', { 'input-error': v$.intervalDays.$error }]" />
                  <div v-if="v$.intervalDays.$error" class="px-1 pt-1">
                    <span class="text-xs text-error font-medium">
                      {{ v$.intervalDays.$errors[0].$message }}
                    </span>
                  </div>
                </label>

                <!-- Hour-only selector (00:00 - 23:00, no minutes) -->
                <label class="form-control w-full">
                  <div class="label pb-1">
                    <span class="label-text font-semibold">{{ $t('notifMaintenance.alertTime') }}</span>
                  </div>
                  <select v-model.number="form.alertHour" @change="recalculateMaintenanceDateFromDays"
                    class="select select-bordered w-full font-mono">
                    <option v-for="h in 24" :key="h - 1" :value="h - 1">
                      {{ String(h - 1).padStart(2, '0') }}:00
                    </option>
                  </select>
                </label>
              </div>

              <!-- Quick Day Presets -->
              <div class="flex flex-wrap gap-1.5">
                <button v-for="preset in [30, 60, 90, 180, 365]" :key="preset" type="button"
                  class="btn btn-xs font-mono"
                  :class="Number(form.intervalDays) === preset ? 'btn-primary text-white' : 'btn-outline border-base-300'"
                  @click="setDayPreset(preset)">
                  +{{ preset }}d
                </button>
              </div>

              <!-- Calculated Maintenance Date Preview + Reset Cycle Button in Edit Mode -->
              <div
                class="p-3 bg-base-100 rounded-lg border border-base-200 flex flex-col sm:flex-row sm:items-center justify-between gap-2">
                <div>
                  <span class="text-xs text-base-content/60 font-semibold block">
                    {{ $t('notifMaintenance.calculatedDueDate') }}
                  </span>
                  <span class="font-mono font-bold text-primary text-sm">
                    {{ form.maintenanceDate ? formatDateHour(form.maintenanceDate) : '-' }}
                  </span>
                </div>

                <!-- Reset Button in Edit Mode when Select by Days is active -->
                <button v-if="isEditing && Number(form.intervalDays) > 0" type="button"
                  class="btn btn-sm btn-warning text-white gap-1.5 shrink-0" @click="handleResetCycle">
                  <Icon icon="lucide:rotate-ccw" class="w-4 h-4" />
                  {{ $t('notifMaintenance.resetCycle', { days: form.intervalDays || 180 }) }}
                </button>
              </div>
            </div>

            <!-- Reason / Maintenance Note -->
            <div class="form-control w-full">
              <div class="label pb-1 flex justify-between">
                <span class="label-text font-semibold">{{ $t('notifMaintenance.maintenanceNote') }}</span>
                <span class="label-text-alt text-base-content/60 font-mono">{{ form.reason?.length || 0 }}/100</span>
              </div>
              <input type="text" v-model="form.reason" maxlength="100" @blur="v$.reason.$touch()"
                :placeholder="$t('notifMaintenance.maintenanceNotePlaceholder')"
                :class="['input input-bordered w-full', { 'input-error': v$.reason.$error }]" />
              <p class="m-0 mt-1.5 px-1 text-xs text-base-content/60 whitespace-normal break-words leading-relaxed">
                {{ $t('notifMaintenance.maintenanceNoteDesc') }}
              </p>
              <div v-if="v$.reason.$error" class="px-1 pt-1">
                <span class="text-xs text-error font-medium">
                  {{ v$.reason.$errors[0].$message }}
                </span>
              </div>
            </div>

            <!-- Active Toggle -->
            <div class="p-4 bg-base-200/40 rounded-xl border border-base-200 flex items-center justify-between gap-3">
              <div class="min-w-0">
                <p class="font-bold text-base-content m-0 text-sm">{{ $t('notifMaintenance.maintenanceStatus') }}</p>
                <p class="text-xs text-base-content/60 m-0 mt-1 whitespace-normal break-words">
                  {{ $t('notifMaintenance.maintenanceStatusDesc') }}
                </p>
              </div>
              <input type="checkbox" v-model="form.active" class="toggle toggle-success toggle-md shrink-0" />
            </div>

          </div>

          <!-- Modal Actions -->
          <div class="px-6 py-4 border-t border-base-200 bg-base-100 flex justify-end gap-3 shrink-0">
            <button type="button" class="btn btn-ghost" @click="closeModal" :disabled="isSaving">
              {{ $t('common.cancel') }}
            </button>
            <button type="submit" class="btn btn-primary px-8" :disabled="isSaving">
              <span v-if="isSaving" class="loading loading-spinner loading-sm"></span>
              {{ isEditing ? $t('common.save') : $t('notifMaintenance.createMaintenance') }}
            </button>
          </div>
        </form>
      </div>
      <form method="dialog" class="modal-backdrop"><button @click="closeModal">close</button></form>
    </dialog>

    <!-- Delete Confirmation Modal -->
    <dialog ref="deleteModal" class="modal z-[200]">
      <div class="modal-box">
        <h3 class="font-bold text-lg text-error flex items-center gap-2">
          <Icon icon="lucide:alert-triangle" class="w-6 h-6" /> {{ $t('common.confirmDelete') }}
        </h3>
        <p class="py-4 text-base-content/80">
          {{ $t('notifMaintenance.deleteWarning', { name: itemToDelete ? getTargetDisplayName(itemToDelete) : '' }) }}
        </p>
        <div class="modal-action">
          <button type="button" @click="closeDeleteModal" class="btn btn-ghost" :disabled="isDeleting">
            {{ $t('common.cancel') }}
          </button>
          <button type="button" @click="confirmDelete" class="btn btn-error text-white" :disabled="isDeleting">
            <span v-if="isDeleting" class="loading loading-spinner loading-sm"></span> {{ $t('common.delete') }}
          </button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop"><button @click="closeDeleteModal">close</button></form>
    </dialog>

  </div>
  <NoAccess v-else />
</template>

<script setup>
import { ref, onMounted, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { useMutation } from '@/composables/useMutation';
import { useFetch } from '@/composables/useFetch';
import { useVuelidate } from '@vuelidate/core';
import { required, requiredIf, maxLength, helpers } from '@vuelidate/validators';
import { toast } from 'vue3-toastify';
import { Icon } from '@iconify/vue';
import { VueDatePicker } from '@vuepic/vue-datepicker';
import thLocale from 'date-fns/locale/th';
import enLocale from 'date-fns/locale/en-US';
import { usePermissionStore } from '@/stores/usePermissionStore';
import { useThemeStore } from '@/stores/useThemeStore';
import NoAccess from '@/components/NoAccess.vue';
import SearchableDropdown from '@/components/SearchableDropdown.vue';
import TableData from '@/components/TableData.vue';
import { useErrorHandler } from '@/composables/useErrorHandler';

const { handleError } = useErrorHandler();
const themeStore = useThemeStore();
const { t, te, locale } = useI18n();
const mainMenuName = 'Notification Maintenance';

const permissionStore = usePermissionStore();
const { hasPermission } = permissionStore;

const { error: saveError, isLoading: isSaving, execute: saveApi } = useMutation();
const { data: maintenanceData, isLoading, error: fetchError, execute: fetchApi } = useFetch();
const { error: deleteError, isLoading: isDeleting, execute: deleteApi } = useMutation();
const { data: fetchDevices, execute: fetchDevicesApi } = useFetch();
const { data: fetchGroups, execute: fetchGroupsApi } = useFetch();

const maintenanceModal = ref(null);
const deleteModal = ref(null);
const isEditing = ref(false);
const editingMaintenanceId = ref(null);
const itemToDelete = ref(null);

const maintenanceTableData = ref([]);
const deviceList = ref([]);
const groupList = ref([]);

const snapToHour = (d) => {
  if (!d) return null;
  const dt = new Date(d);
  dt.setMinutes(0, 0, 0);
  return dt;
};

const presetDates = ref([{ label: t('common.today'), value: snapToHour(new Date()) }]);

const resolveLocale = (loc) => {
  if (!loc) return undefined;
  if (loc.localize) return loc;
  if (loc.default?.localize) return loc.default;
  if (loc.th?.localize) return loc.th;
  if (loc.enUS?.localize) return loc.enUS;
  return loc.default || loc;
};

const dateFnsLocale = computed(() => {
  const isThai = locale.value?.startsWith('th');
  return resolveLocale(isThai ? thLocale : enLocale);
});

const isDeviceActive = (device) => {
  if (!device) return false;
  if (device.isDeleted || device.deleted || device.isDelete) return false;
  if (device.active === false || device.isActive === false) return false;
  if (typeof device.status === 'string' && device.status.toLowerCase() === 'inactive') return false;
  return true;
};

const isUnknownDevice = (deviceId) => {
  if (!deviceId) return true;
  const device = deviceList.value.find(d => d.deviceId === deviceId);
  return !device || !isDeviceActive(device);
};

const isUnknownGroup = (groupId) => {
  if (!groupId) return true;
  const group = groupList.value.find(g => g.groupId === groupId);
  return !group;
};

const getUnknownDeviceText = () => {
  if (te && te('notifDevice.unknownDevice')) return t('notifDevice.unknownDevice');
  if (te && te('common.unknownDevice')) return t('common.unknownDevice');
  return t('scheduler.unknownDevice');
};

const getDeviceName = (deviceId) => {
  if (!deviceId) return getUnknownDeviceText();
  const device = deviceList.value.find(d => d.deviceId === deviceId);
  if (device && isDeviceActive(device)) {
    return device.deviceName;
  }
  return getUnknownDeviceText();
};

const getGroupName = (groupId) => {
  if (!groupId) return t('scheduler.unknownGroup');
  const group = groupList.value.find(g => g.groupId === groupId);
  return group ? group.groupName : t('scheduler.unknownGroup');
};

const getTargetDisplayName = (row) => {
  if (!row) return '';
  if (row.deviceGroupId) return getGroupName(row.deviceGroupId);
  if (row.deviceId) return getDeviceName(row.deviceId);
  return '';
};

const realDeviceList = computed(() => {
  const activePhysicalList = deviceList.value.filter(d => !d.refDeviceId && isDeviceActive(d));
  if (isEditing.value && form.value.deviceId && !activePhysicalList.some(d => d.deviceId === form.value.deviceId)) {
    return [
      { deviceId: form.value.deviceId, deviceName: getDeviceName(form.value.deviceId) },
      ...activePhysicalList
    ];
  }
  return activePhysicalList;
});

// Display date + whole hour (dd/MM/yyyy, HH:00)
const formatDateHour = (ts) => {
  if (!ts) return '-';
  const date = new Date(ts);
  const currentLocale = locale.value === 'th' ? 'th-TH' : 'en-GB';
  const datePart = date.toLocaleDateString(currentLocale, {
    calendar: 'gregory',
    day: '2-digit',
    month: '2-digit',
    year: 'numeric'
  });
  const hourPart = String(date.getHours()).padStart(2, '0');
  return `${datePart}, ${hourPart}:00`;
};

const toDateFloor = (d) => {
  if (!d) return null;
  const dt = new Date(d);
  return new Date(dt.getFullYear(), dt.getMonth(), dt.getDate()).getTime();
};

const getTotalTargetDays = (row) => {
  if (row?.intervalDays && row.intervalDays > 0) return row.intervalDays;
  if (!row?.startDate || !row?.maintenanceDate) return 0;
  const startMs = toDateFloor(row.startDate);
  const maintMs = toDateFloor(row.maintenanceDate);
  const diffDays = Math.round((maintMs - startMs) / (1000 * 60 * 60 * 24));
  return Math.max(0, diffDays);
};

// Status Evaluation: disabled -> late -> maintenance -> near -> counting
const getStatusKey = (row) => {
  if (!row?.active) return 'disabled';
  if (!row?.maintenanceDate) return 'counting';

  const nowMs = Date.now();
  const maintDt = snapToHour(row.maintenanceDate);
  const maintMs = maintDt.getTime();
  const todayFloor = toDateFloor(new Date());
  const maintFloor = toDateFloor(row.maintenanceDate);
  const dayMs = 24 * 60 * 60 * 1000;

  // 1. Late: days have passed beyond the maintenance date (or >= 24h after maintenance timestamp)
  if (todayFloor > maintFloor || nowMs >= maintMs + dayMs) {
    return 'late';
  }

  // 2. Maintenance: maintenance date/hour met (or on the maintenance date after notification fired)
  if (nowMs >= maintMs || (todayFloor === maintFloor && row.notified)) {
    return 'maintenance';
  }

  // 3. Near Maintenance: within 7 days of maintenance date (or on the maintenance day before target hour)
  const remainingDays = Math.ceil((maintFloor - todayFloor) / dayMs);
  if (remainingDays <= 7) {
    return 'near';
  }

  // 4. Counting: normal operation
  return 'counting';
};

const tableColumns = computed(() => [
  { header: t('common.id'), accessorKey: 'maintenanceId', meta: { headerClass: 'w-16', cellClass: 'font-bold' } },
  {
    header: t('scheduler.table.target'),
    id: 'target',
    accessorFn: (row) => getTargetDisplayName(row)
  },
  {
    header: t('notifMaintenance.startDate'),
    id: 'startDate',
    accessorKey: 'startDate'
  },
  {
    header: t('notifMaintenance.maintenanceDate'),
    id: 'maintenanceDate',
    accessorKey: 'maintenanceDate'
  },
  {
    header: t('notifMaintenance.operatingCount'),
    id: 'dayCount',
    accessorKey: 'dayCount'
  },
  { header: t('notifMaintenance.maintenanceNote'), accessorKey: 'reason' },
  {
    header: t('common.status'),
    id: 'status',
    accessorFn: (row) => {
      const key = getStatusKey(row);
      if (key === 'disabled') return t('common.disabled');
      if (key === 'late') return t('notifMaintenance.statusLate');
      if (key === 'maintenance') return t('notifMaintenance.statusMaintenance');
      if (key === 'near') return t('notifMaintenance.statusNearMaintenance');
      return t('notifMaintenance.statusCounting');
    }
  },
  { header: t('common.actions'), id: 'actions', enableSorting: false, meta: { headerClass: 'text-right w-28', cellClass: 'text-right' } }
]);

const nowInit = snapToHour(new Date());
const form = ref({
  targetType: 'device',
  deviceId: null,
  deviceGroupId: null,
  dateMode: 'date', // 'date' | 'days'
  intervalDays: 180,
  alertHour: nowInit.getHours(),
  startDate: nowInit,
  maintenanceDate: null,
  reason: '',
  active: true
});

const recalculateMaintenanceDateFromDays = () => {
  if (form.value.dateMode !== 'days') return;
  const days = Number(form.value.intervalDays);
  if (!days || days < 1 || !form.value.startDate) {
    form.value.maintenanceDate = null;
    return;
  }
  const base = snapToHour(form.value.startDate);
  const due = new Date(base);
  due.setDate(due.getDate() + days);
  due.setHours(Number(form.value.alertHour) || 0, 0, 0, 0);
  form.value.maintenanceDate = due;
};

const handleDateModeChange = () => {
  if (form.value.dateMode === 'days') {
    if (!form.value.intervalDays || form.value.intervalDays < 1) {
      form.value.intervalDays = 180;
    }
    const refTime = form.value.maintenanceDate ? new Date(form.value.maintenanceDate) : new Date(form.value.startDate || new Date());
    form.value.alertHour = refTime.getHours();
    recalculateMaintenanceDateFromDays();
  }
  v$.value.$reset();
};

const handleStartDateChange = (val) => {
  if (val) {
    form.value.startDate = snapToHour(val);
  }
  if (form.value.dateMode === 'days') {
    recalculateMaintenanceDateFromDays();
  }
};

const handleMaintenanceDateChange = (val) => {
  if (val) {
    form.value.maintenanceDate = snapToHour(val);
  }
};

const setDayPreset = (days) => {
  form.value.intervalDays = days;
  recalculateMaintenanceDateFromDays();
};

const handleResetCycle = () => {
  const now = snapToHour(new Date());
  form.value.startDate = now;
  recalculateMaintenanceDateFromDays();
  toast.info(t('notifMaintenance.cycleResetToast', { days: form.value.intervalDays || 180 }));
};

const isMaintenanceAfterOrEqualStart = (value) => {
  if (!value || !form.value.startDate) return true;
  return snapToHour(value).getTime() >= snapToHour(form.value.startDate).getTime();
};

const rules = computed(() => ({
  deviceId: {
    requiredIfDevice: helpers.withMessage(
      t('scheduler.validation.selectDevice'),
      requiredIf(() => form.value.targetType === 'device')
    )
  },
  deviceGroupId: {
    requiredIfGroup: helpers.withMessage(
      t('scheduler.validation.selectGroup'),
      requiredIf(() => form.value.targetType === 'group')
    )
  },
  startDate: {
    required: helpers.withMessage(t('notifMaintenance.validation.startDateRequired'), required)
  },
  intervalDays: {
    requiredIfDays: helpers.withMessage(
      t('notifMaintenance.validation.intervalDaysRequired'),
      requiredIf(() => form.value.dateMode === 'days')
    ),
    minValue: helpers.withMessage(
      t('notifMaintenance.validation.intervalDaysMin'),
      (val) => form.value.dateMode !== 'days' || (Number(val) >= 1)
    )
  },
  maintenanceDate: {
    required: helpers.withMessage(t('notifMaintenance.validation.maintenanceDateRequired'), required),
    afterOrEqualStart: helpers.withMessage(
      t('notifMaintenance.validation.maintenanceAfterStart'),
      isMaintenanceAfterOrEqualStart
    )
  },
  reason: {
    maxLength: helpers.withMessage(t('common.validation.maxLength', { len: 100 }), maxLength(100))
  }
}));

const v$ = useVuelidate(rules, form);

const handleTargetTypeChange = () => {
  if (form.value.targetType === 'device') {
    form.value.deviceGroupId = null;
  } else {
    form.value.deviceId = null;
  }
  v$.value.$reset();
};

const loadTable = async () => {
  await fetchApi('/notification/maintenance/getalldetail');
  if (!fetchError.value && maintenanceData.value) {
    maintenanceTableData.value = maintenanceData.value.data || [];
  } else {
    toast.error(fetchError.value?.message || t('common.messages.loadError'));
  }
};

const loadDevicesAndGroups = async () => {
  await fetchDevicesApi('/device/getalldetail');
  if (fetchDevices.value) {
    deviceList.value = fetchDevices.value.data || [];
  }
  await fetchGroupsApi('/device/group/getalldetail');
  if (fetchGroups.value) {
    groupList.value = fetchGroups.value.data || [];
  }
};

const openCreateModal = () => {
  isEditing.value = false;
  editingMaintenanceId.value = null;
  const now = snapToHour(new Date());
  form.value = {
    targetType: 'device',
    deviceId: null,
    deviceGroupId: null,
    dateMode: 'date',
    intervalDays: 180,
    alertHour: now.getHours(),
    startDate: now,
    maintenanceDate: null,
    reason: '',
    active: true
  };
  v$.value.$reset();
  maintenanceModal.value.showModal();
};

const openEditModal = (row) => {
  isEditing.value = true;
  editingMaintenanceId.value = row.maintenanceId;
  const isGroup = row.deviceGroupId !== null && row.deviceGroupId !== undefined;
  const hasInterval = row.intervalDays !== null && row.intervalDays !== undefined && Number(row.intervalDays) > 0;
  const maintDt = row.maintenanceDate ? snapToHour(row.maintenanceDate) : snapToHour(new Date());

  form.value = {
    targetType: isGroup ? 'group' : 'device',
    deviceId: row.deviceId || null,
    deviceGroupId: row.deviceGroupId || null,
    dateMode: hasInterval ? 'days' : 'date',
    intervalDays: hasInterval ? Number(row.intervalDays) : 180,
    alertHour: maintDt.getHours(),
    startDate: row.startDate ? snapToHour(row.startDate) : snapToHour(new Date()),
    maintenanceDate: row.maintenanceDate ? snapToHour(row.maintenanceDate) : null,
    reason: row.reason || '',
    active: row.active
  };
  v$.value.$reset();
  maintenanceModal.value.showModal();
};

const closeModal = () => maintenanceModal.value.close();
const openDeleteModal = (row) => { itemToDelete.value = row; deleteModal.value.showModal(); };
const closeDeleteModal = () => { deleteModal.value.close(); itemToDelete.value = null; };

const submitForm = async () => {
  form.value.startDate = snapToHour(form.value.startDate);
  if (form.value.dateMode === 'days') {
    recalculateMaintenanceDateFromDays();
  } else if (form.value.maintenanceDate) {
    form.value.maintenanceDate = snapToHour(form.value.maintenanceDate);
  }

  const isFormValid = await v$.value.$validate();
  if (!isFormValid) return;

  const payload = {
    deviceId: form.value.targetType === 'device' ? Number(form.value.deviceId) : null,
    deviceGroupId: form.value.targetType === 'group' ? Number(form.value.deviceGroupId) : null,
    startDate: snapToHour(form.value.startDate).toISOString(),
    maintenanceDate: snapToHour(form.value.maintenanceDate).toISOString(),
    intervalDays: form.value.dateMode === 'days' ? Number(form.value.intervalDays) : null,
    reason: form.value.reason || '',
    active: form.value.active
  };

  if (isEditing.value) {
    payload.maintenanceId = editingMaintenanceId.value;
    await saveApi('/notification/maintenance/update', payload, 'PUT');
  } else {
    await saveApi('/notification/maintenance/create', payload, 'POST');
  }

  if (!saveError.value) {
    toast.success(isEditing.value ? t('common.messages.updated') : t('common.messages.created'));
    await loadTable();
    closeModal();
  } else {
    toast.error(handleError(saveError, 'common.messages.saveError'));
  }
};

const confirmDelete = async () => {
  if (!itemToDelete.value) return;
  await deleteApi(`/notification/maintenance/delete/${itemToDelete.value.maintenanceId}`, null, 'DELETE');
  if (!deleteError.value) {
    toast.success(t('common.messages.deleted'));
    await loadTable();
    closeDeleteModal();
  } else {
    toast.error(handleError(deleteError, 'common.messages.deleteError'));
  }
};

onMounted(async () => {
  if (!hasPermission(mainMenuName, 'Display')) return;
  await loadDevicesAndGroups();
  await loadTable();
});
</script>
