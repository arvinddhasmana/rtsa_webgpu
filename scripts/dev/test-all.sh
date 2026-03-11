#!/usr/bin/env bash
# CLASSIFICATION: UNCLASSIFIED
# RTSA Master Test Runner
# Runs all test types in sequence: unit → frontend → integration → E2E → benchmarks.
# Usage: ./scripts/dev/test-all.sh [--skip-e2e] [--skip-bench]
#
# Test results and logs are saved to:
#   .test-results/<timestamp>/
#     unit/         — per-module Go unit test logs + failures.txt
#     frontend/     — Vitest output log + failures.txt
#     integration/  — integration test logs + failures.txt
#     e2e/          — E2E test log + failures.txt
#     benchmark/    — benchmark run log + failures.txt
#     SUMMARY.txt   — machine-readable aggregate (for AI agents)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

SKIP_E2E=false
SKIP_BENCH=false

for arg in "$@"; do
  case "$arg" in
    --skip-e2e)   SKIP_E2E=true ;;
    --skip-bench) SKIP_BENCH=true ;;
  esac
done

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

# ── Timestamped results directory ────────────────────────────────────────────────
STAGES_RUN=0
STAGES_PASSED=0
STAGES_FAILED=0
STAGES_SKIPPED=0

_TS="$(date +%Y-%m-%dT%H-%M-%S)"
export RTSA_TEST_RESULTS_DIR="${REPO_ROOT}/.test-results/${_TS}"
mkdir -p \
  "${RTSA_TEST_RESULTS_DIR}/unit" \
  "${RTSA_TEST_RESULTS_DIR}/frontend" \
  "${RTSA_TEST_RESULTS_DIR}/integration" \
  "${RTSA_TEST_RESULTS_DIR}/e2e" \
  "${RTSA_TEST_RESULTS_DIR}/benchmark"
# ─────────────────────────────────────────────────────────────────────────────

# Stage tracking — parallel arrays
STAGE_NAMES=()
STAGE_TOTAL=()
STAGE_PASSED=()
STAGE_FAILED=()
STAGE_STATUS=()   # PASSED | FAILED | SKIPPED

start_time=$SECONDS

# ── Read counts.txt written by sub-scripts ─────────────────────────────────
# counts.txt format:  TOTAL=N PASSED=N FAILED=N [...]
read_counts() {
  local file="$1" var_t="$2" var_p="$3" var_f="$4"
  if [ -f "$file" ]; then
    local _t _p _f
    _t="$(grep -m 1 -oP '\bTOTAL=\K[0-9]+' "$file" 2>/dev/null || echo 0)"
    _p="$(grep -m 1 -oP '\bPASSED=\K[0-9]+' "$file" 2>/dev/null || echo 0)"
    _f="$(grep -m 1 -oP '\bFAILED=\K[0-9]+' "$file" 2>/dev/null || echo 0)"
    printf -v "$var_t" '%d' "${_t:-0}"
    printf -v "$var_p" '%d' "${_p:-0}"
    printf -v "$var_f" '%d' "${_f:-0}"
  else
    printf -v "$var_t" '0'; printf -v "$var_p" '0'; printf -v "$var_f" '0'
  fi
}

# ── Print first N failures from failures.txt ─────────────────────────────
print_failures() {
  local file="$1" max_shown="${2:-10}" label="$3"
  if [ ! -f "$file" ] || [ ! -s "$file" ]; then return; fi
  local total_fail
  total_fail="$(wc -l < "$file" | tr -d ' ')"
  local shown=$(( total_fail < max_shown ? total_fail : max_shown ))
  echo ""
  echo -e "${RED}${BOLD}── FAILED TESTS: ${label} [${total_fail} failure(s), showing first ${shown}] ────────${NC}"
  local i=1
  while IFS= read -r line && [ "$i" -le "$max_shown" ]; do
    printf "  %2d. %s\n" "$i" "$line"
    i=$(( i + 1 ))
  done < "$file"
  if [ "$total_fail" -gt "$max_shown" ]; then
    echo "  ... and $(( total_fail - max_shown )) more — full list: ${file}"
  fi
}

