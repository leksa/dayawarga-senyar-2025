package repository

import (
	"context"

	"github.com/google/uuid"
)

// OrgCheckerWrapper wraps OrganizationRepository to implement auth.OrganizationChecker interface
type OrgCheckerWrapper struct {
	repo *OrganizationRepository
}

// NewOrgCheckerWrapper creates a new OrgCheckerWrapper
func NewOrgCheckerWrapper(repo *OrganizationRepository) *OrgCheckerWrapper {
	return &OrgCheckerWrapper{repo: repo}
}

// GetUserOrganizations returns organization IDs that a user belongs to
func (w *OrgCheckerWrapper) GetUserOrganizations(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	return w.repo.GetUserOrganizationIDs(ctx, userID)
}

// IsUserOrgAdmin checks if a user is an admin of a specific organization
func (w *OrgCheckerWrapper) IsUserOrgAdmin(ctx context.Context, userID, orgID uuid.UUID) (bool, error) {
	return w.repo.IsUserOrgAdmin(ctx, userID, orgID)
}
