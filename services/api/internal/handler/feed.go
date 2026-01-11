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

// GetFeeds returns list of information feeds
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
		url := fmt.Sprintf("/api/v1/feeds/photos/%s/file", photo.ID.String())

		result[i] = dto.FeedPhotoResponse{
			ID:       photo.ID.String(),
			Type:     photo.PhotoType,
			Filename: photo.Filename,
			URL:      url,
		}
	}
	return result
}

// GetFeedsByLocation returns feeds for a specific location
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
// Uses calc_nama_* fields (calculated by XLSForm using jr:choice-name)
func extractRegionFromRawData(rawData model.JSONB) *dto.FeedRegion {
	if rawData == nil {
		return nil
	}

	// Cast to map
	data := (map[string]interface{})(rawData)

	region := &dto.FeedRegion{}
	hasData := false

	// Use calc_nama_* fields which contain the actual names (not BPS codes)
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

	if !hasData {
		return nil
	}
	return region
}
