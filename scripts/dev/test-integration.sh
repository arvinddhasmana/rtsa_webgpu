#!/usr/bin/env bash
# CLASSIFICATION: UNCLASSIFIED
# RTSA Integration Test Runner
# Runs integration tests from the centralized tests/ directory and
# per-service integration tests found in svc-*/internal/integration/.
# Requires: Docker (for testcontainers-go)
#
# Test results written to:  ${RTSA_TEST_RESULTS_DIR}/integration/   (set by test-all.sh)
#                     or:  .test-results/<timestamp>/integration/  (standalone run)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# ── Results directory ──────────────────────────────────────────────────────────
_TS="$(date +%Y-%m-%dT%H-%M-%S)"
RESULTS_DIR="${RTSA_TEST_RESULTS_DIR:-.test-results/${_TS}}"
INTEGRATION_DIR="${RESULTS_DIR}/integration"
mkdir -p "${INTEGRATION_DIR}"
INT_LOG="${INTEGRATION_DIR}/run.log"
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

ERRORS=0
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# ─────────────────────────────────────────────────────────────
# Helpers
# ─────────────────────────────────────────────────────────────

# Append failures from a log segment into the shared failures.txt
# Usage: extract_failures <log_file> <label_prefix>
extract_failures() {
  local log="$1" prefix="$2"
  grep -E '^[[:space:]]*--- FAIL:' "${log}" 2>/dev/null \
    | sed "s|.*--- FAIL: \([^ ]*\).*|${prefix}/\1|" \
    >> "${INTEGRATION_DIR}/failures.txt" || true
}

