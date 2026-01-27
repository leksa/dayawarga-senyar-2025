package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/leksa/datamapper-senyar/internal/config"
	"github.com/leksa/datamapper-senyar/internal/email"
	"github.com/leksa/datamapper-senyar/internal/model"
	"github.com/leksa/datamapper-senyar/internal/repository"
)

type InvitationService struct {
	userRepo *repository.UserRepository
	orgRepo  *repository.OrganizationRepository
	emailSvc *email.Service
	config   *config.Config
}

func NewInvitationService(
	userRepo *repository.UserRepository,
	orgRepo *repository.OrganizationRepository,
	emailSvc *email.Service,
	cfg *config.Config,
) *InvitationService {
	return &InvitationService{
		userRepo: userRepo,
		orgRepo:  orgRepo,
		emailSvc: emailSvc,
		config:   cfg,
	}
}

type InviteUserInput struct {
	Email          string
	Name           string
	OrganizationID *uuid.UUID
	OrgRole        model.OrgMemberRole
	InvitedBy      uuid.UUID
}

type InviteUserResult struct {
	User           *model.User
	InvitationLink string
	IsNewUser      bool
	EmailSent      bool
}

type SetPasswordResult struct {
	User       *model.User
	PIN        string
	PINExpires time.Time
}

type VerificationStatusResult struct {
	Verified   bool
	VerifiedAt *time.Time
}

