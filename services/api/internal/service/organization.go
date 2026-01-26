package service

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/leksa/datamapper-senyar/internal/model"
	"github.com/leksa/datamapper-senyar/internal/repository"
)

// OrganizationService handles organization business logic
type OrganizationService struct {
	repo *repository.OrganizationRepository
}

// NewOrganizationService creates a new organization service
func NewOrganizationService(repo *repository.OrganizationRepository) *OrganizationService {
	return &OrganizationService{repo: repo}
}

// CreateOrganizationInput represents input for creating an organization
type CreateOrganizationInput struct {
	Name         string  `json:"name" binding:"required"`
	Slug         string  `json:"slug"`
	Description  *string `json:"description"`
	Email        *string `json:"email"`
	Phone        *string `json:"phone"`
	Address      *string `json:"address"`
	LogoURL      *string `json:"logo_url"`
	ODKProjectID *int    `json:"odk_project_id"`

	// Admin invitation (optional - if provided, invite user as org admin)
	AdminEmail *string `json:"admin_email,omitempty"`
	AdminName  *string `json:"admin_name,omitempty"`
}

// CreateOrganizationResult represents the result of creating an organization with admin
type CreateOrganizationResult struct {
	Organization   *model.Organization `json:"organization"`
	InvitedAdmin   *model.User         `json:"invited_admin,omitempty"`
	InvitationLink string              `json:"invitation_link,omitempty"`
	IsNewAdmin     bool                `json:"is_new_admin,omitempty"`
}

// UpdateOrganizationInput represents input for updating an organization
type UpdateOrganizationInput struct {
	Name         *string `json:"name"`
	Slug         *string `json:"slug"`
	Description  *string `json:"description"`
	Email        *string `json:"email"`
	Phone        *string `json:"phone"`
	Address      *string `json:"address"`
	LogoURL      *string `json:"logo_url"`
	ODKProjectID *int    `json:"odk_project_id"`
	IsActive     *bool   `json:"is_active"`
}

// List returns paginated organizations
func (s *OrganizationService) List(ctx context.Context, filter repository.OrganizationFilter) (*repository.OrganizationListResult, error) {
	return s.repo.List(ctx, filter)
}

// ListWithOrgFilter returns paginated organizations filtered by org IDs
// If orgIDs is nil, returns all organizations (for super_admin)
// If orgIDs is empty, returns empty result
// If orgIDs has values, returns only those organizations
func (s *OrganizationService) ListWithOrgFilter(ctx context.Context, filter repository.OrganizationFilter, orgIDs []uuid.UUID) (*repository.OrganizationListResult, error) {
	if orgIDs == nil {
		// No filter - return all (for super_admin)
		return s.repo.List(ctx, filter)
	}
	// Filter by org IDs
	return s.repo.ListByOrgIDs(ctx, filter, orgIDs)
}

// GetByID retrieves an organization by ID
func (s *OrganizationService) GetByID(ctx context.Context, id string) (*model.Organization, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid organization ID format")
	}
	return s.repo.FindByID(ctx, parsed)
}

// GetByIDWithRelations retrieves an organization with all relations
func (s *OrganizationService) GetByIDWithRelations(ctx context.Context, id string) (*model.Organization, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid organization ID format")
	}
	return s.repo.FindByIDWithRelations(ctx, parsed)
}

// GetBySlug retrieves an organization by slug
func (s *OrganizationService) GetBySlug(ctx context.Context, slug string) (*model.Organization, error) {
	return s.repo.FindBySlug(ctx, slug)
}

// Create creates a new organization
func (s *OrganizationService) Create(ctx context.Context, input CreateOrganizationInput) (*model.Organization, error) {
	// Generate slug if not provided
	slug := input.Slug
	if slug == "" {
		slug = generateSlug(input.Name)
	} else {
		slug = generateSlug(slug)
	}

	// Validate slug
	if !isValidSlug(slug) {
		return nil, errors.New("invalid slug format")
	}

	// Check if slug already exists
	exists, err := s.repo.SlugExists(ctx, slug, nil)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("slug already exists")
	}

	org := &model.Organization{
		Name:         input.Name,
		Slug:         slug,
		Description:  input.Description,
		Email:        input.Email,
		Phone:        input.Phone,
		Address:      input.Address,
		LogoURL:      input.LogoURL,
		ODKProjectID: input.ODKProjectID,
		IsActive:     true,
	}

	if err := s.repo.Create(ctx, org); err != nil {
		return nil, err
	}

	return org, nil
}

