<template>
  <div v-if="hasPermission(mainMenuName, 'Display')" class="w-full mx-auto p-4">
    <!-- Page Header Card -->
    <div
      class="bg-base-100 shadow-sm rounded-box border border-base-200 p-6 mb-6 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
      <div class="flex items-center gap-4">
        <div class="p-3 bg-primary/10 text-primary rounded-xl flex items-center justify-center">
          <Icon icon="lucide:cpu" class="w-7 h-7" />
        </div>
        <div>
          <h2 class="m-0 text-2xl font-extrabold text-base-content tracking-tight">{{ $t('device.title') }}</h2>
          <p class="mt-1 mb-0 text-base-content/60 text-sm font-medium">{{ $t('device.subtitle') }}</p>
        </div>
      </div>
    </div>

    <!-- Devices Table -->
    <TableData :data="deviceTable" :columns="tableColumns" :initial-sorting="[{ id: 'deviceId', desc: false }]"
      :is-loading="isLoading">
      <template #toolbar-actions>
        <div class="dropdown dropdown-end">
          <div tabindex="0" role="button" class="btn btn-outline btn-secondary shadow-sm mr-2 transition-all">
            <Icon icon="lucide:download" class="w-5 h-5 mr-1" />
            {{ $t('common.export') }}
          </div>
          <ul tabindex="0" class="dropdown-content z-[1] menu p-2 shadow bg-base-100 rounded-box w-32">
            <li><a @click="exportData('json')">JSON</a></li>
            <li><a @click="exportData('csv')">CSV</a></li>
            <li><a @click="exportData('excel')">Excel</a></li>
          </ul>
        </div>

        <button class="btn btn-outline btn-accent shadow-sm mr-2 transition-all" @click="openImportModal">
          <Icon icon="lucide:upload" class="w-5 h-5 mr-1" />
          {{ $t('common.import') }}
        </button>

        <button class="btn btn-primary shadow-sm hover:shadow-md transition-all" @click="openCreateModal">
          <Icon icon="lucide:plus" class="w-5 h-5 mr-1" />
          {{ $t('device.addDevice') }}
        </button>
      </template>

      <!-- Device Name with Scaling Badge -->
      <template #cell-deviceName="{ row }">
        <div class="flex items-center gap-2">
          <span class="font-bold">{{ row.deviceName }}</span>
          <span v-if="hasScaling(row)" class="badge badge-accent badge-outline badge-xs font-semibold">
            {{ $t('device.scaledBadge') }}
          </span>
        </div>
      </template>

      <!-- ⚡ Dedicated Reference Device Column -->
      <template #cell-refDeviceId="{ row, value }">
        <div v-if="row.refDeviceId" class="flex items-center gap-1.5">
          <span class="badge badge-info badge-sm font-semibold gap-1">
            <Icon icon="lucide:link-2" class="w-3.5 h-3.5" />
            {{ value }}
          </span>
        </div>
        <span v-else class="text-base-content/40 text-xs font-mono">-</span>
      </template>

      <!-- Protocol Badge -->
      <template #cell-protocol="{ value }">
        <span v-if="value" class="badge badge-outline badge-sm font-bold uppercase tracking-wider text-primary">
          {{ value }}
        </span>
        <span v-else class="badge badge-ghost badge-sm font-bold tracking-wider text-base-content/50">
          {{ $t('common.none') }}
        </span>
      </template>

      <template #cell-status="{ row }">
        <div class="flex items-center gap-3">
          <div class="badge badge-sm font-semibold uppercase tracking-wider"
            :class="row.status ? 'badge-success' : 'badge-ghost text-base-content/50'">
            {{ row.status ? $t('common.active') : $t('common.inactive') }}
          </div>
        </div>
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

    <!-- Create/Edit Modal with Scrollable Body -->
    <dialog ref="deviceModal" class="modal">
      <div class="modal-box sm:w-11/12 sm:max-w-xl p-0 overflow-hidden shadow-2xl flex flex-col max-h-[85vh]">
        <!-- Pinned Header -->
        <div class="px-6 py-5 border-b border-base-200 bg-base-100 flex justify-between items-center shrink-0">
          <h3 class="m-0 text-xl font-extrabold text-base-content">
            {{ isEditing ? $t('device.editDevice') : $t('device.createDevice') }}
          </h3>
          <button class="btn btn-sm btn-circle btn-ghost" @click="closeModal">
            <Icon icon="lucide:x" class="w-4 h-4" />
          </button>
        </div>

        <!-- Form Wrapper -->
        <form @submit.prevent="submitForm" autocomplete="off" class="flex flex-col flex-1 overflow-hidden bg-base-100">
          <!-- Scrollable Body Container -->
          <div class="p-6 overflow-y-auto flex-1">
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-x-5 gap-y-3 pb-6">

              <!-- Device Name -->
              <label class="form-control w-full sm:col-span-2">
                <div class="label pb-1 flex justify-between">
                  <span class="label-text font-semibold">{{ $t('common.deviceName') }}</span>
                  <span class="label-text-alt text-base-content/60 font-mono">
                    {{ form.deviceName?.length || 0 }}/31
                  </span>
                </div>
                <input type="text" v-model="form.deviceName" maxlength="31"
                  :placeholder="$t('device.deviceNamePlaceholder')" @blur="v$.deviceName.$touch()"
                  :class="['input input-bordered w-full', { 'input-error': v$.deviceName.$error }]" />
                <div class="label px-1 py-1 h-6">
                  <span v-if="v$.deviceName.$error" class="label-text-alt text-error font-medium">
                    {{ v$.deviceName.$errors[0].$message }}
                  </span>
                </div>
              </label>

              <!-- Reference Device with SearchableDropdown -->
              <label class="form-control w-full sm:col-span-2">
                <div class="label pb-1 flex justify-between items-center">
                  <span class="label-text font-semibold">{{ $t('device.refDevice') }}</span>
                  <!-- ⚡ Quick clear button appears only when a reference is selected -->
                  <button v-if="form.refDeviceId" type="button"
                    class="btn btn-ghost btn-xs text-error gap-1 hover:bg-error/10 h-auto py-0.5 px-2"
                    @click="form.refDeviceId = null">
                    <Icon icon="lucide:x" class="w-3.5 h-3.5" />
                    {{ $t('common.cancel') }}
                  </button>
                </div>
                <SearchableDropdown v-model="form.refDeviceId" :options="refDeviceOptions" label-key="name"
                  value-key="id" :placeholder="$t('device.refDevicePlaceholder')" />
                <div v-if="form.refDeviceId" class="label px-1 py-1">
                  <span class="label-text-alt text-info font-medium">
                    {{ $t('device.virtualDeviceProtocolHint') }}
                  </span>
                </div>
              </label>

              <!-- Protocol -->
              <label class="form-control w-full sm:col-span-2">
                <div class="label pb-1">
                  <span class="label-text font-semibold">{{ $t('common.protocol') }}</span>
                </div>
                <select v-model="form.protocol" :disabled="!!form.refDeviceId" class="select select-bordered w-full">
                  <option value="" disabled>{{ $t('common.protocolPlaceholder') }}</option>
                  <option value="none">{{ $t('common.none') }}</option>
                  <option v-for="proto in protocolList" :key="proto" :value="proto">
                    {{ proto }}
                  </option>
                </select>
              </label>

              <!-- Administrative Status -->
              <div class="sm:col-span-2 p-4 bg-base-200/50 rounded-box border border-base-200 mt-1">
                <div class="flex items-center justify-between">
                  <div>
                    <span class="font-bold block text-sm">{{ $t('device.activeStatus') }}</span>
                    <span class="text-xs text-base-content/60">{{ $t('device.activeStatusDesc') }}</span>
                  </div>
                  <input type="checkbox" v-model="form.active" class="toggle toggle-primary" />
                </div>
              </div>

              <!-- Scaling Toggle & Fields -->
              <div class="sm:col-span-2 p-4 bg-base-200/50 rounded-box border border-base-200 mt-1">
                <div class="flex items-center justify-between">
                  <div>
                    <span class="font-bold block text-sm">{{ $t('device.enableScaling') }}</span>
                    <span class="text-xs text-base-content/60">{{ $t('device.enableScalingDesc') }}</span>
                  </div>
                  <input type="checkbox" v-model="form.enableScaling" class="toggle toggle-primary" />
                </div>

                <div v-if="form.enableScaling" class="mt-4 pt-4 border-t border-base-200 flex flex-col gap-4">
                  <!-- Raw Range -->
                  <div>
                    <div class="text-xs font-bold uppercase tracking-wider text-base-content/70 mb-2">
                      {{ $t('device.rawRange') }}
                    </div>
                    <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
                      <label class="form-control w-full">
                        <div class="label pb-1">
                          <span class="label-text text-xs font-semibold">{{ $t('device.rawMin') }}</span>
                        </div>
                        <input type="number" step="any" v-model="form.rawMin"
                          :placeholder="$t('device.rawMinPlaceholder')" @blur="v$.rawMin?.$touch()"
                          :class="['input input-sm input-bordered w-full', { 'input-error': v$.rawMin?.$error }]" />
                        <div class="label px-1 py-0.5 h-5">
                          <span v-if="v$.rawMin?.$error" class="label-text-alt text-error text-xs">
                            {{ v$.rawMin.$errors[0].$message }}
                          </span>
                        </div>
                      </label>

                      <label class="form-control w-full">
                        <div class="label pb-1">
                          <span class="label-text text-xs font-semibold">{{ $t('device.rawMax') }}</span>
                        </div>
                        <input type="number" step="any" v-model="form.rawMax"
                          :placeholder="$t('device.rawMaxPlaceholder')" @blur="v$.rawMax?.$touch()"
                          :class="['input input-sm input-bordered w-full', { 'input-error': v$.rawMax?.$error }]" />
                        <div class="label px-1 py-0.5 h-5">
                          <span v-if="v$.rawMax?.$error" class="label-text-alt text-error text-xs">
                            {{ v$.rawMax.$errors[0].$message }}
                          </span>
                        </div>
                      </label>
                    </div>
                  </div>

                  <!-- EU Range -->
                  <div>
                    <div class="text-xs font-bold uppercase tracking-wider text-base-content/70 mb-2">
                      {{ $t('device.euRange') }}
                    </div>
                    <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
                      <label class="form-control w-full">
                        <div class="label pb-1">
                          <span class="label-text text-xs font-semibold">{{ $t('device.euMin') }}</span>
                        </div>
                        <input type="number" step="any" v-model="form.euMin"
                          :placeholder="$t('device.euMinPlaceholder')" @blur="v$.euMin?.$touch()"
                          :class="['input input-sm input-bordered w-full', { 'input-error': v$.euMin?.$error }]" />
                        <div class="label px-1 py-0.5 h-5">
                          <span v-if="v$.euMin?.$error" class="label-text-alt text-error text-xs">
                            {{ v$.euMin.$errors[0].$message }}
                          </span>
                        </div>
                      </label>

                      <label class="form-control w-full">
                        <div class="label pb-1">
                          <span class="label-text text-xs font-semibold">{{ $t('device.euMax') }}</span>
                        </div>
                        <input type="number" step="any" v-model="form.euMax"
                          :placeholder="$t('device.euMaxPlaceholder')" @blur="v$.euMax?.$touch()"
                          :class="['input input-sm input-bordered w-full', { 'input-error': v$.euMax?.$error }]" />
                        <div class="label px-1 py-0.5 h-5">
                          <span v-if="v$.euMax?.$error" class="label-text-alt text-error text-xs">
                            {{ v$.euMax.$errors[0].$message }}
                          </span>
                        </div>
                      </label>
                    </div>
                  </div>

                  <div class="text-[11px] text-base-content/50 italic bg-base-100 p-2 rounded border border-base-200">
                    {{ $t('device.scalingFormulaHint') }}
                  </div>
                </div>
              </div>

              <!-- Network Status (Editing Mode) -->
              <div v-if="isEditing"
                class="sm:col-span-2 p-4 bg-base-200/50 rounded-box border border-base-200 flex flex-col gap-3 mt-1">
                <div class="flex items-center justify-between">
                  <h4 class="font-bold text-sm text-base-content uppercase tracking-wider m-0">
                    {{ $t('device.networkStatus') }}
                  </h4>

                  <div v-if="form.refDeviceId" class="badge badge-info badge-sm font-semibold opacity-80">
                    {{ $t('device.virtualDeviceNoPing') }}
                  </div>
                  <div v-else-if="!form.protocol || form.protocol === 'none'"
                    class="badge badge-warning badge-sm font-semibold opacity-80">
                    {{ $t('device.cannotPingNoProtocol') }}
                  </div>
                  <button v-else type="button" @click="testConnection()" class="btn btn-xs btn-outline btn-secondary"
                    :disabled="isPinging">
                    <span v-if="isPinging" class="loading loading-spinner loading-xs"></span>
                    <Icon v-else icon="lucide:wifi" class="w-3 h-3 mr-1" />
                    {{ $t('device.testConnection') }}
                  </button>
                </div>

                <div class="flex items-center justify-between mt-2">
                  <span class="text-sm font-semibold">{{ $t('device.deviceStatus') }}</span>
                  <div class="badge gap-1 border-none shadow-sm"
                    :class="form.isConnected ? 'bg-success text-white' : 'bg-error text-white'">
                    {{ form.isConnected ? $t('common.connect') : $t('common.disconnect') }}
                  </div>
                </div>

                <div v-if="!form.isConnected" class="flex items-center justify-between">
                  <span class="text-sm font-semibold">{{ $t('device.lastSeenAt') }}</span>
                  <span class="text-sm font-mono text-base-content/70">
                    {{ formatTime(form.lastSeenAt) }}
                  </span>
                </div>
              </div>
            </div>
          </div>

          <!-- Pinned Footer -->
          <div class="border-t border-base-200 p-5 flex justify-end gap-3 shrink-0 bg-base-100">
            <button type="button" class="btn btn-ghost" @click="closeModal">{{ $t('common.cancel') }}</button>
            <button type="submit" class="btn btn-primary px-8 text-white">
              {{ isEditing ? $t('common.save') : $t('device.createDevice') }}
            </button>
          </div>
        </form>
      </div>
      <form method="dialog" class="modal-backdrop"><button @click="closeModal">close</button></form>
    </dialog>

    <!-- Delete Modal -->
    <dialog ref="deleteModal" class="modal z-[200]">
      <div class="modal-box">
        <h3 class="font-bold text-lg text-error flex items-center gap-2">
          <Icon icon="lucide:alert-triangle" class="w-6 h-6" /> {{ $t('common.confirmDelete') }}
        </h3>
        <p class="py-4 text-base-content/80">
          {{ $t('device.deleteWarning', { name: deviceToDelete?.deviceName }) }}
        </p>
        <div class="modal-action">
          <button type="button" @click="closeDeleteModal" class="btn btn-ghost" :disabled="isDeleting">
            {{ $t('common.noCancel') }}
          </button>
          <button type="button" @click="confirmDelete" class="btn btn-error text-white" :disabled="isDeleting">
            <span v-if="isDeleting" class="loading loading-spinner loading-sm"></span> {{ $t('common.yesDelete') }}
          </button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop"><button @click="closeDeleteModal">close</button></form>
    </dialog>

    <!-- Import Modal -->
    <dialog ref="importModal" class="modal">
      <div class="modal-box w-11/12 max-w-4xl p-0 overflow-hidden shadow-2xl flex flex-col max-h-[85vh]">
        <div class="px-6 py-5 border-b border-base-200 bg-base-100 flex justify-between items-center shrink-0">
          <h3 class="m-0 text-xl font-extrabold text-base-content">{{ $t('device.import.title') }}</h3>
          <button class="btn btn-sm btn-circle btn-ghost" @click="closeImportModal">
            <Icon icon="lucide:x" class="w-4 h-4" />
          </button>
        </div>
        <div class="px-6 py-4 bg-info/10 border-b border-info/20 text-sm">
          <div class="flex gap-3 items-start">
            <Icon icon="lucide:info" class="w-5 h-5 text-info shrink-0 mt-0.5" />
            <div>
              <p class="font-bold text-info-content mb-1">{{ $t('device.import.requirementsTitle') }}</p>
              <ul class="list-disc list-inside space-y-1 text-info-content/80 ml-1">
                <li><strong>{{ $t('device.import.supportedFormats') }}</strong></li>
                <li><strong>{{ $t('device.import.requiredColumns') }}</strong>
                  <code class="bg-base-100/50 text-info font-bold px-1.5 py-0.5 rounded">deviceName</code>,
                  <code class="bg-base-100/50 text-info font-bold px-1.5 py-0.5 rounded">protocol</code>,
                  <code class="bg-base-100/50 text-info font-bold px-1.5 py-0.5 rounded">status (active/inactive)</code>
                </li>
              </ul>
            </div>
          </div>
        </div>

        <div class="p-6 bg-base-100 flex-1 overflow-y-auto">
          <div class="flex items-center gap-4 mb-6">
            <input type="file" @change="handleFileSelect" accept=".csv, .json, .xlsx"
              class="file-input file-input-bordered file-input-primary w-full max-w-xs" />
            <button @click="validateImportFile" :disabled="!selectedFile || isLoadingValidateImport"
              class="btn btn-primary text-white">
              <span v-if="isLoadingValidateImport" class="loading loading-spinner loading-sm"></span>
              {{ $t('device.import.validateFile') }}
            </button>
            <div v-show="validateImportError" class="font-semibold text-error">
              <span>{{ validateImportError?.message || "Error" }}</span>
            </div>
          </div>

          <div v-if="validationResults.length > 0" class="flex flex-col gap-2">
            <div class="flex justify-between items-center px-1 text-xs">
              <span class="font-medium text-base-content/70">
                Total rows: <b>{{ validationResults.length }}</b>
              </span>
              <span v-if="invalidCount > 0" class="badge badge-error badge-sm text-white font-bold">
                {{ invalidCount }} Invalid row(s) found
              </span>
              <span v-else class="badge badge-success badge-sm text-white font-bold">
                All rows valid
              </span>
            </div>

            <div class="overflow-x-auto border border-base-200 rounded-box">
              <table class="table table-sm table-zebra w-full">
                <thead class="bg-base-200/50">
                  <tr>
                    <th>{{ $t('common.deviceName') }}</th>
                    <th>{{ $t('device.refDevice') }}</th>
                    <th>{{ $t('common.protocol') }}</th>
                    <th>{{ $t('common.active') }}</th>
                    <th>{{ $t('device.import.validation') }}</th>
                    <th>{{ $t('device.import.message') }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(row, idx) in pagedValidationResults" :key="idx" :class="{ 'bg-error/10': !row.isValid }">
                    <td>{{ row.deviceName }}</td>
                    <td>{{ row.protocol || '-' }}</td>
                    <td>
                      <span class="badge badge-sm font-semibold uppercase tracking-wider"
                        :class="row.active === false ? 'badge-ghost text-base-content/50' : 'badge-success'">
                        {{ row.active === false ? $t('common.inactive') : $t('common.active') }}
                      </span>
                    </td>
                    <td>
                      <span class="badge badge-sm" :class="row.isValid ? 'badge-success' : 'badge-error'">
                        {{ row.isValid ? $t('device.import.pass') : $t('device.import.error') }}
                      </span>
                    </td>
                    <td :class="row.isValid ? 'text-success' : 'text-error font-medium'">
                      {{ row.message }}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            <div class="flex justify-between items-center px-1 pt-1 text-xs">
              <span class="text-base-content/60">
                Page {{ currentPage }} of {{ totalPages }}
              </span>
              <div class="join">
                <button class="join-item btn btn-xs" :disabled="currentPage === 1" @click="currentPage--">‹</button>
                <button class="join-item btn btn-xs" :disabled="currentPage >= totalPages"
                  @click="currentPage++">›</button>
              </div>
            </div>
          </div>
        </div>

        <div class="border-t border-base-200 p-5 flex justify-end gap-3 shrink-0">
          <button type="button" class="btn btn-ghost" @click="closeImportModal">{{ $t('common.cancel') }}</button>
          <button @click="confirmImport" :disabled="!canConfirmImport || isImporting"
            class="btn btn-success text-white px-8">
            <span v-if="isImporting" class="loading loading-spinner loading-sm"></span>
            {{ $t('device.import.confirmImport') }}
          </button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop"><button @click="closeImportModal">close</button></form>
    </dialog>
  </div>
  <NoAccess v-else />
</template>

<script setup>
import { ref, onMounted, computed, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useMutation } from '@/composables/useMutation';
import { useFetch } from '@/composables/useFetch';
import { useVuelidate } from '@vuelidate/core';
import { required, maxLength, helpers } from '@vuelidate/validators';
import { toast } from 'vue3-toastify';
import { Icon } from '@iconify/vue';
import { usePermissionStore } from '@/stores/usePermissionStore';
import NoAccess from '@/components/NoAccess.vue';
import TableData from '@/components/TableData.vue';
import SearchableDropdown from '@/components/SearchableDropdown.vue';
import { useDownload } from '@/composables/useDownload';
import { useErrorHandler } from '@/composables/useErrorHandler';
const { handleError } = useErrorHandler();
import { useFormatter } from '@/composables/useFormatter';
const { formatTime } = useFormatter();

const { t } = useI18n();
const mainMenuName = 'Device';

const permissionStore = usePermissionStore();
const { hasPermission } = permissionStore;

const { error: deviceAddedError, execute: deviceAddedApi } = useMutation();
const { error: deviceUpdatedError, execute: deviceUpdatedApi } = useMutation();
const { data: deviceAllFetch, isLoading, error: deviceAllFetchError, execute: deviceAllFetchApi } = useFetch();
const { error: deviceDeletedError, isLoading: isDeleting, execute: deviceDeletedApi } = useMutation();
const { data: validateImportData, isLoading: isLoadingValidateImport, error: validateImportError, execute: validateImportApi } = useMutation();
const { isDownloading: isExporting, error: exportError, executeDownload } = useDownload();
const { data: protocolData, error: protocolError, execute: protocolFetchApi } = useFetch();

const { data: pingData, isLoading: isPinging, error: pingError, execute: pingApi } = useFetch();

const deviceModal = ref(null);
const isEditing = ref(false);
const editingDeviceId = ref(null);
const deviceTable = ref([]);
const deleteModal = ref(null);
const deviceToDelete = ref(null);
const protocolList = ref([]);

const importModal = ref(null);
const selectedFile = ref(null);
const isImporting = ref(false);
const validationResults = ref([]);
const currentPage = ref(1);
const pageSize = 10;

// ⚡ Added refDeviceId column directly after deviceName
const tableColumns = computed(() => [
  { header: t('common.id'), accessorKey: 'deviceId', meta: { headerClass: 'w-16', cellClass: 'font-bold' } },
  { header: t('common.deviceName'), accessorKey: 'deviceName' },
  // ⚡ Use accessorFn so TanStack Table includes the name & ID in global search
  {
    header: t('device.refDevice'),
    id: 'refDeviceId',
    accessorFn: (row) => row.refDeviceId ? getRefDeviceName(row.refDeviceId) : ''
  },
  { header: t('common.protocol'), accessorKey: 'protocol' },
  { header: t('common.status'), accessorKey: 'status' },
  { header: t('common.actions'), id: 'actions', enableSorting: false, meta: { headerClass: 'text-right', cellClass: 'text-right' } }
]);

const form = ref({
  deviceName: '',
  protocol: 'none',
  active: true,
  isConnected: false,
  lastSeenAt: null,
  refDeviceId: null,
  enableScaling: false,
  rawMin: null,
  rawMax: null,
  euMin: null,
  euMax: null
});

// Helper to look up referenced device name for the table column
const getRefDeviceName = (refId) => {
  if (!refId) return '';
  const target = deviceTable.value.find(d => d.deviceId === refId);
  return target ? `#${target.deviceId} - ${target.deviceName}` : `#${refId}`;
};

// Reference devices (physical sensors only, excludes current self)
const refDeviceOptions = computed(() => {
  return deviceTable.value
    .filter(d => {
      if (isEditing.value && d.deviceId === editingDeviceId.value) return false;
      return !d.refDeviceId; // Only physical devices can serve as source
    })
    .map(d => ({
      id: d.deviceId,
      name: `#${d.deviceId} - ${d.deviceName}`
    }));
});

// Auto-lock protocol to none when referencing another device
watch(() => form.value.refDeviceId, (newRef) => {
  if (newRef) {
    form.value.protocol = 'none';
  }
});

// Reset scaling validation errors if toggle switched off
watch(() => form.value.enableScaling, (enabled) => {
  if (!enabled) {
    v$.value.rawMin?.$reset?.();
    v$.value.rawMax?.$reset?.();
    v$.value.euMin?.$reset?.();
    v$.value.euMax?.$reset?.();
  }
});

const canConfirmImport = computed(() => {
  return validationResults.value.length > 0 && validationResults.value.every(row => row.isValid);
});

const isValidNumber = (val) => {
  if (val === null || val === undefined || val === '') return false;
  return !isNaN(Number(val));
};

const isValPresent = (v) => v !== null && v !== undefined && v !== '';

const hasScaling = (device) => {
  return isValPresent(device.rawMin) &&
    isValPresent(device.rawMax) &&
    isValPresent(device.euMin) &&
    isValPresent(device.euMax);
};

const rules = computed(() => {
  const baseRules = {
    deviceName: {
      required: helpers.withMessage(t('device.validation.deviceNameRequired'), required),
      maxLength: helpers.withMessage(t('common.validation.maxLength', { len: 31 }), maxLength(31))
    },
    rawMin: {},
    rawMax: {},
    euMin: {},
    euMax: {}
  };

  if (form.value.enableScaling) {
    baseRules.rawMin = {
      required: helpers.withMessage(t('device.validation.rawMinRequired'), required),
      validNumber: helpers.withMessage(t('device.validation.mustBeNumber'), isValidNumber)
    };
    baseRules.rawMax = {
      required: helpers.withMessage(t('device.validation.rawMaxRequired'), required),
      validNumber: helpers.withMessage(t('device.validation.mustBeNumber'), isValidNumber),
      notEqual: helpers.withMessage(t('device.validation.rawMaxEqualMin'), (val) => {
        if (!isValidNumber(val) || !isValidNumber(form.value.rawMin)) return true;
        return Number(val) !== Number(form.value.rawMin);
      })
    };
    baseRules.euMin = {
      required: helpers.withMessage(t('device.validation.euMinRequired'), required),
      validNumber: helpers.withMessage(t('device.validation.mustBeNumber'), isValidNumber)
    };
    baseRules.euMax = {
      required: helpers.withMessage(t('device.validation.euMaxRequired'), required),
      validNumber: helpers.withMessage(t('device.validation.mustBeNumber'), isValidNumber),
      notEqual: helpers.withMessage(t('device.validation.euMaxEqualMin'), (val) => {
        if (!isValidNumber(val) || !isValidNumber(form.value.euMin)) return true;
        return Number(val) !== Number(form.value.euMin);
      })
    };
  }

  return baseRules;
});

const v$ = useVuelidate(rules, form);

const sortedValidationResults = computed(() => {
  return [...validationResults.value].sort((a, b) => {
    if (!a.isValid && b.isValid) return -1;
    if (a.isValid && !b.isValid) return 1;
    return 0;
  });
});

const totalPages = computed(() => {
  return Math.ceil(sortedValidationResults.value.length / pageSize) || 1;
});

const pagedValidationResults = computed(() => {
  const start = (currentPage.value - 1) * pageSize;
  return sortedValidationResults.value.slice(start, start + pageSize);
});

const invalidCount = computed(() => {
  return validationResults.value.filter(row => !row.isValid).length;
});

const testConnection = async (isSilent = false) => {
  if (!editingDeviceId.value) return;

  if (form.value.refDeviceId || !form.value.protocol || form.value.protocol === 'none') {
    return;
  }

  await pingApi(`/device/pingdevice?deviceId=${editingDeviceId.value}`);

  if (!pingError.value && pingData.value) {
    const isOnline = !!pingData.value.data.connection;
    form.value.isConnected = isOnline;

    if (isOnline) {
      if (!isSilent) toast.success(t('device.messages.pingSuccess'));
      const tblRecord = deviceTable.value.find(d => d.deviceId === editingDeviceId.value);
      if (tblRecord) tblRecord.isConnected = true;
    } else {
      if (!isSilent) toast.warning(t('device.messages.pingOffline'));
      const tblRecord = deviceTable.value.find(d => d.deviceId === editingDeviceId.value);
      if (tblRecord) tblRecord.isConnected = false;
    }
  } else {
    if (!isSilent) toast.error(handleError(pingError, 'device.messages.pingFailed'));
  }
};

const openCreateModal = () => {
  isEditing.value = false;
  editingDeviceId.value = null;
  form.value = {
    deviceName: '',
    protocol: 'none',
    active: true,
    isConnected: false,
    lastSeenAt: null,
    refDeviceId: null,
    enableScaling: false,
    rawMin: null,
    rawMax: null,
    euMin: null,
    euMax: null
  };
  v$.value.$reset();
  deviceModal.value.showModal();
};

const openEditModal = async (device) => {
  isEditing.value = true;
  editingDeviceId.value = device.deviceId;

  const hasScalingVal = isValPresent(device.rawMin) &&
    isValPresent(device.rawMax) &&
    isValPresent(device.euMin) &&
    isValPresent(device.euMax);

  form.value = {
    deviceName: device.deviceName,
    protocol: device.protocol || 'none',
    active: device.active,
    isConnected: device.isConnected,
    lastSeenAt: device.lastSeenAt,
    refDeviceId: device.refDeviceId ?? null,
    enableScaling: hasScalingVal,
    rawMin: hasScalingVal ? device.rawMin : null,
    rawMax: hasScalingVal ? device.rawMax : null,
    euMin: hasScalingVal ? device.euMin : null,
    euMax: hasScalingVal ? device.euMax : null
  };
  v$.value.$reset();
  deviceModal.value.showModal();

  const isVirtual = !!form.value.refDeviceId;
  const hasNoProtocol = !form.value.protocol || form.value.protocol === 'none';

  if (!isVirtual && !hasNoProtocol) {
    await testConnection(true);
  }
};

const closeModal = () => deviceModal.value.close();
const openDeleteModal = (device) => { deviceToDelete.value = device; deleteModal.value.showModal(); };
const closeDeleteModal = () => { deleteModal.value.close(); deviceToDelete.value = null; };

const confirmDelete = async () => {
  if (!deviceToDelete.value) return;
  await deviceDeletedApi(`/device/delete/${deviceToDelete.value.deviceId}`, null, 'DELETE');
  if (!deviceDeletedError.value) {
    toast.success(t('common.messages.deleteSuccess', { name: deviceToDelete.value.deviceName }));
    await loadTable();
    closeDeleteModal();
  } else {
    toast.error(handleError(deviceDeletedError, 'common.messages.deleteFailed', { item: deviceToDelete.value.deviceName }));
  }
};

const loadProtocols = async () => {
  await protocolFetchApi('/device/getprotocoltype');
  if (!protocolError.value && protocolData.value) {
    protocolList.value = protocolData.value.data || [];
  } else {
    toast.error(t('common.messages.loadError'));
  }
};

const loadTable = async () => {
  await deviceAllFetchApi('/device/getalldetail');
  if (!deviceAllFetchError.value && deviceAllFetch.value) {
    deviceTable.value = [];
    for (let i of deviceAllFetch.value.data) {
      deviceTable.value.push({
        deviceId: i.deviceId,
        deviceName: i.deviceName,
        protocol: i.protocol,
        valueData: i.valueData,
        active: i.active,
        isConnected: i.isConnected,
        lastSeenAt: i.lastSeenAt,
        status: i.active,
        refDeviceId: i.refDeviceId,
        rawMin: i.rawMin,
        rawMax: i.rawMax,
        euMin: i.euMin,
        euMax: i.euMax
      });
    }
  }
};

const submitForm = async () => {
  const isFormValid = await v$.value.$validate();
  if (!isFormValid) return;

  const payload = {
    deviceName: form.value.deviceName,
    protocol: form.value.protocol === 'none' ? null : form.value.protocol,
    active: form.value.active,
    refDeviceId: form.value.refDeviceId ? Number(form.value.refDeviceId) : null,
    rawMin: form.value.enableScaling && isValPresent(form.value.rawMin) ? Number(form.value.rawMin) : null,
    rawMax: form.value.enableScaling && isValPresent(form.value.rawMax) ? Number(form.value.rawMax) : null,
    euMin: form.value.enableScaling && isValPresent(form.value.euMin) ? Number(form.value.euMin) : null,
    euMax: form.value.enableScaling && isValPresent(form.value.euMax) ? Number(form.value.euMax) : null
  };

  if (isEditing.value) {
    payload.deviceId = editingDeviceId.value;
    await deviceUpdatedApi('/device/update', payload, 'PUT');
    if (!deviceUpdatedError.value) {
      closeModal();
      toast.success(t('common.messages.updated'));
      await loadTable();
    } else {
      toast.error(handleError(deviceUpdatedError, 'common.messages.updateFailed', { item: payload.deviceName }));
    }
  } else {
    await deviceAddedApi('/device/create', [payload], 'POST');
    if (!deviceAddedError.value) {
      closeModal();
      toast.success(t('common.messages.created'));
      await loadTable();
    } else {
      toast.error(handleError(deviceAddedError, 'common.messages.createFailed', { item: payload.deviceName }));
    }
  }
};

const openImportModal = () => { validateImportError.value = ""; selectedFile.value = null; validationResults.value = []; importModal.value.showModal(); };
const closeImportModal = () => { importModal.value.close(); selectedFile.value = null; validationResults.value = []; };
const handleFileSelect = (event) => { selectedFile.value = event.target.files[0]; };

const validateImportFile = async () => {
  if (!selectedFile.value) return;
  const formData = new FormData();
  formData.append('file', selectedFile.value);
  await validateImportApi('/device/import/validate', formData, 'POST', 'form');
  if (validateImportError.value) { validationResults.value = []; return; }
  validationResults.value = validateImportData.value.data;
};

const confirmImport = async () => {
  if (!canConfirmImport.value) return;
  isImporting.value = true;
  const payload = validationResults.value.map(row => ({ deviceName: row.deviceName, protocol: row.protocol, active: row.active }));
  await deviceAddedApi('/device/create', payload, 'POST');
  if (!deviceAddedError.value) {
    toast.success(t('device.messages.importSuccess', { count: payload.length }));
    await loadTable();
    closeImportModal();
  } else {
    toast.error(handleError(deviceAddedError, 'device.messages.importFailed'));
  }
  isImporting.value = false;
};

const exportData = async (format) => {
  const extension = format === 'excel' ? 'xlsx' : format;
  const now = new Date();
  const year = now.getFullYear();
  const month = String(now.getMonth() + 1).padStart(2, '0');
  const day = String(now.getDate()).padStart(2, '0');
  const hours = String(now.getHours()).padStart(2, '0');
  const minutes = String(now.getMinutes()).padStart(2, '0');
  const seconds = String(now.getSeconds()).padStart(2, '0');
  const timestamp = `${year}${month}${day}_${hours}${minutes}${seconds}`;
  const filename = `devices_export_${timestamp}.${extension}`;
  const success = await executeDownload(`/device/export/devices?format=${format}`, filename);
  if (success) {
    toast.success(t('device.messages.exportSuccess', { format: format.toUpperCase() }));
  } else {
    toast.error(handleError(exportError, 'device.messages.exportFailed', { format: format.toUpperCase() }));
  }
};

onMounted(async () => {
  if (!hasPermission(mainMenuName, 'Display')) return;
  await loadProtocols();
  await loadTable();
});
</script>