-- Add tenant user unlock permission for org admins
INSERT INTO permissions (name) VALUES ('tenant.user.unlock')
ON CONFLICT (name) DO NOTHING;
