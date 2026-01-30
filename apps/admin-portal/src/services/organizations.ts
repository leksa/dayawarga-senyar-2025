import api from './api'
import type {
  Organization,
  OrganizationStats,
  OrganizationMember,
  CreateOrganizationInput,
  CreateOrganizationWithAdminResult,
  UpdateOrganizationInput,
  OrganizationFilter,
  Group,
  Relawan,
  ApiResponse,
  AssignODKProjectToOrgInput,
  AssignODKProjectResult,
  OrganizationODKInfo,
  OrganizationActivity,
} from './types'

interface OrganizationListResponse {
  organizations: Organization[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export const organizationService = {
  /**
   * List organizations with optional filters
   */
  async list(filter?: OrganizationFilter): Promise<OrganizationListResponse> {
    const params = new URLSearchParams()
    if (filter?.search) params.append('search', filter.search)
    if (filter?.is_active !== undefined) params.append('is_active', String(filter.is_active))
    if (filter?.page) params.append('page', String(filter.page))
    if (filter?.page_size) params.append('page_size', String(filter.page_size))

    const response = await api.get<ApiResponse<OrganizationListResponse>>('/organizations', { params })
    return response.data.data
  },

  /**
   * Get organization by ID
   */
  async get(id: string): Promise<Organization> {
    const response = await api.get<ApiResponse<Organization>>(`/organizations/${id}`)
    return response.data.data
  },

  /**
   * Create new organization
   */
  async create(input: CreateOrganizationInput): Promise<Organization> {
    const response = await api.post<ApiResponse<Organization>>('/organizations', input)
    return response.data.data
  },

  /**
   * Create organization with admin invitation
   * Uses the /organizations/with-admin endpoint
   */
  async createWithAdmin(input: CreateOrganizationInput): Promise<CreateOrganizationWithAdminResult> {
    const response = await api.post<ApiResponse<CreateOrganizationWithAdminResult>>(
      '/organizations/with-admin',
      input
    )
    return response.data.data
  },

  /**
   * Update organization
   */
  async update(id: string, input: UpdateOrganizationInput): Promise<Organization> {
    const response = await api.put<ApiResponse<Organization>>(`/organizations/${id}`, input)
    return response.data.data
  },

  /**
   * Delete organization (soft delete)
   */
  async delete(id: string): Promise<void> {
    await api.delete(`/organizations/${id}`)
  },

  /**
   * Get organization statistics
   */
  async getStats(id: string): Promise<OrganizationStats> {
    const response = await api.get<ApiResponse<OrganizationStats>>(`/organizations/${id}/stats`)
    return response.data.data
  },

  /**
   * Add member to organization
   */
  async addMember(orgId: string, userId: string, role: 'admin' | 'member'): Promise<OrganizationMember> {
    const response = await api.post<ApiResponse<OrganizationMember>>(`/organizations/${orgId}/members`, {
      user_id: userId,
      role,
    })
    return response.data.data
  },

  /**
   * Remove member from organization
   */
  async removeMember(orgId: string, userId: string): Promise<void> {
    await api.delete(`/organizations/${orgId}/members/${userId}`)
  },

  /**
   * Update member role
   */
  async updateMemberRole(orgId: string, userId: string, role: 'admin' | 'member'): Promise<void> {
    await api.put(`/organizations/${orgId}/members/${userId}/role`, { role })
  },

  /**
   * Get groups in organization
   */
  async getGroups(orgId: string): Promise<Group[]> {
    const response = await api.get<ApiResponse<Group[]>>(`/organizations/${orgId}/groups`)
    return response.data.data
  },

  /**
   * Get relawan in organization
   */
  async getRelawan(orgId: string): Promise<Relawan[]> {
    const response = await api.get<ApiResponse<Relawan[]>>(`/organizations/${orgId}/relawan`)
    return response.data.data
  },

  /**
   * Assign ODK project to organization
   * This will:
   * 1. Create ODK Web User for each org admin (with password - immediately active)
   * 2. Assign Project Manager role to each admin
   * 3. Create App User for each admin (ODK Collect access)
   * 4. Update organization with ODK project ID
   */
  async assignODKProject(orgId: string, input: AssignODKProjectToOrgInput): Promise<AssignODKProjectResult> {
    const response = await api.post<ApiResponse<AssignODKProjectResult>>(
      `/organizations/${orgId}/odk-project`,
      input
    )
    return response.data.data
  },

  /**
   * Remove ODK project assignment from organization
   */
  async removeODKProject(orgId: string): Promise<void> {
    await api.delete(`/organizations/${orgId}/odk-project`)
  },

  async getODKInfo(orgId: string): Promise<OrganizationODKInfo> {
    const response = await api.get<ApiResponse<OrganizationODKInfo>>(`/organizations/${orgId}/odk-info`)
    return response.data.data
  },

  async getActivities(
    orgId: string,
    page = 1,
    pageSize = 10
  ): Promise<{ activities: OrganizationActivity[]; total: number }> {
    const response = await api.get<
      ApiResponse<{ activities: OrganizationActivity[]; total: number }>
    >(`/organizations/${orgId}/activities`, {
      params: { page, page_size: pageSize },
    })
    return response.data.data
  },
}
