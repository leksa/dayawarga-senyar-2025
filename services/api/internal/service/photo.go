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

	return &PhotoService{
		db:          db,
		odkClient:   odkClient,
		storagePath: storagePath,
		useS3:       false,
	}
}

// NewPhotoServiceWithS3 creates a new photo service with S3 storage
func NewPhotoServiceWithS3(db *gorm.DB, odkClient *odk.Client, storagePath string, s3Storage *storage.S3Storage) *PhotoService {
	return &PhotoService{
		db:          db,
		odkClient:   odkClient,
		storagePath: storagePath,
		s3Storage:   s3Storage,
		useS3:       s3Storage != nil,
	}
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
func extractS3Key(url string) string {
	// URL format: https://bucket.endpoint/key or https://endpoint/bucket/key
	// We need to extract the path after the bucket
	parts := strings.SplitN(url, "/", 4)
	if len(parts) >= 4 {
		return parts[3]
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
