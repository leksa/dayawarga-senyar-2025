package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/leksa/datamapper-senyar/internal/dto"
	"github.com/leksa/datamapper-senyar/internal/model"
	"github.com/leksa/datamapper-senyar/internal/repository"
)

type FeedHandler struct {
	feedRepo *repository.FeedRepository
	formID   string // ODK form ID for photo URL generation
}

func NewFeedHandler(feedRepo *repository.FeedRepository) *FeedHandler {
	return &FeedHandler{
		feedRepo: feedRepo,
		formID:   "update_informasi", // default form ID
	}
}

// SetFormID sets the ODK form ID for photo URL generation
func (h *FeedHandler) SetFormID(formID string) {
	h.formID = formID
}

// GetFeeds godoc
// @Summary      Get all feeds
// @Description  Retrieve list of information feeds with optional filtering and pagination
// @Tags         feeds
// @Accept       json
// @Produce      json
// @Param        category    query     string  false  "Filter by category"
// @Param        type        query     string  false  "Filter by type"
// @Param        location_id query     string  false  "Filter by location ID"
// @Param        location_name query     string  false  "Filter by location name"
// @Param        search      query     string  false  "Search in content"
// @Param        provinsi    query     string  false  "Filter by province"
// @Param        kota_kab    query     string  false  "Filter by regency/city"
// @Param        kecamatan   query     string  false  "Filter by district"
// @Param        desa        query     string  false  "Filter by village"
// @Param        since       query     string  false  "Get feeds after this timestamp"
// @Param        page        query     int     false  "Page number" minimum(1) default(1)
// @Param        limit       query     int     false  "Items per page" minimum(1) default(50)
// @Success      200  {object}  dto.APIResponse{data=[]dto.FeedResponse}
// @Failure      500  {object}  dto.APIResponse
// @Router       /api/v1/feeds [get]
func (h *FeedHandler) GetFeeds(c *gin.Context) {
	filter := repository.FeedFilter{
		Category:     c.Query("category"),
		Type:         c.Query("type"),
		LocationID:   c.Query("location_id"),
		LocationName: c.Query("location_name"),
		Search:       c.Query("search"),
		Since:        c.Query("since"),
		// Region filters
		Provinsi:  c.Query("provinsi"),
		KotaKab:   c.Query("kota_kab"),
		Kecamatan: c.Query("kecamatan"),
		Desa:      c.Query("desa"),
		Page:      1,
		Limit:     50,
	}

	// Parse pagination
	if page, err := strconv.Atoi(c.Query("page")); err == nil && page > 0 {
		filter.Page = page
	}
	if limit, err := strconv.Atoi(c.Query("limit")); err == nil && limit > 0 {
		filter.Limit = limit
	}

	feeds, total, err := h.feedRepo.FindAll(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to fetch feeds",
			},
		})
		return
	}

	// Collect feed IDs for batch photo query
	feedIDs := make([]uuid.UUID, len(feeds))
	for i, feed := range feeds {
		feedIDs[i] = feed.ID
	}

	// Batch fetch photos for all feeds
	photosMap, _ := h.feedRepo.GetPhotosForFeeds(feedIDs)

	// Convert to response
	feedResponses := make([]dto.FeedResponse, len(feeds))
	for i, feed := range feeds {
		var locationID *string
		if feed.LocationID != nil {
			locIDStr := feed.LocationID.String()
			locationID = &locIDStr
		}

		var faskesID *string
		if feed.FaskesID != nil {
			faskesIDStr := feed.FaskesID.String()
			faskesID = &faskesIDStr
		}

		var coords []float64
		if feed.Longitude != nil && feed.Latitude != nil {
			coords = []float64{*feed.Longitude, *feed.Latitude}
		}

		// Get photos for this feed
		var photoResponses []dto.FeedPhotoResponse
		if photos, ok := photosMap[feed.ID]; ok {
			photoResponses = h.convertPhotosToResponse(photos, feed.ODKSubmissionID)
		}

		// Extract region from raw_data
		var region *dto.FeedRegion
		if feed.RawData != nil {
			region = extractRegionFromRawData(feed.RawData)
		}

		feedResponses[i] = dto.FeedResponse{
			ID:           feed.ID.String(),
			ShortCode:    feed.ShortCode,
			LocationID:   locationID,
			LocationName: feed.LocationName,
			FaskesID:     faskesID,
			FaskesName:   feed.FaskesName,
			Category:     feed.Category,
			Type:         feed.Type,
			Content:      feed.Content,
			Username:     feed.Username,
			Organization: feed.Organization,
			SubmittedAt:  getSubmittedAt(feed.SubmittedAt, feed.CreatedAt),
			Coordinates:  coords,
			Photos:       photoResponses,
			Region:       region,
		}
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Data:    feedResponses,
		Meta: &dto.MetaInfo{
			Total:     total,
			Page:      filter.Page,
			Limit:     filter.Limit,
			Timestamp: time.Now(),
		},
	})
}

