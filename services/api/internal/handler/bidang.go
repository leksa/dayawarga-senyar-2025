package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leksa/datamapper-senyar/internal/service"
)

// BidangHandler handles HTTP requests for bidangs
type BidangHandler struct {
	service *service.BidangService
}

// NewBidangHandler creates a new bidang handler
func NewBidangHandler(service *service.BidangService) *BidangHandler {
	return &BidangHandler{service: service}
}

// List returns all active bidangs
// GET /api/v1/bidang
func (h *BidangHandler) List(c *gin.Context) {
	bidangs, err := h.service.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch bidangs",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    bidangs,
	})
}
