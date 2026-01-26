import axios, { type AxiosInstance, type AxiosError, type InternalAxiosRequestConfig } from 'axios'

// API base URL from environment
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1'

// Create axios instance
const api: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Lazy import auth store to avoid Pinia initialization issues
let authStoreInstance: ReturnType<typeof import('@/stores/auth').useAuthStore> | null = null

async function getAuthStore() {
  if (!authStoreInstance) {
    const { useAuthStore } = await import('@/stores/auth')
    authStoreInstance = useAuthStore()
  }
  return authStoreInstance
}

// Request interceptor - add auth token
api.interceptors.request.use(
  async (config: InternalAxiosRequestConfig) => {
    try {
      const authStore = await getAuthStore()
      const token = authStore.accessToken

      if (token && config.headers) {
        config.headers.Authorization = `Bearer ${token}`
      }
    } catch (e) {
      console.warn('Could not get auth token:', e)
    }

    return config
  },
  (error: AxiosError) => {
    return Promise.reject(error)
  }
)

// Track if we're already redirecting to prevent multiple redirects
let isRedirecting = false

// Response interceptor - handle errors
api.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const status = error.response?.status

    // Handle 401 Unauthorized - token expired or invalid
    // Only redirect if we actually had a token (to avoid redirect loop on initial load)
    if (status === 401 && !isRedirecting) {
      const authStore = await getAuthStore()
      // Only clear and redirect if we thought we were authenticated
      if (authStore.accessToken) {
        isRedirecting = true
        await authStore.clearSession()
        window.location.href = '/login'
      }
    }

    return Promise.reject(error)
  }
)

export default api
