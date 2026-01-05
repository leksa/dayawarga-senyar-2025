package handler

import (
	"net/http"

	"github.com/leksa/datamapper-senyar/internal/dto"
	"github.com/leksa/datamapper-senyar/internal/service"

	"github.com/gin-gonic/gin"
)

// SyncHandler handles sync-related API endpoints
type SyncHandler struct {
	syncService     *service.SyncService
	feedSyncService *service.FeedSyncService
}

// NewSyncHandler creates a new sync handler
func NewSyncHandler(syncService *service.SyncService, feedSyncService *service.FeedSyncService) *SyncHandler {
	return &SyncHandler{
		syncService:     syncService,
		feedSyncService: feedSyncService,
	}
}

// SyncAll triggers a full sync of all submissions
// @Summary Sync all ODK submissions
// @Description Fetches all approved submissions from ODK Central and syncs to PostgreSQL
// @Tags sync
// @Accept json
// @Produce json
// @Success 200 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/v1/sync/posko [post]
func (h *SyncHandler) SyncAll(c *gin.Context) {
	result, err := h.syncService.SyncAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "SYNC_FAILED",
				Message: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Data:    result,
	})
}

// GetSyncStatus returns the current sync status
// @Summary Get sync status
// @Description Returns the current synchronization status for posko form
// @Tags sync
// @Accept json
// @Produce json
// @Success 200 {object} dto.APIResponse
// @Router /api/v1/sync/status [get]
func (h *SyncHandler) GetSyncStatus(c *gin.Context) {
	state, err := h.syncService.GetSyncState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "STATUS_FETCH_FAILED",
				Message: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Data:    state,
	})
}

// SyncFeeds triggers a full sync of all feed submissions
// @Summary Sync all feed submissions
// @Description Fetches all approved feed submissions from ODK Central and syncs to PostgreSQL
// @Tags sync
// @Accept json
// @Produce json
// @Success 200 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/v1/sync/feed [post]
func (h *SyncHandler) SyncFeeds(c *gin.Context) {
	result, err := h.feedSyncService.SyncAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "FEED_SYNC_FAILED",
				Message: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Data:    result,
	})
}

// GetFeedSyncStatus returns the current feed sync status
// @Summary Get feed sync status
// @Description Returns the current synchronization status for feed form
// @Tags sync
// @Accept json
// @Produce json
// @Success 200 {object} dto.APIResponse
// @Router /api/v1/sync/feed/status [get]
func (h *SyncHandler) GetFeedSyncStatus(c *gin.Context) {
	state, err := h.feedSyncService.GetSyncState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "FEED_STATUS_FETCH_FAILED",
				Message: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Data:    state,
	})
}
