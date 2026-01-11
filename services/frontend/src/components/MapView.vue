<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { Locate, Layers, MapPin, ChevronDown, X, Plus, Minus } from 'lucide-vue-next'
import { useLocations } from '@/composables/useLocations'
import { useFaskes } from '@/composables/useFaskes'
import { api, type Feed } from '@/services/api'
import type { MapMarker } from '@/types'

const props = withDefaults(defineProps<{
  showMarkers?: boolean
  showFaskes?: boolean
  showFeeds?: boolean
}>(), {
  showMarkers: true,
  showFaskes: false,
  showFeeds: true
})

const emit = defineEmits<{
  'marker-click': [marker: MapMarker]
  'faskes-click': [marker: any]
  'show-location-detail': [locationId: string]
  'show-faskes-detail': [faskesId: string]
}>()

const { markers, fetchLocations, loading, lastUpdate } = useLocations()
const { markers: faskesMarkers, fetchFaskes } = useFaskes()
const mapContainer = ref<HTMLElement | null>(null)
const showLayerMenu = ref(false)
const activeLayer = ref<'street' | 'satellite' | 'terrain'>('street')
const feedsWithCoords = ref<Feed[]>([])
let map: any = null
let markerLayer: any = null
let faskesLayer: any = null
let feedsLayer: any = null
let currentTileLayer: any = null

// Region filter state - pending (before Apply) and applied (after Apply)
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

