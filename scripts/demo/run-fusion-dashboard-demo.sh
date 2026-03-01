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
# Usage: bash scripts/demo/run-fusion-dashboard-demo.sh [--seed] [--dry-run] [--skip-infra]

# shellcheck source=scripts/demo/_common.sh
source "$(cd "$(dirname "$0")" && pwd)/_common.sh"
parse_common_args "$@"

start_infrastructure_and_services

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

run_simulator "fusion-dashboard-demo.yaml"

echo -e "${GREEN}=== Fusion Dashboard demo complete ===${NC}"
echo "  Key things to point out in the UI:"
echo "  - Raw sensor icons (diamond/triangle/square) around fused track circles"
echo "  - FusionSidePanel shows domain counts, confidence histogram, and track list"
echo "  - Clicking a track in the list opens the Entity Detail Panel with Source Attribution"
echo "  Stop with: bash scripts/demo/stop-demo.sh --volumes"
