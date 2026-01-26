package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/leksa/datamapper-senyar/internal/model"
	"github.com/leksa/datamapper-senyar/internal/odk"
	"github.com/leksa/datamapper-senyar/internal/repository"
)

// UserODKService handles ODK-related user operations
type UserODKService struct {
	userRepo  *repository.UserRepository
	odkClient *odk.Client
	db        *gorm.DB
}

// NewUserODKService creates a new user ODK service
func NewUserODKService(
	userRepo *repository.UserRepository,
	odkClient *odk.Client,
	db *gorm.DB,
) *UserODKService {
	return &UserODKService{
		userRepo:  userRepo,
		odkClient: odkClient,
		db:        db,
	}
}

// AssignProjectRoleInput represents input for assigning a user to an ODK project
type AssignProjectRoleInput struct {
	UserID       uuid.UUID
	ODKProjectID int
	RoleID       int // odk.RoleProjectManager (5) or odk.RoleProjectViewer (6)
}

// AssignProjectRoleResult represents the result of assigning a project role
type AssignProjectRoleResult struct {
	User          *model.User `json:"user"`
	ODKWebUserID  int         `json:"odk_web_user_id"`
	ProjectID     int         `json:"project_id"`
	RoleID        int         `json:"role_id"`
	RoleName      string      `json:"role_name"`
	ODKAppUserID  *int        `json:"odk_app_user_id,omitempty"`  // App User ID if created
	HasQRCode     bool        `json:"has_qr_code"`                // True if QR code is available
}

