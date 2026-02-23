#!/usr/bin/env bash
# CLASSIFICATION: UNCLASSIFIED
# RTSA ClickHouse Schema Initializer
# Creates the RTSA database and applies schema migrations.
# Requires the ClickHouse dev container to be running: docker compose up -d

set -euo pipefail

CLICKHOUSE_HOST="${CLICKHOUSE_HOST:-localhost}"
CLICKHOUSE_PORT="${CLICKHOUSE_PORT:-9000}"
CLICKHOUSE_HTTP_PORT="${CLICKHOUSE_HTTP_PORT:-8123}"
CLICKHOUSE_DATABASE="${CLICKHOUSE_DATABASE:-rtsa}"
CLICKHOUSE_USER="${CLICKHOUSE_USER:-rtsa_dev}"
CLICKHOUSE_PASSWORD="${CLICKHOUSE_PASSWORD:-dev_password_change_me}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
MIGRATIONS_DIR="${REPO_ROOT}/deploy/clickhouse/migrations"

GREEN='\033[0;32m'
RED='\033[0;31m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${CYAN}[ℹ]${NC} $*"; }
log_pass() { echo -e "${GREEN}[✓]${NC} $*"; }
log_fail() { echo -e "${RED}[✗]${NC} $*"; exit 1; }
log_warn() { echo -e "${YELLOW}[!]${NC} $*"; }

# ─────────────────────────────────────────────────────────────
# Execute a ClickHouse query via HTTP interface
# ─────────────────────────────────────────────────────────────
ch_query() {
  local query="$1"
  curl -sf \
    "http://${CLICKHOUSE_HOST}:${CLICKHOUSE_HTTP_PORT}/" \
    --data-urlencode "query=${query}" \
    -u "${CLICKHOUSE_USER}:${CLICKHOUSE_PASSWORD}"
}

# Execute a query against the RTSA database
ch_query_db() {
  local query="$1"
  curl -sf \
    "http://${CLICKHOUSE_HOST}:${CLICKHOUSE_HTTP_PORT}/?database=${CLICKHOUSE_DATABASE}" \
    --data-urlencode "query=${query}" \
    -u "${CLICKHOUSE_USER}:${CLICKHOUSE_PASSWORD}"
}

# ─────────────────────────────────────────────────────────────
# Wait for ClickHouse to be ready
# ─────────────────────────────────────────────────────────────
wait_for_clickhouse() {
  log_info "Waiting for ClickHouse at ${CLICKHOUSE_HOST}:${CLICKHOUSE_HTTP_PORT} ..."
  local max_attempts=30
  local attempt=0

  while ! curl -sf "http://${CLICKHOUSE_HOST}:${CLICKHOUSE_HTTP_PORT}/ping" &>/dev/null; do
    attempt=$(( attempt + 1 ))
    if [ "$attempt" -ge "$max_attempts" ]; then
      log_fail "ClickHouse did not become ready. Is the dev stack running?"
    fi
    printf "."
    sleep 2
  done
  echo ""
  log_pass "ClickHouse is ready"
}

# ─────────────────────────────────────────────────────────────
# Create database
# ─────────────────────────────────────────────────────────────
create_database() {
  log_info "Creating database: ${CLICKHOUSE_DATABASE}"
  ch_query "CREATE DATABASE IF NOT EXISTS ${CLICKHOUSE_DATABASE}" &>/dev/null \
    && log_pass "Database '${CLICKHOUSE_DATABASE}' created/verified"
}

