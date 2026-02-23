#!/usr/bin/env bash
# CLASSIFICATION: UNCLASSIFIED
# RTSA Development Environment Health Check
# Verifies all services and dependencies are running correctly.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

REDPANDA_ADMIN_URL="${REDPANDA_ADMIN_URL:-http://localhost:19644}"
REDPANDA_BROKER="${REDPANDA_BROKER:-localhost:19092}"
CLICKHOUSE_HTTP="${CLICKHOUSE_HTTP:-http://localhost:8123}"
CLICKHOUSE_USER="${CLICKHOUSE_USER:-rtsa_dev}"
CLICKHOUSE_PASSWORD="${CLICKHOUSE_PASSWORD:-dev_password_change_me}"

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

ERRORS=0

log_pass() { echo -e "${PASS} $*"; }
log_fail() { echo -e "${FAIL} $*"; ERRORS=$(( ERRORS + 1 )); }
log_warn() { echo -e "${WARN} $*"; }
log_info() { echo -e "${INFO} $*"; }
log_section() { echo -e "\n${BOLD}${CYAN}── $* ──${NC}"; }

# ─────────────────────────────────────────────────────────────
# Docker checks
# ─────────────────────────────────────────────────────────────
check_docker() {
  log_section "Docker"

  if ! command -v docker &>/dev/null; then
    log_fail "Docker not found"
    return
  fi

  if docker info &>/dev/null; then
    log_pass "Docker daemon running"
  else
    log_fail "Docker daemon not running"
    return
  fi

  # Check compose containers
  local compose_file="${REPO_ROOT}/deploy/docker-compose.yml"
  if [ ! -f "$compose_file" ]; then
    log_warn "deploy/docker-compose.yml not found — skipping container checks"
    return
  fi

  local running_count
  running_count="$(docker compose -f "$compose_file" ps --status running -q 2>/dev/null | wc -l | tr -d ' ')"
  local expected_count
  expected_count="$(docker compose -f "$compose_file" ps -q 2>/dev/null | wc -l | tr -d ' ')"

  if [ "$expected_count" -eq 0 ]; then
    log_warn "No containers defined in compose file"
  elif [ "$running_count" -eq "$expected_count" ]; then
    log_pass "All ${running_count}/${expected_count} dev stack containers running"
  else
    log_fail "Only ${running_count}/${expected_count} containers running"
    log_info "Run: docker compose -f deploy/docker-compose.yml up -d"
  fi
}

# ─────────────────────────────────────────────────────────────
# Redpanda checks
# ─────────────────────────────────────────────────────────────
check_redpanda() {
  log_section "Redpanda"

  if curl -sf "${REDPANDA_ADMIN_URL}/v1/cluster/health_overview" &>/dev/null; then
    log_pass "Redpanda broker reachable (${REDPANDA_ADMIN_URL})"
  else
    log_fail "Redpanda broker not reachable at ${REDPANDA_ADMIN_URL}"
    return
  fi

  # Check cluster health
  local health
  health="$(curl -sf "${REDPANDA_ADMIN_URL}/v1/cluster/health_overview" 2>/dev/null)" || health="{}"
  local is_healthy
  is_healthy="$(echo "$health" | grep -o '"is_healthy":[^,}]*' | cut -d: -f2 | tr -d ' ')"

  if [ "$is_healthy" = "true" ]; then
    log_pass "Redpanda cluster is healthy"
  else
    log_warn "Redpanda cluster health: ${is_healthy:-unknown}"
  fi

  # Count topics
  local topic_count=0
  if command -v docker &>/dev/null; then
    local rp_container
    rp_container="$(docker ps -q -f name=redpanda 2>/dev/null | head -1)" || true
    if [ -n "$rp_container" ]; then
      topic_count="$(docker exec "$rp_container" \
        rpk topic list -X brokers="localhost:9092" 2>/dev/null | tail -n +2 | wc -l | tr -d ' ')" || topic_count=0
    fi
  fi

  if [ "$topic_count" -gt 0 ]; then
    log_pass "Redpanda topics active: ${topic_count}"
  else
    log_warn "No topics found — run: ./scripts/dev/init-topics.sh"
  fi

  # Schema Registry
  if curl -sf "http://localhost:18081/subjects" &>/dev/null; then
    log_pass "Schema Registry reachable (localhost:18081)"
  else
    log_warn "Schema Registry not reachable at localhost:18081"
  fi
}

