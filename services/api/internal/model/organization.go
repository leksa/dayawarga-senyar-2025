package model

import (
	"time"

	"github.com/google/uuid"
)

// Organization represents an NGO or institution managing relawan
type Organization struct {
	ID uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`

	// Basic info
	Name        string  `json:"name" gorm:"not null"`
	Slug        string  `json:"slug" gorm:"uniqueIndex;not null"`
	Description *string `json:"description,omitempty"`

	// Contact
	Email   *string `json:"email,omitempty"`
	Phone   *string `json:"phone,omitempty"`
	Address *string `json:"address,omitempty"`

	// Visual
	LogoURL *string `json:"logo_url,omitempty" gorm:"column:logo_url"`

	// ODK Central integration
	ODKProjectID *int `json:"odk_project_id,omitempty" gorm:"column:odk_project_id"`

	// Status
	IsActive bool `json:"is_active" gorm:"column:is_active;default:true"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" gorm:"column:deleted_at;index"`

	// Relations (loaded when needed)
	Members []OrganizationMember `json:"members,omitempty" gorm:"foreignKey:OrganizationID"`
	Groups  []Group              `json:"groups,omitempty" gorm:"foreignKey:OrganizationID"`
	Relawan []Relawan            `json:"relawan,omitempty" gorm:"foreignKey:OrganizationID"`
}

func (Organization) TableName() string {
	return "organizations"
}

// OrgMemberRole represents the role of a user within an organization
type OrgMemberRole string

const (
	OrgMemberRoleAdmin  OrgMemberRole = "admin"
	OrgMemberRoleMember OrgMemberRole = "member"
)

// OrganizationMember represents user membership in an organization
type OrganizationMember struct {
	ID uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`

	OrganizationID uuid.UUID `json:"organization_id" gorm:"type:uuid;not null"`
	UserID         uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`

	// Role within organization (admin, member)
	Role OrgMemberRole `json:"role" gorm:"not null;default:'member'"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`

	// Relations
	Organization *Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
	User         *User         `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (OrganizationMember) TableName() string {
	return "organization_members"
}

// IsOrgAdmin checks if this member is an admin of the organization
func (m *OrganizationMember) IsOrgAdmin() bool {
	return m.Role == OrgMemberRoleAdmin
}
