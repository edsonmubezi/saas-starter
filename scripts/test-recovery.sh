#!/bin/bash
#
# Test PostgreSQL Point-in-Time Recovery (PITR)
#
# This script tests the recovery process by:
# 1. Downloading a base backup from S3
# 2. Restoring it to a test instance
# 3. Replaying WAL files to a specific point in time
# 4. Verifying data integrity
#
# Usage:
#   ./test-recovery.sh [target-time]
#
# Example:
#   ./test-recovery.sh "2025-01-15 14:30:00"
#
# Prerequisites:
#   - PostgreSQL client tools
#   - AWS CLI configured
#   - Sufficient disk space
#   - Non-production environment!

set -e
set -u

# ============================================
# Configuration
# ============================================
TARGET_TIME="${1:-latest}"  # Default to latest if not specified

S3_BUCKET="${PGBACKUP_S3_BUCKET:-}"
S3_BASE_PREFIX="${PGBACKUP_S3_PREFIX:-base-backups/}"
S3_WAL_PREFIX="${PGBACKUP_S3_PREFIX:-wal-archives/}"
AWS_REGION="${AWS_REGION:-us-east-1}"

TEST_PGDATA="/tmp/test_recovery_pgdata"
TEST_PORT="${TEST_RECOVERY_PORT:-5433}"

echo "============================================"
echo "PostgreSQL Recovery Test"
echo "============================================"
echo "Target Time: $TARGET_TIME"
echo "Test Data Dir: $TEST_PGDATA"
echo "Test Port: $TEST_PORT"
echo "============================================"

# ============================================
# Validation
# ============================================
if [ -z "$S3_BUCKET" ]; then
    echo "ERROR: PGBACKUP_S3_BUCKET not set"
    exit 1
fi

# Warning for production
read -p "WARNING: This will create a test PostgreSQL instance. Continue? (yes/no): " confirm
if [ "$confirm" != "yes" ]; then
    echo "Aborted."
    exit 0
fi

# ============================================
# Cleanup Previous Test
# ============================================
echo "[$(date '+%H:%M:%S')] Cleaning up previous test instance..."

if [ -d "$TEST_PGDATA" ]; then
    # Try to stop any running test instance
    pg_ctl -D "$TEST_PGDATA" stop -m immediate 2>/dev/null || true
    rm -rf "$TEST_PGDATA"
fi

mkdir -p "$TEST_PGDATA"

# ============================================
# Download Latest Base Backup
# ============================================
echo "[$(date '+%H:%M:%S')] Finding latest base backup..."

LATEST_BACKUP=$(aws s3 ls "s3://${S3_BUCKET}/${S3_BASE_PREFIX}" \
    --region "$AWS_REGION" \
    | sort | tail -n 1 | awk '{print $2}')

if [ -z "$LATEST_BACKUP" ]; then
    echo "ERROR: No backups found in s3://${S3_BUCKET}/${S3_BASE_PREFIX}"
    exit 1
fi

echo "[$(date '+%H:%M:%S')] Downloading backup: ${LATEST_BACKUP}"

aws s3 sync "s3://${S3_BUCKET}/${S3_BASE_PREFIX}${LATEST_BACKUP}" "$TEST_PGDATA" \
    --region "$AWS_REGION" \
    --only-show-errors

# ============================================
# Decompress Backup
# ============================================
echo "[$(date '+%H:%M:%S')] Decompressing backup..."

if [ -f "${TEST_PGDATA}/base.tar.gz" ]; then
    cd "$TEST_PGDATA"
    tar -xzf base.tar.gz
    rm base.tar.gz
fi

# ============================================
# Create recovery.conf
# ============================================
echo "[$(date '+%H:%M:%S')] Creating recovery configuration..."

cat > "${TEST_PGDATA}/recovery.signal" <<EOF
# Recovery signal file - triggers recovery mode
EOF

# Create postgresql.conf override for recovery
cat > "${TEST_PGDATA}/postgresql.auto.conf" <<EOF
# Auto-generated recovery configuration
port = ${TEST_PORT}
restore_command = 'aws s3 cp s3://${S3_BUCKET}/${S3_WAL_PREFIX}%f %p --region ${AWS_REGION} --only-show-errors'
EOF

if [ "$TARGET_TIME" != "latest" ]; then
    echo "recovery_target_time = '$TARGET_TIME'" >> "${TEST_PGDATA}/postgresql.auto.conf"
fi

# ============================================
# Start Test Instance
# ============================================
echo "[$(date '+%H:%M:%S')] Starting test PostgreSQL instance..."

pg_ctl -D "$TEST_PGDATA" -l "${TEST_PGDATA}/logfile" start

# Wait for server to start
sleep 5

# Check if server is running
if ! pg_isready -p "$TEST_PORT" -t 30; then
    echo "ERROR: Test instance failed to start"
    echo "Check logs at: ${TEST_PGDATA}/logfile"
    exit 1
fi

echo "[$(date '+%H:%M:%S')] Test instance started successfully on port $TEST_PORT"

# ============================================
# Verify Recovery
# ============================================
echo "[$(date '+%H:%M:%S')] Verifying recovery..."

# Check if we can connect
if psql -p "$TEST_PORT" -U postgres -d postgres -c "SELECT NOW();" > /dev/null 2>&1; then
    echo "✓ Successfully connected to test instance"
else
    echo "✗ Failed to connect to test instance"
    exit 1
fi

# Display recovery status
echo ""
echo "Recovery Status:"
echo "================"
psql -p "$TEST_PORT" -U postgres -d postgres -c "
    SELECT pg_is_in_recovery() AS in_recovery,
           pg_last_wal_receive_lsn() AS last_wal_receive,
           pg_last_wal_replay_lsn() AS last_wal_replay;
"

# Count tables
TABLE_COUNT=$(psql -p "$TEST_PORT" -U postgres -d postgres -t -c "
    SELECT COUNT(*)
    FROM information_schema.tables
    WHERE table_schema = 'public';
" | xargs)

echo ""
echo "Database Statistics:"
echo "===================="
echo "Tables in public schema: $TABLE_COUNT"

# ============================================
# Instructions
# ============================================
echo ""
echo "============================================"
echo "Recovery Test Complete!"
echo "============================================"
echo ""
echo "Test instance is running on port $TEST_PORT"
echo "Data directory: $TEST_PGDATA"
echo ""
echo "To connect:"
echo "  psql -p $TEST_PORT -U postgres -d postgres"
echo ""
echo "To stop and cleanup:"
echo "  pg_ctl -D \"$TEST_PGDATA\" stop"
echo "  rm -rf \"$TEST_PGDATA\""
echo ""
echo "============================================"

exit 0
