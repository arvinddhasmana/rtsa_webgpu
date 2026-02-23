#!/usr/bin/env bash
# CLASSIFICATION: UNCLASSIFIED
# RTSA Go Unit Test Runner
# Runs all Go unit tests with race detection and coverage reporting.
# Usage: ./scripts/dev/test-go.sh [service-name]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
SERVICES_DIR="${REPO_ROOT}/services"
COVERAGE_DIR="${REPO_ROOT}/.coverage"
MIN_COVERAGE=80

# Optional: filter to a single service
TARGET_SERVICE="${1:-}"

GREEN='\033[0;32m'
RED='\033[0;31m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${CYAN}[ℹ]${NC} $*"; }
log_pass() { echo -e "${GREEN}[✓]${NC} $*"; }
log_fail() { echo -e "${RED}[✗]${NC} $*"; }
log_warn() { echo -e "${YELLOW}[!]${NC} $*"; }

TOTAL_SERVICES=0
PASSED_SERVICES=0
FAILED_SERVICES=0

run_tests_for_service() {
  local service_dir="$1"
  local service_name
  service_name="$(basename "$service_dir")"
  local coverage_file="${COVERAGE_DIR}/${service_name}.out"
  local coverage_html="${COVERAGE_DIR}/${service_name}.html"

  TOTAL_SERVICES=$(( TOTAL_SERVICES + 1 ))
  echo ""
  log_info "Testing ${service_name} ..."

  mkdir -p "$COVERAGE_DIR"

  # Run tests with race detector and coverage
  local exit_code=0
  go test \
    -race \
    -timeout 120s \
    -coverprofile="${coverage_file}" \
    -covermode=atomic \
    -v \
    ./... \
    2>&1 | while IFS= read -r line; do
      if echo "$line" | grep -q "^--- PASS"; then
        echo -e "  ${GREEN}${line}${NC}"
      elif echo "$line" | grep -q "^--- FAIL"; then
        echo -e "  ${RED}${line}${NC}"
      else
        echo "  ${line}"
      fi
    done || exit_code=$?

  if [ "${exit_code:-0}" -ne 0 ]; then
    log_fail "${service_name} — TESTS FAILED"
    FAILED_SERVICES=$(( FAILED_SERVICES + 1 ))
    return
  fi

  # Parse coverage percentage
  if [ -f "$coverage_file" ]; then
    local coverage_pct
    coverage_pct="$(go tool cover -func="${coverage_file}" | tail -1 | awk '{print $3}' | sed 's/%//')"
    coverage_pct="${coverage_pct:-0}"

    # Generate HTML coverage report
    go tool cover -html="${coverage_file}" -o "${coverage_html}" 2>/dev/null || true

    local coverage_int
    coverage_int="$(printf "%.0f" "$coverage_pct" 2>/dev/null || echo "0")"

    if [ "${coverage_int}" -ge "${MIN_COVERAGE}" ]; then
      log_pass "${service_name} — ${coverage_pct}% coverage (>= ${MIN_COVERAGE}% required)"
      PASSED_SERVICES=$(( PASSED_SERVICES + 1 ))
    else
      log_fail "${service_name} — ${coverage_pct}% coverage (< ${MIN_COVERAGE}% required)"
      log_info "  Coverage report: ${coverage_html}"
      FAILED_SERVICES=$(( FAILED_SERVICES + 1 ))
    fi
  else
    log_warn "${service_name} — no coverage data (no testable packages?)"
    PASSED_SERVICES=$(( PASSED_SERVICES + 1 ))
  fi
}

main() {
  echo ""
  echo -e "${CYAN}══ RTSA Go Unit Tests (CLASSIFICATION: UNCLASSIFIED) ══${NC}"
  echo "  Min coverage: ${MIN_COVERAGE}%  |  Race detector: ON  |  Timeout: 120s"

  if ! command -v go &>/dev/null; then
    echo -e "${RED}ERROR: go not found${NC}"
    exit 1
  fi

  if [ ! -d "$SERVICES_DIR" ]; then
    log_warn "No services/ directory — nothing to test"
    exit 0
  fi

  if [ -n "$TARGET_SERVICE" ]; then
    local service_path="${SERVICES_DIR}/${TARGET_SERVICE}"
    if [ ! -d "$service_path" ]; then
      echo -e "${RED}ERROR: Service '${TARGET_SERVICE}' not found in ${SERVICES_DIR}/${NC}"
      exit 1
    fi
    (cd "$service_path" && run_tests_for_service "$service_path")
  else
    while IFS= read -r -d '' mod_file; do
      local service_dir
      service_dir="$(dirname "$mod_file")"
      (cd "$service_dir" && run_tests_for_service "$service_dir")
    done < <(find "$SERVICES_DIR" -name "go.mod" -print0 2>/dev/null)
  fi

  echo ""
  echo "─────────────────────────────────────────────────────"
  echo "  Test suite summary:"
  echo "  Services: ${TOTAL_SERVICES} total | ${PASSED_SERVICES} passed | ${FAILED_SERVICES} failed"
  echo "  Coverage reports: ${COVERAGE_DIR}/"
  echo "─────────────────────────────────────────────────────"

  if [ "$FAILED_SERVICES" -gt 0 ]; then
    echo -e "${RED}FAILED — ${FAILED_SERVICES} service(s) did not meet requirements${NC}"
    exit 1
  fi

  if [ "$TOTAL_SERVICES" -eq 0 ]; then
    log_warn "No Go services found to test"
    exit 0
  fi

  echo -e "${GREEN}PASSED — All ${PASSED_SERVICES} service(s) meet coverage requirements${NC}"
}

main "$@"
