#!/usr/bin/env bash
# CLASSIFICATION: UNCLASSIFIED
# RTSA Redpanda Topic Initializer
# Creates all required Redpanda topics for local development.
# Requires the Redpanda dev stack to be running: docker compose up -d

set -euo pipefail

REDPANDA_BROKER="${REDPANDA_BROKER:-localhost:19092}"
REDPANDA_ADMIN_URL="${REDPANDA_ADMIN_URL:-http://localhost:19644}"
TOPIC_MANIFEST="${TOPIC_MANIFEST:-$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)/deploy/charts/redpanda-dev/topics.json}"

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
    }" 2>&1)" || {
    if echo "$response" | grep -qi "already exists"; then
      log_info "  ${topic_name} — already exists (skipped)"
      return 0
    fi
    log_fail "  ${topic_name} — Admin API failed: ${response}"
  }

  if echo "$response" | grep -q "already exists" 2>/dev/null; then
    log_info "  ${topic_name} — already exists (skipped)"
  elif echo "$response" | grep -q '"name"' 2>/dev/null; then
    log_pass "  ${topic_name} — created (${partitions} partitions, retention ${retention_ms}ms)"
  else
    log_fail "  ${topic_name} — unexpected Admin API response: ${response}"
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

  jq -e '.classification == "UNCLASSIFIED" and (.topics | type == "array")' "$TOPIC_MANIFEST" >/dev/null \
    || log_fail "invalid topic manifest: ${TOPIC_MANIFEST}"
  while IFS=$'\t' read -r topic_name partitions retention_ms replication; do
    create_topic "$topic_name" "$partitions" "$retention_ms" "$replication"
  done < <(jq -r '.topics[] | [.name, .partitions, .retention_ms, .replicas] | @tsv' "$TOPIC_MANIFEST")

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
    local redpanda_container
    redpanda_container="$(docker ps -q -f name=redpanda | head -n 1)"
    [[ -n "$redpanda_container" ]] || log_fail "Redpanda container is not running"
    docker exec "$redpanda_container" \
      rpk topic list -X brokers="localhost:9092"
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
