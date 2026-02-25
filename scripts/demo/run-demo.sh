# CLASSIFICATION: UNCLASSIFIED
#!/usr/bin/env bash
# scripts/demo/run-demo.sh
# One-command entrypoint for RTSA demo setup + scenario run.
# Usage: bash scripts/demo/run-demo.sh [maritime|multi-domain] [--setup] [--dry-run] [--stop-on-complete]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

SCENARIO="maritime"
RUN_SETUP="false"
DRY_RUN="false"
STOP_ON_COMPLETE="false"

usage() {
  cat <<'USAGE'
RTSA Demo Runner

Usage:
  bash scripts/demo/run-demo.sh [maritime|multi-domain] [options]

Options:
  --setup             Run scripts/setup/setup-dev.sh before demo launch
  --dry-run           Print commands without executing
  --stop-on-complete  Run scripts/demo/stop-demo.sh after scenario completes
  -h, --help          Show this help
USAGE
}

run_cmd() {
  local cmd="$1"
  if [ "$DRY_RUN" = "true" ]; then
    echo "[dry-run] $cmd"
  else
    eval "$cmd"
  fi
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    maritime)
      SCENARIO="maritime"
      ;;
    multi-domain)
      SCENARIO="multi-domain"
      ;;
    --setup)
      RUN_SETUP="true"
      ;;
    --dry-run)
      DRY_RUN="true"
      ;;
    --stop-on-complete)
      STOP_ON_COMPLETE="true"
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown argument: $1"
      usage
      exit 1
      ;;
  esac
  shift
done

cd "$PROJECT_ROOT"

if [ "$RUN_SETUP" = "true" ]; then
  run_cmd "bash scripts/setup/setup-dev.sh"
fi

if [ "$SCENARIO" = "maritime" ]; then
  run_cmd "bash scripts/demo/run-maritime-demo.sh"
else
  run_cmd "bash scripts/demo/run-multi-domain-demo.sh"
fi

if [ "$STOP_ON_COMPLETE" = "true" ]; then
  run_cmd "bash scripts/demo/stop-demo.sh"
fi

echo "Demo run completed for scenario: $SCENARIO"
echo "Tip: keep services running to showcase dashboards; stop with: bash scripts/demo/stop-demo.sh"
