<template>
  <div v-if="hasPermission(mainMenuName, 'Display')" class="w-full h-full overflow-y-auto p-4 sm:p-6 space-y-5">
    
    <!-- Anchored Page Header Card with Background -->
    <div class="bg-base-100 border border-base-300 rounded-2xl p-5 shadow-xs flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
      <div class="flex items-start sm:items-center gap-3.5">
        <div class="p-3 bg-primary/10 text-primary rounded-xl flex items-center justify-center shrink-0">
          <Icon icon="lucide:users" class="w-6 h-6" />
        </div>
        <div>
          <!-- Breadcrumbs -->
          <div class="flex items-center gap-1.5 text-xs font-semibold text-base-content/50 uppercase tracking-wider mb-0.5">
            <span>{{ $t('menu.management') }}</span>
            <Icon icon="lucide:chevron-right" class="w-3.5 h-3.5" />
            <span class="text-primary">{{ $t('user.title') }}</span>
          </div>

          <!-- Title & Subtitle -->
          <div class="flex items-center gap-2.5">
            <h1 class="m-0 text-xl sm:text-2xl font-black text-base-content tracking-tight">
              {{ $t('user.title') }}
            </h1>
          </div>
          <p class="mt-0.5 mb-0 text-base-content/60 text-xs font-medium">
            {{ $t('user.subtitle') }}
          </p>
        </div>
      </div>

      <!-- Header Actions -->
      <div class="flex items-center gap-2 self-end sm:self-center">
        <button 
          @click="setupData" 
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
        :data="userTable" 
        :columns="tableColumns" 
        :initial-sorting="[{ id: 'userId', desc: false }]"
        :is-loading="isLoading">
        
        <!-- Toolbar Actions -->
        <template #toolbar-actions>
          <button 
            class="btn btn-sm btn-primary rounded-xl font-semibold shadow-xs hover:shadow-md transition-all gap-1 text-white" 
            @click="openCreateModal">
            <Icon icon="lucide:plus" class="w-4 h-4" />
            {{ $t('user.addUser') }}
          </button>
        </template>

        <!-- ID Cell -->
        <template #cell-userId="{ value }">
          <span class="font-mono text-xs font-bold text-base-content/50">#{{ value }}</span>
        </template>

        <!-- Full Name -->
        <template #cell-firstName="{ row }">
          <div class="flex items-center gap-2">
            <div class="w-7 h-7 rounded-full bg-primary/10 text-primary flex items-center justify-center font-bold text-xs shrink-0">
              {{ (row.firstName?.[0] || 'U').toUpperCase() }}
            </div>
            <div class="flex flex-col">
              <span class="font-semibold text-base-content tracking-tight text-sm">
                {{ row.firstName }} {{ row.lastName }}
              </span>
              <span v-if="row.tel" class="text-[11px] text-base-content/50 font-mono">
                {{ row.tel }}
              </span>
            </div>
          </div>
        </template>

        <!-- Username Cell -->
        <template #cell-username="{ value }">
          <span class="font-mono text-xs font-medium text-base-content/80">@{{ value }}</span>
        </template>

        <!-- Email Cell -->
        <template #cell-email="{ value }">
          <span v-if="value" class="text-xs text-base-content/80">{{ value }}</span>
          <span v-else class="text-base-content/30 text-xs font-mono">-</span>
        </template>

        <!-- Role Badge -->
        <template #cell-roleId="{ value }">
          <span class="inline-flex items-center px-2 py-0.5 rounded-md bg-base-200 border border-base-300 text-xs font-semibold text-base-content/80">
            {{ getRoleName(value) }}
          </span>
        </template>

        <!-- Live Status Dot Indicator -->
        <template #cell-active="{ value }">
          <div class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold"
            :class="value ? 'bg-success/10 text-success border border-success/20' : 'bg-base-200 text-base-content/50 border border-base-300'">
            <span class="w-1.5 h-1.5 rounded-full" :class="value ? 'bg-success animate-pulse' : 'bg-base-content/30'"></span>
            {{ value ? $t('common.active') :$t('common.disabled') }}
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
    <dialog ref="userModal" class="modal">
      <div class="modal-box sm:w-11/12 sm:max-w-xl p-0 overflow-hidden shadow-2xl rounded-2xl flex flex-col max-h-[85vh] border border-base-300 bg-base-100">
        <!-- Pinned Header -->
        <div class="px-6 py-4 border-b border-base-200 bg-base-100 flex justify-between items-center shrink-0">
          <div class="flex items-center gap-2.5">
            <div class="p-2 rounded-xl bg-primary/10 text-primary">
              <Icon :icon="isEditing ? 'lucide:pencil' : 'lucide:user-plus'" class="w-5 h-5" />
            </div>
            <div>
              <h3 class="m-0 text-lg font-bold text-base-content">
                {{ isEditing ? $t('user.editUser') :$t('user.createUser') }}
              </h3>
              <p class="m-0 text-xs text-base-content/50">
                {{ isEditing ? $t('user.editSubtitle') :$t('user.createSubtitle') }}
              </p>
            </div>
          </div>
          <button class="btn btn-sm btn-circle btn-ghost" @click="closeModal">
            <Icon icon="lucide:x" class="w-4 h-4" />
          </button>
        </div>

        <!-- Form Wrapper -->
        <form @submit.prevent="submitForm" autocomplete="off" class="flex flex-col flex-1 overflow-hidden bg-base-100">
          <div class="p-6 overflow-y-auto flex-1 flex flex-col gap-4">
            
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <!-- First Name -->
              <label class="form-control w-full">
                <div class="label pb-1 flex justify-between">
                  <span class="label-text font-semibold text-xs uppercase tracking-wider text-base-content/70">{{ $t('user.firstName') }}</span>
                  <span class="label-text-alt text-base-content/50 font-mono text-[11px]">{{ form.firstName?.length || 0 }}/50</span>
                </div>
                <input type="text" v-model="form.firstName" maxlength="50" placeholder="Jane"
                  @blur="v$.firstName.$touch()"
                  :class="['input input-sm h-10 input-bordered w-full rounded-xl', { 'input-error': v$.firstName.$error }]" />
                <div class="label px-1 py-0.5 min-h-[20px]">
                  <span v-if="v$.firstName.$error" class="label-text-alt text-error font-medium text-xs">
                    {{ v$.firstName.$errors[0].$message }}
                  </span>
                </div>
              </label>

              <!-- Last Name -->
              <label class="form-control w-full">
                <div class="label pb-1 flex justify-between">
                  <span class="label-text font-semibold text-xs uppercase tracking-wider text-base-content/70">{{ $t('user.lastName') }}</span>
                  <span class="label-text-alt text-base-content/50 font-mono text-[11px]">{{ form.lastName?.length || 0 }}/50</span>
                </div>
                <input type="text" v-model="form.lastName" maxlength="50" placeholder="Doe" @blur="v$.lastName.$touch()"
                  :class="['input input-sm h-10 input-bordered w-full rounded-xl', { 'input-error': v$.lastName.$error }]" />
                <div class="label px-1 py-0.5 min-h-[20px]">
                  <span v-if="v$.lastName.$error" class="label-text-alt text-error font-medium text-xs">
                    {{ v$.lastName.$errors[0].$message }}
                  </span>
                </div>
              </label>

              <!-- Email -->
              <label class="form-control w-full">
                <div class="label pb-1">
                  <span class="label-text font-semibold text-xs uppercase tracking-wider text-base-content/70">{{ $t('user.email') }}</span>
                </div>
                <input type="email" v-model="form.email" placeholder="jane@example.com" @blur="v$.email.$touch()"
                  :class="['input input-sm h-10 input-bordered w-full rounded-xl', { 'input-error': v$.email.$error }]" />
                <div class="label px-1 py-0.5 min-h-[20px]">
                  <span v-if="v$.email.$error" class="label-text-alt text-error font-medium text-xs">
                    {{ v$.email.$errors[0].$message }}
                  </span>
                </div>
              </label>

              <!-- Telephone -->
              <label class="form-control w-full">
                <div class="label pb-1">
                  <span class="label-text font-semibold text-xs uppercase tracking-wider text-base-content/70">{{ $t('user.phone') }}</span>
                </div>
                <input type="tel" v-model="form.tel" placeholder="+66 81 234 5678" class="input input-sm h-10 input-bordered w-full rounded-xl" />
                <div class="label px-1 py-0.5 min-h-[20px]"></div>
              </label>
            </div>

            <!-- Username -->
            <label class="form-control w-full">
              <div class="label pb-1 flex justify-between items-end">
                <div class="flex items-center gap-1.5">
                  <span class="label-text font-semibold text-xs uppercase tracking-wider text-base-content/70">{{ $t('user.username') }}</span>
                  <span v-if="isEditing" class="badge badge-ghost text-[10px] py-0.5 px-1.5">{{ $t('common.readOnly') }}</span>
                </div>
                <span class="label-text-alt text-base-content/50 font-mono text-[11px]">{{ form.username?.length || 0 }}/31</span>
              </div>
              <input type="text" v-model="form.username" maxlength="31" placeholder="jdoe" @blur="v$.username.$touch()"
                :disabled="isEditing" autocomplete="none"
                :class="['input input-sm h-10 input-bordered w-full rounded-xl disabled:bg-base-200/50 disabled:text-base-content/50', { 'input-error': v$.username.$error }]" />
              <div class="label px-1 py-0.5 min-h-[20px]">
                <span v-if="v$.username.$error" class="label-text-alt text-error font-medium text-xs">
                  {{ v$.username.$errors[0].$message }}
                </span>
              </div>
            </label>

            <!-- Password -->
            <label class="form-control w-full">
              <div class="label pb-1 flex justify-between items-center">
                <span class="label-text font-semibold text-xs uppercase tracking-wider text-base-content/70">{{ $t('user.password') }}</span>
                <span v-if="isEditing" class="label-text-alt text-info font-medium text-xs">{{ $t('user.passwordHint') }}</span>
              </div>
              <input type="password" v-model="form.password" @blur="v$.password.$touch()" placeholder="••••••••"
                autocomplete="new-password" spellcheck="false"
                :class="['input input-sm h-10 input-bordered w-full rounded-xl', { 'input-error': v$.password.$error }]" />
              <div class="label px-1 py-0.5 min-h-[20px]">
                <span v-if="v$.password.$error" class="label-text-alt text-error font-medium text-xs">
                  {{ v$.password.$errors[0].$message }}
                </span>
              </div>
            </label>

            <!-- Role Assignment -->
            <label class="form-control w-full relative">
              <div class="label pb-1">
                <span class="label-text font-semibold text-xs uppercase tracking-wider text-base-content/70">{{ $t('user.assignedRole') }}</span>
                <span class="label-text-alt text-error">*</span>
              </div>
              <SearchableDropdown v-model="form.roleId" :options="Array.from(rolesMaster.values())" label-key="roleName"
                value-key="roleId" :placeholder="$t('common.searchRole')" :error="v$.roleId.$error"
                @blur="v$.roleId.$touch()" />
              <div class="label px-1 py-0.5 min-h-[20px]">
                <span v-if="v$.roleId.$error" class="label-text-alt text-error font-medium text-xs">
                  {{ v$.roleId.$errors[0].$message }}
                </span>
              </div>
            </label>

            <!-- Line Token Integration Box -->
            <div class="p-4 bg-base-200/50 rounded-2xl border border-base-300 flex flex-col gap-2">
              <span class="font-bold text-xs uppercase tracking-wider text-base-content/70">{{ $t('user.lineToken') }}</span>
              <div class="flex gap-2">
                <input type="text" :value="form.lineUserToken ? `${form.lineUserToken.slice(0, 10)}...` : ''"
                  :placeholder="$t('user.lineToken')" disabled
                  class="input input-sm h-10 input-bordered w-full rounded-xl disabled:bg-base-100 disabled:text-base-content/60 font-mono text-xs" />

                <!-- Unlink LINE -->
                <button v-if="isEditing && form.lineUserToken" type="button" @click="handleUnlinkLine"
                  class="btn btn-sm btn-error btn-outline rounded-xl shrink-0 gap-1 text-xs">
                  <Icon icon="lucide:unlink" class="w-3.5 h-3.5" />
                  {{ $t('user.unlinkLine') || 'Unlink' }}
                </button>

                <!-- Connect LINE / QR Code Button -->
                <button type="button" @click="openLineQrCode"
                  class="btn btn-sm bg-[#06c755] hover:bg-[#05a546] text-white border-none rounded-xl shrink-0 gap-1 text-xs font-semibold">
                  <Icon icon="bi:line" class="w-4 h-4" />
                  {{ $t('user.connectLine') }}
                </button>
              </div>
            </div>

            <!-- Active Status Box -->
            <div class="p-4 bg-base-200/50 rounded-2xl border border-base-300 mt-2">
              <div class="flex items-center justify-between">
                <div>
                  <p class="font-bold text-base-content m-0 text-sm">{{ $t('user.accountStatus') }}</p>
                  <p class="text-xs text-base-content/60 m-0 mt-0.5">{{ $t('user.accountStatusDesc') }}</p>
                </div>
                <input type="checkbox" v-model="form.active" class="toggle toggle-primary toggle-sm" />
              </div>
            </div>

          </div>

          <!-- Pinned Footer -->
          <div class="border-t border-base-200 p-4 px-6 flex justify-end gap-2 shrink-0 bg-base-100">
            <button type="button" class="btn btn-sm btn-ghost rounded-xl" @click="closeModal">{{ $t('common.cancel') }}</button>
            <button type="submit" class="btn btn-sm btn-primary rounded-xl px-6 text-white font-semibold">
              {{ isEditing ? $t('common.save') :$t('user.createUser') }}
            </button>
          </div>
        </form>
      </div>
      <form method="dialog" class="modal-backdrop"><button @click="closeModal">close</button></form>
    </dialog>

    <!-- Line QR Code Modal -->
    <dialog ref="lineQrModal" class="modal z-[210]">
      <div class="modal-box max-w-sm p-6 text-center shadow-2xl rounded-2xl border border-base-300 bg-base-100">
        <div class="flex justify-between items-center mb-4">
          <div class="flex items-center gap-2">
            <Icon icon="bi:line" class="w-6 h-6 text-[#06c755]" />
            <h3 class="text-base font-bold text-base-content m-0">{{ $t('user.lineQrTitle') }}</h3>
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
            <QrcodeVue :value="lineAddFriendUrl" :size="180" level="H" render-as="svg" />
          </div>

          <div class="mt-4 pt-3 border-t border-base-200 space-y-2">
            <p class="text-xs font-semibold text-base-content/70">
              {{ $t('user.lineOfficialAccount') }} <span class="text-primary font-mono font-bold">{{ botHandle }}</span>
            </p>

            <a :href="lineAddFriendUrl" target="_blank" rel="noopener noreferrer"
              class="btn bg-[#06c755] hover:bg-[#05a546] text-white btn-sm rounded-xl w-full gap-2 mt-1 border-none shadow-xs font-semibold">
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
      <div class="modal-box rounded-2xl border border-base-300 p-6">
        <h3 class="font-bold text-lg text-error flex items-center gap-2">
          <Icon icon="lucide:alert-triangle" class="w-5 h-5" /> {{ $t('common.confirmDelete') }}
        </h3>
        <p class="py-3 text-sm text-base-content/80">
          {{ $t('user.deleteWarning', { name: userToDelete?.username }) }}
        </p>
        <div class="modal-action mt-4">
          <button type="button" @click="closeDeleteModal" class="btn btn-sm btn-ghost rounded-xl" :disabled="isDeleting">
            {{ $t('common.cancel') }}
          </button>
          <button type="button" @click="confirmDelete" class="btn btn-sm btn-error text-white rounded-xl font-semibold" :disabled="isDeleting">
            <span v-if="isDeleting" class="loading loading-spinner loading-xs"></span> {{ $t('common.delete') }}
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
import StatCardGroup from '@/components/StatCardGroup.vue';
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

// Dynamic bilingual configuration for StatCardGroup
const statCardsData = computed(() => {
  const total = userTable.value.length;
  const active = userTable.value.filter(u => u.active).length;
  const inactive = total - active;
  const lineConnected = userTable.value.filter(u => u.lineUserToken).length;

  return [
    {
      label: t('user.stats.total'),
      value: total,
      icon: 'lucide:users',
      color: 'primary'
    },
    {
      label: t('user.stats.active'),
      value: active,
      icon: 'lucide:user-check',
      color: 'success',
      valueClass: 'text-success'
    },
    {
      label: t('user.stats.disabled'),
      value: inactive,
      icon: 'lucide:user-x',
      color: inactive > 0 ? 'warning' : 'ghost',
      valueClass: inactive > 0 ? 'text-warning' : 'text-base-content/50'
    },
    {
      label: t('user.stats.lineLinked'),
      value: lineConnected,
      icon: 'bi:line',
      color: 'line', // Uses custom LINE color (#06c755)
      valueClass: 'text-[#06c755]'
    }
  ];
});

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