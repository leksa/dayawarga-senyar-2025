<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { X, Search, MapPin, Download } from 'lucide-vue-next'
import Input from './ui/Input.vue'
import Badge from './ui/Badge.vue'
import Button from './ui/Button.vue'
import { useFeeds } from '@/composables/useFeeds'

interface Props {
  filterLocation?: string
}

const props = withDefaults(defineProps<Props>(), {
  filterLocation: ''
})

const emit = defineEmits<{
  close: []
}>()

const { formattedFeeds, fetchFeeds, loading, total } = useFeeds()

const searchQuery = ref(props.filterLocation)
const selectedCategory = ref('Semua')
const selectedType = ref('Semua')

// Watch for filterLocation changes
watch(() => props.filterLocation, (newVal) => {
  if (newVal) {
    searchQuery.value = newVal
    doSearch()
  }
})

onMounted(() => {
  fetchFeeds()
})

const doSearch = async () => {
  const filter: Record<string, string> = {}

  if (searchQuery.value) {
    filter.search = searchQuery.value
  }
  if (selectedCategory.value !== 'Semua') {
    filter.category = selectedCategory.value.toLowerCase()
  }
  if (selectedType.value !== 'Semua') {
    filter.type = selectedType.value.toLowerCase()
  }

  await fetchFeeds(filter)
}

watch([selectedCategory, selectedType], () => {
  doSearch()
})

const categoryColors: Record<string, 'default' | 'success' | 'warning' | 'danger' | 'outline'> = {
  kebutuhan: 'warning',
  informasi: 'outline',
}

const typeColors: Record<string, 'default' | 'success' | 'warning' | 'danger' | 'outline'> = {
  'SAR': 'danger',
  'kesehatan': 'success',
  'logistik': 'default',
  'air bersih': 'outline',
  'transportasi': 'outline',
  'listrik': 'outline',
  'infrastruktur': 'outline',
  'pendidikan': 'outline',
  'psikososial': 'outline',
  'ibadah': 'outline',
  'komunikasi': 'outline',
  'internet': 'outline',
}
</script>

<template>
  <aside class="w-96 bg-white border-l border-gray-200 flex flex-col h-full">
    <!-- Header -->
    <div class="p-4 border-b border-gray-200">
      <div class="flex items-start justify-between">
        <div>
          <h2 class="text-lg font-semibold text-gray-900">Information Updates</h2>
          <div class="flex items-center gap-2 mt-1">
            <span class="w-2 h-2 bg-green-500 rounded-full animate-pulse"></span>
            <span class="text-sm text-gray-500">Live Feed • {{ total }} Updates</span>
          </div>
        </div>
        <button
          class="p-1 hover:bg-gray-100 rounded"
          @click="emit('close')"
        >
          <X class="w-5 h-5 text-gray-400" />
        </button>
      </div>
    </div>

    <!-- Search and filters -->
    <div class="p-4 space-y-3 border-b border-gray-200">
      <div class="relative">
        <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
        <Input
          v-model="searchQuery"
          placeholder="Cari titik posko, faskes, dll"
          class="pl-9"
          @keyup.enter="doSearch"
        />
      </div>
      <div class="flex gap-2">
        <select
          v-model="selectedCategory"
          class="flex-1 text-sm border border-gray-200 rounded-md px-3 py-2 bg-white"
        >
          <option>Semua</option>
          <option>Kebutuhan</option>
          <option>Informasi</option>
        </select>
        <select
          v-model="selectedType"
          class="flex-1 text-sm border border-gray-200 rounded-md px-3 py-2 bg-white"
        >
          <option>Semua</option>
          <option>SAR</option>
          <option>Kesehatan</option>
          <option>Logistik</option>
          <option>Air bersih</option>
          <option>Transportasi</option>
          <option>Listrik</option>
          <option>Infrastruktur</option>
          <option>Pendidikan</option>
          <option>Psikososial</option>
          <option>Ibadah</option>
          <option>Komunikasi</option>
          <option>Internet</option>
        </select>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="p-8 text-center text-gray-500">
      Memuat data...
    </div>

    <!-- Updates list -->
    <div v-else class="flex-1 overflow-y-auto">
      <div
        v-for="update in formattedFeeds"
        :key="update.id"
        class="p-4 border-b border-gray-100 hover:bg-gray-50"
      >
        <div class="flex items-center justify-between mb-2">
          <span class="text-xs text-gray-500">{{ update.timestamp }}</span>
        </div>
        <div class="text-xs text-blue-600 font-medium mb-2">
          {{ update.username }} - {{ update.organization }}
        </div>
        <div class="flex items-center gap-1 mb-2">
          <MapPin class="w-4 h-4 text-blue-500" />
          <span class="text-sm font-medium text-gray-900">{{ update.location }}</span>
        </div>
        <p class="text-sm text-gray-600 mb-3">{{ update.content }}</p>
        <div class="flex gap-2">
          <Badge :variant="categoryColors[update.category] || 'outline'">
            {{ update.category }}
          </Badge>
          <Badge v-if="update.type" :variant="typeColors[update.type] || 'outline'">
            {{ update.type }}
          </Badge>
        </div>
      </div>
      <div v-if="formattedFeeds.length === 0" class="p-8 text-center text-gray-500">
        Tidak ada update yang ditemukan
      </div>
    </div>

    <!-- Footer -->
    <div class="p-4 border-t border-gray-200">
      <Button variant="primary" class="w-full">
        <Download class="w-4 h-4 mr-2" />
        Export Logs
      </Button>
    </div>
  </aside>
</template>
