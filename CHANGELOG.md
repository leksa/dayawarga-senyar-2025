# Changelog

All notable changes to Dayawarga Senyar will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.4.0] - 2026-01-30

### Added
- **Multi-Photo Support** - Feed submissions now support up to 4 photos (foto, foto2, foto3, foto4)
- **FeedDetailView** - New dedicated page for viewing individual feed details with swipe gestures
- **Timeline View** - Location/Posko detail panel now shows internal timeline of all feeds
- **Desa Timeline** - Village-level feed timeline with statistics
- **Share Image Template** - Generate shareable images from feed content with category-based styling

### Changed
- **Standardized Category/Tag Colors** - Created centralized `feedHelpers.ts` utility for consistent badge colors across all components:
  - `kebutuhan` → danger (red)
  - `informasi` → warning (yellow/amber)  
  - `follow-up` → orange
  - `info_bantuan` → success (green)
  - All tags → info (light blue)
- **Mobile Optimization** - Collapsible filters, infinite scroll, pull-to-refresh on FeedsView
- **Safe Area Padding** - Added bottom padding for mobile browser docks on FeedDetailView

### Fixed
- Feed photo extraction now properly handles all 4 photo fields from ODK submissions
- Category badge colors now consistent between FeedsView, DetailPanel, MapView, FeedDetailView, and ShareImageTemplate

### Infrastructure
- ODK Central upgraded from v2025.4.0 to v2025.4.2 on production
- Increased Enketo payload limit from 1MB to 100MB for large photo submissions

## [1.3.0] - 2026-01-28

### Added
- **Admin Portal** - Full relawan management with OIDC authentication via Authentik
- **WhatsApp Verification** - Relawan WA access control integrated with chatbot
- **User Invitation System** - Invite users with WhatsApp PIN verification (6-char alphanumeric, 15-min expiry)
- **ODK Integration** - App User creation with QR code for ODK Collect
- **Org-Scoped Authorization** - org_admin restricted to their own organizations only
- **Project Request Workflow** - Request/approve ODK project assignments to groups
- **Dual-stream Sync** - Conflict resolution for ODK data with entity tracking

### Security
- Removed `.claude/environment.md` from git tracking (contained server IPs)
- Added environment.md to .gitignore

### Documentation
- Updated `claude.md` with complete project constitution and decisions log
- Archived outdated IAM-features-new SRS (described different tech stack)

### Infrastructure
- Database migrations: 000009-000015 (admin portal, ODK, WhatsApp verification, invitation PIN)
- WhatsApp chatbot operational on separate server (dayawarga-chatbot repo)
- CI/CD via GitHub Actions for both main platform and chatbot

## [1.2.0] - 2025-01-14

### Added
- Bailey stats di StatsBox infrastruktur (Bailey Terpasang, Bailey Sedang Dipasang)
- Field `bailey` di API response infrastruktur list
- Detail statistik Faskes (Rumah Sakit, Puskesmas, Posko Kes Darurat, tidak beroperasi)
- Detail statistik Infrastruktur (Jalan/Jembatan Sudah/Sedang Ditangani)
- Kebutuhan Air dalam liter di StatsBox Posko
- Region filter berbasis ID (BPS code) untuk filtering yang lebih akurat

### Changed
- Label "Balita" menjadi "Bayi & Balita" di StatsBox Posko
- Hapus default Leaflet zoom control (custom zoom buttons)
- Cleanup scripts folder - hanya menyimpan essential scripts untuk crontab/automation

### Fixed
- Progress infrastruktur default berdasarkan status_penanganan (belum=0, sedang=50, sudah=100)
- Data inconsistency penanganan_detail "Tuntas" dengan status_penanganan "belum_ditangani"

## [1.1.0] - 2025-01-08

### Added
- Hard sync feature untuk sinkronisasi dan menghapus data yang sudah tidak ada di ODK Central
  - `POST /api/v1/sync/posko/hard` - Hard sync posko/locations
  - `POST /api/v1/sync/feed/hard` - Hard sync feeds
  - `POST /api/v1/sync/faskes/hard` - Hard sync faskes
- CLI sync script (`scripts/sync-all.sh`) dengan commands:
  - `all` - Sync semua data
  - `hard` - Hard sync semua (termasuk delete orphans)
  - `hard-posko`, `hard-feeds`, `hard-faskes` - Hard sync per tipe
  - `photos`, `photos-posko`, `photos-feed`, `photos-faskes` - Sync foto
  - `status` - Lihat status sync
- Versioning system dengan VERSION file dan CHANGELOG.md
- Version info di footer frontend

### Changed
- SyncResult struct sekarang termasuk field `deleted` dan `skipped`
- Photo endpoints sekarang redirect langsung ke S3 URL (HTTP 302)

### Fixed
- S3 path prefix tidak tersimpan dengan benar saat deploy
- Feed photo 404 karena record belum ter-sync ke database
- CORS localhost:8080 error pada frontend production

## [1.0.0] - 2025-01-07

### Added
- Initial release Dayawarga Senyar 2025
- Integrasi ODK Central untuk data collection
- Sync service untuk Posko, Feed, dan Faskes
- S3 storage support (CloudHost is3.cloudhost.id)
- Photo caching dan migration ke S3
- Real-time updates via SSE
- Auto-scheduler untuk periodic sync
- Interactive map dengan Leaflet/MapLibre
- Detail panel untuk location info
- Feed timeline view
- Faskes (health facilities) support

### Infrastructure
- Go API dengan Gin framework
- Vue 3 frontend dengan TypeScript
- PostgreSQL dengan PostGIS
- Docker Compose deployment
- Traefik reverse proxy dengan auto SSL
- GitHub Actions CI/CD
