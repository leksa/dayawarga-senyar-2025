<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft, MapPin, Clock, User, Share2, Navigation, Building2, Megaphone, ChevronLeft, ChevronRight, ExternalLink, ImageOff, X } from 'lucide-vue-next'
import Badge from '@/components/ui/Badge.vue'
import Button from '@/components/ui/Button.vue'
import { api, type Feed, type FeedPhoto } from '@/services/api'
import { categoryColors, tagColor, getCategoryLabel, getTagLabel } from '@/lib/feedHelpers'

const route = useRoute()
const router = useRouter()

const feed = ref<Feed | null>(null)
const loading = ref(true)
const error = ref<string | null>(null)
const currentPhotoIndex = ref(0)
const showShareMenu = ref(false)
const photoLoading = ref(true)
const photoError = ref(false)
const shareButtonRef = ref<HTMLElement | null>(null)
const dropdownPosition = ref({ top: 0, right: 0 })
const photoGalleryRef = ref<HTMLElement | null>(null)

// Touch/swipe gesture state
const touchStartX = ref(0)
const touchStartY = ref(0)
const touchEndX = ref(0)
const isSwiping = ref(false)
const swipeOffset = ref(0)
const minSwipeDistance = 50

// Detect if mobile
const isMobile = ref(false)

const checkMobile = () => {
  isMobile.value = window.innerWidth < 640
}

onMounted(() => {
  checkMobile()
  window.addEventListener('resize', checkMobile)
})

onUnmounted(() => {
  window.removeEventListener('resize', checkMobile)
})

// Touch handlers for swipe gestures
const onTouchStart = (e: TouchEvent) => {
  touchStartX.value = e.changedTouches[0].screenX
  touchStartY.value = e.changedTouches[0].screenY
  isSwiping.value = true
  swipeOffset.value = 0
}

const onTouchMove = (e: TouchEvent) => {
  if (!isSwiping.value) return
  const currentX = e.changedTouches[0].screenX
  const currentY = e.changedTouches[0].screenY
  const diffX = currentX - touchStartX.value
  const diffY = Math.abs(currentY - touchStartY.value)
  
  // Only track horizontal swipes (ignore vertical scrolling)
  if (Math.abs(diffX) > diffY) {
    swipeOffset.value = diffX * 0.3 // Dampen the movement
    e.preventDefault()
  }
}

const onTouchEnd = (e: TouchEvent) => {
  if (!isSwiping.value) return
  touchEndX.value = e.changedTouches[0].screenX
  isSwiping.value = false
  handleSwipe()
  swipeOffset.value = 0
}

const handleSwipe = () => {
  const distance = touchStartX.value - touchEndX.value
  if (Math.abs(distance) > minSwipeDistance) {
    if (distance > 0) {
      nextPhoto()
    } else {
      prevPhoto()
    }
  }
}

const updateDropdownPosition = () => {
  if (shareButtonRef.value) {
    const rect = shareButtonRef.value.getBoundingClientRect()
    dropdownPosition.value = {
      top: rect.bottom + 8,
      right: window.innerWidth - rect.right
    }
  }
}

const feedCode = computed(() => route.params.code as string)

const updateMetaTags = (feedData: Feed) => {
  const loc = feedData.location_name || feedData.faskes_name || feedData.region?.desa || 'Laporan'
  const title = `${getCategoryLabel(feedData.category)} - ${loc} | Dayawarga`
  const description = feedData.content.substring(0, 200) + (feedData.content.length > 200 ? '...' : '')
  const url = `${window.location.origin}/feeds/${feedData.short_code || feedData.id}`
  const image = feedData.photos?.length ? getPhotoUrl(feedData.photos[0]) : 'https://dayawarga.id/og-image.png'
  
  document.title = title
  
  const metaUpdates: Record<string, string> = {
    'description': description,
    'og:title': title,
    'og:description': description,
    'og:url': url,
    'og:image': image,
    'twitter:title': title,
    'twitter:description': description,
    'twitter:url': url,
    'twitter:image': image,
  }
  
  Object.entries(metaUpdates).forEach(([name, content]) => {
    let meta = document.querySelector(`meta[property="${name}"]`) || document.querySelector(`meta[name="${name}"]`)
    if (meta) {
      meta.setAttribute('content', content)
    }
  })
}

