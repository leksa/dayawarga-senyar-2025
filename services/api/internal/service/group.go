package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/leksa/datamapper-senyar/internal/model"
	"github.com/leksa/datamapper-senyar/internal/repository"
)

// GroupService handles group business logic
type GroupService struct {
	repo    *repository.GroupRepository
	orgRepo *repository.OrganizationRepository
}

// NewGroupService creates a new group service
func NewGroupService(repo *repository.GroupRepository, orgRepo *repository.OrganizationRepository) *GroupService {
	return &GroupService{repo: repo, orgRepo: orgRepo}
}

// CreateGroupInput represents input for creating a group
type CreateGroupInput struct {
	OrganizationID string  `json:"organization_id" binding:"required"`
	Name           string  `json:"name" binding:"required"`
	Description    *string `json:"description"`
}

// UpdateGroupInput represents input for updating a group
type UpdateGroupInput struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	IsActive    *bool   `json:"is_active"`
}

// List returns paginated groups
func (s *GroupService) List(ctx context.Context, filter repository.GroupFilter) (*repository.GroupListResult, error) {
	return s.repo.List(ctx, filter)
}

// ListWithOrgFilter returns paginated groups filtered by organization IDs
// If orgIDs is nil, returns all groups (for super_admin)
// If orgIDs is empty slice, returns no groups
func (s *GroupService) ListWithOrgFilter(ctx context.Context, filter repository.GroupFilter, orgIDs []uuid.UUID) (*repository.GroupListResult, error) {
	if orgIDs == nil {
		// Super admin - no org filtering
		return s.repo.List(ctx, filter)
	}
	// Org admin - filter by their org IDs
	return s.repo.ListByOrgIDs(ctx, filter, orgIDs)
}

// GetByID retrieves a group by ID
func (s *GroupService) GetByID(ctx context.Context, id string) (*model.Group, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid group ID format")
	}
	return s.repo.FindByID(ctx, parsed)
}

// GetByIDWithRelawan retrieves a group with relawan members
func (s *GroupService) GetByIDWithRelawan(ctx context.Context, id string) (*model.Group, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid group ID format")
	}
	return s.repo.FindByIDWithRelawan(ctx, parsed)
}

// Create creates a new group
func (s *GroupService) Create(ctx context.Context, input CreateGroupInput) (*model.Group, error) {
	// Parse organization ID
	orgID, err := uuid.Parse(input.OrganizationID)
	if err != nil {
		return nil, errors.New("invalid organization ID format")
	}

	// Verify organization exists
	_, err = s.orgRepo.FindByID(ctx, orgID)
	if err != nil {
		return nil, errors.New("organization not found")
	}

	// Check if group name already exists in organization
	exists, err := s.repo.NameExistsInOrg(ctx, orgID, input.Name, nil)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("group name already exists in organization")
	}

	group := &model.Group{
		OrganizationID: orgID,
		Name:           input.Name,
		Description:    input.Description,
		IsActive:       true,
	}

	if err := s.repo.Create(ctx, group); err != nil {
		return nil, err
	}

	// Reload with organization
	return s.repo.FindByID(ctx, group.ID)
}

// Update updates an existing group
func (s *GroupService) Update(ctx context.Context, id string, input UpdateGroupInput) (*model.Group, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid group ID format")
	}

	group, err := s.repo.FindByID(ctx, parsed)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if input.Name != nil {
		// Check if new name conflicts with existing
		exists, err := s.repo.NameExistsInOrg(ctx, group.OrganizationID, *input.Name, &group.ID)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("group name already exists in organization")
		}
		group.Name = *input.Name
	}
	if input.Description != nil {
		group.Description = input.Description
	}
	if input.IsActive != nil {
		group.IsActive = *input.IsActive
	}

	if err := s.repo.Update(ctx, group); err != nil {
		return nil, err
	}

	return s.repo.FindByID(ctx, group.ID)
}

// Delete soft-deletes a group
func (s *GroupService) Delete(ctx context.Context, id string) error {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return errors.New("invalid group ID format")
	}
	return s.repo.Delete(ctx, parsed)
}

// GetStats returns statistics for a group
func (s *GroupService) GetStats(ctx context.Context, id string) (map[string]int64, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid group ID format")
	}
	return s.repo.GetStats(ctx, parsed)
}

// GetByOrganization returns all groups for an organization
func (s *GroupService) GetByOrganization(ctx context.Context, orgID string) ([]model.Group, error) {
	parsed, err := uuid.Parse(orgID)
	if err != nil {
		return nil, errors.New("invalid organization ID format")
	}
	return s.repo.GetByOrganization(ctx, parsed)
}
