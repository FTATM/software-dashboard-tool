<template>
  <div v-if="hasPermission(mainMenuName, 'Display')" class="w-full h-full overflow-y-auto p-4">

    <div
      class="bg-base-100 shadow-sm rounded-box border border-base-200 p-6 mb-6 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
      <div class="flex items-center gap-4">
        <div class="p-3 bg-error/10 text-error rounded-xl flex items-center justify-center">
          <Icon icon="lucide:siren" class="w-7 h-7" />
        </div>
        <div>
          <h2 class="m-0 text-2xl font-extrabold text-base-content tracking-tight">{{ $t('notifDevice.title') }}</h2>
          <p class="mt-1 mb-0 text-base-content/60 text-sm font-medium">{{ $t('notifDevice.subtitle') }}</p>
        </div>
      </div>
    </div>

    <div class="flex-1 flex flex-col min-h-0">
      <TableData :data="ruleTableData" :columns="tableColumns" :is-loading="isLoading"
        :initial-sorting="[{ id: 'ruleId', desc: false }]">
        <template #toolbar-actions>
          <button class="btn btn-primary shadow-sm hover:shadow-md transition-all" @click="openCreateModal">
            <Icon icon="lucide:plus" class="w-5 h-5" />
            {{ $t('notifDevice.addRule') }}
          </button>
        </template>

        <template #cell-ruleId="{ value }">
          <span class="font-medium text-base-content/50">{{ value }}</span>
        </template>

        <!-- ⚡ Dynamic device name lookup matching SchedulerView -->
        <template #cell-deviceName="{ row }">
          <span :class="[
            'font-bold',
            isUnknownDevice(row.deviceId) ? 'text-base-content/50 italic font-medium' : 'text-primary'
          ]">
            {{ getDeviceName(row.deviceId) }}
          </span>
        </template>

        <template #cell-reason="{ value }">
          <div class="truncate max-w-[200px] text-sm text-base-content/80" :title="value">
            {{ value || $t('notifDevice.noReason') }}
          </div>
        </template>

        <template #cell-logic="{ row }">
          <div class="flex items-center gap-2 font-mono bg-base-200 px-2 py-1 rounded-md w-fit">
            <span class="text-base-content/70 text-xs">{{ $t('notifDevice.ifValue') }}</span>
            <span class="font-bold text-error">{{ row.condition }}</span>
            <span class="font-bold text-base-content">{{ row.threshold }}</span>
          </div>
        </template>

        <template #cell-active="{ value }">
          <span
            :class="['badge badge-sm font-semibold', value ? 'badge-success text-white' : 'badge-ghost text-base-content/40']">
            {{ value ? $t('common.active') : $t('common.disabled') }}
          </span>
        </template>

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
    <dialog ref="ruleModal" class="modal">
      <div class="modal-box sm:w-11/12 sm:max-w-xl p-0 overflow-hidden shadow-2xl">
        <div class="px-6 py-5 border-b border-base-200 bg-base-100 flex justify-between items-center">
          <h3 class="m-0 text-xl font-extrabold text-base-content">
            {{ isEditing ? $t('notifDevice.editRule') : $t('notifDevice.createRule') }}
          </h3>
          <button class="btn btn-sm btn-circle btn-ghost" @click="closeModal">
            <Icon icon="lucide:x" class="w-4 h-4" />
          </button>
        </div>

        <form @submit.prevent="submitForm" autocomplete="off" class="p-6 bg-base-100 flex flex-col gap-4">

          <!-- ⚡ Device Dropdown bound to realDeviceList only -->
          <label class="form-control w-full relative">
            <div class="label pb-1">
              <span class="label-text font-semibold">{{ $t('notifDevice.targetDevice') }}</span>
              <span class="label-text-alt text-error">*</span>
            </div>
            <SearchableDropdown v-model="form.deviceId" :options="realDeviceList" labelKey="deviceName"
              valueKey="deviceId" :placeholder="$t('common.searchDevice')" :error="v$.deviceId.$error"
              @blur="v$.deviceId.$touch()" />
            <div class="label px-1 py-1 h-6">
              <span v-if="v$.deviceId.$error" class="label-text-alt text-error font-medium">
                {{ v$.deviceId.$errors[0].$message }}
              </span>
            </div>
          </label>

          <div class="grid grid-cols-2 gap-4">
            <label class="form-control w-full">
              <div class="label pb-1"><span class="label-text font-semibold">{{ $t('notifDevice.condition') }}</span>
              </div>
              <select v-model="form.condition" @blur="v$.condition.$touch()"
                class="select select-bordered w-full font-mono text-lg font-bold">
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
                <span class="label-text font-semibold">{{ $t('notifDevice.thresholdValue') }}</span>
              </div>
              <input type="number" step="any" v-model="form.threshold" @blur="v$.threshold.$touch()"
                :placeholder="$t('notifDevice.thresholdPlaceholder')"
                class="input input-bordered w-full font-mono text-lg" />
              <div class="label px-1 py-1 h-6">
                <span v-if="v$.threshold.$error" class="label-text-alt text-error font-medium">
                  {{ v$.threshold.$errors[0].$message }}
                </span>
              </div>
            </label>
          </div>

          <label class="form-control w-full">
            <div class="label pb-1 flex justify-between">
              <span class="label-text font-semibold">{{ $t('notifDevice.alertMessageReason') }}</span>
              <span class="label-text-alt text-base-content/60 font-mono">{{ form.reason?.length || 0 }}/100</span>
            </div>
            <input type="text" v-model="form.reason" maxlength="100" @blur="v$.reason.$touch()"
              :placeholder="$t('notifDevice.alertMessagePlaceholder')"
              :class="['input input-bordered w-full', { 'input-error': v$.reason.$error }]" />
            <div class="label px-1 py-1 flex-col items-start gap-1">
              <span class="label-text-alt text-base-content/60">{{ $t('notifDevice.alertMessageDesc') }}</span>
              <span v-if="v$.reason.$error" class="label-text-alt text-error font-medium">
                {{ v$.reason.$errors[0].$message }}
              </span>
            </div>
          </label>

          <div class="p-4 bg-base-200/40 rounded-xl border border-base-200 flex items-center justify-between">
            <div>
              <p class="font-bold text-base-content m-0 text-sm">{{ $t('notifDevice.ruleStatus') }}</p>
              <p class="text-xs text-base-content/60 m-0 mt-1">{{ $t('notifDevice.ruleStatusDesc') }}</p>
            </div>
            <input type="checkbox" v-model="form.active" class="toggle toggle-success toggle-md" />
          </div>

          <div class="border-t border-base-200 mt-2 pt-5 flex justify-end gap-3">
            <button type="button" class="btn btn-ghost" @click="closeModal" :disabled="isSaving">
              {{ $t('common.cancel') }}
            </button>
            <button type="submit" class="btn btn-primary px-8" :disabled="isSaving">
              <span v-if="isSaving" class="loading loading-spinner loading-sm"></span>
              {{ isEditing ? $t('common.save') : $t('notifDevice.createRule') }}
            </button>
          </div>
        </form>
      </div>
      <form method="dialog" class="modal-backdrop"><button @click="closeModal">close</button></form>
    </dialog>

    <dialog ref="deleteModal" class="modal z-[200]">
      <div class="modal-box">
        <h3 class="font-bold text-lg text-error flex items-center gap-2">
          <Icon icon="lucide:alert-triangle" class="w-6 h-6" /> {{ $t('common.confirmDelete') }}
        </h3>
        <p class="py-4 text-base-content/80">
          {{ $t('notifDevice.deleteWarning', { name: ruleToDelete ? getDeviceName(ruleToDelete.deviceId) : '' }) }}
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
import { required, decimal, maxLength, helpers } from '@vuelidate/validators';
import { toast } from 'vue3-toastify';
import { Icon } from '@iconify/vue';
import { usePermissionStore } from '@/stores/usePermissionStore';
import NoAccess from '@/components/NoAccess.vue';
import SearchableDropdown from '@/components/SearchableDropdown.vue';
import TableData from '@/components/TableData.vue';
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

