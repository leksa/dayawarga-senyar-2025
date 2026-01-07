package handler

import (
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/leksa/datamapper-senyar/internal/service"
)

// PhotoHandler handles photo-related HTTP requests
type PhotoHandler struct {
	photoService *service.PhotoService
}

// NewPhotoHandler creates a new photo handler
func NewPhotoHandler(photoService *service.PhotoService) *PhotoHandler {
	return &PhotoHandler{
		photoService: photoService,
	}
}

// GetPhotosByLocation returns all photos for a location
func (h *PhotoHandler) GetPhotosByLocation(c *gin.Context) {
	locationIDStr := c.Param("id")
	locationID, err := uuid.Parse(locationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid location ID",
		})
		return
	}

	photos, err := h.photoService.GetPhotosByLocation(locationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Build photo URLs
	type PhotoResponse struct {
		ID          string  `json:"id"`
		PhotoType   string  `json:"photo_type"`
		Filename    string  `json:"filename"`
		IsCached    bool    `json:"is_cached"`
		FileSize    *int    `json:"file_size,omitempty"`
		URL         string  `json:"url,omitempty"`
		StoragePath string  `json:"storage_path,omitempty"`
		CreatedAt   string  `json:"created_at"`
	}

	var response []PhotoResponse
	for _, photo := range photos {
		pr := PhotoResponse{
			ID:        photo.ID.String(),
			PhotoType: photo.PhotoType,
			Filename:  photo.Filename,
			IsCached:  photo.IsCached,
			FileSize:  photo.FileSize,
			CreatedAt: photo.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
		if photo.StoragePath != nil {
			pr.StoragePath = *photo.StoragePath
		}
		if photo.IsCached {
			pr.URL = "/api/v1/photos/" + photo.ID.String() + "/file"
		}
		response = append(response, pr)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetPhoto returns a single photo's metadata
func (h *PhotoHandler) GetPhoto(c *gin.Context) {
	photoIDStr := c.Param("id")
	photoID, err := uuid.Parse(photoIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid photo ID",
		})
		return
	}

	photos, err := h.photoService.GetPhotosByLocation(uuid.Nil) // This needs adjustment
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	for _, photo := range photos {
		if photo.ID == photoID {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    photo,
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"success": false,
		"error":   "photo not found",
	})
}

// GetPhotoFile serves the actual photo file
func (h *PhotoHandler) GetPhotoFile(c *gin.Context) {
	photoIDStr := c.Param("id")
	photoID, err := uuid.Parse(photoIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid photo ID",
		})
		return
	}

	// Get storage path
	storagePath, err := h.photoService.GetPhotoPath(photoID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// If S3 URL, redirect to it directly (more efficient)
	if strings.HasPrefix(storagePath, "http") {
		c.Redirect(http.StatusFound, storagePath)
		return
	}

	// Local file - stream it
	reader, filename, err := h.photoService.GetPhotoReader(photoID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	defer reader.Close()

	// Determine content type based on extension
	ext := filepath.Ext(filename)
	contentType := "application/octet-stream"
	switch ext {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".gif":
		contentType = "image/gif"
	case ".webp":
		contentType = "image/webp"
	}

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", "inline; filename="+filename)

	c.Stream(func(w io.Writer) bool {
		io.Copy(w, reader)
		return false
	})
}

// SyncPhotos triggers photo synchronization
func (h *PhotoHandler) SyncPhotos(c *gin.Context) {
	result, err := h.photoService.SyncAllPhotos()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// CleanupOrphaned removes orphaned photo files
func (h *PhotoHandler) CleanupOrphaned(c *gin.Context) {
	cleaned, err := h.photoService.CleanupOrphanedFiles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"cleaned_files": cleaned,
		},
	})
}

