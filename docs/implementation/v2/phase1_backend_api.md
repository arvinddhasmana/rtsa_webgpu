<!-- CLASSIFICATION: UNCLASSIFIED -->
# Phase 1 — Backend API Enhancements

> **Phase**: 1 of 4 | **Depends on**: Nothing | **Blocks**: Phase 3
> **Scope**: Protobuf definitions, Go service modifications, ClickHouse schema additions

---

## Purpose

The current backend only streams `FusedTrack` and `AnomalyAlert` data to the UI. The premium dashboards require three additional data streams that do not exist yet:
1. Raw sensor observations (for Fusion Dashboard)
2. A unified event timeline (for Operator UI)
3. Sensor coverage geometry (for Multi-Domain Dashboard)

This phase adds the necessary gRPC APIs, ClickHouse views, and service logic to expose that data.

---

## Step 1: Add Raw Sensor Observation Streaming RPC

### 1.1 Define the new message and RPC in Protobuf

Open `proto/rtsa/entity/v1/track_service.proto`. Add the following new RPC to the existing `TrackService`:

```protobuf
// Server-streaming: receive raw sensor observations for correlation display
// Allows the UI to render individual sensor tracks alongside fused tracks
// Deadline: none (long-lived stream)
rpc StreamSensorObservations(StreamSensorObservationsRequest) returns (stream SensorObservationUpdate);
```

Then define the request and response messages in the same file:

```protobuf
message StreamSensorObservationsRequest {
  // Filter by sensor types (empty = all)
  repeated rtsa.common.v1.SensorType sensor_types = 1;
  // Filter by geographic bounding box (empty = global)
  optional rtsa.common.v1.BoundingBox bounding_box = 2;
  // Caller's classification clearance level
  rtsa.common.v1.ClassificationLevel clearance_level = 3;
}

message SensorObservationUpdate {
  // The raw sensor observation
  rtsa.ingestion.v1.SensorObservation observation = 1;
  // Whether the observation has been correlated to a fused track
  optional string correlated_track_id = 2;
}
```

You will need to add an import at the top of the file:

```protobuf
import "rtsa/ingestion/v1/sensor_observation.proto";
```

### 1.2 Regenerate Go code from Protobuf

Run the buf generation command from the project root:

```bash
buf generate
```

This will regenerate the Go stubs in `gen/go/rtsa/entity/v1/`. Verify the new `StreamSensorObservations` method appears in the generated `track_service_grpc.pb.go`.

### 1.3 Implement the streaming handler in `svc-track`

Open the Track Service implementation directory at `svc-track/`. The service already consumes from `tracks.fused.*` Redpanda topics. You need to add a second consumer group that reads from the raw sensor topics.

1. Create a new file `svc-track/internal/consumer/sensor_consumer.go`. This consumer should:
   - Join a new consumer group named `track-svc-sensor-stream`
   - Subscribe to topics: `sensors.radar.tracks`, `sensors.ew.intercepts`, `sensors.elint.detections`, `sensors.isr.observations`, `sensors.ais.positions`, `sensors.cyber.iocs`
   - Deserialize each message as a `SensorObservation` protobuf
   - Push each observation into an in-memory fan-out channel

2. Create a new file `svc-track/internal/handler/stream_observations.go`. This handler should:
   - Implement the `StreamSensorObservations` RPC
   - Read filtered observations from the fan-out channel
   - Apply classification filtering: only send observations where `observation.classification <= request.clearance_level`
   - If a bounding box is provided, check `observation.position` falls within it
   - If sensor type filters are provided, check `observation.sensor_type` matches
   - Stream each matching observation to the gRPC client

3. Register the new handler in the existing `svc-track/cmd/track/main.go` gRPC server setup.

### 1.4 Generate TypeScript client stub for the UI

After buf generation, verify the TypeScript/gRPC-Web stubs are also updated in `gen/ts/` (or wherever the web-cop consumes generated types from). The web-cop should be able to import and call `streamSensorObservations()`.

---

## Step 2: Add Event Timeline Aggregation RPC

### 2.1 Define the new messages and RPC in Protobuf

Open `proto/rtsa/query/v1/query_service.proto`. Add the following RPC to the existing `QueryService`:

```protobuf
// Unary: get a unified chronological timeline for an entity
// Merges track state changes, anomaly alerts, and operator feedback
// Deadline: 30s
rpc GetEventTimeline(GetEventTimelineRequest) returns (EventTimelineResponse);
```

Define the request and response messages:

```protobuf
message GetEventTimelineRequest {
  // Track ID to get timeline for (required)
  string track_id = 1;
  // Time range (required)
  rtsa.common.v1.TimeRange time_range = 2;
  // Maximum events to return (default: 200)
  int32 max_events = 3;
  // Caller's classification clearance
  rtsa.common.v1.ClassificationLevel clearance_level = 4;
}

message EventTimelineResponse {
  string track_id = 1;
  repeated TimelineEvent events = 2;
}

message TimelineEvent {
  // Event timestamp
  google.protobuf.Timestamp event_time = 1;
  // Event category
  TimelineEventType event_type = 2;
  // Human-readable summary of the event
  string summary = 3;
  // Event-specific payload
  oneof detail {
    TrackStateChange track_change = 10;
    AnomalyEventDetail anomaly = 11;
    FeedbackEventDetail feedback = 12;
    AuditEventDetail audit = 13;
  }
}

enum TimelineEventType {
  TIMELINE_EVENT_TYPE_UNSPECIFIED = 0;
  TIMELINE_EVENT_TYPE_TRACK_CREATED = 1;
  TIMELINE_EVENT_TYPE_TRACK_UPDATED = 2;
  TIMELINE_EVENT_TYPE_TRACK_MERGED = 3;
  TIMELINE_EVENT_TYPE_TRACK_DROPPED = 4;
  TIMELINE_EVENT_TYPE_ANOMALY_DETECTED = 5;
  TIMELINE_EVENT_TYPE_ALERT_ACKNOWLEDGED = 6;
  TIMELINE_EVENT_TYPE_FEEDBACK_SUBMITTED = 7;
  TIMELINE_EVENT_TYPE_CLASSIFICATION_CHANGED = 8;
}

message TrackStateChange {
  string previous_status = 1;
  string new_status = 2;
  rtsa.common.v1.Position position = 3;
  double confidence_score = 4;
}

message AnomalyEventDetail {
  string alert_id = 1;
  rtsa.common.v1.AnomalyType anomaly_type = 2;
  rtsa.common.v1.AlertSeverity severity = 3;
  double confidence_score = 4;
  string explanation = 5;
}

message FeedbackEventDetail {
  string feedback_id = 1;
  rtsa.common.v1.FeedbackType feedback_type = 2;
  double trust_score = 3;
}

message AuditEventDetail {
  string audit_id = 1;
  string action = 2;
  string actor_id = 3;
}
```

### 2.2 Regenerate Go code

```bash
buf generate
```

### 2.3 Implement the timeline handler in `svc-query`

Open the Query Service at `svc-query/`. Create a new file `svc-query/internal/handler/timeline.go`. This handler should:

1. Accept a `GetEventTimelineRequest`.
2. Execute a ClickHouse `UNION ALL` query that pulls events from 4 tables:
   - `tracks_fused` — filter by `track_id`, project `event_time`, `track_status` changes
   - `anomaly_detections` — filter by `track_id`, project `alert_id`, `anomaly_type`, `severity`, `confidence_score`, `explanation`
   - `operator_feedback` — filter by `track_id`, project `feedback_id`, `feedback_type`, `trust_score`
   - `audit_log` — filter by `resource_id = track_id` and `resource_type = 'track'`
3. Apply the classification filter: `WHERE classification_level <= $clearance_level`
4. Order the combined result by `event_time ASC`
5. Limit to `max_events` (default 200)
6. Map each row to a `TimelineEvent` proto message and return in the response.

Register the handler in `svc-query/cmd/query/main.go`.

---

## Step 3: Add Sensor Coverage Geometry

### 3.1 Define new message in Protobuf

Open `proto/rtsa/ingestion/v1/ingestion_service.proto`. Add a new message:

```protobuf
message SensorCoverage {
  // Coverage polygon vertices (for ISR, geo-fence type sensors)
  repeated rtsa.common.v1.Position coverage_polygon = 1;
  // Maximum sensor range in nautical miles (for radar, EW)
  optional double range_nm = 2;
  // Bearing sector start in degrees true (for directional sensors)
  optional double bearing_start_degrees = 3;
  // Bearing sector end in degrees true (for directional sensors)
  optional double bearing_end_degrees = 4;
  // Sensor geographic position
  optional rtsa.common.v1.Position sensor_position = 5;
}
```

Add the `coverage` field to the existing `SensorStatusResponse`:

```protobuf
// Add as field 9 to SensorStatusResponse:
optional SensorCoverage coverage = 9;
```

### 3.2 Add bulk sensor status RPC

In the same file, add a new RPC to the `IngestionService`:

```protobuf
// Unary: get status of all known sensors
// Deadline: 10s
rpc ListSensorStatuses(ListSensorStatusesRequest) returns (ListSensorStatusesResponse);
```

