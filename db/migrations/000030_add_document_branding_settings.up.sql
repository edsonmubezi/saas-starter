CREATE TABLE IF NOT EXISTS document_branding_settings (
    id BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL REFERENCES organizations(id),
    -- Header visibility toggles
    show_logo BOOLEAN NOT NULL DEFAULT true,
    show_org_name BOOLEAN NOT NULL DEFAULT true,
    show_address BOOLEAN NOT NULL DEFAULT true,
    show_contact BOOLEAN NOT NULL DEFAULT true,
    show_tin BOOLEAN NOT NULL DEFAULT false,
    show_reg_number BOOLEAN NOT NULL DEFAULT false,
    -- Footer
    show_footer BOOLEAN NOT NULL DEFAULT true,
    footer_text TEXT DEFAULT '',
    show_page_numbers BOOLEAN NOT NULL DEFAULT true,
    show_generated_date BOOLEAN NOT NULL DEFAULT true,
    -- Watermark
    show_watermark BOOLEAN NOT NULL DEFAULT false,
    watermark_text TEXT DEFAULT '',
    -- Colors
    primary_color VARCHAR(7) DEFAULT '#1a365d',
    header_text_color VARCHAR(7) DEFAULT '#FFFFFF',
    -- Font (gofpdf built-ins: Arial, Helvetica, Times, Courier)
    font_family VARCHAR(50) DEFAULT 'Arial',
    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(organization_id)
);
