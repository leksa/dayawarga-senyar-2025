-- ===========================================
-- Organization Enhancements - Location & Bidang
-- Migration: 000016_organization_enhancements.sql
-- ===========================================

-- ===========================================
-- ORGANIZATIONS - Add location and social media fields
-- ===========================================
ALTER TABLE organizations
    ADD COLUMN IF NOT EXISTS city VARCHAR(100),
    ADD COLUMN IF NOT EXISTS country VARCHAR(100),
    ADD COLUMN IF NOT EXISTS website_url VARCHAR(500),
    ADD COLUMN IF NOT EXISTS social_media JSONB DEFAULT '{}';

-- Indexes for location-based queries
CREATE INDEX IF NOT EXISTS idx_organizations_city ON organizations(city);
CREATE INDEX IF NOT EXISTS idx_organizations_country ON organizations(country);

-- Comments for new columns
COMMENT ON COLUMN organizations.city IS 'City where organization is located';
COMMENT ON COLUMN organizations.country IS 'Country where organization is located';
COMMENT ON COLUMN organizations.website_url IS 'Organization website URL';
COMMENT ON COLUMN organizations.social_media IS 'JSON object containing social media handles (facebook, twitter, instagram, linkedin, etc)';

-- ===========================================
-- BIDANG - Fields/sectors of work
-- ===========================================
CREATE TABLE IF NOT EXISTS bidang (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Basic info
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,

    -- Status
    is_active BOOLEAN DEFAULT true,

    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_bidang_slug ON bidang(slug);
CREATE INDEX IF NOT EXISTS idx_bidang_active ON bidang(is_active) WHERE is_active = true;

-- ===========================================
-- ORGANIZATION_BIDANG - Junction table for many-to-many relationship
-- ===========================================
CREATE TABLE IF NOT EXISTS organization_bidang (
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    bidang_id UUID NOT NULL REFERENCES bidang(id) ON DELETE CASCADE,

    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT NOW(),

    PRIMARY KEY (organization_id, bidang_id)
);

CREATE INDEX IF NOT EXISTS idx_organization_bidang_org ON organization_bidang(organization_id);
CREATE INDEX IF NOT EXISTS idx_organization_bidang_bidang ON organization_bidang(bidang_id);

-- ===========================================
-- SEED DATA - Bidang
-- ===========================================
INSERT INTO bidang (name, slug, description, is_active, created_at)
VALUES
    ('Kesehatan', 'kesehatan', 'Health and medical services', true, NOW()),
    ('Pendidikan', 'pendidikan', 'Education and training', true, NOW()),
    ('Logistik', 'logistik', 'Logistics and supply chain', true, NOW()),
    ('Pangan', 'pangan', 'Food security and nutrition', true, NOW()),
    ('Shelter', 'shelter', 'Housing and shelter', true, NOW()),
    ('WASH', 'wash', 'Water, sanitation and hygiene', true, NOW())
ON CONFLICT (slug) DO NOTHING;

-- ===========================================
-- SUCCESS MESSAGE
-- ===========================================
DO $$
BEGIN
    RAISE NOTICE 'Organization enhancements migration completed successfully!';
    RAISE NOTICE 'Added columns: city, country, website_url, social_media';
    RAISE NOTICE 'Created tables: bidang, organization_bidang';
    RAISE NOTICE 'Seeded 6 bidang entries';
END $$;
