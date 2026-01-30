# Dayawarga Admin Portal Enhancement

## TL;DR

> **Quick Summary**: Implement auto-sync ODK user creation, enable org admin self-service invitations, and enhance organization profile with new fields (city, country, bidang/klaster reference table, social media JSONB) plus activities tab from feeds data.
> 
> **Deliverables**:
> - Auto-create ODK app users on relawan creation and project approval
> - Org admin can invite users to their organization (with WhatsApp PIN verification)
> - Enhanced Organization model with new fields + migration
> - Bidang/Klaster reference table with many-to-many relationship
> - Visual QR code rendering in RelawanQRCodeModal
> - Organization profile page with Relawan and Activities tabs
> 
> **Estimated Effort**: Large (15-20 tasks, ~3-5 days)
> **Parallel Execution**: YES - 4 waves
> **Critical Path**: Task 1 (Migration) → Task 2/3 (Backend Models) → Task 8/9 (Services) → Task 14 (Frontend Profile)

---

## Context

### Original Request
Enhance Dayawarga Admin Portal with:
1. Auto-sync user/relawan creation to ODK (auto-approve, auto QR generation)
2. Admin group login with appropriate roles (org_admin sees only their org)
3. Org admin can invite users to their organization (same invitation flow)
4. Organization profile page with: logo, city, country, website, bidang/klaster, social media, tabs for relawan list and activities

### Interview Summary
**Key Discussions**:
- **Bidang/Klaster**: Separate reference table with IDs (database table of valid bidang with many-to-many)
- **Activities Tab**: Uses existing `feeds` table via `Username` field mapping to relawan
- **Donasi Tab**: Deferred to later phase
- **Social Media**: Single JSONB column `social_media`
- **Test Strategy**: TDD approach (write tests first)

**Research Findings**:
- ODK integration exists: `RelawanODKService.CreateAppUserForRelawan()` and `EnsureAppUserForGroupRelawan()` ready but not auto-triggered
- Feed model has `Username` field for submitter identification
- QR code library (`qrcode`) already installed in frontend
- Test infrastructure exists with `stretchr/testify`
- Invitation flow complete with WhatsApp PIN verification

### Gap Analysis
**Identified Gaps** (addressed in plan):
- `RelawanService.Create()` does not auto-create ODK app user
- `ProjectRequestService.ApproveRequest()` does not call `EnsureAppUserForGroupRelawan()`
- `InvitationHandler.InviteUser()` lacks org_admin permission validation
- Organization model missing: city, country, website_url, social_media
- No bidang/klaster reference table or relationship
- QR code modal shows text placeholder, not actual QR image
- OrganizationDetailView lacks Activities tab

---

## Work Objectives

### Core Objective
Implement complete ODK auto-sync, org admin self-service capabilities, and enhanced organization profile with TDD approach.

### Concrete Deliverables
- Database migration: `000016_organization_enhancements.sql`
- New model: `model/bidang.go` with Bidang and OrganizationBidang
- Modified services: `RelawanService`, `ProjectRequestService`, `InvitationService`
- New service: `OrganizationActivityService` (feeds aggregation)
- Enhanced frontend: `OrganizationDetailView.vue`, `RelawanQRCodeModal.vue`
- Updated types: `types.ts` with new Organization fields

### Definition of Done
- [ ] `go test ./...` passes with all new tests
- [ ] `bun run build` completes without errors
- [ ] Manual verification: Create relawan → ODK app user auto-created
- [ ] Manual verification: Approve project → all group relawan get ODK access
- [ ] Manual verification: Org admin can invite user to their org
- [ ] Manual verification: Organization profile shows new fields and activities tab

### Must Have
- Auto-create ODK app user when relawan created with group that has approved project
- Auto-create ODK app users for all group relawan when project approved
- Org admin permission check before allowing invitation
- Bidang reference table with seed data
- Visual QR code image in modal
- Activities tab showing feeds from organization's relawan

### Must NOT Have (Guardrails)
- DO NOT create Donasi tab (deferred)
- DO NOT modify Authentik/OIDC configuration
- DO NOT change existing invitation email flow
- DO NOT add new authentication methods
- AVOID over-engineering bidang selection UI (simple multi-select is fine)
- AVOID breaking existing API contracts (backward compatible)

---

## Verification Strategy (MANDATORY)

### Test Decision
- **Infrastructure exists**: YES (Go tests with testify)
- **User wants tests**: YES (TDD approach)
- **Framework**: Go testing + testify assertions

### TDD Workflow

Each TODO follows RED-GREEN-REFACTOR:

**Task Structure:**
1. **RED**: Write failing test first
   - Test file: `*_test.go`
   - Test command: `go test ./internal/... -v -run TestName`
   - Expected: FAIL (test exists, implementation doesn't)
2. **GREEN**: Implement minimum code to pass
   - Command: `go test ./internal/... -v`
   - Expected: PASS
3. **REFACTOR**: Clean up while keeping green
   - Command: `go test ./internal/... -v`
   - Expected: PASS (still)

### Frontend Verification
- TypeScript compilation: `bun run build`
- Visual verification via Playwright browser automation

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Start Immediately - Foundation):
├── Task 1: Database migration (organization fields + bidang tables)
├── Task 5: Frontend types update
└── Task 12: QR code visual rendering

Wave 2 (After Wave 1 - Backend Models):
├── Task 2: Organization model enhancement
├── Task 3: Bidang model and repository
├── Task 4: Organization handler updates
└── Task 13: Frontend organization forms

