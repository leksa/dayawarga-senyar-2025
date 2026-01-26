package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/leksa/datamapper-senyar/internal/auth"
	"github.com/leksa/datamapper-senyar/internal/model"
	"github.com/leksa/datamapper-senyar/internal/repository"
	"github.com/leksa/datamapper-senyar/internal/service"
)

// RelawanHandler handles HTTP requests for relawan
type RelawanHandler struct {
	service *service.RelawanService
}

// NewRelawanHandler creates a new relawan handler
func NewRelawanHandler(service *service.RelawanService) *RelawanHandler {
	return &RelawanHandler{service: service}
}

// List returns paginated relawan
// GET /api/v1/relawan
func (h *RelawanHandler) List(c *gin.Context) {
	// Get user's organization IDs for filtering
	userOrgIDs, ok := auth.GetUserOrgIDs(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Unable to determine user organizations",
		})
		return
	}

	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	search := c.Query("search")

	var orgID *uuid.UUID
	if orgIDParam := c.Query("organization_id"); orgIDParam != "" {
		parsed, err := uuid.Parse(orgIDParam)
		if err == nil {
			// If user specifies org_id, verify they can access it
			if userOrgIDs != nil && !auth.CanAccessOrganization(c, parsed) {
				c.JSON(http.StatusForbidden, gin.H{
					"success": false,
					"error":   "Access denied to this organization",
				})
				return
			}
			orgID = &parsed
		}
	}

	var groupID *uuid.UUID
	if groupIDParam := c.Query("group_id"); groupIDParam != "" {
		parsed, err := uuid.Parse(groupIDParam)
		if err == nil {
			groupID = &parsed
		}
	}

	var status *model.RelawanStatus
	if statusParam := c.Query("status"); statusParam != "" {
		s := model.RelawanStatus(statusParam)
		status = &s
	}

	var hasODKAccess *bool
	if odkParam := c.Query("has_odk_access"); odkParam != "" {
		hasODK := odkParam == "true"
		hasODKAccess = &hasODK
	}

	filter := repository.RelawanFilter{
		OrganizationID: orgID,
		GroupID:        groupID,
		Status:         status,
		Search:         search,
		HasODKAccess:   hasODKAccess,
		Page:           page,
		PageSize:       pageSize,
	}

	result, err := h.service.ListWithOrgFilter(c.Request.Context(), filter, userOrgIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch relawan",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// GetByID returns a relawan by ID
// GET /api/v1/relawan/:id
func (h *RelawanHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	relawan, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Relawan not found",
		})
		return
	}

	// Check if user can access this relawan's organization
	if !auth.CanAccessOrganization(c, relawan.OrganizationID) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Access denied to this relawan",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    relawan,
	})
}

// Create creates a new relawan
// POST /api/v1/relawan
func (h *RelawanHandler) Create(c *gin.Context) {
	var input service.CreateRelawanInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	// Parse and verify access to target organization
	orgID, err := uuid.Parse(input.OrganizationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid organization ID format",
		})
		return
	}

	// Check if user can manage this organization
	if !auth.CanManageOrganization(c, orgID) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Access denied to create relawan in this organization",
		})
		return
	}

	relawan, err := h.service.Create(c.Request.Context(), input)
	if err != nil {
		status := http.StatusInternalServerError
		errMsg := err.Error()
		if errMsg == "organization not found" || errMsg == "group not found" ||
			errMsg == "group does not belong to organization" ||
			errMsg == "phone number already exists in organization" ||
			errMsg == "invalid organization ID format" || errMsg == "invalid group ID format" {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{
			"success": false,
			"error":   errMsg,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    relawan,
	})
}

