# CLASSIFICATION: UNCLASSIFIED
#!/usr/bin/env bash
# scripts/demo/seed-demo-data.sh
# Seed ClickHouse with representative demo data for all RTSA use cases (UC001-UC017).
#
# Populates:
#   - Sensor observations: 11 records across 6 sensor types
#   - Fused tracks: 10 current tracks + 48h history for TRK-0002 (MMSI 123456789)
#   - Anomaly detections: 10 alerts (4 CRITICAL + 4 ELEVATED active, 2 acknowledged)
#   - Operator feedback: 5 records (includes 1 anti-poisoning test)
#   - Audit log: 12 entries (track/alert/feedback/NATO lifecycle events)
#   - NATO allied tracks: 5 inbound (REL TO FVEY classification)
#
# Usage: bash scripts/demo/seed-demo-data.sh [--dry-run]
#
# Prerequisite: bash scripts/dev/init-clickhouse.sh must have been run first.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
DRY_RUN="false"

CLICKHOUSE_HOST="${CLICKHOUSE_HOST:-localhost}"
CLICKHOUSE_HTTP_PORT="${CLICKHOUSE_HTTP_PORT:-8123}"
CLICKHOUSE_DATABASE="${CLICKHOUSE_DATABASE:-rtsa}"
CLICKHOUSE_USER="${CLICKHOUSE_USER:-default}"
CLICKHOUSE_PASSWORD="${CLICKHOUSE_PASSWORD:-dev_password_change_me}"

GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

while [ "$#" -gt 0 ]; do
  case "$1" in
    --dry-run) DRY_RUN="true" ;;
    -h|--help)
      echo "Usage: bash scripts/demo/seed-demo-data.sh [--dry-run]"
      echo "Seeds ClickHouse with representative RTSA demo data for all use cases."
      exit 0
      ;;
    *) echo "Unknown argument: $1"; exit 1 ;;
  esac
  shift
done

log_info() { echo -e "${CYAN}[i]${NC} $*"; }
log_pass() { echo -e "${GREEN}[v]${NC} $*"; }
log_fail() { echo -e "${RED}[x]${NC} $*"; exit 1; }
log_warn() { echo -e "${YELLOW}[!]${NC} $*"; }

ch_exec() {
  local sql="$1"
  local label="${2:-query}"
  if [ "$DRY_RUN" = "true" ]; then
    echo -e "${YELLOW}[dry-run]${NC} ClickHouse: $label"
    return 0
  fi
  local http_url="http://${CLICKHOUSE_HOST}:${CLICKHOUSE_HTTP_PORT}/?database=${CLICKHOUSE_DATABASE}"
  curl -sf "${http_url}" \
    -d "${sql}" \
    -u "${CLICKHOUSE_USER}:${CLICKHOUSE_PASSWORD}" >/dev/null 2>&1 || {
    log_fail "ClickHouse query failed: $label"
  }
}

wait_for_clickhouse() {
  log_info "Waiting for ClickHouse at ${CLICKHOUSE_HOST}:${CLICKHOUSE_HTTP_PORT} ..."
  local attempt=0
  local max_attempts=30
  while ! curl -sf "http://${CLICKHOUSE_HOST}:${CLICKHOUSE_HTTP_PORT}/ping" >/dev/null 2>&1; do
    attempt=$(( attempt + 1 ))
    if [ "$attempt" -ge "$max_attempts" ]; then
      log_fail "ClickHouse not ready. Is the stack running? Run: docker compose up -d"
    fi
    printf "."
    sleep 2
  done
  echo ""
  log_pass "ClickHouse is ready"
}

echo -e "${CYAN}============================================================${NC}"
echo -e "${CYAN}  RTSA Demo Data Seeder v2.0${NC}"
echo -e "${CYAN}  Populating representative data for UC001-UC017${NC}"
echo -e "${CYAN}============================================================${NC}"

wait_for_clickhouse

# ─────────────────────────────────────────────────────────────────────────────
# 1. Sensor Observations — 6 sensor types
# ─────────────────────────────────────────────────────────────────────────────
log_info "Seeding sensor observations (6 sensor types)..."

ch_exec "INSERT INTO sensor_observations
  (observation_id, sensor_id, sensor_type, event_time, track_id,
   latitude, longitude, altitude_m, speed_knots, heading_degrees,
   classification_level, confidence, metadata_json)