Wave 3 (After Wave 2 - Core Services):
├── Task 6: Auto-create ODK on relawan creation
├── Task 7: Auto-create ODK on project approval
├── Task 8: Org admin invitation permission
├── Task 9: Organization activity service (feeds)
└── Task 10: Organization activity handler

Wave 4 (After Wave 3 - Frontend Integration):
├── Task 11: Organization profile page enhancement
├── Task 14: Activities tab component
└── Task 15: End-to-end integration testing

Critical Path: Task 1 → Task 2 → Task 6 → Task 11
Parallel Speedup: ~50% faster than sequential
```

### Dependency Matrix

| Task | Depends On | Blocks | Can Parallelize With |
|------|------------|--------|---------------------|
| 1 | None | 2, 3, 4 | 5, 12 |
| 2 | 1 | 4, 6, 9 | 3, 5, 12, 13 |
| 3 | 1 | 4, 11, 13 | 2, 5, 12 |
| 4 | 2, 3 | 11, 13 | 6, 7, 8 |
| 5 | None | 11, 13, 14 | 1, 12 |
| 6 | 2 | 15 | 7, 8, 9 |
| 7 | 2 | 15 | 6, 8, 9 |
| 8 | None | 15 | 6, 7, 9, 10 |
| 9 | 2 | 10, 14 | 6, 7, 8 |
| 10 | 9 | 14 | 11 |
| 11 | 4, 5 | 14, 15 | 10 |
| 12 | None | 15 | 1, 5 |
| 13 | 3, 4, 5 | 15 | 11, 14 |
| 14 | 5, 10, 11 | 15 | 13 |
| 15 | 6, 7, 8, 11, 12, 13, 14 | None | None (final) |

### Agent Dispatch Summary

| Wave | Tasks | Recommended Dispatch |
|------|-------|---------------------|
| 1 | 1, 5, 12 | 3 parallel agents |
| 2 | 2, 3, 4, 13 | 4 parallel agents |
| 3 | 6, 7, 8, 9, 10 | 5 parallel agents |
| 4 | 11, 14, 15 | 3 sequential (11→14→15) |

---

## TODOs

### Wave 1: Foundation

---

- [ ] 1. Database Migration: Organization Enhancement + Bidang Tables

  **Complexity**: Medium

  **What to do**:
  - Create migration file `infrastructure/database/migrations/000016_organization_enhancements.sql`
  - Add columns to `organizations`: `city`, `country`, `website_url`, `social_media` (JSONB)
  - Create `bidang` reference table: `id`, `name`, `slug`, `description`, `is_active`, `created_at`
  - Create `organization_bidang` junction table: `organization_id`, `bidang_id`, `created_at`
  - Add foreign key constraints
  - Seed initial bidang data: Kesehatan, Pendidikan, Logistik, Pangan, Shelter, WASH
  - Write DOWN migration for rollback

  **Must NOT do**:
  - DO NOT modify existing column types
  - DO NOT add NOT NULL without defaults to existing columns
  - DO NOT drop any existing data

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: SQL migration is a focused, single-file task with clear structure
  - **Skills**: [`git-master`]
    - `git-master`: Atomic commit after migration verified
  - **Skills Evaluated but Omitted**:
    - `frontend-ui-ux`: No frontend work in this task

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 5, 12)
  - **Blocks**: Tasks 2, 3, 4
  - **Blocked By**: None (can start immediately)

  **References**:

  **Pattern References**:
  - `infrastructure/database/migrations/000015_add_whatsapp_pin_verification.sql` - Migration file structure pattern
  - `infrastructure/database/migrations/000009_add_admin_portal.sql` - Example of table creation with foreign keys

  **Type References**:
  - `services/api/internal/model/organization.go:Organization` - Existing organization fields to extend

  **Documentation References**:
  - PostgreSQL JSONB documentation for social_media column type

  **Acceptance Criteria**:

  **TDD (tests for migration):**
  - [ ] Migration file syntax valid: `psql -f 000016_organization_enhancements.sql` runs without error
  - [ ] UP migration adds all new columns: `\d organizations` shows city, country, website_url, social_media
  - [ ] Bidang table created with seed data: `SELECT count(*) FROM bidang` returns 6+
  - [ ] Junction table created: `\d organization_bidang` shows structure
  - [ ] DOWN migration removes changes cleanly

  **Manual Execution Verification:**
  - [ ] Using interactive_bash (tmux session):
    - Command: `docker compose exec postgres psql -U dayawarga -d dayawarga -c "\d organizations"`
    - Expected output contains: `city`, `country`, `website_url`, `social_media`
  - [ ] Command: `docker compose exec postgres psql -U dayawarga -d dayawarga -c "SELECT name FROM bidang"`
    - Expected output contains: `Kesehatan`, `Pendidikan`, `Logistik`

  **Commit**: YES
  - Message: `feat(db): add organization enhancement migration with bidang reference table`
  - Files: `infrastructure/database/migrations/000016_organization_enhancements.sql`
  - Pre-commit: Migration syntax validation

---

- [ ] 5. Frontend Types: Update Organization Interface

  **Complexity**: Low

  **What to do**:
  - Add new fields to `Organization` interface in `types.ts`
  - Add `Bidang` interface
  - Add `OrganizationBidang` interface
  - Update `CreateOrganizationInput` and `UpdateOrganizationInput`
  - Ensure backward compatibility (new fields are optional)

  **Must NOT do**:
  - DO NOT change existing required fields to optional
  - DO NOT remove any existing fields

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: TypeScript interface updates are straightforward additions
  - **Skills**: []
    - No special skills needed for type definitions
  - **Skills Evaluated but Omitted**:
    - `frontend-ui-ux`: No visual work, just types

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 12)
  - **Blocks**: Tasks 11, 13, 14
  - **Blocked By**: None (can start immediately)

  **References**:

  **Pattern References**:
  - `apps/admin-portal/src/services/types.ts:Organization` (lines 55-69) - Current Organization interface to extend
  - `apps/admin-portal/src/services/types.ts:OrganizationMember` (lines 78-85) - Pattern for related interfaces

  **Acceptance Criteria**:

  **TypeScript Compilation:**
  - [ ] `bun run build` completes without type errors
  - [ ] New fields present: `city?: string`, `country?: string`, `website_url?: string`, `social_media?: Record<string, string>`, `bidang?: Bidang[]`

  **Manual Execution Verification:**
  - [ ] Using interactive_bash:
    - Command: `cd /Users/leksa/Development/dayawarga-senyar-2025/apps/admin-portal && bun run build`
    - Expected: Build succeeds with no TypeScript errors

  **Commit**: YES
  - Message: `feat(types): add organization enhancement fields and bidang types`
  - Files: `apps/admin-portal/src/services/types.ts`
  - Pre-commit: `bun run build`

---

- [ ] 12. Frontend: Visual QR Code Rendering in Modal

  **Complexity**: Low

  **What to do**:
  - Import and use `qrcode` library (already installed) in `RelawanQRCodeModal.vue`
  - Generate actual QR image from `qr_code_data` JSON string
  - Replace placeholder with canvas/image QR code
  - Add download as PNG functionality
  - Maintain existing copy-to-clipboard functionality

  **Must NOT do**:
  - DO NOT change the QR data format from backend
  - DO NOT remove existing revoke/create functionality

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: Visual component work with UI rendering
  - **Skills**: [`frontend-ui-ux`]
    - `frontend-ui-ux`: QR code visual rendering and download UX
  - **Skills Evaluated but Omitted**:
    - `playwright`: Not needed for implementation, only for verification

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 5)
  - **Blocks**: Task 15
  - **Blocked By**: None (can start immediately)

  **References**:

  **Pattern References**:
  - `apps/admin-portal/src/components/RelawanQRCodeModal.vue` (lines 1-274) - Existing modal to enhance
  - `apps/admin-portal/package.json` line 24 - QR code library: `"qrcode": "^1.5.4"`

  **API/Type References**:
  - `@types/qrcode` - TypeScript types for qrcode library

  **External References**:
  - npm qrcode docs: https://www.npmjs.com/package/qrcode - Usage patterns

  **Acceptance Criteria**:

  **TypeScript Compilation:**
  - [ ] `bun run build` completes without errors

  **Manual Execution Verification:**
  - [ ] Using Playwright browser automation:
    - Navigate to: `http://localhost:5173/relawan/{id}` (relawan with ODK access)
    - Action: Click "QR Code" button
    - Verify: Modal shows actual scannable QR image (not placeholder)
    - Action: Click "Download" button
    - Verify: PNG file downloads with QR code image
    - Screenshot: Save to `.sisyphus/evidence/task-12-qr-visual.png`

  **Commit**: YES
  - Message: `feat(ui): render visual QR code image in relawan modal`
  - Files: `apps/admin-portal/src/components/RelawanQRCodeModal.vue`
  - Pre-commit: `bun run build`

