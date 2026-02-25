#!/usr/bin/env bash
# CLASSIFICATION: UNCLASSIFIED
# RTSA Integration Test Runner
# Runs integration tests against the live local dev stack.
# Requires: docker compose -f deploy/docker-compose.yml up -d

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
SERVICES_DIR="${REPO_ROOT}/services"

GREEN='\033[0;32m'
RED='\033[0;31m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${CYAN}[ℹ]${NC} $*"; }
log_pass() { echo -e "${GREEN}[✓]${NC} $*"; }
log_fail() { echo -e "${RED}[✗]${NC} $*"; }
log_warn() { echo -e "${YELLOW}[!]${NC} $*"; }

ERRORS=0

# ─────────────────────────────────────────────────────────────
# Check the dev stack is running
# ─────────────────────────────────────────────────────────────
check_stack() {
  log_info "Verifying dev stack is running ..."

  local redpanda_ok=false
  local clickhouse_ok=false

  if curl -sf "http://localhost:19644/v1/cluster/health_overview" &>/dev/null; then
    redpanda_ok=true
  fi

  if curl -sf "http://localhost:8123/ping" &>/dev/null; then
    clickhouse_ok=true
  fi

  if $redpanda_ok && $clickhouse_ok; then
    log_pass "Dev stack is running (Redpanda + ClickHouse)"
  else
    echo -e "${RED}ERROR: Dev stack is not fully running.${NC}"
    echo "  Start it with: docker compose -f deploy/docker-compose.yml up -d"
    $redpanda_ok || echo -e "  ${RED}  Redpanda: NOT reachable${NC}"
    $clickhouse_ok || echo -e "  ${RED}  ClickHouse: NOT reachable${NC}"
    exit 1
  fi
}

# ─────────────────────────────────────────────────────────────
# Run integration tests for all services
# ─────────────────────────────────────────────────────────────
run_integration_tests() {
  if [ ! -d "$SERVICES_DIR" ]; then
    log_warn "No services/ directory — nothing to test"
    return
  fi

  log_info "Running integration tests (build tag: integration) ..."
  echo ""

  while IFS= read -r -d '' mod_file; do
    local service_dir
    service_dir="$(dirname "$mod_file")"
    local service_name
    service_name="$(basename "$service_dir")"

    # Skip if no integration test files exist
    local int_test_count
    int_test_count="$(find "$service_dir" -name "*_integration_test.go" 2>/dev/null | wc -l | tr -d ' ')"

    if [ "$int_test_count" -eq 0 ]; then
      log_info "  ${service_name} — no integration tests (skipping)"
      continue
    fi

    log_info "  Running integration tests for ${service_name} (${int_test_count} test files) ..."

    if (cd "$service_dir" && \
      go test \
        -tags=integration \
        -timeout 300s \
        -count=1 \
        ./... 2>&1 | sed "s/^/    /"); then
      log_pass "  ${service_name} — integration tests PASSED"
    else
      log_fail "  ${service_name} — integration tests FAILED"
      ERRORS=$(( ERRORS + 1 ))
    fi

  done < <(find "$SERVICES_DIR" -name "go.mod" -print0 2>/dev/null)
}

# ─────────────────────────────────────────────────────────────
# Main
# ─────────────────────────────────────────────────────────────
main() {
  echo ""
  echo -e "${CYAN}══ RTSA Integration Tests (CLASSIFICATION: UNCLASSIFIED) ══${NC}"

  check_stack
  run_integration_tests

  echo ""
  if [ "$ERRORS" -eq 0 ]; then
    echo -e "${GREEN}PASSED — All integration tests passed.${NC}"
  else
    echo -e "${RED}FAILED — ${ERRORS} integration test suite(s) failed.${NC}"
    exit 1
  fi
}

main "$@"
