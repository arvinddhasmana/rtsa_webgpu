# CLASSIFICATION: UNCLASSIFIED
#!/usr/bin/env bash
# scripts/demo/run-demo.sh
# One-command entrypoint for RTSA demo setup + scenario run (v2.0).
#
# Usage:
#   bash scripts/demo/run-demo.sh [SCENARIO] [OPTIONS]
#
# Scenarios:
#   maritime            Maritime patrol — UC002, UC006, UC008, UC009, UC010
#   multi-domain        All-domain exercise — UC002–UC009, UC012, UC016
#   fusion-dashboard    Fusion Dashboard deep-dive — UC016, UC008, UC012
#   operator-ui         Operator UI + Timeline — UC010, UC012, UC013
#   sensor-health       Sensor Health + Coverage — UC017, UC001
#   nato-exchange       NATO outbound/inbound — UC014, UC015
#   analyst-forensics   Forensics + Intel Search — UC013, UC010, UC011
#   full-suite          All scenarios in sequence — UC001–UC017
#
# Options:
#   --setup             Run scripts/setup/setup-dev.sh before demo launch
#   --seed              Seed ClickHouse with demo data before scenario
#   --dry-run           Print commands without executing
#   --stop-on-complete  Run stop-demo.sh --volumes after scenario completes
#   --quick-switch      Skip optional setup step and launch directly
#   -h, --help          Show this help

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

SCENARIO="maritime"
RUN_SETUP="false"
SEED_DATA="false"
DRY_RUN="false"
STOP_ON_COMPLETE="false"

# Colour helpers
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
NC='\033[0m'

usage() {
  cat <<'USAGE'
RTSA Demo Runner (v2.0)

Usage:
  bash scripts/demo/run-demo.sh [SCENARIO] [OPTIONS]

Scenarios:
  maritime            Maritime patrol — UC002, UC006, UC008, UC009, UC010
  multi-domain        All-domain exercise — UC002–UC009, UC012, UC016
  fusion-dashboard    Fusion Dashboard deep-dive — UC016, UC008, UC012
  operator-ui         Operator UI + Timeline — UC010, UC012, UC013
  sensor-health       Sensor Health + Coverage monitoring — UC017, UC001
  nato-exchange       NATO outbound/inbound data exchange — UC014, UC015
  analyst-forensics   Analyst Forensics + Intel Search — UC013, UC010, UC011
  full-suite          All scenarios in sequence — UC001–UC017 (~60 min)

Options:
  --setup             Run scripts/setup/setup-dev.sh before demo launch
  --seed              Seed ClickHouse with representative demo data before scenario
  --dry-run           Print commands without executing
  --stop-on-complete  Run stop-demo.sh --volumes after scenario completes
  --quick-switch      Skip optional setup; launch scenario directly
  -h, --help          Show this help

Examples:
  bash scripts/demo/run-demo.sh maritime --seed
  bash scripts/demo/run-demo.sh fusion-dashboard --setup --seed
  bash scripts/demo/run-demo.sh full-suite --seed --stop-on-complete
USAGE
}

run_cmd() {
  local cmd="$1"
  if [ "$DRY_RUN" = "true" ]; then
    echo -e "${YELLOW}[dry-run]${NC} $cmd"
  else
    eval "$cmd"
  fi
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    maritime|multi-domain|fusion-dashboard|operator-ui|sensor-health|nato-exchange|analyst-forensics|full-suite)
      SCENARIO="$1"
      ;;
    --setup)
      RUN_SETUP="true"
      ;;
    --seed)
      SEED_DATA="true"
      ;;
    --dry-run)
      DRY_RUN="true"
      ;;
    --stop-on-complete)
      STOP_ON_COMPLETE="true"
      ;;
    --quick-switch)
      RUN_SETUP="false"
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

echo -e "${CYAN}=== RTSA Demo Runner v2.0 — Scenario: ${SCENARIO} ===${NC}"

