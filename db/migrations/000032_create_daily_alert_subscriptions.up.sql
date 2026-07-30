-- Daily alert subscriptions: per-user, per-alert-type notification preferences
CREATE TABLE IF NOT EXISTS daily_alert_subscriptions (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id BIGINT NOT NULL,
    alert_type      VARCHAR(50) NOT NULL,
    enabled         BOOLEAN NOT NULL DEFAULT TRUE,
    company_ids     BIGINT[] DEFAULT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_daily_alert_sub UNIQUE (user_id, organization_id, alert_type)
);

CREATE INDEX IF NOT EXISTS idx_daily_alert_sub_org
    ON daily_alert_subscriptions(organization_id);

CREATE INDEX IF NOT EXISTS idx_daily_alert_sub_enabled
    ON daily_alert_subscriptions(organization_id, enabled) WHERE enabled = TRUE;

-- Permission seeding handled by internal/seeder/permissions.go
-- Role assignment handled by internal/seeder/role_permissions.go
