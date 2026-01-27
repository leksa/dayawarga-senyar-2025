package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/leksa/datamapper-senyar/internal/dto"
	"github.com/leksa/datamapper-senyar/internal/model"
	"github.com/leksa/datamapper-senyar/internal/repository"
	"github.com/leksa/datamapper-senyar/internal/validator"
)

type LocationHandler struct {
	locationRepo *repository.LocationRepository
	feedRepo     *repository.FeedRepository
}

func NewLocationHandler(locationRepo *repository.LocationRepository, feedRepo *repository.FeedRepository) *LocationHandler {
	return &LocationHandler{
		locationRepo: locationRepo,
		feedRepo:     feedRepo,
	}
}

// GetLocations returns GeoJSON FeatureCollection of locations
// @Summary      Get all locations
// @Description  Returns GeoJSON FeatureCollection of locations with optional filtering and pagination
// @Tags         locations
// @Accept       json
// @Produce      json
// @Param        page   query     int    false  "Page number (min: 1)" default(1)
// @Param        limit  query     int    false  "Items per page (min: 1, max: 100)" default(10)
// @Param        type   query     string false  "Filter by location type" Enums(posko, faskes, jembatan)
// @Param        status query     string false  "Filter by status" Enums(active, inactive)
// @Param        search query     string false  "Search in location name"
// @Param        bbox   query     string false  "Bounding box filter: minLng,minLat,maxLng,maxLat"
// @Success      200    {object} dto.LocationListResponse
// @Failure      400    {object} dto.APIResponse
// @Failure      500    {object} dto.APIResponse
// @Router       /api/v1/locations [get]
func (h *LocationHandler) GetLocations(c *gin.Context) {
	// Validate pagination parameters
	pagination := validator.ValidatePagination(c)
	if !pagination.Valid {
		return
	}

	// Validate bounding box if provided
	bbox := validator.ValidateBBox(c)

	filter := repository.LocationFilter{
		Type:   c.Query("type"),
		Status: c.Query("status"),
		Search: validator.SanitizeSearchString(c.Query("search")),
		Page:   pagination.Page,
		Limit:  pagination.Limit,
	}

	// Apply bounding box filter if valid
	if bbox.Valid {
		filter.MinLng = &bbox.MinLng
		filter.MinLat = &bbox.MinLat
		filter.MaxLng = &bbox.MaxLng
		filter.MaxLat = &bbox.MaxLat
	}

	locations, total, err := h.locationRepo.FindAll(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to fetch locations",
			},
		})
		return
	}

	// Convert to GeoJSON
	features := make([]dto.LocationFeatureResponse, len(locations))
	for i, loc := range locations {
		// Build alamat singkat and extract region fields
		alamatSingkat := ""
		namaProvinsi := ""
		namaKotaKab := ""
		namaKecamatan := ""
		namaDesa := ""
		idProvinsi := ""
		idKotaKab := ""
		idKecamatan := ""
		idDesa := ""
		if loc.Alamat != nil {
			parts := []string{}
			// Check both "nama_desa" and "desa" keys
			if desa, ok := loc.Alamat["nama_desa"].(string); ok && desa != "" {
				parts = append(parts, desa)
				namaDesa = desa
			} else if desa, ok := loc.Alamat["desa"].(string); ok && desa != "" {
				parts = append(parts, desa)
				namaDesa = desa
			}
			// Check both "nama_kota_kab" and "kabupaten" keys
			if kab, ok := loc.Alamat["nama_kota_kab"].(string); ok && kab != "" {
				parts = append(parts, kab)
				namaKotaKab = kab
			} else if kab, ok := loc.Alamat["kabupaten"].(string); ok && kab != "" {
				parts = append(parts, kab)
				namaKotaKab = kab
			}
			// Check both "nama_kecamatan" and "kecamatan" keys
			if kec, ok := loc.Alamat["nama_kecamatan"].(string); ok && kec != "" {
				namaKecamatan = kec
			} else if kec, ok := loc.Alamat["kecamatan"].(string); ok && kec != "" {
				namaKecamatan = kec
			}
			// Check both "nama_provinsi" and "provinsi" keys
			if prov, ok := loc.Alamat["nama_provinsi"].(string); ok && prov != "" {
				namaProvinsi = prov
			} else if prov, ok := loc.Alamat["provinsi"].(string); ok && prov != "" {
				namaProvinsi = prov
			}
			// Extract ID wilayah fields
			if id, ok := loc.Alamat["id_provinsi"].(string); ok && id != "" {
				idProvinsi = id
			}
			if id, ok := loc.Alamat["id_kota_kab"].(string); ok && id != "" {
				idKotaKab = id
			}
			if id, ok := loc.Alamat["id_kecamatan"].(string); ok && id != "" {
				idKecamatan = id
			}
			if id, ok := loc.Alamat["id_desa"].(string); ok && id != "" {
				idDesa = id
			}
			alamatSingkat = strings.Join(parts, ", ")
		}

		// Get jumlah_kk and total_jiwa from data_pengungsi
		jumlahKK := 0
		totalJiwa := 0
		jumlahPerempuan := 0
		jumlahLaki := 0
		jumlahBalita := 0
		if loc.DataPengungsi != nil {
			if v, ok := loc.DataPengungsi["jumlah_kk"].(float64); ok {
				jumlahKK = int(v)
			}
			if v, ok := loc.DataPengungsi["total_jiwa"].(float64); ok {
				totalJiwa = int(v)
			}
			// Sum all female categories: dewasa_perempuan, remaja_perempuan, anak_perempuan, balita_perempuan, bayi_perempuan
			if v, ok := loc.DataPengungsi["dewasa_perempuan"].(float64); ok {
				jumlahPerempuan += int(v)
			}
			if v, ok := loc.DataPengungsi["remaja_perempuan"].(float64); ok {
				jumlahPerempuan += int(v)
			}
			if v, ok := loc.DataPengungsi["anak_perempuan"].(float64); ok {
				jumlahPerempuan += int(v)
			}
			if v, ok := loc.DataPengungsi["balita_perempuan"].(float64); ok {
				jumlahPerempuan += int(v)
			}
			if v, ok := loc.DataPengungsi["bayi_perempuan"].(float64); ok {
				jumlahPerempuan += int(v)
			}
			// Sum all male categories: dewasa_laki, remaja_laki, anak_laki, balita_laki, bayi_laki
			if v, ok := loc.DataPengungsi["dewasa_laki"].(float64); ok {
				jumlahLaki += int(v)
			}
			if v, ok := loc.DataPengungsi["remaja_laki"].(float64); ok {
				jumlahLaki += int(v)
			}
			if v, ok := loc.DataPengungsi["anak_laki"].(float64); ok {
				jumlahLaki += int(v)
			}
			if v, ok := loc.DataPengungsi["balita_laki"].(float64); ok {
				jumlahLaki += int(v)
			}
			if v, ok := loc.DataPengungsi["bayi_laki"].(float64); ok {
				jumlahLaki += int(v)
			}
			// Sum balita: balita_perempuan + balita_laki + bayi_perempuan + bayi_laki
			if v, ok := loc.DataPengungsi["balita_perempuan"].(float64); ok {
				jumlahBalita += int(v)
			}
			if v, ok := loc.DataPengungsi["balita_laki"].(float64); ok {
				jumlahBalita += int(v)
			}
			if v, ok := loc.DataPengungsi["bayi_perempuan"].(float64); ok {
				jumlahBalita += int(v)
			}
			if v, ok := loc.DataPengungsi["bayi_laki"].(float64); ok {
				jumlahBalita += int(v)
			}
		}

		// Get kebutuhan_air from fasilitas
		kebutuhanAir := ""
		kebutuhanAirLiter := 0
		if loc.Fasilitas != nil {
			if v, ok := loc.Fasilitas["ketersediaan_air"].(string); ok {
				kebutuhanAir = v
			}
			if v, ok := loc.Fasilitas["kebutuhan_air"].(float64); ok {
				kebutuhanAirLiter = int(v)
			}
		}

		odkSubmissionID := ""
		if loc.ODKSubmissionID != nil {
			odkSubmissionID = *loc.ODKSubmissionID
		}

		// Get baseline_sumber - prefer dedicated column, fallback to identitas JSONB
		baselineSumber := loc.BaselineSumber
		if baselineSumber == "" && loc.Identitas != nil {
			if v, ok := loc.Identitas["baseline_sumber"].(string); ok {
				baselineSumber = v
			}
		}

		features[i] = dto.LocationFeatureResponse{
			Type: "Feature",
			ID:   loc.ID.String(),
			Geometry: &dto.GeoJSONGeometry{
				Type:        "Point",
				Coordinates: []float64{loc.Longitude, loc.Latitude},
			},
			Properties: dto.LocationListProperties{
				ODKSubmissionID:   odkSubmissionID,
				Nama:              loc.Nama,
				Type:              loc.Type,
				Status:            loc.Status,
				AlamatSingkat:     alamatSingkat,
				NamaProvinsi:      namaProvinsi,
				NamaKotaKab:       namaKotaKab,
				NamaKecamatan:     namaKecamatan,
				NamaDesa:          namaDesa,
				IDProvinsi:        idProvinsi,
				IDKotaKab:         idKotaKab,
				IDKecamatan:       idKecamatan,
				IDDesa:            idDesa,
				JumlahKK:          jumlahKK,
				TotalJiwa:         totalJiwa,
				JumlahPerempuan:   jumlahPerempuan,
				JumlahLaki:        jumlahLaki,
				JumlahBalita:      jumlahBalita,
				KebutuhanAir:      kebutuhanAir,
				KebutuhanAirLiter: kebutuhanAirLiter,
				BaselineSumber:    baselineSumber,
				UpdatedAt:         loc.UpdatedAt,
			},
		}
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Data: dto.LocationListResponse{
			Type:     "FeatureCollection",
			Features: features,
		},
		Meta: &dto.MetaInfo{
			Total:     total,
			Page:      filter.Page,
			Limit:     filter.Limit,
			Timestamp: time.Now(),
		},
	})
}

