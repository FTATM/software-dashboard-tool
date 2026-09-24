<template>
  <div v-if="hasPermission(mainMenuName, 'Display')" class="w-full h-full overflow-y-auto p-4 sm:p-6 space-y-5">

    <!-- Anchored Page Header Card with Background -->
    <div class="bg-base-100 border border-base-300 rounded-2xl p-5 shadow-xs flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
      <div class="flex items-start sm:items-center gap-3.5">
        <div class="p-3 bg-primary/10 text-primary rounded-xl flex items-center justify-center shrink-0">
          <Icon icon="lucide:shield" class="w-6 h-6" />
        </div>
        <div>
          <!-- Breadcrumbs -->
          <div class="flex items-center gap-1.5 text-xs font-semibold text-base-content/50 uppercase tracking-wider mb-0.5">
            <span>{{ $t('menu.management') }}</span>
            <Icon icon="lucide:chevron-right" class="w-3.5 h-3.5" />
            <span class="text-primary">{{ $t('role.title') }}</span>
          </div>

          <!-- Title & Subtitle -->
          <div class="flex items-center gap-2.5">
            <h1 class="m-0 text-xl sm:text-2xl font-black text-base-content tracking-tight">
              {{ $t('role.title') }}
            </h1>
          </div>
          <p class="mt-0.5 mb-0 text-base-content/60 text-xs font-medium">
            {{ $t('role.subtitle') }}
          </p>
        </div>
      </div>

      <!-- Header Actions -->
      <div class="flex items-center gap-2 self-end sm:self-center">
        <button 
          @click="loadData" 
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
        :data="roleTable" 
        :columns="tableColumns" 
        :initial-sorting="[{ id: 'roleId', desc: false }]"
        :is-loading="isLoading">

        <!-- Toolbar Actions -->
        <template #toolbar-actions>
          <router-link 
            :to="{ name: 'user' }" 
            class="btn btn-sm btn-outline border-base-300 bg-base-100 hover:bg-base-200 rounded-xl text-xs font-semibold shadow-xs gap-1.5">
            <Icon icon="lucide:users" class="w-4 h-4 text-base-content/70" />
            <span>{{ $t('user.title') }}</span>
          </router-link>

          <button 
            class="btn btn-sm btn-primary rounded-xl font-semibold shadow-xs hover:shadow-md transition-all gap-1 text-white" 
            @click="openCreateModal">
            <Icon icon="lucide:plus" class="w-4 h-4" />
            {{ $t('role.addRole') }}
          </button>
        </template>

        <!-- ID Cell -->
        <template #cell-roleId="{ value }">
          <span class="font-mono text-xs font-bold text-base-content/50">#{{ value }}</span>
        </template>

        <!-- Role Name Cell -->
        <template #cell-roleName="{ row }">
          <div class="flex items-center gap-2">
            <div class="w-7 h-7 rounded-lg bg-primary/10 text-primary flex items-center justify-center font-bold text-xs shrink-0">
              <Icon icon="lucide:shield-check" class="w-4 h-4" />
            </div>
            <span class="font-semibold text-base-content tracking-tight text-sm">{{ row.roleName }}</span>
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
              @click="confirmDelete(row)" 
              class="btn btn-ghost btn-xs btn-square rounded-lg text-base-content/70 hover:text-error hover:bg-error/10 transition-colors"
              :title="$t('common.delete')">
              <Icon icon="lucide:trash-2" class="w-4 h-4" />
            </button>
          </div>
        </template>
      </TableData>
    </div>

    <!-- Create/Edit Modal with Scrollable Body -->
    <dialog ref="roleModal" class="modal">
      <div class="modal-box sm:w-11/12 sm:max-w-xl p-0 overflow-hidden shadow-2xl rounded-2xl flex flex-col max-h-[85vh] border border-base-300 bg-base-100">
        <!-- Pinned Header -->
        <div class="px-6 py-4 border-b border-base-200 bg-base-100 flex justify-between items-center shrink-0">
          <div class="flex items-center gap-2.5">
            <div class="p-2 rounded-xl bg-primary/10 text-primary">
              <Icon :icon="isEditing ? 'lucide:shield-check' : 'lucide:shield-plus'" class="w-5 h-5" />
            </div>
            <div>
              <h3 class="m-0 text-lg font-bold text-base-content">
                {{ isEditing ? $t('role.editRole') :$t('role.createRole') }}
              </h3>
              <p class="m-0 text-xs text-base-content/50">
                {{ isEditing ? $t('role.editSubtitle') :$t('role.createSubtitle') }}
              </p>
            </div>
          </div>
          <button class="btn btn-sm btn-circle btn-ghost" @click="closeModal" :disabled="isSaving">
            <Icon icon="lucide:x" class="w-4 h-4" />
          </button>
        </div>

        <!-- Form Wrapper -->
        <form @submit.prevent="submitForm" autocomplete="off" class="flex flex-col flex-1 overflow-hidden bg-base-100">
          <div class="p-6 overflow-y-auto flex-1 flex flex-col gap-4">

            <!-- Role Name Input -->
            <label class="form-control w-full">
              <div class="label pb-1 flex justify-between items-center">
                <span class="label-text font-semibold text-xs uppercase tracking-wider text-base-content/70">
                  {{ $t('common.roleName') }}
                  <span class="text-error ml-0.5">*</span>
                </span>
                <span class="label-text-alt text-base-content/50 font-mono text-[11px]">{{ form.roleName?.length || 0 }}/31</span>
              </div>
              <input 
                type="text" 
                v-model="form.roleName" 
                maxlength="31" 
                class="input input-sm h-10 input-bordered w-full rounded-xl"
                :placeholder="$t('role.roleNamePlaceholder')" 
                @blur="v$.roleName.$touch()"
                :class="{ 'input-error': v$.roleName.$error }" />
              <div class="label px-1 py-0.5 min-h-[20px]">
                <span v-if="v$.roleName.$error" class="label-text-alt text-error font-medium text-xs">
                  {{ v$.roleName.$errors[0].$message }}
                </span>
              </div>
            </label>

            <!-- Permissions Checklist Box -->
            <div class="flex flex-col gap-2">
              <div class="flex items-center justify-between pb-1 border-b border-base-200">
                <span class="font-bold text-xs uppercase tracking-wider text-base-content/70">
                  {{ $t('role.permissions') }}
                </span>
                <span class="text-xs text-base-content/50 font-mono">
                  {{ form.selectedPermissions.length }} selected
                </span>
              </div>

              <!-- Loading State -->
              <div v-if="isLoadingMenuAvailable || isLoadingRole" class="flex justify-center py-8">
                <span class="loading loading-spinner loading-md text-primary"></span>
              </div>

              <!-- Permissions Matrix -->
              <div v-else class="flex flex-col gap-3 max-h-72 overflow-y-auto pr-1">
                <div 
                  v-for="menu in menuAvailables" 
                  :key="menu.menuId"
                  class="bg-base-200/50 p-3.5 rounded-xl border border-base-300 flex flex-col gap-2">
                  
                  <div class="flex justify-between items-center border-b border-base-300 pb-2">
                    <span class="font-bold text-sm text-base-content">{{ getMenuTranslation(menu.menuName) }}</span>
                    <button 
                      type="button" 
                      class="btn btn-xs rounded-lg font-medium transition-all"
                      :class="isAllSelected(menu) ? 'btn-ghost text-error hover:bg-error/10' : 'btn-ghost text-primary hover:bg-primary/10'"
                      @click="toggleSelectAll(menu)">
                      {{ isAllSelected(menu) ? $t('common.deselectAll') :$t('common.selectAll') }}
                    </button>
                  </div>

                  <!-- Flat Actions -->
                  <div v-if="menu.availableActions && menu.availableActions.length > 0" class="flex flex-wrap gap-3 py-1">
                    <label 
                      v-for="action in getSortedActions(menu.availableActions)" 
                      :key="action.actionId"
                      class="cursor-pointer label p-0 flex items-center gap-1.5 text-xs text-base-content/80 hover:text-base-content select-none">
                      <input 
                        type="checkbox" 
                        class="checkbox checkbox-xs checkbox-primary rounded-md"
                        :value="`${menu.menuId}-${action.actionId}`" 
                        v-model="form.selectedPermissions" />
                      <span class="font-medium">{{ getActionTranslation(action.actionName) }}</span>
                    </label>
                  </div>

                  <!-- Nested Submenus -->
                  <div v-if="menu.submenus && menu.submenus.length > 0" class="flex flex-col gap-2.5 pt-1 border-t border-base-300/60">
                    <div v-for="sub in menu.submenus" :key="sub.menuId" class="pl-2 border-l-2 border-primary/40 space-y-1">
                      <span class="font-semibold text-xs text-base-content/70 block">
                        {{ getMenuTranslation(sub.menuName) }}
                      </span>
                      <div class="flex flex-wrap gap-3">
                        <label 
                          v-for="action in getSortedActions(sub.availableActions)" 
                          :key="action.actionId"
                          class="cursor-pointer label p-0 flex items-center gap-1.5 text-xs text-base-content/80 hover:text-base-content select-none">
                          <input 
                            type="checkbox" 
                            class="checkbox checkbox-xs checkbox-primary rounded-md"
                            :value="`${sub.menuId}-${action.actionId}`" 
                            v-model="form.selectedPermissions" />
                          <span class="font-medium">{{ getActionTranslation(action.actionName) }}</span>
                        </label>
                      </div>
                    </div>
                  </div>
                </div>
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
              {{ isEditing ? $t('common.save') :$t('role.createRole') }}
            </button>
          </div>
        </form>
      </div>

      <form method="dialog" class="modal-backdrop" @click="closeModal"><button>close</button></form>
    </dialog>

    <!-- Delete Modal -->
    <dialog ref="deleteModal" class="modal z-[200]">
      <div class="modal-box rounded-2xl border border-base-300 p-6">
        <h3 class="font-bold text-lg text-error flex items-center gap-2">
          <Icon icon="lucide:alert-triangle" class="w-5 h-5" />
          {{ $t('common.confirmDelete') }}
        </h3>
        <p class="py-3 text-sm text-base-content/80">
          {{ $t('role.deleteWarning', { name: roleToDelete?.roleName }) }}
        </p>
        <div class="modal-action mt-4">
          <button class="btn btn-sm btn-ghost rounded-xl" @click="closeDeleteModal" :disabled="isDeleting">
            {{ $t('common.cancel') }}
          </button>
          <button class="btn btn-sm btn-error text-white rounded-xl font-semibold px-6" @click="executeDelete" :disabled="isDeleting">
            <span v-if="isDeleting" class="loading loading-spinner loading-xs"></span>
            {{ $t('common.delete') }}
          </button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop" @click="closeDeleteModal"><button>close</button></form>
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
import StatCardGroup from '@/components/StatCardGroup.vue';
import { useErrorHandler } from '@/composables/useErrorHandler';
const { handleError } = useErrorHandler();

