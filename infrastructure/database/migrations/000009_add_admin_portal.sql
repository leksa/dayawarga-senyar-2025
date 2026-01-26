-- ===========================================
-- ADMIN PORTAL - Users, Organizations, Groups, Relawan
-- Migration: 000009_add_admin_portal.sql
-- ===========================================

-- ===========================================
-- USERS - Admin Portal users (from Authentik OIDC)
-- ===========================================
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- OIDC identity (from Authentik)
    oidc_subject VARCHAR(255) UNIQUE NOT NULL,  -- Authentik user ID (sub claim)
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),

    -- Profile
    avatar_url VARCHAR(500),

    -- Role (super_admin, org_admin, member)
    role VARCHAR(50) NOT NULL DEFAULT 'member',

    -- Status
    is_active BOOLEAN DEFAULT true,
    last_login_at TIMESTAMPTZ,

    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_oidc_subject ON users(oidc_subject);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

-- ===========================================
-- ORGANIZATIONS - NGOs, institutions managing relawan
-- ===========================================
CREATE TABLE IF NOT EXISTS organizations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Basic info
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,

    -- Contact
    email VARCHAR(255),
    phone VARCHAR(50),
    address TEXT,

    -- Visual
    logo_url VARCHAR(500),

    -- ODK Central integration
    odk_project_id INTEGER,  -- Optional: dedicated ODK project for this org

    -- Status
    is_active BOOLEAN DEFAULT true,

    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_organizations_slug ON organizations(slug);
CREATE INDEX IF NOT EXISTS idx_organizations_active ON organizations(is_active) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_organizations_deleted ON organizations(deleted_at) WHERE deleted_at IS NULL;

-- ===========================================
-- ORGANIZATION_MEMBERS - User membership in organizations
-- ===========================================
CREATE TABLE IF NOT EXISTS organization_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Role within organization (admin, member)
    role VARCHAR(50) NOT NULL DEFAULT 'member',

    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT NOW(),

    CONSTRAINT uq_org_member UNIQUE(organization_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_org_members_org ON organization_members(organization_id);
CREATE INDEX IF NOT EXISTS idx_org_members_user ON organization_members(user_id);

-- ===========================================
-- GROUPS - Teams within organizations
-- ===========================================
CREATE TABLE IF NOT EXISTS groups (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Basic info
    name VARCHAR(255) NOT NULL,
    description TEXT,

    -- Status
    is_active BOOLEAN DEFAULT true,

    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_groups_org ON groups(organization_id);
CREATE INDEX IF NOT EXISTS idx_groups_active ON groups(is_active) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_groups_deleted ON groups(deleted_at) WHERE deleted_at IS NULL;

-- ===========================================
-- RELAWAN - Field volunteers
-- ===========================================
CREATE TABLE IF NOT EXISTS relawan (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Organization and group assignment
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    group_id UUID REFERENCES groups(id) ON DELETE SET NULL,

    -- Basic info
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(50),
    email VARCHAR(255),

    -- ODK App User (created via ODK Central API)
    odk_app_user_id INTEGER,           -- ODK Central App User ID
    odk_app_user_token TEXT,           -- Encrypted token for QR code
    odk_app_user_created_at TIMESTAMPTZ,

    -- Assigned forms (array of form IDs)
    assigned_forms TEXT[],

    -- Status
    status VARCHAR(50) DEFAULT 'active',  -- active, inactive, suspended

    -- Notes
    notes TEXT,

    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_relawan_org ON relawan(organization_id);
CREATE INDEX IF NOT EXISTS idx_relawan_group ON relawan(group_id);
CREATE INDEX IF NOT EXISTS idx_relawan_status ON relawan(status);
CREATE INDEX IF NOT EXISTS idx_relawan_phone ON relawan(phone);
CREATE INDEX IF NOT EXISTS idx_relawan_deleted ON relawan(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_relawan_name_trgm ON relawan USING GIN(name gin_trgm_ops);

-- ===========================================
-- TRIGGERS: Auto-update updated_at
-- ===========================================
CREATE TRIGGER set_updated_at_users
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_updated_at_organizations
    BEFORE UPDATE ON organizations
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_updated_at_groups
    BEFORE UPDATE ON groups
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_updated_at_relawan
    BEFORE UPDATE ON relawan
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

-- ===========================================
-- SUCCESS MESSAGE
-- ===========================================
DO $$
BEGIN
    RAISE NOTICE 'Admin Portal tables created successfully!';
END $$;
