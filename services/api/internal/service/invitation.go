package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/leksa/datamapper-senyar/internal/authentik"
	"github.com/leksa/datamapper-senyar/internal/model"
	"github.com/leksa/datamapper-senyar/internal/repository"
)

// InvitationService handles user invitation logic
type InvitationService struct {
	userRepo        *repository.UserRepository
	orgRepo         *repository.OrganizationRepository
	authentikClient *authentik.Client
	appBaseURL      string // Base URL for invitation links (e.g., "http://localhost:5173")
}

// NewInvitationService creates a new invitation service
func NewInvitationService(
	userRepo *repository.UserRepository,
	orgRepo *repository.OrganizationRepository,
	authentikClient *authentik.Client,
	appBaseURL string,
) *InvitationService {
	return &InvitationService{
		userRepo:        userRepo,
		orgRepo:         orgRepo,
		authentikClient: authentikClient,
		appBaseURL:      appBaseURL,
	}
}

// InviteUserInput represents input for inviting a user
type InviteUserInput struct {
	Email          string              `json:"email" binding:"required,email"`
	Name           string              `json:"name" binding:"required"`
	OrganizationID *uuid.UUID          `json:"organization_id,omitempty"`
	OrgRole        model.OrgMemberRole `json:"org_role,omitempty"`
	InvitedBy      uuid.UUID           `json:"invited_by"`
}

// InviteResult contains the result of an invitation
type InviteResult struct {
	User           *model.User `json:"user"`
	InvitationLink string      `json:"invitation_link,omitempty"`
	IsNewUser      bool        `json:"is_new_user"`
}

// InviteUser invites a user to the platform and optionally to an organization
func (s *InvitationService) InviteUser(ctx context.Context, input InviteUserInput) (*InviteResult, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.FindByEmail(ctx, input.Email)
	if err != nil && err.Error() != "record not found" {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	if existingUser != nil {
		// User exists - just add to organization if specified
		if input.OrganizationID != nil {
			if err := s.addUserToOrg(ctx, existingUser.ID, *input.OrganizationID, input.OrgRole); err != nil {
				return nil, err
			}
		}
		return &InviteResult{
			User:      existingUser,
			IsNewUser: false,
		}, nil
	}

	// Create new user in Authentik first
	authentikUser, err := s.authentikClient.CreateUser(authentik.CreateUserInput{
		Username: input.Email, // Use email as username
		Name:     input.Name,
		Email:    input.Email,
		IsActive: true,
		Path:     "users",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user in Authentik: %w", err)
	}

	// Get recovery link for password setup
	recoveryLink, err := s.authentikClient.GetRecoveryLink(authentikUser.PK)
	if err != nil {
		log.Printf("Warning: failed to get recovery link: %v", err)
		// Don't fail the whole operation, user can request password reset later
	}

	// Fix recovery link domain (Docker internal URL -> public URL)
	if recoveryLink != "" {
		recoveryLink = strings.Replace(recoveryLink, "host.docker.internal", "localhost", 1)
	}

	// Generate invitation token
	invitationToken, err := generateSecureToken(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate invitation token: %w", err)
	}

	// Set expiration (7 days from now)
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	sentAt := time.Now()

	// Determine user role based on org membership role
	// If invited as org admin, set system role to org_admin
	userRole := model.UserRoleMember
	if input.OrgRole == model.OrgMemberRoleAdmin {
		userRole = model.UserRoleOrgAdmin
	}

	// Create user in our database
	user := &model.User{
		OIDCSubject:         authentikUser.UID,
		Email:               input.Email,
		Name:                &input.Name,
		Role:                userRole,
		Status:              model.UserStatusPendingInvitation,
		IsActive:            true,
		InvitationToken:     &invitationToken,
		InvitationExpiresAt: &expiresAt,
		InvitationSentAt:    &sentAt,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		// Rollback: delete user from Authentik
		_ = s.authentikClient.DeleteUser(authentikUser.PK)
		return nil, fmt.Errorf("failed to create user in database: %w", err)
	}

	// Add to organization if specified
	if input.OrganizationID != nil {
		if err := s.addUserToOrg(ctx, user.ID, *input.OrganizationID, input.OrgRole); err != nil {
			log.Printf("Warning: failed to add user to organization: %v", err)
		}
	}

	return &InviteResult{
		User:           user,
		InvitationLink: recoveryLink,
		IsNewUser:      true,
	}, nil
}

// ActivateUser activates a user after they set their password
func (s *InvitationService) ActivateUser(ctx context.Context, userID uuid.UUID) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if user.Status != model.UserStatusPendingInvitation {
		return errors.New("user is not pending invitation")
	}

	// Clear invitation fields and set status to active
	user.Status = model.UserStatusActive
	user.InvitationToken = nil
	user.InvitationExpiresAt = nil

	return s.userRepo.Update(ctx, user)
}

// ResendInvitation resends an invitation email to a pending user
func (s *InvitationService) ResendInvitation(ctx context.Context, userID uuid.UUID) (*InviteResult, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if user.Status != model.UserStatusPendingInvitation {
		return nil, errors.New("user is not pending invitation")
	}

	// Get Authentik user by email
	authentikUser, err := s.authentikClient.GetUserByEmail(user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to get Authentik user: %w", err)
	}
	if authentikUser == nil {
		return nil, errors.New("user not found in Authentik")
	}

	// Get new recovery link
	recoveryLink, err := s.authentikClient.GetRecoveryLink(authentikUser.PK)
	if err != nil {
		return nil, fmt.Errorf("failed to get recovery link: %w", err)
	}

	// Fix recovery link domain (Docker internal URL -> public URL)
	if recoveryLink != "" {
		recoveryLink = strings.Replace(recoveryLink, "host.docker.internal", "localhost", 1)
	}

	// Update invitation timestamps
	newToken, _ := generateSecureToken(32)
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	sentAt := time.Now()

	user.InvitationToken = &newToken
	user.InvitationExpiresAt = &expiresAt
	user.InvitationSentAt = &sentAt

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &InviteResult{
		User:           user,
		InvitationLink: recoveryLink,
		IsNewUser:      false,
	}, nil
}

// addUserToOrg adds a user to an organization with the specified role
func (s *InvitationService) addUserToOrg(ctx context.Context, userID, orgID uuid.UUID, role model.OrgMemberRole) error {
	if role == "" {
		role = model.OrgMemberRoleAdmin // Default to admin for org leader invite
	}

	member := &model.OrganizationMember{
		OrganizationID: orgID,
		UserID:         userID,
		Role:           role,
	}

	return s.orgRepo.AddMember(ctx, member)
}

// generateSecureToken generates a cryptographically secure random token
func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
