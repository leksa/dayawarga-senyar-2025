package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/leksa/datamapper-senyar/internal/model"
	"github.com/leksa/datamapper-senyar/internal/repository"
)

// BidangService handles bidang business logic
type BidangService struct {
	repo *repository.BidangRepository
}

// NewBidangService creates a new bidang service
func NewBidangService(repo *repository.BidangRepository) *BidangService {
	return &BidangService{repo: repo}
}

// List returns all active bidangs
func (s *BidangService) List(ctx context.Context) ([]model.Bidang, error) {
	return s.repo.List(ctx)
}

// GetByID retrieves a bidang by ID
func (s *BidangService) GetByID(ctx context.Context, id string) (*model.Bidang, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid bidang ID format")
	}
	return s.repo.GetByID(ctx, parsed)
}

// GetByOrganization retrieves all bidangs for an organization
func (s *BidangService) GetByOrganization(ctx context.Context, orgID string) ([]model.Bidang, error) {
	parsed, err := uuid.Parse(orgID)
	if err != nil {
		return nil, errors.New("invalid organization ID format")
	}
	return s.repo.GetByOrganization(ctx, parsed)
}

// AddToOrganization adds a bidang to an organization
func (s *BidangService) AddToOrganization(ctx context.Context, orgID, bidangID string) error {
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		return errors.New("invalid organization ID format")
	}
	bidangUUID, err := uuid.Parse(bidangID)
	if err != nil {
		return errors.New("invalid bidang ID format")
	}

	orgBidang := &model.OrganizationBidang{
		OrganizationID: orgUUID,
		BidangID:       bidangUUID,
	}
	return s.repo.AddToOrganization(ctx, orgBidang)
}

// RemoveFromOrganization removes a bidang from an organization
func (s *BidangService) RemoveFromOrganization(ctx context.Context, orgID, bidangID string) error {
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		return errors.New("invalid organization ID format")
	}
	bidangUUID, err := uuid.Parse(bidangID)
	if err != nil {
		return errors.New("invalid bidang ID format")
	}
	return s.repo.RemoveFromOrganization(ctx, orgUUID, bidangUUID)
}
