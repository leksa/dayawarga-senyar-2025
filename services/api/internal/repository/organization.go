package repository

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/leksa/datamapper-senyar/internal/model"
	"gorm.io/gorm"
)

// OrganizationRepository handles database operations for organizations
type OrganizationRepository struct {
	db *gorm.DB
}

// NewOrganizationRepository creates a new organization repository
func NewOrganizationRepository(db *gorm.DB) *OrganizationRepository {
	return &OrganizationRepository{db: db}
}

// OrganizationFilter defines filters for querying organizations
type OrganizationFilter struct {
	Search   string
	IsActive *bool
	Page     int
	PageSize int
}

// OrganizationListResult contains paginated organization results
type OrganizationListResult struct {
	Organizations []model.Organization `json:"organizations"`
	Total         int64                `json:"total"`
	Page          int                  `json:"page"`
	PageSize      int                  `json:"page_size"`
	TotalPages    int                  `json:"total_pages"`
}

// List returns paginated organizations with optional filters
func (r *OrganizationRepository) List(ctx context.Context, filter OrganizationFilter) (*OrganizationListResult, error) {
	query := r.db.WithContext(ctx).Model(&model.Organization{}).Where("deleted_at IS NULL")

	// Apply search filter
	if filter.Search != "" {
		searchTerm := "%" + strings.ToLower(filter.Search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(slug) LIKE ?", searchTerm, searchTerm)
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

	// Fetch organizations
	var organizations []model.Organization
	if err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(filter.PageSize).
		Find(&organizations).Error; err != nil {
		return nil, err
	}

	// Calculate total pages
	totalPages := int(total) / filter.PageSize
	if int(total)%filter.PageSize > 0 {
		totalPages++
	}

	return &OrganizationListResult{
		Organizations: organizations,
		Total:         total,
		Page:          filter.Page,
		PageSize:      filter.PageSize,
		TotalPages:    totalPages,
	}, nil
}

// FindByID retrieves an organization by ID
func (r *OrganizationRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Organization, error) {
	var org model.Organization
	if err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&org).Error; err != nil {
		return nil, err
	}
	return &org, nil
}

// FindByIDWithRelations retrieves an organization with all relations
func (r *OrganizationRepository) FindByIDWithRelations(ctx context.Context, id uuid.UUID) (*model.Organization, error) {
	var org model.Organization
	if err := r.db.WithContext(ctx).
		Preload("Members.User").
		Preload("Groups", "deleted_at IS NULL").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&org).Error; err != nil {
		return nil, err
	}
	return &org, nil
}

// FindBySlug retrieves an organization by slug
func (r *OrganizationRepository) FindBySlug(ctx context.Context, slug string) (*model.Organization, error) {
	var org model.Organization
	if err := r.db.WithContext(ctx).
		Where("slug = ? AND deleted_at IS NULL", slug).
		First(&org).Error; err != nil {
		return nil, err
	}
	return &org, nil
}

// Create creates a new organization
func (r *OrganizationRepository) Create(ctx context.Context, org *model.Organization) error {
	return r.db.WithContext(ctx).Create(org).Error
}

// Update updates an existing organization
func (r *OrganizationRepository) Update(ctx context.Context, org *model.Organization) error {
	return r.db.WithContext(ctx).Save(org).Error
}

// Delete soft-deletes an organization
func (r *OrganizationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&model.Organization{}).
		Where("id = ?", id).
		Update("deleted_at", gorm.Expr("NOW()")).Error
}

