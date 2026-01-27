import api from './api'
import type { User, ApiResponse } from './types'

export const authService = {
  /**
   * Get current authenticated user info
   */
  async me(): Promise<User> {
    const response = await api.get<ApiResponse<User>>('/auth/me')
    return response.data.data
  },
}
