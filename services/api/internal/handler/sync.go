package handler

import (
	"net/http"

	"github.com/leksa/datamapper-senyar/internal/dto"
	"github.com/leksa/datamapper-senyar/internal/service"

	"github.com/gin-gonic/gin"
)

// SyncHandler handles sync-related API endpoints
type SyncHandler struct {
	syncService             *service.SyncService
	feedSyncService         *service.FeedSyncService
	faskesSyncService       *service.FaskesSyncService
	infrastrukturSyncService *service.InfrastrukturSyncService
}

// NewSyncHandler creates a new sync handler
func NewSyncHandler(syncService *service.SyncService, feedSyncService *service.FeedSyncService, faskesSyncService *service.FaskesSyncService) *SyncHandler {
	return &SyncHandler{
		syncService:       syncService,
		feedSyncService:   feedSyncService,
		faskesSyncService: faskesSyncService,
	}
}

// NewSyncHandlerWithInfrastruktur creates a new sync handler with infrastruktur support
func NewSyncHandlerWithInfrastruktur(syncService *service.SyncService, feedSyncService *service.FeedSyncService, faskesSyncService *service.FaskesSyncService, infrastrukturSyncService *service.InfrastrukturSyncService) *SyncHandler {
	return &SyncHandler{
		syncService:              syncService,
		feedSyncService:          feedSyncService,
		faskesSyncService:        faskesSyncService,
		infrastrukturSyncService: infrastrukturSyncService,
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

// SyncFaskes triggers a full sync of all faskes submissions
// @Summary Sync all faskes submissions
// @Description Fetches all approved faskes submissions from ODK Central and syncs to PostgreSQL
// @Tags sync
// @Accept json
// @Produce json
// @Success 200 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/v1/sync/faskes [post]
func (h *SyncHandler) SyncFaskes(c *gin.Context) {
	result, err := h.faskesSyncService.SyncAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "FASKES_SYNC_FAILED",
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

// GetFaskesSyncStatus returns the current faskes sync status
// @Summary Get faskes sync status
// @Description Returns the current synchronization status for faskes form
// @Tags sync
// @Accept json
// @Produce json
// @Success 200 {object} dto.APIResponse
// @Router /api/v1/sync/faskes/status [get]
func (h *SyncHandler) GetFaskesSyncStatus(c *gin.Context) {
	state, err := h.faskesSyncService.GetSyncState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "FASKES_STATUS_FETCH_FAILED",
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

// ========================================
// HARD SYNC ENDPOINTS
// ========================================

// HardSyncPosko triggers a hard sync of posko - syncs and deletes removed submissions
// @Summary Hard sync posko data
// @Description Syncs posko data and deletes records that no longer exist in ODK Central
// @Tags sync
// @Accept json
// @Produce json
// @Success 200 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/v1/sync/posko/hard [post]
func (h *SyncHandler) HardSyncPosko(c *gin.Context) {
	result, err := h.syncService.HardSync()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "HARD_SYNC_FAILED",
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

// HardSyncFeeds triggers a hard sync of feeds - syncs and deletes removed submissions
// @Summary Hard sync feed data
// @Description Syncs feed data and deletes records that no longer exist in ODK Central
// @Tags sync
// @Accept json
// @Produce json
// @Success 200 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/v1/sync/feed/hard [post]
func (h *SyncHandler) HardSyncFeeds(c *gin.Context) {
	result, err := h.feedSyncService.HardSync()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "FEED_HARD_SYNC_FAILED",
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

// HardSyncFaskes triggers a hard sync of faskes - syncs and deletes removed submissions
// @Summary Hard sync faskes data
// @Description Syncs faskes data and deletes records that no longer exist in ODK Central
// @Tags sync
// @Accept json
// @Produce json
// @Success 200 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/v1/sync/faskes/hard [post]
func (h *SyncHandler) HardSyncFaskes(c *gin.Context) {
	result, err := h.faskesSyncService.HardSync()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "FASKES_HARD_SYNC_FAILED",
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

// ========================================
// INFRASTRUKTUR SYNC ENDPOINTS
// ========================================

// SyncInfrastruktur triggers a full sync of all infrastruktur submissions
// @Summary Sync all infrastruktur submissions
// @Description Fetches all approved infrastruktur submissions from ODK Central and syncs to PostgreSQL
// @Tags sync
// @Accept json
// @Produce json
// @Success 200 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/v1/sync/infrastruktur [post]
func (h *SyncHandler) SyncInfrastruktur(c *gin.Context) {
	if h.infrastrukturSyncService == nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "SERVICE_NOT_CONFIGURED",
				Message: "Infrastruktur sync service not configured",
			},
		})
		return
	}

	result, err := h.infrastrukturSyncService.SyncAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INFRASTRUKTUR_SYNC_FAILED",
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

// GetInfrastrukturSyncStatus returns the current infrastruktur sync status
// @Summary Get infrastruktur sync status
// @Description Returns the current synchronization status for infrastruktur form
// @Tags sync
// @Accept json
// @Produce json
// @Success 200 {object} dto.APIResponse
// @Router /api/v1/sync/infrastruktur/status [get]
func (h *SyncHandler) GetInfrastrukturSyncStatus(c *gin.Context) {
	if h.infrastrukturSyncService == nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "SERVICE_NOT_CONFIGURED",
				Message: "Infrastruktur sync service not configured",
			},
		})
		return
	}

	state, err := h.infrastrukturSyncService.GetSyncState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INFRASTRUKTUR_STATUS_FETCH_FAILED",
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

// HardSyncInfrastruktur triggers a hard sync of infrastruktur - syncs and deletes removed submissions
// @Summary Hard sync infrastruktur data
// @Description Syncs infrastruktur data and deletes records that no longer exist in ODK Central
// @Tags sync
// @Accept json
// @Produce json
// @Success 200 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/v1/sync/infrastruktur/hard [post]
func (h *SyncHandler) HardSyncInfrastruktur(c *gin.Context) {
	if h.infrastrukturSyncService == nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "SERVICE_NOT_CONFIGURED",
				Message: "Infrastruktur sync service not configured",
			},
		})
		return
	}

	result, err := h.infrastrukturSyncService.HardSync()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INFRASTRUKTUR_HARD_SYNC_FAILED",
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
