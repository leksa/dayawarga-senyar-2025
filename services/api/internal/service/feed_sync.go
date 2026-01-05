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

	// Map submission to feed
	feed, err := MapFeedSubmission(submission)
	if err != nil {
		return fmt.Errorf("failed to map feed submission %s: %w", odkID, err)
	}

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

	// Check if feed already exists
	var existingFeed model.Feed
	err = s.db.Where("odk_submission_id = ?", odkID).First(&existingFeed).Error

	if err == gorm.ErrRecordNotFound {
		// Create new feed
		if err := s.createFeed(feed); err != nil {
			return fmt.Errorf("failed to create feed for %s: %w", odkID, err)
		}
		result.Created++
		log.Printf("Created feed: %s (%s)", odkID, feed.Category)
	} else if err == nil {
		// Update existing feed
		feed.ID = existingFeed.ID
		if err := s.updateFeed(feed); err != nil {
			return fmt.Errorf("failed to update feed for %s: %w", odkID, err)
		}
		result.Updated++
		log.Printf("Updated feed: %s (%s)", odkID, feed.Category)
	} else {
		return fmt.Errorf("database error checking feed %s: %w", odkID, err)
	}

	return nil
}

// createFeed creates a new feed with PostGIS geometry
func (s *FeedSyncService) createFeed(feed *model.Feed) error {
	feed.ID = uuid.New()
	now := time.Now()
	feed.CreatedAt = now
	feed.UpdatedAt = now

	// Build SQL with geometry
	sql := `
		INSERT INTO information_feeds (
			id, location_id, odk_submission_id,
			content, category, type, username, organization,
			geom, raw_data, submitted_at, created_at, updated_at
		) VALUES (
			?, ?, ?,
			?, ?, ?, ?, ?,
			ST_SetSRID(ST_MakePoint(?, ?), 4326), ?, ?, ?, ?
		)
	`

	lon := float64(0)
	lat := float64(0)
	if feed.Longitude != nil {
		lon = *feed.Longitude
	}
	if feed.Latitude != nil {
		lat = *feed.Latitude
	}

	return s.db.Exec(sql,
		feed.ID, feed.LocationID, feed.ODKSubmissionID,
		feed.Content, feed.Category, feed.Type, feed.Username, feed.Organization,
		lon, lat, feed.RawData, feed.SubmittedAt, feed.CreatedAt, feed.UpdatedAt,
	).Error
}

// updateFeed updates an existing feed
func (s *FeedSyncService) updateFeed(feed *model.Feed) error {
	now := time.Now()
	feed.UpdatedAt = now

	sql := `
		UPDATE information_feeds SET
			location_id = ?,
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

	lon := float64(0)
	lat := float64(0)
	if feed.Longitude != nil {
		lon = *feed.Longitude
	}
	if feed.Latitude != nil {
		lat = *feed.Latitude
	}

	return s.db.Exec(sql,
		feed.LocationID,
		feed.Content,
		feed.Category,
		feed.Type,
		feed.Username,
		lon, lat,
		feed.RawData,
		feed.SubmittedAt,
		feed.UpdatedAt,
		feed.ID,
	).Error
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
