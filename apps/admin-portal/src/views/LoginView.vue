<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { LogIn, ShieldCheck, Loader2 } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import ThemeToggle from '@/components/ThemeToggle.vue'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const isLoggingIn = ref(false)

const passwordRecoveryUrl = computed(() => {
  const baseUrl = import.meta.env.VITE_AUTHENTIK_BASE_URL || 'https://auth.dayawarga.com'
  return `${baseUrl}/if/flow/recovery/`
})

onMounted(async () => {
  await authStore.init()
  if (authStore.isAuthenticated) {
    router.push('/')
  }
})

async function handleLogin() {
  isLoggingIn.value = true
  await authStore.login()
}
</script>

<template>
  <div class="min-h-screen flex flex-col">
    <!-- Header -->
    <header class="flex items-center justify-end p-4">
      <ThemeToggle />
    </header>

    <!-- Main content -->
    <main class="flex-1 flex items-center justify-center p-4">
      <Card class="w-full max-w-md">
        <CardHeader class="text-center space-y-4">
          <!-- Logo -->
          <div class="mx-auto flex h-16 w-16 items-center justify-center rounded-2xl bg-primary text-primary-foreground">
            <ShieldCheck class="h-8 w-8" />
          </div>
          <div>
            <CardTitle class="text-2xl font-bold">Dayawarga Admin</CardTitle>
            <CardDescription class="mt-2">
              Portal administrasi untuk manajemen data bencana
            </CardDescription>
          </div>
        </CardHeader>
        <CardContent class="space-y-6">
          <!-- Error message -->
          <p v-if="authStore.error" class="text-sm text-destructive text-center">
            {{ authStore.error }}
          </p>

          <!-- SSO Login Button -->
          <Button
            @click="handleLogin"
            class="w-full"
            size="lg"
            :disabled="isLoggingIn"
          >
            <Loader2 v-if="isLoggingIn" class="mr-2 h-5 w-5 animate-spin" />
            <LogIn v-else class="mr-2 h-5 w-5" />
            {{ isLoggingIn ? 'Mengarahkan ke login...' : 'Masuk dengan Authentik' }}
          </Button>

          <div class="text-center space-y-2">
            <p class="text-xs text-muted-foreground">
              Anda akan diarahkan ke halaman login Authentik untuk autentikasi.
            </p>
            <a
              :href="passwordRecoveryUrl"
              target="_blank"
              rel="noopener noreferrer"
              class="text-xs text-primary hover:underline"
            >
              Lupa password?
            </a>
          </div>
        </CardContent>
      </Card>
    </main>

    <!-- Footer -->
    <footer class="p-4 text-center text-sm text-muted-foreground">
      <p>Dayawarga &copy; {{ new Date().getFullYear() }}</p>
    </footer>
  </div>
</template>
