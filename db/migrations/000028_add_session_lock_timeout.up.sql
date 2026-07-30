ALTER TABLE organization_settings
  ADD COLUMN session_lock_timeout_minutes INT NOT NULL DEFAULT 15;