// GetFeedPhotoFile serves the actual feed photo file
func (h *PhotoHandler) GetFeedPhotoFile(c *gin.Context) {
	photoIDStr := c.Param("id")
	photoID, err := uuid.Parse(photoIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid photo ID",
		})
		return
	}

	// Get storage path
	storagePath, err := h.photoService.GetFeedPhotoPath(photoID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// If S3 URL, redirect to it directly
	if strings.HasPrefix(storagePath, "http") {
		c.Redirect(http.StatusFound, storagePath)
		return
	}

	// Local file - stream it
	reader, filename, err := h.photoService.GetFeedPhotoReader(photoID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	defer reader.Close()

	// Determine content type based on extension
	ext := filepath.Ext(filename)
	contentType := "application/octet-stream"
	switch ext {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".gif":
		contentType = "image/gif"
	case ".webp":
		contentType = "image/webp"
	}

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", "inline; filename="+filename)

	c.Stream(func(w io.Writer) bool {
		io.Copy(w, reader)
		return false
	})
}

// SyncFeedPhotos triggers feed photo synchronization
func (h *PhotoHandler) SyncFeedPhotos(c *gin.Context) {
	formID := c.Query("form_id")
	if formID == "" {
		formID = "form_feed_v1" // default feed form ID
	}

	result, err := h.photoService.SyncFeedPhotos(formID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// ========================================
// FASKES PHOTOS ENDPOINTS
// ========================================

// GetFaskesPhotoFile serves the actual faskes photo file
func (h *PhotoHandler) GetFaskesPhotoFile(c *gin.Context) {
	photoIDStr := c.Param("id")
	photoID, err := uuid.Parse(photoIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid photo ID",
		})
		return
	}

	// Get storage path
	storagePath, err := h.photoService.GetFaskesPhotoPath(photoID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// If S3 URL, redirect to it directly
	if strings.HasPrefix(storagePath, "http") {
		c.Redirect(http.StatusFound, storagePath)
		return
	}

	// Local file - stream it
	reader, filename, err := h.photoService.GetFaskesPhotoReader(photoID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	defer reader.Close()

	// Determine content type based on extension
	ext := filepath.Ext(filename)
	contentType := "application/octet-stream"
	switch ext {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".gif":
		contentType = "image/gif"
	case ".webp":
		contentType = "image/webp"
	}

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", "inline; filename="+filename)

	c.Stream(func(w io.Writer) bool {
		io.Copy(w, reader)
		return false
	})
}

// GetPhotosByFaskes returns all photos for a faskes
func (h *PhotoHandler) GetPhotosByFaskes(c *gin.Context) {
	faskesIDStr := c.Param("id")
	faskesID, err := uuid.Parse(faskesIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid faskes ID",
		})
		return
	}

	photos, err := h.photoService.GetFaskesPhotosByFaskesID(faskesID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Build photo URLs
	type PhotoResponse struct {
		ID        string `json:"id"`
		PhotoType string `json:"photo_type"`
		Filename  string `json:"filename"`
		IsCached  bool   `json:"is_cached"`
		FileSize  *int   `json:"file_size,omitempty"`
		URL       string `json:"url,omitempty"`
		CreatedAt string `json:"created_at"`
	}

	var response []PhotoResponse
	for _, photo := range photos {
		pr := PhotoResponse{
			ID:        photo.ID.String(),
			PhotoType: photo.PhotoType,
			Filename:  photo.Filename,
			IsCached:  photo.IsCached,
			FileSize:  photo.FileSize,
			CreatedAt: photo.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
		if photo.IsCached {
			pr.URL = "/api/v1/faskes/photos/" + photo.ID.String() + "/file"
		}
		response = append(response, pr)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// SyncFaskesPhotos triggers faskes photo synchronization
func (h *PhotoHandler) SyncFaskesPhotos(c *gin.Context) {
	formID := c.Query("form_id")
	if formID == "" {
		formID = "form_faskes_v1" // default faskes form ID
	}

	result, err := h.photoService.SyncFaskesPhotos(formID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// ========================================
// S3 MIGRATION ENDPOINT
// ========================================

// MigrateToS3 migrates all locally cached photos to S3
func (h *PhotoHandler) MigrateToS3(c *gin.Context) {
	result, err := h.photoService.MigrateToS3()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// ResetCache resets cache status for photos with missing local files
// Use ?force=true to reset ALL cached photos
func (h *PhotoHandler) ResetCache(c *gin.Context) {
	force := c.Query("force") == "true"

	result, err := h.photoService.ResetCacheForMissingFiles(force)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	message := "Cache reset complete. Run /sync/photos to re-download photos to S3."
	if force {
		message = "Force cache reset complete. All photos will be re-downloaded on next sync."
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
		"message": message,
	})
}
