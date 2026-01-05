package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/leksa/datamapper-senyar/internal/dto"
	"github.com/leksa/datamapper-senyar/internal/repository"
)

type FeedHandler struct {
	feedRepo *repository.FeedRepository
}

func NewFeedHandler(feedRepo *repository.FeedRepository) *FeedHandler {
	return &FeedHandler{
		feedRepo: feedRepo,
	}
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
		Page:         1,
		Limit:        50,
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

	// Convert to response
	feedResponses := make([]dto.FeedResponse, len(feeds))
	for i, feed := range feeds {
		var locationID *string
		if feed.LocationID != nil {
			locIDStr := feed.LocationID.String()
			locationID = &locIDStr
		}

		var coords []float64
		if feed.Longitude != nil && feed.Latitude != nil {
			coords = []float64{*feed.Longitude, *feed.Latitude}
		}

		feedResponses[i] = dto.FeedResponse{
			ID:           feed.ID.String(),
			LocationID:   locationID,
			LocationName: feed.LocationName,
			Category:     feed.Category,
			Type:         feed.Type,
			Content:      feed.Content,
			Username:     feed.Username,
			Organization: feed.Organization,
			SubmittedAt:  getSubmittedAt(feed.SubmittedAt, feed.CreatedAt),
			Coordinates:  coords,
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

	// Convert to response
	feedResponses := make([]dto.FeedResponse, len(feeds))
	for i, feed := range feeds {
		var locID *string
		if feed.LocationID != nil {
			locIDStr := feed.LocationID.String()
			locID = &locIDStr
		}

		var coords []float64
		if feed.Longitude != nil && feed.Latitude != nil {
			coords = []float64{*feed.Longitude, *feed.Latitude}
		}

		feedResponses[i] = dto.FeedResponse{
			ID:           feed.ID.String(),
			LocationID:   locID,
			LocationName: feed.LocationName,
			Category:     feed.Category,
			Type:         feed.Type,
			Content:      feed.Content,
			Username:     feed.Username,
			Organization: feed.Organization,
			SubmittedAt:  getSubmittedAt(feed.SubmittedAt, feed.CreatedAt),
			Coordinates:  coords,
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
