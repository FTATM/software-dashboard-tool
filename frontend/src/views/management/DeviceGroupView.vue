<template>
  <div v-if="hasPermission(mainMenuName, 'Display')" class="w-full mx-auto p-4">
    <!-- Page Header Card -->
    <div
      class="bg-base-100 shadow-sm rounded-box border border-base-200 p-6 mb-6 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
      <div class="flex items-center gap-4">
        <div class="p-3 bg-primary/10 text-primary rounded-xl flex items-center justify-center">
          <Icon icon="lucide:layers" class="w-7 h-7" />
        </div>
        <div>
          <h2 class="m-0 text-2xl font-extrabold text-base-content tracking-tight">{{ $t('deviceGroup.title') }}</h2>
          <p class="mt-1 mb-0 text-base-content/60 text-sm font-medium">{{ $t('deviceGroup.subtitle') }}</p>
        </div>
      </div>
    </div>

    <!-- Groups Table -->
    <TableData :data="groupTable" :columns="tableColumns" :initial-sorting="[{ id: 'groupId', desc: false }]"
      :is-loading="isLoadingGroups">
      <template #toolbar-actions>
        <button class="btn btn-primary shadow-sm hover:shadow-md transition-all" @click="openCreateModal">
          <Icon icon="lucide:plus" class="w-5 h-5 mr-1" />
          {{ $t('deviceGroup.addGroup') }}
        </button>
      </template>

      <!-- ⚡ NEW: Protocol Badge Display -->
      <template #cell-protocol="{ value }">
        <span v-if="value" class="badge badge-outline badge-sm font-bold uppercase tracking-wider text-primary">
          {{ value }}
        </span>
        <span v-else class="badge badge-ghost badge-sm font-bold tracking-wider text-base-content/50">
          {{ $t('common.none') }}
        </span>
      </template>

      <template #cell-deviceCount="{ row }">
        <div class="badge badge-outline badge-sm font-bold text-secondary">
          {{ row.deviceIds?.length || 0 }} {{ $t('common.devices') }}
        </div>
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

    <!-- Create/Edit Modal -->
    <dialog ref="groupModal" class="modal">
      <div class="modal-box sm:w-11/12 sm:max-w-xl p-0 overflow-visible shadow-2xl">
        <div class="px-6 py-5 border-b border-base-200 bg-base-100 flex justify-between items-center">
          <h3 class="m-0 text-xl font-extrabold text-base-content">
            {{ isEditing ? $t('deviceGroup.editGroup') : $t('deviceGroup.createGroup') }}
          </h3>
          <button type="button" class="btn btn-sm btn-circle btn-ghost" @click="closeModal">
            <Icon icon="lucide:x" class="w-4 h-4" />
          </button>
        </div>

        <form @submit.prevent="submitForm" autocomplete="off" class="p-6 bg-base-100">
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-x-5 gap-y-3">

            <label class="form-control w-full sm:col-span-2">
              <div class="label pb-1 flex justify-between">
                <span class="label-text font-semibold">{{ $t('deviceGroup.groupName') }}</span>
                <span class="label-text-alt text-base-content/60 font-mono">
                  {{ form.groupName?.length || 0 }}/31
                </span>
              </div>
              <input type="text" v-model="form.groupName" maxlength="31"
                :placeholder="$t('deviceGroup.groupNamePlaceholder')" @blur="v$.groupName.$touch()"
                :class="['input input-bordered w-full', { 'input-error': v$.groupName.$error }]" />
              <div class="label px-1 py-1 h-6">
                <span v-if="v$.groupName.$error" class="label-text-alt text-error font-medium">
                  {{ v$.groupName.$errors[0].$message }}
                </span>
              </div>
            </label>

            <!-- ⚡ NEW: Gateway Protocol Dropdown -->
            <label class="form-control w-full sm:col-span-2">
              <div class="label pb-1">
                <span class="label-text font-semibold">{{ $t('common.protocol') }}</span>
              </div>
              <select v-model="form.protocol" class="select select-bordered w-full">
                <option value="" disabled>{{ $t('common.protocolPlaceholder') }}</option>
                <option value="none">{{ $t('common.none') }}</option>
                <option v-for="proto in protocolList" :key="proto" :value="proto">
                  {{ proto }}
                </option>
              </select>
            </label>

            <label class="form-control w-full sm:col-span-2">
              <div class="label pb-1 flex justify-between">
                <span class="label-text font-semibold">{{ $t('common.description') }}</span>
                <span class="label-text-alt text-base-content/60 font-mono">
                  {{ form.description?.length || 0 }}/100
                </span>
              </div>
              <input type="text" v-model="form.description" maxlength="100"
                :placeholder="$t('deviceGroup.descriptionPlaceholder')" class="input input-bordered w-full" />
            </label>

            <label class="form-control w-full mt-2 sm:col-span-2">
              <div class="label pb-1">
                <span class="label-text font-semibold">{{ $t('deviceGroup.assignDevices') }}</span>
                <span class="text-xs text-base-content/60">{{ $t('deviceGroup.assignDevicesDesc') }}</span>
              </div>
              <SearchableDropdown v-model="form.deviceIds" :options="deviceOptions" labelKey="deviceName"
                valueKey="deviceId" :multiple="true" :placeholder="$t('common.searchDevice')" />
            </label>
          </div>

          <div class="border-t border-base-200 mt-6 pt-5 flex justify-end gap-3">
            <button type="button" class="btn btn-ghost" @click="closeModal">{{ $t('common.cancel') }}</button>
            <button type="submit" class="btn btn-primary px-8 text-white">
              {{ isEditing ? $t('common.save') : $t('deviceGroup.createGroup') }}
            </button>
          </div>
        </form>
      </div>
      <form method="dialog" class="modal-backdrop"><button @click="closeModal">close</button></form>
    </dialog>

    <!-- Delete Modal -->
    <dialog ref="deleteModal" class="modal z-[200]">
      <div class="modal-box">
        <h3 class="font-bold text-lg text-error flex items-center gap-2">
          <Icon icon="lucide:alert-triangle" class="w-6 h-6" /> {{ $t('common.confirmDelete') }}
        </h3>
        <p class="py-4 text-base-content/80">
          {{ $t('deviceGroup.deleteWarning', { name: groupToDelete?.groupName }) }}
        </p>
        <div class="modal-action">
          <button type="button" @click="closeDeleteModal" class="btn btn-ghost" :disabled="isDeleting">
            {{ $t('common.noCancel') }}
          </button>
          <button type="button" @click="confirmDelete" class="btn btn-error text-white" :disabled="isDeleting">
            <span v-if="isDeleting" class="loading loading-spinner loading-sm"></span> {{ $t('common.yesDelete') }}
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
import { required, maxLength, helpers } from '@vuelidate/validators';
import { toast } from 'vue3-toastify';
import { Icon } from '@iconify/vue';
import { usePermissionStore } from '@/stores/usePermissionStore';
import NoAccess from '@/components/NoAccess.vue';
import TableData from '@/components/TableData.vue';
import SearchableDropdown from '@/components/SearchableDropdown.vue';
import { useErrorHandler } from '@/composables/useErrorHandler';
const { handleError } = useErrorHandler();

