# CLASSIFICATION: UNCLASSIFIED
#!/usr/bin/env bash
# scripts/cop-dev/start-backend.sh
# One-command backend startup for COP Web UI development.
#
# Starts all RTSA infrastructure and services EXCEPT web-cop,
# so you can run `npm run dev` locally with hot-reload on port 5173.
#
# Usage:
#   bash scripts/cop-dev/start-backend.sh [OPTIONS]
#
# Options:
#   --setup             Run scripts/setup/setup-dev.sh first (first-time only)
#   --scenario <file>   Simulator scenario YAML (default: sensor-health-demo.yaml)
#   --no-sim            Skip starting the simulator
#   --no-seed           Skip seeding demo data
#   --dry-run           Print commands without executing
#   -h, --help          Show this help

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# ── Defaults ──────────────────────────────────────────────────
RUN_SETUP="false"
SCENARIO="sensor-health-demo.yaml"
RUN_SIMULATOR="true"
SEED_DATA="true"
DRY_RUN="false"

# ── Colours ───────────────────────────────────────────────────
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BOLD='\033[1m'
NC='\033[0m'

usage() {
  cat <<'USAGE'
COP Web UI — Backend Startup Script

Starts all RTSA backend services (infra + microservices + simulator)
while EXCLUDING the web-cop container. Run `npm run dev` locally.

Usage:
  bash scripts/cop-dev/start-backend.sh [OPTIONS]

Options:
  --setup              Run scripts/setup/setup-dev.sh first (first-time only)
  --scenario <file>    Simulator scenario YAML file (default: sensor-health-demo.yaml)
                       Available: maritime-demo.yaml, multi-domain-demo.yaml,
                       fusion-dashboard-demo.yaml, operator-ui-demo.yaml,
                       sensor-health-demo.yaml, nato-exchange-demo.yaml,
                       analyst-forensics-demo.yaml
  --no-sim             Skip starting the simulator
  --no-seed            Skip seeding demo data into ClickHouse
  --dry-run            Print commands without executing
  -h, --help           Show this help

Examples:
  bash scripts/cop-dev/start-backend.sh                           # Default (sensor-health + seed)
  bash scripts/cop-dev/start-backend.sh --scenario maritime-demo.yaml
  bash scripts/cop-dev/start-backend.sh --no-sim                  # Backend only, no simulator
  bash scripts/cop-dev/start-backend.sh --setup                   # First-time setup + start
USAGE
}

run_cmd() {
  if [ "$DRY_RUN" = "true" ]; then
    echo -e "${YELLOW}[dry-run]${NC} $1"
  else
    eval "$1"
  fi
}

# ── Parse arguments ───────────────────────────────────────────
while [ "$#" -gt 0 ]; do
  case "$1" in
    --setup)
      RUN_SETUP="true"
      ;;
    --scenario)
      shift
      SCENARIO="${1:-sensor-health-demo.yaml}"
      ;;
    --no-sim)
      RUN_SIMULATOR="false"
      ;;
    --no-seed)
      SEED_DATA="false"
      ;;
    --dry-run)
      DRY_RUN="true"
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown argument: $1"
      usage
      exit 1
      ;;
  esac
  shift
done

cd "$PROJECT_ROOT"

# Docker Compose with both files
DC="docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.services.yml"

echo ""
echo -e "${BOLD}${CYAN}╔══════════════════════════════════════════════════════╗${NC}"
echo -e "${BOLD}${CYAN}║   RTSA COP — Backend Startup for UI Development     ║${NC}"
echo -e "${BOLD}${CYAN}╚══════════════════════════════════════════════════════╝${NC}"
echo ""

# ── Phase 0: First-time setup ────────────────────────────────
if [ "$RUN_SETUP" = "true" ]; then
  echo -e "${CYAN}[0/6] Running first-time environment setup...${NC}"
  run_cmd "bash scripts/setup/setup-dev.sh"
fi

# ── Phase 1: Infrastructure ──────────────────────────────────
echo -e "${CYAN}[1/6] Starting infrastructure (Redpanda, ClickHouse, Observability)...${NC}"
run_cmd "$DC up -d --build --wait redpanda clickhouse otel-collector prometheus grafana loki tempo"

