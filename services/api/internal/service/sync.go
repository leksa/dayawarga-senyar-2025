package service

import (
	"encoding/json"
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
	db                      *gorm.DB
	odkClient               *odk.Client
	formID                  string
	entityDataset           string
	submissionToEntityCache map[string]string // cache: submission ID -> entity UUID
}

// NewSyncService creates a new sync service
func NewSyncService(db *gorm.DB, odkClient *odk.Client, formID string) *SyncService {
	return &SyncService{
		db:            db,
		odkClient:     odkClient,
		formID:        formID,
		entityDataset: "posko_entities",
	}
}

// SyncResult holds the result of a sync operation
type SyncResult struct {
	TotalFetched int       `json:"total_fetched"`
	Created      int       `json:"created"`
	Updated      int       `json:"updated"`
	Deleted      int       `json:"deleted,omitempty"`
	Skipped      int       `json:"skipped,omitempty"`
	Errors       int       `json:"errors"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	Duration     string    `json:"duration"`
	ErrorDetails []string  `json:"error_details,omitempty"`
}

// SyncAll performs incremental synchronization of approved submissions
// Uses lastSyncTime to only fetch submissions updated since last sync
// Groups submissions by entity_id and only processes the latest submission per entity
func (s *SyncService) SyncAll() (*SyncResult, error) {
	result := &SyncResult{
		StartTime: time.Now(),
	}

	// Update sync state to "syncing"
	s.updateSyncState("syncing", nil)

	// Load entity mapping from ODK (for proper entity ID resolution)
	if err := s.loadEntityMapping(); err != nil {
		log.Printf("Warning: could not load entity mapping: %v", err)
	}

	// Get last sync time for incremental sync
	syncState, err := s.GetSyncState()
	if err != nil {
		log.Printf("Warning: could not get sync state: %v", err)
	}

	// Fetch submissions - incremental if we have last sync time
	var submissions []map[string]interface{}
	if syncState != nil && syncState.LastSyncTime != nil {
		// Incremental sync: fetch only submissions updated since last sync
		submissions, err = s.odkClient.GetApprovedSubmissionsSince(*syncState.LastSyncTime)
		if err != nil {
			errMsg := fmt.Sprintf("failed to fetch submissions: %v", err)
			s.updateSyncState("error", &errMsg)
			return nil, fmt.Errorf(errMsg)
		}
		log.Printf("Incremental sync: fetched %d submissions updated since %s", len(submissions), syncState.LastSyncTime.Format(time.RFC3339))
	} else {
		// Full sync: fetch all approved submissions
		submissions, err = s.odkClient.GetApprovedSubmissions()
		if err != nil {
			errMsg := fmt.Sprintf("failed to fetch submissions: %v", err)
			s.updateSyncState("error", &errMsg)
			return nil, fmt.Errorf(errMsg)
		}
		log.Printf("Full sync: fetched %d submissions from ODK Central", len(submissions))
	}

	result.TotalFetched = len(submissions)

	// If no new submissions, return early
	if len(submissions) == 0 {
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime).String()
		s.updateSyncStateSuccess(0)
		log.Printf("Sync completed: no new submissions to process")
		return result, nil
	}

	// Group submissions by entity_id and keep only the latest per entity
	latestByEntity := s.groupByEntityLatest(submissions)
	log.Printf("Grouped into %d unique entities", len(latestByEntity))

	// Process each entity's latest submission
	for entityID, submission := range latestByEntity {
		if err := s.processEntitySubmission(entityID, submission, result); err != nil {
			result.Errors++
			result.ErrorDetails = append(result.ErrorDetails, err.Error())
			log.Printf("Error processing entity %s: %v", entityID, err)
		}
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime).String()

	// Update sync state
	s.updateSyncStateSuccess(result.TotalFetched)

	log.Printf("Sync completed: %d fetched, %d entities, %d created, %d updated, %d errors",
		result.TotalFetched, len(latestByEntity), result.Created, result.Updated, result.Errors)

	return result, nil
}

// SyncAllFull performs a full synchronization of all approved submissions (ignores lastSyncTime)
// Use this for initial sync or when you need to re-sync everything
func (s *SyncService) SyncAllFull() (*SyncResult, error) {
	result := &SyncResult{
		StartTime: time.Now(),
	}

	// Update sync state to "syncing"
	s.updateSyncState("syncing", nil)

	// Load entity mapping from ODK (for proper entity ID resolution)
	if err := s.loadEntityMapping(); err != nil {
		log.Printf("Warning: could not load entity mapping: %v", err)
	}

	// Fetch all approved submissions
	submissions, err := s.odkClient.GetApprovedSubmissions()
	if err != nil {
		errMsg := fmt.Sprintf("failed to fetch submissions: %v", err)
		s.updateSyncState("error", &errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	result.TotalFetched = len(submissions)
	log.Printf("Full sync: fetched %d submissions from ODK Central", result.TotalFetched)

	// Group submissions by entity_id and keep only the latest per entity
	latestByEntity := s.groupByEntityLatest(submissions)
	log.Printf("Grouped into %d unique entities", len(latestByEntity))

	// Process each entity's latest submission
	for entityID, submission := range latestByEntity {
		if err := s.processEntitySubmission(entityID, submission, result); err != nil {
			result.Errors++
			result.ErrorDetails = append(result.ErrorDetails, err.Error())
			log.Printf("Error processing entity %s: %v", entityID, err)
		}
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime).String()

	// Update sync state
	s.updateSyncStateSuccess(result.TotalFetched)

	log.Printf("Sync completed: %d fetched, %d entities, %d created, %d updated, %d errors",
		result.TotalFetched, len(latestByEntity), result.Created, result.Updated, result.Errors)

	return result, nil
}

// groupByEntityLatest groups submissions by entity_id and returns only the latest submission per entity
// For mode="baru", entity_id is the ODK submission ID (__id)
// For mode="update", entity_id is sel_posko (the entity being updated)
func (s *SyncService) groupByEntityLatest(submissions []map[string]interface{}) map[string]map[string]interface{} {
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

		// Determine entity_id based on mode
		entityID := s.getEntityID(submission)
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

// loadEntityMapping fetches the entity-to-submission mapping from ODK Central
// and inverts it to submission-to-entity for efficient lookup
func (s *SyncService) loadEntityMapping() error {
	if s.submissionToEntityCache != nil {
		return nil // Already loaded
	}

	// Get entity -> submission mapping from ODK
	entityToSubmission, err := s.odkClient.GetEntitySubmissionMapping(s.entityDataset)
	if err != nil {
		log.Printf("Warning: could not load entity mapping: %v (will use submission ID as fallback)", err)
		s.submissionToEntityCache = make(map[string]string) // empty cache
		return nil
	}

	// Invert to submission -> entity mapping
	s.submissionToEntityCache = make(map[string]string)
	for entityUUID, submissionID := range entityToSubmission {
		s.submissionToEntityCache[submissionID] = entityUUID
	}

	log.Printf("Loaded entity mapping: %d entities", len(s.submissionToEntityCache))
	return nil
}

// getEntityID determines the entity ID for a submission
// Priority:
// 1. For mode="update": uses sel_posko (the entity being updated)
// 2. Look up in entity mapping cache (from ODK entity versions)
// 3. Fallback: use submission ID (for dumped data where submission ID = entity ID)
func (s *SyncService) getEntityID(submission map[string]interface{}) string {
	mode, _ := submission["mode"].(string)

	if mode == "update" {
		// For updates, use sel_posko as the entity ID
		if selPosko, ok := submission["sel_posko"].(string); ok && selPosko != "" {
			return selPosko
		}
	}

	// Get submission ID
	odkID, ok := submission["__id"].(string)
	if !ok || odkID == "" {
		return ""
	}

	// Check if we have entity UUID from mapping cache
	if s.submissionToEntityCache != nil {
		if entityUUID, exists := s.submissionToEntityCache[odkID]; exists {
			return entityUUID
		}
	}

	// Fallback: use submission ID as entity ID
	// This works for dumped submissions where we set entity ID = submission ID
	return odkID
}

// processEntitySubmission processes a submission for a specific entity.
// Skips if PostgreSQL record (from WhatsApp) is newer than ODK submission.
func (s *SyncService) processEntitySubmission(entityID string, submission map[string]interface{}, result *SyncResult) error {
	odkID, _ := submission["__id"].(string)

	if system, ok := submission["__system"].(map[string]interface{}); ok {
		if reviewState, ok := system["reviewState"].(string); ok && reviewState != "approved" {
			log.Printf("Skipping non-approved submission %s (state: %s)", odkID, reviewState)
			return nil
		}
	}

	location, err := MapSubmissionToLocation(submission)
	if err != nil {
		return fmt.Errorf("failed to map submission %s: %w", odkID, err)
	}

	if location.RawData == nil {
		location.RawData = model.JSONB{}
	}
	location.RawData["_entity_id"] = entityID
	location.ODKSubmissionID = &odkID

	submissionTime := s.extractSubmissionTime(submission)

	var existingLocation model.Location
	found := false

	err = s.db.Where("raw_data->>'_entity_id' = ?", entityID).First(&existingLocation).Error
	if err == nil {
		found = true
	} else if err == gorm.ErrRecordNotFound {
		err = s.db.Where("linked_entity_id = ?", entityID).First(&existingLocation).Error
		if err == nil {
			found = true
			log.Printf("Found location via linked_entity_id: %s -> %s", entityID, existingLocation.Nama)
		}
	}

	if !found {
		odkSource := "odk"
		location.Source = &odkSource
		if err := s.createLocation(location); err != nil {
			return fmt.Errorf("failed to create location for entity %s: %w", entityID, err)
		}
		result.Created++
		log.Printf("Created location: %s (entity: %s, submission: %s)", location.Nama, entityID, odkID)
	} else {
		if existingLocation.Source != nil && *existingLocation.Source == "whatsapp" {
			if existingLocation.UpdatedAt.After(submissionTime) {
				log.Printf("CONFLICT: Skipping entity %s - PostgreSQL data is newer (pg: %s > odk: %s)",
					entityID, existingLocation.UpdatedAt.Format(time.RFC3339), submissionTime.Format(time.RFC3339))
				s.logConflict("location", &existingLocation.ID, existingLocation.Nama, entityID, odkID,
					"timestamp_newer", existingLocation, submission, existingLocation.UpdatedAt, submissionTime)
				result.Skipped++
				return nil
			}
			log.Printf("CONFLICT: ODK submission is newer, will update WhatsApp record %s (pg: %s < odk: %s)",
				existingLocation.Nama, existingLocation.UpdatedAt.Format(time.RFC3339), submissionTime.Format(time.RFC3339))
			s.logConflict("location", &existingLocation.ID, existingLocation.Nama, entityID, odkID,
				"source_mismatch", existingLocation, submission, existingLocation.UpdatedAt, submissionTime)
		}

		location.ID = existingLocation.ID
		if err := s.updateLocation(location); err != nil {
			return fmt.Errorf("failed to update location for entity %s: %w", entityID, err)
		}
		result.Updated++
		log.Printf("Updated location: %s (entity: %s, submission: %s)", location.Nama, entityID, odkID)
	}

	photos := ExtractPhotos(submission)
	for _, photo := range photos {
		if err := s.processPhoto(location.ID, photo); err != nil {
			log.Printf("Warning: failed to process photo %s: %v", photo.Filename, err)
		}
	}

	return nil
}

func (s *SyncService) extractSubmissionTime(submission map[string]interface{}) time.Time {
	if system, ok := submission["__system"].(map[string]interface{}); ok {
		if dateStr, ok := system["submissionDate"].(string); ok {
			if t, err := time.Parse(time.RFC3339, dateStr); err == nil {
				return t
			}
		}
	}
	return time.Now()
}

func (s *SyncService) logConflict(entityType string, entityID *uuid.UUID, entityName, odkEntityID, odkSubmissionID,
	conflictType string, existing model.Location, incoming map[string]interface{},
	existingUpdatedAt, incomingSubmittedAt time.Time) {

	existingJSON, _ := json.Marshal(existing)
	incomingJSON, _ := json.Marshal(incoming)

	conflict := map[string]interface{}{
		"id":                    uuid.New(),
		"entity_type":           entityType,
		"entity_id":             entityID,
		"entity_name":           entityName,
		"odk_entity_id":         odkEntityID,
		"odk_submission_id":     odkSubmissionID,
		"conflict_type":         conflictType,
		"existing_data":         string(existingJSON),
		"incoming_data":         string(incomingJSON),
		"existing_updated_at":   existingUpdatedAt,
		"incoming_submitted_at": incomingSubmittedAt,
		"status":                "pending",
		"created_at":            time.Now(),
	}

	if err := s.db.Table("sync_conflicts").Create(&conflict).Error; err != nil {
		log.Printf("Warning: failed to log conflict: %v", err)
	}
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

// enrichAlamatWithWilayah looks up wilayah names from database and adds them to alamat
func (s *SyncService) enrichAlamatWithWilayah(alamat model.JSONB) {
	if alamat == nil {
		return
	}

	// Lookup provinsi name
	if idProv, ok := alamat["id_provinsi"].(string); ok && idProv != "" {
		var nama string
		s.db.Raw("SELECT nama FROM wilayah_provinsi WHERE kode = ?", idProv).Scan(&nama)
		if nama != "" {
			alamat["nama_provinsi"] = nama
		}
	}

	// Lookup kota/kab name
	if idKab, ok := alamat["id_kota_kab"].(string); ok && idKab != "" {
		var nama string
		s.db.Raw("SELECT nama FROM wilayah_kota_kab WHERE kode = ?", idKab).Scan(&nama)
		if nama != "" {
			alamat["nama_kota_kab"] = nama
		}
	}

	// Lookup kecamatan name
	if idKec, ok := alamat["id_kecamatan"].(string); ok && idKec != "" {
		var nama string
		s.db.Raw("SELECT nama FROM wilayah_kecamatan WHERE kode = ?", idKec).Scan(&nama)
		if nama != "" {
			alamat["nama_kecamatan"] = nama
		}
	}

	// Lookup desa name
	if idDesa, ok := alamat["id_desa"].(string); ok && idDesa != "" {
		var nama string
		s.db.Raw("SELECT nama FROM wilayah_desa WHERE kode = ?", idDesa).Scan(&nama)
		if nama != "" {
			alamat["nama_desa"] = nama
		}
	}
}

// createLocation creates a new location with PostGIS geometry
func (s *SyncService) createLocation(location *model.Location) error {
	location.ID = uuid.New()
	now := time.Now()
	location.CreatedAt = now
	location.UpdatedAt = now
	location.SyncedAt = &now

	// Enrich alamat with wilayah names if not already set
	if location.Alamat != nil {
		if nama, ok := location.Alamat["nama_provinsi"].(string); !ok || nama == "" {
			s.enrichAlamatWithWilayah(location.Alamat)
		}
	}

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

	// Enrich alamat with wilayah names if not already set
	if location.Alamat != nil {
		if nama, ok := location.Alamat["nama_provinsi"].(string); !ok || nama == "" {
			s.enrichAlamatWithWilayah(location.Alamat)
		}
	}

	sql := `
		UPDATE locations SET
			odk_submission_id = ?,
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
		location.ODKSubmissionID,
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

// HardSync performs a full sync and deletes records that no longer exist in ODK Central
// Uses entity-based grouping to properly handle ODK's append-only submission model
func (s *SyncService) HardSync() (*SyncResult, error) {
	result := &SyncResult{
		StartTime: time.Now(),
	}

	s.updateSyncState("hard_syncing", nil)

	// Load entity mapping from ODK (for proper entity ID resolution)
	// Reset cache to get fresh mapping
	s.submissionToEntityCache = nil
	if err := s.loadEntityMapping(); err != nil {
		log.Printf("Warning: could not load entity mapping: %v", err)
	}

	// Fetch all approved submissions from ODK Central
	submissions, err := s.odkClient.GetApprovedSubmissions()
	if err != nil {
		errMsg := fmt.Sprintf("failed to fetch submissions: %v", err)
		s.updateSyncState("error", &errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	result.TotalFetched = len(submissions)
	log.Printf("HardSync: Fetched %d submissions from ODK Central", result.TotalFetched)

	// Group submissions by entity_id and keep only the latest per entity
	latestByEntity := s.groupByEntityLatest(submissions)
	log.Printf("HardSync: Grouped into %d unique entities", len(latestByEntity))

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
			log.Printf("Error processing entity %s: %v", entityID, err)
		}
	}

	// Find and delete locations that no longer exist in ODK Central
	// Use entity_id for matching (consistent with entity-based upsert)
	var locations []model.Location
	if err := s.db.Where("raw_data->>'_entity_id' IS NOT NULL AND deleted_at IS NULL").Find(&locations).Error; err != nil {
		result.Errors++
		result.ErrorDetails = append(result.ErrorDetails, fmt.Sprintf("failed to fetch existing locations: %v", err))
	} else {
		for _, loc := range locations {
			// Get entity_id from raw_data
			entityID := ""
			if loc.RawData != nil {
				if eid, ok := loc.RawData["_entity_id"].(string); ok {
					entityID = eid
				}
			}

			if entityID != "" && !entityIDSet[entityID] {
				// This entity no longer exists in ODK Central - delete it
				log.Printf("HardSync: Deleting location %s (entity: %s) - no longer in ODK Central", loc.Nama, entityID)

				// Delete associated photos first
				if err := s.db.Where("location_id = ?", loc.ID).Delete(&model.LocationPhoto{}).Error; err != nil {
					log.Printf("Warning: failed to delete photos for location %s: %v", loc.ID, err)
				}

				// Delete the location
				if err := s.db.Delete(&loc).Error; err != nil {
					result.Errors++
					result.ErrorDetails = append(result.ErrorDetails, fmt.Sprintf("failed to delete location %s: %v", loc.ID, err))
				} else {
					result.Deleted++
				}
			}
		}
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime).String()

	s.updateSyncStateSuccess(result.TotalFetched)

	log.Printf("HardSync completed: %d fetched, %d entities, %d created, %d updated, %d deleted, %d errors",
		result.TotalFetched, len(latestByEntity), result.Created, result.Updated, result.Deleted, result.Errors)

	return result, nil
}
