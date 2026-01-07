package repository

import (
	"github.com/google/uuid"
	"github.com/leksa/datamapper-senyar/internal/model"
	"gorm.io/gorm"
)

type FeedRepository struct {
	db *gorm.DB
}

func NewFeedRepository(db *gorm.DB) *FeedRepository {
	return &FeedRepository{db: db}
}

type FeedFilter struct {
	LocationID   string
	LocationName string
	Category     string
	Type         string
	Search       string
	Since        string // ISO date string for filtering feeds since a date
	Page         int
	Limit        int
}

type FeedWithCoords struct {
	model.Feed
	Longitude    *float64 `json:"longitude"`
	Latitude     *float64 `json:"latitude"`
	LocationName *string  `json:"location_name"`
	FaskesName   *string  `json:"faskes_name"`
}

// GetPhotosForFeed retrieves all photos for a specific feed
func (r *FeedRepository) GetPhotosForFeed(feedID uuid.UUID) ([]model.FeedPhoto, error) {
	var photos []model.FeedPhoto
	err := r.db.Where("feed_id = ?", feedID).Find(&photos).Error
	return photos, err
}

// GetPhotosForFeeds retrieves all photos for multiple feeds (batch query)
func (r *FeedRepository) GetPhotosForFeeds(feedIDs []uuid.UUID) (map[uuid.UUID][]model.FeedPhoto, error) {
	var photos []model.FeedPhoto
	err := r.db.Where("feed_id IN ?", feedIDs).Find(&photos).Error
	if err != nil {
		return nil, err
	}

	// Group photos by feed ID
	result := make(map[uuid.UUID][]model.FeedPhoto)
	for _, photo := range photos {
		result[photo.FeedID] = append(result[photo.FeedID], photo)
	}
	return result, nil
}

func (r *FeedRepository) FindAll(filter FeedFilter) ([]FeedWithCoords, int64, error) {
	var feeds []FeedWithCoords
	var total int64

	query := r.db.Table("information_feeds f").
		Select(`
			f.*,
			ST_X(f.geom) as longitude,
			ST_Y(f.geom) as latitude,
			l.nama as location_name,
			fk.nama as faskes_name
		`).
		Joins("LEFT JOIN locations l ON l.id = f.location_id").
		Joins("LEFT JOIN faskes fk ON fk.id = f.faskes_id")

	// Apply filters
	if filter.LocationID != "" {
		query = query.Where("f.location_id = ?", filter.LocationID)
	}
	if filter.LocationName != "" {
		query = query.Where("l.nama ILIKE ?", "%"+filter.LocationName+"%")
	}
	if filter.Category != "" {
		query = query.Where("f.category = ?", filter.Category)
	}
	if filter.Type != "" {
		query = query.Where("f.type = ?", filter.Type)
	}
	if filter.Search != "" {
		query = query.Where("f.content ILIKE ?", "%"+filter.Search+"%")
	}
	if filter.Since != "" {
		query = query.Where("COALESCE(f.submitted_at, f.created_at) >= ?", filter.Since)
	}

	// Count total
	countQuery := r.db.Table("information_feeds f").
		Joins("LEFT JOIN locations l ON l.id = f.location_id").
		Joins("LEFT JOIN faskes fk ON fk.id = f.faskes_id")
	if filter.LocationID != "" {
		countQuery = countQuery.Where("f.location_id = ?", filter.LocationID)
	}
	if filter.LocationName != "" {
		countQuery = countQuery.Where("l.nama ILIKE ?", "%"+filter.LocationName+"%")
	}
	if filter.Category != "" {
		countQuery = countQuery.Where("f.category = ?", filter.Category)
	}
	if filter.Type != "" {
		countQuery = countQuery.Where("f.type = ?", filter.Type)
	}
	if filter.Since != "" {
		countQuery = countQuery.Where("COALESCE(f.submitted_at, f.created_at) >= ?", filter.Since)
	}
	countQuery.Count(&total)

	// Pagination
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	offset := (filter.Page - 1) * filter.Limit
	query = query.Offset(offset).Limit(filter.Limit).Order("f.submitted_at DESC NULLS LAST, f.created_at DESC")

	err := query.Find(&feeds).Error
	return feeds, total, err
}

func (r *FeedRepository) FindByLocationID(locationID uuid.UUID, limit int) ([]FeedWithCoords, error) {
	var feeds []FeedWithCoords

	if limit <= 0 {
		limit = 5
	}

	err := r.db.Table("information_feeds f").
		Select(`
			f.*,
			ST_X(f.geom) as longitude,
			ST_Y(f.geom) as latitude
		`).
		Where("f.location_id = ?", locationID).
		Order("f.submitted_at DESC NULLS LAST").
		Limit(limit).
		Find(&feeds).Error

	return feeds, err
}
