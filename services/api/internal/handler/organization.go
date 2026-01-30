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

// OrganizationHandler handles HTTP requests for organizations
type OrganizationHandler struct {
	service         *service.OrganizationService
	bidangService   *service.BidangService
	activityService *service.OrganizationActivityService
}

// NewOrganizationHandler creates a new organization handler
func NewOrganizationHandler(service *service.OrganizationService) *OrganizationHandler {
	return &OrganizationHandler{service: service}
}

// NewOrganizationHandlerWithBidang creates a new organization handler with bidang service
func NewOrganizationHandlerWithBidang(service *service.OrganizationService, bidangService *service.BidangService) *OrganizationHandler {
	return &OrganizationHandler{service: service, bidangService: bidangService}
}

// NewOrganizationHandlerWithServices creates a new organization handler with all services
func NewOrganizationHandlerWithServices(
	service *service.OrganizationService,
	bidangService *service.BidangService,
	activityService *service.OrganizationActivityService,
) *OrganizationHandler {
	return &OrganizationHandler{
		service:         service,
		bidangService:   bidangService,
		activityService: activityService,
	}
}

// List returns paginated organizations
// GET /api/v1/organizations
// Super admins see all organizations, org admins see only their organizations
func (h *OrganizationHandler) List(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	search := c.Query("search")

	var isActive *bool
	if activeParam := c.Query("is_active"); activeParam != "" {
		active := activeParam == "true"
		isActive = &active
	}

	filter := repository.OrganizationFilter{
		Search:   search,
		IsActive: isActive,
		Page:     page,
		PageSize: pageSize,
	}

	// Get user's organization IDs for filtering
	// For super_admin, orgIDs will be nil (no filtering)
	// For org_admin, orgIDs will contain their organization IDs
	orgIDs, ok := auth.GetUserOrgIDs(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Authentication required",
		})
		return
	}

	result, err := h.service.ListWithOrgFilter(c.Request.Context(), filter, orgIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch organizations",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// GetByID returns an organization by ID
// GET /api/v1/organizations/:id
func (h *OrganizationHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	// Parse and check access
	orgUUID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid organization ID format",
		})
		return
	}

	// Check if user can access this organization
	if !auth.CanAccessOrganization(c, orgUUID) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Access denied to this organization",
		})
		return
	}

	org, err := h.service.GetByIDWithRelations(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Organization not found",
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
			"organization": org,
			"stats":        stats,
		},
	})
}

// Create creates a new organization
// POST /api/v1/organizations
func (h *OrganizationHandler) Create(c *gin.Context) {
	var input service.CreateOrganizationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	org, err := h.service.Create(c.Request.Context(), input)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "slug already exists" || err.Error() == "invalid slug format" {
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
		"data":    org,
	})
}

// Update updates an organization
// PUT /api/v1/organizations/:id
// Org admins can only update their own organization
func (h *OrganizationHandler) Update(c *gin.Context) {
	id := c.Param("id")

	// Parse and check access
	orgUUID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid organization ID format",
		})
		return
	}

	// Check if user can manage this organization
	if !auth.CanManageOrganization(c, orgUUID) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Access denied to manage this organization",
		})
		return
	}

	var input service.UpdateOrganizationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	org, err := h.service.Update(c.Request.Context(), id, input)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "slug already exists" || err.Error() == "invalid slug format" || err.Error() == "invalid organization ID format" {
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
		"data":    org,
	})
}

// Delete soft-deletes an organization
// DELETE /api/v1/organizations/:id
func (h *OrganizationHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to delete organization",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Organization deleted successfully",
	})
}

// GetStats returns statistics for an organization
// GET /api/v1/organizations/:id/stats
func (h *OrganizationHandler) GetStats(c *gin.Context) {
	id := c.Param("id")

	// Parse and check access
	orgUUID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid organization ID format",
		})
		return
	}

	// Check if user can access this organization
	if !auth.CanAccessOrganization(c, orgUUID) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Access denied to this organization",
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

// AddMemberInput represents input for adding a member
type AddMemberInput struct {
	UserID string              `json:"user_id" binding:"required"`
	Role   model.OrgMemberRole `json:"role" binding:"required"`
}

