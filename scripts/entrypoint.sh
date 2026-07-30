#!/bin/bash
# Docker entrypoint script for EkklesiaHR
# Handles initialization tasks before starting the main application
# Runs as root initially, then drops to app user for the server

set -e

echo "=========================================="
echo "EkklesiaHR Backend Starting..."
echo "=========================================="

# Run database migrations if enabled (can run as root)
# NOTE: In production, set RUN_MIGRATIONS=false and run migrations manually
if [ "${RUN_MIGRATIONS}" = "true" ]; then
    echo "Running database migrations..."
    cd /app
    if /app/migrate -command up; then
        echo "Migrations completed successfully"
    else
        echo "Warning: Migration failed - continuing anyway (app may have limited functionality)"
        echo "Run migrations manually to fix: /app/migrate -command up"
    fi
fi

# Ensure uploads directory exists and is writable by app user
mkdir -p /app/uploads
chown -R app:app /app/uploads

echo "Initialization complete, starting server as app user..."

# Drop privileges and start the main application as app user
exec su-exec app /app/myapp "$@"
