<script setup lang="ts">
import { ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { Home, Droplets, Cross, Megaphone, ExternalLink, Info, Map, CloudRain, Mountain, Construction, X } from 'lucide-vue-next'
import Checkbox from './ui/Checkbox.vue'
import { useSidebar } from '@/composables/useSidebar'

const { isOpen, close } = useSidebar()

interface Layer {
  id: string
  name: string
  icon: any
  color: string
  enabled: boolean
  available: boolean // true = dapat diklik, false = disabled (belum ada data)
}

const emit = defineEmits<{
  'layer-toggle': [layerId: string, enabled: boolean]
}>()

const emergencyLayers = ref<Layer[]>([
  { id: 'shelter', name: 'Titik Posko', icon: Home, color: 'bg-blue-500', enabled: true, available: true },
  { id: 'water', name: 'Air Bersih', icon: Droplets, color: 'bg-cyan-500', enabled: false, available: false },
  { id: 'medical', name: 'Fasilitas Kesehatan', icon: Cross, color: 'bg-red-500', enabled: false, available: false },
])

// Watch for changes in emergency layers and emit events
watch(emergencyLayers, (layers) => {
  layers.forEach(layer => {
    if (layer.available) {
      emit('layer-toggle', layer.id, layer.enabled)
    }
  })
}, { deep: true })

const environmentLayers = ref<Layer[]>([
  { id: 'flood', name: 'Area Banjir', icon: CloudRain, color: 'bg-blue-600', enabled: false, available: false },
  { id: 'landslide', name: 'Area Longsor', icon: Mountain, color: 'bg-amber-600', enabled: false, available: false },
])

const infrastructureLayers = ref<Layer[]>([
  { id: 'bridge', name: 'Jembatan', icon: Construction, color: 'bg-gray-500', enabled: false, available: false },
  { id: 'huntara', name: 'Huntara', icon: Home, color: 'bg-orange-500', enabled: false, available: false },
])

// Close sidebar on navigation (mobile)
const handleNavClick = () => {
  close()
}
</script>

<template>
  <!-- Mobile overlay backdrop -->
  <div
    v-if="isOpen"
    class="fixed inset-0 bg-black/50 z-40 lg:hidden"
    @click="close"
  />

  <!-- Sidebar -->
  <aside
    :class="[
      'bg-white border-r border-gray-200 flex flex-col h-full z-50',
      // Mobile: fixed overlay, hidden by default
      'fixed inset-y-0 left-0 w-72 transform transition-transform duration-300 ease-in-out lg:transform-none',
      // Desktop: static
      'lg:relative lg:w-72',
      // Toggle visibility on mobile
      isOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'
    ]"
  >
    <!-- Mobile close button -->
    <div class="lg:hidden flex items-center justify-between p-4 border-b border-gray-200">
      <span class="font-semibold text-gray-900">Menu</span>
      <button
        @click="close"
        class="p-2 -mr-2 rounded-lg hover:bg-gray-100 transition-colors"
        aria-label="Close menu"
      >
        <X class="w-5 h-5 text-gray-600" />
      </button>
    </div>

    <!-- Feeds Section -->
    <div class="p-4 border-b border-gray-200">
      <h3 class="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-2">Feeds</h3>
      <RouterLink
        to="/feeds"
        @click="handleNavClick"
        class="w-full flex items-center gap-3 p-2 rounded-lg bg-blue-50 text-blue-600 hover:bg-blue-100 transition-colors"
      >
        <div class="w-8 h-8 rounded-full bg-blue-500 flex items-center justify-center">
          <Megaphone class="w-4 h-4 text-white" />
        </div>
        <span class="font-medium">Informasi Terbaru</span>
        <span class="ml-auto text-gray-400">&rsaquo;</span>
      </RouterLink>
    </div>

    <!-- Scrollable Content Area -->
    <div class="flex-1 overflow-y-auto">
      <!-- Data Kebencanaan -->
      <div class="p-4">
        <h3 class="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-2">Data Kebencanaan</h3>
        <div class="space-y-1">
          <!-- Peta Bencana (link ke home) -->
          <RouterLink
            to="/"
            @click="handleNavClick"
            class="flex items-center gap-3 p-2 rounded-lg hover:bg-gray-50 transition-colors"
          >
            <div class="w-8 h-8 rounded-full bg-green-500 flex items-center justify-center">
              <Map class="w-4 h-4 text-white" />
            </div>
            <span class="flex-1 text-gray-700 font-medium">Peta Bencana</span>
            <span class="text-gray-400">&rsaquo;</span>
          </RouterLink>

          <!-- Emergency Layers -->
          <div
            v-for="layer in emergencyLayers"
            :key="layer.id"
            :class="[
              'flex items-center gap-3 p-2 rounded-lg',
              layer.available
                ? 'hover:bg-gray-50 cursor-pointer'
                : 'opacity-50 cursor-not-allowed'
            ]"
          >
            <div :class="['w-8 h-8 rounded-full flex items-center justify-center', layer.available ? layer.color : 'bg-gray-300']">
              <component :is="layer.icon" class="w-4 h-4 text-white" />
            </div>
            <span :class="['flex-1', layer.available ? 'text-gray-700' : 'text-gray-400']">{{ layer.name }}</span>
            <Checkbox v-if="layer.available" v-model="layer.enabled" />
            <span v-else class="text-xs text-gray-400 italic">Segera</span>
          </div>
        </div>
      </div>

      <!-- Lingkungan -->
      <div class="p-4 border-t border-gray-200">
        <h3 class="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-2">Lingkungan</h3>
        <div class="space-y-1">
          <div
            v-for="layer in environmentLayers"
            :key="layer.id"
            :class="[
              'flex items-center gap-3 p-2 rounded-lg',
              layer.available
                ? 'hover:bg-gray-50 cursor-pointer'
                : 'opacity-50 cursor-not-allowed'
            ]"
          >
            <div :class="['w-8 h-8 rounded-full flex items-center justify-center', layer.available ? layer.color : 'bg-gray-300']">
              <component :is="layer.icon" class="w-4 h-4 text-white" />
            </div>
            <span :class="['flex-1', layer.available ? 'text-gray-700' : 'text-gray-400']">{{ layer.name }}</span>
            <Checkbox v-if="layer.available" v-model="layer.enabled" />
            <span v-else class="text-xs text-gray-400 italic">Segera</span>
          </div>
        </div>
      </div>

      <!-- Infrastruktur -->
      <div class="p-4 border-t border-gray-200">
        <h3 class="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-2">Infrastruktur</h3>
        <div class="space-y-1">
          <div
            v-for="layer in infrastructureLayers"
            :key="layer.id"
            :class="[
              'flex items-center gap-3 p-2 rounded-lg',
              layer.available
                ? 'hover:bg-gray-50 cursor-pointer'
                : 'opacity-50 cursor-not-allowed'
            ]"
          >
            <div :class="['w-8 h-8 rounded-full flex items-center justify-center', layer.available ? layer.color : 'bg-gray-300']">
              <component :is="layer.icon" class="w-4 h-4 text-white" />
            </div>
            <span :class="['flex-1', layer.available ? 'text-gray-700' : 'text-gray-400']">{{ layer.name }}</span>
            <Checkbox v-if="layer.available" v-model="layer.enabled" />
            <span v-else class="text-xs text-gray-400 italic">Segera</span>
          </div>
        </div>
      </div>

    </div>

    <!-- Tentang Section (sticky at bottom) -->
    <div class="p-4 border-t border-gray-200 bg-white">
      <h3 class="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-2">Tentang</h3>
      <RouterLink
        to="/tentang"
        @click="handleNavClick"
        class="w-full flex items-center gap-3 p-2 rounded-lg bg-gray-50 text-gray-700 hover:bg-gray-100 transition-colors"
      >
        <div class="w-8 h-8 rounded-full bg-gray-600 flex items-center justify-center">
          <Info class="w-4 h-4 text-white" />
        </div>
        <span class="font-medium">Tentang</span>
        <span class="ml-auto text-gray-400">&rsaquo;</span>
      </RouterLink>
    </div>

    <!-- Footer - hidden on mobile -->
    <div class="hidden lg:block p-4 border-t border-gray-200">
      <p class="text-xs text-gray-500 leading-relaxed">
        Kolaborasi inisiatif warga dan relawan. Dikembangkan oleh
        <a href="https://dayawarga.com" target="_blank" class="text-blue-500 hover:underline">dayawarga.com</a>.
        Kode sumber terbuka dan data olahan tersedia di
        <a href="https://github.com/leksa/dayawarga-senyar-2025" target="_blank" class="text-blue-500 hover:underline inline-flex items-center gap-1">
          GitHub
          <ExternalLink class="w-3 h-3" />
        </a>.
      </p>
    </div>
  </aside>
</template>
