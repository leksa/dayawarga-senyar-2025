package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/leksa/datamapper-senyar/internal/model"
	"github.com/leksa/datamapper-senyar/internal/odk"
	"github.com/leksa/datamapper-senyar/internal/repository"
)

// RelawanODKService handles ODK App User operations for relawan
type RelawanODKService struct {
	relawanRepo *repository.RelawanRepository
	groupRepo   *repository.GroupRepository
	odkClient   *odk.Client
	odkBaseURL  string
	db          *gorm.DB
}

// NewRelawanODKService creates a new relawan ODK service
func NewRelawanODKService(
	relawanRepo *repository.RelawanRepository,
	groupRepo *repository.GroupRepository,
	odkClient *odk.Client,
	odkBaseURL string,
	db *gorm.DB,
) *RelawanODKService {
	return &RelawanODKService{
		relawanRepo: relawanRepo,
		groupRepo:   groupRepo,
		odkClient:   odkClient,
		odkBaseURL:  odkBaseURL,
		db:          db,
	}
}

// CreateAppUserForRelawan creates an ODK App User for a relawan
// This is called when:
// 1. Relawan is created and assigned to a group with approved ODK project
// 2. Relawan is moved to a group with approved ODK project
func (s *RelawanODKService) CreateAppUserForRelawan(ctx context.Context, relawanID uuid.UUID) (*model.Relawan, error) {
	// Get relawan with group
	relawan, err := s.relawanRepo.FindByID(ctx, relawanID)
	if err != nil {
		return nil, fmt.Errorf("relawan not found: %w", err)
	}

	// Check if relawan already has ODK App User
	if relawan.HasODKAccess() {
		return relawan, nil // Already has access
	}

	// Check if relawan has a group
	if relawan.GroupID == nil {
		return nil, fmt.Errorf("relawan is not assigned to any group")
	}

	// Get group with ODK project info
	group, err := s.groupRepo.FindByID(ctx, *relawan.GroupID)
	if err != nil {
		return nil, fmt.Errorf("group not found: %w", err)
	}

	// Check if group has approved ODK project
	if !group.HasODKProject() {
		return nil, fmt.Errorf("group does not have an approved ODK project")
	}

	// Create App User in ODK Central
	appUser, err := s.odkClient.CreateAppUser(*group.ODKProjectID, relawan.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to create ODK App User: %w", err)
	}

	// Update relawan with ODK App User info
	now := time.Now()
	relawan.ODKAppUserID = &appUser.ID
	relawan.ODKAppUserToken = &appUser.Token
	relawan.ODKAppUserCreatedAt = &now

	if err := s.relawanRepo.Update(ctx, relawan); err != nil {
		// Try to clean up the created App User
		_ = s.odkClient.DeleteAppUser(*group.ODKProjectID, appUser.ID)
		return nil, fmt.Errorf("failed to update relawan: %w", err)
	}

	return relawan, nil
}

// RevokeAppUserForRelawan revokes ODK App User access for a relawan
func (s *RelawanODKService) RevokeAppUserForRelawan(ctx context.Context, relawanID uuid.UUID) error {
	relawan, err := s.relawanRepo.FindByID(ctx, relawanID)
	if err != nil {
		return fmt.Errorf("relawan not found: %w", err)
	}

	if !relawan.HasODKAccess() {
		return nil // No access to revoke
	}

	// Get group to find project ID
	if relawan.GroupID == nil {
		return fmt.Errorf("relawan has ODK access but no group - data inconsistency")
	}

	group, err := s.groupRepo.FindByID(ctx, *relawan.GroupID)
	if err != nil {
		return fmt.Errorf("group not found: %w", err)
	}

	if !group.HasODKProject() {
		return fmt.Errorf("group has no ODK project - data inconsistency")
	}

	// Delete App User in ODK Central
	if err := s.odkClient.DeleteAppUser(*group.ODKProjectID, *relawan.ODKAppUserID); err != nil {
		return fmt.Errorf("failed to delete ODK App User: %w", err)
	}

	// Clear ODK fields from relawan
	relawan.ODKAppUserID = nil
	relawan.ODKAppUserToken = nil
	relawan.ODKAppUserCreatedAt = nil

	if err := s.relawanRepo.Update(ctx, relawan); err != nil {
		return fmt.Errorf("failed to update relawan: %w", err)
	}

	return nil
}

