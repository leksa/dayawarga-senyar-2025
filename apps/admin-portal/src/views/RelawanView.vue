<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { toast } from 'vue-sonner'
import {
  Plus,
  Search,
  MoreHorizontal,
  UserCircle,
  Eye,
  Pencil,
  Trash2,
  Users,
  Building2,
  ChevronLeft,
  ChevronRight,
} from 'lucide-vue-next'
import AppLayout from '@/layouts/AppLayout.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Label } from '@/components/ui/label'
import { Skeleton } from '@/components/ui/skeleton'
import {
  relawanService,
  organizationService,
  groupService,
  type Relawan,
  type Organization,
  type Group,
} from '@/services'

const router = useRouter()

// State
const isLoading = ref(false)
const isLoadingOrgs = ref(false)
const isLoadingGroups = ref(false)
const isSaving = ref(false)
const searchQuery = ref('')
const selectedOrgFilter = ref<string>('all')
const selectedGroupFilter = ref<string>('all')
const isCreateDialogOpen = ref(false)
const isEditDialogOpen = ref(false)
const isDeleteDialogOpen = ref(false)
const selectedRelawan = ref<Relawan | null>(null)

// Pagination
const currentPage = ref(1)
const pageSize = ref(20)
const totalItems = ref(0)
const totalPages = ref(0)

// Form state
const formData = ref({
  name: '',
  phone: '',
  email: '',
  organizationId: '',
  groupId: '',
})

// Data
const organizations = ref<Organization[]>([])
const allGroups = ref<Group[]>([])
const relawanList = ref<Relawan[]>([])

// Filtered groups based on selected organization
const availableGroups = computed(() => {
  if (!formData.value.organizationId) return []
  return allGroups.value.filter(g => g.organization_id === formData.value.organizationId)
})

// Groups for filter dropdown (based on selected org filter)
const filterGroups = computed(() => {
  if (selectedOrgFilter.value === 'all') return allGroups.value
  return allGroups.value.filter(g => g.organization_id === selectedOrgFilter.value)
})

// Debounce search
let searchTimeout: ReturnType<typeof setTimeout> | null = null

watch(searchQuery, () => {
  if (searchTimeout) clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    currentPage.value = 1
    fetchRelawan()
  }, 300)
})

watch(selectedOrgFilter, () => {
  selectedGroupFilter.value = 'all'
  currentPage.value = 1
  fetchRelawan()
})

watch(selectedGroupFilter, () => {
  currentPage.value = 1
  fetchRelawan()
})

// Reset group when org changes in form
watch(() => formData.value.organizationId, () => {
  formData.value.groupId = ''
})

async function fetchOrganizations() {
  isLoadingOrgs.value = true
  try {
    const response = await organizationService.list({ page_size: 100 })
    organizations.value = response.organizations
  } catch (error) {
    console.error('Failed to fetch organizations:', error)
    toast.error('Gagal memuat data organisasi')
  } finally {
    isLoadingOrgs.value = false
  }
}

async function fetchGroups() {
  isLoadingGroups.value = true
  try {
    const response = await groupService.list({ page_size: 100 })
    allGroups.value = response.groups
  } catch (error) {
    console.error('Failed to fetch groups:', error)
    toast.error('Gagal memuat data tim')
  } finally {
    isLoadingGroups.value = false
  }
}

async function fetchRelawan() {
  isLoading.value = true
  try {
    const response = await relawanService.list({
      organization_id: selectedOrgFilter.value !== 'all' ? selectedOrgFilter.value : undefined,
      group_id: selectedGroupFilter.value !== 'all' ? selectedGroupFilter.value : undefined,
      search: searchQuery.value || undefined,
      page: currentPage.value,
      page_size: pageSize.value,
    })

    relawanList.value = response.relawan
    totalItems.value = response.total
    totalPages.value = response.total_pages
  } catch (error) {
    console.error('Failed to fetch relawan:', error)
    toast.error('Gagal memuat data relawan')
  } finally {
    isLoading.value = false
  }
}

function getOrganizationName(orgId: string): string {
  const org = organizations.value.find((o) => o.id === orgId)
  return org?.name || '-'
}

function getGroupName(groupId: string | null): string {
  if (!groupId) return '-'
  const group = allGroups.value.find((g) => g.id === groupId)
  return group?.name || '-'
}

function viewRelawan(r: Relawan) {
  router.push(`/relawan/${r.id}`)
}

