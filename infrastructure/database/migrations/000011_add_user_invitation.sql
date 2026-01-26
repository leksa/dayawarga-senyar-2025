-- ===========================================
-- User Invitation System
-- Migration: 000011_add_user_invitation.sql
-- ===========================================

-- ===========================================
-- ADD INVITATION FIELDS TO USERS TABLE
-- ===========================================
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS invitation_token VARCHAR(255),
    ADD COLUMN IF NOT EXISTS invitation_expires_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS invitation_sent_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS status VARCHAR(50) DEFAULT 'active';

-- Unique index for invitation token (only non-null values)
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_invitation_token
    ON users(invitation_token) WHERE invitation_token IS NOT NULL;

-- Index for finding pending invitations
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);

COMMENT ON COLUMN users.invitation_token IS 'Token for invitation link, null after activation';
COMMENT ON COLUMN users.invitation_expires_at IS 'When the invitation expires';
COMMENT ON COLUMN users.invitation_sent_at IS 'When invitation email was sent';
COMMENT ON COLUMN users.status IS 'User status: pending_invitation, active, suspended';

-- ===========================================
-- USER INVITATIONS LOG TABLE
-- Track invitation history
-- ===========================================
CREATE TABLE IF NOT EXISTS user_invitations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Invitation details
    email VARCHAR(255) NOT NULL,
    invited_to_org_id UUID REFERENCES organizations(id) ON DELETE SET NULL,
    invited_as_role VARCHAR(50) NOT NULL DEFAULT 'member',

    -- Inviter info
    invited_by UUID REFERENCES users(id) ON DELETE SET NULL,

    -- Status: sent, accepted, expired, cancelled
    status VARCHAR(50) NOT NULL DEFAULT 'sent',

    -- Timestamps
    sent_at TIMESTAMPTZ DEFAULT NOW(),
    accepted_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ NOT NULL,

    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_invitations_user ON user_invitations(user_id);
CREATE INDEX IF NOT EXISTS idx_user_invitations_email ON user_invitations(email);
CREATE INDEX IF NOT EXISTS idx_user_invitations_status ON user_invitations(status);
CREATE INDEX IF NOT EXISTS idx_user_invitations_org ON user_invitations(invited_to_org_id);

COMMENT ON TABLE user_invitations IS 'Log of all user invitations sent';

-- ===========================================
-- SUCCESS MESSAGE
-- ===========================================
DO $$
BEGIN
    RAISE NOTICE 'User invitation system created successfully!';
    RAISE NOTICE '- Added invitation_token, invitation_expires_at, invitation_sent_at, status to users';
    RAISE NOTICE '- Created user_invitations table for tracking';
END $$;
