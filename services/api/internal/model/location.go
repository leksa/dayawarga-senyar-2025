package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// JSONB represents a JSONB column in PostgreSQL
type JSONB map[string]interface{}

func (j JSONB) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, j)
}

// Location represents a posko/shelter location
type Location struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	ODKSubmissionID *string    `json:"odk_submission_id,omitempty" gorm:"column:odk_submission_id"`
	Nama            string     `json:"nama" gorm:"not null"`
	Type            string     `json:"type" gorm:"default:'posko'"`
	Status          string     `json:"status" gorm:"default:'operational'"`

	// Geometry stored as WKT for simplicity, will be converted to GeoJSON in response
	Latitude  *float64 `json:"latitude,omitempty" gorm:"-"`
	Longitude *float64 `json:"longitude,omitempty" gorm:"-"`
	GeoMeta   JSONB    `json:"geo_meta,omitempty" gorm:"type:jsonb"`

	// JSONB fields
	Identitas     JSONB `json:"identitas,omitempty" gorm:"type:jsonb"`
	Alamat        JSONB `json:"alamat,omitempty" gorm:"type:jsonb"`
	DataPengungsi JSONB `json:"data_pengungsi,omitempty" gorm:"type:jsonb;column:data_pengungsi"`
	Fasilitas     JSONB `json:"fasilitas,omitempty" gorm:"type:jsonb"`
	Komunikasi    JSONB `json:"komunikasi,omitempty" gorm:"type:jsonb"`
	Akses         JSONB `json:"akses,omitempty" gorm:"type:jsonb"`
	RawData       JSONB `json:"raw_data,omitempty" gorm:"type:jsonb;column:raw_data"`

	// Source info
	BaselineSumber string `json:"baseline_sumber" gorm:"column:baseline_sumber"`

	// Metadata
	SubmitterName *string    `json:"submitter_name,omitempty" gorm:"column:submitter_name"`
	SubmittedAt   *time.Time `json:"submitted_at,omitempty" gorm:"column:submitted_at"`
	CreatedAt     time.Time  `json:"created_at" gorm:"column:created_at"`
	UpdatedAt     time.Time  `json:"updated_at" gorm:"column:updated_at"`
	SyncedAt      *time.Time `json:"synced_at,omitempty" gorm:"column:synced_at"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty" gorm:"column:deleted_at"`
}

func (Location) TableName() string {
	return "locations"
}

// LocationPhoto represents photo attachments
type LocationPhoto struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	LocationID  uuid.UUID `json:"location_id" gorm:"type:uuid;not null"`
	PhotoType   string    `json:"photo_type" gorm:"not null"`
	Filename    string    `json:"filename" gorm:"not null"`
	StoragePath *string   `json:"storage_path,omitempty"`
	IsCached    bool      `json:"is_cached" gorm:"default:false"`
	FileSize    *int      `json:"file_size,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

func (LocationPhoto) TableName() string {
	return "location_photos"
}
