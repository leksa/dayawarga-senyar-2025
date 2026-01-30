<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { toast } from 'vue-sonner'
import { AxiosError } from 'axios'
import {
  Plus,
  Search,
  MoreHorizontal,
  Building2,
  Eye,
  Pencil,
  Trash2,
  ChevronLeft,
  ChevronRight,
  Mail,
  User,
  Copy,
  Check,
  FolderPlus,
  Link2,
  Loader2,
  CheckCircle2,
  XCircle,
  Globe,
  MapPin,
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
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Checkbox } from '@/components/ui/checkbox'
import { organizationService, type Organization, type Bidang } from '@/services'
import { odkProjectService, bidangService } from '@/services/odk'
import type { ODKProject, AssignODKProjectResult, AdminODKResult } from '@/services/types'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

// Extended type for list display
interface OrganizationDisplay extends Organization {
  groupCount: number
  relawanCount: number
}

// State
const isLoading = ref(false)
const isSaving = ref(false)
const searchQuery = ref('')
const isCreateDialogOpen = ref(false)
const isEditDialogOpen = ref(false)
const isDeleteDialogOpen = ref(false)
const isODKDialogOpen = ref(false)
const selectedOrg = ref<OrganizationDisplay | null>(null)

// ODK Project Assignment state
const isLoadingODKProjects = ref(false)
const isAssigningODK = ref(false)
const odkProjects = ref<ODKProject[]>([])
const selectedODKProjectId = ref<string>('')
const odkAssignmentResult = ref<AssignODKProjectResult | null>(null)
const odkLoadError = ref<string | null>(null)

// Pagination
const currentPage = ref(1)
const pageSize = ref(20)
const totalItems = ref(0)
const totalPages = ref(0)

// Bidang state
const allBidang = ref<Bidang[]>([])
const isLoadingBidang = ref(false)

// Form state
const formData = ref({
  name: '',
  description: '',
  admin_email: '',
  admin_name: '',
  city: '',
  country: '',
  website_url: '',
  social_media: {
    instagram: '',
    facebook: '',
    twitter: '',
  },
  selected_bidang_ids: [] as string[],
})

// Invitation result state
const invitationResult = ref<{
  invitation_link?: string
  admin_email?: string
  is_new_admin?: boolean
} | null>(null)
const isInvitationCopied = ref(false)

// Data
const organizations = ref<OrganizationDisplay[]>([])

// Debounce search
let searchTimeout: ReturnType<typeof setTimeout> | null = null

watch(searchQuery, () => {
  if (searchTimeout) clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    currentPage.value = 1
    fetchOrganizations()
  }, 300)
})

async function fetchOrganizations() {
  isLoading.value = true
  try {
    const response = await organizationService.list({
      search: searchQuery.value || undefined,
      page: currentPage.value,
      page_size: pageSize.value,
    })

    // Fetch stats for each organization
    const orgsWithStats = await Promise.all(
      response.organizations.map(async (org) => {
        try {
          const stats = await organizationService.getStats(org.id)
          return {
            ...org,
            groupCount: stats.total_groups,
            relawanCount: stats.total_relawan,
          }
        } catch {
          return {
            ...org,
            groupCount: 0,
            relawanCount: 0,
          }
        }
      })
    )

    organizations.value = orgsWithStats
    totalItems.value = response.total
    totalPages.value = response.total_pages
  } catch (error) {
    console.error('Failed to fetch organizations:', error)
    toast.error('Gagal memuat data organisasi')
  } finally {
    isLoading.value = false
  }
}

function viewOrganization(org: OrganizationDisplay) {
  router.push(`/organizations/${org.id}`)
}

async function fetchBidang() {
  isLoadingBidang.value = true
  try {
    allBidang.value = await bidangService.list()
  } catch (error) {
    console.error('Failed to fetch bidang:', error)
  } finally {
    isLoadingBidang.value = false
  }
}

function openCreateDialog() {
  formData.value = {
    name: '',
    description: '',
    admin_email: '',
    admin_name: '',
    city: '',
    country: '',
    website_url: '',
    social_media: { instagram: '', facebook: '', twitter: '' },
    selected_bidang_ids: [],
  }
  invitationResult.value = null
  isInvitationCopied.value = false
  isCreateDialogOpen.value = true
  fetchBidang()
}