const isUUID = (str: string): boolean => {
  const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i
  return uuidRegex.test(str)
}

const fetchFeed = async () => {
  loading.value = true
  error.value = null
  
  try {
    const code = feedCode.value
    const response = isUUID(code) 
      ? await api.getFeedById(code)
      : await api.getFeedByShortCode(code)
    
    if (response.success && response.data) {
      feed.value = response.data
      updateMetaTags(response.data)
    } else {
      error.value = 'Feed tidak ditemukan'
    }
  } catch (e) {
    error.value = 'Gagal memuat data'
    console.error('Failed to fetch feed:', e)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchFeed()
})

watch(feedCode, () => {
  fetchFeed()
})

const getPhotoUrl = (photo: FeedPhoto) => {
  const baseUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1'
  return `${baseUrl}/feed-photos/${photo.id}/file`
}

const formatTimestamp = (isoString: string): string => {
  const date = new Date(isoString)
  const options: Intl.DateTimeFormatOptions = {
    weekday: 'long',
    day: 'numeric',
    month: 'long',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    timeZone: 'Asia/Jakarta'
  }
  return date.toLocaleDateString('id-ID', options) + ' WIB'
}



const goBack = () => {
  if (window.history.length > 1) {
    router.back()
  } else {
    router.push('/feeds')
  }
}

const goToMap = () => {
  if (!feed.value) return
  
  if (feed.value.location_id) {
    router.push({ path: '/', query: { location: feed.value.location_id } })
  } else if (feed.value.faskes_id) {
    router.push({ path: '/', query: { faskes: feed.value.faskes_id } })
  } else if (feed.value.coordinates) {
    const query: Record<string, string> = {
      lat: feed.value.coordinates[1].toString(),
      lng: feed.value.coordinates[0].toString(),
      zoom: '18',
      feed: feed.value.id
    }
    if (feed.value.region?.id_desa) {
      query.desa = feed.value.region.id_desa
      if (feed.value.region.desa) query.desa_name = feed.value.region.desa
      if (feed.value.region.kecamatan) query.kecamatan = feed.value.region.kecamatan
      if (feed.value.region.kota_kab) query.kotakab = feed.value.region.kota_kab
    }
    router.push({ path: '/', query })
  }
}

const nextPhoto = () => {
  if (feed.value?.photos && currentPhotoIndex.value < feed.value.photos.length - 1) {
    photoLoading.value = true
    photoError.value = false
    currentPhotoIndex.value++
  }
}

const prevPhoto = () => {
  if (currentPhotoIndex.value > 0) {
    photoLoading.value = true
    photoError.value = false
    currentPhotoIndex.value--
  }
}

const onPhotoLoad = () => {
  photoLoading.value = false
  photoError.value = false
}

const onPhotoError = () => {
  photoLoading.value = false
  photoError.value = true
}

const shareUrl = computed(() => {
  const code = feed.value?.short_code || feedCode.value
  return `${window.location.origin}/feeds/${code}`
})

const shareTitle = computed(() => {
  if (!feed.value) return 'Dayawarga - Update Informasi'
  const loc = feed.value.location_name || feed.value.faskes_name || feed.value.region?.desa || 'Laporan'
  return `${getCategoryLabel(feed.value.category)} - ${loc} | Dayawarga`
})

const shareDescription = computed(() => {
  if (!feed.value) return ''
  return feed.value.content.substring(0, 200) + (feed.value.content.length > 200 ? '...' : '')
})

// Native share API (for mobile)
const useNativeShare = async () => {
  if (typeof navigator.share === 'function') {
    try {
      await navigator.share({
        title: shareTitle.value,
        text: shareDescription.value,
        url: shareUrl.value,
      })
      showShareMenu.value = false
      return true
    } catch {
      // User cancelled or error
      return false
    }
  }
  return false
}

