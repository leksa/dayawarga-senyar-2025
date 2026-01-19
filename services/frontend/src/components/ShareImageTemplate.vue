<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  category: string
  photo?: string
  kabupaten?: string
  kecamatan?: string
  desa?: string
  submitter: string
  content: string
  tags?: string[]
  timestamp: string
}

const props = defineProps<Props>()

// Category config with bold colors
const categoryConfig = computed(() => {
  const configs: Record<string, { label: string; bg: string; bgDark: string; accent: string; overlay: string }> = {
    'kebutuhan': {
      label: 'BUTUH BANTUAN',
      bg: '#DC2626',
      bgDark: '#991B1B',
      accent: '#FECACA',
      overlay: 'rgba(153, 27, 27, 0.25)'
    },
    'informasi': {
      label: 'INFORMASI',
      bg: '#1E40AF',
      bgDark: '#1E3A8A',
      accent: '#BFDBFE',
      overlay: 'rgba(30, 58, 138, 0.25)'
    },
    'follow-up': {
      label: 'FOLLOW UP',
      bg: '#B45309',
      bgDark: '#92400E',
      accent: '#FDE68A',
      overlay: 'rgba(146, 64, 14, 0.25)'
    },
    'info_bantuan': {
      label: 'TERIMA BANTUAN',
      bg: '#047857',
      bgDark: '#065F46',
      accent: '#A7F3D0',
      overlay: 'rgba(6, 95, 70, 0.25)'
    },
  }
  return configs[props.category] || {
    label: props.category.toUpperCase(),
    bg: '#DC2626',
    bgDark: '#991B1B',
    accent: '#FECACA',
    overlay: 'rgba(153, 27, 27, 0.25)'
  }
})

// Tag labels - map field values to display labels
const tagLabels: Record<string, string> = {
  'sar': 'SAR',
  'ambulan': 'Ambulan',
  'medis': 'Medis',
  'transport_roda4': 'Transport Roda 4',
  'transport_roda2': 'Transport Roda 2',
  'air_bersih': 'Air Bersih',
  'sembako': 'Sembako',
  'psikososial': 'Psikososial',
  'sekolah_darurat': 'Sekolah Darurat',
  'dapur_umum': 'Dapur Umum',
  'keamanan': 'Keamanan',
  'listrik': 'Listrik',
  'internet': 'Internet',
  'sinyal_selular': 'Sinyal Selular',
  'sanitasi_mck': 'Sanitasi MCK',
  'lainnya': 'Lainnya',
}

// Get tag label - convert field value to human-readable label
const getTagLabel = (tag: string): string => {
  if (!tag) return ''
  const trimmedTag = tag.trim().toLowerCase()
  return tagLabels[trimmedTag] || tag.replace(/_/g, ' ')
}

// Normalize tags - ensure it's always an array of individual tags
const normalizedTags = computed(() => {
  if (!props.tags) return []

  // If it's a string, split by comma or space
  if (typeof props.tags === 'string') {
    return (props.tags as string)
      .split(/[,\s]+/)
      .map(t => t.trim())
      .filter(t => t.length > 0)
  }

  // If it's an array, flatten any space-separated values
  return props.tags.flatMap(tag =>
    tag.split(/[,\s]+/).map(t => t.trim()).filter(t => t.length > 0)
  )
})

// Location combined
const locationText = computed(() => {
  const parts = []
  if (props.desa) parts.push(props.desa)
  if (props.kecamatan) parts.push(props.kecamatan)
  if (props.kabupaten) parts.push(props.kabupaten)
  return parts.join(' · ')
})

// Truncate content - increased limit for smaller font
const truncatedContent = computed(() => {
  const max = 280
  if (props.content.length <= max) return props.content
  return props.content.substring(0, max).trim() + '...'
})
</script>

<template>
  <div
    class="share-card"
    :style="{
      '--bg-color': categoryConfig.bg,
      '--bg-dark': categoryConfig.bgDark,
      '--accent': categoryConfig.accent,
      '--overlay': categoryConfig.overlay
    }"
  >
    <!-- Photo Hero at top (if exists) with overlaid category/date -->
    <div v-if="photo" class="photo-section">
      <img :src="photo" alt="" class="photo-img" />
      <div class="photo-overlay"></div>
      <!-- Category and Date overlay on photo -->
      <div class="photo-top-bar">
        <span class="photo-category-label">{{ categoryConfig.label }}</span>
        <span class="photo-date-label">{{ timestamp }}</span>
      </div>
    </div>

    <!-- No photo: show category bar normally -->
    <div v-else class="top-bar">
      <span class="category-label">{{ categoryConfig.label }}</span>
      <span class="date-label">{{ timestamp }}</span>
    </div>

    <!-- Background Image Overlay (for content area) -->
    <div class="bg-image"></div>
    <!-- Background Pattern -->
    <div class="bg-pattern"></div>

    <!-- Content Card with slight opacity -->
    <div class="content-card">
      <!-- Location -->
      <div v-if="locationText" class="location-row">
        <svg class="loc-icon" viewBox="0 0 24 24" fill="currentColor">
          <path d="M12 2C8.13 2 5 5.13 5 9c0 5.25 7 13 7 13s7-7.75 7-13c0-3.87-3.13-7-7-7zm0 9.5c-1.38 0-2.5-1.12-2.5-2.5s1.12-2.5 2.5-2.5 2.5 1.12 2.5 2.5-1.12 2.5-2.5 2.5z"/>
        </svg>
        <span class="loc-text">{{ locationText }}</span>
      </div>

      <!-- Message -->
      <p class="message">{{ truncatedContent }}</p>

      <!-- Submitter -->
      <div class="submitter">
        <span class="sub-label">oleh</span>
        <span class="sub-name">{{ submitter }}</span>
      </div>

      <!-- Tags -->
      <div v-if="normalizedTags.length > 0" class="tags-row">
        <span v-for="(tag, index) in normalizedTags.slice(0, 3)" :key="index" class="tag">
          {{ getTagLabel(tag) }}
        </span>
      </div>
    </div>

    <!-- Footer Brand -->
    <div class="brand-bar">
      <div class="brand-left">
        <img src="/logo.png" alt="Dayawarga" class="brand-logo" />
        <span class="brand-tagline">Peta Bencana Senyar</span>
      </div>
      <div class="brand-text">
        <span class="brand-name">DAYAWARGA</span>
        <span class="brand-tld">.COM</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
