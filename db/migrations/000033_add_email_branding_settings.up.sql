CREATE TABLE IF NOT EXISTS email_branding_settings (
    id              BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL REFERENCES organizations(id),
    primary_color   VARCHAR(7) DEFAULT '#4F46E5',
    header_text_color VARCHAR(7) DEFAULT '#FFFFFF',
    accent_color    VARCHAR(7) DEFAULT '#4F46E5',
    font_family     VARCHAR(100) DEFAULT 'Arial, sans-serif',
    show_logo       BOOLEAN DEFAULT true,
    footer_text     TEXT DEFAULT '',
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(organization_id)
);
