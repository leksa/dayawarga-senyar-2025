package validator

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestValidatePagination(t *testing.T) {
	tests := []struct {
		name          string
		queryParams   string
		expectedPage  int
		expectedLimit int
		expectedValid bool
	}{
		{
			name:          "Default values",
			queryParams:   "",
			expectedPage:  1,
			expectedLimit: 50,
			expectedValid: true,
		},
		{
			name:          "Valid page and limit",
			queryParams:   "?page=2&limit=100",
			expectedPage:  2,
			expectedLimit: 100,
			expectedValid: true,
		},
		{
			name:          "Page at boundary (1)",
			queryParams:   "?page=1",
			expectedPage:  1,
			expectedLimit: 50,
			expectedValid: true,
		},
		{
			name:          "Limit at boundary (200)",
			queryParams:   "?limit=200",
			expectedPage:  1,
			expectedLimit: 200,
			expectedValid: true,
		},
		{
			name:          "Limit over cap (should cap at 200)",
			queryParams:   "?limit=500",
			expectedPage:  1,
			expectedLimit: 200,
			expectedValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.GET("/test", func(c *gin.Context) {
				result := ValidatePagination(c)
				assert.Equal(t, tt.expectedPage, result.Page)
				assert.Equal(t, tt.expectedLimit, result.Limit)
				assert.Equal(t, tt.expectedValid, result.Valid)
				c.Status(200)
			})

			req := httptest.NewRequest("GET", "/test"+tt.queryParams, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// If expected valid, status should be 200, otherwise should be 400
			if tt.expectedValid {
				assert.Equal(t, 200, w.Code)
			}
		})
	}
}

func TestValidateBBox(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		expectedValid  bool
		expectedCoords *BBoxValidation
	}{
		{
			name:           "No bbox provided",
			queryParams:    "",
			expectedValid:  true,
			expectedCoords: nil,
		},
		{
			name:          "Valid bbox",
			queryParams:   "?bbox=-180,-90,180,90",
			expectedValid: true,
			expectedCoords: &BBoxValidation{
				MinLng: -180,
				MinLat: -90,
				MaxLng: 180,
				MaxLat: 90,
				Valid:  true,
			},
		},
		{
			name:          "Valid bbox for Aceh",
			queryParams:   "?bbox=95,-6,98,5",
			expectedValid: true,
			expectedCoords: &BBoxValidation{
				MinLng: 95,
				MinLat: -6,
				MaxLng: 98,
				MaxLat: 5,
				Valid:  true,
			},
		},
		{
			name:           "Invalid format (not enough parts)",
			queryParams:    "?bbox=95,-6,98",
			expectedValid:  false,
			expectedCoords: &BBoxValidation{Valid: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.GET("/test", func(c *gin.Context) {
				result := ValidateBBox(c)
				assert.Equal(t, tt.expectedValid, result.Valid)

				if tt.expectedCoords != nil && result.Valid {
					assert.Equal(t, tt.expectedCoords.MinLng, result.MinLng)
					assert.Equal(t, tt.expectedCoords.MinLat, result.MinLat)
					assert.Equal(t, tt.expectedCoords.MaxLng, result.MaxLng)
					assert.Equal(t, tt.expectedCoords.MaxLat, result.MaxLat)
				}

				c.Status(200)
			})

			req := httptest.NewRequest("GET", "/test"+tt.queryParams, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// If expected valid, status should be 200, otherwise should be 400
			if tt.expectedValid {
				assert.Equal(t, 200, w.Code)
			}
		})
	}
}

func TestSanitizeSearchString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normal string",
			input:    "Posko Pengungsi",
			expected: "Posko Pengungsi",
		},
		{
			name:     "String with numbers",
			input:    "Posko 123",
			expected: "Posko 123",
		},
		{
			name:     "String with allowed special chars",
			input:    "Posko-Aceh_Barat.ind",
			expected: "Posko-Aceh_Barat.ind",
		},
		{
			name:     "String with SQL injection attempt",
			input:    "Posko'; DROP TABLE locations--",
			expected: "Posko DROP TABLE locations--",
		},
		{
			name:     "String with quotes",
			input:    "\"Posko 'Test'\"",
			expected: "Posko Test",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeSearchString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateUUID(t *testing.T) {
	tests := []struct {
		name       string
		param      string
		expectedID string
		expectedOK bool
	}{
		{
			name:       "Valid UUID",
			param:      "123e4567-e89b-12d3-a456-426614174000",
			expectedID: "123e4567-e89b-12d3-a456-426614174000",
			expectedOK: true,
		},
		{
			name:       "Invalid UUID (too short)",
			param:      "12345678",
			expectedID: "",
			expectedOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.GET("/test/:id", func(c *gin.Context) {
				id, ok := ValidateUUID(c, "id", "Test ID")
				assert.Equal(t, tt.expectedID, id)
				assert.Equal(t, tt.expectedOK, ok)
				c.Status(200)
			})

			req := httptest.NewRequest("GET", "/test/"+tt.param, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		})
	}
}