@import url('https://fonts.cdnfonts.com/css/overused-grotesk');

.share-card {
  width: 480px;
  background: linear-gradient(145deg, var(--bg-color) 0%, var(--bg-dark) 100%);
  font-family: 'Overused Grotesk', -apple-system, sans-serif;
  position: relative;
  overflow: hidden;
}

/* Background Image */
.bg-image {
  position: absolute;
  inset: 0;
  background-image: url('/background-share-image.jpg');
  background-size: cover;
  background-position: center;
  opacity: 0.15;
  mix-blend-mode: overlay;
  pointer-events: none;
}

/* Background Pattern */
.bg-pattern {
  position: absolute;
  inset: 0;
  opacity: 0.06;
  background-image:
    radial-gradient(circle at 20% 80%, rgba(255,255,255,0.5) 0%, transparent 50%),
    radial-gradient(circle at 80% 20%, rgba(255,255,255,0.4) 0%, transparent 40%),
    repeating-linear-gradient(
      -45deg,
      transparent,
      transparent 30px,
      rgba(255,255,255,0.15) 30px,
      rgba(255,255,255,0.15) 60px
    );
  pointer-events: none;
}

/* Top Bar */
.top-bar {
  position: relative;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px 16px;
}

.category-label {
  font-size: 14px;
  font-weight: 800;
  letter-spacing: 3px;
  text-transform: uppercase;
  color: #FFFFFF;
}

.date-label {
  font-size: 12px;
  font-weight: 500;
  color: rgba(255,255,255,0.7);
}

/* Photo Section - Full width at top with 3:2 aspect ratio */
.photo-section {
  position: relative;
  width: 100%;
  aspect-ratio: 3 / 2;
  overflow: hidden;
}

.photo-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  object-position: center;
  filter: contrast(1.05) saturate(1.1);
}

.photo-overlay {
  position: absolute;
  inset: 0;
  background: var(--overlay);
}

/* Category/Date overlay on photo */
.photo-top-bar {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  background: rgba(0, 0, 0, 0.7);
  z-index: 10;
}

.photo-category-label {
  font-size: 14px;
  font-weight: 800;
  letter-spacing: 3px;
  text-transform: uppercase;
  color: #FFFFFF;
}

.photo-date-label {
  font-size: 12px;
  font-weight: 700;
  color: #FFFFFF;
}

/* Content Card - very transparent with contrasting text */
.content-card {
  position: relative;
  margin: 16px 24px;
  padding: 20px;
  background: rgba(255, 255, 255, 0.2);
  backdrop-filter: blur(4px);
  border-radius: 4px;
  border: 1px solid rgba(255, 255, 255, 0.3);
}

.location-row {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 12px;
  padding-bottom: 12px;
  border-bottom: 2px solid rgba(255, 255, 255, 0.4);
}

.loc-icon {
  width: 14px;
  height: 14px;
  color: #FFFFFF;
  flex-shrink: 0;
}

.loc-text {
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 1.5px;
  text-transform: uppercase;
  color: rgba(255, 255, 255, 0.9);
}

.message {
  font-size: 14px;
  font-weight: 500;
  line-height: 1.55;
  color: #FFFFFF;
  margin: 0 0 14px 0;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.3);
}

.submitter {
  display: flex;
  align-items: baseline;
  gap: 6px;
  margin-bottom: 14px;
}

.sub-label {
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 1px;
  color: rgba(255, 255, 255, 0.6);
}

.sub-name {
  font-size: 13px;
  font-weight: 700;
  color: #FFFFFF;
}

.tags-row {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.tag {
  font-size: 9px;
  font-weight: 700;
  letter-spacing: 0.5px;
  text-transform: uppercase;
  padding: 5px 10px;
  background: rgba(255, 255, 255, 0.25);
  color: #FFFFFF;
  border: 1px solid rgba(255, 255, 255, 0.4);
}

/* Brand Bar */
.brand-bar {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 24px 20px;
}

.brand-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.brand-logo {
  width: 36px;
  height: 36px;
  object-fit: contain;
}

.brand-tagline {
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.5px;
  color: rgba(255, 255, 255, 0.8);
  text-transform: uppercase;
}

.brand-text {
  display: flex;
  align-items: baseline;
}

.brand-name {
  font-size: 16px;
  font-weight: 800;
  letter-spacing: 3px;
  color: #FFFFFF;
}

.brand-tld {
  font-size: 16px;
  font-weight: 800;
  letter-spacing: 3px;
  color: rgba(255,255,255,0.6);
}
</style>
