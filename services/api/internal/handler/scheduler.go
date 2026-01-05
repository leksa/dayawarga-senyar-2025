package handler

import (
	"net/http"

	"github.com/leksa/datamapper-senyar/internal/dto"
	"github.com/leksa/datamapper-senyar/internal/scheduler"

	"github.com/gin-gonic/gin"
)

// SchedulerHandler handles scheduler-related API endpoints
type SchedulerHandler struct {
	scheduler *scheduler.Scheduler
}

// NewSchedulerHandler creates a new scheduler handler
func NewSchedulerHandler(s *scheduler.Scheduler) *SchedulerHandler {
	return &SchedulerHandler{scheduler: s}
}

// GetStatus returns the current scheduler status
// @Summary Get scheduler status
// @Description Returns the current status of the auto-sync scheduler
// @Tags scheduler
// @Accept json
// @Produce json
// @Success 200 {object} dto.APIResponse
// @Router /api/v1/scheduler/status [get]
func (h *SchedulerHandler) GetStatus(c *gin.Context) {
	status := h.scheduler.GetStatus()
	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Data:    status,
	})
}

// SetMode sets the scheduler mode manually
// @Summary Set scheduler mode
// @Description Manually override the scheduler mode (idle, normal, active)
// @Tags scheduler
// @Accept json
// @Produce json
// @Param mode path string true "Mode (idle, normal, active)"
// @Success 200 {object} dto.APIResponse
// @Failure 400 {object} dto.APIResponse
// @Router /api/v1/scheduler/mode/{mode} [post]
func (h *SchedulerHandler) SetMode(c *gin.Context) {
	mode := c.Param("mode")

	var schedulerMode scheduler.Mode
	switch mode {
	case "idle":
		schedulerMode = scheduler.ModeIdle
	case "normal":
		schedulerMode = scheduler.ModeNormal
	case "active":
		schedulerMode = scheduler.ModeActive
	default:
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_MODE",
				Message: "Mode must be one of: idle, normal, active",
			},
		})
		return
	}

	h.scheduler.SetMode(schedulerMode)

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"mode":    mode,
			"message": "Scheduler mode updated",
		},
	})
}

// ClearManualMode clears the manual mode override
// @Summary Clear manual mode
// @Description Clears the manual mode override and returns to automatic scheduling
// @Tags scheduler
// @Accept json
// @Produce json
// @Success 200 {object} dto.APIResponse
// @Router /api/v1/scheduler/mode/auto [post]
func (h *SchedulerHandler) ClearManualMode(c *gin.Context) {
	h.scheduler.ClearManualMode()

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"message": "Manual mode cleared, returning to automatic scheduling",
		},
	})
}

// TriggerSync manually triggers a sync cycle
// @Summary Trigger sync
// @Description Manually triggers an immediate sync cycle
// @Tags scheduler
// @Accept json
// @Produce json
// @Success 200 {object} dto.APIResponse
// @Router /api/v1/scheduler/trigger [post]
func (h *SchedulerHandler) TriggerSync(c *gin.Context) {
	h.scheduler.TriggerSync()

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"message": "Sync triggered",
		},
	})
}

// Start starts the scheduler
// @Summary Start scheduler
// @Description Starts the auto-sync scheduler
// @Tags scheduler
// @Accept json
// @Produce json
// @Success 200 {object} dto.APIResponse
// @Router /api/v1/scheduler/start [post]
func (h *SchedulerHandler) Start(c *gin.Context) {
	h.scheduler.Start()

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"message": "Scheduler started",
		},
	})
}

// Stop stops the scheduler
// @Summary Stop scheduler
// @Description Stops the auto-sync scheduler
// @Tags scheduler
// @Accept json
// @Produce json
// @Success 200 {object} dto.APIResponse
// @Router /api/v1/scheduler/stop [post]
func (h *SchedulerHandler) Stop(c *gin.Context) {
	h.scheduler.Stop()

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"message": "Scheduler stopped",
		},
	})
}
