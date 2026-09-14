<template>
  <div class="flex flex-col items-center justify-center h-full w-full p-8 text-center bg-base-200">

    <template v-if="isChecking">
      <span class="loading loading-spinner loading-lg text-primary"></span>
      <p class="mt-4 text-base-content/70 font-medium">{{ $t('noAccess.verifying') }}</p>
    </template>

    <template v-else-if="isServerDown">
      <div class="text-error mb-4 animate-bounce">
        <Icon icon="lucide:server-crash" class="w-24 h-24 mx-auto drop-shadow-sm" />
      </div>
      <h2 class="text-3xl font-extrabold text-base-content mb-3 tracking-tight">{{ $t('noAccess.connectionLost') }}</h2>
      <p class="text-base-content/70 max-w-md mx-auto text-lg">
        {{ $t('noAccess.connectionLostDesc') }}
      </p>
      <button @click="reloadPage" class="btn btn-outline btn-error mt-6">
        <Icon icon="lucide:refresh-cw" class="w-4 h-4 mr-2" /> {{ $t('noAccess.tryAgain') }}
      </button>
    </template>

    <template v-else>
      <div class="text-warning mb-4">
        <Icon icon="lucide:shield-alert" class="w-24 h-24 mx-auto drop-shadow-sm" />
      </div>
      <h2 class="text-3xl font-extrabold text-base-content mb-3 tracking-tight">{{ $t('noAccess.accessDenied') }}</h2>
      <p class="text-base-content/70 max-w-md mx-auto text-lg">
        {{ $t('noAccess.accessDeniedDesc') }}
      </p>
      <router-link :to="{ name: 'dashboard' }"
        class="btn btn-primary mt-6 text-white shadow-sm hover:scale-105 transition-transform">
        {{ $t('noAccess.returnDashboard') }}
      </router-link>
    </template>

  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { Icon } from '@iconify/vue';

const { t } = useI18n();

const isChecking = ref(true);
const isServerDown = ref(false);

const checkConnection = async () => {
  isChecking.value = true;
  isServerDown.value = false;

  try {
    const response = await fetch(import.meta.env.VITE_API_BASE_URL + '/ping', { cache: 'no-store' });

    if (!response.ok && response.status >= 500) {
      isServerDown.value = true;
    } else {
      isServerDown.value = false;
    }
  } catch (error) {
    isServerDown.value = true;
  } finally {
    isChecking.value = false;
  }
};

const reloadPage = async () => {
  await checkConnection();
  if (!isServerDown.value) {
    location.reload()
  }
}

onMounted(() => {
  checkConnection();
});
</script>