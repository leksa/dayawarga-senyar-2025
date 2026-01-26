package auth

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/leksa/datamapper-senyar/internal/model"
)

// OrganizationChecker interface for checking organization membership
type OrganizationChecker interface {
	GetUserOrganizations(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
	IsUserOrgAdmin(ctx context.Context, userID, orgID uuid.UUID) (bool, error)
}

// Context key for organization checker
const ContextKeyOrgChecker contextKey = "org_checker"

// SetOrgChecker sets the organization checker in context
func SetOrgChecker(c *gin.Context, checker OrganizationChecker) {
	c.Set(string(ContextKeyOrgChecker), checker)
}

// GetOrgChecker retrieves the organization checker from context
func GetOrgChecker(c *gin.Context) OrganizationChecker {
	if checker, exists := c.Get(string(ContextKeyOrgChecker)); exists {
		if orgChecker, ok := checker.(OrganizationChecker); ok {
			return orgChecker
		}
	}
	return nil
}

// RequireRole creates a middleware that requires the user to have one of the specified roles
func RequireRole(roles ...model.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetUser(c)
		if user == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Authentication required",
			})
			return
		}

		// Check if user has one of the required roles
		for _, role := range roles {
			if user.Role == role {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Insufficient permissions",
		})
	}
}

// RequireSuperAdmin creates a middleware that requires super_admin role
func RequireSuperAdmin() gin.HandlerFunc {
	return RequireRole(model.UserRoleSuperAdmin)
}

// RequireOrgAdminOrAbove creates a middleware that requires org_admin or super_admin role
func RequireOrgAdminOrAbove() gin.HandlerFunc {
	return RequireRole(model.UserRoleSuperAdmin, model.UserRoleOrgAdmin)
}

// RequireActiveUser creates a middleware that requires the user to be active
func RequireActiveUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetUser(c)
		if user == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Authentication required",
			})
			return
		}

		if !user.IsActive {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "User account is deactivated",
			})
			return
		}

		c.Next()
	}
}

// CanAccessOrganization checks if user can access a specific organization
// Super admins can access all organizations
// Org admins can only access their own organization(s)
func CanAccessOrganization(c *gin.Context, orgID uuid.UUID) bool {
	user := GetUser(c)
	if user == nil {
		return false
	}

	// Super admin can access everything
	if user.IsSuperAdmin() {
		return true
	}

	// For org_admin, check if they belong to this organization
	checker := GetOrgChecker(c)
	if checker == nil {
		return false
	}

	orgs, err := checker.GetUserOrganizations(c.Request.Context(), user.ID)
	if err != nil {
		return false
	}

	for _, id := range orgs {
		if id == orgID {
			return true
		}
	}

	return false
}

// CanManageOrganization checks if user can manage (edit/delete) a specific organization
// Super admins can manage all organizations
// Org admins can manage their own organization(s) where they are admin
func CanManageOrganization(c *gin.Context, orgID uuid.UUID) bool {
	user := GetUser(c)
	if user == nil {
		return false
	}

	// Super admin can manage everything
	if user.IsSuperAdmin() {
		return true
	}

	// For org_admin, check if they are admin of this organization
	checker := GetOrgChecker(c)
	if checker == nil {
		return false
	}

	isAdmin, err := checker.IsUserOrgAdmin(c.Request.Context(), user.ID, orgID)
	if err != nil {
		return false
	}

	return isAdmin
}

// GetUserOrgIDs returns the organization IDs that the current user can access
// Super admins return nil (meaning all orgs)
// Org admins return their organization IDs
func GetUserOrgIDs(c *gin.Context) ([]uuid.UUID, bool) {
	user := GetUser(c)
	if user == nil {
		return nil, false
	}

	// Super admin can access all - return nil to indicate no filtering needed
	if user.IsSuperAdmin() {
		return nil, true
	}

	// For org_admin, get their organizations
	checker := GetOrgChecker(c)
	if checker == nil {
		return []uuid.UUID{}, false
	}

	orgs, err := checker.GetUserOrganizations(c.Request.Context(), user.ID)
	if err != nil {
		return []uuid.UUID{}, false
	}

	return orgs, true
}
