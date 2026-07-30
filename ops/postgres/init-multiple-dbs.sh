#!/bin/bash
# =============================================================================
# PostgreSQL Multiple Database Initialization Script
# =============================================================================
# Creates multiple databases from POSTGRES_MULTIPLE_DATABASES env variable
# Format: database1,database2,database3

set -e
set -u

function create_database() {
    local database=$1
    echo "Creating database '$database'..."
    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
        SELECT 'CREATE DATABASE $database'
        WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '$database')\gexec
        GRANT ALL PRIVILEGES ON DATABASE $database TO $POSTGRES_USER;
EOSQL
    echo "Database '$database' created successfully."
}

if [ -n "${POSTGRES_MULTIPLE_DATABASES:-}" ]; then
    echo "Multiple database creation requested: $POSTGRES_MULTIPLE_DATABASES"
    for db in $(echo $POSTGRES_MULTIPLE_DATABASES | tr ',' ' '); do
        # Skip if it's the default database (already created)
        if [ "$db" != "$POSTGRES_DB" ]; then
            create_database $db
        fi
    done
    echo "Multiple databases created successfully."
fi
