#!/usr/bin/env bash
# CLASSIFICATION: UNCLASSIFIED
# RTSA End-to-End Test Runner
# Manages the full Docker Compose lifecycle: build → start → init → test → teardown.
# Usage: ./scripts/dev/test-e2e.sh [--keep]
#   --keep: don't tear down the stack after tests (useful for debugging)
#
# Test results written to:  ${RTSA_TEST_RESULTS_DIR}/e2e/   (set by test-all.sh)
#                     or:  .test-results/<timestamp>/e2e/  (standalone run)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

COMPOSE_FILE="${REPO_ROOT}/deploy/docker-compose.yml"
COMPOSE_SERVICES="${REPO_ROOT}/deploy/docker-compose.services.yml"
TESTS_DIR="${REPO_ROOT}/tests"

KEEP_STACK=false
[ "${1:-}" = "--keep" ] && KEEP_STACK=true

# ── Results directory ──────────────────────────────────────────────────────────
_TS="$(date +%Y-%m-%dT%H-%M-%S)"
RESULTS_DIR="${RTSA_TEST_RESULTS_DIR:-.test-results/${_TS}}"
E2E_DIR="${RESULTS_DIR}/e2e"
mkdir -p "${E2E_DIR}"
E2E_LOG="${E2E_DIR}/run.log"
# ──────────────────────────────────────────────────────────────────────────────

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

print_failures() {
  local file="$1" max_shown="${2:-10}" label="$3"
  if [ ! -f "$file" ] || [ ! -s "$file" ]; then return; fi
  local total_fail
  total_fail="$(wc -l < "$file" | tr -d ' ')"
  local shown=$(( total_fail < max_shown ? total_fail : max_shown ))
  echo ""
  echo -e "${RED}${BOLD}── FAILED TESTS: ${label} [${total_fail} failure(s), showing first ${shown}] ──────${NC}"
  local i=1
  while IFS= read -r line && [ "$i" -le "$max_shown" ]; do
    printf "  %2d. %s\n" "$i" "$line"
    i=$(( i + 1 ))
  done < "$file"
  if [ "$total_fail" -gt "$max_shown" ]; then
    echo "  ... and $(( total_fail - max_shown )) more — see: ${file}"
  fi
}

main() {
  echo ""
  echo -e "${BOLD}${CYAN}══ RTSA End-to-End Tests (CLASSIFICATION: UNCLASSIFIED) ══${NC}"
  echo "  Logs: ${E2E_DIR}/"

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

  # 4. Run E2E tests — tee raw -v output to log file
  log_info "Running E2E tests ..."
  set +e
  (cd "$TESTS_DIR" && \
    RTSA_INTEGRATION_TESTS=true go test -v -tags e2e -timeout 15m ./e2e/... 2>&1) \
    | tee "${E2E_LOG}"
  local _pe=("${PIPESTATUS[@]}")
  set -e
  local test_exit=${_pe[0]}

  # 5. Teardown
  cleanup

  # 6. Parse failures from log
  local e2e_passed e2e_failed
  e2e_passed="$(grep -cE '^[[:space:]]*--- PASS:' "${E2E_LOG}" 2>/dev/null || echo 0)"
  e2e_failed="$(grep -cE '^[[:space:]]*--- FAIL:' "${E2E_LOG}" 2>/dev/null || echo 0)"
  local e2e_total=$(( e2e_passed + e2e_failed ))

  grep -E '^[[:space:]]*--- FAIL:' "${E2E_LOG}" 2>/dev/null \
    | sed 's|.*--- FAIL: \([^ ]*\).*|e2e/\1|' \
    >> "${E2E_DIR}/failures.txt" || true

  # Write counts for test-all.sh
  printf 'TOTAL=%d PASSED=%d FAILED=%d\n' \
    "${e2e_total}" "${e2e_passed}" "${e2e_failed}" \
    > "${E2E_DIR}/counts.txt"

  # 7. Failure detail
  print_failures "${E2E_DIR}/failures.txt" 10 "E2E"

  # 8. Report
  local elapsed=$(( SECONDS - start_time ))
  echo ""
  echo "─────────────────────────────────────────────────────"
  echo "  E2E test summary:"
  echo "  Tests : ${e2e_total} run | ${e2e_passed} passed | ${e2e_failed} failed"
  echo "  Full log    : ${E2E_LOG}"
  echo "  Failed list : ${E2E_DIR}/failures.txt"
  echo "─────────────────────────────────────────────────────"

  if [ "$test_exit" -eq 0 ] && [ "${e2e_failed}" -eq 0 ]; then
    echo -e "${GREEN}${BOLD}PASSED${NC} — E2E tests completed successfully (${elapsed}s)"
  else
    echo -e "${RED}${BOLD}FAILED${NC} — E2E tests failed (${elapsed}s)"
    exit 1
  fi
}

main "$@"
