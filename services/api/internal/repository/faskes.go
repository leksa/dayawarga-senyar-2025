package repository

import (
	"github.com/google/uuid"
	"github.com/leksa/datamapper-senyar/internal/model"
	"gorm.io/gorm"
)

type FaskesRepository struct {
	db *gorm.DB
}

func NewFaskesRepository(db *gorm.DB) *FaskesRepository {
	return &FaskesRepository{db: db}
}

type FaskesFilter struct {
	JenisFaskes   string
	StatusFaskes  string
	KondisiFaskes string
	Search        string
	MinLng        *float64
	MinLat        *float64
	MaxLng        *float64
	MaxLat        *float64
	Page          int
	Limit         int
}

type FaskesWithCoords struct {
	model.Faskes
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

func (r *FaskesRepository) FindAll(filter FaskesFilter) ([]FaskesWithCoords, int64, error) {
	var faskesList []FaskesWithCoords
	var total int64

	// Base query with coordinates extraction
	query := r.db.Table("faskes").
		Select(`
			faskes.*,
			ST_X(geom) as longitude,
			ST_Y(geom) as latitude
		`).
		Where("deleted_at IS NULL")

	// Apply filters
	if filter.JenisFaskes != "" {
		query = query.Where("jenis_faskes = ?", filter.JenisFaskes)
	}
	if filter.StatusFaskes != "" {
		query = query.Where("status_faskes = ?", filter.StatusFaskes)
	}
	if filter.KondisiFaskes != "" {
		query = query.Where("kondisi_faskes = ?", filter.KondisiFaskes)
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
	countQuery := r.db.Table("faskes").Where("deleted_at IS NULL")
	if filter.JenisFaskes != "" {
		countQuery = countQuery.Where("jenis_faskes = ?", filter.JenisFaskes)
	}
	if filter.StatusFaskes != "" {
		countQuery = countQuery.Where("status_faskes = ?", filter.StatusFaskes)
	}
	if filter.KondisiFaskes != "" {
		countQuery = countQuery.Where("kondisi_faskes = ?", filter.KondisiFaskes)
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

	err := query.Find(&faskesList).Error
	return faskesList, total, err
}

func (r *FaskesRepository) FindByID(id uuid.UUID) (*FaskesWithCoords, error) {
	var faskes FaskesWithCoords

	err := r.db.Table("faskes").
		Select(`
			faskes.*,
			ST_X(geom) as longitude,
			ST_Y(geom) as latitude
		`).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&faskes).Error

	if err != nil {
		return nil, err
	}

	return &faskes, nil
}

func (r *FaskesRepository) FindPhotos(faskesID uuid.UUID) ([]model.FaskesPhoto, error) {
	var photos []model.FaskesPhoto
	err := r.db.Where("faskes_id = ?", faskesID).Find(&photos).Error
	return photos, err
}
