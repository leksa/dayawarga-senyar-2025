package service

import (
	"fmt"
	"log"
	"time"

	"github.com/leksa/datamapper-senyar/internal/model"
	"github.com/leksa/datamapper-senyar/internal/odk"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FeedSyncService handles synchronization of feeds from ODK Central to PostgreSQL
type FeedSyncService struct {
	db        *gorm.DB
	odkClient *odk.Client
	formID    string
}

// NewFeedSyncService creates a new feed sync service
func NewFeedSyncService(db *gorm.DB, odkClient *odk.Client, formID string) *FeedSyncService {
	return &FeedSyncService{
		db:        db,
		odkClient: odkClient,
		formID:    formID,
	}
}

// FeedSyncResult holds the result of a feed sync operation
type FeedSyncResult struct {
	TotalFetched int       `json:"total_fetched"`
	Created      int       `json:"created"`
	Updated      int       `json:"updated"`
	Deleted      int       `json:"deleted,omitempty"`
	Skipped      int       `json:"skipped"`
	Errors       int       `json:"errors"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	Duration     string    `json:"duration"`
	ErrorDetails []string  `json:"error_details,omitempty"`
}

// SyncAll performs a full synchronization of all approved feed submissions
func (s *FeedSyncService) SyncAll() (*FeedSyncResult, error) {
	result := &FeedSyncResult{
		StartTime: time.Now(),
	}

	// Update sync state to "syncing"
	s.updateSyncState("syncing", nil)

	// Fetch all approved submissions
	submissions, err := s.odkClient.GetApprovedSubmissions()
	if err != nil {
		errMsg := fmt.Sprintf("failed to fetch feed submissions: %v", err)
		s.updateSyncState("error", &errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	result.TotalFetched = len(submissions)
	log.Printf("Fetched %d feed submissions from ODK Central", result.TotalFetched)

	// Process each submission
	for _, submission := range submissions {
		if err := s.processSubmission(submission, result); err != nil {
			result.Errors++
			result.ErrorDetails = append(result.ErrorDetails, err.Error())
			log.Printf("Error processing feed submission: %v", err)
		}
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime).String()

	// Update sync state
	s.updateSyncStateSuccess(result.TotalFetched)

	log.Printf("Feed sync completed: %d fetched, %d created, %d updated, %d skipped, %d errors",
		result.TotalFetched, result.Created, result.Updated, result.Skipped, result.Errors)

	return result, nil
}

// processSubmission processes a single feed submission
func (s *FeedSyncService) processSubmission(submission map[string]interface{}, result *FeedSyncResult) error {
	// Get submission ID
	odkID, ok := submission["__id"].(string)
	if !ok {
		return fmt.Errorf("submission missing __id")
	}

	// Check review state - only process approved submissions
	if system, ok := submission["__system"].(map[string]interface{}); ok {
		if reviewState, ok := system["reviewState"].(string); ok && reviewState != "approved" {
			log.Printf("Skipping non-approved feed submission %s (state: %s)", odkID, reviewState)
			result.Skipped++
			return nil
		}
	}

	// Map submission to feed with photos
	feedResult, err := MapFeedSubmissionWithPhotos(submission)
	if err != nil {
		return fmt.Errorf("failed to map feed submission %s: %w", odkID, err)
	}
	feed := feedResult.Feed

	// Resolve location_id: the calc_location_id from ODK is the entity name, not our DB UUID
	// We need to lookup the location by matching the nama_posko
	if feed.LocationID != nil {
		// Try to find the location by looking up calc_nama_posko in raw_data
		if namaPosko, ok := submission["calc_nama_posko"].(string); ok && namaPosko != "" {
			var location model.Location
			if err := s.db.Where("nama = ?", namaPosko).First(&location).Error; err == nil {
				feed.LocationID = &location.ID
				log.Printf("Resolved location_id for '%s' -> %s", namaPosko, location.ID)
			} else {
				log.Printf("Warning: Could not find location for posko '%s', setting location_id to NULL", namaPosko)
				feed.LocationID = nil
			}
		} else {
			log.Printf("Warning: No calc_nama_posko in submission, setting location_id to NULL")
			feed.LocationID = nil
		}
	}

	// Resolve faskes_id: lookup by nama_faskes
	if feed.FaskesID != nil {
		if namaFaskes, ok := submission["calc_nama_faskes"].(string); ok && namaFaskes != "" {
			var faskes model.Faskes
			if err := s.db.Where("nama = ?", namaFaskes).First(&faskes).Error; err == nil {
				feed.FaskesID = &faskes.ID
				log.Printf("Resolved faskes_id for '%s' -> %s", namaFaskes, faskes.ID)
			} else {
				log.Printf("Warning: Could not find faskes '%s', setting faskes_id to NULL", namaFaskes)
				feed.FaskesID = nil
			}
		} else {
			log.Printf("Warning: No calc_nama_faskes in submission, setting faskes_id to NULL")
			feed.FaskesID = nil
		}
	}

	// Check if feed already exists
	var existingFeed model.Feed
	err = s.db.Where("odk_submission_id = ?", odkID).First(&existingFeed).Error

	if err == gorm.ErrRecordNotFound {
		// Create new feed
		if err := s.createFeed(feed); err != nil {
			return fmt.Errorf("failed to create feed for %s: %w", odkID, err)
		}

		// Save photos
		if len(feedResult.Photos) > 0 {
			if err := s.saveFeedPhotos(feed.ID, feedResult.Photos); err != nil {
				log.Printf("Warning: Failed to save photos for feed %s: %v", odkID, err)
			}
		}

		result.Created++
		log.Printf("Created feed: %s (%s) with %d photos", odkID, feed.Category, len(feedResult.Photos))
	} else if err == nil {
		// Update existing feed
		feed.ID = existingFeed.ID
		if err := s.updateFeed(feed); err != nil {
			return fmt.Errorf("failed to update feed for %s: %w", odkID, err)
		}

		// Update photos (delete existing and re-create)
		if len(feedResult.Photos) > 0 {
			s.db.Where("feed_id = ?", feed.ID).Delete(&model.FeedPhoto{})
			if err := s.saveFeedPhotos(feed.ID, feedResult.Photos); err != nil {
				log.Printf("Warning: Failed to update photos for feed %s: %v", odkID, err)
			}
		}

		result.Updated++
		log.Printf("Updated feed: %s (%s) with %d photos", odkID, feed.Category, len(feedResult.Photos))
	} else {
		return fmt.Errorf("database error checking feed %s: %w", odkID, err)
	}

	return nil
}

// saveFeedPhotos saves photo records for a feed
func (s *FeedSyncService) saveFeedPhotos(feedID uuid.UUID, photos []FeedPhotoInfo) error {
	for _, photo := range photos {
		feedPhoto := model.FeedPhoto{
			ID:        uuid.New(),
			FeedID:    feedID,
			PhotoType: photo.PhotoType,
			Filename:  photo.Filename,
			IsCached:  false,
		}
		if err := s.db.Create(&feedPhoto).Error; err != nil {
			return fmt.Errorf("failed to save photo %s: %w", photo.Filename, err)
		}
	}
	return nil
}

// createFeed creates a new feed with PostGIS geometry
func (s *FeedSyncService) createFeed(feed *model.Feed) error {
	feed.ID = uuid.New()
	now := time.Now()
	feed.CreatedAt = now
	feed.UpdatedAt = now

	// Check if we have valid coordinates
	hasCoords := feed.Longitude != nil && feed.Latitude != nil && *feed.Longitude != 0 && *feed.Latitude != 0

	var sql string
	var args []interface{}

	if hasCoords {
		sql = `
			INSERT INTO information_feeds (
				id, location_id, faskes_id, odk_submission_id,
				content, category, type, username, organization,
				geom, raw_data, submitted_at, created_at, updated_at
			) VALUES (
				?, ?, ?, ?,
				?, ?, ?, ?, ?,
				ST_SetSRID(ST_MakePoint(?, ?), 4326), ?, ?, ?, ?
			)
		`
		args = []interface{}{
			feed.ID, feed.LocationID, feed.FaskesID, feed.ODKSubmissionID,
			feed.Content, feed.Category, feed.Type, feed.Username, feed.Organization,
			*feed.Longitude, *feed.Latitude, feed.RawData, feed.SubmittedAt, feed.CreatedAt, feed.UpdatedAt,
		}
	} else {
		sql = `
			INSERT INTO information_feeds (
				id, location_id, faskes_id, odk_submission_id,
				content, category, type, username, organization,
				geom, raw_data, submitted_at, created_at, updated_at
			) VALUES (
				?, ?, ?, ?,
				?, ?, ?, ?, ?,
				NULL, ?, ?, ?, ?
			)
		`
		args = []interface{}{
			feed.ID, feed.LocationID, feed.FaskesID, feed.ODKSubmissionID,
			feed.Content, feed.Category, feed.Type, feed.Username, feed.Organization,
			feed.RawData, feed.SubmittedAt, feed.CreatedAt, feed.UpdatedAt,
		}
	}

	return s.db.Exec(sql, args...).Error
}

// updateFeed updates an existing feed
func (s *FeedSyncService) updateFeed(feed *model.Feed) error {
	now := time.Now()
	feed.UpdatedAt = now

	// Check if we have valid coordinates
	hasCoords := feed.Longitude != nil && feed.Latitude != nil && *feed.Longitude != 0 && *feed.Latitude != 0

	var sql string
	var args []interface{}

	if hasCoords {
		sql = `
			UPDATE information_feeds SET
				location_id = ?,
				faskes_id = ?,
				content = ?,
				category = ?,
				type = ?,
				username = ?,
				geom = ST_SetSRID(ST_MakePoint(?, ?), 4326),
				raw_data = ?,
				submitted_at = ?,
				updated_at = ?
			WHERE id = ?
		`
		args = []interface{}{
			feed.LocationID,
			feed.FaskesID,
			feed.Content,
			feed.Category,
			feed.Type,
			feed.Username,
			*feed.Longitude, *feed.Latitude,
			feed.RawData,
			feed.SubmittedAt,
			feed.UpdatedAt,
			feed.ID,
		}
	} else {
		sql = `
			UPDATE information_feeds SET
				location_id = ?,
				faskes_id = ?,
				content = ?,
				category = ?,
				type = ?,
				username = ?,
				geom = NULL,
				raw_data = ?,
				submitted_at = ?,
				updated_at = ?
			WHERE id = ?
		`
		args = []interface{}{
			feed.LocationID,
			feed.FaskesID,
			feed.Content,
			feed.Category,
			feed.Type,
			feed.Username,
			feed.RawData,
			feed.SubmittedAt,
			feed.UpdatedAt,
			feed.ID,
		}
	}

	return s.db.Exec(sql, args...).Error
}

// updateSyncState updates the sync_state table for feed form
func (s *FeedSyncService) updateSyncState(status string, errorMsg *string) {
	var syncState odk.SyncState
	result := s.db.Where("form_id = ?", s.formID).First(&syncState)

	now := time.Now()

	if result.Error == gorm.ErrRecordNotFound {
		syncState = odk.SyncState{
			FormID:       s.formID,
			Status:       status,
			ErrorMessage: errorMsg,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		s.db.Create(&syncState)
	} else {
		syncState.Status = status
		syncState.ErrorMessage = errorMsg
		syncState.UpdatedAt = now
		s.db.Save(&syncState)
	}
}

// updateSyncStateSuccess updates sync state after successful sync
func (s *FeedSyncService) updateSyncStateSuccess(recordCount int) {
	var syncState odk.SyncState
	result := s.db.Where("form_id = ?", s.formID).First(&syncState)

	now := time.Now()

	if result.Error == gorm.ErrRecordNotFound {
		syncState = odk.SyncState{
			FormID:          s.formID,
			Status:          "idle",
			LastSyncTime:    &now,
			LastRecordCount: recordCount,
			TotalRecords:    recordCount,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		s.db.Create(&syncState)
	} else {
		syncState.Status = "idle"
		syncState.LastSyncTime = &now
		syncState.LastRecordCount = recordCount
		syncState.TotalRecords += recordCount
		syncState.ErrorMessage = nil
		syncState.UpdatedAt = now
		s.db.Save(&syncState)
	}
}

// GetSyncState returns the current sync state for the feed form
func (s *FeedSyncService) GetSyncState() (*odk.SyncState, error) {
	var syncState odk.SyncState
	err := s.db.Where("form_id = ?", s.formID).First(&syncState).Error
	if err == gorm.ErrRecordNotFound {
		return &odk.SyncState{
			FormID: s.formID,
			Status: "never_synced",
		}, nil
	}
	if err != nil {
		return nil, err
	}
	return &syncState, nil
}

// HardSync performs a full sync and deletes feeds that no longer exist in ODK Central
func (s *FeedSyncService) HardSync() (*FeedSyncResult, error) {
	result := &FeedSyncResult{
		StartTime: time.Now(),
	}

	s.updateSyncState("hard_syncing", nil)

	// Fetch all approved submissions from ODK Central
	submissions, err := s.odkClient.GetApprovedSubmissions()
	if err != nil {
		errMsg := fmt.Sprintf("failed to fetch feed submissions: %v", err)
		s.updateSyncState("error", &errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	result.TotalFetched = len(submissions)
	log.Printf("Feed HardSync: Fetched %d submissions from ODK Central", result.TotalFetched)

	// Build a set of ODK submission IDs from ODK Central
	odkIDSet := make(map[string]bool)
	for _, submission := range submissions {
		if odkID, ok := submission["__id"].(string); ok {
			odkIDSet[odkID] = true
		}
	}

	// Process each submission (create/update)
	for _, submission := range submissions {
		if err := s.processSubmission(submission, result); err != nil {
			result.Errors++
			result.ErrorDetails = append(result.ErrorDetails, err.Error())
			log.Printf("Error processing feed submission: %v", err)
		}
	}

	// Find and delete feeds that no longer exist in ODK Central
	var feeds []model.Feed
	if err := s.db.Where("odk_submission_id IS NOT NULL").Find(&feeds).Error; err != nil {
		result.Errors++
		result.ErrorDetails = append(result.ErrorDetails, fmt.Sprintf("failed to fetch existing feeds: %v", err))
	} else {
		for _, feed := range feeds {
			if feed.ODKSubmissionID != nil && !odkIDSet[*feed.ODKSubmissionID] {
				// This feed no longer exists in ODK Central - delete it
				log.Printf("Feed HardSync: Deleting feed %s (%s) - no longer in ODK Central", feed.ID, *feed.ODKSubmissionID)

				// Delete associated photos first
				if err := s.db.Where("feed_id = ?", feed.ID).Delete(&model.FeedPhoto{}).Error; err != nil {
					log.Printf("Warning: failed to delete photos for feed %s: %v", feed.ID, err)
				}

				// Delete the feed
				if err := s.db.Delete(&feed).Error; err != nil {
					result.Errors++
					result.ErrorDetails = append(result.ErrorDetails, fmt.Sprintf("failed to delete feed %s: %v", feed.ID, err))
				} else {
					result.Deleted++
				}
			}
		}
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime).String()

	s.updateSyncStateSuccess(result.TotalFetched)

	log.Printf("Feed HardSync completed: %d fetched, %d created, %d updated, %d deleted, %d errors",
		result.TotalFetched, result.Created, result.Updated, result.Deleted, result.Errors)

	return result, nil
}
