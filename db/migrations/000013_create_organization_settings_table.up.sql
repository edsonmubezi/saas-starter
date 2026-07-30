-- Create enum type for organization types
CREATE TYPE organization_type_enum AS ENUM (
    'single_company',
    'multiple_company',
    'multi_branch'
);

-- Create organization_settings table
CREATE TABLE IF NOT EXISTS organization_settings (
    id BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL UNIQUE REFERENCES organizations(id) ON DELETE CASCADE,
    organization_type organization_type_enum NOT NULL DEFAULT 'single_company',
    loan_module_enabled BOOLEAN NOT NULL DEFAULT true,
    self_service_enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create index for faster lookups
CREATE INDEX IF NOT EXISTS idx_org_settings_organization_id ON organization_settings(organization_id);

-- Add comment to table
COMMENT ON TABLE organization_settings IS 'Core organization settings for feature toggles and organization type configuration';
COMMENT ON COLUMN organization_settings.organization_type IS 'Type of organization: single_company, multiple_company, or multi_branch';
COMMENT ON COLUMN organization_settings.loan_module_enabled IS 'Enable/disable loan module functionality';
COMMENT ON COLUMN organization_settings.self_service_enabled IS 'Enable/disable employee self-service access';
