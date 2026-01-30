<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted, computed, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Search, MapPin, Download, Filter, Image, Megaphone, ChevronDown, ChevronUp, X, RefreshCw } from 'lucide-vue-next'
import DataLayersSidebar from '@/components/DataLayersSidebar.vue'
import Input from '@/components/ui/Input.vue'
import Badge from '@/components/ui/Badge.vue'
import Button from '@/components/ui/Button.vue'
import { api, type Feed, type FeedFilter, type FeedPhoto } from '@/services/api'
import { categoryColors, tagColor, getCategoryLabel, getTagLabel } from '@/lib/feedHelpers'

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

// Mobile filter collapse state
const showFilters = ref(true)
const isMobile = ref(false)

// Pull-to-refresh state
const scrollContainerRef = ref<HTMLElement | null>(null)
const pullStartY = ref(0)
const pullDistance = ref(0)
const isPulling = ref(false)
const isRefreshing = ref(false)
const pullThreshold = 80

// Infinite scroll
const loadMoreTriggerRef = ref<HTMLElement | null>(null)
const infiniteScrollObserver = ref<IntersectionObserver | null>(null)

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

// Check if mobile
const checkMobile = () => {
  isMobile.value = window.innerWidth < 768
  // On desktop, always show filters
  if (!isMobile.value) {
    showFilters.value = true
  }
}

// Pull-to-refresh handlers
const onTouchStart = (e: TouchEvent) => {
  if (!scrollContainerRef.value) return
  // Only enable pull-to-refresh when scrolled to top
  if (scrollContainerRef.value.scrollTop <= 0) {
    pullStartY.value = e.touches[0].clientY
    isPulling.value = true
  }
}

const onTouchMove = (e: TouchEvent) => {
  if (!isPulling.value || isRefreshing.value) return
  if (!scrollContainerRef.value || scrollContainerRef.value.scrollTop > 0) {
    isPulling.value = false
    pullDistance.value = 0
    return
  }
  
  const currentY = e.touches[0].clientY
  const diff = currentY - pullStartY.value
  
  if (diff > 0) {
    // Dampen the pull distance
    pullDistance.value = Math.min(diff * 0.5, pullThreshold * 1.5)
    e.preventDefault()
  }
}

const onTouchEnd = async () => {
  if (!isPulling.value) return
  isPulling.value = false
  
  if (pullDistance.value >= pullThreshold && !isRefreshing.value) {
    isRefreshing.value = true
    pullDistance.value = pullThreshold
    
    // Refresh data
    await fetchFeeds()
    
    isRefreshing.value = false
  }
  
  pullDistance.value = 0
}

// Setup infinite scroll observer
const setupInfiniteScroll = () => {
  if (infiniteScrollObserver.value) {
    infiniteScrollObserver.value.disconnect()
  }
  
  infiniteScrollObserver.value = new IntersectionObserver(
    (entries) => {
      if (entries[0].isIntersecting && hasMore.value && !loadingMore.value && !loading.value) {
        loadMore()
      }
    },
    { rootMargin: '200px' }
  )
  
  nextTick(() => {
    if (loadMoreTriggerRef.value) {
      infiniteScrollObserver.value?.observe(loadMoreTriggerRef.value)
    }
  })
}

onMounted(() => {
  checkMobile()
  window.addEventListener('resize', checkMobile)
  
  // Initialize from route query
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
  setupInfiniteScroll()
})

onUnmounted(() => {
  window.removeEventListener('resize', checkMobile)
  infiniteScrollObserver.value?.disconnect()
})

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