# ─────────────────────────────────────────────────────────────
# ClickHouse checks
# ─────────────────────────────────────────────────────────────
check_clickhouse() {
  log_section "ClickHouse"

  if curl -sf "${CLICKHOUSE_HTTP}/ping" &>/dev/null; then
    log_pass "ClickHouse reachable (${CLICKHOUSE_HTTP})"
  else
    log_fail "ClickHouse not reachable at ${CLICKHOUSE_HTTP}"
    return
  fi

  # Check RTSA database exists
  local db_exists
  db_exists="$(curl -sf \
    "${CLICKHOUSE_HTTP}/?query=SELECT+name+FROM+system.databases+WHERE+name='rtsa'" \
    -u "${CLICKHOUSE_USER}:${CLICKHOUSE_PASSWORD}" 2>/dev/null)" || db_exists=""

  if [ "$db_exists" = "rtsa" ]; then
    log_pass "Database 'rtsa' exists"
  else
    log_warn "Database 'rtsa' not found — run: ./scripts/dev/init-clickhouse.sh"
    return
  fi

  # Count tables
  local table_count
  table_count="$(curl -sf \
    "${CLICKHOUSE_HTTP}/?database=rtsa&query=SELECT+count()+FROM+system.tables+WHERE+database='rtsa'" \
    -u "${CLICKHOUSE_USER}:${CLICKHOUSE_PASSWORD}" 2>/dev/null)" || table_count=0

  if [ "${table_count:-0}" -gt 0 ]; then
    log_pass "ClickHouse tables in 'rtsa': ${table_count}"
  else
    log_warn "No tables in 'rtsa' — run: ./scripts/dev/init-clickhouse.sh"
  fi
}

# ─────────────────────────────────────────────────────────────
# Observability stack checks
# ─────────────────────────────────────────────────────────────
check_observability() {
  log_section "Observability Stack"

  # Prometheus
  if curl -sf "http://localhost:9090/-/healthy" &>/dev/null; then
    log_pass "Prometheus healthy (localhost:9090)"
  else
    log_warn "Prometheus not reachable at localhost:9090"
  fi

  # Grafana
  if curl -sf "http://localhost:3000/api/health" &>/dev/null; then
    log_pass "Grafana healthy (localhost:3000)"
  else
    log_warn "Grafana not reachable at localhost:3000"
  fi

  # Loki
  if curl -sf "http://localhost:3100/ready" &>/dev/null; then
    log_pass "Loki ready (localhost:3100)"
  else
    log_warn "Loki not reachable at localhost:3100"
  fi
}

# ─────────────────────────────────────────────────────────────
# TLS certificate checks
# ─────────────────────────────────────────────────────────────
check_certs() {
  log_section "TLS Certificates"

  local cert_dir="${REPO_ROOT}/certs/dev"
  local server_cert="${cert_dir}/server.crt"

  if [ ! -d "$cert_dir" ]; then
    log_warn "Dev certificates not found — run: ./scripts/setup/gen-dev-certs.sh"
    return
  fi

  for cert_file in ca.crt server.crt client.crt; do
    if [ -f "${cert_dir}/${cert_file}" ]; then
      if openssl x509 -checkend $((7 * 86400)) -noout -in "${cert_dir}/${cert_file}" 2>/dev/null; then
        local expiry
        expiry="$(openssl x509 -enddate -noout -in "${cert_dir}/${cert_file}" 2>/dev/null | sed 's/notAfter=//')"
        log_pass "  ${cert_file} — valid (expires: ${expiry})"
      else
        log_warn "  ${cert_file} — expires within 7 days; run gen-dev-certs.sh to renew"
      fi
    else
      log_fail "  ${cert_file} — missing"
    fi
  done
}

# ─────────────────────────────────────────────────────────────
# Tool version checks
# ─────────────────────────────────────────────────────────────
check_tools() {
  log_section "Required Tools"

  check_tool_version "go"           "$(go version 2>/dev/null | awk '{print $3}' | sed 's/go//')"     "1.22" "https://go.dev/dl/"
  check_tool_version "buf"          "$(buf --version 2>/dev/null | head -1 | tr -d 'v\n')"             "1.32" "https://github.com/bufbuild/buf/releases"
  check_tool_version "node"         "$(node --version 2>/dev/null | sed 's/v//')"                      "20"   "https://nodejs.org/en/download/"
  check_tool_version "pnpm"         "$(pnpm --version 2>/dev/null | head -1)"                          "9"    "npm install -g pnpm@9"
  check_tool_version "docker"       "$(docker --version 2>/dev/null | awk '{print $3}' | tr -d ',')"   "20"   "https://www.docker.com/products/docker-desktop/"
  check_tool_version "gitleaks"     "$(gitleaks version 2>/dev/null | head -1 | sed 's/[^0-9.]//g')"  "8"    "https://github.com/zricethezav/gitleaks"
  check_tool_version "golangci-lint" "$(golangci-lint --version 2>/dev/null | awk '{print $4}')"       "1.57" "curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh"
  check_tool_version "gosec"        "$(gosec --version 2>/dev/null | head -1 | sed 's/[^0-9.]//g')"   "2"    "go install github.com/securego/gosec/v2/cmd/gosec@latest"
  check_tool_version "trivy"        "$(trivy --version 2>/dev/null | head -1 | sed 's/[^0-9.]//g')"   "0.50" "https://github.com/aquasecurity/trivy/releases"
  check_tool_version "mkcert"       "$(mkcert --version 2>/dev/null | sed 's/v//')"                   "1.4"  "go install filippo.io/mkcert@latest"
}

