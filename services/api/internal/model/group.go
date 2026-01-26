package model

import (
	"time"

	"github.com/google/uuid"
)

// Group represents a team within an organization
type Group struct {
	ID uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`

	OrganizationID uuid.UUID `json:"organization_id" gorm:"type:uuid;not null"`

	// Basic info
	Name        string  `json:"name" gorm:"not null"`
	Description *string `json:"description,omitempty"`

	// Group Leader (becomes ODK Project Manager on approval)
	LeaderID *uuid.UUID `json:"leader_id,omitempty" gorm:"type:uuid"`

	// ODK Integration
	ODKProjectID             *int  `json:"odk_project_id,omitempty" gorm:"column:odk_project_id"`
	ODKProjectManagerCreated bool  `json:"odk_project_manager_created" gorm:"column:odk_project_manager_created;default:false"`

	// Status
	IsActive bool `json:"is_active" gorm:"column:is_active;default:true"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" gorm:"column:deleted_at;index"`

	// Relations
	Organization *Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
	Leader       *User         `json:"leader,omitempty" gorm:"foreignKey:LeaderID"`
	Relawan      []Relawan     `json:"relawan,omitempty" gorm:"foreignKey:GroupID"`
}

// HasODKProject returns true if group has an assigned ODK project
func (g *Group) HasODKProject() bool {
	return g.ODKProjectID != nil && *g.ODKProjectID > 0
}

func (Group) TableName() string {
	return "groups"
}