---

### Wave 2: Backend Models

---

- [ ] 2. Backend Model: Organization Enhancement

  **Complexity**: Medium

  **What to do**:
  - Add new fields to `Organization` struct: `City`, `Country`, `WebsiteURL`, `SocialMedia` (JSONB)
  - Add GORM tags for new columns
  - Add `Bidang` relation (many-to-many via OrganizationBidang)
  - Update `CreateOrganizationInput` struct in service
  - Update `UpdateOrganizationInput` struct in service
  - Write unit tests for model validation

  **Must NOT do**:
  - DO NOT change existing JSON field names (API backward compatibility)
  - DO NOT make new fields required

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Model struct updates with clear patterns
  - **Skills**: []
    - No special skills needed
  - **Skills Evaluated but Omitted**:
    - `git-master`: Can use for commit but not essential

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 3, 5, 12, 13)
  - **Blocks**: Tasks 4, 6, 9
  - **Blocked By**: Task 1 (migration must exist first)

  **References**:

  **Pattern References**:
  - `services/api/internal/model/organization.go` (lines 1-81) - Existing Organization model to extend
  - `services/api/internal/model/feed.go:RawData` (line 26) - JSONB field pattern with `JSONB` type

  **Test References**:
  - `services/api/internal/handler/group_org_scope_test.go` - Test pattern with testify

  **Acceptance Criteria**:

  **TDD:**
  - [ ] Test file created: `services/api/internal/model/organization_test.go`
  - [ ] Test covers: Organization with new fields serializes correctly to JSON
  - [ ] `go test ./internal/model/... -v` → PASS

  **Manual Execution Verification:**
  - [ ] Using interactive_bash:
    - Command: `cd /Users/leksa/Development/dayawarga-senyar-2025/services/api && go build ./...`
    - Expected: Build succeeds with no errors

  **Commit**: YES
  - Message: `feat(model): add organization enhancement fields and bidang relation`
  - Files: `services/api/internal/model/organization.go`, `services/api/internal/model/organization_test.go`
  - Pre-commit: `go test ./internal/model/...`

---

