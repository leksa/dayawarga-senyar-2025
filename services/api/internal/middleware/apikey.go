package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIKeyAuth creates a middleware that validates API key from header or query param
func APIKeyAuth(validKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip if no key is configured (empty string means disabled)
		if validKey == "" {
			c.Next()
			return
		}

		// Check header first: X-API-Key
		apiKey := c.GetHeader("X-API-Key")

		// Fallback to query param: ?api_key=xxx
		if apiKey == "" {
			apiKey = c.Query("api_key")
		}

		if apiKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "API key required. Provide via X-API-Key header or api_key query param.",
			})
			return
		}

		if apiKey != validKey {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Invalid API key",
			})
			return
		}

		c.Next()
	}
}
