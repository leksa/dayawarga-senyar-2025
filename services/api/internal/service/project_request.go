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

// ProjectRequestService handles project request business logic
type ProjectRequestService struct {
	requestRepo      *repository.ProjectRequestRepository
	groupProjectRepo *repository.GroupProjectRepository
	groupRepo        *repository.GroupRepository
	userRepo         *repository.UserRepository
	odkClient        *odk.Client
	db               *gorm.DB
}

// NewProjectRequestService creates a new project request service
func NewProjectRequestService(
	requestRepo *repository.ProjectRequestRepository,
	groupProjectRepo *repository.GroupProjectRepository,
	groupRepo *repository.GroupRepository,
	userRepo *repository.UserRepository,
	odkClient *odk.Client,
	db *gorm.DB,
) *ProjectRequestService {
	return &ProjectRequestService{
		requestRepo:      requestRepo,
		groupProjectRepo: groupProjectRepo,
		groupRepo:        groupRepo,
		userRepo:         userRepo,
		odkClient:        odkClient,
		db:               db,
	}
}

// CreateProjectRequestInput represents input for creating a project request
type CreateProjectRequestInput struct {
	GroupID      uuid.UUID
	ODKProjectID int
	Notes        *string
}

// CreateRequest creates a new project request
func (s *ProjectRequestService) CreateRequest(ctx context.Context, requesterID uuid.UUID, input CreateProjectRequestInput) (*model.ProjectRequest, error) {
	// Get group to verify it exists and get organization ID
	group, err := s.groupRepo.FindByID(ctx, input.GroupID)
	if err != nil {
		return nil, fmt.Errorf("group not found: %w", err)
	}

	// Check if group already has a pending request
	hasPending, err := s.requestRepo.ExistsPendingForGroup(ctx, input.GroupID)
	if err != nil {
		return nil, fmt.Errorf("failed to check pending requests: %w", err)
	}
	if hasPending {
		return nil, fmt.Errorf("group already has a pending project request")
	}

	// Check if group already has this project assigned
	exists, err := s.groupProjectRepo.Exists(ctx, input.GroupID, input.ODKProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing assignment: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("group already has this project assigned")
	}

	// Verify group has a leader
	if group.LeaderID == nil {
		return nil, fmt.Errorf("group must have a leader before requesting a project")
	}

	// Get project name from ODK Central
	var projectName *string
	if s.odkClient != nil {
		project, err := s.odkClient.GetProject(input.ODKProjectID)
		if err == nil && project != nil {
			projectName = &project.Name
		}
	}

	// Create the request
	request := &model.ProjectRequest{
		OrganizationID: group.OrganizationID,
		GroupID:        input.GroupID,
		ODKProjectID:   input.ODKProjectID,
		ODKProjectName: projectName,
		RequestedBy:    requesterID,
		RequestNotes:   input.Notes,
		Status:         model.ProjectRequestStatusPending,
	}

	if err := s.requestRepo.Create(ctx, request); err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return s.requestRepo.GetByID(ctx, request.ID)
}

// GetRequest retrieves a project request by ID
func (s *ProjectRequestService) GetRequest(ctx context.Context, id uuid.UUID) (*model.ProjectRequest, error) {
	return s.requestRepo.GetByID(ctx, id)
}

// ListRequests retrieves project requests with filters
func (s *ProjectRequestService) ListRequests(ctx context.Context, filter repository.ProjectRequestFilter, page, pageSize int) ([]model.ProjectRequest, int64, error) {
	return s.requestRepo.List(ctx, filter, page, pageSize)
}

// ListPendingRequests retrieves all pending project requests (for admin)
func (s *ProjectRequestService) ListPendingRequests(ctx context.Context, page, pageSize int) ([]model.ProjectRequest, int64, error) {
	return s.requestRepo.ListPending(ctx, page, pageSize)
}

// ApproveRequestInput represents input for approving a request
type ApproveRequestInput struct {
	Notes *string
}

