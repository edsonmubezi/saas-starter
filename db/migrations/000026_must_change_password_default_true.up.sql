-- Set default for all future inserts
ALTER TABLE users ALTER COLUMN must_change_password SET DEFAULT true;

-- Fix all existing rows that have NULL or false
UPDATE users SET must_change_password = true WHERE must_change_password IS NULL OR must_change_password = false;

-- Ensure column can never be NULL
ALTER TABLE users ALTER COLUMN must_change_password SET NOT NULL;
