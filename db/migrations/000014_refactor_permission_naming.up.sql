-- Migration: Refactor permission naming from dual-source to prefix-based
-- Changes:
--   - Rename hq-only.* to admin.*
--   - Rename shared.loan.view to admin.loan.view_all
--   - Rename orgs.preloanpayment.* to admin.preloanpayment.* (was HQ_ONLY)
--   - Rename remaining orgs.* to tenant.*
--   - Rename selfservice.* to self.*
--   - Rename unprefixed leave.*, payroll*, adjustment.* to tenant.*
--   - Drop visibility column (now derived from prefix)

BEGIN;

-- Backup old names for reference (optional - can be removed after verification)
ALTER TABLE permissions ADD COLUMN IF NOT EXISTS old_name VARCHAR(100);
UPDATE permissions SET old_name = name WHERE old_name IS NULL;

-- Step 1: Rename hq-only.* to admin.*
UPDATE permissions
SET name = 'admin.' || substring(name from 9)
WHERE name LIKE 'hq-only.%';

-- Step 2: Rename shared.loan.view to admin.loan.view_all
UPDATE permissions
SET name = 'admin.loan.view_all'
WHERE name = 'shared.loan.view';

-- Step 3: Rename orgs.preloanpayment.* to admin.preloanpayment.* (was HQ_ONLY)
UPDATE permissions
SET name = 'admin.' || substring(name from 6)
WHERE name LIKE 'orgs.preloanpayment.%';

-- Step 4: Rename remaining orgs.* to tenant.*
UPDATE permissions
SET name = 'tenant.' || substring(name from 6)
WHERE name LIKE 'orgs.%';

-- Step 5: Rename selfservice.* to self.*
UPDATE permissions
SET name = 'self.' || substring(name from 13)
WHERE name LIKE 'selfservice.%';

-- Step 6: Rename unprefixed leave.* to tenant.leave.*
UPDATE permissions
SET name = 'tenant.' || name
WHERE name LIKE 'leave.%';

-- Step 7: Rename unprefixed payroll.* and payroll_setting.* to tenant.*
UPDATE permissions
SET name = 'tenant.' || name
WHERE name LIKE 'payroll%' AND name NOT LIKE 'tenant.%' AND name NOT LIKE 'self.%';

-- Step 8: Rename unprefixed adjustment.* to tenant.adjustment.*
UPDATE permissions
SET name = 'tenant.' || name
WHERE name LIKE 'adjustment.%';

-- Step 9: Drop visibility column and index (now derived from prefix)
DROP INDEX IF EXISTS idx_permissions_visibility;
ALTER TABLE permissions DROP COLUMN IF EXISTS visibility;

-- Step 10: Create index for prefix-based queries
CREATE INDEX IF NOT EXISTS idx_permissions_scope ON permissions (split_part(name, '.', 1));

COMMIT;