- [ ] 3. Backend Model: Bidang Reference Table

  **Complexity**: Medium

  **What to do**:
  - Create `model/bidang.go` with `Bidang` struct
  - Create `OrganizationBidang` junction struct
  - Create `repository/bidang.go` with CRUD operations
  - Write repository methods: `List()`, `GetByID()`, `GetBySlug()`, `GetByOrganization()`
  - Write unit tests for repository

  **Must NOT do**:
  - DO NOT allow duplicate bidang names
  - DO NOT allow soft-delete (bidang are reference data)

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Standard model + repository creation following existing patterns
  - **Skills**: []
    - No special skills needed
  - **Skills Evaluated but Omitted**:
    - None relevant

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 2, 5, 12)
  - **Blocks**: Tasks 4, 11, 13
  - **Blocked By**: Task 1 (migration must exist)

  **References**:

  **Pattern References**:
  - `services/api/internal/model/organization.go:OrganizationMember` (lines 56-75) - Junction table pattern
  - `services/api/internal/repository/organization.go` - Repository pattern to follow

  **Acceptance Criteria**:

  **TDD:**
  - [ ] Test file created: `services/api/internal/repository/bidang_test.go`
  - [ ] Test covers: List returns all active bidang, GetByOrganization returns correct mapping
  - [ ] `go test ./internal/repository/... -v -run TestBidang` → PASS

  **Manual Execution Verification:**
  - [ ] Using interactive_bash:
    - Command: `cd /Users/leksa/Development/dayawarga-senyar-2025/services/api && go test ./internal/repository/... -v -run TestBidang`
    - Expected: All bidang repository tests pass

  **Commit**: YES
  - Message: `feat(model): add bidang reference table model and repository`
  - Files: `services/api/internal/model/bidang.go`, `services/api/internal/repository/bidang.go`, `services/api/internal/repository/bidang_test.go`
  - Pre-commit: `go test ./internal/repository/...`

---

- [ ] 4. Backend Handler: Organization Updates

  **Complexity**: Medium

  **What to do**:
  - Update `OrganizationHandler` to accept new fields in create/update
  - Add `GET /api/v1/bidang` endpoint to list all bidang
  - Add `POST /api/v1/organizations/:id/bidang` to assign bidang
  - Add `DELETE /api/v1/organizations/:id/bidang/:bidang_id` to remove bidang
  - Update `CreateOrganizationInput` validation
  - Write handler tests

  **Must NOT do**:
  - DO NOT change existing endpoint paths
  - DO NOT break existing request/response formats

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Handler updates following existing patterns
  - **Skills**: []
    - No special skills needed
  - **Skills Evaluated but Omitted**:
    - None relevant

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (after Tasks 2, 3)
  - **Blocks**: Tasks 11, 13
  - **Blocked By**: Tasks 2, 3

  **References**:

  **Pattern References**:
  - `services/api/internal/handler/organization.go` - Existing organization handlers to extend
  - `services/api/internal/handler/relawan.go:Create` (lines 132-184) - Input validation pattern

  **API/Type References**:
  - `services/api/internal/service/organization.go:CreateOrganizationInput` - Input struct to extend

  **Test References**:
  - `services/api/internal/handler/group_org_scope_test.go` - Handler test patterns

  **Acceptance Criteria**:

  **TDD:**
  - [ ] Test file created: `services/api/internal/handler/organization_bidang_test.go`
  - [ ] Test covers: List bidang returns 6 items, Assign bidang to org works, Permission checks enforced
  - [ ] `go test ./internal/handler/... -v -run TestOrganization` → PASS

  **Manual Execution Verification:**
  - [ ] Using interactive_bash:
    - Command: `curl http://localhost:8080/api/v1/bidang`
    - Expected: JSON array with bidang items (Kesehatan, Pendidikan, etc.)

  **Commit**: YES
  - Message: `feat(api): add bidang endpoints and organization enhancement handlers`
  - Files: `services/api/internal/handler/organization.go`, `services/api/internal/handler/bidang.go`, `services/api/internal/handler/organization_bidang_test.go`
  - Pre-commit: `go test ./internal/handler/...`

---

- [ ] 13. Frontend: Organization Create/Edit Forms

  **Complexity**: Medium

  **What to do**:
  - Update organization create form with new fields (city, country, website_url)
  - Add social media inputs (instagram, facebook, twitter as JSONB)
  - Add bidang multi-select component (fetch from `/api/v1/bidang`)
  - Update organization edit form with same fields
  - Handle bidang assignment via API

  **Must NOT do**:
  - DO NOT change existing form field names
  - DO NOT make new fields required in UI

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: Form UI work with multiple components
  - **Skills**: [`frontend-ui-ux`]
    - `frontend-ui-ux`: Form layout and multi-select UX
  - **Skills Evaluated but Omitted**:
    - `playwright`: For verification only, not implementation

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 2, 3, 4)
  - **Blocks**: Task 15
  - **Blocked By**: Tasks 3, 4, 5

  **References**:

  **Pattern References**:
  - `apps/admin-portal/src/views/OrganizationsView.vue` - Existing organization forms
  - `apps/admin-portal/src/components/ui/` - shadcn-vue components for form elements

  **API/Type References**:
  - `apps/admin-portal/src/services/types.ts:CreateOrganizationInput` - Input type to extend

  **Acceptance Criteria**:

  **TypeScript Compilation:**
  - [ ] `bun run build` completes without errors

  **Manual Execution Verification:**
  - [ ] Using Playwright browser automation:
    - Navigate to: `http://localhost:5173/organizations`
    - Action: Click "Tambah Organisasi"
    - Verify: Form shows fields for city, country, website_url, social media, bidang multi-select
    - Action: Fill all fields including selecting 2 bidang
    - Action: Submit form
    - Verify: Organization created with all fields saved
    - Screenshot: Save to `.sisyphus/evidence/task-13-org-form.png`

  **Commit**: YES
  - Message: `feat(ui): add organization enhancement fields to create/edit forms`
  - Files: `apps/admin-portal/src/views/OrganizationsView.vue`, `apps/admin-portal/src/services/organizations.ts`
  - Pre-commit: `bun run build`

