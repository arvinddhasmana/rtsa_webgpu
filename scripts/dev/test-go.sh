#!/usr/bin/env bash
# CLASSIFICATION: UNCLASSIFIED
# RTSA Go Unit Test Runner
# Runs all Go unit tests with race detection and coverage reporting.
# Usage: ./scripts/dev/test-go.sh [module-name]
#   module-name: optional filter, e.g. "svc-radar-ingestion" or "pkg"
#
# Test results written to:  ${RTSA_TEST_RESULTS_DIR}/unit/   (set by test-all.sh)
#                     or:  .test-results/<timestamp>/unit/  (standalone run)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
COVERAGE_DIR="${REPO_ROOT}/.coverage"
MIN_COVERAGE=80

# Optional: filter to a single module
TARGET_MODULE="${1:-}"

# ── Results directory ────────────────────────────────────────────────────────
# Honour parent (test-all.sh) or create a standalone timestamped directory.
_TS="$(date +%Y-%m-%dT%H-%M-%S)"
RESULTS_DIR="${RTSA_TEST_RESULTS_DIR:-.test-results/${_TS}}"
UNIT_DIR="${RESULTS_DIR}/unit"
mkdir -p "${UNIT_DIR}" "${COVERAGE_DIR}"
# ─────────────────────────────────────────────────────────────────────────────

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

TOTAL_MODULES=0
PASSED_MODULES=0
FAILED_MODULES=0

# Aggregate individual test counts (tallied from logs after each module)
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

run_tests_for_module() {
  local module_dir="$1"
  local module_name
  module_name="$(basename "$module_dir")"
  local coverage_file="${COVERAGE_DIR}/${module_name}.out"
  local coverage_html="${COVERAGE_DIR}/${module_name}.html"
  local log_file="${UNIT_DIR}/${module_name}.log"

  TOTAL_MODULES=$(( TOTAL_MODULES + 1 ))
  echo ""
  log_info "Testing ${module_name} ..."

  mkdir -p "$COVERAGE_DIR"

  # Run tests with race detector and coverage.
  # Tee raw -v output to log file; pipe through colorizer for console display.
  # Use set +e / PIPESTATUS so we capture go test's exit code reliably.
  set +e
  (cd "$module_dir" && go test \
    -race \
    -timeout 120s \
    -coverprofile="${coverage_file}" \
    -covermode=atomic \
    -v \
    ./... 2>&1) | tee "${log_file}" | \
    while IFS= read -r line; do
      if [[ "$line" =~ ^[[:space:]]*---\ FAIL: ]]; then
        echo -e "  ${RED}${line}${NC}"
      elif [[ "$line" =~ ^[[:space:]]*---\ PASS: ]]; then
        echo -e "  ${GREEN}${line}${NC}"
      else
        echo "  ${line}"
      fi
    done
  local _pipe_exit=("${PIPESTATUS[@]}")
  set -e
  local go_exit=${_pipe_exit[0]}

  # ── Count individual tests from log ──────────────────────────────────────
  local mod_passed mod_failed
  mod_passed="$(grep -cE '^[[:space:]]*--- PASS:' "${log_file}" 2>/dev/null || true)"
  [ -z "$mod_passed" ] && mod_passed=0
  mod_failed="$(grep -cE '^[[:space:]]*--- FAIL:' "${log_file}" 2>/dev/null || true)"
  [ -z "$mod_failed" ] && mod_failed=0
  TOTAL_TESTS=$(( TOTAL_TESTS + mod_passed + mod_failed ))
  PASSED_TESTS=$(( PASSED_TESTS + mod_passed ))
  FAILED_TESTS=$(( FAILED_TESTS + mod_failed ))

  # ── Append failures to shared failures.txt ────────────────────────────────
  if [ "${mod_failed}" -gt 0 ]; then
    grep -E '^[[:space:]]*--- FAIL:' "${log_file}" \
      | sed "s|.*--- FAIL: \([^ ]*\).*|${module_name}/\1|" \
      >> "${UNIT_DIR}/failures.txt"
  fi

  if [ "${go_exit}" -ne 0 ] && [ "${mod_failed}" -eq 0 ]; then
    # Build error or crash — log the whole outcome as a failure entry
    echo "${module_name}/[BUILD_OR_RUNTIME_ERROR]" >> "${UNIT_DIR}/failures.txt"
    mod_failed=$(( mod_failed + 1 ))
    FAILED_TESTS=$(( FAILED_TESTS + 1 ))
  fi

  # ── Module-level outcome ──────────────────────────────────────────────────
  if [ "${go_exit}" -ne 0 ] || [ "${mod_failed}" -gt 0 ]; then
    log_fail "${module_name} — TESTS FAILED  (${mod_failed} test(s) failed)"
    FAILED_MODULES=$(( FAILED_MODULES + 1 ))
    return
  fi

  # ── Coverage check ────────────────────────────────────────────────────────
  if [ -f "$coverage_file" ]; then
    # Exclude infrastructure wrappers, repositories, and main entry points from coverage
    sed -i -e '/\/testutil/d' \
           -e '/\/redpanda\/consumer\.go/d' \
           -e '/\/redpanda\/producer\.go/d' \
           -e '/clickhouse\.go/d' \
           -e '/\/repository\//d' \
           -e '/\/mock\//d' \
           -e '/main\.go/d' "${coverage_file}"

    local coverage_pct
    coverage_pct="$(go tool cover -func="${coverage_file}" | tail -1 | awk '{print $3}' | sed 's/%//')"
    coverage_pct="${coverage_pct:-0}"

    go tool cover -html="${coverage_file}" -o "${coverage_html}" 2>/dev/null || true

    local coverage_int
    coverage_int="$(printf "%.0f" "$coverage_pct" 2>/dev/null || echo "0")"

    if [ "${coverage_int}" -ge "${MIN_COVERAGE}" ]; then
      log_pass "${module_name} — ${coverage_pct}% coverage (>= ${MIN_COVERAGE}% required)"
      PASSED_MODULES=$(( PASSED_MODULES + 1 ))
    else
      log_fail "${module_name} — ${coverage_pct}% coverage (< ${MIN_COVERAGE}% required)"
      log_info "  Coverage report: ${coverage_html}"
      echo "${module_name}/[COVERAGE_${coverage_pct}%_BELOW_${MIN_COVERAGE}%]" >> "${UNIT_DIR}/failures.txt"
      FAILED_TESTS=$(( FAILED_TESTS + 1 ))
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

  if [ -f "${REPO_ROOT}/pkg/go.mod" ]; then
    modules+=("${REPO_ROOT}/pkg")
  fi

  for dir in "${REPO_ROOT}"/svc-*/; do
    if [ -f "${dir}/go.mod" ]; then
      modules+=("$dir")
    fi
  done

  if [ -f "${REPO_ROOT}/wasm-transforms/go.mod" ]; then
    modules+=("${REPO_ROOT}/wasm-transforms")
  fi

  printf '%s\n' "${modules[@]}"
}

