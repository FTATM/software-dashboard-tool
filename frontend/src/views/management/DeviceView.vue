<template>
  <div v-if="hasPermission(mainMenuName, 'Display')" class="w-full h-full overflow-y-auto p-4 sm:p-6 space-y-5">

    <!-- Anchored Page Header Card with Background -->
    <div
      class="bg-base-100 border border-base-300 rounded-2xl p-5 shadow-xs flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
      <div class="flex items-start sm:items-center gap-3.5">
        <div class="p-3 bg-primary/10 text-primary rounded-xl flex items-center justify-center shrink-0">
          <Icon icon="lucide:cpu" class="w-6 h-6" />
        </div>
        <div>
          <!-- Breadcrumbs -->
          <div
            class="flex items-center gap-1.5 text-xs font-semibold text-base-content/50 uppercase tracking-wider mb-0.5">
            <span>{{ $t('menu.management') }}</span>
            <Icon icon="lucide:chevron-right" class="w-3.5 h-3.5" />
            <span class="text-primary">{{ $t('device.title') }}</span>
          </div>

          <!-- Title & Subtitle -->
          <div class="flex items-center gap-2.5">
            <h1 class="m-0 text-xl sm:text-2xl font-black text-base-content tracking-tight">
              {{ $t('device.title') }}
            </h1>
          </div>
          <p class="mt-0.5 mb-0 text-base-content/60 text-xs font-medium">
            {{ $t('device.subtitle') }}
          </p>
        </div>
      </div>

      <!-- Header Actions -->
      <div class="flex items-center gap-2 self-end sm:self-center">
        <button @click="loadTable"
          class="btn btn-sm btn-ghost border border-base-300 bg-base-100 hover:bg-base-200 rounded-xl gap-1.5 text-xs font-semibold shadow-xs transition-all"
          :title="$t('common.refresh')">
          <Icon icon="lucide:refresh-cw" class="w-3.5 h-3.5" :class="{ 'animate-spin': isLoading }" />
          <span>{{ $t('common.refresh') }}</span>
        </button>
      </div>
    </div>

    <!-- Reusable KPI Summary Status Cards Component -->
    <StatCardGroup :items="statCardsData" />

    <!-- Main Table Card -->
    <div class="bg-base-100 border border-base-300 rounded-2xl p-4 sm:p-5 shadow-xs">
      <TableData :data="deviceTable" :columns="tableColumns" :initial-sorting="[{ id: 'deviceId', desc: false }]"
        :is-loading="isLoading">

        <!-- Toolbar Actions -->
        <template #toolbar-actions>
          <!-- Export Dropdown -->
          <div class="dropdown dropdown-end">
            <div tabindex="0" role="button"
              class="btn btn-sm btn-outline btn-secondary rounded-xl shadow-xs hover:shadow-sm transition-all font-medium">
              <Icon icon="lucide:download" class="w-4 h-4 mr-1" />
              {{ $t('common.export') }}
            </div>
            <ul tabindex="0"
              class="dropdown-content z-50 menu p-1.5 shadow-xl bg-base-100 rounded-xl border border-base-300 w-36 mt-1 text-xs font-medium">
              <li><a @click="exportData('json')" class="rounded-lg py-1.5">JSON</a></li>
              <li><a @click="exportData('csv')" class="rounded-lg py-1.5">CSV</a></li>
              <li><a @click="exportData('excel')" class="rounded-lg py-1.5">Excel</a></li>
            </ul>
          </div>

          <!-- Import Button -->
          <button
            class="btn btn-sm btn-outline btn-accent rounded-xl shadow-xs hover:shadow-sm transition-all font-medium"
            @click="openImportModal">
            <Icon icon="lucide:upload" class="w-4 h-4 mr-1" />
            {{ $t('common.import') }}
          </button>

          <!-- Create Button -->
          <button
            class="btn btn-sm btn-primary rounded-xl font-semibold shadow-xs hover:shadow-md transition-all gap-1 text-white"
            @click="openCreateModal">
            <Icon icon="lucide:plus" class="w-4 h-4" />
            {{ $t('device.addDevice') }}
          </button>
        </template>

        <!-- ID Cell -->
        <template #cell-deviceId="{ value }">
          <span class="font-mono text-xs font-bold text-base-content/50">#{{ value }}</span>
        </template>

        <!-- Device Name with Scaling Badge -->
        <template #cell-deviceName="{ row }">
          <div class="flex items-center gap-2">
            <span class="font-semibold text-base-content tracking-tight text-sm">{{ row.deviceName }}</span>
            <span v-if="hasScaling(row)"
              class="px-1.5 py-0.5 rounded-md text-[10px] font-bold bg-accent/10 text-accent border border-accent/20">
              {{ $t('device.scaledBadge') }}
            </span>
          </div>
        </template>

        <!-- Dedicated Reference Device Column -->
        <template #cell-refDeviceId="{ row, value }">
          <div v-if="row.refDeviceId"
            class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-lg bg-info/10 text-info border border-info/20 text-xs font-medium">
            <Icon icon="lucide:link-2" class="w-3 h-3" />
            <span>{{ value }}</span>
          </div>
          <span v-else class="text-base-content/30 text-xs font-mono">-</span>
        </template>

        <!-- Protocol Badge -->
        <template #cell-protocol="{ value }">
          <span v-if="value && value !== 'none'"
            class="inline-flex items-center px-2 py-0.5 rounded-md bg-base-200 border border-base-300 text-xs font-semibold uppercase tracking-wider text-base-content/80 font-mono">
            {{ value }}
          </span>
          <span v-else class="text-base-content/40 text-xs">
            {{ $t('common.none') }}
          </span>
        </template>

        <!-- Live Status Pill -->
        <template #cell-status="{ row }">
          <div class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold"
            :class="row.status ? 'bg-success/10 text-success border border-success/20' : 'bg-base-200 text-base-content/50 border border-base-300'">
            <span class="w-1.5 h-1.5 rounded-full"
              :class="row.status ? 'bg-success animate-pulse' : 'bg-base-content/30'"></span>
            {{ row.status ? $t('common.active') : $t('common.inactive') }}
          </div>
        </template>

        <!-- Minimal Icon Actions -->
        <template #cell-actions="{ row }">
          <div class="flex justify-end items-center gap-1">
            <button @click="openEditModal(row)"
              class="btn btn-ghost btn-xs btn-square rounded-lg text-base-content/70 hover:text-primary hover:bg-primary/10 transition-colors"
              :title="$t('common.edit')">
              <Icon icon="lucide:pencil" class="w-4 h-4" />
            </button>
            <button @click="openDeleteModal(row)"
              class="btn btn-ghost btn-xs btn-square rounded-lg text-base-content/70 hover:text-error hover:bg-error/10 transition-colors"
              :title="$t('common.delete')">
              <Icon icon="lucide:trash-2" class="w-4 h-4" />
            </button>
          </div>
        </template>
      </TableData>
    </div>

    <!-- Create/Edit Modal with Scrollable Body -->
    <dialog ref="deviceModal" class="modal">
      <div
        class="modal-box sm:w-11/12 sm:max-w-xl p-0 overflow-hidden shadow-2xl rounded-2xl flex flex-col max-h-[85vh] border border-base-300 bg-base-100">
        <!-- Pinned Header -->
        <div class="px-6 py-4 border-b border-base-200 bg-base-100 flex justify-between items-center shrink-0">
          <div class="flex items-center gap-2.5">
            <div class="p-2 rounded-xl bg-primary/10 text-primary">
              <Icon :icon="isEditing ? 'lucide:pencil' : 'lucide:plus'" class="w-5 h-5" />
            </div>
            <div>
              <h3 class="m-0 text-lg font-bold text-base-content">
                {{ isEditing ? $t('device.editDevice') : $t('device.createDevice') }}
              </h3>
              <p class="m-0 text-xs text-base-content/50">
                {{ isEditing ? $t('device.editSubtitle') : $t('device.createSubtitle') }}
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

            <!-- Device Name -->
            <label class="form-control w-full">
              <div class="label pb-1 flex justify-between">
                <span class="label-text font-semibold text-xs uppercase tracking-wider text-base-content/70">{{
                  $t('common.deviceName') }}</span>
                <span class="label-text-alt text-base-content/50 font-mono text-[11px]">
                  {{ form.deviceName?.length || 0 }}/31
                </span>
              </div>
              <input type="text" v-model="form.deviceName" maxlength="31"
                :placeholder="$t('device.deviceNamePlaceholder')" @blur="v$.deviceName.$touch()"
                :class="['input input-sm h-10 input-bordered w-full rounded-xl', { 'input-error': v$.deviceName.$error }]" />
              <div class="label px-1 py-0.5 min-h-[20px]">
                <span v-if="v$.deviceName.$error" class="label-text-alt text-error font-medium">
                  {{ v$.deviceName.$errors[0].$message }}
                </span>
              </div>
            </label>

            <!-- Reference Device -->
            <label class="form-control w-full">
              <div class="label pb-1 flex justify-between items-center">
                <span class="label-text font-semibold text-xs uppercase tracking-wider text-base-content/70">{{
                  $t('device.refDevice') }}</span>
                <button v-if="form.refDeviceId" type="button"
                  class="btn btn-ghost btn-xs text-error gap-1 hover:bg-error/10 h-auto py-0.5 px-2"
                  @click="form.refDeviceId = null">
                  <Icon icon="lucide:x" class="w-3 h-3" />
                  {{ $t('common.cancel') }}
                </button>
              </div>
              <SearchableDropdown v-model="form.refDeviceId" :options="refDeviceOptions" label-key="name" value-key="id"
                :placeholder="$t('device.refDevicePlaceholder')" />
              <div v-if="form.refDeviceId" class="label px-1 py-1">
                <span class="label-text-alt text-info font-medium flex items-center gap-1">
                  <Icon icon="lucide:info" class="w-3.5 h-3.5" />
                  {{ $t('device.virtualDeviceProtocolHint') }}
                </span>
              </div>
            </label>

            <!-- Protocol -->
            <label class="form-control w-full">
              <div class="label pb-1">
                <span class="label-text font-semibold text-xs uppercase tracking-wider text-base-content/70">{{
                  $t('common.protocol') }}</span>
              </div>
              <select v-model="form.protocol" :disabled="!!form.refDeviceId"
                class="select select-sm h-10 select-bordered w-full rounded-xl">
                <option value="" disabled>{{ $t('common.protocolPlaceholder') }}</option>
                <option value="none">{{ $t('common.none') }}</option>
                <option v-for="proto in protocolList" :key="proto" :value="proto">
                  {{ proto }}
                </option>
              </select>
            </label>

            <!-- Active Status Box -->
            <div class="p-4 bg-base-200/50 rounded-2xl border border-base-300 mt-3">
              <div class="flex items-center justify-between">
                <div>
                  <span class="font-bold block text-sm text-base-content">{{ $t('device.activeStatus') }}</span>
                  <span class="text-xs text-base-content/60">{{ $t('device.activeStatusDesc') }}</span>
                </div>
                <input type="checkbox" v-model="form.active" class="toggle toggle-primary toggle-sm" />
              </div>
            </div>

            <!-- Scaling Toggle & Fields -->
            <div class="p-4 bg-base-200/50 rounded-2xl border border-base-300">
              <div class="flex items-center justify-between">
                <div>
                  <span class="font-bold block text-sm text-base-content">{{ $t('device.enableScaling') }}</span>
                  <span class="text-xs text-base-content/60">{{ $t('device.enableScalingDesc') }}</span>
                </div>
                <input type="checkbox" v-model="form.enableScaling" class="toggle toggle-primary toggle-sm" />
              </div>

              <div v-if="form.enableScaling" class="mt-4 pt-4 border-t border-base-300/80 space-y-4">
                <!-- Raw Range -->
                <div>
                  <div class="text-[11px] font-bold uppercase tracking-wider text-base-content/70 mb-2">
                    {{ $t('device.rawRange') }}
                  </div>
                  <div class="grid grid-cols-2 gap-3">
                    <label class="form-control w-full">
                      <span class="label-text text-xs font-semibold mb-1">{{ $t('device.rawMin') }}</span>
                      <input type="number" step="any" v-model="form.rawMin"
                        :placeholder="$t('device.rawMinPlaceholder')" @blur="v$.rawMin?.$touch()"
                        :class="['input input-sm input-bordered w-full rounded-xl', { 'input-error': v$.rawMin?.$error }]" />
                    </label>

                    <label class="form-control w-full">
                      <span class="label-text text-xs font-semibold mb-1">{{ $t('device.rawMax') }}</span>
                      <input type="number" step="any" v-model="form.rawMax"
                        :placeholder="$t('device.rawMaxPlaceholder')" @blur="v$.rawMax?.$touch()"
                        :class="['input input-sm input-bordered w-full rounded-xl', { 'input-error': v$.rawMax?.$error }]" />
                    </label>
                  </div>
                </div>

                <!-- EU Range -->
                <div>
                  <div class="text-[11px] font-bold uppercase tracking-wider text-base-content/70 mb-2">
                    {{ $t('device.euRange') }}
                  </div>
                  <div class="grid grid-cols-2 gap-3">
                    <label class="form-control w-full">
                      <span class="label-text text-xs font-semibold mb-1">{{ $t('device.euMin') }}</span>
                      <input type="number" step="any" v-model="form.euMin" :placeholder="$t('device.euMinPlaceholder')"
                        @blur="v$.euMin?.$touch()"
                        :class="['input input-sm input-bordered w-full rounded-xl', { 'input-error': v$.euMin?.$error }]" />
                    </label>

                    <label class="form-control w-full">
                      <span class="label-text text-xs font-semibold mb-1">{{ $t('device.euMax') }}</span>
                      <input type="number" step="any" v-model="form.euMax" :placeholder="$t('device.euMaxPlaceholder')"
                        @blur="v$.euMax?.$touch()"
                        :class="['input input-sm input-bordered w-full rounded-xl', { 'input-error': v$.euMax?.$error }]" />
                    </label>
                  </div>
                </div>

                <div
                  class="text-[11px] text-base-content/60 italic bg-base-100 p-2.5 rounded-xl border border-base-300/60">
                  {{ $t('device.scalingFormulaHint') }}
                </div>
              </div>
            </div>

            <!-- Network Status (Editing Mode) -->
            <div v-if="isEditing" class="p-4 bg-base-200/50 rounded-2xl border border-base-300 space-y-3">
              <div class="flex items-center justify-between">
                <h4 class="font-bold text-xs uppercase tracking-wider text-base-content/70 m-0">
                  {{ $t('device.networkStatus') }}
                </h4>

                <div v-if="form.refDeviceId" class="badge badge-info badge-sm font-semibold">
                  {{ $t('device.virtualDeviceNoPing') }}
                </div>
                <div v-else-if="!form.protocol || form.protocol === 'none'"
                  class="badge badge-warning badge-sm font-semibold">
                  {{ $t('device.cannotPingNoProtocol') }}
                </div>
                <button v-else type="button" @click="testConnection()" class="btn btn-xs btn-outline rounded-lg"
                  :disabled="isPinging">
                  <span v-if="isPinging" class="loading loading-spinner loading-xs"></span>
                  <Icon v-else icon="lucide:wifi" class="w-3 h-3 mr-1" />
                  {{ $t('device.testConnection') }}
                </button>
              </div>

              <div class="flex items-center justify-between pt-1">
                <span class="text-xs font-semibold text-base-content/80">{{ $t('device.deviceStatus') }}</span>
                <div class="badge gap-1 border-none font-bold text-xs"
                  :class="form.isConnected ? 'bg-success text-white' : 'bg-error text-white'">
                  {{ form.isConnected ? $t('common.connect') : $t('common.disconnect') }}
                </div>
              </div>

              <div v-if="!form.isConnected" class="flex items-center justify-between">
                <span class="text-xs font-semibold text-base-content/80">{{ $t('device.lastSeenAt') }}</span>
                <span class="text-xs font-mono text-base-content/60">
                  {{ formatTime(form.lastSeenAt) }}
                </span>
              </div>
            </div>

          </div>

          <!-- Pinned Footer -->
          <div class="border-t border-base-200 p-4 px-6 flex justify-end gap-2 shrink-0 bg-base-100">
            <button type="button" class="btn btn-sm btn-ghost rounded-xl" @click="closeModal">{{ $t('common.cancel') }}</button>
            <button type="submit" class="btn btn-sm btn-primary rounded-xl px-6 text-white font-semibold">
              {{ isEditing ? $t('common.save') : $t('device.createDevice') }}
            </button>
          </div>
        </form>
      </div>
      <form method="dialog" class="modal-backdrop"><button @click="closeModal">close</button></form>
    </dialog>

    <!-- Delete Modal -->
    <dialog ref="deleteModal" class="modal z-[200]">
      <div class="modal-box rounded-2xl border border-base-300 p-6">
        <h3 class="font-bold text-lg text-error flex items-center gap-2">
          <Icon icon="lucide:alert-triangle" class="w-5 h-5" /> {{ $t('common.confirmDelete') }}
        </h3>
        <p class="py-3 text-sm text-base-content/80">
          {{ $t('device.deleteWarning', { name: deviceToDelete?.deviceName }) }}
        </p>
        <div class="modal-action mt-4">
          <button type="button" @click="closeDeleteModal" class="btn btn-sm btn-ghost rounded-xl"
            :disabled="isDeleting">
            {{ $t('common.noCancel') }}
          </button>
          <button type="button" @click="confirmDelete" class="btn btn-sm btn-error text-white rounded-xl"
            :disabled="isDeleting">
            <span v-if="isDeleting" class="loading loading-spinner loading-xs"></span> {{ $t('common.yesDelete') }}
          </button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop"><button @click="closeDeleteModal">close</button></form>
    </dialog>

    <!-- Import Modal -->
    <dialog ref="importModal" class="modal">
      <div
        class="modal-box w-11/12 max-w-4xl p-0 overflow-hidden shadow-2xl rounded-2xl flex flex-col max-h-[85vh] border border-base-300 bg-base-100">
        <div class="px-6 py-4 border-b border-base-200 bg-base-100 flex justify-between items-center shrink-0">
          <h3 class="m-0 text-lg font-bold text-base-content">{{ $t('device.import.title') }}</h3>
          <button class="btn btn-sm btn-circle btn-ghost" @click="closeImportModal">
            <Icon icon="lucide:x" class="w-4 h-4" />
          </button>
        </div>

        <div class="px-6 py-3.5 bg-info/10 border-b border-info/20 text-xs">
          <div class="flex gap-2.5 items-start">
            <Icon icon="lucide:info" class="w-4 h-4 text-info shrink-0 mt-0.5" />
            <div class="space-y-1">
              <p class="font-bold text-xs m-0">{{ $t('device.import.requirementsTitle') }}</p>
              <p class="m-0 text-base-content/70">
                {{ $t('device.import.supportedFormats') }} | Columns:
                <code class="bg-base-100 px-1 py-0.5 rounded font-mono text-info">deviceName</code>,
                <code class="bg-base-100 px-1 py-0.5 rounded font-mono text-info">protocol</code>,
                <code class="bg-base-100 px-1 py-0.5 rounded font-mono text-info">status</code>
              </p>
            </div>
          </div>
        </div>

        <div class="p-6 bg-base-100 flex-1 overflow-y-auto space-y-4">
          <div class="flex flex-wrap items-center gap-3">
            <input type="file" @change="handleFileSelect" accept=".csv, .json, .xlsx"
              class="file-input file-input-bordered file-input-sm file-input-primary w-full max-w-xs rounded-xl" />
            <button @click="validateImportFile" :disabled="!selectedFile || isLoadingValidateImport"
              class="btn btn-sm btn-primary rounded-xl text-white">
              <span v-if="isLoadingValidateImport" class="loading loading-spinner loading-xs"></span>
              {{ $t('device.import.validateFile') }}
            </button>
            <span v-if="validateImportError" class="font-semibold text-error text-xs">
              {{ validateImportError?.message || "Error" }}
            </span>
          </div>

          <div v-if="validationResults.length > 0" class="space-y-2">
            <div class="flex justify-between items-center px-1 text-xs">
              <span class="font-medium text-base-content/70">Total rows: <b>{{ validationResults.length }}</b></span>
              <span v-if="invalidCount > 0" class="badge badge-error badge-sm text-white font-bold">{{ invalidCount }}
                Invalid row(s)</span>
              <span v-else class="badge badge-success badge-sm text-white font-bold">All rows valid</span>
            </div>

            <div class="overflow-x-auto border border-base-200 rounded-xl">
              <table class="table table-xs table-zebra w-full">
                <thead class="bg-base-200/60">
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
                    <td class="font-semibold">{{ row.deviceName }}</td>
                    <td>{{ row.protocol || '-' }}</td>
                    <td>
                      <span class="badge badge-xs font-semibold uppercase"
                        :class="row.active === false ? 'badge-ghost text-base-content/50' : 'badge-success'">
                        {{ row.active === false ? $t('common.inactive') : $t('common.active') }}
                      </span>
                    </td>
                    <td>
                      <span class="badge badge-xs" :class="row.isValid ? 'badge-success' : 'badge-error'">
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

            <div class="flex justify-between items-center px-1 pt-1 text-xs text-base-content/60">
              <span>Page {{ currentPage }} of {{ totalPages }}</span>
              <div class="join">
                <button class="join-item btn btn-xs" :disabled="currentPage === 1" @click="currentPage--">‹</button>
                <button class="join-item btn btn-xs" :disabled="currentPage >= totalPages"
                  @click="currentPage++">›</button>
              </div>
            </div>
          </div>
        </div>

        <div class="border-t border-base-200 p-4 px-6 flex justify-end gap-2 shrink-0">
          <button type="button" class="btn btn-sm btn-ghost rounded-xl" @click="closeImportModal">{{ $t('common.cancel') }}</button>
          <button @click="confirmImport" :disabled="!canConfirmImport || isImporting"
            class="btn btn-sm btn-success text-white px-6 rounded-xl font-semibold">
            <span v-if="isImporting" class="loading loading-spinner loading-xs"></span>
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
import StatCardGroup from '@/components/StatCardGroup.vue';
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

