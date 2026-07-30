-- Drop indexes
DROP INDEX IF EXISTS idx_organizations_deleted_at;
DROP INDEX IF EXISTS idx_organizations_name;

-- Drop organizations table
DROP TABLE IF EXISTS organizations;