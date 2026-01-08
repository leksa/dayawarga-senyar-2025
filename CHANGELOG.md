# Changelog

All notable changes to Dayawarga Senyar will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
