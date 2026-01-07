# Dayawarga Senyar 2025

Sistem Informasi Geografis (GIS) berbasis collection dan layer data eksisting untuk pemantauan dan pengelolaan data bencana Siklon Senyar di Sumatra 2025. Platform ini menyediakan peta interaktif untuk visualisasi lokasi posko pengungsian, informasi terbaru (feeds), dan dokumentasi foto dari lapangan.

**Live Demo:** [https://dayawarga.com](https://dayawarga.com)

## Arsitektur Sistem

```
┌─────────────────────────────────────────────────────────────────┐
│                        Traefik (Reverse Proxy)                  │
│                    SSL/TLS + Load Balancing                     │
└─────────────────┬───────────────────────────────┬───────────────┘
                  │                               │
                  ▼                               ▼
┌─────────────────────────────┐   ┌─────────────────────────────┐
│      Frontend (Vue 3)       │   │      API (Go/Gin)           │
│  - Peta Leaflet             │   │  - REST API                 │
│  - Responsive UI            │   │  - ODK Central Sync         │
│  - Real-time Feeds          │   │  - Photo Management         │
└─────────────────────────────┘   └──────────────┬──────────────┘
                                                 │
                  ┌──────────────────────────────┼──────────────┐
                  │                              │              │
                  ▼                              ▼              ▼
┌─────────────────────────┐   ┌─────────────────────────┐   ┌───────────┐
│  PostgreSQL + PostGIS   │   │      ODK Central        │   │  Storage  │
│  - Geospatial Data      │   │  - Form Submissions     │   │  - Photos │
│  - Locations & Feeds    │   │  - Field Data           │   │           │
└─────────────────────────┘   └─────────────────────────┘   └───────────┘
```

## Tech Stack

### Backend (services/api)
- **Go 1.21+** dengan Gin framework
- **PostgreSQL 16** dengan PostGIS untuk data geospasial
- **GORM** sebagai ORM
- Integrasi **ODK Central API** untuk sinkronisasi data lapangan
- Scheduler otomatis untuk sync berkala

### Frontend (services/frontend)
- **Vue 3** dengan Composition API
- **TypeScript** untuk type safety
- **Tailwind CSS** untuk styling
- **Leaflet** untuk peta interaktif
- **Vite** sebagai build tool

### Infrastructure
- **Docker** & **Docker Compose** untuk containerization
- **Traefik** sebagai reverse proxy dengan auto SSL
- **GitHub Actions** untuk CI/CD

## Struktur Direktori

```
dayawarga-senyar-2025/
├── services/
│   ├── api/                    # Backend Go API
│   │   ├── cmd/
│   │   │   ├── api/            # Main API server
│   │   │   └── importer/       # CLI tool untuk import data
│   │   ├── internal/
│   │   │   ├── handler/        # HTTP handlers
│   │   │   ├── repository/     # Database queries
│   │   │   ├── service/        # Business logic
│   │   │   ├── model/          # Data models
│   │   │   └── scheduler/      # Background jobs
│   │   └── Dockerfile
│   │
│   └── frontend/               # Vue.js Frontend
│       ├── src/
│       │   ├── components/     # Vue components
│       │   ├── views/          # Page views
│       │   ├── services/       # API client
│       │   └── composables/    # Vue composables
│       └── Dockerfile
│
├── infrastructure/
│   ├── database/               # Database migrations
│   └── traefik/                # Traefik configuration
│
├── .github/workflows/          # CI/CD pipelines
├── docker-compose.yml          # Production compose
└── .env.example                # Environment template
```

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

| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| GET | `/api/v1/locations` | Daftar lokasi posko (GeoJSON) |
| GET | `/api/v1/locations/:id` | Detail lokasi |
| GET | `/api/v1/locations/:id/photos` | Foto lokasi |
| GET | `/api/v1/feeds` | Daftar feeds/update |
| GET | `/api/v1/photos/:id/file` | Download foto |
| POST | `/api/v1/sync/posko` | Trigger sync posko |
| POST | `/api/v1/sync/photos` | Trigger sync foto |

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

## Next release in a week

1. Tambah form ODK, Sync dan tampilan untuk verifikasi Fasilitas Kesehatan (data available Tim Tanggap Darurat Kemenkes)
2. Tambah form ODK, Sync dan tampilan untuk verifikasi titik jembatan (data available tim survei Kementerian PU)
3. Feed Updates bisa koordinat bebas untuk laporan relawan di mana saja
4. Feed Updates tampil di peta dgn foto visual situasi

## Kontak Developer

- **GitHub Issues:** [github.com/leksa/dayawarga-senyar-2025/issues](https://github.com/leksa/dayawarga-senyar-2025/issues)
- **Email:** [dayawarga@gmail.com](mailto:dayawarga@gmail.com)
- **Website:** [dayawarga.com](https://dayawarga.com)

## License

MIT License - Silakan gunakan dan modifikasi sesuai kebutuhan.

---

**Dayawarga** - Platform Pemantauan Bencana Berbasis Komunitas Relawan dan Warga
