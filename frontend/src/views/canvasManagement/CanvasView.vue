<template>
  <div v-if="hasPermission(mainMenuName, 'Display')" class="w-full h-full overflow-y-auto p-4 sm:p-6 space-y-5">

    <!-- Anchored Page Header Card with Background -->
    <div class="bg-base-100 border border-base-300 rounded-2xl p-5 shadow-xs flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
      <div class="flex items-start sm:items-center gap-3.5">
        <div class="p-3 bg-primary/10 text-primary rounded-xl flex items-center justify-center shrink-0">
          <Icon icon="lucide:layout-dashboard" class="w-6 h-6" />
        </div>
        <div>
          <!-- Breadcrumbs -->
          <div class="flex items-center gap-1.5 text-xs font-semibold text-base-content/50 uppercase tracking-wider mb-0.5">
            <span>{{ $t('menu.canvas') }}</span>
            <Icon icon="lucide:chevron-right" class="w-3.5 h-3.5" />
            <span class="text-primary">{{ $t('canvas.title') }}</span>
          </div>

          <!-- Title & Subtitle -->
          <div class="flex items-center gap-2.5">
            <h1 class="m-0 text-xl sm:text-2xl font-black text-base-content tracking-tight">
              {{ $t('canvas.title') }}
            </h1>
          </div>
          <p class="mt-0.5 mb-0 text-base-content/60 text-xs font-medium">
            {{ $t('canvas.subtitle') }}
          </p>
        </div>
      </div>

      <!-- Header Actions -->
      <div class="flex items-center gap-2 self-end sm:self-center">
        <button 
          @click="loadCanvases" 
          class="btn btn-sm btn-ghost border border-base-300 bg-base-100 hover:bg-base-200 rounded-xl gap-1.5 text-xs font-semibold shadow-xs transition-all"
          :title="$t('common.refresh')">
          <Icon icon="lucide:refresh-cw" class="w-3.5 h-3.5" :class="{ 'animate-spin': isLoadingCanvases }" />
          <span>{{ $t('common.refresh') }}</span>
        </button>
      </div>
    </div>

    <!-- Reusable KPI Summary Status Cards -->
    <StatCardGroup :items="statCardsData" />

    <!-- Main Table Card -->
    <div class="bg-base-100 border border-base-300 rounded-2xl p-4 sm:p-5 shadow-xs">
      <TableData 
        :data="canvasTable" 
        :columns="tableColumns" 
        :initial-sorting="[{ id: 'canvasId', desc: false }]"
        :is-loading="isLoadingCanvases">

        <!-- Toolbar Actions -->
        <template #toolbar-actions>
          <router-link 
            :to="{ name: 'canvasAccess' }" 
            class="btn btn-sm btn-outline border-base-300 bg-base-100 hover:bg-base-200 rounded-xl text-xs font-semibold shadow-xs gap-1.5">
            <Icon icon="lucide:shield-check" class="w-4 h-4 text-base-content/70" />
            <span>{{ $t('canvasAccess.title') }}</span>
          </router-link>

          <button 
            @click="openCreateModal" 
            class="btn btn-sm btn-primary rounded-xl font-semibold shadow-xs hover:shadow-md transition-all gap-1 text-white">
            <Icon icon="lucide:plus" class="w-4 h-4" />
            {{ $t('canvas.createCanvas') }}
          </button>
        </template>

        <!-- ID Cell -->
        <template #cell-canvasId="{ value }">
          <span class="font-mono text-xs font-bold text-base-content/50">#{{ value }}</span>
        </template>

        <!-- Canvas Name Cell -->
        <template #cell-canvasName="{ row }">
          <div class="flex items-center gap-2">
            <div class="w-7 h-7 rounded-lg bg-primary/10 text-primary flex items-center justify-center font-bold text-xs shrink-0">
              <Icon icon="lucide:layout-template" class="w-4 h-4" />
            </div>
            <span class="font-semibold text-base-content tracking-tight text-sm">{{ row.canvasName }}</span>
          </div>
        </template>

        <!-- Actions Slot -->
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

    <!-- Create / Edit Modal -->
    <dialog ref="formModal" class="modal z-[200]">
      <div class="modal-box sm:w-11/12 sm:max-w-md p-0 overflow-hidden shadow-2xl rounded-2xl flex flex-col border border-base-300 bg-base-100">
        <!-- Pinned Header -->
        <div class="px-6 py-4 border-b border-base-200 bg-base-100 flex justify-between items-center shrink-0">
          <div class="flex items-center gap-2.5">
            <div class="p-2 rounded-xl bg-primary/10 text-primary">
              <Icon :icon="form.canvasId ? 'lucide:pencil' : 'lucide:layout-dashboard'" class="w-5 h-5" />
            </div>
            <div>
              <h3 class="m-0 text-lg font-bold text-base-content">
                {{ form.canvasId ? $t('canvas.renameCanvas') :$t('canvas.createNewCanvas') }}
              </h3>
              <p class="m-0 text-xs text-base-content/50">
                {{ form.canvasId ? $t('canvas.editSubtitle') :$t('canvas.createSubtitle') }}
              </p>
            </div>
          </div>
          <button class="btn btn-sm btn-circle btn-ghost" @click="closeFormModal" :disabled="isMutating">
            <Icon icon="lucide:x" class="w-4 h-4" />
          </button>
        </div>

        <!-- Form Body -->
        <div class="p-6 flex flex-col gap-4 bg-base-100">
          <label class="form-control w-full">
            <div class="label pb-1 flex justify-between items-center">
              <span class="label-text font-semibold text-xs uppercase tracking-wider text-base-content/70">
                {{ $t('canvas.canvasName') }}
                <span class="text-error ml-0.5">*</span>
              </span>
              <span class="label-text-alt text-base-content/50 font-mono text-[11px]">{{ form.canvasName?.length || 0 }}/31</span>
            </div>
            <input 
              type="text" 
              v-model="form.canvasName" 
              maxlength="31" 
              @blur="v$.canvasName.$touch()"
              :class="['input input-sm h-10 input-bordered w-full rounded-xl', { 'input-error': v$.canvasName.$error }]"
              :placeholder="$t('canvas.canvasNamePlaceholder')" 
              @keyup.enter="submitForm" />
            <div class="label px-1 py-0.5 min-h-[20px]">
              <span v-if="v$.canvasName.$error" class="label-text-alt text-error font-medium text-xs">
                {{ v$.canvasName.$errors[0].$message }}
              </span>
            </div>
          </label>
        </div>

        <!-- Pinned Footer -->
        <div class="border-t border-base-200 p-4 px-6 flex justify-end gap-2 shrink-0 bg-base-100">
          <button type="button" @click="closeFormModal" class="btn btn-sm btn-ghost rounded-xl" :disabled="isMutating">
            {{ $t('common.cancel') }}
          </button>
          <button type="button" @click="submitForm" class="btn btn-sm btn-primary rounded-xl px-6 text-white font-semibold" :disabled="isMutating">
            <span v-if="isMutating" class="loading loading-spinner loading-xs"></span>
            {{ form.canvasId ? $t('common.save') :$t('canvas.createCanvas') }}
          </button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop"><button @click="closeFormModal">close</button></form>
    </dialog>

    <!-- Delete Confirmation Modal -->
    <dialog ref="deleteModal" class="modal z-[200]">
      <div class="modal-box rounded-2xl border border-base-300 p-6">
        <h3 class="font-bold text-lg text-error flex items-center gap-2">
          <Icon icon="lucide:alert-triangle" class="w-5 h-5" />
          {{ $t('canvas.deleteCanvas') }}
        </h3>
        <p class="py-3 text-sm text-base-content/80">
          {{ $t('canvas.deleteWarning', { name: canvasToDelete?.canvasName }) }}
          <br><br>
          <span class="text-xs text-base-content/60">{{ $t('canvas.deleteWarningDesc') }}</span>
        </p>
        <div class="modal-action mt-4">
          <button type="button" @click="closeDeleteModal" class="btn btn-sm btn-ghost rounded-xl" :disabled="isMutating">
            {{ $t('common.noCancel') }}
          </button>
          <button type="button" @click="confirmDelete" class="btn btn-sm btn-error text-white rounded-xl font-semibold px-6" :disabled="isMutating">
            <span v-if="isMutating" class="loading loading-spinner loading-xs"></span>
            {{ $t('common.yesDelete') }}
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
import { useFetch } from '@/composables/useFetch';
import { useMutation } from '@/composables/useMutation';
import { toast } from 'vue3-toastify';
import { Icon } from '@iconify/vue';
import { usePermissionStore } from '@/stores/usePermissionStore';
import NoAccess from '@/components/NoAccess.vue';
import TableData from '@/components/TableData.vue';
import StatCardGroup from '@/components/StatCardGroup.vue';
import { useVuelidate } from '@vuelidate/core';
import { required, maxLength, helpers } from '@vuelidate/validators';
import { useErrorHandler } from '@/composables/useErrorHandler';
const { handleError } = useErrorHandler();

