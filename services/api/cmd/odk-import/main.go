package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/leksa/datamapper-senyar/internal/config"
	"github.com/leksa/datamapper-senyar/internal/model"
	"github.com/leksa/datamapper-senyar/internal/odk"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ImportResult tracks import statistics
type ImportResult struct {
	WebUsersFound      int
	WebUsersCreated    int
	WebUsersUpdated    int
	WebUsersSkipped    int
	AppUsersFound      int
	AppUsersCreated    int
	AppUsersUpdated    int
	AppUsersSkipped    int
	ODKAppUsersCreated int // App users created in ODK for QR codes
	ODKAppUsersFailed  int
	Errors             []string
}

func main() {
	importWebUsers := flag.Bool("web-users", false, "Import web users (email login users)")
	importAppUsers := flag.Bool("app-users", false, "Import app users (ODK Collect field keys)")
	importAll := flag.Bool("all", false, "Import all users (web + app)")
	createODKAppUsers := flag.Bool("create-odk-app-users", false, "Create ODK App Users for all imported users (enables QR codes)")
	backfillAppUsers := flag.Bool("backfill-app-users", false, "Create ODK App Users for existing users without one")
	dryRun := flag.Bool("dry-run", false, "Show what would be imported without making changes")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	listOnly := flag.Bool("list", false, "List users from ODK Central without importing")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `ODK User Importer - Import users from ODK Central to Admin Portal

This tool imports existing users from ODK Central into the admin portal database.
Users are matched by email (for web users). If no match is found, a new user
record is created with status 'pending_invitation'.

Usage:
  odk-import [options]

Options:
`)
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
Examples:
  # List all users in ODK Central (without importing)
  odk-import -list

  # Dry run to see what would be imported
  odk-import -all -dry-run

  # Import web users only
  odk-import -web-users

  # Import all users
  odk-import -all

  # Verbose output
  odk-import -all -verbose

  # Import and create ODK App Users for QR codes
  odk-import -web-users -create-odk-app-users

  # Backfill existing users with ODK App Users
  odk-import -backfill-app-users

Environment Variables:
  DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME
  ODK_BASE_URL, ODK_EMAIL, ODK_PASSWORD, ODK_PROJECT_ID
