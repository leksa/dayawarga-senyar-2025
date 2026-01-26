# claude.md
# Project Constitution, Operating Rules & Long-Context Memory

Claude must read README.md and CHANGELOG.md before proposing any changes.

────────────────────────────────────────────────────────────

## 1. Project Overview

Project Name:
Admin Portal – Authentik + ODK Central Integration

Purpose:
Provide a secure admin portal for organizations to manage relawan (field volunteers) using a unified identity system and controlled ODK App User lifecycle.
ODK Central is ready. Services Production with API services and PostgreSQL database is ready.

Primary goals:
- Single identity via Authentik (OIDC)
- Centralized relawan management
- Programmatic ODK App User creation
- Secure QR-based access for ODK Collect

Non-goals:
- Replacing ODK Central UI
- Rebuilding ODK Collect
- Non-OIDC authentication
- Advanced analytics or billing

Quality bar:
- Production-grade
- Explicit flows over hidden behavior
- Security-first
- Real-world disaster-response readiness

────────────────────────────────────────────────────────────

## 2. Claude Behavior Rules (Hard Rules)

- Do NOT invent requirements or UX decisions
- Do NOT redesign existing features without approval
- Ask before changing architecture or auth flows
- Prefer simple, boring, well-known solutions
- Explain trade-offs when suggesting alternatives
- Never refactor unrelated code
- Respect all logged decisions
- If context is missing or ambiguous, ask
- Assume system is used in high-risk, real operations

────────────────────────────────────────────────────────────

## 3. Authoritative Project Memory

The following files are the source of truth and must always be respected:

- README.md  
  → Describes what already exists and how it works

- CHANGELOG.md  
  → Records completed work and decisions already implemented

Rules:
- Anything in README.md is considered DONE unless stated otherwise
- Anything in CHANGELOG.md must never be re-implemented or undone
- If chat context conflicts with these files, the files win
- New work must extend existing behavior, not replace it

────────────────────────────────────────────────────────────

## 4. System Architecture (Mental Model)

System actors:
1. Admin Portal Frontend (Vue.js)
2. Authentik (OIDC Identity Provider)
3. Backend API (Go)
4. ODK Central (OIDC Client + App User API)

Principles:
- Authentik is the single source of identity
- Backend API is the single source of business logic
- ODK Central is the single source of data collection
- Frontend NEVER talks directly to ODK Central

No component bypasses another’s responsibility.

────────────────────────────────────────────────────────────

## 5. Authentication & Identity Model

- Authentik is the only IdP
- Admin Portal and ODK Central authenticate via OIDC
- Same human identity can log into both systems

Token usage:
- access_token → API authorization
- id_token → identity (email, name, groups)
- refresh_token → session renewal

Identity rules:
- Human users authenticate via Authentik
- ODK App Users are NOT human identities
- ODK App Users do NOT use Authentik
- App User lifecycle is owned by backend only

────────────────────────────────────────────────────────────

## 6. ODK App User Model

- Created via ODK Central API
- Used only by ODK Collect
- Auth via token + QR code
- Limited to assigned forms

Rules:
- App Users must be traceable to relawan records
- Tokens must be generated and stored securely
- Revocation and reassignment must be explicit

────────────────────────────────────────────────────────────

## 7. Canonical Flows

### Unified Login Flow
1. User clicks Login
2. Redirect to Authentik
3. User authenticates
4. Backend exchanges code for tokens
5. Backend validates tokens
6. User auto-provisioned if first login

### Relawan → ODK App User Flow
1. Org Admin adds relawan
2. Backend creates ODK App User
3. Backend assigns forms
4. Backend generates QR code
5. Frontend displays QR
6. Relawan scans in ODK Collect

────────────────────────────────────────────────────────────

## 8. Technical Stack & Constraints

Backend:
- Go
- REST API
- PostgreSQL
- JWT validation

Frontend:
- Vue 3 + Vite
- npm
- Pinia
- Axios
- Tailwind + shadcn-vue
- Language: Indonesian
- Layout: Desktop-first
- Location: apps/admin-portal/

Infrastructure:
- Authentik (Docker)
- ODK Central (existing)
- No managed auth services
- No vendor lock-in assumptions

────────────────────────────────────────────────────────────

## 9. Local Development & CLI Rules

