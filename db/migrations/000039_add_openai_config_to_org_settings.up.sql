ALTER TABLE organization_settings
    ADD COLUMN IF NOT EXISTS openai_api_key TEXT DEFAULT '',
    ADD COLUMN IF NOT EXISTS openai_model VARCHAR(50) DEFAULT 'gpt-4o-mini';
