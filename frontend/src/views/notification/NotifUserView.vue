<template>
  <div v-if="hasPermission(mainMenuName, 'Display')" class="w-full mx-auto p-4">
    <!-- Page Header Card -->
    <div class="bg-base-100 shadow-sm rounded-box border border-base-200 p-6 mb-6 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
      <div class="flex items-center gap-4">
        <div class="p-3 bg-secondary/10 text-secondary rounded-xl flex items-center justify-center">
          <Icon icon="lucide:bell-ring" class="w-7 h-7" />
        </div>
        <div>
          <h2 class="m-0 text-2xl font-extrabold text-base-content tracking-tight">{{ $t('notifUser.title') }}</h2>
          <p class="mt-1 mb-0 text-base-content/60 text-sm font-medium">{{ $t('notifUser.subtitle') }}</p>
        </div>
      </div>
    </div>

    <!-- Notifications Table -->
    <TableData :data="notificationTable" :columns="tableColumns" :initial-sorting="[{ id: 'userId', desc: false }]" :is-loading="isLoading">
      
      <!-- Custom Cell: User Name & Username -->
      <template #cell-user="{ row }">
        <div class="flex flex-col">
          <span class="font-bold text-base-content">{{ row.firstName }} {{ row.lastName }}</span>
          <span class="text-xs text-base-content/60 font-mono mt-0.5">@{{ row.username }}</span>
        </div>
      </template>

      <!-- Custom Cell: Contact Info -->
      <template #cell-contact="{ row }">
        <div class="flex flex-col text-sm">
          <div class="flex items-center gap-1.5">
            <Icon icon="lucide:mail" class="w-3.5 h-3.5 text-base-content/50" />
            <span :class="row.email ? 'text-base-content' : 'text-base-content/40 italic'">
              {{ row.email || $t('notifUser.noEmail') }}
            </span>
          </div>
          <div class="flex items-center gap-1.5 mt-1">
            <Icon icon="lucide:phone" class="w-3.5 h-3.5 text-base-content/50" />
            <span :class="row.tel ? 'text-base-content' : 'text-base-content/40 italic'">
              {{ row.tel || $t('notifUser.noPhone') }}
            </span>
          </div>
        </div>
      </template>

      <!-- Custom Cell: Email Status Badge -->
      <template #cell-emailActive="{ value }">
        <span :class="['badge badge-sm font-semibold', value ? 'badge-info text-white' : 'badge-ghost text-base-content/40']">
          {{ value ? $t('notifUser.enabled') : $t('common.disabled') }}
        </span>
      </template>

      <!-- Custom Cell: SMS Status Badge -->
      <template #cell-smsActive="{ value }">
        <span :class="['badge badge-sm font-semibold', value ? 'badge-success text-white' : 'badge-ghost text-base-content/40']">
          {{ value ? $t('notifUser.enabled') : $t('common.disabled') }}
        </span>
      </template>

      <!-- Custom Cell: Actions Slot -->
      <template #cell-actions="{ row }">
        <div class="flex justify-end">
          <button @click="openEditModal(row)" class="btn btn-sm btn-primary">
            <Icon icon="lucide:pencil" class="w-5 h-5 mr-1" />
            {{ $t('notifUser.configure') }}
          </button>
        </div>
      </template>
    </TableData>

    <!-- Edit Notification Settings Modal -->
    <dialog ref="editModal" class="modal">
      <div class="modal-box sm:w-11/12 sm:max-w-lg p-0 overflow-hidden shadow-2xl">
        <!-- Modal Header -->
        <div class="px-6 py-5 border-b border-base-200 bg-base-100 flex justify-between items-center">
          <div>
            <h3 class="m-0 text-xl font-extrabold text-base-content">
              {{ $t('notifUser.settingsTitle') }}
            </h3>
            <p class="text-xs text-base-content/60 m-0 mt-0.5 font-medium">
              {{ $t('notifUser.settingsSubtitle', { name: `${selectedUser?.firstName || ''} ${selectedUser?.lastName || ''}` }) }}
            </p>
          </div>
          <button class="btn btn-sm btn-circle btn-ghost" @click="closeModal">
            <Icon icon="lucide:x" class="w-4 h-4" />
          </button>
        </div>

        <!-- Modal Body (Form) -->
        <form @submit.prevent="submitForm" class="p-6 bg-base-100 flex flex-col gap-4">
          
          <!-- User Summary Preview Card -->
          <div class="p-3.5 bg-base-200/50 rounded-xl border border-base-200 flex flex-col gap-1 text-sm">
            <div class="flex justify-between items-center">
              <span class="text-base-content/60 text-xs font-semibold uppercase">{{ $t('notifUser.emailDest') }}</span>
              <span :class="selectedUser?.email ? 'font-mono font-medium' : 'text-error text-xs font-medium'">
                {{ selectedUser?.email || $t('notifUser.missingEmail') }}
              </span>
            </div>
            <div class="divider my-0.5"></div>
            <div class="flex justify-between items-center">
              <span class="text-base-content/60 text-xs font-semibold uppercase">{{ $t('notifUser.smsDest') }}</span>
              <span :class="selectedUser?.tel ? 'font-mono font-medium' : 'text-error text-xs font-medium'">
                {{ selectedUser?.tel || $t('notifUser.missingPhone') }}
              </span>
            </div>
          </div>

          <!-- Email Notification Toggle Card -->
          <div class="p-4 bg-base-200/30 rounded-xl border border-base-200 flex items-center justify-between">
            <div class="pr-4">
              <div class="flex items-center gap-2">
                <Icon icon="lucide:mail" class="w-4 h-4 text-info" />
                <p class="font-bold text-base-content m-0 text-sm">{{ $t('notifUser.emailAlerts') }}</p>
              </div>
              <p class="text-xs text-base-content/60 m-0 mt-1">{{ $t('notifUser.emailAlertsDesc') }}</p>
            </div>
            <input 
              type="checkbox" 
              v-model="form.emailActive" 
              :disabled="!selectedUser?.email"
              class="toggle toggle-info toggle-md" 
            />
          </div>

          <!-- SMS Notification Toggle Card -->
          <div class="p-4 bg-base-200/30 rounded-xl border border-base-200 flex items-center justify-between">
            <div class="pr-4">
              <div class="flex items-center gap-2">
                <Icon icon="lucide:phone" class="w-4 h-4 text-success" />
                <p class="font-bold text-base-content m-0 text-sm">{{ $t('notifUser.smsAlerts') }}</p>
              </div>
              <p class="text-xs text-base-content/60 m-0 mt-1">{{ $t('notifUser.smsAlertsDesc') }}</p>
            </div>
            <input 
              type="checkbox" 
              v-model="form.smsActive" 
              :disabled="!selectedUser?.tel"
              class="toggle toggle-success toggle-md" 
            />
          </div>

          <!-- Modal Footer -->
          <div class="border-t border-base-200 mt-2 pt-4 flex justify-end gap-3">
            <button type="button" class="btn btn-ghost" @click="closeModal" :disabled="isSaving">{{ $t('common.cancel') }}</button>
            <button type="submit" class="btn btn-primary px-6" :disabled="isSaving">
              <span v-if="isSaving" class="loading loading-spinner loading-sm"></span>
              {{ $t('common.save') }}
            </button>
          </div>
        </form>
      </div>

      <!-- Clickable backdrop to close -->
      <form method="dialog" class="modal-backdrop">
        <button @click="closeModal">close</button>
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
import { usePermissionStore } from '@/stores/usePermissionStore';
import NoAccess from '@/components/NoAccess.vue';
import TableData from '@/components/TableData.vue';
import { useErrorHandler } from '@/composables/useErrorHandler';
const { handleError } = useErrorHandler();