function openCreateDialog() {
  formData.value = { name: '', phone: '', email: '', organizationId: '', groupId: '' }
  isCreateDialogOpen.value = true
}

function openEditDialog(r: Relawan) {
  selectedRelawan.value = r
  formData.value = {
    name: r.name,
    phone: r.phone || '',
    email: r.email || '',
    organizationId: r.organization_id,
    groupId: r.group_id || '',
  }
  isEditDialogOpen.value = true
}

function openDeleteDialog(r: Relawan) {
  selectedRelawan.value = r
  isDeleteDialogOpen.value = true
}

async function handleCreate() {
  if (!formData.value.name || !formData.value.organizationId) return

  isSaving.value = true
  try {
    await relawanService.create({
      organization_id: formData.value.organizationId,
      group_id: formData.value.groupId || undefined,
      name: formData.value.name,
      phone: formData.value.phone || undefined,
      email: formData.value.email || undefined,
    })
    toast.success('Relawan berhasil ditambahkan')
    isCreateDialogOpen.value = false
    await fetchRelawan()
  } catch (error) {
    console.error('Failed to create relawan:', error)
    toast.error('Gagal menambahkan relawan')
  } finally {
    isSaving.value = false
  }
}

async function handleEdit() {
  if (!selectedRelawan.value || !formData.value.name) return

  isSaving.value = true
  try {
    await relawanService.update(selectedRelawan.value.id, {
      name: formData.value.name,
      phone: formData.value.phone || undefined,
      email: formData.value.email || undefined,
      group_id: formData.value.groupId || null,
    })
    toast.success('Relawan berhasil diperbarui')
    isEditDialogOpen.value = false
    await fetchRelawan()
  } catch (error) {
    console.error('Failed to update relawan:', error)
    toast.error('Gagal memperbarui relawan')
  } finally {
    isSaving.value = false
  }
}

async function handleDelete() {
  if (!selectedRelawan.value) return

  isSaving.value = true
  try {
    await relawanService.delete(selectedRelawan.value.id)
    toast.success('Relawan berhasil dihapus')
    isDeleteDialogOpen.value = false
    await fetchRelawan()
  } catch (error) {
    console.error('Failed to delete relawan:', error)
    toast.error('Gagal menghapus relawan')
  } finally {
    isSaving.value = false
  }
}

function goToPage(page: number) {
  if (page < 1 || page > totalPages.value) return
  currentPage.value = page
  fetchRelawan()
}

const paginationInfo = computed(() => {
  if (totalItems.value === 0) return '0 data'
  const start = (currentPage.value - 1) * pageSize.value + 1
  const end = Math.min(currentPage.value * pageSize.value, totalItems.value)
  return `${start}-${end} dari ${totalItems.value}`
})

function getStatusLabel(status: string): string {
  switch (status) {
    case 'active': return 'Aktif'
    case 'inactive': return 'Nonaktif'
    case 'suspended': return 'Ditangguhkan'
    default: return status
  }
}

function getStatusVariant(status: string): 'default' | 'secondary' | 'destructive' {
  switch (status) {
    case 'active': return 'default'
    case 'inactive': return 'secondary'
    case 'suspended': return 'destructive'
    default: return 'secondary'
  }
}

onMounted(async () => {
  await Promise.all([
    fetchOrganizations(),
    fetchGroups(),
  ])
  await fetchRelawan()
})
</script>

