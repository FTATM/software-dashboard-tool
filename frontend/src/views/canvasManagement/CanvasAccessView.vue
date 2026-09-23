<template>
  <div v-if="hasPermission(mainMenuName, 'Display')" class="p-4 sm:p-6 w-full mx-auto">

    <!-- Header -->
    <div
      class="bg-base-100 shadow-sm rounded-box border border-base-200 p-6 mb-6 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
      <div class="flex items-center gap-4">
        <div class="p-3 bg-primary/10 text-primary rounded-xl flex items-center justify-center">
          <Icon icon="lucide:shield-check" class="w-7 h-7" />
        </div>
        <div>
          <h2 class="m-0 text-2xl font-extrabold text-base-content tracking-tight">{{ $t('canvasAccess.title') }}</h2>
          <p class="mt-1 mb-0 text-base-content/60 text-sm font-medium">{{ $t('canvasAccess.subtitle') }}</p>
        </div>
      </div>
    </div>

    <!-- 3-Column Table -->
    <TableData :data="roleTable" :columns="tableColumns" :initial-sorting="[{ id: 'roleName', desc: false }]"
      :is-loading="isRolesLoading">

      <template #cell-roleName="{ value }">
        <span>{{ value }}</span>
      </template>

      <template #cell-assignedCanvases="{ row }">
        <div class="flex flex-wrap gap-2">
          <span v-for="canvasId in getAssignedCanvases(row.roleId)" :key="canvasId"
            class="badge badge-primary badge-outline badge-sm font-medium">
            {{ getCanvasName(canvasId) }}
          </span>

          <span v-if="getAssignedCanvases(row.roleId).length === 0" class="text-base-content/40 text-sm italic">
            {{ $t('canvasAccess.noAccess') }}
          </span>
        </div>
      </template>

      <template #cell-actions="{ row }">
        <div class="flex justify-end gap-2">
          <button @click="openEditModal(row)" class="btn btn-sm btn-primary">
            <Icon icon="lucide:pencil" class="w-5 h-5" />
          </button>
        </div>
      </template>
    </TableData>

    <!-- Edit Access Modal -->
    <dialog ref="editModal" class="modal z-[200]">
      <div class="modal-box">
        <h3 class="font-bold text-lg text-base-content mb-4 flex items-center gap-2">
          {{ $t('canvasAccess.manageAccessFor') }} <span class="text-primary">{{ editingRole?.roleName }}</span>
        </h3>

        <!-- Select All / Deselect All Bar -->
        <div class="flex justify-between items-center mb-2 px-1">
          <span class="text-sm font-semibold text-base-content/70">
            {{ $t('canvasAccess.assignedCanvases') }} ({{ selectedCanvases.length }}/{{ canvasList.length }})
          </span>
          <button v-if="canvasList.length > 0" type="button" class="btn btn-xs"
            :class="isAllSelected ? 'btn-ghost text-error' : 'btn-outline btn-primary'" @click="toggleSelectAll">
            {{ isAllSelected ? $t('common.deselectAll') : $t('common.selectAll') }}
          </button>
        </div>

        <!-- Multi-Select Checkbox List -->
        <div
          class="flex flex-col w-full max-h-[320px] overflow-y-auto bg-base-200/50 p-2 rounded-box border border-base-200 divide-y divide-base-200">
          <label v-for="canvas in canvasList" :key="canvas.canvasId"
            class="label cursor-pointer flex items-center justify-start gap-3 w-full hover:bg-base-200 px-3 py-2.5 rounded-lg transition-colors">
            <!-- shrink-0 keeps checkbox square even when text is very long -->
            <input type="checkbox" :value="canvas.canvasId" v-model="selectedCanvases"
              class="checkbox checkbox-primary checkbox-sm shrink-0" />
            <!-- truncate cuts off cleanly with ellipsis (...), title shows full name on hover -->
            <span class="label-text font-medium text-sm truncate" :title="canvas.canvasName">
              {{ canvas.canvasName }}
            </span>
          </label>

          <div v-if="canvasList.length === 0" class="text-center py-6 text-sm text-base-content/50">
            {{ $t('canvasAccess.noCanvasesAvailable') }}
          </div>
        </div>

        <!-- Footer Actions -->
        <div class="modal-action mt-6">
          <button type="button" @click="closeEditModal" class="btn btn-ghost" :disabled="isSaving">{{
            $t('common.cancel')
          }}</button>
          <button type="button" @click="saveAccess" class="btn btn-primary text-white" :disabled="isSaving">
            <span v-if="isSaving" class="loading loading-spinner loading-sm"></span>
            {{ $t('common.save') }}
          </button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop">
        <button @click="closeEditModal">close</button>
      </form>
    </dialog>

  </div>
  <NoAccess v-else />