function openEditDialog(org: OrganizationDisplay) {
  selectedOrg.value = org
  formData.value = {
    name: org.name,
    description: org.description || '',
    admin_email: '',
    admin_name: '',
    city: org.city || '',
    country: org.country || '',
    website_url: org.website_url || '',
    social_media: {
      instagram: org.social_media?.instagram || '',
      facebook: org.social_media?.facebook || '',
      twitter: org.social_media?.twitter || '',
    },
    selected_bidang_ids: org.bidang?.map(b => b.id) || [],
  }
  isEditDialogOpen.value = true
  fetchBidang()
}

function openDeleteDialog(org: OrganizationDisplay) {
  selectedOrg.value = org
  isDeleteDialogOpen.value = true
}

function getErrorMessage(error: unknown, fallback: string): string {
  if (error instanceof AxiosError) {
    return error.response?.data?.error || error.message || fallback
  }
  if (error instanceof Error) {
    return error.message
  }
  return fallback
}

async function openODKDialog(org: OrganizationDisplay) {
  selectedOrg.value = org
  selectedODKProjectId.value = ''
  odkAssignmentResult.value = null
  odkLoadError.value = null
  isODKDialogOpen.value = true

  await loadODKProjects()
}

async function loadODKProjects() {
  isLoadingODKProjects.value = true
  odkLoadError.value = null
  try {
    odkProjects.value = await odkProjectService.list()
  } catch (error: unknown) {
    console.error('Failed to load ODK projects:', error)
    odkLoadError.value = getErrorMessage(error, 'Gagal memuat daftar project ODK')
  } finally {
    isLoadingODKProjects.value = false
  }
}

// Available ODK projects (exclude archived)
const availableODKProjects = computed(() => {
  return odkProjects.value.filter(p => !p.archived)
})

async function handleAssignODKProject() {
  if (!selectedOrg.value || !selectedODKProjectId.value) return

  isAssigningODK.value = true
  try {
    const result = await organizationService.assignODKProject(selectedOrg.value.id, {
      odk_project_id: parseInt(selectedODKProjectId.value),
    })
    odkAssignmentResult.value = result
    toast.success('Organisasi berhasil di-assign ke project ODK')
    await fetchOrganizations()
  } catch (error: unknown) {
    console.error('Failed to assign ODK project:', error)
    toast.error(getErrorMessage(error, 'Gagal assign project ODK'))
  } finally {
    isAssigningODK.value = false
  }
}

function closeODKDialog() {
  isODKDialogOpen.value = false
  odkAssignmentResult.value = null
  selectedODKProjectId.value = ''
  odkLoadError.value = null
}

function getAdminStatusIcon(admin: AdminODKResult) {
  if (admin.error) return XCircle
  return CheckCircle2
}

function getAdminStatusColor(admin: AdminODKResult) {
  if (admin.error) return 'text-red-500'
  return 'text-green-500'
}

function buildSocialMedia() {
  const social: Record<string, string> = {}
  if (formData.value.social_media.instagram) social.instagram = formData.value.social_media.instagram
  if (formData.value.social_media.facebook) social.facebook = formData.value.social_media.facebook
  if (formData.value.social_media.twitter) social.twitter = formData.value.social_media.twitter
  return Object.keys(social).length > 0 ? social : undefined
}

async function handleCreate() {
  if (!formData.value.name) return

  isSaving.value = true
  try {
    const hasAdminEmail = formData.value.admin_email?.trim()
    const baseInput = {
      name: formData.value.name,
      description: formData.value.description || undefined,
      city: formData.value.city || undefined,
      country: formData.value.country || undefined,
      website_url: formData.value.website_url || undefined,
      social_media: buildSocialMedia(),
    }

    let createdOrgId: string | null = null

    if (hasAdminEmail) {
      const result = await organizationService.createWithAdmin({
        ...baseInput,
        admin_email: formData.value.admin_email.trim(),
        admin_name: formData.value.admin_name?.trim() || undefined,
      })
      createdOrgId = result.organization?.id || null

      if (result.invitation_link) {
        invitationResult.value = {
          invitation_link: result.invitation_link,
          admin_email: formData.value.admin_email,
          is_new_admin: result.is_new_admin,
        }
        toast.success('Organisasi berhasil dibuat dan undangan dikirim')
      } else {
        toast.success('Organisasi berhasil dibuat')
        isCreateDialogOpen.value = false
      }
    } else {
      const result = await organizationService.create(baseInput)
      createdOrgId = result.id
      toast.success('Organisasi berhasil dibuat')
      isCreateDialogOpen.value = false
    }

    if (createdOrgId && formData.value.selected_bidang_ids.length > 0) {
      for (const bidangId of formData.value.selected_bidang_ids) {
        await bidangService.addToOrganization(createdOrgId, bidangId)
      }
    }

    await fetchOrganizations()
  } catch (error) {
    console.error('Failed to create organization:', error)
    toast.error('Gagal membuat organisasi')
  } finally {
    isSaving.value = false
  }
}