# ── Stage runner ────────────────────────────────────────────────────────────
# Usage: run_stage <label> <counts_file> <failures_file> <command...>
run_stage() {
  local label="$1" counts_file="$2" failures_file="$3"
  shift 3

  STAGES_RUN=$(( STAGES_RUN + 1 ))
  STAGE_NAMES+=("$label")

  echo ""
  echo -e "${BOLD}${CYAN}── ${label} ──${NC}"

  local stage_ok=true
  "$@" || stage_ok=false

  local t p f
  read_counts "$counts_file" t p f

  STAGE_TOTAL+=("$t")
  STAGE_PASSED+=("$p")
  STAGE_FAILED+=("$f")

  if [ "$stage_ok" = true ]; then
    log_pass "${label} — PASSED"
    STAGE_STATUS+=("PASSED")
    STAGES_PASSED=$(( STAGES_PASSED + 1 ))
  else
    log_fail "${label} — FAILED"
    STAGE_STATUS+=("FAILED")
    STAGES_FAILED=$(( STAGES_FAILED + 1 ))
  fi
}


# ── Inline frontend runner ─────────────────────────────────────────────────
run_frontend() {
  local fe_dir="${RTSA_TEST_RESULTS_DIR}/frontend"
  local fe_log="${fe_dir}/run.log"

  STAGES_RUN=$(( STAGES_RUN + 1 ))
  STAGE_NAMES+=("Unit — Frontend")
  echo ""
  echo -e "${BOLD}${CYAN}── Frontend Unit Tests (Vitest) ──${NC}"

  local fe_ok=true
  set +e
  (cd "${REPO_ROOT}/web-cop-gpu" && pnpm test 2>&1) | tee "${fe_log}"
  local _pe=("${PIPESTATUS[@]}")
  set -e
  [ "${_pe[0]}" -eq 0 ] || fe_ok=false

  # Vitest/Jest: lines with ✓/✔ = pass; ✗/×/✕ or 'FAIL ' prefix = fail
  local fp ff ft
  fp="$(grep -cE '✓|✔|PASS ' "${fe_log}" 2>/dev/null || true)"
  ff="$(grep -cE '✗|×|✕|FAIL ' "${fe_log}" 2>/dev/null || true)"
  ft=$(( fp + ff ))

  grep -E '✗|×|✕' "${fe_log}" 2>/dev/null \
    | sed 's/^[[:space:]]*//' \
    | sed 's/^[✗×✕][[:space:]]*/frontend\//' \
    >> "${fe_dir}/failures.txt" || true
  grep -E '^ FAIL ' "${fe_log}" 2>/dev/null \
    | sed 's/^ FAIL /frontend\//' \
    >> "${fe_dir}/failures.txt" || true

  printf 'TOTAL=%d PASSED=%d FAILED=%d\n' "$ft" "$fp" "$ff" > "${fe_dir}/counts.txt"

  STAGE_TOTAL+=("$ft")
  STAGE_PASSED+=("$fp")
  STAGE_FAILED+=("$ff")

  if [ "$fe_ok" = true ]; then
    log_pass "Frontend Unit Tests — PASSED"
    STAGE_STATUS+=("PASSED")
    STAGES_PASSED=$(( STAGES_PASSED + 1 ))
  else
    log_fail "Frontend Unit Tests — FAILED"
    STAGE_STATUS+=("FAILED")
    STAGES_FAILED=$(( STAGES_FAILED + 1 ))
  fi
}

