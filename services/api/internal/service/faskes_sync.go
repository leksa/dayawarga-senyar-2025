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

// FaskesSyncService handles synchronization of faskes data from ODK Central
type FaskesSyncService struct {
	db        *gorm.DB
	odkClient *odk.Client
	formID    string
}

// NewFaskesSyncService creates a new faskes sync service
func NewFaskesSyncService(db *gorm.DB, odkClient *odk.Client, formID string) *FaskesSyncService {
	return &FaskesSyncService{
		db:        db,
		odkClient: odkClient,
		formID:    formID,
	}
}

// SyncAll performs a full synchronization of all approved faskes submissions
func (s *FaskesSyncService) SyncAll() (*SyncResult, error) {
	result := &SyncResult{
		StartTime: time.Now(),
	}

	// Update sync state to "syncing"
	s.updateSyncState("syncing", nil)

	// Fetch all approved submissions
	submissions, err := s.odkClient.GetApprovedSubmissions()
	if err != nil {
		errMsg := fmt.Sprintf("failed to fetch faskes submissions: %v", err)
		s.updateSyncState("error", &errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	result.TotalFetched = len(submissions)
	log.Printf("Fetched %d faskes submissions from ODK Central", result.TotalFetched)

	// Process each submission
	for _, submission := range submissions {
		if err := s.processSubmission(submission, result); err != nil {
			result.Errors++
			result.ErrorDetails = append(result.ErrorDetails, err.Error())
			log.Printf("Error processing faskes submission: %v", err)
		}
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime).String()

	// Update sync state
	s.updateSyncStateSuccess(result.TotalFetched)

	log.Printf("Faskes sync completed: %d fetched, %d created, %d updated, %d errors",
		result.TotalFetched, result.Created, result.Updated, result.Errors)

	return result, nil
}

// processSubmission processes a single faskes submission
func (s *FaskesSyncService) processSubmission(submission map[string]interface{}, result *SyncResult) error {
	// Get submission ID
	odkID, ok := submission["__id"].(string)
	if !ok {
		return fmt.Errorf("submission missing __id")
	}

	// Check review state - only process approved submissions
	if system, ok := submission["__system"].(map[string]interface{}); ok {
		if reviewState, ok := system["reviewState"].(string); ok && reviewState != "approved" {
			log.Printf("Skipping non-approved faskes submission %s (state: %s)", odkID, reviewState)
			return nil
		}
	}

	// Map submission to faskes
	faskes, err := MapSubmissionToFaskes(submission)
	if err != nil {
		return fmt.Errorf("failed to map faskes submission %s: %w", odkID, err)
	}

	// Check if faskes already exists
	var existingFaskes model.Faskes
	err = s.db.Where("odk_submission_id = ?", odkID).First(&existingFaskes).Error

	if err == gorm.ErrRecordNotFound {
		// Create new faskes
		if err := s.createFaskes(faskes); err != nil {
			return fmt.Errorf("failed to create faskes for %s: %w", odkID, err)
		}
		result.Created++
		log.Printf("Created faskes: %s (%s)", faskes.Nama, odkID)
	} else if err == nil {
		// Update existing faskes
		faskes.ID = existingFaskes.ID
		if err := s.updateFaskes(faskes); err != nil {
			return fmt.Errorf("failed to update faskes for %s: %w", odkID, err)
		}
		result.Updated++
		log.Printf("Updated faskes: %s (%s)", faskes.Nama, odkID)
	} else {
		return fmt.Errorf("database error checking faskes %s: %w", odkID, err)
	}

	// Process photos
	photos := ExtractFaskesPhotos(submission)
	for _, photo := range photos {
		if err := s.processPhoto(faskes.ID, photo); err != nil {
			log.Printf("Warning: failed to process faskes photo %s: %v", photo.Filename, err)
		}
	}

	return nil
}

// createFaskes creates a new faskes with PostGIS geometry
func (s *FaskesSyncService) createFaskes(faskes *model.Faskes) error {
	faskes.ID = uuid.New()
	now := time.Now()
	faskes.CreatedAt = now
	faskes.UpdatedAt = now
	faskes.SyncedAt = &now

	// Build SQL with geometry
	sql := `
		INSERT INTO faskes (
			id, odk_submission_id, nama, jenis_faskes, status_faskes, kondisi_faskes,
			geom, alamat, identitas, isolasi, infrastruktur, sdm, perbekalan, klaster, raw_data,
			submitter_name, submitted_at, created_at, updated_at, synced_at
		) VALUES (
			?, ?, ?, ?, ?, ?,
			ST_SetSRID(ST_MakePoint(?, ?), 4326), ?, ?, ?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?
		)
	`

	lon := float64(0)
	lat := float64(0)
	if faskes.Longitude != nil {
		lon = *faskes.Longitude
	}
	if faskes.Latitude != nil {
		lat = *faskes.Latitude
	}

	return s.db.Exec(sql,
		faskes.ID, faskes.ODKSubmissionID, faskes.Nama, faskes.JenisFaskes, faskes.StatusFaskes, faskes.KondisiFaskes,
		lon, lat, faskes.Alamat, faskes.Identitas, faskes.Isolasi, faskes.Infrastruktur, faskes.SDM, faskes.Perbekalan, faskes.Klaster, faskes.RawData,
		faskes.SubmitterName, faskes.SubmittedAt, faskes.CreatedAt, faskes.UpdatedAt, faskes.SyncedAt,
	).Error
}

// updateFaskes updates an existing faskes
func (s *FaskesSyncService) updateFaskes(faskes *model.Faskes) error {
	now := time.Now()
	faskes.UpdatedAt = now
	faskes.SyncedAt = &now

	sql := `
		UPDATE faskes SET
			nama = ?,
			jenis_faskes = ?,
			status_faskes = ?,
			kondisi_faskes = ?,
			geom = ST_SetSRID(ST_MakePoint(?, ?), 4326),
			alamat = ?,
			identitas = ?,
			isolasi = ?,
			infrastruktur = ?,
			sdm = ?,
			perbekalan = ?,
			klaster = ?,
			raw_data = ?,
			submitter_name = ?,
			submitted_at = ?,
			updated_at = ?,
			synced_at = ?
		WHERE id = ?
	`

	lon := float64(0)
	lat := float64(0)
	if faskes.Longitude != nil {
		lon = *faskes.Longitude
	}
	if faskes.Latitude != nil {
		lat = *faskes.Latitude
	}

	return s.db.Exec(sql,
		faskes.Nama,
		faskes.JenisFaskes,
		faskes.StatusFaskes,
		faskes.KondisiFaskes,
		lon, lat,
		faskes.Alamat,
		faskes.Identitas,
		faskes.Isolasi,
		faskes.Infrastruktur,
		faskes.SDM,
		faskes.Perbekalan,
		faskes.Klaster,
		faskes.RawData,
		faskes.SubmitterName,
		faskes.SubmittedAt,
		faskes.UpdatedAt,
		faskes.SyncedAt,
		faskes.ID,
	).Error
}

// processPhoto saves faskes photo metadata
func (s *FaskesSyncService) processPhoto(faskesID uuid.UUID, photo PhotoInfo) error {
	// Check if photo already exists
	var count int64
	s.db.Model(&model.FaskesPhoto{}).
		Where("faskes_id = ? AND filename = ?", faskesID, photo.Filename).
		Count(&count)

	if count > 0 {
		return nil // Photo already exists
	}

	faskesPhoto := &model.FaskesPhoto{
		ID:        uuid.New(),
		FaskesID:  faskesID,
		PhotoType: photo.PhotoType,
		Filename:  photo.Filename,
		IsCached:  false,
		CreatedAt: time.Now(),
	}

	return s.db.Create(faskesPhoto).Error
}

// updateSyncState updates the sync_state table
func (s *FaskesSyncService) updateSyncState(status string, errorMsg *string) {
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
func (s *FaskesSyncService) updateSyncStateSuccess(recordCount int) {
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

// GetSyncState returns the current sync state for faskes form
func (s *FaskesSyncService) GetSyncState() (*odk.SyncState, error) {
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
