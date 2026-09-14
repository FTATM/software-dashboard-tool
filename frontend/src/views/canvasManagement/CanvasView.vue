<template>
  <NoAccess v-if="!hasPermission(mainMenuName, 'Display')" />

  <div v-else class="p-4 sm:p-6 w-full mx-auto">

    <div
      class="bg-base-100 shadow-sm rounded-box border border-base-200 p-6 mb-6 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
      <div class="flex items-center gap-4">
        <div class="p-3 bg-primary/10 text-primary rounded-xl flex items-center justify-center">
          <Icon icon="lucide:layout-dashboard" class="w-7 h-7" />
        </div>
        <div>
          <h2 class="m-0 text-2xl font-extrabold text-base-content tracking-tight">{{ $t('canvas.title') }}</h2>
          <p class="mt-1 mb-0 text-base-content/60 text-sm font-medium">{{ $t('canvas.subtitle') }}</p>
        </div>
      </div>
    </div>

    <!-- Canvas Table -->
    <TableData :data="canvasTable" :columns="tableColumns" :initial-sorting="[{ id: 'canvasId', desc: false }]"
      :is-loading="isLoadingCanvases">

      <template #toolbar-actions>
        <button @click="openCreateModal" class="btn btn-primary text-white">
          <Icon icon="lucide:plus" class="w-5 h-5 mr-1" /> {{ $t('canvas.createCanvas') }}
        </button>
      </template>

      <template #cell-canvasId="{ value }">
        <span class="font-medium text-base-content/50">{{ value }}</span>
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

    <!-- Create / Edit Modal -->
    <dialog ref="formModal" class="modal z-[200]">
      <div class="modal-box">
        <h3 class="font-bold text-lg text-base-content mb-4 flex items-center gap-2">
          <Icon :icon="form.canvasId ? 'lucide:pencil' : 'lucide:plus'" class="w-5 h-5 text-primary" />
          {{ form.canvasId ? $t('canvas.renameCanvas') : $t('canvas.createNewCanvas') }}
        </h3>

        <div class="form-control w-full">
          <label class="label pb-1 flex justify-between">
            <span class="label-text font-bold">{{ $t('canvas.canvasName') }}</span>
            <span class="label-text-alt text-base-content/60 font-mono">{{ form.canvasName?.length || 0 }}/31</span>
          </label>
          <input type="text" v-model="form.canvasName" maxlength="31" @blur="v$.canvasName.$touch()"
            :class="['input input-bordered w-full', { 'input-error': v$.canvasName.$error }]"
            :placeholder="$t('canvas.canvasNamePlaceholder')" @keyup.enter="submitForm" />
          <div class="label px-1 py-1 h-6">
            <span v-if="v$.canvasName.$error" class="label-text-alt text-error font-medium">
              {{ v$.canvasName.$errors[0].$message }}
            </span>
          </div>
        </div>

        <div class="modal-action mt-2">
          <button type="button" @click="closeFormModal" class="btn btn-ghost" :disabled="isMutating">
            {{ $t('common.cancel') }}
          </button>
          <button type="button" @click="submitForm" class="btn btn-primary text-white" :disabled="isMutating">
            <span v-if="isMutating" class="loading loading-spinner loading-sm"></span>
            {{ form.canvasId ? $t('common.save') : $t('canvas.createCanvas') }}
          </button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop"><button @click="closeFormModal">close</button></form>
    </dialog>

    <!-- Delete Confirmation Modal -->
    <dialog ref="deleteModal" class="modal z-[200]">
      <div class="modal-box">
        <h3 class="font-bold text-lg text-error flex items-center gap-2">
          <Icon icon="lucide:alert-triangle" class="w-6 h-6" />
          {{ $t('canvas.deleteCanvas') }}
        </h3>
        <p class="py-4 text-base-content/70">
          {{ $t('canvas.deleteWarning', { name: canvasToDelete?.canvasName }) }}
          <br><br>
          {{ $t('canvas.deleteWarningDesc') }}
        </p>
        <div class="modal-action">
          <button type="button" @click="closeDeleteModal" class="btn btn-ghost" :disabled="isMutating">
            {{ $t('common.noCancel') }}
          </button>
          <button type="button" @click="confirmDelete" class="btn btn-error text-white" :disabled="isMutating">
            <span v-if="isMutating" class="loading loading-spinner loading-sm"></span>
            {{ $t('common.yesDelete') }}
          </button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop"><button @click="closeDeleteModal">close</button></form>
    </dialog>

  </div>
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
import { useVuelidate } from '@vuelidate/core';
import { required, maxLength, helpers } from '@vuelidate/validators';
import { useErrorHandler } from '@/composables/useErrorHandler';
const { handleError } = useErrorHandler();

const { t } = useI18n();
const mainMenuName = 'Canvas'

const permissionStore = usePermissionStore();
const { hasPermission } = permissionStore;

const { data: canvasData, isLoading: isLoadingCanvases, error: canvasError, execute: fetchCanvasesApi } = useFetch();
const { res: mutateRes, isLoading: isMutating, error: mutateError, execute: mutateApi } = useMutation();

const canvasTable = ref([]);
const formModal = ref(null);
const deleteModal = ref(null);

const form = ref({
  canvasId: null,
  canvasName: ''
});
const canvasToDelete = ref(null);

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
    meta: { headerClass: 'text-right', cellClass: 'text-right' }
  }
]);

const loadCanvases = async () => {
  await fetchCanvasesApi('/canvas/getall');
  if (!canvasError.value && canvasData.value) {
    canvasTable.value = canvasData.value.data || [];
  } else {
    toast.error(t('cmmon.messages.loadError'));
  }
};

const openCreateModal = () => {
  form.value = { canvasId: null, canvasName: '' };
  formModal.value.showModal();
};

const openEditModal = (canvas) => {
  form.value = { canvasId: canvas.canvasId, canvasName: canvas.canvasName };
  formModal.value.showModal();
};

const closeFormModal = () => {
  formModal.value.close();
  form.value = { canvasId: null, canvasName: '' };
};

const submitForm = async () => {
  if (!form.value.canvasName.trim()) return;

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
    toast.error(handleError(mutateError , 'common.messages.saveError'));
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
    toast.error(handleError(mutateError , 'common.messages.deleteError'));
  }
};

onMounted(async () => {
  if (!hasPermission(mainMenuName, 'Display')) return;
  await loadCanvases();
});
</script>