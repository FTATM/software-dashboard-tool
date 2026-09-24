<template>
  <div v-if="hasPermission(mainMenuName, 'Display')" class="w-full h-full overflow-y-auto p-4">
    <!-- Page Header Card -->
    <div
      class="bg-base-100 shadow-sm rounded-box border border-base-200 p-6 mb-6 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
      <div class="flex items-center gap-4">
        <div class="p-3 bg-primary/10 text-primary rounded-xl flex items-center justify-center">
          <Icon icon="lucide:users" class="w-7 h-7" />
        </div>
        <div>
          <h2 class="m-0 text-2xl font-extrabold text-base-content tracking-tight">{{ $t('user.title') }}</h2>
          <p class="mt-1 mb-0 text-base-content/60 text-sm font-medium">{{ $t('user.subtitle') }}</p>
        </div>
      </div>
    </div>

    <!-- Users Table -->
    <TableData :data="userTable" :columns="tableColumns" :initial-sorting="[{ id: 'userId', desc: false }]"
      :is-loading="isLoading">
      <template #toolbar-actions>
        <button class="btn btn-primary shadow-sm hover:shadow-md transition-all" @click="openCreateModal">
          <Icon icon="lucide:plus" class="w-5 h-5" />
          {{ $t('user.addUser') }}
        </button>
      </template>

      <template #cell-firstName="{ row }">
        <span class="font-medium">{{ row.firstName }} {{ row.lastName }}</span>
      </template>

      <template #cell-roleId="{ value }">
        <span class="badge badge-outline badge-primary badge-sm font-semibold">
          {{ getRoleName(value) }}
        </span>
      </template>

      <template #cell-active="{ value }">
        <span :class="['badge', value ? 'badge-success text-success-content' : 'badge-ghost']">
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

    <!-- Create/Edit Modal -->
    <dialog ref="userModal" class="modal">
      <div class="modal-box sm:w-11/12 sm:max-w-xl p-0 overflow-hidden shadow-2xl">
        <div class="px-6 py-5 border-b border-base-200 bg-base-100 flex justify-between items-center">
          <h3 class="m-0 text-xl font-extrabold text-base-content">
            {{ isEditing ? $t('user.editUser') : $t('user.createUser') }}
          </h3>
          <button class="btn btn-sm btn-circle btn-ghost" @click="closeModal">
            <Icon icon="lucide:x" class="w-4 h-4" />
          </button>
        </div>

        <form @submit.prevent="submitForm" autocomplete="off" class="p-6 bg-base-100">
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-x-5 gap-y-1">
            <!-- First Name -->
            <label class="form-control w-full">
              <div class="label pb-1 flex justify-between">
                <span class="label-text font-semibold">{{ $t('user.firstName') }}</span>
                <span class="label-text-alt text-base-content/60 font-mono">{{ form.firstName?.length || 0 }}/50</span>
              </div>
              <input type="text" v-model="form.firstName" maxlength="50" placeholder="Jane"
                @blur="v$.firstName.$touch()"
                :class="['input input-bordered w-full', { 'input-error': v$.firstName.$error }]" />
              <div class="label px-1 py-1 h-6">
                <span v-if="v$.firstName.$error" class="label-text-alt text-error font-medium">
                  {{ v$.firstName.$errors[0].$message }}
                </span>
              </div>
            </label>

            <!-- Last Name -->
            <label class="form-control w-full">
              <div class="label pb-1 flex justify-between">
                <span class="label-text font-semibold">{{ $t('user.lastName') }}</span>
                <span class="label-text-alt text-base-content/60 font-mono">{{ form.lastName?.length || 0 }}/50</span>
              </div>
              <input type="text" v-model="form.lastName" maxlength="50" placeholder="Doe" @blur="v$.lastName.$touch()"
                :class="['input input-bordered w-full', { 'input-error': v$.lastName.$error }]" />
              <div class="label px-1 py-1 h-6">
                <span v-if="v$.lastName.$error" class="label-text-alt text-error font-medium">{{
                  v$.lastName.$errors[0].$message }}</span>
              </div>
            </label>

            <!-- Email -->
            <label class="form-control w-full">
              <div class="label pb-1"><span class="label-text font-semibold">{{ $t('user.email') }}</span></div>
              <input type="email" v-model="form.email" placeholder="jane@example.com" @blur="v$.email.$touch()"
                :class="['input input-bordered w-full', { 'input-error': v$.email.$error }]" />
              <div class="label px-1 py-1 h-6">
                <span v-if="v$.email.$error" class="label-text-alt text-error font-medium">{{
                  v$.email.$errors[0].$message
                  }}</span>
              </div>
            </label>

            <!-- Telephone -->
            <label class="form-control w-full">
              <div class="label pb-1"><span class="label-text font-semibold">{{ $t('user.phone') }}</span></div>
              <input type="tel" v-model="form.tel" placeholder="+66 81 234 5678" class="input input-bordered w-full" />
              <div class="label px-1 py-1 h-6"></div>
            </label>

            <!-- Username -->
            <label class="form-control w-full sm:col-span-2">
              <div class="label pb-1 flex justify-between items-end">
                <div>
                  <span class="label-text font-semibold">{{ $t('user.username') }}</span>
                  <span v-if="isEditing" class="badge badge-neutral badge-sm ml-2">{{ $t('common.readOnly') }}</span>
                </div>
                <span class="label-text-alt text-base-content/60 font-mono">{{ form.username?.length || 0 }}/31</span>
              </div>
              <input type="text" v-model="form.username" maxlength="31" placeholder="jdoe" @blur="v$.username.$touch()"
                :disabled="isEditing" autocomplete="none"
                :class="['input input-bordered w-full disabled:bg-base-200/50 disabled:text-base-content/60', { 'input-error': v$.username.$error }]" />
              <div class="label px-1 py-1 h-6">
                <span v-if="v$.username.$error" class="label-text-alt text-error font-medium">
                  {{ v$.username.$errors[0].$message }}</span>
              </div>
            </label>

            <!-- Password -->
            <label class="form-control w-full sm:col-span-2">
              <div class="label pb-1">
                <span class="label-text font-semibold">{{ $t('user.password') }}</span>
                <span v-if="isEditing" class="label-text-alt text-info font-medium">{{ $t('user.passwordHint') }}</span>
              </div>
              <input type="password" v-model="form.password" @blur="v$.password.$touch()" placeholder="••••••••"
                autocomplete="new-password" spellcheck="false"
                :class="['input input-bordered w-full', { 'input-error': v$.password.$error }]" />
              <div class="label px-1 py-1 h-6">
                <span v-if="v$.password.$error" class="label-text-alt text-error font-medium">{{
                  v$.password.$errors[0].$message }}</span>
              </div>
            </label>

            <!-- Line Token Input -->
            <label class="form-control w-full sm:col-span-2 mb-2">
              <div class="label pb-1">
                <span class="label-text font-semibold">{{ $t('user.lineToken') }}</span>
              </div>
              <div class="flex gap-2">
                <input type="text" :value="form.lineUserToken ? `${form.lineUserToken.slice(0, 8)}...` : ''"
                  :placeholder="$t('user.lineToken')" disabled
                  class="input input-bordered w-full disabled:bg-base-200/50 disabled:text-base-content/60" />

                <!-- Unlink button sets local flag and clears field without calling an API yet -->
                <button v-if="isEditing && form.lineUserToken" type="button" @click="handleUnlinkLine"
                  class="btn btn-error btn-outline shrink-0 gap-1.5">
                  <Icon icon="lucide:unlink" class="w-4 h-4" />
                  {{ $t('user.unlinkLine') || 'Unlink LINE' }}
                </button>

                <!-- Connect LINE / QR Code Button -->
                <button type="button" @click="openLineQrCode"
                  class="btn bg-[#06c755] hover:bg-[#05a546] text-white border-none shrink-0">
                  <Icon icon="lucide:qr-code" class="w-5 h-5 mr-1" />
                  {{ $t('user.connectLine') }}
                </button>
              </div>
            </label>
          </div>

          <!-- Role Assignment -->
          <label class="form-control w-full sm:col-span-2 relative">
            <div class="label pb-1">
              <span class="label-text font-semibold">{{ $t('user.assignedRole') }}</span>
              <span class="label-text-alt text-error">*</span>
            </div>
            <SearchableDropdown v-model="form.roleId" :options="Array.from(rolesMaster.values())" label-key="roleName"
              value-key="roleId" :placeholder="$t('common.searchRole')" :error="v$.roleId.$error"
              @blur="v$.roleId.$touch()" />
            <div class="label px-1 py-1 h-6">
              <span v-if="v$.roleId.$error" class="label-text-alt text-error font-medium">{{
                v$.roleId.$errors[0].$message
                }}</span>
            </div>
          </label>

          <!-- Active Toggle -->
          <div class="mt-2 p-5 bg-base-200/40 rounded-xl border border-base-200 flex items-center justify-between">
            <div>
              <p class="font-bold text-base-content m-0 text-sm">{{ $t('user.accountStatus') }}</p>
              <p class="text-xs text-base-content/60 m-0 mt-1">{{ $t('user.accountStatusDesc') }}</p>
            </div>
            <input type="checkbox" v-model="form.active" class="toggle toggle-success toggle-lg" />
          </div>

          <!-- Footer -->
          <div class="border-t border-base-200 mt-6 pt-5 flex justify-end gap-3">
            <button type="button" class="btn btn-ghost" @click="closeModal">{{ $t('common.cancel') }}</button>
            <button type="submit" class="btn btn-primary px-8">
              {{ isEditing ? $t('common.save') : $t('user.createUser') }}
            </button>
          </div>
        </form>
      </div>
      <form method="dialog" class="modal-backdrop"><button @click="closeModal">close</button></form>
    </dialog>

    <!-- Line QR Code Modal -->
    <dialog ref="lineQrModal" class="modal z-[210]">
      <div class="modal-box max-w-sm p-6 text-center shadow-2xl">
        <div class="flex justify-between items-center mb-4">
          <div class="flex items-center gap-2">
            <Icon icon="bi:line" class="w-6 h-6 text-[#06c755]" />
            <h3 class="text-lg font-bold text-base-content">{{ $t('user.lineQrTitle') }}</h3>
          </div>
          <button class="btn btn-sm btn-circle btn-ghost" @click="closeLineQrModal">
            <Icon icon="lucide:x" class="w-4 h-4" />
          </button>
        </div>

        <p class="text-xs text-base-content/60 mb-4">
          {{ $t('user.lineQrSubtitle') }}
        </p>

        <!-- QR Display if URL exists -->
        <template v-if="lineAddFriendUrl">
          <div class="flex justify-center p-4 bg-white rounded-2xl border border-base-300 w-fit mx-auto shadow-inner">
            <QrcodeVue :value="lineAddFriendUrl" :size="190" level="H" render-as="svg" />
          </div>

          <div class="mt-4 pt-3 border-t border-base-200 space-y-2">
            <p class="text-xs font-semibold text-base-content/70">
              {{ $t('user.lineOfficialAccount') }} <span class="text-primary font-mono font-bold">{{ botHandle }}</span>
            </p>

            <a :href="lineAddFriendUrl" target="_blank" rel="noopener noreferrer"
              class="btn bg-[#06c755] hover:bg-[#05a546] text-white btn-sm w-full gap-2 mt-1 border-none shadow-sm">
              <Icon icon="bi:line" class="w-4 h-4" />
              {{ $t('user.openDirectlyInLine') }}
            </a>
          </div>
        </template>

        <!-- Fallback if URL is not configured -->
        <div v-else class="p-6 bg-base-200/60 border border-base-300 rounded-xl text-center">
          <Icon icon="lucide:alert-circle" class="w-8 h-8 text-warning mx-auto mb-2" />
          <p class="text-sm font-medium text-base-content/70">{{ $t('common.noDataAvailable') }}</p>
          <p class="text-xs text-base-content/50 mt-1">{{ $t('user.noLineConfig') }}</p>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop"><button @click="closeLineQrModal">close</button></form>
    </dialog>

    <!-- Delete Modal -->
    <dialog ref="deleteModal" class="modal z-[200]">
      <div class="modal-box">
        <h3 class="font-bold text-lg text-error flex items-center gap-2">
          <Icon icon="lucide:alert-triangle" class="w-6 h-6" /> {{ $t('common.confirmDelete') }}
        </h3>
        <p class="py-4 text-base-content/80">
          {{ $t('user.deleteWarning', { name: userToDelete?.username }) }}
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
import QrcodeVue from 'qrcode.vue';
import { useMutation } from '@/composables/useMutation';
import { useFetch } from '@/composables/useFetch';
import { useVuelidate } from '@vuelidate/core';
import { required, requiredIf, minLength, maxLength, email, helpers } from '@vuelidate/validators';
import { toast } from 'vue3-toastify';
import { Icon } from '@iconify/vue';
import { usePermissionStore } from '@/stores/usePermissionStore';
import NoAccess from '@/components/NoAccess.vue';
import SearchableDropdown from '@/components/SearchableDropdown.vue';
import TableData from '@/components/TableData.vue';
import { useErrorHandler } from '@/composables/useErrorHandler';
const { handleError } = useErrorHandler();

