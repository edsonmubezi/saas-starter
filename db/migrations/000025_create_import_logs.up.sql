CREATE TABLE import_sessions (
    id              BIGSERIAL PRIMARY KEY,
    import_type     VARCHAR(50) NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'completed',
    total_rows      INT NOT NULL DEFAULT 0,
    successful      INT NOT NULL DEFAULT 0,
    failed          INT NOT NULL DEFAULT 0,
    error_file      TEXT,
    created_by      BIGINT NOT NULL REFERENCES users(id),
    created_by_name VARCHAR(200) NOT NULL,
    organization_id BIGINT NOT NULL REFERENCES organizations(id),
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_import_sessions_org ON import_sessions(organization_id);
CREATE INDEX idx_import_sessions_created ON import_sessions(created_at DESC);
CREATE INDEX idx_import_sessions_type ON import_sessions(import_type);

CREATE TABLE import_failed_records (
    id                 BIGSERIAL PRIMARY KEY,
    import_session_id  BIGINT NOT NULL REFERENCES import_sessions(id) ON DELETE CASCADE,
    row_number         INT NOT NULL,
    record_name        VARCHAR(300) NOT NULL DEFAULT '',
    reason             TEXT NOT NULL,
    record_type        VARCHAR(20) NOT NULL DEFAULT 'error',
    column_name        VARCHAR(100),
    value              VARCHAR(500),
    organization_id    BIGINT NOT NULL REFERENCES organizations(id)
);

CREATE INDEX idx_import_failed_records_session ON import_failed_records(import_session_id);
CREATE INDEX idx_import_failed_records_org ON import_failed_records(organization_id);
