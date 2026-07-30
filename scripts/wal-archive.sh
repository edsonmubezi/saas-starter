#!/bin/bash
#
# WAL Archive Script for PostgreSQL → AWS S3
#
# This script is called by PostgreSQL's archive_command to ship
# Write-Ahead Log (WAL) files to S3 for point-in-time recovery.
#
# Usage (called by PostgreSQL):
#   /path/to/wal-archive.sh %p %f
#
# Where:
#   %p = Full path to the WAL file to archive
#   %f = WAL file name
#
# Prerequisites:
#   - AWS CLI installed and configured
#   - Environment variables set (see below)
#   - PostgreSQL user has read access to WAL files
#
# Exit codes:
#   0  = Success
#   1  = Failure (causes PostgreSQL to retry)

set -e  # Exit on any error
set -u  # Exit on undefined variables

# ============================================
# Configuration from Environment Variables
# ============================================
S3_BUCKET="${PGBACKUP_S3_BUCKET:-}"
S3_PREFIX="${PGBACKUP_S3_PREFIX:-wal-archives/}"
AWS_REGION="${AWS_REGION:-us-east-1}"
LOG_FILE="${PGBACKUP_LOG_FILE:-/var/log/postgresql/wal-archive.log}"

# ============================================
# Input Parameters
# ============================================
WAL_FILE_PATH="$1"  # %p - Full path to WAL file
WAL_FILE_NAME="$2"  # %f - Just the file name

# ============================================
# Validation
# ============================================
if [ -z "$S3_BUCKET" ]; then
    echo "$(date '+%Y-%m-%d %H:%M:%S') ERROR: PGBACKUP_S3_BUCKET not set" >&2
    exit 1
fi

if [ ! -f "$WAL_FILE_PATH" ]; then
    echo "$(date '+%Y-%m-%d %H:%M:%S') ERROR: WAL file not found: $WAL_FILE_PATH" >&2
    exit 1
fi

# ============================================
# Upload to S3
# ============================================
S3_DESTINATION="s3://${S3_BUCKET}/${S3_PREFIX}${WAL_FILE_NAME}"

echo "$(date '+%Y-%m-%d %H:%M:%S') Archiving WAL file: $WAL_FILE_NAME to $S3_DESTINATION" >> "$LOG_FILE"

if aws s3 cp "$WAL_FILE_PATH" "$S3_DESTINATION" \
    --region "$AWS_REGION" \
    --storage-class STANDARD_IA \
    --server-side-encryption AES256 \
    --only-show-errors 2>> "$LOG_FILE"; then

    echo "$(date '+%Y-%m-%d %H:%M:%S') SUCCESS: Archived $WAL_FILE_NAME" >> "$LOG_FILE"
    exit 0
else
    echo "$(date '+%Y-%m-%d %H:%M:%S') FAILED: Could not archive $WAL_FILE_NAME" >> "$LOG_FILE"
    exit 1
fi
