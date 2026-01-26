import api from './api'
import type {
  User,
  UserFilter,
  UpdateUserInput,
  ApiResponse,
  UserProjectAssignment,
  AssignProjectRoleInput,
  AssignProjectRoleResult,
  ProjectAssignmentInfo,
  UserQRCodeResponse,
} from './types'

interface UserListResponse {
  users: User[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export const userService = {
  /**
   * List all users (super_admin only)
   */
  async list(filter?: UserFilter): Promise<UserListResponse> {
    const params = new URLSearchParams()
    if (filter?.search) params.append('search', filter.search)
    if (filter?.role) params.append('role', filter.role)
    if (filter?.status) params.append('status', filter.status)
    if (filter?.page) params.append('page', String(filter.page))
    if (filter?.page_size) params.append('page_size', String(filter.page_size))

    const response = await api.get<ApiResponse<UserListResponse>>('/admin/users', { params })
    return response.data.data
  },

  /**
   * Get user by ID (super_admin only)
   */
  async get(id: string): Promise<User> {
    const response = await api.get<ApiResponse<User>>(`/admin/users/${id}`)
    return response.data.data
  },

  /**
   * Update user (super_admin only)
   */
  async update(id: string, input: UpdateUserInput): Promise<User> {
    const response = await api.put<ApiResponse<User>>(`/admin/users/${id}`, input)
    return response.data.data
  },

  /**
   * Get user's ODK project assignments
   */
  async getODKRoles(userId: string): Promise<UserProjectAssignment[]> {
    const response = await api.get<ApiResponse<UserProjectAssignment[]>>(`/admin/users/${userId}/odk-roles`)
    return response.data.data || []
  },

  /**
   * Assign user to an ODK project
   */
  async assignODKRole(userId: string, input: AssignProjectRoleInput): Promise<AssignProjectRoleResult> {
    const response = await api.post<ApiResponse<AssignProjectRoleResult>>(`/admin/users/${userId}/odk-roles`, input)
    return response.data.data
  },

  /**
   * Remove user from an ODK project
   */
  async removeODKRole(userId: string, projectId: number, role: 'manager' | 'viewer'): Promise<void> {
    await api.delete(`/admin/users/${userId}/odk-roles/${projectId}`, { data: { role } })
  },

  /**
   * Get all user assignments for an ODK project
   */
  async getProjectAssignments(projectId: number): Promise<ProjectAssignmentInfo> {
    const response = await api.get<ApiResponse<ProjectAssignmentInfo>>(`/admin/odk-projects/${projectId}/assignments`)
    return response.data.data
  },

  /**
   * Get QR code data for a user's ODK Collect access
   */
  async getQRCode(userId: string): Promise<UserQRCodeResponse> {
    const response = await api.get<ApiResponse<UserQRCodeResponse>>(`/admin/users/${userId}/odk-qr-code`)
    return response.data.data
  },
}