VALUES
  ('OBS-R-001','RADAR-NORTH-01','RADAR',now() - INTERVAL 48 HOUR,'TRK-0001',
   60.21,-9.73,0,12.0,180.0,1,0.88,
   '{\"snr_db\":24.5,\"range_nm\":34.2,\"rcs_m2\":450}'),
  ('OBS-R-002','RADAR-NORTH-01','RADAR',now() - INTERVAL 47 HOUR,'TRK-0001',
   60.09,-9.74,0,12.3,181.0,1,0.91,
   '{\"snr_db\":26.1,\"range_nm\":33.8,\"rcs_m2\":455}'),
  ('OBS-R-003','RADAR-SOUTH-02','RADAR',now() - INTERVAL 46 HOUR,'TRK-0001',
   59.98,-9.75,0,28.5,182.0,1,0.79,
   '{\"snr_db\":19.2,\"range_nm\":67.4,\"rcs_m2\":462}'),
  ('OBS-EW-001','EW-STATION-01','EW_SIGINT',now() - INTERVAL 48 HOUR,'TRK-0001',
   60.20,-9.72,100,0,0,1,0.84,
   '{\"frequency_ghz\":8.5,\"power_dbm\":42.3,\"modulation\":\"BURST\",\"emitter_id\":\"EW-UNK-001\"}'),
  ('OBS-EW-002','EW-STATION-03','EW_SIGINT',now() - INTERVAL 47 HOUR,'TRK-0001',
   60.10,-9.74,100,0,0,1,0.81,
   '{\"frequency_ghz\":8.5,\"power_dbm\":38.7,\"modulation\":\"BURST\",\"emitter_id\":\"EW-UNK-001\"}'),
  ('OBS-EL-001','ELINT-ARRAY-01','ELINT_COMINT',now() - INTERVAL 44 HOUR,'TRK-0005',
   57.50,-10.20,-50,0,0,1,0.72,
   '{\"emitter_id\":\"ELINT-SUB-001\",\"frequency_ghz\":0.15,\"intercept_type\":\"ACOUSTIC\",\"cep_m\":800}'),
  ('OBS-ISR-001','ISR-UAV-ALPHA','ISR',now() - INTERVAL 36 HOUR,'TRK-0006',
   56.80,-8.90,0,15.0,90.0,1,0.95,
   '{\"platform_id\":\"UAV-ALPHA\",\"resolution_m\":0.5,\"coverage_area_km2\":6.25}'),
  ('OBS-AIS-001','AIS-COAST-01','AIS_BFT',now() - INTERVAL 48 HOUR,'TRK-0002',
   60.50,-9.00,0,14.2,215.0,1,0.99,
   '{\"mmsi\":\"123456789\",\"vessel_name\":\"ALLIED-VESSEL-1\",\"vessel_type\":\"MILITARY\",\"nav_status\":\"UNDERWAY\"}'),
  ('OBS-AIS-002','AIS-COAST-01','AIS_BFT',now() - INTERVAL 47 HOUR,'TRK-0002',
   60.38,-9.10,0,14.0,216.0,1,0.99,
   '{\"mmsi\":\"123456789\",\"vessel_name\":\"ALLIED-VESSEL-1\",\"vessel_type\":\"MILITARY\",\"nav_status\":\"UNDERWAY\"}'),
  ('OBS-AIS-003','AIS-COAST-01','AIS_BFT',now() - INTERVAL 46 HOUR,'TRK-0001',
   61.50,-10.20,0,12.5,180.0,1,0.45,
   '{\"mmsi\":\"987654321\",\"vessel_name\":\"UNKNOWN-AIS\",\"vessel_type\":\"UNKNOWN\",\"spoof_flag\":true}'),
  ('OBS-CY-001','CYBER-FEED-01','CYBER',now() - INTERVAL 24 HOUR,'TRK-0007',
   0,0,0,0,0,1,0.88,
   '{\"ioc_type\":\"DOMAIN\",\"ioc_value\":\"malicious-c2.example.com\",\"stix_id\":\"indicator--abc123\",\"mitre_technique\":\"T1071.001\"}')" \
  "sensor observations"
log_pass "Sensor observations seeded (11 records)"

# ─────────────────────────────────────────────────────────────────────────────
# 2. Fused Tracks — 10 tracks across 5 domains
# ─────────────────────────────────────────────────────────────────────────────
log_info "Seeding fused tracks (10 tracks, 5 domains)..."

