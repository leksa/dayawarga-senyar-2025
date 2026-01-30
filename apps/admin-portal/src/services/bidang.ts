import api from './api'
import type { Bidang, ApiResponse } from './types'

export const bidangService = {
  async list(): Promise<Bidang[]> {
    const response = await api.get<ApiResponse<Bidang[]>>('/bidang')
    return response.data.data
  },

  async addToOrganization(orgId: string, bidangId: string): Promise<void> {
    await api.post(`/organizations/${orgId}/bidang`, { bidang_id: bidangId })
  },

  async removeFromOrganization(orgId: string, bidangId: string): Promise<void> {
    await api.delete(`/organizations/${orgId}/bidang/${bidangId}`)
  },
}
