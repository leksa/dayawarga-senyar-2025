<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { RouterLink } from 'vue-router'
import { Home, Package, Cross, Megaphone, ExternalLink, Info, Map, CloudRain, Mountain, Construction, UtensilsCrossed, BookOpen, Users, Newspaper } from 'lucide-vue-next'
import Checkbox from './ui/Checkbox.vue'

// App version from build
const appVersion = computed(() => import.meta.env.VITE_APP_VERSION || '1.0.0')

interface Layer {
  id: string
  name: string
  icon: any
  color: string
  colorEnabled: string
  enabled: boolean
  available: boolean
}

const emit = defineEmits<{
  'layer-toggle': [layerId: string, enabled: boolean]
}>()

const emergencyLayers = ref<Layer[]>([
  { id: 'shelter', name: 'Posko Pengungsi', icon: Home, color: 'bg-gray-300', colorEnabled: 'bg-blue-500', enabled: true, available: true },
  { id: 'medical', name: 'Fasilitas Kesehatan', icon: Cross, color: 'bg-gray-300', colorEnabled: 'bg-red-500', enabled: false, available: true },
  { id: 'infrastructure', name: 'Jalan Jembatan', icon: Construction, color: 'bg-gray-300', colorEnabled: 'bg-amber-600', enabled: false, available: true },
  { id: 'logistics', name: 'Posko Logistik', icon: Package, color: 'bg-gray-300', colorEnabled: 'bg-cyan-500', enabled: false, available: false },
  { id: 'kitchen', name: 'Dapur Umum', icon: UtensilsCrossed, color: 'bg-gray-300', colorEnabled: 'bg-orange-500', enabled: false, available: false },
])

watch(emergencyLayers, (layers) => {
  layers.forEach(layer => {
    if (layer.available) {
      emit('layer-toggle', layer.id, layer.enabled)
    }
  })
}, { deep: true })

const environmentLayers = ref<Layer[]>([
  { id: 'flood', name: 'Area Banjir', icon: CloudRain, color: 'bg-gray-300', colorEnabled: 'bg-blue-600', enabled: false, available: false },
  { id: 'landslide', name: 'Area Longsor', icon: Mountain, color: 'bg-gray-300', colorEnabled: 'bg-amber-600', enabled: false, available: false },
])

const infrastructureLayers = ref<Layer[]>([
  { id: 'huntara', name: 'Huntara', icon: Home, color: 'bg-gray-300', colorEnabled: 'bg-orange-500', enabled: false, available: false },
])

const toggleLayer = (layer: Layer) => {
  if (layer.available) {
    layer.enabled = !layer.enabled
  }
}

const getLayerColor = (layer: Layer) => {
  if (!layer.available) return 'bg-gray-200'
  return layer.enabled ? layer.colorEnabled : layer.color
}
</script>

