package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/leksa/datamapper-senyar/internal/model"
)

// ProjectRequestRepository handles project request database operations
type ProjectRequestRepository struct {
	db *gorm.DB
}

// NewProjectRequestRepository creates a new project request repository
func NewProjectRequestRepository(db *gorm.DB) *ProjectRequestRepository {
	return &ProjectRequestRepository{db: db}
}

// Create creates a new project request
func (r *ProjectRequestRepository) Create(ctx context.Context, request *model.ProjectRequest) error {
	return r.db.WithContext(ctx).Create(request).Error
}

// GetByID retrieves a project request by ID
func (r *ProjectRequestRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.ProjectRequest, error) {
	var request model.ProjectRequest
	err := r.db.WithContext(ctx).
		Preload("Organization").
		Preload("Group").
		Preload("Group.Leader").
		Preload("Requester").
		Preload("Reviewer").
		First(&request, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &request, nil
}

// ProjectRequestFilter defines filter options for listing project requests
type ProjectRequestFilter struct {
	OrganizationID *uuid.UUID
	GroupID        *uuid.UUID
	Status         *model.ProjectRequestStatus
	RequestedBy    *uuid.UUID
}

// List retrieves project requests with optional filters
func (r *ProjectRequestRepository) List(ctx context.Context, filter ProjectRequestFilter, page, pageSize int) ([]model.ProjectRequest, int64, error) {
	var requests []model.ProjectRequest
	var total int64

	query := r.db.WithContext(ctx).Model(&model.ProjectRequest{})

	if filter.OrganizationID != nil {
		query = query.Where("organization_id = ?", *filter.OrganizationID)
	}
	if filter.GroupID != nil {
		query = query.Where("group_id = ?", *filter.GroupID)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.RequestedBy != nil {
		query = query.Where("requested_by = ?", *filter.RequestedBy)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	err := query.
		Preload("Organization").
		Preload("Group").
		Preload("Group.Leader").
		Preload("Requester").
		Preload("Reviewer").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&requests).Error

	return requests, total, err
}

// ListPending retrieves all pending project requests
func (r *ProjectRequestRepository) ListPending(ctx context.Context, page, pageSize int) ([]model.ProjectRequest, int64, error) {
	status := model.ProjectRequestStatusPending
	return r.List(ctx, ProjectRequestFilter{Status: &status}, page, pageSize)
}

// Update updates a project request
func (r *ProjectRequestRepository) Update(ctx context.Context, request *model.ProjectRequest) error {
	return r.db.WithContext(ctx).Save(request).Error
}

// Approve approves a project request
func (r *ProjectRequestRepository) Approve(ctx context.Context, id uuid.UUID, reviewerID uuid.UUID, notes *string) error {
	return r.db.WithContext(ctx).Model(&model.ProjectRequest{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       model.ProjectRequestStatusApproved,
			"reviewed_by":  reviewerID,
			"reviewed_at":  gorm.Expr("NOW()"),
			"review_notes": notes,
		}).Error
}

// Reject rejects a project request
func (r *ProjectRequestRepository) Reject(ctx context.Context, id uuid.UUID, reviewerID uuid.UUID, notes *string) error {
	return r.db.WithContext(ctx).Model(&model.ProjectRequest{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       model.ProjectRequestStatusRejected,
			"reviewed_by":  reviewerID,
			"reviewed_at":  gorm.Expr("NOW()"),
			"review_notes": notes,
		}).Error
}

// ExistsPendingForGroup checks if there's already a pending request for a group
func (r *ProjectRequestRepository) ExistsPendingForGroup(ctx context.Context, groupID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.ProjectRequest{}).
		Where("group_id = ? AND status = ?", groupID, model.ProjectRequestStatusPending).
		Count(&count).Error
	return count > 0, err
}

// GroupProjectRepository handles group project database operations
type GroupProjectRepository struct {
	db *gorm.DB
}

// NewGroupProjectRepository creates a new group project repository
func NewGroupProjectRepository(db *gorm.DB) *GroupProjectRepository {
	return &GroupProjectRepository{db: db}
}

// Create creates a new group project assignment
func (r *GroupProjectRepository) Create(ctx context.Context, gp *model.GroupProject) error {
	return r.db.WithContext(ctx).Create(gp).Error
}

// GetByGroupID retrieves all project assignments for a group
func (r *GroupProjectRepository) GetByGroupID(ctx context.Context, groupID uuid.UUID) ([]model.GroupProject, error) {
	var projects []model.GroupProject
	err := r.db.WithContext(ctx).
		Where("group_id = ?", groupID).
		Find(&projects).Error
	return projects, err
}

// GetByODKProjectID retrieves all groups assigned to an ODK project
func (r *GroupProjectRepository) GetByODKProjectID(ctx context.Context, odkProjectID int) ([]model.GroupProject, error) {
	var projects []model.GroupProject
	err := r.db.WithContext(ctx).
		Preload("Group").
		Where("odk_project_id = ?", odkProjectID).
		Find(&projects).Error
	return projects, err
}

// Exists checks if a group-project assignment exists
func (r *GroupProjectRepository) Exists(ctx context.Context, groupID uuid.UUID, odkProjectID int) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.GroupProject{}).
		Where("group_id = ? AND odk_project_id = ?", groupID, odkProjectID).
		Count(&count).Error
	return count > 0, err
}

// Delete removes a group-project assignment
func (r *GroupProjectRepository) Delete(ctx context.Context, groupID uuid.UUID, odkProjectID int) error {
	return r.db.WithContext(ctx).
		Where("group_id = ? AND odk_project_id = ?", groupID, odkProjectID).
		Delete(&model.GroupProject{}).Error
}