async function copyInvitationLink() {
  if (!invitationResult.value?.invitation_link) return

  try {
    await navigator.clipboard.writeText(invitationResult.value.invitation_link)
    isInvitationCopied.value = true
    toast.success('Link undangan berhasil disalin')
    setTimeout(() => {
      isInvitationCopied.value = false
    }, 2000)
  } catch {
    toast.error('Gagal menyalin link')
  }
}

function closeCreateDialog() {
  isCreateDialogOpen.value = false
  invitationResult.value = null
}

async function handleEdit() {
  if (!selectedOrg.value || !formData.value.name) return

  isSaving.value = true
  try {
    await organizationService.update(selectedOrg.value.id, {
      name: formData.value.name,
      description: formData.value.description || undefined,
      city: formData.value.city || undefined,
      country: formData.value.country || undefined,
      website_url: formData.value.website_url || undefined,
      social_media: buildSocialMedia(),
    })

    const currentBidangIds = selectedOrg.value.bidang?.map(b => b.id) || []
    const newBidangIds = formData.value.selected_bidang_ids

    const toAdd = newBidangIds.filter(id => !currentBidangIds.includes(id))
    const toRemove = currentBidangIds.filter(id => !newBidangIds.includes(id))

    for (const bidangId of toAdd) {
      await bidangService.addToOrganization(selectedOrg.value.id, bidangId)
    }
    for (const bidangId of toRemove) {
      await bidangService.removeFromOrganization(selectedOrg.value.id, bidangId)
    }

    toast.success('Organisasi berhasil diperbarui')
    isEditDialogOpen.value = false
    await fetchOrganizations()
  } catch (error) {
    console.error('Failed to update organization:', error)
    toast.error('Gagal memperbarui organisasi')
  } finally {
    isSaving.value = false
  }
}

async function handleDelete() {
  if (!selectedOrg.value) return

  isSaving.value = true
  try {
    await organizationService.delete(selectedOrg.value.id)
    toast.success('Organisasi berhasil dihapus')
    isDeleteDialogOpen.value = false
    await fetchOrganizations()
  } catch (error) {
    console.error('Failed to delete organization:', error)
    toast.error('Gagal menghapus organisasi')
  } finally {
    isSaving.value = false
  }
}

function goToPage(page: number) {
  if (page < 1 || page > totalPages.value) return
  currentPage.value = page
  fetchOrganizations()
}

const paginationInfo = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value + 1
  const end = Math.min(currentPage.value * pageSize.value, totalItems.value)
  return `${start}-${end} dari ${totalItems.value}`
})

onMounted(() => {
  fetchOrganizations()
})
</script>

