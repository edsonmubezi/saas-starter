-- Create the two_factor_otp_codes table used by totp.go for email/SMS login OTP.
-- This is separate from the otp_codes table (000044) which has a different schema.
CREATE TABLE IF NOT EXISTS two_factor_otp_codes (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code VARCHAR(10) NOT NULL,
    method VARCHAR(20) NOT NULL,  -- 'email', 'sms'
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_2fa_otp_user ON two_factor_otp_codes(user_id);
CREATE INDEX IF NOT EXISTS idx_2fa_otp_expires ON two_factor_otp_codes(expires_at);
CREATE INDEX IF NOT EXISTS idx_2fa_otp_lookup ON two_factor_otp_codes(user_id, method, code) WHERE used_at IS NULL;