</template>

<script setup>
import { ref, onMounted, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { useFetch } from '@/composables/useFetch';
import { useMutation } from '@/composables/useMutation';
import { toast } from 'vue3-toastify';
import { Icon } from '@iconify/vue';
import { usePermissionStore } from '@/stores/usePermissionStore';
import NoAccess from '@/components/NoAccess.vue';
import TableData from '@/components/TableData.vue';
import { useErrorHandler } from '@/composables/useErrorHandler';
const { handleError } = useErrorHandler();

const { t } = useI18n();
const mainMenuName = 'Canvas Access'

const { data: roleData, isLoading: isRolesLoading, error: roleError, execute: fetchRolesApi } = useFetch();
const { data: canvasData, error: canvasError, execute: fetchCanvasesApi } = useFetch();
const { data: mappingData, error: mappingError, execute: fetchMappingApi } = useFetch();
const { res: updateRes, isLoading: isSaving, error: updateError, execute: updateMappingApi } = useMutation();

const permissionStore = usePermissionStore();
const { hasPermission } = permissionStore;

const roleTable = ref([]);
const canvasList = ref([]);
const roleCanvasMap = ref(new Map());

const editModal = ref(null);
const editingRole = ref(null);
const selectedCanvases = ref([]);

const tableColumns = computed(() => [
  {
    header: t('common.roleName'),
    accessorKey: 'roleName'
  },
  {
    header: t('canvasAccess.assignedCanvases'),
    id: 'assignedCanvases',
    enableSorting: false,
  },
  {
    header: t('common.actions'),
    id: 'actions',
    enableSorting: false,
    meta: { headerClass: 'text-right', cellClass: 'text-right' }
  }
]);

const setupData = async () => {
  await fetchRolesApi('/role/getall');
  if (!roleError.value && roleData.value) {
    roleTable.value = roleData.value.data;
  } else {
    toast.error(t('common.messages.loadFailed', { item: "role" }));
  }

  await fetchCanvasesApi('/canvas/getall');
  if (!canvasError.value && canvasData.value) {
    canvasList.value = canvasData.value.data;
  } else {
    toast.error(t('common.messages.loadFailed', { item: "canvas" }));
  }

  await fetchMappingApi('/canvas/getallcanvasroledetail');
  if (!mappingError.value && mappingData.value) {
    roleCanvasMap.value.clear()
    for (let detail of mappingData.value.data) {
      roleCanvasMap.value.set(detail.roleId, detail.canvasIds || []);
    }
  } else if (mappingError.value) {
    toast.error(t('common.messages.loadFailed', { item: "role map" }));
  }
};

const getCanvasName = (id) => {
  const found = canvasList.value.find(c => c.canvasId === id);
  return found ? found.canvasName : t('canvasAccess.unknownCanvas');
};

const getAssignedCanvases = (roleId) => {
  return roleCanvasMap.value.get(roleId) || [];
};

const isAllSelected = computed(() => {
  if (canvasList.value.length === 0) return false;
  return canvasList.value.every(c => selectedCanvases.value.includes(c.canvasId));
});

// Toggle between selecting all and clearing all
const toggleSelectAll = () => {
  if (isAllSelected.value) {
    selectedCanvases.value = [];
  } else {
    selectedCanvases.value = canvasList.value.map(c => c.canvasId);
  }
};

const openEditModal = (role) => {
  editingRole.value = role;
  const currentAccess = roleCanvasMap.value.get(role.roleId) || [];
  selectedCanvases.value = [...currentAccess];
  editModal.value.showModal();
};

const closeEditModal = () => {
  editModal.value.close();
  editingRole.value = null;
  selectedCanvases.value = [];
};

const saveAccess = async () => {
  if (!editingRole.value) return;

  const payload = {
    roleId: editingRole.value.roleId,
    canvasIds: selectedCanvases.value
  };

  await updateMappingApi('/canvas/upsertcanvasrole', payload, 'POST');

  if (!updateError.value && updateRes.value?.ok) {
    roleCanvasMap.value.set(editingRole.value.roleId, [...selectedCanvases.value]);
    toast.success(t('common.messages.updateSuccess', { name: editingRole.value.roleName }));
    closeEditModal();
  } else {
    toast.error(handleError(updateError, 'common.messages.updateError'));
  }
};

onMounted(async () => {
  if (!hasPermission(mainMenuName, 'Display')) return;
  await setupData();
});
</script>