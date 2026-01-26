<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { toast } from 'vue-sonner'
import {
  Search,
  MoreHorizontal,
  UserCog,
  Pencil,
  ChevronLeft,
  ChevronRight,
  Shield,
  ShieldCheck,
  User,
  Mail,
  Clock,
  Link2,
  FolderPlus,
  Trash2,
  Loader2,
  QrCode,
  Smartphone,
} from 'lucide-vue-next'
import QRCode from 'qrcode'
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
import { userService } from '@/services/users'
import { odkProjectService } from '@/services/odk'
import type { User as UserType, UserStatus, ODKProject, UserProjectAssignment, UserQRCodeResponse } from '@/services/types'

// State
const isLoading = ref(false)
const isSaving = ref(false)
const searchQuery = ref('')
const roleFilter = ref<string>('all')
const statusFilter = ref<string>('all')
const isEditDialogOpen = ref(false)
const selectedUser = ref<UserType | null>(null)

// Pagination
const currentPage = ref(1)
const pageSize = ref(20)
const totalItems = ref(0)
const totalPages = ref(0)

// Data
const userList = ref<UserType[]>([])

// Edit form state
const editForm = ref({
  name: '',
  role: '' as 'super_admin' | 'org_admin' | 'member',
  status: '' as UserStatus,
})

// ODK Project Assignment state
const isODKDialogOpen = ref(false)
const isLoadingODKData = ref(false)
const odkProjects = ref<ODKProject[]>([])
const userODKRoles = ref<UserProjectAssignment[]>([])
const assignForm = ref({
  odkProjectId: '',
  role: 'manager' as 'manager' | 'viewer',
})

// QR Code Dialog state
const isQRCodeDialogOpen = ref(false)
const isLoadingQRCode = ref(false)
const qrCodeData = ref<UserQRCodeResponse | null>(null)
const qrCodeImageUrl = ref<string>('')

// Debounce search
let searchTimeout: ReturnType<typeof setTimeout> | null = null

watch(searchQuery, () => {
  if (searchTimeout) clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    currentPage.value = 1
    fetchUsers()
  }, 300)
})

watch([roleFilter, statusFilter], () => {
  currentPage.value = 1
  fetchUsers()
})

async function fetchUsers() {
  isLoading.value = true
  try {
    const response = await userService.list({
      search: searchQuery.value || undefined,
      role: roleFilter.value !== 'all' ? roleFilter.value : undefined,
      status: statusFilter.value !== 'all' ? statusFilter.value : undefined,
      page: currentPage.value,
      page_size: pageSize.value,
    })

    userList.value = response.users
    totalItems.value = response.total
    totalPages.value = response.total_pages
  } catch (error) {
    console.error('Failed to fetch users:', error)
    toast.error('Gagal memuat data pengguna')
  } finally {
    isLoading.value = false
  }
}

function openEditDialog(user: UserType) {
  selectedUser.value = user
  editForm.value = {
    name: user.name || '',
    role: user.role,
    status: user.status,
  }
  isEditDialogOpen.value = true
}

async function handleEdit() {
  if (!selectedUser.value) return

  isSaving.value = true
  try {
    await userService.update(selectedUser.value.id, {
      name: editForm.value.name || undefined,
      role: editForm.value.role,
      status: editForm.value.status,
    })
    toast.success('Pengguna berhasil diperbarui')
    isEditDialogOpen.value = false
    await fetchUsers()
  } catch (error) {
    console.error('Failed to update user:', error)
    toast.error('Gagal memperbarui pengguna')
  } finally {
    isSaving.value = false
  }
}

function goToPage(page: number) {
  if (page < 1 || page > totalPages.value) return
  currentPage.value = page
  fetchUsers()
}

const paginationInfo = computed(() => {
  if (totalItems.value === 0) return '0 data'
  const start = (currentPage.value - 1) * pageSize.value + 1
  const end = Math.min(currentPage.value * pageSize.value, totalItems.value)
  return `${start}-${end} dari ${totalItems.value}`
})

function getRoleLabel(role: string): string {
  switch (role) {
    case 'super_admin': return 'Super Admin'
    case 'org_admin': return 'Admin Organisasi'
    case 'member': return 'Member'
    default: return role
  }
}

function getRoleIcon(role: string) {
  switch (role) {
    case 'super_admin': return ShieldCheck
    case 'org_admin': return Shield
    default: return User
  }
}

