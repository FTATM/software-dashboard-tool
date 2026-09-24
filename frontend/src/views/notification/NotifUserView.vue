<template>
  <div v-if="hasPermission(mainMenuName, 'Display')" class="w-full h-full overflow-y-auto p-4 sm:p-6 space-y-5">
    
    <!-- Anchored Page Header Card with Background -->
    <div class="bg-base-100 border border-base-300 rounded-2xl p-5 shadow-xs flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
      <div class="flex items-start sm:items-center gap-3.5">
        <div class="p-3 bg-secondary/10 text-secondary rounded-xl flex items-center justify-center shrink-0">
          <Icon icon="lucide:bell-ring" class="w-6 h-6" />
        </div>
        <div>
          <!-- Breadcrumbs -->
          <div class="flex items-center gap-1.5 text-xs font-semibold text-base-content/50 uppercase tracking-wider mb-0.5">
            <span>{{ $t('menu.notification') }}</span>
            <Icon icon="lucide:chevron-right" class="w-3.5 h-3.5" />
            <span class="text-primary">{{ $t('notifUser.title') }}</span>
          </div>

          <!-- Title & Subtitle -->
          <div class="flex items-center gap-2.5">
            <h1 class="m-0 text-xl sm:text-2xl font-black text-base-content tracking-tight">
              {{ $t('notifUser.title') }}
            </h1>
          </div>
          <p class="mt-0.5 mb-0 text-base-content/60 text-xs font-medium">
            {{ $t('notifUser.subtitle') }}
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
        :data="notificationTable" 
        :columns="tableColumns" 
        :initial-sorting="[{ id: 'userId', desc: false }]"
        :is-loading="isLoading">

        <!-- ID Cell -->
        <template #cell-userId="{ value }">
          <span class="font-mono text-xs font-bold text-base-content/50">#{{ value }}</span>
        </template>

        <!-- Custom Cell: User Name & Username -->
        <template #cell-user="{ row }">
          <div class="flex items-center gap-2.5">
            <div class="w-7 h-7 rounded-full bg-secondary/10 text-secondary flex items-center justify-center font-bold text-xs shrink-0">
              {{ (row.firstName?.[0] || 'U').toUpperCase() }}
            </div>
            <div class="flex flex-col">
              <span class="font-semibold text-sm text-base-content tracking-tight">{{ row.firstName }} {{ row.lastName }}</span>
              <span class="text-[11px] text-base-content/50 font-mono">@{{ row.username }}</span>
            </div>
          </div>
        </template>

        <!-- Custom Cell: Contact Info -->
        <template #cell-contact="{ row }">
          <div class="flex flex-col gap-0.5 text-xs">
            <div class="flex items-center gap-1.5">
              <Icon icon="lucide:mail" class="w-3.5 h-3.5 text-base-content/40 shrink-0" />
              <span :class="row.email ? 'text-base-content/80' : 'text-base-content/30 italic'">
                {{ row.email || $t('notifUser.noEmail') }}
              </span>
            </div>
            <div class="flex items-center gap-1.5">
              <Icon icon="lucide:phone" class="w-3.5 h-3.5 text-base-content/40 shrink-0" />
              <span :class="row.tel ? 'text-base-content/80 font-mono' : 'text-base-content/30 italic'">
                {{ row.tel || $t('notifUser.noPhone') }}
              </span>
            </div>
          </div>
        </template>

        <!-- Email Status Pill -->
        <template #cell-emailActive="{ value }">
          <div class="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-semibold"
            :class="value ? 'bg-info/10 text-info border border-info/20' : 'bg-base-200 text-base-content/40 border border-base-300'">
            <span class="w-1.5 h-1.5 rounded-full" :class="value ? 'bg-info animate-pulse' : 'bg-base-content/30'"></span>
            {{ value ? $t('notifUser.enabled') :$t('common.disabled') }}
          </div>
        </template>

        <!-- SMS Status Pill -->
        <template #cell-smsActive="{ value }">
          <div class="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-semibold"
            :class="value ? 'bg-success/10 text-success border border-success/20' : 'bg-base-200 text-base-content/40 border border-base-300'">
            <span class="w-1.5 h-1.5 rounded-full" :class="value ? 'bg-success animate-pulse' : 'bg-base-content/30'"></span>
            {{ value ? $t('notifUser.enabled') :$t('common.disabled') }}
          </div>
        </template>

        <!-- LINE Status Pill -->
        <template #cell-lineActive="{ value }">
          <div class="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-semibold"
            :class="value ? 'bg-[#06c755]/10 text-[#06c755] border border-[#06c755]/20' : 'bg-base-200 text-base-content/40 border border-base-300'">
            <span class="w-1.5 h-1.5 rounded-full" :class="value ? 'bg-[#06c755] animate-pulse' : 'bg-base-content/30'"></span>
            {{ value ? $t('notifUser.enabled') :$t('common.disabled') }}
          </div>
        </template>

        <!-- Actions Slot -->
        <template #cell-actions="{ row }">
          <div class="flex justify-end">
            <button 
              @click="openEditModal(row)" 
              class="btn btn-xs btn-outline border-base-300 bg-base-100 hover:bg-base-200 rounded-lg text-xs font-medium shadow-2xs gap-1">
              <Icon icon="lucide:settings-2" class="w-3.5 h-3.5 text-base-content/70" />
              {{ $t('notifUser.configure') }}
            </button>
          </div>
        </template>
      </TableData>
    </div>

    <!-- Edit Notification Settings Modal -->
    <dialog ref="editModal" class="modal">
      <div class="modal-box sm:w-11/12 sm:max-w-lg p-0 overflow-hidden shadow-2xl rounded-2xl flex flex-col max-h-[85vh] border border-base-300 bg-base-100">
        <!-- Modal Header -->
        <div class="px-6 py-4 border-b border-base-200 bg-base-100 flex justify-between items-center shrink-0">
          <div class="flex items-center gap-2.5">
            <div class="p-2 rounded-xl bg-secondary/10 text-secondary">
              <Icon icon="lucide:bell-ring" class="w-5 h-5" />
            </div>
            <div>
              <h3 class="m-0 text-lg font-bold text-base-content">
                {{ $t('notifUser.settingsTitle') }}
              </h3>
              <p class="text-xs text-base-content/50 m-0 mt-0.5">
                {{ $t('notifUser.settingsSubtitle', { name: `${selectedUser?.firstName || ''} ${selectedUser?.lastName || ''}` }) }}
              </p>
            </div>
          </div>
          <button class="btn btn-sm btn-circle btn-ghost" @click="closeModal">
            <Icon icon="lucide:x" class="w-4 h-4" />
          </button>
        </div>

        <!-- Modal Body (Form) -->
        <form @submit.prevent="submitForm" class="flex flex-col flex-1 overflow-hidden bg-base-100">
          <div class="p-6 overflow-y-auto flex-1 flex flex-col gap-4">

            <!-- User Destination Preview Card -->
            <div class="p-4 bg-base-200/50 rounded-2xl border border-base-300 flex flex-col gap-2.5 text-xs">
              <div class="flex justify-between items-center">
                <span class="text-base-content/60 font-semibold uppercase tracking-wider text-[11px]">{{ $t('notifUser.emailDest') }}</span>
                <span :class="selectedUser?.email ? 'font-mono font-medium text-base-content' : 'text-error font-medium'">
                  {{ selectedUser?.email || $t('notifUser.missingEmail') }}
                </span>
              </div>
              <div class="h-px bg-base-300/80 w-full"></div>
              <div class="flex justify-between items-center">
                <span class="text-base-content/60 font-semibold uppercase tracking-wider text-[11px]">{{ $t('notifUser.smsDest') }}</span>
                <span :class="selectedUser?.tel ? 'font-mono font-medium text-base-content' : 'text-error font-medium'">
                  {{ selectedUser?.tel || $t('notifUser.missingPhone') }}
                </span>
              </div>
              <div class="h-px bg-base-300/80 w-full"></div>
              <div class="flex justify-between items-center">
                <span class="text-base-content/60 font-semibold uppercase tracking-wider text-[11px]">LINE TOKEN</span>
                <span :class="selectedUser?.lineUserToken ? 'font-mono font-bold text-[#06c755]' : 'text-error font-medium'">
                  {{ selectedUser?.lineUserToken ? 'Connected' : 'Not Connected' }}
                </span>
              </div>
            </div>

            <!-- Email Notification Toggle Card -->
            <div class="p-4 bg-base-200/40 rounded-2xl border border-base-300 flex items-center justify-between">
              <div class="pr-3">
                <div class="flex items-center gap-1.5">
                  <Icon icon="lucide:mail" class="w-4 h-4 text-info" />
                  <p class="font-bold text-base-content m-0 text-sm">{{ $t('notifUser.emailAlerts') }}</p>
                </div>
                <p class="text-xs text-base-content/60 m-0 mt-0.5">{{ $t('notifUser.emailAlertsDesc') }}</p>
              </div>
              <input type="checkbox" v-model="form.emailActive" :disabled="!selectedUser?.email"
                class="toggle toggle-info toggle-sm shrink-0" />
            </div>

            <!-- SMS Notification Toggle Card -->
            <div class="p-4 bg-base-200/40 rounded-2xl border border-base-300 flex items-center justify-between">
              <div class="pr-3">
                <div class="flex items-center gap-1.5">
                  <Icon icon="lucide:phone" class="w-4 h-4 text-success" />
                  <p class="font-bold text-base-content m-0 text-sm">{{ $t('notifUser.smsAlerts') }}</p>
                </div>
                <p class="text-xs text-base-content/60 m-0 mt-0.5">{{ $t('notifUser.smsAlertsDesc') }}</p>
              </div>
              <input type="checkbox" v-model="form.smsActive" :disabled="!selectedUser?.tel"
                class="toggle toggle-success toggle-sm shrink-0" />
            </div>

            <!-- LINE Notification Toggle Card -->
            <div class="p-4 bg-base-200/40 rounded-2xl border border-base-300 flex items-center justify-between">
              <div class="pr-3">
                <div class="flex items-center gap-1.5">
                  <Icon icon="bi:line" class="w-4 h-4 text-[#06c755]" />
                  <p class="font-bold text-base-content m-0 text-sm">{{ $t('notifUser.lineAlerts') || 'LINE Alerts' }}</p>
                </div>
                <p class="text-xs text-base-content/60 m-0 mt-0.5">
                  {{ $t('notifUser.lineAlertsDesc') || 'Send notifications via LINE Bot' }}
                </p>
              </div>
              <input type="checkbox" v-model="form.lineActive" :disabled="!selectedUser?.lineUserToken"
                class="toggle toggle-success toggle-sm shrink-0" />
            </div>

          </div>

          <!-- Pinned Footer -->
          <div class="border-t border-base-200 p-4 px-6 flex justify-end gap-2 shrink-0 bg-base-100">
            <button type="button" class="btn btn-sm btn-ghost rounded-xl" @click="closeModal" :disabled="isSaving">
              {{ $t('common.cancel') }}
            </button>
            <button type="submit" class="btn btn-sm btn-primary rounded-xl px-6 text-white font-semibold" :disabled="isSaving">
              <span v-if="isSaving" class="loading loading-spinner loading-xs"></span>
              {{ $t('common.save') }}
            </button>
          </div>
        </form>
      </div>

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
import StatCardGroup from '@/components/StatCardGroup.vue';
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

