import api from './api'
import type { User, ApiResponse, ROPCTokenRequest, ROPCTokenResponse } from './types'

// OIDC configuration
const OIDC_AUTHORITY = (import.meta.env.VITE_OIDC_AUTHORITY || 'https://auth.example.com').replace(
  /\/$/,
  '',
)
const OIDC_CLIENT_ID = import.meta.env.VITE_OIDC_CLIENT_ID || 'admin-portal'

export const authService = {
  /**
   * Get current authenticated user info
   */
  async me(): Promise<User> {
    const response = await api.get<ApiResponse<User>>('/auth/me')
    return response.data.data
  },

  /**
   * Exchange username/password for tokens using ROPC flow
   * POST to Authentik's token endpoint directly
   */
  async tokenExchange(credentials: ROPCTokenRequest): Promise<ROPCTokenResponse> {
    // Authentik token endpoint - try application-specific first
    const tokenUrl = `${OIDC_AUTHORITY}/token/`

    const params = new URLSearchParams({
      grant_type: 'password',
      client_id: OIDC_CLIENT_ID,
      username: credentials.username,
      password: credentials.password,
      scope: 'openid profile email offline_access',
    })

    const response = await fetch(tokenUrl, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
      body: params.toString(),
    })

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}))
      // Handle specific error cases
      if (errorData.error === 'invalid_grant') {
        throw new Error('Email/username atau password salah')
      }
      if (errorData.error_description?.includes('mfa') || errorData.error_description?.includes('second_factor')) {
        throw new Error('Akun Anda memerlukan MFA. Silakan gunakan tombol "Masuk dengan SSO".')
      }
      throw new Error(errorData.error_description || errorData.error || 'Login gagal')
    }

    return response.json()
  },
}
