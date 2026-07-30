-- Drop indexes
DROP INDEX IF EXISTS idx_logs_ip_address;
DROP INDEX IF EXISTS idx_logs_action;
DROP INDEX IF EXISTS idx_logs_created_at;
DROP INDEX IF EXISTS idx_logs_user;

-- Drop logs table
DROP TABLE IF EXISTS logs;