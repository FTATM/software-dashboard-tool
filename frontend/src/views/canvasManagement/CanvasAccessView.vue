<template>
  <div v-if="hasPermission(mainMenuName, 'Display')" class="w-full h-full overflow-y-auto p-4 sm:p-6 space-y-5">

    <!-- Anchored Page Header Card with Background -->
    <div class="bg-base-100 border border-base-300 rounded-2xl p-5 shadow-xs flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
      <div class="flex items-start sm:items-center gap-3.5">
        <div class="p-3 bg-primary/10 text-primary rounded-xl flex items-center justify-center shrink-0">
          <Icon icon="lucide:shield-check" class="w-6 h-6" />
        </div>
        <div>
          <!-- Breadcrumbs -->
          <div class="flex items-center gap-1.5 text-xs font-semibold text-base-content/50 uppercase tracking-wider mb-0.5">
            <span>{{ $t('menu.canvas') }}</span>
            <Icon icon="lucide:chevron-right" class="w-3.5 h-3.5" />
            <span class="text-primary">{{ $t('canvasAccess.title') }}</span>
          </div>

          <!-- Title & Subtitle -->
          <div class="flex items-center gap-2.5">
            <h1 class="m-0 text-xl sm:text-2xl font-black text-base-content tracking-tight">
              {{ $t('canvasAccess.title') }}
            </h1>
          </div>
          <p class="mt-0.5 mb-0 text-base-content/60 text-xs font-medium">
            {{ $t('canvasAccess.subtitle') }}
          </p>
        </div>
      </div>

      <!-- Header Actions -->
      <div class="flex items-center gap-2 self-end sm:self-center">
        <button 
          @click="setupData" 
          class="btn btn-sm btn-ghost border border-base-300 bg-base-100 hover:bg-base-200 rounded-xl gap-1.5 text-xs font-semibold shadow-xs transition-all"
          :title="$t('common.refresh')">
          <Icon icon="lucide:refresh-cw" class="w-3.5 h-3.5" :class="{ 'animate-spin': isRolesLoading }" />
          <span>{{ $t('common.refresh') }}</span>
        </button>
      </div>
    </div>

    <!-- Reusable KPI Summary Status Cards -->
    <StatCardGroup :items="statCardsData" />

    <!-- Main Table Card -->
    <div class="bg-base-100 border border-base-300 rounded-2xl p-4 sm:p-5 shadow-xs">
      <TableData 
        :data="roleTable" 
        :columns="tableColumns" 
        :initial-sorting="[{ id: 'roleName', desc: false }]"
        :is-loading="isRolesLoading">

        <!-- Toolbar Actions -->
        <template #toolbar-actions>
          <router-link 
            :to="{ name: 'canvas' }" 
            class="btn btn-sm btn-outline border-base-300 bg-base-100 hover:bg-base-200 rounded-xl text-xs font-semibold shadow-xs gap-1.5">
            <Icon icon="lucide:layout-dashboard" class="w-4 h-4 text-base-content/70" />
            <span>{{ $t('canvas.title') }}</span>
          </router-link>
        </template>

        <!-- Role Name Cell -->
        <template #cell-roleName="{ value }">
          <div class="flex items-center gap-2">
            <div class="w-7 h-7 rounded-lg bg-primary/10 text-primary flex items-center justify-center font-bold text-xs shrink-0">
              <Icon icon="lucide:shield" class="w-4 h-4" />
            </div>
            <span class="font-semibold text-base-content tracking-tight text-sm">{{ value }}</span>
          </div>
        </template>

        <!-- Assigned Canvases Badges -->
        <template #cell-assignedCanvases="{ row }">
          <div class="flex flex-wrap gap-1.5 items-center">
            <span 
              v-for="canvasId in getAssignedCanvases(row.roleId)" 
              :key="canvasId"
              class="inline-flex items-center px-2 py-0.5 rounded-lg text-xs font-semibold bg-primary/10 text-primary border border-primary/20">
              <Icon icon="lucide:layout-dashboard" class="w-3 h-3 mr-1 opacity-70" />
              {{ getCanvasName(canvasId) }}
            </span>

            <span v-if="getAssignedCanvases(row.roleId).length === 0" class="text-base-content/40 text-xs italic">
              {{ $t('canvasAccess.noAccess') }}
            </span>
          </div>
        </template>

        <!-- Minimal Icon Action -->
        <template #cell-actions="{ row }">
          <div class="flex justify-end items-center">
            <button 
              @click="openEditModal(row)" 
              class="btn btn-ghost btn-xs btn-square rounded-lg text-base-content/70 hover:text-primary hover:bg-primary/10 transition-colors"
              :title="$t('common.edit')">
              <Icon icon="lucide:pencil" class="w-4 h-4" />
            </button>
          </div>
        </template>
      </TableData>
    </div>

    <!-- Edit Access Modal -->
    <dialog ref="editModal" class="modal z-[200]">
      <div class="modal-box sm:w-11/12 sm:max-w-md p-0 overflow-hidden shadow-2xl rounded-2xl flex flex-col max-h-[85vh] border border-base-300 bg-base-100">
        <!-- Pinned Header -->
        <div class="px-6 py-4 border-b border-base-200 bg-base-100 flex justify-between items-center shrink-0">
          <div class="flex items-center gap-2.5">
            <div class="p-2 rounded-xl bg-primary/10 text-primary">
              <Icon icon="lucide:shield-check" class="w-5 h-5" />
            </div>
            <div>
              <h3 class="m-0 text-base sm:text-lg font-bold text-base-content">
                {{ $t('canvasAccess.manageAccessFor') }} <span class="text-primary font-black">{{ editingRole?.roleName }}</span>
              </h3>
              <p class="m-0 text-xs text-base-content/50">
                {{ $t('canvasAccess.editSubtitle') }}
              </p>
            </div>
          </div>
          <button class="btn btn-sm btn-circle btn-ghost" @click="closeEditModal" :disabled="isSaving">
            <Icon icon="lucide:x" class="w-4 h-4" />
          </button>
        </div>

        <!-- Modal Body -->
        <div class="p-6 overflow-y-auto flex-1 flex flex-col gap-3">
          <!-- Selection Counter & Toggle All -->
          <div class="flex justify-between items-center px-1 pb-1 border-b border-base-200">
            <span class="text-xs font-bold uppercase tracking-wider text-base-content/70">
              {{ $t('canvasAccess.assignedCanvases') }} ({{ selectedCanvases.length }}/{{ canvasList.length }})
            </span>
            <button 
              v-if="canvasList.length > 0" 
              type="button" 
              class="btn btn-xs rounded-lg font-medium transition-all"
              :class="isAllSelected ? 'btn-ghost text-error hover:bg-error/10' : 'btn-ghost text-primary hover:bg-primary/10'" 
              @click="toggleSelectAll">
              {{ isAllSelected ? $t('common.deselectAll') : $t('common.selectAll') }}
            </button>
          </div>

          <!-- Multi-Select Checkbox List -->
          <div class="flex flex-col w-full max-h-[300px] overflow-y-auto bg-base-200/50 p-2 rounded-xl border border-base-300 gap-1.5">
            <label 
              v-for="canvas in canvasList" 
              :key="canvas.canvasId"
              class="label cursor-pointer flex items-center justify-start gap-3 w-full hover:bg-base-100 px-3 py-2 rounded-lg transition-colors border border-transparent hover:border-base-300">
              <input 
                type="checkbox" 
                :value="canvas.canvasId" 
                v-model="selectedCanvases"
                class="checkbox checkbox-primary checkbox-xs rounded-md shrink-0" />
              <span class="label-text font-medium text-xs truncate text-base-content/90" :title="canvas.canvasName">
                {{ canvas.canvasName }}
              </span>
            </label>

            <div v-if="canvasList.length === 0" class="text-center py-8 text-xs text-base-content/40">
              {{ $t('canvasAccess.noCanvasesAvailable') }}
            </div>
          </div>
        </div>

        <!-- Pinned Footer -->
        <div class="border-t border-base-200 p-4 px-6 flex justify-end gap-2 shrink-0 bg-base-100">
          <button type="button" @click="closeEditModal" class="btn btn-sm btn-ghost rounded-xl" :disabled="isSaving">
            {{ $t('common.cancel') }}
          </button>
          <button type="button" @click="saveAccess" class="btn btn-sm btn-primary rounded-xl px-6 text-white font-semibold" :disabled="isSaving">
            <span v-if="isSaving" class="loading loading-spinner loading-xs"></span>
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
import StatCardGroup from '@/components/StatCardGroup.vue';
import { useErrorHandler } from '@/composables/useErrorHandler';
const { handleError } = useErrorHandler();

