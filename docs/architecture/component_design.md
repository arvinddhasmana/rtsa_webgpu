<!-- CLASSIFICATION: UNCLASSIFIED -->
# Component Design

> **Document**: RTSA Component Design
> **Version**: 2.0
> **Classification**: UNCLASSIFIED
> **Last Updated**: 2026-02-28
> **Compliance**: ITSG-33, NIST 800-53 Rev 5

---

## 1. Overview

This document defines the internal component design for each RTSA service. Components are organized by bounded context. Each service follows the standard Go project layout defined in the coding standards.

---

## 2. Standard Service Structure

All Go services follow this internal structure:

```
svc-<name>/
├── cmd/
│   └── <name>/main.go          # Entry point, wire dependencies
├── internal/
│   ├── config/                  # Configuration loading
│   ├── domain/                  # Domain types, business logic
│   ├── handler/                 # gRPC handler implementations
│   ├── producer/                # Redpanda producers
│   ├── consumer/                # Redpanda consumers
│   ├── repository/              # Data access (ClickHouse, etc.)
│   └── mapper/                  # Schema translation / mapping
├── proto/                       # Service-specific .proto files
├── deploy/
│   └── helm/                    # Helm chart
├── Dockerfile
├── Makefile
└── README.md
```

---

## 3. Sensor Ingestion Services

### 3.1 Component Diagram — Radar Ingestion Service

```mermaid
C4Component
    title Radar Ingestion Service — Component Diagram

    Container_Boundary(svc, "svc-radar-ingestion") {
        Component(handler, "RadarHandler", "gRPC Handler", "Receives radar track<br/>reports via gRPC")
        Component(validator, "TrackValidator", "Domain Logic", "Validates coordinate bounds,<br/>speed, heading, timestamps")
        Component(normalizer, "TrackNormalizer", "Mapper", "Converts radar-specific<br/>format to internal<br/>SensorObservation proto")
        Component(enricher, "MetadataEnricher", "Domain Logic", "Adds classification header,<br/>trace context,<br/>observation timestamp")
        Component(producer, "ObservationProducer", "Redpanda Producer", "Produces to<br/>sensors.radar.tracks")
        Component(dlq, "DLQProducer", "Redpanda Producer", "Routes invalid messages<br/>to sensors.radar.dlq")
        Component(health, "HealthServer", "gRPC Health", "Liveness + readiness<br/>probes")
        Component(metrics, "MetricsExporter", "OpenTelemetry", "Tracks per second,<br/>validation errors,<br/>producer latency")
    }

    Rel(handler, validator, "Validates")
    Rel(validator, normalizer, "Valid tracks")
    Rel(validator, dlq, "Invalid tracks")
    Rel(normalizer, enricher, "Normalized")
    Rel(enricher, producer, "Enriched observation")
```

> **Note**: EW/SIGINT, ELINT/COMINT, ISR, AIS/BFT, and Cyber ingestion services follow the same component pattern with sensor-specific validators and normalizers. The key differences are:

| Service | Validator Specifics | Normalizer Specifics |
|---|---|---|
| EW/SIGINT | Frequency 0.5–40 GHz, power dBm | EmitterObservation with modulation, PRI |
| ELINT/COMINT | Emitter ID format, intercept completeness | InterceptObservation with content classification |
| ISR | Geo-rect validity, resolution bounds | ISRObservation with platform + sensor metadata |
| AIS/BFT | MMSI format (9 digits), AIS spoofing checks | MaritimeObservation with vessel details |
| Cyber | STIX 2.1 schema validation, IOC dedup | ThreatObservation with MITRE ATT&CK mapping |

---

## 4. Fusion Engine

### 4.1 Component Diagram

