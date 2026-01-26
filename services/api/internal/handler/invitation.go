package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/leksa/datamapper-senyar/internal/auth"
	"github.com/leksa/datamapper-senyar/internal/model"
	"github.com/leksa/datamapper-senyar/internal/service"
)

// InvitationHandler handles HTTP requests for user invitations
type InvitationHandler struct {
	invitationService *service.InvitationService
	orgService        *service.OrganizationService
}

// NewInvitationHandler creates a new invitation handler
func NewInvitationHandler(invitationService *service.InvitationService, orgService *service.OrganizationService) *InvitationHandler {
	return &InvitationHandler{
		invitationService: invitationService,
		orgService:        orgService,
	}
}

// CreateOrganizationWithAdminInput represents input for creating org with admin
type CreateOrganizationWithAdminInput struct {
	// Organization fields
	Name         string  `json:"name" binding:"required"`
	Slug         string  `json:"slug"`
	Description  *string `json:"description"`
	Email        *string `json:"email"`
	Phone        *string `json:"phone"`
	Address      *string `json:"address"`
	LogoURL      *string `json:"logo_url"`
	ODKProjectID *int    `json:"odk_project_id"`

	// Admin invitation (required)
	AdminEmail string `json:"admin_email" binding:"required,email"`
	AdminName  string `json:"admin_name" binding:"required"`
}

// CreateOrganizationWithAdmin creates an organization and invites an admin
// POST /api/v1/organizations/with-admin
func (h *InvitationHandler) CreateOrganizationWithAdmin(c *gin.Context) {
	var input CreateOrganizationWithAdminInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	// Get current user from context (set by auth middleware)
	user := auth.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not found in context",
		})
		return
	}

	// Create organization first
	orgInput := service.CreateOrganizationInput{
		Name:         input.Name,
		Slug:         input.Slug,
		Description:  input.Description,
		Email:        input.Email,
		Phone:        input.Phone,
		Address:      input.Address,
		LogoURL:      input.LogoURL,
		ODKProjectID: input.ODKProjectID,
	}

	org, err := h.orgService.Create(c.Request.Context(), orgInput)
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

	// Now invite the admin
	inviteInput := service.InviteUserInput{
		Email:          input.AdminEmail,
		Name:           input.AdminName,
		OrganizationID: &org.ID,
		OrgRole:        model.OrgMemberRoleAdmin,
		InvitedBy:      user.ID,
	}

	inviteResult, err := h.invitationService.InviteUser(c.Request.Context(), inviteInput)
	if err != nil {
		// Log the error but don't fail the whole operation
		// Organization is created, admin invite failed
		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"data": gin.H{
				"organization":     org,
				"invitation_error": err.Error(),
			},
			"message": "Organization created but admin invitation failed",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"organization":    org,
			"invited_admin":   inviteResult.User,
			"invitation_link": inviteResult.InvitationLink,
			"is_new_admin":    inviteResult.IsNewUser,
		},
	})
}

// InviteUserInput represents input for inviting a user
type InviteUserInput struct {
	Email          string  `json:"email" binding:"required,email"`
	Name           string  `json:"name" binding:"required"`
	OrganizationID *string `json:"organization_id,omitempty"`
	OrgRole        string  `json:"org_role,omitempty"`
}

// InviteUser invites a new user to the platform
// POST /api/v1/invitations
func (h *InvitationHandler) InviteUser(c *gin.Context) {
	var input InviteUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	// Get current user from context
	user := auth.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not found in context",
		})
		return
	}

	// Parse organization ID if provided
	var orgID *uuid.UUID
	if input.OrganizationID != nil && *input.OrganizationID != "" {
		parsed, err := uuid.Parse(*input.OrganizationID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid organization ID format",
			})
			return
		}
		orgID = &parsed
	}

	// Default role
	orgRole := model.OrgMemberRoleMember
	if input.OrgRole == "admin" {
		orgRole = model.OrgMemberRoleAdmin
	}

	inviteInput := service.InviteUserInput{
		Email:          input.Email,
		Name:           input.Name,
		OrganizationID: orgID,
		OrgRole:        orgRole,
		InvitedBy:      user.ID,
	}

	result, err := h.invitationService.InviteUser(c.Request.Context(), inviteInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"user":            result.User,
			"invitation_link": result.InvitationLink,
			"is_new_user":     result.IsNewUser,
		},
	})
}

// ResendInvitation resends an invitation to a pending user
// POST /api/v1/invitations/:user_id/resend
func (h *InvitationHandler) ResendInvitation(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid user ID format",
		})
		return
	}

	result, err := h.invitationService.ResendInvitation(c.Request.Context(), userID)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "user not found" || err.Error() == "user is not pending invitation" {
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
		"data": gin.H{
			"user":            result.User,
			"invitation_link": result.InvitationLink,
		},
		"message": "Invitation resent successfully",
	})
}
