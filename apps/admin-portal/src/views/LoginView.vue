<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { LogIn, ShieldCheck, Loader2, Eye, EyeOff } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import ThemeToggle from '@/components/ThemeToggle.vue'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

// Form state
const username = ref('')
const password = ref('')
const showPassword = ref(false)
const isSubmitting = ref(false)

onMounted(async () => {
  await authStore.init()
  if (authStore.isAuthenticated) {
    router.push('/')
  }
})

async function handlePasswordLogin() {
  if (!username.value || !password.value) return

  isSubmitting.value = true
  try {
    const success = await authStore.loginWithPassword(username.value, password.value)
    if (success) {
      router.push('/')
    }
  } finally {
    isSubmitting.value = false
  }
}

async function handleSSOLogin() {
  await authStore.login()
}

function togglePasswordVisibility() {
  showPassword.value = !showPassword.value
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
          <!-- Login Form -->
          <form @submit.prevent="handlePasswordLogin" class="space-y-4">
            <div class="space-y-2">
              <Label for="username">Email atau Username</Label>
              <Input
                id="username"
                v-model="username"
                type="text"
                placeholder="admin@organisasi.com"
                autocomplete="username"
                :disabled="isSubmitting"
                required
              />
            </div>

            <div class="space-y-2">
              <Label for="password">Password</Label>
              <div class="relative">
                <Input
                  id="password"
                  v-model="password"
                  :type="showPassword ? 'text' : 'password'"
                  placeholder="Masukkan password"
                  autocomplete="current-password"
                  :disabled="isSubmitting"
                  class="pr-10"
                  required
                />
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  class="absolute right-0 top-0 h-full px-3 hover:bg-transparent"
                  @click="togglePasswordVisibility"
                  :disabled="isSubmitting"
                >
                  <Eye v-if="!showPassword" class="h-4 w-4 text-muted-foreground" />
                  <EyeOff v-else class="h-4 w-4 text-muted-foreground" />
                </Button>
              </div>
            </div>

            <Button
              type="submit"
              class="w-full"
              size="lg"
              :disabled="isSubmitting || !username || !password"
            >
              <Loader2 v-if="isSubmitting" class="mr-2 h-5 w-5 animate-spin" />
              <LogIn v-else class="mr-2 h-5 w-5" />
              {{ isSubmitting ? 'Memproses...' : 'Masuk' }}
            </Button>
          </form>

          <!-- Error message -->
          <p v-if="authStore.error" class="text-sm text-destructive text-center">
            {{ authStore.error }}
          </p>

          <!-- Divider -->
          <div class="relative">
            <div class="absolute inset-0 flex items-center">
              <Separator class="w-full" />
            </div>
            <div class="relative flex justify-center text-xs uppercase">
              <span class="bg-card px-2 text-muted-foreground">atau</span>
            </div>
          </div>

          <!-- SSO Fallback Button -->
          <Button
            variant="outline"
            @click="handleSSOLogin"
            class="w-full"
            size="lg"
            :disabled="isSubmitting"
          >
            <LogIn class="mr-2 h-5 w-5" />
            Masuk dengan SSO
          </Button>

          <p class="text-xs text-muted-foreground text-center">
            Gunakan SSO jika Anda memiliki autentikasi multi-faktor (MFA) aktif.
          </p>
        </CardContent>
      </Card>
    </main>

    <!-- Footer -->
    <footer class="p-4 text-center text-sm text-muted-foreground">
      <p>Dayawarga &copy; {{ new Date().getFullYear() }}</p>
    </footer>
  </div>
</template>
