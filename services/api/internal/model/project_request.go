package model

import (
	"time"

	"github.com/google/uuid"
)

// ProjectRequestStatus represents the status of a project request
type ProjectRequestStatus string

const (
	ProjectRequestStatusPending  ProjectRequestStatus = "pending"
	ProjectRequestStatusApproved ProjectRequestStatus = "approved"
	ProjectRequestStatusRejected ProjectRequestStatus = "rejected"
)

// ProjectRequest represents a request to assign an ODK project to a group
type ProjectRequest struct {
	ID uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`

	// References
	OrganizationID uuid.UUID `json:"organization_id" gorm:"type:uuid;not null"`
	GroupID        uuid.UUID `json:"group_id" gorm:"type:uuid;not null"`

	// ODK Project
	ODKProjectID   int     `json:"odk_project_id" gorm:"column:odk_project_id;not null"`
	ODKProjectName *string `json:"odk_project_name,omitempty" gorm:"column:odk_project_name"`

	// Request info
	RequestedBy  uuid.UUID `json:"requested_by" gorm:"type:uuid;not null"`
	RequestNotes *string   `json:"request_notes,omitempty" gorm:"column:request_notes"`

	// Status
	Status ProjectRequestStatus `json:"status" gorm:"not null;default:'pending'"`

	// Review info
	ReviewedBy   *uuid.UUID `json:"reviewed_by,omitempty" gorm:"type:uuid;column:reviewed_by"`
	ReviewedAt   *time.Time `json:"reviewed_at,omitempty" gorm:"column:reviewed_at"`
	ReviewNotes  *string    `json:"review_notes,omitempty" gorm:"column:review_notes"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`

	// Relations
	Organization *Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
	Group        *Group        `json:"group,omitempty" gorm:"foreignKey:GroupID"`
	Requester    *User         `json:"requester,omitempty" gorm:"foreignKey:RequestedBy"`
	Reviewer     *User         `json:"reviewer,omitempty" gorm:"foreignKey:ReviewedBy"`
}

func (ProjectRequest) TableName() string {
	return "project_requests"
}

// IsPending returns true if request is pending review
func (pr *ProjectRequest) IsPending() bool {
	return pr.Status == ProjectRequestStatusPending
}

// IsApproved returns true if request was approved
func (pr *ProjectRequest) IsApproved() bool {
	return pr.Status == ProjectRequestStatusApproved
}

// IsRejected returns true if request was rejected
func (pr *ProjectRequest) IsRejected() bool {
	return pr.Status == ProjectRequestStatusRejected
}

// GroupProject represents an approved ODK project assignment to a group
type GroupProject struct {
	ID uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`

	GroupID        uuid.UUID `json:"group_id" gorm:"type:uuid;not null"`
	ODKProjectID   int       `json:"odk_project_id" gorm:"column:odk_project_id;not null"`
	ODKProjectName *string   `json:"odk_project_name,omitempty" gorm:"column:odk_project_name"`

	// Reference to approval
	ApprovedRequestID *uuid.UUID `json:"approved_request_id,omitempty" gorm:"type:uuid;column:approved_request_id"`

	// Timestamps
	AssignedAt time.Time `json:"assigned_at" gorm:"column:assigned_at"`

	// Relations
	Group           *Group          `json:"group,omitempty" gorm:"foreignKey:GroupID"`
	ApprovedRequest *ProjectRequest `json:"approved_request,omitempty" gorm:"foreignKey:ApprovedRequestID"`
}

func (GroupProject) TableName() string {
	return "group_projects"
}
