-- ===========================================
-- DAYAWARGA SENYAR 2025 - Statistik Bencana
-- Aggregated statistics per kecamatan, updated daily
-- ===========================================

-- ===========================================
-- STATISTIK_BENCANA - Aggregated stats per kecamatan
-- ===========================================
CREATE TABLE IF NOT EXISTS statistik_bencana (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Wilayah identifiers
    id_provinsi VARCHAR(20),
    id_kota_kab VARCHAR(20),
    id_kecamatan VARCHAR(20) NOT NULL,

    -- Wilayah names (denormalized for faster queries)
    nama_provinsi VARCHAR(255),
    nama_kota_kab VARCHAR(255),
    nama_kecamatan VARCHAR(255),

    -- Posko stats
    total_posko INTEGER DEFAULT 0,
    total_pengungsi INTEGER DEFAULT 0,
    total_kk INTEGER DEFAULT 0,

    -- Faskes stats
    total_faskes INTEGER DEFAULT 0,
    total_faskes_operasional INTEGER DEFAULT 0,

    -- Metadata
    calculated_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    -- Unique constraint per kecamatan
    CONSTRAINT uq_statistik_bencana_kecamatan UNIQUE(id_kecamatan)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_statistik_bencana_provinsi ON statistik_bencana(id_provinsi);
CREATE INDEX IF NOT EXISTS idx_statistik_bencana_kota_kab ON statistik_bencana(id_kota_kab);
CREATE INDEX IF NOT EXISTS idx_statistik_bencana_calculated ON statistik_bencana(calculated_at DESC);

-- Trigger for updated_at
CREATE TRIGGER set_updated_at_statistik_bencana
    BEFORE UPDATE ON statistik_bencana
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

-- ===========================================
-- SUCCESS MESSAGE
-- ===========================================
DO $$
BEGIN
    RAISE NOTICE 'Statistik bencana table created successfully!';
END $$;
