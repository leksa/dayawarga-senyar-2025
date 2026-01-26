package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/leksa/datamapper-senyar/internal/auth"
	"github.com/leksa/datamapper-senyar/internal/service"
)

// UserHandler handles user management endpoints
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// UserResponse represents a user in API responses
type UserResponse struct {
	ID           string  `json:"id"`
	Email        string  `json:"email"`
	Name         *string `json:"name,omitempty"`
	Role         string  `json:"role"`
	Status       string  `json:"status"`
	IsActive     bool    `json:"is_active"`
	ODKWebUserID *int    `json:"odk_web_user_id,omitempty"`
	ODKAppUserID *int    `json:"odk_app_user_id,omitempty"`
	LastLoginAt  *string `json:"last_login_at,omitempty"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

// List handles GET /users - returns list of all users (super_admin only)
func (h *UserHandler) List(c *gin.Context) {
	// Only super_admin can list all users
	user := auth.GetUser(c)
	if user == nil || !user.IsSuperAdmin() {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Only super admin can access user list",
		})
		return
	}

	// Parse query params
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	search := c.Query("search")
	role := c.Query("role")
	status := c.Query("status")

	filter := service.UserFilter{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
		Role:     role,
		Status:   status,
	}

	result, err := h.userService.List(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch users",
		})
		return
	}

	// Transform to response format
	users := make([]UserResponse, len(result.Users))
	for i, u := range result.Users {
		var lastLogin *string
		if u.LastLoginAt != nil {
			formatted := u.LastLoginAt.Format("2006-01-02T15:04:05Z07:00")
			lastLogin = &formatted
		}

		users[i] = UserResponse{
			ID:           u.ID.String(),
			Email:        u.Email,
			Name:         u.Name,
			Role:         string(u.Role),
			Status:       string(u.Status),
			IsActive:     u.IsActive,
			ODKWebUserID: u.ODKWebUserID,
			ODKAppUserID: u.ODKAppUserID,
			LastLoginAt:  lastLogin,
			CreatedAt:    u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:    u.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"users":       users,
			"total":       result.Total,
			"page":        result.Page,
			"page_size":   result.PageSize,
			"total_pages": result.TotalPages,
		},
	})
}

// Get handles GET /users/:id - returns a specific user
func (h *UserHandler) Get(c *gin.Context) {
	// Only super_admin can view user details
	currentUser := auth.GetUser(c)
	if currentUser == nil || !currentUser.IsSuperAdmin() {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Only super admin can access user details",
		})
		return
	}

	id := c.Param("id")
	user, err := h.userService.GetWithOrganizations(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "User not found",
		})
		return
	}

	var lastLogin *string
	if user.LastLoginAt != nil {
		formatted := user.LastLoginAt.Format("2006-01-02T15:04:05Z07:00")
		lastLogin = &formatted
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": UserResponse{
			ID:           user.ID.String(),
			Email:        user.Email,
			Name:         user.Name,
			Role:         string(user.Role),
			Status:       string(user.Status),
			IsActive:     user.IsActive,
			ODKWebUserID: user.ODKWebUserID,
			LastLoginAt:  lastLogin,
			CreatedAt:    user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:    user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	})
}

// UpdateUserInput represents the input for updating a user
type UpdateUserInput struct {
	Name     *string `json:"name"`
	Role     *string `json:"role"`
	Status   *string `json:"status"`
	IsActive *bool   `json:"is_active"`
}

// Update handles PUT /users/:id - updates a user
func (h *UserHandler) Update(c *gin.Context) {
	// Only super_admin can update users
	currentUser := auth.GetUser(c)
	if currentUser == nil || !currentUser.IsSuperAdmin() {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Only super admin can update users",
		})
		return
	}

	id := c.Param("id")

	var input UpdateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	// Build updates map
	updates := make(map[string]interface{})
	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.Role != nil {
		// Validate role
		if *input.Role != "super_admin" && *input.Role != "org_admin" && *input.Role != "member" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid role. Must be one of: super_admin, org_admin, member",
			})
			return
		}
		updates["role"] = *input.Role
	}
	if input.Status != nil {
		// Validate status
		if *input.Status != "active" && *input.Status != "pending_invitation" && *input.Status != "suspended" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid status. Must be one of: active, pending_invitation, suspended",
			})
			return
		}
		updates["status"] = *input.Status
	}
	if input.IsActive != nil {
		updates["is_active"] = *input.IsActive
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "No fields to update",
		})
		return
	}

	user, err := h.userService.Update(c.Request.Context(), id, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update user",
		})
		return
	}

	var lastLogin *string
	if user.LastLoginAt != nil {
		formatted := user.LastLoginAt.Format("2006-01-02T15:04:05Z07:00")
		lastLogin = &formatted
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": UserResponse{
			ID:           user.ID.String(),
			Email:        user.Email,
			Name:         user.Name,
			Role:         string(user.Role),
			Status:       string(user.Status),
			IsActive:     user.IsActive,
			ODKWebUserID: user.ODKWebUserID,
			LastLoginAt:  lastLogin,
			CreatedAt:    user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:    user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	})
}
