<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ShieldCheck, Loader2, MessageCircle, CheckCircle, RefreshCw, Copy, Check } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import ThemeToggle from '@/components/ThemeToggle.vue'

const route = useRoute()
const router = useRouter()

const userId = route.query.user_id as string
const pin = route.query.pin as string

const verified = ref(false)
const checking = ref(false)
const regenerating = ref(false)
const copied = ref(false)
const currentPIN = ref(pin)

let pollInterval: ReturnType<typeof setInterval> | null = null

const whatsappNumber = import.meta.env.VITE_WHATSAPP_NUMBER || '6281234567890'

const whatsappLink = computed(() => {
  const message = encodeURIComponent(currentPIN.value)
  return `https://wa.me/${whatsappNumber}?text=${message}`
})

async function checkVerificationStatus() {
  if (checking.value || verified.value) return
  
  checking.value = true
  try {
    const response = await fetch(`${import.meta.env.VITE_API_BASE_URL}/invitations/verification-status/${userId}`)
    const data = await response.json()
    
    if (data.success && data.data.verified) {
      verified.value = true
      if (pollInterval) {
        clearInterval(pollInterval)
        pollInterval = null
      }
    }
  } catch {
    // Silent fail, will retry on next poll
  } finally {
    checking.value = false
  }
}

async function regeneratePIN() {
  regenerating.value = true
  try {
    const response = await fetch(`${import.meta.env.VITE_API_BASE_URL}/invitations/regenerate-pin/${userId}`, {
      method: 'POST',
    })
    const data = await response.json()
    
    if (data.success) {
      currentPIN.value = data.data.pin
    }
  } catch {
    // Handle error
  } finally {
    regenerating.value = false
  }
}

function copyPIN() {
  navigator.clipboard.writeText(currentPIN.value)
  copied.value = true
  setTimeout(() => {
    copied.value = false
  }, 2000)
}

function goToLogin() {
  router.push('/login')
}

onMounted(() => {
  checkVerificationStatus()
  pollInterval = setInterval(checkVerificationStatus, 3000)
})

onUnmounted(() => {
  if (pollInterval) {
    clearInterval(pollInterval)
  }
})
</script>

<template>
  <div class="min-h-screen flex flex-col bg-background">
    <header class="flex items-center justify-end p-4">
      <ThemeToggle />
    </header>

    <main class="flex-1 flex items-center justify-center p-4">
      <Card class="w-full max-w-md">
        <CardHeader class="text-center space-y-4">
          <div 
            class="mx-auto flex h-16 w-16 items-center justify-center rounded-2xl transition-colors"
            :class="verified ? 'bg-green-500 text-white' : 'bg-primary text-primary-foreground'"
          >
            <CheckCircle v-if="verified" class="h-8 w-8" />
            <ShieldCheck v-else class="h-8 w-8" />
          </div>
          <div>
            <CardTitle class="text-2xl font-bold">
              {{ verified ? 'Verifikasi Berhasil!' : 'Verifikasi WhatsApp' }}
            </CardTitle>
            <CardDescription class="mt-2">
              {{ verified 
                ? 'Akun Anda telah aktif. Silakan login untuk melanjutkan.' 
                : 'Kirim PIN berikut ke WhatsApp Dayawarga untuk mengaktifkan akun' 
              }}
            </CardDescription>
          </div>
        </CardHeader>

        <CardContent class="space-y-6">
          <template v-if="verified">
            <div class="text-center space-y-4">
              <div class="rounded-lg bg-green-50 dark:bg-green-950 border border-green-200 dark:border-green-800 p-4">
                <p class="text-sm text-green-700 dark:text-green-300">
                  Terima kasih sudah bergabung bersama Dayawarga!
                </p>
              </div>
              
              <Button @click="goToLogin" class="w-full" size="lg">
                Masuk ke Dashboard
              </Button>
            </div>
          </template>

          <template v-else>
            <div class="space-y-4">
              <div class="text-center">
                <p class="text-sm text-muted-foreground mb-2">PIN Verifikasi Anda:</p>
                <div class="flex items-center justify-center gap-2">
                  <div class="font-mono text-4xl font-bold tracking-[0.3em] text-primary bg-muted px-6 py-4 rounded-lg">
                    {{ currentPIN }}
                  </div>
                  <Button variant="ghost" size="icon" @click="copyPIN" class="shrink-0">
                    <Check v-if="copied" class="h-4 w-4 text-green-500" />
                    <Copy v-else class="h-4 w-4" />
                  </Button>
                </div>
                <p class="text-xs text-muted-foreground mt-2">
                  PIN berlaku selama 15 menit
                </p>
              </div>

              <div class="rounded-lg bg-muted p-4 space-y-3">
                <p class="text-sm font-medium">Langkah verifikasi:</p>
                <ol class="text-sm text-muted-foreground space-y-2 list-decimal list-inside">
                  <li>Klik tombol "Kirim via WhatsApp" di bawah</li>
                  <li>Kirim pesan berisi PIN ke nomor Dayawarga</li>
                  <li>Tunggu konfirmasi dari chatbot</li>
                  <li>Halaman ini akan otomatis terupdate</li>
                </ol>
              </div>

              <a :href="whatsappLink" target="_blank" rel="noopener noreferrer" class="block">
                <Button class="w-full bg-green-600 hover:bg-green-700" size="lg">
                  <MessageCircle class="mr-2 h-5 w-5" />
                  Kirim via WhatsApp
                </Button>
              </a>

              <div class="flex items-center justify-between text-sm">
                <button 
                  @click="regeneratePIN" 
                  :disabled="regenerating"
                  class="text-primary hover:underline disabled:opacity-50 flex items-center gap-1"
                >
                  <RefreshCw v-if="regenerating" class="h-3 w-3 animate-spin" />
                  <RefreshCw v-else class="h-3 w-3" />
                  PIN kadaluarsa? Buat baru
                </button>
                
                <span v-if="checking" class="text-muted-foreground flex items-center gap-1">
                  <Loader2 class="h-3 w-3 animate-spin" />
                  Memeriksa...
                </span>
              </div>
            </div>
          </template>
        </CardContent>
      </Card>
    </main>

    <footer class="p-4 text-center text-sm text-muted-foreground">
      <p>Dayawarga &copy; {{ new Date().getFullYear() }}</p>
    </footer>
  </div>
</template>
