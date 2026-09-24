<template>
  <div v-if="hasPermission(mainMenuName, 'Display')" class="w-full h-full overflow-y-auto p-4">

    <!-- Page Header Card -->
    <div
      class="bg-base-100 shadow-sm rounded-box border border-base-200 p-6 mb-6 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
      <div class="flex items-center gap-4">
        <div class="p-3 bg-primary/10 text-primary rounded-xl flex items-center justify-center">
          <Icon icon="lucide:shield" class="w-7 h-7" />
        </div>
        <div>
          <h2 class="m-0 text-2xl font-extrabold text-base-content tracking-tight">{{ $t('role.title') }}</h2>
          <p class="mt-1 mb-0 text-base-content/60 text-sm font-medium">{{ $t('role.subtitle') }}</p>
        </div>
      </div>
    </div>

    <!-- Data Table -->
    <TableData :data="roleTable" :columns="tableColumns" :initial-sorting="[{ id: 'roleId', desc: false }]"
      :is-loading="isLoading">

      <template #toolbar-actions>
        <router-link :to="{ name: 'user' }" class="btn btn-ghost text-blue-600 font-bold underline mr-2">
          <Icon icon="lucide:arrow-right-from-line" class="w-5 h-5 mr-1" />
          {{ $t('user.title') }}
        </router-link>

        <button class="btn btn-primary shadow-sm hover:shadow-md transition-all" @click="openCreateModal">
          <Icon icon="lucide:plus" class="w-5 h-5 stroke-[3]" />
          {{ $t('role.addRole') }}
        </button>
      </template>

      <template #cell-actions="{ row }">
        <div class="flex justify-end gap-2">
          <button class="btn btn-sm btn-primary" @click="openEditModal(row)">
            <Icon icon="lucide:pencil" class="w-5 h-5" />
          </button>
          <button class="btn btn-sm btn-error text-white" @click="confirmDelete(row)">
            <Icon icon="lucide:trash-2" class="w-5 h-5" />
          </button>
        </div>
      </template>
    </TableData>

    <!-- Create / Edit Modal -->
    <dialog ref="roleModal" class="modal modal-bottom sm:modal-middle">
      <div class="modal-box p-0 sm:max-w-md overflow-hidden shadow-2xl">
        <div class="px-6 py-5 border-b border-base-200 bg-base-100 flex justify-between items-center">
          <h3 class="m-0 text-xl font-extrabold text-base-content flex items-center gap-2">
            <Icon :icon="isEditing ? 'lucide:shield-check' : 'lucide:shield-plus'" class="w-5 h-5 text-primary" />
            {{ isEditing ? $t('role.editRole') : $t('role.createRole') }}
          </h3>
          <button class="btn btn-sm btn-circle btn-ghost" @click="closeModal" :disabled="isSaving">
            <Icon icon="lucide:x" class="w-4 h-4" />
          </button>
        </div>

        <form @submit.prevent="submitForm" autocomplete="off" class="p-6 bg-base-100 flex flex-col gap-4">
          <!-- Role Name Input -->
          <label class="form-control w-full">
            <div class="label pb-1 flex justify-between">
              <div>
                <span class="label-text font-semibold">{{ $t('common.roleName') }}</span>
                <span class="label-text-alt text-error ml-1">*</span>
              </div>
              <!-- Added character counter -->
              <span class="label-text-alt text-base-content/60 font-mono">{{ form.roleName?.length || 0 }}/31</span>
            </div>
            <!-- Added maxlength="31" -->
            <input type="text" v-model="form.roleName" maxlength="31" class="input input-bordered w-full"
              :placeholder="$t('role.roleNamePlaceholder')" @blur="v$.roleName.$touch()"
              :class="{ 'input-error': v$.roleName.$error }" />
            <div class="label px-1 py-1 h-6">
              <span v-if="v$.roleName.$error" class="label-text-alt text-error font-medium">
                {{ v$.roleName.$errors[0].$message }}
              </span>
            </div>
          </label>

          <!-- Permissions Checklist -->
          <div class="divider text-sm font-bold text-base-content/50 mt-0 mb-0">{{ $t('role.permissions') }}</div>

          <div v-if="isLoadingMenuAvailable || isLoadingRole" class="flex justify-center py-8">
            <span class="loading loading-spinner loading-lg text-primary"></span>
          </div>

          <div v-else class="flex flex-col gap-4 max-h-64 overflow-y-auto pr-2">
            <div v-for="menu in menuAvailables" :key="menu.menuId"
              class="bg-base-200/30 p-3 rounded-lg border border-base-200">
              <div class="flex justify-between items-center mb-3 border-b border-base-300 pb-2">
                <h4 class="font-extrabold text-base-content text-lg m-0">{{ getMenuTranslation(menu.menuName) }}</h4>
                <button type="button" class="btn btn-xs"
                  :class="isAllSelected(menu) ? 'btn-ghost text-error' : 'btn-outline btn-primary'"
                  @click="toggleSelectAll(menu)">
                  {{ isAllSelected(menu) ? $t('common.deselectAll') : $t('common.selectAll') }}
                </button>
              </div>

              <!-- Flat Actions -->
              <div v-if="menu.availableActions && menu.availableActions.length > 0"
                class="flex flex-wrap gap-4 pl-2 mb-2">
                <label v-for="action in getSortedActions(menu.availableActions)" :key="action.actionId"
                  class="cursor-pointer label p-0 flex gap-2">
                  <input type="checkbox" class="checkbox checkbox-sm checkbox-primary"
                    :value="`${menu.menuId}-${action.actionId}`" v-model="form.selectedPermissions" />
                  <span class="label-text font-medium">{{ getActionTranslation(action.actionName) }}</span>
                </label>
              </div>

              <!-- Nested Submenus -->
              <div v-if="menu.submenus && menu.submenus.length > 0" class="flex flex-col gap-3 pl-2 mt-2">
                <div v-for="sub in menu.submenus" :key="sub.menuId">
                  <h5 class="font-semibold text-xs text-base-content/70 mb-1 border-l-2 border-primary pl-2">
                    {{ getMenuTranslation(sub.menuName) }}
                  </h5>
                  <div class="flex flex-wrap gap-4 pl-3">
                    <label v-for="action in getSortedActions(sub.availableActions)" :key="action.actionId"
                      class="cursor-pointer label p-0 flex gap-2">
                      <input type="checkbox" class="checkbox checkbox-sm checkbox-primary"
                        :value="`${sub.menuId}-${action.actionId}`" v-model="form.selectedPermissions" />
                      <span class="label-text font-medium">{{ getActionTranslation(action.actionName) }}</span>
                    </label>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Modal Actions -->
          <div class="modal-action mt-2 border-t border-base-200 pt-5">
            <button type="button" class="btn btn-ghost" @click="closeModal" :disabled="isSaving">
              {{ $t('common.cancel') }}
            </button>
            <button type="submit" class="btn btn-primary text-white px-8" :disabled="isSaving">
              <span v-if="isSaving" class="loading loading-spinner loading-sm"></span>
              {{ isEditing ? $t('common.save') : $t('role.createRole') }}
            </button>
          </div>
        </form>
      </div>

      <form method="dialog" class="modal-backdrop" @click="closeModal">
        <button>close</button>
      </form>
    </dialog>

    <dialog ref="deleteModal" class="modal modal-bottom sm:modal-middle">
      <div class="modal-box">
        <h3 class="font-bold text-lg text-error flex items-center gap-2">
          <Icon icon="lucide:alert-triangle" class="w-6 h-6" />
          {{ $t('common.confirmDelete') || 'Confirm Deletion' }}
        </h3>
        <p class="py-4">
          {{ $t('role.deleteWarning', { name: roleToDelete?.roleName }) || 'Are you sure you want to delete the role:' }}
        </p>
        <div class="modal-action">
          <button class="btn btn-ghost" @click="closeDeleteModal" :disabled="isDeleting">
            {{ $t('common.cancel') }}
          </button>
          <button class="btn btn-error text-white px-6" @click="executeDelete" :disabled="isDeleting">
            <span v-if="isDeleting" class="loading loading-spinner loading-sm"></span>
            {{ $t('common.delete') || 'Delete' }}
          </button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop" @click="closeDeleteModal">
        <button>close</button>
      </form>
    </dialog>
  </div>
  <NoAccess v-else />
