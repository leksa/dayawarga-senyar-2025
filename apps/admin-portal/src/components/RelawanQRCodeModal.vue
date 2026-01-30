<script setup lang="ts">
import { ref, watch, computed, nextTick } from 'vue'
import { toast } from 'vue-sonner'
import { QrCode, Copy, Download, RefreshCw, Trash2 } from 'lucide-vue-next'
import QRCodeLib from 'qrcode'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Skeleton } from '@/components/ui/skeleton'
import { relawanODKService, type QRCodeResponse, type Relawan } from '@/services'

const props = defineProps<{
  relawan: Relawan | null
  open: boolean
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'refresh'): void
}>()

const isLoading = ref(false)
const isCreating = ref(false)
const isRevoking = ref(false)
const qrData = ref<QRCodeResponse | null>(null)
const error = ref<string | null>(null)
const qrCanvas = ref<HTMLCanvasElement | null>(null)

const dialogOpen = computed({
  get: () => props.open,
  set: (value) => emit('update:open', value),
})

const hasODKAccess = computed(() => {
  return props.relawan?.odk_app_user_id != null
})

watch(
  () => props.open,
  async (isOpen) => {
    if (isOpen && props.relawan && hasODKAccess.value) {
      await fetchQRCode()
    } else {
      qrData.value = null
      error.value = null
    }
  }
)

async function fetchQRCode() {
  if (!props.relawan) return

  isLoading.value = true
  error.value = null

  try {
    qrData.value = await relawanODKService.getQRCode(props.relawan.id)
    // Generate QR code image after data is loaded
    await nextTick()
    await generateQRImage()
  } catch (err: any) {
    console.error('Failed to fetch QR code:', err)
    error.value = err.response?.data?.error || 'Gagal memuat QR code'
  } finally {
    isLoading.value = false
  }
}

async function generateQRImage() {
  if (!qrData.value?.qr_code_data || !qrCanvas.value) return

  try {
    await QRCodeLib.toCanvas(qrCanvas.value, qrData.value.qr_code_data, {
      width: 256,
      margin: 2,
      color: {
        dark: '#000000',
        light: '#FFFFFF',
      },
    })
  } catch (err) {
    console.error('Failed to generate QR image:', err)
    error.value = 'Gagal menghasilkan gambar QR code'
  }
}

async function createAppUser() {
  if (!props.relawan) return

  isCreating.value = true
  try {
    await relawanODKService.createAppUser(props.relawan.id)
    toast.success('ODK App User berhasil dibuat')
    emit('refresh')
    await fetchQRCode()
  } catch (err: any) {
    console.error('Failed to create app user:', err)
    toast.error(err.response?.data?.error || 'Gagal membuat ODK App User')
  } finally {
    isCreating.value = false
  }
}

async function revokeAppUser() {
  if (!props.relawan) return

  isRevoking.value = true
  try {
    await relawanODKService.revokeAppUser(props.relawan.id)
    toast.success('ODK App User berhasil dicabut')
    qrData.value = null
    emit('refresh')
    dialogOpen.value = false
  } catch (err: any) {
    console.error('Failed to revoke app user:', err)
    toast.error(err.response?.data?.error || 'Gagal mencabut ODK App User')
  } finally {
    isRevoking.value = false
  }
}

function copyQRData() {
  if (!qrData.value?.qr_code_data) return

  navigator.clipboard.writeText(qrData.value.qr_code_data)
  toast.success('QR data disalin ke clipboard')
}

async function downloadQRCode() {
  if (!qrData.value?.qr_code_data) return

  try {
    // Generate QR code as data URL
    const dataUrl = await QRCodeLib.toDataURL(qrData.value.qr_code_data, {
      width: 512,
      margin: 2,
      color: {
        dark: '#000000',
        light: '#FFFFFF',
      },
    })

    // Download as PNG
    const a = document.createElement('a')
    a.href = dataUrl
    a.download = `odk-qr-${props.relawan?.name.replace(/\s+/g, '-')}.png`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    toast.success('QR code berhasil diunduh')
  } catch (err) {
    console.error('Failed to download QR code:', err)
    toast.error('Gagal mengunduh QR code')
  }
}

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}
</script>

