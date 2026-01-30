-- Add short_code column to information_feeds for shorter URLs
ALTER TABLE information_feeds ADD COLUMN IF NOT EXISTS short_code VARCHAR(12) UNIQUE;

-- Create index for fast lookup by short_code
CREATE INDEX IF NOT EXISTS idx_feeds_short_code ON information_feeds(short_code) WHERE short_code IS NOT NULL;

-- Generate short_code for existing feeds using base36 of unix timestamp + random suffix
UPDATE information_feeds 
SET short_code = LOWER(
    TO_HEX(EXTRACT(EPOCH FROM COALESCE(submitted_at, created_at))::bigint) || 
    SUBSTRING(REPLACE(id::text, '-', ''), 1, 4)
)
WHERE short_code IS NULL;