// Update updates an existing organization
func (s *OrganizationService) Update(ctx context.Context, id string, input UpdateOrganizationInput) (*model.Organization, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid organization ID format")
	}

	org, err := s.repo.FindByID(ctx, parsed)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if input.Name != nil {
		org.Name = *input.Name
	}
	if input.Slug != nil {
		slug := generateSlug(*input.Slug)
		if !isValidSlug(slug) {
			return nil, errors.New("invalid slug format")
		}
		exists, err := s.repo.SlugExists(ctx, slug, &org.ID)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("slug already exists")
		}
		org.Slug = slug
	}
	if input.Description != nil {
		org.Description = input.Description
	}
	if input.Email != nil {
		org.Email = input.Email
	}
	if input.Phone != nil {
		org.Phone = input.Phone
	}
	if input.Address != nil {
		org.Address = input.Address
	}
	if input.LogoURL != nil {
		org.LogoURL = input.LogoURL
	}
	if input.ODKProjectID != nil {
		org.ODKProjectID = input.ODKProjectID
	}
	if input.IsActive != nil {
		org.IsActive = *input.IsActive
	}

	if err := s.repo.Update(ctx, org); err != nil {
		return nil, err
	}

	return org, nil
}

// Delete soft-deletes an organization
func (s *OrganizationService) Delete(ctx context.Context, id string) error {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return errors.New("invalid organization ID format")
	}
	return s.repo.Delete(ctx, parsed)
}

// GetStats returns statistics for an organization
func (s *OrganizationService) GetStats(ctx context.Context, id string) (map[string]int64, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid organization ID format")
	}
	return s.repo.GetStats(ctx, parsed)
}

// AddMember adds a user to an organization
func (s *OrganizationService) AddMember(ctx context.Context, orgID, userID string, role model.OrgMemberRole) error {
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		return errors.New("invalid organization ID format")
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("invalid user ID format")
	}

	member := &model.OrganizationMember{
		OrganizationID: orgUUID,
		UserID:         userUUID,
		Role:           role,
	}
	return s.repo.AddMember(ctx, member)
}

// RemoveMember removes a user from an organization
func (s *OrganizationService) RemoveMember(ctx context.Context, orgID, userID string) error {
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		return errors.New("invalid organization ID format")
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("invalid user ID format")
	}
	return s.repo.RemoveMember(ctx, orgUUID, userUUID)
}

// UpdateMemberRole updates a member's role in an organization
func (s *OrganizationService) UpdateMemberRole(ctx context.Context, orgID, userID string, role model.OrgMemberRole) error {
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		return errors.New("invalid organization ID format")
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("invalid user ID format")
	}
	return s.repo.UpdateMemberRole(ctx, orgUUID, userUUID, role)
}

// GetUserOrganizations returns all organizations a user belongs to
func (s *OrganizationService) GetUserOrganizations(ctx context.Context, userID string) ([]model.Organization, error) {
	parsed, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}
	return s.repo.GetUserOrganizations(ctx, parsed)
}

// generateSlug creates a URL-safe slug from a string
func generateSlug(s string) string {
	// Convert to lowercase
	slug := strings.ToLower(s)
	// Replace spaces and special chars with hyphens
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	slug = reg.ReplaceAllString(slug, "-")
	// Remove leading/trailing hyphens
	slug = strings.Trim(slug, "-")
	return slug
}

// isValidSlug checks if a slug is valid
func isValidSlug(slug string) bool {
	if len(slug) < 2 || len(slug) > 100 {
		return false
	}
	matched, _ := regexp.MatchString(`^[a-z0-9]+(?:-[a-z0-9]+)*$`, slug)
	return matched
}
