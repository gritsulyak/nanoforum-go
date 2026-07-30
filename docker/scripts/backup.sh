#!/usr/bin/env bash
set -euo pipefail

CONTAINER="${1:-nanoforum}"
DB_PATH_IN_CONTAINER="${DB_PATH:-/app/data/forum.db}"
BACKUP_DIR="./backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/forum_${TIMESTAMP}.db"

mkdir -p "${BACKUP_DIR}"

echo "Создаю бэкап ${DB_PATH_IN_CONTAINER} из контейнера ${CONTAINER}..."

docker exec "${CONTAINER}" sqlite3 "${DB_PATH_IN_CONTAINER}" ".backup /app/data/backup_tmp.db"
docker cp "${CONTAINER}:/app/data/backup_tmp.db" "${BACKUP_FILE}"
docker exec "${CONTAINER}" rm -f /app/data/backup_tmp.db

echo "Бэкап сохранён: ${BACKUP_FILE}"