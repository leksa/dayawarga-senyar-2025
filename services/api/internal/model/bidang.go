package model

import (
	"time"

	"github.com/google/uuid"
)

// Bidang represents a field/domain of work for relawan
type Bidang struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Name        string    `json:"name" gorm:"not null"`
	Slug        string    `json:"slug" gorm:"uniqueIndex;not null"`
	Description *string   `json:"description,omitempty"`
	IsActive    bool      `json:"is_active" gorm:"column:is_active;default:true"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at"`

	// Relations (loaded when needed)
	Organizations []Organization `json:"organizations,omitempty" gorm:"many2many:organization_bidangs;foreignKey:ID;references:ID"`
}

func (Bidang) TableName() string {
	return "bidangs"
}

// OrganizationBidang represents the junction table between organizations and bidangs
type OrganizationBidang struct {
	OrganizationID uuid.UUID `json:"organization_id" gorm:"type:uuid;primaryKey"`
	BidangID       uuid.UUID `json:"bidang_id" gorm:"type:uuid;primaryKey"`
	CreatedAt      time.Time `json:"created_at" gorm:"column:created_at"`

	// Relations
	Organization *Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
	Bidang       *Bidang       `json:"bidang,omitempty" gorm:"foreignKey:BidangID"`
}

func (OrganizationBidang) TableName() string {
	return "organization_bidangs"
}
