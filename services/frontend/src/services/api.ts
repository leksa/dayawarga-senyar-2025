const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1'

export interface LocationFeature {
  type: 'Feature'
  id: string
  geometry: {
    type: 'Point'
    coordinates: [number, number]
  }
  properties: {
    odk_submission_id?: string
    nama: string
    type: string
    status: string
    alamat_singkat?: string
    jumlah_kk: number
    total_jiwa: number
    updated_at: string
  }
}

export interface LocationListResponse {
  type: 'FeatureCollection'
  features: LocationFeature[]
}

export interface LocationDetail {
  id: string
  odk_submission_id?: string
  type: string
  status: string
  geometry: {
    type: 'Point'
    coordinates: [number, number]
    altitude?: number
    accuracy?: number
  }
  identitas: Record<string, unknown>
  alamat: Record<string, unknown>
  data_pengungsi: Record<string, unknown>
  fasilitas: Record<string, unknown>
  komunikasi?: Record<string, unknown>
  akses?: Record<string, unknown>
  photos: {
    type: string
    filename: string
    url: string
  }[]
  meta: {
    submitted_at?: string
    updated_at: string
    submitter?: string
  }
}

export interface Feed {
  id: string
  location_id?: string
  location_name?: string
  content: string
  category: string
  type?: string
  username?: string
  organization?: string
  submitted_at: string
  coordinates?: [number, number]
}

export interface Photo {
  id: string
  photo_type: string
  filename: string
  is_cached: boolean
  file_size?: number
  url?: string
  created_at: string
}

export interface APIResponse<T> {
  success: boolean
  data?: T
  error?: {
    code: string
    message: string
    details?: unknown
  }
  meta?: {
    total: number
    page: number
    limit: number
    timestamp: string
  }
}

export interface LocationFilter {
  type?: string
  status?: string
  search?: string
  bbox?: string
  page?: number
  limit?: number
}

export interface FeedFilter {
  category?: string
  type?: string
  location_id?: string
  location_name?: string
  search?: string
  since?: string // ISO date string for filtering feeds since a date
  page?: number
  limit?: number
}

async function fetchAPI<T>(endpoint: string, options?: RequestInit): Promise<APIResponse<T>> {
  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  })

  if (!response.ok) {
    const error = await response.json().catch(() => ({
      error: { code: 'UNKNOWN_ERROR', message: 'An unknown error occurred' }
    }))
    return error
  }

  return response.json()
}

export interface SyncStatus {
  id: string
  form_id: string
  last_sync_time: string
  last_etag?: string
  last_record_count: number
  total_records: number
  status: string
  error_message?: string
  created_at: string
  updated_at: string
}

// Faskes (Health Facilities) types
export interface FaskesFeature {
  type: 'Feature'
  id: string
  geometry: {
    type: 'Point'
    coordinates: [number, number]
  }
  properties: {
    odk_submission_id?: string
    nama: string
    jenis_faskes: string
    status_faskes: string
    kondisi_faskes?: string
    alamat_singkat?: string
    updated_at: string
  }
}

export interface FaskesListResponse {
  type: 'FeatureCollection'
  features: FaskesFeature[]
}

export interface FaskesDetail {
  id: string
  odk_submission_id?: string
  nama: string
  jenis_faskes: string
  status_faskes: string
  kondisi_faskes?: string
  geometry: {
    type: 'Point'
    coordinates: [number, number]
    altitude?: number
    accuracy?: number
  }
  alamat: Record<string, unknown>
  identitas: Record<string, unknown>
  isolasi?: Record<string, unknown>
  infrastruktur?: Record<string, unknown>
  sdm?: Record<string, unknown>
  perbekalan?: Record<string, unknown>
  klaster?: Record<string, unknown>
  photos: {
    type: string
    filename: string
    url: string
  }[]
  meta: {
    submitted_at?: string
    updated_at: string
    submitter?: string
  }
}

export interface FaskesFilter {
  jenis_faskes?: string
  status_faskes?: string
  kondisi_faskes?: string
  search?: string
  bbox?: string
  page?: number
  limit?: number
}

export const api = {
  async getLocations(filter?: LocationFilter): Promise<APIResponse<LocationListResponse>> {
    const params = new URLSearchParams()
    if (filter?.type) params.append('type', filter.type)
    if (filter?.status) params.append('status', filter.status)
    if (filter?.search) params.append('search', filter.search)
    if (filter?.bbox) params.append('bbox', filter.bbox)
    if (filter?.page) params.append('page', filter.page.toString())
    if (filter?.limit) params.append('limit', filter.limit.toString())

    const query = params.toString()
    return fetchAPI<LocationListResponse>(`/locations${query ? `?${query}` : ''}`)
  },

  async getLocationById(id: string): Promise<APIResponse<LocationDetail>> {
    return fetchAPI<LocationDetail>(`/locations/${id}`)
  },

  async getFeeds(filter?: FeedFilter): Promise<APIResponse<Feed[]>> {
    const params = new URLSearchParams()
    if (filter?.category) params.append('category', filter.category)
    if (filter?.type) params.append('type', filter.type)
    if (filter?.location_id) params.append('location_id', filter.location_id)
    if (filter?.location_name) params.append('location_name', filter.location_name)
    if (filter?.search) params.append('search', filter.search)
    if (filter?.since) params.append('since', filter.since)
    if (filter?.page) params.append('page', filter.page.toString())
    if (filter?.limit) params.append('limit', filter.limit.toString())

    const query = params.toString()
    return fetchAPI<Feed[]>(`/feeds${query ? `?${query}` : ''}`)
  },

  async getFeedsByLocation(locationId: string, filter?: FeedFilter): Promise<APIResponse<Feed[]>> {
    const params = new URLSearchParams()
    if (filter?.page) params.append('page', filter.page.toString())
    if (filter?.limit) params.append('limit', filter.limit.toString())

    const query = params.toString()
    return fetchAPI<Feed[]>(`/locations/${locationId}/feeds${query ? `?${query}` : ''}`)
  },

  async getPhotosByLocation(locationId: string): Promise<APIResponse<Photo[]>> {
    return fetchAPI<Photo[]>(`/locations/${locationId}/photos`)
  },

  getPhotoUrl(photoId: string): string {
    return `${API_BASE_URL}/photos/${photoId}/file`
  },

  async getSyncStatus(): Promise<APIResponse<SyncStatus>> {
    return fetchAPI<SyncStatus>('/sync/status')
  },

  // Faskes (Health Facilities) endpoints
  async getFaskes(filter?: FaskesFilter): Promise<APIResponse<FaskesListResponse>> {
    const params = new URLSearchParams()
    if (filter?.jenis_faskes) params.append('jenis_faskes', filter.jenis_faskes)
    if (filter?.status_faskes) params.append('status_faskes', filter.status_faskes)
    if (filter?.kondisi_faskes) params.append('kondisi_faskes', filter.kondisi_faskes)
    if (filter?.search) params.append('search', filter.search)
    if (filter?.bbox) params.append('bbox', filter.bbox)
    if (filter?.page) params.append('page', filter.page.toString())
    if (filter?.limit) params.append('limit', filter.limit.toString())

    const query = params.toString()
    return fetchAPI<FaskesListResponse>(`/faskes${query ? `?${query}` : ''}`)
  },

  async getFaskesById(id: string): Promise<APIResponse<FaskesDetail>> {
    return fetchAPI<FaskesDetail>(`/faskes/${id}`)
  },
}
