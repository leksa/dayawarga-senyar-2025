# Draft: Dayawarga Admin Portal Enhancement

## Requirements (confirmed)

### Requirement 1: Auto-sync User Creation to ODK
- **Current state**: Relawan created WITHOUT ODK app user; manual button click required
- **Desired**: Auto-create ODK app user when relawan is created (if group has approved project)
- **Gap**: `ApproveRequest` in `project_request.go` does NOT call `EnsureAppUserForGroupRelawan`

### Requirement 2: Admin Group Login with Appropriate Roles
- **Current state**: RBAC system exists (super_admin, org_admin, member)
- **Current state**: Permission middleware exists (`CanAccessOrganization`, `CanManageOrganization`)
- **Desired**: Ensure org_admin can login via Authentik and see only their org data
- **Observation**: Auth flow appears to be working; need to verify invitation flow completeness

### Requirement 3: Org Admin Can Invite Users
- **Current state**: `InviteUser` handler exists in `invitation.go`
- **Current state**: Handler accepts `organization_id` parameter
- **Gap**: Need to verify permission check - does org_admin check happen?
- **Desired**: org_admin can invite users to THEIR organization only

### Requirement 4: Organization Profile Page Enhancement
- **Current state** (Organization model has): name, slug, description, email, phone, address, logo_url, odk_project_id
- **Missing fields**: city, country, website_url, bidang/klaster (JSONB array), instagram, social_media
- **Frontend current tabs**: Tim/Grup, Relawan
- **Desired tabs**: Daftar Relawan, All Activities, Donasi dan Bantuan Disalurkan

## Technical Decisions

### Backend Changes
1. **Organization model**: Add new fields (city, country, website_url, bidang, instagram, social_media)
2. **Migration**: New SQL migration for organization table alteration
3. **RelawanService.Create**: Call ODK app user creation automatically when conditions met
4. **ProjectRequestService.ApproveRequest**: Call `EnsureAppUserForGroupRelawan` after approval
5. **InvitationHandler**: Add permission check for org_admin inviting to their own org

### Frontend Changes
1. **OrganizationDetailView.vue**: Enhance to profile page with new tabs
2. **RelawanQRCodeModal.vue**: Generate actual visual QR code image (not just text)
3. **types.ts**: Add new Organization fields
4. **Organization create/edit forms**: Add new fields

## Research Findings

### Existing ODK Integration
- `odk/app_users.go`: `CreateAppUser(projectID, displayName)` returns token
- `odk.GenerateQRCodeData(baseURL, projectID, token)` generates JSON config
- `RelawanODKService.CreateAppUserForRelawan()` - exists but not auto-called
- `RelawanODKService.EnsureAppUserForGroupRelawan()` - exists, creates for all relawan in group

### Existing Invitation Flow
- Email invitation with token
- Set password via API
- PIN generated for WhatsApp verification
- User becomes active after PIN verification

### QR Code Display Gap
- Current `RelawanQRCodeModal.vue` shows placeholder, not actual QR image
- Need to use a QR library (qrcode.vue or similar) to render visual QR

## Questions Resolved (User Answers)

1. **Bidang/Klaster**: Separate reference table with IDs (database table of valid bidang)
2. **Activities tab**: System activity log from existing `feeds` table - relationship via `submitter_name` → relawan
3. **Donasi tab**: SKIP for now - lower priority
4. **QR Code library**: To be determined (will use standard Vue QR library)
5. **Social media fields**: Single JSONB column `social_media`

## Test Strategy Decision
- **Infrastructure exists**: Need to verify Go test files
- **User wants tests**: YES (TDD approach)
- **QA approach**: TDD - RED-GREEN-REFACTOR for each task

## Scope Boundaries

### INCLUDE
- Auto-create ODK app user on relawan creation (when group has approved project)
- Auto-create app users for all group relawan on project approval
- Org admin invitation capability (to their org only)
- Organization profile enhancement:
  - New fields: city, country, website_url, social_media (JSONB)
  - New reference table: bidang/klaster (many-to-many)
- Visual QR code rendering in modal
- Activities tab using feeds data

### EXCLUDE
- Donasi tab (deferred to later)
- New authentication flows (Authentik already works)