<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Header -->
      <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h2 class="text-2xl font-bold tracking-tight">Organisasi</h2>
          <p class="text-muted-foreground">
            {{ authStore.isSuperAdmin ? 'Kelola organisasi yang terdaftar dalam sistem' : 'Kelola organisasi Anda' }}
          </p>
        </div>
        <Button v-if="authStore.canManageOrganizations" @click="openCreateDialog">
          <Plus class="mr-2 h-4 w-4" />
          Tambah Organisasi
        </Button>
      </div>

      <!-- Search -->
      <div class="flex items-center gap-4">
        <div class="relative flex-1 max-w-sm">
          <Search class="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            v-model="searchQuery"
            placeholder="Cari organisasi..."
            class="pl-10"
          />
        </div>
      </div>

      <!-- Table -->
      <Card>
        <CardContent class="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Nama</TableHead>
                <TableHead>Deskripsi</TableHead>
                <TableHead class="text-center">Tim</TableHead>
                <TableHead class="text-center">Relawan</TableHead>
                <TableHead class="text-center">ODK Project</TableHead>
                <TableHead class="text-center">Status</TableHead>
                <TableHead class="w-[70px]"></TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <template v-if="isLoading">
                <TableRow v-for="i in 5" :key="i">
                  <TableCell><Skeleton class="h-5 w-32" /></TableCell>
                  <TableCell><Skeleton class="h-5 w-48" /></TableCell>
                  <TableCell><Skeleton class="h-5 w-8 mx-auto" /></TableCell>
                  <TableCell><Skeleton class="h-5 w-8 mx-auto" /></TableCell>
                  <TableCell><Skeleton class="h-5 w-16 mx-auto" /></TableCell>
                  <TableCell><Skeleton class="h-5 w-16 mx-auto" /></TableCell>
                  <TableCell><Skeleton class="h-8 w-8" /></TableCell>
                </TableRow>
              </template>
              <template v-else-if="organizations.length === 0">
                <TableRow>
                  <TableCell colspan="7" class="h-32 text-center">
                    <div class="flex flex-col items-center gap-2">
                      <Building2 class="h-8 w-8 text-muted-foreground" />
                      <p class="text-muted-foreground">Tidak ada organisasi ditemukan</p>
                    </div>
                  </TableCell>
                </TableRow>
              </template>
              <template v-else>
                <TableRow
                  v-for="org in organizations"
                  :key="org.id"
                  class="cursor-pointer"
                  @click="viewOrganization(org)"
                >
                  <TableCell class="font-medium">{{ org.name }}</TableCell>
                  <TableCell class="text-muted-foreground max-w-xs truncate">
                    {{ org.description || '-' }}
                  </TableCell>
                  <TableCell class="text-center font-mono">{{ org.groupCount }}</TableCell>
                  <TableCell class="text-center font-mono">{{ org.relawanCount }}</TableCell>
                  <TableCell class="text-center">
                    <Badge v-if="org.odk_project_id" variant="outline" class="gap-1">
                      <Link2 class="h-3 w-3" />
                      {{ org.odk_project_id }}
                    </Badge>
                    <span v-else class="text-muted-foreground text-sm">-</span>
                  </TableCell>
                  <TableCell class="text-center">
                    <Badge
                      :variant="org.is_active ? 'default' : 'secondary'"
                    >
                      {{ org.is_active ? 'Aktif' : 'Nonaktif' }}
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
                        <DropdownMenuItem @click="viewOrganization(org)">
                          <Eye class="mr-2 h-4 w-4" />
                          Lihat Detail
                        </DropdownMenuItem>
                        <DropdownMenuItem @click="openEditDialog(org)">
                          <Pencil class="mr-2 h-4 w-4" />
                          Edit
                        </DropdownMenuItem>
                        <template v-if="authStore.canManageOrganizations">
                          <DropdownMenuSeparator />
                          <DropdownMenuItem
                            v-if="!org.odk_project_id"
                            @click="openODKDialog(org)"
                          >
                            <FolderPlus class="mr-2 h-4 w-4" />
                            Assign ODK Project
                          </DropdownMenuItem>
                          <DropdownMenuItem
                            v-else
                            disabled
                            class="text-muted-foreground"
                          >
                            <Link2 class="mr-2 h-4 w-4" />
                            Sudah ter-assign ke ODK {{ org.odk_project_id }}
                          </DropdownMenuItem>
                        </template>
                        <template v-if="authStore.canManageOrganizations">
                          <DropdownMenuSeparator />
                          <DropdownMenuItem
                            @click="openDeleteDialog(org)"
                            class="text-destructive focus:text-destructive"
                          >
                            <Trash2 class="mr-2 h-4 w-4" />
                            Hapus
                          </DropdownMenuItem>
                        </template>
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
        <DialogContent class="sm:max-w-[500px]">
          <DialogHeader>
            <DialogTitle>Tambah Organisasi</DialogTitle>
            <DialogDescription>
              Buat organisasi baru untuk mengelompokkan tim dan relawan.
            </DialogDescription>
          </DialogHeader>

          <!-- Invitation Result View -->
          <div v-if="invitationResult" class="space-y-4 py-4">
            <div class="rounded-lg border border-green-200 bg-green-50 dark:border-green-900 dark:bg-green-950 p-4">
              <div class="flex items-start gap-3">
                <div class="rounded-full bg-green-100 dark:bg-green-900 p-2">
                  <Check class="h-4 w-4 text-green-600 dark:text-green-400" />
                </div>
                <div class="flex-1 space-y-1">
                  <p class="font-medium text-green-900 dark:text-green-100">
                    Organisasi berhasil dibuat!
                  </p>
                  <p class="text-sm text-green-700 dark:text-green-300">
                    Undangan telah dikirim ke <strong>{{ invitationResult.admin_email }}</strong>
                  </p>
                </div>
              </div>
            </div>

            <div class="space-y-2">
              <Label>Link Undangan</Label>
              <p class="text-xs text-muted-foreground">
                Salin link ini jika admin tidak menerima email undangan.
              </p>
              <div class="flex gap-2">
                <Input
                  :value="invitationResult.invitation_link"
                  readonly
                  class="font-mono text-xs"
                />
                <Button
                  variant="outline"
                  size="icon"
                  @click="copyInvitationLink"
                >
                  <Check v-if="isInvitationCopied" class="h-4 w-4 text-green-600" />
                  <Copy v-else class="h-4 w-4" />
                </Button>
              </div>
            </div>
          </div>

          <!-- Form View -->
          <div v-else class="space-y-4 py-4">
            <div class="space-y-2">
              <Label for="name">Nama Organisasi <span class="text-destructive">*</span></Label>
              <Input
                id="name"
                v-model="formData.name"
                placeholder="Masukkan nama organisasi"
              />
            </div>
            <div class="space-y-2">
              <Label for="description">Deskripsi</Label>
              <Textarea
                id="description"
                v-model="formData.description"
                placeholder="Masukkan deskripsi organisasi"
                rows="3"
              />
            </div>

            <!-- Location Section -->
            <div class="border-t pt-4 mt-4">
              <div class="flex items-center gap-2 mb-3">
                <MapPin class="h-4 w-4 text-muted-foreground" />
                <span class="text-sm font-medium">Lokasi</span>
              </div>
              <div class="grid grid-cols-2 gap-3">
                <div class="space-y-2">
                  <Label for="city">Kota</Label>
                  <Input
                    id="city"
                    v-model="formData.city"
                    placeholder="Jakarta"
                  />
                </div>
                <div class="space-y-2">
                  <Label for="country">Negara</Label>
                  <Input
                    id="country"
                    v-model="formData.country"
                    placeholder="Indonesia"
                  />
                </div>
              </div>
            </div>

            <!-- Web & Social Media Section -->
            <div class="border-t pt-4 mt-4">
              <div class="flex items-center gap-2 mb-3">
                <Globe class="h-4 w-4 text-muted-foreground" />
                <span class="text-sm font-medium">Web & Media Sosial</span>
              </div>
              <div class="space-y-3">
                <div class="space-y-2">
                  <Label for="website_url">Website</Label>
                  <Input
                    id="website_url"
                    v-model="formData.website_url"
                    placeholder="https://organisasi.com"
                  />
                </div>
                <div class="grid grid-cols-3 gap-3">
                  <div class="space-y-2">
                    <Label for="instagram">Instagram</Label>
                    <Input
                      id="instagram"
                      v-model="formData.social_media.instagram"
                      placeholder="@username"
                    />
                  </div>
                  <div class="space-y-2">
                    <Label for="facebook">Facebook</Label>
                    <Input
                      id="facebook"
                      v-model="formData.social_media.facebook"
                      placeholder="pagename"
                    />
                  </div>
                  <div class="space-y-2">
                    <Label for="twitter">Twitter/X</Label>
                    <Input
                      id="twitter"
                      v-model="formData.social_media.twitter"
                      placeholder="@username"
                    />
                  </div>
                </div>
              </div>
            </div>

            <!-- Bidang Section -->
            <div class="border-t pt-4 mt-4">
              <div class="flex items-center gap-2 mb-3">
                <Building2 class="h-4 w-4 text-muted-foreground" />
                <span class="text-sm font-medium">Bidang / Klaster</span>
              </div>
              <div v-if="isLoadingBidang" class="flex items-center gap-2">
                <Loader2 class="h-4 w-4 animate-spin" />
                <span class="text-sm text-muted-foreground">Memuat bidang...</span>
              </div>
              <div v-else class="grid grid-cols-2 gap-2">
                <div
                  v-for="bidang in allBidang"
                  :key="bidang.id"
                  class="flex items-center space-x-2"
                >
                  <Checkbox
                    :id="`bidang-${bidang.id}`"
                    :checked="formData.selected_bidang_ids.includes(bidang.id)"
                    @update:checked="(checked: boolean) => {
                      if (checked) {
                        formData.selected_bidang_ids.push(bidang.id)
                      } else {
                        formData.selected_bidang_ids = formData.selected_bidang_ids.filter(id => id !== bidang.id)
                      }
                    }"
                  />
                  <Label :for="`bidang-${bidang.id}`" class="text-sm font-normal cursor-pointer">
                    {{ bidang.name }}
                  </Label>
                </div>
              </div>
            </div>

            <!-- Admin Invitation Section -->
            <div class="border-t pt-4 mt-4">
              <div class="flex items-center gap-2 mb-3">
                <User class="h-4 w-4 text-muted-foreground" />
                <span class="text-sm font-medium">Undang Admin Organisasi</span>
                <span class="text-xs text-muted-foreground">(opsional)</span>
              </div>
              <div class="space-y-3">
                <div class="space-y-2">
                  <Label for="admin_email">
                    <Mail class="h-3 w-3 inline mr-1" />
                    Email Admin
                  </Label>
                  <Input
                    id="admin_email"
                    v-model="formData.admin_email"
                    type="email"
                    placeholder="admin@organisasi.com"
                  />
                </div>
                <div class="space-y-2">
                  <Label for="admin_name">Nama Admin</Label>
                  <Input
                    id="admin_name"
                    v-model="formData.admin_name"
                    placeholder="Nama lengkap admin"
                  />
                </div>
                <p class="text-xs text-muted-foreground">
                  Admin akan menerima email undangan untuk mengatur password dan mengakses organisasi ini.
                </p>
              </div>
            </div>
          </div>

          <DialogFooter>
            <template v-if="invitationResult">
              <Button @click="closeCreateDialog">
                Selesai
              </Button>
            </template>
            <template v-else>
              <Button variant="outline" @click="closeCreateDialog" :disabled="isSaving">
                Batal
              </Button>
              <Button @click="handleCreate" :disabled="!formData.name || isSaving">
                {{ isSaving ? 'Menyimpan...' : 'Simpan' }}
              </Button>
            </template>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <!-- Edit Dialog -->
      <Dialog v-model:open="isEditDialogOpen">
        <DialogContent class="sm:max-w-[500px] max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Edit Organisasi</DialogTitle>
            <DialogDescription>
              Perbarui informasi organisasi.
            </DialogDescription>
          </DialogHeader>
          <div class="space-y-4 py-4">
            <div class="space-y-2">
              <Label for="edit-name">Nama Organisasi</Label>
              <Input
                id="edit-name"
                v-model="formData.name"
                placeholder="Masukkan nama organisasi"
              />
            </div>
            <div class="space-y-2">
              <Label for="edit-description">Deskripsi</Label>
              <Textarea
                id="edit-description"
                v-model="formData.description"
                placeholder="Masukkan deskripsi organisasi"
                rows="3"
              />
            </div>

            <!-- Location Section -->
            <div class="border-t pt-4 mt-4">
              <div class="flex items-center gap-2 mb-3">
                <MapPin class="h-4 w-4 text-muted-foreground" />
                <span class="text-sm font-medium">Lokasi</span>
              </div>
              <div class="grid grid-cols-2 gap-3">
                <div class="space-y-2">
                  <Label for="edit-city">Kota</Label>
                  <Input
                    id="edit-city"
                    v-model="formData.city"
                    placeholder="Jakarta"
                  />
                </div>
                <div class="space-y-2">
                  <Label for="edit-country">Negara</Label>
                  <Input
                    id="edit-country"
                    v-model="formData.country"
                    placeholder="Indonesia"
                  />
                </div>
              </div>
            </div>

            <!-- Web & Social Media Section -->
            <div class="border-t pt-4 mt-4">
              <div class="flex items-center gap-2 mb-3">
                <Globe class="h-4 w-4 text-muted-foreground" />
                <span class="text-sm font-medium">Web & Media Sosial</span>
              </div>
              <div class="space-y-3">
                <div class="space-y-2">
                  <Label for="edit-website_url">Website</Label>
                  <Input
                    id="edit-website_url"
                    v-model="formData.website_url"
                    placeholder="https://organisasi.com"
                  />
                </div>
                <div class="grid grid-cols-3 gap-3">
                  <div class="space-y-2">
                    <Label for="edit-instagram">Instagram</Label>
                    <Input
                      id="edit-instagram"
                      v-model="formData.social_media.instagram"
                      placeholder="@username"
                    />
                  </div>
                  <div class="space-y-2">
                    <Label for="edit-facebook">Facebook</Label>
                    <Input
                      id="edit-facebook"
                      v-model="formData.social_media.facebook"
                      placeholder="pagename"
                    />
                  </div>
                  <div class="space-y-2">
                    <Label for="edit-twitter">Twitter/X</Label>
                    <Input
                      id="edit-twitter"
                      v-model="formData.social_media.twitter"
                      placeholder="@username"
                    />
                  </div>
                </div>
              </div>
            </div>

            <!-- Bidang Section -->
            <div class="border-t pt-4 mt-4">
              <div class="flex items-center gap-2 mb-3">
                <Building2 class="h-4 w-4 text-muted-foreground" />
                <span class="text-sm font-medium">Bidang / Klaster</span>
              </div>
              <div v-if="isLoadingBidang" class="flex items-center gap-2">
                <Loader2 class="h-4 w-4 animate-spin" />
                <span class="text-sm text-muted-foreground">Memuat bidang...</span>
              </div>
              <div v-else class="grid grid-cols-2 gap-2">
                <div
                  v-for="bidang in allBidang"
                  :key="bidang.id"
                  class="flex items-center space-x-2"
                >
                  <Checkbox
                    :id="`edit-bidang-${bidang.id}`"
                    :checked="formData.selected_bidang_ids.includes(bidang.id)"
                    @update:checked="(checked: boolean) => {
                      if (checked) {
                        formData.selected_bidang_ids.push(bidang.id)
                      } else {
                        formData.selected_bidang_ids = formData.selected_bidang_ids.filter(id => id !== bidang.id)
                      }
                    }"
                  />
                  <Label :for="`edit-bidang-${bidang.id}`" class="text-sm font-normal cursor-pointer">
                    {{ bidang.name }}
                  </Label>
                </div>
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" @click="isEditDialogOpen = false" :disabled="isSaving">
              Batal
            </Button>
            <Button @click="handleEdit" :disabled="!formData.name || isSaving">
              {{ isSaving ? 'Menyimpan...' : 'Simpan' }}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <!-- Delete Dialog -->
      <Dialog v-model:open="isDeleteDialogOpen">
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Hapus Organisasi</DialogTitle>
            <DialogDescription>
              Apakah Anda yakin ingin menghapus organisasi "{{ selectedOrg?.name }}"?
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

      <!-- ODK Project Assignment Dialog -->
      <Dialog v-model:open="isODKDialogOpen">
        <DialogContent class="sm:max-w-[600px]">
          <DialogHeader>
            <DialogTitle class="flex items-center gap-2">
              <FolderPlus class="h-5 w-5" />
              Assign ODK Project
            </DialogTitle>
            <DialogDescription>
              Assign organisasi "{{ selectedOrg?.name }}" ke project ODK Central.
              Admin organisasi akan otomatis didaftarkan sebagai Project Manager.
            </DialogDescription>
          </DialogHeader>

          <!-- Loading State -->
          <div v-if="isLoadingODKProjects" class="flex items-center justify-center py-8">
            <Loader2 class="h-8 w-8 animate-spin text-muted-foreground" />
          </div>

          <!-- Error State -->
          <div v-else-if="odkLoadError" class="space-y-4 py-4">
            <div class="rounded-lg border border-red-200 bg-red-50 dark:border-red-900 dark:bg-red-950 p-4">
              <div class="flex items-start gap-3">
                <XCircle class="h-5 w-5 text-red-500 mt-0.5" />
                <div class="flex-1">
                  <p class="font-medium text-red-900 dark:text-red-100">
                    Gagal memuat daftar project
                  </p>
                  <p class="text-sm text-red-700 dark:text-red-300 mt-1">
                    {{ odkLoadError }}
                  </p>
                </div>
              </div>
            </div>
            <Button variant="outline" @click="loadODKProjects">
              Coba lagi
            </Button>
          </div>

          <!-- Result View -->
          <div v-else-if="odkAssignmentResult" class="space-y-4 py-4">
            <div class="rounded-lg border border-green-200 bg-green-50 dark:border-green-900 dark:bg-green-950 p-4">
              <div class="flex items-start gap-3">
                <div class="rounded-full bg-green-100 dark:bg-green-900 p-2">
                  <CheckCircle2 class="h-5 w-5 text-green-600 dark:text-green-400" />
                </div>
                <div class="flex-1 space-y-1">
                  <p class="font-medium text-green-900 dark:text-green-100">
                    Berhasil di-assign ke ODK Project!
                  </p>
                  <p class="text-sm text-green-700 dark:text-green-300">
                    Project: <strong>{{ odkAssignmentResult.odk_project_name }}</strong>
                  </p>
                </div>
              </div>
            </div>

            <!-- Admin Results -->
            <div class="space-y-2">
              <Label>Status Admin Organisasi:</Label>
              <div class="space-y-2">
                <div
                  v-for="admin in odkAssignmentResult.admins_processed"
                  :key="admin.user_id"
                  class="flex items-center justify-between p-3 rounded-lg border bg-muted/30"
                >
                  <div class="flex items-center gap-3">
                    <component
                      :is="getAdminStatusIcon(admin)"
                      :class="['h-5 w-5', getAdminStatusColor(admin)]"
                    />
                    <div>
                      <p class="font-medium">{{ admin.user_name }}</p>
                      <p class="text-xs text-muted-foreground">{{ admin.user_email }}</p>
                    </div>
                  </div>
                  <div class="text-right text-sm">
                    <template v-if="admin.error">
                      <span class="text-red-500">{{ admin.error }}</span>
                    </template>
                    <template v-else>
                      <div class="flex flex-col gap-0.5">
                        <span class="text-green-600">Web User: {{ admin.odk_web_user_id }}</span>
                        <span class="text-blue-600">App User: {{ admin.odk_app_user_id }}</span>
                        <Badge v-if="admin.has_qr_code" variant="outline" class="text-xs">
                          QR Code tersedia
                        </Badge>
                      </div>
                    </template>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Form View -->
          <div v-else class="space-y-4 py-4">
            <!-- No projects available -->
            <div v-if="availableODKProjects.length === 0" class="rounded-lg border border-yellow-200 bg-yellow-50 dark:border-yellow-900 dark:bg-yellow-950 p-4">
              <p class="text-sm text-yellow-800 dark:text-yellow-200">
                Tidak ada project ODK yang tersedia. Pastikan ODK Central memiliki project yang aktif.
              </p>
              <Button variant="outline" size="sm" class="mt-2" @click="loadODKProjects">
                Muat ulang
              </Button>
            </div>

            <!-- Project Selection -->
            <div v-else class="space-y-2">
              <Label for="odk-project">Pilih ODK Project</Label>
              <Select v-model="selectedODKProjectId">
                <SelectTrigger>
                  <SelectValue placeholder="Pilih project..." />
                </SelectTrigger>
                <SelectContent class="z-[200]">
                  <SelectItem
                    v-for="project in availableODKProjects"
                    :key="project.id"
                    :value="String(project.id)"
                  >
                    {{ project.name }} (ID: {{ project.id }})
                  </SelectItem>
                </SelectContent>
              </Select>
              <p class="text-xs text-muted-foreground">
                Hanya menampilkan project yang tidak di-archive.
              </p>
            </div>

            <!-- Info Box -->
            <div class="rounded-lg border bg-blue-50 dark:bg-blue-950 p-4 text-sm">
              <p class="font-medium text-blue-900 dark:text-blue-100 mb-2">
                Proses yang akan dilakukan:
              </p>
              <ul class="list-disc list-inside text-blue-700 dark:text-blue-300 space-y-1">
                <li>Mendaftarkan admin organisasi ke ODK Central (langsung aktif)</li>
                <li>Memberikan role Project Manager</li>
                <li>Membuat App User untuk ODK Collect</li>
                <li>Generate QR Code untuk login ODK Collect</li>
              </ul>
            </div>
          </div>

          <DialogFooter>
            <template v-if="odkAssignmentResult">
              <Button @click="closeODKDialog">
                Selesai
              </Button>
            </template>
            <template v-else-if="odkLoadError">
              <Button variant="outline" @click="closeODKDialog">
                Tutup
              </Button>
            </template>
            <template v-else>
              <Button variant="outline" @click="closeODKDialog" :disabled="isAssigningODK">
                Batal
              </Button>
              <Button
                @click="handleAssignODKProject"
                :disabled="!selectedODKProjectId || isAssigningODK"
              >
                <Loader2 v-if="isAssigningODK" class="mr-2 h-4 w-4 animate-spin" />
                {{ isAssigningODK ? 'Memproses...' : 'Assign Project' }}
              </Button>
            </template>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  </AppLayout>
</template>
