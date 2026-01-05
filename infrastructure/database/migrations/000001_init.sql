-- ===========================================
-- DAYAWARGA SENYAR 2025 - Database Init
-- ===========================================

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "postgis";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- ===========================================
-- LOCATIONS - Main table for posko/shelters
-- ===========================================
CREATE TABLE IF NOT EXISTS locations (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    odk_submission_id VARCHAR(255) UNIQUE,

    -- Frequently queried fields (dedicated columns)
    nama VARCHAR(500) NOT NULL,
    type VARCHAR(50) NOT NULL DEFAULT 'posko',
    status VARCHAR(50) DEFAULT 'operational',

    -- Spatial data
    geom GEOMETRY(Point, 4326),
    geo_meta JSONB DEFAULT '{}',

    -- Grouped data as JSONB
    identitas JSONB DEFAULT '{}',
    alamat JSONB DEFAULT '{}',
    data_pengungsi JSONB DEFAULT '{}',
    fasilitas JSONB DEFAULT '{}',
    komunikasi JSONB DEFAULT '{}',
    akses JSONB DEFAULT '{}',

    -- Complete raw data for reference
    raw_data JSONB,

    -- ODK metadata
    submitter_name VARCHAR(255),
    submitted_at TIMESTAMPTZ,

    -- System timestamps
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    synced_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Spatial index
CREATE INDEX IF NOT EXISTS idx_locations_geom ON locations USING GIST(geom);

-- Common filters
CREATE INDEX IF NOT EXISTS idx_locations_type ON locations(type);
CREATE INDEX IF NOT EXISTS idx_locations_status ON locations(status);
CREATE INDEX IF NOT EXISTS idx_locations_deleted ON locations(deleted_at) WHERE deleted_at IS NULL;

-- Full-text search
CREATE INDEX IF NOT EXISTS idx_locations_nama_trgm ON locations USING GIN(nama gin_trgm_ops);

-- ===========================================
-- LOCATION_PHOTOS - Photo attachments
-- ===========================================
CREATE TABLE IF NOT EXISTS location_photos (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    location_id UUID NOT NULL REFERENCES locations(id) ON DELETE CASCADE,
    photo_type VARCHAR(50) NOT NULL,
    filename VARCHAR(500) NOT NULL,
    storage_path VARCHAR(1000),
    is_cached BOOLEAN DEFAULT false,
    file_size INTEGER,
    created_at TIMESTAMPTZ DEFAULT NOW(),

    CONSTRAINT uq_location_photos UNIQUE(location_id, photo_type, filename)
);

CREATE INDEX IF NOT EXISTS idx_location_photos_location ON location_photos(location_id);

-- ===========================================
-- INFORMATION_FEEDS - Updates from field
-- ===========================================
CREATE TABLE IF NOT EXISTS information_feeds (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    location_id UUID REFERENCES locations(id) ON DELETE SET NULL,
    odk_submission_id VARCHAR(255) UNIQUE,

    content TEXT NOT NULL,
    category VARCHAR(50) DEFAULT 'informasi',
    type VARCHAR(50),
    username VARCHAR(255),
    organization VARCHAR(255),

    geom GEOMETRY(Point, 4326),

    raw_data JSONB,

    submitted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_feeds_location ON information_feeds(location_id);
CREATE INDEX IF NOT EXISTS idx_feeds_submitted ON information_feeds(submitted_at DESC);
CREATE INDEX IF NOT EXISTS idx_feeds_geom ON information_feeds USING GIST(geom) WHERE geom IS NOT NULL;

-- ===========================================
-- SYNC_STATE - Track sync progress
-- ===========================================
CREATE TABLE IF NOT EXISTS sync_state (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    form_id VARCHAR(255) UNIQUE NOT NULL,
    last_sync_timestamp TIMESTAMPTZ,
    last_etag VARCHAR(500),
    last_record_count INTEGER DEFAULT 0,
    total_records INTEGER DEFAULT 0,
    status VARCHAR(20) DEFAULT 'idle',
    error_message TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- ===========================================
-- TRIGGER: Auto-update updated_at
-- ===========================================
CREATE OR REPLACE FUNCTION trigger_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_updated_at_locations
    BEFORE UPDATE ON locations
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_updated_at_feeds
    BEFORE UPDATE ON information_feeds
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_updated_at_sync_state
    BEFORE UPDATE ON sync_state
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

-- ===========================================
-- SUCCESS MESSAGE
-- ===========================================
DO $$
BEGIN
    RAISE NOTICE 'Database initialized successfully!';
END $$;
