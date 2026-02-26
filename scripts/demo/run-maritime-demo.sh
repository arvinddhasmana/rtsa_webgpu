#!/usr/bin/env bash
# CLASSIFICATION: UNCLASSIFIED
# scripts/demo/run-maritime-demo.sh
# Launch the maritime demo scenario (20 minutes)
# Usage: bash scripts/demo/run-maritime-demo.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "=== RTSA Maritime Demo — Starting infrastructure ==="
cd "$PROJECT_ROOT"
docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.services.yml up -d --wait

echo "=== Waiting for services to stabilise (30s) ==="
sleep 30

echo "=== Running init scripts ==="
bash scripts/dev/init-topics.sh
bash scripts/dev/init-clickhouse.sh

echo "=== Starting simulator with maritime-demo scenario ==="
docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.services.yml \
  run --rm simulator \
  --scenario /app/scenarios/maritime-demo.yaml

echo "=== Maritime demo complete ==="