// convertPhotosToResponse converts feed photos to response format
func (h *FeedHandler) convertPhotosToResponse(photos []model.FeedPhoto, odkSubmissionID *string) []dto.FeedPhotoResponse {
	result := make([]dto.FeedPhotoResponse, len(photos))
	for i, photo := range photos {
		// Build photo URL - use feed photo endpoint (cached group has no prefix)
		url := fmt.Sprintf("/api/v1/feed-photos/%s/file", photo.ID.String())

		result[i] = dto.FeedPhotoResponse{
			ID:       photo.ID.String(),
			Type:     photo.PhotoType,
			Filename: photo.Filename,
			URL:      url,
		}
	}
	return result
}

// GetFeedByID godoc
// @Summary      Get feed by ID
// @Description  Retrieve a single feed by its ID
// @Tags         feeds
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Feed ID"
// @Success      200  {object}  dto.APIResponse{data=dto.FeedResponse}
// @Failure      400  {object}  dto.APIResponse
// @Failure      404  {object}  dto.APIResponse
// @Failure      500  {object}  dto.APIResponse
// @Router       /api/v1/feeds/{id} [get]
func (h *FeedHandler) GetFeedByID(c *gin.Context) {
	idStr := c.Param("id")
	feedID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "VALIDATION_ERROR",
				Message: "Invalid feed ID format",
			},
		})
		return
	}

	feed, err := h.feedRepo.FindByID(feedID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.APIResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "NOT_FOUND",
				Message: "Feed not found",
			},
		})
		return
	}

	// Get photos for this feed
	photos, _ := h.feedRepo.GetPhotosForFeed(feed.ID)
	photoResponses := h.convertPhotosToResponse(photos, feed.ODKSubmissionID)

	var locationID *string
	if feed.LocationID != nil {
		locIDStr := feed.LocationID.String()
		locationID = &locIDStr
	}

	var faskesID *string
	if feed.FaskesID != nil {
		faskesIDStr := feed.FaskesID.String()
		faskesID = &faskesIDStr
	}

	var coords []float64
	if feed.Longitude != nil && feed.Latitude != nil {
		coords = []float64{*feed.Longitude, *feed.Latitude}
	}

	var region *dto.FeedRegion
	if feed.RawData != nil {
		region = extractRegionFromRawData(feed.RawData)
	}

	response := dto.FeedResponse{
		ID:           feed.ID.String(),
		ShortCode:    feed.ShortCode,
		LocationID:   locationID,
		LocationName: feed.LocationName,
		FaskesID:     faskesID,
		FaskesName:   feed.FaskesName,
		Category:     feed.Category,
		Type:         feed.Type,
		Content:      feed.Content,
		Username:     feed.Username,
		Organization: feed.Organization,
		SubmittedAt:  getSubmittedAt(feed.SubmittedAt, feed.CreatedAt),
		Coordinates:  coords,
		Photos:       photoResponses,
		Region:       region,
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Data:    response,
	})
}

