<script setup lang="ts">
import { ref, onMounted } from 'vue'
import {
  Building2,
  Users,
  UserCircle,
  AlertCircle,
} from 'lucide-vue-next'
import AppLayout from '@/layouts/AppLayout.vue'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { useAuthStore } from '@/stores/auth'
import { organizationService } from '@/services/organizations'
import { groupService } from '@/services/groups'
import { relawanService } from '@/services/relawan'

const authStore = useAuthStore()
const isLoading = ref(true)
const loadError = ref<string | null>(null)

const stats = ref({
  organizations: 0,
  groups: 0,
  relawan: 0,
  activeRelawan: 0,
})

async function loadStats() {
  isLoading.value = true
  loadError.value = null

  try {
    // Load data in parallel
    const [orgsData, groupsData, relawanStats] = await Promise.all([
      organizationService.list({ page: 1, page_size: 1 }),
      groupService.list({ page: 1, page_size: 1 }),
      relawanService.getStats(),
    ])

    stats.value = {
      organizations: orgsData.total,
      groups: groupsData.total,
      relawan: relawanStats.total,
      activeRelawan: relawanStats.active,
    }
  } catch (error) {
    console.error('Failed to load dashboard stats:', error)
    loadError.value = 'Gagal memuat data dashboard'
  } finally {
    isLoading.value = false
  }
}

onMounted(async () => {
  await authStore.init()
  await loadStats()
})
</script>

<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Welcome section -->
      <div>
        <h2 class="text-2xl font-bold tracking-tight">
          Selamat datang, {{ authStore.displayName || 'Admin' }}
        </h2>
        <p class="text-muted-foreground">
          Berikut ringkasan data terkini dari sistem.
        </p>
      </div>

      <!-- Error state -->
      <Card v-if="loadError" class="border-destructive">
        <CardContent class="flex items-center gap-3 py-4">
          <AlertCircle class="h-5 w-5 text-destructive" />
          <p class="text-sm text-destructive">{{ loadError }}</p>
        </CardContent>
      </Card>

      <!-- Stats grid -->
      <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <!-- Organizations -->
        <Card>
          <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle class="text-sm font-medium">Total Organisasi</CardTitle>
            <Building2 class="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <Skeleton v-if="isLoading" class="h-8 w-20" />
            <div v-else class="text-2xl font-bold font-mono">{{ stats.organizations }}</div>
          </CardContent>
        </Card>

        <!-- Groups -->
        <Card>
          <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle class="text-sm font-medium">Total Tim/Grup</CardTitle>
            <Users class="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <Skeleton v-if="isLoading" class="h-8 w-20" />
            <div v-else class="text-2xl font-bold font-mono">{{ stats.groups }}</div>
          </CardContent>
        </Card>

        <!-- Relawan -->
        <Card>
          <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle class="text-sm font-medium">Total Relawan</CardTitle>
            <UserCircle class="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <Skeleton v-if="isLoading" class="h-8 w-20" />
            <div v-else class="text-2xl font-bold font-mono">{{ stats.relawan }}</div>
          </CardContent>
        </Card>

        <!-- Active Relawan -->
        <Card>
          <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle class="text-sm font-medium">Relawan Aktif</CardTitle>
            <UserCircle class="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <Skeleton v-if="isLoading" class="h-8 w-20" />
            <div v-else class="text-2xl font-bold font-mono">{{ stats.activeRelawan }}</div>
          </CardContent>
        </Card>
      </div>

      <!-- Quick links -->
      <div class="grid gap-4 md:grid-cols-3">
        <Card class="cursor-pointer hover:bg-muted/50 transition-colors" @click="$router.push('/organizations')">
          <CardHeader>
            <CardTitle class="flex items-center gap-2">
              <Building2 class="h-5 w-5" />
              Kelola Organisasi
            </CardTitle>
            <CardDescription>
              Lihat dan kelola daftar organisasi
            </CardDescription>
          </CardHeader>
        </Card>

        <Card class="cursor-pointer hover:bg-muted/50 transition-colors" @click="$router.push('/groups')">
          <CardHeader>
            <CardTitle class="flex items-center gap-2">
              <Users class="h-5 w-5" />
              Kelola Tim/Grup
            </CardTitle>
            <CardDescription>
              Lihat dan kelola daftar tim atau grup
            </CardDescription>
          </CardHeader>
        </Card>

        <Card class="cursor-pointer hover:bg-muted/50 transition-colors" @click="$router.push('/relawan')">
          <CardHeader>
            <CardTitle class="flex items-center gap-2">
              <UserCircle class="h-5 w-5" />
              Kelola Relawan
            </CardTitle>
            <CardDescription>
              Lihat dan kelola daftar relawan
            </CardDescription>
          </CardHeader>
        </Card>
      </div>
    </div>
  </AppLayout>
</template>
