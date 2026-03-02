#!/usr/bin/env bash
# CLASSIFICATION: UNCLASSIFIED
# RTSA Integration Test Runner
# Runs integration tests from the centralized tests/ directory and
# per-service integration tests found in svc-*/internal/integration/.
# Requires: Docker (for testcontainers-go)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

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

  if (cd "$tests_dir" && \
    RTSA_INTEGRATION_TESTS=true go test \
      -v \
      -tags=integration \
      -timeout 10m \
      -count=1 \
      ./integration/... 2>&1 | sed "s/^/    /"); then
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

    # Check for integration test files
    local int_test_count
    int_test_count="$(find "$dir" -name "*_integration_test.go" -o -name "*_integration_*.go" 2>/dev/null | wc -l | tr -d ' ')"

    if [ "$int_test_count" -eq 0 ]; then
      log_info "  ${service_name} — no integration tests (skipping)"
      continue
    fi

    log_info "  Running integration tests for ${service_name} (${int_test_count} test files) ..."

    if (cd "$dir" && \
      RTSA_INTEGRATION_TESTS=true go test \
        -tags=integration \
        -timeout 300s \
        -count=1 \
        -v \
        ./... 2>&1 | sed "s/^/    /"); then
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

  check_docker
  run_centralized_integration
  run_service_integration_tests

  echo ""
  if [ "$ERRORS" -eq 0 ]; then
    echo -e "${GREEN}PASSED — All integration tests passed.${NC}"
  else
    echo -e "${RED}FAILED — ${ERRORS} integration test suite(s) failed.${NC}"
    exit 1
  fi
}

main "$@"
