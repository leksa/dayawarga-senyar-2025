<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { toast } from 'vue-sonner'
import {
  Plus,
  Search,
  MoreHorizontal,
  Users,
  Eye,
  Pencil,
  Trash2,
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
import { Textarea } from '@/components/ui/textarea'
import { Skeleton } from '@/components/ui/skeleton'
import {
  groupService,
  organizationService,
  type Group,
  type Organization,
} from '@/services'

const router = useRouter()

// Extended type for display
interface GroupDisplay extends Group {
  relawanCount: number
}

// State
const isLoading = ref(false)
const isLoadingOrgs = ref(false)
const isSaving = ref(false)
const searchQuery = ref('')
const selectedOrgFilter = ref<string>('all')
const isCreateDialogOpen = ref(false)
const isEditDialogOpen = ref(false)
const isDeleteDialogOpen = ref(false)
const selectedGroup = ref<GroupDisplay | null>(null)

// Pagination
const currentPage = ref(1)
const pageSize = ref(20)
const totalItems = ref(0)
const totalPages = ref(0)

// Form state
const formData = ref({
  name: '',
  description: '',
  organizationId: '',
})

// Data
const organizations = ref<Organization[]>([])
const groups = ref<GroupDisplay[]>([])

// Debounce search
let searchTimeout: ReturnType<typeof setTimeout> | null = null

watch(searchQuery, () => {
  if (searchTimeout) clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    currentPage.value = 1
    fetchGroups()
  }, 300)
})

watch(selectedOrgFilter, () => {
  currentPage.value = 1
  fetchGroups()
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
  isLoading.value = true
  try {
    const response = await groupService.list({
      organization_id: selectedOrgFilter.value !== 'all' ? selectedOrgFilter.value : undefined,
      search: searchQuery.value || undefined,
      page: currentPage.value,
      page_size: pageSize.value,
    })

    // Fetch stats for each group
    const groupsWithStats = await Promise.all(
      response.groups.map(async (group) => {
        try {
          const stats = await groupService.getStats(group.id)
          return {
            ...group,
            relawanCount: stats.total_relawan,
          }
        } catch {
          return {
            ...group,
            relawanCount: 0,
          }
        }
      })
    )

    groups.value = groupsWithStats
    totalItems.value = response.total
    totalPages.value = response.total_pages
  } catch (error) {
    console.error('Failed to fetch groups:', error)
    toast.error('Gagal memuat data tim')
  } finally {
    isLoading.value = false
  }
}

function getOrganizationName(orgId: string): string {
  const org = organizations.value.find((o) => o.id === orgId)
  return org?.name || '-'
}

function viewGroup(group: GroupDisplay) {
  router.push(`/groups/${group.id}`)
}

function openCreateDialog() {
  formData.value = { name: '', description: '', organizationId: '' }
  isCreateDialogOpen.value = true
}

function openEditDialog(group: GroupDisplay) {
  selectedGroup.value = group
  formData.value = {
    name: group.name,
    description: group.description || '',
    organizationId: group.organization_id,
  }
  isEditDialogOpen.value = true
}

function openDeleteDialog(group: GroupDisplay) {
  selectedGroup.value = group
  isDeleteDialogOpen.value = true
}

async function handleCreate() {
  if (!formData.value.name || !formData.value.organizationId) return

  isSaving.value = true
  try {
    await groupService.create({
      organization_id: formData.value.organizationId,
      name: formData.value.name,
      description: formData.value.description || undefined,
    })
    toast.success('Tim berhasil dibuat')
    isCreateDialogOpen.value = false
    await fetchGroups()
  } catch (error) {
    console.error('Failed to create group:', error)
    toast.error('Gagal membuat tim')
  } finally {
    isSaving.value = false
  }
}

async function handleEdit() {
  if (!selectedGroup.value || !formData.value.name) return

  isSaving.value = true
  try {
    await groupService.update(selectedGroup.value.id, {
      name: formData.value.name,
      description: formData.value.description || undefined,
    })
    toast.success('Tim berhasil diperbarui')
    isEditDialogOpen.value = false
    await fetchGroups()
  } catch (error) {
    console.error('Failed to update group:', error)
    toast.error('Gagal memperbarui tim')
  } finally {
    isSaving.value = false
  }
}

async function handleDelete() {
  if (!selectedGroup.value) return

  isSaving.value = true
  try {
    await groupService.delete(selectedGroup.value.id)
    toast.success('Tim berhasil dihapus')
    isDeleteDialogOpen.value = false
    await fetchGroups()
  } catch (error) {
    console.error('Failed to delete group:', error)
    toast.error('Gagal menghapus tim')
  } finally {
    isSaving.value = false
  }
}

function goToPage(page: number) {
  if (page < 1 || page > totalPages.value) return
  currentPage.value = page
  fetchGroups()
}

const paginationInfo = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value + 1
  const end = Math.min(currentPage.value * pageSize.value, totalItems.value)
  return `${start}-${end} dari ${totalItems.value}`
})

