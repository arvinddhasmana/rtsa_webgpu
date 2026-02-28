<!-- CLASSIFICATION: UNCLASSIFIED -->
# Backend Gaps and Improvement Suggestions — Lead Architect Review

**Status:** COMPLETED | **Date:** 2026-02-27 | **Reviewer:** Lead Architect

---

## Overview

This document analyzes the backend implementation (gRPC Protobuf schemas, Go microservices, Redpanda topic architecture, ClickHouse OLAP layer) against the data requirements of the premium UI dashboards (Operator UI, Fusion Dashboard, Multi-Domain Dashboard). Cross-referenced against `docs/architecture/high_level_architecture.md`, `docs/architecture/data_architecture.md`, and `docs/architecture/component_design.md`.

---

## 1. Critical Backend Gaps

### 1.1 No Raw Sensor Observation Stream to the UI

| Aspect | Detail |
|---|---|
| **Gap** | `TrackService.StreamTracks` streams only `FusedTrack` objects. The UI has no API to receive pre-fusion `SensorObservation` data (Radar plots, EW intercepts, ELINT detections, AIS positions). |
| **Architecture Root Cause** | Per the C4 Container diagram, the Presentation Layer only consumes from `tracks.fused.*` and `alerts.*` topics. The raw `sensors.*` topics are consumed exclusively by the Fusion Engine. This is by design (CQRS) — but the Fusion Dashboard requires a read-model of raw observations. |
| **Impact — Fusion Dashboard** | Cannot render distinct sensor-track icons (Radar ◇, EW △, SIGINT ◻) alongside the fused track (●) to show the correlation process. The `SourceAttribution` inside `FusedTrack` provides only `sensor_id` and `confidence` — not the raw track's independent position, kinematics, or signal metadata. |
| **Impact — Multi-Domain Dashboard** | Cannot display sensor-specific metrics (radar SNR, EW signal power dBm, ELINT CEP accuracy, AIS vessel details). The `metadata_json` field in the ClickHouse `sensor_observations` table contains this data, but no query API exposes it to the UI. |

> [!IMPORTANT]
> This is the single most impactful backend gap. Without raw sensor data reaching the UI, neither the Fusion Dashboard nor the Multi-Domain Dashboard can be built as described.

### 1.2 Missing Event Timeline Aggregation

| Aspect | Detail |
|---|---|
| **Gap** | No unified timeline API exists. `QueryService` offers `QueryTracks`, `QueryAnomalies`, and `QueryAuditLog` as separate RPCs with different response schemas. |
| **Impact — Operator UI** | The chronological event timeline with "correlation markers" requires interleaving track state changes (CREATED → ACTIVE → MERGED), anomaly alerts, and operator feedback into a single time-ordered sequence. |
| **Impact — Entity Detail Panel** | The `EntityTimeline` sub-component (defined in `component_design.md` §9.1) requires the same unified projection per `track_id`. |
| **ClickHouse Readiness** | The data exists across `tracks_fused`, `anomaly_detections`, `operator_feedback`, and `audit_log` tables. A ClickHouse `UNION ALL` query ordered by `event_time` can produce this, but no gRPC RPC wraps it. |

### 1.3 Sensor Health & Coverage Geometry Missing from API

| Aspect | Detail |
|---|---|
| **Gap** | `IngestionService.GetSensorStatus` returns per-sensor statistics (`total_received`, `events_per_second`, `connected`), but does **not** include geographic coverage geometry (radar fan sector, EW listening arc, ISR swath polygon). |
| **Impact — Multi-Domain Dashboard** | Cannot render dynamic sensor coverage overlays on the map. The component design hierarchy defines `SensorCoverageLayer` as a child of `MapView`, but there is no data source to feed it. |
| **Impact — Sensor Operator View** | Sensor health monitoring requires positional context — a sensor's coverage area relative to active tracks. Currently sensor status is positionally blind. |

### 1.4 Alert Lifecycle Gaps