check_tool_version() {
  local name="$1"
  local installed_version="$2"
  local min_version="$3"
  local install_hint="$4"

  if ! command -v "$name" &>/dev/null; then
    log_fail "  ${name} — not installed (install from: ${install_hint})"
    return
  fi

  if [ -z "$installed_version" ]; then
    log_warn "  ${name} — installed (version unknown)"
    return
  fi

  # Simple major.minor comparison — works for most version strings
  local installed_major installed_minor min_major min_minor
  installed_major="$(echo "$installed_version" | cut -d. -f1)"
  min_major="$(echo "$min_version" | cut -d. -f1)"

  if [ "${installed_major:-0}" -gt "${min_major:-0}" ] 2>/dev/null; then
    log_pass "  ${name} ${installed_version}"
  elif [ "${installed_major:-0}" -eq "${min_major:-0}" ] 2>/dev/null; then
    installed_minor="$(echo "$installed_version" | cut -d. -f2)"
    min_minor="$(echo "$min_version" | cut -d. -f2 | sed 's/[^0-9].*//')"
    if [ "${installed_minor:-0}" -ge "${min_minor:-0}" ] 2>/dev/null; then
      log_pass "  ${name} ${installed_version}"
    else
      log_warn "  ${name} ${installed_version} (>= ${min_version} recommended)"
    fi
  else
    log_fail "  ${name} ${installed_version} (>= ${min_version} required)"
  fi
}

# ─────────────────────────────────────────────────────────────
# Go build check
# ─────────────────────────────────────────────────────────────
check_go_build() {
  log_section "Go Compilation"

  local services_dir="${REPO_ROOT}/services"
  if [ ! -d "$services_dir" ]; then
    log_info "No services/ directory found — skipping build check"
    return
  fi

  local ok=0
  local failed=0
  while IFS= read -r -d '' mod_file; do
    local service_dir
    service_dir="$(dirname "$mod_file")"
    local service_name
    service_name="$(basename "$service_dir")"

    if (cd "$service_dir" && go build ./... 2>/dev/null); then
      log_pass "  ${service_name} — compiles OK"
      ok=$(( ok + 1 ))
    else
      log_fail "  ${service_name} — COMPILATION ERROR"
      failed=$(( failed + 1 ))
    fi
  done < <(find "$services_dir" -name "go.mod" -print0 2>/dev/null)

  if [ "$ok" -eq 0 ] && [ "$failed" -eq 0 ]; then
    log_info "No Go services found in services/"
  fi
}

# ─────────────────────────────────────────────────────────────
# Frontend check
# ─────────────────────────────────────────────────────────────
check_frontend() {
  log_section "Frontend"

  local ui_dir="${REPO_ROOT}/ui"
  if [ ! -d "$ui_dir" ]; then
    log_info "No ui/ directory found — skipping frontend check"
    return
  fi

  if [ -d "${ui_dir}/node_modules" ]; then
    log_pass "Frontend dependencies installed (node_modules present)"
  else
    log_warn "Frontend dependencies not installed — run: cd ui && pnpm install"
  fi

  if [ -f "${ui_dir}/package.json" ] && command -v pnpm &>/dev/null; then
    if (cd "$ui_dir" && pnpm tsc --noEmit 2>/dev/null); then
      log_pass "TypeScript compilation OK"
    else
      log_warn "TypeScript compilation had issues — check ui/ for errors"
    fi
  fi
}

# ─────────────────────────────────────────────────────────────
# Protobuf check
# ─────────────────────────────────────────────────────────────
check_proto() {
  log_section "Protobuf"

  local proto_dir="${REPO_ROOT}/proto"
  if [ ! -d "$proto_dir" ]; then
    log_info "No proto/ directory found — skipping proto check"
    return
  fi

  if ! command -v buf &>/dev/null; then
    log_warn "buf not found — skipping proto lint"
    return
  fi

  if (cd "$REPO_ROOT" && buf lint proto/ 2>/dev/null); then
    log_pass "Protobuf lint passed"
  else
    log_warn "Protobuf lint has warnings — check proto/ definitions"
  fi
}

# ─────────────────────────────────────────────────────────────
# Summary
# ─────────────────────────────────────────────────────────────
print_summary() {
  echo ""
  echo "══════════════════════════════════════════════════════"
  if [ "$ERRORS" -eq 0 ]; then
    echo -e "${GREEN}${BOLD}  All checks passed. Development environment is ready.${NC}"
  else
    echo -e "${RED}${BOLD}  ${ERRORS} check(s) failed. Review issues above.${NC}"
    echo "  See GETTING_STARTED.md for resolution steps."
  fi
  echo "══════════════════════════════════════════════════════"
}

# ─────────────────────────────────────────────────────────────
# Main
# ─────────────────────────────────────────────────────────────
main() {
  echo ""
  echo -e "${BOLD}${CYAN}RTSA Development Environment Health Check${NC}"
  echo -e "${CYAN}CLASSIFICATION: UNCLASSIFIED${NC}"
  echo -e "${CYAN}$(date -u '+%Y-%m-%dT%H:%M:%SZ')${NC}"

  check_tools
  check_docker
  check_redpanda
  check_clickhouse
  check_observability
  check_certs
  check_go_build
  check_frontend
  check_proto

  print_summary
  exit "$ERRORS"
}

main "$@"
