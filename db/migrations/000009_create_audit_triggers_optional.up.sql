-- Optional: Automatic audit triggers for crown-jewel tables
-- These capture before/after state automatically at the database level
-- IMPORTANT: These are DISABLED by default. Enable per table as needed.

-- Generic audit trigger function
CREATE OR REPLACE FUNCTION audit.log_data_changes()
RETURNS TRIGGER AS $$
DECLARE
    v_tenant_id BIGINT;
    v_user_id BIGINT;
    v_resource_id BIGINT;
    v_action TEXT;
BEGIN
    -- Try to get tenant_id and user_id from session variables
    -- These should be set by application using SET LOCAL
    BEGIN
        v_tenant_id := current_setting('app.tenant_id', true)::BIGINT;
    EXCEPTION WHEN OTHERS THEN
        v_tenant_id := NULL;
    END;

    BEGIN
        v_user_id := current_setting('app.user_id', true)::BIGINT;
    EXCEPTION WHEN OTHERS THEN
        v_user_id := NULL;
    END;

    -- Determine action
    v_action := TG_OP || '_' || UPPER(TG_TABLE_NAME);

    -- Get resource ID from NEW or OLD
    IF TG_OP = 'DELETE' THEN
        v_resource_id := (to_jsonb(OLD)->>'id')::BIGINT;
    ELSE
        v_resource_id := (to_jsonb(NEW)->>'id')::BIGINT;
    END IF;

    -- Insert into outbox
    INSERT INTO audit.outbox (event_json, created_at)
    VALUES (
        jsonb_build_object(
            'occurred_at', NOW(),
            'actor_type', 'user',
            'actor_id', v_user_id,
            'tenant_id', COALESCE(v_tenant_id, 0),
            'action', v_action,
            'resource_type', TG_TABLE_NAME,
            'resource_id', v_resource_id,
            'before_json', CASE
                WHEN TG_OP IN ('UPDATE', 'DELETE') THEN row_to_json(OLD)
                ELSE NULL
            END,
            'after_json', CASE
                WHEN TG_OP IN ('INSERT', 'UPDATE') THEN row_to_json(NEW)
                ELSE NULL
            END,
            'severity', CASE
                WHEN TG_OP = 'DELETE' THEN 'HIGH'
                WHEN TG_OP = 'UPDATE' THEN 'MEDIUM'
                ELSE 'LOW'
            END,
            'origin_service', 'database-trigger'
        ),
        NOW()
    );

    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Example triggers (COMMENTED OUT - enable as needed)
--
-- Uncomment these to enable automatic auditing for specific tables:

-- For employees table:
-- CREATE TRIGGER audit_employees_changes
-- AFTER INSERT OR UPDATE OR DELETE ON employees
-- FOR EACH ROW EXECUTE FUNCTION audit.log_data_changes();

-- For loans table:
-- CREATE TRIGGER audit_loans_changes
-- AFTER INSERT OR UPDATE OR DELETE ON loans
-- FOR EACH ROW EXECUTE FUNCTION audit.log_data_changes();

-- For users table (high security):
-- CREATE TRIGGER audit_users_changes
-- AFTER INSERT OR UPDATE OR DELETE ON users
-- FOR EACH ROW EXECUTE FUNCTION audit.log_data_changes();

-- For roles table (high security):
-- CREATE TRIGGER audit_roles_changes
-- AFTER INSERT OR UPDATE OR DELETE ON roles
-- FOR EACH ROW EXECUTE FUNCTION audit.log_data_changes();

-- For role_permissions (privilege escalation detection):
-- CREATE TRIGGER audit_role_permissions_changes
-- AFTER INSERT OR UPDATE OR DELETE ON role_permissions
-- FOR EACH ROW EXECUTE FUNCTION audit.log_data_changes();

COMMENT ON FUNCTION audit.log_data_changes() IS
'Generic trigger function for automatic audit logging. Captures before/after state and writes to outbox. Enable per table by creating AFTER triggers.';

-- Note: For production use, prefer application-level audit calls for better control
-- These triggers are useful for:
-- 1. Crown-jewel tables where you want defense-in-depth
-- 2. Catching unauthorized direct database modifications
-- 3. Compliance requirements for specific tables