---

### Wave 3: Core Services

---

- [ ] 6. Backend Service: Auto-Create ODK App User on Relawan Creation

  **Complexity**: High

  **What to do**:
  - Modify `RelawanService.Create()` to check if group has approved ODK project
  - If yes, call `RelawanODKService.CreateAppUserForRelawan()` automatically
  - Handle case where relawan is created without group (skip ODK creation)
  - Handle case where group exists but has no approved project (skip ODK creation)
  - Add error handling - ODK failure should not fail relawan creation (log warning)
  - Write comprehensive unit tests

  **Must NOT do**:
  - DO NOT fail relawan creation if ODK creation fails (graceful degradation)
  - DO NOT create ODK user if group has no approved project
  - DO NOT change the relawan creation API response format

  **Recommended Agent Profile**:
  - **Category**: `ultrabrain`
    - Reason: Complex service logic with error handling and edge cases
  - **Skills**: []
    - No special skills needed, but requires careful logic
  - **Skills Evaluated but Omitted**:
    - None relevant

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 7, 8, 9)
  - **Blocks**: Task 15
  - **Blocked By**: Task 2

  **References**:

  **Pattern References**:
  - `services/api/internal/service/relawan.go:Create` (lines 74-133) - Current Create method to modify
  - `services/api/internal/service/relawan_odk.go:CreateAppUserForRelawan` (lines 42-93) - ODK creation method to call
  - `services/api/internal/model/group.go:HasODKProject()` - Check if group has approved project

  **API/Type References**:
  - `services/api/internal/repository/group.go:FindByID` - Get group with ODK info

  **WHY Each Reference Matters**:
  - `relawan.go:Create` - This is the method we're modifying, need to understand current flow
  - `relawan_odk.go:CreateAppUserForRelawan` - This is the method we'll call, need to understand its requirements
  - `group.go:HasODKProject()` - Gate condition for auto-creation

  **Acceptance Criteria**:

  **TDD:**
  - [ ] Test file updated: `services/api/internal/service/relawan_test.go`
  - [ ] Test covers: Create relawan with group+project → ODK user created
  - [ ] Test covers: Create relawan with group but no project → no ODK user
  - [ ] Test covers: Create relawan without group → no ODK user
  - [ ] Test covers: ODK creation fails → relawan still created, warning logged
  - [ ] `go test ./internal/service/... -v -run TestRelawanCreate` → PASS

  **Manual Execution Verification:**
  - [ ] Using interactive_bash:
    - Prerequisite: Have a group with approved ODK project
    - Command: `curl -X POST http://localhost:8080/api/v1/relawan -H "Authorization: Bearer $TOKEN" -d '{"organization_id":"...","group_id":"...","name":"Test Auto ODK"}'`
    - Verify response includes: `odk_app_user_id` is NOT null
    - Command: `curl http://localhost:8080/api/v1/relawan/{id}/odk-qr-code -H "Authorization: Bearer $TOKEN"`
    - Verify: QR code data returned (not error "relawan does not have ODK access")

  **Commit**: YES
  - Message: `feat(service): auto-create ODK app user on relawan creation`
  - Files: `services/api/internal/service/relawan.go`, `services/api/internal/service/relawan_test.go`
  - Pre-commit: `go test ./internal/service/...`

---

- [ ] 7. Backend Service: Auto-Create ODK App Users on Project Approval

  **Complexity**: High

  **What to do**:
  - Modify `ProjectRequestService.ApproveRequest()` to call `EnsureAppUserForGroupRelawan()`
  - Call after successful transaction (Project Manager created, group updated)
  - Handle partial failures gracefully (some relawan may fail, continue with others)
  - Return count of created app users in response
  - Write comprehensive unit tests

  **Must NOT do**:
  - DO NOT fail entire approval if some relawan ODK creation fails
  - DO NOT rollback project approval on ODK user creation failure
  - DO NOT create duplicate ODK users for relawan who already have access

  **Recommended Agent Profile**:
  - **Category**: `ultrabrain`
    - Reason: Transaction handling with partial failure recovery
  - **Skills**: []
    - No special skills needed, but requires careful transaction management
  - **Skills Evaluated but Omitted**:
    - None relevant

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 6, 8, 9)
  - **Blocks**: Task 15
  - **Blocked By**: Task 2

  **References**:

  **Pattern References**:
  - `services/api/internal/service/project_request.go:ApproveRequest` (lines 130-223) - Current approval flow to modify
  - `services/api/internal/service/relawan_odk.go:EnsureAppUserForGroupRelawan` (lines 189-213) - Method to call after approval

  **WHY Each Reference Matters**:
  - `project_request.go:ApproveRequest` - This is the method we're modifying, need to understand the transaction boundary
  - `relawan_odk.go:EnsureAppUserForGroupRelawan` - Bulk creation method that handles existing users gracefully

  **Acceptance Criteria**:

  **TDD:**
  - [ ] Test file created: `services/api/internal/service/project_request_test.go`
  - [ ] Test covers: Approve request → all group relawan get ODK access
  - [ ] Test covers: Relawan who already have access are skipped (no error)
  - [ ] Test covers: Partial ODK failure → approval succeeds, count reflects actual created
  - [ ] `go test ./internal/service/... -v -run TestProjectRequestApprove` → PASS

  **Manual Execution Verification:**
  - [ ] Using interactive_bash:
    - Prerequisite: Have a group with 3 relawan, pending project request
    - Command: `curl -X POST http://localhost:8080/api/v1/project-requests/{id}/approve -H "Authorization: Bearer $TOKEN"`
    - Verify response includes: Success
    - Command: `curl http://localhost:8080/api/v1/groups/{id}/relawan -H "Authorization: Bearer $TOKEN"`
    - Verify: All 3 relawan have `odk_app_user_id` NOT null

  **Commit**: YES
  - Message: `feat(service): auto-create ODK app users for group relawan on project approval`
  - Files: `services/api/internal/service/project_request.go`, `services/api/internal/service/project_request_test.go`
  - Pre-commit: `go test ./internal/service/...`