// Count active filters
const activeFilterCount = computed(() => {
  let count = 0
  if (searchQuery.value) count++
  if (selectedCategory.value) count++
  if (selectedTag.value) count++
  if (hasActiveRegionFilter.value) count++
  return count
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



const formattedFeeds = computed(() => {
  return feeds.value.map(feed => ({
    id: feed.id,
    shortCode: feed.short_code,
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
    tags: feed.type ? feed.type.split(',').map(t => t.trim()).filter(t => t) : [],
    coordinates: feed.coordinates,
    photos: feed.photos ?? [],
    desaId: feed.region?.id_desa,
    desaName: feed.region?.desa,
    kecamatan: feed.region?.kecamatan,
    kotaKab: feed.region?.kota_kab,
    provinsi: feed.region?.provinsi,
    submitter: `${feed.username ?? 'anonymous'}${feed.organization ? ` - ${feed.organization}` : ''}`,
  }))
})

const getPhotoUrl = (photo: FeedPhoto) => {
  const baseUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1'
  return `${baseUrl}/feed-photos/${photo.id}/file`
}

const goToFeedDetail = (shortCode: string | undefined, feedId: string) => {
  const code = shortCode || feedId
  router.push({ path: `/feeds/${code}` })
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
      since: getThirtyDaysAgo(),
    }

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
  if (loadingMore.value || loading.value) return
  page.value++
  fetchFeeds(false)
}

// Handle search
const handleSearch = () => {
  fetchFeeds()
}

// Watch filter changes
const isRouteSyncing = ref(false)

watch([selectedCategory, selectedTag], () => {
  if (!isRouteSyncing.value) {
    fetchFeeds()
  }
})

// Watch route query changes
watch(() => route.query, (newQuery) => {
  isRouteSyncing.value = true

  const newCategory = (newQuery.category as string) || ''
  const newTag = (newQuery.tag as string) || ''
  const newSearch = (newQuery.search as string) || ''

  const categoryChanged = selectedCategory.value !== newCategory
  const tagChanged = selectedTag.value !== newTag
  const searchChanged = searchQuery.value !== newSearch

  if (categoryChanged || tagChanged || searchChanged) {
    selectedCategory.value = newCategory
    selectedTag.value = newTag
    searchQuery.value = newSearch
    fetchFeeds()
  }

  isRouteSyncing.value = false
}, { immediate: false })

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
  { value: 'lainnya', label: 'Lainnya' },
]

// Predefined categories
const allCategories = [
  { value: 'informasi', label: 'Informasi' },
  { value: 'kebutuhan', label: 'Butuh Bantuan' },
  { value: 'follow-up', label: 'Follow-up' },
  { value: 'info_bantuan', label: 'Terima Bantuan' },
]

// Toggle filters visibility on mobile
const toggleFilters = () => {
  if (isMobile.value) {
    showFilters.value = !showFilters.value
  }
}
</script>

