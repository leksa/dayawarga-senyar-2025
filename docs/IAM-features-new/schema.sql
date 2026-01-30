-- ============================================================================
-- ODK Admin Portal - Database Schema
-- PostgreSQL 16+
-- ============================================================================

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================================
-- ENUM TYPES
-- ============================================================================

CREATE TYPE user_system_role AS ENUM (
    'administrator',      -- Full system access
    'admin_lembaga',      -- Manage one lembaga
    'member'              -- Regular member (PM or Relawan assigned per sub-lembaga)
);

CREATE TYPE sub_lembaga_role AS ENUM (
    'project_manager',    -- Can manage relawan in sub-lembaga
    'relawan'             -- Can only submit forms
);

CREATE TYPE sync_job_status AS ENUM (
    'pending',
    'processing',
    'completed',
    'failed'
);

CREATE TYPE sync_job_type AS ENUM (
    'create_app_user',
    'revoke_app_user',
    'sync_projects',
    'sync_forms',
    'sync_submissions'
);

-- ============================================================================
-- CORE TABLES
-- ============================================================================

-- Users table (synced with Authentik)
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    authentik_id VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(50),
    system_role user_system_role NOT NULL DEFAULT 'member',
    lembaga_id UUID,  -- NULL for administrator, required for others
    avatar_url TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Lembaga (Organization) table
CREATE TABLE lembaga (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL,  -- Short code for identification
    description TEXT,
    address TEXT,
    logo_url TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_by UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Sub-Lembaga (maps to ODK Project)
CREATE TABLE sub_lembaga (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    lembaga_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- ODK Central mapping
    odk_project_id INTEGER NOT NULL,  -- ODK Central project ID
    odk_project_name VARCHAR(255) NOT NULL,
    
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_by UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Ensure one ODK project can only be linked once
    UNIQUE(odk_project_id)
);

-- Sub-Lembaga Members (assignment of users to sub-lembaga with roles)
CREATE TABLE sub_lembaga_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    sub_lembaga_id UUID NOT NULL,
    user_id UUID NOT NULL,
    role sub_lembaga_role NOT NULL DEFAULT 'relawan',
    
    -- ODK Central App User mapping
    odk_app_user_id INTEGER,          -- ODK Central app user ID (actorId)
    odk_app_user_token VARCHAR(255),  -- ODK app user token
    qr_code_data TEXT,                -- Cached QR code data (base64 or JSON)
    
    assigned_by UUID NOT NULL,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_active BOOLEAN NOT NULL DEFAULT true,
    revoked_at TIMESTAMPTZ,
    revoked_by UUID,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- One user can only be assigned once per sub-lembaga
    UNIQUE(sub_lembaga_id, user_id)
);

-- ============================================================================
-- ODK CACHE TABLES (for faster access, synced periodically)
-- ============================================================================

-- Cached ODK Projects
CREATE TABLE odk_projects_cache (
    id SERIAL PRIMARY KEY,
    odk_project_id INTEGER UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    archived BOOLEAN DEFAULT false,
    forms_count INTEGER DEFAULT 0,
    app_users_count INTEGER DEFAULT 0,
    last_submission_at TIMESTAMPTZ,
    synced_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Cached ODK Forms
CREATE TABLE odk_forms_cache (
    id SERIAL PRIMARY KEY,
    odk_project_id INTEGER NOT NULL,
    odk_form_id VARCHAR(255) NOT NULL,  -- xmlFormId
    name VARCHAR(255) NOT NULL,
    version VARCHAR(50),
    state VARCHAR(50),  -- open, closing, closed
    submissions_count INTEGER DEFAULT 0,
    synced_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE(odk_project_id, odk_form_id)
);

-- ============================================================================
-- SYNC & JOB TABLES
-- ============================================================================

-- Background sync jobs
CREATE TABLE sync_jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    job_type sync_job_type NOT NULL,
    entity_type VARCHAR(50),  -- 'sub_lembaga_member', 'sub_lembaga', etc.
    entity_id UUID,
    payload JSONB NOT NULL DEFAULT '{}',
    status sync_job_status NOT NULL DEFAULT 'pending',
    error_message TEXT,
    retry_count INTEGER NOT NULL DEFAULT 0,
    max_retries INTEGER NOT NULL DEFAULT 3,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    
    -- For job ordering
    priority INTEGER NOT NULL DEFAULT 0
);

