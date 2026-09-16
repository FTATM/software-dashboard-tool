<template>
  <!-- ⚡ Swapped bg-slate-100 to bg-base-200 -->
  <div class="relative flex justify-center items-center min-h-screen bg-base-200 transition-colors duration-300">

    <!-- Top Left Controls (Language + Theme) -->
    <div class="absolute top-6 left-6 flex gap-3 items-center">
      <!-- Language Toggle -->
      <div class="flex gap-1 bg-base-100 p-1 rounded-lg shadow-sm border border-base-300">
        <button @click="changeLang('en')"
          :class="['px-3 py-1.5 text-sm font-bold rounded-md transition-colors', locale === 'en' ? 'bg-primary text-primary-content' : 'text-base-content/60 hover:bg-base-200']">
          EN
        </button>
        <button @click="changeLang('th')"
          :class="['px-3 py-1.5 text-sm font-bold rounded-md transition-colors', locale === 'th' ? 'bg-primary text-primary-content' : 'text-base-content/60 hover:bg-base-200']">
          TH
        </button>
      </div>

      <!-- ⚡ NEW: Theme Toggle Button using your store -->
      <button @click="themeStore.toggleTheme()"
        class="btn btn-square btn-sm btn-ghost bg-base-100 border border-base-300 shadow-sm text-base-content">
        <span v-if="themeStore.isDarkTheme">🌙</span>
        <span v-else>☀️</span>
      </button>
    </div>

    <!-- ⚡ Swapped bg-white to bg-base-100 -->
    <div
      class="bg-base-100 w-full max-w-[400px] p-10 rounded-xl shadow-[0_4px_20px_rgba(0,0,0,0.08)] border border-base-300">

      <div class="text-center mb-[30px]">
        <!-- ⚡ Swapped text colors to text-base-content -->
        <h2 class="m-0 text-base-content text-[1.75rem] font-bold">{{ $t('login.title') }}</h2>
        <p class="text-base-content/70 mt-2 text-[0.95rem]">{{ $t('login.subtitle') }}</p>
      </div>

      <form @submit.prevent="handleLogin" class="block">
        <div v-if="messageLogin"
          class="bg-error/10 text-error p-3 rounded-md mb-5 text-[0.9rem] text-center border border-error/30 font-medium">
          {{ messageLogin }}
        </div>

        <div class="mb-5">
          <label for="username" class="block mb-2 text-base-content/90 font-medium text-[0.95rem]">
            {{ $t('login.username') }}
          </label>
          <!-- ⚡ Swapped border and background to base-content and base-200 -->
          <input type="text" id="username" v-model="username" :placeholder="$t('login.usernamePlaceholder')" required
            :disabled="isUserLoginLoading"
            class="w-full p-3 border border-base-300 rounded-md text-base-content text-base bg-base-200 box-border transition-colors duration-200 focus:outline-none focus:border-primary focus:ring-[3px] focus:ring-primary/20 disabled:opacity-60 disabled:cursor-not-allowed" />
        </div>

        <div class="mb-5">
          <label for="password" class="block mb-2 text-base-content/90 font-medium text-[0.95rem]">
            {{ $t('login.password') }}
          </label>
          <input type="password" id="password" v-model="password" :placeholder="$t('login.passwordPlaceholder')"
            required :disabled="isUserLoginLoading"
            class="w-full p-3 border border-base-300 rounded-md text-base-content text-base bg-base-200 box-border transition-colors duration-200 focus:outline-none focus:border-primary focus:ring-[3px] focus:ring-primary/20 disabled:opacity-60 disabled:cursor-not-allowed" />
        </div>

        <button type="submit" :disabled="isUserLoginLoading"
          class="w-full p-3 btn btn-primary border-none rounded-md text-base font-semibold cursor-pointer transition-colors mt-2.5 disabled:opacity-60 disabled:cursor-not-allowed">
          {{ isUserLoginLoading ? $t('login.authenticating') : $t('login.signIn') }}
        </button>
      </form>

      <div class="text-center mt-6 text-[0.8rem] text-base-content/50">
        {{ $t('login.footerNote') }}
      </div>

    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useMutation } from '@/composables/useMutation';
import { useUserStore } from '@/stores/useUserStore';
import { useI18n } from 'vue-i18n';
import { useThemeStore } from '@/stores/useThemeStore'; // ⚡ NEW: Import your theme store
import { useErrorHandler } from '@/composables/useErrorHandler';
const { handleError } = useErrorHandler();

const router = useRouter();

// i18n Setup
const { t, locale } = useI18n();

const changeLang = (lang) => {
  locale.value = lang;
  localStorage.setItem('lang', lang);
};

// --- STORES ---
const userStore = useUserStore();
const { setUser } = userStore;

// ⚡ NEW: Initialize Theme Store
const themeStore = useThemeStore();

const messageLogin = ref("");
const username = ref('');
const password = ref('');

const {
  data: userLogin,
  res: userLoginRes,
  isLoading: isUserLoginLoading,
  error: userLoginError,
  execute: userLoginApi
} = useMutation();

const handleLogin = async () => {
  messageLogin.value = ""
  await userLoginApi('/user/login', { username: username.value, password: password.value }, "POST")

  if (!userLoginRes.value.ok) {
    if (userLoginRes.value.status >= 500 || !userLoginError.value?.message) {
      messageLogin.value = t('login.errorConnection')
    } else {
      if (userLoginError.value?.message === "t_invalid_user_password") {
        messageLogin.value = t('login.invalidLogin');
      } else {
        messageLogin.value = handleError(userLoginError, 'common.messages.loadError')
      }
    }
    return
  }

  if (userLogin.value.data.token) {
    localStorage.setItem('token', userLogin.value.data.token);
  }

  setUser(
    {
      id: userLogin.value.data.userId,
      firstName: userLogin.value.data.firstName,
      lastName: userLogin.value.data.lastName,
      fullName: `${userLogin.value.data.firstName} ${userLogin.value.data.lastName}`
    }
  )
  router.replace('/dashboard');
};

onMounted(() => {
  const savedLang = localStorage.getItem('lang');
  if (savedLang) {
    locale.value = savedLang;
  }
});
</script>