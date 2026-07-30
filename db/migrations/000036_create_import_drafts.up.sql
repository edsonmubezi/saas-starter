CREATE TABLE import_drafts (
    id              BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL REFERENCES organizations(id),
    mapping_id      BIGINT REFERENCES import_column_mappings(id),
    import_type     VARCHAR(50) NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    file_name       VARCHAR(500),
    total_rows      INT NOT NULL DEFAULT 0,
    valid_rows      INT NOT NULL DEFAULT 0,
    error_rows      INT NOT NULL DEFAULT 0,
    confirmed_rows  INT NOT NULL DEFAULT 0,
    payroll_month   DATE,
    created_by      BIGINT,
    created_by_name VARCHAR(255),
    confirmed_by    BIGINT,
    confirmed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (status IN ('pending', 'validated', 'confirmed', 'cancelled')),
    CHECK (import_type IN ('employee', 'payroll'))
);

CREATE INDEX idx_import_drafts_org ON import_drafts(organization_id, status);

CREATE TABLE import_draft_rows (
    id                BIGSERIAL PRIMARY KEY,
    draft_id          BIGINT NOT NULL REFERENCES import_drafts(id) ON DELETE CASCADE,
    row_number        INT NOT NULL,
    status            VARCHAR(20) NOT NULL DEFAULT 'pending',
    row_data          JSONB NOT NULL,
    validation_errors JSONB DEFAULT '[]',
    match_info        JSONB,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (status IN ('pending', 'valid', 'error', 'duplicate', 'confirmed', 'skipped'))
);

CREATE INDEX idx_idr_draft ON import_draft_rows(draft_id, status);
CREATE INDEX idx_idr_draft_num ON import_draft_rows(draft_id, row_number);
