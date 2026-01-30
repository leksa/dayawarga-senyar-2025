package service

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/leksa/datamapper-senyar/internal/model"
	"github.com/leksa/datamapper-senyar/internal/odk"
	"github.com/leksa/datamapper-senyar/internal/repository"
)

type AuthentikWebhookService struct {
	userRepo  *repository.UserRepository
	odkClient *odk.Client
	db        *gorm.DB
}

func NewAuthentikWebhookService(
	userRepo *repository.UserRepository,
	odkClient *odk.Client,
	db *gorm.DB,
) *AuthentikWebhookService {
	return &AuthentikWebhookService{
		userRepo:  userRepo,
		odkClient: odkClient,
		db:        db,
	}
}

type AuthentikUserCreatedInput struct {
	AuthentikUserID int
	Username        string
	Email           string
	Name            string
	IsActive        bool
}

type AuthentikUserCreatedResult struct {
	UserID       string `json:"user_id"`
	Email        string `json:"email"`
	ODKWebUserID int    `json:"odk_web_user_id"`
	ODKStatus    string `json:"odk_status"`
	Message      string `json:"message"`
}

func (s *AuthentikWebhookService) HandleUserCreated(ctx context.Context, input AuthentikUserCreatedInput) (*AuthentikUserCreatedResult, error) {
	if !input.IsActive {
		return &AuthentikUserCreatedResult{
			Email:     input.Email,
			ODKStatus: "skipped",
			Message:   "User is not active, skipping ODK sync",
		}, nil
	}

	user, err := s.userRepo.FindByEmail(ctx, input.Email)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	if user == nil {
		return &AuthentikUserCreatedResult{
			Email:     input.Email,
			ODKStatus: "skipped",
			Message:   "User not found in local database, skipping ODK sync",
		}, nil
	}

	if user.ODKWebUserID != nil && *user.ODKWebUserID > 0 {
		return &AuthentikUserCreatedResult{
			UserID:       user.ID.String(),
			Email:        input.Email,
			ODKWebUserID: *user.ODKWebUserID,
			ODKStatus:    "exists",
			Message:      "User already has ODK Web User",
		}, nil
	}

	if s.odkClient == nil {
		return nil, fmt.Errorf("ODK client not configured")
	}

	webUser, err := s.odkClient.GetWebUserByEmail(input.Email)
	if err != nil {
		fmt.Printf("[Authentik Webhook] Warning: failed to check existing ODK user: %v\n", err)
	}

	if webUser == nil {
		displayName := input.Name
		if displayName == "" {
			displayName = input.Username
		}

		password := generateSecurePassword()
		webUser, err = s.odkClient.CreateWebUserWithPassword(input.Email, displayName, password)
		if err != nil {
			return nil, fmt.Errorf("failed to create ODK Web User: %w", err)
		}
		fmt.Printf("[Authentik Webhook] Created ODK Web User for %s with ID: %d\n", input.Email, webUser.ID)
	} else if webUser.LastLoginAt == nil {
		fmt.Printf("[Authentik Webhook] User %s exists but is pending, recreating with password\n", input.Email)
		if err := s.odkClient.DeleteWebUser(webUser.ID); err != nil {
			fmt.Printf("[Authentik Webhook] Warning: failed to delete pending user: %v\n", err)
		}

		displayName := input.Name
		if displayName == "" {
			displayName = input.Username
		}

		password := generateSecurePassword()
		webUser, err = s.odkClient.CreateWebUserWithPassword(input.Email, displayName, password)
		if err != nil {
			return nil, fmt.Errorf("failed to recreate ODK Web User: %w", err)
		}
		fmt.Printf("[Authentik Webhook] Recreated ODK Web User for %s with ID: %d\n", input.Email, webUser.ID)
	} else {
		fmt.Printf("[Authentik Webhook] ODK Web User already exists for %s with ID: %d\n", input.Email, webUser.ID)
	}

	now := time.Now()
	if err := s.db.Model(&model.User{}).
		Where("id = ?", user.ID).
		Updates(map[string]interface{}{
			"odk_web_user_id": webUser.ID,
			"updated_at":      now,
		}).Error; err != nil {
		return nil, fmt.Errorf("failed to update user with ODK Web User ID: %w", err)
	}

	return &AuthentikUserCreatedResult{
		UserID:       user.ID.String(),
		Email:        input.Email,
		ODKWebUserID: webUser.ID,
		ODKStatus:    "created",
		Message:      "ODK Web User created successfully. Assign to project to get App User + QR code.",
	}, nil
}