// ⚡ Check device active & soft-delete status
const isDeviceActive = (device) => {
  if (!device) return false;
  if (device.isDeleted || device.deleted || device.isDelete) return false;
  if (device.active === false || device.isActive === false) return false;
  if (typeof device.status === 'string' && device.status.toLowerCase() === 'inactive') return false;
  return true;
};

// ⚡ Check if device ID is missing, soft-deleted, or inactive
const isUnknownDevice = (deviceId) => {
  if (!deviceId) return true;
  const device = deviceList.value.find(d => d.deviceId === deviceId);
  return !device || !isDeviceActive(device);
};

// ⚡ Translation resolver with fallback to scheduler key
const getUnknownDeviceText = () => {
  if (te && te('notifDevice.unknownDevice')) return t('notifDevice.unknownDevice');
  if (te && te('common.unknownDevice')) return t('common.unknownDevice');
  return t('scheduler.unknownDevice');
};

// ⚡ Resolve device name like SchedulerView does
const getDeviceName = (deviceId) => {
  if (!deviceId) return getUnknownDeviceText();
  const device = deviceList.value.find(d => d.deviceId === deviceId);
  if (device && isDeviceActive(device)) {
    return device.deviceName;
  }
  return getUnknownDeviceText();
};

// ⚡ Filter out virtual sensors and inactive/deleted devices; retain current selection if editing
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
  // ⚡ Use /device/getalldetail so refDeviceId is populated on each device
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