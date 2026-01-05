<script setup lang="ts">
import { computed } from 'vue'
import { Check } from 'lucide-vue-next'

interface Props {
  modelValue: boolean
  id?: string
}

const props = defineProps<Props>()
const emit = defineEmits<{
  'update:modelValue': [value: boolean]
}>()

const checked = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})
</script>

<template>
  <button
    type="button"
    role="checkbox"
    :aria-checked="checked"
    :class="[
      'h-5 w-5 rounded border-2 flex items-center justify-center transition-colors',
      checked ? 'bg-blue-500 border-blue-500' : 'border-gray-300 bg-white'
    ]"
    @click="checked = !checked"
  >
    <Check v-if="checked" class="h-3 w-3 text-white" />
  </button>
</template>
