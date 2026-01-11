package repository

import (
	"github.com/google/uuid"
	"github.com/leksa/datamapper-senyar/internal/model"
	"gorm.io/gorm"
)

type InfrastrukturRepository struct {
	db *gorm.DB
}

func NewInfrastrukturRepository(db *gorm.DB) *InfrastrukturRepository {
	return &InfrastrukturRepository{db: db}
}

type InfrastrukturFilter struct {
	Jenis            string // "Jalan" or "Jembatan"
	StatusJln        string // "Nasional" or "Daerah"
	StatusAkses      string // "dapat_diakses" or "akses_terputus"
	StatusPenanganan string
	NamaKabupaten    string
	Search           string
	MinLng           *float64
	MinLat           *float64
	MaxLng           *float64
	MaxLat           *float64
	Page             int
	Limit            int
}

type InfrastrukturWithCoords struct {
	model.Infrastruktur
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

func (r *InfrastrukturRepository) FindAll(filter InfrastrukturFilter) ([]InfrastrukturWithCoords, int64, error) {
	var items []InfrastrukturWithCoords
	var total int64

	// Base query with coordinates extraction
	query := r.db.Table("infrastruktur").
		Select(`
			infrastruktur.*,
			ST_X(geom) as longitude,
			ST_Y(geom) as latitude
		`).
		Where("deleted_at IS NULL")

	// Apply filters
	if filter.Jenis != "" {
		query = query.Where("jenis = ?", filter.Jenis)
	}
	if filter.StatusJln != "" {
		query = query.Where("status_jln = ?", filter.StatusJln)
	}
	if filter.StatusAkses != "" {
		query = query.Where("status_akses = ?", filter.StatusAkses)
	}
	if filter.StatusPenanganan != "" {
		query = query.Where("status_penanganan = ?", filter.StatusPenanganan)
	}
	if filter.NamaKabupaten != "" {
		query = query.Where("nama_kabupaten ILIKE ?", "%"+filter.NamaKabupaten+"%")
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
	countQuery := r.db.Table("infrastruktur").Where("deleted_at IS NULL")
	if filter.Jenis != "" {
		countQuery = countQuery.Where("jenis = ?", filter.Jenis)
	}
	if filter.StatusJln != "" {
		countQuery = countQuery.Where("status_jln = ?", filter.StatusJln)
	}
	if filter.StatusAkses != "" {
		countQuery = countQuery.Where("status_akses = ?", filter.StatusAkses)
	}
	if filter.StatusPenanganan != "" {
		countQuery = countQuery.Where("status_penanganan = ?", filter.StatusPenanganan)
	}
	if filter.NamaKabupaten != "" {
		countQuery = countQuery.Where("nama_kabupaten ILIKE ?", "%"+filter.NamaKabupaten+"%")
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
	if filter.Limit > 500 {
		filter.Limit = 500
	}

	offset := (filter.Page - 1) * filter.Limit
	query = query.Offset(offset).Limit(filter.Limit).Order("updated_at DESC")

	err := query.Find(&items).Error
	return items, total, err
}

func (r *InfrastrukturRepository) FindByID(id uuid.UUID) (*InfrastrukturWithCoords, error) {
	var item InfrastrukturWithCoords

	err := r.db.Table("infrastruktur").
		Select(`
			infrastruktur.*,
			ST_X(geom) as longitude,
			ST_Y(geom) as latitude
		`).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&item).Error

	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (r *InfrastrukturRepository) FindPhotos(infrastrukturID uuid.UUID) ([]model.InfrastrukturPhoto, error) {
	var photos []model.InfrastrukturPhoto
	err := r.db.Where("infrastruktur_id = ?", infrastrukturID).Find(&photos).Error
	return photos, err
}

// GetStats returns statistics about infrastructure
func (r *InfrastrukturRepository) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total by jenis
	var jenisStats []struct {
		Jenis string
		Count int64
	}
	r.db.Table("infrastruktur").
		Select("jenis, count(*) as count").
		Where("deleted_at IS NULL").
		Group("jenis").
		Scan(&jenisStats)
	stats["by_jenis"] = jenisStats

	// Total by status_akses
	var aksesStats []struct {
		StatusAkses string `gorm:"column:status_akses"`
		Count       int64
	}
	r.db.Table("infrastruktur").
		Select("status_akses, count(*) as count").
		Where("deleted_at IS NULL").
		Group("status_akses").
		Scan(&aksesStats)
	stats["by_status_akses"] = aksesStats

	// Total by status_penanganan
	var penangananStats []struct {
		StatusPenanganan string `gorm:"column:status_penanganan"`
		Count            int64
	}
	r.db.Table("infrastruktur").
		Select("status_penanganan, count(*) as count").
		Where("deleted_at IS NULL").
		Group("status_penanganan").
		Scan(&penangananStats)
	stats["by_status_penanganan"] = penangananStats

	// Average progress
	var avgProgress float64
	r.db.Table("infrastruktur").
		Select("COALESCE(AVG(progress), 0)").
		Where("deleted_at IS NULL").
		Scan(&avgProgress)
	stats["avg_progress"] = avgProgress

	return stats, nil
}