```mermaid
C4Component
    title Fusion Engine — Component Diagram

    Container_Boundary(svc, "svc-fusion-engine") {
        Component(consumer, "SensorConsumer", "Redpanda Consumer", "Consumes from all<br/>sensors.* topics<br/>(consumer group: fusion)")
        Component(gate, "GatingFilter", "Domain Logic", "Spatial/temporal gating:<br/>- Distance ≤ threshold<br/>- Time ≤ 30s<br/>- Type compatibility")
        Component(scorer, "CorrelationScorer", "Domain Logic", "Weighted scoring:<br/>- Position proximity (0.35)<br/>- Velocity similarity (0.25)<br/>- Type match (0.20)<br/>- Temporal proximity (0.20)")
        Component(assigner, "TrackAssigner", "Domain Logic", "Assigns observation to<br/>existing track or<br/>creates new track")
        Component(estimator, "StateEstimator", "Domain Logic", "Kalman filter for<br/>position/velocity<br/>state estimation")
        Component(merger, "TrackMerger", "Domain Logic", "Merges tracks when<br/>correlation exceeds<br/>threshold (≥ 0.85)")
        Component(producer, "FusedTrackProducer", "Redpanda Producer", "Produces to<br/>tracks.fused.{type}")
        Component(stale, "StaleTrackMonitor", "Background Worker", "Marks tracks STALE<br/>after 60s no update.<br/>Drops after 5 min.")
        Component(metrics, "FusionMetrics", "OpenTelemetry", "Correlation rate,<br/>track count,<br/>fusion latency")
    }

    Rel(consumer, gate, "Sensor observations")
    Rel(gate, scorer, "Gated candidates")
    Rel(scorer, assigner, "Scored pairs")
    Rel(assigner, estimator, "Track + observation")
    Rel(estimator, producer, "Updated fused track")
    Rel(assigner, merger, "High-correlation tracks")
    Rel(merger, producer, "Merged track")
    Rel(stale, producer, "Track status updates")
```

### 4.2 Key Algorithms

#### Gating Criteria

| Criterion | Surface Track | Air Track | Subsurface |
|---|---|---|---|
| Max distance | 5 NM | 20 NM | 2 NM |
| Max time delta | 30 s | 15 s | 60 s |
| Type compatibility | Surface only | Air only | Sub only |

#### Correlation Score Thresholds

| Score Range | Action |
|---|---|
| ≥ 0.85 | Auto-correlate (merge into existing track) |
| 0.60 – 0.84 | Tentative correlation (flag for review) |
| < 0.60 | No correlation (create new track or discard) |

---

## 5. Anomaly Detection Service

### 5.1 Component Diagram

```mermaid
C4Component
    title Anomaly Detection Service — Component Diagram

    Container_Boundary(svc, "svc-anomaly-detection") {
        Component(consumer, "TrackConsumer", "Redpanda Consumer", "Consumes from<br/>tracks.fused.* topics")
        Component(preprocessor, "FeatureExtractor", "Domain Logic", "Extracts features:<br/>- Speed delta<br/>- Heading change rate<br/>- Pattern deviation<br/>- Spatial anomaly score")
        Component(inference, "InferenceEngine", "AI/ML Runtime", "Runs inference on<br/>pre-trained model.<br/>Outputs anomaly scores<br/>per anomaly type.")
        Component(thresholder, "ThresholdEvaluator", "Domain Logic", "Maps scores to<br/>severity levels:<br/>NORMAL/WATCH/<br/>ELEVATED/CRITICAL")
        Component(explainer, "ExplanationGenerator", "Domain Logic", "Produces human-readable<br/>explanation of<br/>why alert was raised")
        Component(producer, "AlertProducer", "Redpanda Producer", "Produces to<br/>alerts.anomaly.{severity}")
        Component(modelloader, "ModelLoader", "Background Worker", "Watches models.*<br/>topic for updates.<br/>Hot-swaps models.")
        Component(degradation, "DegradationMonitor", "Background Worker", "Monitors false positive<br/>rate vs. baseline.<br/>Triggers rollback.")
        Component(metrics, "InferenceMetrics", "OpenTelemetry", "Inference latency,<br/>alert counts by severity,<br/>model version")
    }

    Rel(consumer, preprocessor, "Fused tracks")
    Rel(preprocessor, inference, "Feature vectors")
    Rel(inference, thresholder, "Anomaly scores")
    Rel(thresholder, explainer, "Elevated+ alerts")
    Rel(explainer, producer, "Alert with explanation")
    Rel(modelloader, inference, "Updated model")
    Rel(degradation, modelloader, "Rollback trigger")
```

### 5.2 Anomaly Types

