-- ===========================================
-- DAYAWARGA SENYAR 2025 - Add baseline_sumber to locations
-- ===========================================

-- Add baseline_sumber column to locations table
ALTER TABLE locations ADD COLUMN IF NOT EXISTS baseline_sumber VARCHAR(100) DEFAULT 'BNPB';

-- Update all existing records to have 'BNPB' as baseline_sumber
UPDATE locations SET baseline_sumber = 'BNPB' WHERE baseline_sumber IS NULL;

-- Success message
DO $$
BEGIN
    RAISE NOTICE 'baseline_sumber column added to locations table!';
END $$;
