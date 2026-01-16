<script setup lang="ts">
import { ref, watch, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Search, MapPin, Download, Filter, Image, Navigation, Megaphone, ChevronDown, X } from 'lucide-vue-next'
import DataLayersSidebar from '@/components/DataLayersSidebar.vue'
import Input from '@/components/ui/Input.vue'
import Badge from '@/components/ui/Badge.vue'
import Button from '@/components/ui/Button.vue'
import { api, type Feed, type FeedFilter, type FeedPhoto } from '@/services/api'

const route = useRoute()
const router = useRouter()

// State
const feeds = ref<Feed[]>([])
const loading = ref(false)
const loadingMore = ref(false)
const total = ref(0)
const page = ref(1)
const limit = 20

// Filters
const searchQuery = ref('')
const selectedCategory = ref('')
const selectedTag = ref('')

// Region cascade filter state
const showRegionFilter = ref(false)
const filterLevel = ref<'provinsi' | 'kotakab' | 'kecamatan' | 'desa'>('provinsi')

// Pending filter values (before clicking Apply)
const pendingProvinsi = ref<string>('')
const pendingKotaKab = ref<string>('')
const pendingKecamatan = ref<string>('')
const pendingDesa = ref<string>('')

// Applied filter values (after clicking Apply)
const appliedProvinsi = ref<string>('')
const appliedKotaKab = ref<string>('')
const appliedKecamatan = ref<string>('')
const appliedDesa = ref<string>('')

// Available options for each level (dynamically populated from data)
const availableKotaKab = ref<string[]>([])
const availableKecamatan = ref<string[]>([])
const availableDesa = ref<string[]>([])

// Collect unique region values from feeds
const collectRegionData = () => {
  const kotaKabSet = new Set<string>()
  const kecamatanMap = new Map<string, Set<string>>()
  const desaMap = new Map<string, Set<string>>()

  feeds.value.forEach(f => {
    if (f.region?.kota_kab) kotaKabSet.add(f.region.kota_kab)
    if (f.region?.kota_kab && f.region?.kecamatan) {
      if (!kecamatanMap.has(f.region.kota_kab)) kecamatanMap.set(f.region.kota_kab, new Set())
      kecamatanMap.get(f.region.kota_kab)!.add(f.region.kecamatan)
    }
    if (f.region?.kecamatan && f.region?.desa) {
      if (!desaMap.has(f.region.kecamatan)) desaMap.set(f.region.kecamatan, new Set())
      desaMap.get(f.region.kecamatan)!.add(f.region.desa)
    }
  })

  return { kotaKabSet, kecamatanMap, desaMap }
}

// Update available kota/kab when provinsi selected
const updateAvailableKotaKab = () => {
  const { kotaKabSet } = collectRegionData()
  availableKotaKab.value = Array.from(kotaKabSet).sort()
}

// Update available kecamatan when kota/kab selected
const updateAvailableKecamatan = (kotaKab: string) => {
  const { kecamatanMap } = collectRegionData()
  const kecSet = kecamatanMap.get(kotaKab)
  availableKecamatan.value = kecSet ? Array.from(kecSet).sort() : []
}

// Update available desa when kecamatan selected
const updateAvailableDesa = (kecamatan: string) => {
  const { desaMap } = collectRegionData()
  const desaSet = desaMap.get(kecamatan)
  availableDesa.value = desaSet ? Array.from(desaSet).sort() : []
}

// Filter label for display (shows applied filter)
const regionFilterLabel = computed(() => {
  if (appliedDesa.value) return appliedDesa.value
  if (appliedKecamatan.value) return appliedKecamatan.value
  if (appliedKotaKab.value) return appliedKotaKab.value
  if (appliedProvinsi.value) return appliedProvinsi.value
  return 'Lokasi'
})

// Check if any region filter is active
const hasActiveRegionFilter = computed(() => {
  return !!(appliedProvinsi.value || appliedKotaKab.value || appliedKecamatan.value || appliedDesa.value)
})

// Handle provinsi selection
const selectProvinsi = (provinsi: string) => {
  pendingProvinsi.value = provinsi
  pendingKotaKab.value = ''
  pendingKecamatan.value = ''
  pendingDesa.value = ''
  if (provinsi === 'Aceh') {
    updateAvailableKotaKab()
    filterLevel.value = 'kotakab'
  }
}

