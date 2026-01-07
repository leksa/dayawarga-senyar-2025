<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import DataLayersSidebar from '@/components/DataLayersSidebar.vue'
import MapView from '@/components/MapView.vue'
import DetailPanel from '@/components/DetailPanel.vue'
import type { MapMarker } from '@/types'

const router = useRouter()

const showDetail = ref(false)
const selectedMarker = ref<MapMarker | null>(null)
const selectedFaskes = ref<any | null>(null)
const showPoskoMarkers = ref(true)
const showFaskesMarkers = ref(false)

const handleMarkerClick = (marker: MapMarker) => {
  selectedMarker.value = marker
  selectedFaskes.value = null
  showDetail.value = true
}

const handleFaskesClick = (marker: any) => {
  selectedFaskes.value = marker
  selectedMarker.value = null
  showDetail.value = true
}

const showLocationUpdates = (locationId: string) => {
  router.push({ path: '/feeds', query: { search: locationId } })
}

const closeDetailPanel = () => {
  showDetail.value = false
  selectedMarker.value = null
  selectedFaskes.value = null
}

const handleLayerToggle = (layerId: string, enabled: boolean) => {
  if (layerId === 'shelter') {
    showPoskoMarkers.value = enabled
  } else if (layerId === 'medical') {
    showFaskesMarkers.value = enabled
  }
}
</script>

<template>
  <div class="flex-1 flex overflow-hidden">
    <DataLayersSidebar @layer-toggle="handleLayerToggle" />
    <MapView
      @marker-click="handleMarkerClick"
      @faskes-click="handleFaskesClick"
      :show-markers="showPoskoMarkers"
      :show-faskes="showFaskesMarkers"
    />
    <DetailPanel
      v-if="showDetail"
      :marker="selectedMarker"
      :faskes="selectedFaskes"
      @close="closeDetailPanel"
      @show-location-updates="showLocationUpdates"
    />
  </div>
</template>
