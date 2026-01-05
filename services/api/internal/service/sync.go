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

// SyncService handles synchronization between ODK Central and PostgreSQL
type SyncService struct {
	db        *gorm.DB
	odkClient *odk.Client
	formID    string
}

// NewSyncService creates a new sync service
func NewSyncService(db *gorm.DB, odkClient *odk.Client, formID string) *SyncService {
	return &SyncService{
		db:        db,
		odkClient: odkClient,
		formID:    formID,
	}
}

// SyncResult holds the result of a sync operation
type SyncResult struct {
	TotalFetched int       `json:"total_fetched"`
	Created      int       `json:"created"`
	Updated      int       `json:"updated"`
	Errors       int       `json:"errors"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	Duration     string    `json:"duration"`
	ErrorDetails []string  `json:"error_details,omitempty"`
}

// SyncAll performs a full synchronization of all approved submissions
func (s *SyncService) SyncAll() (*SyncResult, error) {
	result := &SyncResult{
		StartTime: time.Now(),
	}

	// Update sync state to "syncing"
	s.updateSyncState("syncing", nil)

	// Fetch all approved submissions
	submissions, err := s.odkClient.GetApprovedSubmissions()
	if err != nil {
		errMsg := fmt.Sprintf("failed to fetch submissions: %v", err)
		s.updateSyncState("error", &errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	result.TotalFetched = len(submissions)
	log.Printf("Fetched %d submissions from ODK Central", result.TotalFetched)

	// Process each submission
	for _, submission := range submissions {
		if err := s.processSubmission(submission, result); err != nil {
			result.Errors++
			result.ErrorDetails = append(result.ErrorDetails, err.Error())
			log.Printf("Error processing submission: %v", err)
		}
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime).String()

	// Update sync state
	s.updateSyncStateSuccess(result.TotalFetched)

	log.Printf("Sync completed: %d fetched, %d created, %d updated, %d errors",
		result.TotalFetched, result.Created, result.Updated, result.Errors)

	return result, nil
}

// SyncSince performs incremental sync since last sync time
func (s *SyncService) SyncSince(since time.Time) (*SyncResult, error) {
	result := &SyncResult{
		StartTime: time.Now(),
	}

	s.updateSyncState("syncing", nil)

	submissions, err := s.odkClient.GetSubmissionsSince(since)
	if err != nil {
		errMsg := fmt.Sprintf("failed to fetch submissions: %v", err)
		s.updateSyncState("error", &errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	result.TotalFetched = len(submissions)

	for _, submission := range submissions {
		if err := s.processSubmission(submission, result); err != nil {
			result.Errors++
			result.ErrorDetails = append(result.ErrorDetails, err.Error())
		}
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime).String()

	s.updateSyncStateSuccess(result.TotalFetched)

	return result, nil
}

// processSubmission processes a single submission
func (s *SyncService) processSubmission(submission map[string]interface{}, result *SyncResult) error {
	// Get submission ID
	odkID, ok := submission["__id"].(string)
	if !ok {
		return fmt.Errorf("submission missing __id")
	}

	// Check review state - only process approved submissions
	if system, ok := submission["__system"].(map[string]interface{}); ok {
		if reviewState, ok := system["reviewState"].(string); ok && reviewState != "approved" {
			log.Printf("Skipping non-approved submission %s (state: %s)", odkID, reviewState)
			return nil
		}
	}

	// Map submission to location
	location, err := MapSubmissionToLocation(submission)
	if err != nil {
		return fmt.Errorf("failed to map submission %s: %w", odkID, err)
	}

	// Check if location already exists
	var existingLocation model.Location
	err = s.db.Where("odk_submission_id = ?", odkID).First(&existingLocation).Error

	if err == gorm.ErrRecordNotFound {
		// Create new location
		if err := s.createLocation(location); err != nil {
			return fmt.Errorf("failed to create location for %s: %w", odkID, err)
		}
		result.Created++
		log.Printf("Created location: %s (%s)", location.Nama, odkID)
	} else if err == nil {
		// Update existing location
		location.ID = existingLocation.ID
		if err := s.updateLocation(location); err != nil {
			return fmt.Errorf("failed to update location for %s: %w", odkID, err)
		}
		result.Updated++
		log.Printf("Updated location: %s (%s)", location.Nama, odkID)
	} else {
		return fmt.Errorf("database error checking location %s: %w", odkID, err)
	}

	// Process photos
	photos := ExtractPhotos(submission)
	for _, photo := range photos {
		if err := s.processPhoto(location.ID, photo); err != nil {
			log.Printf("Warning: failed to process photo %s: %v", photo.Filename, err)
		}
	}

	return nil
}

// createLocation creates a new location with PostGIS geometry
func (s *SyncService) createLocation(location *model.Location) error {
	location.ID = uuid.New()
	now := time.Now()
	location.CreatedAt = now
	location.UpdatedAt = now
	location.SyncedAt = &now

	// Build SQL with geometry
	sql := `
		INSERT INTO locations (
			id, odk_submission_id, nama, type, status,
			geom, geo_meta, identitas, alamat, data_pengungsi,
			fasilitas, komunikasi, akses, raw_data,
			submitter_name, submitted_at, created_at, updated_at, synced_at
		) VALUES (
			?, ?, ?, ?, ?,
			ST_SetSRID(ST_MakePoint(?, ?), 4326), ?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?, ?
		)
	`

	lon := float64(0)
	lat := float64(0)
	if location.Longitude != nil {
		lon = *location.Longitude
	}
	if location.Latitude != nil {
		lat = *location.Latitude
	}

	return s.db.Exec(sql,
		location.ID, location.ODKSubmissionID, location.Nama, location.Type, location.Status,
		lon, lat, location.GeoMeta, location.Identitas, location.Alamat, location.DataPengungsi,
		location.Fasilitas, location.Komunikasi, location.Akses, location.RawData,
		location.SubmitterName, location.SubmittedAt, location.CreatedAt, location.UpdatedAt, location.SyncedAt,
	).Error
}

// updateLocation updates an existing location
func (s *SyncService) updateLocation(location *model.Location) error {
	now := time.Now()
	location.UpdatedAt = now
	location.SyncedAt = &now

	sql := `
		UPDATE locations SET
			nama = ?,
			geom = ST_SetSRID(ST_MakePoint(?, ?), 4326),
			geo_meta = ?,
			identitas = ?,
			alamat = ?,
			data_pengungsi = ?,
			fasilitas = ?,
			komunikasi = ?,
			akses = ?,
			raw_data = ?,
			submitter_name = ?,
			submitted_at = ?,
			updated_at = ?,
			synced_at = ?
		WHERE id = ?
	`

	lon := float64(0)
	lat := float64(0)
	if location.Longitude != nil {
		lon = *location.Longitude
	}
	if location.Latitude != nil {
		lat = *location.Latitude
	}

	return s.db.Exec(sql,
		location.Nama,
		lon, lat,
		location.GeoMeta,
		location.Identitas,
		location.Alamat,
		location.DataPengungsi,
		location.Fasilitas,
		location.Komunikasi,
		location.Akses,
		location.RawData,
		location.SubmitterName,
		location.SubmittedAt,
		location.UpdatedAt,
		location.SyncedAt,
		location.ID,
	).Error
}

// processPhoto saves photo metadata (actual download can be done separately)
func (s *SyncService) processPhoto(locationID uuid.UUID, photo PhotoInfo) error {
	// Check if photo already exists
	var count int64
	s.db.Model(&model.LocationPhoto{}).
		Where("location_id = ? AND filename = ?", locationID, photo.Filename).
		Count(&count)

	if count > 0 {
		return nil // Photo already exists
	}

	locationPhoto := &model.LocationPhoto{
		ID:         uuid.New(),
		LocationID: locationID,
		PhotoType:  photo.PhotoType,
		Filename:   photo.Filename,
		IsCached:   false,
		CreatedAt:  time.Now(),
	}

	return s.db.Create(locationPhoto).Error
}

// updateSyncState updates the sync_state table
func (s *SyncService) updateSyncState(status string, errorMsg *string) {
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
func (s *SyncService) updateSyncStateSuccess(recordCount int) {
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

// GetSyncState returns the current sync state for a form
func (s *SyncService) GetSyncState() (*odk.SyncState, error) {
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