// Handle kota/kab selection
const selectKotaKab = (kotaKab: string) => {
  pendingKotaKab.value = kotaKab
  pendingKecamatan.value = ''
  pendingDesa.value = ''
  if (kotaKab) {
    updateAvailableKecamatan(kotaKab)
    filterLevel.value = 'kecamatan'
  }
}

// Handle kecamatan selection
const selectKecamatan = (kecamatan: string) => {
  pendingKecamatan.value = kecamatan
  pendingDesa.value = ''
  if (kecamatan) {
    updateAvailableDesa(kecamatan)
    filterLevel.value = 'desa'
  }
}

// Handle desa selection
const selectDesa = (desa: string) => {
  pendingDesa.value = desa
}

// Apply the filter
const applyRegionFilter = () => {
  appliedProvinsi.value = pendingProvinsi.value
  appliedKotaKab.value = pendingKotaKab.value
  appliedKecamatan.value = pendingKecamatan.value
  appliedDesa.value = pendingDesa.value
  showRegionFilter.value = false
  fetchFeeds()
}

// Clear region filter
const clearRegionFilter = () => {
  pendingProvinsi.value = ''
  pendingKotaKab.value = ''
  pendingKecamatan.value = ''
  pendingDesa.value = ''
  appliedProvinsi.value = ''
  appliedKotaKab.value = ''
  appliedKecamatan.value = ''
  appliedDesa.value = ''
  filterLevel.value = 'provinsi'
  fetchFeeds()
}

// Go back one level in cascade filter
const goBackLevel = () => {
  if (filterLevel.value === 'desa') {
    filterLevel.value = 'kecamatan'
    pendingDesa.value = ''
  } else if (filterLevel.value === 'kecamatan') {
    filterLevel.value = 'kotakab'
    pendingKecamatan.value = ''
  } else if (filterLevel.value === 'kotakab') {
    filterLevel.value = 'provinsi'
    pendingKotaKab.value = ''
  }
}

// Calculate 30 days ago date
const getThirtyDaysAgo = (): string => {
  const date = new Date()
  date.setDate(date.getDate() - 30)
  return date.toISOString()
}

// Category and type colors
const categoryColors: Record<string, 'default' | 'success' | 'warning' | 'danger' | 'outline'> = {
  kebutuhan: 'warning',
  informasi: 'outline',
  'follow-up': 'danger',
  'info_bantuan': 'success',
}

const typeColors: Record<string, 'default' | 'success' | 'warning' | 'danger' | 'outline'> = {
  'sar': 'danger',
  'ambulan': 'danger',
  'medis': 'success',
  'transport_roda4': 'default',
  'transport_roda2': 'default',
  'air_bersih': 'outline',
  'sembako': 'warning',
  'psikososial': 'outline',
  'sekolah_darurat': 'outline',
  'dapur_umum': 'outline',
  'keamanan': 'danger',
  'listrik': 'outline',
  'internet': 'outline',
  'sinyal_selular': 'outline',
}

// Format timestamp
const formatTimestamp = (isoString: string): string => {
  const date = new Date(isoString)
  const day = date.getDate().toString().padStart(2, '0')
  const month = (date.getMonth() + 1).toString().padStart(2, '0')
  const year = date.getFullYear()
  const hours = date.getHours().toString().padStart(2, '0')
  const minutes = date.getMinutes().toString().padStart(2, '0')
  return `${day}-${month}-${year} ${hours}:${minutes}`
}

// Format tag display
const formatTagDisplay = (tag: string): string => {
  const tagMap: Record<string, string> = {
    'sar': 'SAR',
    'ambulan': 'Ambulan',
    'medis': 'Medis',
    'transport_roda4': 'Transport Roda 4',
    'transport_roda2': 'Transport Roda 2',
    'air_bersih': 'Air Bersih',
    'sembako': 'Sembako',
    'psikososial': 'Psikososial',
    'sekolah_darurat': 'Sekolah Darurat',
    'dapur_umum': 'Dapur Umum',
    'keamanan': 'Keamanan',
    'listrik': 'Listrik',
    'internet': 'Internet',
    'sinyal_selular': 'Sinyal Selular',
  }
  return tagMap[tag] || tag
}

