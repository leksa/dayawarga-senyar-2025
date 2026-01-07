package model

import (
	"time"

	"github.com/google/uuid"
)

// Feed represents information updates from field
type Feed struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	LocationID      *uuid.UUID `json:"location_id,omitempty" gorm:"type:uuid"`
	FaskesID        *uuid.UUID `json:"faskes_id,omitempty" gorm:"type:uuid"`
	ODKSubmissionID *string    `json:"odk_submission_id,omitempty" gorm:"column:odk_submission_id"`

	Content      string  `json:"content" gorm:"not null"`
	Category     string  `json:"category" gorm:"default:'informasi'"`
	Type         *string `json:"type,omitempty"`
	Username     *string `json:"username,omitempty"`
	Organization *string `json:"organization,omitempty"`

	// Geometry
	Latitude  *float64 `json:"latitude,omitempty" gorm:"-"`
	Longitude *float64 `json:"longitude,omitempty" gorm:"-"`

	RawData JSONB `json:"raw_data,omitempty" gorm:"type:jsonb;column:raw_data"`

	SubmittedAt *time.Time `json:"submitted_at,omitempty" gorm:"column:submitted_at"`
	CreatedAt   time.Time  `json:"created_at" gorm:"column:created_at"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"column:updated_at"`

	// Joined fields
	LocationName *string `json:"location_name,omitempty" gorm:"-"`
	FaskesName   *string `json:"faskes_name,omitempty" gorm:"-"`

	// Relations
	Photos []FeedPhoto `json:"photos,omitempty" gorm:"foreignKey:FeedID"`
}

func (Feed) TableName() string {
	return "information_feeds"
}

// FeedPhoto represents a photo attachment for a feed
type FeedPhoto struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	FeedID      uuid.UUID `json:"feed_id" gorm:"type:uuid;not null"`
	PhotoType   string    `json:"photo_type" gorm:"default:'foto'"`
	Filename    string    `json:"filename" gorm:"not null"`
	StoragePath *string   `json:"storage_path,omitempty"`
	IsCached    bool      `json:"is_cached" gorm:"default:false"`
	FileSize    *int      `json:"file_size,omitempty"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at"`
}

func (FeedPhoto) TableName() string {
	return "feed_photos"
}