onMounted(async () => {
  await fetchOrganizations()
  await fetchGroups()
})
</script>

<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Header -->
      <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h2 class="text-2xl font-bold tracking-tight">Tim/Grup</h2>
          <p class="text-muted-foreground">
            Kelola tim dan grup relawan dalam organisasi
          </p>
        </div>
        <Button @click="openCreateDialog">
          <Plus class="mr-2 h-4 w-4" />
          Tambah Tim
        </Button>
      </div>

      <!-- Filters -->
      <div class="flex flex-col sm:flex-row items-start sm:items-center gap-4">
        <div class="relative flex-1 max-w-sm">
          <Search class="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            v-model="searchQuery"
            placeholder="Cari tim..."
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
      </div>

      <!-- Table -->
      <Card>
        <CardContent class="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Nama Tim</TableHead>
                <TableHead>Organisasi</TableHead>
                <TableHead>Deskripsi</TableHead>
                <TableHead class="text-center">Relawan</TableHead>
                <TableHead class="text-center">Status</TableHead>
                <TableHead class="w-[70px]"></TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <template v-if="isLoading">
                <TableRow v-for="i in 5" :key="i">
                  <TableCell><Skeleton class="h-5 w-32" /></TableCell>
                  <TableCell><Skeleton class="h-5 w-24" /></TableCell>
                  <TableCell><Skeleton class="h-5 w-48" /></TableCell>
                  <TableCell><Skeleton class="h-5 w-8 mx-auto" /></TableCell>
                  <TableCell><Skeleton class="h-5 w-16 mx-auto" /></TableCell>
                  <TableCell><Skeleton class="h-8 w-8" /></TableCell>
                </TableRow>
              </template>
              <template v-else-if="groups.length === 0">
                <TableRow>
                  <TableCell colspan="6" class="h-32 text-center">
                    <div class="flex flex-col items-center gap-2">
                      <Users class="h-8 w-8 text-muted-foreground" />
                      <p class="text-muted-foreground">Tidak ada tim ditemukan</p>
                    </div>
                  </TableCell>
                </TableRow>
              </template>
              <template v-else>
                <TableRow
                  v-for="group in groups"
                  :key="group.id"
                  class="cursor-pointer"
                  @click="viewGroup(group)"
                >
                  <TableCell class="font-medium">{{ group.name }}</TableCell>
                  <TableCell>
                    <div class="flex items-center gap-2">
                      <Building2 class="h-4 w-4 text-muted-foreground" />
                      <span class="text-sm">{{ group.organization?.name || getOrganizationName(group.organization_id) }}</span>
                    </div>
                  </TableCell>
                  <TableCell class="text-muted-foreground max-w-xs truncate">
                    {{ group.description || '-' }}
                  </TableCell>
                  <TableCell class="text-center font-mono">{{ group.relawanCount }}</TableCell>
                  <TableCell class="text-center">
                    <Badge
                      :variant="group.is_active ? 'default' : 'secondary'"
                    >
                      {{ group.is_active ? 'Aktif' : 'Nonaktif' }}
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
                        <DropdownMenuItem @click="viewGroup(group)">
                          <Eye class="mr-2 h-4 w-4" />
                          Lihat Detail
                        </DropdownMenuItem>
                        <DropdownMenuItem @click="openEditDialog(group)">
                          <Pencil class="mr-2 h-4 w-4" />
                          Edit
                        </DropdownMenuItem>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem
                          @click="openDeleteDialog(group)"
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
            <DialogTitle>Tambah Tim</DialogTitle>
            <DialogDescription>
              Buat tim baru untuk mengelompokkan relawan.
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
              <Label for="name">Nama Tim</Label>
              <Input
                id="name"
                v-model="formData.name"
                placeholder="Masukkan nama tim"
              />
            </div>
            <div class="space-y-2">
              <Label for="description">Deskripsi</Label>
              <Textarea
                id="description"
                v-model="formData.description"
                placeholder="Masukkan deskripsi tim"
                rows="3"
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
            <DialogTitle>Edit Tim</DialogTitle>
            <DialogDescription>
              Perbarui informasi tim.
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
              <Label for="edit-name">Nama Tim</Label>
              <Input
                id="edit-name"
                v-model="formData.name"
                placeholder="Masukkan nama tim"
              />
            </div>
            <div class="space-y-2">
              <Label for="edit-description">Deskripsi</Label>
              <Textarea
                id="edit-description"
                v-model="formData.description"
                placeholder="Masukkan deskripsi tim"
                rows="3"
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
            <DialogTitle>Hapus Tim</DialogTitle>
            <DialogDescription>
              Apakah Anda yakin ingin menghapus tim "{{ selectedGroup?.name }}"?
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
