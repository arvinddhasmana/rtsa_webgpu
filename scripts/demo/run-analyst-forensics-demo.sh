# CLASSIFICATION: UNCLASSIFIED
#!/usr/bin/env bash
# scripts/demo/run-analyst-forensics-demo.sh
# Launch the Intelligence Analyst Forensics + Intel Search demo scenario (~20 minutes).
#
# Scenario: Historical forensic investigation of a suspect vessel with 72-hour track history.
# Use Cases: UC013 (Historical Data Query), UC010 (Operator Feedback), UC011 (Model Retraining)
# Demonstrates: Forensics dashboard, Intel Search (by MMSI/callsign), GetEventTimeline v2.0 RPC,
# ClickHouse time-range and spatial queries, feedback submission with trust scoring.
#
# Usage: bash scripts/demo/run-analyst-forensics-demo.sh [--seed] [--dry-run]

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

echo -e "${CYAN}=== RTSA Analyst Forensics Demo — Starting infrastructure ===${NC}"
run_cmd "docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.services.yml up -d --build --wait"

echo -e "${CYAN}=== Waiting for services to stabilise (30s) ===${NC}"
run_cmd "sleep 30"

echo -e "${CYAN}=== Initialising Redpanda topics and ClickHouse schema ===${NC}"
run_cmd "bash scripts/dev/init-topics.sh"
run_cmd "bash scripts/dev/init-clickhouse.sh"

# Seed data is required — provides the 72-hour track history for forensics investigation
if [ "$SEED_DATA" = "true" ]; then
  echo -e "${CYAN}=== Seeding ClickHouse demo data (72-hour historical dataset) ===${NC}"
  run_cmd "bash scripts/demo/seed-demo-data.sh"
else
  echo -e "${YELLOW}[!] Note: --seed not specified. Analyst Forensics demo requires seeded data.${NC}"
  echo -e "${YELLOW}    Re-run with: bash scripts/demo/run-analyst-forensics-demo.sh --seed${NC}"
fi

echo ""
echo -e "${CYAN}--- Scenario: Forensic Investigation — Suspect Vessel MMSI 123456789 ---${NC}"
echo "  Duration     : ~20 minutes"
echo "  Use Cases    : UC013, UC010, UC011"
echo "  Entry Point  : http://localhost:5173"
echo "  Suggested UI : Intelligence Analyst (auto-selects Forensics dashboard)"
echo ""
echo "  v2.0 APIs exercised:"
echo "    - QueryService.GetEventTimeline (new v2.0 UNION ALL across 4 ClickHouse tables)"
echo "    - QueryService.QueryTracks (parameterized ClickHouse query with classification filter)"
echo "    - QueryService.QueryAnomalies (historical anomaly detection records)"
echo "    - FeedbackService.SubmitFeedback (with trust score returned in response)"
echo ""
echo "  Pre-seeded forensic dataset (from seed-demo-data.sh):"
echo "    Entity       : Vessel MMSI 123456789 (track TRK-0002 and historical trail)"
echo "    Time window  : 72 hours of track history"
echo "    Anomaly window: 4-hour anomalous activity period within the 72-hour dataset"
echo "    Data sources : tracks_fused, sensor_observations (4 sensor types), anomaly_detections,"
echo "                   operator_feedback, audit_log"
echo ""
echo "  Scenario timeline (analyst-driven actions):"
echo "    T+0   : Forensics panel open — time range set to last 72 hours, Surface domain"
echo "    T+2   : Filter by MMSI 123456789 — entity timeline opens for this vessel"
echo "    T+3   : Entity timeline shows: detection → pattern deviation → ELINT intercept → feedback"
echo "    T+5   : Switch to Intel Search — search by callsign VIPER-01 (air track TRK-0004)"
echo "    T+8   : Analyst submits Confirm Hostile feedback — trust score returned and displayed"
echo "    T+12  : Model retraining log shown — batch threshold reached, retrain triggered (UC011)"
echo ""
echo "  Intel Search capabilities:"
echo "    - Search by MMSI (exact): 123456789 → surface tracks"
echo "    - Search by callsign (partial): VIPER → air tracks"
echo "    - Search by track_id (prefix): TRK-0 → all seeded tracks"
echo "    - Search by sensor_id: RADAR-NORTH-01 → all observations from that sensor"
echo ""

echo -e "${CYAN}=== Starting analyst forensics scenario simulator ===${NC}"
run_cmd "docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.services.yml \
  run --rm simulator \
  --scenario /app/scenarios/analyst-forensics-demo.yaml \
  --log-level info"

echo -e "${GREEN}=== Analyst Forensics demo complete ===${NC}"
echo "  Key things to point out in the UI:"
echo "  - Forensics panel time-range filter and entity type filter"
echo "  - Entity timeline showing 4-source event interleaving via GetEventTimeline"
echo "  - Intel Search tab — multi-field entity lookup"
echo "  - Feedback form in Entity Detail Panel — trust score returned on submit"
echo "  Stop with: bash scripts/demo/stop-demo.sh --volumes"
