package service

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/leksa/datamapper-senyar/internal/model"
	"github.com/leksa/datamapper-senyar/internal/odk"
	"gorm.io/gorm"
)

// PhotoService handles photo storage and retrieval
type PhotoService struct {
	db          *gorm.DB
	odkClient   *odk.Client
	storagePath string
}

// NewPhotoService creates a new photo service
func NewPhotoService(db *gorm.DB, odkClient *odk.Client, storagePath string) *PhotoService {
	// Create storage directory if it doesn't exist
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		log.Printf("Warning: failed to create storage directory: %v", err)
	}

	return &PhotoService{
		db:          db,
		odkClient:   odkClient,
		storagePath: storagePath,
	}
}

// DownloadAndSavePhoto downloads a photo from ODK Central and saves it to local storage
func (s *PhotoService) DownloadAndSavePhoto(photo *model.LocationPhoto, submissionID string) error {
	// Download from ODK Central
	data, err := s.odkClient.GetAttachment(submissionID, photo.Filename)
	if err != nil {
		return fmt.Errorf("failed to download attachment: %w", err)
	}

	// Create location-specific directory
	locationDir := filepath.Join(s.storagePath, photo.LocationID.String())
	if err := os.MkdirAll(locationDir, 0755); err != nil {
		return fmt.Errorf("failed to create location directory: %w", err)
	}

	// Generate unique filename to avoid conflicts
	ext := filepath.Ext(photo.Filename)
	newFilename := fmt.Sprintf("%s_%s%s", photo.PhotoType, uuid.New().String()[:8], ext)
	storagePath := filepath.Join(locationDir, newFilename)

	// Write file
	if err := os.WriteFile(storagePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Update database record
	fileSize := len(data)
	photo.StoragePath = &storagePath
	photo.IsCached = true
	photo.FileSize = &fileSize

	if err := s.db.Save(photo).Error; err != nil {
		// Clean up file if database update fails
		os.Remove(storagePath)
		return fmt.Errorf("failed to update database: %w", err)
	}

	log.Printf("Downloaded photo: %s -> %s", photo.Filename, storagePath)
	return nil
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
	path, err := s.GetPhotoPath(photoID)
	if err != nil {
		return nil, "", err
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, "", fmt.Errorf("failed to open file: %w", err)
	}

	return file, filepath.Base(path), nil
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

// DownloadAndSaveFeedPhoto downloads a feed photo from ODK Central and saves it to local storage
func (s *PhotoService) DownloadAndSaveFeedPhoto(photo *model.FeedPhoto, submissionID string, formID string) error {
	// Download from ODK Central using the feed form
	data, err := s.odkClient.GetAttachmentForForm(formID, submissionID, photo.Filename)
	if err != nil {
		return fmt.Errorf("failed to download feed attachment: %w", err)
	}

	// Create feed-specific directory
	feedDir := filepath.Join(s.storagePath, "feeds", photo.FeedID.String())
	if err := os.MkdirAll(feedDir, 0755); err != nil {
		return fmt.Errorf("failed to create feed directory: %w", err)
	}

	// Generate unique filename to avoid conflicts
	ext := filepath.Ext(photo.Filename)
	newFilename := fmt.Sprintf("%s_%s%s", photo.PhotoType, uuid.New().String()[:8], ext)
	storagePath := filepath.Join(feedDir, newFilename)

	// Write file
	if err := os.WriteFile(storagePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Update database record
	fileSize := len(data)
	photo.StoragePath = &storagePath
	photo.IsCached = true
	photo.FileSize = &fileSize

	if err := s.db.Save(photo).Error; err != nil {
		// Clean up file if database update fails
		os.Remove(storagePath)
		return fmt.Errorf("failed to update database: %w", err)
	}

	log.Printf("Downloaded feed photo: %s -> %s", photo.Filename, storagePath)
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
	path, err := s.GetFeedPhotoPath(photoID)
	if err != nil {
		return nil, "", err
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, "", fmt.Errorf("failed to open file: %w", err)
	}

	return file, filepath.Base(path), nil
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

// DownloadAndSaveFaskesPhoto downloads a faskes photo from ODK Central and saves it to local storage
func (s *PhotoService) DownloadAndSaveFaskesPhoto(photo *model.FaskesPhoto, submissionID string, formID string) error {
	// Download from ODK Central using the faskes form
	data, err := s.odkClient.GetAttachmentForForm(formID, submissionID, photo.Filename)
	if err != nil {
		return fmt.Errorf("failed to download faskes attachment: %w", err)
	}

	// Create faskes-specific directory
	faskesDir := filepath.Join(s.storagePath, "faskes", photo.FaskesID.String())
	if err := os.MkdirAll(faskesDir, 0755); err != nil {
		return fmt.Errorf("failed to create faskes directory: %w", err)
	}

	// Generate unique filename to avoid conflicts
	ext := filepath.Ext(photo.Filename)
	newFilename := fmt.Sprintf("%s_%s%s", photo.PhotoType, uuid.New().String()[:8], ext)
	storagePath := filepath.Join(faskesDir, newFilename)

	// Write file
	if err := os.WriteFile(storagePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Update database record
	fileSize := len(data)
	photo.StoragePath = &storagePath
	photo.IsCached = true
	photo.FileSize = &fileSize

	if err := s.db.Save(photo).Error; err != nil {
		// Clean up file if database update fails
		os.Remove(storagePath)
		return fmt.Errorf("failed to update database: %w", err)
	}

	log.Printf("Downloaded faskes photo: %s -> %s", photo.Filename, storagePath)
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
	path, err := s.GetFaskesPhotoPath(photoID)
	if err != nil {
		return nil, "", err
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, "", fmt.Errorf("failed to open file: %w", err)
	}

	return file, filepath.Base(path), nil
}

// GetFaskesPhotosByFaskesID returns all photos for a faskes
func (s *PhotoService) GetFaskesPhotosByFaskesID(faskesID uuid.UUID) ([]model.FaskesPhoto, error) {
	var photos []model.FaskesPhoto
	if err := s.db.Where("faskes_id = ?", faskesID).Find(&photos).Error; err != nil {
		return nil, err
	}
	return photos, nil
}