<template>
  <!-- Sidebar - Icon only on mobile, full on desktop -->
  <aside class="bg-white border-r border-gray-200 flex flex-col h-full w-14 lg:w-72 flex-shrink-0">
    <!-- Navigation Links - Icon only on mobile -->
    <div class="p-2 lg:p-4 border-b border-gray-200">
      <h3 class="hidden lg:block text-xs font-semibold text-gray-400 uppercase tracking-wider mb-2">Feeds</h3>
      <RouterLink
        to="/feeds"
        class="flex items-center justify-center lg:justify-start gap-2 p-1.5 lg:p-2 rounded-lg bg-blue-50 text-blue-600 hover:bg-blue-100 transition-colors"
        title="Informasi Terbaru"
      >
        <div class="w-7 h-7 lg:w-6 lg:h-6 rounded-full bg-blue-500 flex items-center justify-center flex-shrink-0">
          <Megaphone class="w-3.5 h-3.5 lg:w-3 lg:h-3 text-white" />
        </div>
        <span class="hidden lg:inline text-sm font-medium">Informasi Terbaru</span>
        <span class="hidden lg:inline ml-auto text-gray-400">&rsaquo;</span>
      </RouterLink>
    </div>

    <!-- Scrollable Content Area -->
    <div class="flex-1 overflow-y-auto">
      <!-- Data Kebencanaan -->
      <div class="p-2 lg:p-4">
        <h3 class="hidden lg:block text-xs font-semibold text-gray-400 uppercase tracking-wider mb-2">Data Kebencanaan</h3>
        <div class="space-y-0.5">
          <!-- Peta Bencana -->
          <RouterLink
            to="/"
            class="flex items-center justify-center lg:justify-start gap-2 p-1.5 lg:p-2 rounded-lg hover:bg-gray-50 transition-colors"
            title="Peta Bencana"
          >
            <div class="w-7 h-7 lg:w-6 lg:h-6 rounded-full bg-green-500 flex items-center justify-center flex-shrink-0">
              <Map class="w-3.5 h-3.5 lg:w-3 lg:h-3 text-white" />
            </div>
            <span class="hidden lg:inline flex-1 text-sm text-gray-700">Peta Bencana</span>
            <span class="hidden lg:inline text-gray-400">&rsaquo;</span>
          </RouterLink>

          <!-- Emergency Layers -->
          <div
            v-for="layer in emergencyLayers"
            :key="layer.id"
            :class="[
              'flex items-center justify-center lg:justify-start gap-2 p-1.5 lg:p-2 rounded-lg',
              layer.available ? 'hover:bg-gray-50 cursor-pointer' : 'opacity-50 cursor-not-allowed'
            ]"
            :title="layer.name"
            @click="toggleLayer(layer)"
          >
            <div :class="['w-7 h-7 lg:w-6 lg:h-6 rounded-full flex items-center justify-center flex-shrink-0 transition-colors', getLayerColor(layer)]">
              <component :is="layer.icon" class="w-3.5 h-3.5 lg:w-3 lg:h-3 text-white" />
            </div>
            <span :class="['hidden lg:inline flex-1 text-sm', layer.available ? 'text-gray-700' : 'text-gray-400']">{{ layer.name }}</span>
            <Checkbox v-if="layer.available" v-model="layer.enabled" class="hidden lg:block" @click.stop />
            <span v-else class="hidden lg:inline text-xs text-gray-400 italic">Segera</span>
          </div>
        </div>
      </div>

      <!-- Lingkungan -->
      <div class="p-2 lg:p-4 border-t border-gray-200">
        <h3 class="hidden lg:block text-xs font-semibold text-gray-400 uppercase tracking-wider mb-2">Lingkungan</h3>
        <div class="space-y-0.5">
          <div
            v-for="layer in environmentLayers"
            :key="layer.id"
            :class="[
              'flex items-center justify-center lg:justify-start gap-2 p-1.5 lg:p-2 rounded-lg',
              layer.available ? 'hover:bg-gray-50 cursor-pointer' : 'opacity-50 cursor-not-allowed'
            ]"
            :title="layer.name"
            @click="toggleLayer(layer)"
          >
            <div :class="['w-7 h-7 lg:w-6 lg:h-6 rounded-full flex items-center justify-center flex-shrink-0 transition-colors', getLayerColor(layer)]">
              <component :is="layer.icon" class="w-3.5 h-3.5 lg:w-3 lg:h-3 text-white" />
            </div>
            <span :class="['hidden lg:inline flex-1 text-sm', layer.available ? 'text-gray-700' : 'text-gray-400']">{{ layer.name }}</span>
            <Checkbox v-if="layer.available" v-model="layer.enabled" class="hidden lg:block" @click.stop />
            <span v-else class="hidden lg:inline text-xs text-gray-400 italic">Segera</span>
          </div>
        </div>
      </div>

      <!-- Infrastruktur -->
      <div class="p-2 lg:p-4 border-t border-gray-200">
        <h3 class="hidden lg:block text-xs font-semibold text-gray-400 uppercase tracking-wider mb-2">Infrastruktur</h3>
        <div class="space-y-0.5">
          <div
            v-for="layer in infrastructureLayers"
            :key="layer.id"
            :class="[
              'flex items-center justify-center lg:justify-start gap-2 p-1.5 lg:p-2 rounded-lg',
              layer.available ? 'hover:bg-gray-50 cursor-pointer' : 'opacity-50 cursor-not-allowed'
            ]"
            :title="layer.name"
            @click="toggleLayer(layer)"
          >
            <div :class="['w-7 h-7 lg:w-6 lg:h-6 rounded-full flex items-center justify-center flex-shrink-0 transition-colors', getLayerColor(layer)]">
              <component :is="layer.icon" class="w-3.5 h-3.5 lg:w-3 lg:h-3 text-white" />
            </div>
            <span :class="['hidden lg:inline flex-1 text-sm', layer.available ? 'text-gray-700' : 'text-gray-400']">{{ layer.name }}</span>
            <Checkbox v-if="layer.available" v-model="layer.enabled" class="hidden lg:block" @click.stop />
            <span v-else class="hidden lg:inline text-xs text-gray-400 italic">Segera</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Tentang Section -->
    <div class="p-2 lg:p-4 border-t border-gray-200 bg-white">
      <h3 class="hidden lg:block text-xs font-semibold text-gray-400 uppercase tracking-wider mb-2">Tentang</h3>
      <div class="space-y-0.5">
        <RouterLink
          to="/tentang"
          class="flex items-center justify-center lg:justify-start gap-2 p-1.5 lg:p-2 rounded-lg text-gray-600 hover:bg-gray-100 transition-colors"
          title="Penjelasan"
        >
          <div class="w-7 h-7 lg:w-6 lg:h-6 rounded-full bg-gray-500 flex items-center justify-center flex-shrink-0">
            <Info class="w-3.5 h-3.5 lg:w-3 lg:h-3 text-white" />
          </div>
          <span class="hidden lg:inline text-sm">Penjelasan</span>
        </RouterLink>
        <RouterLink
          to="/pakai-dayawarga"
          class="flex items-center justify-center lg:justify-start gap-2 p-1.5 lg:p-2 rounded-lg text-gray-600 hover:bg-gray-100 transition-colors"
          title="Pakai Dayawarga"
        >
          <div class="w-7 h-7 lg:w-6 lg:h-6 rounded-full bg-blue-500 flex items-center justify-center flex-shrink-0">
            <BookOpen class="w-3.5 h-3.5 lg:w-3 lg:h-3 text-white" />
          </div>
          <span class="hidden lg:inline text-sm">Pakai Dayawarga</span>
        </RouterLink>
        <RouterLink
          to="/belakang-layar"
          class="flex items-center justify-center lg:justify-start gap-2 p-1.5 lg:p-2 rounded-lg text-gray-600 hover:bg-gray-100 transition-colors"
          title="Belakang Layar"
        >
          <div class="w-7 h-7 lg:w-6 lg:h-6 rounded-full bg-purple-500 flex items-center justify-center flex-shrink-0">
            <Users class="w-3.5 h-3.5 lg:w-3 lg:h-3 text-white" />
          </div>
          <span class="hidden lg:inline text-sm">Belakang Layar</span>
        </RouterLink>
        <a
          href="https://stories.dayawarga.com"
          target="_blank"
          class="flex items-center justify-center lg:justify-start gap-2 p-1.5 lg:p-2 rounded-lg text-gray-600 hover:bg-gray-100 transition-colors"
          title="Blog"
        >
          <div class="w-7 h-7 lg:w-6 lg:h-6 rounded-full bg-green-500 flex items-center justify-center flex-shrink-0">
            <Newspaper class="w-3.5 h-3.5 lg:w-3 lg:h-3 text-white" />
          </div>
          <span class="hidden lg:inline text-sm">Blog</span>
          <ExternalLink class="hidden lg:inline w-3 h-3 text-gray-400 ml-auto" />
        </a>
      </div>
    </div>

    <!-- Footer - GitHub icon on mobile, full text on desktop -->
    <div class="p-2 lg:p-4 border-t border-gray-200">
      <!-- Mobile: just GitHub icon (using SVG) -->
      <a
        href="https://github.com/leksa/dayawarga-senyar-2025"
        target="_blank"
        class="lg:hidden flex items-center justify-center p-2 rounded-lg hover:bg-gray-100 transition-colors"
        title="GitHub Repository"
      >
        <svg class="w-6 h-6 text-gray-600" viewBox="0 0 24 24" fill="currentColor">
          <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
        </svg>
      </a>
      <!-- Desktop: full footer text -->
      <p class="hidden lg:block text-xs text-gray-500 leading-relaxed">
        Kolaborasi inisiatif warga dan relawan. Dikembangkan oleh
        <a href="https://dayawarga.com" target="_blank" class="text-blue-500 hover:underline">dayawarga.com</a>.
        Kode sumber terbuka dan data olahan tersedia di
        <a href="https://github.com/leksa/dayawarga-senyar-2025" target="_blank" class="text-blue-500 hover:underline inline-flex items-center gap-1">
          GitHub
          <ExternalLink class="w-3 h-3" />
        </a>.
        <span class="text-gray-400">v{{ appVersion }}</span>
      </p>
    </div>
  </aside>
</template>
