-- Drop indexes
DROP INDEX IF EXISTS idx_permissions_name;
DROP INDEX IF EXISTS idx_permissions_visibility;

-- Drop permissions table
DROP TABLE IF EXISTS permissions;