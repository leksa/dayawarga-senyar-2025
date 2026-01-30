package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/leksa/datamapper-senyar/internal/model"
	"gorm.io/gorm"
)

// BidangRepository handles database operations for bidangs
type BidangRepository struct {
	db *gorm.DB
}

// NewBidangRepository creates a new bidang repository
func NewBidangRepository(db *gorm.DB) *BidangRepository {
	return &BidangRepository{db: db}
}

// List returns all active bidangs
func (r *BidangRepository) List(ctx context.Context) ([]model.Bidang, error) {
	var bidangs []model.Bidang
	if err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("name ASC").
		Find(&bidangs).Error; err != nil {
		return nil, err
	}
	return bidangs, nil
}

// GetByID retrieves a bidang by ID
func (r *BidangRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Bidang, error) {
	var bidang model.Bidang
	if err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&bidang).Error; err != nil {
		return nil, err
	}
	return &bidang, nil
}

// GetBySlug retrieves a bidang by slug
func (r *BidangRepository) GetBySlug(ctx context.Context, slug string) (*model.Bidang, error) {
	var bidang model.Bidang
	if err := r.db.WithContext(ctx).
		Where("slug = ?", slug).
		First(&bidang).Error; err != nil {
		return nil, err
	}
	return &bidang, nil
}

// GetByOrganization retrieves all bidangs for an organization
func (r *BidangRepository) GetByOrganization(ctx context.Context, orgID uuid.UUID) ([]model.Bidang, error) {
	var bidangs []model.Bidang
	if err := r.db.WithContext(ctx).
		Joins("JOIN organization_bidangs ON bidangs.id = organization_bidangs.bidang_id").
		Where("organization_bidangs.organization_id = ? AND bidangs.is_active = ?", orgID, true).
		Order("bidangs.name ASC").
		Find(&bidangs).Error; err != nil {
		return nil, err
	}
	return bidangs, nil
}

// Create creates a new bidang
func (r *BidangRepository) Create(ctx context.Context, bidang *model.Bidang) error {
	return r.db.WithContext(ctx).Create(bidang).Error
}

// Update updates an existing bidang
func (r *BidangRepository) Update(ctx context.Context, bidang *model.Bidang) error {
	return r.db.WithContext(ctx).Save(bidang).Error
}

// SlugExists checks if a slug already exists (excluding given ID)
func (r *BidangRepository) SlugExists(ctx context.Context, slug string, excludeID *uuid.UUID) (bool, error) {
	query := r.db.WithContext(ctx).Model(&model.Bidang{}).Where("slug = ?", slug)
	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// AddToOrganization adds a bidang to an organization
func (r *BidangRepository) AddToOrganization(ctx context.Context, orgBidang *model.OrganizationBidang) error {
	return r.db.WithContext(ctx).Create(orgBidang).Error
}

// RemoveFromOrganization removes a bidang from an organization
func (r *BidangRepository) RemoveFromOrganization(ctx context.Context, orgID, bidangID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("organization_id = ? AND bidang_id = ?", orgID, bidangID).
		Delete(&model.OrganizationBidang{}).Error
}
