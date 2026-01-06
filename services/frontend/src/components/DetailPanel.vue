<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { X, Phone, User, ChevronDown, ChevronUp, ChevronRight, Clock, MapPin, Users, Home, Wifi, Navigation, Camera, ArrowLeft } from 'lucide-vue-next'
import Badge from './ui/Badge.vue'
import Button from './ui/Button.vue'
import PhotoModal from './PhotoModal.vue'
import type { MapMarker } from '@/types'
import { api, type LocationDetail, type Feed, type Photo } from '@/services/api'

interface Props {
  marker: MapMarker | null
}

const props = defineProps<Props>()
const emit = defineEmits<{
  close: []
  'show-location-updates': [locationId: string]
}>()

const locationDetail = ref<LocationDetail | null>(null)
const latestFeed = ref<Feed | null>(null)
const loading = ref(false)

// Photo state
const photos = ref<Photo[]>([])
const viewMode = ref<'detail' | 'gallery'>('detail')
const selectedPhotoUrl = ref<string | null>(null)
const isModalOpen = ref(false)

// Collapsible sections state
const expandedSections = ref<Record<string, boolean>>({
  pengungsi: false,
  fasilitas: false,
  komunikasi: false,
  akses: false,
})

const toggleSection = (section: string) => {
  expandedSections.value[section] = !expandedSections.value[section]
}

// Fetch location detail, feeds, and photos when marker changes
watch(() => props.marker, async (newMarker) => {
  if (!newMarker) {
    locationDetail.value = null
    latestFeed.value = null
    photos.value = []
    viewMode.value = 'detail'
    return
  }

  loading.value = true
  try {
    // Fetch location detail, feeds, and photos in parallel
    const [detailRes, feedsRes, photosRes] = await Promise.all([
      api.getLocationById(newMarker.id),
      api.getFeedsByLocation(newMarker.id, { limit: 1 }),
      api.getPhotosByLocation(newMarker.id)
    ])

    if (detailRes.success && detailRes.data) {
      locationDetail.value = detailRes.data
    }

    if (feedsRes.success && feedsRes.data && feedsRes.data.length > 0) {
      latestFeed.value = feedsRes.data[0]
    } else {
      latestFeed.value = null
    }

    if (photosRes.success && photosRes.data) {
      photos.value = photosRes.data
    } else {
      photos.value = []
    }
  } catch (e) {
    console.error('Failed to fetch location detail:', e)
  } finally {
    loading.value = false
  }
}, { immediate: true })

// Format ID from alamat
const locationId = computed(() => {
  if (!locationDetail.value?.alamat) return '-'
  const alamat = locationDetail.value.alamat as Record<string, string>
  const idDesa = alamat.id_desa || ''
  const nama = locationDetail.value.identitas?.nama || props.marker?.name || ''
  return `${idDesa} - ${nama}`
})

// Format status badge
const statusVariant = computed(() => {
  const status = locationDetail.value?.status || props.marker?.status
  switch (status) {
    case 'operasional': return 'success'
    case 'operational': return 'success'
    case 'non_aktif': return 'danger'
    case 'evakuasi': return 'warning'
    case 'persiapan_huntara': return 'default'
    default: return 'outline'
  }
})

const statusLabel = computed(() => {
  const status = locationDetail.value?.status || props.marker?.status
  switch (status) {
    case 'operasional':
    case 'operational': return 'Operasional'
    case 'non_aktif': return 'Non-aktif'
    case 'evakuasi': return 'Evakuasi'
    case 'persiapan_huntara': return 'Persiapan Huntara'
    default: return status || 'Unknown'
  }
})

// Format date
const formatDate = (dateStr: string | undefined) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleDateString('id-ID', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

// Format relative time
const formatRelativeTime = (dateStr: string | undefined) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMins = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMins / 60)
  const diffDays = Math.floor(diffHours / 24)

  if (diffMins < 60) return `${diffMins} menit yang lalu`
  if (diffHours < 24) return `${diffHours} jam yang lalu`
  return `${diffDays} hari yang lalu`
}

// Category and type colors
const categoryColors: Record<string, 'default' | 'success' | 'warning' | 'danger' | 'outline'> = {
  kebutuhan: 'warning',
  informasi: 'outline',
  'follow-up': 'danger',
}