const { t } = useI18n();
const mainMenuName = 'Device Group'

const permissionStore = usePermissionStore();
const { hasPermission } = permissionStore;

const { error: groupAddedError, execute: groupAddedApi } = useMutation();
const { error: groupUpdatedError, execute: groupUpdatedApi } = useMutation();
const { data: groupAllFetch, isLoading: isLoadingGroups, error: groupAllFetchError, execute: groupAllFetchApi } = useFetch();
const { error: groupDeletedError, isLoading: isDeleting, execute: groupDeletedApi } = useMutation();
const { data: deviceData, error: deviceError, execute: deviceFetchApi } = useFetch();

// ⚡ NEW: Protocol Fetching
const { data: protocolData, error: protocolError, execute: protocolFetchApi } = useFetch();

const groupModal = ref(null);
const deleteModal = ref(null);
const isEditing = ref(false);
const editingGroupId = ref(null);
const groupTable = ref([]);
const groupToDelete = ref(null);
const deviceOptions = ref([]);
const protocolList = ref([]);

const tableColumns = computed(() => [
  { header: t('common.id'), accessorKey: 'groupId', meta: { headerClass: 'w-16', cellClass: 'font-bold' } },
  { header: t('deviceGroup.groupName'), accessorKey: 'groupName' },
  { header: t('common.description'), accessorKey: 'description' },
  { header: t('common.protocol'), accessorKey: 'protocol' }, // ⚡ Added to Table
  { header: t('common.devices'), id: 'deviceCount', enableSorting: false },
  { header: t('common.actions'), id: 'actions', enableSorting: false, meta: { headerClass: 'text-right', cellClass: 'text-right' } }
]);

