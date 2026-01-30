package handler

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/leksa/datamapper-senyar/internal/auth"
	"github.com/leksa/datamapper-senyar/internal/model"
	"github.com/stretchr/testify/assert"
)

// MockOrgChecker implements auth.OrganizationChecker for testing
type MockOrgChecker struct {
	userOrgs  map[uuid.UUID][]uuid.UUID
	orgAdmins map[string]bool // key: "userID:orgID"
}

func NewMockOrgChecker() *MockOrgChecker {
	return &MockOrgChecker{
		userOrgs:  make(map[uuid.UUID][]uuid.UUID),
		orgAdmins: make(map[string]bool),
	}
}

func (m *MockOrgChecker) SetUserOrganizations(userID uuid.UUID, orgIDs []uuid.UUID) {
	m.userOrgs[userID] = orgIDs
}

func (m *MockOrgChecker) SetOrgAdmin(userID, orgID uuid.UUID, isAdmin bool) {
	key := userID.String() + ":" + orgID.String()
	m.orgAdmins[key] = isAdmin
}

func (m *MockOrgChecker) GetUserOrganizations(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	if orgs, ok := m.userOrgs[userID]; ok {
		return orgs, nil
	}
	return []uuid.UUID{}, nil
}

func (m *MockOrgChecker) IsUserOrgAdmin(ctx context.Context, userID, orgID uuid.UUID) (bool, error) {
	key := userID.String() + ":" + orgID.String()
	return m.orgAdmins[key], nil
}

// setupTestContext creates a gin context with user and org checker for testing
func setupTestContext(user *model.User, orgChecker *MockOrgChecker) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Create a dummy request so c.Request.Context() works
	req := httptest.NewRequest("GET", "/", nil)
	c.Request = req

	// Set user in context using the same key as auth package
	c.Set("auth_user", user)

	// Set org checker in context
	auth.SetOrgChecker(c, orgChecker)

	return c, w
}

func TestGetUserOrgIDs_SuperAdmin(t *testing.T) {
	superAdmin := &model.User{
		ID:   uuid.New(),
		Role: model.UserRoleSuperAdmin,
	}

	orgChecker := NewMockOrgChecker()
	c, _ := setupTestContext(superAdmin, orgChecker)

	// Super admin should get nil (no filtering)
	orgIDs, ok := auth.GetUserOrgIDs(c)
	assert.True(t, ok)
	assert.Nil(t, orgIDs, "Super admin should get nil orgIDs (no filtering)")
}

func TestGetUserOrgIDs_OrgAdmin(t *testing.T) {
	orgAdminID := uuid.New()
	org1 := uuid.New()
	org2 := uuid.New()

	orgAdmin := &model.User{
		ID:   orgAdminID,
		Role: model.UserRoleOrgAdmin,
	}

	orgChecker := NewMockOrgChecker()
	orgChecker.SetUserOrganizations(orgAdminID, []uuid.UUID{org1, org2})

	c, _ := setupTestContext(orgAdmin, orgChecker)

	// Org admin should get their org IDs
	orgIDs, ok := auth.GetUserOrgIDs(c)
	assert.True(t, ok)
	assert.NotNil(t, orgIDs)
	assert.Len(t, orgIDs, 2)
	assert.Contains(t, orgIDs, org1)
	assert.Contains(t, orgIDs, org2)
}

func TestCanAccessOrganization_SuperAdmin(t *testing.T) {
	superAdmin := &model.User{
		ID:   uuid.New(),
		Role: model.UserRoleSuperAdmin,
	}

	orgChecker := NewMockOrgChecker()
	c, _ := setupTestContext(superAdmin, orgChecker)

	anyOrgID := uuid.New()

	// Super admin can access any org
	canAccess := auth.CanAccessOrganization(c, anyOrgID)
	assert.True(t, canAccess, "Super admin should be able to access any organization")
}