```protobuf
message ListSensorStatusesRequest {
  // Filter by sensor types (empty = all)
  repeated rtsa.common.v1.SensorType sensor_types = 1;
  // Only return sensors with observations in the last N seconds (0 = all)
  int32 active_within_seconds = 2;
}

message ListSensorStatusesResponse {
  repeated SensorStatusResponse sensors = 1;
}
```

### 3.3 Regenerate and implement

Run `buf generate`. Then implement the `ListSensorStatuses` handler in the relevant ingestion services. This handler should:
1. Query an in-memory sensor registry (populated from configuration or from the first observation received from each sensor)
2. For each registered sensor, return its current `SensorStatusResponse` including the `SensorCoverage` geometry
3. The coverage geometry for common sensor types should be populated from environment/config variables:
   - Radar: `range_nm`, `bearing_start_degrees`, `bearing_end_degrees`, `sensor_position`
   - EW/SIGINT: `range_nm`, `sensor_position` (omnidirectional)
   - ISR: `coverage_polygon` from the most recent `ISRObservation.coverage_polygon`

---

## Step 4: Add Alert Assignment RPC

### 4.1 Define new RPC in Protobuf

Open `proto/rtsa/inference/v1/alert_service.proto`. Add:

```protobuf
// Unary: assign an alert to another operator for follow-up
// Deadline: 5s
rpc AssignAlert(AssignAlertRequest) returns (AssignAlertResponse);
```

```protobuf
message AssignAlertRequest {
  string alert_id = 1;
  string assigner_operator_id = 2;
  string assignee_operator_id = 3;
  string comment = 4;
}

message AssignAlertResponse {
  bool success = 1;
  google.protobuf.Timestamp assigned_at = 2;
}
```

### 4.2 Implement in `svc-alert`

Create `svc-alert/internal/handler/assign.go`. The handler should:
1. Validate the alert exists in the in-memory priority queue
2. Set an `assigned_to` field on the alert
3. Produce an audit event to `audit.events` via the existing audit emitter pattern
4. Return success with timestamp

---

## Step 5: Add ClickHouse Materialized Views

### 5.1 Real-time track count by domain

Create a SQL migration file or add to your ClickHouse init scripts. Execute the following SQL against ClickHouse:

```sql
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_active_tracks_by_domain
ENGINE = AggregatingMergeTree()
ORDER BY (entity_type, ten_second_bucket)
AS SELECT
    entity_type,
    toStartOfInterval(event_time, INTERVAL 10 SECOND) AS ten_second_bucket,
    uniqExactState(track_id) AS unique_tracks,
    countState() AS observation_count
FROM tracks_fused
WHERE track_status IN ('ACTIVE', 'NEW')
GROUP BY entity_type, ten_second_bucket;
```

### 5.2 Sensor throughput (5-minute rolling)

```sql
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_sensor_throughput_5min
ENGINE = AggregatingMergeTree()
ORDER BY (sensor_type, sensor_id, five_min_bucket)
AS SELECT
    sensor_type,
    sensor_id,
    toStartOfFiveMinutes(event_time) AS five_min_bucket,
    countState() AS observation_count
FROM sensor_observations
GROUP BY sensor_type, sensor_id, five_min_bucket;
```

### 5.3 Alert acknowledgement latency

```sql
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_alert_ack_latency
ENGINE = AggregatingMergeTree()
ORDER BY (severity, hour)
AS SELECT
    severity,
    toStartOfHour(event_time) AS hour,
    countState() AS alert_count,
    avgState(toUnixTimestamp(now()) - toUnixTimestamp(event_time)) AS avg_ack_delay_seconds
FROM anomaly_detections
WHERE alert_id IN (SELECT alert_id FROM audit_log WHERE event_type = 'alert_acknowledged')
GROUP BY severity, hour;
```

---

## Verification

### Automated

1. Regenerate protos: `buf generate` — must complete without errors
2. Build all modified services:
   ```bash
   cd svc-track && go build ./...
   cd svc-query && go build ./...
   cd svc-alert && go build ./...
   ```
3. Run existing unit tests:
   ```bash
   cd svc-track && go test ./...
   cd svc-query && go test ./...
   cd svc-alert && go test ./...
   ```
4. Write new unit tests for:
   - `stream_observations.go` — verify classification filtering, bounding box filtering, sensor type filtering
   - `timeline.go` — verify UNION ALL query produces correctly ordered events
   - `assign.go` — verify audit event is produced

### Integration

1. Start the full stack: `docker compose -f deploy/docker-compose.yml up -d`
2. Use `grpcurl` or a test script to:
   - Call `StreamSensorObservations` and verify raw sensor data arrives
   - Call `GetEventTimeline` for a known track and verify events from multiple tables appear in order
   - Call `ListSensorStatuses` and verify coverage geometry is present
   - Call `AssignAlert` and verify audit log entry is created