// SlugExists checks if a slug already exists (excluding given ID)
func (r *OrganizationRepository) SlugExists(ctx context.Context, slug string, excludeID *uuid.UUID) (bool, error) {
	query := r.db.WithContext(ctx).Model(&model.Organization{}).Where("slug = ? AND deleted_at IS NULL", slug)
	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetStats returns statistics for an organization
func (r *OrganizationRepository) GetStats(ctx context.Context, id uuid.UUID) (map[string]int64, error) {
	stats := make(map[string]int64)

	// Count groups
	var groupCount int64
	if err := r.db.WithContext(ctx).Model(&model.Group{}).
		Where("organization_id = ? AND deleted_at IS NULL", id).
		Count(&groupCount).Error; err != nil {
		return nil, err
	}
	stats["groups"] = groupCount

	// Count relawan
	var relawanCount int64
	if err := r.db.WithContext(ctx).Model(&model.Relawan{}).
		Where("organization_id = ? AND deleted_at IS NULL", id).
		Count(&relawanCount).Error; err != nil {
		return nil, err
	}
	stats["relawan"] = relawanCount

	// Count active relawan
	var activeRelawanCount int64
	if err := r.db.WithContext(ctx).Model(&model.Relawan{}).
		Where("organization_id = ? AND deleted_at IS NULL AND status = ?", id, model.RelawanStatusActive).
		Count(&activeRelawanCount).Error; err != nil {
		return nil, err
	}
	stats["active_relawan"] = activeRelawanCount

	// Count members
	var memberCount int64
	if err := r.db.WithContext(ctx).Model(&model.OrganizationMember{}).
		Where("organization_id = ?", id).
		Count(&memberCount).Error; err != nil {
		return nil, err
	}
	stats["members"] = memberCount

	return stats, nil
}

// AddMember adds a user to an organization
func (r *OrganizationRepository) AddMember(ctx context.Context, member *model.OrganizationMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

// RemoveMember removes a user from an organization
func (r *OrganizationRepository) RemoveMember(ctx context.Context, orgID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("organization_id = ? AND user_id = ?", orgID, userID).
		Delete(&model.OrganizationMember{}).Error
}

// GetMember retrieves a member of an organization
func (r *OrganizationRepository) GetMember(ctx context.Context, orgID, userID uuid.UUID) (*model.OrganizationMember, error) {
	var member model.OrganizationMember
	if err := r.db.WithContext(ctx).
		Where("organization_id = ? AND user_id = ?", orgID, userID).
		First(&member).Error; err != nil {
		return nil, err
	}
	return &member, nil
}

// UpdateMemberRole updates a member's role in an organization
func (r *OrganizationRepository) UpdateMemberRole(ctx context.Context, orgID, userID uuid.UUID, role model.OrgMemberRole) error {
	return r.db.WithContext(ctx).
		Model(&model.OrganizationMember{}).
		Where("organization_id = ? AND user_id = ?", orgID, userID).
		Update("role", role).Error
}

// GetUserOrganizations returns all organizations a user belongs to
func (r *OrganizationRepository) GetUserOrganizations(ctx context.Context, userID uuid.UUID) ([]model.Organization, error) {
	var orgs []model.Organization
	if err := r.db.WithContext(ctx).
		Joins("JOIN organization_members ON organizations.id = organization_members.organization_id").
		Where("organization_members.user_id = ? AND organizations.deleted_at IS NULL", userID).
		Find(&orgs).Error; err != nil {
		return nil, err
	}
	return orgs, nil
}

// GetUserOrganizationIDs returns organization IDs that a user belongs to (for auth.OrganizationChecker interface)
func (r *OrganizationRepository) GetUserOrganizationIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	var orgIDs []uuid.UUID
	if err := r.db.WithContext(ctx).
		Model(&model.OrganizationMember{}).
		Where("user_id = ?", userID).
		Pluck("organization_id", &orgIDs).Error; err != nil {
		return nil, err
	}
	return orgIDs, nil
}

// IsUserOrgAdmin checks if a user is an admin of a specific organization (for auth.OrganizationChecker interface)
func (r *OrganizationRepository) IsUserOrgAdmin(ctx context.Context, userID, orgID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.OrganizationMember{}).
		Where("user_id = ? AND organization_id = ? AND role = ?", userID, orgID, model.OrgMemberRoleAdmin).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetOrgAdmins returns all admin members of an organization with their User data
func (r *OrganizationRepository) GetOrgAdmins(ctx context.Context, orgID uuid.UUID) ([]model.OrganizationMember, error) {
	var members []model.OrganizationMember
	if err := r.db.WithContext(ctx).
		Preload("User").
		Where("organization_id = ? AND role = ?", orgID, model.OrgMemberRoleAdmin).
		Find(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}

// ListByOrgIDs returns organizations filtered by IDs (for org-scoped access)
func (r *OrganizationRepository) ListByOrgIDs(ctx context.Context, filter OrganizationFilter, orgIDs []uuid.UUID) (*OrganizationListResult, error) {
	query := r.db.WithContext(ctx).Model(&model.Organization{}).Where("deleted_at IS NULL")

	// Filter by organization IDs
	if len(orgIDs) > 0 {
		query = query.Where("id IN ?", orgIDs)
	} else {
		// No organizations - return empty result
		return &OrganizationListResult{
			Organizations: []model.Organization{},
			Total:         0,
			Page:          filter.Page,
			PageSize:      filter.PageSize,
			TotalPages:    0,
		}, nil
	}

	// Apply search filter
	if filter.Search != "" {
		searchTerm := "%" + strings.ToLower(filter.Search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(slug) LIKE ?", searchTerm, searchTerm)
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

	// Fetch organizations
	var organizations []model.Organization
	if err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(filter.PageSize).
		Find(&organizations).Error; err != nil {
		return nil, err
	}

	// Calculate total pages
	totalPages := int(total) / filter.PageSize
	if int(total)%filter.PageSize > 0 {
		totalPages++
	}

	return &OrganizationListResult{
		Organizations: organizations,
		Total:         total,
		Page:          filter.Page,
		PageSize:      filter.PageSize,
		TotalPages:    totalPages,
	}, nil
}