`)
	}

	flag.Parse()

	if !*importWebUsers && !*importAppUsers && !*importAll && !*listOnly && !*backfillAppUsers {
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
	}
	odkClient := odk.NewClient(odkConfig)

	startTime := time.Now()

	// List only mode
	if *listOnly {
		if err := listODKUsers(odkClient, cfg.ODKProjectID, *verbose); err != nil {
			log.Fatalf("Failed to list users: %v", err)
		}
		return
	}

	// Import mode
	result := &ImportResult{}

	if *importAll || *importWebUsers {
		if err := importWebUsersFromODK(db, odkClient, cfg.ODKProjectID, *dryRun, *verbose, *createODKAppUsers, result); err != nil {
			log.Printf("Error importing web users: %v", err)
		}
	}

	if *importAll || *importAppUsers {
		if err := importAppUsersFromODK(db, odkClient, cfg.ODKProjectID, *dryRun, *verbose, result); err != nil {
			log.Printf("Error importing app users: %v", err)
		}
	}

	if *backfillAppUsers {
		if err := backfillODKAppUsers(db, odkClient, cfg.ODKProjectID, *dryRun, *verbose, result); err != nil {
			log.Printf("Error backfilling ODK app users: %v", err)
		}
	}

	// Print summary
	log.Println("")
	log.Println("=== Import Summary ===")
	if *dryRun {
		log.Println("[DRY-RUN MODE - No changes made]")
	}
	log.Printf("Web Users - Found: %d, Created: %d, Updated: %d, Skipped: %d",
		result.WebUsersFound, result.WebUsersCreated, result.WebUsersUpdated, result.WebUsersSkipped)
	log.Printf("App Users - Found: %d, Created: %d, Updated: %d, Skipped: %d",
		result.AppUsersFound, result.AppUsersCreated, result.AppUsersUpdated, result.AppUsersSkipped)
	if result.ODKAppUsersCreated > 0 || result.ODKAppUsersFailed > 0 {
		log.Printf("ODK App Users (QR codes) - Created: %d, Failed: %d",
			result.ODKAppUsersCreated, result.ODKAppUsersFailed)
	}

	if len(result.Errors) > 0 {
		log.Printf("Errors: %d", len(result.Errors))
		for _, e := range result.Errors {
			log.Printf("  - %s", e)
		}
	}

	log.Printf("Completed in %v", time.Since(startTime))
}

// listODKUsers lists all users from ODK Central without importing
func listODKUsers(odkClient *odk.Client, projectID int, verbose bool) error {
	log.Println("=== ODK Central Users ===")

	// List web users
	log.Println("\n--- Web Users (Email Login) ---")
	webUsers, err := odkClient.ListWebUsers()
	if err != nil {
		return fmt.Errorf("failed to fetch web users: %w", err)
	}

	log.Printf("Found %d web users:", len(webUsers))
	for _, u := range webUsers {
		status := "active"
		if u.DeletedAt != nil {
			status = "deleted"
		}
		log.Printf("  [%d] %s <%s> (%s)", u.ID, u.DisplayName, u.Email, status)
	}

	// Get project assignments
	assignments, err := odkClient.ListProjectAssignments(projectID)
	if err != nil {
		log.Printf("Warning: Could not fetch project assignments: %v", err)
	} else if verbose {
		log.Printf("\nProject %d Role Assignments:", projectID)
		for _, a := range assignments {
			roleName := roleIDToName(a.RoleID)
			log.Printf("  Actor %d: %s", a.ActorID, roleName)
		}
	}

	// List app users
	log.Println("\n--- App Users (ODK Collect) ---")
	appUsers, err := odkClient.ListAppUsers(projectID)
	if err != nil {
		return fmt.Errorf("failed to fetch app users: %w", err)
	}

	log.Printf("Found %d app users in project %d:", len(appUsers), projectID)
	for _, u := range appUsers {
		status := "active"
		if u.DeletedAt != nil {
			status = "deleted"
		}
		log.Printf("  [%d] %s (%s)", u.ID, u.DisplayName, status)
	}

	return nil
}

func importWebUsersFromODK(db *gorm.DB, odkClient *odk.Client, projectID int, dryRun, verbose, createAppUsers bool, result *ImportResult) error {
	log.Println("\n=== Importing Web Users ===")

	// Fetch web users from ODK
	webUsers, err := odkClient.ListWebUsers()
	if err != nil {
		return fmt.Errorf("failed to fetch web users: %w", err)
	}

	result.WebUsersFound = len(webUsers)
	log.Printf("Found %d web users in ODK Central", len(webUsers))

	// Fetch project assignments to determine roles
	assignments, err := odkClient.ListProjectAssignments(projectID)
	if err != nil {
		log.Printf("Warning: Could not fetch project assignments: %v", err)
	}

	// Build assignment map: actorID -> roleID
	assignmentMap := make(map[int]int)
	for _, a := range assignments {
		assignmentMap[a.ActorID] = a.RoleID
	}

	ctx := context.Background()

	for _, odkUser := range webUsers {
		// Skip deleted users
		if odkUser.DeletedAt != nil {
			if verbose {
				log.Printf("  Skipping deleted user: %s", odkUser.Email)
			}
			result.WebUsersSkipped++
			continue
		}

		// Skip users without email
		if odkUser.Email == "" {
			if verbose {
				log.Printf("  Skipping user without email: ID %d", odkUser.ID)
			}
			result.WebUsersSkipped++
			continue
		}

		// Determine role from ODK assignments
		portalRole := determinePortalRole(odkUser.ID, assignmentMap)

		if verbose {
			log.Printf("  Processing: %s <%s> -> role: %s",
				odkUser.DisplayName, odkUser.Email, portalRole)
		}

		if dryRun {
			// Check if user exists
			var existingUser model.User
			err := db.WithContext(ctx).Where("email = ?", strings.ToLower(odkUser.Email)).First(&existingUser).Error
			if err == nil {
				log.Printf("  [DRY-RUN] Would update existing user: %s (ODK ID: %d)", odkUser.Email, odkUser.ID)
				result.WebUsersUpdated++
			} else {
				log.Printf("  [DRY-RUN] Would create new user: %s <%s> (ODK ID: %d, role: %s)",
					odkUser.DisplayName, odkUser.Email, odkUser.ID, portalRole)
				result.WebUsersCreated++
			}
			continue
		}

		// Check if user exists by email
		var existingUser model.User
		err := db.WithContext(ctx).Where("email = ?", strings.ToLower(odkUser.Email)).First(&existingUser).Error

		if err == nil {
			// User exists - update ODK web user ID if not set
			if existingUser.ODKWebUserID == nil || *existingUser.ODKWebUserID != odkUser.ID {
				odkID := odkUser.ID
				existingUser.ODKWebUserID = &odkID
				if err := db.WithContext(ctx).Save(&existingUser).Error; err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("Failed to update user %s: %v", odkUser.Email, err))
					continue
				}
				log.Printf("  Updated user: %s (linked ODK ID: %d)", odkUser.Email, odkUser.ID)
				result.WebUsersUpdated++
			} else {
				if verbose {
					log.Printf("  Skipped (already linked): %s", odkUser.Email)
				}
				result.WebUsersSkipped++
			}
		} else {
			// Create new user with pending invitation status
			odkID := odkUser.ID
			newUser := &model.User{
				ID:           uuid.New(),
				OIDCSubject:  fmt.Sprintf("odk-import-%d", odkUser.ID), // Placeholder, will be updated on first login
				Email:        strings.ToLower(odkUser.Email),
				Name:         &odkUser.DisplayName,
				Role:         portalRole,
				Status:       model.UserStatusPendingInvitation,
				IsActive:     true,
				ODKWebUserID: &odkID,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}

			if err := db.WithContext(ctx).Create(newUser).Error; err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Failed to create user %s: %v", odkUser.Email, err))
				continue
			}

			log.Printf("  Created user: %s <%s> (ODK ID: %d, role: %s)",
				odkUser.DisplayName, odkUser.Email, odkUser.ID, portalRole)
			result.WebUsersCreated++

			if createAppUsers && !dryRun {
				if err := createODKAppUserForUser(db, odkClient, projectID, newUser, verbose, result); err != nil {
					log.Printf("  Warning: Failed to create ODK App User for %s: %v", odkUser.Email, err)
				}
			}
		}
	}

	return nil
}

// importAppUsersFromODK imports app users (field keys) from ODK Central
// App users are linked to existing relawan records by matching display name
func importAppUsersFromODK(db *gorm.DB, odkClient *odk.Client, projectID int, dryRun, verbose bool, result *ImportResult) error {
	log.Println("\n=== Importing App Users (ODK Collect) ===")

	// Fetch app users from ODK
	appUsers, err := odkClient.ListAppUsers(projectID)
	if err != nil {
		return fmt.Errorf("failed to fetch app users: %w", err)
	}

	result.AppUsersFound = len(appUsers)
	log.Printf("Found %d app users in project %d", len(appUsers), projectID)

	// App users don't have email and require organization_id in relawan table
	// We can only update existing relawan records that match by name or ID
	log.Println("Note: App users can only be linked to existing relawan records")
	log.Println("      Create relawan via admin portal first, then run this import to link ODK app user IDs")
	log.Println("")

	for _, appUser := range appUsers {
		// Skip deleted users
		if appUser.DeletedAt != nil {
			if verbose {
				log.Printf("  Skipping deleted app user: %s", appUser.DisplayName)
			}
			result.AppUsersSkipped++
			continue
		}

		if verbose {
			log.Printf("  Processing app user: %s (ID: %d)", appUser.DisplayName, appUser.ID)
		}

		// Check if relawan already has this ODK app user ID
		var existingByODKID int64
		db.Table("relawan").Where("odk_app_user_id = ? AND deleted_at IS NULL", appUser.ID).Count(&existingByODKID)
		if existingByODKID > 0 {
			if verbose {
				log.Printf("  Skipped (already linked): %s", appUser.DisplayName)
			}
			result.AppUsersSkipped++
			continue
		}

		// Try to find relawan by exact name match (case-insensitive)
		var relawan struct {
			ID           uuid.UUID
			Name         string
			ODKAppUserID *int
		}
		err := db.Table("relawan").
			Select("id, name, odk_app_user_id").
			Where("LOWER(name) = LOWER(?) AND deleted_at IS NULL AND odk_app_user_id IS NULL", appUser.DisplayName).
			First(&relawan).Error

		if err != nil {
			// No matching relawan found
			log.Printf("  No match found for: %s (ODK ID: %d) - create relawan manually first",
				appUser.DisplayName, appUser.ID)
			result.AppUsersSkipped++
			continue
		}

		if dryRun {
			log.Printf("  [DRY-RUN] Would link relawan '%s' to ODK App User ID: %d",
				relawan.Name, appUser.ID)
			result.AppUsersUpdated++
			continue
		}

		// Update relawan with ODK app user ID
		now := time.Now()
		err = db.Table("relawan").
			Where("id = ?", relawan.ID).
			Updates(map[string]interface{}{
				"odk_app_user_id":         appUser.ID,
				"odk_app_user_created_at": appUser.CreatedAt,
				"updated_at":              now,
			}).Error

		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to link relawan %s: %v", relawan.Name, err))
			continue
		}

		log.Printf("  Linked relawan '%s' to ODK App User ID: %d", relawan.Name, appUser.ID)
		result.AppUsersUpdated++
	}

	return nil
}

// determinePortalRole determines the admin portal role based on ODK assignments
func determinePortalRole(odkUserID int, assignmentMap map[int]int) model.UserRole {
	roleID, hasAssignment := assignmentMap[odkUserID]
	if !hasAssignment {
		return model.UserRoleMember
	}

	switch roleID {
	case odk.RoleAdmin:
		// Site Administrator -> super_admin
		return model.UserRoleSuperAdmin
	case odk.RoleProjectManager:
		// Project Manager -> org_admin
		return model.UserRoleOrgAdmin
	case odk.RoleProjectViewer, odk.RoleDataCollector:
		// Project Viewer or Data Collector -> member
		return model.UserRoleMember
	default:
		return model.UserRoleMember
	}
}

func roleIDToName(roleID int) string {
	switch roleID {
	case odk.RoleAdmin:
		return "Site Administrator"
	case odk.RoleProjectManager:
		return "Project Manager"
	case odk.RoleProjectViewer:
		return "Project Viewer"
	case odk.RoleDataCollector:
		return "Data Collector"
	default:
		return fmt.Sprintf("Unknown (%d)", roleID)
	}
}

func createODKAppUserForUser(db *gorm.DB, odkClient *odk.Client, projectID int, user *model.User, verbose bool, result *ImportResult) error {
	if user.ODKAppUserID != nil && *user.ODKAppUserID > 0 {
		if verbose {
			log.Printf("    User %s already has ODK App User ID: %d", user.Email, *user.ODKAppUserID)
		}
		return nil
	}

	displayName := user.Email
	if user.Name != nil && *user.Name != "" {
		displayName = *user.Name
	}

	appUser, err := odkClient.CreateAppUser(projectID, displayName)
	if err != nil {
		result.ODKAppUsersFailed++
		return fmt.Errorf("failed to create ODK app user: %w", err)
	}

	updates := map[string]interface{}{
		"odk_app_user_id":         appUser.ID,
		"odk_app_user_project_id": projectID,
		"updated_at":              time.Now(),
	}
	if appUser.Token != "" {
		updates["odk_app_user_token"] = appUser.Token
	}

	if err := db.Model(&model.User{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
		result.ODKAppUsersFailed++
		return fmt.Errorf("failed to update user with ODK app user ID: %w", err)
	}

	log.Printf("    Created ODK App User for %s: ID=%d (token %s)",
		user.Email, appUser.ID, tokenStatus(appUser.Token))
	result.ODKAppUsersCreated++
	return nil
}

func tokenStatus(token string) string {
	if token != "" {
		return "saved"
	}
	return "not available"
}

func backfillODKAppUsers(db *gorm.DB, odkClient *odk.Client, projectID int, dryRun, verbose bool, result *ImportResult) error {
	log.Println("\n=== Backfilling ODK App Users ===")

	var users []model.User
	if err := db.Where("odk_app_user_id IS NULL OR odk_app_user_token IS NULL").Find(&users).Error; err != nil {
		return fmt.Errorf("failed to fetch users without ODK app users: %w", err)
	}

	log.Printf("Found %d users without ODK App Users", len(users))

	for _, user := range users {
		displayName := user.Email
		if user.Name != nil && *user.Name != "" {
			displayName = *user.Name
		}

		if verbose {
			log.Printf("  Processing: %s <%s>", displayName, user.Email)
		}

		if dryRun {
			log.Printf("  [DRY-RUN] Would create ODK App User for: %s", user.Email)
			result.ODKAppUsersCreated++
			continue
		}

		if err := createODKAppUserForUser(db, odkClient, projectID, &user, verbose, result); err != nil {
			log.Printf("  Error creating ODK App User for %s: %v", user.Email, err)
		}
	}

	return nil
}
