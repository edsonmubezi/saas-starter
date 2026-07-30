-- Add security-related fields to users table for account lockout, 2FA, and tracking
ALTER TABLE users ADD COLUMN IF NOT EXISTS failed_login_attempts INT DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS locked_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS lock_reason VARCHAR(100);
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_login_ip VARCHAR(45);
ALTER TABLE users ADD COLUMN IF NOT EXISTS must_change_password BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS two_factor_enabled BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS two_factor_method VARCHAR(20); -- 'totp', 'email', 'sms'
ALTER TABLE users ADD COLUMN IF NOT EXISTS totp_secret VARCHAR(255); -- encrypted TOTP secret
ALTER TABLE users ADD COLUMN IF NOT EXISTS phone_number VARCHAR(20); -- for SMS OTP

-- Indexes for security queries
CREATE INDEX IF NOT EXISTS idx_users_locked_at ON users(locked_at) WHERE locked_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_two_factor ON users(two_factor_enabled) WHERE two_factor_enabled = TRUE;