const { t } = useI18n();
const mainMenuName = 'User';

const { error: userAddedError, execute: userAddedApi } = useMutation();
const { error: userUpdatedError, execute: userUpdatedApi } = useMutation();
const { data: userAllFetch, isLoading, error: userAllFetchError, execute: userAllFetchApi } = useFetch();
const { error: userDeletedError, isLoading: isDeleting, execute: userDeletedApi } = useMutation();
const { data: roleData, error: roleAllError, execute: roleFetchApi } = useFetch();

const permissionStore = usePermissionStore();
const { hasPermission } = permissionStore;

const URL_OA_BOT = import.meta.env.VITE_LINE_URL_OA_BOT || '';
const userModal = ref(null);
const lineQrModal = ref(null);
const lineAddFriendUrl = ref(URL_OA_BOT);
const isEditing = ref(false);
const editingUserId = ref(null);
const userTable = ref([]);
const deleteModal = ref(null);
const userToDelete = ref(null);
const rolesMaster = ref(new Map());

const tableColumns = computed(() => [
  { header: t('common.id'), accessorKey: 'userId', meta: { headerClass: 'w-16', cellClass: 'font-bold' } },
  { header: t('user.name'), accessorKey: 'firstName' },
  { header: t('user.username'), accessorKey: 'username' },
  { header: t('user.email'), accessorKey: 'email' },
  { header: t('user.assignedRole'), accessorKey: 'roleId' },
  { header: t('user.accountStatus'), accessorKey: 'active' },
  { header: t('common.actions'), id: 'actions', enableSorting: false, meta: { headerClass: 'text-right', cellClass: 'text-right' } }
]);

