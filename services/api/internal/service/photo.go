package service

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/leksa/datamapper-senyar/internal/model"
	"github.com/leksa/datamapper-senyar/internal/odk"
	"github.com/leksa/datamapper-senyar/internal/storage"
	"gorm.io/gorm"
)

// PhotoService handles photo storage and retrieval
type PhotoService struct {
	db          *gorm.DB
	odkClient   *odk.Client
	storagePath string
	s3Storage   *storage.S3Storage
	useS3       bool
}

// NewPhotoService creates a new photo service with local storage
func NewPhotoService(db *gorm.DB, odkClient *odk.Client, storagePath string) *PhotoService {
	// Create storage directory if it doesn't exist
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		log.Printf("Warning: failed to create storage directory: %v", err)
	}

	svc := &PhotoService{
		db:          db,
		odkClient:   odkClient,
		storagePath: storagePath,
		useS3:       false,
	}

	// Validate cache on startup - verify files exist for cached photos
	svc.ValidateCacheOnStartup()

	return svc
}

// NewPhotoServiceWithS3 creates a new photo service with S3 storage
func NewPhotoServiceWithS3(db *gorm.DB, odkClient *odk.Client, storagePath string, s3Storage *storage.S3Storage) *PhotoService {
	svc := &PhotoService{
		db:          db,
		odkClient:   odkClient,
		storagePath: storagePath,
		s3Storage:   s3Storage,
		useS3:       s3Storage != nil,
	}

	// Validate cache on startup - verify files exist for cached photos
	svc.ValidateCacheOnStartup()

	return svc
}

// DownloadAndSavePhoto downloads a photo from ODK Central and saves it to storage (S3 or local)
func (s *PhotoService) DownloadAndSavePhoto(photo *model.LocationPhoto, submissionID string) error {
	// Download from ODK Central
	data, err := s.odkClient.GetAttachment(submissionID, photo.Filename)
	if err != nil {
		return fmt.Errorf("failed to download attachment: %w", err)
	}

	// Generate unique filename
	ext := filepath.Ext(photo.Filename)
	newFilename := fmt.Sprintf("%s_%s%s", photo.PhotoType, uuid.New().String()[:8], ext)
	fileSize := len(data)

	var storagePath string

	if s.useS3 {
		// Upload to S3
		key := fmt.Sprintf("locations/%s/%s", photo.LocationID.String(), newFilename)
		contentType := getContentType(ext)
		url, err := s.s3Storage.Upload(context.Background(), key, data, contentType)
		if err != nil {
			return fmt.Errorf("failed to upload to S3: %w", err)
		}
		storagePath = url
		log.Printf("Uploaded photo to S3: %s -> %s", photo.Filename, url)
	} else {
		// Save to local filesystem
		locationDir := filepath.Join(s.storagePath, photo.LocationID.String())
		if err := os.MkdirAll(locationDir, 0755); err != nil {
			return fmt.Errorf("failed to create location directory: %w", err)
		}
		storagePath = filepath.Join(locationDir, newFilename)
		if err := os.WriteFile(storagePath, data, 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
		log.Printf("Downloaded photo: %s -> %s", photo.Filename, storagePath)
	}

	// Update database record
	photo.StoragePath = &storagePath
	photo.IsCached = true
	photo.FileSize = &fileSize

	if err := s.db.Save(photo).Error; err != nil {
		// Clean up if database update fails
		if s.useS3 {
			key := fmt.Sprintf("locations/%s/%s", photo.LocationID.String(), newFilename)
			s.s3Storage.Delete(context.Background(), key)
		} else {
			os.Remove(storagePath)
		}
		return fmt.Errorf("failed to update database: %w", err)
	}

	return nil
}

// getContentType returns the MIME type based on file extension
func getContentType(ext string) string {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}

// SyncPhotos downloads all uncached photos for a location
func (s *PhotoService) SyncPhotos(locationID uuid.UUID, submissionID string) (int, error) {
	var photos []model.LocationPhoto
	if err := s.db.Where("location_id = ? AND is_cached = false", locationID).Find(&photos).Error; err != nil {
		return 0, fmt.Errorf("failed to fetch photos: %w", err)
	}

	downloaded := 0
	for _, photo := range photos {
		if err := s.DownloadAndSavePhoto(&photo, submissionID); err != nil {
			log.Printf("Warning: failed to download photo %s: %v", photo.Filename, err)
			continue
		}
		downloaded++
	}

	return downloaded, nil
}

// SyncAllPhotos syncs all uncached photos across all locations
func (s *PhotoService) SyncAllPhotos() (*PhotoSyncResult, error) {
	result := &PhotoSyncResult{
		StartTime: time.Now(),
	}

	// Get all uncached photos with their location's submission ID
	var photos []struct {
		model.LocationPhoto
		ODKSubmissionID string `gorm:"column:odk_submission_id"`
	}

	err := s.db.Table("location_photos").
		Select("location_photos.*, locations.odk_submission_id").
		Joins("LEFT JOIN locations ON locations.id = location_photos.location_id").
		Where("location_photos.is_cached = false").
		Find(&photos).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch uncached photos: %w", err)
	}

	result.TotalFound = len(photos)

	for _, p := range photos {
		photo := p.LocationPhoto
		if err := s.DownloadAndSavePhoto(&photo, p.ODKSubmissionID); err != nil {
			result.Errors++
			result.ErrorDetails = append(result.ErrorDetails, fmt.Sprintf("%s: %v", photo.Filename, err))
			continue
		}
		result.Downloaded++
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime).String()

	return result, nil
}