-- ============================================================================
-- AUDIT LOG
-- ============================================================================

CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    actor_id UUID,  -- User who performed the action (NULL for system)
    actor_email VARCHAR(255),  -- Denormalized for historical record
    action VARCHAR(100) NOT NULL,  -- e.g., 'user.create', 'member.assign'
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID,
    old_value JSONB,
    new_value JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- DASHBOARD STATISTICS (materialized/cached)
-- ============================================================================

CREATE TABLE dashboard_stats (
    id SERIAL PRIMARY KEY,
    stat_date DATE NOT NULL DEFAULT CURRENT_DATE,
    lembaga_id UUID,  -- NULL for system-wide stats
    
    total_members INTEGER DEFAULT 0,
    active_members INTEGER DEFAULT 0,
    total_sub_lembaga INTEGER DEFAULT 0,
    total_submissions_today INTEGER DEFAULT 0,
    total_submissions_week INTEGER DEFAULT 0,
    total_submissions_month INTEGER DEFAULT 0,
    
    calculated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE(stat_date, lembaga_id)
);

-- ============================================================================
-- FOREIGN KEYS
-- ============================================================================

ALTER TABLE users 
    ADD CONSTRAINT fk_users_lembaga 
    FOREIGN KEY (lembaga_id) REFERENCES lembaga(id) ON DELETE SET NULL;

ALTER TABLE lembaga 
    ADD CONSTRAINT fk_lembaga_created_by 
    FOREIGN KEY (created_by) REFERENCES users(id);

ALTER TABLE sub_lembaga 
    ADD CONSTRAINT fk_sub_lembaga_lembaga 
    FOREIGN KEY (lembaga_id) REFERENCES lembaga(id) ON DELETE CASCADE;

ALTER TABLE sub_lembaga 
    ADD CONSTRAINT fk_sub_lembaga_created_by 
    FOREIGN KEY (created_by) REFERENCES users(id);

ALTER TABLE sub_lembaga_members 
    ADD CONSTRAINT fk_slm_sub_lembaga 
    FOREIGN KEY (sub_lembaga_id) REFERENCES sub_lembaga(id) ON DELETE CASCADE;

ALTER TABLE sub_lembaga_members 
    ADD CONSTRAINT fk_slm_user 
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE sub_lembaga_members 
    ADD CONSTRAINT fk_slm_assigned_by 
    FOREIGN KEY (assigned_by) REFERENCES users(id);

ALTER TABLE sub_lembaga_members 
    ADD CONSTRAINT fk_slm_revoked_by 
    FOREIGN KEY (revoked_by) REFERENCES users(id);

ALTER TABLE audit_logs 
    ADD CONSTRAINT fk_audit_actor 
    FOREIGN KEY (actor_id) REFERENCES users(id) ON DELETE SET NULL;

-- ============================================================================
-- INDEXES
-- ============================================================================

-- Users indexes
CREATE INDEX idx_users_authentik_id ON users(authentik_id);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_lembaga_id ON users(lembaga_id);
CREATE INDEX idx_users_system_role ON users(system_role);
CREATE INDEX idx_users_is_active ON users(is_active);

-- Lembaga indexes
CREATE INDEX idx_lembaga_code ON lembaga(code);
CREATE INDEX idx_lembaga_is_active ON lembaga(is_active);

-- Sub-lembaga indexes
CREATE INDEX idx_sub_lembaga_lembaga_id ON sub_lembaga(lembaga_id);
CREATE INDEX idx_sub_lembaga_odk_project_id ON sub_lembaga(odk_project_id);
CREATE INDEX idx_sub_lembaga_is_active ON sub_lembaga(is_active);

