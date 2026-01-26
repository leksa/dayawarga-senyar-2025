package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/leksa/datamapper-senyar/internal/odk"
	"github.com/leksa/datamapper-senyar/internal/service"
)

// UserODKHandler handles user ODK-related API endpoints
type UserODKHandler struct {
	userODKService *service.UserODKService
}

// NewUserODKHandler creates a new user ODK handler
func NewUserODKHandler(userODKService *service.UserODKService) *UserODKHandler {
	return &UserODKHandler{
		userODKService: userODKService,
	}
}

// AssignProjectRoleInput represents input for assigning a project role
type AssignProjectRoleInput struct {
	ODKProjectID int    `json:"odk_project_id" binding:"required"`
	Role         string `json:"role" binding:"required,oneof=manager viewer"`
}

// AssignProjectRole assigns a user to an ODK project
// POST /api/v1/admin/users/:id/odk-roles
func (h *UserODKHandler) AssignProjectRole(c *gin.Context) {
	ctx := c.Request.Context()

	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid user ID",
		})
		return
	}

	var input AssignProjectRoleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	// Convert role string to role ID
	roleID := odk.RoleProjectViewer
	if input.Role == "manager" {
		roleID = odk.RoleProjectManager
	}

	result, err := h.userODKService.AssignProjectRole(ctx, service.AssignProjectRoleInput{
		UserID:       userID,
		ODKProjectID: input.ODKProjectID,
		RoleID:       roleID,
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
		"message": "User assigned to ODK project successfully",
	})
}

// RemoveProjectRoleInput represents input for removing a project role
type RemoveProjectRoleInput struct {
	Role string `json:"role" binding:"required,oneof=manager viewer"`
}

// RemoveProjectRole removes a user from an ODK project
// DELETE /api/v1/admin/users/:id/odk-roles/:projectId
func (h *UserODKHandler) RemoveProjectRole(c *gin.Context) {
	ctx := c.Request.Context()

	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid user ID",
		})
		return
	}

	projectID, err := strconv.Atoi(c.Param("projectId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid project ID",
		})
		return
	}

	var input RemoveProjectRoleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	// Convert role string to role ID
	roleID := odk.RoleProjectViewer
	if input.Role == "manager" {
		roleID = odk.RoleProjectManager
	}

	err = h.userODKService.RemoveProjectRole(ctx, service.RemoveProjectRoleInput{
		UserID:       userID,
		ODKProjectID: projectID,
		RoleID:       roleID,
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
		"message": "User removed from ODK project successfully",
	})
}

// GetUserProjectAssignments gets all ODK project assignments for a user
// GET /api/v1/admin/users/:id/odk-roles
func (h *UserODKHandler) GetUserProjectAssignments(c *gin.Context) {
	ctx := c.Request.Context()

	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid user ID",
		})
		return
	}

	assignments, err := h.userODKService.GetUserProjectAssignments(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    assignments,
	})
}

// GetProjectAssignments gets all user assignments for an ODK project
// GET /api/v1/admin/odk-projects/:id/assignments
func (h *UserODKHandler) GetProjectAssignments(c *gin.Context) {
	ctx := c.Request.Context()

	projectID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid project ID",
		})
		return
	}

	info, err := h.userODKService.GetProjectAssignments(ctx, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
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

// GetUserQRCode gets QR code data for a user's ODK Collect access
// GET /api/v1/admin/users/:id/odk-qr-code
func (h *UserODKHandler) GetUserQRCode(c *gin.Context) {
	ctx := c.Request.Context()

	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid user ID",
		})
		return
	}

	qrCode, err := h.userODKService.GetUserQRCode(ctx, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    qrCode,
	})
}
