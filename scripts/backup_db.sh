#!/usr/bin/env bash
set -euo pipefail

BACKUP_DIR="${HERASAT_BACKUP_DIR:-./backups}"
mkdir -p "$BACKUP_DIR"

TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
TARGET_PATH="$BACKUP_DIR/backup_herasat_${TIMESTAMP}.sql.gz"

CONTAINER_NAME="herasat_postgres"
DB_USER="herasat_user"
DB_NAME="herasat_db"

if command -v pg_dump >/dev/null 2>&1; then
    if [ -n "${HERASAT_DB_CONN:-}" ]; then
        pg_dump --dbname "$HERASAT_DB_CONN" --no-owner --no-acl --clean --if-exists | gzip -9 > "$TARGET_PATH"
    else
        pg_dump -h 127.0.0.1 -p 5432 -U "$DB_USER" --no-owner --no-acl --clean --if-exists "$DB_NAME" | gzip -9 > "$TARGET_PATH"
    fi
elif command -v docker >/dev/null 2>&1; then
    docker exec "$CONTAINER_NAME" pg_dump -U "$DB_USER" --no-owner --no-acl --clean --if-exists "$DB_NAME" | gzip -9 > "$TARGET_PATH"
else
    echo "ERROR: Neither pg_dump nor docker found" >&2
    exit 1
fi

echo "BACKUP_SUCCESS: $TARGET_PATH"
