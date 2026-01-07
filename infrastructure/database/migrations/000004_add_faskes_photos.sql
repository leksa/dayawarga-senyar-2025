-- ===========================================
-- DAYAWARGA SENYAR 2025 - Add Faskes Photos
-- ===========================================

-- ===========================================
-- FASKES_PHOTOS - Photo attachments for faskes
-- ===========================================
CREATE TABLE IF NOT EXISTS faskes_photos (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    faskes_id UUID NOT NULL REFERENCES faskes(id) ON DELETE CASCADE,
    photo_type VARCHAR(50) NOT NULL DEFAULT 'foto_depan',
    filename VARCHAR(500) NOT NULL,
    storage_path VARCHAR(1000),
    is_cached BOOLEAN DEFAULT false,
    file_size INTEGER,
    created_at TIMESTAMPTZ DEFAULT NOW(),

    CONSTRAINT uq_faskes_photos UNIQUE(faskes_id, photo_type, filename)
);

CREATE INDEX IF NOT EXISTS idx_faskes_photos_faskes ON faskes_photos(faskes_id);

-- ===========================================
-- SUCCESS MESSAGE
-- ===========================================
DO $$
BEGIN
    RAISE NOTICE 'Faskes photos table created successfully!';
END $$;
