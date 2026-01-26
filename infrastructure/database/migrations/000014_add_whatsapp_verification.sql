-- ===========================================
-- WHATSAPP VERIFICATION - Enable relawan to use WhatsApp chatbot
-- Migration: 000014_add_whatsapp_verification.sql
-- ===========================================

-- Add WhatsApp verification fields to relawan table
ALTER TABLE relawan
    ADD COLUMN IF NOT EXISTS wa_verified BOOLEAN DEFAULT false,
    ADD COLUMN IF NOT EXISTS wa_verified_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS wa_last_activity TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS wa_session_count INTEGER DEFAULT 0;

-- Index for quick lookup by phone + verification status (used by chatbot)
CREATE INDEX IF NOT EXISTS idx_relawan_wa_verified ON relawan(phone, wa_verified) WHERE wa_verified = true;

-- Index for activity tracking
CREATE INDEX IF NOT EXISTS idx_relawan_wa_activity ON relawan(wa_last_activity) WHERE wa_last_activity IS NOT NULL;

-- ===========================================
-- SUCCESS MESSAGE
-- ===========================================
DO $$
BEGIN
    RAISE NOTICE 'WhatsApp verification fields added to relawan table successfully!';
END $$;