const openShareMenu = async () => {
  // Try native share first on mobile
  if (isMobile.value && typeof navigator.share === 'function') {
    const shared = await useNativeShare()
    if (shared) return
  }
  
  // Fall back to custom menu
  updateDropdownPosition()
  showShareMenu.value = !showShareMenu.value
}

const shareToTwitter = () => {
  const title = shareTitle.value
  const urlLength = 23
  const maxTextLength = 280 - urlLength - 5
  let text = title
  
  if (text.length < maxTextLength) {
    const remainingChars = maxTextLength - text.length - 2
    if (remainingChars > 20) {
      const desc = feed.value?.content.substring(0, remainingChars - 3) + '...'
      text = `${title}\n\n${desc}`
    }
  }
  
  if (text.length > maxTextLength) {
    text = text.substring(0, maxTextLength - 3) + '...'
  }
  
  const url = `https://twitter.com/intent/tweet?text=${encodeURIComponent(text)}&url=${encodeURIComponent(shareUrl.value)}`
  window.open(url, '_blank', 'width=600,height=400')
  showShareMenu.value = false
}

const shareToFacebook = () => {
  const url = `https://www.facebook.com/sharer/sharer.php?u=${encodeURIComponent(shareUrl.value)}`
  window.open(url, '_blank', 'width=600,height=400')
  showShareMenu.value = false
}

const shareToWhatsapp = () => {
  const text = `${shareTitle.value}\n\n${shareDescription.value}\n\n${shareUrl.value}`
  const url = `https://wa.me/?text=${encodeURIComponent(text)}`
  window.open(url, '_blank')
  showShareMenu.value = false
}

const shareToThreads = () => {
  const text = `${shareTitle.value}\n\n${shareDescription.value}\n\n${shareUrl.value}`
  const url = `https://www.threads.net/intent/post?text=${encodeURIComponent(text)}`
  window.open(url, '_blank', 'width=600,height=600')
  showShareMenu.value = false
}

const shareToInstagram = async () => {
  const text = `${shareTitle.value}\n\n${shareDescription.value}\n\n${shareUrl.value}`
  try {
    await navigator.clipboard.writeText(text)
    alert('Teks telah disalin! Buka Instagram dan paste di caption atau story.')
  } catch {
    alert('Buka Instagram dan bagikan link: ' + shareUrl.value)
  }
  showShareMenu.value = false
}

const copyLink = async () => {
  try {
    await navigator.clipboard.writeText(shareUrl.value)
    alert('Link berhasil disalin!')
  } catch {
    const input = document.createElement('input')
    input.value = shareUrl.value
    document.body.appendChild(input)
    input.select()
    document.execCommand('copy')
    document.body.removeChild(input)
    alert('Link berhasil disalin!')
  }
  showShareMenu.value = false
}

const locationName = computed(() => {
  if (!feed.value) return ''
  return feed.value.location_name || feed.value.faskes_name || ''
})

const regionText = computed(() => {
  if (!feed.value?.region) return ''
  const parts = []
  if (feed.value.region.desa) parts.push(feed.value.region.desa)
  if (feed.value.region.kecamatan) parts.push(feed.value.region.kecamatan)
  if (feed.value.region.kota_kab) parts.push(feed.value.region.kota_kab)
  return parts.join(', ')
})

const typeLabel = computed(() => {
  if (!feed.value) return { icon: Megaphone, text: 'Laporan', color: 'text-orange-600' }
  if (feed.value.location_id) return { icon: MapPin, text: 'Posko', color: 'text-blue-600' }
  if (feed.value.faskes_id) return { icon: Building2, text: 'Faskes', color: 'text-green-600' }
  return { icon: Megaphone, text: 'Laporan Situasi', color: 'text-orange-600' }
})

const openGoogleMaps = () => {
  if (feed.value?.coordinates) {
    window.open(`https://www.google.com/maps?q=${feed.value.coordinates[1]},${feed.value.coordinates[0]}`, '_blank')
  }
}

