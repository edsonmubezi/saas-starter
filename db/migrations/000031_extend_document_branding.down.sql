ALTER TABLE document_branding_settings
    DROP COLUMN IF EXISTS watermark_type,
    DROP COLUMN IF EXISTS watermark_image_path,
    DROP COLUMN IF EXISTS header_org_name,
    DROP COLUMN IF EXISTS header_address,
    DROP COLUMN IF EXISTS header_phone,
    DROP COLUMN IF EXISTS header_email,
    DROP COLUMN IF EXISTS header_tin,
    DROP COLUMN IF EXISTS footer_org_name;
