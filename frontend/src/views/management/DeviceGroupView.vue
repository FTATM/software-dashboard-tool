<template>
  <div v-if="hasPermission(mainMenuName, 'Display')" class="w-full h-full overflow-y-auto p-4 sm:p-6 space-y-5">
    
    <!-- Anchored Page Header Card with Background -->
    <div class="bg-base-100 border border-base-300 rounded-2xl p-5 shadow-xs flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
      <div class="flex items-start sm:items-center gap-3.5">
        <div class="p-3 bg-primary/10 text-primary rounded-xl flex items-center justify-center shrink-0">
          <Icon icon="lucide:layers" class="w-6 h-6" />
        </div>
        <div>
          <!-- Breadcrumbs -->
          <div class="flex items-center gap-1.5 text-xs font-semibold text-base-content/50 uppercase tracking-wider mb-0.5">
            <span>{{ $t('menu.management') }}</span>
            <Icon icon="lucide:chevron-right" class="w-3.5 h-3.5" />
            <span class="text-primary">{{ $t('deviceGroup.title') }}</span>
          </div>

          <!-- Title & Subtitle -->
          <div class="flex items-center gap-2.5">
            <h1 class="m-0 text-xl sm:text-2xl font-black text-base-content tracking-tight">
              {{ $t('deviceGroup.title') }}
            </h1>
          </div>
          <p class="mt-0.5 mb-0 text-base-content/60 text-xs font-medium">
            {{ $t('deviceGroup.subtitle') }}
          </p>
        </div>
      </div>

      <!-- Header Actions -->
      <div class="flex items-center gap-2 self-end sm:self-center">
        <button 
          @click="loadTable" 
          class="btn btn-sm btn-ghost border border-base-300 bg-base-100 hover:bg-base-200 rounded-xl gap-1.5 text-xs font-semibold shadow-xs transition-all"
          :title="$t('common.refresh')">
          <Icon icon="lucide:refresh-cw" class="w-3.5 h-3.5" :class="{ 'animate-spin': isLoadingGroups }" />
          <span>{{ $t('common.refresh') }}</span>
        </button>
      </div>
    </div>

    <!-- Reusable KPI Summary Status Cards -->
    <StatCardGroup :items="statCardsData" />

    <!-- Main Table Card -->
    <div class="bg-base-100 border border-base-300 rounded-2xl p-4 sm:p-5 shadow-xs">
      <TableData 
        :data="groupTable" 
        :columns="tableColumns" 
        :initial-sorting="[{ id: 'groupId', desc: false }]"
        :is-loading="isLoadingGroups">
        
        <!-- Toolbar Actions -->
        <template #toolbar-actions>
          <button 
            class="btn btn-sm btn-primary rounded-xl font-semibold shadow-xs hover:shadow-md transition-all gap-1 text-white" 
            @click="openCreateModal">
            <Icon icon="lucide:plus" class="w-4 h-4" />
            {{ $t('deviceGroup.addGroup') }}
          </button>
        </template>

        <!-- ID Cell -->
        <template #cell-groupId="{ value }">
          <span class="font-mono text-xs font-bold text-base-content/50">#{{ value }}</span>
        </template>

        <!-- Group Name Cell -->
        <template #cell-groupName="{ row }">
          <span class="font-semibold text-base-content tracking-tight text-sm">{{ row.groupName }}</span>
        </template>

        <!-- Description Cell -->
        <template #cell-description="{ value }">
          <span v-if="value" class="text-xs text-base-content/70">{{ value }}</span>
          <span v-else class="text-base-content/30 text-xs font-mono">-</span>
        </template>

        <!-- Protocol Badge Display -->
        <template #cell-protocol="{ value }">
          <span v-if="value && value !== 'none'" class="inline-flex items-center px-2 py-0.5 rounded-md bg-base-200 border border-base-300 text-xs font-semibold uppercase tracking-wider text-base-content/80 font-mono">
            {{ value }}
          </span>
          <span v-else class="text-base-content/40 text-xs">
            {{ $t('common.none') }}
          </span>
        </template>

        <!-- Device Count Pill -->
        <template #cell-deviceCount="{ row }">
          <div class="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-lg bg-info/10 text-info border border-info/20 text-xs font-semibold">
            <Icon icon="lucide:cpu" class="w-3.5 h-3.5" />
            <span>{{ row.deviceIds?.length || 0 }} {{ $t('common.devices') }}</span>
          </div>
        </template>

        <!-- Minimal Icon Actions -->
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

    <!-- Create/Edit Modal with Scrollable Body -->
    <dialog ref="groupModal" class="modal">
      <div class="modal-box sm:w-11/12 sm:max-w-xl p-0 overflow-hidden shadow-2xl rounded-2xl flex flex-col max-h-[85vh] border border-base-300 bg-base-100">
        <!-- Pinned Header -->
        <div class="px-6 py-4 border-b border-base-200 bg-base-100 flex justify-between items-center shrink-0">
          <div class="flex items-center gap-2.5">
            <div class="p-2 rounded-xl bg-primary/10 text-primary">
              <Icon :icon="isEditing ? 'lucide:pencil' : 'lucide:folder-plus'" class="w-5 h-5" />
            </div>
            <div>
              <h3 class="m-0 text-lg font-bold text-base-content">
                {{ isEditing ? $t('deviceGroup.editGroup') : $t('deviceGroup.createGroup') }}
              </h3>
              <p class="m-0 text-xs text-base-content/50">
                {{ isEditing ? $t('deviceGroup.editSubtitle') : $t('deviceGroup.createSubtitle') }}
              </p>
            </div>
          </div>
          <button type="button" class="btn btn-sm btn-circle btn-ghost" @click="closeModal">
            <Icon icon="lucide:x" class="w-4 h-4" />
          </button>
        </div>

        <!-- Form Wrapper -->
        <form @submit.prevent="submitForm" autocomplete="off" class="flex flex-col flex-1 overflow-hidden bg-base-100">
          <div class="p-6 overflow-y-auto flex-1 flex flex-col gap-4">

            <!-- Group Name -->
            <label class="form-control w-full">
              <div class="label pb-1 flex justify-between">
                <span class="label-text font-semibold text-xs uppercase tracking-wider text-base-content/70">{{ $t('deviceGroup.groupName') }}</span>
                <span class="label-text-alt text-base-content/50 font-mono text-[11px]">
                  {{ form.groupName?.length || 0 }}/31
                </span>
              </div>
              <input type="text" v-model="form.groupName" maxlength="31"
                :placeholder="$t('deviceGroup.groupNamePlaceholder')" @blur="v$.groupName.$touch()"
                :class="['input input-sm h-10 input-bordered w-full rounded-xl', { 'input-error': v$.groupName.$error }]" />
              <div class="label px-1 py-0.5 min-h-[20px]">
                <span v-if="v$.groupName.$error" class="label-text-alt text-error font-medium text-xs">
                  {{ v$.groupName.$errors[0].$message }}
                </span>
              </div>
            </label>

            <!-- Gateway Protocol Dropdown -->
            <label class="form-control w-full">
              <div class="label pb-1">
                <span class="label-text font-semibold text-xs uppercase tracking-wider text-base-content/70">{{ $t('common.protocol') }}</span>
              </div>
              <select v-model="form.protocol" class="select select-sm h-10 select-bordered w-full rounded-xl">
                <option value="" disabled>{{ $t('common.protocolPlaceholder') }}</option>
                <option value="none">{{ $t('common.none') }}</option>
                <option v-for="proto in protocolList" :key="proto" :value="proto">
                  {{ proto }}
                </option>
              </select>
            </label>

            <!-- Description -->
            <label class="form-control w-full">
              <div class="label pb-1 flex justify-between">
                <span class="label-text font-semibold text-xs uppercase tracking-wider text-base-content/70">{{ $t('common.description') }}</span>
                <span class="label-text-alt text-base-content/50 font-mono text-[11px]">
                  {{ form.description?.length || 0 }}/100
                </span>
              </div>
              <input type="text" v-model="form.description" maxlength="100"
                :placeholder="$t('deviceGroup.descriptionPlaceholder')" class="input input-sm h-10 input-bordered w-full rounded-xl" />
            </label>

            <!-- Assign Devices Dropdown -->
            <label class="form-control w-full mt-1">
              <div class="label pb-1 flex justify-between items-center">
                <span class="label-text font-semibold text-xs uppercase tracking-wider text-base-content/70">{{ $t('deviceGroup.assignDevices') }}</span>
                <span class="text-xs text-base-content/50">{{ $t('deviceGroup.assignDevicesDesc') }}</span>
              </div>
              <SearchableDropdown v-model="form.deviceIds" :options="deviceOptions" labelKey="deviceName"
                valueKey="deviceId" :multiple="true" :placeholder="$t('common.searchDevice')" />
            </label>

          </div>

          <!-- Pinned Footer -->
          <div class="border-t border-base-200 p-4 px-6 flex justify-end gap-2 shrink-0 bg-base-100">
            <button type="button" class="btn btn-sm btn-ghost rounded-xl" @click="closeModal">{{ $t('common.cancel') }}</button>
            <button type="submit" class="btn btn-sm btn-primary rounded-xl px-6 text-white font-semibold">
              {{ isEditing ? $t('common.save') : $t('deviceGroup.createGroup') }}
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
          {{ $t('deviceGroup.deleteWarning', { name: groupToDelete?.groupName }) }}
        </p>
        <div class="modal-action mt-4">
          <button type="button" @click="closeDeleteModal" class="btn btn-sm btn-ghost rounded-xl" :disabled="isDeleting">
            {{ $t('common.noCancel') }}
          </button>
          <button type="button" @click="confirmDelete" class="btn btn-sm btn-error text-white rounded-xl font-semibold" :disabled="isDeleting">
            <span v-if="isDeleting" class="loading loading-spinner loading-xs"></span> {{ $t('common.yesDelete') }}
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
import StatCardGroup from '@/components/StatCardGroup.vue';
import { useErrorHandler } from '@/composables/useErrorHandler';
const { handleError } = useErrorHandler();