const form = ref({
  groupName: '',
  description: '',
  protocol: 'none', // ⚡ Default to none
  deviceIds: []
});

const rules = computed(() => ({
  groupName: {
    required: helpers.withMessage(t('deviceGroup.validation.groupNameRequired'), required),
    maxLength: helpers.withMessage(t('common.validation.maxLength', { len: 31 }), maxLength(31))
  }
}));
const v$ = useVuelidate(rules, form);

const loadProtocols = async () => {
  await protocolFetchApi('/device/getprotocoltype');
  if (!protocolError.value && protocolData.value) {
    protocolList.value = protocolData.value.data || [];
  }
};

const loadDevicesForDropdown = async () => {
  await deviceFetchApi('/device/getalldetail');
  if (!deviceError.value && deviceData.value) {
    deviceOptions.value = deviceData.value.data || [];
  }
};

const loadTable = async () => {
  await groupAllFetchApi('/device/group/getalldetail');
  if (!groupAllFetchError.value && groupAllFetch.value) {
    groupTable.value = groupAllFetch.value.data || [];
  }
};

const openCreateModal = () => {
  isEditing.value = false;
  editingGroupId.value = null;
  form.value = { groupName: '', description: '', protocol: 'none', deviceIds: [] };
  v$.value.$reset();
  groupModal.value.showModal();
};

const openEditModal = (group) => {
  isEditing.value = true;
  editingGroupId.value = group.groupId;
  form.value = {
    groupName: group.groupName,
    description: group.description || '',
    protocol: group.protocol || 'none', // ⚡ Map null to 'none' for UI
    deviceIds: group.deviceIds || []
  };
  v$.value.$reset();
  groupModal.value.showModal();
};

const closeModal = () => groupModal.value.close();
const openDeleteModal = (group) => { groupToDelete.value = group; deleteModal.value.showModal(); };
const closeDeleteModal = () => { deleteModal.value.close(); groupToDelete.value = null; };

const confirmDelete = async () => {
  if (!groupToDelete.value) return;
  await groupDeletedApi(`/device/group/delete/${groupToDelete.value.groupId}`, null, 'DELETE');
  if (!groupDeletedError.value) {
    toast.success(t('common.messages.deleteSuccess', { name: groupToDelete.value.groupName }));
    await loadTable();
    closeDeleteModal();
  } else {
    toast.error(handleError(groupDeletedError, 'common.messages.deleteFailed', { item: groupToDelete.value.groupName }));
  }
};

const submitForm = async () => {
  const isFormValid = await v$.value.$validate();
  if (!isFormValid) return;

  const payload = {
    groupName: form.value.groupName,
    description: form.value.description,
    protocol: form.value.protocol === 'none' ? null : form.value.protocol, // ⚡ Strip 'none' back to null for DB
    deviceIds: form.value.deviceIds
  };

  if (isEditing.value) {
    payload.groupId = editingGroupId.value;
    await groupUpdatedApi('/device/group/update', payload, 'PUT');
    if (!groupUpdatedError.value) {
      closeModal();
      toast.success(t('common.messages.updated'));
      await loadTable();
    } else {
      toast.error(handleError(groupUpdatedError.value?.message, 'common.messages.updateFailed', { item: payload.groupName }));
    }
  } else {
    await groupAddedApi('/device/group/create', payload, 'POST');
    if (!groupAddedError.value) {
      closeModal();
      toast.success(t('common.messages.created'));
      await loadTable();
    } else {
      toast.error(handleError(groupAddedError.value?.message, 'common.messages.createFailed', { item: payload.groupName }));
    }
  }
};

onMounted(async () => {
  if (!hasPermission(mainMenuName, 'Display')) return;
  await loadProtocols();
  await loadDevicesForDropdown();
  await loadTable();
});
</script>