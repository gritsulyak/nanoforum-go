#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
USERNAME="${LOAD_USER:-loadtest}"
PASSWORD="${LOAD_PASSWORD:-loadtest}"
COMPOSE="${COMPOSE:-docker compose -f ${ROOT}/docker/docker-compose.yml}"

echo "Creating load-test user '${USERNAME}' inside the forum container..."
if printf '%s\n%s\n' "${USERNAME}" "${PASSWORD}" | ${COMPOSE} --profile tools run --rm createuser; then
  echo "User '${USERNAME}' created."
else
  echo "User '${USERNAME}' may already exist; continuing."
fi
