import { ref, computed } from 'vue'
import { api, type FaskesFeature, type FaskesDetail, type FaskesFilter } from '@/services/api'

const faskesList = ref<FaskesFeature[]>([])
const selectedFaskes = ref<FaskesDetail | null>(null)
const loading = ref(false)
const error = ref<string | null>(null)
const total = ref(0)
const lastUpdate = ref<Date | null>(null)

export function useFaskes() {
  const fetchFaskes = async (filter?: FaskesFilter) => {
    loading.value = true
    error.value = null

    try {
      const response = await api.getFaskes(filter)
      if (response.success && response.data) {
        faskesList.value = response.data.features
        total.value = response.meta?.total ?? response.data.features.length
        lastUpdate.value = new Date()
      } else {
        error.value = response.error?.message ?? 'Failed to fetch faskes'
      }
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to fetch faskes'
    } finally {
      loading.value = false
    }
  }

  const fetchFaskesById = async (id: string) => {
    loading.value = true
    error.value = null

    try {
      const response = await api.getFaskesById(id)
      if (response.success && response.data) {
        selectedFaskes.value = response.data
      } else {
        error.value = response.error?.message ?? 'Failed to fetch faskes'
      }
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to fetch faskes'
    } finally {
      loading.value = false
    }
  }

  const clearSelectedFaskes = () => {
    selectedFaskes.value = null
  }

  const markers = computed(() => {
    return faskesList.value.map(f => ({
      id: f.id,
      name: f.properties.nama,
      type: 'faskes',
      jenisFaskes: f.properties.jenis_faskes,
      statusFaskes: f.properties.status_faskes,
      kondisiFaskes: f.properties.kondisi_faskes,
      lat: f.geometry.coordinates[1],
      lng: f.geometry.coordinates[0],
      alamatSingkat: f.properties.alamat_singkat,
      updatedAt: f.properties.updated_at,
    }))
  })

  return {
    faskesList,
    selectedFaskes,
    loading,
    error,
    total,
    lastUpdate,
    markers,
    fetchFaskes,
    fetchFaskesById,
    clearSelectedFaskes,
  }
}
