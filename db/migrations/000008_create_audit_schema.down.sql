-- Drop audit schema and all tables
DROP TABLE IF EXISTS audit.access_log CASCADE;
DROP TABLE IF EXISTS audit.dead_letter CASCADE;
DROP TABLE IF EXISTS audit.events CASCADE;
DROP TABLE IF EXISTS audit.outbox CASCADE;
DROP SCHEMA IF EXISTS audit CASCADE;
