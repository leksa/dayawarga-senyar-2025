package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/leksa/datamapper-senyar/internal/auth"
	"github.com/leksa/datamapper-senyar/internal/model"
	"github.com/leksa/datamapper-senyar/internal/service"
)

type InvitationHandler struct {
	invitationService *service.InvitationService
	orgService        *service.OrganizationService
}

func NewInvitationHandler(invitationService *service.InvitationService, orgService *service.OrganizationService) *InvitationHandler {
	return &InvitationHandler{
		invitationService: invitationService,
		orgService:        orgService,
	}
}

type CreateOrganizationWithAdminInput struct {
	Name         string  `json:"name" binding:"required"`
	Slug         string  `json:"slug"`
	Description  *string `json:"description"`
	Email        *string `json:"email"`
	Phone        *string `json:"phone"`
	Address      *string `json:"address"`
	LogoURL      *string `json:"logo_url"`
	ODKProjectID *int    `json:"odk_project_id"`
	AdminEmail   string  `json:"admin_email" binding:"required,email"`
	AdminName    string  `json:"admin_name" binding:"required"`
}

func (h *InvitationHandler) CreateOrganizationWithAdmin(c *gin.Context) {
	var input CreateOrganizationWithAdminInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	user := auth.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not found in context",
		})
		return
	}

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

	inviteInput := service.InviteUserInput{
		Email:          input.AdminEmail,
		Name:           input.AdminName,
		OrganizationID: &org.ID,
		OrgRole:        model.OrgMemberRoleAdmin,
		InvitedBy:      user.ID,
	}

	inviteResult, err := h.invitationService.InviteUser(c.Request.Context(), inviteInput)
	if err != nil {
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
			"email_sent":      inviteResult.EmailSent,
		},
	})
}

type InviteUserInput struct {
	Email          string  `json:"email" binding:"required,email"`
	Name           string  `json:"name" binding:"required"`
	OrganizationID *string `json:"organization_id,omitempty"`
	OrgRole        string  `json:"org_role,omitempty"`
}

func (h *InvitationHandler) InviteUser(c *gin.Context) {
	var input InviteUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	user := auth.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not found in context",
		})
		return
	}

	// Permission checks for org_admin (super_admin bypasses all restrictions)
	if user.Role != model.UserRoleSuperAdmin {
		// org_admin must provide organization_id
		if input.OrganizationID == nil || *input.OrganizationID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "organization_id is required for org_admin",
			})
			return
		}

		// Check if org_admin belongs to the organization
		if !auth.CanManageOrganization(c, uuid.MustParse(*input.OrganizationID)) {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Access denied to this organization",
			})
			return
		}

		// org_admin cannot invite with admin role
		if input.OrgRole == "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "org_admin cannot invite with admin role",
			})
			return
		}
	}

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

func (h *InvitationHandler) CancelInvitation(c *gin.Context) {
	userID := c.Param("user_id")
	if _, err := uuid.Parse(userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid user ID format",
		})
		return
	}

	if err := h.invitationService.CancelInvitation(c.Request.Context(), userID); err != nil {
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
		"message": "Invitation cancelled successfully",
	})
}

func (h *InvitationHandler) ValidateToken(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Token is required",
		})
		return
	}

	user, err := h.invitationService.ValidateToken(c.Request.Context(), token)
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
			"email": user.Email,
			"name":  user.Name,
			"role":  user.Role,
		},
	})
}

type AcceptInvitationInput struct {
	Token       string `json:"token" binding:"required"`
	OIDCSubject string `json:"oidc_subject" binding:"required"`
}

func (h *InvitationHandler) AcceptInvitation(c *gin.Context) {
	var input AcceptInvitationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	user, err := h.invitationService.AcceptInvitation(c.Request.Context(), input.Token, input.OIDCSubject)
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
			"id":     user.ID.String(),
			"email":  user.Email,
			"name":   user.Name,
			"role":   user.Role,
			"status": user.Status,
		},
	})
}

type SetPasswordInput struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

func (h *InvitationHandler) SetPassword(c *gin.Context) {
	var input SetPasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	result, err := h.invitationService.SetPassword(c.Request.Context(), input.Token, input.Password)
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
			"user_id":     result.User.ID.String(),
			"email":       result.User.Email,
			"pin":         result.PIN,
			"pin_expires": result.PINExpires,
		},
	})
}

func (h *InvitationHandler) GetVerificationStatus(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid user ID format",
		})
		return
	}

	result, err := h.invitationService.GetVerificationStatus(c.Request.Context(), userID)
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
			"verified":    result.Verified,
			"verified_at": result.VerifiedAt,
		},
	})
}

func (h *InvitationHandler) RegeneratePIN(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid user ID format",
		})
		return
	}

	result, err := h.invitationService.RegeneratePIN(c.Request.Context(), userID)
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
			"pin":         result.PIN,
			"pin_expires": result.PINExpires,
		},
	})
}

type VerifyPINInput struct {
	PIN   string `json:"pin" binding:"required,len=6"`
	Phone string `json:"phone" binding:"required"`
}

func (h *InvitationHandler) VerifyPIN(c *gin.Context) {
	var input VerifyPINInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	user, err := h.invitationService.VerifyPIN(c.Request.Context(), input.PIN, input.Phone)
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
			"user_id": user.ID.String(),
			"email":   user.Email,
			"name":    user.Name,
			"status":  user.Status,
		},
		"message": "User verified successfully",
	})
}