const { t } = useI18n();
const mainMenuName = 'Role';

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

// Dynamic bilingual summary stats
const statCardsData = computed(() => {
  const totalRoles = roleTable.value.length;
  const totalModules = menuAvailables.value.length;
  
  let totalActions = 0;
  menuAvailables.value.forEach(menu => {
    totalActions += (menu.availableActions?.length || 0);
    if (menu.submenus) {
      menu.submenus.forEach(sub => {
        totalActions += (sub.availableActions?.length || 0);
      });
    }
  });

  return [
    {
      label: t('role.stats.total'),
      value: totalRoles,
      icon: 'lucide:shield',
      color: 'primary'
    },
    {
      label: t('role.stats.modules'),
      value: totalModules,
      icon: 'lucide:layout-grid',
      color: 'info',
      valueClass: 'text-info'
    },
    {
      label: t('role.stats.permissions'),
      value: totalActions,
      icon: 'lucide:key',
      color: 'success',
      valueClass: 'text-success'
    }
  ];
});

const tableColumns = computed(() => [
  { header: t('common.id'), accessorKey: 'roleId', meta: { headerClass: 'w-16', cellClass: 'font-bold' } },
  { header: t('common.roleName'), accessorKey: 'roleName', meta: { headerClass: 'w-48', cellClass: 'font-bold' } },
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

const getMenuTranslation = (rawMenuName) => {
  const i18nKey = menuTranslationMap[rawMenuName];
  if (i18nKey) {
    return t(`menu.${i18nKey}`);
  }
  return rawMenuName;
};

const getActionTranslation = (rawActionName) => {
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
      return orderA - orderB;
    }
    
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
  await roleDeleteApi(`/role/delete/${roleToDelete.value.roleId}`, null, 'DELETE');

  if (!roleDeleteError.value) {
    toast.success(t('common.messages.deleted') || 'Role deleted successfully');
    await loadData();
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