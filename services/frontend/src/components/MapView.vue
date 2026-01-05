<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { Plus, Minus, Locate, Layers } from 'lucide-vue-next'
import { useLocations } from '@/composables/useLocations'
import type { MapMarker } from '@/types'

const props = withDefaults(defineProps<{
  showMarkers?: boolean
}>(), {
  showMarkers: true
})

const emit = defineEmits<{
  'marker-click': [marker: MapMarker]
}>()

const { markers, fetchLocations, loading, lastUpdate } = useLocations()
const mapContainer = ref<HTMLElement | null>(null)
const showLayerMenu = ref(false)
const activeLayer = ref<'street' | 'satellite' | 'terrain'>('street')
let map: any = null
let markerLayer: any = null
let currentTileLayer: any = null

const tileLayers = {
  street: {
    url: 'https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png',
    attribution: '&copy; OpenStreetMap contributors',
    name: 'Peta Jalan'
  },
  satellite: {
    url: 'https://server.arcgisonline.com/ArcGIS/rest/services/World_Imagery/MapServer/tile/{z}/{y}/{x}',
    attribution: '&copy; Esri',
    name: 'Satelit'
  },
  terrain: {
    url: 'https://{s}.tile.opentopomap.org/{z}/{x}/{y}.png',
    attribution: '&copy; OpenTopoMap contributors',
    name: 'Terrain'
  }
}

const markerIcons: Record<string, { color: string; icon: string }> = {
  posko: { color: '#3B82F6', icon: '🏠' },
  air_bersih: { color: '#06B6D4', icon: '💧' },
  kesehatan: { color: '#EF4444', icon: '🏥' },
  logistik: { color: '#F59E0B', icon: '📦' },
  dapur_umum: { color: '#10B981', icon: '🍲' },
}

const getMarkerIcon = (type: string) => {
  return markerIcons[type] || { color: '#6B7280', icon: '📍' }
}

onMounted(async () => {
  if (!mapContainer.value) return

  const L = await import('leaflet')

  // Center on Aceh (Siklon Senyar affected area)
  map = L.map(mapContainer.value).setView([4.7, 97.5], 8)

  // Add default tile layer
  currentTileLayer = L.tileLayer(tileLayers.street.url, {
    attribution: tileLayers.street.attribution
  }).addTo(map)

  markerLayer = L.layerGroup().addTo(map)

  // Fetch locations from API
  await fetchLocations()
})

// Watch for showMarkers prop changes
watch(() => props.showMarkers, (show) => {
  if (!markerLayer) return
  if (show) {
    markerLayer.addTo(map)
  } else {
    markerLayer.remove()
  }
})

// Watch for marker changes and update map
watch(markers, async (newMarkers) => {
  if (!map || !markerLayer) return

  const L = await import('leaflet')
  markerLayer.clearLayers()

  if (newMarkers.length === 0) return

  const bounds: [number, number][] = []

  newMarkers.forEach((marker) => {
    const iconConfig = getMarkerIcon(marker.type)
    const customIcon = L.divIcon({
      className: 'custom-marker',
      html: `<div style="background-color: ${iconConfig.color}; width: 32px; height: 32px; border-radius: 50%; display: flex; align-items: center; justify-content: center; border: 2px solid white; box-shadow: 0 2px 4px rgba(0,0,0,0.3); font-size: 14px;">${iconConfig.icon}</div>`,
      iconSize: [32, 32],
      iconAnchor: [16, 16],
    })

    L.marker([marker.lat, marker.lng], { icon: customIcon })
      .addTo(markerLayer)
      .on('click', () => {
        emit('marker-click', marker)
      })

    bounds.push([marker.lat, marker.lng])
  })

  // Auto-fit to show all markers
  if (bounds.length > 0) {
    map.fitBounds(bounds, { padding: [50, 50], maxZoom: 12 })
  }
}, { immediate: true })

const zoomIn = () => map?.zoomIn()
const zoomOut = () => map?.zoomOut()
const locateMe = () => {
  if (navigator.geolocation) {
    navigator.geolocation.getCurrentPosition((pos) => {
      map?.setView([pos.coords.latitude, pos.coords.longitude], 14)
    })
  }
}

const toggleLayerMenu = () => {
  showLayerMenu.value = !showLayerMenu.value
}

const switchLayer = async (layerKey: 'street' | 'satellite' | 'terrain') => {
  if (!map || activeLayer.value === layerKey) return

  const L = await import('leaflet')
  const layerConfig = tileLayers[layerKey]

  // Remove current tile layer
  if (currentTileLayer) {
    map.removeLayer(currentTileLayer)
  }

  // Add new tile layer
  currentTileLayer = L.tileLayer(layerConfig.url, {
    attribution: layerConfig.attribution
  }).addTo(map)

  activeLayer.value = layerKey
  showLayerMenu.value = false
}

defineExpose({
  lastUpdate,
  loading,
  refreshLocations: fetchLocations,
})
</script>

<template>
  <div class="relative flex-1 h-full">
    <!-- Loading indicator -->
    <div v-if="loading" class="absolute top-4 left-1/2 -translate-x-1/2 z-[1001] bg-white px-4 py-2 rounded-lg shadow-md">
      <span class="text-sm text-gray-600">Memuat data...</span>
    </div>

    <!-- Map container -->
    <div ref="mapContainer" class="w-full h-full"></div>

    <!-- Map controls -->
    <div class="absolute right-4 bottom-24 z-[1000] flex flex-col gap-2">
      <button
        class="w-10 h-10 bg-white rounded-lg shadow-md flex items-center justify-center hover:bg-gray-50"
        @click="zoomIn"
      >
        <Plus class="w-5 h-5 text-gray-600" />
      </button>
      <button
        class="w-10 h-10 bg-white rounded-lg shadow-md flex items-center justify-center hover:bg-gray-50"
        @click="zoomOut"
      >
        <Minus class="w-5 h-5 text-gray-600" />
      </button>
      <button
        class="w-10 h-10 bg-white rounded-lg shadow-md flex items-center justify-center hover:bg-gray-50"
        @click="locateMe"
      >
        <Locate class="w-5 h-5 text-gray-600" />
      </button>
      <div class="relative">
        <button
          class="w-10 h-10 bg-white rounded-lg shadow-md flex items-center justify-center hover:bg-gray-50"
          @click="toggleLayerMenu"
        >
          <Layers class="w-5 h-5 text-gray-600" />
        </button>

        <!-- Layer menu popup -->
        <div
          v-if="showLayerMenu"
          class="absolute bottom-0 right-12 bg-white rounded-lg shadow-lg py-2 min-w-[140px]"
        >
          <button
            v-for="(layer, key) in tileLayers"
            :key="key"
            :class="[
              'w-full px-4 py-2 text-left text-sm hover:bg-gray-100 flex items-center gap-2',
              activeLayer === key ? 'text-blue-600 font-medium bg-blue-50' : 'text-gray-700'
            ]"
            @click="switchLayer(key as 'street' | 'satellite' | 'terrain')"
          >
            <span v-if="activeLayer === key" class="w-2 h-2 bg-blue-600 rounded-full"></span>
            <span v-else class="w-2 h-2"></span>
            {{ layer.name }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style>
.custom-marker {
  background: transparent;
  border: none;
}
</style>
