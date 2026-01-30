-- Performance indexes for locations table
-- These indexes improve query performance for common filter patterns

-- Index for type filter (common in location queries)
CREATE INDEX IF NOT EXISTS idx_locations_type 
ON locations(type) 
WHERE deleted_at IS NULL;

-- Index for status filter
CREATE INDEX IF NOT EXISTS idx_locations_status 
ON locations(status) 
WHERE deleted_at IS NULL;

-- Composite index for type + status (common filter combination)
CREATE INDEX IF NOT EXISTS idx_locations_type_status 
ON locations(type, status) 
WHERE deleted_at IS NULL;

-- Index for updated_at (used for ordering and incremental sync)
CREATE INDEX IF NOT EXISTS idx_locations_updated_at 
ON locations(updated_at DESC) 
WHERE deleted_at IS NULL;

-- Index for ODK submission ID (used in sync operations)
CREATE INDEX IF NOT EXISTS idx_locations_odk_submission_id 
ON locations(odk_submission_id);

-- GIST index for spatial queries (PostGIS)
-- This is CRITICAL for bbox queries and distance calculations
CREATE INDEX IF NOT EXISTS idx_locations_geom 
ON locations USING GIST (geom)
WHERE deleted_at IS NULL;

-- Composite index for type + updated_at (common for filtered lists with sorting)
CREATE INDEX IF NOT EXISTS idx_locations_type_updated 
ON locations(type, updated_at DESC) 
WHERE deleted_at IS NULL;

-- Partial index for active locations (status = 'active')
CREATE INDEX IF NOT EXISTS idx_locations_active 
ON locations(updated_at DESC)
WHERE status = 'active' AND deleted_at IS NULL;

-- Performance indexes for feeds table
CREATE INDEX IF NOT EXISTS idx_feeds_created_at 
ON feeds(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_feeds_type 
ON feeds(type);

CREATE INDEX IF NOT EXISTS idx_feeds_location_id 
ON feeds(location_id);

-- Performance indexes for location_photos table
CREATE INDEX IF NOT EXISTS idx_location_photos_location_id 
ON location_photos(location_id);

CREATE INDEX IF NOT EXISTS idx_location_photos_type 
ON location_photos(photo_type);

-- Performance indexes for sync_state table
CREATE INDEX IF NOT EXISTS idx_sync_state_form_id 
ON sync_state(form_id);

CREATE INDEX IF NOT EXISTS idx_sync_state_status 
ON sync_state(status);

-- Performance indexes for wilayah tables
CREATE INDEX IF NOT EXISTS idx_wilayah_provinsi_kode 
ON wilayah_provinsi(kode);

CREATE INDEX IF NOT EXISTS idx_wilayah_kota_kab_kode 
ON wilayah_kota_kab(kode);

CREATE INDEX IF NOT EXISTS idx_wilayah_kecamatan_kode 
ON wilayah_kecamatan(kode);

CREATE INDEX IF NOT EXISTS idx_wilayah_desa_kode 
ON wilayah_desa(kode);

-- Add comments for documentation
COMMENT ON INDEX idx_locations_type IS 'Index for filtering locations by type';
COMMENT ON INDEX idx_locations_status IS 'Index for filtering locations by status';
COMMENT ON INDEX idx_locations_type_status IS 'Composite index for type and status filters';
COMMENT ON INDEX idx_locations_updated_at IS 'Index for ordering by updated_at (used in pagination)';
COMMENT ON INDEX idx_locations_odk_submission_id IS 'Index for ODK submission ID lookups';
COMMENT ON INDEX idx_locations_geom IS 'PostGIS spatial index for bbox and distance queries';
COMMENT ON INDEX idx_locations_type_updated IS 'Composite index for type filter with sorting';
COMMENT ON INDEX idx_locations_active IS 'Partial index for active locations only';

COMMENT ON INDEX idx_feeds_created_at IS 'Index for ordering feeds by created_at';
COMMENT ON INDEX idx_feeds_type IS 'Index for filtering feeds by type';
COMMENT ON INDEX idx_feeds_location_id IS 'Index for feeds by location';

COMMENT ON INDEX idx_location_photos_location_id IS 'Index for photos by location';
COMMENT ON INDEX idx_location_photos_type IS 'Index for photos by type';

COMMENT ON INDEX idx_sync_state_form_id IS 'Index for sync state by form';
COMMENT ON INDEX idx_sync_state_status IS 'Index for sync state by status';