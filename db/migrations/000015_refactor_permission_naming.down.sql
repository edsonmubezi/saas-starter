-- Rollback migration: Restore original permission naming

BEGIN;

-- Drop the new index
DROP INDEX IF EXISTS idx_permissions_scope;

-- Add back visibility column
ALTER TABLE permissions ADD COLUMN IF NOT EXISTS visibility VARCHAR(10) NOT NULL DEFAULT 'ALL';

-- Restore original names from backup
UPDATE permissions
SET name = old_name
WHERE old_name IS NOT NULL;

-- Set visibility based on restored names
UPDATE permissions
SET visibility = 'HQ_ONLY'
WHERE name LIKE 'hq-only.%'
   OR name = 'shared.loan.view'
   OR name LIKE 'orgs.preloanpayment.%';

UPDATE permissions
SET visibility = 'ALL'
WHERE visibility IS NULL OR visibility = '';

-- Add back visibility index
CREATE INDEX IF NOT EXISTS idx_permissions_visibility ON permissions(visibility);

-- Remove backup column
ALTER TABLE permissions DROP COLUMN IF EXISTS old_name;

-- Add visibility check constraint
ALTER TABLE permissions ADD CONSTRAINT permissions_visibility_check
    CHECK (visibility IN ('ALL', 'HQ_ONLY'));

COMMIT;