# ── Summary printer ────────────────────────────────────────────────────────────
print_summary() {
  local elapsed=$(( SECONDS - start_time ))
  local results_dir="${RTSA_TEST_RESULTS_DIR}"

  echo ""
  echo -e "${BOLD}${CYAN}════════════════════════════════════════════════════════════════${NC}"
  echo -e "${BOLD}${CYAN}  RTSA Master Test Summary  (CLASSIFICATION: UNCLASSIFIED)${NC}"
  echo -e "${BOLD}${CYAN}════════════════════════════════════════════════════════════════${NC}"
  echo ""

  # Table header
  printf "  %-38s  %5s  %7s  %7s  %8s\n" "Stage" "Tests" "Passed" "Failed" "Status"
  printf "  %-38s  %5s  %7s  %7s  %8s\n" \
    "──────────────────────────────────────" "─────" "───────" "───────" "────────"

  local i
  for (( i=0; i<${#STAGE_NAMES[@]}; i++ )); do
    local status="${STAGE_STATUS[$i]}"
    local color="$NC"
    case "$status" in
      PASSED)  color="$GREEN" ;;
      FAILED)  color="$RED" ;;
      SKIPPED) color="$YELLOW" ;;
    esac
    printf "  %-38s  %5s  %7s  %7s  " \
      "${STAGE_NAMES[$i]}" \
      "${STAGE_TOTAL[$i]}" \
      "${STAGE_PASSED[$i]}" \
      "${STAGE_FAILED[$i]}"
    echo -e "${color}${BOLD}${status}${NC}"
  done

  echo ""
  echo "  Stages: ${STAGES_RUN} run | ${STAGES_PASSED} passed | ${STAGES_FAILED} failed | ${STAGES_SKIPPED} skipped"
  echo "  Elapsed: ${elapsed}s"

  # ── Detailed failure breakdown, per failing stage ──────────────────────
  local any_failures=false
  for (( i=0; i<${#STAGE_NAMES[@]}; i++ )); do
    if [ "${STAGE_FAILED[$i]}" -gt 0 ] || [ "${STAGE_STATUS[$i]}" = "FAILED" ]; then
      any_failures=true
      break
    fi
  done

  if [ "$any_failures" = true ]; then
    echo ""
    echo -e "${RED}${BOLD}════════════════════════════════════════════════════════════════${NC}"
    echo -e "${RED}${BOLD}  FAILURE DETAILS BY STAGE${NC}"
    echo -e "${RED}${BOLD}════════════════════════════════════════════════════════════════${NC}"

    local stage_failures_file stage_dir
    for (( i=0; i<${#STAGE_NAMES[@]}; i++ )); do
      [ "${STAGE_STATUS[$i]}" = "SKIPPED" ] && continue
      [ "${STAGE_FAILED[$i]}" -eq 0 ] && [ "${STAGE_STATUS[$i]}" = "PASSED" ] && continue

      case "${STAGE_NAMES[$i]}" in
        "Unit — Go"*)           stage_dir="${results_dir}/unit" ;;
        "Unit — Frontend"*)     stage_dir="${results_dir}/frontend" ;;
        "Integration"*)          stage_dir="${results_dir}/integration" ;;
        "E2E"*)                  stage_dir="${results_dir}/e2e" ;;
        "Benchmark"*)            stage_dir="${results_dir}/benchmark" ;;
        *)                       stage_dir="" ;;
      esac

      if [ -n "$stage_dir" ]; then
        stage_failures_file="${stage_dir}/failures.txt"
        print_failures "${stage_failures_file}" 10 "${STAGE_NAMES[$i]}"
      fi
    done
  fi

  # ── Results directory info ────────────────────────────────────────────
  echo ""
  echo -e "${CYAN}${BOLD}── Test Results Location ────────────────────────────────────────────${NC}"
  echo "  ${results_dir}/"
  echo ""
  echo "  Subdirectories:"
  echo "    unit/         — per-module Go unit test logs (*.log, failures.txt)"
  echo "    frontend/     — Vitest output (run.log, failures.txt)"
  echo "    integration/  — integration test logs (*.log, run.log, failures.txt)"
  echo "    e2e/          — E2E test log (run.log, failures.txt)"
  echo "    benchmark/    — benchmark log (run.log, failures.txt)"
  echo "    SUMMARY.txt   — machine-readable aggregate for AI agents"
  echo ""
  echo "  Each failures.txt  — one <stage/TestName> per line"
  echo "  Each counts.txt    — TOTAL=N PASSED=N FAILED=N for scripted consumers"

  # ── Write SUMMARY.txt ─────────────────────────────────────────────────────
  {
    echo "RTSA TEST SUMMARY"
    echo "CLASSIFICATION: UNCLASSIFIED"
    echo "Run timestamp : ${_TS}"
    echo "Elapsed       : ${elapsed}s"
    echo ""
    echo "STAGES"
    printf "  %-38s  %5s  %7s  %7s  %8s\n" "Stage" "Tests" "Passed" "Failed" "Status"
    for (( i=0; i<${#STAGE_NAMES[@]}; i++ )); do
      printf "  %-38s  %5s  %7s  %7s  %8s\n" \
        "${STAGE_NAMES[$i]}" \
        "${STAGE_TOTAL[$i]}" \
        "${STAGE_PASSED[$i]}" \
        "${STAGE_FAILED[$i]}" \
        "${STAGE_STATUS[$i]}"
    done
    echo ""
    echo "STAGES_RUN=${STAGES_RUN} STAGES_PASSED=${STAGES_PASSED} STAGES_FAILED=${STAGES_FAILED} STAGES_SKIPPED=${STAGES_SKIPPED}"
    echo ""
    echo "RESULTS_DIR=${results_dir}"
    echo ""
    echo "FAILURE DETAILS"
    local stage_dir
    for (( i=0; i<${#STAGE_NAMES[@]}; i++ )); do
      [ "${STAGE_STATUS[$i]}" = "SKIPPED" ] && continue
      [ "${STAGE_FAILED[$i]}" -eq 0 ] && [ "${STAGE_STATUS[$i]}" = "PASSED" ] && continue
      case "${STAGE_NAMES[$i]}" in
        "Unit — Go"*)       stage_dir="${results_dir}/unit" ;;
        "Unit — Frontend"*) stage_dir="${results_dir}/frontend" ;;
        "Integration"*)     stage_dir="${results_dir}/integration" ;;
        "E2E"*)             stage_dir="${results_dir}/e2e" ;;
        "Benchmark"*)       stage_dir="${results_dir}/benchmark" ;;
        *)                  stage_dir="" ;;
      esac
      echo "  [${STAGE_NAMES[$i]}] failures:"
      if [ -n "$stage_dir" ] && [ -f "${stage_dir}/failures.txt" ]; then
        sed 's/^/    /' "${stage_dir}/failures.txt" || true
      else
        echo "    (no failures.txt — check full log)"
      fi
      echo ""
    done
  } > "${results_dir}/SUMMARY.txt"

  log_info "SUMMARY.txt written to: ${results_dir}/SUMMARY.txt"
}

