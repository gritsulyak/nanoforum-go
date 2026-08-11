#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

BASE_URL="${BASE_URL:-http://localhost:8084}"
BASE_PATH="${BASE_PATH:-/forum}"
SCRIPT="${SCRIPT:-posts}"
RPS="${RPS:-5000}"
DURATION="${DURATION:-60s}"
PROFILE_SECONDS="${PROFILE_SECONDS:-30}"
OUTPUT_DIR="${SCRIPT_DIR}/profiles"
PPROF_URL="${BASE_URL}${BASE_PATH}/debug/pprof"

mkdir -p "${OUTPUT_DIR}"

echo "Profile target: ${BASE_URL}${BASE_PATH}"
echo "Script: ${SCRIPT} | RPS: ${RPS} | duration: ${DURATION}"
echo "Profiling for ${PROFILE_SECONDS}s (CPU + trace) + heap/mutex/block snapshots"
echo

echo ">>> Starting load test in background..."
k6 run \
  --quiet \
  -e BASE_URL="${BASE_URL}" \
  -e BASE_PATH="${BASE_PATH}" \
  -e RPS="${RPS}" \
  -e DURATION="${DURATION}" \
  "${SCRIPT_DIR}/${SCRIPT}.js" &
K6_PID=$!

sleep 5
echo ">>> Capturing CPU profile (${PROFILE_SECONDS}s)..."
curl -s -o "${OUTPUT_DIR}/cpu.pprof" "${PPROF_URL}/profile?seconds=${PROFILE_SECONDS}"

echo ">>> Capturing trace (${PROFILE_SECONDS}s)..."
curl -s -o "${OUTPUT_DIR}/trace.out" "${PPROF_URL}/trace?seconds=${PROFILE_SECONDS}"

echo ">>> Capturing heap/mutex/block snapshots..."
curl -s -o "${OUTPUT_DIR}/heap.pprof" "${PPROF_URL}/heap"
curl -s -o "${OUTPUT_DIR}/mutex.pprof" "${PPROF_URL}/mutex"
curl -s -o "${OUTPUT_DIR}/block.pprof" "${PPROF_URL}/block"

echo ">>> Waiting for load test to finish..."
wait "${K6_PID}" || echo "k6 exited non-zero (expected at high RPS)"

echo
echo "Profiles saved to ${OUTPUT_DIR}"
echo
echo "Analyze:"
echo "  go tool pprof -http :9999 ${OUTPUT_DIR}/cpu.pprof"
echo "  go tool pprof -http :9999 ${OUTPUT_DIR}/heap.pprof"
echo "  go tool pprof -http :9999 ${OUTPUT_DIR}/mutex.pprof"
echo "  go tool pprof -http :9999 ${OUTPUT_DIR}/block.pprof"
echo "  go tool trace ${OUTPUT_DIR}/trace.out"
