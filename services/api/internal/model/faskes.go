package model

import (
	"time"

	"github.com/google/uuid"
)

// Faskes represents a health facility (fasilitas kesehatan)
type Faskes struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	ODKSubmissionID *string    `json:"odk_submission_id,omitempty" gorm:"column:odk_submission_id;uniqueIndex"`
	Nama            string     `json:"nama" gorm:"not null"`
	JenisFaskes     string     `json:"jenis_faskes" gorm:"column:jenis_faskes"` // rumah_sakit, puskesmas, klinik, posko_kes_darurat
	StatusFaskes    string     `json:"status_faskes" gorm:"column:status_faskes;default:'operasional'"` // operasional, non_aktif
	KondisiFaskes   *string    `json:"kondisi_faskes,omitempty" gorm:"column:kondisi_faskes"` // tidak_rusak, rusak_ringan, rusak_sedang, rusak_berat, hancur_total

	// Geometry stored as WKT for simplicity, will be converted to GeoJSON in response
	Latitude  *float64 `json:"latitude,omitempty" gorm:"-"`
	Longitude *float64 `json:"longitude,omitempty" gorm:"-"`

	// JSONB fields
	Alamat        JSONB `json:"alamat,omitempty" gorm:"type:jsonb"`
	Identitas     JSONB `json:"identitas,omitempty" gorm:"type:jsonb"`
	Isolasi       JSONB `json:"isolasi,omitempty" gorm:"type:jsonb"`
	Infrastruktur JSONB `json:"infrastruktur,omitempty" gorm:"type:jsonb"`
	SDM           JSONB `json:"sdm,omitempty" gorm:"type:jsonb"`
	Perbekalan    JSONB `json:"perbekalan,omitempty" gorm:"type:jsonb"`
	Klaster       JSONB `json:"klaster,omitempty" gorm:"type:jsonb"`
	RawData       JSONB `json:"raw_data,omitempty" gorm:"type:jsonb;column:raw_data"`

	// Metadata
	SubmitterName *string    `json:"submitter_name,omitempty" gorm:"column:submitter_name"`
	SubmittedAt   *time.Time `json:"submitted_at,omitempty" gorm:"column:submitted_at"`
	CreatedAt     time.Time  `json:"created_at" gorm:"column:created_at"`
	UpdatedAt     time.Time  `json:"updated_at" gorm:"column:updated_at"`
	SyncedAt      *time.Time `json:"synced_at,omitempty" gorm:"column:synced_at"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty" gorm:"column:deleted_at"`
}

func (Faskes) TableName() string {
	return "faskes"
}

// FaskesPhoto represents photo attachments for faskes
type FaskesPhoto struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	FaskesID    uuid.UUID `json:"faskes_id" gorm:"type:uuid;not null"`
	PhotoType   string    `json:"photo_type" gorm:"not null"`
	Filename    string    `json:"filename" gorm:"not null"`
	StoragePath *string   `json:"storage_path,omitempty"`
	IsCached    bool      `json:"is_cached" gorm:"default:false"`
	FileSize    *int      `json:"file_size,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

func (FaskesPhoto) TableName() string {
	return "faskes_photos"
}