# Print the first N failures from failures.txt with index numbers
print_failures() {
  local file="$1"
  local max_shown="${2:-10}"
  local label="$3"
  if [ ! -f "$file" ] || [ ! -s "$file" ]; then
    return
  fi
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
  echo -e "${CYAN}══ RTSA Go Unit Tests (CLASSIFICATION: UNCLASSIFIED) ══${NC}"
  echo "  Min coverage: ${MIN_COVERAGE}%  |  Race detector: ON  |  Timeout: 120s"
  echo "  Logs: ${UNIT_DIR}/"

  if ! command -v go &>/dev/null; then
    echo -e "${RED}ERROR: go not found${NC}"
    exit 1
  fi

  if [ -n "$TARGET_MODULE" ]; then
    local module_path=""
    for candidate in "${REPO_ROOT}/${TARGET_MODULE}" "${REPO_ROOT}/svc-${TARGET_MODULE}"; do
      if [ -f "${candidate}/go.mod" ]; then
        module_path="$candidate"
        break
      fi
    done

    if [ -z "$module_path" ]; then
      echo -e "${RED}ERROR: Module '${TARGET_MODULE}' not found${NC}"
      exit 1
    fi
    run_tests_for_module "$module_path"
  else
    while IFS= read -r module_dir; do
      [ -z "$module_dir" ] && continue
      run_tests_for_module "$module_dir"
    done < <(discover_modules)
  fi

  # ── Write counts for test-all.sh to read ─────────────────────────────────
  printf 'TOTAL=%d PASSED=%d FAILED=%d MODULES_TOTAL=%d MODULES_PASSED=%d MODULES_FAILED=%d\n' \
    "${TOTAL_TESTS}" "${PASSED_TESTS}" "${FAILED_TESTS}" \
    "${TOTAL_MODULES}" "${PASSED_MODULES}" "${FAILED_MODULES}" \
    > "${UNIT_DIR}/counts.txt"

  # ── Failure detail (shown when running standalone or within test-all.sh) ──
  print_failures "${UNIT_DIR}/failures.txt" 10 "Unit — Go"

  echo ""
  echo "─────────────────────────────────────────────────────"
  echo "  Go unit test summary:"
  echo "  Tests  : ${TOTAL_TESTS} run | ${PASSED_TESTS} passed | ${FAILED_TESTS} failed"
  echo "  Modules: ${TOTAL_MODULES} total | ${PASSED_MODULES} passed | ${FAILED_MODULES} failed"
  echo "  Coverage reports : ${COVERAGE_DIR}/"
  echo "  Per-module logs  : ${UNIT_DIR}/*.log"
  echo "  Failed test list : ${UNIT_DIR}/failures.txt"
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
