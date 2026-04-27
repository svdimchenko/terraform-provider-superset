#!/usr/bin/env bash
set -euo pipefail

HOST="${SUPERSET_HOST:-http://localhost:8088}"
MAX_ATTEMPTS="${MAX_ATTEMPTS:-60}"
SLEEP_SECONDS="${SLEEP_SECONDS:-5}"

echo "Waiting for Superset at ${HOST} ..."

for i in $(seq 1 "$MAX_ATTEMPTS"); do
  if curl -sf "${HOST}/api/v1/security/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"admin","provider":"db"}' \
    -o /dev/null 2>/dev/null; then
    echo "Superset is ready (attempt ${i})."
    exit 0
  fi
  echo "  attempt ${i}/${MAX_ATTEMPTS} — not ready yet"
  sleep "$SLEEP_SECONDS"
done

echo "ERROR: Superset did not become ready in time."
exit 1