// Formatted feeds
const formattedFeeds = computed(() => {
  return feeds.value.map(feed => ({
    id: feed.id,
    timestamp: formatTimestamp(feed.submitted_at),
    username: feed.username ?? 'anonymous',
    organization: feed.organization ?? '',
    location: feed.location_name ?? '',
    locationId: feed.location_id,
    faskesName: feed.faskes_name ?? '',
    faskesId: feed.faskes_id,
    content: feed.content,
    category: feed.category,
    type: feed.type ?? '',
    coordinates: feed.coordinates,
    photos: feed.photos ?? [],
  }))
})

// Get photo URL helper - use direct photo ID approach like MapView
const getPhotoUrl = (photo: FeedPhoto) => {
  const baseUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1'
  return `${baseUrl}/feeds/photos/${photo.id}/file`
}

// Navigate to map with coordinates and show popup
const goToMapWithFeed = (feed: { id: string, coordinates?: [number, number], locationId?: string, faskesId?: string }) => {
  // Priority: coordinates > location > faskes
  if (feed.coordinates) {
    router.push({
      path: '/',
      query: {
        lat: feed.coordinates[1].toString(),
        lng: feed.coordinates[0].toString(),
        zoom: '18',
        feed: feed.id
      }
    })
  } else if (feed.locationId) {
    router.push({ path: '/', query: { location: feed.locationId } })
  } else if (feed.faskesId) {
    router.push({ path: '/', query: { faskes: feed.faskesId } })
  }
}

// Has more data to load
const hasMore = computed(() => feeds.value.length < total.value)

// Fetch feeds
const fetchFeeds = async (reset = true) => {
  if (reset) {
    loading.value = true
    page.value = 1
    feeds.value = []
  } else {
    loadingMore.value = true
  }

  try {
    const filter: FeedFilter = {
      page: page.value,
      limit,
      since: getThirtyDaysAgo(), // Always filter last 30 days
    }

    // Check if search is UUID or name
    if (searchQuery.value) {
      const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i
      if (uuidRegex.test(searchQuery.value)) {
        filter.location_id = searchQuery.value
      } else {
        filter.location_name = searchQuery.value
      }
    }
    if (selectedCategory.value) {
      filter.category = selectedCategory.value
    }
    if (selectedTag.value) {
      filter.type = selectedTag.value
    }
    // Region filters
    if (appliedProvinsi.value) {
      filter.provinsi = appliedProvinsi.value
    }
    if (appliedKotaKab.value) {
      filter.kota_kab = appliedKotaKab.value
    }
    if (appliedKecamatan.value) {
      filter.kecamatan = appliedKecamatan.value
    }
    if (appliedDesa.value) {
      filter.desa = appliedDesa.value
    }

    const response = await api.getFeeds(filter)
    if (response.success && response.data) {
      if (reset) {
        feeds.value = response.data
      } else {
        feeds.value = [...feeds.value, ...response.data]
      }
      total.value = response.meta?.total ?? response.data.length
    }
  } catch (e) {
    console.error('Failed to fetch feeds:', e)
  } finally {
    loading.value = false
    loadingMore.value = false
  }
}

// Load more
const loadMore = () => {
  page.value++
  fetchFeeds(false)
}

// Handle search
const handleSearch = () => {
  fetchFeeds()
}

// Watch filter changes
watch([selectedCategory, selectedTag], () => {
  fetchFeeds()
})

// Initialize from route query
onMounted(() => {
  if (route.query.search) {
    searchQuery.value = route.query.search as string
  }
  if (route.query.category) {
    selectedCategory.value = route.query.category as string
  }
  if (route.query.tag) {
    selectedTag.value = route.query.tag as string
  }
  fetchFeeds()
})