<template>
  <Dialog v-model:open="dialogOpen">
    <DialogContent class="sm:max-w-md">
      <DialogHeader>
        <DialogTitle class="flex items-center gap-2">
          <QrCode class="h-5 w-5" />
          QR Code ODK Collect
        </DialogTitle>
        <DialogDescription>
          {{ relawan?.name }}
        </DialogDescription>
      </DialogHeader>

      <div class="space-y-4 py-4">
        <!-- No ODK Access -->
        <template v-if="!hasODKAccess">
          <div class="text-center py-8 space-y-4">
            <div class="mx-auto w-16 h-16 rounded-full bg-muted flex items-center justify-center">
              <QrCode class="h-8 w-8 text-muted-foreground" />
            </div>
            <div class="space-y-2">
              <p class="font-medium">Belum memiliki akses ODK</p>
              <p class="text-sm text-muted-foreground">
                Relawan ini belum memiliki ODK App User. Buat App User untuk
                menghasilkan QR code.
              </p>
            </div>
            <Button @click="createAppUser" :disabled="isCreating">
              <template v-if="isCreating">
                <RefreshCw class="h-4 w-4 mr-2 animate-spin" />
                Membuat...
              </template>
              <template v-else>
                <QrCode class="h-4 w-4 mr-2" />
                Buat ODK App User
              </template>
            </Button>
          </div>
        </template>

        <!-- Loading -->
        <template v-else-if="isLoading">
          <div class="space-y-4">
            <Skeleton class="h-48 w-48 mx-auto" />
            <Skeleton class="h-4 w-32 mx-auto" />
            <Skeleton class="h-4 w-24 mx-auto" />
          </div>
        </template>

        <!-- Error -->
        <template v-else-if="error">
          <div class="text-center py-8 space-y-4">
            <p class="text-sm text-destructive">{{ error }}</p>
            <Button variant="outline" @click="fetchQRCode">
              <RefreshCw class="h-4 w-4 mr-2" />
              Coba Lagi
            </Button>
          </div>
        </template>

        <!-- QR Code Display -->
        <template v-else-if="qrData">
          <div class="text-center space-y-4">
            <!-- QR Code Image -->
            <div class="mx-auto w-64 h-64 flex items-center justify-center bg-white rounded-lg p-2 border">
              <canvas ref="qrCanvas" class="max-w-full max-h-full"></canvas>
            </div>

            <!-- QR Data (truncated) -->
            <div class="bg-muted rounded-lg p-3 text-xs font-mono break-all max-h-20 overflow-hidden">
              {{ qrData.qr_code_data.substring(0, 100) }}...
            </div>

            <!-- Info -->
            <div class="text-sm space-y-1 text-muted-foreground">
              <p>
                <span class="font-medium text-foreground">Grup:</span>
                {{ qrData.group_name }}
              </p>
              <p>
                <span class="font-medium text-foreground">Project ID:</span>
                {{ qrData.project_id }}
              </p>
              <p>
                <span class="font-medium text-foreground">Dibuat:</span>
                {{ formatDate(qrData.created_at) }}
              </p>
            </div>

            <!-- Actions -->
            <div class="flex justify-center gap-2">
              <Button variant="outline" size="sm" @click="copyQRData">
                <Copy class="h-4 w-4 mr-1" />
                Salin
              </Button>
              <Button variant="outline" size="sm" @click="downloadQRCode">
                <Download class="h-4 w-4 mr-1" />
                Unduh
              </Button>
            </div>
          </div>
        </template>
      </div>

      <DialogFooter class="flex-col sm:flex-row gap-2">
        <Button
          v-if="hasODKAccess"
          variant="destructive"
          size="sm"
          :disabled="isRevoking"
          @click="revokeAppUser"
        >
          <template v-if="isRevoking">
            <RefreshCw class="h-4 w-4 mr-2 animate-spin" />
            Mencabut...
          </template>
          <template v-else>
            <Trash2 class="h-4 w-4 mr-2" />
            Cabut Akses
          </template>
        </Button>
        <Button variant="outline" @click="dialogOpen = false">
          Tutup
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