# Tally PASS/FAIL lines from a log segment
tally_counts() {
  local log="$1"
  local p f
  p="$(grep -cE '^[[:space:]]*--- PASS:' "${log}" 2>/dev/null; true)"
  f="$(grep -cE '^[[:space:]]*--- FAIL:' "${log}" 2>/dev/null; true)"
  # Strip any trailing whitespace/newlines; default to 0 if empty (file missing).
  p="${p//[[:space:]]/}"; p="${p:-0}"
  f="${f//[[:space:]]/}"; f="${f:-0}"
  TOTAL_TESTS=$(( TOTAL_TESTS + p + f ))
  PASSED_TESTS=$(( PASSED_TESTS + p ))
  FAILED_TESTS=$(( FAILED_TESTS + f ))
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

# ─────────────────────────────────────────────────────────────
# Check Docker is available (needed for testcontainers-go)
# ─────────────────────────────────────────────────────────────
check_docker() {
  log_info "Verifying Docker is available ..."

  if docker info &>/dev/null; then
    log_pass "Docker is running"
  else
    echo -e "${RED}ERROR: Docker is not running or not available.${NC}"
    echo "  testcontainers-go requires Docker to spin up test containers."
    exit 1
  fi
}

# ─────────────────────────────────────────────────────────────
# Run centralized integration tests (tests/integration/)
# ─────────────────────────────────────────────────────────────
run_centralized_integration() {
  local tests_dir="${REPO_ROOT}/tests"

  if [ ! -d "${tests_dir}/integration" ]; then
    log_warn "No tests/integration/ directory found — skipping centralized tests"
    return
  fi

  log_info "Running centralized integration tests (tests/integration/) ..."
  echo ""

  local seg_log="${INTEGRATION_DIR}/centralized.log"
  set +e
  (cd "$tests_dir" && \
    RTSA_INTEGRATION_TESTS=true go test \
      -v \
      -tags=integration \
      -timeout 10m \
      -count=1 \
      ./integration/... 2>&1) | tee "${seg_log}" | sed "s/^/    /"
  local _pe=("${PIPESTATUS[@]}")
  set -e
  local go_exit=${_pe[0]}

  # Append to combined run.log and tally
  cat "${seg_log}" >> "${INT_LOG}" 2>/dev/null || true
  tally_counts "${seg_log}"
  extract_failures "${seg_log}" "integration/centralized"

  if [ "${go_exit}" -eq 0 ] && ! grep -qE '^(FAIL|--- FAIL)' "${seg_log}" 2>/dev/null; then
    log_pass "Centralized integration tests — PASSED"
  else
    log_fail "Centralized integration tests — FAILED"
    ERRORS=$(( ERRORS + 1 ))
  fi
}

# ─────────────────────────────────────────────────────────────
# Run per-service integration tests (svc-*/internal/integration/)
# ─────────────────────────────────────────────────────────────
run_service_integration_tests() {
  log_info "Scanning for per-service integration tests ..."
  echo ""

  for dir in "${REPO_ROOT}"/svc-*/; do
    local service_name
    service_name="$(basename "$dir")"

    local int_test_count
    int_test_count="$(find "$dir" -name "*_integration_test.go" -o -name "*_integration_*.go" -o -path "*/integration/*_test.go" 2>/dev/null | wc -l | tr -d ' ')"

    if [ "$int_test_count" -eq 0 ]; then
      log_info "  ${service_name} — no integration tests (skipping)"
      continue
    fi

    log_info "  Running integration tests for ${service_name} (${int_test_count} test files) ..."

    local seg_log="${INTEGRATION_DIR}/${service_name}.log"
    set +e
    (cd "$dir" && \
      RTSA_INTEGRATION_TESTS=true go test \
        -tags=integration \
        -timeout 300s \
        -count=1 \
        -v \
        ./... 2>&1) | tee "${seg_log}" | sed "s/^/    /"
    local _pe=("${PIPESTATUS[@]}")
    set -e
    local go_exit=${_pe[0]}

    cat "${seg_log}" >> "${INT_LOG}" 2>/dev/null || true
    tally_counts "${seg_log}"
    extract_failures "${seg_log}" "integration/${service_name}"

    if [ "${go_exit}" -eq 0 ] && ! grep -qE '^(FAIL|--- FAIL)' "${seg_log}" 2>/dev/null; then
      log_pass "  ${service_name} — integration tests PASSED"
    else
      log_fail "  ${service_name} — integration tests FAILED"
      ERRORS=$(( ERRORS + 1 ))
    fi
  done
}

# ─────────────────────────────────────────────────────────────
# Main
# ─────────────────────────────────────────────────────────────
main() {
  echo ""
  echo -e "${CYAN}══ RTSA Integration Tests (CLASSIFICATION: UNCLASSIFIED) ══${NC}"
  echo "  Logs: ${INTEGRATION_DIR}/"

  check_docker
  run_centralized_integration
  run_service_integration_tests

  # ── Write counts for test-all.sh ──
  printf 'TOTAL=%d PASSED=%d FAILED=%d SUITES_FAILED=%d\n' \
    "${TOTAL_TESTS}" "${PASSED_TESTS}" "${FAILED_TESTS}" "${ERRORS}" \
    > "${INTEGRATION_DIR}/counts.txt"

  # ── Failure detail ──
  print_failures "${INTEGRATION_DIR}/failures.txt" 10 "Integration"

  echo ""
  echo "─────────────────────────────────────────────────────────"
  echo "  Integration test summary:"
  echo "  Tests : ${TOTAL_TESTS} run | ${PASSED_TESTS} passed | ${FAILED_TESTS} failed"
  echo "  Suites: ${ERRORS} failed"
  echo "  Full log    : ${INT_LOG}"
  echo "  Failed list : ${INTEGRATION_DIR}/failures.txt"
  echo "─────────────────────────────────────────────────────────"

  if [ "$ERRORS" -eq 0 ]; then
    echo -e "${GREEN}PASSED — All integration tests passed.${NC}"
  else
    echo -e "${RED}FAILED — ${ERRORS} integration test suite(s) failed.${NC}"
    exit 1
  fi
}

main "$@"