Claude may:
- Provide shell commands
- Explain Docker and server workflows
- Propose scripts or Makefiles
- Suggest `.env.example`
- Explain how to generate credentials

Claude must NOT:
- Assume real passwords or tokens
- Hardcode secrets in docs or code
- Reuse example secrets
- Output sensitive values unless provided

────────────────────────────────────────────────────────────

## 10. Secrets & Environment Handling

- All secrets use environment variables
- Never commit secrets
- Use placeholders only
- If you dont find login or secrets, ask me. Don't guess it

Example:
AUTHENTIK_CLIENT_ID=your_client_id  
AUTHENTIK_CLIENT_SECRET=your_client_secret  

Environments:
- local
- docker
- staging
- production

Secrets differ per environment.

### Production Servers

See private infrastructure documentation for server IPs and credentials.

| Server | Purpose |
|--------|---------|
| Main Platform | Frontend, API, Admin Portal, Authentik, Ghost |
| ODK Central | Data collection forms |
| WhatsApp Chatbot | Relawan chatbot (Node.js + PM2) |

Login: Public key attached to all servers

DO NOT ACCESS SERVER PRODUCTION WITHOUT MY PERMISSION

### DNS Records (dayawarga.com)

| Subdomain | Service |
|-----------|---------|
| dayawarga.com | Main Frontend |
| www.dayawarga.com | Main Frontend |
| api.dayawarga.com | Backend API |
| data.dayawarga.com | ODK Central |
| stories.dayawarga.com | Ghost CMS |
| admin.dayawarga.com | Admin Portal |
| auth.dayawarga.com | Authentik SSO |

────────────────────────────────────────────────────────────

## 11. Docker & Server Assumptions

- Docker or native tooling for local dev
- Docker Compose preferred locally
- Production runs in controlled servers
- Commands must be environment-aware
- Do not assume root access

────────────────────────────────────────────────────────────

## 12. Implementation Phases (Authoritative)

Phase 0 – Infrastructure
- Frontend scaffold
- Authentik setup
- OIDC providers (Admin Portal, ODK Central)
- ODK Central OIDC config


Phase 1 – Auth & Core
Backend:
- DB migrations
- OIDC middleware
- User auto-provisioning
- RBAC
- /auth/me

Frontend:
- Login + callback
- Token handling
- Protected routes
- Base layout

Phase 2 – Organization & Relawan
Backend:
- Org CRUD
- ODK API wrapper
- Relawan CRUD
- App User creation
- QR generation

Frontend:
- Org pages
- Relawan UI
- QR display

Phase 3 – Dashboard
- Org dashboard
- Activity feed
- Submission history
- Public profile

────────────────────────────────────────────────────────────

## 13. Security & Risk Awareness

Known risks:
- App User token leakage
- Role misconfiguration
- Backend as single failure point

Rules:
- Warn before destructive commands
- Explain rollback for migrations
- Ask before production-impacting actions

────────────────────────────────────────────────────────────

## 14. Coding Style & Quality Standards

Code must prioritize:
- Readability over cleverness
- Explicit behavior over magic
- Maintainability over short-term speed

All code should be understandable by a new engineer in <30 minutes.

────────────────────────────────────────────────────────────

### Go (Backend) Style Rules

General:
- Follow standard Go conventions (gofmt is mandatory)
- Small files, single responsibility
- Prefer explicit error handling
- Avoid global state
- No hidden side effects

Structure:
- Handlers → Services → Repositories
- Business logic must NOT live in handlers
- Database access must NOT live in handlers
- ODK Central API access via dedicated client layer

Naming:
- Functions: verbs (CreateRelawan, AssignForm)
- Structs: nouns (Relawan, Organization)
- Interfaces: behavior-based (RelawanRepository)

Errors:
- Always return errors explicitly
- Wrap errors with context
- No silent failures
- No panic in request flow

Example:
return fmt.Errorf("create app user failed: %w", err)

Auth & Context:
- Request context must be passed down
- Auth data extracted once per request
- No re-parsing tokens in inner layers

Logging:
- Log at boundaries (API, external calls)
- Do NOT log secrets or tokens
- Prefer structured logs

Testing:
- Unit-test business logic
- Mock external services (ODK, Authentik)
- Avoid over-mocking internals

### Vue 3 (Frontend) Style Rules

