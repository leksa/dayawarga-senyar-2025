package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/leksa/datamapper-senyar/internal/service"
)

// RelawanODKHandler handles ODK-related endpoints for relawan
type RelawanODKHandler struct {
	service *service.RelawanODKService
}

// NewRelawanODKHandler creates a new relawan ODK handler
func NewRelawanODKHandler(service *service.RelawanODKService) *RelawanODKHandler {
	return &RelawanODKHandler{service: service}
}

// CreateAppUser creates an ODK App User for a relawan
// POST /api/v1/relawan/:id/odk-app-user
func (h *RelawanODKHandler) CreateAppUser(c *gin.Context) {
	ctx := c.Request.Context()

	relawanID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid relawan ID",
		})
		return
	}

	relawan, err := h.service.CreateAppUserForRelawan(ctx, relawanID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    relawan,
		"message": "ODK App User created successfully",
	})
}

// RevokeAppUser revokes ODK App User access for a relawan
// DELETE /api/v1/relawan/:id/odk-app-user
func (h *RelawanODKHandler) RevokeAppUser(c *gin.Context) {
	ctx := c.Request.Context()

	relawanID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid relawan ID",
		})
		return
	}

	if err := h.service.RevokeAppUserForRelawan(ctx, relawanID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "ODK App User revoked successfully",
	})
}

// GetQRCode gets the QR code data for a relawan's ODK Collect
// GET /api/v1/relawan/:id/odk-qr-code
func (h *RelawanODKHandler) GetQRCode(c *gin.Context) {
	ctx := c.Request.Context()

	relawanID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid relawan ID",
		})
		return
	}

	qrData, err := h.service.GetQRCodeForRelawan(ctx, relawanID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    qrData,
	})
}

// AssignFormsInput represents input for assigning forms to a relawan
type AssignFormsInput struct {
	FormIDs []string `json:"form_ids" binding:"required"`
}

// AssignForms assigns forms to a relawan's ODK App User
// POST /api/v1/relawan/:id/odk-forms
func (h *RelawanODKHandler) AssignForms(c *gin.Context) {
	ctx := c.Request.Context()

	relawanID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid relawan ID",
		})
		return
	}

	var input AssignFormsInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	if err := h.service.AssignFormsToRelawan(ctx, relawanID, input.FormIDs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Forms assigned successfully",
	})
}

// CreateGroupAppUsers creates ODK App Users for all relawan in a group
// POST /api/v1/groups/:id/odk-app-users
func (h *RelawanODKHandler) CreateGroupAppUsers(c *gin.Context) {
	ctx := c.Request.Context()

	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid group ID",
		})
		return
	}

	created, err := h.service.EnsureAppUserForGroupRelawan(ctx, groupID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"created_count": created,
		},
		"message": "ODK App Users created for group relawan",
	})
}
