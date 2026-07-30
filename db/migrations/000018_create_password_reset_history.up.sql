-- Track password reset history for auditing
CREATE TABLE IF NOT EXISTS password_reset_history (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reset_by_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL, -- admin who initiated (null for self-reset)
    reset_method VARCHAR(20) NOT NULL, -- 'email', 'form', 'self'
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_password_reset_history_user ON password_reset_history(user_id);
CREATE INDEX idx_password_reset_history_reset_by ON password_reset_history(reset_by_user_id) WHERE reset_by_user_id IS NOT NULL;
CREATE INDEX idx_password_reset_history_date ON password_reset_history(created_at);