<template>
  <div class="flex-1 flex overflow-hidden">
    <DataLayersSidebar />

    <!-- Feeds Content -->
    <main class="flex-1 bg-gray-50 flex flex-col overflow-hidden">
      <!-- Header - Sticky -->
      <div class="bg-white border-b border-gray-200 sticky top-0 z-10">
        <div class="px-3 md:px-6 py-3 md:py-4">
          <div class="max-w-4xl mx-auto">
            <!-- Title row -->
            <div class="flex items-center justify-between mb-3">
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
              <div class="flex items-center gap-2">
                <!-- Mobile filter toggle -->
                <button
                  v-if="isMobile"
                  @click="toggleFilters"
                  class="flex items-center gap-1.5 px-3 py-2 text-sm font-medium rounded-lg border border-gray-200 bg-white hover:bg-gray-50 active:bg-gray-100 transition-colors"
                  :class="{ 'text-blue-600 border-blue-300 bg-blue-50': activeFilterCount > 0 }"
                >
                  <Filter class="w-4 h-4" />
                  <span>Filter</span>
                  <span v-if="activeFilterCount > 0" class="w-5 h-5 rounded-full bg-blue-600 text-white text-xs flex items-center justify-center">
                    {{ activeFilterCount }}
                  </span>
                  <component :is="showFilters ? ChevronUp : ChevronDown" class="w-4 h-4 text-gray-400" />
                </button>
                <Button variant="outline" class="gap-2 hidden md:flex">
                  <Download class="w-4 h-4" />
                  Export
                </Button>
              </div>
            </div>

            <!-- Collapsible Filters -->
            <Transition
              enter-active-class="transition-all duration-200 ease-out"
              enter-from-class="opacity-0 max-h-0"
              enter-to-class="opacity-100 max-h-[500px]"
              leave-active-class="transition-all duration-200 ease-in"
              leave-from-class="opacity-100 max-h-[500px]"
              leave-to-class="opacity-0 max-h-0"
            >
              <div v-show="showFilters" class="overflow-hidden">
                <!-- Filters Row -->
                <div class="flex flex-col md:flex-row gap-2 md:gap-3">
                  <!-- Search by Location Name -->
                  <div class="flex-1 relative">
                    <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
                    <Input
                      v-model="searchQuery"
                      placeholder="Cari nama posko..."
                      class="pl-9 w-full h-11 md:h-10"
                      @keyup.enter="handleSearch"
                    />
                  </div>

                  <!-- Filter selects -->
                  <div class="flex gap-2">
                    <!-- Category Filter -->
                    <div class="flex-1 md:w-40 md:flex-initial">
                      <select
                        v-model="selectedCategory"
                        class="w-full h-11 md:h-10 text-sm border border-gray-200 rounded-lg px-3 bg-white focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      >
                        <option value="">Kategori</option>
                        <option v-for="cat in allCategories" :key="cat.value" :value="cat.value">
                          {{ cat.label }}
                        </option>
                      </select>
                    </div>

                    <!-- Tags Filter -->
                    <div class="flex-1 md:w-40 md:flex-initial">
                      <select
                        v-model="selectedTag"
                        class="w-full h-11 md:h-10 text-sm border border-gray-200 rounded-lg px-3 bg-white focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      >
                        <option value="">Tags</option>
                        <option v-for="tag in allTags" :key="tag.value" :value="tag.value">
                          {{ tag.label }}
                        </option>
                      </select>
                    </div>

                    <!-- Region Cascade Filter -->
                    <div class="relative flex-1 md:w-40 md:flex-initial">
                      <button
                        class="w-full h-11 md:h-10 text-sm border border-gray-200 rounded-lg px-3 bg-white flex items-center justify-between hover:bg-gray-50 active:bg-gray-100 transition-colors"
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

                      <!-- Region Dropdown Menu -->
                      <Teleport to="body">
                        <div
                          v-if="showRegionFilter"
                          class="fixed inset-0 z-[9998]"
                          @click="showRegionFilter = false"
                        ></div>
                      </Teleport>
                      <div
                        v-if="showRegionFilter"
                        class="absolute top-full left-0 right-0 md:left-auto md:right-auto mt-1 bg-white rounded-lg shadow-lg border border-gray-200 min-w-[280px] max-h-[400px] overflow-hidden flex flex-col z-[9999]"
                      >
                        <!-- Breadcrumb -->
                        <div v-if="filterLevel !== 'provinsi'" class="px-3 py-2 bg-gray-50 border-b border-gray-200">
                          <div class="flex items-center gap-1 text-xs text-gray-500 flex-wrap">
                            <button class="hover:text-blue-600 active:text-blue-700" @click="filterLevel = 'provinsi'; pendingKotaKab = ''; pendingKecamatan = ''; pendingDesa = ''">
                              {{ pendingProvinsi || 'Provinsi' }}
                            </button>
                            <span v-if="pendingKotaKab">›</span>
                            <button v-if="pendingKotaKab" class="hover:text-blue-600 active:text-blue-700" @click="filterLevel = 'kecamatan'; pendingKecamatan = ''; pendingDesa = ''">
                              {{ pendingKotaKab }}
                            </button>
                            <span v-if="pendingKecamatan">›</span>
                            <button v-if="pendingKecamatan" class="hover:text-blue-600 active:text-blue-700" @click="filterLevel = 'desa'; pendingDesa = ''">
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
                            class="text-blue-500 hover:text-blue-700 active:text-blue-800 text-xs font-medium"
                            @click="goBackLevel"
                          >
                            ← Kembali
                          </button>
                        </div>

                        <!-- List Content -->
                        <div class="max-h-[250px] overflow-y-auto flex-1 overscroll-contain">
                          <!-- Provinsi Selection -->
                          <template v-if="filterLevel === 'provinsi'">
                            <button
                              class="w-full px-3 py-3 text-left text-sm hover:bg-blue-50 active:bg-blue-100 hover:text-blue-600 flex items-center gap-2 transition-colors"
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
                              class="w-full px-3 py-3 text-left text-sm hover:bg-blue-50 active:bg-blue-100 hover:text-blue-600 transition-colors"
                              :class="{ 'bg-blue-50 text-blue-600 font-medium': pendingKotaKab === '' }"
                              @click="pendingKotaKab = ''; pendingKecamatan = ''; pendingDesa = ''"
                            >
                              Semua Kota/Kabupaten
                            </button>
                            <button
                              v-for="kota in availableKotaKab"
                              :key="kota"
                              class="w-full px-3 py-3 text-left text-sm hover:bg-blue-50 active:bg-blue-100 hover:text-blue-600 transition-colors"
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
                              class="w-full px-3 py-3 text-left text-sm hover:bg-blue-50 active:bg-blue-100 hover:text-blue-600 transition-colors"
                              :class="{ 'bg-blue-50 text-blue-600 font-medium': pendingKecamatan === '' }"
                              @click="pendingKecamatan = ''; pendingDesa = ''"
                            >
                              Semua Kecamatan
                            </button>
                            <button
                              v-for="kec in availableKecamatan"
                              :key="kec"
                              class="w-full px-3 py-3 text-left text-sm hover:bg-blue-50 active:bg-blue-100 hover:text-blue-600 transition-colors"
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
                              class="w-full px-3 py-3 text-left text-sm hover:bg-blue-50 active:bg-blue-100 hover:text-blue-600 transition-colors"
                              :class="{ 'bg-blue-50 text-blue-600 font-medium': pendingDesa === '' }"
                              @click="pendingDesa = ''"
                            >
                              Semua Desa
                            </button>
                            <button
                              v-for="desa in availableDesa"
                              :key="desa"
                              class="w-full px-3 py-3 text-left text-sm hover:bg-blue-50 active:bg-blue-100 hover:text-blue-600 transition-colors"
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
                        <div class="px-3 py-3 border-t border-gray-200 bg-gray-50">
                          <button
                            class="w-full px-4 py-3 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 active:bg-blue-800 transition-colors"
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
                <div v-if="activeFilterCount > 0" class="flex flex-wrap gap-2 mt-3">
                  <Badge
                    v-if="searchQuery"
                    variant="default"
                    class="cursor-pointer hover:bg-gray-200 active:bg-gray-300 transition-colors"
                    @click="searchQuery = ''; handleSearch()"
                  >
                    Lokasi: {{ searchQuery }}
                    <span class="ml-1">&times;</span>
                  </Badge>
                  <Badge
                    v-if="selectedCategory"
                    variant="default"
                    class="cursor-pointer hover:bg-gray-200 active:bg-gray-300 transition-colors"
                    @click="selectedCategory = ''"
                  >
                    Kategori: {{ allCategories.find(c => c.value === selectedCategory)?.label }}
                    <span class="ml-1">&times;</span>
                  </Badge>
                  <Badge
                    v-if="selectedTag"
                    variant="default"
                    class="cursor-pointer hover:bg-gray-200 active:bg-gray-300 transition-colors"
                    @click="selectedTag = ''"
                  >
                    Tag: {{ getTagLabel(selectedTag) }}
                    <span class="ml-1">&times;</span>
                  </Badge>
                  <Badge
                    v-if="hasActiveRegionFilter"
                    variant="default"
                    class="cursor-pointer hover:bg-gray-200 active:bg-gray-300 transition-colors"
                    @click="clearRegionFilter"
                  >
                    Wilayah: {{ regionFilterLabel }}
                    <span class="ml-1">&times;</span>
                  </Badge>
                </div>
              </div>
            </Transition>
          </div>
        </div>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="flex-1 flex items-center justify-center">
        <div class="flex flex-col items-center gap-3">
          <div class="w-8 h-8 border-3 border-gray-300 border-t-blue-500 rounded-full animate-spin"></div>
          <span class="text-gray-500 text-sm">Memuat data...</span>
        </div>
      </div>

      <!-- Feeds List with Pull-to-refresh -->
      <div
        v-else
        ref="scrollContainerRef"
        class="flex-1 overflow-y-auto overscroll-contain"
        @touchstart.passive="onTouchStart"
        @touchmove="onTouchMove"
        @touchend.passive="onTouchEnd"
      >
        <!-- Pull-to-refresh indicator -->
        <div
          class="flex items-center justify-center transition-all duration-200 overflow-hidden"
          :style="{ height: `${pullDistance}px` }"
        >
          <div v-if="pullDistance > 0" class="flex items-center gap-2 text-gray-500">
            <RefreshCw
              class="w-5 h-5 transition-transform duration-200"
              :class="{ 'animate-spin': isRefreshing }"
              :style="{ transform: isRefreshing ? '' : `rotate(${pullDistance * 2}deg)` }"
            />
            <span class="text-sm">
              {{ isRefreshing ? 'Memperbarui...' : pullDistance >= pullThreshold ? 'Lepaskan untuk refresh' : 'Tarik untuk refresh' }}
            </span>
          </div>
        </div>

        <div class="max-w-4xl mx-auto py-3 md:py-4 px-3 md:px-6 pb-safe">
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
              :id="`feed-item-${update.id}`"
              class="bg-white rounded-lg border border-gray-200 overflow-hidden hover:shadow-md hover:border-blue-300 active:bg-gray-50 transition-all cursor-pointer"
              @click="goToFeedDetail(update.shortCode, update.id)"
            >
              <div class="flex min-h-[120px] md:min-h-[140px]">
                <!-- Photo on the left -->
                <div v-if="update.photos.length > 0" class="flex-shrink-0 w-28 md:w-36 bg-gray-200 relative overflow-hidden">
                  <div class="absolute inset-0 bg-gray-200 flex items-center justify-center">
                    <div class="w-6 h-6 border-2 border-gray-300 border-t-gray-400 rounded-full animate-spin feed-img-loader"></div>
                  </div>
                  <img
                    :src="getPhotoUrl(update.photos[0])"
                    :alt="update.photos[0].filename"
                    class="w-full h-full object-cover relative z-10 bg-gray-200"
                    loading="lazy"
                    decoding="async"
                    @load="(e: Event) => (e.target as HTMLElement).classList.add('feed-img-loaded')"
                    @error="(e: Event) => (e.target as HTMLElement).classList.add('feed-img-error')"
                  />
                  <div
                    v-if="update.photos.length > 1"
                    class="absolute bottom-1.5 right-1.5 bg-black/60 text-white text-xs px-1.5 py-0.5 rounded flex items-center gap-1 z-20"
                  >
                    <Image class="w-3 h-3" />
                    <span>+{{ update.photos.length - 1 }}</span>
                  </div>
                </div>

                <!-- Content on the right -->
                <div class="flex-1 min-w-0 p-3 md:p-4 flex flex-col">
                  <!-- Top row -->
                  <div class="flex items-center justify-between mb-1.5">
                    <div class="flex items-center gap-1.5">
                      <template v-if="update.locationId">
                        <MapPin class="w-4 h-4 text-blue-500 flex-shrink-0" />
                        <span class="text-xs font-medium text-blue-600">Posko</span>
                      </template>
                      <template v-else-if="update.faskesId">
                        <MapPin class="w-4 h-4 text-green-500 flex-shrink-0" />
                        <span class="text-xs font-medium text-green-600">Faskes</span>
                      </template>
                      <template v-else>
                        <Megaphone class="w-4 h-4 text-orange-500 flex-shrink-0" />
                        <span class="text-xs font-medium text-orange-600">Laporan Situasi</span>
                      </template>
                    </div>
                    <span class="text-xs text-gray-400">{{ update.timestamp }}</span>
                  </div>

                  <!-- Location name -->
                  <div class="mb-1">
                    <template v-if="update.locationId">
                      <span class="text-sm font-semibold text-gray-900 line-clamp-1">{{ update.location }}</span>
                      <span v-if="update.desaName || update.kecamatan" class="text-xs text-gray-500 block line-clamp-1">
                        {{ update.desaName || update.kecamatan }}{{ update.kotaKab ? `, ${update.kotaKab}` : '' }}
                      </span>
                    </template>
                    <template v-else-if="update.faskesId">
                      <span class="text-sm font-semibold text-gray-900 line-clamp-1">{{ update.faskesName }}</span>
                      <span v-if="update.desaName || update.kecamatan" class="text-xs text-gray-500 block line-clamp-1">
                        {{ update.desaName || update.kecamatan }}{{ update.kotaKab ? `, ${update.kotaKab}` : '' }}
                      </span>
                    </template>
                    <template v-else-if="update.desaName || update.kecamatan">
                      <span class="text-sm font-semibold text-gray-900 line-clamp-1">
                        {{ update.desaName || update.kecamatan }}{{ update.kotaKab ? `, ${update.kotaKab}` : '' }}
                      </span>
                    </template>
                  </div>

                  <!-- Content text -->
                  <p class="text-sm text-gray-600 leading-relaxed line-clamp-3 md:line-clamp-4 flex-1">{{ update.content }}</p>

                  <!-- Bottom row -->
                  <div class="flex items-end justify-between mt-2 gap-2">
                    <div class="flex flex-wrap gap-1">
                      <Badge :variant="categoryColors[update.category] || 'outline'" class="text-xs">
                        {{ getCategoryLabel(update.category) }}
                      </Badge>
                      <template v-if="update.type">
                        <Badge
                          v-for="t in update.type.split(/[\s,]+/).filter((tag: string) => tag).slice(0, 2)"
                          :key="t"
                          :variant="tagColor"
                          class="text-xs"
                        >
                          {{ getTagLabel(t) }}
                        </Badge>
                      </template>
                    </div>
                    <span class="text-xs text-gray-500 truncate max-w-[100px] md:max-w-[180px] text-right flex-shrink-0">
                      {{ update.username }}{{ update.organization ? ` • ${update.organization}` : '' }}
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Infinite scroll trigger & loading more indicator -->
          <div ref="loadMoreTriggerRef" class="py-6 text-center">
            <div v-if="loadingMore" class="flex items-center justify-center gap-2">
              <div class="w-5 h-5 border-2 border-gray-300 border-t-blue-500 rounded-full animate-spin"></div>
              <span class="text-sm text-gray-500">Memuat lebih banyak...</span>
            </div>
            <div v-else-if="!hasMore && formattedFeeds.length > 0" class="text-sm text-gray-400">
              Semua update telah ditampilkan
            </div>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<style scoped>
.feed-img-loaded {
  opacity: 1;
}

.feed-img-loaded + .feed-img-loader,
.feed-img-loaded ~ .feed-img-loader {
  display: none;
}

.feed-img-error {
  opacity: 0;
}

.feed-img-loaded + div .feed-img-loader {
  display: none;
}

/* Safe area for bottom */
.pb-safe {
  padding-bottom: env(safe-area-inset-bottom, 0);
}

/* Better touch scrolling */
.overscroll-contain {
  overscroll-behavior: contain;
}
</style>
