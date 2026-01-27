// @title           Daya Warga API
// @version         1.0
// @description     API untuk sistem informasi Daya Warga - platform manajemen data posko pengungsi, fasilitas kesehatan, dan infrastruktur.
// @termsOfService  https://dayawarga.com/terms

// @contact.name   Daya Warga Team
// @contact.url    https://dayawarga.com
// @contact.email  support@dayawarga.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description Enter the token with the `Bearer ` prefix, e.g. "Bearer abcde12345"

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
	_ "github.com/leksa/datamapper-senyar/docs" // Swagger docs
	"github.com/leksa/datamapper-senyar/internal/auth"
	"github.com/leksa/datamapper-senyar/internal/config"
	"github.com/leksa/datamapper-senyar/internal/email"
	"github.com/leksa/datamapper-senyar/internal/handler"
	"github.com/leksa/datamapper-senyar/internal/middleware"
	"github.com/leksa/datamapper-senyar/internal/odk"
	"github.com/leksa/datamapper-senyar/internal/repository"
	"github.com/leksa/datamapper-senyar/internal/scheduler"
	"github.com/leksa/datamapper-senyar/internal/service"
	"github.com/leksa/datamapper-senyar/internal/sse"
	"github.com/leksa/datamapper-senyar/internal/storage"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
	infrastrukturRepo := repository.NewInfrastrukturRepository(db)

	// Admin Portal repositories
	userRepo := repository.NewUserRepository(db)
	orgRepo := repository.NewOrganizationRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	relawanRepo := repository.NewRelawanRepository(db)
	projectRequestRepo := repository.NewProjectRequestRepository(db)
	groupProjectRepo := repository.NewGroupProjectRepository(db)
	orgChecker := repository.NewOrgCheckerWrapper(orgRepo)

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

	// Initialize ODK client for infrastruktur form
	odkInfrastrukturConfig := &odk.ODKConfig{
		BaseURL:   cfg.ODKBaseURL,
		Email:     cfg.ODKEmail,
		Password:  cfg.ODKPassword,
		ProjectID: cfg.ODKProjectID,
		FormID:    cfg.ODKInfrastrukturFormID,
	}
	odkInfrastrukturClient := odk.NewClient(odkInfrastrukturConfig)

	// Initialize services
	syncService := service.NewSyncService(db, odkPoskoClient, cfg.ODKFormID)
	feedSyncService := service.NewFeedSyncService(db, odkFeedClient, cfg.ODKFeedFormID)
	faskesSyncService := service.NewFaskesSyncService(db, odkFaskesClient, cfg.ODKFaskesFormID)
	infrastrukturSyncService := service.NewInfrastrukturSyncService(db, odkInfrastrukturClient, cfg.ODKInfrastrukturFormID)

	// Admin Portal services
	userService := service.NewUserService(userRepo)
	orgService := service.NewOrganizationService(orgRepo)
	groupService := service.NewGroupService(groupRepo, orgRepo)
	relawanService := service.NewRelawanService(relawanRepo, orgRepo, groupRepo)
	odkProjectService := service.NewODKProjectService(odkPoskoClient)
	projectRequestService := service.NewProjectRequestService(projectRequestRepo, groupProjectRepo, groupRepo, userRepo, odkPoskoClient, db)
	relawanODKService := service.NewRelawanODKService(relawanRepo, groupRepo, odkPoskoClient, cfg.ODKBaseURL, db)
	userODKService := service.NewUserODKService(userRepo, odkPoskoClient, db)
	orgODKService := service.NewOrganizationODKService(orgRepo, userRepo, odkPoskoClient, db)

	// Email and Invitation services
	emailService := email.NewService(cfg)
	invitationService := service.NewInvitationService(userRepo, orgRepo, emailService, cfg)

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
	chatbotHandler := handler.NewChatbotHandler(db, locationRepo, feedRepo)
	locationHandler := handler.NewLocationHandler(locationRepo, feedRepo)
	feedHandler := handler.NewFeedHandler(feedRepo)
	faskesHandler := handler.NewFaskesHandler(faskesRepo)
	infrastrukturHandler := handler.NewInfrastrukturHandler(infrastrukturRepo)
	healthHandler := handler.NewHealthHandler(db)
	syncHandler := handler.NewSyncHandlerWithInfrastruktur(syncService, feedSyncService, faskesSyncService, infrastrukturSyncService)
	photoHandler := handler.NewPhotoHandler(photoService)
	sseHandler := handler.NewSSEHandler(sseHub)
	schedulerHandler := handler.NewSchedulerHandler(autoScheduler)

	// Admin Portal handlers
	authHandler := handler.NewAuthHandler(userService)
	userHandler := handler.NewUserHandler(userService)
	userODKHandler := handler.NewUserODKHandler(userODKService)
	orgHandler := handler.NewOrganizationHandler(orgService)
	groupHandler := handler.NewGroupHandler(groupService)
	relawanHandler := handler.NewRelawanHandler(relawanService)
	odkHandler := handler.NewODKHandler(odkProjectService, projectRequestService)
	relawanODKHandler := handler.NewRelawanODKHandler(relawanODKService)
	orgODKHandler := handler.NewOrganizationODKHandler(orgODKService)
	invitationHandler := handler.NewInvitationHandler(invitationService, orgService)

	// Initialize OIDC validator for Admin Portal (if configured)
	var oidcValidator *auth.OIDCValidator
	if cfg.OIDCIssuerURL != "" && cfg.OIDCClientID != "" {
		oidcValidator = auth.NewOIDCValidator(auth.OIDCConfig{
			IssuerURL: cfg.OIDCIssuerURL,
			ClientID:  cfg.OIDCClientID,
		})
		log.Printf("OIDC validator initialized: %s", cfg.OIDCIssuerURL)
	} else {
		log.Println("OIDC not configured - Admin Portal authentication disabled")
	}

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
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:5179", "http://localhost:3000", "https://dayawarga.com", "https://www.dayawarga.com"},
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

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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

			// Infrastruktur - Roads/Bridges (cached)
			cached.GET("/infrastruktur", infrastrukturHandler.GetInfrastruktur)
			cached.GET("/infrastruktur/:id", infrastrukturHandler.GetInfrastrukturByID)
			cached.GET("/infrastruktur/stats", infrastrukturHandler.GetInfrastrukturStats)

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
			// Chatbot endpoints
			chatbot := protected.Group("/chatbot")
			{
				chatbot.POST("/feeds", chatbotHandler.CreateFeed)
				chatbot.POST("/locations", chatbotHandler.CreateLocation)
				chatbot.PUT("/locations/:id", chatbotHandler.UpdateLocation)
				chatbot.GET("/wilayah/kabupaten", chatbotHandler.GetWilayahKabupaten)
				chatbot.GET("/wilayah/kecamatan", chatbotHandler.GetWilayahKecamatan)
				chatbot.GET("/wilayah/desa", chatbotHandler.GetWilayahDesa)
			}

			// WhatsApp relawan validation (for chatbot)
			wa := protected.Group("/wa")
			{
				wa.GET("/validate", relawanHandler.ValidateWAAccess)
				wa.POST("/activity", relawanHandler.RecordWAActivity)
			}

			// Sync endpoints
			protected.POST("/sync/posko", syncHandler.SyncAll)
			protected.POST("/sync/feed", syncHandler.SyncFeeds)
			protected.POST("/sync/faskes", syncHandler.SyncFaskes)
			protected.POST("/sync/infrastruktur", syncHandler.SyncInfrastruktur)
			protected.POST("/sync/photos", photoHandler.SyncPhotos)              // Posko photos
			protected.POST("/sync/feed-photos", photoHandler.SyncFeedPhotos)     // Feed photos
			protected.POST("/sync/faskes-photos", photoHandler.SyncFaskesPhotos) // Faskes photos
			protected.POST("/migrate/s3", photoHandler.MigrateToS3)              // Migrate local photos to S3
			protected.POST("/photos/reset-cache", photoHandler.ResetCache)       // Reset cache for missing files

			// Hard sync endpoints - sync AND delete records not in ODK Central
			protected.POST("/sync/posko/hard", syncHandler.HardSyncPosko)
			protected.POST("/sync/feed/hard", syncHandler.HardSyncFeeds)
			protected.POST("/sync/faskes/hard", syncHandler.HardSyncFaskes)
			protected.POST("/sync/infrastruktur/hard", syncHandler.HardSyncInfrastruktur)

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
		v1.GET("/sync/infrastruktur/status", syncHandler.GetInfrastrukturSyncStatus)

		// ============================================
		// Admin Portal Routes (OIDC Protected)
		// ============================================
		if oidcValidator != nil {
			// Create auth middleware
			authMiddleware := auth.Middleware(oidcValidator, userService)

			// Middleware to inject org checker into context
			orgCheckerMiddleware := func(c *gin.Context) {
				auth.SetOrgChecker(c, orgChecker)
				c.Next()
			}

			// Admin Portal group with OIDC auth
			admin := v1.Group("")
			admin.Use(authMiddleware)
			admin.Use(orgCheckerMiddleware)
			{
				// Auth endpoints
				admin.GET("/auth/me", authHandler.Me)

				// Organizations - require org_admin or super_admin role
				orgs := admin.Group("/organizations")
				orgs.Use(auth.RequireOrgAdminOrAbove())
				{
					orgs.GET("", orgHandler.List)
					orgs.POST("", auth.RequireSuperAdmin(), orgHandler.Create)
					orgs.GET("/:id", orgHandler.GetByID)
					orgs.PUT("/:id", orgHandler.Update)
					orgs.DELETE("/:id", auth.RequireSuperAdmin(), orgHandler.Delete)
					orgs.GET("/:id/stats", orgHandler.GetStats)
					orgs.POST("/:id/members", orgHandler.AddMember)
					orgs.DELETE("/:id/members/:user_id", orgHandler.RemoveMember)
					orgs.PUT("/:id/members/:user_id/role", orgHandler.UpdateMemberRole)
					// ODK Project assignment (super_admin only)
					orgs.POST("/:id/odk-project", auth.RequireSuperAdmin(), orgODKHandler.AssignODKProject)
					orgs.DELETE("/:id/odk-project", auth.RequireSuperAdmin(), orgODKHandler.RemoveODKProject)
					orgs.GET("/:id/odk-info", orgODKHandler.GetODKInfo)
					// Create organization with admin (super_admin only)
					orgs.POST("/with-admin", auth.RequireSuperAdmin(), invitationHandler.CreateOrganizationWithAdmin)
				}

				// Groups - require org_admin or super_admin role
				groups := admin.Group("/groups")
				groups.Use(auth.RequireOrgAdminOrAbove())
				{
					groups.GET("", groupHandler.List)
					groups.POST("", groupHandler.Create)
					groups.GET("/:id", groupHandler.GetByID)
					groups.PUT("/:id", groupHandler.Update)
					groups.DELETE("/:id", groupHandler.Delete)
					groups.GET("/:id/stats", groupHandler.GetStats)
					// Project request for group
					groups.POST("/:id/project-request", odkHandler.CreateProjectRequest)
					groups.GET("/:id/project-requests", odkHandler.GetGroupProjectRequests)
					// Bulk create ODK app users for group
					groups.POST("/:id/odk-app-users", relawanODKHandler.CreateGroupAppUsers)
				}

				// Relawan - require org_admin or super_admin role
				relawan := admin.Group("/relawan")
				relawan.Use(auth.RequireOrgAdminOrAbove())
				{
					relawan.GET("", relawanHandler.List)
					relawan.GET("/stats", relawanHandler.GetStats)
					relawan.POST("", relawanHandler.Create)
					relawan.GET("/:id", relawanHandler.GetByID)
					relawan.PUT("/:id", relawanHandler.Update)
					relawan.DELETE("/:id", relawanHandler.Delete)
					relawan.PUT("/:id/status", relawanHandler.UpdateStatus)
					relawan.PUT("/:id/group", relawanHandler.MoveToGroup)
					relawan.POST("/bulk/move-to-group", relawanHandler.BulkMoveToGroup)
					// ODK App User management
					relawan.POST("/:id/odk-app-user", relawanODKHandler.CreateAppUser)
					relawan.DELETE("/:id/odk-app-user", relawanODKHandler.RevokeAppUser)
					relawan.GET("/:id/odk-qr-code", relawanODKHandler.GetQRCode)
					relawan.POST("/:id/odk-forms", relawanODKHandler.AssignForms)
					// WhatsApp verification management
					relawan.POST("/:id/wa-verify", relawanHandler.SetWAVerified)
					relawan.DELETE("/:id/wa-verify", relawanHandler.RevokeWAVerified)
					relawan.GET("/:id/wa-status", relawanHandler.GetWAStatus)
				}

				// ODK Projects (read from ODK Central)
				odk := admin.Group("/odk")
				odk.Use(auth.RequireOrgAdminOrAbove())
				{
					odk.GET("/projects", odkHandler.ListProjects)
					odk.GET("/projects/:id", odkHandler.GetProject)
					odk.GET("/projects/:id/forms", odkHandler.ListProjectForms)
				}

				// Invitations - require org_admin or super_admin role
				invitations := admin.Group("/invitations")
				invitations.Use(auth.RequireOrgAdminOrAbove())
				{
					invitations.POST("", invitationHandler.InviteUser)
					invitations.POST("/:user_id/resend", invitationHandler.ResendInvitation)
					invitations.DELETE("/:user_id", invitationHandler.CancelInvitation)
				}

				// Admin-only endpoints
				adminOnly := admin.Group("/admin")
				adminOnly.Use(auth.RequireSuperAdmin())
				{
					adminOnly.GET("/users", userHandler.List)
					adminOnly.GET("/users/:id", userHandler.Get)
					adminOnly.PUT("/users/:id", userHandler.Update)

					adminOnly.GET("/users/:id/odk-roles", userODKHandler.GetUserProjectAssignments)
					adminOnly.POST("/users/:id/odk-roles", userODKHandler.AssignProjectRole)
					adminOnly.DELETE("/users/:id/odk-roles/:projectId", userODKHandler.RemoveProjectRole)
					adminOnly.GET("/users/:id/odk-qr-code", userODKHandler.GetUserQRCode)

					adminOnly.GET("/odk-projects/:id/assignments", userODKHandler.GetProjectAssignments)

					adminOnly.GET("/project-requests", odkHandler.ListProjectRequests)
					adminOnly.GET("/project-requests/:id", odkHandler.GetProjectRequest)
					adminOnly.PUT("/project-requests/:id", odkHandler.ReviewProjectRequest)

					invitations.POST("/organizations/with-admin", invitationHandler.CreateOrganizationWithAdmin)
				}
			}
		}

		// Public invitation endpoints (no auth required)
		v1.GET("/invitations/validate", invitationHandler.ValidateToken)
		v1.POST("/invitations/accept", invitationHandler.AcceptInvitation)
		v1.POST("/invitations/set-password", invitationHandler.SetPassword)
		v1.GET("/invitations/verification-status/:user_id", invitationHandler.GetVerificationStatus)
		v1.POST("/invitations/regenerate-pin/:user_id", invitationHandler.RegeneratePIN)
		v1.POST("/invitations/verify-pin", invitationHandler.VerifyPIN)
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
