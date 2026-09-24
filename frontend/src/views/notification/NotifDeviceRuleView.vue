<template>
  <div v-if="hasPermission(mainMenuName, 'Display')" class="w-full h-full overflow-y-auto p-4 sm:p-6 space-y-5">

    <!-- Anchored Page Header Card with Background -->
    <div class="bg-base-100 border border-base-300 rounded-2xl p-5 shadow-xs flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
      <div class="flex items-start sm:items-center gap-3.5">
        <div class="p-3 bg-error/10 text-error rounded-xl flex items-center justify-center shrink-0">
          <Icon icon="lucide:siren" class="w-6 h-6" />
        </div>
        <div>
          <!-- Breadcrumbs -->
          <div class="flex items-center gap-1.5 text-xs font-semibold text-base-content/50 uppercase tracking-wider mb-0.5">
            <span>{{ $t('menu.notification') }}</span>
            <Icon icon="lucide:chevron-right" class="w-3.5 h-3.5" />
            <span class="text-primary">{{ $t('notifDevice.title') }}</span>
          </div>

          <!-- Title & Subtitle -->
          <div class="flex items-center gap-2.5">
            <h1 class="m-0 text-xl sm:text-2xl font-black text-base-content tracking-tight">
              {{ $t('notifDevice.title') }}
            </h1>
          </div>
          <p class="mt-0.5 mb-0 text-base-content/60 text-xs font-medium">
            {{ $t('notifDevice.subtitle') }}
          </p>
        </div>
      </div>

      <!-- Header Actions -->
      <div class="flex items-center gap-2 self-end sm:self-center">
        <button 
          @click="loadTable" 
          class="btn btn-sm btn-ghost border border-base-300 bg-base-100 hover:bg-base-200 rounded-xl gap-1.5 text-xs font-semibold shadow-xs transition-all"
          :title="$t('common.refresh')">
          <Icon icon="lucide:refresh-cw" class="w-3.5 h-3.5" :class="{ 'animate-spin': isLoading }" />
          <span>{{ $t('common.refresh') }}</span>
        </button>
      </div>
    </div>

    <!-- Reusable KPI Summary Status Cards -->
    <StatCardGroup :items="statCardsData" />

    <!-- Main Table Card -->
    <div class="bg-base-100 border border-base-300 rounded-2xl p-4 sm:p-5 shadow-xs">
      <TableData 
        :data="ruleTableData" 
        :columns="tableColumns" 
        :is-loading="isLoading"
        :initial-sorting="[{ id: 'ruleId', desc: false }]">
        
        <!-- Toolbar Actions -->
        <template #toolbar-actions>
          <button 
            class="btn btn-sm btn-primary rounded-xl font-semibold shadow-xs hover:shadow-md transition-all gap-1 text-white" 
            @click="openCreateModal">
            <Icon icon="lucide:plus" class="w-4 h-4" />
            {{ $t('notifDevice.addRule') }}
          </button>
        </template>

        <!-- ID Cell -->
        <template #cell-ruleId="{ value }">
          <span class="font-mono text-xs font-bold text-base-content/50">#{{ value }}</span>
        </template>

        <!-- Target Device Name -->
        <template #cell-deviceName="{ row }">
          <div class="flex items-center gap-2">
            <div class="w-7 h-7 rounded-lg bg-primary/10 text-primary flex items-center justify-center shrink-0">
              <Icon icon="lucide:cpu" class="w-3.5 h-3.5" />
            </div>
            <span :class="[
              'font-semibold text-sm tracking-tight',
              isUnknownDevice(row.deviceId) ? 'text-base-content/40 italic' : 'text-base-content'
            ]">
              {{ getDeviceName(row.deviceId) }}
            </span>
          </div>
        </template>

        <!-- Reason / Alert Message -->
        <template #cell-reason="{ value }">
          <div class="truncate max-w-[220px] text-xs text-base-content/80 font-medium" :title="value">
            {{ value || $t('notifDevice.noReason') }}
          </div>
        </template>

        <!-- Logic Condition Pill -->
        <template #cell-logic="{ row }">
          <div class="inline-flex items-center gap-1.5 font-mono bg-base-200 px-2.5 py-1 rounded-lg border border-base-300 text-xs">
            <span class="text-base-content/60 text-[10px] uppercase font-bold">{{ $t('notifDevice.ifValue') }}</span>
            <span class="font-bold text-error">{{ row.condition }}</span>
            <span class="font-black text-base-content">{{ row.threshold }}</span>
          </div>
        </template>

        <!-- Active Status Live Pill -->
        <template #cell-active="{ value }">
          <div class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold"
            :class="value ? 'bg-success/10 text-success border border-success/20' : 'bg-base-200 text-base-content/50 border border-base-300'">
            <span class="w-1.5 h-1.5 rounded-full" :class="value ? 'bg-success animate-pulse' : 'bg-base-content/30'"></span>
            {{ value ? $t('common.active') : $t('common.disabled') }}
          </div>
        </template>

        <!-- Action Buttons -->
        <template #cell-actions="{ row }">
          <div class="flex justify-end items-center gap-1">
            <button 
              @click="openEditModal(row)" 
              class="btn btn-ghost btn-xs btn-square rounded-lg text-base-content/70 hover:text-primary hover:bg-primary/10 transition-colors"
              :title="$t('common.edit')">
              <Icon icon="lucide:pencil" class="w-4 h-4" />
            </button>
            <button 
              @click="openDeleteModal(row)" 
              class="btn btn-ghost btn-xs btn-square rounded-lg text-base-content/70 hover:text-error hover:bg-error/10 transition-colors"
              :title="$t('common.delete')">
              <Icon icon="lucide:trash-2" class="w-4 h-4" />
            </button>
          </div>
        </template>
      </TableData>
    </div>

    <!-- Create / Edit Modal with Scrollable Body -->
    <dialog ref="ruleModal" class="modal">
      <div class="modal-box sm:w-11/12 sm:max-w-xl p-0 overflow-hidden shadow-2xl rounded-2xl flex flex-col max-h-[85vh] border border-base-300 bg-base-100">
        <!-- Pinned Header -->
        <div class="px-6 py-4 border-b border-base-200 bg-base-100 flex justify-between items-center shrink-0">
          <div class="flex items-center gap-2.5">
            <div class="p-2 rounded-xl bg-primary/10 text-primary">
              <Icon :icon="isEditing ? 'lucide:pencil' : 'lucide:bell-plus'" class="w-5 h-5" />
            </div>
            <div>
              <h3 class="m-0 text-lg font-bold text-base-content">
                {{ isEditing ? $t('notifDevice.editRule') : $t('notifDevice.createRule') }}
              </h3>
              <p class="m-0 text-xs text-base-content/50">
                {{ isEditing ? $t('notifDevice.editSubtitle') : $t('notifDevice.createSubtitle') }}
              </p>
            </div>
          </div>
          <button class="btn btn-sm btn-circle btn-ghost" @click="closeModal">
            <Icon icon="lucide:x" class="w-4 h-4" />
          </button>
        </div>

        <!-- Form Wrapper -->
        <form @submit.prevent="submitForm" autocomplete="off" class="flex flex-col flex-1 overflow-hidden bg-base-100">
          <div class="p-6 overflow-y-auto flex-1 flex flex-col gap-4">

            <!-- Target Device -->
            <label class="form-control w-full relative">
              <div class="label pb-1">
                <span class="label-text font-semibold text-xs uppercase tracking-wider text-base-content/70">
                  {{ $t('notifDevice.targetDevice') }}
                  <span class="text-error ml-0.5">*</span>
                </span>
              </div>
              <SearchableDropdown v-model="form.deviceId" :options="realDeviceList" labelKey="deviceName"
                valueKey="deviceId" :placeholder="$t('common.searchDevice')" :error="v$.deviceId.$error"
                @blur="v$.deviceId.$touch()" />
              <div class="label px-1 py-0.5 min-h-[20px]">
                <span v-if="v$.deviceId.$error" class="label-text-alt text-error font-medium text-xs">
                  {{ v$.deviceId.$errors[0].$message }}
                </span>
              </div>
            </label>

            <!-- Condition & Threshold -->
            <div class="grid grid-cols-2 gap-3">
              <label class="form-control w-full">
                <div class="label pb-1">
                  <span class="label-text font-semibold text-xs uppercase tracking-wider text-base-content/70">{{ $t('notifDevice.condition') }}</span>
                </div>
                <select v-model="form.condition" @blur="v$.condition.$touch()"
                  class="select select-sm h-10 select-bordered w-full rounded-xl font-mono text-base font-bold">
                  <option value=">">&gt;</option>
                  <option value=">=">&gt;=</option>
                  <option value="==">==</option>
                  <option value="!=">!=</option>
                  <option value="<">&lt;</option>
                  <option value="<=">&lt;=</option>
                </select>
              </label>

              <label class="form-control w-full">
                <div class="label pb-1">
                  <span class="label-text font-semibold text-xs uppercase tracking-wider text-base-content/70">{{ $t('notifDevice.thresholdValue') }}</span>
                </div>
                <input type="number" step="any" v-model="form.threshold" @blur="v$.threshold.$touch()"
                  :placeholder="$t('notifDevice.thresholdPlaceholder')"
                  class="input input-sm h-10 input-bordered w-full rounded-xl font-mono text-base font-medium"
                  :class="{ 'input-error': v$.threshold.$error }" />
                <div class="label px-1 py-0.5 min-h-[20px]">
                  <span v-if="v$.threshold.$error" class="label-text-alt text-error font-medium text-xs">
                    {{ v$.threshold.$errors[0].$message }}
                  </span>
                </div>
              </label>
            </div>

            <!-- Alert Message / Reason -->
            <label class="form-control w-full">
              <div class="label pb-1 flex justify-between">
                <span class="label-text font-semibold text-xs uppercase tracking-wider text-base-content/70">{{ $t('notifDevice.alertMessageReason') }}</span>
                <span class="label-text-alt text-base-content/50 font-mono text-[11px]">{{ form.reason?.length || 0 }}/100</span>
              </div>
              <input type="text" v-model="form.reason" maxlength="100" @blur="v$.reason.$touch()"
                :placeholder="$t('notifDevice.alertMessagePlaceholder')"
                :class="['input input-sm h-10 input-bordered w-full rounded-xl', { 'input-error': v$.reason.$error }]" />
              <div class="label px-1 py-1 flex-col items-start gap-1">
                <span class="label-text-alt text-base-content/50 text-[11px]">{{ $t('notifDevice.alertMessageDesc') }}</span>
                <span v-if="v$.reason.$error" class="label-text-alt text-error font-medium text-xs">
                  {{ v$.reason.$errors[0].$message }}
                </span>
              </div>
            </label>

            <!-- Status Box -->
            <div class="p-4 bg-base-200/50 rounded-2xl border border-base-300 mt-1">
              <div class="flex items-center justify-between">
                <div>
                  <p class="font-bold text-base-content m-0 text-sm">{{ $t('notifDevice.ruleStatus') }}</p>
                  <p class="text-xs text-base-content/60 m-0 mt-0.5">{{ $t('notifDevice.ruleStatusDesc') }}</p>
                </div>
                <input type="checkbox" v-model="form.active" class="toggle toggle-primary toggle-sm" />
              </div>
            </div>

          </div>

          <!-- Pinned Footer -->
          <div class="border-t border-base-200 p-4 px-6 flex justify-end gap-2 shrink-0 bg-base-100">
            <button type="button" class="btn btn-sm btn-ghost rounded-xl" @click="closeModal" :disabled="isSaving">
              {{ $t('common.cancel') }}
            </button>
            <button type="submit" class="btn btn-sm btn-primary rounded-xl px-6 text-white font-semibold" :disabled="isSaving">
              <span v-if="isSaving" class="loading loading-spinner loading-xs"></span>
              {{ isEditing ? $t('common.save') : $t('notifDevice.createRule') }}
            </button>
          </div>
        </form>
      </div>
      <form method="dialog" class="modal-backdrop"><button @click="closeModal">close</button></form>
    </dialog>

    <!-- Delete Modal -->
    <dialog ref="deleteModal" class="modal z-[200]">
      <div class="modal-box rounded-2xl border border-base-300 p-6">
        <h3 class="font-bold text-lg text-error flex items-center gap-2">
          <Icon icon="lucide:alert-triangle" class="w-5 h-5" /> {{ $t('common.confirmDelete') }}
        </h3>
        <p class="py-3 text-sm text-base-content/80">
          {{ $t('notifDevice.deleteWarning', { name: ruleToDelete ? getDeviceName(ruleToDelete.deviceId) : '' }) }}
        </p>
        <div class="modal-action mt-4">
          <button type="button" @click="closeDeleteModal" class="btn btn-sm btn-ghost rounded-xl" :disabled="isDeleting">
            {{ $t('common.cancel') }}
          </button>
          <button type="button" @click="confirmDelete" class="btn btn-sm btn-error text-white rounded-xl font-semibold px-6" :disabled="isDeleting">
            <span v-if="isDeleting" class="loading loading-spinner loading-xs"></span> {{ $t('common.delete') }}
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
import { required, decimal, maxLength, helpers } from '@vuelidate/validators';
import { toast } from 'vue3-toastify';
import { Icon } from '@iconify/vue';
import { usePermissionStore } from '@/stores/usePermissionStore';
import NoAccess from '@/components/NoAccess.vue';
import SearchableDropdown from '@/components/SearchableDropdown.vue';
import TableData from '@/components/TableData.vue';
import StatCardGroup from '@/components/StatCardGroup.vue';
import { useErrorHandler } from '@/composables/useErrorHandler';
const { handleError } = useErrorHandler();

