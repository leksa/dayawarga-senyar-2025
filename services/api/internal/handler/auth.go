package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leksa/datamapper-senyar/internal/auth"
	"github.com/leksa/datamapper-senyar/internal/service"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	userService *service.UserService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(userService *service.UserService) *AuthHandler {
	return &AuthHandler{userService: userService}
}

// MeResponse represents the /auth/me response
type MeResponse struct {
	ID        string               `json:"id"`
	Email     string               `json:"email"`
	Name      *string              `json:"name,omitempty"`
	AvatarURL *string              `json:"avatar_url,omitempty"`
	Role      string               `json:"role"`
	IsActive  bool                 `json:"is_active"`
	Orgs      []OrgMembershipBrief `json:"organizations,omitempty"`
}

// OrgMembershipBrief represents a brief organization membership
type OrgMembershipBrief struct {
	OrgID   string `json:"id"`
	OrgName string `json:"name"`
	OrgSlug string `json:"slug"`
	Role    string `json:"role"`
}

// Me handles GET /auth/me - returns current authenticated user info
func (h *AuthHandler) Me(c *gin.Context) {
	user := auth.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Not authenticated",
		})
		return
	}

	// Get user with organization memberships
	userWithOrgs, err := h.userService.GetWithOrganizations(c.Request.Context(), user.ID.String())
	if err != nil {
		// If we can't load orgs, just return basic user info
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": MeResponse{
				ID:        user.ID.String(),
				Email:     user.Email,
				Name:      user.Name,
				AvatarURL: user.AvatarURL,
				Role:      string(user.Role),
				IsActive:  user.IsActive,
				Orgs:      []OrgMembershipBrief{},
			},
		})
		return
	}

	// Build organization memberships
	var orgs []OrgMembershipBrief
	for _, membership := range userWithOrgs.OrganizationMemberships {
		if membership.Organization != nil {
			orgs = append(orgs, OrgMembershipBrief{
				OrgID:   membership.Organization.ID.String(),
				OrgName: membership.Organization.Name,
				OrgSlug: membership.Organization.Slug,
				Role:    string(membership.Role),
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": MeResponse{
			ID:        userWithOrgs.ID.String(),
			Email:     userWithOrgs.Email,
			Name:      userWithOrgs.Name,
			AvatarURL: userWithOrgs.AvatarURL,
			Role:      string(userWithOrgs.Role),
			IsActive:  userWithOrgs.IsActive,
			Orgs:      orgs,
		},
	})
}
