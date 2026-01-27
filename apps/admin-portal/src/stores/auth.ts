import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { UserManager, WebStorageStateStore, type User as OIDCUser, type UserManagerSettings } from 'oidc-client-ts'
import { authService } from '@/services/auth'
import type { User as ApiUser } from '@/services/types'

const oidcSettings: UserManagerSettings = {
  authority: import.meta.env.VITE_OIDC_AUTHORITY || 'https://auth.example.com',
  client_id: import.meta.env.VITE_OIDC_CLIENT_ID || 'admin-portal',
  redirect_uri: `${window.location.origin}/callback`,
  post_logout_redirect_uri: window.location.origin,
  response_type: 'code',
  scope: 'openid profile email offline_access',
  automaticSilentRenew: false,
  userStore: new WebStorageStateStore({ store: window.localStorage }),
}

export const useAuthStore = defineStore('auth', () => {
  const userManager = new UserManager(oidcSettings)
  const user = ref<OIDCUser | null>(null)
  const apiUser = ref<ApiUser | null>(null)
  const isLoading = ref(true)
  const error = ref<string | null>(null)

  const isAuthenticated = computed(() => user.value !== null && !user.value.expired)
  const accessToken = computed(() => user.value?.access_token ?? null)
  const profile = computed(() => user.value?.profile ?? null)
  const displayName = computed(() => {
    if (apiUser.value?.name) return apiUser.value.name
    if (!profile.value) return null
    return profile.value.name || profile.value.preferred_username || profile.value.email || 'User'
  })

  // Role-based computed properties
  const userRole = computed(() => apiUser.value?.role ?? null)
  const isSuperAdmin = computed(() => apiUser.value?.role === 'super_admin')
  const isOrgAdmin = computed(() => apiUser.value?.role === 'org_admin')
  const canManageOrganizations = computed(() => isSuperAdmin.value) // Only super_admin can create/delete orgs
  const canInviteUsers = computed(() => isSuperAdmin.value) // Only super_admin can invite users

  async function init() {
    isLoading.value = true
    error.value = null
    try {
      // Try to get user from oidc-client-ts
      let storedUser = await userManager.getUser()

      // If not found, try to read from our manual storage
      if (!storedUser) {
        const authority = (import.meta.env.VITE_OIDC_AUTHORITY || '').replace(/\/$/, '')
        const clientId = import.meta.env.VITE_OIDC_CLIENT_ID
        const storageKey = `oidc.user:${authority}:${clientId}`
        const storedUserStr = localStorage.getItem(storageKey)
        if (storedUserStr) {
          try {
            storedUser = JSON.parse(storedUserStr)
          } catch {
            // Invalid stored user, ignore
          }
        }
      }

      if (storedUser) {
        // Check expiration - expires_at is in seconds
        const now = Math.floor(Date.now() / 1000)
        const isExpired = storedUser.expires_at ? storedUser.expires_at <= now : true

        if (!isExpired) {
          user.value = storedUser
          // Fetch user info from API to get role
          await fetchUserInfo()
        } else {
          // Token expired, clear and let user re-login
          await clearSession()
        }
      }
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to initialize auth'
      console.error('Auth init error:', e)
    } finally {
      isLoading.value = false
    }
  }

  // Fetch user info from API (includes role)
  async function fetchUserInfo() {
    try {
      const userInfo = await authService.me()
      apiUser.value = userInfo
    } catch (e) {
      console.error('Failed to fetch user info:', e)
      // Don't fail auth if user info fetch fails - user might still be able to use the app
    }
  }

  async function login() {
    error.value = null
    try {
      await userManager.signinRedirect()
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to initiate login'
      console.error('Login error:', e)
    }
  }

  async function handleCallback() {
    isLoading.value = true
    error.value = null
    try {
      const callbackUser = await userManager.signinRedirectCallback()
      user.value = callbackUser

      // If oidc-client-ts didn't store, manually store user
      const keys = Object.keys(localStorage).filter(k => k.startsWith('oidc.'))
      if (keys.length === 0 && callbackUser) {
        const authority = (import.meta.env.VITE_OIDC_AUTHORITY || '').replace(/\/$/, '')
        const clientId = import.meta.env.VITE_OIDC_CLIENT_ID
        const storageKey = `oidc.user:${authority}:${clientId}`
        localStorage.setItem(storageKey, JSON.stringify(callbackUser))
      }

      // Fetch user info from API to get role
      await fetchUserInfo()

      return true
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to handle callback'
      console.error('Callback error:', e)
      return false
    } finally {
      isLoading.value = false
    }
  }

  async function logout() {
    error.value = null
    try {
      await userManager.signoutRedirect()
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to logout'
      console.error('Logout error:', e)
    }
  }

  async function clearSession() {
    await userManager.removeUser()
    user.value = null
    apiUser.value = null
    // Also clear manually stored user
    const authority = (import.meta.env.VITE_OIDC_AUTHORITY || '').replace(/\/$/, '')
    const clientId = import.meta.env.VITE_OIDC_CLIENT_ID
    const storageKey = `oidc.user:${authority}:${clientId}`
    localStorage.removeItem(storageKey)
  }

  // Event listeners for token changes
  userManager.events.addUserLoaded(async (loadedUser) => {
    user.value = loadedUser
    await fetchUserInfo()
  })

  userManager.events.addUserUnloaded(() => {
    user.value = null
    apiUser.value = null
  })

  userManager.events.addAccessTokenExpired(async () => {
    console.log('Access token expired, clearing session...')
    await clearSession()
    window.location.href = '/login'
  })

  return {
    user,
    apiUser,
    isLoading,
    error,
    isAuthenticated,
    accessToken,
    profile,
    displayName,
    userRole,
    isSuperAdmin,
    isOrgAdmin,
    canManageOrganizations,
    canInviteUsers,
    init,
    login,
    logout,
    handleCallback,
    clearSession,
    fetchUserInfo,
  }
})
