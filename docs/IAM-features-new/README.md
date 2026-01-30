# ODK Admin Portal

> **⚠️ ARCHIVED DOCUMENTATION**
> 
> This document describes an **ALTERNATIVE design** that was NOT implemented.
> The actual implementation uses:
> - **Go + Gin** backend (not Node.js + Express)
> - **GORM** ORM (not Prisma)
> - **3-level hierarchy**: Organization → Group → Relawan (not 4-level Lembaga structure)
> 
> For current implementation details, see:
> - `/README.md` - System overview
> - `/claude.md` - Project constitution and decisions
> - `/CHANGELOG.md` - Implementation history
> 
> **Relevant files from this folder:**
> - `flowcharts.md` - Flow diagrams (still relevant)
> - `wireframes.md` - UI designs (still relevant)
> - `invitation-whatsapp-verification.md` - Invitation flow (implemented)

---

Admin portal untuk manajemen user dan organisasi yang terintegrasi dengan **Authentik** (Identity Provider) dan **ODK Central** (Data Collection Backend).

## 🎯 Features

- ✅ Authentication via Authentik (OIDC)
- ✅ Multi-level organization hierarchy (Lembaga → Sub-Lembaga)
- ✅ ODK Project linking per Sub-Lembaga
- ✅ Automatic ODK App User creation with QR codes
- ✅ Role-based submission viewing
- ✅ Simple dashboard statistics

## 👥 User Hierarchy

```
ADMINISTRATOR (Level 1)
└── LEMBAGA
    └── ADMIN LEMBAGA (Level 2)
        └── SUB-LEMBAGA (= ODK Project)
            ├── PROJECT MANAGER (Level 3)
            └── RELAWAN (Level 4)
```

## 🛠 Technology Stack

| Layer | Technology |
|-------|------------|
| Frontend | Vue.js 3, Tailwind CSS, shadcn-vue |
| Backend | Node.js, Express, TypeScript |
| Database | PostgreSQL 16, Prisma ORM |
| Queue | Redis, BullMQ |
| Auth | Authentik (OIDC) |
| ODK | ODK Central API |

## 📁 Project Structure

```
odk-admin-portal/
├── frontend/          # Vue.js application
├── backend/           # Node.js API
├── database/          # SQL schema
├── prisma/            # Prisma schema
├── docs/              # Documentation
│   ├── SRS.md         # Requirements specification
│   ├── api-spec.yaml  # OpenAPI 3.0 specification
│   ├── flowcharts.md  # System flowcharts
│   └── wireframes.md  # UI wireframes
├── docker-compose.yml
└── .env.example
```

## 🚀 Quick Start

### Prerequisites

- Docker & Docker Compose
- Node.js 20+ (for local development)
- Running Authentik instance
- Running ODK Central instance

### Setup

1. Clone repository
```bash
git clone <repository-url>
cd odk-admin-portal
```

2. Copy environment file
```bash
cp .env.example .env
```

3. Configure `.env` with your:
   - Authentik OIDC credentials
   - ODK Central credentials
   - Database settings

4. Start with Docker Compose
```bash
docker-compose up -d
```

5. Access the application
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:3001

## 📖 Documentation

- [Software Requirements Specification (SRS)](./docs/SRS.md)
- [API Specification (OpenAPI 3.0)](./docs/api-spec.yaml)
- [System Flowcharts](./docs/flowcharts.md)
- [UI Wireframes](./docs/wireframes.md)

## 🔐 Authentik Setup

1. Create a new OAuth2/OIDC Provider in Authentik
2. Configure:
   - Client Type: Confidential
   - Redirect URIs: `http://localhost:3001/api/auth/callback`
   - Scopes: `openid email profile`
3. Create an Application linked to the provider
4. Copy Client ID and Secret to `.env`

## 📱 ODK Central Setup

1. Ensure you have admin access to ODK Central
2. Note the admin email and password
3. Add credentials to `.env`
4. Projects are created in ODK Central, then linked via Admin Portal

## 📝 License

[MIT License](LICENSE)
