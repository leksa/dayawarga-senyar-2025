package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db *gorm.DB
}

func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{
		db: db,
	}
}

type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Services  map[string]string `json:"services"`
}

// Check returns health status of the API
// @Summary      Health check
// @Description  Returns the health status of the API and its dependencies
// @Tags         health
// @Accept       json
// @Produce      json
// @Success      200  {object} HealthResponse
// @Router       /health [get]
func (h *HealthHandler) Check(c *gin.Context) {
	services := make(map[string]string)

	// Check database connection
	sqlDB, err := h.db.DB()
	if err != nil {
		services["database"] = "unhealthy"
	} else if err := sqlDB.Ping(); err != nil {
		services["database"] = "unhealthy"
	} else {
		services["database"] = "healthy"
	}

	// Determine overall status
	status := "healthy"
	for _, v := range services {
		if v != "healthy" {
			status = "degraded"
			break
		}
	}

	c.JSON(http.StatusOK, HealthResponse{
		Status:    status,
		Timestamp: time.Now(),
		Services:  services,
	})
}

// Ready returns readiness status
// @Summary      Readiness check
// @Description  Returns whether the API is ready to handle requests
// @Tags         health
// @Accept       json
// @Produce      json
// @Success      200  {object} map[string]interface{}
// @Failure      503  {object} map[string]interface{}
// @Router       /ready [get]
func (h *HealthHandler) Ready(c *gin.Context) {
	sqlDB, err := h.db.DB()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"error":  "database connection failed",
		})
		return
	}

	if err := sqlDB.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"error":  "database ping failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}
