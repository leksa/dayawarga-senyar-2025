package dto

import "time"

// APIResponse is the standard response wrapper
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
	Meta    *MetaInfo   `json:"meta,omitempty"`
}

type ErrorInfo struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

type MetaInfo struct {
	Total     int64     `json:"total,omitempty"`
	Page      int       `json:"page,omitempty"`
	Limit     int       `json:"limit,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// GeoJSON types
type GeoJSONFeatureCollection struct {
	Type     string           `json:"type"`
	Features []GeoJSONFeature `json:"features"`
}

type GeoJSONFeature struct {
	Type       string                 `json:"type"`
	ID         string                 `json:"id"`
	Geometry   *GeoJSONGeometry       `json:"geometry"`
	Properties map[string]interface{} `json:"properties"`
}

type GeoJSONGeometry struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

// LocationListResponse for GET /locations
type LocationListResponse struct {
	Type     string                    `json:"type"`
	Features []LocationFeatureResponse `json:"features"`
}

type LocationFeatureResponse struct {
	Type       string                 `json:"type"`
	ID         string                 `json:"id"`
	Geometry   *GeoJSONGeometry       `json:"geometry"`
	Properties LocationListProperties `json:"properties"`
}

type LocationListProperties struct {
	ODKSubmissionID string    `json:"odk_submission_id,omitempty"`
	Nama            string    `json:"nama"`
	Type            string    `json:"type"`
	Status          string    `json:"status"`
	AlamatSingkat   string    `json:"alamat_singkat,omitempty"`
	JumlahKK        int       `json:"jumlah_kk"`
	TotalJiwa       int       `json:"total_jiwa"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// LocationDetailResponse for GET /locations/:id
type LocationDetailResponse struct {
	ID              string                 `json:"id"`
	ODKSubmissionID string                 `json:"odk_submission_id,omitempty"`
	Type            string                 `json:"type"`
	Status          string                 `json:"status"`
	Geometry        *LocationGeometry      `json:"geometry"`
	Identitas       map[string]interface{} `json:"identitas"`
	Alamat          map[string]interface{} `json:"alamat"`
	DataPengungsi   map[string]interface{} `json:"data_pengungsi"`
	Fasilitas       map[string]interface{} `json:"fasilitas"`
	Komunikasi      map[string]interface{} `json:"komunikasi,omitempty"`
	Akses           map[string]interface{} `json:"akses,omitempty"`
	Photos          []PhotoResponse        `json:"photos"`
	Meta            LocationMeta           `json:"meta"`
}

type LocationGeometry struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
	Altitude    *float64  `json:"altitude,omitempty"`
	Accuracy    *float64  `json:"accuracy,omitempty"`
}

type PhotoResponse struct {
	Type     string `json:"type"`
	Filename string `json:"filename"`
	URL      string `json:"url"`
}

type LocationMeta struct {
	SubmittedAt   *time.Time `json:"submitted_at,omitempty"`
	UpdatedAt     time.Time  `json:"updated_at"`
	SubmitterName string     `json:"submitter,omitempty"`
}

// FeedResponse for GET /feeds
type FeedResponse struct {
	ID           string              `json:"id"`
	LocationID   *string             `json:"location_id,omitempty"`
	LocationName *string             `json:"location_name,omitempty"`
	FaskesID     *string             `json:"faskes_id,omitempty"`
	FaskesName   *string             `json:"faskes_name,omitempty"`
	Content      string              `json:"content"`
	Category     string              `json:"category"`
	Type         *string             `json:"type,omitempty"`
	Username     *string             `json:"username,omitempty"`
	Organization *string             `json:"organization,omitempty"`
	SubmittedAt  time.Time           `json:"submitted_at"`
	Coordinates  []float64           `json:"coordinates,omitempty"`
	Photos       []FeedPhotoResponse `json:"photos,omitempty"`
	Region       *FeedRegion         `json:"region,omitempty"`
}

// FeedRegion contains regional information from ODK submission
type FeedRegion struct {
	Provinsi  string `json:"provinsi,omitempty"`
	KotaKab   string `json:"kota_kab,omitempty"`
	Kecamatan string `json:"kecamatan,omitempty"`
	Desa      string `json:"desa,omitempty"`
}

// FeedPhotoResponse for feed photo data
type FeedPhotoResponse struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Filename string `json:"filename"`
	URL      string `json:"url"`
}

// FaskesListResponse for GET /faskes
type FaskesListResponse struct {
	Type     string                  `json:"type"`
	Features []FaskesFeatureResponse `json:"features"`
}

type FaskesFeatureResponse struct {
	Type       string               `json:"type"`
	ID         string               `json:"id"`
	Geometry   *GeoJSONGeometry     `json:"geometry"`
	Properties FaskesListProperties `json:"properties"`
}

type FaskesListProperties struct {
	ODKSubmissionID string    `json:"odk_submission_id,omitempty"`
	Nama            string    `json:"nama"`
	JenisFaskes     string    `json:"jenis_faskes"`
	StatusFaskes    string    `json:"status_faskes"`
	KondisiFaskes   string    `json:"kondisi_faskes,omitempty"`
	AlamatSingkat   string    `json:"alamat_singkat,omitempty"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// FaskesDetailResponse for GET /faskes/:id
type FaskesDetailResponse struct {
	ID              string                 `json:"id"`
	ODKSubmissionID string                 `json:"odk_submission_id,omitempty"`
	Nama            string                 `json:"nama"`
	JenisFaskes     string                 `json:"jenis_faskes"`
	StatusFaskes    string                 `json:"status_faskes"`
	KondisiFaskes   string                 `json:"kondisi_faskes,omitempty"`
	Geometry        *LocationGeometry      `json:"geometry"`
	Alamat          map[string]interface{} `json:"alamat"`
	Identitas       map[string]interface{} `json:"identitas"`
	Isolasi         map[string]interface{} `json:"isolasi,omitempty"`
	Infrastruktur   map[string]interface{} `json:"infrastruktur,omitempty"`
	SDM             map[string]interface{} `json:"sdm,omitempty"`
	Perbekalan      map[string]interface{} `json:"perbekalan,omitempty"`
	Klaster         map[string]interface{} `json:"klaster,omitempty"`
	Photos          []PhotoResponse        `json:"photos"`
	Meta            LocationMeta           `json:"meta"`
}

// HealthResponse for GET /health
type HealthResponse struct {
	Status    string            `json:"status"`
	Version   string            `json:"version"`
	Checks    map[string]Check  `json:"checks"`
	Timestamp time.Time         `json:"timestamp"`
}

type Check struct {
	Status    string     `json:"status"`
	LatencyMs int64      `json:"latency_ms,omitempty"`
	LastSync  *time.Time `json:"last_sync,omitempty"`
}
