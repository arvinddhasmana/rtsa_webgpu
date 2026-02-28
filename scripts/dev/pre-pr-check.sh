#!/usr/bin/env bash
# CLASSIFICATION: UNCLASSIFIED
# RTSA Pre-PR Local Gate
# Mirrors all 5 CI security gates locally before opening a Pull Request.
# Run from the repository root: ./scripts/dev/pre-pr-check.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

GREEN='\033[0;32m'
RED='\033[0;31m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
BOLD='\033[1m'
NC='\033[0m'

PASS="${GREEN}[✓]${NC}"
FAIL="${RED}[✗]${NC}"
WARN="${YELLOW}[!]${NC}"
INFO="${CYAN}[ℹ]${NC}"

GATE_ERRORS=0
TOTAL_GATES=0

start_time="$SECONDS"

log_gate()  { echo -e "\n${BOLD}${CYAN}═══ SG-${1}: ${2} ═══${NC}"; TOTAL_GATES=$(( TOTAL_GATES + 1 )); }
log_pass()  { echo -e "${PASS} $*"; }
log_fail()  { echo -e "${FAIL} $*"; GATE_ERRORS=$(( GATE_ERRORS + 1 )); }
log_warn()  { echo -e "${WARN} $*"; }
log_info()  { echo -e "${INFO} $*"; }

