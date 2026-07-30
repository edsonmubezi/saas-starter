CREATE TABLE IF NOT EXISTS organization_email_configs (
    id BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    smtp_enabled BOOLEAN NOT NULL DEFAULT true,
    smtp_host VARCHAR(255) NOT NULL,
    smtp_port INTEGER NOT NULL DEFAULT 587,
    smtp_user VARCHAR(255) NOT NULL,
    smtp_password TEXT NOT NULL,
    from_address VARCHAR(255) NOT NULL,
    from_name VARCHAR(255) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(organization_id)
);

CREATE INDEX IF NOT EXISTS idx_org_email_configs_org_id ON organization_email_configs(organization_id);
