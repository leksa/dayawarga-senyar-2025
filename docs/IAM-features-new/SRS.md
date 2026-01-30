# Software Requirements Specification (SRS)
# ODK Admin Portal

> **⚠️ ARCHIVED - NOT IMPLEMENTED**
> 
> This SRS describes a **Node.js + Prisma** system that was NOT built.
> The actual implementation uses **Go + Gin + GORM**.
> 
> See `/claude.md` section "Implementation Phases (Authoritative)" for actual implementation.

**Version:** 1.0  
**Date:** January 2025  
**Status:** ~~Final Draft~~ **ARCHIVED**

---

## 1. Executive Summary

ODK Admin Portal adalah sistem manajemen user dan organisasi yang terintegrasi dengan **Authentik** (Identity Provider) dan **ODK Central** (Data Collection Backend). Sistem ini menyediakan hierarki 4 level: Administrator → Admin Lembaga → Project Manager → Relawan.

### Key Features
- ✅ Authentication via Authentik (OIDC)
- ✅ Multi-level organization hierarchy (Lembaga → Sub-Lembaga)
- ✅ ODK Project linking per Sub-Lembaga
- ✅ Automatic ODK App User creation with QR codes
- ✅ Role-based submission viewing
- ✅ Simple dashboard statistics

---

## 2. User Roles & Permissions

### 2.1 Hierarchy

```
ADMINISTRATOR (Level 1)
    │
    └── LEMBAGA
        └── ADMIN LEMBAGA (Level 2) - 1 per lembaga
            │
            └── SUB-LEMBAGA (= ODK Project)
                ├── PROJECT MANAGER (Level 3) - can be multiple
                │   └── can assign Relawan only
                │
                └── RELAWAN (Level 4)
                    └── submit forms, view own QR & submissions
```

### 2.2 Permission Matrix

| Permission | Admin | Admin Lembaga | PM | Relawan |
|------------|:-----:|:-------------:|:--:|:-------:|
| Manage Lembaga | ✅ | ❌ | ❌ | ❌ |
| Create Sub-Lembaga | ✅ | ✅ | ❌ | ❌ |
| Assign PM | ✅ | ✅ | ❌ | ❌ |
| Assign Relawan | ✅ | ✅ | ✅ | ❌ |
| View Lembaga Submissions | ✅ | ✅ | ❌ | ❌ |
| View Sub-Lembaga Submissions | ✅ | ✅ | ✅ | ❌ |
| View Own Submissions | ✅ | ✅ | ✅ | ✅ |
| View Own QR Codes | ✅ | ✅ | ✅ | ✅ |

### 2.3 Constraints
- Satu Lembaga = Satu Admin Lembaga
- User tidak bisa jadi Admin Lembaga DAN member di lembaga yang sama
- PM hanya bisa assign Relawan, tidak bisa assign PM lain
- Remove member = auto revoke ODK App User

---

## 3. Functional Requirements

### 3.1 Authentication
| ID | Requirement |
|----|-------------|
| AUTH-01 | Login via Authentik OIDC |
| AUTH-02 | Auto-create user on first login |
| AUTH-03 | Session management with JWT |
| AUTH-04 | Logout with Authentik session termination |

### 3.2 User Management
| ID | Requirement |
|----|-------------|
| USER-01 | CRUD users (role-based access) |
| USER-02 | Sync user creation to Authentik |
| USER-03 | Deactivate user = revoke all ODK App Users |
| USER-04 | Filter by lembaga, role, status |

### 3.3 Lembaga Management
| ID | Requirement |
|----|-------------|
| LEMB-01 | CRUD Lembaga (Administrator only) |
| LEMB-02 | Unique lembaga code |
| LEMB-03 | Assign Admin Lembaga |
| LEMB-04 | Cascade delete to Sub-Lembaga |

### 3.4 Sub-Lembaga Management
| ID | Requirement |
|----|-------------|
| SUBL-01 | Create Sub-Lembaga by linking ODK Project |
| SUBL-02 | One ODK Project = One Sub-Lembaga (unique) |
| SUBL-03 | Display available (unlinked) ODK Projects |
| SUBL-04 | Delete = revoke all member App Users |