ch_exec "INSERT INTO tracks_fused
  (track_id, entity_type, hostile_class, track_status, event_time,
   latitude, longitude, altitude_m, speed_knots, heading_degrees,
   confidence_score, source_count, classification_level, label)
VALUES
  ('TRK-0001','SURFACE','SUSPECT','ACTIVE',now() - INTERVAL 46 HOUR,
   59.98,-9.75,0,28.5,182.0,0.82,4,1,'USV-SUSPECT-001'),
  ('TRK-0002','SURFACE','FRIENDLY','ACTIVE',now() - INTERVAL 47 HOUR,
   60.38,-9.10,0,14.0,216.0,0.95,2,1,'ALLIED-VESSEL-1'),
  ('TRK-0008','SURFACE','HOSTILE','ACTIVE',now() - INTERVAL 12 HOUR,
   58.50,-11.20,0,32.1,90.0,0.91,3,1,'FAST-MOVER-001'),
  ('TRK-0010','SURFACE','UNKNOWN','ACTIVE',now() - INTERVAL 6 HOUR,
   57.80,-10.50,0,8.5,270.0,0.83,2,1,'AIS-SPOOF-SUSPECT'),
  ('TRK-0003','AIR','UNKNOWN','ACTIVE',now() - INTERVAL 4 HOUR,
   59.20,-8.40,8000,420.0,45.0,0.67,2,1,'UAV-UNKNOWN-001'),
  ('TRK-0004','AIR','FRIENDLY','ACTIVE',now() - INTERVAL 2 HOUR,
   60.10,-7.80,12000,480.0,220.0,0.98,3,1,'VIPER-01'),
  ('TRK-0009','AIR','SUSPECT','ACTIVE',now() - INTERVAL 3 HOUR,
   58.90,-9.10,15000,650.0,180.0,0.76,2,1,'CRUISE-SUSPECT-001'),
  ('TRK-0005','SUBSURFACE','TENTATIVE','ACTIVE',now() - INTERVAL 8 HOUR,
   57.50,-10.20,-150,8.0,120.0,0.51,1,1,'ACOUSTIC-CONTACT-001'),
  ('TRK-0006','LAND','UNKNOWN','ACTIVE',now() - INTERVAL 10 HOUR,
   56.80,-8.90,150,15.0,90.0,0.73,1,1,'VEHICLE-CONVOY-001'),
  ('TRK-0007','CYBER','UNKNOWN','ACTIVE',now() - INTERVAL 24 HOUR,
   0,0,0,0,0,0.88,1,1,'IOC-C2-DOMAIN')" \
  "fused tracks"
log_pass "Fused tracks seeded (10 tracks)"

# ─────────────────────────────────────────────────────────────────────────────
# 3. 48-hour track history for MMSI 123456789 (TRK-0002)
# Hours 20-24 contain anomalous speed/pattern behaviour (forensics window)
# ─────────────────────────────────────────────────────────────────────────────
log_info "Seeding 48-hour track history for MMSI 123456789 (TRK-0002)..."

ch_exec "INSERT INTO tracks_fused
  (track_id, entity_type, hostile_class, track_status, event_time,
   latitude, longitude, altitude_m, speed_knots, heading_degrees,
   confidence_score, source_count, classification_level, label)
SELECT
  'TRK-0002' AS track_id,
  'SURFACE' AS entity_type,
  multiIf(h BETWEEN 20 AND 24, 'SUSPECT', 'FRIENDLY') AS hostile_class,
  'ACTIVE' AS track_status,
  now() - INTERVAL (96 - h * 2) HOUR AS event_time,
  60.50 + h * 0.02 * sin(h * 0.3) AS latitude,
  -9.00 + h * 0.03 * cos(h * 0.2) AS longitude,
  0 AS altitude_m,
  multiIf(h BETWEEN 20 AND 24, 28.5 + (h - 20) * 1.5, 14.2 + h * 0.1) AS speed_knots,
  215.0 + h * 2.0 AS heading_degrees,
  multiIf(h BETWEEN 20 AND 24, 0.62 - (h - 20) * 0.04, 0.95) AS confidence_score,
  2 AS source_count,
  1 AS classification_level,
  'ALLIED-VESSEL-1' AS label
FROM (SELECT arrayJoin(range(0, 48)) AS h)" \
  "48h track history"
log_pass "48-hour track history seeded (48 records for TRK-0002)"

# ─────────────────────────────────────────────────────────────────────────────
# 4. Anomaly Detection Records
# ─────────────────────────────────────────────────────────────────────────────
log_info "Seeding anomaly detection records (10 alerts)..."