// ApproveRequest approves a project request and creates the Project Manager in ODK
func (s *ProjectRequestService) ApproveRequest(ctx context.Context, requestID uuid.UUID, reviewerID uuid.UUID, input ApproveRequestInput) (*model.ProjectRequest, error) {
	// Get the request
	request, err := s.requestRepo.GetByID(ctx, requestID)
	if err != nil {
		return nil, fmt.Errorf("request not found: %w", err)
	}

	if !request.IsPending() {
		return nil, fmt.Errorf("request is not pending")
	}

	// Get the group with leader
	group, err := s.groupRepo.FindByID(ctx, request.GroupID)
	if err != nil {
		return nil, fmt.Errorf("group not found: %w", err)
	}

	if group.LeaderID == nil {
		return nil, fmt.Errorf("group has no leader")
	}

	// Get the leader user
	leader, err := s.userRepo.FindByID(ctx, *group.LeaderID)
	if err != nil {
		return nil, fmt.Errorf("leader not found: %w", err)
	}

	// Use transaction for consistency
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Approve the request
		if err := s.requestRepo.Approve(ctx, requestID, reviewerID, input.Notes); err != nil {
			return fmt.Errorf("failed to approve request: %w", err)
		}

		// 2. Create group project assignment
		now := time.Now()
		groupProject := &model.GroupProject{
			GroupID:           request.GroupID,
			ODKProjectID:      request.ODKProjectID,
			ODKProjectName:    request.ODKProjectName,
			ApprovedRequestID: &requestID,
			AssignedAt:        now,
		}
		if err := s.groupProjectRepo.Create(ctx, groupProject); err != nil {
			return fmt.Errorf("failed to create group project: %w", err)
		}

		// 3. Update group with ODK project ID
		odkProjectID := request.ODKProjectID
		if err := tx.Model(&model.Group{}).
			Where("id = ?", request.GroupID).
			Updates(map[string]interface{}{
				"odk_project_id": odkProjectID,
			}).Error; err != nil {
			return fmt.Errorf("failed to update group: %w", err)
		}

		// 4. Create Project Manager in ODK Central (if client available)
		if s.odkClient != nil {
			displayName := leader.Email
			if leader.Name != nil && *leader.Name != "" {
				displayName = *leader.Name
			}

			webUser, err := s.odkClient.GetOrCreateProjectManager(request.ODKProjectID, leader.Email, displayName)
			if err != nil {
				return fmt.Errorf("failed to create ODK project manager: %w", err)
			}

			// Update user with ODK web user ID
			if err := tx.Model(&model.User{}).
				Where("id = ?", leader.ID).
				Update("odk_web_user_id", webUser.ID).Error; err != nil {
				return fmt.Errorf("failed to update user ODK ID: %w", err)
			}

			// Mark group as having Project Manager created
			if err := tx.Model(&model.Group{}).
				Where("id = ?", request.GroupID).
				Update("odk_project_manager_created", true).Error; err != nil {
				return fmt.Errorf("failed to update group PM status: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return s.requestRepo.GetByID(ctx, requestID)
}

// RejectRequestInput represents input for rejecting a request
type RejectRequestInput struct {
	Notes *string
}

// RejectRequest rejects a project request
func (s *ProjectRequestService) RejectRequest(ctx context.Context, requestID uuid.UUID, reviewerID uuid.UUID, input RejectRequestInput) (*model.ProjectRequest, error) {
	// Get the request
	request, err := s.requestRepo.GetByID(ctx, requestID)
	if err != nil {
		return nil, fmt.Errorf("request not found: %w", err)
	}

	if !request.IsPending() {
		return nil, fmt.Errorf("request is not pending")
	}

	if err := s.requestRepo.Reject(ctx, requestID, reviewerID, input.Notes); err != nil {
		return nil, fmt.Errorf("failed to reject request: %w", err)
	}

	return s.requestRepo.GetByID(ctx, requestID)
}

// ODKProjectService provides ODK project related functionality
type ODKProjectService struct {
	odkClient *odk.Client
}

// NewODKProjectService creates a new ODK project service
func NewODKProjectService(odkClient *odk.Client) *ODKProjectService {
	return &ODKProjectService{odkClient: odkClient}
}

// ListProjects lists all available ODK projects
func (s *ODKProjectService) ListProjects(ctx context.Context) ([]odk.Project, error) {
	if s.odkClient == nil {
		return nil, fmt.Errorf("ODK client not configured")
	}
	return s.odkClient.ListProjects()
}

// GetProject gets a specific ODK project
func (s *ODKProjectService) GetProject(ctx context.Context, projectID int) (*odk.Project, error) {
	if s.odkClient == nil {
		return nil, fmt.Errorf("ODK client not configured")
	}
	return s.odkClient.GetProject(projectID)
}

// ListProjectForms lists all forms in an ODK project
func (s *ODKProjectService) ListProjectForms(ctx context.Context, projectID int) ([]odk.Form, error) {
	if s.odkClient == nil {
		return nil, fmt.Errorf("ODK client not configured")
	}
	return s.odkClient.ListProjectForms(projectID)
}
