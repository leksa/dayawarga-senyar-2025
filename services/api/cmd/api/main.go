package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/leksa/datamapper-senyar/internal/config"
	"github.com/leksa/datamapper-senyar/internal/handler"
	"github.com/leksa/datamapper-senyar/internal/middleware"
	"github.com/leksa/datamapper-senyar/internal/odk"
	"github.com/leksa/datamapper-senyar/internal/repository"
	"github.com/leksa/datamapper-senyar/internal/scheduler"
	"github.com/leksa/datamapper-senyar/internal/service"
	"github.com/leksa/datamapper-senyar/internal/sse"
	"github.com/leksa/datamapper-senyar/internal/storage"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Setup database connection
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)

	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Connected to database successfully")

	// Initialize repositories
	locationRepo := repository.NewLocationRepository(db)
	feedRepo := repository.NewFeedRepository(db)
	faskesRepo := repository.NewFaskesRepository(db)

	// Initialize ODK client for posko form
	odkPoskoConfig := &odk.ODKConfig{
		BaseURL:   cfg.ODKBaseURL,
		Email:     cfg.ODKEmail,
		Password:  cfg.ODKPassword,
		ProjectID: cfg.ODKProjectID,
		FormID:    cfg.ODKFormID,
	}
	odkPoskoClient := odk.NewClient(odkPoskoConfig)

	// Initialize ODK client for feed form
	odkFeedConfig := &odk.ODKConfig{
		BaseURL:   cfg.ODKBaseURL,
		Email:     cfg.ODKEmail,
		Password:  cfg.ODKPassword,
		ProjectID: cfg.ODKProjectID,
		FormID:    cfg.ODKFeedFormID,
	}
	odkFeedClient := odk.NewClient(odkFeedConfig)

	// Initialize ODK client for faskes form
	odkFaskesConfig := &odk.ODKConfig{
		BaseURL:   cfg.ODKBaseURL,
		Email:     cfg.ODKEmail,
		Password:  cfg.ODKPassword,
		ProjectID: cfg.ODKProjectID,
		FormID:    cfg.ODKFaskesFormID,
	}
	odkFaskesClient := odk.NewClient(odkFaskesConfig)

	// Initialize services
	syncService := service.NewSyncService(db, odkPoskoClient, cfg.ODKFormID)
	feedSyncService := service.NewFeedSyncService(db, odkFeedClient, cfg.ODKFeedFormID)
	faskesSyncService := service.NewFaskesSyncService(db, odkFaskesClient, cfg.ODKFaskesFormID)

	// Initialize photo service (with optional S3 storage)
	var photoService *service.PhotoService
	if cfg.S3Enabled {
		s3Config := storage.S3Config{
			Endpoint:        cfg.S3Endpoint,
			Bucket:          cfg.S3Bucket,
			AccessKeyID:     cfg.S3AccessKeyID,
			SecretAccessKey: cfg.S3SecretAccessKey,
			Region:          cfg.S3Region,
			PathPrefix:      cfg.S3PathPrefix,
			UsePathStyle:    true, // Required for S3-compatible storage like CloudHost
		}
		s3Storage, err := storage.NewS3Storage(s3Config)
		if err != nil {
			log.Fatalf("Failed to initialize S3 storage: %v", err)
		}
		photoService = service.NewPhotoServiceWithS3(db, odkPoskoClient, cfg.PhotoStoragePath, s3Storage)
		log.Printf("S3 storage enabled: %s/%s", cfg.S3Endpoint, cfg.S3Bucket)
	} else {
		photoService = service.NewPhotoService(db, odkPoskoClient, cfg.PhotoStoragePath)
		log.Println("Using local filesystem for photo storage")
	}

	// Initialize SSE Hub for real-time updates
	sseHub := sse.NewHub()

	// Initialize Scheduler
	schedulerConfig := scheduler.DefaultConfig()
	autoScheduler := scheduler.NewScheduler(schedulerConfig, syncService, feedSyncService, sseHub)

	// Start scheduler if enabled
	if os.Getenv("SCHEDULER_ENABLED") != "false" {
		autoScheduler.Start()
		log.Println("Auto-scheduler started")
	}

	// Initialize handlers
	locationHandler := handler.NewLocationHandler(locationRepo, feedRepo)
	feedHandler := handler.NewFeedHandler(feedRepo)
	faskesHandler := handler.NewFaskesHandler(faskesRepo)
	healthHandler := handler.NewHealthHandler(db)
	syncHandler := handler.NewSyncHandler(syncService, feedSyncService, faskesSyncService)
	photoHandler := handler.NewPhotoHandler(photoService)
	sseHandler := handler.NewSSEHandler(sseHub)
	schedulerHandler := handler.NewSchedulerHandler(autoScheduler)

	// Initialize middleware
	rateLimiter := middleware.DefaultRateLimiter()
	cache := middleware.DefaultCache()

	// Setup Gin router
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// Configure CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000", "https://dayawarga.com", "https://www.dayawarga.com"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "X-Cache", "X-RateLimit-Limit", "X-RateLimit-Remaining"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Apply global middleware
	r.Use(rateLimiter.Middleware())

	// Health endpoints (no cache, no rate limit heavy)
	r.GET("/health", healthHandler.Check)
	r.GET("/ready", healthHandler.Ready)

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Apply cache middleware to read endpoints
		cached := v1.Group("")
		cached.Use(cache.Middleware())
		{
			// Locations (cached)
			cached.GET("/locations", locationHandler.GetLocations)
			cached.GET("/locations/:id", locationHandler.GetLocationByID)

			// Faskes - Health facilities (cached)
			cached.GET("/faskes", faskesHandler.GetFaskes)
			cached.GET("/faskes/:id", faskesHandler.GetFaskesByID)

			// Feeds (cached)
			cached.GET("/feeds", feedHandler.GetFeeds)
			cached.GET("/locations/:id/feeds", feedHandler.GetFeedsByLocation)

			// Photos (cached)
			// Posko photos
			cached.GET("/locations/:id/photos", photoHandler.GetPhotosByLocation)
			cached.GET("/photos/:id/file", photoHandler.GetPhotoFile)
			// Feed photos
			cached.GET("/feeds/photos/:id/file", photoHandler.GetFeedPhotoFile)
			// Faskes photos
			cached.GET("/faskes/:id/photos", photoHandler.GetPhotosByFaskes)
			cached.GET("/faskes/photos/:id/file", photoHandler.GetFaskesPhotoFile)
		}

		// SSE Events (no cache, streaming)
		v1.GET("/events", sseHandler.Stream)

		// Protected endpoints - require API key
		protected := v1.Group("")
		protected.Use(middleware.APIKeyAuth(cfg.SyncAPIKey))
		{
			// Sync endpoints
			protected.POST("/sync/posko", syncHandler.SyncAll)
			protected.POST("/sync/feed", syncHandler.SyncFeeds)
			protected.POST("/sync/faskes", syncHandler.SyncFaskes)
			protected.POST("/sync/photos", photoHandler.SyncPhotos)              // Posko photos
			protected.POST("/sync/feed-photos", photoHandler.SyncFeedPhotos)     // Feed photos
			protected.POST("/sync/faskes-photos", photoHandler.SyncFaskesPhotos) // Faskes photos
			protected.POST("/migrate/s3", photoHandler.MigrateToS3)              // Migrate local photos to S3
			protected.POST("/photos/reset-cache", photoHandler.ResetCache)       // Reset cache for missing files

			// Hard sync endpoints - sync AND delete records not in ODK Central
			protected.POST("/sync/posko/hard", syncHandler.HardSyncPosko)
			protected.POST("/sync/feed/hard", syncHandler.HardSyncFeeds)
			protected.POST("/sync/faskes/hard", syncHandler.HardSyncFaskes)

			// Scheduler endpoints
			protected.GET("/scheduler/status", schedulerHandler.GetStatus)
			protected.POST("/scheduler/start", schedulerHandler.Start)
			protected.POST("/scheduler/stop", schedulerHandler.Stop)
			protected.POST("/scheduler/trigger", schedulerHandler.TriggerSync)
			protected.POST("/scheduler/mode/:mode", schedulerHandler.SetMode)
			protected.POST("/scheduler/mode/auto", schedulerHandler.ClearManualMode)
		}

		// Sync status endpoints (read-only, no auth required)
		v1.GET("/sync/status", syncHandler.GetSyncStatus)
		v1.GET("/sync/feed/status", syncHandler.GetFeedSyncStatus)
		v1.GET("/sync/faskes/status", syncHandler.GetFaskesSyncStatus)
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down gracefully...")
		autoScheduler.Stop()
		sqlDB.Close()
		os.Exit(0)
	}()

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Starting server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
