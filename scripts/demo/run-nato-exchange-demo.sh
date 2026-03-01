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
# Usage: bash scripts/demo/run-nato-exchange-demo.sh [--seed] [--dry-run] [--skip-infra]

# shellcheck source=scripts/demo/_common.sh
source "$(cd "$(dirname "$0")" && pwd)/_common.sh"
parse_common_args "$@"

start_infrastructure_and_services

# Seed data is required for this scenario — provides inbound NATO tracks and nomination queue
if [ "$SEED_DATA" != "true" ]; then
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

run_simulator "nato-exchange-demo.yaml"

echo -e "${GREEN}=== NATO Exchange demo complete ===${NC}"
echo "  What to show in the UI:"
echo "  - NATO Exchange dashboard connectivity header (Link 16 + NFFI status)"
echo "  - Track Nomination Queue with [Nominate] and [Revoke] buttons"
echo "  - Map shows nominated tracks with NATO compass-star overlay icon"
echo "  - Classification guard BLOCKED entry in nomination queue"
echo "  - Inbound allied tracks with REL TO FVEY markers on map"
echo "  Stop with: bash scripts/demo/stop-demo.sh --volumes"
