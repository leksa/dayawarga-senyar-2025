<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import { HelpCircle, Clock } from 'lucide-vue-next'
import { api } from '@/services/api'

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
  <header class="h-14 bg-white border-b border-gray-200 flex items-center justify-between px-4">
    <!-- Logo -->
    <RouterLink to="/" class="flex items-center gap-3 hover:opacity-80 transition-opacity">
      <img src="/logo.png" alt="Dayawarga Logo" class="w-10 h-10 object-contain" />
      <span class="font-semibold text-gray-900">Peta Kondisi Bencana Siklon Senyar Sumatra 2025</span>
    </RouterLink>

    <!-- Right side -->
    <div class="flex items-center gap-4">
      <div class="flex items-center gap-1 text-gray-500 text-sm">
        <Clock class="w-4 h-4" />
        <span>Data terakhir diperbaharui: {{ lastSyncTime || '-' }}</span>
      </div>

      <RouterLink to="/tentang" class="flex items-center gap-1 text-gray-600 hover:text-gray-900">
        <HelpCircle class="w-4 h-4" />
        <span class="text-sm">Support</span>
      </RouterLink>
    </div>
  </header>
</template>
