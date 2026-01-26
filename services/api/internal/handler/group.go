package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/leksa/datamapper-senyar/internal/auth"
	"github.com/leksa/datamapper-senyar/internal/repository"
	"github.com/leksa/datamapper-senyar/internal/service"
)

// GroupHandler handles HTTP requests for groups
type GroupHandler struct {
	service *service.GroupService
}

// NewGroupHandler creates a new group handler
func NewGroupHandler(service *service.GroupService) *GroupHandler {
	return &GroupHandler{service: service}
}

// List returns paginated groups
// GET /api/v1/groups
func (h *GroupHandler) List(c *gin.Context) {
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

	var isActive *bool
	if activeParam := c.Query("is_active"); activeParam != "" {
		active := activeParam == "true"
		isActive = &active
	}

	filter := repository.GroupFilter{
		OrganizationID: orgID,
		Search:         search,
		IsActive:       isActive,
		Page:           page,
		PageSize:       pageSize,
	}

	result, err := h.service.ListWithOrgFilter(c.Request.Context(), filter, userOrgIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch groups",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// GetByID returns a group by ID
// GET /api/v1/groups/:id
func (h *GroupHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	group, err := h.service.GetByIDWithRelawan(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Group not found",
		})
		return
	}

	// Check if user can access this group's organization
	if !auth.CanAccessOrganization(c, group.OrganizationID) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Access denied to this group",
		})
		return
	}

	// Get stats
	stats, err := h.service.GetStats(c.Request.Context(), id)
	if err != nil {
		stats = make(map[string]int64)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"group": group,
			"stats": stats,
		},
	})
}

// Create creates a new group
// POST /api/v1/groups
func (h *GroupHandler) Create(c *gin.Context) {
	var input service.CreateGroupInput
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
			"error":   "Access denied to create groups in this organization",
		})
		return
	}

	group, err := h.service.Create(c.Request.Context(), input)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "organization not found" || err.Error() == "group name already exists in organization" || err.Error() == "invalid organization ID format" {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    group,
	})
}

// Update updates a group
// PUT /api/v1/groups/:id
func (h *GroupHandler) Update(c *gin.Context) {
	id := c.Param("id")

	// First get the group to check organization access
	existingGroup, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Group not found",
		})
		return
	}

	// Check if user can manage this group's organization
	if !auth.CanManageOrganization(c, existingGroup.OrganizationID) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Access denied to update this group",
		})
		return
	}

	var input service.UpdateGroupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	group, err := h.service.Update(c.Request.Context(), id, input)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "group name already exists in organization" || err.Error() == "invalid group ID format" {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    group,
	})
}

// Delete soft-deletes a group
// DELETE /api/v1/groups/:id
func (h *GroupHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	// First get the group to check organization access
	existingGroup, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Group not found",
		})
		return
	}

	// Check if user can manage this group's organization
	if !auth.CanManageOrganization(c, existingGroup.OrganizationID) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Access denied to delete this group",
		})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to delete group",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Group deleted successfully",
	})
}

// GetStats returns statistics for a group
// GET /api/v1/groups/:id/stats
func (h *GroupHandler) GetStats(c *gin.Context) {
	id := c.Param("id")

	// First get the group to check organization access
	group, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Group not found",
		})
		return
	}

	// Check if user can access this group's organization
	if !auth.CanAccessOrganization(c, group.OrganizationID) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Access denied to this group",
		})
		return
	}

	stats, err := h.service.GetStats(c.Request.Context(), id)
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

// GetByOrganization returns all groups for an organization
// GET /api/v1/organizations/:id/groups
func (h *GroupHandler) GetByOrganization(c *gin.Context) {
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

	groups, err := h.service.GetByOrganization(c.Request.Context(), orgIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch groups",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    groups,
	})
}
