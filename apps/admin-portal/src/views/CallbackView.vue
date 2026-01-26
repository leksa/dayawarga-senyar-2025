<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { Loader2, AlertCircle } from 'lucide-vue-next'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()
const hasError = ref(false)

onMounted(async () => {
  const success = await authStore.handleCallback()
  if (success) {
    router.push('/')
  } else {
    hasError.value = true
  }
})

function goToLogin() {
  router.push('/login')
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center p-4">
    <Card class="w-full max-w-md">
      <CardHeader class="text-center">
        <div
          v-if="!hasError"
          class="mx-auto flex h-12 w-12 items-center justify-center"
        >
          <Loader2 class="h-8 w-8 animate-spin text-primary" />
        </div>
        <div
          v-else
          class="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-destructive/10"
        >
          <AlertCircle class="h-6 w-6 text-destructive" />
        </div>

        <CardTitle v-if="!hasError">Memproses Login</CardTitle>
        <CardTitle v-else>Login Gagal</CardTitle>

        <CardDescription v-if="!hasError">
          Mohon tunggu sebentar...
        </CardDescription>
        <CardDescription v-else>
          {{ authStore.error || 'Terjadi kesalahan saat memproses login.' }}
        </CardDescription>
      </CardHeader>

      <CardContent v-if="hasError">
        <Button @click="goToLogin" class="w-full">
          Kembali ke Login
        </Button>
      </CardContent>
    </Card>
  </div>
</template>