const { t } = useI18n();
const mainMenuName = 'Device Group';

const permissionStore = usePermissionStore();
const { hasPermission } = permissionStore;

const { error: groupAddedError, execute: groupAddedApi } = useMutation();
const { error: groupUpdatedError, execute: groupUpdatedApi } = useMutation();
const { data: groupAllFetch, isLoading: isLoadingGroups, error: groupAllFetchError, execute: groupAllFetchApi } = useFetch();
const { error: groupDeletedError, isLoading: isDeleting, execute: groupDeletedApi } = useMutation();
const { data: deviceData, error: deviceError, execute: deviceFetchApi } = useFetch();
const { data: protocolData, error: protocolError, execute: protocolFetchApi } = useFetch();

const groupModal = ref(null);
const deleteModal = ref(null);
const isEditing = ref(false);
const editingGroupId = ref(null);
const groupTable = ref([]);
const groupToDelete = ref(null);
const deviceOptions = ref([]);
const protocolList = ref([]);

// Dynamic bilingual summary stats
const statCardsData = computed(() => {
  const total = groupTable.value.length;
  const assignedDevices = groupTable.value.reduce((acc, g) => acc + (g.deviceIds?.length || 0), 0);
  const withProtocol = groupTable.value.filter(g => g.protocol && g.protocol !== 'none').length;
  const empty = groupTable.value.filter(g => !g.deviceIds || g.deviceIds.length === 0).length;

  return [
    {
      label: t('deviceGroup.stats.total'),
      value: total,
      icon: 'lucide:layers',
      color: 'primary'
    },
    {
      label: t('deviceGroup.stats.assignedDevices'),
      value: assignedDevices,
      icon: 'lucide:cpu',
      color: 'info',
      valueClass: 'text-info'
    },
    {
      label: t('deviceGroup.stats.withProtocol'),
      value: withProtocol,
      icon: 'lucide:network',
      color: 'success',
      valueClass: 'text-success'
    },
    {
      label: t('deviceGroup.stats.empty'),
      value: empty,
      icon: 'lucide:folder-x',
      color: empty > 0 ? 'warning' : 'ghost',
      valueClass: empty > 0 ? 'text-warning' : 'text-base-content/50'
    }
  ];
});

const tableColumns = computed(() => [
  { header: t('common.id'), accessorKey: 'groupId', meta: { headerClass: 'w-16', cellClass: 'font-bold' } },
  { header: t('deviceGroup.groupName'), accessorKey: 'groupName' },
  { header: t('common.description'), accessorKey: 'description' },
  { header: t('common.protocol'), accessorKey: 'protocol' },
  { header: t('common.devices'), id: 'deviceCount', enableSorting: false },
  { header: t('common.actions'), id: 'actions', enableSorting: false, meta: { headerClass: 'text-right', cellClass: 'text-right' } }
]);

const form = ref({
  groupName: '',
  description: '',
  protocol: 'none',
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
    protocol: group.protocol || 'none',
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
    protocol: form.value.protocol === 'none' ? null : form.value.protocol,
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