| Type | Features Used | Threshold (ELEVATED) |
|---|---|---|
| Speed anomaly | Speed delta, historical avg | > 3σ from mean |
| Route deviation | Heading change rate, expected route | > 30° sustained deviation |
| AIS manipulation | AIS vs. radar position delta | > 0.5 NM discrepancy |
| Behavioral pattern | Activity sequence encoding | Model confidence > 0.75 |
| Temporal anomaly | Time-of-day activity pattern | Outside normal pattern, p < 0.05 |
| Proximity alert | Distance to critical assets | < defined exclusion zone |

---

## 6. Feedback Service

### 6.1 Component Diagram

```mermaid
C4Component
    title Feedback Service — Component Diagram

    Container_Boundary(svc, "svc-feedback") {
        Component(handler, "FeedbackHandler", "gRPC Handler", "Receives operator<br/>feedback submissions<br/>via gRPC")
        Component(validator, "FeedbackValidator", "Domain Logic", "Validates feedback type,<br/>track ID exists,<br/>required fields present")
        Component(trust, "TrustScorer", "Domain Logic", "Calculates trust score:<br/>- Clearance (0.2)<br/>- Accuracy (0.3)<br/>- Temporal (0.2)<br/>- Deviation (0.3)")
        Component(antipoison, "AntiPoisonGuard", "Domain Logic", "Detects bulk anomalous<br/>feedback patterns.<br/>Rate limiting per operator.")
        Component(producer_raw, "RawFeedbackProducer", "Redpanda Producer", "Produces to<br/>feedback.operator.submissions")
        Component(producer_val, "ValidatedFeedbackProducer", "Redpanda Producer", "Produces to<br/>feedback.operator.validated<br/>(trust ≥ 0.5 only)")
        Component(audit, "AuditEmitter", "Redpanda Producer", "Produces audit events<br/>for all feedback actions")
        Component(metrics, "FeedbackMetrics", "OpenTelemetry", "Feedback rate,<br/>trust distribution,<br/>rejection rate")
    }

    Rel(handler, validator, "Raw feedback")
    Rel(validator, trust, "Valid feedback")
    Rel(trust, antipoison, "Scored feedback")
    Rel(antipoison, producer_raw, "All feedback")
    Rel(antipoison, producer_val, "Trusted feedback (≥ 0.5)")
    Rel(antipoison, audit, "All actions")
```

---

## 7. NATO Adapter Service

### 7.1 Component Diagram

```mermaid
C4Component
    title NATO Adapter Service — Component Diagram

    Container_Boundary(svc, "svc-nato-adapter") {
        Component(l16_rx, "Link16Receiver", "Protocol Handler", "Receives J-Series<br/>messages from<br/>Link 16 terminal")
        Component(nffi_rx, "NFFIReceiver", "HTTP/XML Handler", "Receives NFFI XML<br/>from MIP gateway")
        Component(cdg_in, "InboundGuard", "Cross-Domain Guard", "Classification ceiling<br/>check, content inspection<br/>for inbound data")
        Component(xlat_in, "InboundTranslator", "Mapper", "J-Series/NFFI →<br/>Internal proto format")

        Component(consumer, "TrackConsumer", "Redpanda Consumer", "Consumes from<br/>tracks.fused.* for<br/>outbound export")
        Component(cdg_out, "OutboundGuard", "Cross-Domain Guard", "Release policy check,<br/>source sanitization,<br/>REL TO marking")
        Component(xlat_out, "OutboundTranslator", "Mapper", "Internal proto →<br/>J-Series/NFFI format")
        Component(l16_tx, "Link16Transmitter", "Protocol Handler", "Sends J-Series<br/>messages to<br/>Link 16 terminal")
        Component(nffi_tx, "NFFITransmitter", "HTTP/XML Handler", "Sends NFFI XML<br/>to MIP gateway")

        Component(producer, "NATOProducer", "Redpanda Producer", "Produces inbound to<br/>sensors.nato.*")
        Component(audit, "AuditEmitter", "Redpanda Producer", "All NATO exchange<br/>events audited")
    }

    Rel(l16_rx, cdg_in, "J-Series")
    Rel(nffi_rx, cdg_in, "NFFI XML")
    Rel(cdg_in, xlat_in, "Cleared data")
    Rel(xlat_in, producer, "Internal proto")

    Rel(consumer, cdg_out, "Fused tracks")
    Rel(cdg_out, xlat_out, "Sanitized tracks")
    Rel(xlat_out, l16_tx, "J-Series messages")
    Rel(xlat_out, nffi_tx, "NFFI XML")
```

