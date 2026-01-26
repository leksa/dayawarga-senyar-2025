package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/leksa/datamapper-senyar/internal/model"
	"github.com/leksa/datamapper-senyar/internal/repository"
)

// RelawanService handles relawan business logic
type RelawanService struct {
	repo      *repository.RelawanRepository
	orgRepo   *repository.OrganizationRepository
	groupRepo *repository.GroupRepository
}

// NewRelawanService creates a new relawan service
func NewRelawanService(repo *repository.RelawanRepository, orgRepo *repository.OrganizationRepository, groupRepo *repository.GroupRepository) *RelawanService {
	return &RelawanService{repo: repo, orgRepo: orgRepo, groupRepo: groupRepo}
}

// CreateRelawanInput represents input for creating a relawan
type CreateRelawanInput struct {
	OrganizationID string   `json:"organization_id" binding:"required"`
	GroupID        *string  `json:"group_id"`
	Name           string   `json:"name" binding:"required"`
	Phone          *string  `json:"phone"`
	Email          *string  `json:"email"`
	AssignedForms  []string `json:"assigned_forms"`
	Notes          *string  `json:"notes"`
}

// UpdateRelawanInput represents input for updating a relawan
type UpdateRelawanInput struct {
	GroupID       *string              `json:"group_id"`
	Name          *string              `json:"name"`
	Phone         *string              `json:"phone"`
	Email         *string              `json:"email"`
	AssignedForms []string             `json:"assigned_forms"`
	Notes         *string              `json:"notes"`
	Status        *model.RelawanStatus `json:"status"`
}

// List returns paginated relawan
func (s *RelawanService) List(ctx context.Context, filter repository.RelawanFilter) (*repository.RelawanListResult, error) {
	return s.repo.List(ctx, filter)
}

// ListWithOrgFilter returns paginated relawan filtered by organization IDs
// If orgIDs is nil, returns all relawan (for super_admin)
// If orgIDs is empty slice, returns no relawan
func (s *RelawanService) ListWithOrgFilter(ctx context.Context, filter repository.RelawanFilter, orgIDs []uuid.UUID) (*repository.RelawanListResult, error) {
	if orgIDs == nil {
		// Super admin - no org filtering
		return s.repo.List(ctx, filter)
	}
	// Org admin - filter by their org IDs
	return s.repo.ListByOrgIDs(ctx, filter, orgIDs)
}

// GetByID retrieves a relawan by ID
func (s *RelawanService) GetByID(ctx context.Context, id string) (*model.Relawan, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid relawan ID format")
	}
	return s.repo.FindByID(ctx, parsed)
}

// Create creates a new relawan
func (s *RelawanService) Create(ctx context.Context, input CreateRelawanInput) (*model.Relawan, error) {
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

	// Parse group ID if provided
	var groupID *uuid.UUID
	if input.GroupID != nil && *input.GroupID != "" {
		parsed, err := uuid.Parse(*input.GroupID)
		if err != nil {
			return nil, errors.New("invalid group ID format")
		}
		// Verify group exists and belongs to organization
		group, err := s.groupRepo.FindByID(ctx, parsed)
		if err != nil {
			return nil, errors.New("group not found")
		}
		if group.OrganizationID != orgID {
			return nil, errors.New("group does not belong to organization")
		}
		groupID = &parsed
	}

	// Check phone uniqueness if provided
	if input.Phone != nil && *input.Phone != "" {
		exists, err := s.repo.PhoneExistsInOrg(ctx, orgID, *input.Phone, nil)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("phone number already exists in organization")
		}
	}

	relawan := &model.Relawan{
		OrganizationID: orgID,
		GroupID:        groupID,
		Name:           input.Name,
		Phone:          input.Phone,
		Email:          input.Email,
		AssignedForms:  input.AssignedForms,
		Notes:          input.Notes,
		Status:         model.RelawanStatusActive,
	}

	if err := s.repo.Create(ctx, relawan); err != nil {
		return nil, err
	}

	// Reload with relations
	return s.repo.FindByID(ctx, relawan.ID)
}