// Update updates a relawan
// PUT /api/v1/relawan/:id
func (h *RelawanHandler) Update(c *gin.Context) {
	id := c.Param("id")

	// First get the relawan to check organization access
	existingRelawan, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Relawan not found",
		})
		return
	}

	// Check if user can manage this relawan's organization
	if !auth.CanManageOrganization(c, existingRelawan.OrganizationID) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Access denied to update this relawan",
		})
		return
	}

	var input service.UpdateRelawanInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	relawan, err := h.service.Update(c.Request.Context(), id, input)
	if err != nil {
		status := http.StatusInternalServerError
		errMsg := err.Error()
		if errMsg == "group not found" || errMsg == "group does not belong to organization" ||
			errMsg == "phone number already exists in organization" ||
			errMsg == "invalid relawan ID format" || errMsg == "invalid group ID format" {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{
			"success": false,
			"error":   errMsg,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    relawan,
	})
}

// Delete soft-deletes a relawan
// DELETE /api/v1/relawan/:id
func (h *RelawanHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	// First get the relawan to check organization access
	existingRelawan, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Relawan not found",
		})
		return
	}

	// Check if user can manage this relawan's organization
	if !auth.CanManageOrganization(c, existingRelawan.OrganizationID) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Access denied to delete this relawan",
		})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to delete relawan",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Relawan deleted successfully",
	})
}

// UpdateStatusInput represents input for updating relawan status
type UpdateStatusInput struct {
	Status model.RelawanStatus `json:"status" binding:"required"`
}

// UpdateStatus updates a relawan's status
// PUT /api/v1/relawan/:id/status
func (h *RelawanHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")

	// First get the relawan to check organization access
	existingRelawan, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Relawan not found",
		})
		return
	}

	// Check if user can manage this relawan's organization
	if !auth.CanManageOrganization(c, existingRelawan.OrganizationID) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Access denied to update this relawan",
		})
		return
	}

	var input UpdateStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	if err := h.service.UpdateStatus(c.Request.Context(), id, input.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Status updated successfully",
	})
}

// MoveToGroupInput represents input for moving relawan to group
type MoveToGroupInput struct {
	GroupID *string `json:"group_id"`
}

// MoveToGroup moves a relawan to a different group
// PUT /api/v1/relawan/:id/group
func (h *RelawanHandler) MoveToGroup(c *gin.Context) {
	id := c.Param("id")

	// First get the relawan to check organization access
	existingRelawan, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Relawan not found",
		})
		return
	}

	// Check if user can manage this relawan's organization
	if !auth.CanManageOrganization(c, existingRelawan.OrganizationID) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Access denied to move this relawan",
		})
		return
	}

	var input MoveToGroupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	if err := h.service.MoveToGroup(c.Request.Context(), id, input.GroupID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Relawan moved to group successfully",
	})
}

// GetStats returns statistics for relawan
// GET /api/v1/relawan/stats
func (h *RelawanHandler) GetStats(c *gin.Context) {
	// Get user's organization IDs for filtering
	userOrgIDs, ok := auth.GetUserOrgIDs(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Unable to determine user organizations",
		})
		return
	}

	var orgID *string
	if orgIDParam := c.Query("organization_id"); orgIDParam != "" {
		// If user specifies org_id, verify they can access it
		parsed, err := uuid.Parse(orgIDParam)
		if err == nil && userOrgIDs != nil && !auth.CanAccessOrganization(c, parsed) {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Access denied to this organization",
			})
			return
		}
		orgID = &orgIDParam
	}

	stats, err := h.service.GetStatsWithOrgFilter(c.Request.Context(), orgID, userOrgIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch statistics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// BulkMoveToGroupInput represents input for bulk moving relawan
type BulkMoveToGroupInput struct {
	IDs     []string `json:"ids" binding:"required"`
	GroupID *string  `json:"group_id"`
}

// BulkMoveToGroup moves multiple relawan to a group
// POST /api/v1/relawan/bulk/move-to-group
func (h *RelawanHandler) BulkMoveToGroup(c *gin.Context) {
	var input BulkMoveToGroupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	// Verify access to all relawan being moved
	for _, relawanID := range input.IDs {
		relawan, err := h.service.GetByID(c.Request.Context(), relawanID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Relawan not found: " + relawanID,
			})
			return
		}
		if !auth.CanManageOrganization(c, relawan.OrganizationID) {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Access denied to move relawan: " + relawanID,
			})
			return
		}
	}

	if err := h.service.BulkMoveToGroup(c.Request.Context(), input.IDs, input.GroupID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Relawan moved to group successfully",
	})
}