ch_exec "INSERT INTO anomaly_detections
  (alert_id, track_id, anomaly_type, severity, confidence_score, explanation,
   event_time, alert_status, classification_level)
VALUES
  ('ALT-001','TRK-0001','SPEED_ANOMALY','CRITICAL',0.91,
   'USV accelerated from 12 to 28 knots over 90 seconds — 3.2 standard deviations above historical mean.',
   now() - INTERVAL 46 HOUR,'ACTIVE',1),
  ('ALT-002','TRK-0001','AIS_MANIPULATION','CRITICAL',0.87,
   'AIS position diverges 1.2 NM from co-located radar plot. MMSI 987654321 may be spoofing.',
   now() - INTERVAL 46 HOUR,'ACTIVE',1),
  ('ALT-003','TRK-0008','SPEED_ANOMALY','CRITICAL',0.94,
   'Surface contact accelerating at 3.8 m/s^2 — consistent with missile-boat attack profile.',
   now() - INTERVAL 12 HOUR,'ACTIVE',1),
  ('ALT-010','TRK-0010','AIS_MANIPULATION','CRITICAL',0.89,
   'AIS MMSI 987654321 reused at different location. Cross-referenced with RADAR-SOUTH-02 — spoofing probable.',
   now() - INTERVAL 6 HOUR,'ACTIVE',1),
  ('ALT-004','TRK-0003','BEHAVIORAL_PATTERN','ELEVATED',0.78,
   'UAV flight pattern matches ISR loitering profile. Operating outside designated exercise areas.',
   now() - INTERVAL 4 HOUR,'ACTIVE',1),
  ('ALT-005','TRK-0009','ROUTE_DEVIATION','ELEVATED',0.81,
   'Air track deviating 35 degrees from filed flight plan for 8 minutes — exceeds 30-degree threshold.',
   now() - INTERVAL 3 HOUR,'ACTIVE',1),
  ('ALT-006','TRK-0005','TEMPORAL_ANOMALY','ELEVATED',0.69,
   'Acoustic contact active outside normal patrol hours — inconsistent with allied submarine schedule.',
   now() - INTERVAL 8 HOUR,'ACTIVE',1),
  ('ALT-007','TRK-0006','ROUTE_DEVIATION','ELEVATED',0.74,
   'ISR vehicle convoy deviated from expected route. Now approaching restricted area perimeter.',
   now() - INTERVAL 10 HOUR,'ACTIVE',1),
  ('ALT-008','TRK-0002','SPEED_ANOMALY','ELEVATED',0.73,
   'Vessel exceeded normal transit speed during 4-hour anomalous activity window.',
   now() - INTERVAL 24 HOUR,'ACKNOWLEDGED',1),
  ('ALT-009','TRK-0002','BEHAVIORAL_PATTERN','ELEVATED',0.68,
   'Activity pattern deviation during anomalous window — correlated with ELINT intercept.',
   now() - INTERVAL 22 HOUR,'ACKNOWLEDGED',1)" \
  "anomaly detections"
log_pass "Anomaly detections seeded (4 CRITICAL + 4 ELEVATED active, 2 acknowledged)"

# ─────────────────────────────────────────────────────────────────────────────
# 5. Operator Feedback Records
# ─────────────────────────────────────────────────────────────────────────────
log_info "Seeding operator feedback records (5 records)..."

ch_exec "INSERT INTO operator_feedback
  (feedback_id, operator_id, track_id, alert_id, feedback_type, trust_score,
   justification, event_time, classification_level)
VALUES
  ('FB-001','OP-001','TRK-0002','ALT-008','CONFIRM_ANOMALY',0.85,
   'Vessel acceleration confirmed by independent ISR observation. Anomaly legitimate.',
   now() - INTERVAL 23 HOUR,1),
  ('FB-002','OP-001','TRK-0002','ALT-009','REJECT_ANOMALY',0.82,
   'Behavioral pattern explained by evasive action during active submarine contact nearby. False positive.',
   now() - INTERVAL 21 HOUR,1),
  ('FB-003','OP-002','TRK-0001','ALT-001','CONFIRM_ANOMALY',0.79,
   'Speed anomaly corroborated by EW intercept burst at same time. Strongly suspicious.',
   now() - INTERVAL 45 HOUR,1),
  ('FB-004','OP-002','TRK-0001','ALT-002','CONFIRM_ANOMALY',0.79,
   'AIS spoofing confirmed — radar and AIS positions irreconcilable at 1.2 NM separation.',
   now() - INTERVAL 45 HOUR,1),
  ('FB-005','OP-999','TRK-0003','ALT-004','REJECT_ANOMALY',0.22,
   'test_invalid_justification',  -- deliberately too short to trigger anti-poisoning guard
   now() - INTERVAL 2 HOUR,1)" \
  "operator feedback"