---

- [ ] 8. Backend Handler: Org Admin Invitation Permission Check

  **Complexity**: Medium

  **What to do**:
  - Modify `InvitationHandler.InviteUser()` to validate org_admin permissions
  - If inviter is org_admin (not super_admin), enforce:
    - Must provide organization_id
    - organization_id must be one of inviter's organizations
    - Invited user role must be "member" (not "admin")
  - Super_admin bypasses these restrictions
  - Write unit tests for permission logic

  **Must NOT do**:
  - DO NOT change super_admin behavior
  - DO NOT modify the invitation flow (email, token, PIN)
  - DO NOT allow org_admin to invite another org_admin

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Permission check addition is focused scope
  - **Skills**: []
    - No special skills needed
  - **Skills Evaluated but Omitted**:
    - None relevant

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 6, 7, 9)
  - **Blocks**: Task 15
  - **Blocked By**: None (independent of model changes)

  **References**:

  **Pattern References**:
  - `services/api/internal/handler/invitation.go:InviteUser` (lines 121-183) - Handler to modify
  - `services/api/internal/auth/context.go` - Auth context helpers (GetUser, CanManageOrganization)
  - `services/api/internal/handler/relawan.go:Create` (lines 154-161) - Permission check pattern

  **WHY Each Reference Matters**:
  - `invitation.go:InviteUser` - This is the handler we're modifying
  - `auth/context.go` - Pattern for checking user permissions
  - `relawan.go:Create` - Example of org permission check in handlers

  **Acceptance Criteria**:

  **TDD:**
  - [ ] Test file created: `services/api/internal/handler/invitation_test.go`
  - [ ] Test covers: super_admin can invite to any org with any role
  - [ ] Test covers: org_admin can invite member to their org
  - [ ] Test covers: org_admin CANNOT invite admin to their org
  - [ ] Test covers: org_admin CANNOT invite to org they don't belong to
  - [ ] Test covers: org_admin CANNOT invite without organization_id
  - [ ] `go test ./internal/handler/... -v -run TestInviteUser` → PASS

  **Manual Execution Verification:**
  - [ ] Using interactive_bash:
    - Login as org_admin (not super_admin)
    - Command: `curl -X POST http://localhost:8080/api/v1/invitations -H "Authorization: Bearer $ORG_ADMIN_TOKEN" -d '{"email":"test@example.com","name":"Test","organization_id":"their-org-id","org_role":"member"}'`
    - Verify: Success, invitation created
    - Command: `curl -X POST http://localhost:8080/api/v1/invitations -H "Authorization: Bearer $ORG_ADMIN_TOKEN" -d '{"email":"test2@example.com","name":"Test2","organization_id":"other-org-id","org_role":"member"}'`
    - Verify: Error "Access denied to this organization"

  **Commit**: YES
  - Message: `feat(auth): add org_admin permission check for user invitation`
  - Files: `services/api/internal/handler/invitation.go`, `services/api/internal/handler/invitation_test.go`
  - Pre-commit: `go test ./internal/handler/...`

---

- [ ] 9. Backend Service: Organization Activity Service (Feeds Aggregation)

  **Complexity**: Medium

  **What to do**:
  - Create `service/organization_activity.go`
  - Implement `GetOrganizationActivities(orgID, pagination)` method
  - Query feeds where `username` matches any relawan name in the organization
  - Join with relawan table to get submitter info
  - Return paginated list with feed content, photos, timestamps
  - Write unit tests

  **Must NOT do**:
  - DO NOT modify the feeds table structure
  - DO NOT duplicate feed data
  - DO NOT return feeds from other organizations' relawan

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Query construction following existing patterns
  - **Skills**: []
    - No special skills needed
  - **Skills Evaluated but Omitted**:
    - None relevant

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 6, 7, 8)
  - **Blocks**: Tasks 10, 14
  - **Blocked By**: Task 2 (needs organization model for relation)

  **References**:

  **Pattern References**:
  - `services/api/internal/service/feed.go` - If exists, feed service patterns
  - `services/api/internal/repository/relawan.go:GetByOrganization` - Get relawan by org pattern
  - `services/api/internal/model/feed.go:Feed` (lines 9-38) - Feed model with Username field

  **API/Type References**:
  - Feed.Username field (line 19) - This is the field that maps to relawan name

  **WHY Each Reference Matters**:
  - Feed model shows `Username` field is what we'll join against relawan names
  - GetByOrganization shows how to get all relawan for an org

  **Acceptance Criteria**:

  **TDD:**
  - [ ] Test file created: `services/api/internal/service/organization_activity_test.go`
  - [ ] Test covers: Get activities returns feeds from org's relawan only
  - [ ] Test covers: Pagination works correctly
  - [ ] Test covers: Empty result when org has no relawan with feeds
  - [ ] `go test ./internal/service/... -v -run TestOrganizationActivity` → PASS

  **Manual Execution Verification:**
  - [ ] Using interactive_bash:
    - Prerequisite: Have organization with relawan who have submitted feeds
    - Command: Check database for feeds with matching usernames
    - `docker compose exec postgres psql -U dayawarga -d dayawarga -c "SELECT f.content, f.username FROM information_feeds f LIMIT 5"`

  **Commit**: YES
  - Message: `feat(service): add organization activity service for feeds aggregation`
  - Files: `services/api/internal/service/organization_activity.go`, `services/api/internal/service/organization_activity_test.go`
  - Pre-commit: `go test ./internal/service/...`