General:
- Composition API only
- No Options API
- Explicit state over implicit reactivity
- Avoid clever watchers

Structure:
- Pages: route-level components
- Components: reusable UI
- Composables: business logic
- Stores (Pinia): global state only

State Management:
- Pinia for shared state
- Local state stays in components
- No API calls inside components
- API calls live in service layer

API Handling:
- Axios wrapper per backend service
- No raw Axios usage in components
- Centralized error handling

Naming:
- Components: PascalCase
- Composables: useXxx()
- Stores: useXxxStore()

UI Rules:
- No hardcoded colors (use theme)
- Desktop-first layout
- Accessibility matters (labels, focus)

Language:
- UI text in Indonesian
- No hardcoded strings in logic

────────────────────────────────────────────────────────────

## 15. Folder & File Organization

Backend (example):
- internal/handlers
- internal/services
- internal/repositories
- internal/odk
- internal/auth
- internal/middleware

Frontend:
apps/admin-portal/
├── src/
│   ├── pages/
│   ├── components/
│   ├── composables/
│   ├── stores/
│   ├── services/
│   └── router/

Rules:
- No circular dependencies
- One concern per folder
- No dumping ground folders (utils/)

────────────────────────────────────────────────────────────

## 16. Forbidden Patterns

- Business logic in HTTP handlers
- API calls inside Vue components
- Magic environment variables
- Silent error swallowing
- Over-abstracted interfaces
- “Just works” undocumented behavior

────────────────────────────────────────────────────────────

## 17. Code Review Expectations

Before suggesting code:
- Explain where the code lives
- Explain why this structure was chosen
- Explain how it integrates with existing code
- Mention risks or trade-offs

If code increases complexity, justify it.

────────────────────────────────────────────────────────────

## 18. Decisions Log (Append Only)

### 2026-01-15
- Read README.mD and CHANGELOG.md for what have been done

### 2026-01-15
- Authentik as single IdP
- ODK Central uses Authentik via OIDC
- App Users managed by backend only
- Frontend never talks to ODK Central

### 2026-01-15 - Phase 0 Completed
- Admin Portal frontend: Vue 3 + Vite + TypeScript + pnpm
- UI Library: shadcn-vue (104 components)
- Design: Bold theme, dark mode default, Plus Jakarta Sans + JetBrains Mono
- Color: Deep Teal primary (OKLCH format)
- Auth: OIDC via oidc-client-ts, Pinia auth store
- Views: Dashboard, Organizations, Groups, Relawan (list + detail)
- Authentik: Docker setup at infrastructure/authentik/
- OIDC Provider: Public client type for SPA
- Login flow: Tested and working

### 2026-01-15 - Phase 1 Completed (Backend OIDC Auth)
- Database migration: 000009_add_admin_portal.sql (users, organizations, organization_members, groups, relawan tables)
- User model with OIDC fields: internal/model/user.go
- OIDC JWT validator with JWKS caching: internal/auth/oidc.go
- Auth middleware: internal/auth/middleware.go
- RBAC middleware: internal/auth/rbac.go
- User repository: internal/repository/user.go
- User service with auto-provisioning: internal/service/user.go
- Auth handler (/auth/me): internal/handler/auth.go
- Config updated with OIDC_ISSUER_URL, OIDC_CLIENT_ID
- Docker compose override with host-gateway for localhost access

### 2026-01-15 - Phase 2 Completed (Organizations, Groups, Relawan CRUD API)
Backend API endpoints (all OIDC protected):

Organizations:
- GET /api/v1/organizations (list with pagination & search)
- GET /api/v1/organizations/:id (detail with stats)
- POST /api/v1/organizations (create)
- PUT /api/v1/organizations/:id (update)
- DELETE /api/v1/organizations/:id (soft delete)
- GET /api/v1/organizations/:id/stats
- POST /api/v1/organizations/:id/members (add member)
- DELETE /api/v1/organizations/:id/members/:user_id (remove member)
- PUT /api/v1/organizations/:id/members/:user_id/role (update role)
- GET /api/v1/organizations/:id/groups
- GET /api/v1/organizations/:id/relawan

Groups:
- GET /api/v1/groups (list with filters)
- GET /api/v1/groups/:id (detail with relawan)
- POST /api/v1/groups (create)
- PUT /api/v1/groups/:id (update)
- DELETE /api/v1/groups/:id (soft delete)
- GET /api/v1/groups/:id/stats
- GET /api/v1/groups/:id/relawan

