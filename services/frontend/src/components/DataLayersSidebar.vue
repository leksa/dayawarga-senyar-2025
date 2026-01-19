<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { RouterLink } from 'vue-router'
import { Home, Package, Cross, Megaphone, ExternalLink, Info, Map, CloudRain, Mountain, Construction, UtensilsCrossed, BookOpen, Users, Newspaper, HandHelping, Gift } from 'lucide-vue-next'
import Checkbox from './ui/Checkbox.vue'
import { useSidebar } from '@/composables/useSidebar'

const { isOpen: isSidebarOpen, close: closeSidebar } = useSidebar()

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

// Handle link click - close sidebar on mobile
const handleLinkClick = () => {
  closeSidebar()
}
</script>

<template>
  <!-- Mobile: Teleported sidebar overlay -->
  <Teleport to="body">
    <!-- Backdrop - starts below header (top-14 = 3.5rem) -->
    <Transition name="fade">
      <div
        v-if="isSidebarOpen"
        class="lg:hidden fixed top-14 left-0 right-0 bottom-0 bg-black/50 z-[9998]"
        @click="closeSidebar"
      ></div>
    </Transition>

    <!-- Mobile Sidebar -->
    <Transition name="slide">
      <aside
        v-if="isSidebarOpen"
        class="lg:hidden bg-white border-r border-gray-200 flex flex-col fixed top-14 left-0 z-[9999] w-72 h-[calc(100vh-3.5rem)] overflow-y-auto"
      >
        <!-- Navigation Links -->
        <div class="p-4 border-b border-gray-200">
          <h3 class="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-2">Feeds</h3>
          <div class="space-y-0.5">
            <RouterLink
              to="/feeds"
              class="flex items-center gap-2 p-2 rounded-lg bg-blue-50 text-blue-600 hover:bg-blue-100 transition-colors"
              @click="handleLinkClick"
            >
              <div class="w-6 h-6 rounded-full bg-blue-500 flex items-center justify-center flex-shrink-0">
                <Megaphone class="w-3 h-3 text-white" />
              </div>
              <span class="text-sm font-medium">Informasi Terbaru</span>
              <span class="ml-auto text-gray-400">&rsaquo;</span>
            </RouterLink>
            <!-- Sub-menu: Butuh Bantuan -->
            <RouterLink
              to="/feeds?category=kebutuhan"
              class="flex items-center gap-2 p-2 pl-4 rounded-lg hover:bg-orange-50 text-gray-600 hover:text-orange-600 transition-colors"
              @click="handleLinkClick"
            >
              <div class="w-5 h-5 rounded-full bg-orange-500 flex items-center justify-center flex-shrink-0">
                <HandHelping class="w-2.5 h-2.5 text-white" />
              </div>
              <span class="text-sm">Butuh Bantuan</span>
              <span class="ml-auto text-gray-400">&rsaquo;</span>
            </RouterLink>
            <!-- Sub-menu: Terima Bantuan -->
            <RouterLink
              to="/feeds?category=info_bantuan"
              class="flex items-center gap-2 p-2 pl-4 rounded-lg hover:bg-green-50 text-gray-600 hover:text-green-600 transition-colors"
              @click="handleLinkClick"
            >
              <div class="w-5 h-5 rounded-full bg-green-500 flex items-center justify-center flex-shrink-0">
                <Gift class="w-2.5 h-2.5 text-white" />
              </div>
              <span class="text-sm">Terima Bantuan</span>
              <span class="ml-auto text-gray-400">&rsaquo;</span>
            </RouterLink>
          </div>
        </div>

        <!-- Scrollable Content Area -->
        <div class="flex-1 overflow-y-auto">
          <!-- Data Kebencanaan -->
          <div class="p-4">
            <h3 class="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-2">Data Kebencanaan</h3>
            <div class="space-y-0.5">
              <RouterLink
                to="/"
                class="flex items-center gap-2 p-2 rounded-lg hover:bg-gray-50 transition-colors"
                @click="handleLinkClick"
              >
                <div class="w-6 h-6 rounded-full bg-green-500 flex items-center justify-center flex-shrink-0">
                  <Map class="w-3 h-3 text-white" />
                </div>
                <span class="flex-1 text-sm text-gray-700">Peta Bencana</span>
                <span class="text-gray-400">&rsaquo;</span>
              </RouterLink>

              <div
                v-for="layer in emergencyLayers"
                :key="layer.id"
                :class="[
                  'flex items-center gap-2 p-2 rounded-lg',
                  layer.available ? 'hover:bg-gray-50 cursor-pointer' : 'opacity-50 cursor-not-allowed'
                ]"
                @click="toggleLayer(layer)"
              >
                <div :class="['w-6 h-6 rounded-full flex items-center justify-center flex-shrink-0 transition-colors', getLayerColor(layer)]">
                  <component :is="layer.icon" class="w-3 h-3 text-white" />
                </div>
                <span :class="['flex-1 text-sm', layer.available ? 'text-gray-700' : 'text-gray-400']">{{ layer.name }}</span>
                <Checkbox v-if="layer.available" v-model="layer.enabled" @click.stop />
                <span v-else class="text-xs text-gray-400 italic">Segera</span>
              </div>
            </div>
          </div>

          <!-- Lingkungan -->
          <div class="p-4 border-t border-gray-200">
            <h3 class="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-2">Lingkungan</h3>
            <div class="space-y-0.5">
              <div
                v-for="layer in environmentLayers"
                :key="layer.id"
                :class="[
                  'flex items-center gap-2 p-2 rounded-lg',
                  layer.available ? 'hover:bg-gray-50 cursor-pointer' : 'opacity-50 cursor-not-allowed'
                ]"
                @click="toggleLayer(layer)"
              >
                <div :class="['w-6 h-6 rounded-full flex items-center justify-center flex-shrink-0 transition-colors', getLayerColor(layer)]">
                  <component :is="layer.icon" class="w-3 h-3 text-white" />
                </div>
                <span :class="['flex-1 text-sm', layer.available ? 'text-gray-700' : 'text-gray-400']">{{ layer.name }}</span>
                <Checkbox v-if="layer.available" v-model="layer.enabled" @click.stop />
                <span v-else class="text-xs text-gray-400 italic">Segera</span>
              </div>
            </div>
          </div>

          <!-- Infrastruktur -->
          <div class="p-4 border-t border-gray-200">
            <h3 class="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-2">Infrastruktur</h3>
            <div class="space-y-0.5">
              <div
                v-for="layer in infrastructureLayers"
                :key="layer.id"
                :class="[
                  'flex items-center gap-2 p-2 rounded-lg',
                  layer.available ? 'hover:bg-gray-50 cursor-pointer' : 'opacity-50 cursor-not-allowed'
                ]"
                @click="toggleLayer(layer)"
              >
                <div :class="['w-6 h-6 rounded-full flex items-center justify-center flex-shrink-0 transition-colors', getLayerColor(layer)]">
                  <component :is="layer.icon" class="w-3 h-3 text-white" />
                </div>
                <span :class="['flex-1 text-sm', layer.available ? 'text-gray-700' : 'text-gray-400']">{{ layer.name }}</span>
                <Checkbox v-if="layer.available" v-model="layer.enabled" @click.stop />
                <span v-else class="text-xs text-gray-400 italic">Segera</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Tentang Section -->
        <div class="p-4 border-t border-gray-200 bg-white">
          <h3 class="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-2">Tentang</h3>
          <div class="space-y-0.5">
            <RouterLink
              to="/tentang"
              class="flex items-center gap-2 p-2 rounded-lg text-gray-600 hover:bg-gray-100 transition-colors"
              @click="handleLinkClick"
            >
              <div class="w-6 h-6 rounded-full bg-gray-500 flex items-center justify-center flex-shrink-0">
                <Info class="w-3 h-3 text-white" />
              </div>
              <span class="text-sm">Penjelasan</span>
            </RouterLink>
            <RouterLink
              to="/pakai-dayawarga"
              class="flex items-center gap-2 p-2 rounded-lg text-gray-600 hover:bg-gray-100 transition-colors"
              @click="handleLinkClick"
            >
              <div class="w-6 h-6 rounded-full bg-blue-500 flex items-center justify-center flex-shrink-0">
                <BookOpen class="w-3 h-3 text-white" />
              </div>
              <span class="text-sm">Pakai Dayawarga</span>
            </RouterLink>
            <RouterLink
              to="/belakang-layar"
              class="flex items-center gap-2 p-2 rounded-lg text-gray-600 hover:bg-gray-100 transition-colors"
              @click="handleLinkClick"
            >
              <div class="w-6 h-6 rounded-full bg-purple-500 flex items-center justify-center flex-shrink-0">
                <Users class="w-3 h-3 text-white" />
              </div>
              <span class="text-sm">Belakang Layar</span>
            </RouterLink>
            <a
              href="https://stories.dayawarga.com"
              target="_blank"
              class="flex items-center gap-2 p-2 rounded-lg text-gray-600 hover:bg-gray-100 transition-colors"
            >
              <div class="w-6 h-6 rounded-full bg-green-500 flex items-center justify-center flex-shrink-0">
                <Newspaper class="w-3 h-3 text-white" />
              </div>
              <span class="text-sm">Blog</span>
              <ExternalLink class="w-3 h-3 text-gray-400 ml-auto" />
            </a>
          </div>
        </div>

        <!-- Footer -->
        <div class="p-4 border-t border-gray-200">
          <p class="text-xs text-gray-500 leading-relaxed">
            Kolaborasi inisiatif warga dan relawan. Dikembangkan oleh
            <a href="https://dayawarga.com" target="_blank" class="text-blue-500 hover:underline">dayawarga.com</a>.
            <span class="text-gray-400">v{{ appVersion }}</span>
          </p>
        </div>
      </aside>
    </Transition>
  </Teleport>

  <!-- Desktop Sidebar - always visible on lg screens -->
  <aside class="hidden lg:flex bg-white border-r border-gray-200 flex-col flex-shrink-0 w-72 h-full">
    <!-- Navigation Links -->
    <div class="p-4 border-b border-gray-200">
      <h3 class="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-2">Feeds</h3>
      <div class="space-y-0.5">
        <RouterLink
          to="/feeds"
          class="flex items-center gap-2 p-2 rounded-lg bg-blue-50 text-blue-600 hover:bg-blue-100 transition-colors"
        >
          <div class="w-6 h-6 rounded-full bg-blue-500 flex items-center justify-center flex-shrink-0">
            <Megaphone class="w-3 h-3 text-white" />
          </div>
          <span class="text-sm font-medium">Informasi Terbaru</span>
          <span class="ml-auto text-gray-400">&rsaquo;</span>
        </RouterLink>
        <!-- Sub-menu: Butuh Bantuan -->
        <RouterLink
          to="/feeds?category=kebutuhan"
          class="flex items-center gap-2 p-2 pl-4 rounded-lg hover:bg-orange-50 text-gray-600 hover:text-orange-600 transition-colors"
        >
          <div class="w-5 h-5 rounded-full bg-orange-500 flex items-center justify-center flex-shrink-0">
            <HandHelping class="w-2.5 h-2.5 text-white" />
          </div>
          <span class="text-sm">Butuh Bantuan</span>
          <span class="ml-auto text-gray-400">&rsaquo;</span>
        </RouterLink>
        <!-- Sub-menu: Terima Bantuan -->
        <RouterLink
          to="/feeds?category=info_bantuan"
          class="flex items-center gap-2 p-2 pl-4 rounded-lg hover:bg-green-50 text-gray-600 hover:text-green-600 transition-colors"
        >
          <div class="w-5 h-5 rounded-full bg-green-500 flex items-center justify-center flex-shrink-0">
            <Gift class="w-2.5 h-2.5 text-white" />
          </div>
          <span class="text-sm">Terima Bantuan</span>
          <span class="ml-auto text-gray-400">&rsaquo;</span>
        </RouterLink>
      </div>
    </div>

    <!-- Scrollable Content Area -->
    <div class="flex-1 overflow-y-auto">
      <!-- Data Kebencanaan -->
      <div class="p-4">
        <h3 class="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-2">Data Kebencanaan</h3>
        <div class="space-y-0.5">
          <RouterLink
            to="/"
            class="flex items-center gap-2 p-2 rounded-lg hover:bg-gray-50 transition-colors"
          >
            <div class="w-6 h-6 rounded-full bg-green-500 flex items-center justify-center flex-shrink-0">
              <Map class="w-3 h-3 text-white" />
            </div>
            <span class="flex-1 text-sm text-gray-700">Peta Bencana</span>
            <span class="text-gray-400">&rsaquo;</span>
          </RouterLink>

          <div
            v-for="layer in emergencyLayers"
            :key="layer.id"
            :class="[
              'flex items-center gap-2 p-2 rounded-lg',
              layer.available ? 'hover:bg-gray-50 cursor-pointer' : 'opacity-50 cursor-not-allowed'
            ]"
            @click="toggleLayer(layer)"
          >
            <div :class="['w-6 h-6 rounded-full flex items-center justify-center flex-shrink-0 transition-colors', getLayerColor(layer)]">
              <component :is="layer.icon" class="w-3 h-3 text-white" />
            </div>
            <span :class="['flex-1 text-sm', layer.available ? 'text-gray-700' : 'text-gray-400']">{{ layer.name }}</span>
            <Checkbox v-if="layer.available" v-model="layer.enabled" @click.stop />
            <span v-else class="text-xs text-gray-400 italic">Segera</span>
          </div>
        </div>
      </div>

      <!-- Lingkungan -->
      <div class="p-4 border-t border-gray-200">
        <h3 class="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-2">Lingkungan</h3>
        <div class="space-y-0.5">
          <div
            v-for="layer in environmentLayers"
            :key="layer.id"
            :class="[
              'flex items-center gap-2 p-2 rounded-lg',
              layer.available ? 'hover:bg-gray-50 cursor-pointer' : 'opacity-50 cursor-not-allowed'
            ]"
            @click="toggleLayer(layer)"
          >
            <div :class="['w-6 h-6 rounded-full flex items-center justify-center flex-shrink-0 transition-colors', getLayerColor(layer)]">
              <component :is="layer.icon" class="w-3 h-3 text-white" />
            </div>
            <span :class="['flex-1 text-sm', layer.available ? 'text-gray-700' : 'text-gray-400']">{{ layer.name }}</span>
            <Checkbox v-if="layer.available" v-model="layer.enabled" @click.stop />
            <span v-else class="text-xs text-gray-400 italic">Segera</span>
          </div>
        </div>
      </div>

      <!-- Infrastruktur -->
      <div class="p-4 border-t border-gray-200">
        <h3 class="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-2">Infrastruktur</h3>
        <div class="space-y-0.5">
          <div
            v-for="layer in infrastructureLayers"
            :key="layer.id"
            :class="[
              'flex items-center gap-2 p-2 rounded-lg',
              layer.available ? 'hover:bg-gray-50 cursor-pointer' : 'opacity-50 cursor-not-allowed'
            ]"
            @click="toggleLayer(layer)"
          >
            <div :class="['w-6 h-6 rounded-full flex items-center justify-center flex-shrink-0 transition-colors', getLayerColor(layer)]">
              <component :is="layer.icon" class="w-3 h-3 text-white" />
            </div>
            <span :class="['flex-1 text-sm', layer.available ? 'text-gray-700' : 'text-gray-400']">{{ layer.name }}</span>
            <Checkbox v-if="layer.available" v-model="layer.enabled" @click.stop />
            <span v-else class="text-xs text-gray-400 italic">Segera</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Tentang Section -->
    <div class="p-4 border-t border-gray-200 bg-white">
      <h3 class="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-2">Tentang</h3>
      <div class="space-y-0.5">
        <RouterLink
          to="/tentang"
          class="flex items-center gap-2 p-2 rounded-lg text-gray-600 hover:bg-gray-100 transition-colors"
        >
          <div class="w-6 h-6 rounded-full bg-gray-500 flex items-center justify-center flex-shrink-0">
            <Info class="w-3 h-3 text-white" />
          </div>
          <span class="text-sm">Penjelasan</span>
        </RouterLink>
        <RouterLink
          to="/pakai-dayawarga"
          class="flex items-center gap-2 p-2 rounded-lg text-gray-600 hover:bg-gray-100 transition-colors"
        >
          <div class="w-6 h-6 rounded-full bg-blue-500 flex items-center justify-center flex-shrink-0">
            <BookOpen class="w-3 h-3 text-white" />
          </div>
          <span class="text-sm">Pakai Dayawarga</span>
        </RouterLink>
        <RouterLink
          to="/belakang-layar"
          class="flex items-center gap-2 p-2 rounded-lg text-gray-600 hover:bg-gray-100 transition-colors"
        >
          <div class="w-6 h-6 rounded-full bg-purple-500 flex items-center justify-center flex-shrink-0">
            <Users class="w-3 h-3 text-white" />
          </div>
          <span class="text-sm">Belakang Layar</span>
        </RouterLink>
        <a
          href="https://stories.dayawarga.com"
          target="_blank"
          class="flex items-center gap-2 p-2 rounded-lg text-gray-600 hover:bg-gray-100 transition-colors"
        >
          <div class="w-6 h-6 rounded-full bg-green-500 flex items-center justify-center flex-shrink-0">
            <Newspaper class="w-3 h-3 text-white" />
          </div>
          <span class="text-sm">Blog</span>
          <ExternalLink class="w-3 h-3 text-gray-400 ml-auto" />
        </a>
      </div>
    </div>

    <!-- Footer -->
    <div class="p-4 border-t border-gray-200">
      <p class="text-xs text-gray-500 leading-relaxed">
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

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.slide-enter-active,
.slide-leave-active {
  transition: transform 0.3s ease;
}

.slide-enter-from,
.slide-leave-to {
  transform: translateX(-100%);
}
</style>