func (s *InvitationService) InviteUser(ctx context.Context, input InviteUserInput) (*InviteUserResult, error) {
	var org *model.Organization
	if input.OrganizationID != nil {
		var err error
		org, err = s.orgRepo.FindByID(ctx, *input.OrganizationID)
		if err != nil {
			return nil, fmt.Errorf("organization not found")
		}
	}

	existingUser, err := s.userRepo.FindByEmail(ctx, input.Email)
	if err == nil && existingUser != nil {
		if existingUser.Status == model.UserStatusActive {
			return nil, fmt.Errorf("user with this email already exists and is active")
		}
		return s.doResendInvitation(ctx, existingUser, org, input.InvitedBy)
	}

	token, err := generateToken(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	now := time.Now()

	var name *string
	if input.Name != "" {
		name = &input.Name
	}

	tempSubject := fmt.Sprintf("pending_%s", token[:16])

	user := &model.User{
		ID:                  uuid.New(),
		OIDCSubject:         tempSubject,
		Email:               input.Email,
		Name:                name,
		Role:                model.UserRoleMember,
		Status:              model.UserStatusPendingInvitation,
		IsActive:            false,
		InvitationToken:     &token,
		InvitationExpiresAt: &expiresAt,
		InvitationSentAt:    &now,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	if input.OrganizationID != nil {
		role := input.OrgRole
		if role == "" {
			role = model.OrgMemberRoleMember
		}
		member := &model.OrganizationMember{
			ID:             uuid.New(),
			OrganizationID: *input.OrganizationID,
			UserID:         user.ID,
			Role:           role,
		}
		if err := s.orgRepo.AddMember(ctx, member); err != nil {
			return nil, fmt.Errorf("failed to add user to organization: %w", err)
		}
	}

	inviteLink := fmt.Sprintf("%s/invite/accept?token=%s", s.config.AppBaseURL, token)

	emailSent := false
	if s.emailSvc != nil && s.emailSvc.IsConfigured() && org != nil {
		inviter, _ := s.userRepo.FindByID(ctx, input.InvitedBy)
		inviterName := ""
		if inviter != nil && inviter.Name != nil {
			inviterName = *inviter.Name
		}

		roleName := "Anggota"
		if input.OrgRole == model.OrgMemberRoleAdmin {
			roleName = "Admin Organisasi"
		}

		err := s.emailSvc.SendInvitation(input.Email, email.InvitationData{
			RecipientName: input.Name,
			InviterName:   inviterName,
			OrgName:       org.Name,
			Role:          roleName,
			InviteLink:    inviteLink,
			ExpiresIn:     "7 hari",
		})
		if err != nil {
			fmt.Printf("Warning: failed to send invitation email: %v\n", err)
		} else {
			emailSent = true
		}
	}

	return &InviteUserResult{
		User:           user,
		InvitationLink: inviteLink,
		IsNewUser:      true,
		EmailSent:      emailSent,
	}, nil
}

func (s *InvitationService) doResendInvitation(ctx context.Context, user *model.User, org *model.Organization, inviterID uuid.UUID) (*InviteUserResult, error) {
	token, err := generateToken(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	now := time.Now()

	updates := map[string]interface{}{
		"invitation_token":      token,
		"invitation_expires_at": expiresAt,
		"invitation_sent_at":    now,
	}

	updatedUser, err := s.userRepo.Update(ctx, user.ID.String(), updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update invitation: %w", err)
	}

	inviteLink := fmt.Sprintf("%s/invite/accept?token=%s", s.config.AppBaseURL, token)

	emailSent := false
	if s.emailSvc != nil && s.emailSvc.IsConfigured() && org != nil {
		inviter, _ := s.userRepo.FindByID(ctx, inviterID)
		inviterName := ""
		if inviter != nil && inviter.Name != nil {
			inviterName = *inviter.Name
		}

		userName := ""
		if user.Name != nil {
			userName = *user.Name
		}

		roleName := "Anggota"
		if user.Role == model.UserRoleOrgAdmin {
			roleName = "Admin Organisasi"
		}

		err := s.emailSvc.SendInvitation(user.Email, email.InvitationData{
			RecipientName: userName,
			InviterName:   inviterName,
			OrgName:       org.Name,
			Role:          roleName,
			InviteLink:    inviteLink,
			ExpiresIn:     "7 hari",
		})
		if err != nil {
			fmt.Printf("Warning: failed to send invitation email: %v\n", err)
		} else {
			emailSent = true
		}
	}

	return &InviteUserResult{
		User:           updatedUser,
		InvitationLink: inviteLink,
		IsNewUser:      false,
		EmailSent:      emailSent,
	}, nil
}

func (s *InvitationService) ResendInvitation(ctx context.Context, userID uuid.UUID) (*InviteUserResult, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	if user.Status != model.UserStatusPendingInvitation {
		return nil, fmt.Errorf("user is not pending invitation")
	}

	userWithOrgs, err := s.userRepo.FindWithOrganizations(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user organizations: %w", err)
	}

	var org *model.Organization
	if len(userWithOrgs.OrganizationMemberships) > 0 {
		org = userWithOrgs.OrganizationMemberships[0].Organization
	}

	return s.doResendInvitation(ctx, user, org, uuid.Nil)
}

func (s *InvitationService) ValidateToken(ctx context.Context, token string) (*model.User, error) {
	user, err := s.userRepo.FindByInvitationToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired invitation token")
	}

	if user.IsInvitationExpired() {
		return nil, fmt.Errorf("invitation has expired")
	}

	return user, nil
}

func (s *InvitationService) SetPassword(ctx context.Context, token string, password string) (*SetPasswordResult, error) {
	user, err := s.ValidateToken(ctx, token)
	if err != nil {
		return nil, err
	}

	pin := generatePIN()
	pinExpires := time.Now().Add(15 * time.Minute)

	updates := map[string]interface{}{
		"status":                      model.UserStatusPendingVerification,
		"verification_pin":            pin,
		"verification_pin_expires_at": pinExpires,
		"invitation_token":            nil,
		"invitation_expires_at":       nil,
	}

	updatedUser, err := s.userRepo.Update(ctx, user.ID.String(), updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &SetPasswordResult{
		User:       updatedUser,
		PIN:        pin,
		PINExpires: pinExpires,
	}, nil
}

func (s *InvitationService) VerifyPIN(ctx context.Context, pin string, phone string) (*model.User, error) {
	user, err := s.userRepo.FindByVerificationPIN(ctx, pin)
	if err != nil {
		return nil, fmt.Errorf("invalid PIN")
	}

	if user.IsPINExpired() {
		return nil, fmt.Errorf("PIN has expired")
	}

	if user.Status != model.UserStatusPendingVerification {
		return nil, fmt.Errorf("user is not pending verification")
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":                      model.UserStatusActive,
		"is_active":                   true,
		"verification_pin":            nil,
		"verification_pin_expires_at": nil,
		"verification_phone":          phone,
		"verified_at":                 now,
	}

	updatedUser, err := s.userRepo.Update(ctx, user.ID.String(), updates)
	if err != nil {
		return nil, fmt.Errorf("failed to verify user: %w", err)
	}

	return updatedUser, nil
}

func (s *InvitationService) GetVerificationStatus(ctx context.Context, userID uuid.UUID) (*VerificationStatusResult, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	return &VerificationStatusResult{
		Verified:   user.Status == model.UserStatusActive && user.VerifiedAt != nil,
		VerifiedAt: user.VerifiedAt,
	}, nil
}

func (s *InvitationService) RegeneratePIN(ctx context.Context, userID uuid.UUID) (*SetPasswordResult, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	if user.Status != model.UserStatusPendingVerification {
		return nil, fmt.Errorf("user is not pending verification")
	}

	pin := generatePIN()
	pinExpires := time.Now().Add(15 * time.Minute)

	updates := map[string]interface{}{
		"verification_pin":            pin,
		"verification_pin_expires_at": pinExpires,
	}

	updatedUser, err := s.userRepo.Update(ctx, user.ID.String(), updates)
	if err != nil {
		return nil, fmt.Errorf("failed to regenerate PIN: %w", err)
	}

	return &SetPasswordResult{
		User:       updatedUser,
		PIN:        pin,
		PINExpires: pinExpires,
	}, nil
}

func (s *InvitationService) AcceptInvitation(ctx context.Context, token string, oidcSubject string) (*model.User, error) {
	user, err := s.ValidateToken(ctx, token)
	if err != nil {
		return nil, err
	}

	updates := map[string]interface{}{
		"oidc_subject":          oidcSubject,
		"status":                model.UserStatusActive,
		"is_active":             true,
		"invitation_token":      nil,
		"invitation_expires_at": nil,
	}

	updatedUser, err := s.userRepo.Update(ctx, user.ID.String(), updates)
	if err != nil {
		return nil, fmt.Errorf("failed to accept invitation: %w", err)
	}

	return updatedUser, nil
}

func (s *InvitationService) CancelInvitation(ctx context.Context, userID string) error {
	user, err := s.userRepo.FindByIDStr(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	if user.Status != model.UserStatusPendingInvitation {
		return fmt.Errorf("user is not pending invitation")
	}

	return s.userRepo.Delete(ctx, userID)
}

func generateToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func generatePIN() string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	result := make([]byte, 6)
	rand.Read(result)
	for i := range result {
		result[i] = chars[int(result[i])%len(chars)]
	}
	return strings.ToUpper(string(result))
}
