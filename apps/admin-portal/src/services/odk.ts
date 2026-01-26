import api from './api'
import type {
  ApiResponse,
  PaginatedResponse,
  ODKProject,
  ODKForm,
  ProjectRequest,
  CreateProjectRequestInput,
  ReviewProjectRequestInput,
  ProjectRequestFilter,
  QRCodeResponse,
  Relawan
} from './types'

// ODK Projects Service
export const odkProjectService = {
  // List all available ODK projects
  async list(): Promise<ODKProject[]> {
    const response = await api.get<ApiResponse<ODKProject[]>>('/odk/projects')
    return response.data.data
  },

  // Get a specific ODK project
  async getById(id: number): Promise<ODKProject> {
    const response = await api.get<ApiResponse<ODKProject>>(`/odk/projects/${id}`)
    return response.data.data
  },

  // List forms in an ODK project
  async listForms(projectId: number): Promise<ODKForm[]> {
    const response = await api.get<ApiResponse<ODKForm[]>>(`/odk/projects/${projectId}/forms`)
    return response.data.data
  }
}

// Project Request Service
export const projectRequestService = {
  // Create a project request for a group
  async create(groupId: string, input: CreateProjectRequestInput): Promise<ProjectRequest> {
    const response = await api.post<ApiResponse<ProjectRequest>>(
      `/groups/${groupId}/project-request`,
      input
    )
    return response.data.data
  },

  // List project requests for a group
  async listByGroup(
    groupId: string,
    page = 1,
    pageSize = 20
  ): Promise<PaginatedResponse<ProjectRequest>> {
    const response = await api.get<
      ApiResponse<ProjectRequest[]> & { meta: { total: number; page: number; page_size: number; total_pages: number } }
    >(`/groups/${groupId}/project-requests`, {
      params: { page, page_size: pageSize }
    })
    return {
      data: response.data.data,
      total: response.data.meta.total,
      page: response.data.meta.page,
      page_size: response.data.meta.page_size,
      total_pages: response.data.meta.total_pages
    }
  },

  // Admin: List all project requests
  async list(filter: ProjectRequestFilter = {}): Promise<PaginatedResponse<ProjectRequest>> {
    const response = await api.get<
      ApiResponse<ProjectRequest[]> & { meta: { total: number; page: number; page_size: number; total_pages: number } }
    >('/admin/project-requests', {
      params: {
        organization_id: filter.organization_id,
        status: filter.status,
        page: filter.page || 1,
        page_size: filter.page_size || 20
      }
    })
    return {
      data: response.data.data,
      total: response.data.meta.total,
      page: response.data.meta.page,
      page_size: response.data.meta.page_size,
      total_pages: response.data.meta.total_pages
    }
  },

  // Admin: Get a specific project request
  async getById(id: string): Promise<ProjectRequest> {
    const response = await api.get<ApiResponse<ProjectRequest>>(`/admin/project-requests/${id}`)
    return response.data.data
  },

  // Admin: Approve or reject a project request
  async review(id: string, input: ReviewProjectRequestInput): Promise<ProjectRequest> {
    const response = await api.put<ApiResponse<ProjectRequest>>(
      `/admin/project-requests/${id}`,
      input
    )
    return response.data.data
  }
}

// Relawan ODK Service
export const relawanODKService = {
  // Create ODK App User for a relawan
  async createAppUser(relawanId: string): Promise<Relawan> {
    const response = await api.post<ApiResponse<Relawan>>(`/relawan/${relawanId}/odk-app-user`)
    return response.data.data
  },

  // Revoke ODK App User for a relawan
  async revokeAppUser(relawanId: string): Promise<void> {
    await api.delete(`/relawan/${relawanId}/odk-app-user`)
  },

  // Get QR code data for a relawan
  async getQRCode(relawanId: string): Promise<QRCodeResponse> {
    const response = await api.get<ApiResponse<QRCodeResponse>>(`/relawan/${relawanId}/odk-qr-code`)
    return response.data.data
  },

  // Assign forms to a relawan
  async assignForms(relawanId: string, formIds: string[]): Promise<void> {
    await api.post(`/relawan/${relawanId}/odk-forms`, { form_ids: formIds })
  },

  // Bulk create ODK App Users for all relawan in a group
  async createGroupAppUsers(groupId: string): Promise<{ created_count: number }> {
    const response = await api.post<ApiResponse<{ created_count: number }>>(
      `/groups/${groupId}/odk-app-users`
    )
    return response.data.data
  }
}

export default {
  projects: odkProjectService,
  requests: projectRequestService,
  relawan: relawanODKService
}
