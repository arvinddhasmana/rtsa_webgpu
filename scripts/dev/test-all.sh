#!/usr/bin/env bash
# CLASSIFICATION: UNCLASSIFIED
# RTSA Master Test Runner
# Runs all test types in sequence: unit → frontend → integration → E2E → benchmarks.
# Usage: ./scripts/dev/test-all.sh [--skip-e2e] [--skip-bench]

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

TOTAL=0
PASSED=0
FAILED=0
SKIPPED=0

run_stage() {
  local name="$1"
  shift
  TOTAL=$(( TOTAL + 1 ))

  echo ""
  echo -e "${BOLD}${CYAN}── ${name} ──${NC}"

  if "$@"; then
    log_pass "${name} — PASSED"
    PASSED=$(( PASSED + 1 ))
  else
    log_fail "${name} — FAILED"
    FAILED=$(( FAILED + 1 ))
  fi
}

start_time=$SECONDS

main() {
  echo ""
  echo -e "${BOLD}${CYAN}╔══════════════════════════════════════════════════╗${NC}"
  echo -e "${BOLD}${CYAN}║  RTSA Master Test Runner                         ║${NC}"
  echo -e "${BOLD}${CYAN}║  CLASSIFICATION: UNCLASSIFIED                    ║${NC}"
  echo -e "${BOLD}${CYAN}╚══════════════════════════════════════════════════╝${NC}"

  # 1. Go Unit Tests
  run_stage "Go Unit Tests" bash "${SCRIPT_DIR}/test-go.sh"

  # 2. Frontend Unit Tests
  if [ -d "${REPO_ROOT}/web-cop" ] && command -v pnpm &>/dev/null; then
    run_stage "Frontend Unit Tests (Vitest)" bash -c "cd '${REPO_ROOT}/web-cop' && pnpm test"
  else
    log_warn "Frontend tests — skipped (web-cop/ not found or pnpm not available)"
    SKIPPED=$(( SKIPPED + 1 ))
  fi

  # 3. Integration Tests
  run_stage "Integration Tests (testcontainers)" bash "${SCRIPT_DIR}/test-integration.sh"

  # 4. E2E Tests
  if [ "$SKIP_E2E" = true ]; then
    log_warn "E2E Tests — skipped (--skip-e2e)"
    SKIPPED=$(( SKIPPED + 1 ))
  else
    run_stage "E2E Tests (Docker Compose)" bash "${SCRIPT_DIR}/test-e2e.sh"
  fi

  # 5. Benchmarks
  if [ "$SKIP_BENCH" = true ]; then
    log_warn "Benchmarks — skipped (--skip-bench)"
    SKIPPED=$(( SKIPPED + 1 ))
  else
    run_stage "Performance Benchmarks" bash -c \
      "cd '${REPO_ROOT}/tests' && RTSA_INTEGRATION_TESTS=true go test -v -tags integration -bench=. -benchtime=10s -timeout 10m ./benchmark/..."
  fi

  # Summary
  local elapsed=$(( SECONDS - start_time ))
  echo ""
  echo "══════════════════════════════════════════════════════"
  echo "  Master Test Summary"
  echo "  Stages: ${TOTAL} run | ${PASSED} passed | ${FAILED} failed | ${SKIPPED} skipped"
  echo "  Elapsed: ${elapsed}s"
  echo "══════════════════════════════════════════════════════"

  if [ "$FAILED" -gt 0 ]; then
    echo -e "${RED}${BOLD}FAILED — ${FAILED} stage(s) failed${NC}"
    exit 1
  fi

  echo -e "${GREEN}${BOLD}ALL STAGES PASSED${NC}"
}

main "$@"
