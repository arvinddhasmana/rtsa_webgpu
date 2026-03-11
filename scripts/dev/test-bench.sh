#!/usr/bin/env bash
# CLASSIFICATION: UNCLASSIFIED
# RTSA Benchmark Test Runner
# Runs the performance benchmark suite from tests/benchmark/ (build tag: integration).
# All benchmarks are guarded by RTSA_INTEGRATION_TESTS=true.
#
# Usage:
#   ./scripts/dev/test-bench.sh
#   ./scripts/dev/test-bench.sh --benchtime 30s     # override per-benchmark time
#   ./scripts/dev/test-bench.sh --timeout 20m        # override suite timeout
#
# Test results written to:  ${RTSA_TEST_RESULTS_DIR}/benchmark/   (set by test-all.sh)
#                     or:  .test-results/<timestamp>/benchmark/  (standalone run)
#
# Benchmarks (B01–B04):
#   B01 BenchmarkIngestionThroughput  — protobuf marshal/unmarshal throughput
#   B02 BenchmarkFusionLatency        — spatial gating p95 latency (100 synthetic tracks)
#   B03 BenchmarkAnomalyLatency       — feature extraction p95 latency
#   B04 BenchmarkQueryPerformance     — 100-row and 10K-row deserialization p95

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
TESTS_DIR="${REPO_ROOT}/tests"

# ── Defaults ──────────────────────────────────────────────────────────────────
BENCH_TIME="${RTSA_BENCH_TIME:-10s}"
BENCH_TIMEOUT="${RTSA_BENCH_TIMEOUT:-10m}"
BENCH_FILTER="${RTSA_BENCH_FILTER:-.}"   # regex passed to -bench=; "." runs all

# ── Parse optional CLI overrides ──────────────────────────────────────────────
while [[ $# -gt 0 ]]; do
  case "$1" in
    --benchtime)  BENCH_TIME="$2";    shift 2 ;;
    --timeout)    BENCH_TIMEOUT="$2"; shift 2 ;;
    --bench)      BENCH_FILTER="$2";  shift 2 ;;
    *) echo "Unknown option: $1"; exit 1 ;;
  esac
done

# ── Results directory ──────────────────────────────────────────────────────────
_TS="$(date +%Y-%m-%dT%H-%M-%S)"
RESULTS_DIR="${RTSA_TEST_RESULTS_DIR:-.test-results/${_TS}}"
BENCH_DIR="${RESULTS_DIR}/benchmark"
mkdir -p "${BENCH_DIR}"
BENCH_LOG="${BENCH_DIR}/run.log"

# ── Colours ───────────────────────────────────────────────────────────────────
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

# ── Preflight checks ──────────────────────────────────────────────────────────
check_go() {
  if ! command -v go &>/dev/null; then
    echo -e "${RED}ERROR: go binary not found in PATH${NC}"
    exit 1
  fi
}

check_benchmark_pkg() {
  if [ ! -f "${TESTS_DIR}/benchmark/go.mod" ]; then
    echo -e "${RED}ERROR: tests/benchmark/go.mod not found — is REPO_ROOT correct?${NC}"
    echo "  REPO_ROOT=${REPO_ROOT}"
    exit 1
  fi
}

# ── Failure extraction helpers ────────────────────────────────────────────────
extract_failures() {
  local log="$1"
  grep -E '^[[:space:]]*--- FAIL:' "${log}" 2>/dev/null \
    | sed 's|.*--- FAIL: \([^ ]*\).*|benchmark/\1|' \
    >> "${BENCH_DIR}/failures.txt" || true
  # Also capture threshold-exceeded lines from b.Errorf (they appear as FAIL in log)
  grep -E 'THRESHOLD EXCEEDED' "${log}" 2>/dev/null \
    | sed 's/^[[:space:]]*//' \
    >> "${BENCH_DIR}/threshold_violations.txt" || true
}

tally_counts() {
  local log="$1"
  # Benchmarks don't emit --- PASS:/--- FAIL: lines.
  # Count the number of benchmark result lines (lines starting with 'Benchmark').
  local bench_lines
  bench_lines="$(grep -cE '^Benchmark[A-Za-z0-9_]+-[0-9]+' "${log}" 2>/dev/null || echo 0)"
  bench_lines="${bench_lines//[[:space:]]/}"; bench_lines="${bench_lines:-0}"
  echo "${bench_lines} 0"  # format: "<total> <fail>" — fail determined from exit code
}

