# CLASSIFICATION: UNCLASSIFIED
#!/usr/bin/env bash
# scripts/demo/_common.sh
# Shared functions for RTSA demo scripts.
# Source this file: source "$(dirname "$0")/_common.sh"

set -euo pipefail

# ── Paths ─────────────────────────────────────────────────────────────────────
COMMON_SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$COMMON_SCRIPT_DIR/../.." && pwd)"

# ── Colour helpers ────────────────────────────────────────────────────────────
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# ── Shared state ──────────────────────────────────────────────────────────────
SEED_DATA="${SEED_DATA:-false}"
DRY_RUN="${DRY_RUN:-false}"
SKIP_INFRA="${SKIP_INFRA:-false}"

# ── Functions ─────────────────────────────────────────────────────────────────
run_cmd() {
  if [ "$DRY_RUN" = "true" ]; then
    echo -e "${YELLOW}[dry-run]${NC} $1"
  else
    eval "$1"
  fi
}

parse_common_args() {
  while [ "$#" -gt 0 ]; do
    case "$1" in
      --seed) SEED_DATA="true" ;;
      --dry-run) DRY_RUN="true" ;;
      --skip-infra) SKIP_INFRA="true" ;;
      *) echo "Unknown argument: $1"; exit 1 ;;
    esac
    shift
  done
}

# ── Docker Compose helper ─────────────────────────────────────────────────────
DC="docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.services.yml"

# wait_for_service_health polls the gRPC health endpoint of a container using
# grpc_health_probe (built into every distroless service image) until it returns
# SERVING, or fails after a timeout.
# Args: container_name max_attempts
wait_for_service_health() {
  local container="$1"
  local max="${2:-60}"
  local attempt=0
  while [ "$attempt" -lt "$max" ]; do
    if docker exec "$container" /bin/grpc_health_probe -addr=:50051 >/dev/null 2>&1; then
      return 0
    fi
    attempt=$(( attempt + 1 ))
    sleep 1
  done
  echo -e "${YELLOW}[!] $container did not pass health check within ${max}s${NC}"
  return 1
}

# ── Orchestrated startup ──────────────────────────────────────────────────────
# Phase 1: Infrastructure only (Redpanda, ClickHouse, etc.)
# Phase 2: Init topics + schema (while infra is healthy)
# Phase 3: Application services
# Phase 4: Wait for application services to be healthy
start_infrastructure_and_services() {
  cd "$PROJECT_ROOT"

  if [ "$SKIP_INFRA" = "true" ]; then
    echo -e "${CYAN}=== Skipping infrastructure setup (--skip-infra) ===${NC}"
    return 0
  fi

  echo -e "${CYAN}=== Phase 1/4: Starting infrastructure (Redpanda, ClickHouse, observability) ===${NC}"
  run_cmd "$DC up -d --build --wait redpanda clickhouse otel-collector prometheus grafana loki tempo"

  echo -e "${CYAN}=== Phase 2/4: Initialising Redpanda topics and ClickHouse schema ===${NC}"
  run_cmd "bash scripts/dev/init-topics.sh"
  run_cmd "bash scripts/dev/init-clickhouse.sh"

  if [ "$SEED_DATA" = "true" ]; then
    echo -e "${CYAN}=== Seeding ClickHouse demo data ===${NC}"
    run_cmd "bash scripts/demo/seed-demo-data.sh"
  fi

  echo -e "${CYAN}=== Phase 3/4: Starting application services ===${NC}"
  run_cmd "$DC up -d --build"

  echo -e "${CYAN}=== Phase 4/4: Waiting for application services to be healthy ===${NC}"
  local services=(
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
  local all_healthy=true
  for svc in "${services[@]}"; do
    echo -n "  Checking $svc... "
    if wait_for_service_health "$svc" 90; then
      echo -e "${GREEN}healthy${NC}"
    else
      echo -e "${RED}timeout${NC}"
      all_healthy=false
    fi
  done

  if [ "$all_healthy" = "true" ]; then
    echo -e "${GREEN}=== All services healthy ===${NC}"
  else
    echo -e "${YELLOW}=== Some services did not pass health check — demo may be degraded ===${NC}"
  fi

  # Wait for web-cop-gpu frontend container (non-fatal — warn only)
  wait_for_web_cop_gpu || true
}

# ── Web-COP-GPU health check ─────────────────────────────────────────────────
# Waits for the web-cop-gpu container to serve HTTP on port 5173.
wait_for_web_cop_gpu() {
  echo -n "  Checking web-cop-gpu (port 5173)... "
  local attempt=0
  local max=30
  while [ "$attempt" -lt "$max" ]; do
    if curl -sf --max-time 2 http://localhost:5173/ >/dev/null 2>&1; then
      echo -e "${GREEN}healthy${NC}"
      return 0
    fi
    attempt=$(( attempt + 1 ))
    sleep 1
  done
  echo -e "${YELLOW}not reachable (ensure web-cop-gpu container is running or start local dev server)${NC}"
  return 1
}

# ── Open browser helper ───────────────────────────────────────────────────────
open_cop_browser() {
  local url="${1:-http://localhost:5173}"
  if command -v xdg-open &>/dev/null; then
    xdg-open "$url" 2>/dev/null &
  elif command -v open &>/dev/null; then
    open "$url" 2>/dev/null &
  else
    echo -e "${CYAN}  Open COP in browser: ${url}${NC}"
  fi
}

# ── Simulator runner ──────────────────────────────────────────────────────────
run_simulator() {
  local scenario_file="$1"

  echo -e "${CYAN}=== Starting simulator (scenario: ${scenario_file}) ===${NC}"
  run_cmd "$DC run --rm simulator --scenario /app/scenarios/${scenario_file}"
}
