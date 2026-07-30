ALTER TABLE document_branding_settings
    ADD COLUMN IF NOT EXISTS watermark_type VARCHAR(10) DEFAULT 'text',
    ADD COLUMN IF NOT EXISTS watermark_image_path TEXT DEFAULT '',
    ADD COLUMN IF NOT EXISTS header_org_name TEXT DEFAULT '',
    ADD COLUMN IF NOT EXISTS header_address TEXT DEFAULT '',
    ADD COLUMN IF NOT EXISTS header_phone TEXT DEFAULT '',
    ADD COLUMN IF NOT EXISTS header_email TEXT DEFAULT '',
    ADD COLUMN IF NOT EXISTS header_tin TEXT DEFAULT '',
    ADD COLUMN IF NOT EXISTS footer_org_name TEXT DEFAULT '';