print_failures() {
  local file="$1" max_shown="${2:-10}" label="$3"
  if [ ! -f "$file" ] || [ ! -s "$file" ]; then return; fi
  local total_fail
  total_fail="$(wc -l < "$file" | tr -d ' ')"
  local shown=$(( total_fail < max_shown ? total_fail : max_shown ))
  echo ""
  echo -e "${RED}${BOLD}── FAILED BENCHMARKS: ${label} [${total_fail} failure(s), showing first ${shown}] ──────${NC}"
  local i=1
  while IFS= read -r line && [ "$i" -le "$max_shown" ]; do
    printf "  %2d. %s\n" "$i" "$line"
    i=$(( i + 1 ))
  done < "$file"
  if [ "$total_fail" -gt "$max_shown" ]; then
    echo "  ... and $(( total_fail - max_shown )) more — see: ${file}"
  fi
}

# ── Main benchmark runner ─────────────────────────────────────────────────────
run_benchmarks() {
  log_info "Running RTSA performance benchmarks ..."
  log_info "  Filter     : -bench=${BENCH_FILTER}"
  log_info "  Bench time : -benchtime=${BENCH_TIME}"
  log_info "  Timeout    : -timeout=${BENCH_TIMEOUT}"
  log_info "  Log        : ${BENCH_LOG}"
  echo ""

  local bench_ok=true
  set +e
  (cd "${TESTS_DIR}" && \
    RTSA_INTEGRATION_TESTS=true go test \
      -v \
      -tags integration \
      -bench="${BENCH_FILTER}" \
      -benchtime="${BENCH_TIME}" \
      -timeout "${BENCH_TIMEOUT}" \
      -count=1 \
      ./benchmark/... 2>&1) | tee "${BENCH_LOG}" | sed 's/^/    /'
  local _pe=("${PIPESTATUS[@]}")
  set -e

  [ "${_pe[0]}" -eq 0 ] || bench_ok=false

  # Tally how many benchmark functions ran and whether any failed.
  # Go benchmarks don't emit --- PASS:/--- FAIL: lines.
  # Use suite exit code to determine overall pass/fail.
  local bt
  bt="$(grep -cE '^Benchmark[A-Za-z0-9_]+-[0-9]+' "${BENCH_LOG}" 2>/dev/null || true)"
  bt="${bt//[[:space:]]/}"; bt="${bt:-0}"
  local bf
  if [ "$bench_ok" = true ]; then
    bf=0
  else
    # Count threshold violations as failures; fall back to marking all as failed
    bf="$(grep -cE 'THRESHOLD EXCEEDED' "${BENCH_LOG}" 2>/dev/null || true)"
    bf="${bf//[[:space:]]/}"; bf="${bf:-0}"
    [ "$bf" -eq 0 ] && bf="$bt"
  fi
  local bp=$(( bt - bf ))

  extract_failures "${BENCH_LOG}"

  printf 'TOTAL=%d PASSED=%d FAILED=%d\n' "$bt" "$bp" "$bf" > "${BENCH_DIR}/counts.txt"

  # Print result lines from log for easy reading
  echo ""
  log_info "Benchmark result lines from log:"
  grep -E '^Benchmark' "${BENCH_LOG}" 2>/dev/null | sed 's/^/  /' || true

  # Print threshold violations if any
  if [ -s "${BENCH_DIR}/threshold_violations.txt" ]; then
    echo ""
    log_warn "Threshold violations:"
    sed 's/^/  /' "${BENCH_DIR}/threshold_violations.txt"
  fi

  echo ""
  echo "─────────────────────────────────────────────────────────────────────"
  echo "  Benchmark summary:"
  echo "    Benchmarks : ${bt} run | ${bp} passed | ${bf} failed"
  echo "    Log        : ${BENCH_LOG}"
  if [ -s "${BENCH_DIR}/failures.txt" ]; then
    echo "    Failures   : ${BENCH_DIR}/failures.txt"
  fi
  if [ -s "${BENCH_DIR}/threshold_violations.txt" ]; then
    echo "    Violations : ${BENCH_DIR}/threshold_violations.txt"
  fi
  echo "─────────────────────────────────────────────────────────────────────"

  print_failures "${BENCH_DIR}/failures.txt" 10 "Benchmarks"

  if [ "$bench_ok" = true ] && [ "$bf" -eq 0 ]; then
    log_pass "Performance Benchmarks — PASSED (${bp} benchmark(s) OK)"
  else
    log_fail "Performance Benchmarks — FAILED"
    exit 1
  fi
}

# ── Entry point ───────────────────────────────────────────────────────────────
main() {
  echo ""
  echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
  echo -e "${CYAN}  RTSA Performance Benchmarks (CLASSIFICATION: UNCLASSIFIED)  ${NC}"
  echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
  echo "  Results : ${BENCH_DIR}/"

  check_go
  check_benchmark_pkg
  run_benchmarks
}

main "$@"