</template>

<script setup>
import { ref, onMounted, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { Icon } from '@iconify/vue';
import { toast } from 'vue3-toastify';
import { useFetch } from '@/composables/useFetch';
import { useMutation } from '@/composables/useMutation';
import { useVuelidate } from '@vuelidate/core';
import { required, maxLength, helpers } from '@vuelidate/validators';

import { usePermissionStore } from '@/stores/usePermissionStore';
import NoAccess from '@/components/NoAccess.vue';
import TableData from '@/components/TableData.vue';
import { useErrorHandler } from '@/composables/useErrorHandler';
const { handleError } = useErrorHandler();

const { t } = useI18n();
const mainMenuName = 'Role'

const { data: roleAllFetch, isLoading, error: roleAllFetchError, execute: roleAllFetchApi } = useFetch();
const { data: menuAvailableData, isLoading: isLoadingMenuAvailable, execute: menuAvailableFetchApi } = useFetch();
const { data: roleDetailData, error: roleDetailDataError, isLoading: isLoadingRole, execute: roleDetailFetchApi } = useFetch();
const { error: roleUpsertError, execute: roleUpsertApi } = useMutation();
const { error: roleDeleteError, execute: roleDeleteApi } = useMutation();

const permissionStore = usePermissionStore();
const { hasPermission } = permissionStore;

const roleTable = ref([]);
const menuAvailables = ref([]);
const roleModal = ref(null);
const isEditing = ref(false);
const editingRoleId = ref(null);
const isSaving = ref(false);
const deleteModal = ref(null);
const roleToDelete = ref(null);
const isDeleting = ref(false);

const tableColumns = computed(() => [
  { header: t('common.id'), accessorKey: 'roleId', meta: { headerClass: 'w-16', cellClass: 'font-bold' } },
  { header: t('common.roleName'), accessorKey: 'roleName', meta: { headerClass: 'w-20', cellClass: 'font-bold' } },
  { header: t('common.actions'), id: 'actions', enableSorting: false, meta: { headerClass: 'text-right', cellClass: 'text-right' } }
]);

const form = ref({
  roleName: '',
  selectedPermissions: []
});

const menuTranslationMap = {
  'Dashboard': 'dashboard',
  'Canvas': 'canvas',
  'Canvas Design': 'canvasDesign',
  'Canvas Access': 'canvasAccess',
  'Scheduler': 'scheduler',
  'Log Report': 'logReport',
  'Notification User': 'notifUser',
  'Notification Device': 'notifDeviceRule',
  'User': 'user',
  'Role': 'role',
  'Device': 'device',
  'Device Group': 'deviceGroup'
};

// Helper function to translate Menu names
const getMenuTranslation = (rawMenuName) => {
  const i18nKey = menuTranslationMap[rawMenuName];
  if (i18nKey) {
    return t(`menu.${i18nKey}`);
  }
  return rawMenuName; // Fallback to raw DB name if not found
};

// Helper function to translate Action names
const getActionTranslation = (rawActionName) => {
  // Convert "Display" to "display" to match the common JSON key
  const safeKey = rawActionName.toLowerCase();
  return t(`common.${safeKey}`);
};

const rules = computed(() => ({
  roleName: {
    required: helpers.withMessage(t('role.validation.roleNameRequired'), required),
    maxLength: helpers.withMessage(t('common.validation.maxLength', { len: 31 }), maxLength(31))
  }
}));

const v$ = useVuelidate(rules, form);

const loadData = async () => {
  await roleAllFetchApi('/role/getall');
  if (!roleAllFetchError.value && roleAllFetch.value) {
    roleTable.value = roleAllFetch.value.data.map(r => ({
      roleId: r.roleId || r.role_id,
      roleName: r.roleName || r.role_name
    }));
  }

  await menuAvailableFetchApi('/role/getmenuavailable');
  if (menuAvailableData.value) {
    menuAvailables.value = menuAvailableData.value.data;
  }
};

const getAllPermsForMenu = (menu) => {
  const allPerms = [];
  if (menu.availableActions) {
    menu.availableActions.forEach(action => allPerms.push(`${menu.menuId}-${action.actionId}`));
  }
  if (menu.submenus) {
    menu.submenus.forEach(sub => {
      if (sub.availableActions) {
        sub.availableActions.forEach(action => allPerms.push(`${sub.menuId}-${action.actionId}`));
      }
    });
  }
  return allPerms;
};

const isAllSelected = (menu) => {
  const allPerms = getAllPermsForMenu(menu);
  if (allPerms.length === 0) return false;
  return allPerms.every(perm => form.value.selectedPermissions.includes(perm));
};

const toggleSelectAll = (menu) => {
  const allPerms = getAllPermsForMenu(menu);
  if (isAllSelected(menu)) {
    form.value.selectedPermissions = form.value.selectedPermissions.filter(p => !allPerms.includes(p));
  } else {
    allPerms.forEach(p => {
      if (!form.value.selectedPermissions.includes(p)) form.value.selectedPermissions.push(p);
    });
  }
};

const getSortedActions = (actions) => {
  if (!actions || !actions.length) return [];
  
  const orderMap = { 
    'Display': 1, 
    'Create': 2, 
    'Update': 3, 
    'Delete': 4 
  };

  return [...actions].sort((a, b) => {
    const orderA = orderMap[a.actionName] || 99;
    const orderB = orderMap[b.actionName] || 99;

    if (orderA !== orderB) {
      return orderA - orderB; // Sort by the defined priority
    }
    
    // If neither is in the orderMap (both are 99), sort alphabetically A-Z
    return a.actionName.localeCompare(b.actionName);
  });
};

const openCreateModal = () => {
  isEditing.value = false;
  editingRoleId.value = null;
  form.value = { roleName: '', selectedPermissions: [] };
  v$.value.$reset();
  roleModal.value.showModal();
};

const openEditModal = async (role) => {
  isEditing.value = true;
  editingRoleId.value = role.roleId;
  form.value = { roleName: role.roleName, selectedPermissions: [] };
  v$.value.$reset();
  roleModal.value.showModal();

  await roleDetailFetchApi(`/role/getdetailbyid/${role.roleId}`);
  if (roleDetailData.value) {
    const detail = roleDetailData.value.data;
    if (detail.rolePermissions) {
      form.value.selectedPermissions = detail.rolePermissions.map(p => `${p.menuId}-${p.actionId}`);
    } else {
      toast.error(roleDetailDataError || t('common.messages.loadError'));
    }
  }
};

const closeModal = () => {
  if (isSaving.value) return;
  roleModal.value.close();
};

const submitForm = async () => {
  const isFormValid = await v$.value.$validate();
  if (!isFormValid) return;

  isSaving.value = true;
  const formattedPermissions = form.value.selectedPermissions.map(str => {
    const [menuId, actionId] = str.split('-');
    return { menuId: parseInt(menuId), actionId: parseInt(actionId) };
  });

  const payload = {
    roleId: isEditing.value ? editingRoleId.value : 0,
    roleName: form.value.roleName,
    rolePermissions: formattedPermissions
  };

  await roleUpsertApi('/role/upsert', payload, 'POST');

  if (!roleUpsertError.value) {
    toast.success(isEditing.value ? t('common.messages.updated') : t('common.messages.created'));
    await loadData();
    isSaving.value = false;
    closeModal();
  } else {
    toast.error(handleError(roleUpsertError, 'common.messages.saveError'));
  }

  isSaving.value = false;
};

// --- Delete Functionality ---
const confirmDelete = (role) => {
  roleToDelete.value = role;
  deleteModal.value.showModal();
};

const closeDeleteModal = () => {
  if (isDeleting.value) return;
  deleteModal.value.close();
  roleToDelete.value = null;
};

const executeDelete = async () => {
  if (!roleToDelete.value) return;

  isDeleting.value = true;

  // Assuming your backend uses a DELETE method and the ID in the URL
  await roleDeleteApi(`/role/delete/${roleToDelete.value.roleId}`, null, 'DELETE');

  if (!roleDeleteError.value) {
    toast.success(t('common.messages.deleted') || 'Role deleted successfully');
    await loadData(); // Refresh the table
    closeDeleteModal();
  } else {
    toast.error(handleError(roleDeleteError, 'common.messages.deleteFailed', { item: roleToDelete.value.roleName }));
  }

  isDeleting.value = false;
};

onMounted(async () => {
  if (!hasPermission(mainMenuName, 'Display')) return;
  await loadData();
});
</script>