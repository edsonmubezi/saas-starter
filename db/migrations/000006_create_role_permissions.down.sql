-- Drop indexes
DROP INDEX IF EXISTS idx_role_permissions_permission;
DROP INDEX IF EXISTS idx_role_permissions_role;

-- Drop role_permissions table
DROP TABLE IF EXISTS role_permissions;