<script setup lang="ts">
import { computed } from 'vue'
import { Users, Building2, Construction } from 'lucide-vue-next'

const props = defineProps<{
  // Visibility flags
  showPosko: boolean
  showFaskes: boolean
  showInfrastruktur: boolean
  // Posko stats
  poskoCount: number
  totalPengungsi: number
  jumlahKk: number
  jumlahPerempuan: number
  jumlahLaki: number
  jumlahBalita: number
  kebutuhanAirLiter: number
  // Faskes stats
  faskesCount: number
  faskesRumahSakit: number
  faskesPuskesmas: number
  faskesPoskoKesDarurat: number
  faskesRsNotOperational: number
  faskesPuskesmasNotOperational: number
  // Infrastruktur stats
  infrastrukturCount: number
  infraJalanSudahDitangani: number
  infraJalanSedangDitangani: number
  infraJembatanSudahDitangani: number
  infraJembatanSedangDitangani: number
  infraBaileyTerpasang: number
  infraBaileySedangDipasang: number
}>()

// Format number with thousand separator
const formatNumber = (num: number): string => {
  return num.toLocaleString('id-ID')
}

// Check if any stats should be shown
const hasVisibleStats = computed(() => {
  return props.showPosko || props.showFaskes || props.showInfrastruktur
})
</script>

