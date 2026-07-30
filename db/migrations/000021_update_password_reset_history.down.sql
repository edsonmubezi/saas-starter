-- Revert password_reset_history table changes
ALTER TABLE password_reset_history
DROP COLUMN IF EXISTS initiated_by_id,
DROP COLUMN IF EXISTS status,
DROP COLUMN IF EXISTS form_attached,
DROP COLUMN IF EXISTS form_url,
DROP COLUMN IF EXISTS notes,
DROP COLUMN IF EXISTS token_expires_at,
DROP COLUMN IF EXISTS completed_at,
DROP COLUMN IF EXISTS updated_at,
DROP COLUMN IF EXISTS organization_id,
DROP COLUMN IF EXISTS temporary_password_generated,
DROP COLUMN IF EXISTS email_sent,
DROP COLUMN IF EXISTS pdf_generated;

-- Restore the old reset_by_user_id column
ALTER TABLE password_reset_history
ADD COLUMN IF NOT EXISTS reset_by_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL;

-- Drop indexes
DROP INDEX IF EXISTS idx_password_reset_history_initiated_by;
DROP INDEX IF EXISTS idx_password_reset_history_org;
DROP INDEX IF EXISTS idx_password_reset_history_status;
