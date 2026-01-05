<script setup lang="ts">
import { ref, watch, onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import { Search, MapPin, Download, Filter } from 'lucide-vue-next'
import DataLayersSidebar from '@/components/DataLayersSidebar.vue'
import Input from '@/components/ui/Input.vue'
import Badge from '@/components/ui/Badge.vue'
import Button from '@/components/ui/Button.vue'
import { api, type Feed, type FeedFilter } from '@/services/api'

const route = useRoute()

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
    content: feed.content,
    category: feed.category,
    type: feed.type ?? '',
    coordinates: feed.coordinates,
  }))
})

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

    // Search by location_name only
    if (searchQuery.value) {
      filter.location_name = searchQuery.value
    }
    if (selectedCategory.value) {
      filter.category = selectedCategory.value
    }
    if (selectedTag.value) {
      filter.type = selectedTag.value
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
]

// Predefined categories
const allCategories = [
  { value: 'informasi', label: 'Informasi' },
  { value: 'kebutuhan', label: 'Kebutuhan' },
  { value: 'follow-up', label: 'Follow-up' },
]
</script>

<template>
  <div class="flex-1 flex overflow-hidden">
    <DataLayersSidebar />

    <!-- Feeds Content -->
    <main class="flex-1 bg-gray-50 flex flex-col overflow-hidden">
      <!-- Header -->
      <div class="bg-white border-b border-gray-200 px-6 py-4">
        <div class="max-w-4xl mx-auto">
          <div class="flex items-center justify-between mb-4">
            <div>
              <h1 class="text-2xl font-bold text-gray-900">Informasi Terbaru</h1>
              <div class="flex items-center gap-2 mt-1">
                <span class="relative flex h-2.5 w-2.5">
                  <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
                  <span class="relative inline-flex rounded-full h-2.5 w-2.5 bg-green-500"></span>
                </span>
                <span class="text-sm text-gray-500">Live Feeds - {{ total }} updates sebulan terakhir</span>
              </div>
            </div>
            <Button variant="outline" class="gap-2">
              <Download class="w-4 h-4" />
              Export
            </Button>
          </div>

          <!-- Filters Row -->
          <div class="flex gap-3">
            <!-- Search by Location Name -->
            <div class="flex-1 relative">
              <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
              <Input
                v-model="searchQuery"
                placeholder="Cari nama posko, faskes, dll..."
                class="pl-9 w-full"
                @keyup.enter="handleSearch"
              />
            </div>

            <!-- Category Filter -->
            <div class="w-48">
              <select
                v-model="selectedCategory"
                class="w-full h-10 text-sm border border-gray-200 rounded-lg px-3 bg-white focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              >
                <option value="">Semua Kategori</option>
                <option v-for="cat in allCategories" :key="cat.value" :value="cat.value">
                  {{ cat.label }}
                </option>
              </select>
            </div>

            <!-- Tags Filter (Single Select) -->
            <div class="w-48">
              <select
                v-model="selectedTag"
                class="w-full h-10 text-sm border border-gray-200 rounded-lg px-3 bg-white focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              >
                <option value="">Semua Tags</option>
                <option v-for="tag in allTags" :key="tag.value" :value="tag.value">
                  {{ tag.label }}
                </option>
              </select>
            </div>
          </div>

          <!-- Active Filters Display -->
          <div v-if="searchQuery || selectedCategory || selectedTag" class="flex flex-wrap gap-2 mt-3">
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
          </div>
        </div>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="flex-1 flex items-center justify-center">
        <div class="text-gray-500">Memuat data...</div>
      </div>

      <!-- Feeds List -->
      <div v-else class="flex-1 overflow-y-auto">
        <div class="max-w-4xl mx-auto py-4 px-6">
          <!-- Empty State -->
          <div v-if="formattedFeeds.length === 0" class="text-center py-12">
            <Filter class="w-12 h-12 text-gray-300 mx-auto mb-4" />
            <p class="text-gray-500">Tidak ada update yang ditemukan</p>
            <p class="text-sm text-gray-400 mt-1">Coba ubah filter atau kata kunci pencarian</p>
          </div>

          <!-- Feed Items -->
          <div class="space-y-3">
            <div
              v-for="update in formattedFeeds"
              :key="update.id"
              class="bg-white rounded-lg border border-gray-200 p-4 hover:shadow-sm transition-shadow"
            >
              <div class="flex items-center justify-between mb-2">
                <span class="text-xs text-gray-500">{{ update.timestamp }}</span>
              </div>
              <div class="text-xs text-blue-600 font-medium mb-2">
                {{ update.username }}{{ update.organization ? ` - ${update.organization}` : '' }}
              </div>
              <div class="flex items-center gap-1.5 mb-2">
                <MapPin class="w-4 h-4 text-blue-500 flex-shrink-0" />
                <span class="text-sm font-medium text-gray-900">{{ update.location || 'Lokasi tidak diketahui' }}</span>
              </div>
              <p class="text-sm text-gray-600 mb-3 leading-relaxed">{{ update.content }}</p>
              <div class="flex flex-wrap gap-2">
                <Badge :variant="categoryColors[update.category] || 'outline'">
                  {{ update.category }}
                </Badge>
                <template v-if="update.type">
                  <Badge
                    v-for="t in update.type.split(',')"
                    :key="t"
                    :variant="typeColors[t.trim()] || 'outline'"
                  >
                    {{ formatTagDisplay(t.trim()) }}
                  </Badge>
                </template>
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
