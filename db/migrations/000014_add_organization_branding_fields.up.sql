-- Migration 000031: Add organization branding fields
-- Add fields for company branding on PDF documents and other exports:
-- 1. logo_url - Path to company logo in storage
-- 2. email - Company email address
-- 3. tin - Tax Identification Number
-- 4. registration_number - Business registration number

-- Add logo_url column for storing company logo path
ALTER TABLE organizations
    ADD COLUMN IF NOT EXISTS logo_url VARCHAR(500);

-- Add email column for company email
ALTER TABLE organizations
    ADD COLUMN IF NOT EXISTS email VARCHAR(255);

-- Add TIN (Tax Identification Number) column
ALTER TABLE organizations
    ADD COLUMN IF NOT EXISTS tin VARCHAR(100);

-- Add business registration number column
ALTER TABLE organizations
    ADD COLUMN IF NOT EXISTS registration_number VARCHAR(100);

-- Add indexes for commonly queried fields
CREATE INDEX IF NOT EXISTS idx_organizations_email ON organizations(email) WHERE email IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_organizations_tin ON organizations(tin) WHERE tin IS NOT NULL;

-- Add comments for documentation
COMMENT ON COLUMN organizations.logo_url IS 'Path to company logo file in storage system (S3 or local)';
COMMENT ON COLUMN organizations.email IS 'Official company email address for correspondence and documents';
COMMENT ON COLUMN organizations.tin IS 'Tax Identification Number for official documents';
COMMENT ON COLUMN organizations.registration_number IS 'Business registration number for legal compliance';
