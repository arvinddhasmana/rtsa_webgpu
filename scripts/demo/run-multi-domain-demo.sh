# CLASSIFICATION: UNCLASSIFIED
#!/usr/bin/env bash
# scripts/demo/run-multi-domain-demo.sh
# Launch the multi-domain combined-arms exercise scenario (~30 minutes).
#
# Scenario: Combined-arms exercise — all five domains (Air, Surface, Subsurface, Land, Cyber)
# Use Cases: UC002–UC009 (all ingestion + fusion + detection), UC012 (UI), UC016 (Fusion Dashboard)
# Demonstrates: High-density multi-domain track picture, sensor coverage overlays, domain metrics.
#
# Usage: bash scripts/demo/run-multi-domain-demo.sh [--seed] [--dry-run]

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

echo -e "${CYAN}=== RTSA Multi-Domain Demo — Starting infrastructure ===${NC}"
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
echo -e "${CYAN}--- Scenario: Combined-Arms Multi-Domain Exercise ---${NC}"
echo "  Duration     : ~30 minutes"
echo "  Use Cases    : UC002–UC009, UC012, UC016"
echo "  Entry Point  : http://localhost:5173"
echo "  Suggested UI : Operations Commander → Multi-Domain Dashboard"
echo ""
echo "  Domain Entity Summary:"
echo "    Air        : 8 tracks (4 friendly F-35, 2 unknown UAV, 2 suspect cruise-missile)"
echo "    Surface    : 12 tracks (6 allied, 3 merchant, 2 fast-mover, 1 spoofed AIS)"
echo "    Subsurface : 2 tracks (acoustic contact, TENTATIVE classification)"
echo "    Land       : 5 tracks (ISR vehicle convoys, 1 route deviation)"
echo "    Cyber      : 3 IOCs (C2 domain, lateral movement, data exfiltration)"
echo ""
echo "  Sensor Coverage (visible on Multi-Domain map):"
echo "    RADAR-NORTH-01 : 150 NM fan sector 000°–120°"
echo "    RADAR-SOUTH-02 : 120 NM fan sector 090°–270°"
echo "    EW-STATION-01  : 200 NM omnidirectional"
echo "    EW-STATION-03  : 180 NM omnidirectional"
echo "    ISR-UAV-ALPHA  : 50x50 NM swath polygon (dynamic)"
echo ""

echo -e "${CYAN}=== Starting multi-domain scenario simulator ===${NC}"
run_cmd "docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.services.yml \
  run --rm simulator \
  --scenario /app/scenarios/multi-domain-demo.yaml \
  --log-level info"

echo -e "${GREEN}=== Multi-domain demo complete ===${NC}"
echo "  Recommended next step: switch to Fusion Dashboard to see raw sensor icons."
echo "  Stop with: bash scripts/demo/stop-demo.sh --volumes"
