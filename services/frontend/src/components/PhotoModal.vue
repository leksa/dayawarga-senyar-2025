<script setup lang="ts">
import { X } from 'lucide-vue-next'

interface Props {
  imageUrl: string
  isOpen: boolean
}

defineProps<Props>()
const emit = defineEmits<{
  close: []
}>()

const handleBackdropClick = (e: MouseEvent) => {
  if (e.target === e.currentTarget) {
    emit('close')
  }
}
</script>

<template>
  <Teleport to="body">
    <Transition name="modal">
      <div
        v-if="isOpen"
        class="fixed inset-0 z-[2000] flex items-center justify-center bg-black/80 backdrop-blur-sm"
        @click="handleBackdropClick"
      >
        <!-- Close button -->
        <button
          class="absolute top-4 right-4 p-2 bg-white/10 hover:bg-white/20 rounded-full transition-colors z-10"
          @click="emit('close')"
        >
          <X class="w-6 h-6 text-white" />
        </button>

        <!-- Image -->
        <img
          :src="imageUrl"
          alt="Photo"
          class="max-w-[90vw] max-h-[90vh] object-contain rounded-lg shadow-2xl"
        />
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.2s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-active img,
.modal-leave-active img {
  transition: transform 0.2s ease;
}

.modal-enter-from img,
.modal-leave-to img {
  transform: scale(0.95);
}
</style>
