-- ===========================================
-- DAYAWARGA SENYAR 2025 - Add Infrastruktur Table
-- For roads/bridges (Jalan/Jembatan) status tracking
-- ===========================================

-- ===========================================
-- INFRASTRUKTUR - Roads and bridges table
-- ===========================================
CREATE TABLE IF NOT EXISTS infrastruktur (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    odk_submission_id VARCHAR(255),
    entity_id VARCHAR(255) NOT NULL, -- Links to ODK entity (jembatan_entities)
    object_id VARCHAR(100), -- Original BNPB/PU object ID

    -- Frequently queried fields (dedicated columns)
    nama VARCHAR(500) NOT NULL,
    jenis VARCHAR(50) NOT NULL, -- Jalan, Jembatan
    status_jln VARCHAR(100), -- Nasional, Provinsi, Kabupaten/Kota, Desa

    -- Location
    nama_provinsi VARCHAR(255),
    nama_kabupaten VARCHAR(255),

    -- Spatial data
    geom GEOMETRY(Point, 4326),

    -- Status fields (updateable via ODK)
    status_akses VARCHAR(100), -- dapat_diakses, akses_terputus
    keterangan_bencana TEXT,
    dampak TEXT,
    status_penanganan VARCHAR(100), -- belum_ditangani, sedang_ditangani, selesai
    penanganan_detail TEXT,
    bailey VARCHAR(100),
    progress INTEGER DEFAULT 0, -- 0-100%
    target_selesai VARCHAR(255),

    -- Source tracking
    baseline_sumber VARCHAR(100) DEFAULT 'BNPB/PU',
    update_by VARCHAR(255),

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

-- Unique constraint: one record per entity (latest submission wins)
CREATE UNIQUE INDEX IF NOT EXISTS idx_infrastruktur_entity ON infrastruktur(entity_id) WHERE deleted_at IS NULL;

-- Spatial index
CREATE INDEX IF NOT EXISTS idx_infrastruktur_geom ON infrastruktur USING GIST(geom);

-- Common filters
CREATE INDEX IF NOT EXISTS idx_infrastruktur_jenis ON infrastruktur(jenis);
CREATE INDEX IF NOT EXISTS idx_infrastruktur_status_jln ON infrastruktur(status_jln);
CREATE INDEX IF NOT EXISTS idx_infrastruktur_status_akses ON infrastruktur(status_akses);
CREATE INDEX IF NOT EXISTS idx_infrastruktur_status_penanganan ON infrastruktur(status_penanganan);
CREATE INDEX IF NOT EXISTS idx_infrastruktur_kabupaten ON infrastruktur(nama_kabupaten);
CREATE INDEX IF NOT EXISTS idx_infrastruktur_deleted ON infrastruktur(deleted_at) WHERE deleted_at IS NULL;

-- Full-text search
CREATE INDEX IF NOT EXISTS idx_infrastruktur_nama_trgm ON infrastruktur USING GIN(nama gin_trgm_ops);

-- ===========================================
-- INFRASTRUKTUR_PHOTOS - Photo attachments
-- ===========================================
CREATE TABLE IF NOT EXISTS infrastruktur_photos (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    infrastruktur_id UUID NOT NULL REFERENCES infrastruktur(id) ON DELETE CASCADE,
    photo_type VARCHAR(50) NOT NULL, -- foto_1, foto_2, foto_3, foto_4
    filename VARCHAR(500) NOT NULL,
    storage_path VARCHAR(1000),
    is_cached BOOLEAN DEFAULT false,
    file_size INTEGER,
    created_at TIMESTAMPTZ DEFAULT NOW(),

    CONSTRAINT uq_infrastruktur_photos UNIQUE(infrastruktur_id, photo_type, filename)
);

CREATE INDEX IF NOT EXISTS idx_infrastruktur_photos_infra ON infrastruktur_photos(infrastruktur_id);

-- ===========================================
-- TRIGGER: Auto-update updated_at for infrastruktur
-- ===========================================
CREATE TRIGGER set_updated_at_infrastruktur
    BEFORE UPDATE ON infrastruktur
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

-- ===========================================
-- SUCCESS MESSAGE
-- ===========================================
DO $$
BEGIN
    RAISE NOTICE 'Infrastruktur tables created successfully!';
END $$;
