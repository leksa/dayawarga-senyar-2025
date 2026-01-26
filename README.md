# Dayawarga Senyar 2025

Sistem Informasi Geografis (GIS) berbasis collection dan layer data eksisting untuk pemantauan dan pengelolaan data bencana Siklon Senyar di Sumatra 2025. Platform ini menyediakan peta interaktif untuk visualisasi lokasi posko pengungsian, informasi terbaru (feeds), dan dokumentasi foto dari lapangan.

**Live Demo:** [https://dayawarga.com](https://dayawarga.com)

## Infrastructure

### Production Services

| Service | Domain | Description |
|---------|--------|-------------|
| **Main Platform** | dayawarga.com | Frontend, API, Ghost CMS |
| **Admin Portal** | admin.dayawarga.com | Relawan management |
| **Auth (SSO)** | auth.dayawarga.com | Authentik Identity Provider |
| **ODK Central** | data.dayawarga.com | Data collection forms |
| **WhatsApp Chatbot** | (internal) | Relawan chatbot service |

## Arsitektur Sistem

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              Main Platform Server                               │
├─────────────────────────────────────────────────────────────────────────────────┤
│                        Traefik (Reverse Proxy)                                  │
│                    SSL/TLS + Load Balancing                                     │
└───────┬───────────────┬───────────────┬───────────────┬───────────────┬─────────┘
        │               │               │               │               │
        ▼               ▼               ▼               ▼               ▼
┌───────────────┐ ┌───────────┐ ┌───────────────┐ ┌───────────┐ ┌─────────────┐
│   Frontend    │ │    API    │ │ Admin Portal  │ │ Authentik │ │  Ghost CMS  │
│   (Vue 3)     │ │  (Go/Gin) │ │   (Vue 3)     │ │   (SSO)   │ │   (Blog)    │
│ dayawarga.com │ │ api.daya..│ │ admin.daya... │ │ auth.daya.│ │ stories...  │
└───────────────┘ └─────┬─────┘ └───────────────┘ └───────────┘ └─────────────┘
                        │
        ┌───────────────┼───────────────┐
        │               │               │
        ▼               ▼               ▼
┌───────────────┐ ┌───────────────┐ ┌───────────┐
│  PostgreSQL   │ │  ODK Central  │ │  Storage  │
│  + PostGIS    │ │ data.daya...  │ │  (S3/R2)  │
└───────────────┘ └───────────────┘ └───────────┘

┌─────────────────────────────────────────────────────────────────────────────────┐
│                            Chatbot Server (Separate)                            │
├─────────────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────────────────────────────┐    │
│  │                    WhatsApp Chatbot (Node.js + PM2)                     │    │
│  │  - Relawan verification via API                                         │    │
│  │  - LLM-powered conversation                                             │    │
│  │  - Feed & Posko submission                                              │    │
│  └─────────────────────────────────────────────────────────────────────────┘    │
│                                    │                                            │
│                                    ▼                                            │
│                    ┌───────────────────────────────┐                            │
│                    │      Dayawarga API            │                            │
│                    │  - WA validation endpoint     │                            │
│                    │  - Activity tracking          │                            │
│                    └───────────────────────────────┘                            │
└─────────────────────────────────────────────────────────────────────────────────┘
```

## Tech Stack

### Backend (services/api)
- **Go 1.24+** dengan Gin framework
- **PostgreSQL 16** dengan PostGIS untuk data geospasial
- **GORM** sebagai ORM
- Integrasi **ODK Central API** untuk sinkronisasi data lapangan
- Scheduler otomatis untuk sync berkala
- **OIDC Auth** via Authentik untuk Admin Portal

### Frontend (services/frontend)
- **Vue 3** dengan Composition API
- **TypeScript** untuk type safety
- **Tailwind CSS** untuk styling
- **Leaflet** untuk peta interaktif
- **Vite** sebagai build tool

### Admin Portal (apps/admin-portal)
- **Vue 3** dengan Composition API + TypeScript
- **shadcn-vue** untuk UI components
- **OIDC authentication** via Authentik
- Manajemen organisasi, grup, dan relawan
- Integrasi ODK App User untuk QR code
- WhatsApp verification untuk relawan

### WhatsApp Chatbot (separate repo: dayawarga-chatbot)
- **Node.js** dengan TypeScript
- **Claude AI** untuk conversational UX
- **SQLite** untuk session management
- Integrasi dengan Dayawarga API untuk feed/posko submission
- Relawan verification via Admin Portal

### Infrastructure
- **Docker** & **Docker Compose** untuk containerization
- **Traefik** sebagai reverse proxy dengan auto SSL
- **Authentik** sebagai Identity Provider (SSO)
- **GitHub Actions** untuk CI/CD
- **PM2** untuk chatbot process management

## Struktur Direktori

```
dayawarga-senyar-2025/
├── services/
│   ├── api/                    # Backend Go API
│   │   ├── cmd/api/            # Main API server
│   │   ├── internal/
│   │   │   ├── auth/           # OIDC + RBAC middleware
│   │   │   ├── handler/        # HTTP handlers
│   │   │   ├── repository/     # Database queries
│   │   │   ├── service/        # Business logic
│   │   │   ├── model/          # Data models
│   │   │   ├── odk/            # ODK Central API client
│   │   │   └── scheduler/      # Background jobs
│   │   └── Dockerfile
│   │
│   └── frontend/               # Vue.js Public Frontend
│       ├── src/
│       │   ├── components/     # Vue components
│       │   ├── views/          # Page views
│       │   └── services/       # API client
│       └── Dockerfile
│
├── apps/
│   └── admin-portal/           # Admin Portal (Vue.js)
│       ├── src/
│       │   ├── views/          # Dashboard, Org, Group, Relawan views
│       │   ├── services/       # API client with OIDC
│       │   ├── stores/         # Pinia stores (auth)
│       │   └── components/     # shadcn-vue components
│       └── Dockerfile
│
├── infrastructure/
│   ├── database/migrations/    # SQL migrations
│   ├── traefik/                # Traefik configuration
│   └── authentik/              # Authentik IdP setup
│
├── .github/workflows/          # CI/CD pipelines
├── docker-compose.yml          # Production compose (with profiles)
└── .env.example                # Environment template
```

### Docker Compose Profiles

| Profile | Services | Command |
|---------|----------|---------|
| (default) | traefik, postgres, api, frontend, ghost | `docker compose up -d` |
| `admin-portal` | + authentik, admin-portal | `docker compose --profile admin-portal up -d` |
| `tools` | + pgadmin | `docker compose --profile tools up -d` |
| `docs` | + api-docs | `docker compose --profile docs up -d` |

## Fitur Utama

- **Peta Interaktif** - Visualisasi lokasi posko pengungsian dengan marker berwarna berdasarkan status
- **Detail Posko** - Informasi lengkap termasuk jumlah pengungsi, fasilitas, kontak, dan foto
- **Feeds/Update** - Timeline informasi terbaru dari lapangan dengan filter kategori dan tags
- **Photo Gallery** - Dokumentasi foto dari setiap lokasi posko
- **Responsive Design** - Optimal untuk desktop dan mobile
- **Auto Sync** - Sinkronisasi otomatis dengan ODK Central setiap 5 menit

## Development Setup

### Prerequisites
- Node.js 18+
- Go 1.21+
- Docker & Docker Compose
- PostgreSQL 16 dengan PostGIS (atau gunakan Docker)

### Quick Start

1. **Clone repository**
   ```bash
   git clone https://github.com/leksa/dayawarga-senyar-2025.git
   cd dayawarga-senyar-2025
   ```

2. **Setup environment**
   ```bash
   cp .env.example .env
   # Edit .env dengan konfigurasi lokal
   ```

3. **Jalankan dengan Docker Compose**
   ```bash
   docker-compose up -d
   ```

4. **Atau jalankan secara terpisah untuk development:**

   **Backend:**
   ```bash
   cd services/api
   go mod download
   go run cmd/api/main.go
   ```

   **Frontend:**
   ```bash
   cd services/frontend
   npm install
   npm run dev
   ```

5. **Akses aplikasi:**
   - Frontend: http://localhost:5173
   - API: http://localhost:8080/api/v1

## API Endpoints

### Public API (No Auth)

| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| GET | `/api/v1/locations` | Daftar lokasi posko (GeoJSON) |
| GET | `/api/v1/locations/:id` | Detail lokasi |
| GET | `/api/v1/locations/:id/photos` | Foto lokasi |
| GET | `/api/v1/feeds` | Daftar feeds/update |
| GET | `/api/v1/photos/:id/file` | Download foto |
| POST | `/api/v1/sync/posko` | Trigger sync posko |
| POST | `/api/v1/sync/photos` | Trigger sync foto |

### Admin Portal API (OIDC Protected)

| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| GET | `/api/v1/auth/me` | Current user info |
| GET/POST/PUT/DELETE | `/api/v1/organizations/*` | Org management |
| GET/POST/PUT/DELETE | `/api/v1/groups/*` | Group management |
| GET/POST/PUT/DELETE | `/api/v1/relawan/*` | Relawan management |
| POST | `/api/v1/relawan/:id/wa-verify` | Enable WA access |
| DELETE | `/api/v1/relawan/:id/wa-verify` | Revoke WA access |
| GET | `/api/v1/relawan/:id/wa-status` | WA verification status |

### WhatsApp Chatbot API (API Key Protected)

| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| GET | `/api/v1/wa/validate?phone=xxx` | Validate phone has WA access |
| POST | `/api/v1/wa/activity` | Record chatbot activity |

## Branching Strategy

```
main (production)
  │
  └── development (staging)
        │
        └── feature/* atau fix/* (working branches)
```

- **`main`** - Branch production, deploy otomatis ke server. **Hanya menerima Pull Request**, tidak ada push langsung.
- **`development`** - Branch staging untuk testing. Merge dari feature branches.
- **`feature/*`** atau **`fix/*`** - Branch untuk pengembangan fitur atau perbaikan bug.

## Cara Berkontribusi

Kami sangat terbuka untuk kontribusi! Berikut langkah-langkahnya:

1. **Fork repository ini**

2. **Clone fork Anda**
   ```bash
   git clone https://github.com/YOUR_USERNAME/dayawarga-senyar-2025.git
   ```

3. **Buat branch baru dari `development`**
   ```bash
   git checkout development
   git pull origin development
   git checkout -b feature/nama-fitur
   ```

4. **Lakukan perubahan dan commit**
   ```bash
   git add .
   git commit -m "feat: deskripsi perubahan"
   ```

5. **Push ke fork Anda**
   ```bash
   git push origin feature/nama-fitur
   ```

6. **Buat Pull Request ke branch `development`**

### Panduan Commit Message

Gunakan format [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` - Fitur baru
- `fix:` - Perbaikan bug
- `docs:` - Perubahan dokumentasi
- `refactor:` - Refactoring kode
- `test:` - Penambahan/perbaikan test
- `chore:` - Maintenance

### Code Review

- Semua Pull Request akan di-review sebelum merge
- Pastikan tidak ada conflict dengan branch target
- Pastikan build dan test berhasil

## Issues & Support

Jika Anda menemukan bug atau memiliki ide fitur baru:

1. **Cek [Issues](https://github.com/leksa/dayawarga-senyar-2025/issues)** yang sudah ada
2. **Buat Issue baru** dengan deskripsi yang jelas
3. **Gunakan label** yang sesuai (bug, enhancement, question, dll)

## Recent Updates

### January 2026
- ✅ **Admin Portal** - Full relawan management with OIDC auth
- ✅ **WhatsApp Integration** - Chatbot validates relawan via Admin Portal
- ✅ **ODK Integration** - App User creation with QR code for ODK Collect
- ✅ **Dual-stream Sync** - Conflict resolution for ODK data

### Roadmap
- [ ] Tambah form ODK untuk verifikasi titik jembatan (data Kementerian PU)
- [ ] Halaman repo data terupdate (sync daily) dalam csv/json
- [ ] Analytics dashboard di Admin Portal

## Kontak Developer

- **GitHub Issues:** [github.com/leksa/dayawarga-senyar-2025/issues](https://github.com/leksa/dayawarga-senyar-2025/issues)
- **Email:** [dayawarga@gmail.com](mailto:leksa.rizal@gmail.com)
- **Website:** [dayawarga.com](https://dayawarga.com)

## License

MIT License - Silakan gunakan dan modifikasi sesuai kebutuhan.

---

**Dayawarga** - Platform Pemantauan Bencana Berbasis Komunitas Relawan dan Warga