log_pass "Operator feedback seeded (5 records, 1 anti-poisoning test at trust=0.22)"

# ─────────────────────────────────────────────────────────────────────────────
# 6. Audit Log Records
# ─────────────────────────────────────────────────────────────────────────────
log_info "Seeding audit log records (12 entries)..."

ch_exec "INSERT INTO audit_log
  (audit_id, event_type, actor_id, resource_type, resource_id,
   action, event_time, classification_level, details_json)
VALUES
  ('AUD-001','TRACK_CREATED','svc-fusion-engine','track','TRK-0001',
   'CREATE',now() - INTERVAL 48 HOUR,1,
   '{\"source_count\":4,\"initial_confidence\":0.82,\"entity_type\":\"SURFACE\"}'),
  ('AUD-002','TRACK_CREATED','svc-fusion-engine','track','TRK-0002',
   'CREATE',now() - INTERVAL 48 HOUR,1,
   '{\"source_count\":2,\"initial_confidence\":0.95,\"mmsi\":\"123456789\"}'),
  ('AUD-003','TRACK_MERGED','svc-fusion-engine','track','TRK-0001',
   'MERGE',now() - INTERVAL 38 HOUR,1,
   '{\"merged_from\":\"TRK-GHOST-001\",\"new_confidence\":0.87}'),
  ('AUD-004','TRACK_STATUS_CHANGE','svc-track','track','TRK-GHOST-001',
   'STATUS_UPDATE',now() - INTERVAL 38 HOUR,1,
   '{\"previous_status\":\"ACTIVE\",\"new_status\":\"DROPPED\",\"reason\":\"merged\"}'),
  ('AUD-005','ALERT_CREATED','svc-anomaly-detection','alert','ALT-001',
   'CREATE',now() - INTERVAL 46 HOUR,1,
   '{\"severity\":\"CRITICAL\",\"anomaly_type\":\"SPEED_ANOMALY\",\"track_id\":\"TRK-0001\"}'),
  ('AUD-006','ALERT_ACKNOWLEDGED','OP-001','alert','ALT-008',
   'ACKNOWLEDGE',now() - INTERVAL 23 HOUR,1,
   '{\"operator_id\":\"OP-001\",\"track_id\":\"TRK-0002\"}'),
  ('AUD-007','FEEDBACK_SUBMITTED','OP-001','feedback','FB-001',
   'SUBMIT',now() - INTERVAL 23 HOUR,1,
   '{\"feedback_type\":\"CONFIRM_ANOMALY\",\"trust_score\":0.85,\"track_id\":\"TRK-0002\"}'),
  ('AUD-008','FEEDBACK_SUBMITTED','OP-001','feedback','FB-002',
   'SUBMIT',now() - INTERVAL 21 HOUR,1,
   '{\"feedback_type\":\"REJECT_ANOMALY\",\"trust_score\":0.82,\"track_id\":\"TRK-0002\"}'),
  ('AUD-009','ANTIPOISON_BLOCKED','svc-feedback','feedback','FB-005',
   'BLOCK',now() - INTERVAL 2 HOUR,1,
   '{\"operator_id\":\"OP-999\",\"reason\":\"justification_too_short\",\"trust_score\":0.22}'),
  ('AUD-010','ALERT_ASSIGNED','OP-001','alert','ALT-003',
   'ASSIGN',now() - INTERVAL 11 HOUR,1,
   '{\"assigner_operator_id\":\"OP-001\",\"assignee_operator_id\":\"OP-007\",\"comment\":\"Priority follow-up required\"}'),
  ('AUD-011','NATO_NOMINATED','OP-003','nato_nomination','TRK-0002',
   'NOMINATE',now() - INTERVAL 20 HOUR,1,
   '{\"track_id\":\"TRK-0002\",\"format\":\"LINK16\",\"classification_check\":\"PASS\"}'),
  ('AUD-012','NATO_BLOCKED','svc-nato-adapter','nato_nomination','TRK-0008',
   'BLOCK',now() - INTERVAL 19 HOUR,1,
   '{\"track_id\":\"TRK-0008\",\"reason\":\"classification_ceiling_exceeded\",\"max_release_level\":1}')" \
  "audit log"
