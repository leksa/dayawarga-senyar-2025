package validator

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// BBoxValidation represents validated bounding box coordinates
type BBoxValidation struct {
	MinLng float64
	MinLat float64
	MaxLng float64
	MaxLat float64
	Valid  bool
}

// PaginationValidation represents validated pagination parameters
type PaginationValidation struct {
	Page  int
	Limit int
	Valid bool
}

// ValidatePagination validates and sanitizes pagination parameters
func ValidatePagination(c *gin.Context) PaginationValidation {
	result := PaginationValidation{
		Valid: true,
	}

	// Default values
	result.Page = 1
	result.Limit = 50

	// Parse page
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			if page > 0 {
				result.Page = page
			} else {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error": gin.H{
						"code":    "VALIDATION_ERROR",
						"message": "Page must be greater than 0",
					},
				})
				result.Valid = false
				return result
			}
		}
	}

	// Parse limit
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			if limit > 0 && limit <= 200 {
				result.Limit = limit
			} else if limit > 200 {
				result.Limit = 200 // Cap at 200
			} else {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error": gin.H{
						"code":    "VALIDATION_ERROR",
						"message": "Limit must be between 1 and 200",
					},
				})
				result.Valid = false
				return result
			}
		}
	}

	return result
}

// ValidateBBox validates bounding box coordinates
// Format: bbox=minLng,minLat,maxLng,maxLat
func ValidateBBox(c *gin.Context) BBoxValidation {
	result := BBoxValidation{
		Valid: false, // Invalid by default, only valid if bbox parameter is provided and valid
	}

	bbox := c.Query("bbox")
	if bbox == "" {
		return result // No bbox parameter provided
	}

	// Parse bbox format
	parts, err := parseBBoxParts(bbox)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": fmt.Sprintf("Invalid bbox format. Expected: minLng,minLat,maxLng,maxLat. Error: %v", err),
			},
		})
		result.Valid = false
		return result
	}

	if len(parts) != 4 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid bbox format. Expected 4 coordinates: minLng,minLat,maxLng,maxLat",
			},
		})
		result.Valid = false
		return result
	}

	minLng, err := strconv.ParseFloat(parts[0], 64)
	if err != nil || !isValidLongitude(minLng) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid minLng. Must be between -180 and 180",
			},
		})
		result.Valid = false
		return result
	}

	minLat, err := strconv.ParseFloat(parts[1], 64)
	if err != nil || !isValidLatitude(minLat) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid minLat. Must be between -90 and 90",
			},
		})
		result.Valid = false
		return result
	}

	maxLng, err := strconv.ParseFloat(parts[2], 64)
	if err != nil || !isValidLongitude(maxLng) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid maxLng. Must be between -180 and 180",
			},
		})
		result.Valid = false
		return result
	}

	maxLat, err := strconv.ParseFloat(parts[3], 64)
	if err != nil || !isValidLatitude(maxLat) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid maxLat. Must be between -90 and 90",
			},
		})
		result.Valid = false
		return result
	}

	// Validate bbox is properly ordered
	if minLng >= maxLng {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid bbox: minLng must be less than maxLng",
			},
		})
		result.Valid = false
		return result
	}

	if minLat >= maxLat {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid bbox: minLat must be less than maxLat",
			},
		})
		result.Valid = false
		return result
	}

	result.MinLng = minLng
	result.MinLat = minLat
	result.MaxLng = maxLng
	result.MaxLat = maxLat

	return result
}

// parseBBoxParts splits bbox string into parts
func parseBBoxParts(bbox string) ([]string, error) {
	parts := make([]string, 0, 4)
	current := ""

	for _, ch := range bbox {
		if ch == ',' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

	if len(parts) != 4 {
		return nil, fmt.Errorf("expected 4 coordinates, got %d", len(parts))
	}

	return parts, nil
}

// isValidLongitude checks if longitude is valid
func isValidLongitude(lng float64) bool {
	return lng >= -180 && lng <= 180
}

// isValidLatitude checks if latitude is valid
func isValidLatitude(lat float64) bool {
	return lat >= -90 && lat <= 90
}

// ValidateUUID validates if a string is a valid UUID format
func ValidateUUID(c *gin.Context, param string, fieldName string) (string, bool) {
	value := c.Param(param)
	if value == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": fmt.Sprintf("%s is required", fieldName),
			},
		})
		return "", false
	}

	// Basic UUID format validation (8-4-4-4-12)
	if len(value) != 36 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": fmt.Sprintf("Invalid %s format", fieldName),
			},
		})
		return "", false
	}

	return value, true
}

// SanitizeSearchString sanitizes search string to prevent injection
func SanitizeSearchString(search string) string {
	// Remove special SQL characters
	sanitized := ""
	for _, ch := range search {
		if (ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == ' ' || ch == '-' || ch == '_' || ch == '.' {
			sanitized += string(ch)
		}
	}
	return sanitized
}