// Update updates an existing relawan
func (s *RelawanService) Update(ctx context.Context, id string, input UpdateRelawanInput) (*model.Relawan, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid relawan ID format")
	}

	relawan, err := s.repo.FindByID(ctx, parsed)
	if err != nil {
		return nil, err
	}

	// Update group if provided
	if input.GroupID != nil {
		if *input.GroupID == "" {
			relawan.GroupID = nil
		} else {
			groupUUID, err := uuid.Parse(*input.GroupID)
			if err != nil {
				return nil, errors.New("invalid group ID format")
			}
			// Verify group exists and belongs to same organization
			group, err := s.groupRepo.FindByID(ctx, groupUUID)
			if err != nil {
				return nil, errors.New("group not found")
			}
			if group.OrganizationID != relawan.OrganizationID {
				return nil, errors.New("group does not belong to organization")
			}
			relawan.GroupID = &groupUUID
		}
	}

	// Update other fields if provided
	if input.Name != nil {
		relawan.Name = *input.Name
	}
	if input.Phone != nil {
		if *input.Phone != "" {
			// Check phone uniqueness
			exists, err := s.repo.PhoneExistsInOrg(ctx, relawan.OrganizationID, *input.Phone, &relawan.ID)
			if err != nil {
				return nil, err
			}
			if exists {
				return nil, errors.New("phone number already exists in organization")
			}
		}
		relawan.Phone = input.Phone
	}
	if input.Email != nil {
		relawan.Email = input.Email
	}
	if input.AssignedForms != nil {
		relawan.AssignedForms = input.AssignedForms
	}
	if input.Notes != nil {
		relawan.Notes = input.Notes
	}
	if input.Status != nil {
		relawan.Status = *input.Status
	}

	if err := s.repo.Update(ctx, relawan); err != nil {
		return nil, err
	}

	return s.repo.FindByID(ctx, relawan.ID)
}

// Delete soft-deletes a relawan
func (s *RelawanService) Delete(ctx context.Context, id string) error {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return errors.New("invalid relawan ID format")
	}
	return s.repo.Delete(ctx, parsed)
}

// UpdateStatus updates a relawan's status
func (s *RelawanService) UpdateStatus(ctx context.Context, id string, status model.RelawanStatus) error {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return errors.New("invalid relawan ID format")
	}
	return s.repo.UpdateStatus(ctx, parsed, status)
}

// MoveToGroup moves a relawan to a different group
func (s *RelawanService) MoveToGroup(ctx context.Context, id string, groupID *string) error {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return errors.New("invalid relawan ID format")
	}

	// Get relawan to verify ownership
	relawan, err := s.repo.FindByID(ctx, parsed)
	if err != nil {
		return errors.New("relawan not found")
	}

	var groupUUID *uuid.UUID
	if groupID != nil && *groupID != "" {
		gid, err := uuid.Parse(*groupID)
		if err != nil {
			return errors.New("invalid group ID format")
		}
		// Verify group belongs to same organization
		group, err := s.groupRepo.FindByID(ctx, gid)
		if err != nil {
			return errors.New("group not found")
		}
		if group.OrganizationID != relawan.OrganizationID {
			return errors.New("group does not belong to organization")
		}
		groupUUID = &gid
	}

	return s.repo.UpdateGroup(ctx, parsed, groupUUID)
}

// GetByOrganization returns all relawan for an organization
func (s *RelawanService) GetByOrganization(ctx context.Context, orgID string) ([]model.Relawan, error) {
	parsed, err := uuid.Parse(orgID)
	if err != nil {
		return nil, errors.New("invalid organization ID format")
	}
	return s.repo.GetByOrganization(ctx, parsed)
}

// GetByGroup returns all relawan for a group
func (s *RelawanService) GetByGroup(ctx context.Context, groupID string) ([]model.Relawan, error) {
	parsed, err := uuid.Parse(groupID)
	if err != nil {
		return nil, errors.New("invalid group ID format")
	}
	return s.repo.GetByGroup(ctx, parsed)
}