func TestCanAccessOrganization_OrgAdmin_OwnOrg(t *testing.T) {
	orgAdminID := uuid.New()
	ownOrg := uuid.New()

	orgAdmin := &model.User{
		ID:   orgAdminID,
		Role: model.UserRoleOrgAdmin,
	}

	orgChecker := NewMockOrgChecker()
	orgChecker.SetUserOrganizations(orgAdminID, []uuid.UUID{ownOrg})

	c, _ := setupTestContext(orgAdmin, orgChecker)

	// Org admin can access their own org
	canAccess := auth.CanAccessOrganization(c, ownOrg)
	assert.True(t, canAccess, "Org admin should be able to access their own organization")
}

func TestCanAccessOrganization_OrgAdmin_OtherOrg(t *testing.T) {
	orgAdminID := uuid.New()
	ownOrg := uuid.New()
	otherOrg := uuid.New()

	orgAdmin := &model.User{
		ID:   orgAdminID,
		Role: model.UserRoleOrgAdmin,
	}

	orgChecker := NewMockOrgChecker()
	orgChecker.SetUserOrganizations(orgAdminID, []uuid.UUID{ownOrg})

	c, _ := setupTestContext(orgAdmin, orgChecker)

	// Org admin cannot access other org
	canAccess := auth.CanAccessOrganization(c, otherOrg)
	assert.False(t, canAccess, "Org admin should NOT be able to access other organization")
}

func TestCanManageOrganization_SuperAdmin(t *testing.T) {
	superAdmin := &model.User{
		ID:   uuid.New(),
		Role: model.UserRoleSuperAdmin,
	}

	orgChecker := NewMockOrgChecker()
	c, _ := setupTestContext(superAdmin, orgChecker)

	anyOrgID := uuid.New()

	// Super admin can manage any org
	canManage := auth.CanManageOrganization(c, anyOrgID)
	assert.True(t, canManage, "Super admin should be able to manage any organization")
}

func TestCanManageOrganization_OrgAdmin_OwnOrg(t *testing.T) {
	orgAdminID := uuid.New()
	ownOrg := uuid.New()

	orgAdmin := &model.User{
		ID:   orgAdminID,
		Role: model.UserRoleOrgAdmin,
	}

	orgChecker := NewMockOrgChecker()
	orgChecker.SetOrgAdmin(orgAdminID, ownOrg, true)

	c, _ := setupTestContext(orgAdmin, orgChecker)

	// Org admin can manage their own org where they are admin
	canManage := auth.CanManageOrganization(c, ownOrg)
	assert.True(t, canManage, "Org admin should be able to manage their own organization")
}

func TestCanManageOrganization_OrgAdmin_OtherOrg(t *testing.T) {
	orgAdminID := uuid.New()
	ownOrg := uuid.New()
	otherOrg := uuid.New()

	orgAdmin := &model.User{
		ID:   orgAdminID,
		Role: model.UserRoleOrgAdmin,
	}

	orgChecker := NewMockOrgChecker()
	orgChecker.SetOrgAdmin(orgAdminID, ownOrg, true)
	orgChecker.SetOrgAdmin(orgAdminID, otherOrg, false)

	c, _ := setupTestContext(orgAdmin, orgChecker)

	// Org admin cannot manage other org
	canManage := auth.CanManageOrganization(c, otherOrg)
	assert.False(t, canManage, "Org admin should NOT be able to manage other organization")
}

func TestCanAccessOrganization_MemberRole(t *testing.T) {
	memberID := uuid.New()
	someOrg := uuid.New()

	member := &model.User{
		ID:   memberID,
		Role: model.UserRoleMember,
	}

	orgChecker := NewMockOrgChecker()
	// Even if member is in org list, they're not super_admin or org_admin
	orgChecker.SetUserOrganizations(memberID, []uuid.UUID{someOrg})

	c, _ := setupTestContext(member, orgChecker)

	// Member cannot access org (not super_admin or org_admin)
	canAccess := auth.CanAccessOrganization(c, someOrg)
	assert.True(t, canAccess, "Member with org membership should be able to access organization")
}

func TestGetUserOrgIDs_NoUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// No user set in context
	orgIDs, ok := auth.GetUserOrgIDs(c)
	assert.False(t, ok)
	assert.Nil(t, orgIDs)
}
