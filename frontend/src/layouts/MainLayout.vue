<template>
  <div class="drawer h-screen w-screen overflow-hidden bg-base-200/50 font-sans">
    <input id="mobile-nav-drawer" type="checkbox" class="drawer-toggle" v-model="isMobileNavOpen" />

    <div class="drawer-content flex flex-col h-screen overflow-hidden">

      <!-- Top Navigation Ribbon -->
      <header
        class="navbar min-h-[58px] h-[58px] shrink-0 bg-base-100 border-b border-base-300 px-3 sm:px-5 flex items-center justify-between z-30 select-none shadow-xs">

        <!-- Left Section: Mobile Toggle & Brand -->
        <div class="flex items-center gap-2">
          <!-- Mobile Menu Trigger (< lg) -->
          <label for="mobile-nav-drawer" class="btn btn-ghost btn-square btn-xs sm:btn-sm text-base-content lg:hidden">
            <Icon icon="lucide:menu" class="w-5 h-5" />
          </label>

          <!-- Brand Logo & Name -->
          <router-link to="/"
            class="flex items-center gap-2.5 px-2 py-1.5 rounded-lg hover:bg-base-200 transition-colors">
            <div
              class="p-1.5 rounded-md bg-primary/10 border border-primary/20 text-primary flex items-center justify-center">
              <img src="/favicon.svg" alt="Logo" class="w-4 h-4 object-contain" />
            </div>
            <span class="font-bold text-sm tracking-tight text-base-content">
              {{ $t('header.brand') }}
            </span>
          </router-link>
        </div>

        <!-- Center Section: Desktop Horizontal Navigation Menu (>= lg) -->
        <nav class="hidden lg:flex items-center">
          <ul class="menu menu-horizontal menu-sm p-0 gap-1.5 text-base-content font-medium text-sm">

            <!-- Dashboard -->
            <li>
              <router-link :to="{ name: 'dashboard' }" @click="closeAllDropdowns"
                active-class="!bg-primary/10 !text-primary font-bold"
                class="py-2 px-3 rounded-lg flex items-center gap-2 transition-colors">
                <Icon icon="lucide:layout-dashboard" class="w-[18px] h-[18px] opacity-80" />
                <span>{{ $t('menu.dashboard') }}</span>
              </router-link>
            </li>

            <!-- Canvas Group -->
            <li
              v-if="hasPermission('Canvas Design', 'Display') || hasPermission('Canvas Access', 'Display') || hasPermission('Canvas', 'Display')">
              <details class="dropdown" name="header-dropdown">
                <summary
                  class="py-2 px-3 rounded-lg flex items-center gap-1.5 transition-colors hover:bg-base-200 cursor-pointer">
                  <Icon icon="lucide:layout-list" class="w-[18px] h-[18px] opacity-80" />
                  <span>{{ $t('menu.canvas') }}</span>
                </summary>
                <ul
                  class="dropdown-content z-50 menu menu-sm p-1.5 shadow-lg bg-base-100 rounded-xl border border-base-300 w-48 mt-2">
                  <li v-if="hasPermission('Canvas', 'Display')">
                    <router-link :to="{ name: 'canvas' }" @click="closeAllDropdowns"
                      active-class="!bg-primary/10 !text-primary font-bold" class="py-2 px-3 rounded-lg">
                      {{ $t('menu.canvas') }}
                    </router-link>
                  </li>
                  <li v-if="hasPermission('Canvas Design', 'Display')">
                    <router-link :to="{ name: 'canvasDesign' }" @click="closeAllDropdowns"
                      active-class="!bg-primary/10 !text-primary font-bold" class="py-2 px-3 rounded-lg">
                      {{ $t('menu.canvasDesign') }}
                    </router-link>
                  </li>
                  <li v-if="hasPermission('Canvas Access', 'Display')">
                    <router-link :to="{ name: 'canvasAccess' }" @click="closeAllDropdowns"
                      active-class="!bg-primary/10 !text-primary font-bold" class="py-2 px-3 rounded-lg">
                      {{ $t('menu.canvasAccess') }}
                    </router-link>
                  </li>
                </ul>
              </details>
            </li>

            <!-- Scheduler -->
            <li v-if="hasPermission('Scheduler', 'Display')">
              <router-link :to="{ name: 'scheduler' }" @click="closeAllDropdowns"
                active-class="!bg-primary/10 !text-primary font-bold"
                class="py-2 px-3 rounded-lg flex items-center gap-2 transition-colors">
                <Icon icon="lucide:calendar-clock" class="w-[18px] h-[18px] opacity-80" />
                <span>{{ $t('menu.scheduler') }}</span>
              </router-link>
            </li>

            <!-- Notification Group -->
            <li v-if="hasPermission('Notification User', 'Display') || hasPermission('Notification Device', 'Display')">
              <details class="dropdown" name="header-dropdown">
                <summary
                  class="py-2 px-3 rounded-lg flex items-center gap-1.5 transition-colors hover:bg-base-200 cursor-pointer">
                  <Icon icon="lucide:bell" class="w-[18px] h-[18px] opacity-80" />
                  <span>{{ $t('menu.notification') }}</span>
                </summary>
                <ul
                  class="dropdown-content z-50 menu menu-sm p-1.5 shadow-lg bg-base-100 rounded-xl border border-base-300 w-48 mt-2">
                  <li v-if="hasPermission('Notification User', 'Display')">
                    <router-link :to="{ name: 'notifUser' }" @click="closeAllDropdowns"
                      active-class="!bg-primary/10 !text-primary font-bold" class="py-2 px-3 rounded-lg">
                      {{ $t('menu.notifUser') }}
                    </router-link>
                  </li>
                  <li v-if="hasPermission('Notification Device', 'Display')">
                    <router-link :to="{ name: 'notifDevicerule' }" @click="closeAllDropdowns"
                      active-class="!bg-primary/10 !text-primary font-bold" class="py-2 px-3 rounded-lg">
                      {{ $t('menu.notifDeviceRule') }}
                    </router-link>
                  </li>
                </ul>
              </details>
            </li>

            <!-- Log Report -->
            <li v-if="hasPermission('Log Report', 'Display')">
              <router-link :to="{ name: 'logReport' }" @click="closeAllDropdowns"
                active-class="!bg-primary/10 !text-primary font-bold"
                class="py-2 px-3 rounded-lg flex items-center gap-2 transition-colors">
                <Icon icon="lucide:file-text" class="w-[18px] h-[18px] opacity-80" />
                <span>{{ $t('menu.logReport') }}</span>
              </router-link>
            </li>

            <!-- Management Group -->
            <li
              v-if="hasPermission('User', 'Display') || hasPermission('Role', 'Display') || hasPermission('Device', 'Display') || hasPermission('Device Group', 'Display')">
              <details class="dropdown" name="header-dropdown">
                <summary
                  class="py-2 px-3 rounded-lg flex items-center gap-1.5 transition-colors hover:bg-base-200 cursor-pointer">
                  <Icon icon="lucide:settings" class="w-[18px] h-[18px] opacity-80" />
                  <span>{{ $t('menu.management') }}</span>
                </summary>
                <ul
                  class="dropdown-content z-50 menu menu-sm p-1.5 shadow-lg bg-base-100 rounded-xl border border-base-300 w-48 mt-2">
                  <li v-if="hasPermission('User', 'Display')">
                    <router-link :to="{ name: 'user' }" @click="closeAllDropdowns"
                      active-class="!bg-primary/10 !text-primary font-bold" class="py-2 px-3 rounded-lg">
                      {{ $t('menu.user') }}
                    </router-link>
                  </li>
                  <li v-if="hasPermission('Role', 'Display')">
                    <router-link :to="{ name: 'role' }" @click="closeAllDropdowns"
                      active-class="!bg-primary/10 !text-primary font-bold" class="py-2 px-3 rounded-lg">
                      {{ $t('menu.role') }}
                    </router-link>
                  </li>
                  <li v-if="hasPermission('Device', 'Display')">
                    <router-link :to="{ name: 'device' }" @click="closeAllDropdowns"
                      active-class="!bg-primary/10 !text-primary font-bold" class="py-2 px-3 rounded-lg">
                      {{ $t('menu.device') }}
                    </router-link>
                  </li>
                  <li v-if="hasPermission('Device Group', 'Display')">
                    <router-link :to="{ name: 'deviceGroup' }" @click="closeAllDropdowns"
                      active-class="!bg-primary/10 !text-primary font-bold" class="py-2 px-3 rounded-lg">
                      {{ $t('menu.deviceGroup') }}
                    </router-link>
                  </li>
                </ul>
              </details>
            </li>

          </ul>
        </nav>

        <!-- Right Section: Utilities & User -->
        <div class="flex items-center gap-2.5">

          <!-- Company Badge -->
          <div
            class="hidden md:flex items-center gap-1.5 px-2 py-1 rounded-lg bg-base-200/50 border border-base-300 opacity-70 hover:opacity-100 transition-opacity">
            <span class="text-[10px] text-base-content/60 font-medium">{{ $t('header.poweredBy') }}</span>
            <img src="/LOGO_FT.png" alt="Company Logo" class="h-3.5 w-auto object-contain" />
          </div>

          <div class="h-4 w-px bg-base-300 mx-0.5 hidden md:block"></div>

          <!-- Language Switcher -->
          <div class="join bg-base-200 p-0.5 rounded-md border border-base-300 hidden sm:inline-flex">
            <button type="button"
              class="btn h-6 min-h-0 px-2.5 join-item text-[11px] font-bold rounded-sm border-none transition-colors"
              :class="locale === 'en' ? 'bg-base-100 text-base-content shadow-xs' : 'bg-transparent text-base-content/50 hover:text-base-content'"
              @click="setLanguage('en')">
              EN
            </button>
            <button type="button"
              class="btn h-6 min-h-0 px-2.5 join-item text-[11px] font-bold rounded-sm border-none transition-colors"
              :class="locale === 'th' ? 'bg-base-100 text-base-content shadow-xs' : 'bg-transparent text-base-content/50 hover:text-base-content'"
              @click="setLanguage('th')">
              TH
            </button>
          </div>

          <!-- Theme Switcher -->
          <label
            class="btn btn-ghost btn-sm btn-circle bg-base-200 border border-base-300 flex items-center justify-center cursor-pointer hover:bg-base-300/70 transition-colors">
            <input type="checkbox" v-model="themeStore.isDarkTheme" class="hidden" />
            <Icon :icon="themeStore.isDarkTheme ? 'lucide:moon' : 'lucide:sun'" class="w-4 h-4"
              :class="themeStore.isDarkTheme ? 'text-info' : 'text-warning'" />
          </label>

          <div class="h-4 w-px bg-base-300 mx-0.5 hidden sm:block"></div>

          <!-- User Profile Pill with Hover Tooltip -->
          <div 
            class="tooltip tooltip-bottom before:text-xs before:max-w-xs before:font-medium before:shadow-md" 
            :data-tip="userStore.user?.fullName || $t('header.defaultUser')">
            <div class="flex items-center gap-2 px-2.5 py-1.5 rounded-full bg-base-200 border border-base-300 text-xs cursor-pointer hover:bg-base-300/60 transition-colors">
              <div class="w-5 h-5 rounded-full bg-primary/20 text-primary flex items-center justify-center shrink-0">
                <Icon icon="lucide:user" class="w-3.5 h-3.5" />
              </div>
              <span class="font-medium text-base-content max-w-[100px] sm:max-w-[160px] md:max-w-[220px] truncate text-xs">
                {{ userStore.user?.fullName || $t('header.defaultUser') }}
              </span>
            </div>
          </div>

          <!-- Logout Button -->
          <button @click="handleLogout"
            class="btn btn-ghost btn-sm text-error/80 hover:text-error hover:bg-error/10 p-1.5 font-medium transition-all"
            :title="$t('header.logout')">
            <Icon icon="lucide:log-out" class="w-4 h-4" />
          </button>
        </div>

      </header>

      <!-- Main Slot Content (Full-Width Viewport) -->
      <main class="flex-1 min-h-0 h-full overflow-hidden w-full relative">
        <slot />
      </main>

    </div>

    <!-- Mobile Navigation Drawer Panel (< lg) -->
    <aside class="drawer-side z-50">
      <label for="mobile-nav-drawer" aria-label="close sidebar" class="drawer-overlay"></label>
      <div class="flex flex-col w-[260px] min-h-full bg-base-100 border-r border-base-300 p-3 select-none">

        <!-- Mobile Drawer Header -->
        <div class="flex items-center gap-2 pb-3 mb-2 border-b border-base-300">
          <img src="/favicon.svg" alt="Logo" class="w-5 h-5 object-contain" />
          <span class="font-bold text-sm text-base-content">{{ $t('header.brand') }}</span>
        </div>

        <!-- Navigation Links -->
        <div class="flex-1 overflow-y-auto">
          <ul class="menu menu-sm p-0 gap-1 text-base-content">
            <li>
              <router-link :to="{ name: 'dashboard' }" active-class="!bg-primary/10 !text-primary font-bold"
                @click="isMobileNavOpen = false" class="py-2">
                <Icon icon="lucide:layout-dashboard" class="w-4 h-4" /> {{ $t('menu.dashboard') }}
              </router-link>
            </li>

            <li
              v-if="hasPermission('Canvas Design', 'Display') || hasPermission('Canvas Access', 'Display') || hasPermission('Canvas', 'Display')">
              <details name="mobile-header-dropdown">
                <summary class="py-2">
                  <Icon icon="lucide:layout-list" class="w-4 h-4" /> {{ $t('menu.canvas') }}
                </summary>
                <ul class="ml-2 pl-2 border-l border-base-300">
                  <li v-if="hasPermission('Canvas', 'Display')"><router-link :to="{ name: 'canvas' }"
                      @click="isMobileNavOpen = false">{{ $t('menu.canvas') }}</router-link></li>
                  <li v-if="hasPermission('Canvas Design', 'Display')"><router-link :to="{ name: 'canvasDesign' }"
                      @click="isMobileNavOpen = false">{{ $t('menu.canvasDesign') }}</router-link></li>
                  <li v-if="hasPermission('Canvas Access', 'Display')"><router-link :to="{ name: 'canvasAccess' }"
                      @click="isMobileNavOpen = false">{{ $t('menu.canvasAccess') }}</router-link></li>
                </ul>
              </details>
            </li>

            <li v-if="hasPermission('Scheduler', 'Display')">
              <router-link :to="{ name: 'scheduler' }" active-class="!bg-primary/10 !text-primary font-bold"
                @click="isMobileNavOpen = false" class="py-2">
                <Icon icon="lucide:calendar-clock" class="w-4 h-4" /> {{ $t('menu.scheduler') }}
              </router-link>
            </li>

            <li v-if="hasPermission('Notification User', 'Display') || hasPermission('Notification Device', 'Display')">
              <details name="mobile-header-dropdown">
                <summary class="py-2">
                  <Icon icon="lucide:bell" class="w-4 h-4" /> {{ $t('menu.notification') }}
                </summary>
                <ul class="ml-2 pl-2 border-l border-base-300">
                  <li v-if="hasPermission('Notification User', 'Display')"><router-link :to="{ name: 'notifUser' }"
                      @click="isMobileNavOpen = false">{{ $t('menu.notifUser') }}</router-link></li>
                  <li v-if="hasPermission('Notification Device', 'Display')"><router-link
                      :to="{ name: 'notifDevicerule' }" @click="isMobileNavOpen = false">{{ $t('menu.notifDeviceRule')
                      }}</router-link></li>
                </ul>
              </details>
            </li>

            <li v-if="hasPermission('Log Report', 'Display')">
              <router-link :to="{ name: 'logReport' }" active-class="!bg-primary/10 !text-primary font-bold"
                @click="isMobileNavOpen = false" class="py-2">
                <Icon icon="lucide:file-text" class="w-4 h-4" /> {{ $t('menu.logReport') }}
              </router-link>
            </li>

            <li
              v-if="hasPermission('User', 'Display') || hasPermission('Role', 'Display') || hasPermission('Device', 'Display') || hasPermission('Device Group', 'Display')">
              <details name="mobile-header-dropdown">
                <summary class="py-2">
                  <Icon icon="lucide:settings" class="w-4 h-4" /> {{ $t('menu.management') }}
                </summary>
                <ul class="ml-2 pl-2 border-l border-base-300">
                  <li v-if="hasPermission('User', 'Display')"><router-link :to="{ name: 'user' }"
                      @click="isMobileNavOpen = false">{{ $t('menu.user') }}</router-link></li>
                  <li v-if="hasPermission('Role', 'Display')"><router-link :to="{ name: 'role' }"
                      @click="isMobileNavOpen = false">{{ $t('menu.role') }}</router-link></li>
                  <li v-if="hasPermission('Device', 'Display')"><router-link :to="{ name: 'device' }"
                      @click="isMobileNavOpen = false">{{ $t('menu.device') }}</router-link></li>
                  <li v-if="hasPermission('Device Group', 'Display')"><router-link :to="{ name: 'deviceGroup' }"
                      @click="isMobileNavOpen = false">{{ $t('menu.deviceGroup') }}</router-link></li>
                </ul>
              </details>
            </li>
          </ul>
        </div>

        <!-- Mobile Drawer Footer Controls -->
        <div class="pt-3 border-t border-base-300 flex flex-col gap-2">
          <div class="join w-full bg-base-200 p-0.5 rounded-md border border-base-300">
            <button type="button" class="btn h-6 min-h-0 join-item flex-1 text-xs font-bold border-none"
              :class="locale === 'en' ? 'bg-base-100 text-base-content' : 'bg-transparent text-base-content/50'"
              @click="setLanguage('en')">EN</button>
            <button type="button" class="btn h-6 min-h-0 join-item flex-1 text-xs font-bold border-none"
              :class="locale === 'th' ? 'bg-base-100 text-base-content' : 'bg-transparent text-base-content/50'"
              @click="setLanguage('th')">TH</button>
          </div>
          <div class="flex items-center justify-center gap-1.5 opacity-50 py-1">
            <span class="text-[9px] uppercase tracking-wider text-base-content/60 font-semibold">{{
              $t('header.poweredBy') }}</span>
            <img src="/LOGO_FT.png" alt="Company Logo" class="h-3 w-auto object-contain" />
          </div>
        </div>

      </div>
    </aside>

  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { Icon } from '@iconify/vue';
