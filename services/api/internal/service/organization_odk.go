package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/leksa/datamapper-senyar/internal/model"
	"github.com/leksa/datamapper-senyar/internal/odk"
	"github.com/leksa/datamapper-senyar/internal/repository"
)

// OrganizationODKService handles ODK-related organization operations
type OrganizationODKService struct {
	orgRepo   *repository.OrganizationRepository
	userRepo  *repository.UserRepository
	odkClient *odk.Client
	db        *gorm.DB
}

// NewOrganizationODKService creates a new organization ODK service
func NewOrganizationODKService(
	orgRepo *repository.OrganizationRepository,
	userRepo *repository.UserRepository,
	odkClient *odk.Client,
	db *gorm.DB,
) *OrganizationODKService {
	return &OrganizationODKService{
		orgRepo:   orgRepo,
		userRepo:  userRepo,
		odkClient: odkClient,
		db:        db,
	}
}

// AssignODKProjectInput represents input for assigning an ODK project to an organization
type AssignODKProjectInput struct {
	OrganizationID uuid.UUID
	ODKProjectID   int
}

// AssignODKProjectResult represents the result of assigning an ODK project
type AssignODKProjectResult struct {
	Organization    *model.Organization `json:"organization"`
	ODKProjectID    int                 `json:"odk_project_id"`
	ODKProjectName  string              `json:"odk_project_name"`
	AdminsProcessed []AdminODKResult    `json:"admins_processed"`
}

// AdminODKResult represents the ODK setup result for an org admin
type AdminODKResult struct {
	UserID       uuid.UUID `json:"user_id"`
	UserEmail    string    `json:"user_email"`
	UserName     string    `json:"user_name"`
	ODKWebUserID int       `json:"odk_web_user_id"`
	ODKAppUserID int       `json:"odk_app_user_id"`
	HasQRCode    bool      `json:"has_qr_code"`
	Error        string    `json:"error,omitempty"`
}