if [ "$RUN_SETUP" = "true" ]; then
  echo -e "${CYAN}[1/4] Running first-time environment setup...${NC}"
  run_cmd "bash scripts/setup/setup-dev.sh"
fi

echo -e "${CYAN}[2/4] Starting infrastructure and services...${NC}"
# Source _common.sh for start_infrastructure_and_services
source "$SCRIPT_DIR/_common.sh"
start_infrastructure_and_services

echo -e "${CYAN}[3/4] Seeding and scenario preparation complete.${NC}"

echo -e "${CYAN}[4/4] Launching scenario: ${SCENARIO}${NC}"

# Build child script args: pass --skip-infra to avoid duplicate setup;
# forward --seed and --dry-run if they were set.
CHILD_ARGS="--skip-infra"
if [ "$SEED_DATA" = "true" ]; then
  CHILD_ARGS="$CHILD_ARGS --seed"
fi
if [ "$DRY_RUN" = "true" ]; then
  CHILD_ARGS="$CHILD_ARGS --dry-run"
fi

case "$SCENARIO" in
  maritime)
    run_cmd "bash scripts/demo/run-maritime-demo.sh $CHILD_ARGS"
    ;;
  multi-domain)
    run_cmd "bash scripts/demo/run-multi-domain-demo.sh $CHILD_ARGS"
    ;;
  fusion-dashboard)
    run_cmd "bash scripts/demo/run-fusion-dashboard-demo.sh $CHILD_ARGS"
    ;;
  operator-ui)
    run_cmd "bash scripts/demo/run-operator-ui-demo.sh $CHILD_ARGS"
    ;;
  sensor-health)
    run_cmd "bash scripts/demo/run-sensor-health-demo.sh $CHILD_ARGS"
    ;;
  nato-exchange)
    run_cmd "bash scripts/demo/run-nato-exchange-demo.sh $CHILD_ARGS"
    ;;
  analyst-forensics)
    run_cmd "bash scripts/demo/run-analyst-forensics-demo.sh $CHILD_ARGS"
    ;;
  full-suite)
    echo -e "${CYAN}=== Full Suite: running all 7 scenarios in sequence (~60 min) ===${NC}"
    run_cmd "bash scripts/demo/run-maritime-demo.sh $CHILD_ARGS"
    run_cmd "sleep 300"  # 5-minute transition
    run_cmd "bash scripts/demo/run-fusion-dashboard-demo.sh $CHILD_ARGS"
    run_cmd "sleep 300"
    run_cmd "bash scripts/demo/run-operator-ui-demo.sh $CHILD_ARGS"
    run_cmd "sleep 300"
    run_cmd "bash scripts/demo/run-multi-domain-demo.sh $CHILD_ARGS"
    run_cmd "sleep 300"
    run_cmd "bash scripts/demo/run-sensor-health-demo.sh $CHILD_ARGS"
    run_cmd "sleep 300"
    run_cmd "bash scripts/demo/run-nato-exchange-demo.sh $CHILD_ARGS"
    run_cmd "sleep 300"
    run_cmd "bash scripts/demo/run-analyst-forensics-demo.sh $CHILD_ARGS"
    ;;
esac

echo -e "${GREEN}=== Demo run completed for scenario: ${SCENARIO} ===${NC}"
echo ""
echo "  COP (WebGPU)   : http://localhost:5173  (Chrome 113+ / Edge 113+ required)"
echo "  API Gateway    : https://localhost:8443"
echo "  Redpanda Admin : http://localhost:8080"
echo "  ClickHouse UI  : http://localhost:8123/play"
echo ""
echo "  Tip: services remain running for live UI exploration."
echo "  Stop with: bash scripts/demo/stop-demo.sh --volumes"

# Open COP in default browser
open_cop_browser "http://localhost:5173"

if [ "$STOP_ON_COMPLETE" = "true" ]; then
  echo -e "${CYAN}--stop-on-complete flag set. Tearing down...${NC}"
  run_cmd "bash scripts/demo/stop-demo.sh --volumes"
fi