const form = ref({
  firstName: '',
  lastName: '',
  email: '',
  tel: '',
  username: '',
  password: '',
  active: true,
  roleId: null,
  lineUserToken: '',
});

const rules = computed(() => ({
  firstName: {
    required: helpers.withMessage(t('user.validation.firstNameRequired'), required),
    maxLength: helpers.withMessage(t('common.validation.maxLength', { len: 50 }), maxLength(50))
  },
  lastName: {
    required: helpers.withMessage(t('user.validation.lastNameRequired'), required),
    maxLength: helpers.withMessage(t('common.validation.maxLength', { len: 50 }), maxLength(50))
  },
  email: { email: helpers.withMessage(t('user.validation.emailInvalid'), email) },
  username: {
    required: helpers.withMessage(t('user.validation.usernameRequired'), required),
    minLength: helpers.withMessage(t('user.validation.usernameMin'), minLength(4)),
    maxLength: helpers.withMessage(t('common.validation.maxLength', { len: 31 }), maxLength(31))
  },
  password: {
    required: helpers.withMessage(t('user.validation.passwordRequired'), requiredIf(() => !isEditing.value)),
    validLength: helpers.withMessage(t('user.validation.passwordMin'), (value) => {
      if (isEditing.value && (!value || value.length === 0)) return true;
      return value && value.length >= 6;
    })
  },
  roleId: { required: helpers.withMessage(t('user.validation.roleRequired'), required) }
}));

