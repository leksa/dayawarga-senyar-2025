import api from './api'
import type {
  Group,
  GroupStats,
  CreateGroupInput,
  UpdateGroupInput,
  GroupFilter,
  Relawan,
  ApiResponse,
} from './types'

interface GroupListResponse {
  groups: Group[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export const groupService = {
  /**
   * List groups with optional filters
   */
  async list(filter?: GroupFilter): Promise<GroupListResponse> {
    const params = new URLSearchParams()
    if (filter?.organization_id) params.append('organization_id', filter.organization_id)
    if (filter?.search) params.append('search', filter.search)
    if (filter?.is_active !== undefined) params.append('is_active', String(filter.is_active))
    if (filter?.page) params.append('page', String(filter.page))
    if (filter?.page_size) params.append('page_size', String(filter.page_size))

    const response = await api.get<ApiResponse<GroupListResponse>>('/groups', { params })
    return response.data.data
  },

  /**
   * Get group by ID
   */
  async get(id: string): Promise<Group> {
    const response = await api.get<ApiResponse<Group>>(`/groups/${id}`)
    return response.data.data
  },

  /**
   * Create new group
   */
  async create(input: CreateGroupInput): Promise<Group> {
    const response = await api.post<ApiResponse<Group>>('/groups', input)
    return response.data.data
  },

  /**
   * Update group
   */
  async update(id: string, input: UpdateGroupInput): Promise<Group> {
    const response = await api.put<ApiResponse<Group>>(`/groups/${id}`, input)
    return response.data.data
  },

  /**
   * Delete group (soft delete)
   */
  async delete(id: string): Promise<void> {
    await api.delete(`/groups/${id}`)
  },

  /**
   * Get group statistics
   */
  async getStats(id: string): Promise<GroupStats> {
    const response = await api.get<ApiResponse<GroupStats>>(`/groups/${id}/stats`)
    return response.data.data
  },

  /**
   * Get relawan in group
   */
  async getRelawan(groupId: string): Promise<Relawan[]> {
    const response = await api.get<ApiResponse<Relawan[]>>(`/groups/${groupId}/relawan`)
    return response.data.data
  },
}
