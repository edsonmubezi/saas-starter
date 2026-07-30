-- Track all login attempts for security auditing
CREATE TABLE IF NOT EXISTS login_attempts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE, -- null if user not found
    email VARCHAR(255) NOT NULL,
    ip_address VARCHAR(45) NOT NULL,
    user_agent TEXT,
    success BOOLEAN NOT NULL,
    failure_reason VARCHAR(100), -- 'invalid_password', 'account_locked', 'account_inactive', '2fa_failed', 'user_not_found'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for efficient querying
CREATE INDEX idx_login_attempts_user ON login_attempts(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX idx_login_attempts_email ON login_attempts(email);
CREATE INDEX idx_login_attempts_ip ON login_attempts(ip_address);
CREATE INDEX idx_login_attempts_date ON login_attempts(created_at);
CREATE INDEX idx_login_attempts_success ON login_attempts(success);

-- Cleanup function for old login attempts (keep last 90 days)
CREATE OR REPLACE FUNCTION cleanup_old_login_attempts() RETURNS void AS $$
BEGIN
    DELETE FROM login_attempts WHERE created_at < NOW() - INTERVAL '90 days';
END;
$$ LANGUAGE plpgsql;
