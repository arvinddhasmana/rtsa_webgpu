# CLASSIFICATION: UNCLASSIFIED
#!/usr/bin/env bash
# scripts/cop-dev/stop-backend.sh
# Stop RTSA backend services used for COP Web UI development.
#
# Usage:
#   bash scripts/cop-dev/stop-backend.sh [OPTIONS]
#
# Options:
#   --sim        Stop only the simulator container (default)
#   --services   Stop all containers, keep Docker volumes
#   --reset      Full teardown including Docker volumes
#   -h, --help   Show this help

set -euo pipefail

GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BOLD='\033[1m'
NC='\033[0m'

usage() {
  cat <<'USAGE'
COP Web UI — Backend Stop Script

Usage:
  bash scripts/cop-dev/stop-backend.sh [OPTIONS]

Options:
  --sim        Stop only the simulator container (default if no flag given)
  --services   Stop all containers, keep Docker volumes (data preserved)
  --reset      Full teardown: stop all containers AND remove Docker volumes
  -h, --help   Show this help

Examples:
  bash scripts/cop-dev/stop-backend.sh            # Stop simulator only
  bash scripts/cop-dev/stop-backend.sh --sim       # Same as default
  bash scripts/cop-dev/stop-backend.sh --services  # Stop all, keep data
  bash scripts/cop-dev/stop-backend.sh --reset     # Clean slate
USAGE
}

MODE="sim"  # default

while [ "$#" -gt 0 ]; do
  case "$1" in
    --sim)
      MODE="sim"
      ;;
    --services)
      MODE="services"
      ;;
    --reset)
      MODE="reset"
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

DC="docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.services.yml"

echo ""
echo -e "${BOLD}${CYAN}╔══════════════════════════════════════════════════════╗${NC}"
echo -e "${BOLD}${CYAN}║   RTSA COP — Backend Stop (mode: ${MODE})${NC}"
echo -e "${BOLD}${CYAN}╚══════════════════════════════════════════════════════╝${NC}"
echo ""

case "$MODE" in
  sim)
    echo -e "${CYAN}Stopping simulator container...${NC}"
    $DC stop simulator 2>/dev/null || true
    docker rm -f rtsa-simulator 2>/dev/null || true
    echo -e "${GREEN}✓ Simulator stopped${NC}"
    echo ""
    echo "  Backend services are still running."
    echo "  Frontend: http://localhost:5173 (if npm run dev is running)"
    echo "  To stop everything: bash scripts/cop-dev/stop-backend.sh --services"
    ;;

  services)
    echo -e "${CYAN}Stopping all containers (keeping Docker volumes)...${NC}"
    $DC down
    echo -e "${GREEN}✓ All containers stopped. Docker volumes preserved.${NC}"
    echo ""
    echo "  Data is preserved — next start will be fast."
    echo "  Restart: bash scripts/cop-dev/start-backend.sh"
    ;;

  reset)
    echo -e "${RED}Full teardown: stopping all containers and removing Docker volumes...${NC}"
    $DC down -v
    echo -e "${GREEN}✓ Full reset complete. All containers and volumes removed.${NC}"
    echo ""
    echo "  Next start will re-initialize everything from scratch."
    echo "  Restart: bash scripts/cop-dev/start-backend.sh"
    ;;
esac

echo ""
