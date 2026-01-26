<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  ArrowLeft,
  Users,
  Building2,
  Phone,
  Mail,
  Calendar,
  Pencil,
  Trash2,
  MessageSquare,
  Loader2,
} from 'lucide-vue-next'
import AppLayout from '@/layouts/AppLayout.vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import { Skeleton } from '@/components/ui/skeleton'
import { relawanService } from '@/services/relawan'
import type { Relawan } from '@/services/types'
import { toast } from 'vue-sonner'

const route = useRoute()
const router = useRouter()

const isLoading = ref(true)
const error = ref<string | null>(null)
const relawan = ref<Relawan | null>(null)
const isTogglingWA = ref(false)

const formattedCreatedAt = computed(() => {
  if (!relawan.value?.created_at) return '-'
  return new Date(relawan.value.created_at).toLocaleDateString('id-ID', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  })
})

const formattedWALastActivity = computed(() => {
  if (!relawan.value?.wa_last_activity) return null
  return new Date(relawan.value.wa_last_activity).toLocaleDateString('id-ID', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
})

async function fetchRelawan() {
  const id = route.params.id as string
  if (!id) {
    error.value = 'ID relawan tidak valid'
    isLoading.value = false
    return
  }

  try {
    isLoading.value = true
    error.value = null
    relawan.value = await relawanService.get(id)
  } catch (err: any) {
    error.value = err?.response?.data?.error || err?.message || 'Gagal memuat data relawan'
    toast.error(error.value || 'Terjadi kesalahan')
  } finally {
    isLoading.value = false
  }
}

async function toggleWAAccess() {
  if (!relawan.value || isTogglingWA.value) return

  const id = relawan.value.id
  const currentlyVerified = relawan.value.wa_verified

  try {
    isTogglingWA.value = true

    if (currentlyVerified) {
      await relawanService.revokeWAAccess(id)
      relawan.value.wa_verified = false
      relawan.value.wa_verified_at = null
      toast.success('Akses WhatsApp telah dinonaktifkan')
    } else {
      await relawanService.enableWAAccess(id)
      relawan.value.wa_verified = true
      relawan.value.wa_verified_at = new Date().toISOString()
      toast.success('Akses WhatsApp telah diaktifkan')
    }
  } catch (err: any) {
    const errorMsg = err?.response?.data?.error || err?.message || 'Gagal mengubah status WhatsApp'
    toast.error(errorMsg)
  } finally {
    isTogglingWA.value = false
  }
}

onMounted(() => {
  fetchRelawan()
})

function goBack() {
  router.push('/relawan')
}

function viewGroup() {
  if (relawan.value?.group_id) {
    router.push(`/groups/${relawan.value.group_id}`)
  }
}

function viewOrganization() {
  if (relawan.value?.organization_id) {
    router.push(`/organizations/${relawan.value.organization_id}`)
  }
}
</script>

<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Header -->
      <div class="flex items-center gap-4">
        <Button variant="ghost" size="icon" @click="goBack">
          <ArrowLeft class="h-5 w-5" />
        </Button>
        <div class="flex-1">
          <Skeleton v-if="isLoading" class="h-8 w-48" />
          <h2 v-else class="text-2xl font-bold tracking-tight">
            {{ relawan?.name }}
          </h2>
          <Skeleton v-if="isLoading" class="h-5 w-64 mt-1" />
          <div v-else class="flex items-center gap-2 mt-1">
            <Badge :variant="relawan?.status === 'active' ? 'default' : 'secondary'">
              {{ relawan?.status === 'active' ? 'Aktif' : 'Nonaktif' }}
            </Badge>
            <Badge
              v-if="relawan?.wa_verified"
              variant="outline"
              class="border-green-500 text-green-600 bg-green-50"
            >
              <MessageSquare class="mr-1 h-3 w-3" />
              WhatsApp Aktif
            </Badge>
          </div>
        </div>
        <div class="flex items-center gap-2">
          <Button variant="outline" size="sm">
            <Pencil class="mr-2 h-4 w-4" />
            Edit
          </Button>
          <Button variant="destructive" size="sm">
            <Trash2 class="mr-2 h-4 w-4" />
            Hapus
          </Button>
        </div>
      </div>

      <!-- Content -->
      <div class="grid gap-6 md:grid-cols-3">
        <!-- Main Info -->
        <Card class="md:col-span-2">
          <CardHeader>
            <CardTitle>Informasi Relawan</CardTitle>
          </CardHeader>
          <CardContent class="space-y-6">
            <!-- Contact Info -->
            <div class="space-y-4">
              <h4 class="text-sm font-medium text-muted-foreground uppercase tracking-wide">
                Kontak
              </h4>
              <div class="grid gap-4 sm:grid-cols-2">
                <div class="flex items-center gap-3">
                  <div class="flex h-10 w-10 items-center justify-center rounded-lg bg-muted">
                    <Phone class="h-5 w-5 text-muted-foreground" />
                  </div>
                  <div>
                    <p class="text-sm text-muted-foreground">Telepon</p>
                    <Skeleton v-if="isLoading" class="h-5 w-28" />
                    <p v-else class="font-mono">{{ relawan?.phone }}</p>
                  </div>
                </div>
                <div class="flex items-center gap-3">
                  <div class="flex h-10 w-10 items-center justify-center rounded-lg bg-muted">
                    <Mail class="h-5 w-5 text-muted-foreground" />
                  </div>
                  <div>
                    <p class="text-sm text-muted-foreground">Email</p>
                    <Skeleton v-if="isLoading" class="h-5 w-40" />
                    <p v-else>{{ relawan?.email }}</p>
                  </div>
                </div>
              </div>
            </div>

            <Separator />

            <!-- Organization Info -->
            <div class="space-y-4">
              <h4 class="text-sm font-medium text-muted-foreground uppercase tracking-wide">
                Afiliasi
              </h4>
              <div class="grid gap-4 sm:grid-cols-2">
                <div class="flex items-center gap-3">
                  <div class="flex h-10 w-10 items-center justify-center rounded-lg bg-muted">
                    <Users class="h-5 w-5 text-muted-foreground" />
                  </div>
                  <div>
                    <p class="text-sm text-muted-foreground">Tim</p>
                    <Skeleton v-if="isLoading" class="h-5 w-32" />
                    <template v-else>
                      <Button
                        v-if="relawan?.group"
                        variant="link"
                        class="h-auto p-0 text-base"
                        @click="viewGroup"
                      >
                        {{ relawan?.group?.name }}
                      </Button>
                      <span v-else class="text-muted-foreground">-</span>
                    </template>
                  </div>
                </div>
                <div class="flex items-center gap-3">
                  <div class="flex h-10 w-10 items-center justify-center rounded-lg bg-muted">
                    <Building2 class="h-5 w-5 text-muted-foreground" />
                  </div>
                  <div>
                    <p class="text-sm text-muted-foreground">Organisasi</p>
                    <Skeleton v-if="isLoading" class="h-5 w-28" />
                    <template v-else>
                      <Button
                        v-if="relawan?.organization"
                        variant="link"
                        class="h-auto p-0 text-base"
                        @click="viewOrganization"
                      >
                        {{ relawan?.organization?.name }}
                      </Button>
                      <span v-else class="text-muted-foreground">-</span>
                    </template>
                  </div>
                </div>
              </div>
            </div>

            <Separator />

            <!-- Notes -->
            <div class="space-y-4">
              <h4 class="text-sm font-medium text-muted-foreground uppercase tracking-wide">
                Catatan
              </h4>
              <Skeleton v-if="isLoading" class="h-16 w-full" />
              <p v-else class="text-sm leading-relaxed">
                {{ relawan?.notes || 'Tidak ada catatan.' }}
              </p>
            </div>
          </CardContent>
        </Card>

        <!-- Side Info -->
        <div class="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle class="text-base">Status</CardTitle>
            </CardHeader>
            <CardContent>
              <Skeleton v-if="isLoading" class="h-6 w-16" />
              <Badge
                v-else
                :variant="relawan?.status === 'active' ? 'default' : 'secondary'"
                class="text-sm"
              >
                {{ relawan?.status === 'active' ? 'Aktif' : 'Nonaktif' }}
              </Badge>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle class="text-base">Terdaftar</CardTitle>
            </CardHeader>
            <CardContent>
              <div class="flex items-center gap-2">
                <Calendar class="h-4 w-4 text-muted-foreground" />
                <Skeleton v-if="isLoading" class="h-5 w-24" />
                <span v-else class="text-sm">{{ formattedCreatedAt }}</span>
              </div>
            </CardContent>
          </Card>

          <!-- WhatsApp Access Card -->
          <Card>
            <CardHeader>
              <CardTitle class="text-base flex items-center gap-2">
                <MessageSquare class="h-4 w-4" />
                Akses WhatsApp
              </CardTitle>
            </CardHeader>
            <CardContent class="space-y-4">
              <Skeleton v-if="isLoading" class="h-6 w-28" />
              <template v-else>
                <!-- Status Badge -->
                <div class="flex items-center justify-between">
                  <span class="text-sm text-muted-foreground">Status</span>
                  <Badge
                    :variant="relawan?.wa_verified ? 'default' : 'secondary'"
                    :class="relawan?.wa_verified ? 'bg-green-500 hover:bg-green-600' : ''"
                  >
                    {{ relawan?.wa_verified ? 'Aktif' : 'Nonaktif' }}
                  </Badge>
                </div>

                <!-- Session Count -->
                <div class="flex items-center justify-between">
                  <span class="text-sm text-muted-foreground">Sesi</span>
                  <span class="text-sm font-medium">{{ relawan?.wa_session_count || 0 }}</span>
                </div>

                <!-- Last Activity -->
                <div v-if="formattedWALastActivity" class="flex items-center justify-between">
                  <span class="text-sm text-muted-foreground">Aktivitas Terakhir</span>
                  <span class="text-sm">{{ formattedWALastActivity }}</span>
                </div>

                <Separator />

                <!-- Toggle Button -->
                <Button
                  :variant="relawan?.wa_verified ? 'destructive' : 'default'"
                  size="sm"
                  class="w-full"
                  :disabled="isTogglingWA || !relawan?.phone"
                  @click="toggleWAAccess"
                >
                  <Loader2 v-if="isTogglingWA" class="mr-2 h-4 w-4 animate-spin" />
                  <MessageSquare v-else class="mr-2 h-4 w-4" />
                  {{ relawan?.wa_verified ? 'Nonaktifkan' : 'Aktifkan' }}
                </Button>

                <p v-if="!relawan?.phone" class="text-xs text-muted-foreground text-center">
                  Nomor telepon diperlukan untuk mengaktifkan WhatsApp
                </p>
              </template>
            </CardContent>
          </Card>
        </div>
      </div>

      <!-- Error State -->
      <Card v-if="error && !isLoading" class="border-destructive">
        <CardContent class="pt-6">
          <div class="text-center space-y-2">
            <p class="text-destructive font-medium">{{ error }}</p>
            <Button variant="outline" size="sm" @click="fetchRelawan">
              Coba Lagi
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  </AppLayout>
</template>
