# CLASSIFICATION: UNCLASSIFIED
#!/usr/bin/env bash
# scripts/demo/run-maritime-demo.sh
# Launch the maritime patrol demo scenario (~20 minutes).
#
# Scenario: North Atlantic maritime patrol — UC002, UC003, UC006, UC008, UC009, UC010
# Demonstrates: Radar + AIS + EW/SIGINT ingestion, multi-source fusion, speed anomaly
# detection, AIS spoofing detection, and operator feedback workflow.
#
# Usage: bash scripts/demo/run-maritime-demo.sh [--seed] [--dry-run]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
SEED_DATA="false"
DRY_RUN="false"

GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
NC='\033[0m'

while [ "$#" -gt 0 ]; do
  case "$1" in
    --seed) SEED_DATA="true" ;;
    --dry-run) DRY_RUN="true" ;;
    *) echo "Unknown argument: $1"; exit 1 ;;
  esac
  shift
done

run_cmd() {
  if [ "$DRY_RUN" = "true" ]; then
    echo -e "${YELLOW}[dry-run]${NC} $1"
  else
    eval "$1"
  fi
}

cd "$PROJECT_ROOT"

echo -e "${CYAN}=== RTSA Maritime Demo — Starting infrastructure ===${NC}"
run_cmd "docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.services.yml up -d --build --wait"

echo -e "${CYAN}=== Waiting for services to stabilise (30s) ===${NC}"
run_cmd "sleep 30"

echo -e "${CYAN}=== Initialising Redpanda topics and ClickHouse schema ===${NC}"
run_cmd "bash scripts/dev/init-topics.sh"
run_cmd "bash scripts/dev/init-clickhouse.sh"

if [ "$SEED_DATA" = "true" ]; then
  echo -e "${CYAN}=== Seeding ClickHouse demo data ===${NC}"
  run_cmd "bash scripts/demo/seed-demo-data.sh"
fi

echo ""
echo -e "${CYAN}--- Scenario: North Atlantic Maritime Patrol ---${NC}"
echo "  Duration     : ~20 minutes"
echo "  Use Cases    : UC002 (Radar), UC003 (EW/SIGINT), UC006 (AIS), UC008 (Fusion),"
echo "                 UC009 (Anomaly Detection), UC010 (Operator Feedback)"
echo "  Entry Point  : http://localhost:5173"
echo "  Suggested UI : Operations Commander → Fusion Dashboard (default)"
echo ""
echo "  Timeline of events:"
echo "    T+0:00  Radar detects USV at 60.2°N 9.7°W — 12 knots, heading 180°"
echo "    T+1:30  AIS signal near USV — MMSI 987654321 (UNKNOWN-AIS, possible spoof)"
echo "    T+3:00  EW intercept — encrypted comms 8.5 GHz correlated to USV"
echo "    T+3:15  Fusion creates TRK-0001 (confidence 0.82, 4 sensors)"
echo "    T+5:00  CRITICAL ALERT — speed anomaly, USV accelerates to 28 knots"
echo "    T+6:30  CRITICAL ALERT — AIS spoofing, position delta 1.2 NM from radar"
echo "    T+8:00  Friendly vessel MMSI 123456789 detected, TRK-0002 created"
echo "    T+10:00 Track merge — TRK-0001 and radar ghost merged (confidence 0.87)"
echo "    T+12:00 Operator feedback submitted — CONFIRM_ANOMALY, trust 0.85"
echo "    T+18:00 Vessel exits bounding box — TRK-0001 status: STALE"
echo ""

echo -e "${CYAN}=== Starting maritime scenario simulator ===${NC}"
run_cmd "docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.services.yml \
  run --rm simulator \
  --scenario /app/scenarios/maritime-demo.yaml \
  --log-level info"

echo -e "${GREEN}=== Maritime demo complete ===${NC}"
echo "  Check the UI at http://localhost:5173 to see the full track picture."
echo "  Stop with: bash scripts/demo/stop-demo.sh --volumes"