---

- [ ] 10. Backend Handler: Organization Activities Endpoint

  **Complexity**: Low

  **What to do**:
  - Create `GET /api/v1/organizations/:id/activities` endpoint
  - Use `OrganizationActivityService.GetOrganizationActivities()`
  - Support pagination query params (page, page_size)
  - Check organization access permissions
  - Write handler tests

  **Must NOT do**:
  - DO NOT expose sensitive feed data
  - DO NOT allow access to other organizations' activities

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Standard handler creation
  - **Skills**: []
    - No special skills needed
  - **Skills Evaluated but Omitted**:
    - None relevant

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (after Task 9)
  - **Blocks**: Task 14
  - **Blocked By**: Task 9

  **References**:

  **Pattern References**:
  - `services/api/internal/handler/organization.go` - Organization handlers pattern
  - `services/api/internal/handler/relawan.go:List` (lines 24-101) - Pagination pattern

  **Acceptance Criteria**:

  **TDD:**
  - [ ] Test file updated: `services/api/internal/handler/organization_test.go`
  - [ ] Test covers: Get activities returns feeds for org
  - [ ] Test covers: Pagination params work
  - [ ] Test covers: Access denied for non-member
  - [ ] `go test ./internal/handler/... -v -run TestOrganizationActivities` → PASS

  **Manual Execution Verification:**
  - [ ] Using interactive_bash:
    - Command: `curl http://localhost:8080/api/v1/organizations/{id}/activities?page=1&page_size=10 -H "Authorization: Bearer $TOKEN"`
    - Verify: JSON response with `data` array of activities, `total`, `page`, `page_size`

  **Commit**: YES
  - Message: `feat(api): add organization activities endpoint`
  - Files: `services/api/internal/handler/organization.go`, `services/api/internal/handler/organization_test.go`
  - Pre-commit: `go test ./internal/handler/...`

---

### Wave 4: Frontend Integration

---

- [ ] 11. Frontend: Organization Profile Page Enhancement

  **Complexity**: High

  **What to do**:
  - Enhance `OrganizationDetailView.vue` with profile card showing all fields
  - Display: logo, name, description, city, country, website, social media links, bidang badges
  - Make social media links clickable
  - Display bidang as colored badges
  - Maintain existing tabs structure
  - Add loading states for all new data

  **Must NOT do**:
  - DO NOT remove existing tabs (Tim/Grup, Relawan)
  - DO NOT break mobile responsiveness

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: Complex UI layout with multiple components
  - **Skills**: [`frontend-ui-ux`]
    - `frontend-ui-ux`: Profile layout, badge styling, responsive design
  - **Skills Evaluated but Omitted**:
    - `playwright`: For verification only

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4 (with Task 10)
  - **Blocks**: Tasks 14, 15
  - **Blocked By**: Tasks 4, 5

  **References**:

  **Pattern References**:
  - `apps/admin-portal/src/views/OrganizationDetailView.vue` (lines 1-277) - Current view to enhance
  - `apps/admin-portal/src/components/ui/badge/Badge.vue` - Badge component for bidang

  **Acceptance Criteria**:

  **TypeScript Compilation:**
  - [ ] `bun run build` completes without errors

  **Manual Execution Verification:**
  - [ ] Using Playwright browser automation:
    - Navigate to: `http://localhost:5173/organizations/{id}`
    - Verify: Profile card shows logo, name, city, country, website
    - Verify: Social media icons are displayed and clickable
    - Verify: Bidang displayed as badges
    - Verify: Existing tabs (Tim/Grup, Relawan) still work
    - Screenshot: Save to `.sisyphus/evidence/task-11-org-profile.png`

  **Commit**: YES
  - Message: `feat(ui): enhance organization profile page with new fields display`
  - Files: `apps/admin-portal/src/views/OrganizationDetailView.vue`
  - Pre-commit: `bun run build`

---

- [ ] 14. Frontend: Activities Tab Component

  **Complexity**: Medium

  **What to do**:
  - Add "Aktivitas" tab to OrganizationDetailView
  - Create activity feed component showing feeds from organization's relawan
  - Display: content, username (relawan name), timestamp, photos if any
  - Implement pagination (load more button)
  - Handle empty state gracefully

  **Must NOT do**:
  - DO NOT remove existing tabs
  - DO NOT load all activities at once (use pagination)

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: List UI with pagination and photo handling
  - **Skills**: [`frontend-ui-ux`]
    - `frontend-ui-ux`: Activity feed design, pagination UX
  - **Skills Evaluated but Omitted**:
    - `playwright`: For verification only

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Sequential after Task 11
  - **Blocks**: Task 15
  - **Blocked By**: Tasks 5, 10, 11

  **References**:

  **Pattern References**:
  - `apps/admin-portal/src/views/OrganizationDetailView.vue:TabsContent` (lines 168-273) - Tab content pattern
  - `apps/admin-portal/src/services/types.ts` - Will need Feed type

  **API/Type References**:
  - New endpoint: `GET /api/v1/organizations/:id/activities`

  **Acceptance Criteria**:

  **TypeScript Compilation:**
  - [ ] `bun run build` completes without errors

  **Manual Execution Verification:**
  - [ ] Using Playwright browser automation:
    - Navigate to: `http://localhost:5173/organizations/{id}`
    - Action: Click "Aktivitas" tab
    - Verify: Activity feed shows with content, submitter name, timestamp
    - Verify: Photos displayed if present
    - Action: Scroll/click "Load more" if available
    - Verify: More activities load
    - Screenshot: Save to `.sisyphus/evidence/task-14-activities-tab.png`

  **Commit**: YES
  - Message: `feat(ui): add activities tab to organization profile`
  - Files: `apps/admin-portal/src/views/OrganizationDetailView.vue`, `apps/admin-portal/src/services/types.ts`
  - Pre-commit: `bun run build`