---

## 8. Presentation Services

### 8.1 Track Service *(v2.0: dual consumer group)*

```mermaid
C4Component
    title Track Service — Component Diagram (v2.0)

    Container_Boundary(svc, "svc-track") {
        Component(consumer, "FusedTrackConsumer", "Redpanda Consumer", "Consumer group: track-service<br/>Consumes tracks.fused.* topics")
        Component(sensor_consumer, "SensorObsConsumer", "Redpanda Consumer", "Consumer group: track-svc-sensor-stream<br/>Consumes sensors.* topics (v2.0)")
        Component(cache, "TrackStateCache", "In-Memory Store", "Current state of all<br/>active tracks.<br/>Indexed by track_id.")
        Component(fanout, "SensorObsFanout", "In-Memory Fan-Out", "Distributes raw observations<br/>to all registered<br/>stream handlers (v2.0)")
        Component(stream_handler, "StreamTracksHandler", "gRPC Server-Stream", "Streams track updates<br/>to connected clients<br/>with filtering")
        Component(sensor_stream_handler, "StreamSensorObsHandler", "gRPC Server-Stream", "Streams raw sensor observations<br/>with classification + bbox + type<br/>filtering (v2.0)")
        Component(detail_handler, "GetTrackDetailsHandler", "gRPC Unary", "Returns full track<br/>details with source<br/>attribution and history")
        Component(filter, "TrackFilterEngine", "Domain Logic", "Applies classification,<br/>type, and spatial<br/>filters per client")
        Component(metrics, "TrackMetrics", "OpenTelemetry", "Active track count,<br/>stream client count,<br/>update latency")
    }

    Rel(consumer, cache, "Track updates")
    Rel(sensor_consumer, fanout, "Raw sensor observations")
    Rel(cache, filter, "Active tracks")
    Rel(filter, stream_handler, "Filtered updates")
    Rel(cache, detail_handler, "Track details")
    Rel(fanout, sensor_stream_handler, "Observed sensor data")
```

**New RPC — `StreamSensorObservations` *(v2.0)***: Server-streaming RPC consuming from `sensors.*` Redpanda topics via a dedicated consumer group. Classification filtering, bounding box filtering, and sensor type filtering are applied per connected client before forwarding each `SensorObservationUpdate`.

### 8.2 Alert Service *(v2.0: AssignAlert)*

```mermaid
C4Component
    title Alert Service — Component Diagram (v2.0)

    Container_Boundary(svc, "svc-alert") {
        Component(consumer, "AlertConsumer", "Redpanda Consumer", "Consumes from<br/>alerts.anomaly.* topics")
        Component(queue, "AlertQueue", "In-Memory Priority Queue", "Priority queue sorted<br/>by severity then time.<br/>CRITICAL first.")
        Component(stream_handler, "StreamAlertsHandler", "gRPC Server-Stream", "Streams alerts to<br/>connected UI clients")
        Component(ack_handler, "AcknowledgeAlertHandler", "gRPC Unary", "Records operator<br/>acknowledgment of alert")
        Component(assign_handler, "AssignAlertHandler", "gRPC Unary", "Assigns alert to another<br/>operator; produces audit event (v2.0)")
        Component(audit_emitter, "AuditEmitter", "Redpanda Producer", "Emits audit events for<br/>acknowledgment and assignment")
        Component(metrics, "AlertMetrics", "OpenTelemetry", "Alert rate by severity,<br/>unacknowledged count,<br/>time-to-acknowledge")
    }

    Rel(consumer, queue, "New alerts")
    Rel(queue, stream_handler, "Prioritized alerts")
    Rel(ack_handler, queue, "Remove acknowledged")
    Rel(ack_handler, audit_emitter, "alert_acknowledged event")
    Rel(assign_handler, queue, "Set assigned_to field")
    Rel(assign_handler, audit_emitter, "alert_assigned event")
```

**New RPC — `AssignAlert` *(v2.0)***: Unary RPC with fields `alert_id`, `assigner_operator_id`, `assignee_operator_id`, `comment`. Sets an `assigned_to` field on the in-memory alert and produces an audit event to `audit.events`. Returns `success` and `assigned_at` timestamp.

### 8.3 Query Service *(v2.0: GetEventTimeline)*