const tags = computed(() => {
  if (!feed.value?.type) return []
  return feed.value.type.split(/[\s,]+/).filter(t => t)
})

// Photo dot indicators
const goToPhoto = (index: number) => {
  if (index !== currentPhotoIndex.value) {
    photoLoading.value = true
    photoError.value = false
    currentPhotoIndex.value = index
  }
}
</script>

<template>
  <div class="flex-1 overflow-y-auto bg-gray-50">
    <!-- Share Menu Modal -->
    <Teleport to="body">
      <Transition name="share-menu">
        <div v-if="showShareMenu" class="fixed inset-0 z-[9999]">
          <!-- Backdrop -->
          <div 
            class="absolute inset-0 bg-black/50 transition-opacity"
            @click="showShareMenu = false"
          ></div>
          
          <!-- Mobile: Bottom Sheet -->
          <div class="sm:hidden fixed bottom-0 left-0 right-0 bg-white rounded-t-2xl shadow-2xl animate-slide-up pb-safe">
            <!-- Handle bar -->
            <div class="flex justify-center pt-3 pb-2">
              <div class="w-10 h-1 bg-gray-300 rounded-full"></div>
            </div>
            
            <!-- Header -->
            <div class="px-4 pb-3 flex items-center justify-between border-b border-gray-100">
              <h3 class="text-lg font-semibold text-gray-900">Bagikan</h3>
              <button 
                @click="showShareMenu = false"
                class="w-8 h-8 flex items-center justify-center rounded-full hover:bg-gray-100 active:bg-gray-200 transition-colors"
              >
                <X class="w-5 h-5 text-gray-500" />
              </button>
            </div>
            
            <!-- Share options grid -->
            <div class="p-4 grid grid-cols-4 gap-4">
              <button 
                @click="shareToWhatsapp"
                class="flex flex-col items-center gap-2 p-3 rounded-xl hover:bg-gray-50 active:bg-gray-100 transition-colors"
              >
                <div class="w-12 h-12 bg-green-500 rounded-full flex items-center justify-center">
                  <span class="text-2xl">💬</span>
                </div>
                <span class="text-xs text-gray-600">WhatsApp</span>
              </button>
              
              <button 
                @click="shareToTwitter"
                class="flex flex-col items-center gap-2 p-3 rounded-xl hover:bg-gray-50 active:bg-gray-100 transition-colors"
              >
                <div class="w-12 h-12 bg-black rounded-full flex items-center justify-center">
                  <span class="text-2xl text-white">𝕏</span>
                </div>
                <span class="text-xs text-gray-600">Twitter</span>
              </button>
              
              <button 
                @click="shareToFacebook"
                class="flex flex-col items-center gap-2 p-3 rounded-xl hover:bg-gray-50 active:bg-gray-100 transition-colors"
              >
                <div class="w-12 h-12 bg-blue-600 rounded-full flex items-center justify-center">
                  <span class="text-2xl">📘</span>
                </div>
                <span class="text-xs text-gray-600">Facebook</span>
              </button>
              
              <button 
                @click="shareToInstagram"
                class="flex flex-col items-center gap-2 p-3 rounded-xl hover:bg-gray-50 active:bg-gray-100 transition-colors"
              >
                <div class="w-12 h-12 bg-gradient-to-br from-purple-600 via-pink-500 to-orange-400 rounded-full flex items-center justify-center">
                  <span class="text-2xl">📷</span>
                </div>
                <span class="text-xs text-gray-600">Instagram</span>
              </button>
              
              <button 
                @click="shareToThreads"
                class="flex flex-col items-center gap-2 p-3 rounded-xl hover:bg-gray-50 active:bg-gray-100 transition-colors"
              >
                <div class="w-12 h-12 bg-black rounded-full flex items-center justify-center">
                  <span class="text-2xl">🧵</span>
                </div>
                <span class="text-xs text-gray-600">Threads</span>
              </button>
              
              <button 
                @click="copyLink"
                class="flex flex-col items-center gap-2 p-3 rounded-xl hover:bg-gray-50 active:bg-gray-100 transition-colors"
              >
                <div class="w-12 h-12 bg-gray-500 rounded-full flex items-center justify-center">
                  <span class="text-2xl">🔗</span>
                </div>
                <span class="text-xs text-gray-600">Salin Link</span>
              </button>
            </div>
          </div>
          
          <!-- Desktop: Dropdown positioned relative to button -->
          <div 
            class="hidden sm:block fixed w-48 bg-white rounded-lg shadow-lg border border-gray-200 py-2 animate-fade-in"
            :style="{ top: dropdownPosition.top + 'px', right: dropdownPosition.right + 'px' }"
          >
            <button 
              @click="shareToTwitter"
              class="w-full px-4 py-2.5 text-left text-sm hover:bg-gray-50 active:bg-gray-100 flex items-center gap-3 transition-colors"
            >
              <span class="text-lg">𝕏</span>
              <span>Twitter / X</span>
            </button>
            <button 
              @click="shareToFacebook"
              class="w-full px-4 py-2.5 text-left text-sm hover:bg-gray-50 active:bg-gray-100 flex items-center gap-3 transition-colors"
            >
              <span class="text-lg">📘</span>
              <span>Facebook</span>
            </button>
            <button 
              @click="shareToWhatsapp"
              class="w-full px-4 py-2.5 text-left text-sm hover:bg-gray-50 active:bg-gray-100 flex items-center gap-3 transition-colors"
            >
              <span class="text-lg">💬</span>
              <span>WhatsApp</span>
            </button>
            <button 
              @click="shareToInstagram"
              class="w-full px-4 py-2.5 text-left text-sm hover:bg-gray-50 active:bg-gray-100 flex items-center gap-3 transition-colors"
            >
              <span class="text-lg">📷</span>
              <span>Instagram</span>
            </button>
            <button 
              @click="shareToThreads"
              class="w-full px-4 py-2.5 text-left text-sm hover:bg-gray-50 active:bg-gray-100 flex items-center gap-3 transition-colors"
            >
              <span class="text-lg">🧵</span>
              <span>Threads</span>
            </button>
            <hr class="my-2 border-gray-200" />
            <button 
              @click="copyLink"
              class="w-full px-4 py-2.5 text-left text-sm hover:bg-gray-50 active:bg-gray-100 flex items-center gap-3 transition-colors"
            >
              <span class="text-lg">🔗</span>
              <span>Salin Link</span>
            </button>
          </div>
        </div>
      </Transition>
    </Teleport>
    
    <!-- Header -->
    <header class="bg-white border-b border-gray-200 sticky top-0 z-20">
      <div class="max-w-3xl mx-auto px-4 py-3 flex items-center justify-between safe-area-top">
        <button 
          @click="goBack"
          class="flex items-center gap-2 text-gray-600 hover:text-gray-900 active:text-gray-900 transition-colors min-h-[44px] -ml-2 pl-2 pr-3"
        >
          <ArrowLeft class="w-5 h-5" />
          <span class="text-sm font-medium">Kembali</span>
        </button>
        
        <div ref="shareButtonRef">
          <Button 
            variant="outline" 
            class="gap-2 min-h-[44px]"
            @click="openShareMenu"
          >
            <Share2 class="w-4 h-4" />
            <span class="hidden sm:inline">Bagikan</span>
          </Button>
        </div>
      </div>
    </header>
    
    <!-- Loading -->
    <div v-if="loading" class="max-w-3xl mx-auto px-4 py-12 text-center">
      <div class="animate-pulse">
        <div class="aspect-[4/3] bg-gray-200 rounded-lg mb-4"></div>
        <div class="h-8 bg-gray-200 rounded w-3/4 mx-auto mb-2"></div>
        <div class="h-4 bg-gray-200 rounded w-1/2 mx-auto"></div>
      </div>
    </div>
    
    <!-- Error -->
    <div v-else-if="error" class="max-w-3xl mx-auto px-4 py-12 text-center">
      <p class="text-gray-500 mb-4">{{ error }}</p>
      <Button @click="goBack" class="min-h-[44px]">Kembali ke Feeds</Button>
    </div>
    
    <!-- Content -->
    <main v-else-if="feed" class="max-w-3xl mx-auto pb-safe">
      <!-- Photo Gallery with swipe support -->
      <div 
        v-if="feed.photos && feed.photos.length > 0" 
        ref="photoGalleryRef"
        class="relative bg-gray-200 aspect-[4/3] sm:aspect-auto sm:min-h-[300px] sm:max-h-[70vh] overflow-hidden touch-pan-y"
        @touchstart.passive="onTouchStart"
        @touchmove="onTouchMove"
        @touchend.passive="onTouchEnd"
      >
        <!-- Loading Placeholder -->
        <div 
          v-if="photoLoading && !photoError" 
          class="absolute inset-0 flex items-center justify-center bg-gray-200"
        >
          <div class="w-12 h-12 border-4 border-gray-300 border-t-gray-500 rounded-full animate-spin"></div>
        </div>
        
        <!-- Error Placeholder -->
        <div 
          v-if="photoError" 
          class="absolute inset-0 flex flex-col items-center justify-center bg-gray-200 text-gray-400"
        >
          <ImageOff class="w-16 h-16 mb-2" />
          <span class="text-sm">Gambar tidak tersedia</span>
        </div>
        
        <!-- Actual Image with swipe offset -->
        <img 
          :src="getPhotoUrl(feed.photos[currentPhotoIndex])"
          :alt="feed.photos[currentPhotoIndex].filename"
          class="w-full h-full object-contain transition-transform duration-150"
          :class="{ 'opacity-0': photoLoading || photoError }"
          :style="{ transform: `translateX(${swipeOffset}px)` }"
          loading="lazy"
          decoding="async"
          @load="onPhotoLoad"
          @error="onPhotoError"
        />
        
        <!-- Photo Navigation Buttons (hidden on mobile, visible on desktop) -->
        <div v-if="feed.photos.length > 1 && !photoError" class="hidden sm:flex absolute inset-x-0 bottom-0 top-0 items-center justify-between px-4">
          <button 
            v-if="currentPhotoIndex > 0"
            @click="prevPhoto"
            class="w-10 h-10 bg-black/50 hover:bg-black/70 active:bg-black/80 rounded-full flex items-center justify-center text-white transition-colors"
          >
            <ChevronLeft class="w-6 h-6" />
          </button>
          <div v-else class="w-10"></div>
          
          <button 
            v-if="currentPhotoIndex < feed.photos.length - 1"
            @click="nextPhoto"
            class="w-10 h-10 bg-black/50 hover:bg-black/70 active:bg-black/80 rounded-full flex items-center justify-center text-white transition-colors"
          >
            <ChevronRight class="w-6 h-6" />
          </button>
          <div v-else class="w-10"></div>
        </div>
        
        <!-- Photo Dot Indicators -->
        <div v-if="feed.photos.length > 1 && !photoError" class="absolute bottom-4 left-1/2 -translate-x-1/2 flex items-center gap-2">
          <button
            v-for="(_, index) in feed.photos"
            :key="index"
            @click="goToPhoto(index)"
            class="w-2 h-2 rounded-full transition-all duration-200"
            :class="index === currentPhotoIndex ? 'bg-white w-4' : 'bg-white/50 hover:bg-white/70'"
            :aria-label="`Foto ${index + 1}`"
          />
        </div>
        
        <!-- Swipe hint for mobile (shown briefly) -->
        <div v-if="feed.photos.length > 1 && !photoError" class="sm:hidden absolute bottom-12 left-1/2 -translate-x-1/2 text-white/70 text-xs pointer-events-none">
          Geser untuk foto lainnya
        </div>
      </div>
      
      <!-- Card Content -->
      <div class="bg-white mx-4 -mt-4 relative z-10 rounded-t-2xl shadow-lg" :class="{ 'mt-0 rounded-2xl': !feed.photos?.length }">
        <!-- Type Badge -->
        <div class="px-5 sm:px-6 pt-5 sm:pt-6 pb-2">
          <div class="flex items-center gap-2">
            <component :is="typeLabel.icon" class="w-5 h-5" :class="typeLabel.color" />
            <span class="text-sm font-semibold" :class="typeLabel.color">{{ typeLabel.text }}</span>
          </div>
        </div>
        
        <!-- Location Name -->
        <div class="px-5 sm:px-6 pb-2">
          <h1 v-if="locationName" class="text-xl sm:text-2xl font-bold text-gray-900">
            {{ locationName }}
          </h1>
          <p v-if="regionText" class="text-sm text-gray-500 mt-1">
            {{ regionText }}
          </p>
          <p v-else-if="!locationName && regionText" class="text-xl sm:text-2xl font-bold text-gray-900">
            {{ regionText }}
          </p>
        </div>
        
        <!-- Category & Tags -->
        <div class="px-5 sm:px-6 py-3 flex flex-wrap gap-2">
          <Badge :variant="categoryColors[feed.category] || 'default'" class="text-sm">
            {{ getCategoryLabel(feed.category) }}
          </Badge>
          <Badge 
            v-for="tag in tags" 
            :key="tag" 
            :variant="tagColor" 
            class="text-sm"
          >
            {{ getTagLabel(tag) }}
          </Badge>
        </div>
        
        <!-- Timestamp & Submitter -->
        <div class="px-5 sm:px-6 py-3 border-t border-gray-100 flex flex-wrap items-center gap-3 sm:gap-4 text-sm text-gray-500">
          <div class="flex items-center gap-2">
            <Clock class="w-4 h-4 flex-shrink-0" />
            <span class="text-xs sm:text-sm">{{ formatTimestamp(feed.submitted_at) }}</span>
          </div>
          <div v-if="feed.username" class="flex items-center gap-2">
            <User class="w-4 h-4 flex-shrink-0" />
            <span class="text-xs sm:text-sm">{{ feed.username }}{{ feed.organization ? ` • ${feed.organization}` : '' }}</span>
          </div>
        </div>
        
        <!-- Content -->
        <div class="px-5 sm:px-6 py-5 sm:py-6 border-t border-gray-100">
          <p class="text-gray-800 text-base sm:text-lg leading-relaxed whitespace-pre-wrap">{{ feed.content }}</p>
        </div>
        
        <!-- Actions -->
        <div class="px-5 sm:px-6 py-4 pb-8 border-t border-gray-100 flex gap-3 pb-safe-extra">
          <Button 
            v-if="feed.coordinates || feed.location_id || feed.faskes_id"
            variant="primary" 
            class="flex-1 gap-2 min-h-[48px] text-base"
            @click="goToMap"
          >
            <Navigation class="w-5 h-5" />
            Lihat di Peta
          </Button>
          
          <Button 
            v-if="feed.coordinates"
            variant="outline"
            class="gap-2 min-h-[48px]"
            @click="openGoogleMaps"
          >
            <ExternalLink class="w-5 h-5" />
            <span class="hidden sm:inline">Google Maps</span>
          </Button>
        </div>
      </div>
    </main>
    
  </div>
</template>

<style scoped>
/* Safe area utilities */
.safe-area-top {
  padding-top: env(safe-area-inset-top, 0);
}

.pb-safe {
  padding-bottom: env(safe-area-inset-bottom, 0);
}

.pb-safe-extra {
  padding-bottom: max(2rem, env(safe-area-inset-bottom, 2rem));
}

/* Animations */
.animate-slide-up {
  animation: slideUp 0.3s ease-out;
}

.animate-fade-in {
  animation: fadeIn 0.15s ease-out;
}

@keyframes slideUp {
  from {
    transform: translateY(100%);
  }
  to {
    transform: translateY(0);
  }
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: scale(0.95);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}

/* Share menu transitions */
.share-menu-enter-active,
.share-menu-leave-active {
  transition: opacity 0.2s ease;
}

.share-menu-enter-from,
.share-menu-leave-to {
  opacity: 0;
}

/* Touch pan for scrolling while allowing horizontal swipe */
.touch-pan-y {
  touch-action: pan-y;
}
</style>