// Dynamic bilingual configuration for StatCardGroup
const statCardsData = computed(() => {
  const total = deviceTable.value.length;
  const active = deviceTable.value.filter(d => d.active).length;
  const virtual = deviceTable.value.filter(d => d.refDeviceId).length;
  const physical = total - virtual;

  return [
    {
      label: t('device.stats.total'),
      value: total,
      icon: 'lucide:cpu',
      color: 'primary'
    },
    {
      label: t('device.stats.active'),
      value: active,
      icon: 'lucide:activity',
      color: 'success',
      valueClass: 'text-success'
    },
    {
      label: t('device.stats.physical'),
      value: physical,
      icon: 'lucide:hard-drive',
      color: 'accent'
    },
    {
      label: t('device.stats.virtual'),
      value: virtual,
      icon: 'lucide:link-2',
      color: 'info',
      valueClass: 'text-info'
    }
  ];
});

const tableColumns = computed(() => [
  { header: t('common.id'), accessorKey: 'deviceId', meta: { headerClass: 'w-16', cellClass: 'font-bold' } },
  { header: t('common.deviceName'), accessorKey: 'deviceName' },
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

const getRefDeviceName = (refId) => {
  if (!refId) return '';
  const target = deviceTable.value.find(d => d.deviceId === refId);
  return target ? `#${target.deviceId} - ${target.deviceName}` : `#${refId}`;
};

const refDeviceOptions = computed(() => {
  return deviceTable.value
    .filter(d => {
      if (isEditing.value && d.deviceId === editingDeviceId.value) return false;
      return !d.refDeviceId;
    })
    .map(d => ({
      id: d.deviceId,
      name: `#${d.deviceId} - ${d.deviceName}`
    }));
});

watch(() => form.value.refDeviceId, (newRef) => {
  if (newRef) {
    form.value.protocol = 'none';
  }
});

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