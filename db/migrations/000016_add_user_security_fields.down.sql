-- Remove security-related fields from users table
DROP INDEX IF EXISTS idx_users_two_factor;
DROP INDEX IF EXISTS idx_users_locked_at;

ALTER TABLE users DROP COLUMN IF EXISTS phone_number;
ALTER TABLE users DROP COLUMN IF EXISTS totp_secret;
ALTER TABLE users DROP COLUMN IF EXISTS two_factor_method;
ALTER TABLE users DROP COLUMN IF EXISTS two_factor_enabled;
ALTER TABLE users DROP COLUMN IF EXISTS must_change_password;
ALTER TABLE users DROP COLUMN IF EXISTS last_login_ip;
ALTER TABLE users DROP COLUMN IF EXISTS last_login_at;
ALTER TABLE users DROP COLUMN IF EXISTS lock_reason;
ALTER TABLE users DROP COLUMN IF EXISTS locked_at;
ALTER TABLE users DROP COLUMN IF EXISTS failed_login_attempts;