<template>
  <div
    v-if="hasVisibleStats"
    class="absolute bottom-6 left-4 z-[1000] flex flex-col gap-2"
  >
    <!-- Posko Pengungsi Stats -->
    <div
      v-if="showPosko"
      class="bg-white/95 backdrop-blur-sm rounded-lg shadow-md px-4 py-2.5 min-w-[200px]"
    >
      <div class="flex items-center gap-2 text-gray-600 mb-1.5">
        <Users class="w-4 h-4 text-orange-500" />
        <span class="text-xs font-medium uppercase tracking-wide">Posko Pengungsi</span>
      </div>
      <div class="space-y-1">
        <div class="flex justify-between items-baseline gap-4">
          <span class="text-xs text-gray-500">Jumlah Posko</span>
          <span class="text-sm font-semibold text-gray-800">{{ formatNumber(poskoCount) }}</span>
        </div>
        <div class="flex justify-between items-baseline gap-4">
          <span class="text-xs text-gray-500">Total Pengungsi</span>
          <span class="text-sm font-semibold text-gray-800">{{ formatNumber(totalPengungsi) }}</span>
        </div>
        <div class="flex justify-between items-baseline gap-4">
          <span class="text-xs text-gray-500">Jumlah KK</span>
          <span class="text-sm font-semibold text-gray-800">{{ formatNumber(jumlahKk) }}</span>
        </div>
        <div class="flex justify-between items-baseline gap-4">
          <span class="text-xs text-gray-500">Perempuan</span>
          <span class="text-sm font-semibold text-gray-800">{{ formatNumber(jumlahPerempuan) }}</span>
        </div>
        <div class="flex justify-between items-baseline gap-4">
          <span class="text-xs text-gray-500">Laki-laki</span>
          <span class="text-sm font-semibold text-gray-800">{{ formatNumber(jumlahLaki) }}</span>
        </div>
        <div class="flex justify-between items-baseline gap-4">
          <span class="text-xs text-gray-500">Bayi & Balita</span>
          <span class="text-sm font-semibold text-gray-800">{{ formatNumber(jumlahBalita) }}</span>
        </div>
        <div class="flex justify-between items-baseline gap-4">
          <span class="text-xs text-gray-500">Kebutuhan Air/Hari</span>
          <span class="text-sm font-semibold text-gray-800">{{ formatNumber(kebutuhanAirLiter) }} L</span>
        </div>
      </div>
    </div>

    <!-- Fasilitas Kesehatan Stats -->
    <div
      v-if="showFaskes"
      class="bg-white/95 backdrop-blur-sm rounded-lg shadow-md px-4 py-2.5 min-w-[200px]"
    >
      <div class="flex items-center gap-2 text-gray-600 mb-1.5">
        <Building2 class="w-4 h-4 text-green-500" />
        <span class="text-xs font-medium uppercase tracking-wide">Fasilitas Kesehatan</span>
      </div>
      <div class="space-y-1">
        <div class="flex justify-between items-baseline gap-4">
          <span class="text-xs text-gray-500">Jumlah Faskes</span>
          <span class="text-sm font-semibold text-gray-800">{{ formatNumber(faskesCount) }}</span>
        </div>
        <div class="flex justify-between items-baseline gap-4">
          <span class="text-xs text-gray-500">Rumah Sakit</span>
          <span class="text-sm font-semibold text-gray-800">{{ formatNumber(faskesRumahSakit) }}</span>
        </div>
        <div class="flex justify-between items-baseline gap-4">
          <span class="text-xs text-gray-500">Puskesmas</span>
          <span class="text-sm font-semibold text-gray-800">{{ formatNumber(faskesPuskesmas) }}</span>
        </div>
        <div class="flex justify-between items-baseline gap-4">
          <span class="text-xs text-gray-500">Posko Kes Darurat</span>
          <span class="text-sm font-semibold text-gray-800">{{ formatNumber(faskesPoskoKesDarurat) }}</span>
        </div>
        <div v-if="faskesRsNotOperational > 0" class="flex justify-between items-baseline gap-4">
          <span class="text-xs text-red-600">RS Tidak Beroperasi</span>
          <span class="text-sm font-semibold text-red-600">{{ formatNumber(faskesRsNotOperational) }}</span>
        </div>
        <div v-if="faskesPuskesmasNotOperational > 0" class="flex justify-between items-baseline gap-4">
          <span class="text-xs text-red-600">Puskesmas Tidak Beroperasi</span>
          <span class="text-sm font-semibold text-red-600">{{ formatNumber(faskesPuskesmasNotOperational) }}</span>
        </div>
      </div>
    </div>

    <!-- Infra Terdampak Stats -->
    <div
      v-if="showInfrastruktur"
      class="bg-white/95 backdrop-blur-sm rounded-lg shadow-md px-4 py-2.5 min-w-[200px]"
    >
      <div class="flex items-center gap-2 text-gray-600 mb-1.5">
        <Construction class="w-4 h-4 text-blue-500" />
        <span class="text-xs font-medium uppercase tracking-wide">Infra Terdampak</span>
      </div>
      <div class="space-y-1">
        <div class="flex justify-between items-baseline gap-4">
          <span class="text-xs text-gray-500">Total Jalan/Jembatan</span>
          <span class="text-sm font-semibold text-gray-800">{{ formatNumber(infrastrukturCount) }}</span>
        </div>
        <div v-if="infraJalanSudahDitangani > 0" class="flex justify-between items-baseline gap-4">
          <span class="text-xs text-green-600">Jalan Sudah Ditangani</span>
          <span class="text-sm font-semibold text-green-600">{{ formatNumber(infraJalanSudahDitangani) }}</span>
        </div>
        <div v-if="infraJalanSedangDitangani > 0" class="flex justify-between items-baseline gap-4">
          <span class="text-xs text-yellow-600">Jalan Sedang Ditangani</span>
          <span class="text-sm font-semibold text-yellow-600">{{ formatNumber(infraJalanSedangDitangani) }}</span>
        </div>
        <div v-if="infraJembatanSudahDitangani > 0" class="flex justify-between items-baseline gap-4">
          <span class="text-xs text-green-600">Jembatan Sudah Ditangani</span>
          <span class="text-sm font-semibold text-green-600">{{ formatNumber(infraJembatanSudahDitangani) }}</span>
        </div>
        <div v-if="infraJembatanSedangDitangani > 0" class="flex justify-between items-baseline gap-4">
          <span class="text-xs text-yellow-600">Jembatan Sedang Ditangani</span>
          <span class="text-sm font-semibold text-yellow-600">{{ formatNumber(infraJembatanSedangDitangani) }}</span>
        </div>
        <div v-if="infraBaileyTerpasang > 0" class="flex justify-between items-baseline gap-4">
          <span class="text-xs text-blue-600">Bailey Terpasang</span>
          <span class="text-sm font-semibold text-blue-600">{{ formatNumber(infraBaileyTerpasang) }}</span>
        </div>
        <div v-if="infraBaileySedangDipasang > 0" class="flex justify-between items-baseline gap-4">
          <span class="text-xs text-cyan-600">Bailey Sedang Dipasang</span>
          <span class="text-sm font-semibold text-cyan-600">{{ formatNumber(infraBaileySedangDipasang) }}</span>
        </div>
      </div>
    </div>
  </div>
</template>
