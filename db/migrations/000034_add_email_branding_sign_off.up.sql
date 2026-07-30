ALTER TABLE email_branding_settings
    ADD COLUMN IF NOT EXISTS sign_off TEXT DEFAULT '';
