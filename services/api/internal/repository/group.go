package repository

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/leksa/datamapper-senyar/internal/model"
	"gorm.io/gorm"
)

// GroupRepository handles database operations for groups
type GroupRepository struct {
	db *gorm.DB
}

// NewGroupRepository creates a new group repository
func NewGroupRepository(db *gorm.DB) *GroupRepository {
	return &GroupRepository{db: db}
}

// GroupFilter defines filters for querying groups
type GroupFilter struct {
	OrganizationID *uuid.UUID
	Search         string
	IsActive       *bool
	Page           int
	PageSize       int
}

// GroupListResult contains paginated group results
type GroupListResult struct {
	Groups     []model.Group `json:"groups"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
	TotalPages int           `json:"total_pages"`
}

// List returns paginated groups with optional filters
func (r *GroupRepository) List(ctx context.Context, filter GroupFilter) (*GroupListResult, error) {
	query := r.db.WithContext(ctx).Model(&model.Group{}).Where("deleted_at IS NULL")

	// Apply organization filter
	if filter.OrganizationID != nil {
		query = query.Where("organization_id = ?", *filter.OrganizationID)
	}

	// Apply search filter
	if filter.Search != "" {
		searchTerm := "%" + strings.ToLower(filter.Search) + "%"
		query = query.Where("LOWER(name) LIKE ?", searchTerm)
	}

	// Apply active filter
	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// Set default pagination
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}

	// Calculate offset
	offset := (filter.Page - 1) * filter.PageSize

	// Fetch groups with organization
	var groups []model.Group
	if err := query.
		Preload("Organization").
		Order("created_at DESC").
		Offset(offset).
		Limit(filter.PageSize).
		Find(&groups).Error; err != nil {
		return nil, err
	}

	// Calculate total pages
	totalPages := int(total) / filter.PageSize
	if int(total)%filter.PageSize > 0 {
		totalPages++
	}

	return &GroupListResult{
		Groups:     groups,
		Total:      total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: totalPages,
	}, nil
}

// FindByID retrieves a group by ID
func (r *GroupRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Group, error) {
	var group model.Group
	if err := r.db.WithContext(ctx).
		Preload("Organization").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&group).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

// FindByIDWithRelawan retrieves a group with relawan members
func (r *GroupRepository) FindByIDWithRelawan(ctx context.Context, id uuid.UUID) (*model.Group, error) {
	var group model.Group
	if err := r.db.WithContext(ctx).
		Preload("Organization").
		Preload("Relawan", "deleted_at IS NULL").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&group).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

// Create creates a new group
func (r *GroupRepository) Create(ctx context.Context, group *model.Group) error {
	return r.db.WithContext(ctx).Create(group).Error
}

// Update updates an existing group
func (r *GroupRepository) Update(ctx context.Context, group *model.Group) error {
	return r.db.WithContext(ctx).Save(group).Error
}

// Delete soft-deletes a group
func (r *GroupRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&model.Group{}).
		Where("id = ?", id).
		Update("deleted_at", gorm.Expr("NOW()")).Error
}

// GetStats returns statistics for a group
func (r *GroupRepository) GetStats(ctx context.Context, id uuid.UUID) (map[string]int64, error) {
	stats := make(map[string]int64)

	// Count relawan in this group
	var relawanCount int64
	if err := r.db.WithContext(ctx).Model(&model.Relawan{}).
		Where("group_id = ? AND deleted_at IS NULL", id).
		Count(&relawanCount).Error; err != nil {
		return nil, err
	}
	stats["relawan"] = relawanCount

	// Count active relawan
	var activeRelawanCount int64
	if err := r.db.WithContext(ctx).Model(&model.Relawan{}).
		Where("group_id = ? AND deleted_at IS NULL AND status = ?", id, model.RelawanStatusActive).
		Count(&activeRelawanCount).Error; err != nil {
		return nil, err
	}
	stats["active_relawan"] = activeRelawanCount

	return stats, nil
}

// GetByOrganization returns all groups for an organization
func (r *GroupRepository) GetByOrganization(ctx context.Context, orgID uuid.UUID) ([]model.Group, error) {
	var groups []model.Group
	if err := r.db.WithContext(ctx).
		Where("organization_id = ? AND deleted_at IS NULL", orgID).
		Order("name ASC").
		Find(&groups).Error; err != nil {
		return nil, err
	}
	return groups, nil
}

// NameExistsInOrg checks if a group name already exists in an organization
func (r *GroupRepository) NameExistsInOrg(ctx context.Context, orgID uuid.UUID, name string, excludeID *uuid.UUID) (bool, error) {
	query := r.db.WithContext(ctx).Model(&model.Group{}).
		Where("organization_id = ? AND LOWER(name) = ? AND deleted_at IS NULL", orgID, strings.ToLower(name))
	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ListByOrgIDs returns paginated groups filtered by organization IDs (for org-scoped access)
func (r *GroupRepository) ListByOrgIDs(ctx context.Context, filter GroupFilter, orgIDs []uuid.UUID) (*GroupListResult, error) {
	query := r.db.WithContext(ctx).Model(&model.Group{}).Where("deleted_at IS NULL")

	// Apply organization IDs filter
	if len(orgIDs) > 0 {
		query = query.Where("organization_id IN ?", orgIDs)
	}

	// Apply specific organization filter (if user also filters by org_id)
	if filter.OrganizationID != nil {
		query = query.Where("organization_id = ?", *filter.OrganizationID)
	}

	// Apply search filter
	if filter.Search != "" {
		searchTerm := "%" + strings.ToLower(filter.Search) + "%"
		query = query.Where("LOWER(name) LIKE ?", searchTerm)
	}

	// Apply active filter
	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// Set default pagination
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}

	// Calculate offset
	offset := (filter.Page - 1) * filter.PageSize

	// Fetch groups with organization
	var groups []model.Group
	if err := query.
		Preload("Organization").
		Order("created_at DESC").
		Offset(offset).
		Limit(filter.PageSize).
		Find(&groups).Error; err != nil {
		return nil, err
	}

	// Calculate total pages
	totalPages := int(total) / filter.PageSize
	if int(total)%filter.PageSize > 0 {
		totalPages++
	}

	return &GroupListResult{
		Groups:     groups,
		Total:      total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: totalPages,
	}, nil
}