// Dynamic bilingual summary stats
const statCardsData = computed(() => {
  const total = notificationTable.value.length;
  const emailCount = notificationTable.value.filter(u => u.emailActive).length;
  const smsCount = notificationTable.value.filter(u => u.smsActive).length;
  const lineCount = notificationTable.value.filter(u => u.lineActive).length;

  return [
    {
      label: t('notifUser.stats.total'),
      value: total,
      icon: 'lucide:users',
      color: 'primary'
    },
    {
      label: t('notifUser.stats.email'),
      value: emailCount,
      icon: 'lucide:mail',
      color: 'info',
      valueClass: 'text-info'
    },
    {
      label: t('notifUser.stats.sms'),
      value: smsCount,
      icon: 'lucide:phone',
      color: 'success',
      valueClass: 'text-success'
    },
    {
      label: t('notifUser.stats.line'),
      value: lineCount,
      icon: 'bi:line',
      color: 'accent',
      valueClass: 'text-[#06c755]'
    }
  ];
});

const form = ref({
  userId: null,
  emailActive: false,
  smsActive: false,
  lineActive: false
});

const tableColumns = computed(() => [
  { header: t('common.id'), accessorKey: 'userId', meta: { headerClass: 'w-16', cellClass: 'font-bold' } },
  { header: t('common.user'), id: 'user', enableSorting: true, accessorFn: row => `${row.firstName} ${row.lastName}` },
  { header: t('notifUser.contactInfo'), id: 'contact', enableSorting: false },
  { header: t('notifUser.emailAlerts'), accessorKey: 'emailActive', meta: { headerClass: 'w-32 text-center', cellClass: 'text-center' } },
  { header: t('notifUser.smsAlerts'), accessorKey: 'smsActive', meta: { headerClass: 'w-32 text-center', cellClass: 'text-center' } },
  { header: t('notifUser.lineAlerts') || 'LINE Alerts', accessorKey: 'lineActive', meta: { headerClass: 'w-32 text-center', cellClass: 'text-center' } },
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
      lineUserToken: i.lineUserToken || null,
      emailActive: i.emailActive || false,
      smsActive: i.smsActive || false,
      lineActive: i.lineActive || false,
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
    smsActive: user.tel ? user.smsActive : false,
    lineActive: user.lineUserToken ? user.lineActive : false,
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
    smsActive: form.value.smsActive,
    lineActive: form.value.lineActive,
  };

  await saveApi('/notification/user/upsert', payload, 'PUT');

  if (!saveError.value) {
    toast.success(t('common.messages.updated'));
    closeModal();
    await loadData();
  } else {
    toast.error(handleError(saveError, 'common.messages.saveError'));
  }
};

onMounted(async () => {
  if (!hasPermission(mainMenuName, 'Display')) return;
  await loadData();
});
</script>