// GetLocationByID returns detailed location info
// @Summary      Get location by ID
// @Description  Returns detailed information about a specific location including photos
// @Tags         locations
// @Accept       json
// @Produce      json
// @Param        id   path      string true  "Location UUID"
// @Success      200   {object} dto.LocationDetailResponse
// @Failure      400   {object} dto.APIResponse
// @Failure      404   {object} dto.APIResponse
// @Failure      500   {object} dto.APIResponse
// @Router       /api/v1/locations/{id} [get]
func (h *LocationHandler) GetLocationByID(c *gin.Context) {
	idStr, ok := validator.ValidateUUID(c, "id", "Location ID")
	if !ok {
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "VALIDATION_ERROR",
				Message: "Invalid location ID format",
			},
		})
		return
	}

	location, err := h.locationRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.APIResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "NOT_FOUND",
				Message: "Location not found",
			},
		})
		return
	}

	// Get photos
	photos, err := h.locationRepo.FindPhotos(id)
	if err != nil {
		// Log error but continue with empty photos list
		// This is a non-critical error, we still want to return location data
		photos = []model.LocationPhoto{}
	}
	photoResponses := make([]dto.PhotoResponse, len(photos))
	for i, p := range photos {
		photoResponses[i] = dto.PhotoResponse{
			Type:     p.PhotoType,
			Filename: p.Filename,
			URL:      "/api/v1/photos/" + p.ID.String() + "/file",
		}
	}

	// Build geometry with metadata
	var altitude, accuracy *float64
	if location.GeoMeta != nil {
		if v, ok := location.GeoMeta["altitude"].(float64); ok {
			altitude = &v
		}
		if v, ok := location.GeoMeta["accuracy"].(float64); ok {
			accuracy = &v
		}
	}

	odkSubmissionID := ""
	if location.ODKSubmissionID != nil {
		odkSubmissionID = *location.ODKSubmissionID
	}

	submitterName := ""
	if location.SubmitterName != nil {
		submitterName = *location.SubmitterName
	}

	// Convert JSONB to map
	identitas := map[string]interface{}{}
	if location.Identitas != nil {
		identitas = location.Identitas
	}
	// Add nama to identitas
	identitas["nama"] = location.Nama

	alamat := map[string]interface{}{}
	if location.Alamat != nil {
		alamat = location.Alamat
	}

	dataPengungsi := map[string]interface{}{}
	if location.DataPengungsi != nil {
		dataPengungsi = location.DataPengungsi
	}

	fasilitas := map[string]interface{}{}
	if location.Fasilitas != nil {
		fasilitas = location.Fasilitas
	}

	komunikasi := map[string]interface{}{}
	if location.Komunikasi != nil {
		komunikasi = location.Komunikasi
	}

	akses := map[string]interface{}{}
	if location.Akses != nil {
		akses = location.Akses
	}

	// Get baseline_sumber - prefer dedicated column, fallback to identitas JSONB
	baselineSumber := location.BaselineSumber
	if baselineSumber == "" && location.Identitas != nil {
		if v, ok := location.Identitas["baseline_sumber"].(string); ok {
			baselineSumber = v
		}
	}

	response := dto.LocationDetailResponse{
		ID:              location.ID.String(),
		ODKSubmissionID: odkSubmissionID,
		Type:            location.Type,
		Status:          location.Status,
		BaselineSumber:  baselineSumber,
		Geometry: &dto.LocationGeometry{
			Type:        "Point",
			Coordinates: []float64{location.Longitude, location.Latitude},
			Altitude:    altitude,
			Accuracy:    accuracy,
		},
		Identitas:     identitas,
		Alamat:        alamat,
		DataPengungsi: dataPengungsi,
		Fasilitas:     fasilitas,
		Komunikasi:    komunikasi,
		Akses:         akses,
		Photos:        photoResponses,
		Meta: dto.LocationMeta{
			SubmittedAt:   location.SubmittedAt,
			UpdatedAt:     location.UpdatedAt,
			SubmitterName: submitterName,
		},
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Data:    response,
	})
}
