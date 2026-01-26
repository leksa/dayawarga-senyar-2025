<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  ArrowLeft,
  Building2,
  Users,
  UserCircle,
  Pencil,
  Trash2,
  Plus,
  AlertCircle,
} from 'lucide-vue-next'
import AppLayout from '@/layouts/AppLayout.vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Skeleton } from '@/components/ui/skeleton'
import { organizationService } from '@/services/organizations'
import { groupService } from '@/services/groups'
import { relawanService } from '@/services/relawan'
import type { Organization, OrganizationStats, Group, Relawan } from '@/services/types'

const route = useRoute()
const router = useRouter()

const isLoading = ref(true)
const loadError = ref<string | null>(null)
const organization = ref<Organization | null>(null)
const stats = ref<OrganizationStats | null>(null)
const groups = ref<Group[]>([])
const relawan = ref<Relawan[]>([])

const orgId = computed(() => route.params.id as string)

async function loadData() {
  isLoading.value = true
  loadError.value = null

  try {
    // Load organization and stats in parallel
    const [orgData, orgStats, groupsData, relawanData] = await Promise.all([
      organizationService.get(orgId.value),
      organizationService.getStats(orgId.value),
      groupService.list({ organization_id: orgId.value, page_size: 100 }),
      relawanService.list({ organization_id: orgId.value, page_size: 100 }),
    ])

    organization.value = orgData
    stats.value = orgStats
    groups.value = groupsData.groups || []
    relawan.value = relawanData.relawan || []
  } catch (error) {
    console.error('Failed to load organization:', error)
    loadError.value = 'Gagal memuat data organisasi'
  } finally {
    isLoading.value = false
  }
}

onMounted(() => {
  loadData()
})

function goBack() {
  router.push('/organizations')
}

function viewGroup(group: Group) {
  router.push(`/groups/${group.id}`)
}

function viewRelawan(r: Relawan) {
  router.push(`/relawan/${r.id}`)
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
            {{ organization?.name || 'Organisasi' }}
          </h2>
          <Skeleton v-if="isLoading" class="h-5 w-96 mt-1" />
          <p v-else class="text-muted-foreground">
            {{ organization?.description || '-' }}
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

      <!-- Error state -->
      <Card v-if="loadError" class="border-destructive">
        <CardContent class="flex items-center gap-3 py-4">
          <AlertCircle class="h-5 w-5 text-destructive" />
          <p class="text-sm text-destructive">{{ loadError }}</p>
        </CardContent>
      </Card>

      <!-- Stats -->
      <div class="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle class="text-sm font-medium">Status</CardTitle>
            <Building2 class="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <Skeleton v-if="isLoading" class="h-6 w-16" />
            <Badge v-else :variant="organization?.is_active ? 'default' : 'secondary'">
              {{ organization?.is_active ? 'Aktif' : 'Nonaktif' }}
            </Badge>
          </CardContent>
        </Card>

        <Card>
          <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle class="text-sm font-medium">Total Tim</CardTitle>
            <Users class="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <Skeleton v-if="isLoading" class="h-8 w-12" />
            <div v-else class="text-2xl font-bold font-mono">
              {{ stats?.total_groups || 0 }}
            </div>
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
              {{ stats?.total_relawan || 0 }}
            </div>
          </CardContent>
        </Card>
      </div>

      <!-- Tabs -->
      <Tabs default-value="groups" class="space-y-4">
        <TabsList>
          <TabsTrigger value="groups">Tim/Grup</TabsTrigger>
          <TabsTrigger value="relawan">Relawan</TabsTrigger>
        </TabsList>

        <TabsContent value="groups" class="space-y-4">
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-semibold">Daftar Tim</h3>
            <Button size="sm" @click="$router.push({ path: '/groups', query: { org: orgId, action: 'create' } })">
              <Plus class="mr-2 h-4 w-4" />
              Tambah Tim
            </Button>
          </div>

          <Card>
            <CardContent class="p-0">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Nama Tim</TableHead>
                    <TableHead class="text-center">Status</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  <TableRow v-if="isLoading">
                    <TableCell colspan="2">
                      <Skeleton class="h-10 w-full" />
                    </TableCell>
                  </TableRow>
                  <TableRow v-else-if="groups.length === 0">
                    <TableCell colspan="2" class="text-center text-muted-foreground py-8">
                      Belum ada tim dalam organisasi ini
                    </TableCell>
                  </TableRow>
                  <TableRow
                    v-else
                    v-for="group in groups"
                    :key="group.id"
                    class="cursor-pointer"
                    @click="viewGroup(group)"
                  >
                    <TableCell class="font-medium">{{ group.name }}</TableCell>
                    <TableCell class="text-center">
                      <Badge :variant="group.is_active ? 'default' : 'secondary'">
                        {{ group.is_active ? 'Aktif' : 'Nonaktif' }}
                      </Badge>
                    </TableCell>
                  </TableRow>
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="relawan" class="space-y-4">
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-semibold">Daftar Relawan</h3>
            <Button size="sm" @click="$router.push({ path: '/relawan', query: { org: orgId, action: 'create' } })">
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
                    <TableHead class="text-center">Status</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  <TableRow v-if="isLoading">
                    <TableCell colspan="3">
                      <Skeleton class="h-10 w-full" />
                    </TableCell>
                  </TableRow>
                  <TableRow v-else-if="relawan.length === 0">
                    <TableCell colspan="3" class="text-center text-muted-foreground py-8">
                      Belum ada relawan dalam organisasi ini
                    </TableCell>
                  </TableRow>
                  <TableRow
                    v-else
                    v-for="r in relawan"
                    :key="r.id"
                    class="cursor-pointer"
                    @click="viewRelawan(r)"
                  >
                    <TableCell class="font-medium">{{ r.name }}</TableCell>
                    <TableCell class="font-mono text-sm">{{ r.phone || '-' }}</TableCell>
                    <TableCell class="text-center">
                      <Badge :variant="r.status === 'active' ? 'default' : 'secondary'">
                        {{ r.status === 'active' ? 'Aktif' : r.status === 'inactive' ? 'Nonaktif' : 'Ditangguhkan' }}
                      </Badge>
                    </TableCell>
                  </TableRow>
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  </AppLayout>
</template>