const { t } = useI18n();
const mainMenuName = 'Canvas Access';

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

// Dynamic bilingual summary stats
const statCardsData = computed(() => {
  const totalRoles = roleTable.value.length;
  const withAccess = roleTable.value.filter(r => (roleCanvasMap.value.get(r.roleId) || []).length > 0).length;
  const noAccess = totalRoles - withAccess;
  const totalCanvases = canvasList.value.length;

  return [
    {
      label: t('canvasAccess.stats.totalRoles'),
      value: totalRoles,
      icon: 'lucide:shield',
      color: 'primary'
    },
    {
      label: t('canvasAccess.stats.withAccess'),
      value: withAccess,
      icon: 'lucide:shield-check',
      color: 'success',
      valueClass: 'text-success'
    },
    {
      label: t('canvasAccess.stats.noAccess'),
      value: noAccess,
      icon: 'lucide:shield-x',
      color: noAccess > 0 ? 'warning' : 'ghost',
      valueClass: noAccess > 0 ? 'text-warning' : 'text-base-content/50'
    },
    {
      label: t('canvasAccess.stats.totalCanvases'),
      value: totalCanvases,
      icon: 'lucide:layout-dashboard',
      color: 'info',
      valueClass: 'text-info'
    }
  ];
});

const tableColumns = computed(() => [
  {
    header: t('common.roleName'),
    accessorKey: 'roleName',
    meta: { headerClass: 'w-56', cellClass: 'font-bold' }
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
    meta: { headerClass: 'text-right w-24', cellClass: 'text-right' }
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
    roleCanvasMap.value.clear();
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