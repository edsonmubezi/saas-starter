-- Two-factor authentication backup codes
CREATE TABLE IF NOT EXISTS two_factor_backup_codes (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code_hash VARCHAR(255) NOT NULL, -- bcrypt hash of backup code
    used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_2fa_backup_user ON two_factor_backup_codes(user_id);
CREATE INDEX idx_2fa_backup_unused ON two_factor_backup_codes(user_id) WHERE used_at IS NULL;

-- OTP codes for email/SMS verification
CREATE TABLE IF NOT EXISTS otp_codes (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code_hash VARCHAR(255) NOT NULL, -- bcrypt hash of OTP
    type VARCHAR(20) NOT NULL, -- 'email', 'sms', 'login_2fa'
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used_at TIMESTAMP WITH TIME ZONE,
    ip_address VARCHAR(45),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_otp_user ON otp_codes(user_id);
CREATE INDEX idx_otp_expires ON otp_codes(expires_at);
CREATE INDEX idx_otp_type ON otp_codes(type);

-- Cleanup function for expired OTPs
CREATE OR REPLACE FUNCTION cleanup_expired_otps() RETURNS void AS $$
BEGIN
    DELETE FROM otp_codes WHERE expires_at < NOW() - INTERVAL '1 day';
END;
$$ LANGUAGE plpgsql;
