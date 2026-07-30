-- Platform schema: application logs, access logs, security events
-- These tables are auto-created by ensurePlatformTables() on startup but
-- having them as migrations ensures they exist before the first request.

CREATE SCHEMA IF NOT EXISTS platform;

CREATE TABLE IF NOT EXISTS platform.application_logs (
    id          BIGSERIAL PRIMARY KEY,
    timestamp   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    level       VARCHAR(10) NOT NULL DEFAULT 'INFO',
    category    VARCHAR(20) NOT NULL DEFAULT 'system',
    message     TEXT NOT NULL,
    service     VARCHAR(50) NOT NULL DEFAULT 'saas-api',
    trace_id    VARCHAR(100),
    span_id     VARCHAR(100),
    tenant_id   BIGINT,
    user_id     BIGINT,
    request_id  VARCHAR(100),
    method      VARCHAR(10),
    path        TEXT,
    status_code INT,
    duration_ns BIGINT,
    ip_address  VARCHAR(45),
    user_agent  TEXT,
    error_type  VARCHAR(255),
    error_msg   TEXT,
    error_stack TEXT,
    error_code  VARCHAR(50),
    fields      JSONB,
    stack       TEXT
);

CREATE TABLE IF NOT EXISTS platform.access_logs (
    id              BIGSERIAL PRIMARY KEY,
    timestamp       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    method          VARCHAR(10) NOT NULL,
    path            TEXT NOT NULL,
    query           TEXT,
    status_code     INT NOT NULL,
    duration_ns     BIGINT NOT NULL,
    request_size    BIGINT,
    response_size   BIGINT,
    ip_address      VARCHAR(45) NOT NULL,
    user_agent      TEXT,
    referer         TEXT,
    tenant_id       BIGINT,
    user_id         BIGINT,
    request_id      VARCHAR(100),
    trace_id        VARCHAR(100),
    span_id         VARCHAR(100),
    error           TEXT,
    content_type    VARCHAR(100),
    protocol        VARCHAR(20),
    tls_version     VARCHAR(20),
    geo_location    JSONB
);

CREATE TABLE IF NOT EXISTS platform.security_events (
    id                BIGSERIAL PRIMARY KEY,
    timestamp         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    tenant_id         BIGINT,
    severity          VARCHAR(20) NOT NULL DEFAULT 'INFO',
    category          VARCHAR(30) NOT NULL DEFAULT 'authentication',
    event_type        VARCHAR(60) NOT NULL,
    actor_id          BIGINT,
    actor_email       VARCHAR(255),
    ip_address        VARCHAR(45),
    user_agent        TEXT,
    geo_location      JSONB,
    details           JSONB,
    threat_indicators TEXT[],
    alert_sent        BOOLEAN NOT NULL DEFAULT FALSE,
    alert_channels    TEXT[],
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for application logs
CREATE INDEX IF NOT EXISTS idx_app_logs_tenant_ts   ON platform.application_logs (tenant_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_app_logs_level_ts    ON platform.application_logs (level, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_app_logs_category_ts ON platform.application_logs (category, timestamp DESC);

-- Indexes for access logs
CREATE INDEX IF NOT EXISTS idx_access_logs_tenant_ts ON platform.access_logs (tenant_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_access_logs_status_ts ON platform.access_logs (status_code, timestamp DESC);

-- Indexes for security events
CREATE INDEX IF NOT EXISTS idx_sec_events_tenant_ts   ON platform.security_events (tenant_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_sec_events_type_ts     ON platform.security_events (event_type, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_sec_events_severity_ts ON platform.security_events (severity, timestamp DESC);
