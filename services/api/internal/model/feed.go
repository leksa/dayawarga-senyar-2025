package model

import (
	"time"

	"github.com/google/uuid"
)

// Feed represents information updates from field
type Feed struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	LocationID      *uuid.UUID `json:"location_id,omitempty" gorm:"type:uuid"`
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
}

func (Feed) TableName() string {
	return "information_feeds"
}
