# CLASSIFICATION: UNCLASSIFIED
#!/usr/bin/env bash
# scripts/demo/run-sensor-health-demo.sh
# Launch the Sensor Health Monitoring demo scenario (~10 minutes).
#
# Scenario: All sensors healthy at start; deliberate degradation and recovery injected.
# Use Cases: UC017 (Sensor Health Monitoring), UC001 (System Initialization)
# Demonstrates: SensorHealthDashboard, SensorStatusCard, ListSensorStatuses v2.0 RPC,
# SensorCoverageLayer fan sectors and arcs, live acceptance rate monitoring.
#
# Usage: bash scripts/demo/run-sensor-health-demo.sh [--seed] [--dry-run] [--skip-infra]

# shellcheck source=scripts/demo/_common.sh
source "$(cd "$(dirname "$0")" && pwd)/_common.sh"
parse_common_args "$@"

start_infrastructure_and_services

echo ""
echo -e "${CYAN}--- Scenario: Sensor Degradation Detection and Recovery ---${NC}"
echo "  Duration     : ~10 minutes"
echo "  Use Cases    : UC017, UC001"
echo "  Entry Point  : http://localhost:5173"
echo "  Suggested UI : Sensor Operator (auto-selects Sensor Health dashboard)"
echo ""
echo "  v2.0 APIs exercised:"
echo "    - IngestionService.ListSensorStatuses (new v2.0 bulk RPC)"
echo "    - SensorStatusResponse.coverage (new SensorCoverage geometry field)"
echo "    - SensorCoverageLayer MapLibre layers (fan sectors, arcs, polygons)"
echo ""
echo "  Registered sensors in this demo:"
echo "    RADAR-NORTH-01  : 150 NM, sector 000°–120° (60.5°N 8.2°W)"
echo "    RADAR-SOUTH-02  : 120 NM, sector 090°–270° (55.3°N 12.1°W)"
echo "    EW-STATION-01   : 200 NM omnidirectional (59.1°N 9.8°W)"
echo "    EW-STATION-03   : 180 NM omnidirectional (56.7°N 11.3°W)"
echo "    ELINT-ARRAY-01  : 300 NM omnidirectional (58.4°N 7.6°W)"
echo "    ISR-UAV-ALPHA   : 50x50 NM swath polygon (airborne, dynamic)"
echo "    AIS-COAST-01    : 40 NM VHF range (61.0°N 10.5°W)"
echo "    AIS-COAST-02    : 40 NM VHF range (54.9°N 13.0°W)"
echo ""
echo "  Scenario timeline:"
echo "    T+0  : All 12 sensors healthy — green status dots, nominal events/sec"
echo "    T+2  : RADAR-NORTH-01 degrades — events/sec drops 80%, card turns amber"
echo "    T+5  : EW-STATION-03 disconnects — card turns red, coverage arc disappears"
echo "    T+6  : Sensor degradation alert injected into alert panel"
echo "    T+8  : Both sensors recover — green status restored, coverage overlays return"
echo ""
echo "  Sensor card information displayed:"
echo "    - sensor_id and type icon"
echo "    - Connection status (green=connected, amber=degraded, red=disconnected)"
echo "    - events/sec (live observation rate)"
echo "    - total_received counter"
echo "    - last_observation_time timestamp"
echo "    - acceptance_rate = total_accepted / total_received * 100 (percentage bar)"
echo "    - Coverage geometry link → click to centre map"
echo ""

run_simulator "sensor-health-demo.yaml"

echo -e "${GREEN}=== Sensor Health demo complete ===${NC}"
echo "  What happened:"
echo "  - RADAR-NORTH-01 degraded and recovered (amber → green)"
echo "  - EW-STATION-03 disconnected and recovered (red → green)"
echo "  - SensorCoverageLayer showed coverage gaps during degradation"
echo "  Stop with: bash scripts/demo/stop-demo.sh --volumes"
