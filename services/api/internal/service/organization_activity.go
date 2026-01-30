package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/leksa/datamapper-senyar/internal/model"
	"github.com/leksa/datamapper-senyar/internal/repository"
	"gorm.io/gorm"
)

// OrganizationActivityService aggregates feeds from organization's relawan
type OrganizationActivityService struct {
	relawanRepo *repository.RelawanRepository
	feedRepo    *repository.FeedRepository
	db          *gorm.DB
}

// NewOrganizationActivityService creates a new organization activity service
func NewOrganizationActivityService(
	relawanRepo *repository.RelawanRepository,
	feedRepo *repository.FeedRepository,
	db *gorm.DB,
) *OrganizationActivityService {
	return &OrganizationActivityService{
		relawanRepo: relawanRepo,
		feedRepo:    feedRepo,
		db:          db,
	}
}

// OrganizationActivity represents a feed with relawan information
type OrganizationActivity struct {
	Feed        model.Feed `json:"feed"`
	RelawanName string     `json:"relawan_name,omitempty"`
}

// GetOrganizationActivities returns paginated feeds from organization's relawan
func (s *OrganizationActivityService) GetOrganizationActivities(
	ctx context.Context,
	orgID uuid.UUID,
	page, pageSize int,
) ([]OrganizationActivity, int64, error) {
	relawan, err := s.relawanRepo.GetByOrganization(ctx, orgID)
	if err != nil {
		return nil, 0, err
	}

	relawanNames := make([]string, 0, len(relawan))
	for _, r := range relawan {
		relawanNames = append(relawanNames, r.Name)
	}

	if len(relawanNames) == 0 {
		return []OrganizationActivity{}, 0, nil
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var feeds []model.Feed
	var total int64

	query := s.db.WithContext(ctx).
		Where("username IN ?", relawanNames).
		Order("created_at DESC")

	if err := query.Model(&model.Feed{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize

	if err := query.
		Preload("Photos").
		Offset(offset).
		Limit(pageSize).
		Find(&feeds).Error; err != nil {
		return nil, 0, err
	}

	activities := make([]OrganizationActivity, 0, len(feeds))
	for _, feed := range feeds {
		activity := OrganizationActivity{
			Feed: feed,
		}
		if feed.Username != nil {
			activity.RelawanName = *feed.Username
		}
		activities = append(activities, activity)
	}

	return activities, total, nil
}
