package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/leksa/datamapper-senyar/internal/dto"
	"github.com/leksa/datamapper-senyar/internal/repository"
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
func (h *LocationHandler) GetLocations(c *gin.Context) {
	filter := repository.LocationFilter{
		Type:   c.Query("type"),
		Status: c.Query("status"),
		Search: c.Query("search"),
		Page:   1,
		Limit:  50,
	}

	// Parse pagination
	if page, err := strconv.Atoi(c.Query("page")); err == nil && page > 0 {
		filter.Page = page
	}
	if limit, err := strconv.Atoi(c.Query("limit")); err == nil && limit > 0 {
		filter.Limit = limit
	}

	// Parse bounding box: bbox=minLng,minLat,maxLng,maxLat
	if bbox := c.Query("bbox"); bbox != "" {
		parts := strings.Split(bbox, ",")
		if len(parts) == 4 {
			if minLng, err := strconv.ParseFloat(parts[0], 64); err == nil {
				filter.MinLng = &minLng
			}
			if minLat, err := strconv.ParseFloat(parts[1], 64); err == nil {
				filter.MinLat = &minLat
			}
			if maxLng, err := strconv.ParseFloat(parts[2], 64); err == nil {
				filter.MaxLng = &maxLng
			}
			if maxLat, err := strconv.ParseFloat(parts[3], 64); err == nil {
				filter.MaxLat = &maxLat
			}
		}
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
			alamatSingkat = strings.Join(parts, ", ")
		}

		// Get jumlah_kk and total_jiwa from data_pengungsi
		jumlahKK := 0
		totalJiwa := 0
		if loc.DataPengungsi != nil {
			if v, ok := loc.DataPengungsi["jumlah_kk"].(float64); ok {
				jumlahKK = int(v)
			}
			if v, ok := loc.DataPengungsi["total_jiwa"].(float64); ok {
				totalJiwa = int(v)
			}
		}

		odkSubmissionID := ""
		if loc.ODKSubmissionID != nil {
			odkSubmissionID = *loc.ODKSubmissionID
		}

		// Get baseline_sumber from identitas
		baselineSumber := ""
		if loc.Identitas != nil {
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
				ODKSubmissionID: odkSubmissionID,
				Nama:            loc.Nama,
				Type:            loc.Type,
				Status:          loc.Status,
				AlamatSingkat:   alamatSingkat,
				NamaProvinsi:    namaProvinsi,
				NamaKotaKab:     namaKotaKab,
				NamaKecamatan:   namaKecamatan,
				NamaDesa:        namaDesa,
				JumlahKK:        jumlahKK,
				TotalJiwa:       totalJiwa,
				BaselineSumber:  baselineSumber,
				UpdatedAt:       loc.UpdatedAt,
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
func (h *LocationHandler) GetLocationByID(c *gin.Context) {
	idStr := c.Param("id")
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
	photos, _ := h.locationRepo.FindPhotos(id)
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

	response := dto.LocationDetailResponse{
		ID:              location.ID.String(),
		ODKSubmissionID: odkSubmissionID,
		Type:            location.Type,
		Status:          location.Status,
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
