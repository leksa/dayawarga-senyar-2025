-- ===========================================
-- ODK Integration - Project Requests & Group Projects
-- Migration: 000010_add_odk_integration.sql
-- ===========================================

-- ===========================================
-- ADD GROUP LEADER TO GROUPS TABLE
-- ===========================================
ALTER TABLE groups
    ADD COLUMN IF NOT EXISTS leader_id UUID REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS odk_project_id INTEGER,
    ADD COLUMN IF NOT EXISTS odk_project_manager_created BOOLEAN DEFAULT false;

CREATE INDEX IF NOT EXISTS idx_groups_leader ON groups(leader_id);
CREATE INDEX IF NOT EXISTS idx_groups_odk_project ON groups(odk_project_id);

COMMENT ON COLUMN groups.leader_id IS 'User who leads this group, becomes ODK Project Manager on approval';
COMMENT ON COLUMN groups.odk_project_id IS 'Assigned ODK Central project ID';
COMMENT ON COLUMN groups.odk_project_manager_created IS 'Whether Project Manager was created in ODK Central';

-- ===========================================
-- PROJECT REQUESTS - Approval workflow for ODK project assignment
-- ===========================================
CREATE TABLE IF NOT EXISTS project_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- References
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,

    -- ODK Project requested
    odk_project_id INTEGER NOT NULL,
    odk_project_name VARCHAR(255),  -- Cached project name

    -- Request info
    requested_by UUID NOT NULL REFERENCES users(id),
    request_notes TEXT,

    -- Approval status: pending, approved, rejected
    status VARCHAR(50) NOT NULL DEFAULT 'pending',

    -- Approval info (filled when approved/rejected)
    reviewed_by UUID REFERENCES users(id),
    reviewed_at TIMESTAMPTZ,
    review_notes TEXT,

    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_project_requests_org ON project_requests(organization_id);
CREATE INDEX IF NOT EXISTS idx_project_requests_group ON project_requests(group_id);
CREATE INDEX IF NOT EXISTS idx_project_requests_status ON project_requests(status);
CREATE INDEX IF NOT EXISTS idx_project_requests_requested_by ON project_requests(requested_by);
CREATE INDEX IF NOT EXISTS idx_project_requests_pending ON project_requests(status) WHERE status = 'pending';

COMMENT ON TABLE project_requests IS 'Approval workflow for assigning ODK projects to groups';

-- ===========================================
-- GROUP_PROJECTS - Track approved group-project assignments
-- ===========================================
CREATE TABLE IF NOT EXISTS group_projects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    odk_project_id INTEGER NOT NULL,
    odk_project_name VARCHAR(255),

    -- Approval reference
    approved_request_id UUID REFERENCES project_requests(id),

    -- Timestamps
    assigned_at TIMESTAMPTZ DEFAULT NOW(),

    CONSTRAINT uq_group_project UNIQUE(group_id, odk_project_id)
);

CREATE INDEX IF NOT EXISTS idx_group_projects_group ON group_projects(group_id);
CREATE INDEX IF NOT EXISTS idx_group_projects_odk ON group_projects(odk_project_id);

COMMENT ON TABLE group_projects IS 'Approved ODK project assignments to groups';

-- ===========================================
-- ADD ODK WEB USER ID TO USERS TABLE
-- Track when a portal user becomes ODK Project Manager
-- ===========================================
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS odk_web_user_id INTEGER;

CREATE INDEX IF NOT EXISTS idx_users_odk_web_user ON users(odk_web_user_id) WHERE odk_web_user_id IS NOT NULL;

COMMENT ON COLUMN users.odk_web_user_id IS 'ODK Central Web User ID when user is a Project Manager';

-- ===========================================
-- TRIGGERS: Auto-update updated_at
-- ===========================================
CREATE TRIGGER set_updated_at_project_requests
    BEFORE UPDATE ON project_requests
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

-- ===========================================
-- SUCCESS MESSAGE
-- ===========================================
DO $$
BEGIN
    RAISE NOTICE 'ODK Integration tables created successfully!';
    RAISE NOTICE '- Added leader_id, odk_project_id, odk_project_manager_created to groups';
    RAISE NOTICE '- Added odk_web_user_id to users';
    RAISE NOTICE '- Created project_requests table (approval workflow)';
    RAISE NOTICE '- Created group_projects table (approved assignments)';
END $$;
