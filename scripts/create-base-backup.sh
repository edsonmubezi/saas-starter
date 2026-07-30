#!/bin/bash
#
# Create PostgreSQL Base Backup and Upload to S3
#
# This script creates a full base backup of PostgreSQL using pg_basebackup,
# compresses it, and uploads to S3. This backup + WAL archives enable
# point-in-time recovery (PITR).
#
# Usage:
#   ./create-base-backup.sh
#
# Prerequisites:
#   - PostgreSQL client tools (pg_basebackup)
#   - AWS CLI installed and configured
#   - Environment variables set
#   - Sufficient disk space in /tmp or $BACKUP_DIR

set -e
set -u

# ============================================
# Configuration
# ============================================
PGHOST="${PGHOST:-localhost}"
PGPORT="${PGPORT:-5432}"
PGUSER="${PGUSER:-postgres}"
PGDATABASE="${PGDATABASE:-postgres}"

S3_BUCKET="${PGBACKUP_S3_BUCKET:-}"
S3_PREFIX="${PGBACKUP_S3_PREFIX:-base-backups/}"
AWS_REGION="${AWS_REGION:-us-east-1}"

BACKUP_DIR="${BACKUP_DIR:-/tmp/pg_backups}"
RETENTION_DAYS="${BACKUP_RETENTION_DAYS:-7}"

TIMESTAMP=$(date '+%Y%m%d_%H%M%S')
BACKUP_NAME="base-backup-${TIMESTAMP}"
BACKUP_PATH="${BACKUP_DIR}/${BACKUP_NAME}"

# ============================================
# Validation
# ============================================
if [ -z "$S3_BUCKET" ]; then
    echo "ERROR: PGBACKUP_S3_BUCKET environment variable not set"
    exit 1
fi

# Ensure backup directory exists
mkdir -p "$BACKUP_DIR"

echo "============================================"
echo "PostgreSQL Base Backup to S3"
echo "============================================"
echo "Host:       $PGHOST:$PGPORT"
echo "Database:   $PGDATABASE"
echo "S3 Bucket:  s3://${S3_BUCKET}/${S3_PREFIX}"
echo "Backup Dir: $BACKUP_PATH"
echo "============================================"

# ============================================
# Create Base Backup
# ============================================
echo "[$(date '+%Y-%m-%d %H:%M:%S')] Creating base backup..."

pg_basebackup \
    --host="$PGHOST" \
    --port="$PGPORT" \
    --username="$PGUSER" \
    --pgdata="$BACKUP_PATH" \
    --format=tar \
    --gzip \
    --compress=9 \
    --checkpoint=fast \
    --label="$BACKUP_NAME" \
    --progress \
    --verbose

if [ $? -ne 0 ]; then
    echo "ERROR: pg_basebackup failed"
    exit 1
fi

echo "[$(date '+%Y-%m-%d %H:%M:%S')] Base backup created successfully"

# ============================================
# Create backup manifest
# ============================================
MANIFEST_FILE="${BACKUP_PATH}/backup_manifest.txt"
cat > "$MANIFEST_FILE" <<EOF
Backup Name: ${BACKUP_NAME}
Timestamp: ${TIMESTAMP}
Database: ${PGDATABASE}
Host: ${PGHOST}:${PGPORT}
PostgreSQL Version: $(psql --version | head -n 1)
Backup Size: $(du -sh "$BACKUP_PATH" | cut -f1)
Created: $(date '+%Y-%m-%d %H:%M:%S %Z')
EOF

# ============================================
# Upload to S3
# ============================================
echo "[$(date '+%Y-%m-%d %H:%M:%S')] Uploading backup to S3..."

aws s3 sync "$BACKUP_PATH" "s3://${S3_BUCKET}/${S3_PREFIX}${BACKUP_NAME}/" \
    --region "$AWS_REGION" \
    --storage-class STANDARD_IA \
    --server-side-encryption AES256 \
    --delete \
    --only-show-errors

if [ $? -ne 0 ]; then
    echo "ERROR: S3 upload failed"
    exit 1
fi

echo "[$(date '+%Y-%m-%d %H:%M:%S')] Backup uploaded successfully"

# ============================================
# Cleanup Local Backup
# ============================================
echo "[$(date '+%Y-%m-%d %H:%M:%S')] Cleaning up local backup..."
rm -rf "$BACKUP_PATH"

# ============================================
# Delete Old Backups from S3
# ============================================
echo "[$(date '+%Y-%m-%d %H:%M:%S')] Cleaning up old backups (older than ${RETENTION_DAYS} days)..."

CUTOFF_DATE=$(date -d "${RETENTION_DAYS} days ago" '+%Y-%m-%d' 2>/dev/null || date -v -${RETENTION_DAYS}d '+%Y-%m-%d')

aws s3 ls "s3://${S3_BUCKET}/${S3_PREFIX}" --region "$AWS_REGION" | while read -r line; do
    BACKUP_DATE=$(echo "$line" | awk '{print $1}')
    BACKUP_DIR=$(echo "$line" | awk '{print $2}')

    if [[ "$BACKUP_DATE" < "$CUTOFF_DATE" ]]; then
        echo "  Deleting old backup: ${BACKUP_DIR}"
        aws s3 rm "s3://${S3_BUCKET}/${S3_PREFIX}${BACKUP_DIR}" \
            --region "$AWS_REGION" \
            --recursive \
            --only-show-errors
    fi
done

# ============================================
# Summary
# ============================================
echo "============================================"
echo "Backup Complete!"
echo "============================================"
echo "Backup Name:    ${BACKUP_NAME}"
echo "S3 Location:    s3://${S3_BUCKET}/${S3_PREFIX}${BACKUP_NAME}/"
echo "Completed:      $(date '+%Y-%m-%d %H:%M:%S')"
echo "============================================"

exit 0
