<script setup lang="ts">
import { onMounted, computed, ref } from 'vue'
import { Toaster } from '@/components/ui/sonner'
import { useColorMode } from '@/composables/useColorMode'
import { useAuthStore } from '@/stores/auth'

// Initialize color mode (dark as default)
const { mode } = useColorMode()
const authStore = useAuthStore()
const isAuthReady = ref(false)

// Convert mode to sonner theme type
const toasterTheme = computed(() => {
  if (mode.value === 'dark') return 'dark'
  if (mode.value === 'light') return 'light'
  return 'system'
})

onMounted(async () => {
  // Ensure dark mode class is set on initial load
  if (!localStorage.getItem('admin-portal-color-mode')) {
    document.documentElement.classList.add('dark')
  }

  // Initialize auth store before rendering routes
  await authStore.init()
  isAuthReady.value = true
})
</script>

<template>
  <RouterView v-if="isAuthReady" />
  <div v-else class="min-h-screen flex items-center justify-center">
    <div class="animate-spin h-8 w-8 border-4 border-primary border-t-transparent rounded-full"></div>
  </div>
  <Toaster position="top-right" :theme="toasterTheme" richColors />
</template>
