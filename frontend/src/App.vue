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
</script>

<style>

#app {
  width: 100%;
  height: 100%;
}
</style>