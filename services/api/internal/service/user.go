package service

import (
	"context"
	"errors"
	"log"

	"github.com/google/uuid"
	"github.com/leksa/datamapper-senyar/internal/auth"
	"github.com/leksa/datamapper-senyar/internal/model"
	"github.com/leksa/datamapper-senyar/internal/repository"
	"gorm.io/gorm"
)

// UserService handles user business logic
type UserService struct {
	repo *repository.UserRepository
}

// NewUserService creates a new user service
func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// FindOrCreateFromOIDC finds an existing user by OIDC subject or creates a new one
// This implements the auto-provisioning flow
func (s *UserService) FindOrCreateFromOIDC(ctx context.Context, claims *auth.OIDCClaims) (*model.User, error) {
	// Try to find existing user by OIDC subject
	user, err := s.repo.FindByOIDCSubject(ctx, claims.Subject)
	if err == nil {
		// User exists, update profile if needed and return
		updated := s.updateUserFromClaims(user, claims)
		if updated {
			if err := s.repo.Save(ctx, user); err != nil {
				log.Printf("[UserService] Failed to update user profile: %v", err)
			}
		}

		// Update last login
		if err := s.repo.UpdateLastLogin(ctx, user.ID); err != nil {
			log.Printf("[UserService] Failed to update last login: %v", err)
		}

		return user, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// User doesn't exist, create new one
	log.Printf("[UserService] Creating new user from OIDC: subject=%s, email=%s", claims.Subject, claims.Email)

	newUser := &model.User{
		OIDCSubject: claims.Subject,
		Email:       claims.Email,
		Role:        model.UserRoleMember, // Default role for new users
		IsActive:    true,
	}

	// Set optional fields
	if claims.Name != "" {
		newUser.Name = &claims.Name
	}
	if claims.AvatarURL != "" {
		newUser.AvatarURL = &claims.AvatarURL
	}

	if err := s.repo.Create(ctx, newUser); err != nil {
		return nil, err
	}

	log.Printf("[UserService] Created new user: id=%s, email=%s", newUser.ID, newUser.Email)

	return newUser, nil
}

// updateUserFromClaims updates user fields from OIDC claims if they changed
func (s *UserService) updateUserFromClaims(user *model.User, claims *auth.OIDCClaims) bool {
	updated := false

	// Update email if changed
	if user.Email != claims.Email {
		user.Email = claims.Email
		updated = true
	}

	// Update name if changed
	if claims.Name != "" {
		if user.Name == nil || *user.Name != claims.Name {
			user.Name = &claims.Name
			updated = true
		}
	}

	// Update avatar if changed
	if claims.AvatarURL != "" {
		if user.AvatarURL == nil || *user.AvatarURL != claims.AvatarURL {
			user.AvatarURL = &claims.AvatarURL
			updated = true
		}
	}

	return updated
}

// GetByOIDCSubject retrieves a user by OIDC subject
func (s *UserService) GetByOIDCSubject(ctx context.Context, subject string) (*model.User, error) {
	return s.repo.FindByOIDCSubject(ctx, subject)
}

// GetByID retrieves a user by ID
func (s *UserService) GetByID(ctx context.Context, id string) (*model.User, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}
	return s.repo.FindByID(ctx, parsed)
}

// GetWithOrganizations retrieves a user with their organization memberships
func (s *UserService) GetWithOrganizations(ctx context.Context, id string) (*model.User, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}
	return s.repo.FindWithOrganizations(ctx, parsed)
}

// UserFilter defines filter options for listing users
type UserFilter struct {
	Role     string
	Status   string
	IsActive *bool
	Search   string
	Page     int
	PageSize int
}

// UserListResult represents the result of listing users
type UserListResult struct {
	Users      []model.User
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
}

// List returns a paginated list of users with optional filters
func (s *UserService) List(ctx context.Context, filter UserFilter) (*UserListResult, error) {
	// Set defaults
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}

	repoFilter := repository.UserFilter{
		Role:     filter.Role,
		IsActive: filter.IsActive,
		Search:   filter.Search,
		Page:     filter.Page,
		Limit:    filter.PageSize,
	}

	users, total, err := s.repo.FindAll(ctx, repoFilter)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / filter.PageSize
	if int(total)%filter.PageSize > 0 {
		totalPages++
	}

	return &UserListResult{
		Users:      users,
		Total:      total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: totalPages,
	}, nil
}

// Update updates a user
func (s *UserService) Update(ctx context.Context, id string, updates map[string]interface{}) (*model.User, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	user, err := s.repo.FindByID(ctx, parsed)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if name, ok := updates["name"].(string); ok {
		user.Name = &name
	}
	if role, ok := updates["role"].(string); ok {
		user.Role = model.UserRole(role)
	}
	if status, ok := updates["status"].(string); ok {
		user.Status = model.UserStatus(status)
	}
	if isActive, ok := updates["is_active"].(bool); ok {
		user.IsActive = isActive
	}

	if err := s.repo.Save(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}
