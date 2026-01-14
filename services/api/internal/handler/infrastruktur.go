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

type InfrastrukturHandler struct {
	infraRepo *repository.InfrastrukturRepository
}

func NewInfrastrukturHandler(infraRepo *repository.InfrastrukturRepository) *InfrastrukturHandler {
	return &InfrastrukturHandler{
		infraRepo: infraRepo,
	}
}

// GetInfrastruktur returns GeoJSON FeatureCollection of infrastruktur (roads/bridges)
// @Summary Get all infrastruktur
// @Description Returns a GeoJSON FeatureCollection of infrastructure (roads and bridges)
// @Tags infrastruktur
// @Accept json
// @Produce json
// @Param jenis query string false "Filter by jenis (Jalan, Jembatan)"
// @Param status_jln query string false "Filter by status_jln (Nasional, Daerah)"
// @Param status_akses query string false "Filter by status_akses (dapat_diakses, akses_terputus)"
// @Param status_penanganan query string false "Filter by status_penanganan"
// @Param kabupaten query string false "Filter by kabupaten name"
// @Param search query string false "Search by name"
// @Param bbox query string false "Bounding box (minLng,minLat,maxLng,maxLat)"
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} dto.APIResponse
// @Router /api/v1/infrastruktur [get]
func (h *InfrastrukturHandler) GetInfrastruktur(c *gin.Context) {
	filter := repository.InfrastrukturFilter{
		Jenis:            c.Query("jenis"),
		StatusJln:        c.Query("status_jln"),
		StatusAkses:      c.Query("status_akses"),
		StatusPenanganan: c.Query("status_penanganan"),
		NamaKabupaten:    c.Query("kabupaten"),
		Search:           c.Query("search"),
		Page:             1,
		Limit:            50,
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

	infraList, total, err := h.infraRepo.FindAll(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to fetch infrastruktur",
			},
		})
		return
	}

	// Convert to GeoJSON
	features := make([]dto.InfrastrukturFeatureResponse, len(infraList))
	for i, infra := range infraList {
		features[i] = dto.InfrastrukturFeatureResponse{
			Type: "Feature",
			ID:   infra.ID.String(),
			Geometry: &dto.GeoJSONGeometry{
				Type:        "Point",
				Coordinates: []float64{infra.Longitude, infra.Latitude},
			},
			Properties: dto.InfrastrukturListProperties{
				EntityID:         infra.EntityID,
				Nama:             infra.Nama,
				Jenis:            infra.Jenis,
				StatusJln:        infra.StatusJln,
				NamaProvinsi:     infra.NamaProvinsi,
				NamaKabupaten:    infra.NamaKabupaten,
				StatusAkses:      infra.StatusAkses,
				StatusPenanganan: infra.StatusPenanganan,
				Bailey:           infra.Bailey,
				Progress:         infra.Progress,
				UpdatedAt:        infra.UpdatedAt,
			},
		}
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Data: dto.InfrastrukturListResponse{
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

// GetInfrastrukturByID returns detailed infrastruktur info
// @Summary Get infrastruktur by ID
// @Description Returns detailed information about a specific infrastructure
// @Tags infrastruktur
// @Accept json
// @Produce json
// @Param id path string true "Infrastruktur ID"
// @Success 200 {object} dto.APIResponse
// @Failure 400 {object} dto.APIResponse
// @Failure 404 {object} dto.APIResponse
// @Router /api/v1/infrastruktur/{id} [get]
func (h *InfrastrukturHandler) GetInfrastrukturByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "VALIDATION_ERROR",
				Message: "Invalid infrastruktur ID format",
			},
		})
		return
	}

	infra, err := h.infraRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.APIResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "NOT_FOUND",
				Message: "Infrastruktur not found",
			},
		})
		return
	}

	// Get photos
	photos, _ := h.infraRepo.FindPhotos(id)
	photoResponses := make([]dto.PhotoResponse, len(photos))
	for i, p := range photos {
		photoResponses[i] = dto.PhotoResponse{
			Type:     p.PhotoType,
			Filename: p.Filename,
			URL:      "/api/v1/infrastruktur/photos/" + p.ID.String() + "/file",
		}
	}

	submitterName := ""
	if infra.SubmitterName != nil {
		submitterName = *infra.SubmitterName
	}

	response := dto.InfrastrukturDetailResponse{
		ID:            infra.ID.String(),
		EntityID:      infra.EntityID,
		ObjectID:      infra.ObjectID,
		Nama:          infra.Nama,
		Jenis:         infra.Jenis,
		StatusJln:     infra.StatusJln,
		NamaProvinsi:  infra.NamaProvinsi,
		NamaKabupaten: infra.NamaKabupaten,
		Geometry: &dto.LocationGeometry{
			Type:        "Point",
			Coordinates: []float64{infra.Longitude, infra.Latitude},
		},
		StatusAkses:       infra.StatusAkses,
		KeteranganBencana: infra.KeteranganBencana,
		Dampak:            infra.Dampak,
		StatusPenanganan:  infra.StatusPenanganan,
		PenangananDetail:  infra.PenangananDetail,
		Bailey:            infra.Bailey,
		Progress:          infra.Progress,
		TargetSelesai:     infra.TargetSelesai,
		BaselineSumber:    infra.BaselineSumber,
		UpdateBy:          infra.UpdateBy,
		Photos:            photoResponses,
		Meta: dto.LocationMeta{
			SubmittedAt:   infra.SubmittedAt,
			UpdatedAt:     infra.UpdatedAt,
			SubmitterName: submitterName,
		},
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Data:    response,
	})
}

// GetInfrastrukturStats returns statistics about infrastructure
// @Summary Get infrastruktur statistics
// @Description Returns statistics about infrastructure (by jenis, status_akses, etc.)
// @Tags infrastruktur
// @Accept json
// @Produce json
// @Success 200 {object} dto.APIResponse
// @Router /api/v1/infrastruktur/stats [get]
func (h *InfrastrukturHandler) GetInfrastrukturStats(c *gin.Context) {
	stats, err := h.infraRepo.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to fetch statistics",
			},
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Data:    stats,
		Meta: &dto.MetaInfo{
			Timestamp: time.Now(),
		},
	})
}
