package model

import (
	"time"

	"github.com/google/uuid"
)

// UserRole represents the role of a user in the system
type UserRole string

const (
	UserRoleSuperAdmin UserRole = "super_admin"
	UserRoleOrgAdmin   UserRole = "org_admin"
	UserRoleMember     UserRole = "member"
)

// UserStatus represents the status of a user
type UserStatus string

const (
	UserStatusPendingInvitation   UserStatus = "pending_invitation"
	UserStatusPendingVerification UserStatus = "pending_verification"
	UserStatusActive              UserStatus = "active"
	UserStatusSuspended           UserStatus = "suspended"
)

// User represents an admin portal user (from Authentik OIDC)
type User struct {
	ID uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`

	// OIDC identity (from Authentik)
	OIDCSubject string  `json:"oidc_subject" gorm:"column:oidc_subject;uniqueIndex;not null"`
	Email       string  `json:"email" gorm:"uniqueIndex;not null"`
	Name        *string `json:"name,omitempty"`

	// Profile
	AvatarURL *string `json:"avatar_url,omitempty" gorm:"column:avatar_url"`

	// Role (super_admin, org_admin, member)
	Role UserRole `json:"role" gorm:"not null;default:'member'"`

	// ODK Integration - Web User (browser/enketo access)
	ODKWebUserID *int `json:"odk_web_user_id,omitempty" gorm:"column:odk_web_user_id"`

	// ODK Integration - App User (ODK Collect access)
	ODKAppUserID        *int    `json:"odk_app_user_id,omitempty" gorm:"column:odk_app_user_id"`
	ODKAppUserToken     *string `json:"-" gorm:"column:odk_app_user_token"` // Token for QR code, not exposed in JSON
	ODKAppUserProjectID *int    `json:"odk_app_user_project_id,omitempty" gorm:"column:odk_app_user_project_id"`

	// Status
	IsActive    bool       `json:"is_active" gorm:"column:is_active;default:true"`
	Status      UserStatus `json:"status" gorm:"column:status;default:'active'"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty" gorm:"column:last_login_at"`

	// Invitation fields
	InvitationToken     *string    `json:"-" gorm:"column:invitation_token"`
	InvitationExpiresAt *time.Time `json:"-" gorm:"column:invitation_expires_at"`
	InvitationSentAt    *time.Time `json:"invitation_sent_at,omitempty" gorm:"column:invitation_sent_at"`

	// WhatsApp PIN verification
	VerificationPIN          *string    `json:"-" gorm:"column:verification_pin"`
	VerificationPINExpiresAt *time.Time `json:"-" gorm:"column:verification_pin_expires_at"`
	VerificationPhone        *string    `json:"verification_phone,omitempty" gorm:"column:verification_phone"`
	VerifiedAt               *time.Time `json:"verified_at,omitempty" gorm:"column:verified_at"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`

	// Relations (loaded when needed)
	OrganizationMemberships []OrganizationMember `json:"organization_memberships,omitempty" gorm:"foreignKey:UserID"`
}

// HasODKAccess returns true if user has an ODK web user account
func (u *User) HasODKAccess() bool {
	return u.ODKWebUserID != nil && *u.ODKWebUserID > 0
}

// HasODKAppUser returns true if user has an ODK app user (for ODK Collect)
func (u *User) HasODKAppUser() bool {
	return u.ODKAppUserID != nil && *u.ODKAppUserID > 0
}

// HasODKQRCode returns true if user has QR code data available
func (u *User) HasODKQRCode() bool {
	return u.ODKAppUserToken != nil && *u.ODKAppUserToken != ""
}

func (User) TableName() string {
	return "users"
}

// IsSuperAdmin checks if user has super admin role
func (u *User) IsSuperAdmin() bool {
	return u.Role == UserRoleSuperAdmin
}

// IsOrgAdmin checks if user has org admin role
func (u *User) IsOrgAdmin() bool {
	return u.Role == UserRoleOrgAdmin
}

// IsPendingInvitation checks if user is pending invitation
func (u *User) IsPendingInvitation() bool {
	return u.Status == UserStatusPendingInvitation
}

func (u *User) IsInvitationExpired() bool {
	if u.InvitationExpiresAt == nil {
		return true
	}
	return time.Now().After(*u.InvitationExpiresAt)
}

func (u *User) IsPINExpired() bool {
	if u.VerificationPINExpiresAt == nil {
		return true
	}
	return time.Now().After(*u.VerificationPINExpiresAt)
}

func (u *User) IsPendingVerification() bool {
	return u.Status == UserStatusPendingVerification
}

func (u *User) IsVerified() bool {
	return u.VerifiedAt != nil
}
