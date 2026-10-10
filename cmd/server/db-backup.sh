#!/bin/bash
set -e

ENV_FILE_PATH="ПУТЬ ДО .ENV ПОПРАВЬ !!! /.env"
CONTAINER_NAME="site-backend-postgres"
BACKUP_DIR="/var/backups/db-backend"
KEEP_DAYS=7

if [ ! -f "$ENV_FILE_PATH" ]; then
    echo "❌ Error: .env file not found at $ENV_FILE_PATH"
    exit 1
fi

set -a
source "$ENV_FILE_PATH"
set +a

CONTAINER_NAME=${DB_CONTAINER_NAME:-"site-backend-postgres"}
DB_USER=${DB_USER:-"postgres"}
DB_NAME=${DB_NAME:-"postgres"}
DB_PASSWORD=${DB_PASSWORD:-""}


TIMESTAMP=$(date +'%Y-%m-%d_%H-%M-%S')
BACKUP_FILE="$BACKUP_DIR/db_backup_$TIMESTAMP.sql.gz"

echo "Starting database backup for '$DB_NAME'..."

mkdir -p "$BACKUP_DIR"

docker exec -i "$CONTAINER_NAME" pg_dump -U "$DB_USER" "$DB_NAME" | gzip > "$BACKUP_FILE"

echo "Backup saved to: $BACKUP_FILE"

echo "Cleaning up backups older than $KEEP_DAYS days..."
find "$BACKUP_DIR" -type f -name "db_backup_*.sql.gz" -mtime +"$KEEP_DAYS" -exec rm {} \;

echo "Backup process completed!!!"