# ─────────────────────────────────────────────────────────────
# SG-1: Pre-Build — Secret scan, lint, format, classification headers
# ─────────────────────────────────────────────────────────────
gate_1_pre_build() {
  log_gate 1 "Pre-Build"

  # Secret detection
  if command -v gitleaks &>/dev/null; then
    if gitleaks detect --source="$REPO_ROOT" --no-banner -q 2>/dev/null; then
      log_pass "Secret scan (gitleaks) — no secrets detected"
    else
      log_fail "Secret scan (gitleaks) — potential secrets found! Review output above."
    fi
  else
    log_warn "gitleaks not installed — skipping secret scan (REQUIRED for CI)"
  fi

  # Classification header check
  local missing_headers=0
  while IFS= read -r -d '' file; do
    # Skip generated files and binaries
    case "$file" in
      */.git/*|*/node_modules/*|*/vendor/*|*/.coverage/*|*/certs/*) continue ;;
      *.go|*.ts|*.tsx|*.proto|*.yaml|*.yml|*.sh|*.sql|*.md)
        if ! head -5 "$file" | grep -qi "CLASSIFICATION"; then
          log_warn "  Missing classification header: ${file#${REPO_ROOT}/}"
          missing_headers=$(( missing_headers + 1 ))
        fi
        ;;
    esac
  done < <(find "$REPO_ROOT" \
    -not -path '*/.git/*' \
    -not -path '*/node_modules/*' \
    -not -path '*/vendor/*' \
    -not -path '*/.coverage/*' \
    -not -path '*/certs/*' \
    -not -path '*/generated/*' \
    \( -name "*.go" -o -name "*.ts" -o -name "*.tsx" -o -name "*.proto" \
       -o -name "*.yaml" -o -name "*.yml" -o -name "*.sh" -o -name "*.sql" \) \
    -print0 2>/dev/null)

  if [ "$missing_headers" -eq 0 ]; then
    log_pass "Classification headers — all source files have headers"
  else
    log_fail "Classification headers — ${missing_headers} file(s) missing CLASSIFICATION header"
  fi

  # Go formatting
  if command -v gofmt &>/dev/null && [ -d "${REPO_ROOT}/services" ]; then
    local unformatted
    unformatted="$(find "${REPO_ROOT}/services" -name "*.go" \
      -not -path '*/vendor/*' \
      -exec gofmt -l {} \; 2>/dev/null | head -20)"

    if [ -z "$unformatted" ]; then
      log_pass "Go formatting (gofmt) — all files formatted"
    else
      log_fail "Go formatting (gofmt) — unformatted files:"
      echo "$unformatted" | while read -r f; do echo "  ${f}"; done
      log_info "Fix with: gofmt -w ./services/..."
    fi
  fi

  # Buf proto formatting
  if command -v buf &>/dev/null && [ -d "${REPO_ROOT}/proto" ]; then
    if (cd "$REPO_ROOT" && buf format --diff proto/ 2>/dev/null | grep -q '.') 2>/dev/null; then
      log_fail "Proto formatting (buf) — files need formatting. Run: buf format -w proto/"
    else
      log_pass "Proto formatting (buf) — all proto files formatted"
    fi
  fi

  # TypeScript formatting (prettier check)
  if [ -d "${REPO_ROOT}/web-cop" ] && command -v pnpm &>/dev/null; then
    if (cd "${REPO_ROOT}/web-cop" && pnpm prettier --check . 2>/dev/null); then
      log_pass "TypeScript formatting (prettier) — all files formatted"
    else
      log_fail "TypeScript formatting (prettier) — files need formatting. Run: pnpm prettier --write ."
    fi
  fi
}

# ─────────────────────────────────────────────────────────────
# SG-2: Build — Compile, proto gen, TypeScript compile
# ─────────────────────────────────────────────────────────────
gate_2_build() {
  log_gate 2 "Build"

  # Proto generation
  if command -v buf &>/dev/null && [ -d "${REPO_ROOT}/proto" ]; then
    if (cd "$REPO_ROOT" && buf lint proto/ 2>/dev/null); then
      log_pass "Protobuf lint (buf) — no issues"
    else
      log_fail "Protobuf lint (buf) — lint errors found"
    fi

    if (cd "$REPO_ROOT" && buf generate proto/ 2>/dev/null); then
      log_pass "Protobuf generation — code generated successfully"
    else
      log_fail "Protobuf generation — code generation failed"
    fi
  fi

  # Go compilation
  if command -v go &>/dev/null && [ -d "${REPO_ROOT}/services" ]; then
    local build_errors=0
    while IFS= read -r -d '' mod_file; do
      local service_dir
      service_dir="$(dirname "$mod_file")"
      local service_name
      service_name="$(basename "$service_dir")"

      if (cd "$service_dir" && go build ./... 2>/dev/null); then
        log_pass "  Go build: ${service_name}"
      else
        log_fail "  Go build: ${service_name} — COMPILATION ERROR"
        build_errors=$(( build_errors + 1 ))
      fi
    done < <(find "${REPO_ROOT}/services" -name "go.mod" -print0 2>/dev/null)
  fi

  # TypeScript compilation
  if [ -d "${REPO_ROOT}/web-cop" ] && command -v pnpm &>/dev/null; then
    if (cd "${REPO_ROOT}/web-cop" && pnpm tsc --noEmit 2>/dev/null); then
      log_pass "TypeScript compilation (tsc) — no errors"
    else
      log_fail "TypeScript compilation (tsc) — type errors found"
    fi
  fi
}

# ─────────────────────────────────────────────────────────────
# SG-3: Test — Unit tests, coverage ≥ 80%
# ─────────────────────────────────────────────────────────────
gate_3_test() {
  log_gate 3 "Test (Unit Tests + Coverage)"

  local test_script="${SCRIPT_DIR}/test-go.sh"
  if [ -f "$test_script" ]; then
    if bash "$test_script"; then
      log_pass "Go unit tests — PASSED (≥ 80% coverage)"
    else
      log_fail "Go unit tests — FAILED (coverage or test failures)"
    fi
  else
    if command -v go &>/dev/null && [ -d "${REPO_ROOT}/services" ]; then
      log_info "Running quick Go tests (no coverage threshold)..."
      while IFS= read -r -d '' mod_file; do
        local service_dir
        service_dir="$(dirname "$mod_file")"
        local service_name
        service_name="$(basename "$service_dir")"
        if (cd "$service_dir" && go test -race -timeout 120s ./... 2>/dev/null); then
          log_pass "  ${service_name} — tests PASSED"
        else
          log_fail "  ${service_name} — tests FAILED"
        fi
      done < <(find "${REPO_ROOT}/services" -name "go.mod" -print0 2>/dev/null)
    fi
  fi

  # Frontend tests
  if [ -d "${REPO_ROOT}/web-cop" ] && command -v pnpm &>/dev/null; then
    if (cd "${REPO_ROOT}/web-cop" && pnpm test --coverage --run 2>/dev/null); then
      log_pass "Frontend unit tests (vitest) — PASSED"
    else
      log_fail "Frontend unit tests (vitest) — FAILED"
    fi
  fi
}

# ─────────────────────────────────────────────────────────────
# SG-4: Security — SAST, dependency scan
# ─────────────────────────────────────────────────────────────
gate_4_security() {
  log_gate 4 "Security (SAST + Dependency Scan)"

  # gosec — Go SAST
  if command -v gosec &>/dev/null && [ -d "${REPO_ROOT}/services" ]; then
    local gosec_output
    if gosec_output="$(gosec -quiet -severity high ./services/... 2>&1)"; then
      log_pass "Go SAST (gosec) — no HIGH/CRITICAL issues"
    else
      log_fail "Go SAST (gosec) — HIGH or CRITICAL issues found:"
      echo "$gosec_output" | head -30 | sed 's/^/  /'
    fi
  else
    log_warn "gosec not installed — skipping Go SAST (REQUIRED for CI)"
  fi

  # govulncheck — Go dependency vulnerabilities
  if command -v govulncheck &>/dev/null && [ -d "${REPO_ROOT}/services" ]; then
    local vuln_errors=0
    while IFS= read -r -d '' mod_file; do
      local service_dir
      service_dir="$(dirname "$mod_file")"
      local service_name
      service_name="$(basename "$service_dir")"
      if (cd "$service_dir" && govulncheck ./... 2>/dev/null); then
        log_pass "  govulncheck: ${service_name} — no vulnerabilities"
      else
        log_fail "  govulncheck: ${service_name} — vulnerabilities found"
        vuln_errors=$(( vuln_errors + 1 ))
      fi
    done < <(find "${REPO_ROOT}/services" -name "go.mod" -print0 2>/dev/null)
  else
    log_warn "govulncheck not installed — skipping Go vulnerability scan (REQUIRED for CI)"
  fi

  # golangci-lint — Go linting
  if command -v golangci-lint &>/dev/null && [ -d "${REPO_ROOT}/services" ]; then
    local lint_errors=0
    while IFS= read -r -d '' mod_file; do
      local service_dir
      service_dir="$(dirname "$mod_file")"
      local service_name
      service_name="$(basename "$service_dir")"
      if (cd "$service_dir" && golangci-lint run ./... 2>/dev/null); then
        log_pass "  golangci-lint: ${service_name} — no issues"
      else
        log_fail "  golangci-lint: ${service_name} — lint issues found"
        lint_errors=$(( lint_errors + 1 ))
      fi
    done < <(find "${REPO_ROOT}/services" -name "go.mod" -print0 2>/dev/null)
  else
    log_warn "golangci-lint not installed — skipping Go lint (REQUIRED for CI)"
  fi

  # semgrep — multi-language SAST
  if command -v semgrep &>/dev/null; then
    if semgrep scan --config=auto --error --quiet "$REPO_ROOT" 2>/dev/null; then
      log_pass "Multi-language SAST (semgrep) — no critical findings"
    else
      log_warn "semgrep found issues — review before PR (will block CI if Critical/High)"
    fi
  else
    log_warn "semgrep not installed — skipping multi-language SAST"
  fi

  # npm audit — frontend dependencies
  if [ -d "${REPO_ROOT}/web-cop" ] && command -v pnpm &>/dev/null; then
    if (cd "${REPO_ROOT}/web-cop" && pnpm audit --audit-level=high 2>/dev/null); then
      log_pass "Frontend dependency audit (npm audit) — no HIGH/CRITICAL vulnerabilities"
    else
      log_fail "Frontend dependency audit — HIGH or CRITICAL vulnerabilities found"
      log_info "  Run: cd web-cop && pnpm audit --fix"
    fi
  fi
}

# ─────────────────────────────────────────────────────────────
# SG-5: Integration (optional — requires running dev stack)
# ─────────────────────────────────────────────────────────────
gate_5_integration() {
  log_gate 5 "Integration (optional — skips if dev stack not running)"

  # Check if dev stack is up
  if ! curl -sf "http://localhost:19644/v1/cluster/health_overview" &>/dev/null; then
    log_warn "Dev stack not running — skipping integration tests"
    log_info "Start it with: docker compose -f deploy/docker-compose.yml up -d"
    return
  fi

  local int_script="${SCRIPT_DIR}/test-integration.sh"
  if [ -f "$int_script" ]; then
    if bash "$int_script"; then
      log_pass "Integration tests — PASSED"
    else
      log_fail "Integration tests — FAILED"
    fi
  else
    log_info "Integration test script not found — skipping"
  fi
}

# ─────────────────────────────────────────────────────────────
# Final summary
# ─────────────────────────────────────────────────────────────
print_summary() {
  local elapsed=$(( SECONDS - start_time ))

  echo ""
  echo "══════════════════════════════════════════════════════"
  echo "  Pre-PR Gate Results"
  echo "  Gates run: ${TOTAL_GATES} | Elapsed: ${elapsed}s"
  echo "──────────────────────────────────────────────────────"

  if [ "$GATE_ERRORS" -eq 0 ]; then
    echo -e "${GREEN}${BOLD}  ALL GATES PASSED — Ready to open Pull Request${NC}"
    echo ""
    echo "  Open PR against main:"
    echo "  git push origin \$(git rev-parse --abbrev-ref HEAD)"
    echo "  Then create PR at: https://github.com/<org>/rtsa/pulls"
  else
    echo -e "${RED}${BOLD}  ${GATE_ERRORS} GATE(S) FAILED — Fix issues before opening PR${NC}"
    echo ""
    echo "  Resolve all failures listed above, then re-run:"
    echo "  ./scripts/dev/pre-pr-check.sh"
  fi
  echo "══════════════════════════════════════════════════════"
}

# ─────────────────────────────────────────────────────────────
# Main
# ─────────────────────────────────────────────────────────────
main() {
  echo ""
  echo -e "${BOLD}${CYAN}╔══════════════════════════════════════════════════╗${NC}"
  echo -e "${BOLD}${CYAN}║  RTSA Pre-PR Local Gate                          ║${NC}"
  echo -e "${BOLD}${CYAN}║  CLASSIFICATION: UNCLASSIFIED                    ║${NC}"
  echo -e "${BOLD}${CYAN}╚══════════════════════════════════════════════════╝${NC}"

  cd "$REPO_ROOT"

  gate_1_pre_build
  gate_2_build
  gate_3_test
  gate_4_security
  gate_5_integration

  print_summary

  [ "$GATE_ERRORS" -eq 0 ]
}

main "$@"