// GetStats returns statistics for relawan
func (s *RelawanService) GetStats(ctx context.Context, orgID *string) (map[string]int64, error) {
	var orgUUID *uuid.UUID
	if orgID != nil && *orgID != "" {
		parsed, err := uuid.Parse(*orgID)
		if err != nil {
			return nil, errors.New("invalid organization ID format")
		}
		orgUUID = &parsed
	}
	return s.repo.GetStats(ctx, orgUUID)
}

// GetStatsWithOrgFilter returns statistics for relawan filtered by organization IDs
// If orgIDs is nil, returns stats for all orgs (super_admin)
// If orgID is provided, it must be within the allowed orgIDs
func (s *RelawanService) GetStatsWithOrgFilter(ctx context.Context, orgID *string, userOrgIDs []uuid.UUID) (map[string]int64, error) {
	// If specific org is requested
	if orgID != nil && *orgID != "" {
		parsed, err := uuid.Parse(*orgID)
		if err != nil {
			return nil, errors.New("invalid organization ID format")
		}
		return s.repo.GetStats(ctx, &parsed)
	}

	// Super admin - get all stats
	if userOrgIDs == nil {
		return s.repo.GetStats(ctx, nil)
	}

	// Org admin without specific org filter - get stats for their orgs
	return s.repo.GetStatsByOrgIDs(ctx, userOrgIDs)
}

// BulkMoveToGroup moves multiple relawan to a group
func (s *RelawanService) BulkMoveToGroup(ctx context.Context, ids []string, groupID *string) error {
	// Parse all IDs
	uuids := make([]uuid.UUID, len(ids))
	for i, id := range ids {
		parsed, err := uuid.Parse(id)
		if err != nil {
			return errors.New("invalid relawan ID format")
		}
		uuids[i] = parsed
	}

	var groupUUID *uuid.UUID
	if groupID != nil && *groupID != "" {
		gid, err := uuid.Parse(*groupID)
		if err != nil {
			return errors.New("invalid group ID format")
		}
		groupUUID = &gid
	}

	return s.repo.BulkUpdateGroup(ctx, uuids, groupUUID)
}

// SetWAVerified enables or disables WhatsApp verification for a relawan
func (s *RelawanService) SetWAVerified(ctx context.Context, id string, verified bool) error {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return errors.New("invalid relawan ID format")
	}

	relawan, err := s.repo.FindByID(ctx, parsed)
	if err != nil {
		return errors.New("relawan not found")
	}

	if relawan.Phone == nil || *relawan.Phone == "" {
		return errors.New("relawan must have a phone number to enable WhatsApp access")
	}

	return s.repo.SetWAVerified(ctx, parsed, verified)
}

// GetWAStatus returns WhatsApp status for a relawan
func (s *RelawanService) GetWAStatus(ctx context.Context, id string) (*WAStatus, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid relawan ID format")
	}

	relawan, err := s.repo.FindByID(ctx, parsed)
	if err != nil {
		return nil, errors.New("relawan not found")
	}

	return &WAStatus{
		Verified:     relawan.WAVerified,
		VerifiedAt:   relawan.WAVerifiedAt,
		LastActivity: relawan.WALastActivity,
		SessionCount: relawan.WASessionCount,
		HasPhone:     relawan.Phone != nil && *relawan.Phone != "",
	}, nil
}

// ValidateWAAccess checks if a phone number has WhatsApp access (for chatbot)
func (s *RelawanService) ValidateWAAccess(ctx context.Context, phone string) (*model.Relawan, error) {
	return s.repo.FindByPhoneWithWAAccess(ctx, phone)
}

// UpdateWAActivity updates activity timestamp for a relawan (called by chatbot)
func (s *RelawanService) UpdateWAActivity(ctx context.Context, phone string) error {
	relawan, err := s.repo.FindByPhoneWithWAAccess(ctx, phone)
	if err != nil {
		return errors.New("relawan not found or not verified")
	}
	return s.repo.UpdateWAActivity(ctx, relawan.ID)
}

// WAStatus represents WhatsApp verification status
type WAStatus struct {
	Verified     bool       `json:"verified"`
	VerifiedAt   *time.Time `json:"verified_at,omitempty"`
	LastActivity *time.Time `json:"last_activity,omitempty"`
	SessionCount int        `json:"session_count"`
	HasPhone     bool       `json:"has_phone"`
}
