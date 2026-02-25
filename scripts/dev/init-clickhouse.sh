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
CLICKHOUSE_USER="${CLICKHOUSE_USER:-default}"
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
    -d "${query}" \
    -u "${CLICKHOUSE_USER}:${CLICKHOUSE_PASSWORD}"
}

# Execute a query against the RTSA database
ch_query_db() {
  local query="$1"
  curl -sf \
    "http://${CLICKHOUSE_HOST}:${CLICKHOUSE_HTTP_PORT}/?database=${CLICKHOUSE_DATABASE}" \
    -d "${query}" \
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

  # sensor_observations table
  ch_query_db "
-- CLASSIFICATION: UNCLASSIFIED
CREATE TABLE IF NOT EXISTS sensor_observations
(
    observation_id       String,
    sensor_id            String,
    sensor_type          LowCardinality(String),
    latitude             Float64,
    longitude            Float64,
    altitude_meters      Float64,
    speed_knots          Float64,
    heading_degrees      Float64,
    classification_level LowCardinality(String),
    metadata_json        String,
    event_time           DateTime64(3, 'UTC'),
    ingestion_time       DateTime64(3, 'UTC') DEFAULT now64(3)
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(event_time)
ORDER BY (sensor_type, sensor_id, event_time)
TTL toDateTime(event_time) + INTERVAL 90 DAY
SETTINGS index_granularity = 8192;
" &>/dev/null && log_pass "  sensor_observations table ready"

  # tracks_fused table
  ch_query_db "
-- CLASSIFICATION: UNCLASSIFIED
CREATE TABLE IF NOT EXISTS tracks_fused
(
    track_id               String,
    entity_type            LowCardinality(String),
    hostile_classification LowCardinality(String),
    latitude               Float64,
    longitude              Float64,
    altitude_meters        Float64,
    speed_knots            Float64,
    heading_degrees        Float64,
    confidence_score       Float32,
    source_count           UInt16,
    source_sensors         Array(String),
    classification_level   LowCardinality(String),
    track_status           LowCardinality(String),
    event_time             DateTime64(3, 'UTC'),
    ingestion_time         DateTime64(3, 'UTC') DEFAULT now64(3)
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(event_time)
ORDER BY (entity_type, track_id, event_time)
TTL toDateTime(event_time) + INTERVAL 90 DAY
SETTINGS index_granularity = 8192;
" &>/dev/null && log_pass "  tracks_fused table ready"

  # anomaly_detections table
  ch_query_db "
-- CLASSIFICATION: UNCLASSIFIED
CREATE TABLE IF NOT EXISTS anomaly_detections
(
    alert_id             String,
    track_id             String,
    anomaly_type         LowCardinality(String),
    severity             LowCardinality(String),
    confidence_score     Float32,
    explanation          String,
    model_version        String,
    classification_level LowCardinality(String),
    event_time           DateTime64(3, 'UTC'),
    ingestion_time       DateTime64(3, 'UTC') DEFAULT now64(3)
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(event_time)
ORDER BY (track_id, event_time)
TTL toDateTime(event_time) + INTERVAL 90 DAY
SETTINGS index_granularity = 8192;
" &>/dev/null && log_pass "  anomaly_detections table ready"

  # audit_log table — append-only, no TTL (immutable audit trail)
  ch_query_db "
-- CLASSIFICATION: UNCLASSIFIED
-- ITSG-33: AU-3 — Content of Audit Records
-- NO TTL — immutable audit trail (CR-SEC-003)
CREATE TABLE IF NOT EXISTS audit_log
(
    audit_id             String,
    service_id           LowCardinality(String),
    event_type           LowCardinality(String),
    actor_id             String,
    actor_type           LowCardinality(String),
    resource_type        LowCardinality(String),
    resource_id          String,
    action               LowCardinality(String),
    detail_json          String,
    classification_level LowCardinality(String),
    event_time           DateTime64(3, 'UTC'),
    ingestion_time       DateTime64(3, 'UTC') DEFAULT now64(3)
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(event_time)
ORDER BY (service_id, event_type, event_time)
SETTINGS index_granularity = 8192;
" &>/dev/null && log_pass "  audit_log table ready"

  # operator_feedback table
  ch_query_db "
-- CLASSIFICATION: UNCLASSIFIED
CREATE TABLE IF NOT EXISTS operator_feedback
(
    feedback_id          String,
    track_id             String,
    operator_id          String,
    feedback_type        LowCardinality(String),
    justification        String,
    trust_score          Float32,
    clearance_score      Float32,
    accuracy_score       Float32,
    temporal_score       Float32,
    deviation_score      Float32,
    validated            UInt8 DEFAULT 0,
    classification_level LowCardinality(String),
    event_time           DateTime64(3, 'UTC'),
    ingestion_time       DateTime64(3, 'UTC') DEFAULT now64(3)
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(event_time)
ORDER BY (operator_id, event_time)
SETTINGS index_granularity = 8192;
" &>/dev/null && log_pass "  operator_feedback table ready"
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
