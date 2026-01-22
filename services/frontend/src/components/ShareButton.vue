<script setup lang="ts">
import { ref, createApp, h, nextTick } from 'vue'
import { Download, Loader2 } from 'lucide-vue-next'
import html2canvas from 'html2canvas'
import ShareImageTemplate from './ShareImageTemplate.vue'

interface Props {
  feedId: string
  // Feed data for template
  category?: string
  photo?: string
  kabupaten?: string
  kecamatan?: string
  desa?: string
  submitter?: string
  content?: string
  tags?: string[]
  timestamp?: string
}

const props = defineProps<Props>()

const emit = defineEmits<{
  (e: 'downloaded'): void
  (e: 'error', message: string): void
}>()

const isDownloading = ref(false)

// Download feed item as styled image using ShareImageTemplate
const handleDownload = async (e: Event) => {
  e.stopPropagation()

  if (isDownloading.value) return

  isDownloading.value = true

  try {
    // Create a temporary container for the template
    const container = document.createElement('div')
    container.style.position = 'absolute'
    container.style.left = '-9999px'
    container.style.top = '0'
    document.body.appendChild(container)

    // Create and mount the ShareImageTemplate component
    const app = createApp({
      render() {
        return h(ShareImageTemplate, {
          category: props.category || 'informasi',
          photo: props.photo,
          kabupaten: props.kabupaten,
          kecamatan: props.kecamatan,
          desa: props.desa,
          submitter: props.submitter || 'Anonim',
          content: props.content || '',
          tags: props.tags,
          timestamp: props.timestamp || new Date().toLocaleDateString('id-ID'),
        })
      },
    })

    app.mount(container)

    // Wait for images to load and component to render
    await nextTick()
    await new Promise((resolve) => setTimeout(resolve, 500))

    // Find the share-card element
    const element = container.querySelector('.share-card') as HTMLElement
    if (!element) {
      throw new Error('Template element not found')
    }

    const canvas = await html2canvas(element, {
      backgroundColor: null,
      scale: 2,
      logging: false,
      useCORS: true,
      allowTaint: false,
    })

    // Convert to blob and download
    canvas.toBlob((blob) => {
      if (blob) {
        const url = URL.createObjectURL(blob)
        const link = document.createElement('a')
        link.href = url
        link.download = `dayawarga-update-${props.feedId.substring(0, 8)}.jpg`
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)
        URL.revokeObjectURL(url)
        emit('downloaded')
      }
    }, 'image/jpeg', 0.95)

    // Cleanup
    app.unmount()
    document.body.removeChild(container)
  } catch (error) {
    emit('error', error instanceof Error ? error.message : 'Gagal membuat gambar')
  } finally {
    isDownloading.value = false
  }
}
</script>

<template>
  <button
    class="flex items-center gap-1 px-2 py-1 rounded hover:bg-gray-100 text-gray-500 hover:text-blue-600 transition-colors text-xs font-medium"
    :class="{ 'opacity-50 cursor-wait': isDownloading }"
    title="Download gambar untuk share"
    :disabled="isDownloading"
    @click="handleDownload"
  >
    <Loader2 v-if="isDownloading" class="w-3.5 h-3.5 animate-spin" />
    <Download v-else class="w-3.5 h-3.5" />
    <span class="hidden sm:inline">SHARE</span>
  </button>
</template>
