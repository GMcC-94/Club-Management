#!/bin/sh
set -e

BACKUP_DIR="/backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/club_management_$TIMESTAMP.sql"

# Create backup directory if it doesn't exist
mkdir -p $BACKUP_DIR

# Create backup
pg_dump > $BACKUP_FILE

# Compress
gzip $BACKUP_FILE

# Delete backups older than BACKUP_KEEP_DAYS
find $BACKUP_DIR -name "*.sql.gz" -mtime +${BACKUP_KEEP_DAYS:-7} -delete

echo "Backup completed: ${BACKUP_FILE}.gz"