const handleShowMoreUpdates = () => {
  if (props.marker) {
    emit('show-location-updates', props.marker.id)
  }
}

// Helper to display Yes/No
const formatYesNo = (value: unknown) => {
  if (value === 'yes' || value === true) return 'Ya'
  if (value === 'no' || value === false) return 'Tidak'
  return value || '-'
}

// Helper to format terisolir
const formatTerisolir = (value: unknown) => {
  if (value === 'yes' || value === true) return 'Terisolir'
  if (value === 'no' || value === false) return 'Tidak Terisolir'
  return value || '-'
}

// Photo helpers
const cachedPhotos = computed(() => photos.value.filter(p => p.is_cached && p.url))

const getPhotoUrl = (photo: Photo) => {
  return api.getPhotoUrl(photo.id)
}

const openPhotoGallery = () => {
  if (cachedPhotos.value.length > 0) {
    viewMode.value = 'gallery'
  }
}

const backToDetail = () => {
  viewMode.value = 'detail'
}

const openPhotoModal = (photo: Photo) => {
  selectedPhotoUrl.value = getPhotoUrl(photo)
  isModalOpen.value = true
}

const closePhotoModal = () => {
  isModalOpen.value = false
  selectedPhotoUrl.value = null
}
</script>

<template>
  <!-- Mobile backdrop (covers area behind panel, clicking closes it) -->
  <div
    v-if="marker"
    class="fixed inset-0 left-14 bg-black/30 z-40 lg:hidden"
    @click="emit('close')"
  />

  <aside
    v-if="marker"
    class="fixed inset-y-0 left-14 right-0 lg:h-full lg:relative lg:inset-auto lg:w-96 bg-white border-l border-gray-200 flex flex-col overflow-hidden z-50 lg:border-t-0"
  >
    <!-- Header -->
    <div class="p-3 lg:p-4 border-b border-gray-200">
      <div class="flex items-start justify-between gap-2">
        <div class="flex-1 min-w-0">
          <h2 class="text-base lg:text-lg font-semibold text-gray-900 truncate">{{ marker.name }}</h2>
          <div class="flex items-center gap-2 mt-1">
            <Badge :variant="statusVariant">Status Posko: {{ statusLabel }}</Badge>
          </div>
          <div class="text-xs text-gray-500 mt-1">ID: {{ locationId }}</div>
        </div>
        <button
          class="p-1 hover:bg-gray-100 rounded"
          @click="emit('close')"
        >
          <X class="w-5 h-5 text-gray-400" />
        </button>
      </div>
    </div>

    <!-- Photo Section (when photos available and in detail mode) -->
    <div v-if="!loading && cachedPhotos.length > 0 && viewMode === 'detail'" class="relative">
      <div
        class="cursor-pointer group"
        @click="openPhotoGallery"
      >
        <img
          :src="getPhotoUrl(cachedPhotos[0])"
          :alt="cachedPhotos[0].filename"
          class="w-full h-48 object-cover"
        />
        <!-- Photo count badge -->
        <div class="absolute bottom-3 right-3 bg-black/70 text-white px-3 py-1.5 rounded-full text-sm flex items-center gap-1.5 group-hover:bg-black/80 transition-colors">
          <Camera class="w-4 h-4" />
          <span>{{ cachedPhotos.length }} Foto</span>
        </div>
        <!-- Hover overlay -->
        <div class="absolute inset-0 bg-black/0 group-hover:bg-black/20 transition-colors flex items-center justify-center">
          <span class="text-white opacity-0 group-hover:opacity-100 transition-opacity text-sm font-medium">
            Lihat Galeri
          </span>
        </div>
      </div>
    </div>

    <!-- Gallery View -->
    <div v-if="!loading && viewMode === 'gallery'" class="flex-1 flex flex-col overflow-hidden">
      <!-- Gallery Header -->
      <div class="p-4 border-b border-gray-200 flex items-center gap-3">
        <button
          class="p-1.5 hover:bg-gray-100 rounded-full transition-colors"
          @click="backToDetail"
        >
          <ArrowLeft class="w-5 h-5 text-gray-600" />
        </button>
        <h3 class="font-semibold text-gray-900">Galeri Foto ({{ cachedPhotos.length }})</h3>
      </div>

      <!-- Thumbnails Grid -->
      <div class="flex-1 overflow-y-auto p-3 lg:p-4">
        <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-2 gap-2 lg:gap-3">
          <div
            v-for="photo in cachedPhotos"
            :key="photo.id"
            class="relative aspect-square cursor-pointer group rounded-lg overflow-hidden"
            @click="openPhotoModal(photo)"
          >
            <img
              :src="getPhotoUrl(photo)"
              :alt="photo.filename"
              class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-200"
            />
            <div class="absolute inset-0 bg-black/0 group-hover:bg-black/20 transition-colors" />
            <div class="absolute bottom-2 left-2 bg-black/60 text-white text-xs px-2 py-1 rounded capitalize">
              {{ photo.photo_type.replace('_', ' ') }}
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="flex-1 flex items-center justify-center">
      <span class="text-gray-500">Memuat data...</span>
    </div>

    <!-- Content (Detail View) -->
    <div v-if="!loading && viewMode === 'detail'" class="flex-1 overflow-y-auto">
      <!-- Latest Update for this location -->
      <div v-if="latestFeed" class="p-4 border-b border-gray-200 bg-blue-50">
        <div class="flex items-center gap-2 mb-2">
          <span class="w-2 h-2 bg-green-500 rounded-full animate-pulse"></span>
          <span class="text-xs font-medium text-gray-600">Update Terbaru</span>
        </div>
        <div class="flex items-center justify-between mb-1">
          <span class="text-xs text-gray-500">{{ formatDate(latestFeed.submitted_at) }}</span>
        </div>
        <div class="text-xs text-blue-600 font-medium mb-2">
          {{ latestFeed.username || '-' }}{{ latestFeed.organization ? ` - ${latestFeed.organization}` : '' }}
        </div>
        <p class="text-sm text-gray-600 mb-3">{{ latestFeed.content }}</p>
        <div class="flex gap-2 mb-3">
          <Badge :variant="categoryColors[latestFeed.category] || 'outline'">
            {{ latestFeed.category }}
          </Badge>
          <Badge v-if="latestFeed.type" variant="outline">
            {{ latestFeed.type }}
          </Badge>
        </div>
        <button
          class="w-full flex items-center justify-center gap-1 py-2 text-sm text-blue-600 hover:bg-blue-100 rounded-lg transition-colors"
          @click="handleShowMoreUpdates"
        >
          <span>Update lainnya</span>
          <ChevronRight class="w-4 h-4" />
        </button>
      </div>

      <!-- No update message -->
      <div v-else class="p-4 border-b border-gray-200 bg-gray-50">
        <div class="flex items-center gap-2 mb-2">
          <span class="w-2 h-2 bg-gray-400 rounded-full"></span>
          <span class="text-xs font-medium text-gray-600">Update Terbaru</span>
        </div>
        <p class="text-sm text-gray-500">Belum ada update untuk lokasi ini.</p>
      </div>

      <!-- Informasi Umum Posko -->
      <div class="p-4 border-b border-gray-200">
        <h3 class="font-semibold text-gray-900 mb-3 flex items-center gap-2">
          <Home class="w-4 h-4" />
          Informasi Umum Posko
        </h3>
        <div class="grid grid-cols-2 gap-4">
          <div>
            <div class="text-xs text-gray-500">Kapasitas</div>
            <div class="text-sm font-medium">
              {{ (locationDetail?.data_pengungsi as any)?.total_pengungsi || '-' }} Jiwa
            </div>
          </div>
          <div>
            <div class="text-xs text-gray-500">Jenis Posko Pengungsian</div>
            <div class="text-sm font-medium capitalize">
              {{ (locationDetail?.data_pengungsi as any)?.jenis_pengungsian || '-' }}
            </div>
          </div>
          <div>
            <div class="text-xs text-gray-500">Terisolir</div>
            <div class="text-sm font-medium">
              {{ formatTerisolir((locationDetail?.data_pengungsi as any)?.terisolir) }}
            </div>
          </div>
          <div>
            <div class="text-xs text-gray-500">Penanggung Jawab Posko</div>
            <div class="text-sm font-medium">
              {{ (locationDetail?.identitas as any)?.nama_penanggungjawab || '-' }}
            </div>
          </div>
        </div>
      </div>

      <!-- Kontak Posko -->
      <div class="p-4 border-b border-gray-200">
        <h3 class="font-semibold text-gray-900 mb-3 flex items-center gap-2">
          <Phone class="w-4 h-4" />
          Kontak Posko
        </h3>
        <div class="space-y-3">
          <div class="flex items-start gap-3">
            <Phone class="w-5 h-5 text-blue-500 mt-0.5" />
            <div>
              <div class="text-sm font-medium">
                {{ (locationDetail?.identitas as any)?.contact_penanggungjawab || '-' }}
              </div>
              <div class="text-xs text-gray-500">Nomor Kontak Penanggung Jawab</div>
            </div>
          </div>
          <div class="flex items-start gap-3">
            <Phone class="w-5 h-5 text-green-500 mt-0.5" />
            <div>
              <div class="text-sm font-medium">
                {{ (locationDetail?.identitas as any)?.contact_relawan || '-' }}
              </div>
              <div class="text-xs text-gray-500">Nomor Kontak Relawan</div>
            </div>
          </div>
          <div class="flex items-start gap-3">
            <User class="w-5 h-5 text-gray-400 mt-0.5" />
            <div>
              <div class="text-sm font-medium">
                {{ (locationDetail?.identitas as any)?.nama_relawan || '-' }}
              </div>
              <div class="text-xs text-gray-500">Nama Relawan</div>
            </div>
          </div>
          <div class="flex items-start gap-3">
            <Clock class="w-5 h-5 text-gray-400 mt-0.5" />
            <div>
              <div class="text-sm font-medium">
                {{ formatRelativeTime(locationDetail?.meta?.updated_at) }}
              </div>
              <div class="text-xs text-gray-500">Last Update</div>
            </div>
          </div>
        </div>
      </div>

      <!-- Data Pengungsi (Collapsible) -->
      <div class="border-b border-gray-200">
        <button
          class="w-full flex items-center justify-between p-4 hover:bg-gray-50"
          @click="toggleSection('pengungsi')"
        >
          <span class="font-medium text-gray-700 flex items-center gap-2">
            <Users class="w-4 h-4" />
            Data Pengungsi
          </span>
          <component :is="expandedSections.pengungsi ? ChevronUp : ChevronDown" class="w-5 h-5 text-gray-400" />
        </button>
        <div v-if="expandedSections.pengungsi" class="px-4 pb-4">
          <div class="grid grid-cols-2 gap-3 text-sm">
            <div>
              <span class="text-gray-500">Jumlah KK:</span>
              <span class="ml-1 font-medium">{{ (locationDetail?.data_pengungsi as any)?.jumlah_kk || '-' }}</span>
            </div>
            <div>
              <span class="text-gray-500">Total Pengungsi:</span>
              <span class="ml-1 font-medium">{{ (locationDetail?.data_pengungsi as any)?.total_pengungsi || '-' }}</span>
            </div>
            <div>
              <span class="text-gray-500">Lansia:</span>
              <span class="ml-1 font-medium">{{ (locationDetail?.data_pengungsi as any)?.lansia || '-' }}</span>
            </div>
            <div>
              <span class="text-gray-500">Ibu Hamil:</span>
              <span class="ml-1 font-medium">{{ (locationDetail?.data_pengungsi as any)?.ibu_hamil || '-' }}</span>
            </div>
            <div>
              <span class="text-gray-500">Ibu Menyusui:</span>
              <span class="ml-1 font-medium">{{ (locationDetail?.data_pengungsi as any)?.ibu_menyusui || '-' }}</span>
            </div>
            <div>
              <span class="text-gray-500">Balita:</span>
              <span class="ml-1 font-medium">
                {{ ((locationDetail?.data_pengungsi as any)?.balita_laki || 0) + ((locationDetail?.data_pengungsi as any)?.balita_perempuan || 0) || '-' }}
              </span>
            </div>
            <div>
              <span class="text-gray-500">Difabel:</span>
              <span class="ml-1 font-medium">{{ (locationDetail?.data_pengungsi as any)?.difabel || '-' }}</span>
            </div>
            <div>
              <span class="text-gray-500">Komorbid:</span>
              <span class="ml-1 font-medium">{{ (locationDetail?.data_pengungsi as any)?.komorbid || '-' }}</span>
            </div>
            <div>
              <span class="text-gray-500">Anak Tanpa Ortu:</span>
              <span class="ml-1 font-medium">{{ (locationDetail?.data_pengungsi as any)?.anak_tanpa_ortu || '-' }}</span>
            </div>
            <div>
              <span class="text-gray-500">Remaja Tanpa Ortu:</span>
              <span class="ml-1 font-medium">{{ (locationDetail?.data_pengungsi as any)?.remaja_tanpa_ortu || '-' }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Fasilitas Layanan (Collapsible) -->
      <div class="border-b border-gray-200">
        <button
          class="w-full flex items-center justify-between p-4 hover:bg-gray-50"
          @click="toggleSection('fasilitas')"
        >
          <span class="font-medium text-gray-700 flex items-center gap-2">
            <Home class="w-4 h-4" />
            Fasilitas Layanan
          </span>
          <component :is="expandedSections.fasilitas ? ChevronUp : ChevronDown" class="w-5 h-5 text-gray-400" />
        </button>
        <div v-if="expandedSections.fasilitas" class="px-4 pb-4">
          <div class="grid grid-cols-2 gap-3 text-sm">
            <div>
              <span class="text-gray-500">Posko Logistik:</span>
              <span class="ml-1 font-medium">{{ formatYesNo((locationDetail?.fasilitas as any)?.posko_logistik) }}</span>
            </div>
            <div>
              <span class="text-gray-500">Posko Faskes:</span>
              <span class="ml-1 font-medium">{{ formatYesNo((locationDetail?.fasilitas as any)?.posko_faskes) }}</span>
            </div>
            <div>
              <span class="text-gray-500">Dapur Umum:</span>
              <span class="ml-1 font-medium">{{ formatYesNo((locationDetail?.fasilitas as any)?.dapur_umum) }}</span>
            </div>
            <div>
              <span class="text-gray-500">Kapasitas Dapur:</span>
              <span class="ml-1 font-medium">{{ (locationDetail?.fasilitas as any)?.kapasitas_dapur || '-' }}</span>
            </div>
            <div>
              <span class="text-gray-500">Toilet Perempuan:</span>
              <span class="ml-1 font-medium">{{ (locationDetail?.fasilitas as any)?.toilet_perempuan || '-' }}</span>
            </div>
            <div>
              <span class="text-gray-500">Toilet Laki-laki:</span>
              <span class="ml-1 font-medium">{{ (locationDetail?.fasilitas as any)?.toilet_laki || '-' }}</span>
            </div>
            <div>
              <span class="text-gray-500">Tempat Sampah:</span>
              <span class="ml-1 font-medium">{{ (locationDetail?.fasilitas as any)?.tempat_sampah || '-' }}</span>
            </div>
            <div>
              <span class="text-gray-500">Tenaga Medis:</span>
              <span class="ml-1 font-medium">{{ formatYesNo((locationDetail?.fasilitas as any)?.posko_tenaga_medis) }}</span>
            </div>
            <div>
              <span class="text-gray-500">Ruang Laktasi:</span>
              <span class="ml-1 font-medium">{{ formatYesNo((locationDetail?.fasilitas as any)?.ruang_laktasi) }}</span>
            </div>
            <div>
              <span class="text-gray-500">Sekolah Darurat:</span>
              <span class="ml-1 font-medium">{{ formatYesNo((locationDetail?.fasilitas as any)?.sekolah_darurat) }}</span>
            </div>
            <div>
              <span class="text-gray-500">Area Bermain:</span>
              <span class="ml-1 font-medium">{{ formatYesNo((locationDetail?.fasilitas as any)?.area_bermain) }}</span>
            </div>
            <div>
              <span class="text-gray-500">Psikososial:</span>
              <span class="ml-1 font-medium">{{ formatYesNo((locationDetail?.fasilitas as any)?.posko_psikososial) }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Sarana Komunikasi (Collapsible) -->
      <div class="border-b border-gray-200">
        <button
          class="w-full flex items-center justify-between p-4 hover:bg-gray-50"
          @click="toggleSection('komunikasi')"
        >
          <span class="font-medium text-gray-700 flex items-center gap-2">
            <Wifi class="w-4 h-4" />
            Sarana Komunikasi
          </span>
          <component :is="expandedSections.komunikasi ? ChevronUp : ChevronDown" class="w-5 h-5 text-gray-400" />
        </button>
        <div v-if="expandedSections.komunikasi" class="px-4 pb-4">
          <div class="space-y-2 text-sm">
            <div>
              <span class="text-gray-500">Ketersediaan Sinyal:</span>
              <span class="ml-1 font-medium capitalize">{{ (locationDetail?.komunikasi as any)?.ketersediaan_sinyal || '-' }}</span>
            </div>
            <div>
              <span class="text-gray-500">Ketersediaan Internet:</span>
              <span class="ml-1 font-medium capitalize">{{ (locationDetail?.komunikasi as any)?.ketersediaan_internet || '-' }}</span>
            </div>
            <div>
              <span class="text-gray-500">Jaringan ORARI:</span>
              <span class="ml-1 font-medium">{{ formatYesNo((locationDetail?.komunikasi as any)?.jaringan_orari) }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Akses Fasilitas Umum (Collapsible) -->
      <div class="border-b border-gray-200">
        <button
          class="w-full flex items-center justify-between p-4 hover:bg-gray-50"
          @click="toggleSection('akses')"
        >
          <span class="font-medium text-gray-700 flex items-center gap-2">
            <Navigation class="w-4 h-4" />
            Akses Fasilitas Umum
          </span>
          <component :is="expandedSections.akses ? ChevronUp : ChevronDown" class="w-5 h-5 text-gray-400" />
        </button>
        <div v-if="expandedSections.akses" class="px-4 pb-4">
          <div class="space-y-2 text-sm">
            <div>
              <span class="text-gray-500">Jarak ke Puskesmas:</span>
              <span class="ml-1 font-medium">{{ (locationDetail?.akses as any)?.jarak_pkm || '-' }} km</span>
            </div>
            <div>
              <span class="text-gray-500">Jarak ke Posko Logistik:</span>
              <span class="ml-1 font-medium">{{ (locationDetail?.akses as any)?.jarak_posko_logistik || '-' }} km</span>
            </div>
            <div>
              <span class="text-gray-500">Sumber Air:</span>
              <span class="ml-1 font-medium capitalize">{{ (locationDetail?.fasilitas as any)?.sumber_air || '-' }}</span>
            </div>
            <div>
              <span class="text-gray-500">Sumber Listrik:</span>
              <span class="ml-1 font-medium capitalize">{{ (locationDetail?.fasilitas as any)?.sumber_listrik || '-' }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Alamat -->
      <div class="p-4 border-b border-gray-200">
        <h3 class="font-semibold text-gray-900 mb-3 flex items-center gap-2">
          <MapPin class="w-4 h-4" />
          Alamat
        </h3>
        <div class="text-sm text-gray-600">
          <p>{{ (locationDetail?.identitas as any)?.alamat_dusun || '' }}</p>
          <p>
            Desa {{ (locationDetail?.alamat as any)?.nama_desa || '-' }},
            Kec. {{ (locationDetail?.alamat as any)?.nama_kecamatan || '-' }}
          </p>
          <p>
            {{ (locationDetail?.alamat as any)?.nama_kota_kab || '-' }},
            {{ (locationDetail?.alamat as any)?.nama_provinsi || '-' }}
          </p>
        </div>
      </div>
    </div>

    <!-- Footer (only in detail mode) -->
    <div v-if="viewMode === 'detail'" class="p-4 border-t border-gray-200">
      <Button variant="primary" class="w-full">
        Generate Report
      </Button>
    </div>

    <!-- Photo Modal -->
    <PhotoModal
      :image-url="selectedPhotoUrl || ''"
      :is-open="isModalOpen"
      @close="closePhotoModal"
    />
  </aside>
</template>