const { t } = useI18n();
const mainMenuName = 'Notification User';

const permissionStore = usePermissionStore();
const { hasPermission } = permissionStore;

const { data: notifyData, isLoading, error: fetchError, execute: fetchApi } = useFetch();
const { isLoading: isSaving, error: saveError, execute: saveApi } = useMutation();

const notificationTable = ref([]);
const editModal = ref(null);
const selectedUser = ref(null);

const form = ref({
  userId: null,
  emailActive: false,
  smsActive: false
});

const tableColumns = computed(() => [
  { header: t('common.id'), accessorKey: 'userId', meta: { headerClass: 'w-16', cellClass: 'font-bold' } },
  { header: t('common.user'), id: 'user', enableSorting: true, accessorFn: row => `${row.firstName} ${row.lastName}` },
  { header: t('notifUser.contactInfo'), id: 'contact', enableSorting: false },
  { header: t('notifUser.emailAlerts'), accessorKey: 'emailActive', meta: { headerClass: 'w-32 text-center', cellClass: 'text-center' } },
  { header: t('notifUser.smsAlerts'), accessorKey: 'smsActive', meta: { headerClass: 'w-32 text-center', cellClass: 'text-center' } },
  { header: t('common.actions'), id: 'actions', enableSorting: false, meta: { headerClass: 'text-right w-28', cellClass: 'text-right' } }
]);

const loadData = async () => {
  await fetchApi('/notification/user/getalldetail');
  
  if (!fetchError.value && notifyData.value) {
    notificationTable.value = notifyData.value.data.map(i => ({
      userId: i.userId,
      firstName: i.firstName,
      lastName: i.lastName,
      username: i.username,
      email: i.email,
      tel: i.tel,
      emailActive: i.emailActive || false,
      smsActive: i.smsActive || false
    }));
  } else {
    toast.error(fetchError.value?.message || t('common.messages.loadError'));
  }
};

const openEditModal = (user) => {
  selectedUser.value = user;
  form.value = {
    userId: user.userId,
    emailActive: user.email ? user.emailActive : false,
    smsActive: user.tel ? user.smsActive : false
  };
  editModal.value.showModal();
};

const closeModal = () => {
  editModal.value.close();
  selectedUser.value = null;
};

const submitForm = async () => {
  const payload = {
    userId: form.value.userId,
    emailActive: form.value.emailActive,
    smsActive: form.value.smsActive
  };

  await saveApi('/notification/user/upsert', payload, 'PUT');

  if (!saveError.value) {
    toast.success(t('common.messages.updated'));
    closeModal();
    await loadData();
  } else {
    toast.error(handleError(saveError , 'common.messages.saveError'));
  }
};

onMounted(async () => {
  if (!hasPermission(mainMenuName, 'Display')) return;
  await loadData();
});
</script>