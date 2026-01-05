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
const showPoskoMarkers = ref(true)

const handleMarkerClick = (marker: MapMarker) => {
  selectedMarker.value = marker
  showDetail.value = true
}

const showLocationUpdates = (locationId: string) => {
  router.push({ path: '/feeds', query: { search: locationId } })
}

const closeDetailPanel = () => {
  showDetail.value = false
  selectedMarker.value = null
}

const handleLayerToggle = (layerId: string, enabled: boolean) => {
  if (layerId === 'shelter') {
    showPoskoMarkers.value = enabled
  }
}
</script>

<template>
  <div class="flex-1 flex overflow-hidden">
    <DataLayersSidebar @layer-toggle="handleLayerToggle" />
    <MapView @marker-click="handleMarkerClick" :show-markers="showPoskoMarkers" />
    <DetailPanel
      v-if="showDetail"
      :marker="selectedMarker"
      @close="closeDetailPanel"
      @show-location-updates="showLocationUpdates"
    />
  </div>
</template>
