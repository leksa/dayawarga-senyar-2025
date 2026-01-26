<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { toast } from 'vue-sonner'
import {
  Clock,
  CheckCircle2,
  XCircle,
  ChevronLeft,
  ChevronRight,
  Check,
  X,
  Building2,
  Users,
  FolderOpen,
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
  projectRequestService,
  organizationService,
  type ProjectRequest,
  type Organization,
  type ProjectRequestStatus,
} from '@/services'

// State
const isLoading = ref(false)
const isLoadingOrgs = ref(false)
const isProcessing = ref(false)
const statusFilter = ref<ProjectRequestStatus | 'all'>('all')
const selectedOrgFilter = ref<string>('all')
const isReviewDialogOpen = ref(false)
const selectedRequest = ref<ProjectRequest | null>(null)
const reviewAction = ref<'approve' | 'reject'>('approve')
const reviewNotes = ref('')

// Pagination
const currentPage = ref(1)
const pageSize = ref(20)
const totalItems = ref(0)
const totalPages = ref(0)

// Data
const organizations = ref<Organization[]>([])
const requests = ref<ProjectRequest[]>([])

// Stats
const stats = computed(() => {
  const pending = requests.value.filter(r => r.status === 'pending').length
  const approved = requests.value.filter(r => r.status === 'approved').length
  const rejected = requests.value.filter(r => r.status === 'rejected').length
  return { pending, approved, rejected, total: requests.value.length }
})

