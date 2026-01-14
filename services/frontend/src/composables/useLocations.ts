import { ref, computed } from 'vue'
import { api, type LocationFeature, type LocationDetail, type LocationFilter } from '@/services/api'

const locations = ref<LocationFeature[]>([])
const selectedLocation = ref<LocationDetail | null>(null)
const loading = ref(false)
const error = ref<string | null>(null)
const total = ref(0)
const lastUpdate = ref<Date | null>(null)

export function useLocations() {
  const fetchLocations = async (filter?: LocationFilter) => {
    loading.value = true
    error.value = null

    try {
      const response = await api.getLocations(filter)
      if (response.success && response.data) {
        locations.value = response.data.features
        total.value = response.meta?.total ?? response.data.features.length
        lastUpdate.value = new Date()
      } else {
        error.value = response.error?.message ?? 'Failed to fetch locations'
      }
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to fetch locations'
    } finally {
      loading.value = false
    }
  }

  const fetchLocationById = async (id: string) => {
    loading.value = true
    error.value = null

    try {
      const response = await api.getLocationById(id)
      if (response.success && response.data) {
        selectedLocation.value = response.data
      } else {
        error.value = response.error?.message ?? 'Failed to fetch location'
      }
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to fetch location'
    } finally {
      loading.value = false
    }
  }

  const clearSelectedLocation = () => {
    selectedLocation.value = null
  }

  const markers = computed(() => {
    return locations.value.map(loc => ({
      id: loc.id,
      name: loc.properties.nama,
      type: loc.properties.type,
      status: loc.properties.status,
      lat: loc.geometry.coordinates[1],
      lng: loc.geometry.coordinates[0],
      alamatSingkat: loc.properties.alamat_singkat,
      namaProvinsi: loc.properties.nama_provinsi,
      namaKotaKab: loc.properties.nama_kota_kab,
      namaKecamatan: loc.properties.nama_kecamatan,
      namaDesa: loc.properties.nama_desa,
      idProvinsi: loc.properties.id_provinsi,
      idKotaKab: loc.properties.id_kota_kab,
      idKecamatan: loc.properties.id_kecamatan,
      idDesa: loc.properties.id_desa,
      jumlahKK: loc.properties.jumlah_kk,
      totalJiwa: loc.properties.total_jiwa,
      jumlahPerempuan: loc.properties.jumlah_perempuan,
      jumlahLaki: loc.properties.jumlah_laki,
      jumlahBalita: loc.properties.jumlah_balita,
      kebutuhanAir: loc.properties.kebutuhan_air,
      kebutuhanAirLiter: loc.properties.kebutuhan_air_liter,
      updatedAt: loc.properties.updated_at,
    }))
  })

  return {
    locations,
    selectedLocation,
    loading,
    error,
    total,
    lastUpdate,
    markers,
    fetchLocations,
    fetchLocationById,
    clearSelectedLocation,
  }
}
