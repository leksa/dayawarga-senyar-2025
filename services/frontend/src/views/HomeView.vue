<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import DataLayersSidebar from '@/components/DataLayersSidebar.vue'
import MapView from '@/components/MapView.vue'
import DetailPanel from '@/components/DetailPanel.vue'
import { api } from '@/services/api'
import type { MapMarker } from '@/types'

const router = useRouter()
const route = useRoute()

const mapViewRef = ref<InstanceType<typeof MapView> | null>(null)
const showDetail = ref(false)
const selectedMarker = ref<MapMarker | null>(null)
const selectedFaskes = ref<any | null>(null)
const selectedInfrastruktur = ref<any | null>(null)
const showPoskoMarkers = ref(true)
const showFaskesMarkers = ref(false)
const showInfrastrukturMarkers = ref(false)
const showFeedsMarkers = ref(false)

// Handle URL query params for map navigation and detail panel
onMounted(() => {
  checkQueryParams()
})

watch(() => route.query, () => {
  checkQueryParams()
})

const checkQueryParams = async () => {
  const { lat, lng, zoom, location, faskes, infrastruktur, feed } = route.query

  // Handle map navigation with optional feed popup
  if (lat && lng && mapViewRef.value) {
    const latitude = parseFloat(lat as string)
    const longitude = parseFloat(lng as string)
    const zoomLevel = zoom ? parseInt(zoom as string) : 17
    if (!isNaN(latitude) && !isNaN(longitude)) {
      mapViewRef.value.flyTo(latitude, longitude, zoomLevel)

      // If feed ID is provided, show feed popup after flying to location
      if (feed) {
        setTimeout(() => {
          mapViewRef.value?.showFeedPopup(feed as string, latitude, longitude)
        }, 1600)
      }
    }
  }

  // Handle location detail from query param
  if (location) {
    await showLocationDetail(location as string)
  }

  // Handle faskes detail from query param
  if (faskes) {
    await showFaskesDetail(faskes as string)
  }

  // Handle infrastruktur detail from query param
  if (infrastruktur) {
    await showInfrastrukturDetail(infrastruktur as string)
  }
}

// Fetch and show location detail
const showLocationDetail = async (locationId: string) => {
  try {
    const response = await api.getLocationById(locationId)
    if (response.success && response.data) {
      const loc = response.data
      // Convert to MapMarker format
      const marker: MapMarker = {
        id: loc.id,
        name: (loc.identitas as any)?.nama || 'Unknown',
        type: loc.type,
        status: loc.status,
        lat: loc.geometry.coordinates[1],
        lng: loc.geometry.coordinates[0],
        jumlahKK: (loc.data_pengungsi as any)?.jumlah_kk || 0,
        totalJiwa: (loc.data_pengungsi as any)?.total_jiwa || 0,
      }
      selectedMarker.value = marker
      selectedFaskes.value = null
      showDetail.value = true

      // Fly to location
      if (mapViewRef.value) {
        mapViewRef.value.flyTo(marker.lat, marker.lng, 15)
      }
    }
  } catch (e) {
    console.error('Failed to fetch location:', e)
  }
}

// Fetch and show faskes detail
const showFaskesDetail = async (faskesId: string) => {
  try {
    const response = await api.getFaskesById(faskesId)
    if (response.success && response.data) {
      const fk = response.data
      // Convert to faskes marker format
      const marker = {
        id: fk.id,
        nama: fk.nama,
        jenisFaskes: fk.jenis_faskes,
        statusFaskes: fk.status_faskes,
        kondisiFaskes: fk.kondisi_faskes,
        lat: fk.geometry.coordinates[1],
        lng: fk.geometry.coordinates[0],
      }
      selectedFaskes.value = marker
      selectedMarker.value = null
      selectedInfrastruktur.value = null
      showDetail.value = true

      // Fly to faskes
      if (mapViewRef.value) {
        mapViewRef.value.flyTo(marker.lat, marker.lng, 15)
      }
    }
  } catch (e) {
    console.error('Failed to fetch faskes:', e)
  }
}

// Fetch and show infrastruktur detail
const showInfrastrukturDetail = async (infrastrukturId: string) => {
  try {
    const response = await api.getInfrastrukturById(infrastrukturId)
    if (response.success && response.data) {
      const inf = response.data
      // Convert to infrastruktur marker format
      const marker = {
        id: inf.id,
        name: inf.nama,
        jenis: inf.jenis,
        statusJln: inf.status_jln,
        statusAkses: inf.status_akses,
        statusPenanganan: inf.status_penanganan,
        progress: inf.progress,
        lat: inf.geometry.coordinates[1],
        lng: inf.geometry.coordinates[0],
      }
      selectedInfrastruktur.value = marker
      selectedMarker.value = null
      selectedFaskes.value = null
      showDetail.value = true

      // Fly to infrastruktur
      if (mapViewRef.value) {
        mapViewRef.value.flyTo(marker.lat, marker.lng, 15)
      }
    }
  } catch (e) {
    console.error('Failed to fetch infrastruktur:', e)
  }
}

const handleMarkerClick = (marker: MapMarker) => {
  selectedMarker.value = marker
  selectedFaskes.value = null
  selectedInfrastruktur.value = null
  showDetail.value = true
}

const handleFaskesClick = (marker: any) => {
  selectedFaskes.value = marker
  selectedMarker.value = null
  selectedInfrastruktur.value = null
  showDetail.value = true
}

const handleInfrastrukturClick = (marker: any) => {
  selectedInfrastruktur.value = marker
  selectedMarker.value = null
  selectedFaskes.value = null
  showDetail.value = true
}

const showLocationUpdates = (locationId: string) => {
  router.push({ path: '/feeds', query: { search: locationId } })
}

const closeDetailPanel = () => {
  showDetail.value = false
  selectedMarker.value = null
  selectedFaskes.value = null
  selectedInfrastruktur.value = null
}

const handleLayerToggle = (layerId: string, enabled: boolean) => {
  if (layerId === 'shelter') {
    showPoskoMarkers.value = enabled
  } else if (layerId === 'medical') {
    showFaskesMarkers.value = enabled
  } else if (layerId === 'infrastructure') {
    showInfrastrukturMarkers.value = enabled
  } else if (layerId === 'feeds') {
    showFeedsMarkers.value = enabled
  }
}
</script>

<template>
  <div class="flex-1 flex overflow-hidden">
    <DataLayersSidebar @layer-toggle="handleLayerToggle" />
    <MapView
      ref="mapViewRef"
      @marker-click="handleMarkerClick"
      @faskes-click="handleFaskesClick"
      @infrastruktur-click="handleInfrastrukturClick"
      @show-location-detail="showLocationDetail"
      @show-faskes-detail="showFaskesDetail"
      @show-infrastruktur-detail="showInfrastrukturDetail"
      :show-markers="showPoskoMarkers"
      :show-faskes="showFaskesMarkers"
      :show-infrastruktur="showInfrastrukturMarkers"
      :show-feeds="showFeedsMarkers"
    />
    <DetailPanel
      v-if="showDetail"
      :marker="selectedMarker"
      :faskes="selectedFaskes"
      :infrastruktur="selectedInfrastruktur"
      @close="closeDetailPanel"
      @show-location-updates="showLocationUpdates"
    />
  </div>
</template>
