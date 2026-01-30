export type BadgeVariant = 'default' | 'success' | 'warning' | 'danger' | 'outline' | 'orange' | 'info'

export const categoryColors: Record<string, BadgeVariant> = {
  kebutuhan: 'danger',
  informasi: 'warning',
  'follow-up': 'orange',
  'info_bantuan': 'success',
}

export const categoryLabels: Record<string, string> = {
  kebutuhan: 'Butuh Bantuan',
  informasi: 'Informasi',
  'follow-up': 'Follow-up',
  'info_bantuan': 'Terima Bantuan',
}

export const tagLabels: Record<string, string> = {
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

export const tagColor: BadgeVariant = 'info'

export const getCategoryColor = (category: string): BadgeVariant => {
  return categoryColors[category] || 'default'
}

export const getCategoryLabel = (category: string): string => {
  return categoryLabels[category] || category
}

export const getTagLabel = (tag: string): string => {
  if (!tag) return ''
  const trimmedTag = tag.trim().toLowerCase()
  return tagLabels[trimmedTag] || tag.replace(/_/g, ' ')
}

export const categoryHexColors: Record<string, { label: string; bg: string; bgDark: string; accent: string; overlay: string }> = {
  'kebutuhan': {
    label: 'BUTUH BANTUAN',
    bg: '#DC2626',
    bgDark: '#991B1B',
    accent: '#FECACA',
    overlay: 'rgba(153, 27, 27, 0.25)'
  },
  'informasi': {
    label: 'INFORMASI',
    bg: '#D97706',
    bgDark: '#B45309',
    accent: '#FDE68A',
    overlay: 'rgba(180, 83, 9, 0.25)'
  },
  'follow-up': {
    label: 'FOLLOW UP',
    bg: '#EA580C',
    bgDark: '#C2410C',
    accent: '#FDBA74',
    overlay: 'rgba(194, 65, 12, 0.25)'
  },
  'info_bantuan': {
    label: 'TERIMA BANTUAN',
    bg: '#047857',
    bgDark: '#065F46',
    accent: '#A7F3D0',
    overlay: 'rgba(6, 95, 70, 0.25)'
  },
}

export const getCategoryPopupClass = (category: string): string => {
  switch (category) {
    case 'kebutuhan': return 'cat-kebutuhan'
    case 'follow-up': return 'cat-followup'
    case 'info_bantuan': return 'cat-info-bantuan'
    default: return 'cat-info'
  }
}