log_pass "Audit log seeded (12 entries covering all lifecycle event types)"

# ─────────────────────────────────────────────────────────────────────────────
# 7. NATO Inbound Allied Tracks (UC015)
# ─────────────────────────────────────────────────────────────────────────────
log_info "Seeding NATO inbound allied tracks (5 tracks, REL TO FVEY)..."

ch_exec "INSERT INTO tracks_fused
  (track_id, entity_type, hostile_class, track_status, event_time,
   latitude, longitude, altitude_m, speed_knots, heading_degrees,
   confidence_score, source_count, classification_level, label)
VALUES
  ('TRK-NATO-001','SURFACE','FRIENDLY','ACTIVE',now() - INTERVAL 30 HOUR,
   61.20,-5.80,0,18.5,135.0,0.99,1,1,'NATO-ALLIED-FRIG-01'),
  ('TRK-NATO-002','SURFACE','FRIENDLY','ACTIVE',now() - INTERVAL 28 HOUR,
   60.80,-6.20,0,16.2,160.0,0.99,1,1,'NATO-ALLIED-DEST-01'),
  ('TRK-NATO-003','AIR','FRIENDLY','ACTIVE',now() - INTERVAL 25 HOUR,
   61.50,-4.50,8500,380.0,270.0,0.99,1,1,'NATO-ALLIED-P8-01'),
  ('TRK-NATO-004','AIR','FRIENDLY','ACTIVE',now() - INTERVAL 22 HOUR,
   60.90,-5.10,6000,350.0,290.0,0.99,1,1,'NATO-ALLIED-MPA-02'),
  ('TRK-NATO-005','SURFACE','FRIENDLY','ACTIVE',now() - INTERVAL 18 HOUR,
   59.50,-6.80,0,12.0,180.0,0.99,1,1,'NATO-ALLIED-SUP-01')" \
  "NATO allied tracks"
log_pass "NATO allied tracks seeded (3 surface + 2 air, REL TO FVEY classification)"

# ─────────────────────────────────────────────────────────────────────────────
# 8. Optimize v2.0 materialized views
# ─────────────────────────────────────────────────────────────────────────────
log_info "Optimizing v2.0 materialized views..."

for view in mv_active_tracks_by_domain mv_sensor_throughput_5min mv_alert_ack_latency; do
  ch_exec "OPTIMIZE TABLE ${view} FINAL" "optimize ${view}" 2>/dev/null || \
    log_warn "${view} not yet created — run bash scripts/dev/init-clickhouse.sh first"
done

log_pass "Materialized views optimized"

# ─────────────────────────────────────────────────────────────────────────────
# Summary
# ─────────────────────────────────────────────────────────────────────────────
echo ""
echo -e "${GREEN}============================================================${NC}"
echo -e "${GREEN}  RTSA Demo Data Seeding Complete (v2.0)${NC}"
echo -e "${GREEN}============================================================${NC}"
echo ""
echo "  Seeded records:"
echo "    Sensor observations : 11 (RADAR x3, EW_SIGINT x2, ELINT x1, ISR x1, AIS_BFT x3, CYBER x1)"
echo "    Fused tracks        : 10 current + 48 history records for TRK-0002"
echo "    Anomaly detections  : 10 (4 CRITICAL + 4 ELEVATED active, 2 acknowledged)"
echo "    Operator feedback   : 5 (4 valid, 1 anti-poisoning test at trust=0.22)"
echo "    Audit log entries   : 12 (track, alert, feedback, NATO lifecycle)"
echo "    NATO allied tracks  : 5 inbound (REL TO FVEY)"
echo ""
echo "  Key entities for demo:"
echo "    TRK-0001         : Surface Suspect — speed anomaly + AIS spoofing (2 CRITICAL alerts)"
echo "    TRK-0002         : Surface Friendly — MMSI 123456789, 48h forensics history"
echo "    TRK-0004         : Air Friendly — callsign VIPER-01 (Intel Search target)"
echo "    TRK-0008         : Surface Hostile — fast-mover CRITICAL alert"
echo "    TRK-NATO-001..5  : NATO inbound tracks (REL TO FVEY)"
echo ""
echo "  Dashboard: http://localhost:5173"
