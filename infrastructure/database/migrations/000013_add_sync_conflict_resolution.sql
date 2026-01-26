ALTER TABLE locations 
    ADD COLUMN IF NOT EXISTS source VARCHAR(20) DEFAULT 'odk',
    ADD COLUMN IF NOT EXISTS last_modified_by VARCHAR(100),
    ADD COLUMN IF NOT EXISTS linked_entity_id VARCHAR(100);

COMMENT ON COLUMN locations.source IS 'Data source: odk, whatsapp, api, manual';
COMMENT ON COLUMN locations.last_modified_by IS 'Identifier of last modifier (phone number, user ID, etc)';
COMMENT ON COLUMN locations.linked_entity_id IS 'Links WhatsApp-created records to ODK entity IDs for sync';

CREATE INDEX IF NOT EXISTS idx_locations_linked_entity 
    ON locations(linked_entity_id) 
    WHERE linked_entity_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_locations_source ON locations(source);

ALTER TABLE information_feeds 
    ADD COLUMN IF NOT EXISTS source VARCHAR(20) DEFAULT 'odk',
    ADD COLUMN IF NOT EXISTS last_modified_by VARCHAR(100);

COMMENT ON COLUMN information_feeds.source IS 'Data source: odk, whatsapp, api, manual';
COMMENT ON COLUMN information_feeds.last_modified_by IS 'Identifier of last modifier';

CREATE INDEX IF NOT EXISTS idx_feeds_source ON information_feeds(source);

CREATE TABLE IF NOT EXISTS sync_conflicts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID,
    entity_name VARCHAR(500),
    odk_entity_id VARCHAR(100),
    odk_submission_id VARCHAR(255),
    conflict_type VARCHAR(50) NOT NULL,
    existing_data JSONB,
    incoming_data JSONB,
    existing_updated_at TIMESTAMPTZ,
    incoming_submitted_at TIMESTAMPTZ,
    status VARCHAR(20) DEFAULT 'pending',
    resolved_by VARCHAR(100),
    resolved_at TIMESTAMPTZ,
    resolution_notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sync_conflicts_status ON sync_conflicts(status);
CREATE INDEX IF NOT EXISTS idx_sync_conflicts_entity ON sync_conflicts(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_sync_conflicts_created ON sync_conflicts(created_at DESC);

COMMENT ON TABLE sync_conflicts IS 'Logs sync conflicts between ODK and other data sources for manual resolution';

UPDATE locations SET source = 'odk' WHERE source IS NULL AND odk_submission_id IS NOT NULL;
UPDATE locations SET source = 'manual' WHERE source IS NULL AND odk_submission_id IS NULL;
UPDATE information_feeds SET source = 'odk' WHERE source IS NULL AND odk_submission_id IS NOT NULL;
UPDATE information_feeds SET source = 'manual' WHERE source IS NULL AND odk_submission_id IS NULL;