func (h *FeedHandler) GetFeedByShortCode(c *gin.Context) {
	shortCode := c.Param("code")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "VALIDATION_ERROR",
				Message: "Short code is required",
			},
		})
		return
	}

	feed, err := h.feedRepo.FindByShortCode(shortCode)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.APIResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "NOT_FOUND",
				Message: "Feed not found",
			},
		})
		return
	}

	photos, _ := h.feedRepo.GetPhotosForFeed(feed.ID)
	photoResponses := h.convertPhotosToResponse(photos, feed.ODKSubmissionID)

	var locationID *string
	if feed.LocationID != nil {
		locIDStr := feed.LocationID.String()
		locationID = &locIDStr
	}

	var faskesID *string
	if feed.FaskesID != nil {
		faskesIDStr := feed.FaskesID.String()
		faskesID = &faskesIDStr
	}

	var coords []float64
	if feed.Longitude != nil && feed.Latitude != nil {
		coords = []float64{*feed.Longitude, *feed.Latitude}
	}

	var region *dto.FeedRegion
	if feed.RawData != nil {
		region = extractRegionFromRawData(feed.RawData)
	}

	response := dto.FeedResponse{
		ID:           feed.ID.String(),
		ShortCode:    feed.ShortCode,
		LocationID:   locationID,
		LocationName: feed.LocationName,
		FaskesID:     faskesID,
		FaskesName:   feed.FaskesName,
		Category:     feed.Category,
		Type:         feed.Type,
		Content:      feed.Content,
		Username:     feed.Username,
		Organization: feed.Organization,
		SubmittedAt:  getSubmittedAt(feed.SubmittedAt, feed.CreatedAt),
		Coordinates:  coords,
		Photos:       photoResponses,
		Region:       region,
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Data:    response,
	})
}

// GetFeedsByLocation godoc
// @Summary      Get feeds by location
// @Description  Retrieve list of feeds for a specific location with pagination
// @Tags         feeds
// @Accept       json
// @Produce      json
// @Param        id     path      string  true   "Location ID"
// @Param        page   query     int     false  "Page number" minimum(1) default(1)
// @Param        limit  query     int     false  "Items per page" minimum(1) default(50)
// @Success      200  {object}  dto.APIResponse{data=[]dto.FeedResponse}
// @Failure      400  {object}  dto.APIResponse
// @Failure      500  {object}  dto.APIResponse
// @Router       /api/v1/locations/{id}/feeds [get]
func (h *FeedHandler) GetFeedsByLocation(c *gin.Context) {
	idStr := c.Param("id")
	locationID, err := uuid.Parse(idStr)
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

	filter := repository.FeedFilter{
		LocationID: locationID.String(),
		Page:       1,
		Limit:      50,
	}

	// Parse pagination
	if page, err := strconv.Atoi(c.Query("page")); err == nil && page > 0 {
		filter.Page = page
	}
	if limit, err := strconv.Atoi(c.Query("limit")); err == nil && limit > 0 {
		filter.Limit = limit
	}

	feeds, total, err := h.feedRepo.FindAll(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to fetch feeds",
			},
		})
		return
	}

	// Collect feed IDs for batch photo query
	locFeedIDs := make([]uuid.UUID, len(feeds))
	for i, feed := range feeds {
		locFeedIDs[i] = feed.ID
	}

	// Batch fetch photos for all feeds
	locPhotosMap, _ := h.feedRepo.GetPhotosForFeeds(locFeedIDs)

	// Convert to response
	feedResponses := make([]dto.FeedResponse, len(feeds))
	for i, feed := range feeds {
		var locID *string
		if feed.LocationID != nil {
			locIDStr := feed.LocationID.String()
			locID = &locIDStr
		}

		var faskesID *string
		if feed.FaskesID != nil {
			faskesIDStr := feed.FaskesID.String()
			faskesID = &faskesIDStr
		}

		var coords []float64
		if feed.Longitude != nil && feed.Latitude != nil {
			coords = []float64{*feed.Longitude, *feed.Latitude}
		}

		// Get photos for this feed
		var photoResponses []dto.FeedPhotoResponse
		if photos, ok := locPhotosMap[feed.ID]; ok {
			photoResponses = h.convertPhotosToResponse(photos, feed.ODKSubmissionID)
		}

		feedResponses[i] = dto.FeedResponse{
			ID:           feed.ID.String(),
			LocationID:   locID,
			LocationName: feed.LocationName,
			FaskesID:     faskesID,
			FaskesName:   feed.FaskesName,
			Category:     feed.Category,
			Type:         feed.Type,
			Content:      feed.Content,
			Username:     feed.Username,
			Organization: feed.Organization,
			SubmittedAt:  getSubmittedAt(feed.SubmittedAt, feed.CreatedAt),
			Coordinates:  coords,
			Photos:       photoResponses,
		}
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Data:    feedResponses,
		Meta: &dto.MetaInfo{
			Total:     total,
			Page:      filter.Page,
			Limit:     filter.Limit,
			Timestamp: time.Now(),
		},
	})
}

