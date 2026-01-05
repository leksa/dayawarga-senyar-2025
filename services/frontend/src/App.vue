<script setup lang="ts">
import AppHeader from './components/AppHeader.vue'
</script>

<template>
  <div class="h-screen flex flex-col bg-gray-100">
    <AppHeader />
    <router-view />
  </div>
</template>