| Aspect | Detail |
|---|---|
| **Gap** | `AlertService.AcknowledgeAlert` is the only alert lifecycle RPC. The documented alert quick-actions (`[Inspect]`, `[Confirm]`, `[Reject]`, `[Assign]`) require additional operations. |
| **[Confirm] / [Reject]** | These map to `FeedbackService.SubmitFeedback` with `CONFIRM_ANOMALY` / `REJECT_ANOMALY` types. This path exists but is not wired in the UI. **Backend: OK, UI gap only.** |
| **[Assign]** | No backend concept for alert assignment to a specific operator. Requires a new `AssignAlert` RPC or an extension to `AcknowledgeAlert` with an `assignee_operator_id` field. |

### 1.5 Missing ClickHouse Materialized Views for Dashboard Metrics

| Metric | Needed By | Current State |
|---|---|---|
| Active tracks by domain (Surface/Air/Sub/Land/Cyber) in real-time | Fusion Dashboard, Multi-Domain Dashboard | `mv_track_count_by_type` exists but aggregates hourly, not real-time |
| Sensor observation rate by sensor_type (last 5 min) | Multi-Domain Dashboard | No materialized view |
| Alert-to-acknowledge latency distribution | Operator UI metrics | No materialized view |
| Sensor coverage freshness (last observation per sensor_id) | Multi-Domain Dashboard | `SensorStatusResponse.last_observation_time` exists per-sensor but not aggregated |

---

## 2. Improvements (Prioritised by Dashboard Dependency)

### P0 — Required for Fusion Dashboard & Multi-Domain Dashboard

#### 2.1 New RPC: `StreamSensorObservations`
Add to `TrackService` or create a dedicated `SensorStreamService`:

```protobuf
rpc StreamSensorObservations(StreamSensorObservationsRequest)
    returns (stream SensorObservationUpdate);
```

**Implementation**: A new consumer group in `svc-track` (or a new BFF service) that reads from `sensors.*` Redpanda topics, applies classification filtering, and streams to UI clients via gRPC-Web. This is consistent with AP-01 (Event-Driven First) — the service consumes from Redpanda, not from another service.

#### 2.2 New RPC: `GetEventTimeline`
Add to `QueryService`:

```protobuf
rpc GetEventTimeline(GetEventTimelineRequest)
    returns (EventTimelineResponse);
```

**Implementation**: A ClickHouse `UNION ALL` query across `tracks_fused`, `anomaly_detections`, `operator_feedback`, `audit_log` filtered by `track_id` and `time_range`, ordered by `event_time`. Returns a `repeated TimelineEvent` with a `oneof` payload.

#### 2.3 Extend `SensorStatusResponse` with Coverage Geometry
Add to `ingestion_service.proto`:

```protobuf
message SensorCoverage {
    repeated Position coverage_polygon = 1;  // For ISR, Geo-fences
    optional double range_nm = 2;            // For Radar
    optional double bearing_start_deg = 3;   // Sector start
    optional double bearing_end_deg = 4;     // Sector end
    optional Position sensor_position = 5;   // Sensor location
}
```

Add a `SensorCoverage coverage` field to `SensorStatusResponse`. Populate from sensor registry configuration or from ingestion metadata.

### P1 — Required for Operator UI

#### 2.4 New RPC: `AssignAlert`
Add to `AlertService`:

```protobuf
rpc AssignAlert(AssignAlertRequest) returns (AssignAlertResponse);
```

With fields: `alert_id`, `assignee_operator_id`, `assigner_operator_id`, `comment`. Produces audit event.

#### 2.5 Real-Time Materialized Views
Add ClickHouse materialized views for dashboard KPIs:
- `mv_active_tracks_by_domain` — `AggregatingMergeTree` with 10-second granularity
- `mv_sensor_throughput_5min` — rolling 5-minute observation rate by `sensor_type`
- `mv_alert_ack_latency` — time-to-acknowledge distribution by severity

### P2 — Desirable Enhancements

#### 2.6 Bulk Sensor Status RPC
Add `ListSensorStatuses` to `IngestionService` — returns all active sensors with their status and coverage in a single call, avoiding N+1 queries from the UI.

#### 2.7 Extend `QueryService` for Sensor-Specific Queries
Add `QuerySensorObservations` RPC that exposes the `sensor_observations` ClickHouse table to the UI, enabling the Intelligence Analyst to search raw intercepts by frequency, emitter ID, or MMSI.