watch([statusFilter, selectedOrgFilter], () => {
  currentPage.value = 1
  fetchRequests()
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

async function fetchRequests() {
  isLoading.value = true
  try {
    const response = await projectRequestService.list({
      status: statusFilter.value !== 'all' ? statusFilter.value : undefined,
      organization_id: selectedOrgFilter.value !== 'all' ? selectedOrgFilter.value : undefined,
      page: currentPage.value,
      page_size: pageSize.value,
    })
    requests.value = response.data
    totalItems.value = response.total
    totalPages.value = response.total_pages
  } catch (error) {
    console.error('Failed to fetch requests:', error)
    toast.error('Gagal memuat data permintaan')
  } finally {
    isLoading.value = false
  }
}

function openReviewDialog(request: ProjectRequest, action: 'approve' | 'reject') {
  selectedRequest.value = request
  reviewAction.value = action
  reviewNotes.value = ''
  isReviewDialogOpen.value = true
}

async function submitReview() {
  if (!selectedRequest.value) return

  isProcessing.value = true
  try {
    await projectRequestService.review(selectedRequest.value.id, {
      action: reviewAction.value,
      notes: reviewNotes.value || undefined,
    })
    toast.success(
      reviewAction.value === 'approve'
        ? 'Permintaan berhasil disetujui'
        : 'Permintaan berhasil ditolak'
    )
    isReviewDialogOpen.value = false
    fetchRequests()
  } catch (error: any) {
    console.error('Failed to review request:', error)
    toast.error(error.response?.data?.error || 'Gagal memproses permintaan')
  } finally {
    isProcessing.value = false
  }
}

function getStatusBadge(status: ProjectRequestStatus) {
  switch (status) {
    case 'pending':
      return { variant: 'outline' as const, label: 'Menunggu', icon: Clock }
    case 'approved':
      return { variant: 'default' as const, label: 'Disetujui', icon: CheckCircle2 }
    case 'rejected':
      return { variant: 'destructive' as const, label: 'Ditolak', icon: XCircle }
    default:
      return { variant: 'outline' as const, label: status, icon: Clock }
  }
}

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function goToPage(page: number) {
  if (page >= 1 && page <= totalPages.value) {
    currentPage.value = page
    fetchRequests()
  }
}

onMounted(() => {
  fetchOrganizations()
  fetchRequests()
})
</script>

<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Header -->
      <div>
        <h1 class="text-2xl font-bold tracking-tight">Permintaan Proyek ODK</h1>
        <p class="text-muted-foreground">
          Kelola permintaan penugasan proyek ODK untuk grup
        </p>
      </div>

      <!-- Stats Cards -->
      <div class="grid gap-4 md:grid-cols-4">
        <Card>
          <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle class="text-sm font-medium">Total Permintaan</CardTitle>
            <FolderOpen class="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div class="text-2xl font-bold">{{ totalItems }}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle class="text-sm font-medium">Menunggu</CardTitle>
            <Clock class="h-4 w-4 text-yellow-500" />
          </CardHeader>
          <CardContent>
            <div class="text-2xl font-bold text-yellow-500">{{ stats.pending }}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle class="text-sm font-medium">Disetujui</CardTitle>
            <CheckCircle2 class="h-4 w-4 text-green-500" />
          </CardHeader>
          <CardContent>
            <div class="text-2xl font-bold text-green-500">{{ stats.approved }}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle class="text-sm font-medium">Ditolak</CardTitle>
            <XCircle class="h-4 w-4 text-red-500" />
          </CardHeader>
          <CardContent>
            <div class="text-2xl font-bold text-red-500">{{ stats.rejected }}</div>
          </CardContent>
        </Card>
      </div>

      <!-- Filters -->
      <Card>
        <CardContent class="pt-6">
          <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
            <div class="flex flex-1 gap-4">
              <Select v-model="statusFilter">
                <SelectTrigger class="w-[180px]">
                  <SelectValue placeholder="Filter Status" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">Semua Status</SelectItem>
                  <SelectItem value="pending">Menunggu</SelectItem>
                  <SelectItem value="approved">Disetujui</SelectItem>
                  <SelectItem value="rejected">Ditolak</SelectItem>
                </SelectContent>
              </Select>

              <Select v-model="selectedOrgFilter">
                <SelectTrigger class="w-[250px]">
                  <SelectValue placeholder="Filter Organisasi" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">Semua Organisasi</SelectItem>
                  <SelectItem
                    v-for="org in organizations"
                    :key="org.id"
                    :value="org.id"
                  >
                    {{ org.name }}
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
        </CardContent>
      </Card>

      <!-- Table -->
      <Card>
        <CardContent class="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Grup</TableHead>
                <TableHead>Organisasi</TableHead>
                <TableHead>Proyek ODK</TableHead>
                <TableHead>Pemohon</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Tanggal</TableHead>
                <TableHead class="text-right">Aksi</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <!-- Loading state -->
              <template v-if="isLoading">
                <TableRow v-for="i in 5" :key="i">
                  <TableCell><Skeleton class="h-4 w-32" /></TableCell>
                  <TableCell><Skeleton class="h-4 w-28" /></TableCell>
                  <TableCell><Skeleton class="h-4 w-36" /></TableCell>
                  <TableCell><Skeleton class="h-4 w-24" /></TableCell>
                  <TableCell><Skeleton class="h-6 w-20" /></TableCell>
                  <TableCell><Skeleton class="h-4 w-28" /></TableCell>
                  <TableCell><Skeleton class="h-8 w-20 ml-auto" /></TableCell>
                </TableRow>
              </template>

              <!-- Empty state -->
              <template v-else-if="requests.length === 0">
                <TableRow>
                  <TableCell colspan="7" class="h-32 text-center">
                    <div class="flex flex-col items-center justify-center text-muted-foreground">
                      <FolderOpen class="h-8 w-8 mb-2" />
                      <p>Tidak ada permintaan ditemukan</p>
                    </div>
                  </TableCell>
                </TableRow>
              </template>

              <!-- Data rows -->
              <template v-else>
                <TableRow v-for="request in requests" :key="request.id">
                  <TableCell>
                    <div class="flex items-center gap-2">
                      <Users class="h-4 w-4 text-muted-foreground" />
                      <span class="font-medium">{{ request.group?.name || '-' }}</span>
                    </div>
                  </TableCell>
                  <TableCell>
                    <div class="flex items-center gap-2">
                      <Building2 class="h-4 w-4 text-muted-foreground" />
                      <span>{{ request.organization?.name || '-' }}</span>
                    </div>
                  </TableCell>
                  <TableCell>
                    <span class="font-mono text-sm">
                      {{ request.odk_project_name || `Project #${request.odk_project_id}` }}
                    </span>
                  </TableCell>
                  <TableCell>
                    {{ request.requester?.name || request.requester?.email || '-' }}
                  </TableCell>
                  <TableCell>
                    <Badge :variant="getStatusBadge(request.status).variant">
                      <component :is="getStatusBadge(request.status).icon" class="h-3 w-3 mr-1" />
                      {{ getStatusBadge(request.status).label }}
                    </Badge>
                  </TableCell>
                  <TableCell class="text-sm text-muted-foreground">
                    {{ formatDate(request.created_at) }}
                  </TableCell>
                  <TableCell class="text-right">
                    <div class="flex items-center justify-end gap-2">
                      <template v-if="request.status === 'pending'">
                        <Button
                          size="sm"
                          variant="outline"
                          class="text-green-600 hover:text-green-700"
                          @click="openReviewDialog(request, 'approve')"
                        >
                          <Check class="h-4 w-4 mr-1" />
                          Setujui
                        </Button>
                        <Button
                          size="sm"
                          variant="outline"
                          class="text-red-600 hover:text-red-700"
                          @click="openReviewDialog(request, 'reject')"
                        >
                          <X class="h-4 w-4 mr-1" />
                          Tolak
                        </Button>
                      </template>
                      <template v-else>
                        <span class="text-sm text-muted-foreground">
                          {{ request.reviewer?.name || '-' }}
                        </span>
                      </template>
                    </div>
                  </TableCell>
                </TableRow>
              </template>
            </TableBody>
          </Table>
        </CardContent>

        <!-- Pagination -->
        <div
          v-if="totalPages > 1"
          class="flex items-center justify-between px-6 py-4 border-t"
        >
          <div class="text-sm text-muted-foreground">
            Menampilkan {{ (currentPage - 1) * pageSize + 1 }} -
            {{ Math.min(currentPage * pageSize, totalItems) }} dari {{ totalItems }}
          </div>
          <div class="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              :disabled="currentPage === 1"
              @click="goToPage(currentPage - 1)"
            >
              <ChevronLeft class="h-4 w-4" />
            </Button>
            <span class="text-sm">{{ currentPage }} / {{ totalPages }}</span>
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

      <!-- Review Dialog -->
      <Dialog v-model:open="isReviewDialogOpen">
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {{ reviewAction === 'approve' ? 'Setujui Permintaan' : 'Tolak Permintaan' }}
            </DialogTitle>
            <DialogDescription>
              <template v-if="reviewAction === 'approve'">
                Dengan menyetujui, Group Leader akan menjadi Project Manager di ODK Central
                untuk proyek ini.
              </template>
              <template v-else>
                Berikan alasan penolakan permintaan ini.
              </template>
            </DialogDescription>
          </DialogHeader>

          <div class="space-y-4 py-4">
            <!-- Request Info -->
            <div class="rounded-lg border p-4 space-y-2 text-sm">
              <div class="flex justify-between">
                <span class="text-muted-foreground">Grup:</span>
                <span class="font-medium">{{ selectedRequest?.group?.name }}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-muted-foreground">Proyek ODK:</span>
                <span class="font-mono">
                  {{ selectedRequest?.odk_project_name || `#${selectedRequest?.odk_project_id}` }}
                </span>
              </div>
              <div class="flex justify-between">
                <span class="text-muted-foreground">Pemohon:</span>
                <span>{{ selectedRequest?.requester?.name || selectedRequest?.requester?.email }}</span>
              </div>
              <div v-if="selectedRequest?.request_notes" class="pt-2 border-t">
                <span class="text-muted-foreground">Catatan pemohon:</span>
                <p class="mt-1">{{ selectedRequest.request_notes }}</p>
              </div>
            </div>

            <!-- Notes input -->
            <div class="space-y-2">
              <Label for="review-notes">
                {{ reviewAction === 'approve' ? 'Catatan (opsional)' : 'Alasan penolakan' }}
              </Label>
              <Textarea
                id="review-notes"
                v-model="reviewNotes"
                :placeholder="reviewAction === 'approve' ? 'Tambahkan catatan...' : 'Berikan alasan penolakan...'"
                :required="reviewAction === 'reject'"
              />
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" @click="isReviewDialogOpen = false">
              Batal
            </Button>
            <Button
              :variant="reviewAction === 'approve' ? 'default' : 'destructive'"
              :disabled="isProcessing || (reviewAction === 'reject' && !reviewNotes)"
              @click="submitReview"
            >
              <template v-if="isProcessing">Memproses...</template>
              <template v-else>
                {{ reviewAction === 'approve' ? 'Setujui' : 'Tolak' }}
              </template>
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  </AppLayout>
</template>