const v$ = useVuelidate(rules, form);

const getRoleName = (id) => {
  const found = rolesMaster.value.has(id) ? rolesMaster.value.get(id) : null;
  return found ? found.roleName : t('common.none');
};

const openLineQrCode = () => {
  lineQrModal.value.showModal();
};

const closeLineQrModal = () => {
  lineQrModal.value.close();
};

const botHandle = computed(() => {
  if (!lineAddFriendUrl.value) return '-';
  const parts = lineAddFriendUrl.value.split('/');
  return parts[parts.length - 1] || '-';
});

// Clear token locally and flag for deletion on submit
const handleUnlinkLine = () => {
  form.value.lineUserToken = '';
};

const openCreateModal = () => {
  isEditing.value = false;
  editingUserId.value = null;
  form.value = {
    firstName: '',
    lastName: '',
    email: '',
    tel: '',
    username: '',
    password: '',
    active: true,
    roleId: null,
    lineUserToken: '',
  };
  v$.value.$reset();
  userModal.value.showModal();
};

const openEditModal = (user) => {
  isEditing.value = true;
  editingUserId.value = user.userId;
  form.value = {
    firstName: user.firstName,
    lastName: user.lastName,
    email: user.email || '',
    tel: user.tel || '',
    username: user.username,
    password: '',
    active: user.active,
    roleId: user.roleId,
    lineUserToken: user.lineUserToken || '',
  };
  v$.value.$reset();
  userModal.value.showModal();
};