# ── Phase 2: Init topics + schema ────────────────────────────
echo -e "${CYAN}[2/6] Initializing Redpanda topics and ClickHouse schema...${NC}"
run_cmd "bash scripts/dev/init-topics.sh"
run_cmd "bash scripts/dev/init-clickhouse.sh"

# ── Phase 3: Seed demo data ─────────────────────────────────
if [ "$SEED_DATA" = "true" ]; then
  echo -e "${CYAN}[3/6] Seeding ClickHouse with demo data...${NC}"
  run_cmd "bash scripts/demo/seed-demo-data.sh"
else
  echo -e "${YELLOW}[3/6] Skipping data seeding (--no-seed)${NC}"
fi

# ── Phase 4: Application services (EXCLUDING web-cop) ───────
echo -e "${CYAN}[4/6] Starting application services (web-cop excluded)...${NC}"
run_cmd "$DC up -d --build --scale web-cop=0"

# ── Phase 5: Health checks ──────────────────────────────────
echo -e "${CYAN}[5/6] Waiting for services to pass health checks...${NC}"

SERVICES=(
  rtsa-radar-ingestion
  rtsa-ew-ingestion
  rtsa-elint-ingestion
  rtsa-isr-ingestion
  rtsa-ais-ingestion
  rtsa-cyber-ingestion
  rtsa-fusion-engine
  rtsa-track
  rtsa-alert
  rtsa-anomaly-detection
)

ALL_HEALTHY=true
for svc in "${SERVICES[@]}"; do
  printf "  %-35s" "$svc"
  attempt=0
  max=90
  healthy=false
  while [ "$attempt" -lt "$max" ]; do
    if docker exec "$svc" /bin/grpc_health_probe -addr=:50051 >/dev/null 2>&1; then
      healthy=true
      break
    fi
    attempt=$(( attempt + 1 ))
    sleep 1
  done
  if $healthy; then
    echo -e "${GREEN}✓ healthy${NC}"
  else
    echo -e "${RED}✗ timeout${NC}"
    ALL_HEALTHY=false
  fi
done

# Also check Envoy (HTTP health, not gRPC)
printf "  %-35s" "rtsa-envoy"
if curl -sf --max-time 5 http://localhost:9901/ready >/dev/null 2>&1; then
  echo -e "${GREEN}✓ healthy${NC}"
else
  echo -e "${YELLOW}! check manually${NC}"
fi

if ! $ALL_HEALTHY; then
  echo -e "${YELLOW}  ⚠ Some services did not pass health checks — backend may be degraded${NC}"
fi

# ── Phase 6: Simulator ──────────────────────────────────────
if [ "$RUN_SIMULATOR" = "true" ]; then
  echo -e "${CYAN}[6/6] Starting simulator (scenario: ${SCENARIO})...${NC}"
  # Run simulator detached so shell returns to developer
  run_cmd "$DC run -d --rm simulator --scenario /app/scenarios/${SCENARIO}"
  echo -e "${GREEN}  ✓ Simulator running in background${NC}"
else
  echo -e "${YELLOW}[6/6] Skipping simulator (--no-sim)${NC}"
fi

# ── Summary ──────────────────────────────────────────────────
echo ""
echo -e "${GREEN}╔══════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║   Backend ready for UI development!                  ║${NC}"
echo -e "${GREEN}╚══════════════════════════════════════════════════════╝${NC}"
echo ""
echo "  Next steps:"
echo "    cd web-cop && npm run dev"
echo ""
echo "  Connection URLs:"
echo "    Frontend (Vite)      : http://localhost:5173"
echo "    gRPC-Web (Envoy)     : https://localhost:8443"
echo "    Envoy Admin          : http://localhost:9901"
echo "    Redpanda Console     : http://localhost:8080"
echo "    ClickHouse HTTP      : http://localhost:8123/play"
echo "    Grafana              : http://localhost:3000 (admin/admin)"
echo ""
echo "  Simulator scenario: ${SCENARIO}"
echo "    Stop sim only : bash scripts/cop-dev/stop-backend.sh --sim"
echo "    Stop all      : bash scripts/cop-dev/stop-backend.sh --services"
echo "    Full reset    : bash scripts/cop-dev/stop-backend.sh --reset"
echo ""
