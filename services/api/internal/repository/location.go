package repository

import (
	"github.com/google/uuid"
	"github.com/leksa/datamapper-senyar/internal/model"
	"gorm.io/gorm"
)

type LocationRepository struct {
	db *gorm.DB
}

func NewLocationRepository(db *gorm.DB) *LocationRepository {
	return &LocationRepository{db: db}
}

type LocationFilter struct {
	Type   string
	Status string
	Search string
	MinLng *float64
	MinLat *float64
	MaxLng *float64
	MaxLat *float64
	Page   int
	Limit  int
}

type LocationWithCoords struct {
	model.Location
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

func (r *LocationRepository) FindAll(filter LocationFilter) ([]LocationWithCoords, int64, error) {
	var locations []LocationWithCoords
	var total int64

	// Base query with coordinates extraction
	query := r.db.Table("locations").
		Select(`
			locations.*,
			ST_X(geom) as longitude,
			ST_Y(geom) as latitude
		`).
		Where("deleted_at IS NULL")

	// Apply filters
	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Search != "" {
		query = query.Where("nama ILIKE ?", "%"+filter.Search+"%")
	}

	// Bounding box filter
	if filter.MinLng != nil && filter.MinLat != nil && filter.MaxLng != nil && filter.MaxLat != nil {
		query = query.Where(`
			ST_Within(
				geom,
				ST_MakeEnvelope(?, ?, ?, ?, 4326)
			)
		`, *filter.MinLng, *filter.MinLat, *filter.MaxLng, *filter.MaxLat)
	}

	// Count total
	countQuery := r.db.Table("locations").Where("deleted_at IS NULL")
	if filter.Type != "" {
		countQuery = countQuery.Where("type = ?", filter.Type)
	}
	if filter.Status != "" {
		countQuery = countQuery.Where("status = ?", filter.Status)
	}
	if filter.Search != "" {
		countQuery = countQuery.Where("nama ILIKE ?", "%"+filter.Search+"%")
	}
	countQuery.Count(&total)

	// Pagination
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Limit > 200 {
		filter.Limit = 200
	}

	offset := (filter.Page - 1) * filter.Limit
	query = query.Offset(offset).Limit(filter.Limit).Order("updated_at DESC")

	err := query.Find(&locations).Error
	return locations, total, err
}

func (r *LocationRepository) FindByID(id uuid.UUID) (*LocationWithCoords, error) {
	var location LocationWithCoords

	err := r.db.Table("locations").
		Select(`
			locations.*,
			ST_X(geom) as longitude,
			ST_Y(geom) as latitude
		`).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&location).Error

	if err != nil {
		return nil, err
	}

	return &location, nil
}

func (r *LocationRepository) FindPhotos(locationID uuid.UUID) ([]model.LocationPhoto, error) {
	var photos []model.LocationPhoto
	err := r.db.Where("location_id = ?", locationID).Find(&photos).Error
	return photos, err
}