func getSubmittedAt(submittedAt *time.Time, createdAt time.Time) time.Time {
	if submittedAt != nil {
		return *submittedAt
	}
	return createdAt
}

// extractRegionFromRawData extracts region info from ODK raw_data
// Uses calc_nama_* fields for names and sel_* fields for IDs (BPS codes)
// Also checks grp_relasi group for feed forms that store region data there
func extractRegionFromRawData(rawData model.JSONB) *dto.FeedRegion {
	if rawData == nil {
		return nil
	}

	// Cast to map
	data := (map[string]interface{})(rawData)

	region := &dto.FeedRegion{}
	hasData := false

	// Use calc_nama_* fields which contain the actual names
	if v, ok := data["calc_nama_provinsi"].(string); ok && v != "" {
		region.Provinsi = v
		hasData = true
	}
	if v, ok := data["calc_nama_kota_kab"].(string); ok && v != "" {
		region.KotaKab = v
		hasData = true
	}
	if v, ok := data["calc_nama_kecamatan"].(string); ok && v != "" {
		region.Kecamatan = v
		hasData = true
	}
	if v, ok := data["calc_nama_desa"].(string); ok && v != "" {
		region.Desa = v
		hasData = true
	}

	// Extract ID codes from sel_* fields (BPS codes from ODK selection)
	// First try root level (location/faskes forms)
	if v, ok := data["sel_provinsi"].(string); ok && v != "" {
		region.IDProvinsi = v
		hasData = true
	}
	if v, ok := data["sel_kota_kab"].(string); ok && v != "" {
		region.IDKotaKab = v
		hasData = true
	}
	if v, ok := data["sel_kecamatan"].(string); ok && v != "" {
		region.IDKecamatan = v
		hasData = true
	}
	if v, ok := data["sel_desa"].(string); ok && v != "" {
		region.IDDesa = v
		hasData = true
	}

	// Also check grp_relasi for feed forms that might store sel_* there
	if grpRelasi, ok := data["grp_relasi"].(map[string]interface{}); ok {
		if region.IDProvinsi == "" {
			if v, ok := grpRelasi["sel_provinsi"].(string); ok && v != "" {
				region.IDProvinsi = v
				hasData = true
			}
		}
		if region.IDKotaKab == "" {
			if v, ok := grpRelasi["sel_kota_kab"].(string); ok && v != "" {
				region.IDKotaKab = v
				hasData = true
			}
		}
		if region.IDKecamatan == "" {
			if v, ok := grpRelasi["sel_kecamatan"].(string); ok && v != "" {
				region.IDKecamatan = v
				hasData = true
			}
		}
		if region.IDDesa == "" {
			if v, ok := grpRelasi["sel_desa"].(string); ok && v != "" {
				region.IDDesa = v
				hasData = true
			}
		}
	}

	if !hasData {
		return nil
	}
	return region
}
