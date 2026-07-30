#!/usr/bin/env bash
set -euo pipefail

CONTAINER="${1:?Укажите имя контейнера: restore.sh <container> <backup_file>}"
BACKUP_FILE="${2:?Укажите путь к файлу бэкапа: restore.sh <container> <backup_file>}"
DB_PATH_IN_CONTAINER="${DB_PATH:-/app/data/forum.db}"

if [ ! -f "${BACKUP_FILE}" ]; then
  echo "Файл бэкапа не найден: ${BACKUP_FILE}" >&2
  exit 1
fi

echo "Останавливаю форум перед восстановлением..."
docker stop "${CONTAINER}"

echo "Копирую ${BACKUP_FILE} в контейнер..."
docker cp "${BACKUP_FILE}" "${CONTAINER}:${DB_PATH_IN_CONTAINER}"

echo "Запускаю контейнер обратно..."
docker start "${CONTAINER}"

echo "Восстановление завершено."