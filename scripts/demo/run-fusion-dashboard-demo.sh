# CLASSIFICATION: UNCLASSIFIED
#!/usr/bin/env bash
# scripts/demo/run-fusion-dashboard-demo.sh
# Launch the Fusion Dashboard deep-dive scenario (~15 minutes).
#
# Scenario: Sensor-by-sensor startup that makes the multi-source fusion algorithm visible.
# Use Cases: UC016 (Fusion Dashboard), UC008 (Multi-Source Fusion), UC012 (UI)
# Demonstrates: StreamSensorObservations v2.0 RPC, raw sensor icons alongside fused tracks,
# FusionSidePanel confidence histograms, track list with per-sensor attribution.
#
# Usage: bash scripts/demo/run-fusion-dashboard-demo.sh [--seed] [--dry-run]

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

echo -e "${CYAN}=== RTSA Fusion Dashboard Demo — Starting infrastructure ===${NC}"
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
echo -e "${CYAN}--- Scenario: Fusion Algorithm Transparency ---${NC}"
echo "  Duration     : ~15 minutes"
echo "  Use Cases    : UC016, UC008, UC012"
echo "  Entry Point  : http://localhost:5173"
echo "  Suggested UI : Operations Commander → Fusion Dashboard (default)"
echo ""
echo "  v2.0 APIs exercised:"
echo "    - TrackService.StreamSensorObservations (new v2.0 server-streaming RPC)"
echo "    - track-svc-sensor-stream consumer group (new dual consumer in svc-track)"
echo "    - FusionSidePanel: domain counts, confidence histogram, active track list"
echo ""
echo "  Scenario timeline (sensors start sequentially):"
echo "    T+0  : Radar only — raw diamond plots, no fused tracks yet"
echo "    T+2  : EW/SIGINT online — amber triangle intercepts near radar plots"
echo "    T+4  : AIS feed starts — green circle positions near EW contacts"
echo "    T+6  : First fused track (confidence >=0.85, 3-sensor correlation)"
echo "    T+8  : ISR corroboration — confidence increases to 0.93"
echo "    T+10 : Deliberate conflicting observation — confidence drops, tentative flag"
echo ""
echo -e "${CYAN}  Raw sensor icon legend:${NC}"
echo "    Light-blue diamond : Radar plot (StreamSensorObservations — RADAR type)"
echo "    Amber triangle     : EW/SIGINT intercept (EW_SIGINT type)"
echo "    Purple square      : ELINT detection (ELINT_COMINT type)"
echo "    White pentagon     : ISR observation (ISR type)"
echo "    Green circle       : AIS position (AIS_BFT type)"
echo "    Red hexagon        : Cyber IOC (CYBER type)"
echo ""

echo -e "${CYAN}=== Starting fusion dashboard scenario simulator ===${NC}"
run_cmd "docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.services.yml \
  run --rm simulator \
  --scenario /app/scenarios/fusion-dashboard-demo.yaml \
  --log-level info"

echo -e "${GREEN}=== Fusion Dashboard demo complete ===${NC}"
echo "  Key things to point out in the UI:"
echo "  - Raw sensor icons (diamond/triangle/square) around fused track circles"
echo "  - FusionSidePanel shows domain counts, confidence histogram, and track list"
echo "  - Clicking a track in the list opens the Entity Detail Panel with Source Attribution"
echo "  Stop with: bash scripts/demo/stop-demo.sh --volumes"
