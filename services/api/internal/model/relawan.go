package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// RelawanStatus represents the status of a relawan
type RelawanStatus string

const (
	RelawanStatusActive    RelawanStatus = "active"
	RelawanStatusInactive  RelawanStatus = "inactive"
	RelawanStatusSuspended RelawanStatus = "suspended"
)

// Relawan represents a field volunteer
type Relawan struct {
	ID uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`

	// Organization and group assignment
	OrganizationID uuid.UUID  `json:"organization_id" gorm:"type:uuid;not null"`
	GroupID        *uuid.UUID `json:"group_id,omitempty" gorm:"type:uuid"`

	// Basic info
	Name  string  `json:"name" gorm:"not null"`
	Phone *string `json:"phone,omitempty"`
	Email *string `json:"email,omitempty"`

	// ODK App User (created via ODK Central API)
	ODKAppUserID        *int       `json:"odk_app_user_id,omitempty" gorm:"column:odk_app_user_id"`
	ODKAppUserToken     *string    `json:"-" gorm:"column:odk_app_user_token"` // Hidden from JSON, encrypted
	ODKAppUserCreatedAt *time.Time `json:"odk_app_user_created_at,omitempty" gorm:"column:odk_app_user_created_at"`

	// Assigned forms (array of form IDs)
	AssignedForms pq.StringArray `json:"assigned_forms,omitempty" gorm:"type:text[];column:assigned_forms"`

	// WhatsApp verification
	WAVerified     bool       `json:"wa_verified" gorm:"column:wa_verified;default:false"`
	WAVerifiedAt   *time.Time `json:"wa_verified_at,omitempty" gorm:"column:wa_verified_at"`
	WALastActivity *time.Time `json:"wa_last_activity,omitempty" gorm:"column:wa_last_activity"`
	WASessionCount int        `json:"wa_session_count" gorm:"column:wa_session_count;default:0"`

	// Status
	Status RelawanStatus `json:"status" gorm:"default:'active'"`

	// Notes
	Notes *string `json:"notes,omitempty"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" gorm:"column:deleted_at;index"`

	// Relations
	Organization *Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
	Group        *Group        `json:"group,omitempty" gorm:"foreignKey:GroupID"`
}

func (Relawan) TableName() string {
	return "relawan"
}

// IsActive checks if relawan is active
func (r *Relawan) IsActive() bool {
	return r.Status == RelawanStatusActive
}

// HasODKAccess checks if relawan has ODK app user configured
func (r *Relawan) HasODKAccess() bool {
	return r.ODKAppUserID != nil && r.ODKAppUserToken != nil
}

// HasWAAccess checks if relawan is verified for WhatsApp chatbot
func (r *Relawan) HasWAAccess() bool {
	return r.WAVerified && r.Status == RelawanStatusActive
}
