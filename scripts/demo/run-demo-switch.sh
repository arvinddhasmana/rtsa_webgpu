#!/usr/bin/env bash
# CLASSIFICATION: UNCLASSIFIED
# scripts/demo/run-demo-switch.sh
# Convenience wrapper to switch between maritime and multi-domain demos.
# Usage: bash scripts/demo/run-demo-switch.sh [maritime|multi-domain] [options]
# This script simply forwards all arguments to run-demo.sh.

set -euo pipefail

# Resolve project root (same as other scripts)
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
cd "$PROJECT_ROOT"

# Forward arguments to run-demo.sh
bash scripts/demo/run-demo.sh "$@"
