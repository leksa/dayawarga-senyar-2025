-- ===========================================
-- DAYAWARGA SENYAR 2025 - Add Faskes Table
-- ===========================================

-- ===========================================
-- FASKES - Health facilities table
-- ===========================================
CREATE TABLE IF NOT EXISTS faskes (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    odk_submission_id VARCHAR(255) UNIQUE,

    -- Frequently queried fields (dedicated columns)
    nama VARCHAR(500) NOT NULL,
    jenis_faskes VARCHAR(50), -- rumah_sakit, puskesmas, klinik, posko_kes_darurat
    status_faskes VARCHAR(50) DEFAULT 'operasional', -- operasional, non_aktif
    kondisi_faskes VARCHAR(50), -- tidak_rusak, rusak_ringan, rusak_sedang, rusak_berat, hancur_total

    -- Spatial data
    geom GEOMETRY(Point, 4326),

    -- Grouped data as JSONB
    alamat JSONB DEFAULT '{}',
    identitas JSONB DEFAULT '{}',
    isolasi JSONB DEFAULT '{}',
    infrastruktur JSONB DEFAULT '{}',
    sdm JSONB DEFAULT '{}',
    perbekalan JSONB DEFAULT '{}',
    klaster JSONB DEFAULT '{}',

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
CREATE INDEX IF NOT EXISTS idx_faskes_geom ON faskes USING GIST(geom);

-- Common filters
CREATE INDEX IF NOT EXISTS idx_faskes_jenis ON faskes(jenis_faskes);
CREATE INDEX IF NOT EXISTS idx_faskes_status ON faskes(status_faskes);
CREATE INDEX IF NOT EXISTS idx_faskes_kondisi ON faskes(kondisi_faskes);
CREATE INDEX IF NOT EXISTS idx_faskes_deleted ON faskes(deleted_at) WHERE deleted_at IS NULL;

-- Full-text search
CREATE INDEX IF NOT EXISTS idx_faskes_nama_trgm ON faskes USING GIN(nama gin_trgm_ops);

-- ===========================================
-- FASKES_PHOTOS - Photo attachments for faskes
-- ===========================================
CREATE TABLE IF NOT EXISTS faskes_photos (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    faskes_id UUID NOT NULL REFERENCES faskes(id) ON DELETE CASCADE,
    photo_type VARCHAR(50) NOT NULL,
    filename VARCHAR(500) NOT NULL,
    storage_path VARCHAR(1000),
    is_cached BOOLEAN DEFAULT false,
    file_size INTEGER,
    created_at TIMESTAMPTZ DEFAULT NOW(),

    CONSTRAINT uq_faskes_photos UNIQUE(faskes_id, photo_type, filename)
);

CREATE INDEX IF NOT EXISTS idx_faskes_photos_faskes ON faskes_photos(faskes_id);

-- ===========================================
-- TRIGGER: Auto-update updated_at for faskes
-- ===========================================
CREATE TRIGGER set_updated_at_faskes
    BEFORE UPDATE ON faskes
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

-- ===========================================
-- SUCCESS MESSAGE
-- ===========================================
DO $$
BEGIN
    RAISE NOTICE 'Faskes tables created successfully!';
END $$;
