-- Add 2FA requirement flag to roles (admin-configurable per role)
ALTER TABLE roles ADD COLUMN IF NOT EXISTS requires_2fa BOOLEAN DEFAULT FALSE;
