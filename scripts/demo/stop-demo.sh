# CLASSIFICATION: UNCLASSIFIED
#!/usr/bin/env bash
# scripts/demo/stop-demo.sh
# Stop RTSA demo services with flexible scope options (v2.0).
#
# Usage: bash scripts/demo/stop-demo.sh [OPTIONS]
#
# Options:
#   --live-feed   Stop only the live-feed / simulator containers (default if no option given)
#   --containers  Stop all demo containers (docker compose down)
#   --volumes     Stop all containers and remove Docker volumes (full reset)
#   -h, --help    Show this help

set -euo pipefail

GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
NC='\033[0m'

usage() {
  cat <<'USAGE'
Stop RTSA demo services.

Usage:
  bash scripts/demo/stop-demo.sh [OPTIONS]

Options:
  --live-feed    Stop only the live-feed / simulator containers (default).
  --containers   Stop all demo containers (docker compose down).
  --volumes      Stop all containers and remove Docker volumes (full teardown).
  -h, --help     Show this help message.

Examples:
  bash scripts/demo/stop-demo.sh                  # Stop simulators only (default)
  bash scripts/demo/stop-demo.sh --containers     # Stop all containers
  bash scripts/demo/stop-demo.sh --volumes        # Full teardown including volumes
USAGE
}

STOP_LIVE_FEED=false
STOP_CONTAINERS=false
STOP_VOLUMES=false

# Default: if no flags given, stop live-feed only
if [ "$#" -eq 0 ]; then
  STOP_LIVE_FEED=true
fi

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

COMPOSE_ARGS="-f deploy/docker-compose.yml -f deploy/docker-compose.services.yml"

echo -e "${CYAN}=== Stopping RTSA demo services ===${NC}"

if $STOP_LIVE_FEED; then
  echo -e "${CYAN}Stopping simulator / live-feed containers...${NC}"
  # Stop simulator and live-feed containers specifically; tolerate missing containers
  docker compose ${COMPOSE_ARGS} stop simulator live-feed 2>/dev/null || true
  echo -e "${GREEN}[v]${NC} Simulator and live-feed stopped"
  echo ""
  echo "  Note: Infrastructure services (Redpanda, ClickHouse, microservices) are still running."
  echo "  The UI at http://localhost:5173 remains accessible."
  echo "  To stop everything: bash scripts/demo/stop-demo.sh --containers"
fi

if $STOP_CONTAINERS; then
  echo -e "${CYAN}Stopping all demo containers...${NC}"
  docker compose ${COMPOSE_ARGS} down
  echo -e "${GREEN}[v]${NC} All containers stopped"
fi

if $STOP_VOLUMES; then
  echo -e "${CYAN}Removing all containers and Docker volumes (full teardown)...${NC}"
  docker compose ${COMPOSE_ARGS} down -v
  echo -e "${GREEN}[v]${NC} Full teardown complete — all containers and volumes removed"
  echo ""
  echo "  To restart from scratch:"
  echo "    bash scripts/demo/run-demo.sh multi-domain --setup --seed"
fi

echo -e "${GREEN}=== Demo stop script completed ===${NC}"
