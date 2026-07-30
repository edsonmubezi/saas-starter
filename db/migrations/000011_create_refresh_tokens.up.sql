-- Create refresh_tokens table for managing JWT refresh token rotation
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_fingerprint VARCHAR(255) UNIQUE NOT NULL,
    issued_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    revoked_at TIMESTAMP WITH TIME ZONE,
    replaced_by_token_id BIGINT REFERENCES refresh_tokens(id) ON DELETE SET NULL,
    device_info TEXT,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token_fingerprint ON refresh_tokens(token_fingerprint);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_revoked_at ON refresh_tokens(revoked_at);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_issued_at ON refresh_tokens(issued_at);

-- Create a function to automatically clean up old tokens (keep last 30 days)
CREATE OR REPLACE FUNCTION cleanup_old_refresh_tokens()
RETURNS void AS $$
BEGIN
    DELETE FROM refresh_tokens
    WHERE (expires_at < NOW() - INTERVAL '30 days')
       OR (revoked_at IS NOT NULL AND revoked_at < NOW() - INTERVAL '30 days');
END;
$$ LANGUAGE plpgsql;

-- Add comments for documentation
COMMENT ON TABLE refresh_tokens IS 'Stores refresh token fingerprints for rotation and revocation tracking';
COMMENT ON COLUMN refresh_tokens.token_fingerprint IS 'SHA-256 hash of the refresh token (never store plain tokens)';
COMMENT ON COLUMN refresh_tokens.expires_at IS 'Token expiration timestamp (typically 7 days from issuance)';
COMMENT ON COLUMN refresh_tokens.revoked_at IS 'Timestamp when token was revoked (password change, logout, etc.)';
COMMENT ON COLUMN refresh_tokens.replaced_by_token_id IS 'Reference to the new token that replaced this one during rotation';