main() {
  echo ""
  echo -e "${BOLD}${CYAN}╔════════════════════════════════════════════════════════════╗${NC}"
  echo -e "${BOLD}${CYAN}║  RTSA Master Test Runner  (CLASSIFICATION: UNCLASSIFIED)  ║${NC}"
  echo -e "${BOLD}${CYAN}╚════════════════════════════════════════════════════════════╝${NC}"
  echo "  Results: ${RTSA_TEST_RESULTS_DIR}/"

  # 1. Go Unit Tests
  run_stage "Unit — Go" \
    "${RTSA_TEST_RESULTS_DIR}/unit/counts.txt" \
    "${RTSA_TEST_RESULTS_DIR}/unit/failures.txt" \
    bash "${SCRIPT_DIR}/test-go.sh"

  # 2. Frontend Unit Tests
  if [ -d "${REPO_ROOT}/web-cop-gpu" ] && command -v pnpm &>/dev/null; then
    run_frontend
  else
    log_warn "Frontend tests — skipped (web-cop-gpu/ not found or pnpm not available)"
    STAGE_NAMES+=("Unit — Frontend")
    STAGE_TOTAL+=(0); STAGE_PASSED+=(0); STAGE_FAILED+=(0)
    STAGE_STATUS+=("SKIPPED")
    STAGES_SKIPPED=$(( STAGES_SKIPPED + 1 ))
  fi

  # 3. Integration Tests
  run_stage "Integration" \
    "${RTSA_TEST_RESULTS_DIR}/integration/counts.txt" \
    "${RTSA_TEST_RESULTS_DIR}/integration/failures.txt" \
    bash "${SCRIPT_DIR}/test-integration.sh"

  # 4. E2E Tests
  if [ "$SKIP_E2E" = true ]; then
    log_warn "E2E Tests — skipped (--skip-e2e)"
    STAGE_NAMES+=("E2E")
    STAGE_TOTAL+=(0); STAGE_PASSED+=(0); STAGE_FAILED+=(0)
    STAGE_STATUS+=("SKIPPED")
    STAGES_SKIPPED=$(( STAGES_SKIPPED + 1 ))
  else
    run_stage "E2E" \
      "${RTSA_TEST_RESULTS_DIR}/e2e/counts.txt" \
      "${RTSA_TEST_RESULTS_DIR}/e2e/failures.txt" \
      bash "${SCRIPT_DIR}/test-e2e.sh"
  fi

  # 5. Benchmarks
  if [ "$SKIP_BENCH" = true ]; then
    log_warn "Benchmarks — skipped (--skip-bench)"
    STAGE_NAMES+=("Benchmarks")
    STAGE_TOTAL+=(0); STAGE_PASSED+=(0); STAGE_FAILED+=(0)
    STAGE_STATUS+=("SKIPPED")
    STAGES_SKIPPED=$(( STAGES_SKIPPED + 1 ))
  else
    run_stage "Benchmarks" \
      "${RTSA_TEST_RESULTS_DIR}/benchmark/counts.txt" \
      "${RTSA_TEST_RESULTS_DIR}/benchmark/failures.txt" \
      bash "${SCRIPT_DIR}/test-bench.sh"
  fi

  # ── Final summary ────────────────────────────────────────────────────────────
  print_summary

  if [ "$STAGES_FAILED" -gt 0 ]; then
    echo ""
    echo -e "${RED}${BOLD}OVERALL: FAILED — ${STAGES_FAILED} stage(s) failed${NC}"
    exit 1
  fi

  echo ""
  echo -e "${GREEN}${BOLD}OVERALL: ALL STAGES PASSED${NC}"
}

main "$@"
