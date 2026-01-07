-- ===========================================
-- DAYAWARGA SENYAR 2025 - Add Feed Photos and Faskes Reference
-- ===========================================

-- ===========================================
-- Add faskes_id column to information_feeds
-- ===========================================
ALTER TABLE information_feeds
ADD COLUMN IF NOT EXISTS faskes_id UUID REFERENCES faskes(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_feeds_faskes ON information_feeds(faskes_id);

-- ===========================================
-- FEED_PHOTOS - Photo attachments for feeds
-- ===========================================
CREATE TABLE IF NOT EXISTS feed_photos (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    feed_id UUID NOT NULL REFERENCES information_feeds(id) ON DELETE CASCADE,
    photo_type VARCHAR(50) NOT NULL DEFAULT 'foto',
    filename VARCHAR(500) NOT NULL,
    storage_path VARCHAR(1000),
    is_cached BOOLEAN DEFAULT false,
    file_size INTEGER,
    created_at TIMESTAMPTZ DEFAULT NOW(),

    CONSTRAINT uq_feed_photos UNIQUE(feed_id, photo_type, filename)
);

CREATE INDEX IF NOT EXISTS idx_feed_photos_feed ON feed_photos(feed_id);

-- ===========================================
-- SUCCESS MESSAGE
-- ===========================================
DO $$
BEGIN
    RAISE NOTICE 'Feed photos table created successfully!';
END $$;
