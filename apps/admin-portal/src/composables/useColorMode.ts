import { useColorMode as useVueUseColorMode, type BasicColorSchema } from '@vueuse/core'
import { computed, watch } from 'vue'

export function useColorMode() {
  const mode = useVueUseColorMode({
    attribute: 'class',
    modes: {
      light: 'light',
      dark: 'dark',
    },
    initialValue: 'dark' as BasicColorSchema,
    storageKey: 'admin-portal-color-mode',
  })

  const isDark = computed(() => mode.value === 'dark')
  const isLight = computed(() => mode.value === 'light')

  function setMode(value: BasicColorSchema) {
    mode.value = value
  }

  function toggleMode() {
    mode.value = mode.value === 'dark' ? 'light' : 'dark'
  }

  // Ensure dark class is applied on mount
  watch(
    mode,
    (value) => {
      if (value === 'dark') {
        document.documentElement.classList.add('dark')
        document.documentElement.classList.remove('light')
      } else {
        document.documentElement.classList.add('light')
        document.documentElement.classList.remove('dark')
      }
    },
    { immediate: true }
  )

  return {
    mode,
    isDark,
    isLight,
    setMode,
    toggleMode,
  }
}