// QRCodeResponse represents the QR code data for ODK Collect
type QRCodeResponse struct {
	RelawanID   uuid.UUID `json:"relawan_id"`
	RelawanName string    `json:"relawan_name"`
	GroupName   string    `json:"group_name"`
	ProjectID   int       `json:"project_id"`
	QRCodeData  string    `json:"qr_code_data"` // Base64 encoded JSON for ODK Collect
	CreatedAt   time.Time `json:"created_at"`
}

// GetQRCodeForRelawan generates QR code data for a relawan's ODK Collect
func (s *RelawanODKService) GetQRCodeForRelawan(ctx context.Context, relawanID uuid.UUID) (*QRCodeResponse, error) {
	relawan, err := s.relawanRepo.FindByID(ctx, relawanID)
	if err != nil {
		return nil, fmt.Errorf("relawan not found: %w", err)
	}

	if !relawan.HasODKAccess() {
		return nil, fmt.Errorf("relawan does not have ODK access")
	}

	if relawan.GroupID == nil {
		return nil, fmt.Errorf("relawan is not assigned to any group")
	}

	group, err := s.groupRepo.FindByID(ctx, *relawan.GroupID)
	if err != nil {
		return nil, fmt.Errorf("group not found: %w", err)
	}

	if !group.HasODKProject() {
		return nil, fmt.Errorf("group does not have an ODK project")
	}

	// Generate QR code data
	qrData := odk.GenerateQRCodeData(s.odkBaseURL, *group.ODKProjectID, *relawan.ODKAppUserToken)

	groupName := ""
	if group != nil {
		groupName = group.Name
	}

	return &QRCodeResponse{
		RelawanID:   relawan.ID,
		RelawanName: relawan.Name,
		GroupName:   groupName,
		ProjectID:   *group.ODKProjectID,
		QRCodeData:  qrData,
		CreatedAt:   *relawan.ODKAppUserCreatedAt,
	}, nil
}

// EnsureAppUserForGroupRelawan creates ODK App Users for all relawan in a group
// This is called after a project request is approved
func (s *RelawanODKService) EnsureAppUserForGroupRelawan(ctx context.Context, groupID uuid.UUID) (int, error) {
	// Get all relawan in group
	relawanList, err := s.relawanRepo.GetByGroup(ctx, groupID)
	if err != nil {
		return 0, fmt.Errorf("failed to get group relawan: %w", err)
	}

	created := 0
	for _, r := range relawanList {
		if r.HasODKAccess() {
			continue // Already has access
		}

		_, err := s.CreateAppUserForRelawan(ctx, r.ID)
		if err != nil {
			// Log error but continue with other relawan
			continue
		}
		created++
	}

	return created, nil
}

// AssignFormsToRelawan assigns forms to a relawan's ODK App User
func (s *RelawanODKService) AssignFormsToRelawan(ctx context.Context, relawanID uuid.UUID, formIDs []string) error {
	relawan, err := s.relawanRepo.FindByID(ctx, relawanID)
	if err != nil {
		return fmt.Errorf("relawan not found: %w", err)
	}

	if !relawan.HasODKAccess() {
		return fmt.Errorf("relawan does not have ODK access")
	}

	if relawan.GroupID == nil {
		return fmt.Errorf("relawan is not assigned to any group")
	}

	group, err := s.groupRepo.FindByID(ctx, *relawan.GroupID)
	if err != nil {
		return fmt.Errorf("group not found: %w", err)
	}

	if !group.HasODKProject() {
		return fmt.Errorf("group does not have an ODK project")
	}

	// Assign forms in ODK Central
	for _, formID := range formIDs {
		if err := s.odkClient.AssignFormToAppUser(*group.ODKProjectID, formID, *relawan.ODKAppUserID); err != nil {
			return fmt.Errorf("failed to assign form %s: %w", formID, err)
		}
	}

	// Update assigned forms in database
	relawan.AssignedForms = formIDs
	if err := s.relawanRepo.Update(ctx, relawan); err != nil {
		return fmt.Errorf("failed to update relawan: %w", err)
	}

	return nil
}
