<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ShieldCheck, Loader2, Mail, Lock, Eye, EyeOff, AlertCircle } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import ThemeToggle from '@/components/ThemeToggle.vue'

const route = useRoute()
const router = useRouter()

const loading = ref(true)
const submitting = ref(false)
const error = ref('')
const showPassword = ref(false)
const showConfirmPassword = ref(false)

const invitationData = ref<{
  email: string
  name: string | null
  role: string
} | null>(null)

const password = ref('')
const confirmPassword = ref('')
const passwordError = ref('')

const token = route.query.token as string

onMounted(async () => {
  if (!token) {
    error.value = 'Token undangan tidak ditemukan'
    loading.value = false
    return
  }

  try {
    const response = await fetch(`${import.meta.env.VITE_API_BASE_URL}/invitations/validate?token=${token}`)
    const data = await response.json()

    if (!data.success) {
      error.value = data.error || 'Token undangan tidak valid atau sudah kadaluarsa'
      loading.value = false
      return
    }

    invitationData.value = data.data
  } catch {
    error.value = 'Gagal memvalidasi undangan. Silakan coba lagi.'
  } finally {
    loading.value = false
  }
})

function validatePassword() {
  passwordError.value = ''
  
  if (password.value.length < 8) {
    passwordError.value = 'Password minimal 8 karakter'
    return false
  }
  
  if (password.value !== confirmPassword.value) {
    passwordError.value = 'Password tidak cocok'
    return false
  }
  
  return true
}

async function handleSubmit() {
  if (!validatePassword()) return

  submitting.value = true
  error.value = ''

  try {
    const response = await fetch(`${import.meta.env.VITE_API_BASE_URL}/invitations/set-password`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ token, password: password.value }),
    })

    const data = await response.json()

    if (!data.success) {
      error.value = data.error || 'Gagal menyimpan password'
      submitting.value = false
      return
    }

    router.push({
      name: 'invite-verify',
      query: {
        user_id: data.data.user_id,
        pin: data.data.pin,
        email: data.data.email,
      },
    })
  } catch {
    error.value = 'Terjadi kesalahan. Silakan coba lagi.'
    submitting.value = false
  }
}
</script>

<template>
  <div class="min-h-screen flex flex-col bg-background">
    <header class="flex items-center justify-end p-4">
      <ThemeToggle />
    </header>

    <main class="flex-1 flex items-center justify-center p-4">
      <Card class="w-full max-w-md">
        <CardHeader class="text-center space-y-4">
          <div class="mx-auto flex h-16 w-16 items-center justify-center rounded-2xl bg-primary text-primary-foreground">
            <ShieldCheck class="h-8 w-8" />
          </div>
          <div>
            <CardTitle class="text-2xl font-bold">Bergabung dengan Dayawarga</CardTitle>
            <CardDescription class="mt-2">
              Anda diundang untuk bergabung sebagai anggota tim
            </CardDescription>
          </div>
        </CardHeader>

        <CardContent class="space-y-6">
          <div v-if="loading" class="flex justify-center py-8">
            <Loader2 class="h-8 w-8 animate-spin text-primary" />
          </div>

          <div v-else-if="error && !invitationData" class="text-center space-y-4">
            <div class="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-destructive/10">
              <AlertCircle class="h-6 w-6 text-destructive" />
            </div>
            <p class="text-sm text-destructive">{{ error }}</p>
            <Button variant="outline" @click="router.push('/login')">
              Kembali ke Login
            </Button>
          </div>

          <template v-else-if="invitationData">
            <div class="rounded-lg bg-muted p-4 space-y-2">
              <div class="flex items-center gap-2 text-sm">
                <Mail class="h-4 w-4 text-muted-foreground" />
                <span class="text-muted-foreground">Email:</span>
                <span class="font-medium">{{ invitationData.email }}</span>
              </div>
              <p class="text-sm text-muted-foreground">
                Silakan buat password untuk mengaktifkan akun Anda.
              </p>
            </div>

            <form @submit.prevent="handleSubmit" class="space-y-4">
              <div class="space-y-2">
                <Label for="password">Password</Label>
                <div class="relative">
                  <Lock class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                  <Input
                    id="password"
                    v-model="password"
                    :type="showPassword ? 'text' : 'password'"
                    placeholder="Minimal 8 karakter"
                    class="pl-10 pr-10"
                    required
                  />
                  <button
                    type="button"
                    @click="showPassword = !showPassword"
                    class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                  >
                    <Eye v-if="!showPassword" class="h-4 w-4" />
                    <EyeOff v-else class="h-4 w-4" />
                  </button>
                </div>
              </div>

              <div class="space-y-2">
                <Label for="confirmPassword">Konfirmasi Password</Label>
                <div class="relative">
                  <Lock class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                  <Input
                    id="confirmPassword"
                    v-model="confirmPassword"
                    :type="showConfirmPassword ? 'text' : 'password'"
                    placeholder="Ulangi password"
                    class="pl-10 pr-10"
                    required
                  />
                  <button
                    type="button"
                    @click="showConfirmPassword = !showConfirmPassword"
                    class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                  >
                    <Eye v-if="!showConfirmPassword" class="h-4 w-4" />
                    <EyeOff v-else class="h-4 w-4" />
                  </button>
                </div>
              </div>

              <p v-if="passwordError" class="text-sm text-destructive">{{ passwordError }}</p>
              <p v-if="error" class="text-sm text-destructive">{{ error }}</p>

              <Button type="submit" class="w-full" size="lg" :disabled="submitting">
                <Loader2 v-if="submitting" class="mr-2 h-4 w-4 animate-spin" />
                {{ submitting ? 'Menyimpan...' : 'Lanjutkan' }}
              </Button>
            </form>
          </template>
        </CardContent>
      </Card>
    </main>

    <footer class="p-4 text-center text-sm text-muted-foreground">
      <p>Dayawarga &copy; {{ new Date().getFullYear() }}</p>
    </footer>
  </div>
</template>