// Collect unique region values from all markers
const collectRegionData = () => {
  const kotaKabSet = new Set<string>()
  const kecamatanMap = new Map<string, Set<string>>()
  const desaMap = new Map<string, Set<string>>()

  // From location markers
  markers.value.forEach(m => {
    if (m.namaKotaKab) kotaKabSet.add(m.namaKotaKab)
    if (m.namaKotaKab && m.namaKecamatan) {
      if (!kecamatanMap.has(m.namaKotaKab)) kecamatanMap.set(m.namaKotaKab, new Set())
      kecamatanMap.get(m.namaKotaKab)!.add(m.namaKecamatan)
    }
    if (m.namaKecamatan && m.namaDesa) {
      if (!desaMap.has(m.namaKecamatan)) desaMap.set(m.namaKecamatan, new Set())
      desaMap.get(m.namaKecamatan)!.add(m.namaDesa)
    }
  })

  // From faskes markers
  faskesMarkers.value.forEach(m => {
    if (m.namaKotaKab) kotaKabSet.add(m.namaKotaKab)
    if (m.namaKotaKab && m.namaKecamatan) {
      if (!kecamatanMap.has(m.namaKotaKab)) kecamatanMap.set(m.namaKotaKab, new Set())
      kecamatanMap.get(m.namaKotaKab)!.add(m.namaKecamatan)
    }
    if (m.namaKecamatan && m.namaDesa) {
      if (!desaMap.has(m.namaKecamatan)) desaMap.set(m.namaKecamatan, new Set())
      desaMap.get(m.namaKecamatan)!.add(m.namaDesa)
    }
  })

  // From feeds
  feedsWithCoords.value.forEach(f => {
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
const filterLabel = computed(() => {
  if (appliedDesa.value) return appliedDesa.value
  if (appliedKecamatan.value) return appliedKecamatan.value
  if (appliedKotaKab.value) return appliedKotaKab.value
  if (appliedProvinsi.value) return appliedProvinsi.value
  return 'Semua Wilayah'
})

const hasActiveFilter = computed(() => {
  return appliedProvinsi.value !== '' || appliedKotaKab.value !== '' || appliedKecamatan.value !== '' || appliedDesa.value !== ''
})

// Handle province selection
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

// Apply filter
const applyFilter = () => {
  appliedProvinsi.value = pendingProvinsi.value
  appliedKotaKab.value = pendingKotaKab.value
  appliedKecamatan.value = pendingKecamatan.value
  appliedDesa.value = pendingDesa.value
  showRegionFilter.value = false
}

// Clear filter
const clearFilter = () => {
  pendingProvinsi.value = ''
  pendingKotaKab.value = ''
  pendingKecamatan.value = ''
  pendingDesa.value = ''
  appliedProvinsi.value = ''
  appliedKotaKab.value = ''
  appliedKecamatan.value = ''
  appliedDesa.value = ''
  availableKotaKab.value = []
  availableKecamatan.value = []
  availableDesa.value = []
  filterLevel.value = 'provinsi'
}

// Go back to previous level
const goBackLevel = () => {
  if (filterLevel.value === 'desa') {
    filterLevel.value = 'kecamatan'
    pendingDesa.value = ''
  } else if (filterLevel.value === 'kecamatan') {
    filterLevel.value = 'kotakab'
    pendingKecamatan.value = ''
    pendingDesa.value = ''
  } else if (filterLevel.value === 'kotakab') {
    filterLevel.value = 'provinsi'
    pendingKotaKab.value = ''
    pendingKecamatan.value = ''
    pendingDesa.value = ''
  }
}

// Toggle filter dropdown
const toggleRegionFilter = () => {
  if (!showRegionFilter.value) {
    // Opening: sync pending with applied
    pendingProvinsi.value = appliedProvinsi.value
    pendingKotaKab.value = appliedKotaKab.value
    pendingKecamatan.value = appliedKecamatan.value
    pendingDesa.value = appliedDesa.value
    // Set appropriate level
    if (appliedDesa.value) filterLevel.value = 'desa'
    else if (appliedKecamatan.value) filterLevel.value = 'kecamatan'
    else if (appliedKotaKab.value) filterLevel.value = 'kotakab'
    else if (appliedProvinsi.value) filterLevel.value = 'kotakab'
    else filterLevel.value = 'provinsi'
    // Update available options
    if (appliedProvinsi.value) updateAvailableKotaKab()
    if (appliedKotaKab.value) updateAvailableKecamatan(appliedKotaKab.value)
    if (appliedKecamatan.value) updateAvailableDesa(appliedKecamatan.value)
  }
  showRegionFilter.value = !showRegionFilter.value
}

// Fetch feeds with coordinates (free geolocation feeds)
const fetchFeedsWithCoords = async () => {
  try {
    const thirtyDaysAgo = new Date()
    thirtyDaysAgo.setDate(thirtyDaysAgo.getDate() - 30)

    const response = await api.getFeeds({
      since: thirtyDaysAgo.toISOString(),
      limit: 100
    })

    if (response.success && response.data) {
      // Filter only feeds that have coordinates (free geolocation or related)
      feedsWithCoords.value = response.data.filter(feed =>
        feed.coordinates && feed.coordinates.length === 2
      )
    }
  } catch (e) {
    console.error('Failed to fetch feeds:', e)
  }
}

// Get feed photo URL
const getFeedPhotoUrl = (photoId: string) => {
  const baseUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1'
  return `${baseUrl}/feeds/photos/${photoId}/file`
}

// Format timestamp for popup
const formatTimestamp = (isoString: string): string => {
  const date = new Date(isoString)
  const day = date.getDate().toString().padStart(2, '0')
  const month = (date.getMonth() + 1).toString().padStart(2, '0')
  const hours = date.getHours().toString().padStart(2, '0')
  const minutes = date.getMinutes().toString().padStart(2, '0')
  return `${day}/${month} ${hours}:${minutes}`
}

// Build popup content for feed
const buildFeedPopupContent = (feed: Feed): string => {
  // Photo at top (full width)
  const photoHtml = feed.photos && feed.photos.length > 0
    ? `<div class="popup-photo"><img src="${getFeedPhotoUrl(feed.photos[0].id)}" alt="Foto" loading="lazy" /></div>`
    : ''

  // Google Maps link (if coordinates available)
  const gmapsLink = feed.coordinates
    ? `<a href="https://www.google.com/maps?q=${feed.coordinates[1]},${feed.coordinates[0]}" target="_blank" class="gmaps-link" title="Buka di Google Maps">
        <span class="gmaps-text">GMaps</span>
        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 2C8.13 2 5 5.13 5 9c0 5.25 7 13 7 13s7-7.75 7-13c0-3.87-3.13-7-7-7z"/><circle cx="12" cy="9" r="2.5"/></svg>
      </a>`
    : ''

  // Location name (clickable if has location_id or faskes_id)
  let locationHtml = ''
  if (feed.location_id && feed.location_name) {
    locationHtml = `<div class="popup-location clickable" data-location-id="${feed.location_id}">
      <span class="location-icon">📍</span>
      <span class="location-name">${feed.location_name}</span>
      ${gmapsLink}
    </div>`
  } else if (feed.faskes_id && feed.faskes_name) {
    locationHtml = `<div class="popup-location clickable" data-faskes-id="${feed.faskes_id}">
      <span class="location-icon">🏥</span>
      <span class="location-name">${feed.faskes_name}</span>
      <span class="faskes-badge">Faskes</span>
      ${gmapsLink}
    </div>`
  } else {
    locationHtml = `<div class="popup-location free">
      <span class="location-icon"><svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m3 11 18-5v12L3 13v-2z"/><path d="M11.6 16.8a3 3 0 1 1-5.8-1.6"/></svg></span>
      <span class="location-name">Laporan Situasi</span>
      ${gmapsLink}
    </div>`
  }

  // Submitter info
  const submitterHtml = feed.username
    ? `<div class="popup-submitter">oleh: ${feed.username}</div>`
    : ''

  // Content
  const contentHtml = `<div class="popup-content">${feed.content}</div>`

  // Region info
  let regionHtml = ''
  if (feed.region) {
    const parts = []
    if (feed.region.desa) parts.push(feed.region.desa)
    if (feed.region.kecamatan) parts.push(feed.region.kecamatan)
    if (feed.region.kota_kab) parts.push(feed.region.kota_kab)
    if (parts.length > 0) {
      regionHtml = `<div class="popup-region">${parts.join(' • ')}</div>`
    }
  }

  // Category badge
  const categoryClass = feed.category === 'kebutuhan' ? 'cat-kebutuhan' :
    feed.category === 'follow-up' ? 'cat-followup' : 'cat-info'

  // Type/tags badges
  let tagsHtml = ''
  if (feed.type) {
    const tags = feed.type.split(/[\s,]+/).filter(t => t)
    tagsHtml = tags.map(tag => `<span class="popup-tag">${tag}</span>`).join('')
  }

  // Bottom row: date, category, tags
  const bottomHtml = `<div class="popup-bottom">
    <span class="popup-date">${formatTimestamp(feed.submitted_at)}</span>
    <span class="popup-category ${categoryClass}">${feed.category}</span>
    ${tagsHtml}
  </div>`

  return `
    <div class="feed-popup-new">
      ${photoHtml}
      ${locationHtml}
      ${submitterHtml}
      ${contentHtml}
      ${regionHtml}
      ${bottomHtml}
    </div>
  `
}

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

const faskesIcons: Record<string, { color: string; icon: string }> = {
  rumah_sakit: { color: '#940000', icon: '🏥' },      // Dark Red - Rumah Sakit
  puskesmas: { color: '#EC4899', icon: '🏥' },        // Pink - Puskesmas (sama dengan RS, warna merah muda)
  klinik: { color: '#2563EB', icon: '💊' },           // Blue - Klinik
  posko_kes_darurat: { color: '#F59E0B', icon: '⛑️' }, // Orange/Amber - Posko Darurat
}

const getMarkerIcon = (type: string) => {
  return markerIcons[type] || { color: '#6B7280', icon: '📍' }
}

const getFaskesIcon = (jenisFaskes: string) => {
  return faskesIcons[jenisFaskes] || { color: '#EF4444', icon: '🏥' }
}

// Setup click handlers for popup location links
const setupPopupClickHandlers = () => {
  document.addEventListener('click', (e) => {
    const target = e.target as HTMLElement
    const locationEl = target.closest('[data-location-id]') as HTMLElement
    const faskesEl = target.closest('[data-faskes-id]') as HTMLElement

    if (locationEl) {
      const locationId = locationEl.dataset.locationId
      if (locationId) {
        emit('show-location-detail', locationId)
        map?.closePopup()
      }
    } else if (faskesEl) {
      const faskesId = faskesEl.dataset.faskesId
      if (faskesId) {
        emit('show-faskes-detail', faskesId)
        map?.closePopup()
      }
    }
  })
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
  faskesLayer = L.layerGroup()
  feedsLayer = L.layerGroup()

  // Setup popup click handlers
  setupPopupClickHandlers()

  // Fetch locations from API (use high limit to get all markers on map)
  await fetchLocations({ limit: 1000 })

  // Fetch faskes if enabled
  if (props.showFaskes) {
    await fetchFaskes({ limit: 1000 })
    faskesLayer.addTo(map)
  }

  // Fetch feeds with coordinates and show on map
  if (props.showFeeds) {
    await fetchFeedsWithCoords()
    feedsLayer.addTo(map)
  }
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

// Watch for showFaskes prop changes
watch(() => props.showFaskes, async (show) => {
  if (!faskesLayer) return
  if (show) {
    // Fetch faskes data if not already loaded
    if (faskesMarkers.value.length === 0) {
      await fetchFaskes({ limit: 1000 })
    }
    faskesLayer.addTo(map)
  } else {
    faskesLayer.remove()
  }
})

// Watch for showFeeds prop changes
watch(() => props.showFeeds, async (show) => {
  if (!feedsLayer) return
  if (show) {
    // Fetch feeds data if not already loaded
    if (feedsWithCoords.value.length === 0) {
      await fetchFeedsWithCoords()
    }
    feedsLayer.addTo(map)
  } else {
    feedsLayer.remove()
  }
})

// Computed filtered location markers based on region filter
const filteredLocationMarkers = computed(() => {
  // Access reactive values directly for proper reactivity
  const prov = appliedProvinsi.value
  const kab = appliedKotaKab.value
  const kec = appliedKecamatan.value
  const desa = appliedDesa.value

  // No filter - return all
  if (!prov && !kab && !kec && !desa) {
    return markers.value
  }

  return markers.value.filter(m => {
    const normalize = (s?: string) => (s || '').toLowerCase().trim()

    if (desa) {
      const mDesa = normalize(m.namaDesa)
      const fDesa = normalize(desa)
      return mDesa.includes(fDesa) || fDesa.includes(mDesa)
    }
    if (kec) {
      const mKec = normalize(m.namaKecamatan)
      const fKec = normalize(kec)
      return mKec.includes(fKec) || fKec.includes(mKec)
    }
    if (kab) {
      const mKab = normalize(m.namaKotaKab)
      const fKab = normalize(kab)
      return mKab.includes(fKab) || fKab.includes(mKab)
    }
    if (prov) {
      const mProv = normalize(m.namaProvinsi)
      const fProv = normalize(prov)
      return mProv.includes(fProv) || fProv.includes(mProv)
    }
    return true
  })
})

// Computed filtered faskes markers based on region filter
const filteredFaskesMarkers = computed(() => {
  // Access reactive values directly for proper reactivity
  const prov = appliedProvinsi.value
  const kab = appliedKotaKab.value
  const kec = appliedKecamatan.value
  const desa = appliedDesa.value

  // No filter - return all
  if (!prov && !kab && !kec && !desa) {
    return faskesMarkers.value
  }

  return faskesMarkers.value.filter(m => {
    const normalize = (s?: string) => (s || '').toLowerCase().trim()

    if (desa) {
      const mDesa = normalize(m.namaDesa)
      const fDesa = normalize(desa)
      return mDesa.includes(fDesa) || fDesa.includes(mDesa)
    }
    if (kec) {
      const mKec = normalize(m.namaKecamatan)
      const fKec = normalize(kec)
      return mKec.includes(fKec) || fKec.includes(mKec)
    }
    if (kab) {
      const mKab = normalize(m.namaKotaKab)
      const fKab = normalize(kab)
      return mKab.includes(fKab) || fKab.includes(mKab)
    }
    if (prov) {
      const mProv = normalize(m.namaProvinsi)
      const fProv = normalize(prov)
      return mProv.includes(fProv) || fProv.includes(mProv)
    }
    return true
  })
})

// Computed filtered feeds based on region filter
const filteredFeeds = computed(() => {
  // Access reactive values directly for proper reactivity
  const prov = appliedProvinsi.value
  const kab = appliedKotaKab.value
  const kec = appliedKecamatan.value
  const desa = appliedDesa.value

  // No filter - return all
  if (!prov && !kab && !kec && !desa) {
    return feedsWithCoords.value
  }

  return feedsWithCoords.value.filter(f => {
    const normalize = (s?: string) => (s || '').toLowerCase().trim()

    if (desa) {
      const mDesa = normalize(f.region?.desa)
      const fDesa = normalize(desa)
      return mDesa.includes(fDesa) || fDesa.includes(mDesa)
    }
    if (kec) {
      const mKec = normalize(f.region?.kecamatan)
      const fKec = normalize(kec)
      return mKec.includes(fKec) || fKec.includes(mKec)
    }
    if (kab) {
      const mKab = normalize(f.region?.kota_kab)
      const fKab = normalize(kab)
      return mKab.includes(fKab) || fKab.includes(mKab)
    }
    if (prov) {
      const mProv = normalize(f.region?.provinsi)
      const fProv = normalize(prov)
      return mProv.includes(fProv) || fProv.includes(mProv)
    }
    return true
  })
})

// Watch for filtered location marker changes and update map
watch(filteredLocationMarkers, async (newMarkers) => {
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

// Watch for filtered faskes marker changes and update map
watch(filteredFaskesMarkers, async (newMarkers) => {
  if (!map || !faskesLayer) return

  const L = await import('leaflet')
  faskesLayer.clearLayers()

  if (newMarkers.length === 0) return

  newMarkers.forEach((marker) => {
    const iconConfig = getFaskesIcon(marker.jenisFaskes)
    const customIcon = L.divIcon({
      className: 'custom-marker',
      html: `<div style="background-color: ${iconConfig.color}; width: 32px; height: 32px; border-radius: 50%; display: flex; align-items: center; justify-content: center; border: 2px solid white; box-shadow: 0 2px 4px rgba(0,0,0,0.3); font-size: 14px;">${iconConfig.icon}</div>`,
      iconSize: [32, 32],
      iconAnchor: [16, 16],
    })

    L.marker([marker.lat, marker.lng], { icon: customIcon })
      .addTo(faskesLayer)
      .on('click', () => {
        emit('faskes-click', marker)
      })
  })
}, { immediate: true })

// Watch for filtered feeds changes and update map
watch(filteredFeeds, async (newFeeds) => {
  if (!map || !feedsLayer) return

  const L = await import('leaflet')
  feedsLayer.clearLayers()

  if (newFeeds.length === 0) return

  newFeeds.forEach((feed) => {
    if (!feed.coordinates) return

    // Determine icon based on feed type
    const hasPhoto = feed.photos && feed.photos.length > 0
    const isLaporSituasi = !feed.location_id && !feed.faskes_id
    const iconColor = isLaporSituasi ? '#F97316' : '#8B5CF6' // orange for lapor situasi, purple for related feeds
    const iconEmoji = hasPhoto ? '📷' : '📝'

    const customIcon = L.divIcon({
      className: 'custom-marker feed-marker',
      html: `<div style="background-color: ${iconColor}; width: 28px; height: 28px; border-radius: 50%; display: flex; align-items: center; justify-content: center; border: 2px solid white; box-shadow: 0 2px 4px rgba(0,0,0,0.3); font-size: 12px;">${iconEmoji}</div>`,
      iconSize: [28, 28],
      iconAnchor: [14, 14],
    })

    const marker = L.marker([feed.coordinates[1], feed.coordinates[0]], { icon: customIcon })
      .addTo(feedsLayer)

    // Add popup with feed content and photo
    const popupContent = buildFeedPopupContent(feed)
    marker.bindPopup(popupContent, {
      maxWidth: 350,
      className: 'feed-popup-container',
      closeButton: false
    })
  })
}, { immediate: true })

const locateMe = () => {
  if (navigator.geolocation) {
    navigator.geolocation.getCurrentPosition((pos) => {
      map?.setView([pos.coords.latitude, pos.coords.longitude], 14)
    })
  }
}

const zoomIn = () => {
  if (map) {
    map.zoomIn()
  }
}

const zoomOut = () => {
  if (map) {
    map.zoomOut()
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

const flyTo = (lat: number, lng: number, zoom: number = 15) => {
  if (map) {
    map.flyTo([lat, lng], zoom, { duration: 1.5 })
  }
}

// Show popup for a specific feed at given coordinates
const showFeedPopup = async (feedId: string, lat: number, lng: number) => {
  if (!map) return

  const L = await import('leaflet')

  // Try to find feed in existing data first
  let feed = feedsWithCoords.value.find(f => f.id === feedId)

  // If not found, fetch it
  if (!feed) {
    try {
      const response = await api.getFeeds({ limit: 100 })
      if (response.success && response.data) {
        feed = response.data.find(f => f.id === feedId)
      }
    } catch (e) {
      console.error('Failed to fetch feed for popup:', e)
    }
  }

  if (feed) {
    const popupContent = buildFeedPopupContent(feed)
    L.popup({
      closeButton: false,
      className: 'feed-popup-container',
      maxWidth: 350,
      autoPanPaddingTopLeft: [50, 150],
      autoPanPaddingBottomRight: [50, 50]
    })
      .setLatLng([lat, lng])
      .setContent(popupContent)
      .openOn(map)
  }
}

defineExpose({
  lastUpdate,
  loading,
  refreshLocations: fetchLocations,
  refreshFaskes: fetchFaskes,
  refreshFeeds: fetchFeedsWithCoords,
  flyTo,
  showFeedPopup,
})
</script>

<template>
  <div class="relative flex-1 h-full">
    <!-- Loading indicator -->
    <div v-if="loading" class="absolute top-4 left-1/2 -translate-x-1/2 z-[1001] bg-white px-4 py-2 rounded-lg shadow-md">
      <span class="text-sm text-gray-600">Memuat data...</span>
    </div>

    <!-- Region Filter (top-left) -->
    <div class="absolute top-4 left-4 z-[1000]">
      <div class="relative">
        <!-- Filter Button -->
        <button
          class="flex items-center gap-2 bg-white px-3 py-2 rounded-lg shadow-md hover:bg-gray-50 text-sm"
          @click="toggleRegionFilter"
        >
          <MapPin class="w-4 h-4 text-gray-500" />
          <span class="text-gray-700 max-w-[150px] truncate">{{ filterLabel }}</span>
          <ChevronDown class="w-4 h-4 text-gray-400" :class="{ 'rotate-180': showRegionFilter }" />
          <button
            v-if="hasActiveFilter"
            class="ml-1 p-0.5 hover:bg-gray-200 rounded"
            @click.stop="clearFilter"
          >
            <X class="w-3 h-3 text-gray-500" />
          </button>
        </button>

        <!-- Dropdown Menu -->
        <div
          v-if="showRegionFilter"
          class="absolute top-full left-0 mt-1 bg-white rounded-lg shadow-lg border border-gray-200 min-w-[250px] max-h-[450px] overflow-hidden flex flex-col"
        >
          <!-- Breadcrumb / Current Selection -->
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
          <div class="max-h-[280px] overflow-y-auto flex-1">
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
              @click="applyFilter"
            >
              Terapkan Filter
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Map container -->
    <div ref="mapContainer" class="w-full h-full"></div>

    <!-- Map controls -->
    <div class="absolute right-4 bottom-24 z-[1000] flex flex-col gap-2">
      <!-- Zoom controls -->
      <div class="flex flex-col bg-white rounded-lg shadow-md overflow-hidden">
        <button
          class="w-10 h-10 flex items-center justify-center hover:bg-gray-50 border-b border-gray-200"
          @click="zoomIn"
        >
          <Plus class="w-5 h-5 text-gray-600" />
        </button>
        <button
          class="w-10 h-10 flex items-center justify-center hover:bg-gray-50"
          @click="zoomOut"
        >
          <Minus class="w-5 h-5 text-gray-600" />
        </button>
      </div>
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

.feed-marker {
  z-index: 500 !important;
}

/* Leaflet popup customization for feeds */
.feed-popup-container .leaflet-popup-content-wrapper {
  border-radius: 12px;
  padding: 0;
  overflow: hidden;
}

.feed-popup-container .leaflet-popup-content {
  margin: 0;
  width: 350px !important;
}

.feed-popup-container .leaflet-popup-tip {
  background: white;
}

/* New popup styles */
.feed-popup-new {
  font-family: system-ui, -apple-system, sans-serif;
}

.feed-popup-new .popup-photo {
  margin: -1px -1px 0 -1px;
}

.feed-popup-new .popup-photo img {
  width: 100%;
  max-height: 180px;
  object-fit: cover;
  display: block;
}

.feed-popup-new .popup-location {
  padding: 10px 12px 4px;
  font-size: 13px;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 6px;
}

.feed-popup-new .popup-location .location-name {
  flex: 1;
}

.feed-popup-new .popup-location .gmaps-link {
  margin-left: auto;
  color: #4285f4;
  display: flex;
  align-items: center;
  gap: 3px;
  padding: 3px 6px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
  text-decoration: none;
  transition: background-color 0.2s;
}

.feed-popup-new .popup-location .gmaps-link:hover {
  background-color: #e8f0fe;
  color: #1a73e8;
}

.feed-popup-new .popup-location .gmaps-link .gmaps-text {
  white-space: nowrap;
}

.feed-popup-new .popup-location.clickable {
  cursor: pointer;
  color: #2563eb;
}

.feed-popup-new .popup-location.clickable:hover {
  color: #1d4ed8;
  text-decoration: underline;
}

.feed-popup-new .popup-location.free {
  color: #ea580c;
}

.feed-popup-new .popup-location .location-icon {
  font-size: 14px;
}

.feed-popup-new .popup-location .faskes-badge {
  background: #dcfce7;
  color: #166534;
  font-size: 10px;
  padding: 2px 6px;
  border-radius: 4px;
  font-weight: 500;
}

.feed-popup-new .popup-submitter {
  padding: 0 12px 6px;
  font-size: 11px;
  color: #6b7280;
}

.feed-popup-new .popup-content {
  padding: 0 12px 8px;
  font-size: 12px;
  color: #374151;
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 4;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.feed-popup-new .popup-region {
  padding: 0 12px 8px;
  font-size: 10px;
  color: #9ca3af;
}

.feed-popup-new .popup-bottom {
  padding: 8px 12px;
  background: #f9fafb;
  border-top: 1px solid #e5e7eb;
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  align-items: center;
}

.feed-popup-new .popup-date {
  font-size: 10px;
  color: #6b7280;
}

.feed-popup-new .popup-category {
  font-size: 10px;
  padding: 2px 6px;
  border-radius: 4px;
  font-weight: 500;
}

.feed-popup-new .popup-category.cat-kebutuhan {
  background: #fef3c7;
  color: #92400e;
}

.feed-popup-new .popup-category.cat-followup {
  background: #fee2e2;
  color: #991b1b;
}

.feed-popup-new .popup-category.cat-info {
  background: #f3f4f6;
  color: #374151;
}

.feed-popup-new .popup-tag {
  font-size: 9px;
  padding: 2px 5px;
  border-radius: 3px;
  background: #e0e7ff;
  color: #3730a3;
}
</style>
