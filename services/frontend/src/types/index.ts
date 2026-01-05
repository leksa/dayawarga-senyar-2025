export interface MapMarker {
  id: string
  name: string
  type: string
  lat: number
  lng: number
  status?: string
  alamatSingkat?: string
  jumlahKK?: number
  totalJiwa?: number
  updatedAt?: string
}

export interface LocationDetails {
  id: string
  name: string
  status: string
  statusId: string
  image?: string
  capacity?: string
  type?: string
  personnel?: string
  manager?: string
  hours?: string
  phone?: string
  verified?: string
  lastInspection?: string
}

export interface Layer {
  id: string
  name: string
  icon: string
  color: string
  enabled: boolean
  category: 'emergency' | 'environment' | 'infrastructure' | 'feeds'
}

export interface InfoUpdate {
  id: string
  timestamp: string
  username: string
  organization: string
  location: string
  locationId?: string
  content: string
  category: string
  type: string
}
