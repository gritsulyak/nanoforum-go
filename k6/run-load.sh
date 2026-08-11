#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

BASE_URL="${BASE_URL:-http://localhost:8084}"
BASE_PATH="${BASE_PATH:-/forum}"
DURATION="${DURATION:-10s}"
RPS_LEVELS="${RPS_LEVELS:-100 5000 10000 20000}"
if [ "$#" -gt 0 ]; then
  SCRIPTS="$*"
else
  SCRIPTS="${SCRIPTS:-login posts post-create}"
fi
OUTPUT_DIR="${SCRIPT_DIR}/results"

mkdir -p "${OUTPUT_DIR}"

echo "Target: ${BASE_URL}${BASE_PATH} (${DURATION} per run)"
echo "RPS levels: ${RPS_LEVELS}"
echo "Scripts: ${SCRIPTS}"
echo

run_script() {
  local script="$1"
  local rps="$2"
  echo "=============================================="
  echo "Script: ${script} | RPS: ${rps}"
  echo "=============================================="
  k6 run \
    --out json="${OUTPUT_DIR}/${script}-${rps}rps.json" \
    --summary-export="${OUTPUT_DIR}/${script}-${rps}rps-summary.json" \
    -e BASE_URL="${BASE_URL}" \
    -e BASE_PATH="${BASE_PATH}" \
    -e RPS="${rps}" \
    -e DURATION="${DURATION}" \
    -e USERNAME="${LOAD_USER:-loadtest}" \
    -e PASSWORD="${LOAD_PASSWORD:-loadtest}" \
    "${SCRIPT_DIR}/${script}.js"
  echo
}

for script in ${SCRIPTS}; do
  for rps in ${RPS_LEVELS}; do
    run_script "${script}" "${rps}"
  done
done

echo "All done. Results saved to ${OUTPUT_DIR}"