const { t } = useI18n();
const mainMenuName = 'Canvas';

const permissionStore = usePermissionStore();
const { hasPermission } = permissionStore;

const { data: canvasData, isLoading: isLoadingCanvases, error: canvasError, execute: fetchCanvasesApi } = useFetch();
const { data: mappingData, execute: fetchMappingApi } = useFetch();
const { res: mutateRes, isLoading: isMutating, error: mutateError, execute: mutateApi } = useMutation();

const canvasTable = ref([]);
const roleMappings = ref([]);
const formModal = ref(null);
const deleteModal = ref(null);

const form = ref({
  canvasId: null,
  canvasName: ''
});
const canvasToDelete = ref(null);

// Dynamic bilingual summary stats
const statCardsData = computed(() => {
  const total = canvasTable.value.length;
  
  // Collect all canvas IDs assigned across all roles
  const assignedCanvasIds = new Set(roleMappings.value.flatMap(m => m.canvasIds || []));
  const assignedCount = canvasTable.value.filter(c => assignedCanvasIds.has(c.canvasId)).length;
  const unassignedCount = total - assignedCount;
  const configuredRolesCount = roleMappings.value.filter(m => (m.canvasIds || []).length > 0).length;

  return [
    {
      label: t('canvas.stats.total'),
      value: total,
      icon: 'lucide:layout-dashboard',
      color: 'primary'
    },
    {
      label: t('canvas.stats.assigned'),
      value: assignedCount,
      icon: 'lucide:shield-check',
      color: 'success',
      valueClass: 'text-success'
    },
    {
      label: t('canvas.stats.unassigned'),
      value: unassignedCount,
      icon: 'lucide:shield-alert',
      color: unassignedCount > 0 ? 'warning' : 'ghost',
      valueClass: unassignedCount > 0 ? 'text-warning' : 'text-base-content/50'
    },
    {
      label: t('canvas.stats.roles'),
      value: configuredRolesCount,
      icon: 'lucide:users',
      color: 'info',
      valueClass: 'text-info'
    }
  ];
});

