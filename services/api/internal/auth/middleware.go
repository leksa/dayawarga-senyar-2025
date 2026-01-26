package auth

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leksa/datamapper-senyar/internal/model"
)

// Context keys for storing auth data
type contextKey string

const (
	ContextKeyClaims contextKey = "oidc_claims"
	ContextKeyUser   contextKey = "auth_user"
)

// UserService defines the interface for user operations needed by auth middleware
type UserService interface {
	FindOrCreateFromOIDC(ctx context.Context, claims *OIDCClaims) (*model.User, error)
	GetByOIDCSubject(ctx context.Context, subject string) (*model.User, error)
}

// Middleware creates a Gin middleware for OIDC authentication
func Middleware(validator *OIDCValidator, userService UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		token, err := ExtractBearerToken(c.GetHeader("Authorization"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		// Validate token
		claims, err := validator.ValidateToken(c.Request.Context(), token)
		if err != nil {
			log.Printf("[Auth] Token validation failed: %v", err)
			status := http.StatusUnauthorized
			if err == ErrTokenExpired {
				status = http.StatusUnauthorized
			}

			c.AbortWithStatusJSON(status, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		// Store claims in context
		c.Set(string(ContextKeyClaims), claims)

		// Find or create user in database
		user, err := userService.FindOrCreateFromOIDC(c.Request.Context(), claims)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to process user authentication",
			})
			return
		}

		// Check if user is active
		if !user.IsActive {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "User account is deactivated",
			})
			return
		}

		// Store user in context
		c.Set(string(ContextKeyUser), user)

		c.Next()
	}
}

// OptionalMiddleware creates a middleware that doesn't require auth but will process it if present
func OptionalMiddleware(validator *OIDCValidator, userService UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to extract token
		token, err := ExtractBearerToken(c.GetHeader("Authorization"))
		if err != nil {
			// No token, continue without auth
			c.Next()
			return
		}

		// Validate token
		claims, err := validator.ValidateToken(c.Request.Context(), token)
		if err != nil {
			// Invalid token, continue without auth
			c.Next()
			return
		}

		// Store claims in context
		c.Set(string(ContextKeyClaims), claims)

		// Find or create user
		user, err := userService.FindOrCreateFromOIDC(c.Request.Context(), claims)
		if err != nil {
			c.Next()
			return
		}

		if user.IsActive {
			c.Set(string(ContextKeyUser), user)
		}

		c.Next()
	}
}

// GetClaims retrieves OIDC claims from Gin context
func GetClaims(c *gin.Context) *OIDCClaims {
	if claims, exists := c.Get(string(ContextKeyClaims)); exists {
		if oidcClaims, ok := claims.(*OIDCClaims); ok {
			return oidcClaims
		}
	}
	return nil
}

// GetUser retrieves the authenticated user from Gin context
func GetUser(c *gin.Context) *model.User {
	if user, exists := c.Get(string(ContextKeyUser)); exists {
		if authUser, ok := user.(*model.User); ok {
			return authUser
		}
	}
	return nil
}

// IsAuthenticated checks if the request is authenticated
func IsAuthenticated(c *gin.Context) bool {
	return GetUser(c) != nil
}
