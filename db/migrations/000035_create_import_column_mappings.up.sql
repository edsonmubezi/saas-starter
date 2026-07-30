CREATE TABLE import_column_mappings (
    id              BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL REFERENCES organizations(id),
    import_type     VARCHAR(50) NOT NULL,
    name            VARCHAR(255) NOT NULL,
    columns         JSONB NOT NULL DEFAULT '[]',
    is_default      BOOLEAN NOT NULL DEFAULT false,
    delete_status   INT NOT NULL DEFAULT 0,
    created_by      BIGINT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (import_type IN ('employee', 'payroll'))
);

CREATE INDEX idx_icm_org ON import_column_mappings(organization_id, import_type) WHERE delete_status = 0;
CREATE UNIQUE INDEX idx_icm_default ON import_column_mappings(organization_id, import_type) WHERE is_default = true AND delete_status = 0;