// Predefined tags list for filter
const allTags = [
  { value: 'sar', label: 'SAR' },
  { value: 'ambulan', label: 'Ambulan' },
  { value: 'medis', label: 'Medis' },
  { value: 'transport_roda4', label: 'Transport Roda 4' },
  { value: 'transport_roda2', label: 'Transport Roda 2' },
  { value: 'air_bersih', label: 'Air Bersih' },
  { value: 'sembako', label: 'Sembako' },
  { value: 'psikososial', label: 'Psikososial' },
  { value: 'sekolah_darurat', label: 'Sekolah Darurat' },
  { value: 'dapur_umum', label: 'Dapur Umum' },
  { value: 'keamanan', label: 'Keamanan' },
  { value: 'listrik', label: 'Listrik' },
  { value: 'internet', label: 'Internet' },
  { value: 'sinyal_selular', label: 'Sinyal Selular' },
  { value: 'sanitasi_mck', label: 'Sanitasi MCK' },
]

// Predefined categories
const allCategories = [
  { value: 'informasi', label: 'Informasi' },
  { value: 'kebutuhan', label: 'Kebutuhan' },
  { value: 'follow-up', label: 'Follow-up' },
  { value: 'info_bantuan', label: 'Info Bantuan' },
]
</script>

<template>
  <div class="flex-1 flex overflow-hidden">
    <DataLayersSidebar />

    <!-- Feeds Content -->
    <main class="flex-1 bg-gray-50 flex flex-col overflow-hidden">
      <!-- Header -->
      <div class="bg-white border-b border-gray-200 px-3 md:px-6 py-3 md:py-4">
        <div class="max-w-4xl mx-auto">
          <div class="flex items-center justify-between mb-3 md:mb-4">
            <div>
              <h1 class="text-lg md:text-2xl font-bold text-gray-900">Informasi Terbaru</h1>
              <div class="flex items-center gap-2 mt-1">
                <span class="relative flex h-2 w-2 md:h-2.5 md:w-2.5">
                  <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
                  <span class="relative inline-flex rounded-full h-2 w-2 md:h-2.5 md:w-2.5 bg-green-500"></span>
                </span>
                <span class="text-xs md:text-sm text-gray-500">{{ total }} updates</span>
              </div>
            </div>
            <Button variant="outline" class="gap-2 hidden md:flex">
              <Download class="w-4 h-4" />
              Export
            </Button>
          </div>

          <!-- Filters Row - Stack on mobile -->
          <div class="flex flex-col md:flex-row gap-2 md:gap-3">
            <!-- Search by Location Name -->
            <div class="flex-1 relative">
              <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
              <Input
                v-model="searchQuery"
                placeholder="Cari nama posko..."
                class="pl-9 w-full"
                @keyup.enter="handleSearch"
              />
            </div>

            <!-- Filter selects in row on mobile -->
            <div class="flex gap-2">
              <!-- Category Filter -->
              <div class="flex-1 md:w-48 md:flex-initial">
                <select
                  v-model="selectedCategory"
                  class="w-full h-10 text-sm border border-gray-200 rounded-lg px-2 md:px-3 bg-white focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                >
                  <option value="">Kategori</option>
                  <option v-for="cat in allCategories" :key="cat.value" :value="cat.value">
                    {{ cat.label }}
                  </option>
                </select>
              </div>

              <!-- Tags Filter (Single Select) -->
              <div class="flex-1 md:w-48 md:flex-initial">
                <select
                  v-model="selectedTag"
                  class="w-full h-10 text-sm border border-gray-200 rounded-lg px-2 md:px-3 bg-white focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                >
                  <option value="">Tags</option>
                  <option v-for="tag in allTags" :key="tag.value" :value="tag.value">
                    {{ tag.label }}
                  </option>
                </select>
              </div>

              <!-- Region Cascade Filter -->
              <div class="relative flex-1 md:w-48 md:flex-initial">
                <button
                  class="w-full h-10 text-sm border border-gray-200 rounded-lg px-2 md:px-3 bg-white flex items-center justify-between hover:bg-gray-50"
                  :class="{ 'ring-2 ring-blue-500 border-transparent': showRegionFilter, 'text-blue-600': hasActiveRegionFilter }"
                  @click="showRegionFilter = !showRegionFilter"
                >
                  <span class="flex items-center gap-1.5 truncate">
                    <MapPin class="w-4 h-4 flex-shrink-0" />
                    <span class="truncate">{{ regionFilterLabel }}</span>
                  </span>
                  <div class="flex items-center gap-1">
                    <button
                      v-if="hasActiveRegionFilter"
                      class="p-0.5 hover:bg-gray-200 rounded"
                      @click.stop="clearRegionFilter"
                    >
                      <X class="w-3 h-3 text-gray-500" />
                    </button>
                    <ChevronDown class="w-4 h-4 text-gray-400" />
                  </div>
                </button>

                <!-- Dropdown Menu -->
                <div
                  v-if="showRegionFilter"
                  class="absolute top-full left-0 mt-1 bg-white rounded-lg shadow-lg border border-gray-200 min-w-[250px] max-h-[400px] overflow-hidden flex flex-col z-50"
                >
                  <!-- Breadcrumb -->
                  <div v-if="filterLevel !== 'provinsi'" class="px-3 py-2 bg-gray-50 border-b border-gray-200">
                    <div class="flex items-center gap-1 text-xs text-gray-500 flex-wrap">
                      <button class="hover:text-blue-600" @click="filterLevel = 'provinsi'; pendingKotaKab = ''; pendingKecamatan = ''; pendingDesa = ''">
                        {{ pendingProvinsi || 'Provinsi' }}
                      </button>
                      <span v-if="pendingKotaKab">›</span>
                      <button v-if="pendingKotaKab" class="hover:text-blue-600" @click="filterLevel = 'kecamatan'; pendingKecamatan = ''; pendingDesa = ''">
                        {{ pendingKotaKab }}
                      </button>
                      <span v-if="pendingKecamatan">›</span>
                      <button v-if="pendingKecamatan" class="hover:text-blue-600" @click="filterLevel = 'desa'; pendingDesa = ''">
                        {{ pendingKecamatan }}
                      </button>
                      <span v-if="pendingDesa">›</span>
                      <span v-if="pendingDesa" class="font-medium text-gray-700">{{ pendingDesa }}</span>
                    </div>
                  </div>

                  <!-- Header with Back button -->
                  <div class="px-3 py-2 border-b border-gray-200 flex items-center justify-between">
                    <span class="text-xs font-medium text-gray-500 uppercase">
                      {{ filterLevel === 'provinsi' ? 'Provinsi' : filterLevel === 'kotakab' ? 'Kota/Kabupaten' : filterLevel === 'kecamatan' ? 'Kecamatan' : 'Desa' }}
                    </span>
                    <button
                      v-if="filterLevel !== 'provinsi'"
                      class="text-blue-500 hover:text-blue-700 text-xs"
                      @click="goBackLevel"
                    >
                      ← Kembali
                    </button>
                  </div>

                  <!-- List Content -->
                  <div class="max-h-[250px] overflow-y-auto flex-1">
                    <!-- Provinsi Selection -->
                    <template v-if="filterLevel === 'provinsi'">
                      <button
                        class="w-full px-3 py-2 text-left text-sm hover:bg-blue-50 hover:text-blue-600 flex items-center gap-2"
                        :class="{ 'bg-blue-50 text-blue-600 font-medium': pendingProvinsi === 'Aceh' }"
                        @click="selectProvinsi('Aceh')"
                      >
                        <MapPin class="w-4 h-4" />
                        Aceh
                      </button>
                    </template>

                    <!-- Kota/Kab Selection -->
                    <template v-else-if="filterLevel === 'kotakab'">
                      <button
                        class="w-full px-3 py-2 text-left text-sm hover:bg-blue-50 hover:text-blue-600"
                        :class="{ 'bg-blue-50 text-blue-600 font-medium': pendingKotaKab === '' }"
                        @click="pendingKotaKab = ''; pendingKecamatan = ''; pendingDesa = ''"
                      >
                        Semua Kota/Kabupaten
                      </button>
                      <button
                        v-for="kota in availableKotaKab"
                        :key="kota"
                        class="w-full px-3 py-2 text-left text-sm hover:bg-blue-50 hover:text-blue-600"
                        :class="{ 'bg-blue-50 text-blue-600 font-medium': pendingKotaKab === kota }"
                        @click="selectKotaKab(kota)"
                      >
                        {{ kota }}
                      </button>
                      <div v-if="availableKotaKab.length === 0" class="px-3 py-4 text-sm text-gray-400 text-center">
                        Tidak ada data
                      </div>
                    </template>

                    <!-- Kecamatan Selection -->
                    <template v-else-if="filterLevel === 'kecamatan'">
                      <button
                        class="w-full px-3 py-2 text-left text-sm hover:bg-blue-50 hover:text-blue-600"
                        :class="{ 'bg-blue-50 text-blue-600 font-medium': pendingKecamatan === '' }"
                        @click="pendingKecamatan = ''; pendingDesa = ''"
                      >
                        Semua Kecamatan
                      </button>
                      <button
                        v-for="kec in availableKecamatan"
                        :key="kec"
                        class="w-full px-3 py-2 text-left text-sm hover:bg-blue-50 hover:text-blue-600"
                        :class="{ 'bg-blue-50 text-blue-600 font-medium': pendingKecamatan === kec }"
                        @click="selectKecamatan(kec)"
                      >
                        {{ kec }}
                      </button>
                      <div v-if="availableKecamatan.length === 0" class="px-3 py-4 text-sm text-gray-400 text-center">
                        Tidak ada data kecamatan
                      </div>
                    </template>

                    <!-- Desa Selection -->
                    <template v-else-if="filterLevel === 'desa'">
                      <button
                        class="w-full px-3 py-2 text-left text-sm hover:bg-blue-50 hover:text-blue-600"
                        :class="{ 'bg-blue-50 text-blue-600 font-medium': pendingDesa === '' }"
                        @click="pendingDesa = ''"
                      >
                        Semua Desa
                      </button>
                      <button
                        v-for="desa in availableDesa"
                        :key="desa"
                        class="w-full px-3 py-2 text-left text-sm hover:bg-blue-50 hover:text-blue-600"
                        :class="{ 'bg-blue-50 text-blue-600 font-medium': pendingDesa === desa }"
                        @click="selectDesa(desa)"
                      >
                        {{ desa }}
                      </button>
                      <div v-if="availableDesa.length === 0" class="px-3 py-4 text-sm text-gray-400 text-center">
                        Tidak ada data desa
                      </div>
                    </template>
                  </div>

                  <!-- Apply Button -->
                  <div class="px-3 py-2 border-t border-gray-200 bg-gray-50">
                    <button
                      class="w-full px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors"
                      @click="applyRegionFilter"
                    >
                      Terapkan Filter
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Active Filters Display -->
          <div v-if="searchQuery || selectedCategory || selectedTag || hasActiveRegionFilter" class="flex flex-wrap gap-2 mt-3">
            <Badge
              v-if="searchQuery"
              variant="default"
              class="cursor-pointer hover:bg-gray-200"
              @click="searchQuery = ''; handleSearch()"
            >
              Lokasi: {{ searchQuery }}
              <span class="ml-1">&times;</span>
            </Badge>
            <Badge
              v-if="selectedCategory"
              variant="default"
              class="cursor-pointer hover:bg-gray-200"
              @click="selectedCategory = ''"
            >
              Kategori: {{ allCategories.find(c => c.value === selectedCategory)?.label }}
              <span class="ml-1">&times;</span>
            </Badge>
            <Badge
              v-if="selectedTag"
              variant="default"
              class="cursor-pointer hover:bg-gray-200"
              @click="selectedTag = ''"
            >
              Tag: {{ formatTagDisplay(selectedTag) }}
              <span class="ml-1">&times;</span>
            </Badge>
            <Badge
              v-if="hasActiveRegionFilter"
              variant="default"
              class="cursor-pointer hover:bg-gray-200"
              @click="clearRegionFilter"
            >
              Wilayah: {{ regionFilterLabel }}
              <span class="ml-1">&times;</span>
            </Badge>
          </div>
        </div>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="flex-1 flex items-center justify-center">
        <div class="text-gray-500">Memuat data...</div>
      </div>

      <!-- Feeds List -->
      <div v-else class="flex-1 overflow-y-auto">
        <div class="max-w-4xl mx-auto py-3 md:py-4 px-3 md:px-6">
          <!-- Empty State -->
          <div v-if="formattedFeeds.length === 0" class="text-center py-12">
            <Filter class="w-12 h-12 text-gray-300 mx-auto mb-4" />
            <p class="text-gray-500">Tidak ada update yang ditemukan</p>
            <p class="text-sm text-gray-400 mt-1">Coba ubah filter atau kata kunci pencarian</p>
          </div>

          <!-- Feed Items -->
          <div class="space-y-2 md:space-y-3">
            <div
              v-for="update in formattedFeeds"
              :key="update.id"
              class="bg-white rounded-lg border border-gray-200 p-3 md:p-4 hover:shadow-md hover:border-blue-300 transition-all cursor-pointer"
              @click="goToMapWithFeed({ id: update.id, coordinates: update.coordinates, locationId: update.locationId, faskesId: update.faskesId })"
            >
              <div class="flex gap-3">
                <!-- Photo on the left -->
                <div v-if="update.photos.length > 0" class="flex-shrink-0">
                  <div class="relative">
                    <img
                      :src="getPhotoUrl(update.photos[0])"
                      :alt="update.photos[0].filename"
                      class="w-24 h-24 md:w-32 md:h-32 object-cover rounded-lg border border-gray-200"
                      loading="lazy"
                    />
                    <div
                      v-if="update.photos.length > 1"
                      class="absolute bottom-1 right-1 bg-black/60 text-white text-xs px-1.5 py-0.5 rounded flex items-center gap-1"
                    >
                      <Image class="w-3 h-3" />
                      <span>+{{ update.photos.length - 1 }}</span>
                    </div>
                  </div>
                </div>

                <!-- Content on the right -->
                <div class="flex-1 min-w-0">
                  <div class="flex items-center justify-between mb-1">
                    <span class="text-xs text-gray-500">{{ update.timestamp }}</span>
                    <!-- Navigate to map indicator -->
                    <div
                      v-if="update.coordinates || update.locationId || update.faskesId"
                      class="flex items-center gap-1 text-xs text-blue-500"
                      title="Klik untuk lihat di peta"
                    >
                      <Navigation class="w-3.5 h-3.5" />
                      <span class="hidden sm:inline">Peta</span>
                    </div>
                  </div>
                  <div class="text-xs text-blue-600 font-medium mb-1">
                    {{ update.username }}{{ update.organization ? ` - ${update.organization}` : '' }}
                  </div>
                  <div class="flex items-center gap-1.5 mb-1">
                    <!-- Show different icon based on whether it's related to a location/faskes -->
                    <template v-if="update.locationId">
                      <MapPin class="w-4 h-4 text-blue-500 flex-shrink-0" />
                      <span class="text-sm font-medium text-blue-600 truncate">{{ update.location }}</span>
                    </template>
                    <template v-else-if="update.faskesId">
                      <MapPin class="w-4 h-4 text-green-500 flex-shrink-0" />
                      <span class="text-sm font-medium text-green-600 truncate">{{ update.faskesName }}</span>
                      <Badge variant="success" class="ml-1 text-xs">Faskes</Badge>
                    </template>
                    <template v-else-if="update.coordinates">
                      <Megaphone class="w-4 h-4 text-orange-500 flex-shrink-0" />
                      <span class="text-sm font-medium text-orange-600">Laporan Situasi</span>
                    </template>
                  </div>

                  <p class="text-sm text-gray-600 mb-2 leading-relaxed line-clamp-2">{{ update.content }}</p>
                  <div class="flex flex-wrap gap-1.5">
                    <Badge :variant="categoryColors[update.category] || 'outline'" class="text-xs">
                      {{ update.category }}
                    </Badge>
                    <template v-if="update.type">
                      <Badge
                        v-for="t in update.type.split(',')"
                        :key="t"
                        :variant="typeColors[t.trim()] || 'outline'"
                        class="text-xs"
                      >
                        {{ formatTagDisplay(t.trim()) }}
                      </Badge>
                    </template>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Load More -->
          <div v-if="hasMore" class="py-6 text-center">
            <Button
              variant="outline"
              :disabled="loadingMore"
              @click="loadMore"
            >
              {{ loadingMore ? 'Memuat...' : 'Muat Lebih Banyak' }}
            </Button>
          </div>

          <!-- End of list -->
          <div v-else-if="formattedFeeds.length > 0" class="py-6 text-center text-sm text-gray-400">
            Semua update telah ditampilkan
          </div>
        </div>
      </div>
    </main>
  </div>
</template>
