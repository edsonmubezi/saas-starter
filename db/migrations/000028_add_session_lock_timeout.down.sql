ALTER TABLE organization_settings
  DROP COLUMN IF EXISTS session_lock_timeout_minutes;