const { t, te } = useI18n();
const mainMenuName = 'Notification Device';

const permissionStore = usePermissionStore();
const { hasPermission } = permissionStore;

const { error: saveError, isLoading: isSaving, execute: saveApi } = useMutation();
const { data: ruleData, isLoading, error: fetchError, execute: fetchApi } = useFetch();
const { error: deleteError, isLoading: isDeleting, execute: deleteApi } = useMutation();
const { data: fetchDevices, execute: fetchDevicesApi } = useFetch();

const ruleModal = ref(null);
const deleteModal = ref(null);
const isEditing = ref(false);
const editingRuleId = ref(null);
const ruleToDelete = ref(null);

const ruleTableData = ref([]);
const deviceList = ref([]);

// Dynamic bilingual summary stats
const statCardsData = computed(() => {
  const total = ruleTableData.value.length;
  const active = ruleTableData.value.filter(r => r.active).length;
  const disabled = total - active;
  const uniqueDevices = new Set(ruleTableData.value.map(r => r.deviceId)).size;

  return [
    {
      label: t('notifDevice.stats.total'),
      value: total,
      icon: 'lucide:siren',
      color: 'primary'
    },
    {
      label: t('notifDevice.stats.active'),
      value: active,
      icon: 'lucide:bell-ring',
      color: 'success',
      valueClass: 'text-success'
    },
    {
      label: t('notifDevice.stats.disabled'),
      value: disabled,
      icon: 'lucide:bell-off',
      color: disabled > 0 ? 'warning' : 'ghost',
      valueClass: disabled > 0 ? 'text-warning' : 'text-base-content/50'
    },
    {
      label: t('notifDevice.stats.devices'),
      value: uniqueDevices,
      icon: 'lucide:cpu',
      color: 'info',
      valueClass: 'text-info'
    }
  ];
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

const tableColumns = computed(() => [
  { header: t('common.id'), accessorKey: 'ruleId', meta: { headerClass: 'w-16', cellClass: 'font-bold' } },
  {
    header: t('common.device'),
    id: 'deviceName',
    accessorFn: (row) => getDeviceName(row.deviceId)
  },
  { header: t('notifDevice.alertMessage'), accessorKey: 'reason' },
  { header: t('notifDevice.logicCondition'), id: 'logic', enableSorting: false },
  { header: t('common.status'), accessorKey: 'active' },
  { header: t('common.actions'), id: 'actions', enableSorting: false, meta: { headerClass: 'text-right w-28', cellClass: 'text-right' } }
]);

const form = ref({
  deviceId: null,
  condition: '>',
  threshold: null,
  reason: '',
  active: true
});

const rules = computed(() => ({
  deviceId: { required: helpers.withMessage(t('notifDevice.validation.deviceRequired'), required) },
  condition: { required },
  threshold: {
    required: helpers.withMessage(t('notifDevice.validation.thresholdRequired'), required),
    decimal: helpers.withMessage(t('notifDevice.validation.mustBeNumber'), decimal)
  },
  reason: {
    maxLength: helpers.withMessage(t('common.validation.maxLength', { len: 100 }), maxLength(100))
  }
}));

const v$ = useVuelidate(rules, form);

const loadTable = async () => {
  await fetchApi('/notification/devicerule/getalldetail');
  if (!fetchError.value && ruleData.value) {
    ruleTableData.value = ruleData.value.data || [];
  } else {
    toast.error(fetchError.value?.message || t('common.messages.loadError'));
  }
};

const loadDevices = async () => {
  await fetchDevicesApi('/device/getalldetail');
  if (fetchDevices.value) {
    deviceList.value = fetchDevices.value.data || [];
  }
};

const openCreateModal = () => {
  isEditing.value = false;
  editingRuleId.value = null;
  form.value = { deviceId: null, condition: '>', threshold: null, reason: '', active: true };
  v$.value.$reset();
  ruleModal.value.showModal();
};

const openEditModal = (rule) => {
  isEditing.value = true;
  editingRuleId.value = rule.ruleId;
  form.value = {
    deviceId: rule.deviceId,
    condition: rule.condition,
    threshold: rule.threshold,
    reason: rule.reason || '',
    active: rule.active
  };
  v$.value.$reset();
  ruleModal.value.showModal();
};

const closeModal = () => ruleModal.value.close();
const openDeleteModal = (rule) => { ruleToDelete.value = rule; deleteModal.value.showModal(); };
const closeDeleteModal = () => { deleteModal.value.close(); ruleToDelete.value = null; };

const submitForm = async () => {
  const isFormValid = await v$.value.$validate();
  if (!isFormValid) return;

  const payload = { ...form.value, threshold: Number(form.value.threshold) };

  if (isEditing.value) {
    payload.ruleId = editingRuleId.value;
    await saveApi('/notification/devicerule/update', payload, 'PUT');
  } else {
    await saveApi('/notification/devicerule/create', payload, 'POST');
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
  if (!ruleToDelete.value) return;
  await deleteApi(`/notification/devicerule/delete/${ruleToDelete.value.ruleId}`, null, 'DELETE');
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
  await loadDevices();
  await loadTable();
});
</script>