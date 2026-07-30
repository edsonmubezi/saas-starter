-- Create audit schema for tamper-evident logging
CREATE SCHEMA IF NOT EXISTS audit;

-- Outbox table (transactional buffer)
-- This table is written in the same transaction as business logic
CREATE TABLE IF NOT EXISTS audit.outbox (
    id BIGSERIAL PRIMARY KEY,
    event_json JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    delivered_at TIMESTAMPTZ,
    retry_count INT NOT NULL DEFAULT 0,
    last_error TEXT
);

-- Index for undelivered events (worker polls this)
CREATE INDEX idx_outbox_undelivered ON audit.outbox (created_at)
WHERE delivered_at IS NULL;

-- Events table (permanent immutable audit log)
CREATE TABLE IF NOT EXISTS audit.events (
    audit_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Actor information (who did it)
    actor_type VARCHAR(20) NOT NULL CHECK (actor_type IN ('user', 'service', 'system')),
    actor_id BIGINT,
    tenant_id BIGINT NOT NULL,

    -- Action details (what was done)
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100) NOT NULL,
    resource_id BIGINT,

    -- Request context (where/when/how)
    request_id VARCHAR(100),
    ip VARCHAR(45),
    user_agent TEXT,
    origin_service VARCHAR(50) NOT NULL DEFAULT 'hrm-api',

    -- Severity classification
    severity VARCHAR(20) NOT NULL DEFAULT 'LOW'
        CHECK (severity IN ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL')),

    -- State changes (before/after snapshots)
    before_json JSONB,
    after_json JSONB,

    -- Versioning and integrity
    schema_version INT NOT NULL DEFAULT 1,
    sig_sha256 BYTEA NOT NULL
);

-- Primary indexes for audit queries
CREATE INDEX idx_audit_tenant_resource ON audit.events
    (tenant_id, resource_type, resource_id, occurred_at DESC);

CREATE INDEX idx_audit_action ON audit.events
    (action, occurred_at DESC);

CREATE INDEX idx_audit_actor ON audit.events
    (actor_id, occurred_at DESC) WHERE actor_id IS NOT NULL;

CREATE INDEX idx_audit_occurred ON audit.events
    (occurred_at DESC);

CREATE INDEX idx_audit_severity_high ON audit.events
    (severity, occurred_at DESC) WHERE severity IN ('HIGH', 'CRITICAL');

CREATE INDEX idx_audit_tenant_action ON audit.events
    (tenant_id, action, occurred_at DESC);

-- Dead letter queue (for failed events after max retries)
CREATE TABLE IF NOT EXISTS audit.dead_letter (
    id BIGSERIAL PRIMARY KEY,
    event_json JSONB NOT NULL,
    error_message TEXT NOT NULL,
    retry_count INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    moved_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Audit access log (audit-the-audit: who viewed audit logs)
CREATE TABLE IF NOT EXISTS audit.access_log (
    id BIGSERIAL PRIMARY KEY,
    accessed_by BIGINT NOT NULL,
    accessed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    query_filter JSONB,
    reason TEXT,
    ip VARCHAR(45),
    result_count INT
);

-- Add comments for documentation
COMMENT ON SCHEMA audit IS 'Tamper-evident audit logging with outbox pattern';
COMMENT ON TABLE audit.outbox IS 'Transactional outbox for reliable audit event delivery';
COMMENT ON TABLE audit.events IS 'Immutable audit trail with SHA-256 signatures';
COMMENT ON TABLE audit.dead_letter IS 'Failed audit events after exhausted retries';
COMMENT ON TABLE audit.access_log IS 'Log of who accessed audit logs and why (audit-the-audit)';

COMMENT ON COLUMN audit.events.sig_sha256 IS 'SHA-256 signature of canonical event representation for integrity verification';
COMMENT ON COLUMN audit.events.tenant_id IS 'Multi-tenant isolation: each tenant can only see their own events';
COMMENT ON COLUMN audit.events.severity IS 'LOW=informational, MEDIUM=normal ops, HIGH=security event, CRITICAL=breach attempt';