// AssignODKProject assigns an ODK project to an organization
// This performs the following atomic operations:
// 1. Validates organization exists and admin(s) are active
// 2. Creates/gets ODK Web User for each org admin (with password - immediately active)
// 3. Assigns Project Manager role to each admin
// 4. Creates App User for each admin (ODK Collect access)
// 5. Updates organization with ODK project ID
func (s *OrganizationODKService) AssignODKProject(ctx context.Context, input AssignODKProjectInput) (*AssignODKProjectResult, error) {
	if s.odkClient == nil {
		return nil, fmt.Errorf("ODK client not configured")
	}

	// 1. Get organization
	org, err := s.orgRepo.FindByID(ctx, input.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	// Check if organization already has an ODK project
	if org.ODKProjectID != nil && *org.ODKProjectID > 0 {
		return nil, fmt.Errorf("organization already assigned to ODK project %d", *org.ODKProjectID)
	}

	// 2. Get organization admins
	admins, err := s.orgRepo.GetOrgAdmins(ctx, input.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization admins: %w", err)
	}

	if len(admins) == 0 {
		return nil, fmt.Errorf("organization has no admin users")
	}

	// 3. Verify ODK project exists
	project, err := s.odkClient.GetProject(input.ODKProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ODK project: %w", err)
	}
	if project == nil {
		return nil, fmt.Errorf("ODK project %d not found", input.ODKProjectID)
	}

	// 4. Process each admin
	var adminResults []AdminODKResult
	var hasSuccessfulAdmin bool

	for _, adminMember := range admins {
		if adminMember.User == nil {
			continue
		}

		user := adminMember.User
		result := AdminODKResult{
			UserID:    user.ID,
			UserEmail: user.Email,
		}

		if user.Name != nil {
			result.UserName = *user.Name
		} else {
			result.UserName = user.Email
		}

		// Check if user is active
		if user.Status != model.UserStatusActive {
			result.Error = fmt.Sprintf("user status is %s, must be active", user.Status)
			adminResults = append(adminResults, result)
			continue
		}

		// Process this admin
		adminResult, err := s.processAdminForODK(ctx, user, input.ODKProjectID)
		if err != nil {
			result.Error = err.Error()
			adminResults = append(adminResults, result)
			continue
		}

		result.ODKWebUserID = adminResult.WebUserID
		result.ODKAppUserID = adminResult.AppUserID
		result.HasQRCode = adminResult.HasQRCode
		adminResults = append(adminResults, result)
		hasSuccessfulAdmin = true
	}

	// 5. Check if at least one admin was processed successfully
	if !hasSuccessfulAdmin {
		return nil, fmt.Errorf("failed to process any organization admin for ODK")
	}

	// 6. Update organization with ODK project ID
	org.ODKProjectID = &input.ODKProjectID
	if err := s.orgRepo.Update(ctx, org); err != nil {
		return nil, fmt.Errorf("failed to update organization: %w", err)
	}

	return &AssignODKProjectResult{
		Organization:    org,
		ODKProjectID:    input.ODKProjectID,
		ODKProjectName:  project.Name,
		AdminsProcessed: adminResults,
	}, nil
}

// adminODKResult holds the result of processing an admin for ODK
type adminODKResult struct {
	WebUserID int
	AppUserID int
	HasQRCode bool
}

// processAdminForODK handles all ODK operations for a single admin user
func (s *OrganizationODKService) processAdminForODK(ctx context.Context, user *model.User, projectID int) (*adminODKResult, error) {
	displayName := user.Email
	if user.Name != nil && *user.Name != "" {
		displayName = *user.Name
	}

	var webUser *odk.WebUser
	var err error

	// Step 1: Get or create ODK Web User
	// Always verify user exists in ODK Central by email lookup
	webUser, err = s.odkClient.GetWebUserByEmail(user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing ODK user: %w", err)
	}

	if webUser == nil {
		// Create new ODK user with password (immediately active)
		password := generateSecurePassword()
		webUser, err = s.odkClient.CreateWebUserWithPassword(user.Email, displayName, password)
		if err != nil {
			return nil, fmt.Errorf("failed to create ODK user: %w", err)
		}
		fmt.Printf("[ODK] Created new ODK Web User for %s with ID: %d\n", user.Email, webUser.ID)
	} else if webUser.LastLoginAt == nil {
		// User exists but never logged in (pending invitation)
		// Delete and recreate with password to force-activate
		fmt.Printf("[ODK] User %s (ID: %d) is pending, recreating with password\n", user.Email, webUser.ID)
		if err := s.odkClient.DeleteWebUser(webUser.ID); err != nil {
			return nil, fmt.Errorf("failed to delete pending ODK user: %w", err)
		}
		password := generateSecurePassword()
		webUser, err = s.odkClient.CreateWebUserWithPassword(user.Email, displayName, password)
		if err != nil {
			return nil, fmt.Errorf("failed to recreate ODK user: %w", err)
		}
		fmt.Printf("[ODK] Recreated ODK Web User for %s with ID: %d (now active)\n", user.Email, webUser.ID)
	} else {
		fmt.Printf("[ODK] Found existing active ODK Web User for %s with ID: %d\n", user.Email, webUser.ID)
	}

	// Update user with ODK web user ID (in case it changed or was stale)
	if user.ODKWebUserID == nil || *user.ODKWebUserID != webUser.ID {
		if err := s.db.Model(&model.User{}).
			Where("id = ?", user.ID).
			Update("odk_web_user_id", webUser.ID).Error; err != nil {
			return nil, fmt.Errorf("failed to update user ODK ID: %w", err)
		}
		user.ODKWebUserID = &webUser.ID
	}

	// Step 2: Assign Project Manager role
	fmt.Printf("[ODK] Assigning Project Manager role to user %d on project %d\n", webUser.ID, projectID)
	if err := s.odkClient.AssignProjectRole(projectID, webUser.ID, odk.RoleProjectManager); err != nil {
		return nil, fmt.Errorf("failed to assign project manager role: %w", err)
	}
	fmt.Printf("[ODK] Successfully assigned Project Manager role\n")

	// Step 3: Create App User for ODK Collect access
	var appUserID int
	var hasQRCode bool

	if user.ODKAppUserID != nil && *user.ODKAppUserID > 0 &&
		user.ODKAppUserProjectID != nil && *user.ODKAppUserProjectID == projectID {
		// User already has App User for this project
		appUserID = *user.ODKAppUserID
		hasQRCode = user.ODKAppUserToken != nil && *user.ODKAppUserToken != ""
		fmt.Printf("[ODK] User %s already has App User ID: %d for project %d\n", user.Email, appUserID, projectID)
	} else {
		// Create new App User
		appUser, err := s.odkClient.CreateAppUser(projectID, displayName)
		if err != nil {
			return nil, fmt.Errorf("failed to create App User: %w", err)
		}
		appUserID = appUser.ID
		hasQRCode = appUser.Token != ""
		fmt.Printf("[ODK] Created App User for %s with ID: %d\n", user.Email, appUserID)

		// Update user with App User details
		updates := map[string]interface{}{
			"odk_app_user_id":         appUser.ID,
			"odk_app_user_project_id": projectID,
		}
		if appUser.Token != "" {
			updates["odk_app_user_token"] = appUser.Token
		}

		if err := s.db.Model(&model.User{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update user with App User details: %w", err)
		}
	}

	return &adminODKResult{
		WebUserID: webUser.ID,
		AppUserID: appUserID,
		HasQRCode: hasQRCode,
	}, nil
}

// generateSecurePassword generates a secure random password
func generateSecurePassword() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// RemoveODKProjectInput represents input for removing an ODK project from an organization
type RemoveODKProjectInput struct {
	OrganizationID uuid.UUID
}

// RemoveODKProject removes the ODK project assignment from an organization
// Note: This does NOT delete the ODK project or remove users from ODK Central
func (s *OrganizationODKService) RemoveODKProject(ctx context.Context, input RemoveODKProjectInput) error {
	org, err := s.orgRepo.FindByID(ctx, input.OrganizationID)
	if err != nil {
		return fmt.Errorf("organization not found: %w", err)
	}

	if org.ODKProjectID == nil || *org.ODKProjectID == 0 {
		return fmt.Errorf("organization is not assigned to any ODK project")
	}

	// Clear ODK project ID
	org.ODKProjectID = nil
	if err := s.orgRepo.Update(ctx, org); err != nil {
		return fmt.Errorf("failed to update organization: %w", err)
	}

	return nil
}

// GetOrganizationODKInfo gets ODK-related information for an organization
type OrganizationODKInfo struct {
	Organization   *model.Organization        `json:"organization"`
	ODKProject     *odk.Project               `json:"odk_project,omitempty"`
	AdminODKStatus []AdminODKStatus           `json:"admin_odk_status"`
}

// AdminODKStatus represents an admin's ODK status
type AdminODKStatus struct {
	UserID       uuid.UUID `json:"user_id"`
	UserEmail    string    `json:"user_email"`
	UserName     string    `json:"user_name"`
	UserStatus   string    `json:"user_status"`
	HasWebUser   bool      `json:"has_web_user"`
	HasAppUser   bool      `json:"has_app_user"`
	HasQRCode    bool      `json:"has_qr_code"`
	ODKWebUserID *int      `json:"odk_web_user_id,omitempty"`
	ODKAppUserID *int      `json:"odk_app_user_id,omitempty"`
}

// GetOrganizationODKInfo gets ODK information for an organization
func (s *OrganizationODKService) GetOrganizationODKInfo(ctx context.Context, orgID uuid.UUID) (*OrganizationODKInfo, error) {
	org, err := s.orgRepo.FindByID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	info := &OrganizationODKInfo{
		Organization:   org,
		AdminODKStatus: []AdminODKStatus{},
	}

	// Get ODK project info if assigned
	if org.ODKProjectID != nil && *org.ODKProjectID > 0 && s.odkClient != nil {
		project, err := s.odkClient.GetProject(*org.ODKProjectID)
		if err == nil {
			info.ODKProject = project
		}
	}

	// Get admin statuses
	admins, err := s.orgRepo.GetOrgAdmins(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization admins: %w", err)
	}

	for _, adminMember := range admins {
		if adminMember.User == nil {
			continue
		}

		user := adminMember.User
		status := AdminODKStatus{
			UserID:     user.ID,
			UserEmail:  user.Email,
			UserStatus: string(user.Status),
			HasWebUser: user.ODKWebUserID != nil && *user.ODKWebUserID > 0,
			HasAppUser: user.ODKAppUserID != nil && *user.ODKAppUserID > 0,
			HasQRCode:  user.ODKAppUserToken != nil && *user.ODKAppUserToken != "",
		}

		if user.Name != nil {
			status.UserName = *user.Name
		} else {
			status.UserName = user.Email
		}

		if user.ODKWebUserID != nil {
			status.ODKWebUserID = user.ODKWebUserID
		}
		if user.ODKAppUserID != nil {
			status.ODKAppUserID = user.ODKAppUserID
		}

		info.AdminODKStatus = append(info.AdminODKStatus, status)
	}

	return info, nil
}