// PhotoSyncResult holds the result of a photo sync operation
type PhotoSyncResult struct {
	TotalFound   int       `json:"total_found"`
	Downloaded   int       `json:"downloaded"`
	Errors       int       `json:"errors"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	Duration     string    `json:"duration"`
	ErrorDetails []string  `json:"error_details,omitempty"`
}

// GetPhotoPath returns the storage path for a photo
func (s *PhotoService) GetPhotoPath(photoID uuid.UUID) (string, error) {
	var photo model.LocationPhoto
	if err := s.db.First(&photo, photoID).Error; err != nil {
		return "", fmt.Errorf("photo not found: %w", err)
	}

	if photo.StoragePath == nil || *photo.StoragePath == "" {
		return "", fmt.Errorf("photo not cached")
	}

	return *photo.StoragePath, nil
}

// GetPhotosByLocation returns all photos for a location
func (s *PhotoService) GetPhotosByLocation(locationID uuid.UUID) ([]model.LocationPhoto, error) {
	var photos []model.LocationPhoto
	if err := s.db.Where("location_id = ?", locationID).Find(&photos).Error; err != nil {
		return nil, err
	}
	return photos, nil
}

// GetPhotoReader returns a reader for the photo file
func (s *PhotoService) GetPhotoReader(photoID uuid.UUID) (io.ReadCloser, string, error) {
	var photo model.LocationPhoto
	if err := s.db.First(&photo, photoID).Error; err != nil {
		return nil, "", fmt.Errorf("photo not found: %w", err)
	}

	if photo.StoragePath == nil || *photo.StoragePath == "" {
		return nil, "", fmt.Errorf("photo not cached")
	}

	storagePath := *photo.StoragePath

	// Check if it's an S3 URL
	if s.useS3 && strings.HasPrefix(storagePath, "http") {
		// Extract key from URL and get from S3
		key := extractS3Key(storagePath)
		reader, contentType, err := s.s3Storage.GetReader(context.Background(), key)
		if err != nil {
			return nil, "", fmt.Errorf("failed to get from S3: %w", err)
		}
		_ = contentType // We return filename, not content type
		return reader, filepath.Base(key), nil
	}

	// Local file
	file, err := os.Open(storagePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to open file: %w", err)
	}

	return file, filepath.Base(storagePath), nil
}

// extractS3Key extracts the S3 key from a full URL
// URL format: https://is3.cloudhost.id/bucket/prefix/path/to/file.ext
// Returns key WITHOUT the prefix (since S3Storage.GetReader adds prefix via buildKey)
func extractS3Key(url string) string {
	// Parse URL: https://is3.cloudhost.id/media/dayawarga/locations/uuid/file.png
	// We need: locations/uuid/file.png (without bucket and prefix)
	parts := strings.SplitN(url, "/", 6)
	// parts[0] = "https:"
	// parts[1] = ""
	// parts[2] = "is3.cloudhost.id"
	// parts[3] = "media" (bucket)
	// parts[4] = "dayawarga" (prefix)
	// parts[5] = "locations/uuid/file.png" (actual key)
	if len(parts) >= 6 {
		return parts[5]
	}
	// Fallback for URLs without prefix
	if len(parts) >= 5 {
		return parts[4]
	}
	return url
}

// DeletePhoto deletes a photo from storage and database
func (s *PhotoService) DeletePhoto(photoID uuid.UUID) error {
	var photo model.LocationPhoto
	if err := s.db.First(&photo, photoID).Error; err != nil {
		return fmt.Errorf("photo not found: %w", err)
	}

	// Delete file if exists
	if photo.StoragePath != nil && *photo.StoragePath != "" {
		os.Remove(*photo.StoragePath)
	}

	// Delete database record
	return s.db.Delete(&photo).Error
}

// CleanupOrphanedFiles removes files that don't have database records
func (s *PhotoService) CleanupOrphanedFiles() (int, error) {
	cleaned := 0

	err := filepath.Walk(s.storagePath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		// Check if file exists in database (location photos or feed photos)
		var count int64
		s.db.Model(&model.LocationPhoto{}).Where("storage_path = ?", path).Count(&count)

		if count == 0 {
			s.db.Model(&model.FeedPhoto{}).Where("storage_path = ?", path).Count(&count)
		}

		if count == 0 {
			if err := os.Remove(path); err == nil {
				cleaned++
				log.Printf("Cleaned up orphaned file: %s", path)
			}
		}

		return nil
	})

	return cleaned, err
}

// DownloadAndSaveFeedPhoto downloads a feed photo from ODK Central and saves it to storage (S3 or local)
func (s *PhotoService) DownloadAndSaveFeedPhoto(photo *model.FeedPhoto, submissionID string, formID string) error {
	// Download from ODK Central using the feed form
	data, err := s.odkClient.GetAttachmentForForm(formID, submissionID, photo.Filename)
	if err != nil {
		return fmt.Errorf("failed to download feed attachment: %w", err)
	}

	// Generate unique filename
	ext := filepath.Ext(photo.Filename)
	newFilename := fmt.Sprintf("%s_%s%s", photo.PhotoType, uuid.New().String()[:8], ext)
	fileSize := len(data)

	var storagePath string

	if s.useS3 {
		// Upload to S3
		key := fmt.Sprintf("feeds/%s/%s", photo.FeedID.String(), newFilename)
		contentType := getContentType(ext)
		url, err := s.s3Storage.Upload(context.Background(), key, data, contentType)
		if err != nil {
			return fmt.Errorf("failed to upload feed photo to S3: %w", err)
		}
		storagePath = url
		log.Printf("Uploaded feed photo to S3: %s -> %s", photo.Filename, url)
	} else {
		// Save to local filesystem
		feedDir := filepath.Join(s.storagePath, "feeds", photo.FeedID.String())
		if err := os.MkdirAll(feedDir, 0755); err != nil {
			return fmt.Errorf("failed to create feed directory: %w", err)
		}
		storagePath = filepath.Join(feedDir, newFilename)
		if err := os.WriteFile(storagePath, data, 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
		log.Printf("Downloaded feed photo: %s -> %s", photo.Filename, storagePath)
	}

	// Update database record
	photo.StoragePath = &storagePath
	photo.IsCached = true
	photo.FileSize = &fileSize

	if err := s.db.Save(photo).Error; err != nil {
		// Clean up if database update fails
		if s.useS3 {
			key := fmt.Sprintf("feeds/%s/%s", photo.FeedID.String(), newFilename)
			s.s3Storage.Delete(context.Background(), key)
		} else {
			os.Remove(storagePath)
		}
		return fmt.Errorf("failed to update database: %w", err)
	}

	return nil
}

// SyncFeedPhotos downloads all uncached feed photos
func (s *PhotoService) SyncFeedPhotos(formID string) (*PhotoSyncResult, error) {
	result := &PhotoSyncResult{
		StartTime: time.Now(),
	}

	// Get all uncached feed photos with their feed's submission ID
	var photos []struct {
		model.FeedPhoto
		ODKSubmissionID string `gorm:"column:odk_submission_id"`
	}

	err := s.db.Table("feed_photos").
		Select("feed_photos.*, information_feeds.odk_submission_id").
		Joins("LEFT JOIN information_feeds ON information_feeds.id = feed_photos.feed_id").
		Where("feed_photos.is_cached = false").
		Find(&photos).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch uncached feed photos: %w", err)
	}

	result.TotalFound = len(photos)

	for _, p := range photos {
		photo := p.FeedPhoto
		if p.ODKSubmissionID == "" {
			result.Errors++
			result.ErrorDetails = append(result.ErrorDetails, fmt.Sprintf("%s: missing submission ID", photo.Filename))
			continue
		}
		if err := s.DownloadAndSaveFeedPhoto(&photo, p.ODKSubmissionID, formID); err != nil {
			result.Errors++
			result.ErrorDetails = append(result.ErrorDetails, fmt.Sprintf("%s: %v", photo.Filename, err))
			continue
		}
		result.Downloaded++
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime).String()

	return result, nil
}

// GetFeedPhotoPath returns the storage path for a feed photo
func (s *PhotoService) GetFeedPhotoPath(photoID uuid.UUID) (string, error) {
	var photo model.FeedPhoto
	if err := s.db.First(&photo, photoID).Error; err != nil {
		return "", fmt.Errorf("feed photo not found: %w", err)
	}

	if photo.StoragePath == nil || *photo.StoragePath == "" {
		return "", fmt.Errorf("feed photo not cached")
	}

	return *photo.StoragePath, nil
}

// GetFeedPhotoReader returns a reader for the feed photo file
func (s *PhotoService) GetFeedPhotoReader(photoID uuid.UUID) (io.ReadCloser, string, error) {
	var photo model.FeedPhoto
	if err := s.db.First(&photo, photoID).Error; err != nil {
		return nil, "", fmt.Errorf("feed photo not found: %w", err)
	}

	if photo.StoragePath == nil || *photo.StoragePath == "" {
		return nil, "", fmt.Errorf("feed photo not cached")
	}

	storagePath := *photo.StoragePath

	// Check if it's an S3 URL
	if s.useS3 && strings.HasPrefix(storagePath, "http") {
		key := extractS3Key(storagePath)
		reader, _, err := s.s3Storage.GetReader(context.Background(), key)
		if err != nil {
			return nil, "", fmt.Errorf("failed to get feed photo from S3: %w", err)
		}
		return reader, filepath.Base(key), nil
	}

	// Local file
	file, err := os.Open(storagePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to open file: %w", err)
	}

	return file, filepath.Base(storagePath), nil
}

// GetFeedPhotoByID returns a feed photo by ID
func (s *PhotoService) GetFeedPhotoByID(photoID uuid.UUID) (*model.FeedPhoto, error) {
	var photo model.FeedPhoto
	if err := s.db.First(&photo, photoID).Error; err != nil {
		return nil, fmt.Errorf("feed photo not found: %w", err)
	}
	return &photo, nil
}

// ========================================
// FASKES PHOTOS
// ========================================

// DownloadAndSaveFaskesPhoto downloads a faskes photo from ODK Central and saves it to storage (S3 or local)
func (s *PhotoService) DownloadAndSaveFaskesPhoto(photo *model.FaskesPhoto, submissionID string, formID string) error {
	// Download from ODK Central using the faskes form
	data, err := s.odkClient.GetAttachmentForForm(formID, submissionID, photo.Filename)
	if err != nil {
		return fmt.Errorf("failed to download faskes attachment: %w", err)
	}

	// Generate unique filename
	ext := filepath.Ext(photo.Filename)
	newFilename := fmt.Sprintf("%s_%s%s", photo.PhotoType, uuid.New().String()[:8], ext)
	fileSize := len(data)

	var storagePath string

	if s.useS3 {
		// Upload to S3
		key := fmt.Sprintf("faskes/%s/%s", photo.FaskesID.String(), newFilename)
		contentType := getContentType(ext)
		url, err := s.s3Storage.Upload(context.Background(), key, data, contentType)
		if err != nil {
			return fmt.Errorf("failed to upload faskes photo to S3: %w", err)
		}
		storagePath = url
		log.Printf("Uploaded faskes photo to S3: %s -> %s", photo.Filename, url)
	} else {
		// Save to local filesystem
		faskesDir := filepath.Join(s.storagePath, "faskes", photo.FaskesID.String())
		if err := os.MkdirAll(faskesDir, 0755); err != nil {
			return fmt.Errorf("failed to create faskes directory: %w", err)
		}
		storagePath = filepath.Join(faskesDir, newFilename)
		if err := os.WriteFile(storagePath, data, 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
		log.Printf("Downloaded faskes photo: %s -> %s", photo.Filename, storagePath)
	}

	// Update database record
	photo.StoragePath = &storagePath
	photo.IsCached = true
	photo.FileSize = &fileSize

	if err := s.db.Save(photo).Error; err != nil {
		// Clean up if database update fails
		if s.useS3 {
			key := fmt.Sprintf("faskes/%s/%s", photo.FaskesID.String(), newFilename)
			s.s3Storage.Delete(context.Background(), key)
		} else {
			os.Remove(storagePath)
		}
		return fmt.Errorf("failed to update database: %w", err)
	}

	return nil
}

// SyncFaskesPhotos downloads all uncached faskes photos
func (s *PhotoService) SyncFaskesPhotos(formID string) (*PhotoSyncResult, error) {
	result := &PhotoSyncResult{
		StartTime: time.Now(),
	}

	// Get all uncached faskes photos with their faskes's submission ID
	var photos []struct {
		model.FaskesPhoto
		ODKSubmissionID string `gorm:"column:odk_submission_id"`
	}

	err := s.db.Table("faskes_photos").
		Select("faskes_photos.*, faskes.odk_submission_id").
		Joins("LEFT JOIN faskes ON faskes.id = faskes_photos.faskes_id").
		Where("faskes_photos.is_cached = false").
		Find(&photos).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch uncached faskes photos: %w", err)
	}

	result.TotalFound = len(photos)

	for _, p := range photos {
		photo := p.FaskesPhoto
		if p.ODKSubmissionID == "" {
			result.Errors++
			result.ErrorDetails = append(result.ErrorDetails, fmt.Sprintf("%s: missing submission ID", photo.Filename))
			continue
		}
		if err := s.DownloadAndSaveFaskesPhoto(&photo, p.ODKSubmissionID, formID); err != nil {
			result.Errors++
			result.ErrorDetails = append(result.ErrorDetails, fmt.Sprintf("%s: %v", photo.Filename, err))
			continue
		}
		result.Downloaded++
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime).String()

	return result, nil
}

// GetFaskesPhotoPath returns the storage path for a faskes photo
func (s *PhotoService) GetFaskesPhotoPath(photoID uuid.UUID) (string, error) {
	var photo model.FaskesPhoto
	if err := s.db.First(&photo, photoID).Error; err != nil {
		return "", fmt.Errorf("faskes photo not found: %w", err)
	}

	if photo.StoragePath == nil || *photo.StoragePath == "" {
		return "", fmt.Errorf("faskes photo not cached")
	}

	return *photo.StoragePath, nil
}

// GetFaskesPhotoReader returns a reader for the faskes photo file
func (s *PhotoService) GetFaskesPhotoReader(photoID uuid.UUID) (io.ReadCloser, string, error) {
	var photo model.FaskesPhoto
	if err := s.db.First(&photo, photoID).Error; err != nil {
		return nil, "", fmt.Errorf("faskes photo not found: %w", err)
	}

	if photo.StoragePath == nil || *photo.StoragePath == "" {
		return nil, "", fmt.Errorf("faskes photo not cached")
	}

	storagePath := *photo.StoragePath

	// Check if it's an S3 URL
	if s.useS3 && strings.HasPrefix(storagePath, "http") {
		key := extractS3Key(storagePath)
		reader, _, err := s.s3Storage.GetReader(context.Background(), key)
		if err != nil {
			return nil, "", fmt.Errorf("failed to get faskes photo from S3: %w", err)
		}
		return reader, filepath.Base(key), nil
	}

	// Local file
	file, err := os.Open(storagePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to open file: %w", err)
	}

	return file, filepath.Base(storagePath), nil
}

// GetFaskesPhotosByFaskesID returns all photos for a faskes
func (s *PhotoService) GetFaskesPhotosByFaskesID(faskesID uuid.UUID) ([]model.FaskesPhoto, error) {
	var photos []model.FaskesPhoto
	if err := s.db.Where("faskes_id = ?", faskesID).Find(&photos).Error; err != nil {
		return nil, err
	}
	return photos, nil
}

// ========================================
// CACHE VALIDATION ON STARTUP
// ========================================

// ValidateCacheOnStartup checks all photos marked as cached and verifies files exist
// For photos where files exist but is_cached=false, it updates the database
// For photos where is_cached=true but files are missing, it resets the cache status
func (s *PhotoService) ValidateCacheOnStartup() {
	log.Println("Validating photo cache on startup...")

	// Validate location photos
	locFixed, locReset := s.validateLocationPhotosCache()

	// Validate feed photos
	feedFixed, feedReset := s.validateFeedPhotosCache()

	// Validate faskes photos
	faskesFixed, faskesReset := s.validateFaskesPhotosCache()

	log.Printf("Photo cache validation complete: fixed %d, reset %d (location: %d/%d, feed: %d/%d, faskes: %d/%d)",
		locFixed+feedFixed+faskesFixed,
		locReset+feedReset+faskesReset,
		locFixed, locReset,
		feedFixed, feedReset,
		faskesFixed, faskesReset)
}

func (s *PhotoService) validateLocationPhotosCache() (fixed, reset int) {
	var photos []model.LocationPhoto
	// Get all photos with storage_path set (both cached and not cached, local files only)
	if err := s.db.Where("storage_path IS NOT NULL AND storage_path NOT LIKE 'http%'").Find(&photos).Error; err != nil {
		log.Printf("Warning: failed to fetch location photos for validation: %v", err)
		return 0, 0
	}

	for _, photo := range photos {
		if photo.StoragePath == nil {
			continue
		}
		_, err := os.Stat(*photo.StoragePath)
		fileExists := err == nil

		if fileExists && !photo.IsCached {
			// File exists but is_cached is false - fix it
			photo.IsCached = true
			if err := s.db.Save(&photo).Error; err == nil {
				fixed++
			}
		} else if !fileExists && photo.IsCached {
			// File missing but is_cached is true - reset it
			photo.IsCached = false
			photo.StoragePath = nil
			photo.FileSize = nil
			if err := s.db.Save(&photo).Error; err == nil {
				reset++
			}
		}
	}
	return fixed, reset
}

func (s *PhotoService) validateFeedPhotosCache() (fixed, reset int) {
	var photos []model.FeedPhoto
	if err := s.db.Where("storage_path IS NOT NULL AND storage_path NOT LIKE 'http%'").Find(&photos).Error; err != nil {
		log.Printf("Warning: failed to fetch feed photos for validation: %v", err)
		return 0, 0
	}

	for _, photo := range photos {
		if photo.StoragePath == nil {
			continue
		}
		_, err := os.Stat(*photo.StoragePath)
		fileExists := err == nil

		if fileExists && !photo.IsCached {
			photo.IsCached = true
			if err := s.db.Save(&photo).Error; err == nil {
				fixed++
			}
		} else if !fileExists && photo.IsCached {
			photo.IsCached = false
			photo.StoragePath = nil
			photo.FileSize = nil
			if err := s.db.Save(&photo).Error; err == nil {
				reset++
			}
		}
	}
	return fixed, reset
}

func (s *PhotoService) validateFaskesPhotosCache() (fixed, reset int) {
	var photos []model.FaskesPhoto
	if err := s.db.Where("storage_path IS NOT NULL AND storage_path NOT LIKE 'http%'").Find(&photos).Error; err != nil {
		log.Printf("Warning: failed to fetch faskes photos for validation: %v", err)
		return 0, 0
	}

	for _, photo := range photos {
		if photo.StoragePath == nil {
			continue
		}
		_, err := os.Stat(*photo.StoragePath)
		fileExists := err == nil

		if fileExists && !photo.IsCached {
			photo.IsCached = true
			if err := s.db.Save(&photo).Error; err == nil {
				fixed++
			}
		} else if !fileExists && photo.IsCached {
			photo.IsCached = false
			photo.StoragePath = nil
			photo.FileSize = nil
			if err := s.db.Save(&photo).Error; err == nil {
				reset++
			}
		}
	}
	return fixed, reset
}

// ========================================
// CACHE RESET
// ========================================

// ResetCacheResult holds the result of a cache reset operation
type ResetCacheResult struct {
	LocationPhotos int `json:"location_photos"`
	FeedPhotos     int `json:"feed_photos"`
	FaskesPhotos   int `json:"faskes_photos"`
	TotalReset     int `json:"total_reset"`
}

// ResetCacheForMissingFiles resets is_cached flag for photos whose local files are missing
// This allows them to be re-downloaded (to S3 if enabled)
// If force is true, it resets ALL cached photos regardless of file existence
func (s *PhotoService) ResetCacheForMissingFiles(force bool) (*ResetCacheResult, error) {
	result := &ResetCacheResult{}

	// If force mode, reset all cached photos that are not already on S3
	if force {
		// Reset all location photos not on S3
		res := s.db.Model(&model.LocationPhoto{}).
			Where("is_cached = true").
			Updates(map[string]interface{}{
				"is_cached":    false,
				"storage_path": nil,
				"file_size":    nil,
			})
		result.LocationPhotos = int(res.RowsAffected)

		// Reset all feed photos not on S3
		res = s.db.Model(&model.FeedPhoto{}).
			Where("is_cached = true").
			Updates(map[string]interface{}{
				"is_cached":    false,
				"storage_path": nil,
				"file_size":    nil,
			})
		result.FeedPhotos = int(res.RowsAffected)

		// Reset all faskes photos not on S3
		res = s.db.Model(&model.FaskesPhoto{}).
			Where("is_cached = true").
			Updates(map[string]interface{}{
				"is_cached":    false,
				"storage_path": nil,
				"file_size":    nil,
			})
		result.FaskesPhotos = int(res.RowsAffected)

		result.TotalReset = result.LocationPhotos + result.FeedPhotos + result.FaskesPhotos
		log.Printf("Force reset cache: %d location, %d feed, %d faskes photos",
			result.LocationPhotos, result.FeedPhotos, result.FaskesPhotos)
		return result, nil
	}

	// Reset location photos with missing local files
	var locationPhotos []model.LocationPhoto
	if err := s.db.Where("is_cached = true AND storage_path IS NOT NULL AND storage_path NOT LIKE 'http%'").Find(&locationPhotos).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch location photos: %w", err)
	}

	for _, photo := range locationPhotos {
		if photo.StoragePath == nil {
			continue
		}
		// Check if local file exists
		if _, err := os.Stat(*photo.StoragePath); os.IsNotExist(err) {
			// File doesn't exist, reset cache status
			photo.IsCached = false
			photo.StoragePath = nil
			photo.FileSize = nil
			if err := s.db.Save(&photo).Error; err != nil {
				log.Printf("Failed to reset cache for location photo %s: %v", photo.ID, err)
				continue
			}
			result.LocationPhotos++
			log.Printf("Reset cache for location photo: %s (file missing: %s)", photo.ID, *photo.StoragePath)
		}
	}

	// Reset feed photos with missing local files
	var feedPhotos []model.FeedPhoto
	if err := s.db.Where("is_cached = true AND storage_path IS NOT NULL AND storage_path NOT LIKE 'http%'").Find(&feedPhotos).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch feed photos: %w", err)
	}

	for _, photo := range feedPhotos {
		if photo.StoragePath == nil {
			continue
		}
		if _, err := os.Stat(*photo.StoragePath); os.IsNotExist(err) {
			photo.IsCached = false
			photo.StoragePath = nil
			photo.FileSize = nil
			if err := s.db.Save(&photo).Error; err != nil {
				log.Printf("Failed to reset cache for feed photo %s: %v", photo.ID, err)
				continue
			}
			result.FeedPhotos++
			log.Printf("Reset cache for feed photo: %s", photo.ID)
		}
	}

	// Reset faskes photos with missing local files
	var faskesPhotos []model.FaskesPhoto
	if err := s.db.Where("is_cached = true AND storage_path IS NOT NULL AND storage_path NOT LIKE 'http%'").Find(&faskesPhotos).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch faskes photos: %w", err)
	}

	for _, photo := range faskesPhotos {
		if photo.StoragePath == nil {
			continue
		}
		if _, err := os.Stat(*photo.StoragePath); os.IsNotExist(err) {
			photo.IsCached = false
			photo.StoragePath = nil
			photo.FileSize = nil
			if err := s.db.Save(&photo).Error; err != nil {
				log.Printf("Failed to reset cache for faskes photo %s: %v", photo.ID, err)
				continue
			}
			result.FaskesPhotos++
			log.Printf("Reset cache for faskes photo: %s", photo.ID)
		}
	}

	result.TotalReset = result.LocationPhotos + result.FeedPhotos + result.FaskesPhotos
	return result, nil
}

// ========================================
// S3 MIGRATION
// ========================================

// MigrationResult holds the result of a migration operation
type MigrationResult struct {
	LocationPhotos *PhotoSyncResult `json:"location_photos"`
	FeedPhotos     *PhotoSyncResult `json:"feed_photos"`
	FaskesPhotos   *PhotoSyncResult `json:"faskes_photos"`
	TotalMigrated  int              `json:"total_migrated"`
	TotalErrors    int              `json:"total_errors"`
	Duration       string           `json:"duration"`
}

// MigrateToS3 migrates all locally cached photos to S3
func (s *PhotoService) MigrateToS3() (*MigrationResult, error) {
	if !s.useS3 {
		return nil, fmt.Errorf("S3 storage is not enabled")
	}

	startTime := time.Now()
	result := &MigrationResult{}

	// Migrate location photos
	locationResult, err := s.migrateLocationPhotosToS3()
	if err != nil {
		log.Printf("Error migrating location photos: %v", err)
	}
	result.LocationPhotos = locationResult

	// Migrate feed photos
	feedResult, err := s.migrateFeedPhotosToS3()
	if err != nil {
		log.Printf("Error migrating feed photos: %v", err)
	}
	result.FeedPhotos = feedResult

	// Migrate faskes photos
	faskesResult, err := s.migrateFaskesPhotosToS3()
	if err != nil {
		log.Printf("Error migrating faskes photos: %v", err)
	}
	result.FaskesPhotos = faskesResult

	// Calculate totals
	if result.LocationPhotos != nil {
		result.TotalMigrated += result.LocationPhotos.Downloaded
		result.TotalErrors += result.LocationPhotos.Errors
	}
	if result.FeedPhotos != nil {
		result.TotalMigrated += result.FeedPhotos.Downloaded
		result.TotalErrors += result.FeedPhotos.Errors
	}
	if result.FaskesPhotos != nil {
		result.TotalMigrated += result.FaskesPhotos.Downloaded
		result.TotalErrors += result.FaskesPhotos.Errors
	}

	result.Duration = time.Since(startTime).String()

	return result, nil
}

// migrateLocationPhotosToS3 migrates location photos from local storage to S3
func (s *PhotoService) migrateLocationPhotosToS3() (*PhotoSyncResult, error) {
	result := &PhotoSyncResult{
		StartTime: time.Now(),
	}

	// Find all cached photos that are NOT yet on S3 (storage_path doesn't start with http)
	var photos []model.LocationPhoto
	err := s.db.Where("is_cached = true AND storage_path IS NOT NULL AND storage_path NOT LIKE 'http%'").
		Find(&photos).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch local photos: %w", err)
	}

	result.TotalFound = len(photos)
	log.Printf("Found %d location photos to migrate to S3", len(photos))

	for _, photo := range photos {
		if photo.StoragePath == nil {
			continue
		}

		localPath := *photo.StoragePath

		// Read local file
		data, err := os.ReadFile(localPath)
		if err != nil {
			result.Errors++
			result.ErrorDetails = append(result.ErrorDetails, fmt.Sprintf("%s: failed to read local file: %v", photo.Filename, err))
			continue
		}

		// Generate S3 key
		ext := filepath.Ext(localPath)
		key := fmt.Sprintf("locations/%s/%s", photo.LocationID.String(), filepath.Base(localPath))
		contentType := getContentType(ext)

		// Upload to S3
		url, err := s.s3Storage.Upload(context.Background(), key, data, contentType)
		if err != nil {
			result.Errors++
			result.ErrorDetails = append(result.ErrorDetails, fmt.Sprintf("%s: failed to upload to S3: %v", photo.Filename, err))
			continue
		}

		// Update database with S3 URL
		photo.StoragePath = &url
		if err := s.db.Save(&photo).Error; err != nil {
			result.Errors++
			result.ErrorDetails = append(result.ErrorDetails, fmt.Sprintf("%s: failed to update database: %v", photo.Filename, err))
			// Try to delete from S3 since we couldn't update the DB
			s.s3Storage.Delete(context.Background(), key)
			continue
		}

		log.Printf("Migrated location photo to S3: %s -> %s", localPath, url)
		result.Downloaded++
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime).String()

	return result, nil
}

// migrateFeedPhotosToS3 migrates feed photos from local storage to S3
func (s *PhotoService) migrateFeedPhotosToS3() (*PhotoSyncResult, error) {
	result := &PhotoSyncResult{
		StartTime: time.Now(),
	}

	var photos []model.FeedPhoto
	err := s.db.Where("is_cached = true AND storage_path IS NOT NULL AND storage_path NOT LIKE 'http%'").
		Find(&photos).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch local feed photos: %w", err)
	}

	result.TotalFound = len(photos)
	log.Printf("Found %d feed photos to migrate to S3", len(photos))

	for _, photo := range photos {
		if photo.StoragePath == nil {
			continue
		}

		localPath := *photo.StoragePath

		data, err := os.ReadFile(localPath)
		if err != nil {
			result.Errors++
			result.ErrorDetails = append(result.ErrorDetails, fmt.Sprintf("%s: failed to read local file: %v", photo.Filename, err))
			continue
		}

		ext := filepath.Ext(localPath)
		key := fmt.Sprintf("feeds/%s/%s", photo.FeedID.String(), filepath.Base(localPath))
		contentType := getContentType(ext)

		url, err := s.s3Storage.Upload(context.Background(), key, data, contentType)
		if err != nil {
			result.Errors++
			result.ErrorDetails = append(result.ErrorDetails, fmt.Sprintf("%s: failed to upload to S3: %v", photo.Filename, err))
			continue
		}

		photo.StoragePath = &url
		if err := s.db.Save(&photo).Error; err != nil {
			result.Errors++
			result.ErrorDetails = append(result.ErrorDetails, fmt.Sprintf("%s: failed to update database: %v", photo.Filename, err))
			s.s3Storage.Delete(context.Background(), key)
			continue
		}

		log.Printf("Migrated feed photo to S3: %s -> %s", localPath, url)
		result.Downloaded++
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime).String()

	return result, nil
}

// migrateFaskesPhotosToS3 migrates faskes photos from local storage to S3
func (s *PhotoService) migrateFaskesPhotosToS3() (*PhotoSyncResult, error) {
	result := &PhotoSyncResult{
		StartTime: time.Now(),
	}

	var photos []model.FaskesPhoto
	err := s.db.Where("is_cached = true AND storage_path IS NOT NULL AND storage_path NOT LIKE 'http%'").
		Find(&photos).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch local faskes photos: %w", err)
	}

	result.TotalFound = len(photos)
	log.Printf("Found %d faskes photos to migrate to S3", len(photos))

	for _, photo := range photos {
		if photo.StoragePath == nil {
			continue
		}

		localPath := *photo.StoragePath

		data, err := os.ReadFile(localPath)
		if err != nil {
			result.Errors++
			result.ErrorDetails = append(result.ErrorDetails, fmt.Sprintf("%s: failed to read local file: %v", photo.Filename, err))
			continue
		}

		ext := filepath.Ext(localPath)
		key := fmt.Sprintf("faskes/%s/%s", photo.FaskesID.String(), filepath.Base(localPath))
		contentType := getContentType(ext)

		url, err := s.s3Storage.Upload(context.Background(), key, data, contentType)
		if err != nil {
			result.Errors++
			result.ErrorDetails = append(result.ErrorDetails, fmt.Sprintf("%s: failed to upload to S3: %v", photo.Filename, err))
			continue
		}

		photo.StoragePath = &url
		if err := s.db.Save(&photo).Error; err != nil {
			result.Errors++
			result.ErrorDetails = append(result.ErrorDetails, fmt.Sprintf("%s: failed to update database: %v", photo.Filename, err))
			s.s3Storage.Delete(context.Background(), key)
			continue
		}

		log.Printf("Migrated faskes photo to S3: %s -> %s", localPath, url)
		result.Downloaded++
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime).String()

	return result, nil
}