<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Header -->
      <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h2 class="text-2xl font-bold tracking-tight">Relawan</h2>
          <p class="text-muted-foreground">
            Kelola data relawan yang terdaftar dalam sistem
          </p>
        </div>
        <Button @click="openCreateDialog">
          <Plus class="mr-2 h-4 w-4" />
          Tambah Relawan
        </Button>
      </div>

      <!-- Filters -->
      <div class="flex flex-col sm:flex-row items-start sm:items-center gap-4">
        <div class="relative flex-1 max-w-sm">
          <Search class="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            v-model="searchQuery"
            placeholder="Cari relawan..."
            class="pl-10"
          />
        </div>
        <Select v-model="selectedOrgFilter">
          <SelectTrigger class="w-[200px]">
            <SelectValue placeholder="Filter organisasi" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">Semua Organisasi</SelectItem>
            <SelectItem v-for="org in organizations" :key="org.id" :value="org.id">
              {{ org.name }}
            </SelectItem>
          </SelectContent>
        </Select>
        <Select v-model="selectedGroupFilter">
          <SelectTrigger class="w-[200px]">
            <SelectValue placeholder="Filter tim" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">Semua Tim</SelectItem>
            <SelectItem v-for="group in filterGroups" :key="group.id" :value="group.id">
              {{ group.name }}
            </SelectItem>
          </SelectContent>
        </Select>
      </div>

      <!-- Table -->
      <Card>
        <CardContent class="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Nama</TableHead>
                <TableHead>Telepon</TableHead>
                <TableHead>Tim</TableHead>
                <TableHead>Organisasi</TableHead>
                <TableHead class="text-center">Status</TableHead>
                <TableHead class="w-[70px]"></TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <template v-if="isLoading">
                <TableRow v-for="i in 5" :key="i">
                  <TableCell><Skeleton class="h-5 w-32" /></TableCell>
                  <TableCell><Skeleton class="h-5 w-28" /></TableCell>
                  <TableCell><Skeleton class="h-5 w-36" /></TableCell>
                  <TableCell><Skeleton class="h-5 w-28" /></TableCell>
                  <TableCell><Skeleton class="h-5 w-16 mx-auto" /></TableCell>
                  <TableCell><Skeleton class="h-8 w-8" /></TableCell>
                </TableRow>
              </template>
              <template v-else-if="relawanList.length === 0">
                <TableRow>
                  <TableCell colspan="6" class="h-32 text-center">
                    <div class="flex flex-col items-center gap-2">
                      <UserCircle class="h-8 w-8 text-muted-foreground" />
                      <p class="text-muted-foreground">Tidak ada relawan ditemukan</p>
                    </div>
                  </TableCell>
                </TableRow>
              </template>
              <template v-else>
                <TableRow
                  v-for="r in relawanList"
                  :key="r.id"
                  class="cursor-pointer"
                  @click="viewRelawan(r)"
                >
                  <TableCell class="font-medium">{{ r.name }}</TableCell>
                  <TableCell class="font-mono text-sm">{{ r.phone || '-' }}</TableCell>
                  <TableCell>
                    <div class="flex items-center gap-2">
                      <Users class="h-4 w-4 text-muted-foreground" />
                      <span class="text-sm">{{ r.group?.name || getGroupName(r.group_id) }}</span>
                    </div>
                  </TableCell>
                  <TableCell>
                    <div class="flex items-center gap-2">
                      <Building2 class="h-4 w-4 text-muted-foreground" />
                      <span class="text-sm text-muted-foreground">{{ r.organization?.name || getOrganizationName(r.organization_id) }}</span>
                    </div>
                  </TableCell>
                  <TableCell class="text-center">
                    <Badge :variant="getStatusVariant(r.status)">
                      {{ getStatusLabel(r.status) }}
                    </Badge>
                  </TableCell>
                  <TableCell @click.stop>
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="icon">
                          <MoreHorizontal class="h-4 w-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem @click="viewRelawan(r)">
                          <Eye class="mr-2 h-4 w-4" />
                          Lihat Detail
                        </DropdownMenuItem>
                        <DropdownMenuItem @click="openEditDialog(r)">
                          <Pencil class="mr-2 h-4 w-4" />
                          Edit
                        </DropdownMenuItem>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem
                          @click="openDeleteDialog(r)"
                          class="text-destructive focus:text-destructive"
                        >
                          <Trash2 class="mr-2 h-4 w-4" />
                          Hapus
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </TableCell>
                </TableRow>
              </template>
            </TableBody>
          </Table>
        </CardContent>

        <!-- Pagination -->
        <div v-if="!isLoading && totalPages > 1" class="flex items-center justify-between border-t px-4 py-3">
          <p class="text-sm text-muted-foreground">{{ paginationInfo }}</p>
          <div class="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              :disabled="currentPage === 1"
              @click="goToPage(currentPage - 1)"
            >
              <ChevronLeft class="h-4 w-4" />
            </Button>
            <span class="text-sm">
              Halaman {{ currentPage }} dari {{ totalPages }}
            </span>
            <Button
              variant="outline"
              size="sm"
              :disabled="currentPage === totalPages"
              @click="goToPage(currentPage + 1)"
            >
              <ChevronRight class="h-4 w-4" />
            </Button>
          </div>
        </div>
      </Card>

      <!-- Create Dialog -->
      <Dialog v-model:open="isCreateDialogOpen">
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Tambah Relawan</DialogTitle>
            <DialogDescription>
              Daftarkan relawan baru ke dalam sistem.
            </DialogDescription>
          </DialogHeader>
          <div class="space-y-4 py-4">
            <div class="space-y-2">
              <Label for="org">Organisasi</Label>
              <Select v-model="formData.organizationId">
                <SelectTrigger>
                  <SelectValue placeholder="Pilih organisasi" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem v-for="org in organizations" :key="org.id" :value="org.id">
                    {{ org.name }}
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div class="space-y-2">
              <Label for="group">Tim (opsional)</Label>
              <Select v-model="formData.groupId" :disabled="!formData.organizationId">
                <SelectTrigger>
                  <SelectValue placeholder="Pilih tim" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem v-for="group in availableGroups" :key="group.id" :value="group.id">
                    {{ group.name }}
                  </SelectItem>
                </SelectContent>
              </Select>
              <p v-if="!formData.organizationId" class="text-xs text-muted-foreground">Pilih organisasi terlebih dahulu</p>
            </div>
            <div class="space-y-2">
              <Label for="name">Nama Lengkap</Label>
              <Input
                id="name"
                v-model="formData.name"
                placeholder="Masukkan nama lengkap"
              />
            </div>
            <div class="space-y-2">
              <Label for="phone">Nomor Telepon</Label>
              <Input
                id="phone"
                v-model="formData.phone"
                placeholder="081234567890"
              />
            </div>
            <div class="space-y-2">
              <Label for="email">Email (opsional)</Label>
              <Input
                id="email"
                v-model="formData.email"
                type="email"
                placeholder="email@example.com"
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" @click="isCreateDialogOpen = false" :disabled="isSaving">
              Batal
            </Button>
            <Button
              @click="handleCreate"
              :disabled="!formData.name || !formData.organizationId || isSaving"
            >
              {{ isSaving ? 'Menyimpan...' : 'Simpan' }}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <!-- Edit Dialog -->
      <Dialog v-model:open="isEditDialogOpen">
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit Relawan</DialogTitle>
            <DialogDescription>
              Perbarui informasi relawan.
            </DialogDescription>
          </DialogHeader>
          <div class="space-y-4 py-4">
            <div class="space-y-2">
              <Label for="edit-org">Organisasi</Label>
              <Select v-model="formData.organizationId" disabled>
                <SelectTrigger>
                  <SelectValue placeholder="Pilih organisasi" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem v-for="org in organizations" :key="org.id" :value="org.id">
                    {{ org.name }}
                  </SelectItem>
                </SelectContent>
              </Select>
              <p class="text-xs text-muted-foreground">Organisasi tidak dapat diubah</p>
            </div>
            <div class="space-y-2">
              <Label for="edit-group">Tim</Label>
              <Select v-model="formData.groupId">
                <SelectTrigger>
                  <SelectValue placeholder="Pilih tim" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem v-for="group in availableGroups" :key="group.id" :value="group.id">
                    {{ group.name }}
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div class="space-y-2">
              <Label for="edit-name">Nama Lengkap</Label>
              <Input
                id="edit-name"
                v-model="formData.name"
                placeholder="Masukkan nama lengkap"
              />
            </div>
            <div class="space-y-2">
              <Label for="edit-phone">Nomor Telepon</Label>
              <Input
                id="edit-phone"
                v-model="formData.phone"
                placeholder="081234567890"
              />
            </div>
            <div class="space-y-2">
              <Label for="edit-email">Email</Label>
              <Input
                id="edit-email"
                v-model="formData.email"
                type="email"
                placeholder="email@example.com"
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" @click="isEditDialogOpen = false" :disabled="isSaving">
              Batal
            </Button>
            <Button
              @click="handleEdit"
              :disabled="!formData.name || isSaving"
            >
              {{ isSaving ? 'Menyimpan...' : 'Simpan' }}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <!-- Delete Dialog -->
      <Dialog v-model:open="isDeleteDialogOpen">
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Hapus Relawan</DialogTitle>
            <DialogDescription>
              Apakah Anda yakin ingin menghapus relawan "{{ selectedRelawan?.name }}"?
              Tindakan ini tidak dapat dibatalkan.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" @click="isDeleteDialogOpen = false" :disabled="isSaving">
              Batal
            </Button>
            <Button variant="destructive" @click="handleDelete" :disabled="isSaving">
              {{ isSaving ? 'Menghapus...' : 'Hapus' }}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  </AppLayout>
</template>
