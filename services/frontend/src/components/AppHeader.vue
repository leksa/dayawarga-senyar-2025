<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import { HelpCircle, Clock, Menu } from 'lucide-vue-next'
import { api } from '@/services/api'
import { useSidebar } from '@/composables/useSidebar'

const { toggle: toggleSidebar } = useSidebar()

const lastSyncTime = ref<string>('')

const formatDateTime = (isoString: string): string => {
  if (!isoString) return '-'
  const date = new Date(isoString)
  const day = date.getDate().toString().padStart(2, '0')
  const month = (date.getMonth() + 1).toString().padStart(2, '0')
  const year = date.getFullYear()
  const hours = date.getHours().toString().padStart(2, '0')
  const minutes = date.getMinutes().toString().padStart(2, '0')
  return `${day}-${month}-${year} ${hours}:${minutes}`
}

onMounted(async () => {
  const response = await api.getSyncStatus()
  if (response.success && response.data) {
    lastSyncTime.value = formatDateTime(response.data.last_sync_time)
  }
})
</script>

<template>
  <header class="h-14 bg-white border-b border-gray-200 flex items-center justify-between px-2 md:px-4">
    <!-- Left: Menu button (mobile) + Logo -->
    <div class="flex items-center gap-2">
      <!-- Mobile menu button -->
      <button
        @click="toggleSidebar"
        class="lg:hidden p-2 -ml-1 rounded-lg hover:bg-gray-100 transition-colors"
        aria-label="Toggle menu"
      >
        <Menu class="w-6 h-6 text-gray-600" />
      </button>

      <!-- Logo -->
      <RouterLink to="/" class="flex items-center gap-2 md:gap-3 hover:opacity-80 transition-opacity">
        <img src="/logo.png" alt="Dayawarga Logo" class="w-8 h-8 md:w-10 md:h-10 object-contain" />
        <span class="font-semibold text-gray-900 text-sm md:text-base hidden sm:inline">
          Peta Bencana Siklon Senyar
        </span>
        <span class="font-semibold text-gray-900 text-sm md:text-base sm:hidden">
          Dayawarga
        </span>
      </RouterLink>
    </div>

    <!-- Right side -->
    <div class="flex items-center gap-2 md:gap-4">
      <!-- Last sync - hidden on mobile -->
      <div class="hidden md:flex items-center gap-1 text-gray-500 text-sm">
        <Clock class="w-4 h-4" />
        <span>Update: {{ lastSyncTime || '-' }}</span>
      </div>

      <!-- Support link -->
      <RouterLink to="/tentang" class="flex items-center gap-1 text-gray-600 hover:text-gray-900 p-2 -mr-2 md:mr-0">
        <HelpCircle class="w-5 h-5 md:w-4 md:h-4" />
        <span class="text-sm hidden md:inline">Support</span>
      </RouterLink>
    </div>
  </header>
</template>
