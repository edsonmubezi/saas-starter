-- Rollback migration 000031: Remove organization branding fields

-- Drop indexes
DROP INDEX IF EXISTS idx_organizations_email;
DROP INDEX IF EXISTS idx_organizations_tin;

-- Remove columns
ALTER TABLE organizations
    DROP COLUMN IF EXISTS logo_url,
    DROP COLUMN IF EXISTS email,
    DROP COLUMN IF EXISTS tin,
    DROP COLUMN IF EXISTS registration_number;
