package repository

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/leksa/datamapper-senyar/internal/model"
	"gorm.io/gorm"
)

// RelawanRepository handles database operations for relawan
type RelawanRepository struct {
	db *gorm.DB
}

// NewRelawanRepository creates a new relawan repository
func NewRelawanRepository(db *gorm.DB) *RelawanRepository {
	return &RelawanRepository{db: db}
}

// RelawanFilter defines filters for querying relawan
type RelawanFilter struct {
	OrganizationID *uuid.UUID
	GroupID        *uuid.UUID
	Status         *model.RelawanStatus
	Search         string
	HasODKAccess   *bool
	Page           int
	PageSize       int
}

// RelawanListResult contains paginated relawan results
type RelawanListResult struct {
	Relawan    []model.Relawan `json:"relawan"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int             `json:"total_pages"`
}

// List returns paginated relawan with optional filters
func (r *RelawanRepository) List(ctx context.Context, filter RelawanFilter) (*RelawanListResult, error) {
	query := r.db.WithContext(ctx).Model(&model.Relawan{}).Where("deleted_at IS NULL")

	// Apply organization filter
	if filter.OrganizationID != nil {
		query = query.Where("organization_id = ?", *filter.OrganizationID)
	}

	// Apply group filter
	if filter.GroupID != nil {
		query = query.Where("group_id = ?", *filter.GroupID)
	}

	// Apply status filter
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}

	// Apply search filter
	if filter.Search != "" {
		searchTerm := "%" + strings.ToLower(filter.Search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR phone LIKE ? OR LOWER(email) LIKE ?", searchTerm, searchTerm, searchTerm)
	}

	// Apply ODK access filter
	if filter.HasODKAccess != nil {
		if *filter.HasODKAccess {
			query = query.Where("odk_app_user_id IS NOT NULL")
		} else {
			query = query.Where("odk_app_user_id IS NULL")
		}
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

	// Fetch relawan with relations
	var relawan []model.Relawan
	if err := query.
		Preload("Organization").
		Preload("Group").
		Order("created_at DESC").
		Offset(offset).
		Limit(filter.PageSize).
		Find(&relawan).Error; err != nil {
		return nil, err
	}

	// Calculate total pages
	totalPages := int(total) / filter.PageSize
	if int(total)%filter.PageSize > 0 {
		totalPages++
	}

	return &RelawanListResult{
		Relawan:    relawan,
		Total:      total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: totalPages,
	}, nil
}

// FindByID retrieves a relawan by ID
func (r *RelawanRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Relawan, error) {
	var relawan model.Relawan
	if err := r.db.WithContext(ctx).
		Preload("Organization").
		Preload("Group").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&relawan).Error; err != nil {
		return nil, err
	}
	return &relawan, nil
}

// FindByPhone retrieves a relawan by phone number within an organization
func (r *RelawanRepository) FindByPhone(ctx context.Context, orgID uuid.UUID, phone string) (*model.Relawan, error) {
	var relawan model.Relawan
	if err := r.db.WithContext(ctx).
		Where("organization_id = ? AND phone = ? AND deleted_at IS NULL", orgID, phone).
		First(&relawan).Error; err != nil {
		return nil, err
	}
	return &relawan, nil
}

// Create creates a new relawan
func (r *RelawanRepository) Create(ctx context.Context, relawan *model.Relawan) error {
	return r.db.WithContext(ctx).Create(relawan).Error
}

// Update updates an existing relawan
func (r *RelawanRepository) Update(ctx context.Context, relawan *model.Relawan) error {
	return r.db.WithContext(ctx).Save(relawan).Error
}

// Delete soft-deletes a relawan
func (r *RelawanRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&model.Relawan{}).
		Where("id = ?", id).
		Update("deleted_at", gorm.Expr("NOW()")).Error
}

// UpdateStatus updates a relawan's status
func (r *RelawanRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.RelawanStatus) error {
	return r.db.WithContext(ctx).
		Model(&model.Relawan{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// UpdateGroup moves a relawan to a different group
func (r *RelawanRepository) UpdateGroup(ctx context.Context, id uuid.UUID, groupID *uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&model.Relawan{}).
		Where("id = ?", id).
		Update("group_id", groupID).Error
}

// UpdateODKAppUser updates ODK app user information
func (r *RelawanRepository) UpdateODKAppUser(ctx context.Context, id uuid.UUID, odkUserID *int, odkToken *string) error {
	updates := map[string]interface{}{
		"odk_app_user_id":    odkUserID,
		"odk_app_user_token": odkToken,
	}
	if odkUserID != nil {
		updates["odk_app_user_created_at"] = gorm.Expr("NOW()")
	}
	return r.db.WithContext(ctx).
		Model(&model.Relawan{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// GetByOrganization returns all relawan for an organization
func (r *RelawanRepository) GetByOrganization(ctx context.Context, orgID uuid.UUID) ([]model.Relawan, error) {
	var relawan []model.Relawan
	if err := r.db.WithContext(ctx).
		Preload("Group").
		Where("organization_id = ? AND deleted_at IS NULL", orgID).
		Order("name ASC").
		Find(&relawan).Error; err != nil {
		return nil, err
	}
	return relawan, nil
}

// GetByGroup returns all relawan for a group
func (r *RelawanRepository) GetByGroup(ctx context.Context, groupID uuid.UUID) ([]model.Relawan, error) {
	var relawan []model.Relawan
	if err := r.db.WithContext(ctx).
		Where("group_id = ? AND deleted_at IS NULL", groupID).
		Order("name ASC").
		Find(&relawan).Error; err != nil {
		return nil, err
	}
	return relawan, nil
}

// PhoneExistsInOrg checks if a phone number already exists in an organization
func (r *RelawanRepository) PhoneExistsInOrg(ctx context.Context, orgID uuid.UUID, phone string, excludeID *uuid.UUID) (bool, error) {
	query := r.db.WithContext(ctx).Model(&model.Relawan{}).
		Where("organization_id = ? AND phone = ? AND deleted_at IS NULL", orgID, phone)
	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// CountByOrganization returns the count of relawan in an organization
func (r *RelawanRepository) CountByOrganization(ctx context.Context, orgID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Relawan{}).
		Where("organization_id = ? AND deleted_at IS NULL", orgID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// CountByGroup returns the count of relawan in a group
func (r *RelawanRepository) CountByGroup(ctx context.Context, groupID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Relawan{}).
		Where("group_id = ? AND deleted_at IS NULL", groupID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// BulkUpdateGroup updates group for multiple relawan
func (r *RelawanRepository) BulkUpdateGroup(ctx context.Context, ids []uuid.UUID, groupID *uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&model.Relawan{}).
		Where("id IN ?", ids).
		Update("group_id", groupID).Error
}

// GetStats returns overall statistics for relawan
func (r *RelawanRepository) GetStats(ctx context.Context, orgID *uuid.UUID) (map[string]int64, error) {
	stats := make(map[string]int64)
	query := r.db.WithContext(ctx).Model(&model.Relawan{}).Where("deleted_at IS NULL")

	if orgID != nil {
		query = query.Where("organization_id = ?", *orgID)
	}

	// Total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	stats["total"] = total

	// Active count
	var activeCount int64
	activeQuery := r.db.WithContext(ctx).Model(&model.Relawan{}).
		Where("deleted_at IS NULL AND status = ?", model.RelawanStatusActive)
	if orgID != nil {
		activeQuery = activeQuery.Where("organization_id = ?", *orgID)
	}
	if err := activeQuery.Count(&activeCount).Error; err != nil {
		return nil, err
	}
	stats["active"] = activeCount

	// With ODK access
	var odkCount int64
	odkQuery := r.db.WithContext(ctx).Model(&model.Relawan{}).
		Where("deleted_at IS NULL AND odk_app_user_id IS NOT NULL")
	if orgID != nil {
		odkQuery = odkQuery.Where("organization_id = ?", *orgID)
	}
	if err := odkQuery.Count(&odkCount).Error; err != nil {
		return nil, err
	}
	stats["with_odk_access"] = odkCount

	return stats, nil
}

// ListByOrgIDs returns paginated relawan filtered by organization IDs (for org-scoped access)
func (r *RelawanRepository) ListByOrgIDs(ctx context.Context, filter RelawanFilter, orgIDs []uuid.UUID) (*RelawanListResult, error) {
	query := r.db.WithContext(ctx).Model(&model.Relawan{}).Where("deleted_at IS NULL")

	// Apply organization IDs filter
	if len(orgIDs) > 0 {
		query = query.Where("organization_id IN ?", orgIDs)
	}

	// Apply specific organization filter (if user also filters by org_id)
	if filter.OrganizationID != nil {
		query = query.Where("organization_id = ?", *filter.OrganizationID)
	}

	// Apply group filter
	if filter.GroupID != nil {
		query = query.Where("group_id = ?", *filter.GroupID)
	}

	// Apply status filter
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}

	// Apply search filter
	if filter.Search != "" {
		searchTerm := "%" + strings.ToLower(filter.Search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR phone LIKE ? OR LOWER(email) LIKE ?", searchTerm, searchTerm, searchTerm)
	}

	// Apply ODK access filter
	if filter.HasODKAccess != nil {
		if *filter.HasODKAccess {
			query = query.Where("odk_app_user_id IS NOT NULL")
		} else {
			query = query.Where("odk_app_user_id IS NULL")
		}
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

	// Fetch relawan with relations
	var relawan []model.Relawan
	if err := query.
		Preload("Organization").
		Preload("Group").
		Order("created_at DESC").
		Offset(offset).
		Limit(filter.PageSize).
		Find(&relawan).Error; err != nil {
		return nil, err
	}

	// Calculate total pages
	totalPages := int(total) / filter.PageSize
	if int(total)%filter.PageSize > 0 {
		totalPages++
	}

	return &RelawanListResult{
		Relawan:    relawan,
		Total:      total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: totalPages,
	}, nil
}

// SetWAVerified enables or disables WhatsApp verification for a relawan
func (r *RelawanRepository) SetWAVerified(ctx context.Context, id uuid.UUID, verified bool) error {
	updates := map[string]interface{}{
		"wa_verified": verified,
	}
	if verified {
		updates["wa_verified_at"] = gorm.Expr("NOW()")
	} else {
		updates["wa_verified_at"] = nil
	}
	return r.db.WithContext(ctx).
		Model(&model.Relawan{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// UpdateWAActivity updates the last WhatsApp activity timestamp and increments session count
func (r *RelawanRepository) UpdateWAActivity(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&model.Relawan{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"wa_last_activity": gorm.Expr("NOW()"),
			"wa_session_count": gorm.Expr("wa_session_count + 1"),
		}).Error
}

// FindByPhoneWithWAAccess finds a verified relawan by phone number (for chatbot validation)
func (r *RelawanRepository) FindByPhoneWithWAAccess(ctx context.Context, phone string) (*model.Relawan, error) {
	var relawan model.Relawan
	if err := r.db.WithContext(ctx).
		Where("phone = ? AND wa_verified = true AND status = ? AND deleted_at IS NULL", phone, model.RelawanStatusActive).
		First(&relawan).Error; err != nil {
		return nil, err
	}
	return &relawan, nil
}

// GetStatsByOrgIDs returns overall statistics for relawan filtered by organization IDs
func (r *RelawanRepository) GetStatsByOrgIDs(ctx context.Context, orgIDs []uuid.UUID) (map[string]int64, error) {
	stats := make(map[string]int64)
	query := r.db.WithContext(ctx).Model(&model.Relawan{}).Where("deleted_at IS NULL")

	if len(orgIDs) > 0 {
		query = query.Where("organization_id IN ?", orgIDs)
	}

	// Total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	stats["total"] = total

	// Active count
	var activeCount int64
	activeQuery := r.db.WithContext(ctx).Model(&model.Relawan{}).
		Where("deleted_at IS NULL AND status = ?", model.RelawanStatusActive)
	if len(orgIDs) > 0 {
		activeQuery = activeQuery.Where("organization_id IN ?", orgIDs)
	}
	if err := activeQuery.Count(&activeCount).Error; err != nil {
		return nil, err
	}
	stats["active"] = activeCount

	// With ODK access
	var odkCount int64
	odkQuery := r.db.WithContext(ctx).Model(&model.Relawan{}).
		Where("deleted_at IS NULL AND odk_app_user_id IS NOT NULL")
	if len(orgIDs) > 0 {
		odkQuery = odkQuery.Where("organization_id IN ?", orgIDs)
	}
	if err := odkQuery.Count(&odkCount).Error; err != nil {
		return nil, err
	}
	stats["with_odk_access"] = odkCount

	return stats, nil
}