```mermaid
C4Component
    title Query Service — Component Diagram (v2.0)

    Container_Boundary(svc, "svc-query") {
        Component(handler, "QueryHandler", "gRPC Handler", "Receives historical<br/>query requests")
        Component(timeline_handler, "EventTimelineHandler", "gRPC Unary", "Unified timeline for a track_id<br/>across 4 ClickHouse tables (v2.0)")
        Component(authz, "ClassificationFilter", "Security", "Injects classification<br/>filter based on<br/>caller's clearance")
        Component(guardrail, "QueryGuardrail", "Domain Logic", "Enforces time range limits,<br/>result limits,<br/>query timeout")
        Component(executor, "QueryExecutor", "ClickHouse Client", "Executes parameterized<br/>SQL against ClickHouse")
        Component(paginator, "ResultPaginator", "Domain Logic", "Paginates large result<br/>sets for streaming<br/>gRPC response")
        Component(audit, "QueryAuditEmitter", "Redpanda Producer", "Logs query execution<br/>to audit trail")
    }

    Rel(handler, authz, "Query request")
    Rel(timeline_handler, authz, "Timeline request")
    Rel(authz, guardrail, "Filtered query")
    Rel(guardrail, executor, "Guarded query")
    Rel(executor, paginator, "Result rows")
    Rel(paginator, handler, "Paginated response")
    Rel(executor, audit, "Query audit event")
```

**New RPC — `GetEventTimeline` *(v2.0)***: Unary RPC accepting `track_id`, `time_range`, `max_events`, and `clearance_level`. Executes a ClickHouse `UNION ALL` query across `tracks_fused`, `anomaly_detections`, `operator_feedback`, and `audit_log`, ordered by `event_time ASC`. Returns `repeated TimelineEvent` with a `oneof` payload discriminated by `TimelineEventType`.

---

## 9. COP Web Application (React) *(v2.0: Two-Level RBAC Shell)*

### 9.1 Architecture Overview

v2.0 introduces a **Two-Level RBAC Shell**:
- **Level 1** — Role Selector: Operations Commander, Intelligence Analyst, Security Officer, Sensor Operator, NATO Liaison
- **Level 2** — Dashboard Selector: role-appropriate views rendered dynamically by `MainLayout`

```mermaid
flowchart TD
    APP[App] --> AUTH[AuthProvider]
    AUTH --> LAYOUT[MainLayout]
    LAYOUT --> TOOLBAR[Toolbar: RoleSelector + DashboardSelector]
    LAYOUT --> ROUTER{activeDashboardView}
    ROUTER -->|fusion| FD[FusionDashboard]
    ROUTER -->|multi-domain| MD[MultiDomainDashboard]
    ROUTER -->|operator| OD[OperatorDashboard]
    ROUTER -->|forensics| FP[ForensicsPanel]
    ROUTER -->|audit| AP[AlertPanel + AuditQueue]
    ROUTER -->|sensor-health| SH[SensorHealthDashboard]
    ROUTER -->|nato-exchange| NE[NATOExchangeDashboard]

    FD --> MAP1[MapView: FusedTrack + SensorObs icons]
    FD --> FSP[FusionSidePanel]
    MD --> MAP2[MapView + SensorCoverageLayer]
    MD --> DMO[DomainMetricsOverlay]
    OD --> MAP3[MapView with backdrop-blur]
    OD --> TL[TimelineView]
    OD --> ALP[AlertPanel with quick-actions]
```

### 9.2 Component Hierarchy *(v2.0)*

```mermaid
flowchart TD
    APP[App] --> AUTH[AuthProvider]
    AUTH --> LAYOUT[MainLayout]
    LAYOUT --> MAP[MapView]
    LAYOUT --> ALERTS[AlertPanel]
    LAYOUT --> DETAIL[DetailPanel]
    LAYOUT --> FORENSICS[ForensicsPanel]
    LAYOUT --> FUSION_SP[FusionSidePanel]
    LAYOUT --> TIMELINE[TimelineView]
    LAYOUT --> DOMAIN_O[DomainMetricsOverlay]

    MAP --> TRACKS[TrackLayer]
    MAP --> SENSOR_ICONS[SensorObsLayer]
    MAP --> HALOS[ThreatHaloLayer]
    MAP --> GEO[GeoFenceLayer]
    MAP --> COVERAGE[SensorCoverageLayer]
    MAP --> LAYER_TOGGLE[MapLayerToggle]

    ALERTS --> ALERTCARD[AlertCard: Inspect / Confirm / Reject / Assign]
    ALERTS --> ALERTFILTER[AlertFilter]

    DETAIL --> IDENTITY[IdentitySection]
    DETAIL --> POSITION[PositionSection]
    DETAIL --> SOURCES[SourceAttribution]
    DETAIL --> ENTITY_TL[EntityTimeline]
    DETAIL --> FEEDBACKFORM[FeedbackForm]

    FORENSICS --> QUERYBUILDER[QueryBuilder]
    FORENSICS --> RESULTS[ResultsView]
    FORENSICS --> REPLAY[MapReplay]
```