// AssignProjectRole assigns a user to an ODK project with a specific role
// Creates the ODK web user if it doesn't exist
func (s *UserODKService) AssignProjectRole(ctx context.Context, input AssignProjectRoleInput) (*AssignProjectRoleResult, error) {
	if s.odkClient == nil {
		return nil, fmt.Errorf("ODK client not configured")
	}

	// Validate role
	if input.RoleID != odk.RoleProjectManager && input.RoleID != odk.RoleProjectViewer {
		return nil, fmt.Errorf("invalid role ID: must be %d (Project Manager) or %d (Project Viewer)",
			odk.RoleProjectManager, odk.RoleProjectViewer)
	}

	// Get the user
	user, err := s.userRepo.FindByID(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Verify project exists in ODK
	project, err := s.odkClient.GetProject(input.ODKProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ODK project: %w", err)
	}
	if project == nil {
		return nil, fmt.Errorf("ODK project %d not found", input.ODKProjectID)
	}

	// Get or create ODK web user
	displayName := user.Email
	if user.Name != nil && *user.Name != "" {
		displayName = *user.Name
	}

	var webUser *odk.WebUser

	// Check if user already has ODK web user ID
	if user.ODKWebUserID != nil && *user.ODKWebUserID > 0 {
		// User already has ODK account, just assign the role
		webUser = &odk.WebUser{ID: *user.ODKWebUserID}
	} else {
		// Need to get or create ODK user
		webUser, err = s.odkClient.GetWebUserByEmail(user.Email)
		if err != nil {
			return nil, fmt.Errorf("failed to check existing ODK user: %w", err)
		}

		if webUser == nil {
			// Create new ODK user
			webUser, err = s.odkClient.CreateWebUser(user.Email, displayName)
			if err != nil {
				return nil, fmt.Errorf("failed to create ODK user: %w", err)
			}
		}

		// Update user with ODK web user ID
		if err := s.db.Model(&model.User{}).
			Where("id = ?", user.ID).
			Update("odk_web_user_id", webUser.ID).Error; err != nil {
			return nil, fmt.Errorf("failed to update user ODK ID: %w", err)
		}
		user.ODKWebUserID = &webUser.ID
	}

	// Assign role to project
	fmt.Printf("Assigning role %d to user %d on project %d\n", input.RoleID, webUser.ID, input.ODKProjectID)
	if err := s.odkClient.AssignProjectRole(input.ODKProjectID, webUser.ID, input.RoleID); err != nil {
		fmt.Printf("Failed to assign project role: %v\n", err)
		return nil, fmt.Errorf("failed to assign project role: %w", err)
	}
	fmt.Printf("Successfully assigned role to user\n")

	roleName := "Project Viewer"
	if input.RoleID == odk.RoleProjectManager {
		roleName = "Project Manager"
	}

	result := &AssignProjectRoleResult{
		User:         user,
		ODKWebUserID: webUser.ID,
		ProjectID:    input.ODKProjectID,
		RoleID:       input.RoleID,
		RoleName:     roleName,
		HasQRCode:    false,
	}

	// For Project Manager role, also create App User for ODK Collect access
	if input.RoleID == odk.RoleProjectManager {
		appUserResult, err := s.createOrGetAppUser(ctx, user, input.ODKProjectID, displayName)
		if err != nil {
			// Log error but don't fail the main operation
			// App User can be created later if needed
			fmt.Printf("Warning: failed to create App User for user %s: %v\n", user.Email, err)
		} else if appUserResult != nil {
			result.ODKAppUserID = &appUserResult.AppUserID
			result.HasQRCode = appUserResult.HasToken
			// Reload user to get updated fields
			updatedUser, _ := s.userRepo.FindByID(ctx, user.ID)
			if updatedUser != nil {
				result.User = updatedUser
			}
		}
	}

	return result, nil
}

// appUserResult represents the result of creating/getting an app user
type appUserResult struct {
	AppUserID int
	HasToken  bool
}

// createOrGetAppUser creates or retrieves an App User for ODK Collect access
func (s *UserODKService) createOrGetAppUser(ctx context.Context, user *model.User, projectID int, displayName string) (*appUserResult, error) {
	// Check if user already has an App User for this project
	if user.ODKAppUserID != nil && *user.ODKAppUserID > 0 &&
		user.ODKAppUserProjectID != nil && *user.ODKAppUserProjectID == projectID {
		// User already has App User for this project
		return &appUserResult{
			AppUserID: *user.ODKAppUserID,
			HasToken:  user.ODKAppUserToken != nil && *user.ODKAppUserToken != "",
		}, nil
	}

	// Check if user has App User for a different project - we need to handle this
	if user.ODKAppUserID != nil && *user.ODKAppUserID > 0 &&
		user.ODKAppUserProjectID != nil && *user.ODKAppUserProjectID != projectID {
		// User already has App User for a different project
		// For now, we keep the existing one (App Users are per-project in ODK)
		// In the future, we might want to support multiple project App Users
		return &appUserResult{
			AppUserID: *user.ODKAppUserID,
			HasToken:  user.ODKAppUserToken != nil && *user.ODKAppUserToken != "",
		}, nil
	}

	// Create new App User
	appUser, err := s.odkClient.CreateAppUser(projectID, displayName)
	if err != nil {
		return nil, fmt.Errorf("failed to create App User: %w", err)
	}

	// Update user with App User details
	updates := map[string]interface{}{
		"odk_app_user_id":         appUser.ID,
		"odk_app_user_project_id": projectID,
	}

	// Store token if available (token is only returned on creation)
	if appUser.Token != "" {
		updates["odk_app_user_token"] = appUser.Token
	}

	if err := s.db.Model(&model.User{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update user with App User details: %w", err)
	}

	return &appUserResult{
		AppUserID: appUser.ID,
		HasToken:  appUser.Token != "",
	}, nil
}

// RemoveProjectRoleInput represents input for removing a user from an ODK project
type RemoveProjectRoleInput struct {
	UserID       uuid.UUID
	ODKProjectID int
	RoleID       int
}

// RemoveProjectRole removes a user's role from an ODK project
func (s *UserODKService) RemoveProjectRole(ctx context.Context, input RemoveProjectRoleInput) error {
	if s.odkClient == nil {
		return fmt.Errorf("ODK client not configured")
	}

	// Get the user
	user, err := s.userRepo.FindByID(ctx, input.UserID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Check if user has ODK web user ID
	if user.ODKWebUserID == nil || *user.ODKWebUserID == 0 {
		return fmt.Errorf("user does not have an ODK account")
	}

	// Remove role from project
	if err := s.odkClient.RemoveProjectRole(input.ODKProjectID, *user.ODKWebUserID, input.RoleID); err != nil {
		return fmt.Errorf("failed to remove project role: %w", err)
	}

	return nil
}

// UserProjectAssignment represents a user's assignment to an ODK project
type UserProjectAssignment struct {
	ProjectID   int    `json:"project_id"`
	ProjectName string `json:"project_name"`
	RoleID      int    `json:"role_id"`
	RoleName    string `json:"role_name"`
}

// GetUserProjectAssignments gets all ODK project assignments for a user
func (s *UserODKService) GetUserProjectAssignments(ctx context.Context, userID uuid.UUID) ([]UserProjectAssignment, error) {
	if s.odkClient == nil {
		return nil, fmt.Errorf("ODK client not configured")
	}

	// Get the user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Check if user has ODK web user ID
	if user.ODKWebUserID == nil || *user.ODKWebUserID == 0 {
		return []UserProjectAssignment{}, nil
	}

	// Get all projects
	projects, err := s.odkClient.ListProjects()
	if err != nil {
		return nil, fmt.Errorf("failed to list ODK projects: %w", err)
	}

	var assignments []UserProjectAssignment

	// Check each project for assignments
	for _, project := range projects {
		projectAssignments, err := s.odkClient.ListProjectAssignments(project.ID)
		if err != nil {
			continue // Skip projects we can't access
		}

		for _, assignment := range projectAssignments {
			if assignment.ActorID == *user.ODKWebUserID {
				roleName := getRoleName(assignment.RoleID)
				assignments = append(assignments, UserProjectAssignment{
					ProjectID:   project.ID,
					ProjectName: project.Name,
					RoleID:      assignment.RoleID,
					RoleName:    roleName,
				})
			}
		}
	}

	return assignments, nil
}

// ProjectAssignmentInfo represents full info about project assignments
type ProjectAssignmentInfo struct {
	Project     *odk.Project                `json:"project"`
	Assignments []ProjectUserAssignmentInfo `json:"assignments"`
}

// ProjectUserAssignmentInfo represents a user's assignment to a project
type ProjectUserAssignmentInfo struct {
	UserID       *uuid.UUID `json:"user_id,omitempty"`       // nil if user not in portal
	ODKWebUserID int        `json:"odk_web_user_id"`
	Email        string     `json:"email"`
	DisplayName  string     `json:"display_name"`
	RoleID       int        `json:"role_id"`
	RoleName     string     `json:"role_name"`
}

// GetProjectAssignments gets all user assignments for an ODK project
func (s *UserODKService) GetProjectAssignments(ctx context.Context, odkProjectID int) (*ProjectAssignmentInfo, error) {
	if s.odkClient == nil {
		return nil, fmt.Errorf("ODK client not configured")
	}

	// Get project info
	project, err := s.odkClient.GetProject(odkProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ODK project: %w", err)
	}

	// Get project assignments
	assignments, err := s.odkClient.ListProjectAssignments(odkProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list project assignments: %w", err)
	}

	// Get all ODK web users to map IDs to emails
	webUsers, err := s.odkClient.ListWebUsers()
	if err != nil {
		return nil, fmt.Errorf("failed to list ODK web users: %w", err)
	}

	// Create map for quick lookup
	webUserMap := make(map[int]*odk.WebUser)
	for i := range webUsers {
		webUserMap[webUsers[i].ID] = &webUsers[i]
	}

	var assignmentInfos []ProjectUserAssignmentInfo

	for _, assignment := range assignments {
		webUser := webUserMap[assignment.ActorID]
		if webUser == nil {
			continue // Skip if web user not found
		}

		info := ProjectUserAssignmentInfo{
			ODKWebUserID: assignment.ActorID,
			Email:        webUser.Email,
			DisplayName:  webUser.DisplayName,
			RoleID:       assignment.RoleID,
			RoleName:     getRoleName(assignment.RoleID),
		}

		// Try to find matching portal user
		portalUser, _ := s.userRepo.FindByEmail(ctx, webUser.Email)
		if portalUser != nil {
			info.UserID = &portalUser.ID
		}

		assignmentInfos = append(assignmentInfos, info)
	}

	return &ProjectAssignmentInfo{
		Project:     project,
		Assignments: assignmentInfos,
	}, nil
}

func getRoleName(roleID int) string {
	switch roleID {
	case odk.RoleAdmin:
		return "Site Administrator"
	case odk.RoleProjectManager:
		return "Project Manager"
	case odk.RoleProjectViewer:
		return "Project Viewer"
	case odk.RoleDataCollector:
		return "Data Collector"
	default:
		return fmt.Sprintf("Unknown (%d)", roleID)
	}
}

// UserQRCodeResponse represents QR code data for a user
type UserQRCodeResponse struct {
	UserID      uuid.UUID `json:"user_id"`
	UserName    string    `json:"user_name"`
	UserEmail   string    `json:"user_email"`
	ProjectID   int       `json:"project_id"`
	ProjectName string    `json:"project_name"`
	QRCodeData  string    `json:"qr_code_data"`
}

// GetUserQRCode generates QR code data for a user's ODK Collect access
func (s *UserODKService) GetUserQRCode(ctx context.Context, userID uuid.UUID) (*UserQRCodeResponse, error) {
	if s.odkClient == nil {
		return nil, fmt.Errorf("ODK client not configured")
	}

	// Get the user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Check if user has App User
	if user.ODKAppUserID == nil || *user.ODKAppUserID == 0 {
		return nil, fmt.Errorf("user does not have ODK Collect access")
	}

	// Check if user has token
	if user.ODKAppUserToken == nil || *user.ODKAppUserToken == "" {
		return nil, fmt.Errorf("user does not have QR code token")
	}

	// Check if user has project ID
	if user.ODKAppUserProjectID == nil || *user.ODKAppUserProjectID == 0 {
		return nil, fmt.Errorf("user does not have associated ODK project")
	}

	// Get project info
	project, err := s.odkClient.GetProject(*user.ODKAppUserProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project info: %w", err)
	}

	// Generate QR code data
	qrCodeData := odk.GenerateQRCodeData(s.odkClient.GetBaseURL(), *user.ODKAppUserProjectID, *user.ODKAppUserToken)

	userName := user.Email
	if user.Name != nil && *user.Name != "" {
		userName = *user.Name
	}

	return &UserQRCodeResponse{
		UserID:      user.ID,
		UserName:    userName,
		UserEmail:   user.Email,
		ProjectID:   *user.ODKAppUserProjectID,
		ProjectName: project.Name,
		QRCodeData:  qrCodeData,
	}, nil
}
