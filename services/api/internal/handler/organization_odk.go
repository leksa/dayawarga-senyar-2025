package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/leksa/datamapper-senyar/internal/service"
)

// OrganizationODKHandler handles organization ODK endpoints
type OrganizationODKHandler struct {
	orgODKService *service.OrganizationODKService
}

// NewOrganizationODKHandler creates a new organization ODK handler
func NewOrganizationODKHandler(orgODKService *service.OrganizationODKService) *OrganizationODKHandler {
	return &OrganizationODKHandler{
		orgODKService: orgODKService,
	}
}

// AssignODKProjectRequest represents the request body for assigning ODK project
type AssignODKProjectRequest struct {
	ODKProjectID int `json:"odk_project_id" binding:"required"`
}

// AssignODKProject assigns an ODK project to an organization
// POST /api/v1/admin/organizations/:id/odk-project
func (h *OrganizationODKHandler) AssignODKProject(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse organization ID
	orgID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid organization ID",
		})
		return
	}

	// Parse request body
	var req AssignODKProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	// Assign ODK project
	result, err := h.orgODKService.AssignODKProject(ctx, service.AssignODKProjectInput{
		OrganizationID: orgID,
		ODKProjectID:   req.ODKProjectID,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// RemoveODKProject removes the ODK project assignment from an organization
// DELETE /api/v1/admin/organizations/:id/odk-project
func (h *OrganizationODKHandler) RemoveODKProject(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse organization ID
	orgID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid organization ID",
		})
		return
	}

	// Remove ODK project
	if err := h.orgODKService.RemoveODKProject(ctx, service.RemoveODKProjectInput{
		OrganizationID: orgID,
	}); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "ODK project assignment removed",
	})
}

// GetODKInfo gets ODK-related information for an organization
// GET /api/v1/admin/organizations/:id/odk-info
func (h *OrganizationODKHandler) GetODKInfo(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse organization ID
	orgID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid organization ID",
		})
		return
	}

	// Get ODK info
	info, err := h.orgODKService.GetOrganizationODKInfo(ctx, orgID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    info,
	})
}

// Helper to parse int from param
func parseIntParam(c *gin.Context, param string) (int, error) {
	return strconv.Atoi(c.Param(param))
}