---

- [ ] 15. End-to-End Integration Testing

  **Complexity**: High

  **What to do**:
  - Run full integration test of all features
  - Test 1: Create relawan with auto ODK user creation
  - Test 2: Approve project request with bulk ODK user creation
  - Test 3: Org admin invites user to their organization
  - Test 4: View organization profile with all new fields
  - Test 5: View activities tab with feeds
  - Test 6: Scan QR code visual in modal
  - Document any issues found

  **Must NOT do**:
  - DO NOT skip any test scenario
  - DO NOT proceed if critical tests fail

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: End-to-end testing across frontend and backend
  - **Skills**: [`playwright`, `frontend-ui-ux`]
    - `playwright`: Browser automation for e2e testing
    - `frontend-ui-ux`: Verify UI elements correct
  - **Skills Evaluated but Omitted**:
    - `git-master`: Final commit after all tests pass

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Sequential (final task)
  - **Blocks**: None (final)
  - **Blocked By**: Tasks 6, 7, 8, 11, 12, 13, 14

  **References**:

  All previous tasks provide the features to test.

  **Acceptance Criteria**:

  **All Integration Tests Pass:**
  - [ ] Test 1 PASS: Create relawan → odk_app_user_id populated
  - [ ] Test 2 PASS: Approve project → all group relawan have odk_app_user_id
  - [ ] Test 3 PASS: Org admin invite → user receives invitation, can set password, gets PIN
  - [ ] Test 4 PASS: Org profile → shows city, country, website, bidang, social media
  - [ ] Test 5 PASS: Activities tab → shows feeds from organization's relawan
  - [ ] Test 6 PASS: QR modal → shows scannable QR image, download works

  **Manual Execution Verification:**
  - [ ] Using Playwright browser automation for frontend tests
  - [ ] Using curl/httpie for API tests
  - [ ] Screenshots saved for each visual verification
  - [ ] All evidence in `.sisyphus/evidence/` directory

  **Commit**: YES
  - Message: `test(e2e): verify all admin portal enhancement features`
  - Files: `.sisyphus/evidence/*`
  - Pre-commit: All individual task tests pass

---

## Commit Strategy

| After Task | Message | Files | Verification |
|------------|---------|-------|--------------|
| 1 | `feat(db): add organization enhancement migration with bidang reference table` | migrations/000016_*.sql | psql migration runs |
| 2 | `feat(model): add organization enhancement fields and bidang relation` | model/organization.go, *_test.go | go test |
| 3 | `feat(model): add bidang reference table model and repository` | model/bidang.go, repository/bidang.go, *_test.go | go test |
| 4 | `feat(api): add bidang endpoints and organization enhancement handlers` | handler/organization.go, handler/bidang.go, *_test.go | go test |
| 5 | `feat(types): add organization enhancement fields and bidang types` | services/types.ts | bun run build |
| 6 | `feat(service): auto-create ODK app user on relawan creation` | service/relawan.go, *_test.go | go test |
| 7 | `feat(service): auto-create ODK app users for group relawan on project approval` | service/project_request.go, *_test.go | go test |
| 8 | `feat(auth): add org_admin permission check for user invitation` | handler/invitation.go, *_test.go | go test |
| 9 | `feat(service): add organization activity service for feeds aggregation` | service/organization_activity.go, *_test.go | go test |
| 10 | `feat(api): add organization activities endpoint` | handler/organization.go, *_test.go | go test |
| 11 | `feat(ui): enhance organization profile page with new fields display` | views/OrganizationDetailView.vue | bun run build |
| 12 | `feat(ui): render visual QR code image in relawan modal` | components/RelawanQRCodeModal.vue | bun run build |
| 13 | `feat(ui): add organization enhancement fields to create/edit forms` | views/OrganizationsView.vue | bun run build |
| 14 | `feat(ui): add activities tab to organization profile` | views/OrganizationDetailView.vue | bun run build |
| 15 | `test(e2e): verify all admin portal enhancement features` | .sisyphus/evidence/* | all tests pass |

---

## Success Criteria

### Verification Commands
```bash
# Backend tests
cd /Users/leksa/Development/dayawarga-senyar-2025/services/api && go test ./... -v

# Frontend build
cd /Users/leksa/Development/dayawarga-senyar-2025/apps/admin-portal && bun run build

# Migration check
docker compose exec postgres psql -U dayawarga -d dayawarga -c "\d organizations"
# Expected: city, country, website_url, social_media columns exist

# Bidang seed data
docker compose exec postgres psql -U dayawarga -d dayawarga -c "SELECT name FROM bidang"
# Expected: Kesehatan, Pendidikan, Logistik, Pangan, Shelter, WASH
```

### Final Checklist
- [ ] All "Must Have" features implemented and tested
- [ ] All "Must NOT Have" guardrails respected
- [ ] All backend tests pass (`go test ./...`)
- [ ] Frontend builds without errors (`bun run build`)
- [ ] E2E integration tests pass
- [ ] No breaking changes to existing API endpoints