// GetByOrganization returns all relawan for an organization
// GET /api/v1/organizations/:id/relawan
func (h *RelawanHandler) GetByOrganization(c *gin.Context) {
	orgIDStr := c.Param("id")

	// Parse organization ID
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid organization ID format",
		})
		return
	}

	// Check if user can access this organization
	if !auth.CanAccessOrganization(c, orgID) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Access denied to this organization",
		})
		return
	}

	relawan, err := h.service.GetByOrganization(c.Request.Context(), orgIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch relawan",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    relawan,
	})
}

// GetByGroup returns all relawan for a group
// GET /api/v1/groups/:id/relawan
func (h *RelawanHandler) GetByGroup(c *gin.Context) {
	groupID := c.Param("id")

	// We need to get the group first to check organization access
	// For now, we'll let the service handle this and add access check
	// Note: This endpoint should ideally verify the group belongs to an accessible org

	relawan, err := h.service.GetByGroup(c.Request.Context(), groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch relawan",
		})
		return
	}

	// If there are relawan, check access to first one's organization
	// This is a simplified check - all relawan in a group belong to same org
	if len(relawan) > 0 && !auth.CanAccessOrganization(c, relawan[0].OrganizationID) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Access denied to this group",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    relawan,
	})
}

// SetWAVerified enables WhatsApp verification for a relawan
// POST /api/v1/relawan/:id/wa-verify
func (h *RelawanHandler) SetWAVerified(c *gin.Context) {
	id := c.Param("id")

	existingRelawan, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Relawan not found",
		})
		return
	}

	if !auth.CanManageOrganization(c, existingRelawan.OrganizationID) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Access denied to manage this relawan",
		})
		return
	}

	if err := h.service.SetWAVerified(c.Request.Context(), id, true); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "WhatsApp access enabled",
	})
}

// RevokeWAVerified disables WhatsApp verification for a relawan
// DELETE /api/v1/relawan/:id/wa-verify
func (h *RelawanHandler) RevokeWAVerified(c *gin.Context) {
	id := c.Param("id")

	existingRelawan, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Relawan not found",
		})
		return
	}

	if !auth.CanManageOrganization(c, existingRelawan.OrganizationID) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Access denied to manage this relawan",
		})
		return
	}

	if err := h.service.SetWAVerified(c.Request.Context(), id, false); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "WhatsApp access revoked",
	})
}

// GetWAStatus returns WhatsApp status for a relawan
// GET /api/v1/relawan/:id/wa-status
func (h *RelawanHandler) GetWAStatus(c *gin.Context) {
	id := c.Param("id")

	existingRelawan, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Relawan not found",
		})
		return
	}

	if !auth.CanAccessOrganization(c, existingRelawan.OrganizationID) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Access denied to view this relawan",
		})
		return
	}

	status, err := h.service.GetWAStatus(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    status,
	})
}

// ValidateWAAccess validates if a phone number has WhatsApp access (public endpoint for chatbot)
// GET /api/v1/wa/validate?phone=xxx
func (h *RelawanHandler) ValidateWAAccess(c *gin.Context) {
	phone := c.Query("phone")
	if phone == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "phone parameter required",
		})
		return
	}

	relawan, err := h.service.ValidateWAAccess(c.Request.Context(), phone)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success":    true,
			"has_access": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"has_access": true,
		"relawan": gin.H{
			"id":              relawan.ID,
			"name":            relawan.Name,
			"organization_id": relawan.OrganizationID,
		},
	})
}

// RecordWAActivity records activity for a phone number (public endpoint for chatbot)
// POST /api/v1/wa/activity
func (h *RelawanHandler) RecordWAActivity(c *gin.Context) {
	var input struct {
		Phone string `json:"phone" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "phone required",
		})
		return
	}

	if err := h.service.UpdateWAActivity(c.Request.Context(), input.Phone); err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Activity recorded",
	})
}
