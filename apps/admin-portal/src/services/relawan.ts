import api from './api'
import type {
  Relawan,
  RelawanStats,
  CreateRelawanInput,
  UpdateRelawanInput,
  RelawanFilter,
  RelawanStatus,
  ApiResponse,
  WAStatus,
} from './types'

interface RelawanListResponse {
  relawan: Relawan[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export const relawanService = {
  /**
   * List relawan with optional filters
   */
  async list(filter?: RelawanFilter): Promise<RelawanListResponse> {
    const params = new URLSearchParams()
    if (filter?.organization_id) params.append('organization_id', filter.organization_id)
    if (filter?.group_id) params.append('group_id', filter.group_id)
    if (filter?.status) params.append('status', filter.status)
    if (filter?.search) params.append('search', filter.search)
    if (filter?.has_odk_access !== undefined) params.append('has_odk_access', String(filter.has_odk_access))
    if (filter?.page) params.append('page', String(filter.page))
    if (filter?.page_size) params.append('page_size', String(filter.page_size))

    const response = await api.get<ApiResponse<RelawanListResponse>>('/relawan', { params })
    return response.data.data
  },

  /**
   * Get relawan by ID
   */
  async get(id: string): Promise<Relawan> {
    const response = await api.get<ApiResponse<Relawan>>(`/relawan/${id}`)
    return response.data.data
  },

  /**
   * Create new relawan
   */
  async create(input: CreateRelawanInput): Promise<Relawan> {
    const response = await api.post<ApiResponse<Relawan>>('/relawan', input)
    return response.data.data
  },

  /**
   * Update relawan
   */
  async update(id: string, input: UpdateRelawanInput): Promise<Relawan> {
    const response = await api.put<ApiResponse<Relawan>>(`/relawan/${id}`, input)
    return response.data.data
  },

  /**
   * Delete relawan (soft delete)
   */
  async delete(id: string): Promise<void> {
    await api.delete(`/relawan/${id}`)
  },

  /**
   * Update relawan status
   */
  async updateStatus(id: string, status: RelawanStatus): Promise<void> {
    await api.put(`/relawan/${id}/status`, { status })
  },

  /**
   * Move relawan to a different group
   */
  async moveToGroup(id: string, groupId: string | null): Promise<void> {
    await api.put(`/relawan/${id}/group`, { group_id: groupId })
  },

  /**
   * Bulk move relawan to a group
   */
  async bulkMoveToGroup(ids: string[], groupId: string | null): Promise<void> {
    await api.post('/relawan/bulk/move-to-group', {
      ids,
      group_id: groupId,
    })
  },

  async getStats(organizationId?: string): Promise<RelawanStats> {
    const params = new URLSearchParams()
    if (organizationId) params.append('organization_id', organizationId)

    const response = await api.get<ApiResponse<RelawanStats>>('/relawan/stats', { params })
    return response.data.data
  },

  async getWAStatus(id: string): Promise<WAStatus> {
    const response = await api.get<ApiResponse<WAStatus>>(`/relawan/${id}/wa-status`)
    return response.data.data
  },

  async enableWAAccess(id: string): Promise<void> {
    await api.post(`/relawan/${id}/wa-verify`)
  },

  async revokeWAAccess(id: string): Promise<void> {
    await api.delete(`/relawan/${id}/wa-verify`)
  },
}
