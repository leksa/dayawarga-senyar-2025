<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  ArrowLeft,
  Users,
  UserCircle,
  Building2,
  Pencil,
  Trash2,
  Plus,
} from 'lucide-vue-next'
import AppLayout from '@/layouts/AppLayout.vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Skeleton } from '@/components/ui/skeleton'

const route = useRoute()
const router = useRouter()

const isLoading = ref(true)
const group = ref<any>(null)

// Mock relawan
const relawan = ref([
  { id: '1', name: 'Ahmad Suryadi', phone: '081234567890', role: 'Koordinator', status: 'active' },
  { id: '2', name: 'Siti Nurhaliza', phone: '081234567891', role: 'Anggota', status: 'active' },
  { id: '3', name: 'Budi Santoso', phone: '081234567892', role: 'Anggota', status: 'active' },
  { id: '4', name: 'Dewi Lestari', phone: '081234567893', role: 'Anggota', status: 'active' },
])

onMounted(() => {
  setTimeout(() => {
    group.value = {
      id: route.params.id,
      name: 'Tim Logistik Bandung',
      description: 'Tim penanganan logistik wilayah Bandung',
      organizationId: '1',
      organizationName: 'PMI Jawa Barat',
      relawanCount: 15,
      status: 'active',
      createdAt: '2024-01-20',
    }
    isLoading.value = false
  }, 500)
})

function goBack() {
  router.push('/groups')
}

function viewRelawan(r: any) {
  router.push(`/relawan/${r.id}`)
}

function viewOrganization() {
  if (group.value?.organizationId) {
    router.push(`/organizations/${group.value.organizationId}`)
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
            {{ group?.name }}
          </h2>
          <Skeleton v-if="isLoading" class="h-5 w-96 mt-1" />
          <p v-else class="text-muted-foreground">
            {{ group?.description }}
          </p>
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

      <!-- Info Cards -->
      <div class="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle class="text-sm font-medium">Organisasi</CardTitle>
            <Building2 class="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <Skeleton v-if="isLoading" class="h-6 w-32" />
            <Button
              v-else
              variant="link"
              class="h-auto p-0 text-base font-semibold"
              @click="viewOrganization"
            >
              {{ group?.organizationName }}
            </Button>
          </CardContent>
        </Card>

        <Card>
          <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle class="text-sm font-medium">Status</CardTitle>
            <Users class="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <Skeleton v-if="isLoading" class="h-6 w-16" />
            <Badge v-else :variant="group?.status === 'active' ? 'default' : 'secondary'">
              {{ group?.status === 'active' ? 'Aktif' : 'Nonaktif' }}
            </Badge>
          </CardContent>
        </Card>

        <Card>
          <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle class="text-sm font-medium">Total Relawan</CardTitle>
            <UserCircle class="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <Skeleton v-if="isLoading" class="h-8 w-12" />
            <div v-else class="text-2xl font-bold font-mono">
              {{ group?.relawanCount }}
            </div>
          </CardContent>
        </Card>
      </div>

      <!-- Relawan Table -->
      <div class="space-y-4">
        <div class="flex items-center justify-between">
          <h3 class="text-lg font-semibold">Daftar Relawan</h3>
          <Button size="sm">
            <Plus class="mr-2 h-4 w-4" />
            Tambah Relawan
          </Button>
        </div>

        <Card>
          <CardContent class="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Nama</TableHead>
                  <TableHead>Telepon</TableHead>
                  <TableHead>Peran</TableHead>
                  <TableHead class="text-center">Status</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow
                  v-for="r in relawan"
                  :key="r.id"
                  class="cursor-pointer"
                  @click="viewRelawan(r)"
                >
                  <TableCell class="font-medium">{{ r.name }}</TableCell>
                  <TableCell class="font-mono text-sm">{{ r.phone }}</TableCell>
                  <TableCell>
                    <Badge variant="outline">{{ r.role }}</Badge>
                  </TableCell>
                  <TableCell class="text-center">
                    <Badge :variant="r.status === 'active' ? 'default' : 'secondary'">
                      {{ r.status === 'active' ? 'Aktif' : 'Nonaktif' }}
                    </Badge>
                  </TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      </div>
    </div>
  </AppLayout>
</template>
