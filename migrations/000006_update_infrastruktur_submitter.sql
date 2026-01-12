-- Migration: Update infrastruktur submitter names to bot_DayaWarga
-- This updates records imported via script to show proper bot name

-- Update update_by field
UPDATE infrastruktur
SET update_by = 'bot_DayaWarga'
WHERE update_by IN ('Fakhrizal', 'data-import', 'Rizal Dasi');

-- Update submitter_name field
UPDATE infrastruktur
SET submitter_name = 'bot_DayaWarga'
WHERE submitter_name IN ('Fakhrizal', 'data-import', 'Rizal Dasi');

-- Verify the update
SELECT update_by, submitter_name, COUNT(*) as count
FROM infrastruktur
GROUP BY update_by, submitter_name;
