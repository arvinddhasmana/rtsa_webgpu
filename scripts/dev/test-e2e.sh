#!/usr/bin/env bash
# CLASSIFICATION: UNCLASSIFIED
# RTSA End-to-End Test Runner
# Manages the full Docker Compose lifecycle: build → start → init → test → teardown.
# Usage: ./scripts/dev/test-e2e.sh [--keep]
#   --keep: don't tear down the stack after tests (useful for debugging)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

COMPOSE_FILE="${REPO_ROOT}/deploy/docker-compose.yml"
COMPOSE_SERVICES="${REPO_ROOT}/deploy/docker-compose.services.yml"
TESTS_DIR="${REPO_ROOT}/tests"

KEEP_STACK=false
[ "${1:-}" = "--keep" ] && KEEP_STACK=true

GREEN='\033[0;32m'
RED='\033[0;31m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
BOLD='\033[1m'
NC='\033[0m'

log_info() { echo -e "${CYAN}[ℹ]${NC} $*"; }
log_pass() { echo -e "${GREEN}[✓]${NC} $*"; }
log_fail() { echo -e "${RED}[✗]${NC} $*"; }
log_warn() { echo -e "${YELLOW}[!]${NC} $*"; }

cleanup() {
  if [ "$KEEP_STACK" = true ]; then
    log_warn "Stack left running (--keep flag). Tear down manually with:"
    echo "  docker compose -f ${COMPOSE_FILE} -f ${COMPOSE_SERVICES} down -v"
    return
  fi
  log_info "Tearing down Docker Compose stack ..."
  docker compose -f "$COMPOSE_FILE" -f "$COMPOSE_SERVICES" down -v 2>/dev/null || true
}

main() {
  echo ""
  echo -e "${BOLD}${CYAN}══ RTSA End-to-End Tests (CLASSIFICATION: UNCLASSIFIED) ══${NC}"

  local start_time=$SECONDS

  # 1. Build and start
  log_info "Building and starting Docker Compose stack ..."
  docker compose -f "$COMPOSE_FILE" -f "$COMPOSE_SERVICES" up -d --build --wait
  log_pass "All containers are healthy"

  # 2. Initialize topics
  log_info "Initializing Redpanda topics ..."
  "${SCRIPT_DIR}/init-topics.sh"
  log_pass "Topics initialized"

  # 3. Initialize ClickHouse
  log_info "Initializing ClickHouse schema ..."
  "${SCRIPT_DIR}/init-clickhouse.sh"
  log_pass "ClickHouse schema initialized"

  # 4. Run E2E tests
  log_info "Running E2E tests ..."
  local test_exit=0
  (cd "$TESTS_DIR" && \
    RTSA_INTEGRATION_TESTS=true go test -v -tags e2e -timeout 15m ./e2e/...) || test_exit=$?

  # 5. Teardown
  cleanup

  # 6. Report
  local elapsed=$(( SECONDS - start_time ))
  echo ""
  echo "─────────────────────────────────────────────────────"

  if [ "$test_exit" -eq 0 ]; then
    echo -e "${GREEN}${BOLD}PASSED${NC} — E2E tests completed successfully (${elapsed}s)"
  else
    echo -e "${RED}${BOLD}FAILED${NC} — E2E tests failed (${elapsed}s)"
    exit 1
  fi
}

main "$@"