const rules = computed(() => ({
  canvasName: {
    required: helpers.withMessage(t('canvas.validation.canvasNameRequired') || 'Canvas name is required', required),
    maxLength: helpers.withMessage(t('common.validation.maxLength', { len: 31 }), maxLength(31))
  }
}));

const v$ = useVuelidate(rules, form);

const tableColumns = computed(() => [
  {
    header: t('common.id'),
    accessorKey: 'canvasId',
    meta: { headerClass: 'w-16', cellClass: 'font-bold' }
  },
  {
    header: t('canvas.canvasName'),
    accessorKey: 'canvasName'
  },
  {
    header: t('common.actions'),
    id: 'actions',
    enableSorting: false,
    meta: { headerClass: 'text-right w-24', cellClass: 'text-right' }
  }
]);

const loadCanvases = async () => {
  await fetchCanvasesApi('/canvas/getall');
  if (!canvasError.value && canvasData.value) {
    canvasTable.value = canvasData.value.data || [];
  } else {
    toast.error(t('common.messages.loadError'));
  }

  await fetchMappingApi('/canvas/getallcanvasroledetail');
  if (mappingData.value) {
    roleMappings.value = mappingData.value.data || [];
  }
};

const openCreateModal = () => {
  form.value = { canvasId: null, canvasName: '' };
  v$.value.$reset();
  formModal.value.showModal();
};

const openEditModal = (canvas) => {
  form.value = { canvasId: canvas.canvasId, canvasName: canvas.canvasName };
  v$.value.$reset();
  formModal.value.showModal();
};

const closeFormModal = () => {
  formModal.value.close();
  form.value = { canvasId: null, canvasName: '' };
};

const submitForm = async () => {
  const isFormValid = await v$.value.$validate();
  if (!isFormValid) return;

  const isUpdate = !!form.value.canvasId;
  const endpoint = isUpdate ? '/canvas/update' : '/canvas/create';
  const httpMethod = isUpdate ? 'PUT' : 'POST';

  const payload = {
    canvasName: form.value.canvasName
  };

  if (isUpdate) {
    payload.canvasId = form.value.canvasId;
  }

  await mutateApi(endpoint, payload, httpMethod);

  if (!mutateError.value && mutateRes.value?.ok) {
    toast.success(isUpdate ? t('canvas.messages.renameSuccess') : t('common.messages.created'));
    closeFormModal();
    await loadCanvases();
  } else {
    toast.error(handleError(mutateError, 'common.messages.saveError'));
  }
};

const openDeleteModal = (canvas) => {
  canvasToDelete.value = canvas;
  deleteModal.value.showModal();
};

const closeDeleteModal = () => {
  deleteModal.value.close();
  canvasToDelete.value = null;
};

const confirmDelete = async () => {
  if (!canvasToDelete.value) return;

  await mutateApi(`/canvas/delete/${canvasToDelete.value.canvasId}`, null, 'DELETE');

  if (!mutateError.value && mutateRes.value?.ok) {
    toast.success(t('common.messages.deleted'));
    closeDeleteModal();
    await loadCanvases();
  } else {
    toast.error(handleError(mutateError, 'common.messages.deleteError'));
  }
};

onMounted(async () => {
  if (!hasPermission(mainMenuName, 'Display')) return;
  await loadCanvases();
});
</script>