function getRoleVariant(role: string): 'default' | 'secondary' | 'outline' {
  switch (role) {
    case 'super_admin': return 'default'
    case 'org_admin': return 'secondary'
    default: return 'outline'
  }
}

function getStatusLabel(status: string): string {
  switch (status) {
    case 'active': return 'Aktif'
    case 'pending_invitation': return 'Menunggu Undangan'
    case 'suspended': return 'Ditangguhkan'
    default: return status
  }
}

function getStatusVariant(status: string): 'default' | 'secondary' | 'destructive' {
  switch (status) {
    case 'active': return 'default'
    case 'pending_invitation': return 'secondary'
    case 'suspended': return 'destructive'
    default: return 'secondary'
  }
}

function formatDate(dateString: string | null): string {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

// ODK Project Assignment functions
async function openODKDialog(user: UserType) {
  selectedUser.value = user
  isODKDialogOpen.value = true
  isLoadingODKData.value = true
  assignForm.value = { odkProjectId: '', role: 'manager' }
  odkProjects.value = []
  userODKRoles.value = []

  try {
    // Load ODK projects first
    const projectsRes = await odkProjectService.list()
    odkProjects.value = projectsRes
    console.log('ODK Projects loaded:', projectsRes)

    // Then load user's current assignments
    try {
      const rolesRes = await userService.getODKRoles(user.id)
      userODKRoles.value = rolesRes
      console.log('User ODK Roles loaded:', rolesRes)
    } catch (roleError) {
      console.error('Failed to load user ODK roles:', roleError)
      // User roles failed but projects are still available
      userODKRoles.value = []
    }
  } catch (error) {
    console.error('Failed to load ODK projects:', error)
    toast.error('Gagal memuat data ODK projects')
  } finally {
    isLoadingODKData.value = false
  }
}

async function handleAssignODKRole() {
  if (!selectedUser.value || !assignForm.value.odkProjectId) return

  isSaving.value = true
  try {
    await userService.assignODKRole(selectedUser.value.id, {
      odk_project_id: parseInt(assignForm.value.odkProjectId),
      role: assignForm.value.role,
    })
    toast.success('User berhasil ditambahkan ke project ODK')

    // Refresh user's ODK roles
    userODKRoles.value = await userService.getODKRoles(selectedUser.value.id)
    assignForm.value = { odkProjectId: '', role: 'manager' }

    // Refresh user list to update ODK link status
    await fetchUsers()
  } catch (error) {
    console.error('Failed to assign ODK role:', error)
    toast.error('Gagal menambahkan user ke project ODK')
  } finally {
    isSaving.value = false
  }
}

async function handleRemoveODKRole(assignment: UserProjectAssignment) {
  if (!selectedUser.value) return

  const roleType = assignment.role_id === 5 ? 'manager' : 'viewer'

  isSaving.value = true
  try {
    await userService.removeODKRole(selectedUser.value.id, assignment.project_id, roleType)
    toast.success('User berhasil dihapus dari project ODK')

    // Refresh user's ODK roles
    userODKRoles.value = await userService.getODKRoles(selectedUser.value.id)

    // Refresh user list
    await fetchUsers()
  } catch (error) {
    console.error('Failed to remove ODK role:', error)
    toast.error('Gagal menghapus user dari project ODK')
  } finally {
    isSaving.value = false
  }
}

// Projects that user is not yet assigned to (exclude archived projects)
const availableProjects = computed(() => {
  const assignedProjectIds = userODKRoles.value.map(r => r.project_id)
  return odkProjects.value.filter(p => !p.archived && !assignedProjectIds.includes(p.id))
})

// QR Code functions
async function openQRCodeDialog(user: UserType) {
  selectedUser.value = user
  isQRCodeDialogOpen.value = true
  isLoadingQRCode.value = true
  qrCodeData.value = null
  qrCodeImageUrl.value = ''

  try {
    const response = await userService.getQRCode(user.id)
    qrCodeData.value = response

    // Generate QR code image from the qr_code_data JSON
    qrCodeImageUrl.value = await QRCode.toDataURL(response.qr_code_data, {
      width: 300,
      margin: 2,
      color: {
        dark: '#000000',
        light: '#ffffff',
      },
    })
  } catch (error) {
    console.error('Failed to load QR code:', error)
    toast.error('Gagal memuat QR code. Pastikan user memiliki akses ODK Collect.')
    isQRCodeDialogOpen.value = false
  } finally {
    isLoadingQRCode.value = false
  }
}

function downloadQRCode() {
  if (!qrCodeImageUrl.value || !selectedUser.value) return

  const link = document.createElement('a')
  link.download = `odk-collect-${selectedUser.value.name || selectedUser.value.email}.png`
  link.href = qrCodeImageUrl.value
  link.click()
}

onMounted(() => {
  fetchUsers()
})
</script>

<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Header -->
      <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h2 class="text-2xl font-bold tracking-tight">Pengguna</h2>
          <p class="text-muted-foreground">
            Kelola semua pengguna portal admin (diimport dari ODK Central)
          </p>
        </div>
      </div>

      <!-- Filters -->
      <div class="flex flex-col sm:flex-row items-start sm:items-center gap-4">
        <div class="relative flex-1 max-w-sm">
          <Search class="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            v-model="searchQuery"
            placeholder="Cari nama atau email..."
            class="pl-10"
          />
        </div>
        <Select v-model="roleFilter">
          <SelectTrigger class="w-[180px]">
            <SelectValue placeholder="Filter role" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">Semua Role</SelectItem>
            <SelectItem value="super_admin">Super Admin</SelectItem>
            <SelectItem value="org_admin">Admin Organisasi</SelectItem>
            <SelectItem value="member">Member</SelectItem>
          </SelectContent>
        </Select>
        <Select v-model="statusFilter">
          <SelectTrigger class="w-[180px]">
            <SelectValue placeholder="Filter status" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">Semua Status</SelectItem>
            <SelectItem value="active">Aktif</SelectItem>
            <SelectItem value="pending_invitation">Menunggu Undangan</SelectItem>
            <SelectItem value="suspended">Ditangguhkan</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <!-- Table -->
      <Card>
        <CardContent class="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Pengguna</TableHead>
                <TableHead>Role</TableHead>
                <TableHead class="text-center">Status</TableHead>
                <TableHead>ODK Link</TableHead>
                <TableHead class="text-center">QR Code</TableHead>
                <TableHead>Login Terakhir</TableHead>
                <TableHead class="w-[70px]"></TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <template v-if="isLoading">
                <TableRow v-for="i in 5" :key="i">
                  <TableCell><Skeleton class="h-10 w-48" /></TableCell>
                  <TableCell><Skeleton class="h-5 w-24" /></TableCell>
                  <TableCell><Skeleton class="h-5 w-28 mx-auto" /></TableCell>
                  <TableCell><Skeleton class="h-5 w-16" /></TableCell>
                  <TableCell><Skeleton class="h-8 w-8 mx-auto" /></TableCell>
                  <TableCell><Skeleton class="h-5 w-32" /></TableCell>
                  <TableCell><Skeleton class="h-8 w-8" /></TableCell>
                </TableRow>
              </template>
              <template v-else-if="userList.length === 0">
                <TableRow>
                  <TableCell colspan="7" class="h-32 text-center">
                    <div class="flex flex-col items-center gap-2">
                      <UserCog class="h-8 w-8 text-muted-foreground" />
                      <p class="text-muted-foreground">Tidak ada pengguna ditemukan</p>
                    </div>
                  </TableCell>
                </TableRow>
              </template>
              <template v-else>
                <TableRow v-for="user in userList" :key="user.id">
                  <TableCell>
                    <div class="flex flex-col gap-0.5">
                      <span class="font-medium">{{ user.name || '-' }}</span>
                      <span class="text-sm text-muted-foreground flex items-center gap-1">
                        <Mail class="h-3 w-3" />
                        {{ user.email }}
                      </span>
                    </div>
                  </TableCell>
                  <TableCell>
                    <Badge :variant="getRoleVariant(user.role)" class="gap-1">
                      <component :is="getRoleIcon(user.role)" class="h-3 w-3" />
                      {{ getRoleLabel(user.role) }}
                    </Badge>
                  </TableCell>
                  <TableCell class="text-center">
                    <Badge :variant="getStatusVariant(user.status)">
                      {{ getStatusLabel(user.status) }}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    <div class="flex flex-col gap-1">
                      <span v-if="user.odk_web_user_id" class="flex items-center gap-1 text-sm text-green-600">
                        <Link2 class="h-3 w-3" />
                        Web: {{ user.odk_web_user_id }}
                      </span>
                      <span v-if="user.odk_app_user_id" class="flex items-center gap-1 text-sm text-blue-600">
                        <Smartphone class="h-3 w-3" />
                        App: {{ user.odk_app_user_id }}
                      </span>
                      <span v-if="!user.odk_web_user_id && !user.odk_app_user_id" class="text-sm text-muted-foreground">-</span>
                    </div>
                  </TableCell>
                  <TableCell class="text-center">
                    <Button
                      v-if="user.odk_app_user_id"
                      variant="ghost"
                      size="icon"
                      @click="openQRCodeDialog(user)"
                      title="Lihat QR Code"
                    >
                      <QrCode class="h-4 w-4 text-blue-600" />
                    </Button>
                    <span v-else class="text-sm text-muted-foreground">-</span>
                  </TableCell>
                  <TableCell>
                    <span v-if="user.last_login_at" class="flex items-center gap-1 text-sm text-muted-foreground">
                      <Clock class="h-3 w-3" />
                      {{ formatDate(user.last_login_at) }}
                    </span>
                    <span v-else class="text-sm text-muted-foreground">Belum pernah login</span>
                  </TableCell>
                  <TableCell>
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="icon">
                          <MoreHorizontal class="h-4 w-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem @click="openEditDialog(user)">
                          <Pencil class="mr-2 h-4 w-4" />
                          Edit Role/Status
                        </DropdownMenuItem>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem @click="openODKDialog(user)">
                          <FolderPlus class="mr-2 h-4 w-4" />
                          Kelola Akses ODK
                        </DropdownMenuItem>
                        <DropdownMenuItem
                          v-if="user.odk_app_user_id"
                          @click="openQRCodeDialog(user)"
                        >
                          <QrCode class="mr-2 h-4 w-4" />
                          Lihat QR Code
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

      <!-- Edit Dialog -->
      <Dialog v-model:open="isEditDialogOpen">
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit Pengguna</DialogTitle>
            <DialogDescription>
              Ubah role atau status pengguna "{{ selectedUser?.email }}".
            </DialogDescription>
          </DialogHeader>
          <div class="space-y-4 py-4">
            <div class="space-y-2">
              <Label for="edit-name">Nama</Label>
              <Input
                id="edit-name"
                v-model="editForm.name"
                placeholder="Nama pengguna"
              />
            </div>
            <div class="space-y-2">
              <Label for="edit-role">Role</Label>
              <Select v-model="editForm.role">
                <SelectTrigger>
                  <SelectValue placeholder="Pilih role" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="super_admin">Super Admin</SelectItem>
                  <SelectItem value="org_admin">Admin Organisasi</SelectItem>
                  <SelectItem value="member">Member</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div class="space-y-2">
              <Label for="edit-status">Status</Label>
              <Select v-model="editForm.status">
                <SelectTrigger>
                  <SelectValue placeholder="Pilih status" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="active">Aktif</SelectItem>
                  <SelectItem value="pending_invitation">Menunggu Undangan</SelectItem>
                  <SelectItem value="suspended">Ditangguhkan</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" @click="isEditDialogOpen = false" :disabled="isSaving">
              Batal
            </Button>
            <Button @click="handleEdit" :disabled="isSaving">
              {{ isSaving ? 'Menyimpan...' : 'Simpan' }}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <!-- ODK Project Assignment Dialog -->
      <Dialog v-model:open="isODKDialogOpen">
        <DialogContent class="max-w-lg">
          <DialogHeader>
            <DialogTitle>Kelola Akses ODK</DialogTitle>
            <DialogDescription>
              Kelola akses ODK Central untuk "{{ selectedUser?.name || selectedUser?.email }}".
            </DialogDescription>
          </DialogHeader>

          <div class="space-y-6 py-4">
            <!-- Loading State -->
            <div v-if="isLoadingODKData" class="flex items-center justify-center py-8">
              <Loader2 class="h-6 w-6 animate-spin text-muted-foreground" />
            </div>

            <template v-else>
              <!-- Current Assignments -->
              <div class="space-y-3">
                <Label class="text-sm font-medium">Project yang Diassign</Label>
                <div v-if="userODKRoles.length === 0" class="text-sm text-muted-foreground py-2">
                  Belum ada project yang diassign
                </div>
                <div v-else class="space-y-2">
                  <div
                    v-for="assignment in userODKRoles"
                    :key="assignment.project_id"
                    class="flex items-center justify-between p-3 border rounded-lg"
                  >
                    <div>
                      <p class="font-medium text-sm">{{ assignment.project_name }}</p>
                      <Badge variant="secondary" class="mt-1">
                        {{ assignment.role_name }}
                      </Badge>
                    </div>
                    <Button
                      variant="ghost"
                      size="icon"
                      class="text-destructive hover:text-destructive"
                      :disabled="isSaving"
                      @click="handleRemoveODKRole(assignment)"
                    >
                      <Trash2 class="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              </div>

              <!-- Add New Assignment -->
              <div class="space-y-3 pt-4 border-t">
                <Label class="text-sm font-medium">Tambah ke Project</Label>
                <div v-if="odkProjects.length === 0" class="text-sm text-muted-foreground py-2">
                  Tidak ada project ODK tersedia
                </div>
                <div v-else-if="availableProjects.length === 0" class="text-sm text-muted-foreground py-2">
                  {{ odkProjects.filter(p => !p.archived).length === 0
                    ? 'Semua project di-archive'
                    : 'User sudah diassign ke semua project aktif' }}
                </div>
                <template v-else>
                  <div class="flex gap-2">
                    <Select v-model="assignForm.odkProjectId" class="flex-1">
                      <SelectTrigger>
                        <SelectValue placeholder="Pilih project..." />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem
                          v-for="project in availableProjects"
                          :key="project.id"
                          :value="String(project.id)"
                        >
                          {{ project.name }}
                        </SelectItem>
                      </SelectContent>
                    </Select>
                    <Select v-model="assignForm.role" class="w-36">
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="manager">Manager</SelectItem>
                        <SelectItem value="viewer">Viewer</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  <Button
                    class="w-full"
                    :disabled="!assignForm.odkProjectId || isSaving"
                    @click="handleAssignODKRole"
                  >
                    <FolderPlus v-if="!isSaving" class="mr-2 h-4 w-4" />
                    <Loader2 v-else class="mr-2 h-4 w-4 animate-spin" />
                    {{ isSaving ? 'Menambahkan...' : 'Tambahkan ke Project' }}
                  </Button>
                </template>
              </div>
            </template>
          </div>

          <DialogFooter>
            <Button variant="outline" @click="isODKDialogOpen = false">
              Tutup
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <!-- QR Code Dialog -->
      <Dialog v-model:open="isQRCodeDialogOpen">
        <DialogContent class="max-w-md">
          <DialogHeader>
            <DialogTitle class="flex items-center gap-2">
              <QrCode class="h-5 w-5" />
              QR Code ODK Collect
            </DialogTitle>
            <DialogDescription>
              Scan QR code ini dengan aplikasi ODK Collect untuk login.
            </DialogDescription>
          </DialogHeader>

          <div class="py-4">
            <!-- Loading State -->
            <div v-if="isLoadingQRCode" class="flex items-center justify-center py-12">
              <Loader2 class="h-8 w-8 animate-spin text-muted-foreground" />
            </div>

            <!-- QR Code Content -->
            <div v-else-if="qrCodeData" class="space-y-4">
              <!-- User Info -->
              <div class="text-center space-y-1">
                <p class="font-medium">{{ qrCodeData.user_name }}</p>
                <p class="text-sm text-muted-foreground">{{ qrCodeData.user_email }}</p>
                <Badge variant="secondary" class="mt-2">
                  {{ qrCodeData.project_name }}
                </Badge>
              </div>

              <!-- QR Code Image -->
              <div class="flex justify-center p-4 bg-white rounded-lg border">
                <img
                  :src="qrCodeImageUrl"
                  :alt="`QR Code for ${qrCodeData.user_name}`"
                  class="w-64 h-64"
                />
              </div>

              <!-- Instructions -->
              <div class="text-sm text-muted-foreground text-center space-y-1">
                <p>1. Buka aplikasi ODK Collect di perangkat Android</p>
                <p>2. Pilih "Configure with QR code"</p>
                <p>3. Scan QR code di atas</p>
              </div>
            </div>
          </div>

          <DialogFooter class="flex gap-2">
            <Button variant="outline" @click="isQRCodeDialogOpen = false">
              Tutup
            </Button>
            <Button v-if="qrCodeImageUrl" @click="downloadQRCode">
              Download QR Code
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  </AppLayout>
</template>