-- Sub-lembaga members indexes
CREATE INDEX idx_slm_sub_lembaga_id ON sub_lembaga_members(sub_lembaga_id);
CREATE INDEX idx_slm_user_id ON sub_lembaga_members(user_id);
CREATE INDEX idx_slm_role ON sub_lembaga_members(role);
CREATE INDEX idx_slm_is_active ON sub_lembaga_members(is_active);
CREATE INDEX idx_slm_odk_app_user_id ON sub_lembaga_members(odk_app_user_id);

-- ODK cache indexes
CREATE INDEX idx_odk_projects_cache_synced ON odk_projects_cache(synced_at);
CREATE INDEX idx_odk_forms_cache_project ON odk_forms_cache(odk_project_id);

-- Sync jobs indexes
CREATE INDEX idx_sync_jobs_status ON sync_jobs(status);
CREATE INDEX idx_sync_jobs_type ON sync_jobs(job_type);
CREATE INDEX idx_sync_jobs_created ON sync_jobs(created_at);
CREATE INDEX idx_sync_jobs_priority_status ON sync_jobs(priority DESC, status, created_at);

-- Audit logs indexes
CREATE INDEX idx_audit_logs_actor ON audit_logs(actor_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX idx_audit_logs_created ON audit_logs(created_at DESC);

-- Dashboard stats indexes
CREATE INDEX idx_dashboard_stats_date ON dashboard_stats(stat_date);
CREATE INDEX idx_dashboard_stats_lembaga ON dashboard_stats(lembaga_id);

-- ============================================================================
-- TRIGGERS
-- ============================================================================

-- Auto-update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_lembaga_updated_at
    BEFORE UPDATE ON lembaga
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_sub_lembaga_updated_at
    BEFORE UPDATE ON sub_lembaga
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_sub_lembaga_members_updated_at
    BEFORE UPDATE ON sub_lembaga_members
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- INITIAL DATA (Administrator user - to be updated with real Authentik ID)
-- ============================================================================

-- This will be created during first login via Authentik
-- INSERT INTO users (authentik_id, email, name, system_role)
-- VALUES ('authentik-admin-uuid', 'admin@example.com', 'System Administrator', 'administrator');

-- ============================================================================
-- VIEWS
-- ============================================================================

-- View for member assignments with full details
CREATE VIEW v_member_assignments AS
SELECT 
    slm.id,
    slm.sub_lembaga_id,
    slm.user_id,
    slm.role,
    slm.odk_app_user_id,
    slm.qr_code_data,
    slm.is_active,
    slm.assigned_at,
    u.name AS user_name,
    u.email AS user_email,
    u.phone AS user_phone,
    sl.name AS sub_lembaga_name,
    sl.odk_project_id,
    sl.odk_project_name,
    l.id AS lembaga_id,
    l.name AS lembaga_name,
    l.code AS lembaga_code
FROM sub_lembaga_members slm
JOIN users u ON slm.user_id = u.id
JOIN sub_lembaga sl ON slm.sub_lembaga_id = sl.id
JOIN lembaga l ON sl.lembaga_id = l.id;

-- View for lembaga statistics
CREATE VIEW v_lembaga_stats AS
SELECT 
    l.id,
    l.name,
    l.code,
    l.is_active,
    COUNT(DISTINCT u.id) FILTER (WHERE u.is_active) AS active_members,
    COUNT(DISTINCT sl.id) FILTER (WHERE sl.is_active) AS active_sub_lembaga,
    COUNT(DISTINCT slm.id) FILTER (WHERE slm.is_active) AS total_assignments
FROM lembaga l
LEFT JOIN users u ON u.lembaga_id = l.id
LEFT JOIN sub_lembaga sl ON sl.lembaga_id = l.id
LEFT JOIN sub_lembaga_members slm ON slm.sub_lembaga_id = sl.id
GROUP BY l.id, l.name, l.code, l.is_active;
