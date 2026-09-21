<template>
  <div class="relative min-h-screen bg-base-200 flex items-center justify-center p-4">
    <!-- Top Right Language Control -->
    <div class="absolute top-4 right-4 z-20 flex gap-1 bg-base-100 p-1 rounded-lg shadow-sm border border-base-300">
      <button type="button" @click="changeLang('en')"
        :class="['px-2.5 py-1 text-xs sm:text-sm font-bold rounded-md transition-colors', locale === 'en' ? 'bg-primary text-primary-content' : 'text-base-content/60 hover:bg-base-200']">
        EN
      </button>
      <button type="button" @click="changeLang('th')"
        :class="['px-2.5 py-1 text-xs sm:text-sm font-bold rounded-md transition-colors', locale === 'th' ? 'bg-primary text-primary-content' : 'text-base-content/60 hover:bg-base-200']">
        TH
      </button>
    </div>

    <!-- Non-LINE Browser Blocker Screen -->
    <div v-if="!isInLineApp && !isLiffInitializing"
      class="card w-full max-w-md bg-base-100 shadow-xl border border-base-200">
      <div class="card-body items-center text-center py-10">
        <div class="p-3 bg-warning/10 text-warning rounded-2xl mb-2">
          <Icon icon="lucide:alert-circle" class="w-10 h-10" />
        </div>
        <h3 class="card-title text-lg font-bold">{{ $t('lineRegister.appOnlyTitle') }}</h3>
        <p class="text-sm text-base-content/70">
          {{ $t('lineRegister.appOnlyDesc') }}
        </p>
      </div>
    </div>

    <!-- Loading State -->
    <div v-else-if="isLiffInitializing" class="flex flex-col items-center gap-3">
      <span class="loading loading-spinner loading-lg text-primary"></span>
      <p class="text-sm font-medium text-base-content/70">{{ $t('lineRegister.connecting') }}</p>
    </div>

    <!-- Registration / Link Form -->
    <div v-else class="card w-full max-w-md bg-base-100 shadow-xl border border-base-200">
      <div class="card-body p-6">

        <!-- Header -->
        <div class="flex items-center gap-3 pb-4 border-b border-base-200">
          <div class="p-2.5 bg-primary/10 text-primary rounded-xl flex items-center justify-center">
            <Icon icon="lucide:link" class="w-6 h-6" />
          </div>
          <div>
            <h2 class="text-lg font-extrabold text-base-content tracking-tight">{{ $t('lineRegister.title') }}</h2>
            <p class="text-xs text-base-content/60">{{ $t('lineRegister.subtitle') }}</p>
          </div>
        </div>

        <!-- User Profile Preview (Fetched from LINE) -->
        <div v-if="lineProfile" class="flex items-center gap-3 p-3 bg-base-200/50 rounded-xl my-3">
          <img :src="lineProfile.pictureUrl || '/default-avatar.png'" alt="LINE Avatar"
            class="w-10 h-10 rounded-full border border-base-300 object-cover" />
          <div class="overflow-hidden">
            <p class="text-xs text-base-content/50 uppercase font-semibold">{{ $t('lineRegister.profileLabel') }}</p>
            <p class="text-sm font-bold text-base-content truncate">{{ lineProfile.displayName }}</p>
          </div>
        </div>

        <form @submit.prevent="handleSubmit" class="block mt-2">
          <!-- Error Alert Banner -->
          <div v-if="errorMessage"
            class="bg-error/10 text-error p-3 rounded-md mb-4 text-[0.9rem] text-center border border-error/30 font-medium">
            {{ errorMessage }}
          </div>

          <!-- Username / Email -->
          <div class="mb-4">
            <label class="block mb-2 text-base-content/90 font-medium text-[0.95rem]">
              {{ $t('lineRegister.username') }}
            </label>
            <input type="text" v-model="form.username" :placeholder="$t('lineRegister.usernamePlaceholder')" required
              :disabled="isLinking"
              class="w-full p-3 border border-base-300 rounded-md text-base-content text-base bg-base-200 box-border transition-colors duration-200 focus:outline-none focus:border-primary focus:ring-[3px] focus:ring-primary/20 disabled:opacity-60 disabled:cursor-not-allowed" />
          </div>

          <!-- Password -->
          <div class="mb-5">
            <label class="block mb-2 text-base-content/90 font-medium text-[0.95rem]">
              {{ $t('lineRegister.password') }}
            </label>
            <input type="password" v-model="form.password" :placeholder="$t('lineRegister.passwordPlaceholder')"
              required :disabled="isLinking"
              class="w-full p-3 border border-base-300 rounded-md text-base-content text-base bg-base-200 box-border transition-colors duration-200 focus:outline-none focus:border-primary focus:ring-[3px] focus:ring-primary/20 disabled:opacity-60 disabled:cursor-not-allowed" />
          </div>

          <button type="submit"
            class="w-full p-3 btn btn-primary border-none rounded-md text-base font-semibold cursor-pointer transition-colors mt-2 disabled:opacity-60 disabled:cursor-not-allowed"
            :disabled="isLinking">
            <span v-if="isLinking" class="loading loading-spinner loading-sm"></span>
            <span v-else>{{ $t('lineRegister.signInAndLink') }}</span>
          </button>
        </form>

      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import liff from '@line/liff';
import { toast } from 'vue3-toastify';
import { Icon } from '@iconify/vue';
import { useMutation } from '@/composables/useMutation';
import { useErrorHandler } from '@/composables/useErrorHandler';

const { t, locale } = useI18n();
const { handleError } = useErrorHandler();
const LIFF_ID = import.meta.env.VITE_LINE_LIFF_ID || '';

const isLiffInitializing = ref(true);
const isInLineApp = ref(true);
const lineProfile = ref(null);
const errorMessage = ref('');

const form = ref({
  username: '',
  password: ''
});

const changeLang = (lang) => {
  locale.value = lang;
  localStorage.setItem('lang', lang);
};

const {
  res: linkRes,
  isLoading: isLinking,
  error: linkError,
  execute: linkAccountApi
} = useMutation();

onMounted(async () => {
  const savedLang = localStorage.getItem('lang');
  if (savedLang) {
    locale.value = savedLang;
  }

  try {
    await liff.init({ liffId: LIFF_ID });

    if (!liff.isInClient()) {
      isInLineApp.value = false;
      isLiffInitializing.value = false;
      return;
    }

    if (liff.isLoggedIn()) {
      lineProfile.value = await liff.getProfile();
    }
  } catch (err) {
    toast.error(err.message || 'Failed to initialize LINE LIFF');
  } finally {
    isLiffInitializing.value = false;
  }
});

const handleSubmit = async () => {
  errorMessage.value = '';

  const idToken = liff.getIDToken();
  if (!idToken) {
    errorMessage.value = t('lineRegister.sessionExpired');
    return;
  }

  await linkAccountApi(
    '/user/linkline',
    {
      username: form.value.username,
      password: form.value.password,
      idToken: idToken
    },
    'POST'
  );

  if (!linkRes.value?.ok) {
    if (linkRes.value?.status >= 500 || !linkError.value?.message) {
      errorMessage.value = t('login.errorConnection');
    } else {
      if (linkError.value?.message === 't_invalid_user_password') {
        errorMessage.value = t('login.invalidLogin');
      } else {
        errorMessage.value = handleError(linkError, 'common.messages.loadError');
      }
    }
    return;
  }

  toast.success(t('lineRegister.linkSuccess'));

  setTimeout(() => {
    if (liff.isInClient()) {
      liff.closeWindow();
    }
  }, 1500);
};
</script>