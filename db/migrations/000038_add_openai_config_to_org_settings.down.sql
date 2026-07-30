ALTER TABLE organization_settings
    DROP COLUMN IF EXISTS openai_api_key,
    DROP COLUMN IF EXISTS openai_model;
