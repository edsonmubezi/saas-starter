-- Drop the cleanup function
DROP FUNCTION IF EXISTS cleanup_expired_password_reset_tokens();

-- Drop indexes
DROP INDEX IF EXISTS idx_password_reset_tokens_used_at;
DROP INDEX IF EXISTS idx_password_reset_tokens_expires_at;
DROP INDEX IF EXISTS idx_password_reset_tokens_token_hash;
DROP INDEX IF EXISTS idx_password_reset_tokens_user_id;

-- Drop the password_reset_tokens table
DROP TABLE IF EXISTS password_reset_tokens;
