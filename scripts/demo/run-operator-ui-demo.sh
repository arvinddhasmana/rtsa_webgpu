# CLASSIFICATION: UNCLASSIFIED
#!/usr/bin/env bash
# scripts/demo/run-operator-ui-demo.sh
# Launch the Operator UI + Event Timeline demo scenario (~15 minutes).
#
# Scenario: Alert workflow and unified event timeline demonstration.
# Use Cases: UC010 (Operator Feedback), UC012 (Situational Awareness UI), UC013 (Historical Query)
# Demonstrates: AlertCard quick-actions (Inspect/Confirm/Reject/Assign), TimelineView,
# GetEventTimeline v2.0 RPC, AssignAlert v2.0 RPC, blurred-map Operator UI layout.
#
# Usage: bash scripts/demo/run-operator-ui-demo.sh [--seed] [--dry-run]

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

echo -e "${CYAN}=== RTSA Operator UI Demo — Starting infrastructure ===${NC}"
run_cmd "docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.services.yml up -d --wait"

echo -e "${CYAN}=== Waiting for services to stabilise (30s) ===${NC}"
run_cmd "sleep 30"

echo -e "${CYAN}=== Initialising Redpanda topics and ClickHouse schema ===${NC}"
run_cmd "bash scripts/dev/init-topics.sh"
run_cmd "bash scripts/dev/init-clickhouse.sh"

# Seed data is strongly recommended for this scenario (pre-loads alerts and timeline events)
if [ "$SEED_DATA" = "true" ]; then
  echo -e "${CYAN}=== Seeding ClickHouse demo data ===${NC}"
  run_cmd "bash scripts/demo/seed-demo-data.sh"
else
  echo -e "${YELLOW}[!] Note: --seed not specified. Operator UI demo works best with pre-seeded data.${NC}"
  echo -e "${YELLOW}    Re-run with: bash scripts/demo/run-operator-ui-demo.sh --seed${NC}"
fi

echo ""
echo -e "${CYAN}--- Scenario: Operator Alert Workflow and Timeline ---${NC}"
echo "  Duration     : ~15 minutes"
echo "  Use Cases    : UC010, UC012, UC013"
echo "  Entry Point  : http://localhost:5173"
echo "  Suggested UI : Operations Commander → Operator UI (Level-2 tab)"
echo ""
echo "  v2.0 APIs exercised:"
echo "    - QueryService.GetEventTimeline (new v2.0 UNION ALL query across 4 ClickHouse tables)"
echo "    - AlertService.AssignAlert (new v2.0 RPC with audit event)"
echo "    - FeedbackService.SubmitFeedback (CONFIRM_ANOMALY / REJECT_ANOMALY)"
echo "    - DetailPanel: SourceAttribution, EntityTimeline, FeedbackForm (all new in v2.0)"
echo ""
echo "  Scenario timeline:"
echo "    T+0   : 5 CRITICAL + 10 ELEVATED alerts pre-loaded"
echo "    T+2   : New CRITICAL alert — USV speed anomaly"
echo "    T+5   : New ELEVATED alert — route deviation"
echo "    T+6   : Track merge event — timeline shows dual-source correlation marker"
echo "    T+9   : New CRITICAL alert — AIS spoofing detection"
echo "    T+10  : Alert assignment event — AssignAlert RPC, badge appears on alert card"
echo "    T+13  : Feedback-confirmed alert resolved, closed in alert panel"
echo ""
echo "  Alert quick-action buttons:"
echo "    [Inspect] — opens Entity Detail Panel (SourceAttribution + EntityTimeline)"
echo "    [Confirm] — calls FeedbackService.SubmitFeedback(CONFIRM_ANOMALY)"
echo "    [Reject]  — calls FeedbackService.SubmitFeedback(REJECT_ANOMALY)"
echo "    [Assign]  — calls AlertService.AssignAlert (v2.0) + produces audit event"
echo ""

echo -e "${CYAN}=== Starting operator UI scenario simulator ===${NC}"
run_cmd "docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.services.yml \
  run --rm simulator \
  --scenario /app/scenarios/operator-ui-demo.yaml \
  --log-level info"

echo -e "${GREEN}=== Operator UI demo complete ===${NC}"
echo "  Key things to demonstrate in the UI:"
echo "  - Map is blurred in background (CSS backdrop-filter)"
echo "  - AlertCard quick-action buttons: [Inspect] [Confirm] [Reject] [Assign]"
echo "  - Entity Detail Panel: Source Attribution table, Entity Timeline, Feedback Form"
echo "  - Timeline on right: interleaved events from tracks, anomalies, feedback, audit"
echo "  Stop with: bash scripts/demo/stop-demo.sh --volumes"