Relawan:
- GET /api/v1/relawan (list with filters)
- GET /api/v1/relawan/stats
- GET /api/v1/relawan/:id (detail)
- POST /api/v1/relawan (create)
- PUT /api/v1/relawan/:id (update)
- DELETE /api/v1/relawan/:id (soft delete)
- PUT /api/v1/relawan/:id/status
- PUT /api/v1/relawan/:id/group
- POST /api/v1/relawan/bulk/move-to-group

Files created:
- internal/repository/organization.go
- internal/repository/group.go
- internal/repository/relawan.go
- internal/service/organization.go
- internal/service/group.go
- internal/service/relawan.go
- internal/handler/organization.go
- internal/handler/group.go
- internal/handler/relawan.go

CORS updated to include localhost:5174 and admin.dayawarga.com

### 2026-01-15 - Phase 2 Completed (Frontend API Integration)
Frontend connected to backend API with full auth flow working:

Service layer:
- services/api.ts - Axios client with auth interceptors
- services/types.ts - TypeScript interfaces + ApiResponse wrapper
- services/organizations.ts - Organization CRUD
- services/groups.ts - Group CRUD
- services/relawan.ts - Relawan CRUD

Auth flow fixes:
- WebStorageStateStore for explicit localStorage persistence
- Manual localStorage fallback when oidc-client-ts fails to store
- Storage key normalization (trailing slash in authority URL)
- App.vue initializes auth before rendering routes (loading spinner)
- Router guard reads localStorage directly for consistency
- API interceptor only redirects on 401 if token was present
- Backend OIDC issuer validation normalizes trailing slash

Key files modified:
- src/App.vue - Auth init with isAuthReady gate
- src/stores/auth.ts - Manual storage, clearSession, init improvements
- src/router/index.ts - localStorage key normalization
- src/services/api.ts - 401 handling fix
- services/api/internal/auth/oidc.go - Issuer trailing slash normalization

────────────────────────────────────────────────────────────

## 19. Current Focus (Update Often)

Active:
- Admin Portal production deployment with Authentik
- WhatsApp chatbot operational testing

Completed:
- Phase 0: Frontend scaffold + Authentik setup + OIDC login
- Phase 1: Backend OIDC auth + user auto-provisioning + RBAC
- Phase 2 (Backend): Organizations, Groups, Relawan CRUD API
- Phase 2 (Frontend): API integration + auth flow fixes
- Org-Scoped Authorization: Restrict org_admin to their own organizations only
- WhatsApp Verification: Relawan WA access controlled via Admin Portal
- Chatbot Integration: WA chatbot validates relawan via Dayawarga API

Out of scope:
- Analytics
- Mobile optimization

────────────────────────────────────────────────────────────

## 20. Open Questions

- App User token rotation policy?
- Multi-project ODK support per organization?

────────────────────────────────────────────────────────────

## 21. ODK Integration Decisions (2026-01-16)

### Complete User Flow
1. Admin invites Organization Leader → registers via Authentik
2. Org Leader views available ODK Projects & Forms (from ODK Central API)
3. Org Leader creates Group → assigns ODK Project → assigns Group Leader
4. Admin reviews & approves request → Group Leader becomes Project Manager in ODK
5. Group Leader adds Relawan → auto-creates App User + QR Code
6. Relawan scans QR in ODK Collect → can submit forms

### Role Hierarchy
| Role | Scope | ODK Central Role |
|------|-------|------------------|
| Super Admin | System | - |
| Organization Leader | Organization | - |
| Group Leader (approved) | Group | Project Manager |
| Relawan | - | App User (Field Key) |

### Database Changes Required
- NEW: `project_requests` table (approval workflow)
- NEW: `group_projects` table (group ↔ project mapping)
- MODIFY: `groups` add `leader_id`, `odk_project_manager_created`

### Relawan App User
- Display name di ODK: nama relawan dari Admin Portal
- Tracking via `relawan.odk_app_user_id` (numeric, unique, immutable)
- Token disimpan di `relawan.odk_app_user_token` untuk QR code
- Nama di frontend/Dayawarga.com: dari database Admin Portal (bukan ODK)

