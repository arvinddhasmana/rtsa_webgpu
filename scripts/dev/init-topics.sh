#!/usr/bin/env bash
# CLASSIFICATION: UNCLASSIFIED
# RTSA Redpanda Topic Initializer
# Creates all required Redpanda topics for local development.
# Requires the Redpanda dev stack to be running: docker compose up -d

set -euo pipefail

REDPANDA_BROKER="${REDPANDA_BROKER:-localhost:19092}"
REDPANDA_ADMIN_URL="${REDPANDA_ADMIN_URL:-http://localhost:19644}"

GREEN='\033[0;32m'
RED='\033[0;31m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${CYAN}[ℹ]${NC} $*"; }
log_pass() { echo -e "${GREEN}[✓]${NC} $*"; }
log_fail() { echo -e "${RED}[✗]${NC} $*"; }
log_warn() { echo -e "${YELLOW}[!]${NC} $*"; }

# ─────────────────────────────────────────────────────────────
# Wait for Redpanda to be ready
# ─────────────────────────────────────────────────────────────
wait_for_redpanda() {
  log_info "Waiting for Redpanda to be ready at ${REDPANDA_ADMIN_URL} ..."
  local max_attempts=30
  local attempt=0

  while ! curl -sf "${REDPANDA_ADMIN_URL}/v1/cluster/health_overview" &>/dev/null; do
    attempt=$(( attempt + 1 ))
    if [ "$attempt" -ge "$max_attempts" ]; then
      log_fail "Redpanda did not become ready in time. Is the dev stack running?"
      log_info "Start it with: docker compose -f deploy/docker-compose.yml up -d"
      exit 1
    fi
    printf "."
    sleep 2
  done
  echo ""
  log_pass "Redpanda is ready"
}

# ─────────────────────────────────────────────────────────────
# Create a single topic via Admin API
# Args: topic_name partitions retention_ms
# ─────────────────────────────────────────────────────────────
create_topic() {
  local topic_name="$1"
  local partitions="${2:-3}"
  local retention_ms="${3:-86400000}"  # default 24h
  local replication="${4:-1}"          # dev uses single-node replication

  local response
  response="$(curl -sf -X POST \
    "${REDPANDA_ADMIN_URL}/v1/topics" \
    -H "Content-Type: application/json" \
    -d "{
      \"name\": \"${topic_name}\",
      \"partition_count\": ${partitions},
      \"replication_factor\": ${replication},
      \"configs\": [
        {\"name\": \"retention.ms\", \"value\": \"${retention_ms}\"}
      ]
    }" 2>&1)" || true

  if echo "$response" | grep -q "already exists" 2>/dev/null; then
    log_info "  ${topic_name} — already exists (skipped)"
  elif echo "$response" | grep -q '"name"' 2>/dev/null; then
    log_pass "  ${topic_name} — created (${partitions} partitions, retention ${retention_ms}ms)"
  else
    # Fall back to rpk if Admin API fails
    if command -v docker &>/dev/null; then
      docker exec -i "$(docker ps -q -f name=redpanda)" \
        rpk topic create "$topic_name" \
          --partitions "$partitions" \
          --replicas "$replication" \
          -X brokers="localhost:9092" \
          2>/dev/null && log_pass "  ${topic_name} — created via rpk" \
          || log_warn "  ${topic_name} — create failed (may already exist)"
    fi
  fi
}

# ─────────────────────────────────────────────────────────────
# Topic definitions
# See: docs/sdlc_guidelines/08_tech_specific/redpanda_guidelines.md
# Retention values for dev are shorter than production.
# ─────────────────────────────────────────────────────────────
create_all_topics() {
  echo ""
  log_info "Creating Redpanda topics for local development..."
  echo ""

  # Sensor raw topics — 24h retention in dev (12 partitions in DC; 3 here)
  create_topic "sensor.raw.radar"          3 86400000  1
  create_topic "sensor.raw.ew_sigint"      3 86400000  1
  create_topic "sensor.raw.elint_comint"   3 86400000  1
  create_topic "sensor.raw.isr"            3 86400000  1
  create_topic "sensor.raw.ais_bft"        3 86400000  1
  create_topic "sensor.raw.cyber"          3 86400000  1

  # Dead-letter queues for each sensor raw topic
  create_topic "sensor.raw.radar.dlq"       1 604800000 1
  create_topic "sensor.raw.ew_sigint.dlq"   1 604800000 1
  create_topic "sensor.raw.elint_comint.dlq" 1 604800000 1
  create_topic "sensor.raw.isr.dlq"         1 604800000 1
  create_topic "sensor.raw.ais_bft.dlq"     1 604800000 1
  create_topic "sensor.raw.cyber.dlq"        1 604800000 1

  # Entity / track topics — 48h retention
  create_topic "entity.tracks.fused"          3 172800000 1

  # Inference / anomaly topics — 48h retention
  create_topic "inference.anomaly.scores"     2 172800000 1

  # Feedback topics — 7 day retention
  create_topic "feedback.operator.submissions" 1 604800000 1
  create_topic "feedback.operator.validated"   1 604800000 1

  # Audit events — 30 day retention in dev
  create_topic "audit.events"                 2 2592000000 1

  # NATO exchange topics — 24h retention
  create_topic "nato.exchange.inbound"        1 86400000  1
  create_topic "nato.exchange.outbound"       1 86400000  1

  echo ""
  log_pass "All topics created/verified."
}

# ─────────────────────────────────────────────────────────────
# List topics for verification
# ─────────────────────────────────────────────────────────────
list_topics() {
  echo ""
  log_info "Listing all topics in Redpanda:"

  if command -v docker &>/dev/null; then
    docker exec "$(docker ps -q -f name=redpanda)" \
      rpk topic list -X brokers="localhost:9092" 2>/dev/null || true
  fi
}

# ─────────────────────────────────────────────────────────────
# Main
# ─────────────────────────────────────────────────────────────
main() {
  echo ""
  echo -e "${CYAN}══ RTSA Redpanda Topic Initializer (UNCLASSIFIED) ══${NC}"

  wait_for_redpanda
  create_all_topics
  list_topics

  echo ""
  echo -e "${GREEN}Done.${NC} Redpanda topics are ready for development."
}

main "$@"
