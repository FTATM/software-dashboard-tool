<template>
  <!-- If the route asks to hide the layout (like the Login page) -->
  <template v-if="$route.meta.hideLayout">
    <router-view />
  </template>

  <!-- Default behavior: Use the MainLayout (like the Dashboard) -->
  <template v-else>
    <MainLayout>
      <router-view />
    </MainLayout>
  </template>
</template>

<script setup>
import MainLayout from './layouts/MainLayout.vue';

const appVersion = import.meta.env.VITE_APP_VERSION || 'dev';
const commitSha = import.meta.env.VITE_COMMIT_SHA || 'none';
const buildTime = import.meta.env.VITE_BUILD_TIME || 'unknown';

console.log(`Version: ${appVersion} (${commitSha}) built at ${buildTime}`);
</script>

<style>

#app {
  width: 100%;
  height: 100%;
}
</style>