import { useMutation } from '@/composables/useMutation';
import { usePermissionStore } from '@/stores/usePermissionStore';
import { useUserStore } from '@/stores/useUserStore';
import { useThemeStore } from '@/stores/useThemeStore';

const router = useRouter();
const { locale } = useI18n();

const isMobileNavOpen = ref(false);

const permissionStore = usePermissionStore();
const { hasPermission, setPermissions } = permissionStore;
const userStore = useUserStore();
const { setUser } = userStore;
const themeStore = useThemeStore();

const { res: logoutRes, execute: logoutApi } = useMutation();

const closeAllDropdowns = () => {
  document.querySelectorAll('details[name="header-dropdown"]').forEach((el) => {
    el.removeAttribute('open');
  });
};

const handleClickOutside = (event) => {
  if (!event.target.closest('details[name="header-dropdown"]')) {
    closeAllDropdowns();
  }
};

onMounted(() => {
  const savedLang = localStorage.getItem('lang') || 'en';
  locale.value = savedLang;
  window.addEventListener('click', handleClickOutside);
});

onUnmounted(() => {
  window.removeEventListener('click', handleClickOutside);
});

const setLanguage = (lang) => {
  locale.value = lang;
  localStorage.setItem('lang', lang);
};

const handleLogout = async () => {
  try {
    await logoutApi('/user/logout', null, 'POST');
    if (!logoutRes.value.ok) {
      console.warn("Server-side logout failed (Response not OK), but proceeding with local logout.");
    }
  } catch (error) {
    console.error("Network error during logout:", error);
  } finally {
    setUser(null);
    setPermissions(null);
    router.push('/login');
  }
};
</script>