package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/leksa/datamapper-senyar/internal/config"
	"github.com/leksa/datamapper-senyar/internal/odk"
	"github.com/leksa/datamapper-senyar/internal/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Parse command line flags
	syncPhotos := flag.Bool("photos", false, "Sync all uncached photos from ODK")
	syncPosko := flag.Bool("posko", false, "Sync posko data from ODK")
	syncAll := flag.Bool("all", false, "Sync everything (posko + photos)")
	dryRun := flag.Bool("dry-run", false, "Show what would be done without making changes")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	locationID := flag.String("location", "", "Sync photos for specific location UUID")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `ODK Data Importer - Import data and images from ODK Central

Usage:
  importer [options]

Options:
`)
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
Examples:
  # Sync all uncached photos
  importer -photos

  # Sync posko data only
  importer -posko

  # Sync everything
  importer -all

  # Dry run to see what would be synced
  importer -photos -dry-run

  # Sync photos for specific location
  importer -photos -location=<uuid>

Environment Variables:
  DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME
  ODK_BASE_URL, ODK_EMAIL, ODK_PASSWORD, ODK_PROJECT_ID, ODK_FORM_ID
  PHOTO_STORAGE_PATH
`)
	}

	flag.Parse()

	if !*syncPhotos && !*syncPosko && !*syncAll {
		flag.Usage()
		os.Exit(1)
	}

	// Load configuration
	cfg := config.Load()

	// Setup logging
	logLevel := logger.Silent
	if *verbose {
		logLevel = logger.Info
	}

	// Connect to database
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Connected to database")

	// Create ODK client
	odkConfig := &odk.ODKConfig{
		BaseURL:   cfg.ODKBaseURL,
		Email:     cfg.ODKEmail,
		Password:  cfg.ODKPassword,
		ProjectID: cfg.ODKProjectID,
		FormID:    cfg.ODKFormID,
	}
	odkClient := odk.NewClient(odkConfig)

	// Run requested operations
	startTime := time.Now()

	if *syncAll || *syncPosko {
		if err := runPoskoSync(db, odkClient, cfg.ODKFormID, *dryRun, *verbose); err != nil {
			log.Printf("Posko sync error: %v", err)
		}
	}

	if *syncAll || *syncPhotos {
		if err := runPhotoSync(db, odkClient, cfg.PhotoStoragePath, *dryRun, *verbose, *locationID); err != nil {
			log.Printf("Photo sync error: %v", err)
		}
	}

	log.Printf("Import completed in %v", time.Since(startTime))
}

func runPoskoSync(db *gorm.DB, odkClient *odk.Client, formID string, dryRun, verbose bool) error {
	log.Println("=== Starting Posko Sync ===")

	syncService := service.NewSyncService(db, odkClient, formID)

	if dryRun {
		// Just fetch and show stats
		submissions, err := odkClient.GetApprovedSubmissions()
		if err != nil {
			return fmt.Errorf("failed to fetch submissions: %w", err)
		}
		log.Printf("[DRY-RUN] Found %d approved submissions in ODK", len(submissions))

		// Count existing in DB
		var count int64
		db.Table("locations").Where("deleted_at IS NULL").Count(&count)
		log.Printf("[DRY-RUN] Currently %d locations in database", count)
		return nil
	}

	result, err := syncService.SyncAll()
	if err != nil {
		return err
	}

	log.Printf("Posko sync completed:")
	log.Printf("  - Fetched: %d", result.TotalFetched)
	log.Printf("  - Created: %d", result.Created)
	log.Printf("  - Updated: %d", result.Updated)
	log.Printf("  - Errors: %d", result.Errors)

	return nil
}

func runPhotoSync(db *gorm.DB, odkClient *odk.Client, storagePath string, dryRun, verbose bool, locationID string) error {
	log.Println("=== Starting Photo Sync ===")

	photoService := service.NewPhotoService(db, odkClient, storagePath)

	if dryRun {
		// Count uncached photos
		var count int64
		query := db.Table("location_photos").Where("is_cached = false")
		if locationID != "" {
			query = query.Where("location_id = ?", locationID)
		}
		query.Count(&count)

		log.Printf("[DRY-RUN] Found %d uncached photos to download", count)

		if verbose && count > 0 {
			var photos []struct {
				Filename  string
				PhotoType string
				Nama      string `gorm:"column:nama"`
			}
			db.Table("location_photos").
				Select("location_photos.filename, location_photos.photo_type, locations.nama").
				Joins("LEFT JOIN locations ON locations.id = location_photos.location_id").
				Where("location_photos.is_cached = false").
				Limit(20).
				Find(&photos)

			log.Println("[DRY-RUN] Sample of photos to download:")
			for _, p := range photos {
				log.Printf("  - %s (%s) from %s", p.Filename, p.PhotoType, p.Nama)
			}
			if count > 20 {
				log.Printf("  ... and %d more", count-20)
			}
		}
		return nil
	}

	result, err := photoService.SyncAllPhotos()
	if err != nil {
		return err
	}

	log.Printf("Photo sync completed:")
	log.Printf("  - Total found: %d", result.TotalFound)
	log.Printf("  - Downloaded: %d", result.Downloaded)
	log.Printf("  - Errors: %d", result.Errors)
	log.Printf("  - Duration: %s", result.Duration)

	if verbose && len(result.ErrorDetails) > 0 {
		log.Println("Error details:")
		for _, e := range result.ErrorDetails {
			log.Printf("  - %s", e)
		}
	}

	return nil
}
