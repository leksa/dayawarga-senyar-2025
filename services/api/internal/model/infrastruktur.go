package model

import (
	"time"

	"github.com/google/uuid"
)

// Infrastruktur represents a road/bridge infrastructure record
type Infrastruktur struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	ODKSubmissionID *string    `json:"odk_submission_id,omitempty" gorm:"column:odk_submission_id"`
	EntityID        string     `json:"entity_id" gorm:"column:entity_id;index"`
	ObjectID        string     `json:"object_id" gorm:"column:object_id"`

	// Basic info
	Nama      string `json:"nama" gorm:"not null"`
	Jenis     string `json:"jenis" gorm:"not null"`            // "Jalan" or "Jembatan"
	StatusJln string `json:"status_jln" gorm:"column:status_jln"` // "Nasional" or "Daerah"

	// Location
	NamaProvinsi  string   `json:"nama_provinsi" gorm:"column:nama_provinsi"`
	NamaKabupaten string   `json:"nama_kabupaten" gorm:"column:nama_kabupaten"`
	Latitude      *float64 `json:"latitude,omitempty" gorm:"-"`
	Longitude     *float64 `json:"longitude,omitempty" gorm:"-"`

	// Status fields (dynamic - updated by relawan)
	StatusAkses       string `json:"status_akses" gorm:"column:status_akses"`             // "dapat_diakses" or "akses_terputus"
	KeteranganBencana string `json:"keterangan_bencana" gorm:"column:keterangan_bencana"` // multi-select as comma-separated
	Dampak            string `json:"dampak" gorm:"column:dampak;type:text"`

	// Penanganan fields
	StatusPenanganan string `json:"status_penanganan" gorm:"column:status_penanganan"`
	PenangananDetail string `json:"penanganan_detail" gorm:"column:penanganan_detail;type:text"`
	Bailey           string `json:"bailey" gorm:"column:bailey"`     // For bridges only
	Progress         int    `json:"progress" gorm:"column:progress"` // 0-100
	TargetSelesai    string `json:"target_selesai" gorm:"column:target_selesai"`

	// Source info
	BaselineSumber string `json:"baseline_sumber" gorm:"column:baseline_sumber"` // "BNPB/PU"
	UpdateBy       string `json:"update_by" gorm:"column:update_by"`

	// Raw data from ODK
	RawData JSONB `json:"raw_data,omitempty" gorm:"type:jsonb;column:raw_data"`

	// Metadata
	SubmitterName *string    `json:"submitter_name,omitempty" gorm:"column:submitter_name"`
	SubmittedAt   *time.Time `json:"submitted_at,omitempty" gorm:"column:submitted_at"`
	CreatedAt     time.Time  `json:"created_at" gorm:"column:created_at"`
	UpdatedAt     time.Time  `json:"updated_at" gorm:"column:updated_at"`
	SyncedAt      *time.Time `json:"synced_at,omitempty" gorm:"column:synced_at"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty" gorm:"column:deleted_at"`
}

func (Infrastruktur) TableName() string {
	return "infrastruktur"
}

// InfrastrukturPhoto represents photo attachments for infrastructure
type InfrastrukturPhoto struct {
	ID              uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	InfrastrukturID uuid.UUID `json:"infrastruktur_id" gorm:"type:uuid;not null;index"`
	PhotoType       string    `json:"photo_type" gorm:"not null"` // foto_1, foto_2, foto_3, foto_4
	Filename        string    `json:"filename" gorm:"not null"`
	StoragePath     *string   `json:"storage_path,omitempty"`
	IsCached        bool      `json:"is_cached" gorm:"default:false"`
	FileSize        *int      `json:"file_size,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

func (InfrastrukturPhoto) TableName() string {
	return "infrastruktur_photos"
}
