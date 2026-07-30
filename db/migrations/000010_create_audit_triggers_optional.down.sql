-- Drop all audit triggers if they exist
DROP TRIGGER IF EXISTS audit_employees_changes ON employees CASCADE;
DROP TRIGGER IF EXISTS audit_loans_changes ON loans CASCADE;
DROP TRIGGER IF EXISTS audit_users_changes ON users CASCADE;
DROP TRIGGER IF EXISTS audit_roles_changes ON roles CASCADE;
DROP TRIGGER IF EXISTS audit_role_permissions_changes ON role_permissions CASCADE;

-- Drop the trigger function
DROP FUNCTION IF EXISTS audit.log_data_changes() CASCADE;
