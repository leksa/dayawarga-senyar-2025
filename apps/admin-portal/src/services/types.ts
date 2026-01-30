// API Response Types

// Standard API Response wrapper (backend returns { success, data, error })
export interface ApiResponse<T> {
  success: boolean
  data: T
  error?: string
}

// Pagination
export interface PaginatedResponse<T> {
  data: T[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

// User status type
export type UserStatus = 'active' | 'pending_invitation' | 'suspended'

// User (from /auth/me and /admin/users)
export interface User {
  id: string
  email: string
  name: string | null
  role: 'super_admin' | 'org_admin' | 'member'
  status: UserStatus
  is_active: boolean
  odk_web_user_id?: number | null
  odk_app_user_id?: number | null
  odk_app_user_project_id?: number | null
  last_login_at: string | null
  created_at: string
  updated_at: string
}

// User filter for listing
export interface UserFilter {
  search?: string
  role?: string
  status?: string
  page?: number
  page_size?: number
}

// Update user input
export interface UpdateUserInput {
  name?: string
  role?: 'super_admin' | 'org_admin' | 'member'
  status?: UserStatus
  is_active?: boolean
}

// Organization
export interface Organization {
  id: string
  name: string
  slug: string
  description: string | null
  email: string | null
  phone: string | null
  address: string | null
  logo_url: string | null
  odk_project_id: number | null
  is_active: boolean
  city?: string | null
  country?: string | null
  website_url?: string | null
  social_media?: Record<string, string> | null
  bidang?: Bidang[]
  created_at: string
  updated_at: string
}

// Bidang (Field/Domain)
export interface Bidang {
  id: string
  name: string
  slug: string
  description: string | null
  is_active: boolean
  created_at: string
}

// Organization Bidang (Junction table)
export interface OrganizationBidang {
  organization_id: string
  bidang_id: string
  created_at: string
  bidang?: Bidang
}

export interface OrganizationStats {
  total_members: number
  total_groups: number
  total_relawan: number
  active_relawan: number
}

export interface OrganizationMember {
  id: string
  user_id: string
  organization_id: string
  role: 'admin' | 'member'
  created_at: string
  user?: User
}

export interface CreateOrganizationInput {
  name: string
  slug?: string
  description?: string
  email?: string
  phone?: string
  address?: string
  logo_url?: string
  odk_project_id?: number
  city?: string
  country?: string
  website_url?: string
  social_media?: Record<string, string>
  // Admin invitation (optional)
  admin_email?: string
  admin_name?: string
}

// Result from creating organization with admin
export interface CreateOrganizationWithAdminResult {
  organization: Organization
  invited_admin?: User
  invitation_link?: string
  is_new_admin?: boolean
}

export interface UpdateOrganizationInput {
  name?: string
  slug?: string
  description?: string
  email?: string
  phone?: string
  address?: string
  logo_url?: string
  odk_project_id?: number
  is_active?: boolean
  city?: string
  country?: string
  website_url?: string
  social_media?: Record<string, string>
}

// Group
export interface Group {
  id: string
  organization_id: string
  name: string
  description: string | null
  is_active: boolean
  created_at: string
  updated_at: string
  organization?: Organization
}

export interface GroupStats {
  total_relawan: number
  active_relawan: number
}

export interface CreateGroupInput {
  organization_id: string
  name: string
  description?: string
}

export interface UpdateGroupInput {
  name?: string
  description?: string
  is_active?: boolean
}

// Relawan
export type RelawanStatus = 'active' | 'inactive' | 'suspended'

export interface Relawan {
  id: string
  organization_id: string
  group_id: string | null
  name: string
  phone: string | null
  email: string | null
  odk_app_user_id: number | null
  odk_app_user_token: string | null
  odk_app_user_created_at: string | null
  assigned_forms: string[]
  status: RelawanStatus
  notes: string | null
  wa_verified: boolean
  wa_verified_at: string | null
  wa_last_activity: string | null
  wa_session_count: number
  created_at: string
  updated_at: string
  organization?: Organization
  group?: Group
}

export interface WAStatus {
  verified: boolean
  verified_at: string | null
  last_activity: string | null
  session_count: number
  has_phone: boolean
}

export interface RelawanStats {
  total: number
  active: number
  with_odk_access: number
}

export interface CreateRelawanInput {
  organization_id: string
  group_id?: string
  name: string
  phone?: string
  email?: string
  assigned_forms?: string[]
  notes?: string
}

export interface UpdateRelawanInput {
  group_id?: string | null
  name?: string
  phone?: string
  email?: string
  assigned_forms?: string[]
  notes?: string
  status?: RelawanStatus
}

// API Error
export interface ApiError {
  error: string
  message?: string
  details?: Record<string, string[]>
}

// ODK Project (from ODK Central)
export interface ODKProject {
  id: number
  name: string
  description: string | null
  archived: boolean
  keyId: number | null
  created_at: string
  updated_at: string | null
}

// ODK Form (from ODK Central)
export interface ODKForm {
  xmlFormId: string
  name: string
  version: string
  state: 'open' | 'closing' | 'closed'
  enketoId: string | null
  enketoOnceId: string | null
  createdAt: string
  updatedAt: string | null
  publishedAt: string | null
}

// Project Request Status
export type ProjectRequestStatus = 'pending' | 'approved' | 'rejected'

// Project Request
export interface ProjectRequest {
  id: string
  organization_id: string
  group_id: string
  odk_project_id: number
  odk_project_name: string | null
  requested_by: string
  request_notes: string | null
  status: ProjectRequestStatus
  reviewed_by: string | null
  reviewed_at: string | null
  review_notes: string | null
  created_at: string
  updated_at: string
  organization?: Organization
  group?: Group
  requester?: User
  reviewer?: User
}

// Create Project Request Input
export interface CreateProjectRequestInput {
  odk_project_id: number
  notes?: string
}

// Review Project Request Input
export interface ReviewProjectRequestInput {
  action: 'approve' | 'reject'
  notes?: string
}

// Project Request Filter
export interface ProjectRequestFilter {
  organization_id?: string
  group_id?: string
  status?: ProjectRequestStatus
  page?: number
  page_size?: number
}

// QR Code Response
export interface QRCodeResponse {
  relawan_id: string
  relawan_name: string
  group_name: string
  project_id: number
  qr_code_data: string
  created_at: string
}

// Group with ODK fields (extended)
export interface GroupWithODK extends Group {
  leader_id: string | null
  odk_project_id: number | null
  odk_project_manager_created: boolean
  leader?: User
}

// Filter types
export interface OrganizationFilter {
  search?: string
  is_active?: boolean
  page?: number
  page_size?: number
}

export interface GroupFilter {
  organization_id?: string
  search?: string
  is_active?: boolean
  page?: number
  page_size?: number
}

export interface RelawanFilter {
  organization_id?: string
  group_id?: string
  status?: RelawanStatus
  search?: string
  has_odk_access?: boolean
  page?: number
  page_size?: number
}

// ROPC Token Exchange (Resource Owner Password Credentials)
export interface ROPCTokenRequest {
  username: string
  password: string
}

export interface ROPCTokenResponse {
  access_token: string
  token_type: string
  expires_in: number
  refresh_token?: string
  id_token?: string
  scope: string
}

// User ODK Project Assignment
export interface UserProjectAssignment {
  project_id: number
  project_name: string
  role_id: number
  role_name: string
}

// Assign User to ODK Project Input
export interface AssignProjectRoleInput {
  odk_project_id: number
  role: 'manager' | 'viewer'
}

// Assign Project Role Result
export interface AssignProjectRoleResult {
  user: User
  odk_web_user_id: number
  project_id: number
  role_id: number
  role_name: string
  odk_app_user_id?: number | null
  has_qr_code: boolean
}

// User QR Code Response
export interface UserQRCodeResponse {
  user_id: string
  user_name: string
  user_email: string
  project_id: number
  project_name: string
  qr_code_data: string
}

// Project Assignment Info (for viewing all users assigned to a project)
export interface ProjectAssignmentInfo {
  project: ODKProject
  assignments: ProjectUserAssignment[]
}

export interface ProjectUserAssignment {
  user_id?: string
  odk_web_user_id: number
  email: string
  display_name: string
  role_id: number
  role_name: string
}

// Organization ODK Assignment
export interface AssignODKProjectToOrgInput {
  odk_project_id: number
}

export interface AdminODKResult {
  user_id: string
  user_email: string
  user_name: string
  odk_web_user_id: number
  odk_app_user_id: number
  has_qr_code: boolean
  error?: string
}

export interface AssignODKProjectResult {
  organization: Organization
  odk_project_id: number
  odk_project_name: string
  admins_processed: AdminODKResult[]
}

export interface AdminODKStatus {
  user_id: string
  user_email: string
  user_name: string
  user_status: string
  has_web_user: boolean
  has_app_user: boolean
  has_qr_code: boolean
  odk_web_user_id?: number
  odk_app_user_id?: number
}

export interface OrganizationODKInfo {
  organization: Organization
  odk_project?: ODKProject
  admin_odk_status: AdminODKStatus[]
}

// Feed Photo
export interface FeedPhoto {
  id: string
  feed_id: string
  filename: string
  original_path: string
  storage_path?: string
  storage_url?: string
  size_bytes?: number
  content_type?: string
  sync_status: string
}

// Feed (from ODK submissions)
export interface Feed {
  id: string
  project_id: number
  form_id: string
  submission_id: string
  instance_id: string
  username?: string
  device_id?: string
  notes?: string
  sync_status: string
  created_at: string
  updated_at: string
  photos?: FeedPhoto[]
}

// Organization Activity (feed with relawan info)
export interface OrganizationActivity {
  feed: Feed
  relawan_name?: string
}