### Implementation Phases (Revised)
1. Phase A: ODK API Client (Backend) - list projects, manage users
2. Phase B: Database Migration - approval tables, group leader
3. Phase C: Project Request Workflow - request, approve, create PM
4. Phase D: Relawan App User - auto-create on add, QR generation
5. Phase E: Frontend - complete UI for workflow

See: `.claude/skills/agent-memory/memories/admin-portal/odk-integration-plan.md`

### 2026-01-19 - Org-Scoped Authorization Completed

Implemented organization-scoped authorization to restrict org_admin users to only access their own organizations' data.

Backend Changes:
- `internal/auth/rbac.go` - RBAC helper functions:
  - `GetUserOrgIDs()` - returns org IDs user can access (nil for super_admin)
  - `CanAccessOrganization()` - check read access to org
  - `CanManageOrganization()` - check write access to org
- `internal/repository/group.go` - Added `ListByOrgIDs()`, `GetStatsByOrgIDs()`
- `internal/repository/relawan.go` - Added `ListByOrgIDs()`, `GetStatsByOrgIDs()`
- `internal/service/group.go` - Added `ListWithOrgFilter()`
- `internal/service/relawan.go` - Added `ListWithOrgFilter()`, `GetStatsWithOrgFilter()`
- `internal/handler/group.go` - All handlers check org access
- `internal/handler/relawan.go` - All handlers check org access
- `internal/handler/organization.go` - List filtered by user's orgs
- `internal/handler/group_org_scope_test.go` - 10 unit tests (ALL PASS)

Frontend Changes:
- `apps/admin-portal/src/stores/auth.ts` - Role-based computed properties:
  - `isSuperAdmin`, `isOrgAdmin`, `canManageOrganizations`, `canInviteUsers`
- `apps/admin-portal/src/services/auth.ts` - Fixed API response parsing
- `apps/admin-portal/src/views/OrganizationsView.vue` - "Tambah Organisasi" button restricted to super_admin

Authorization Matrix:
| Action | super_admin | org_admin |
|--------|-------------|-----------|
| View all orgs | Yes | No (own only) |
| Create/delete org | Yes | No |
| Manage org data | Yes (all) | Yes (own) |
| View groups/relawan | Yes (all) | Yes (own org) |

### 2026-01-26 - WhatsApp Verification Integration Completed

Implemented WhatsApp access control for relawan, integrated with external chatbot service.

Database Changes:
- `000014_add_whatsapp_verification.sql` - Added to relawan table:
  - `wa_verified` (boolean) - WA access enabled
  - `wa_verified_at` (timestamp) - When WA was enabled
  - `wa_last_activity` (timestamp) - Last chatbot interaction
  - `wa_session_count` (integer) - Total chatbot sessions

Backend API (Go):
- `internal/model/relawan.go` - Added WA fields + `HasWAAccess()` method
- `internal/repository/relawan.go` - Added `SetWAVerified()`, `UpdateWAActivity()`, `FindByPhoneWithWAAccess()`
- `internal/service/relawan.go` - Added WA service methods + `WAStatus` struct
- `internal/handler/relawan.go` - Added WA verification endpoints:
  - `POST /api/v1/relawan/:id/wa-verify` - Enable WA (OIDC protected)
  - `DELETE /api/v1/relawan/:id/wa-verify` - Revoke WA (OIDC protected)
  - `GET /api/v1/relawan/:id/wa-status` - Get status (OIDC protected)
  - `GET /api/v1/wa/validate?phone=xxx` - Validate phone (API key, for chatbot)
  - `POST /api/v1/wa/activity` - Record activity (API key, for chatbot)

Frontend (Admin Portal):
- `apps/admin-portal/src/services/types.ts` - Added WA fields to Relawan interface
- `apps/admin-portal/src/services/relawan.ts` - Added WA service methods
- `apps/admin-portal/src/views/RelawanDetailView.vue` - Added WhatsApp toggle card

Chatbot Integration (separate repo: dayawarga-chatbot):
- `src/dayawarga-api.ts` - Added `validateWAAccess()`, `recordWAActivity()`
- `src/webhook.ts` - Validates relawan via API instead of local SQLite

Deployment:
- Main platform deployed via GitHub Actions
- Chatbot deployed via GitHub Actions (PM2)
- Both repos have CI/CD workflows configured
