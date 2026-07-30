-- Add is_assignable flag to roles table.
-- Roles with is_assignable = false cannot be assigned by org admins (e.g. SuperAdmin, OrgsAdmin).
ALTER TABLE roles
    ADD COLUMN IF NOT EXISTS is_assignable BOOLEAN NOT NULL DEFAULT true;

-- Mark system-level roles as non-assignable for every organization
UPDATE roles SET is_assignable = false
WHERE name IN ('SuperAdmin', 'OrgsAdmin');