const closeModal = () => userModal.value.close();
const openDeleteModal = (user) => { userToDelete.value = user; deleteModal.value.showModal(); };
const closeDeleteModal = () => { deleteModal.value.close(); userToDelete.value = null; };

const confirmDelete = async () => {
  if (!userToDelete.value) return;
  await userDeletedApi(`/user/delete/${userToDelete.value.userId}`, null, 'DELETE');
  if (!userDeletedError.value) {
    toast.success(t('common.messages.deleteSuccess', { name: userToDelete.value.username }));
    await loadTable();
    closeDeleteModal();
  } else {
    toast.error(handleError(userDeletedError, 'common.messages.deleteError'));
  }
};

const loadTable = async () => {
  await userAllFetchApi('/user/getalldetail');
  if (!userAllFetchError.value && userAllFetch.value) {
    userTable.value = userAllFetch.value.data.map(i => ({
      userId: i.userId,
      firstName: i.firstName,
      lastName: i.lastName,
      email: i.email,
      tel: i.tel,
      username: i.username,
      active: i.active,
      roleId: i.roleId,
      lineUserToken: i.lineUserToken
    }));
  }
};

const loadRoles = async () => {
  await roleFetchApi('/role/getall');
  if (!roleAllError.value && roleData.value) {
    rolesMaster.value.clear();
    for (let i of roleData.value.data) rolesMaster.value.set(i.roleId, i);
  } else {
    toast.error(roleAllError.value?.message || t('common.messages.loadError'));
  }
};

const setupData = async () => {
  await loadTable();
  await loadRoles();
};

const submitForm = async () => {
  const isFormValid = await v$.value.$validate();
  if (!isFormValid) return;

  if (isEditing.value) {
    const payload = { ...form.value, userId: editingUserId.value };
    if (!payload.password) delete payload.password;
    delete payload.username;

    await userUpdatedApi('/user/update', payload, 'PUT');
    if (!userUpdatedError.value) {
      await loadTable();
      closeModal();
      toast.success(t('common.messages.updated'));
    } else {
      toast.error(handleError(userUpdatedError, 'common.messages.updateError'));
    }
  } else {
    await userAddedApi('/user/create', form.value, 'POST');
    if (!userAddedError.value) {
      await loadTable();
      closeModal();
      toast.success(t('common.messages.created'));
    } else {
      toast.error(handleError(userAddedError, 'common.messages.createError'));
    }
  }
};

onMounted(async () => {
  if (!hasPermission(mainMenuName, 'Display')) return;
  await setupData();
});
</script>