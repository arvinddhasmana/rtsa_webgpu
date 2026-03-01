# CLASSIFICATION: UNCLASSIFIED
#!/usr/bin/env bash
# scripts/demo/run-maritime-demo.sh
# Launch the maritime patrol demo scenario (~20 minutes).
#
# Scenario: North Atlantic maritime patrol — UC002, UC003, UC006, UC008, UC009, UC010
# Demonstrates: Radar + AIS + EW/SIGINT ingestion, multi-source fusion, speed anomaly
# detection, AIS spoofing detection, and operator feedback workflow.
#
# Usage: bash scripts/demo/run-maritime-demo.sh [--seed] [--dry-run] [--skip-infra]

# shellcheck source=scripts/demo/_common.sh
source "$(cd "$(dirname "$0")" && pwd)/_common.sh"
parse_common_args "$@"

start_infrastructure_and_services

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

run_simulator "maritime-demo.yaml"

echo -e "${GREEN}=== Maritime demo complete ===${NC}"
echo "  Check the UI at http://localhost:5173 to see the full track picture."
echo "  Stop with: bash scripts/demo/stop-demo.sh --volumes"
