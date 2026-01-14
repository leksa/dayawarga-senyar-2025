import { ref, computed } from 'vue'
import { api, type InfrastrukturFeature, type InfrastrukturDetail, type InfrastrukturFilter } from '@/services/api'

const infrastrukturList = ref<InfrastrukturFeature[]>([])
const selectedInfrastruktur = ref<InfrastrukturDetail | null>(null)
const loading = ref(false)
const error = ref<string | null>(null)
const total = ref(0)
const lastUpdate = ref<Date | null>(null)

export function useInfrastruktur() {
  const fetchInfrastruktur = async (filter?: InfrastrukturFilter) => {
    loading.value = true
    error.value = null

    try {
      const response = await api.getInfrastruktur(filter)
      if (response.success && response.data) {
        infrastrukturList.value = response.data.features
        total.value = response.meta?.total ?? response.data.features.length
        lastUpdate.value = new Date()
      } else {
        error.value = response.error?.message ?? 'Failed to fetch infrastruktur'
      }
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to fetch infrastruktur'
    } finally {
      loading.value = false
    }
  }

  const fetchInfrastrukturById = async (id: string) => {
    loading.value = true
    error.value = null

    try {
      const response = await api.getInfrastrukturById(id)
      if (response.success && response.data) {
        selectedInfrastruktur.value = response.data
      } else {
        error.value = response.error?.message ?? 'Failed to fetch infrastruktur'
      }
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to fetch infrastruktur'
    } finally {
      loading.value = false
    }
  }

  const clearSelectedInfrastruktur = () => {
    selectedInfrastruktur.value = null
  }

  const markers = computed(() => {
    return infrastrukturList.value.map(i => ({
      id: i.id,
      name: i.properties.nama,
      type: 'infrastruktur',
      jenis: i.properties.jenis,
      statusJln: i.properties.status_jln,
      statusAkses: i.properties.status_akses,
      statusPenanganan: i.properties.status_penanganan,
      bailey: i.properties.bailey,
      progress: i.properties.progress,
      lat: i.geometry.coordinates[1],
      lng: i.geometry.coordinates[0],
      namaProvinsi: i.properties.nama_provinsi,
      namaKabupaten: i.properties.nama_kabupaten,
      updatedAt: i.properties.updated_at,
    }))
  })

  return {
    infrastrukturList,
    selectedInfrastruktur,
    loading,
    error,
    total,
    lastUpdate,
    markers,
    fetchInfrastruktur,
    fetchInfrastrukturById,
    clearSelectedInfrastruktur,
  }
}