### 3.5 Member Assignment
| ID | Requirement |
|----|-------------|
| MEMB-01 | Assign members to Sub-Lembaga |
| MEMB-02 | Set role: PM or Relawan |
| MEMB-03 | Auto-create ODK App User on assignment |
| MEMB-04 | Auto-generate QR code |
| MEMB-05 | One user can be in multiple Sub-Lembaga |
| MEMB-06 | Each assignment = separate ODK App User & QR |
| MEMB-07 | Remove = revoke ODK App User |
| MEMB-08 | Regenerate QR = new App User |

### 3.6 QR Code Management
| ID | Requirement |
|----|-------------|
| QR-01 | ODK-compatible QR format |
| QR-02 | Contains: server URL + App User token |
| QR-03 | Display SubmitterID with QR |
| QR-04 | Download as PNG |

### 3.7 Submissions
| ID | Requirement |
|----|-------------|
| SUBM-01 | View submissions (role-filtered) |
| SUBM-02 | Map SubmitterID to user name |
| SUBM-03 | Filter by date, form |

### 3.8 Dashboard
| ID | Requirement |
|----|-------------|
| DASH-01 | Role-appropriate statistics |
| DASH-02 | Submission counts: today, week, month |
| DASH-03 | Recent submissions list |

---

## 4. Non-Functional Requirements

| Category | Requirement |
|----------|-------------|
| **Performance** | Page load < 3s, API < 500ms |
| **Scale** | 100-500 concurrent users |
| **Availability** | 99% uptime |
| **Security** | HTTPS, JWT, RBAC |
| **Usability** | Indonesian language, responsive |

---

## 5. Technology Stack

| Layer | Technology |
|-------|------------|
| Frontend | Vue.js 3, Tailwind CSS, shadcn-vue |
| Backend | Node.js, Express, TypeScript |
| Database | PostgreSQL 16, Prisma ORM |
| Queue | Redis, BullMQ |
| Auth | Authentik (OIDC) |
| ODK | ODK Central API |

---

## 6. Data Model Summary

```
users
  ├── id, authentik_id, email, name
  ├── system_role (administrator|admin_lembaga|member)
  └── lembaga_id

lembaga
  ├── id, name, code
  └── is_active

sub_lembaga
  ├── id, lembaga_id, name
  ├── odk_project_id (unique)
  └── is_active

sub_lembaga_members
  ├── id, sub_lembaga_id, user_id
  ├── role (project_manager|relawan)
  ├── odk_app_user_id, qr_code_data
  └── is_active
```

---

## 7. Key Flows

### 7.1 Create Sub-Lembaga
1. Admin Lembaga opens Sub-Lembaga page
2. System fetches available ODK Projects (unlinked)
3. Admin selects ODK Project
4. System creates Sub-Lembaga record with ODK Project ID
5. Sub-Lembaga ready for member assignment

### 7.2 Assign Member
1. Admin/PM selects member from lembaga pool
2. Admin/PM selects Sub-Lembaga and role
3. System creates ODK App User via API
4. System assigns App User to all forms in project
5. System stores QR code data
6. Member can now view their QR code

### 7.3 Remove Member
1. Admin/PM removes member from Sub-Lembaga
2. System revokes ODK App User via API
3. System marks assignment as inactive
4. QR code becomes invalid

---

## 8. UI Pages

| Page | Roles | Purpose |
|------|-------|---------|
| Dashboard | All | Statistics overview |
| Lembaga | Admin | Manage organizations |
| Sub-Lembaga | Admin, Admin Lembaga | Manage projects |
| Members | Admin, Admin Lembaga, PM | Manage assignments |
| My Projects | All | View own QR codes |
| Submissions | All | View submissions (filtered) |

---

## Appendix: Glossary

| Term | Definition |
|------|------------|
| Lembaga | Organization/institution |
| Sub-Lembaga | Project unit within Lembaga, maps 1:1 to ODK Project |
| Admin Lembaga | Administrator for one Lembaga |
| PM | Project Manager for one Sub-Lembaga |
| Relawan | Field enumerator/data collector |
| App User | ODK Central user for mobile authentication |
| SubmitterID | ODK identifier tracking who submitted |
