#!/usr/bin/env bash
# CLASSIFICATION: UNCLASSIFIED
# scripts/demo/stop-demo.sh
# Stop RTSA demo services with flexible options
# Usage: bash scripts/demo/stop-demo.sh [OPTIONS]
#   Options: --live-feed (default), --containers, --volumes
#   If no options are provided, defaults to stopping only the live-feed container.

set -euo pipefail

usage() {
  cat <<'USAGE'
Stop RTSA demo services.

Usage:
  bash scripts/demo/stop-demo.sh [OPTIONS]

Options:
  --live-feed   Stop only the live-feed container (default).
  --containers   Stop all demo containers.
  --volumes     Stop all containers and remove Docker volumes.
  -h, --help    Show this help message.
USAGE
}

# Default behavior flags
STOP_LIVE_FEED=true
STOP_CONTAINERS=false
STOP_VOLUMES=false

# Parse arguments
while [[ "$#" -gt 0 ]]; do
  case "$1" in
    --live-feed)
      STOP_LIVE_FEED=true
      ;;
    --containers)
      STOP_CONTAINERS=true
      ;;
    --volumes)
      STOP_VOLUMES=true
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      usage
      exit 1
      ;;
  esac
  shift
done

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
cd "$PROJECT_ROOT"

echo "=== Stopping RTSA services ==="

if $STOP_LIVE_FEED; then
  echo "Stopping live-feed container only..."
  docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.services.yml stop live-feed || true
fi

if $STOP_CONTAINERS; then
  echo "Stopping all demo containers..."
  docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.services.yml down
fi

if $STOP_VOLUMES; then
  echo "Removing Docker volumes..."
  docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.services.yml down -v
fi

echo "=== Demo stop script completed ==="
