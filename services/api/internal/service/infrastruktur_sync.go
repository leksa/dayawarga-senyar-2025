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

// InfrastrukturSyncService handles synchronization of infrastruktur data from ODK Central
type InfrastrukturSyncService struct {
	db            *gorm.DB
	odkClient     *odk.Client
	formID        string
	entityDataset string
}

// NewInfrastrukturSyncService creates a new infrastruktur sync service
func NewInfrastrukturSyncService(db *gorm.DB, odkClient *odk.Client, formID string) *InfrastrukturSyncService {
	return &InfrastrukturSyncService{
		db:            db,
		odkClient:     odkClient,
		formID:        formID,
		entityDataset: "jembatan_entities",
	}
}

// SyncAll performs a full synchronization of all approved infrastruktur submissions
func (s *InfrastrukturSyncService) SyncAll() (*SyncResult, error) {
	result := &SyncResult{
		StartTime: time.Now(),
	}

	// Update sync state to "syncing"
	s.updateSyncState("syncing", nil)

	// Fetch all approved submissions
	submissions, err := s.odkClient.GetApprovedSubmissions()
	if err != nil {
		errMsg := fmt.Sprintf("failed to fetch infrastruktur submissions: %v", err)
		s.updateSyncState("error", &errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	result.TotalFetched = len(submissions)
	log.Printf("Fetched %d infrastruktur submissions from ODK Central", result.TotalFetched)

	// Group submissions by entity_id and keep only the latest per entity
	latestByEntity := s.groupByEntityLatest(submissions)
	log.Printf("Grouped into %d unique entities", len(latestByEntity))

	// Process each entity's latest submission
	for entityID, submission := range latestByEntity {
		if err := s.processEntitySubmission(entityID, submission, result); err != nil {
			result.Errors++
			result.ErrorDetails = append(result.ErrorDetails, err.Error())
			log.Printf("Error processing infrastruktur entity %s: %v", entityID, err)
		}
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime).String()

	// Update sync state
	s.updateSyncStateSuccess(result.TotalFetched)

	log.Printf("Infrastruktur sync completed: %d fetched, %d entities, %d created, %d updated, %d errors",
		result.TotalFetched, len(latestByEntity), result.Created, result.Updated, result.Errors)

	return result, nil
}

// groupByEntityLatest groups submissions by entity_id (sel_jembatan) and returns only the latest per entity
func (s *InfrastrukturSyncService) groupByEntityLatest(submissions []map[string]interface{}) map[string]map[string]interface{} {
	latestByEntity := make(map[string]map[string]interface{})
	latestTimeByEntity := make(map[string]time.Time)

	for _, submission := range submissions {
		// Get submission timestamp
		var submittedAt time.Time
		if system, ok := submission["__system"].(map[string]interface{}); ok {
			if dateStr, ok := system["submissionDate"].(string); ok {
				if t, err := time.Parse(time.RFC3339, dateStr); err == nil {
					submittedAt = t
				}
			}
		}

		// Get entity ID from sel_jembatan (the entity being updated)
		entityID, _ := submission["sel_jembatan"].(string)
		if entityID == "" {
			continue
		}

		// Keep only the latest submission per entity
		if existingTime, exists := latestTimeByEntity[entityID]; !exists || submittedAt.After(existingTime) {
			latestByEntity[entityID] = submission
			latestTimeByEntity[entityID] = submittedAt
		}
	}

	return latestByEntity
}

// processEntitySubmission processes a submission for a specific entity
func (s *InfrastrukturSyncService) processEntitySubmission(entityID string, submission map[string]interface{}, result *SyncResult) error {
	// Get submission ID for logging
	odkID, _ := submission["__id"].(string)

	// Check review state - only process approved submissions
	if system, ok := submission["__system"].(map[string]interface{}); ok {
		if reviewState, ok := system["reviewState"].(string); ok && reviewState != "approved" {
			log.Printf("Skipping non-approved infrastruktur submission %s (state: %s)", odkID, reviewState)
			return nil
		}
	}

	// Map submission to infrastruktur
	infra, err := MapSubmissionToInfrastruktur(submission)
	if err != nil {
		return fmt.Errorf("failed to map infrastruktur submission %s: %w", odkID, err)
	}

	// Ensure entity_id is set
	infra.EntityID = entityID

	// Update odk_submission_id to the latest submission ID
	infra.ODKSubmissionID = &odkID

	// Check if infrastruktur already exists by entity_id
	var existingInfra model.Infrastruktur
	err = s.db.Where("entity_id = ?", entityID).First(&existingInfra).Error

	if err == gorm.ErrRecordNotFound {
		// Create new infrastruktur
		if err := s.createInfrastruktur(infra); err != nil {
			return fmt.Errorf("failed to create infrastruktur for entity %s: %w", entityID, err)
		}
		result.Created++
		log.Printf("Created infrastruktur: %s (entity: %s, submission: %s)", infra.Nama, entityID, odkID)
	} else if err == nil {
		// Update existing infrastruktur
		infra.ID = existingInfra.ID
		if err := s.updateInfrastruktur(infra); err != nil {
			return fmt.Errorf("failed to update infrastruktur for entity %s: %w", entityID, err)
		}
		result.Updated++
		log.Printf("Updated infrastruktur: %s (entity: %s, submission: %s)", infra.Nama, entityID, odkID)
	} else {
		return fmt.Errorf("database error checking infrastruktur entity %s: %w", entityID, err)
	}

	// Process photos
	photos := ExtractInfrastrukturPhotos(submission)
	for _, photo := range photos {
		if err := s.processPhoto(infra.ID, photo); err != nil {
			log.Printf("Warning: failed to process infrastruktur photo %s: %v", photo.Filename, err)
		}
	}

	return nil
}

// createInfrastruktur creates a new infrastruktur record with PostGIS geometry
func (s *InfrastrukturSyncService) createInfrastruktur(infra *model.Infrastruktur) error {
	infra.ID = uuid.New()
	now := time.Now()
	infra.CreatedAt = now
	infra.UpdatedAt = now
	infra.SyncedAt = &now

	// Build SQL with geometry
	sql := `
		INSERT INTO infrastruktur (
			id, odk_submission_id, entity_id, object_id, nama, jenis, status_jln,
			nama_provinsi, nama_kabupaten, geom,
			status_akses, keterangan_bencana, dampak,
			status_penanganan, penanganan_detail, bailey, progress, target_selesai,
			baseline_sumber, update_by, raw_data,
			submitter_name, submitted_at, created_at, updated_at, synced_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?,
			?, ?, ST_SetSRID(ST_MakePoint(?, ?), 4326),
			?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?,
			?, ?, ?, ?, ?
		)
	`

	lon := float64(0)
	lat := float64(0)
	if infra.Longitude != nil {
		lon = *infra.Longitude
	}
	if infra.Latitude != nil {
		lat = *infra.Latitude
	}

	return s.db.Exec(sql,
		infra.ID, infra.ODKSubmissionID, infra.EntityID, infra.ObjectID, infra.Nama, infra.Jenis, infra.StatusJln,
		infra.NamaProvinsi, infra.NamaKabupaten, lon, lat,
		infra.StatusAkses, infra.KeteranganBencana, infra.Dampak,
		infra.StatusPenanganan, infra.PenangananDetail, infra.Bailey, infra.Progress, infra.TargetSelesai,
		infra.BaselineSumber, infra.UpdateBy, infra.RawData,
		infra.SubmitterName, infra.SubmittedAt, infra.CreatedAt, infra.UpdatedAt, infra.SyncedAt,
	).Error
}

// updateInfrastruktur updates an existing infrastruktur record
func (s *InfrastrukturSyncService) updateInfrastruktur(infra *model.Infrastruktur) error {
	now := time.Now()
	infra.UpdatedAt = now
	infra.SyncedAt = &now

	sql := `
		UPDATE infrastruktur SET
			odk_submission_id = ?,
			nama = ?,
			geom = ST_SetSRID(ST_MakePoint(?, ?), 4326),
			status_akses = ?,
			keterangan_bencana = ?,
			dampak = ?,
			status_penanganan = ?,
			penanganan_detail = ?,
			bailey = ?,
			progress = ?,
			update_by = ?,
			raw_data = ?,
			submitter_name = ?,
			submitted_at = ?,
			updated_at = ?,
			synced_at = ?
		WHERE id = ?
	`

	lon := float64(0)
	lat := float64(0)
	if infra.Longitude != nil {
		lon = *infra.Longitude
	}
	if infra.Latitude != nil {
		lat = *infra.Latitude
	}

	return s.db.Exec(sql,
		infra.ODKSubmissionID,
		infra.Nama,
		lon, lat,
		infra.StatusAkses,
		infra.KeteranganBencana,
		infra.Dampak,
		infra.StatusPenanganan,
		infra.PenangananDetail,
		infra.Bailey,
		infra.Progress,
		infra.UpdateBy,
		infra.RawData,
		infra.SubmitterName,
		infra.SubmittedAt,
		infra.UpdatedAt,
		infra.SyncedAt,
		infra.ID,
	).Error
}

// processPhoto saves photo metadata
func (s *InfrastrukturSyncService) processPhoto(infrastrukturID uuid.UUID, photo InfrastrukturPhotoInfo) error {
	// Check if photo already exists
	var count int64
	s.db.Model(&model.InfrastrukturPhoto{}).
		Where("infrastruktur_id = ? AND filename = ?", infrastrukturID, photo.Filename).
		Count(&count)

	if count > 0 {
		return nil // Photo already exists
	}

	infraPhoto := &model.InfrastrukturPhoto{
		ID:              uuid.New(),
		InfrastrukturID: infrastrukturID,
		PhotoType:       photo.PhotoType,
		Filename:        photo.Filename,
		IsCached:        false,
		CreatedAt:       time.Now(),
	}

	return s.db.Create(infraPhoto).Error
}

// updateSyncState updates the sync_state table
func (s *InfrastrukturSyncService) updateSyncState(status string, errorMsg *string) {
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
func (s *InfrastrukturSyncService) updateSyncStateSuccess(recordCount int) {
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

// GetSyncState returns the current sync state
func (s *InfrastrukturSyncService) GetSyncState() (*odk.SyncState, error) {
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

// HardSync performs a full sync and deletes records that no longer exist in ODK Central
func (s *InfrastrukturSyncService) HardSync() (*SyncResult, error) {
	result := &SyncResult{
		StartTime: time.Now(),
	}

	s.updateSyncState("hard_syncing", nil)

	// Fetch all approved submissions from ODK Central
	submissions, err := s.odkClient.GetApprovedSubmissions()
	if err != nil {
		errMsg := fmt.Sprintf("failed to fetch infrastruktur submissions: %v", err)
		s.updateSyncState("error", &errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	result.TotalFetched = len(submissions)
	log.Printf("HardSync Infrastruktur: Fetched %d submissions", result.TotalFetched)

	// Group submissions by entity_id and keep only the latest per entity
	latestByEntity := s.groupByEntityLatest(submissions)
	log.Printf("HardSync Infrastruktur: Grouped into %d unique entities", len(latestByEntity))

	// Build a set of entity IDs from ODK Central
	entityIDSet := make(map[string]bool)
	for entityID := range latestByEntity {
		entityIDSet[entityID] = true
	}

	// Process each entity's latest submission (create/update)
	for entityID, submission := range latestByEntity {
		if err := s.processEntitySubmission(entityID, submission, result); err != nil {
			result.Errors++
			result.ErrorDetails = append(result.ErrorDetails, err.Error())
			log.Printf("Error processing infrastruktur entity %s: %v", entityID, err)
		}
	}

	// Find and delete infrastruktur that no longer exist in ODK Central
	var infraList []model.Infrastruktur
	if err := s.db.Where("entity_id != '' AND deleted_at IS NULL").Find(&infraList).Error; err != nil {
		result.Errors++
		result.ErrorDetails = append(result.ErrorDetails, fmt.Sprintf("failed to fetch existing infrastruktur: %v", err))
	} else {
		for _, infra := range infraList {
			if infra.EntityID != "" && !entityIDSet[infra.EntityID] {
				log.Printf("HardSync: Deleting infrastruktur %s (entity: %s) - no longer in ODK", infra.Nama, infra.EntityID)

				// Delete associated photos first
				if err := s.db.Where("infrastruktur_id = ?", infra.ID).Delete(&model.InfrastrukturPhoto{}).Error; err != nil {
					log.Printf("Warning: failed to delete photos for infrastruktur %s: %v", infra.ID, err)
				}

				// Delete the infrastruktur
				if err := s.db.Delete(&infra).Error; err != nil {
					result.Errors++
					result.ErrorDetails = append(result.ErrorDetails, fmt.Sprintf("failed to delete infrastruktur %s: %v", infra.ID, err))
				} else {
					result.Deleted++
				}
			}
		}
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime).String()

	s.updateSyncStateSuccess(result.TotalFetched)

	log.Printf("HardSync Infrastruktur completed: %d fetched, %d entities, %d created, %d updated, %d deleted, %d errors",
		result.TotalFetched, len(latestByEntity), result.Created, result.Updated, result.Deleted, result.Errors)

	return result, nil
}