// AddMember adds a user to an organization
// POST /api/v1/organizations/:id/members
// Only org admin can add members to their organization
func (h *OrganizationHandler) AddMember(c *gin.Context) {
	orgID := c.Param("id")

	// Parse and check access
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid organization ID format",
		})
		return
	}

	// Check if user can manage this organization
	if !auth.CanManageOrganization(c, orgUUID) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Access denied to manage this organization",
		})
		return
	}

	var input AddMemberInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	if err := h.service.AddMember(c.Request.Context(), orgID, input.UserID, input.Role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Member added successfully",
	})
}

// RemoveMember removes a user from an organization
// DELETE /api/v1/organizations/:id/members/:user_id
// Only org admin can remove members from their organization
func (h *OrganizationHandler) RemoveMember(c *gin.Context) {
	orgID := c.Param("id")
	userID := c.Param("user_id")

	// Parse and check access
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid organization ID format",
		})
		return
	}

	// Check if user can manage this organization
	if !auth.CanManageOrganization(c, orgUUID) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Access denied to manage this organization",
		})
		return
	}

	if err := h.service.RemoveMember(c.Request.Context(), orgID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Member removed successfully",
	})
}

// UpdateMemberRoleInput represents input for updating a member's role
type UpdateMemberRoleInput struct {
	Role model.OrgMemberRole `json:"role" binding:"required"`
}

// UpdateMemberRole updates a member's role in an organization
// PUT /api/v1/organizations/:id/members/:user_id/role
// Only org admin can update member roles in their organization
func (h *OrganizationHandler) UpdateMemberRole(c *gin.Context) {
	orgID := c.Param("id")
	userID := c.Param("user_id")

	// Parse and check access
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid organization ID format",
		})
		return
	}

	// Check if user can manage this organization
	if !auth.CanManageOrganization(c, orgUUID) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Access denied to manage this organization",
		})
		return
	}

	var input UpdateMemberRoleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	if err := h.service.UpdateMemberRole(c.Request.Context(), orgID, userID, input.Role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Member role updated successfully",
	})
}

// AddBidangInput represents input for adding a bidang to an organization
type AddBidangInput struct {
	BidangID string `json:"bidang_id" binding:"required"`
}

// AddBidang adds a bidang to an organization
// POST /api/v1/organizations/:id/bidang
// Only org admin can add bidangs to their organization
func (h *OrganizationHandler) AddBidang(c *gin.Context) {
	orgID := c.Param("id")

	// Parse and check access
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid organization ID format",
		})
		return
	}

	// Check if user can manage this organization
	if !auth.CanManageOrganization(c, orgUUID) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Access denied to manage this organization",
		})
		return
	}

	var input AddBidangInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	if err := h.bidangService.AddToOrganization(c.Request.Context(), orgID, input.BidangID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Bidang added successfully",
	})
}

// RemoveBidang removes a bidang from an organization
// DELETE /api/v1/organizations/:id/bidang/:bidang_id
// Only org admin can remove bidangs from their organization
func (h *OrganizationHandler) RemoveBidang(c *gin.Context) {
	orgID := c.Param("id")
	bidangID := c.Param("bidang_id")

	// Parse and check access
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid organization ID format",
		})
		return
	}

	// Check if user can manage this organization
	if !auth.CanManageOrganization(c, orgUUID) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Access denied to manage this organization",
		})
		return
	}

	if err := h.bidangService.RemoveFromOrganization(c.Request.Context(), orgID, bidangID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Bidang removed successfully",
	})
}

// GetActivities returns paginated activities (feeds) from organization's relawan
// GET /api/v1/organizations/:id/activities
// Only org members can access their organization's activities
func (h *OrganizationHandler) GetActivities(c *gin.Context) {
	id := c.Param("id")

	// Parse and check access
	orgUUID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid organization ID format",
		})
		return
	}

	// Check if user can access this organization
	if !auth.CanAccessOrganization(c, orgUUID) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Access denied to this organization",
		})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// Get activities from service
	activities, total, err := h.activityService.GetOrganizationActivities(c.Request.Context(), orgUUID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch activities",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"activities": activities,
			"pagination": gin.H{
				"page":       page,
				"page_size":  pageSize,
				"total":      total,
				"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
			},
		},
	})
}
