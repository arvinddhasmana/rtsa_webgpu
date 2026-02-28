# CLASSIFICATION: UNCLASSIFIED
#!/usr/bin/env bash
# scripts/demo/run-demo-switch.sh
# Convenience wrapper to switch between any RTSA demo scenario (v2.0).
#
# Forwards all arguments to run-demo.sh. Supports all v2.0 scenarios:
#   maritime, multi-domain, fusion-dashboard, operator-ui,
#   sensor-health, nato-exchange, analyst-forensics, full-suite
#
# Usage:
#   bash scripts/demo/run-demo-switch.sh [SCENARIO] [OPTIONS]
#
# Examples:
#   bash scripts/demo/run-demo-switch.sh maritime --seed
#   bash scripts/demo/run-demo-switch.sh fusion-dashboard --setup --seed
#   bash scripts/demo/run-demo-switch.sh full-suite --seed --stop-on-complete

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
cd "$PROJECT_ROOT"

# Forward all arguments to the main demo runner
bash scripts/demo/run-demo.sh "$@"
