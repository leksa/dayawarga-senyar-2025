-- Migration: Add ODK App User fields to users table
-- Description: Adds support for ODK Collect (App User) access for admin portal users
-- When a user is assigned as Project Manager, they automatically get an App User created
-- which allows them to use ODK Collect mobile app

-- Add ODK App User columns
ALTER TABLE users
ADD COLUMN IF NOT EXISTS odk_app_user_id INTEGER,
ADD COLUMN IF NOT EXISTS odk_app_user_token TEXT,
ADD COLUMN IF NOT EXISTS odk_app_user_project_id INTEGER;

-- Add comments for documentation
COMMENT ON COLUMN users.odk_app_user_id IS 'ODK Central App User ID for ODK Collect access';
COMMENT ON COLUMN users.odk_app_user_token IS 'ODK App User token for QR code generation (encrypted at rest)';
COMMENT ON COLUMN users.odk_app_user_project_id IS 'ODK Project ID where the App User was created';

-- Create index for lookup by app user id
CREATE INDEX IF NOT EXISTS idx_users_odk_app_user_id ON users(odk_app_user_id) WHERE odk_app_user_id IS NOT NULL;
