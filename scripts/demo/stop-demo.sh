#!/usr/bin/env bash
# CLASSIFICATION: UNCLASSIFIED
# scripts/demo/stop-demo.sh
# Stop all RTSA demo services and clean up volumes
# Usage: bash scripts/demo/stop-demo.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

cd "$PROJECT_ROOT"
echo "=== Stopping all RTSA services ==="
docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.services.yml down -v
echo "=== Demo stopped and volumes cleaned ==="