### 9.3 State Management *(v2.0)*

| Store | Library | Content |
|---|---|---|
| TrackStore | Zustand | Active tracks indexed by `track_id` |
| AlertStore | Zustand | Alert queue + acknowledgment state |
| AuthStore | Zustand | Operator identity, clearance level |
| UIStore | Zustand | Panel visibility, selected entity, filters, **`activeRole`**, **`activeDashboardView`** *(v2.0)* |

**New `uiStore` fields** *(v2.0)*:
- `activeDashboardView: DashboardView` — current Level-2 view selection
- `setDashboardView(view: DashboardView): void` — Level-2 view switch action
- `ActiveRole` extended to include `sensor_operator` and `nato_liaison`
- `setActiveRole` auto-resets `activeDashboardView` to the role's default

### 9.4 Real-Time Hooks *(v2.0)*

| Hook | Transport | Purpose |
|---|---|---|
| `useTrackStream` | gRPC-Web server-stream | Subscribe to filtered fused track updates |
| `useAlertStream` | gRPC-Web server-stream | Subscribe to severity-filtered alerts |
| `useSensorStream` | gRPC-Web server-stream | Subscribe to raw sensor observations *(v2.0 — Fusion Dashboard)* |
| `useConnectionStatus` | WebSocket heartbeat | Monitor backend connectivity |
| `useOfflineMode` | Service Worker | Cache tracks and queue feedback when offline |

### 9.5 Role-to-Dashboard Mapping *(v2.0)*

| Role | Default Dashboard | Available Dashboards |
|---|---|---|
| Operations Commander | Fusion | Fusion, Multi-Domain, Operator UI |
| Intelligence Analyst | Forensics | Forensics, Intelligence Search |
| Security Officer | Audit & Feedback | Audit & Feedback |
| Sensor Operator | Sensor Health | Sensor Health |
| NATO Liaison | NATO Exchange | NATO Exchange |

### 9.6 Design System *(v2.0)*

| Category | Specification |
|---|---|
| Typography | Inter (Google Fonts) — weights 300/400/500/600/700 |
| Base background | `#0B1120` (near-black) |
| Panel background | `rgba(15, 23, 42, 0.60)` — glassmorphism with `backdrop-filter: blur(12px)` |
| Primary accent | `#3B82F6` (blue) |
| Warning accent | `#F59E0B` (amber) |
| Critical accent | `#EF4444` (red) |
| NVG theme | Green-on-black palette via `[data-theme="nvg"]` CSS override |
| Panel collapse | CSS transitions on width/height with `.panel-collapsible.collapsed` class |

| Library | Package Path | Purpose |
|---|---|---|
| `pkg/interceptors` | `github.com/rtsa/pkg/interceptors` | gRPC interceptor chain (auth, logging, metrics, tracing, classification) |
| `pkg/redpanda` | `github.com/rtsa/pkg/redpanda` | franz-go wrapper with standard producer/consumer config |
| `pkg/health` | `github.com/rtsa/pkg/health` | Standard health check server |
| `pkg/shutdown` | `github.com/rtsa/pkg/shutdown` | Graceful shutdown orchestrator |
| `pkg/classification` | `github.com/rtsa/pkg/classification` | Classification level types and enforcement |
| `pkg/audit` | `github.com/rtsa/pkg/audit` | Audit event builder and producer |
| `pkg/telemetry` | `github.com/rtsa/pkg/telemetry` | OpenTelemetry initialization |
| `pkg/config` | `github.com/rtsa/pkg/config` | Environment-based configuration loader |
