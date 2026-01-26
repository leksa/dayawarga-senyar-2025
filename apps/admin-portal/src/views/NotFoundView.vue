<script setup lang="ts">
import { useRouter } from 'vue-router'
import { Home, ArrowLeft } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'

const router = useRouter()

function goHome() {
  router.push('/')
}

function goBack() {
  router.back()
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center p-4">
    <div class="text-center space-y-6">
      <div class="space-y-2">
        <h1 class="text-8xl font-bold text-primary font-mono">404</h1>
        <h2 class="text-2xl font-semibold">Halaman Tidak Ditemukan</h2>
        <p class="text-muted-foreground max-w-md mx-auto">
          Halaman yang Anda cari tidak ada atau telah dipindahkan.
        </p>
      </div>

      <div class="flex items-center justify-center gap-4">
        <Button variant="outline" @click="goBack">
          <ArrowLeft class="mr-2 h-4 w-4" />
          Kembali
        </Button>
        <Button @click="goHome">
          <Home class="mr-2 h-4 w-4" />
          Ke Dashboard
        </Button>
      </div>
    </div>
  </div>
</template>
