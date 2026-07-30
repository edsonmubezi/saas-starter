-- Create permissions table
CREATE TABLE IF NOT EXISTS permissions (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    visibility VARCHAR(10) NOT NULL DEFAULT 'ALL' CHECK (visibility IN ('ALL', 'HQ_ONLY')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_permissions_visibility ON permissions(visibility);
CREATE INDEX IF NOT EXISTS idx_permissions_name ON permissions(name);