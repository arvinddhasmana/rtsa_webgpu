#!/usr/bin/env bash
# CLASSIFICATION: UNCLASSIFIED
# RTSA Go Unit Test Runner
# Runs all Go unit tests with race detection and coverage reporting.
# Usage: ./scripts/dev/test-go.sh [module-name]
#   module-name: optional filter, e.g. "svc-radar-ingestion" or "pkg"

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
COVERAGE_DIR="${REPO_ROOT}/.coverage"
MIN_COVERAGE=80

# Optional: filter to a single module
TARGET_MODULE="${1:-}"

GREEN='\033[0;32m'
RED='\033[0;31m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${CYAN}[ℹ]${NC} $*"; }
log_pass() { echo -e "${GREEN}[✓]${NC} $*"; }
log_fail() { echo -e "${RED}[✗]${NC} $*"; }
log_warn() { echo -e "${YELLOW}[!]${NC} $*"; }

TOTAL_MODULES=0
PASSED_MODULES=0
FAILED_MODULES=0

run_tests_for_module() {
  local module_dir="$1"
  local module_name
  module_name="$(basename "$module_dir")"
  local coverage_file="${COVERAGE_DIR}/${module_name}.out"
  local coverage_html="${COVERAGE_DIR}/${module_name}.html"

  TOTAL_MODULES=$(( TOTAL_MODULES + 1 ))
  echo ""
  log_info "Testing ${module_name} ..."

  mkdir -p "$COVERAGE_DIR"

  # Run tests with race detector and coverage
  local exit_code=0
  (cd "$module_dir" && go test \
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
    done) || exit_code=$?

  if [ "${exit_code:-0}" -ne 0 ]; then
    log_fail "${module_name} — TESTS FAILED"
    FAILED_MODULES=$(( FAILED_MODULES + 1 ))
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
      log_pass "${module_name} — ${coverage_pct}% coverage (>= ${MIN_COVERAGE}% required)"
      PASSED_MODULES=$(( PASSED_MODULES + 1 ))
    else
      log_fail "${module_name} — ${coverage_pct}% coverage (< ${MIN_COVERAGE}% required)"
      log_info "  Coverage report: ${coverage_html}"
      FAILED_MODULES=$(( FAILED_MODULES + 1 ))
    fi
  else
    log_warn "${module_name} — no coverage data (no testable packages?)"
    PASSED_MODULES=$(( PASSED_MODULES + 1 ))
  fi
}

# Discover all Go modules in the repository (svc-*, pkg, wasm-transforms, tools/*)
discover_modules() {
  local modules=()

  # pkg/ shared library
  if [ -f "${REPO_ROOT}/pkg/go.mod" ]; then
    modules+=("${REPO_ROOT}/pkg")
  fi

  # svc-* service modules
  for dir in "${REPO_ROOT}"/svc-*/; do
    if [ -f "${dir}/go.mod" ]; then
      modules+=("$dir")
    fi
  done

  # wasm-transforms
  if [ -f "${REPO_ROOT}/wasm-transforms/go.mod" ]; then
    modules+=("${REPO_ROOT}/wasm-transforms")
  fi

  printf '%s\n' "${modules[@]}"
}

main() {
  echo ""
  echo -e "${CYAN}══ RTSA Go Unit Tests (CLASSIFICATION: UNCLASSIFIED) ══${NC}"
  echo "  Min coverage: ${MIN_COVERAGE}%  |  Race detector: ON  |  Timeout: 120s"

  if ! command -v go &>/dev/null; then
    echo -e "${RED}ERROR: go not found${NC}"
    exit 1
  fi

  if [ -n "$TARGET_MODULE" ]; then
    # Try to find the module by name
    local module_path=""
    for candidate in "${REPO_ROOT}/${TARGET_MODULE}" "${REPO_ROOT}/svc-${TARGET_MODULE}"; do
      if [ -f "${candidate}/go.mod" ]; then
        module_path="$candidate"
        break
      fi
    done

    if [ -z "$module_path" ]; then
      echo -e "${RED}ERROR: Module '${TARGET_MODULE}' not found (tried ${TARGET_MODULE}/ and svc-${TARGET_MODULE}/)${NC}"
      exit 1
    fi
    run_tests_for_module "$module_path"
  else
    while IFS= read -r module_dir; do
      [ -z "$module_dir" ] && continue
      run_tests_for_module "$module_dir"
    done < <(discover_modules)
  fi

  echo ""
  echo "─────────────────────────────────────────────────────"
  echo "  Test suite summary:"
  echo "  Modules: ${TOTAL_MODULES} total | ${PASSED_MODULES} passed | ${FAILED_MODULES} failed"
  echo "  Coverage reports: ${COVERAGE_DIR}/"
  echo "─────────────────────────────────────────────────────"

  if [ "$FAILED_MODULES" -gt 0 ]; then
    echo -e "${RED}FAILED — ${FAILED_MODULES} module(s) did not meet requirements${NC}"
    exit 1
  fi

  if [ "$TOTAL_MODULES" -eq 0 ]; then
    log_warn "No Go modules found to test"
    exit 0
  fi

  echo -e "${GREEN}PASSED — All ${PASSED_MODULES} module(s) meet coverage requirements${NC}"
}

main "$@"
