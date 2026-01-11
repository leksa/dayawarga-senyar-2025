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

type FaskesHandler struct {
	faskesRepo *repository.FaskesRepository
}

func NewFaskesHandler(faskesRepo *repository.FaskesRepository) *FaskesHandler {
	return &FaskesHandler{
		faskesRepo: faskesRepo,
	}
}

// GetFaskes returns GeoJSON FeatureCollection of faskes (health facilities)
// @Summary Get all faskes
// @Description Returns a GeoJSON FeatureCollection of health facilities
// @Tags faskes
// @Accept json
// @Produce json
// @Param jenis_faskes query string false "Filter by jenis_faskes (rumah_sakit, puskesmas, klinik, posko_kes_darurat)"
// @Param status_faskes query string false "Filter by status_faskes (operasional, non_aktif)"
// @Param kondisi_faskes query string false "Filter by kondisi_faskes"
// @Param search query string false "Search by name"
// @Param bbox query string false "Bounding box (minLng,minLat,maxLng,maxLat)"
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} dto.APIResponse
// @Router /api/v1/faskes [get]
func (h *FaskesHandler) GetFaskes(c *gin.Context) {
	filter := repository.FaskesFilter{
		JenisFaskes:   c.Query("jenis_faskes"),
		StatusFaskes:  c.Query("status_faskes"),
		KondisiFaskes: c.Query("kondisi_faskes"),
		Search:        c.Query("search"),
		Page:          1,
		Limit:         50,
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

	faskesList, total, err := h.faskesRepo.FindAll(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to fetch faskes",
			},
		})
		return
	}

	// Convert to GeoJSON
	features := make([]dto.FaskesFeatureResponse, len(faskesList))
	for i, f := range faskesList {
		// Extract alamat fields
		alamatSingkat := ""
		namaProvinsi := ""
		namaKotaKab := ""
		namaKecamatan := ""
		namaDesa := ""
		if f.Alamat != nil {
			parts := []string{}
			if desa, ok := f.Alamat["nama_desa"].(string); ok && desa != "" {
				parts = append(parts, desa)
				namaDesa = desa
			}
			if kab, ok := f.Alamat["nama_kota_kab"].(string); ok && kab != "" {
				parts = append(parts, kab)
				namaKotaKab = kab
			}
			if kec, ok := f.Alamat["nama_kecamatan"].(string); ok && kec != "" {
				namaKecamatan = kec
			}
			if prov, ok := f.Alamat["nama_provinsi"].(string); ok && prov != "" {
				namaProvinsi = prov
			}
			alamatSingkat = strings.Join(parts, ", ")
		}

		odkSubmissionID := ""
		if f.ODKSubmissionID != nil {
			odkSubmissionID = *f.ODKSubmissionID
		}

		kondisiFaskes := ""
		if f.KondisiFaskes != nil {
			kondisiFaskes = *f.KondisiFaskes
		}

		features[i] = dto.FaskesFeatureResponse{
			Type: "Feature",
			ID:   f.ID.String(),
			Geometry: &dto.GeoJSONGeometry{
				Type:        "Point",
				Coordinates: []float64{f.Longitude, f.Latitude},
			},
			Properties: dto.FaskesListProperties{
				ODKSubmissionID: odkSubmissionID,
				Nama:            f.Nama,
				JenisFaskes:     f.JenisFaskes,
				StatusFaskes:    f.StatusFaskes,
				KondisiFaskes:   kondisiFaskes,
				AlamatSingkat:   alamatSingkat,
				NamaProvinsi:    namaProvinsi,
				NamaKotaKab:     namaKotaKab,
				NamaKecamatan:   namaKecamatan,
				NamaDesa:        namaDesa,
				UpdatedAt:       f.UpdatedAt,
			},
		}
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Data: dto.FaskesListResponse{
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

// GetFaskesByID returns detailed faskes info
// @Summary Get faskes by ID
// @Description Returns detailed information about a specific health facility
// @Tags faskes
// @Accept json
// @Produce json
// @Param id path string true "Faskes ID"
// @Success 200 {object} dto.APIResponse
// @Failure 400 {object} dto.APIResponse
// @Failure 404 {object} dto.APIResponse
// @Router /api/v1/faskes/{id} [get]
func (h *FaskesHandler) GetFaskesByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "VALIDATION_ERROR",
				Message: "Invalid faskes ID format",
			},
		})
		return
	}

	faskes, err := h.faskesRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.APIResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "NOT_FOUND",
				Message: "Faskes not found",
			},
		})
		return
	}

	// Get photos
	photos, _ := h.faskesRepo.FindPhotos(id)
	photoResponses := make([]dto.PhotoResponse, len(photos))
	for i, p := range photos {
		photoResponses[i] = dto.PhotoResponse{
			Type:     p.PhotoType,
			Filename: p.Filename,
			URL:      "/api/v1/faskes/" + id.String() + "/photos/" + p.Filename,
		}
	}

	odkSubmissionID := ""
	if faskes.ODKSubmissionID != nil {
		odkSubmissionID = *faskes.ODKSubmissionID
	}

	kondisiFaskes := ""
	if faskes.KondisiFaskes != nil {
		kondisiFaskes = *faskes.KondisiFaskes
	}

	submitterName := ""
	if faskes.SubmitterName != nil {
		submitterName = *faskes.SubmitterName
	}

	// Convert JSONB to map
	alamat := map[string]interface{}{}
	if faskes.Alamat != nil {
		alamat = faskes.Alamat
	}

	identitas := map[string]interface{}{}
	if faskes.Identitas != nil {
		identitas = faskes.Identitas
	}
	identitas["nama"] = faskes.Nama

	isolasi := map[string]interface{}{}
	if faskes.Isolasi != nil {
		isolasi = faskes.Isolasi
	}

	infrastruktur := map[string]interface{}{}
	if faskes.Infrastruktur != nil {
		infrastruktur = faskes.Infrastruktur
	}

	sdm := map[string]interface{}{}
	if faskes.SDM != nil {
		sdm = faskes.SDM
	}

	perbekalan := map[string]interface{}{}
	if faskes.Perbekalan != nil {
		perbekalan = faskes.Perbekalan
	}

	klaster := map[string]interface{}{}
	if faskes.Klaster != nil {
		klaster = faskes.Klaster
	}

	response := dto.FaskesDetailResponse{
		ID:              faskes.ID.String(),
		ODKSubmissionID: odkSubmissionID,
		Nama:            faskes.Nama,
		JenisFaskes:     faskes.JenisFaskes,
		StatusFaskes:    faskes.StatusFaskes,
		KondisiFaskes:   kondisiFaskes,
		Geometry: &dto.LocationGeometry{
			Type:        "Point",
			Coordinates: []float64{faskes.Longitude, faskes.Latitude},
		},
		Alamat:        alamat,
		Identitas:     identitas,
		Isolasi:       isolasi,
		Infrastruktur: infrastruktur,
		SDM:           sdm,
		Perbekalan:    perbekalan,
		Klaster:       klaster,
		Photos:        photoResponses,
		Meta: dto.LocationMeta{
			SubmittedAt:   faskes.SubmittedAt,
			UpdatedAt:     faskes.UpdatedAt,
			SubmitterName: submitterName,
		},
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Data:    response,
	})
}