# ─────────────────────────────────────────────────────────────
# Run inline schema for core RTSA tables
# See: docs/sdlc_guidelines/08_tech_specific/clickhouse_guidelines.md
# ─────────────────────────────────────────────────────────────
apply_inline_schema() {
  log_info "Applying core RTSA schema..."

  # sensor_events table
  ch_query_db "
-- CLASSIFICATION: UNCLASSIFIED
CREATE TABLE IF NOT EXISTS sensor_events
(
    event_id        String,
    sensor_id       String,
    sensor_type     Enum8(
        'RADAR' = 1,
        'EW_SIGINT' = 2,
        'ELINT_COMINT' = 3,
        'ISR' = 4,
        'AIS_BFT' = 5,
        'CYBER' = 6
    ),
    event_time      DateTime64(3, 'UTC'),
    latitude        Float64,
    longitude       Float64,
    altitude        Float64,
    speed_ms        Float64,
    heading_deg     Float64,
    classification  Enum8(
        'UNCLASSIFIED' = 0,
        'PROTECTED_A' = 1,
        'PROTECTED_B' = 2,
        'PROTECTED_C' = 3,
        'SECRET' = 4
    ),
    raw_payload     String,
    ingestion_time  DateTime64(3, 'UTC') DEFAULT now64(3)
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(event_time)
ORDER BY (sensor_type, sensor_id, event_time)
TTL toDateTime(event_time) + INTERVAL 90 DAY
SETTINGS index_granularity = 8192;
" &>/dev/null && log_pass "  sensor_events table ready"

  # entity_tracks table
  ch_query_db "
-- CLASSIFICATION: UNCLASSIFIED
CREATE TABLE IF NOT EXISTS entity_tracks
(
    track_id        String,
    entity_type     Enum8(
        'AIR' = 1,
        'SURFACE' = 2,
        'SUBSURFACE' = 3,
        'LAND' = 4,
        'SPACE' = 5,
        'CYBER' = 6
    ),
    hostile_status  Enum8(
        'UNKNOWN' = 0,
        'PENDING' = 1,
        'FRIENDLY' = 2,
        'NEUTRAL' = 3,
        'HOSTILE' = 4,
        'SUSPECT' = 5
    ),
    track_time      DateTime64(3, 'UTC'),
    latitude        Float64,
    longitude       Float64,
    altitude        Float64,
    speed_ms        Float64,
    heading_deg     Float64,
    confidence      Float32,
    source_sensors  Array(String),
    classification  Enum8(
        'UNCLASSIFIED' = 0,
        'PROTECTED_A' = 1,
        'PROTECTED_B' = 2,
        'PROTECTED_C' = 3,
        'SECRET' = 4
    ),
    ingestion_time  DateTime64(3, 'UTC') DEFAULT now64(3)
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(track_time)
ORDER BY (entity_type, track_id, track_time)
TTL toDateTime(track_time) + INTERVAL 90 DAY
SETTINGS index_granularity = 8192;
" &>/dev/null && log_pass "  entity_tracks table ready"

  # anomaly_scores table
  ch_query_db "
-- CLASSIFICATION: UNCLASSIFIED
CREATE TABLE IF NOT EXISTS anomaly_scores
(
    score_id        String,
    track_id        String,
    anomaly_type    LowCardinality(String),
    score           Float32,
    confidence      Float32,
    explanation     String,
    score_time      DateTime64(3, 'UTC'),
    model_version   String,
    classification  Enum8(
        'UNCLASSIFIED' = 0,
        'PROTECTED_A' = 1,
        'PROTECTED_B' = 2,
        'PROTECTED_C' = 3,
        'SECRET' = 4
    ),
    ingestion_time  DateTime64(3, 'UTC') DEFAULT now64(3)
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(score_time)
ORDER BY (track_id, score_time)
TTL toDateTime(score_time) + INTERVAL 90 DAY
SETTINGS index_granularity = 8192;
" &>/dev/null && log_pass "  anomaly_scores table ready"

  # audit_events table — append-only, no TTL (immutable audit trail)
  ch_query_db "
-- CLASSIFICATION: UNCLASSIFIED
-- ITSG-33: AU-3 — Content of Audit Records
CREATE TABLE IF NOT EXISTS audit_events
(
    event_id        String,
    event_type      LowCardinality(String),
    service         LowCardinality(String),
    operator_id     String,
    resource_type   LowCardinality(String),
    resource_id     String,
    action          LowCardinality(String),
    outcome         Enum8('SUCCESS' = 1, 'FAILURE' = 2),
    event_time      DateTime64(3, 'UTC'),
    client_ip       IPv6,
    details         String,
    classification  Enum8(
        'UNCLASSIFIED' = 0,
        'PROTECTED_A' = 1,
        'PROTECTED_B' = 2,
        'PROTECTED_C' = 3,
        'SECRET' = 4
    ),
    ingestion_time  DateTime64(3, 'UTC') DEFAULT now64(3)
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(event_time)
ORDER BY (service, event_type, event_time)
SETTINGS index_granularity = 8192;
" &>/dev/null && log_pass "  audit_events table ready"

  # feedback_events table
  ch_query_db "
-- CLASSIFICATION: UNCLASSIFIED
CREATE TABLE IF NOT EXISTS feedback_events
(
    feedback_id     String,
    operator_id     String,
    track_id        String,
    score_id        String,
    feedback_type   Enum8('CONFIRM' = 1, 'DISMISS' = 2, 'ESCALATE' = 3, 'RECLASSIFY' = 4),
    trust_score     Float32,
    validated       UInt8 DEFAULT 0,
    feedback_time   DateTime64(3, 'UTC'),
    classification  Enum8(
        'UNCLASSIFIED' = 0,
        'PROTECTED_A' = 1,
        'PROTECTED_B' = 2,
        'PROTECTED_C' = 3,
        'SECRET' = 4
    ),
    ingestion_time  DateTime64(3, 'UTC') DEFAULT now64(3)
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(feedback_time)
ORDER BY (operator_id, feedback_time)
SETTINGS index_granularity = 8192;
" &>/dev/null && log_pass "  feedback_events table ready"
}

# ─────────────────────────────────────────────────────────────
# Run SQL migration files if they exist
# ─────────────────────────────────────────────────────────────
apply_migration_files() {
  if [ ! -d "$MIGRATIONS_DIR" ]; then
    log_info "No migration files directory found at ${MIGRATIONS_DIR} — using inline schema only"
    return
  fi

  log_info "Applying SQL migration files from ${MIGRATIONS_DIR}..."

  local applied=0
  while IFS= read -r -d '' sql_file; do
    local filename
    filename="$(basename "$sql_file")"
    log_info "  Applying migration: ${filename}"

    ch_query_db "$(cat "$sql_file")" &>/dev/null \
      && log_pass "  ${filename} — applied" \
      || log_warn "  ${filename} — had warnings (may already be applied)"

    applied=$(( applied + 1 ))
  done < <(find "$MIGRATIONS_DIR" -name "*.sql" -print0 | sort -z)

  if [ "$applied" -eq 0 ]; then
    log_info "No SQL migration files found in ${MIGRATIONS_DIR}"
  else
    log_pass "${applied} migration file(s) applied"
  fi
}

# ─────────────────────────────────────────────────────────────
# Verify schema
# ─────────────────────────────────────────────────────────────
verify_schema() {
  echo ""
  log_info "Tables in '${CLICKHOUSE_DATABASE}' database:"

  local tables
  tables="$(ch_query "SHOW TABLES FROM ${CLICKHOUSE_DATABASE}" 2>/dev/null)" || {
    log_warn "Could not list tables"
    return
  }

  echo "$tables" | while read -r table; do
    [ -n "$table" ] && log_pass "  ${table}"
  done
}

# ─────────────────────────────────────────────────────────────
# Main
# ─────────────────────────────────────────────────────────────
main() {
  echo ""
  echo -e "${CYAN}══ RTSA ClickHouse Schema Initializer (UNCLASSIFIED) ══${NC}"

  wait_for_clickhouse
  create_database
  apply_inline_schema
  apply_migration_files
  verify_schema

  echo ""
  echo -e "${GREEN}Done.${NC} ClickHouse schema is ready for development."
}

main "$@"
