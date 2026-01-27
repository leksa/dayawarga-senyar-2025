-- ===========================================
-- WhatsApp PIN Verification for User Invitation
-- Migration: 000015_add_whatsapp_pin_verification.sql
-- ===========================================

-- Add PIN verification fields to users table
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS verification_pin VARCHAR(6),
    ADD COLUMN IF NOT EXISTS verification_pin_expires_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS verification_phone VARCHAR(20),
    ADD COLUMN IF NOT EXISTS verified_at TIMESTAMPTZ;

-- Index for PIN lookup (only non-null values)
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_verification_pin
    ON users(verification_pin) WHERE verification_pin IS NOT NULL;

COMMENT ON COLUMN users.verification_pin IS '6-digit alphanumeric PIN for WhatsApp verification';
COMMENT ON COLUMN users.verification_pin_expires_at IS 'When the PIN expires (15 minutes)';
COMMENT ON COLUMN users.verification_phone IS 'Phone number used for verification';
COMMENT ON COLUMN users.verified_at IS 'When user completed WhatsApp verification';

-- Add new status for pending verification
-- Status flow: pending_invitation -> pending_verification -> active
COMMENT ON COLUMN users.status IS 'User status: pending_invitation, pending_verification, active, suspended';

DO $$
BEGIN
    RAISE NOTICE 'WhatsApp PIN verification fields added to users table';
END $$;
