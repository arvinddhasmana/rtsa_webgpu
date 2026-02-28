# CLASSIFICATION: UNCLASSIFIED
#!/usr/bin/env bash
# scripts/demo/run-nato-exchange-demo.sh
# Launch the NATO Data Exchange demo scenario (~15 minutes).
#
# Scenario: Bidirectional NATO allied data exchange via Link 16, NFFI, and MIP formats.
# Use Cases: UC014 (NATO Outbound), UC015 (NATO Inbound)
# Demonstrates: NATOExchangeDashboard, connectivity header, track nomination queue,
# cross-domain classification guard, inbound allied tracks with REL TO markings.
#
# Usage: bash scripts/demo/run-nato-exchange-demo.sh [--seed] [--dry-run]

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

echo -e "${CYAN}=== RTSA NATO Exchange Demo — Starting infrastructure ===${NC}"
run_cmd "docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.services.yml up -d --build --wait"

echo -e "${CYAN}=== Waiting for services to stabilise (30s) ===${NC}"
run_cmd "sleep 30"

echo -e "${CYAN}=== Initialising Redpanda topics and ClickHouse schema ===${NC}"
run_cmd "bash scripts/dev/init-topics.sh"
run_cmd "bash scripts/dev/init-clickhouse.sh"

# Seed data is required for this scenario — provides inbound NATO tracks and nomination queue
if [ "$SEED_DATA" = "true" ]; then
  echo -e "${CYAN}=== Seeding ClickHouse demo data (includes NATO records) ===${NC}"
  run_cmd "bash scripts/demo/seed-demo-data.sh"
else
  echo -e "${YELLOW}[!] Note: --seed not specified. NATO Exchange demo requires seeded data.${NC}"
  echo -e "${YELLOW}    Re-run with: bash scripts/demo/run-nato-exchange-demo.sh --seed${NC}"
fi

echo ""
echo -e "${CYAN}--- Scenario: NATO Bidirectional Data Exchange ---${NC}"
echo "  Duration     : ~15 minutes"
echo "  Use Cases    : UC014 (NATO Outbound), UC015 (NATO Inbound)"
echo "  Entry Point  : http://localhost:5173"
echo "  Suggested UI : NATO Liaison (auto-selects NATO Exchange dashboard)"
echo ""
echo "  Protocol formats demonstrated:"
echo "    - STANAG 5516 Link 16 J-Series messages (outbound track reports)"
echo "    - NFFI (NATO Friendly Force Information) XML (bidirectional)"
echo "    - MIP (Multilateral Interoperability Programme) data share"
echo ""
echo "  Pre-seeded NATO data (from seed-demo-data.sh):"
echo "    Inbound tracks   : 5 allied tracks with REL TO FVEY classification"
echo "    Nomination queue : 3 organic tracks awaiting outbound review"
echo ""
echo "  Scenario timeline:"
echo "    T+0   : Dashboard shows connectivity: Link 16 connected, NFFI connected"
echo "    T+0   : 5 inbound allied tracks visible on map (NATO icon, REL TO label)"
echo "    T+0   : 3 organic tracks in nomination queue"
echo "    T+3   : NATO partner acknowledgment — nominated tracks show Acknowledged status"
echo "    T+5   : New inbound allied track arrives — appears on map with NATO icon"
echo "    T+8   : Cross-domain guard blocks one outbound — classification ceiling exceeded"
echo "    T+12  : Track revocation — nominated track recalled from sharing queue"
echo ""
echo "  Classification guard rules (enforced by svc-nato-adapter):"
echo "    - Maximum outbound classification: NATO UNCLASSIFIED"
echo "    - Inbound tracks mapped to internal classification levels"
echo "    - All exchange actions logged to audit_log with resource_type=nato_nomination"
echo ""

echo -e "${CYAN}=== Starting NATO exchange scenario simulator ===${NC}"
run_cmd "docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.services.yml \
  run --rm simulator \
  --scenario /app/scenarios/nato-exchange-demo.yaml \
  --log-level info"

echo -e "${GREEN}=== NATO Exchange demo complete ===${NC}"
echo "  What to show in the UI:"
echo "  - NATO Exchange dashboard connectivity header (Link 16 + NFFI status)"
echo "  - Track Nomination Queue with [Nominate] and [Revoke] buttons"
echo "  - Map shows nominated tracks with NATO compass-star overlay icon"
echo "  - Classification guard BLOCKED entry in nomination queue"
echo "  - Inbound allied tracks with REL TO FVEY markers on map"
echo "  Stop with: bash scripts/demo/stop-demo.sh --volumes"
