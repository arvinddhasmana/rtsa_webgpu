<!-- CLASSIFICATION: UNCLASSIFIED -->

# RTSA Implementation Overview — Master Orchestration Document

> **Document**: RTSA Implementation Plan — Module Orchestration
> **Version**: 1.0
> **Classification**: UNCLASSIFIED
> **Last Updated**: 2026-02-23
> **Target Agent**: Greatest Ever Developer (`@greatest-ever-developer`)

---

## 1. Purpose

This document is the master orchestration guide for implementing the RTSA system. It decomposes the full system into **18 independently implementable modules**, each with its own detailed instruction file. Multiple AI Agents can execute modules in parallel using **interface contracts** and **mock dependencies**.

**Scope**: Core flow — Sensors → Ingestion → Fusion → Anomaly Detection → Track/Alert → UI, plus Feedback loop, Historical Queries, and Audit. **NATO Adapter (UC014/UC015) and Training Pipeline (UC011) are stubbed** with noop implementations.

**Target Environment**: Development (Docker Compose, single-node Redpanda, single-node ClickHouse).

---

## 2. Module Inventory

| Module | File                            | Description                                              | Phase |
| ------ | ------------------------------- | -------------------------------------------------------- | ----- |
| 00     | `00-implementation-overview.md` | This file — orchestration, contracts, traceability       | P0    |
| 01     | `01-dev-infrastructure.md`      | Docker Compose, Makefiles, project scaffolding           | P0    |
| 02     | `02-protobuf-schemas.md`        | All `.proto` files, buf config, code generation          | P0    |
| 03     | `03-shared-go-libraries.md`     | `pkg/*` shared libraries (Redpanda, health, etc.)        | P1    |
| 04     | `04-sensor-ingestion-radar.md`  | Reference ingestion service (Radar) — full detail        | P2    |
| 05     | `05-sensor-ingestion-batch.md`  | 5 remaining sensor services (EW, ELINT, ISR, AIS, Cyber) | P2    |
| 06     | `06-wasm-data-transforms.md`    | Redpanda Wasm validation transforms                      | P2    |
| 07     | `07-fusion-engine.md`           | Multi-source track fusion with Kalman filter             | P3    |
| 08     | `08-anomaly-detection.md`       | Rule-based anomaly detection (simulating ML)             | P3    |
| 09     | `09-feedback-trust-scoring.md`  | Operator feedback, trust scoring, anti-poisoning         | P4    |
| 10     | `10-track-service.md`           | Real-time track state cache + gRPC streaming             | P4    |
| 11     | `11-alert-service.md`           | Priority alert queue + gRPC streaming                    | P4    |
| 12     | `12-query-service-etl.md`       | ClickHouse queries + Redpanda Connect ETL pipelines      | P4    |
| 13     | `13-audit-service.md`           | Immutable audit trail (Redpanda → ClickHouse)            | P4    |
| 14     | `14-api-gateway.md`             | Envoy gRPC-Web proxy configuration                       | P4    |
| 15     | `15-cop-web-application.md`     | React COP UI (map, alerts, feedback, forensics)          | P5    |
| 16     | `16-test-data-simulators.md`    | Sensor simulators + realistic scenario data              | P2    |
| 17     | `17-integration-e2e-testing.md` | Cross-module integration + E2E test harness              | P6    |

---

## 3. Implementation Phases & Parallelism

```mermaid
gantt
    title RTSA Implementation Phases
    dateFormat  YYYY-MM-DD
    axisFormat  %b %d

    section P0 Foundation
    Module 01 Dev Infrastructure    :p0a, 2026-02-24, 3d
    Module 02 Protobuf Schemas      :p0b, 2026-02-24, 3d

    section P1 Core Libraries
    Module 03 Shared Go Libraries   :p1, after p0a p0b, 3d

    section P2 Ingestion
    Module 04 Radar Ingestion       :p2a, after p1, 3d
    Module 05 Other Ingestion (5)   :p2b, after p1, 4d
    Module 06 Wasm Transforms       :p2c, after p1, 2d
    Module 16 Test Data Simulators  :p2d, after p1, 3d

    section P3 Processing
    Module 07 Fusion Engine         :p3a, after p1, 4d
    Module 08 Anomaly Detection     :p3b, after p1, 3d

    section P4 Serving
    Module 09 Feedback Service      :p4a, after p1, 3d
    Module 10 Track Service         :p4b, after p1, 2d
    Module 11 Alert Service         :p4c, after p1, 2d
    Module 12 Query Service + ETL   :p4d, after p1, 3d
    Module 13 Audit Service         :p4e, after p1, 2d
    Module 14 API Gateway           :p4f, after p1, 1d

    section P5 Presentation
    Module 15 COP Web Application   :p5, after p1, 5d

    section P6 Integration
    Module 17 Integration + E2E     :p6, after p2a p2b p3a p3b p4a p4b p4c p4d p4e p5, 4d
```

### Parallelism Rules

| Phase  | Modules                | Can Start After              | Max Parallel Agents |
| ------ | ---------------------- | ---------------------------- | ------------------- |
| **P0** | 01, 02                 | Immediately                  | 2                   |
| **P1** | 03                     | P0 complete                  | 1                   |
| **P2** | 04, 05, 06, 16         | P1 complete                  | 4                   |
| **P3** | 07, 08                 | P1 complete                  | 2                   |
| **P4** | 09, 10, 11, 12, 13, 14 | P1 complete                  | 6                   |
| **P5** | 15                     | P1 complete (uses mock gRPC) | 1                   |
| **P6** | 17                     | P2–P5 complete               | 1                   |

> **Key insight**: All modules in P2–P5 depend only on **Module 03 (shared libraries)** for interface contracts. Each module mocks its upstream/downstream dependencies using the interfaces defined in Module 03. This enables maximum parallelism.

---

## 4. Interface Contract Registry

Every module communicates through well-defined interfaces. Parallel agents implement against these contracts and mock counterparts.

### 4.1 Go Interfaces (defined in Module 03 — `pkg/`)

```go
// pkg/redpanda/producer.go
type MessageProducer interface {
    Produce(ctx context.Context, topic string, key []byte, value []byte, headers map[string]string) error
    Close() error
}

// pkg/redpanda/consumer.go
type MessageConsumer interface {
    Consume(ctx context.Context, topics []string, handler MessageHandler) error
    Close() error
}

type MessageHandler func(ctx context.Context, msg *Message) error

type Message struct {
    Topic     string
    Key       []byte
    Value     []byte
    Headers   map[string]string
    Offset    int64
    Partition int32
    Timestamp time.Time
}

// pkg/health/checker.go
type HealthChecker interface {
    Check(ctx context.Context) HealthStatus
    Name() string
}

type HealthStatus struct {
    Status  Status // UP, DOWN, DEGRADED
    Details map[string]string
}

// pkg/audit/emitter.go
type AuditEmitter interface {
    Emit(ctx context.Context, event AuditEvent) error
}

// pkg/classification/guard.go
type ClassificationGuard interface {
    Enforce(dataLevel, requiredLevel ClassificationLevel) error
    MaxLevel(levels ...ClassificationLevel) ClassificationLevel
}
```

### 4.2 gRPC Service Interfaces (defined in Module 02 — `.proto` files)

| Service            | Package             | Key RPCs                                                                 |
| ------------------ | ------------------- | ------------------------------------------------------------------------ |
| `IngestionService` | `rtsa.ingestion.v1` | `IngestSensorData` (client-stream), `GetSensorStatus` (unary)            |
| `TrackService`     | `rtsa.entity.v1`    | `StreamTracks` (server-stream), `GetTrackDetails` (unary)                |
| `AlertService`     | `rtsa.inference.v1` | `StreamAlerts` (server-stream), `AcknowledgeAlert` (unary)               |
| `FeedbackService`  | `rtsa.feedback.v1`  | `SubmitFeedback` (unary), `GetFeedbackHistory` (unary)                   |
| `QueryService`     | `rtsa.query.v1`     | `QueryTracks` (unary), `QueryAnomalies` (unary), `QueryAuditLog` (unary) |
| `HealthService`    | `rtsa.common.v1`    | `Check` (unary), `Watch` (server-stream)                                 |

### 4.3 Redpanda Topic Contracts

| Topic                           | Key            | Producer Module       | Consumer Module(s) |
| ------------------------------- | -------------- | --------------------- | ------------------ |
| `sensors.radar.tracks`          | sensor_id      | 04                    | 07                 |
| `sensors.ew.intercepts`         | sensor_id      | 05                    | 07                 |
| `sensors.elint.detections`      | sensor_id      | 05                    | 07                 |
| `sensors.isr.observations`      | sensor_id      | 05                    | 07                 |
| `sensors.ais.positions`         | mmsi           | 05                    | 07                 |
| `sensors.cyber.iocs`            | ioc_type       | 05                    | 07                 |
| `tracks.fused.surface`          | track_id       | 07                    | 08, 10             |
| `tracks.fused.air`              | track_id       | 07                    | 08, 10             |
| `tracks.fused.subsurface`       | track_id       | 07                    | 08, 10             |
| `tracks.fused.land`             | track_id       | 07                    | 08, 10             |
| `tracks.fused.cyber`            | track_id       | 07                    | 08, 10             |
| `alerts.anomaly.critical`       | track_id       | 08                    | 11                 |
| `alerts.anomaly.elevated`       | track_id       | 08                    | 11                 |
| `alerts.anomaly.watch`          | track_id       | 08                    | 11                 |
| `feedback.operator.submissions` | operator_id    | 09                    | 12 (ETL)           |
| `feedback.operator.validated`   | operator_id    | 09                    | (stub: training)   |
| `models.anomaly.published`      | model_version  | (stub)                | 08                 |
| `audit.events`                  | service_id     | All (via interceptor) | 13                 |
| `dlq.sensors.*`                 | original_topic | 04, 05, 06            | —                  |

### 4.4 ClickHouse Table Contracts (defined in Module 12)

| Table                 | Writer Module | Reader Module  |
| --------------------- | ------------- | -------------- |
| `sensor_observations` | 12 (ETL)      | 12 (Query)     |
| `tracks_fused`        | 12 (ETL)      | 12 (Query)     |
| `anomaly_detections`  | 12 (ETL)      | 12 (Query)     |
| `operator_feedback`   | 12 (ETL)      | 09, 12 (Query) |
| `audit_log`           | 13            | 12 (Query)     |

---

## 5. Requirement → Module Traceability

| Requirement Group      | Requirements       | Modules                    | Use Cases   |
| ---------------------- | ------------------ | -------------------------- | ----------- |
| CR-ING (Ingestion)     | CR-ING-001..010    | 01, 02, 03, 04, 05, 06     | UC002–UC007 |
| CR-FUS (Fusion)        | CR-FUS-001..007    | 07                         | UC008       |
| CR-INF (Inference)     | CR-INF-001..007    | 08                         | UC009       |
| CR-FB (Feedback)       | CR-FB-001..007     | 09                         | UC010       |
| CR-UI (User Interface) | CR-UI-001..008     | 10, 11, 14, 15             | UC012       |
| CR-HIS (Historical)    | CR-HIS-001..007    | 12                         | UC013       |
| CR-NATO (Interop)      | CR-NATO-001..005   | (Stubbed)                  | (Stubbed)   |
| CR-SEC (Security)      | CR-SEC-001..008    | 01, 02, 03 (cross-cutting) | All         |
| NFR-PERF               | NFR-PERF-001..006  | 07, 08, 12, 16, 17         | All         |
| NFR-AVAIL              | NFR-AVAIL-001..004 | 01, 03                     | UC001       |

---

## 6. Feature → Module Mapping

| Feature                          | Modules        | Status      |
| -------------------------------- | -------------- | ----------- |
| FEAT-01 Platform Infrastructure  | 01             | Implement   |
| FEAT-02 Security Framework       | 02, 03         | Implement   |
| FEAT-03 Event Streaming Backbone | 01, 06         | Implement   |
| FEAT-04 Radar Ingestion          | 04             | Implement   |
| FEAT-05 EW/SIGINT Ingestion      | 05             | Implement   |
| FEAT-06 ELINT/COMINT Ingestion   | 05             | Implement   |
| FEAT-07 ISR Metadata Ingestion   | 05             | Implement   |
| FEAT-08 AIS/BFT Ingestion        | 05             | Implement   |
| FEAT-09 Cyber Threat Ingestion   | 05             | Implement   |
| FEAT-10 Multi-Source Fusion      | 07             | Implement   |
| FEAT-11 Anomaly Detection        | 08             | Implement   |
| FEAT-12 Operator Feedback        | 09             | Implement   |
| FEAT-13 Situational Awareness UI | 10, 11, 14, 15 | Implement   |
| FEAT-14 Historical Analysis      | 12             | Implement   |
| FEAT-15 NATO Data Exchange       | —              | **Stubbed** |

---

## 7. Stubbed Modules

The following are **explicitly out of scope** for the core implementation but receive minimal stub implementations that satisfy interface contracts:

### 7.1 NATO Adapter (UC014/UC015)

Create a minimal `svc-nato-adapter/` with:

- A noop consumer that logs `tracks.fused.*` messages it would export
- A noop producer that can inject test NATO data into `sensors.nato.link16` / `sensors.nato.nffi`
- Full gRPC health check so it shows healthy in the dev stack
- No Link 16 or NFFI protocol implementation

### 7.2 Training Pipeline (UC011)

Create a minimal `svc-training/` with:

- A noop consumer on `feedback.operator.validated` that logs received feedback
- No actual model training or anti-poisoning batch validation
- Full gRPC health check
- Produces no events to `models.anomaly.published` (anomaly detection uses pre-configured rule thresholds)

---

## 8. Global Conventions

### 8.1 Classification Header

Every generated source file must begin with:

```
// CLASSIFICATION: UNCLASSIFIED
```

(Adjust comment syntax per file type: `<!-- -->` for HTML/XML, `# ` for YAML/Makefile/Python, `// ` for Go/Proto/TypeScript)

### 8.2 Go Module Path

All Go modules use the base path: `github.com/arvinddhasmana/RTSA_VS_Opus`

- Shared libraries: `github.com/arvinddhasmana/RTSA_VS_Opus/pkg/<package>`
- Services: `github.com/arvinddhasmana/RTSA_VS_Opus/svc-<name>`
- Proto generated: `github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/<context>/v1`

### 8.3 Error Format

```go
fmt.Errorf("[package].[function](%s): %w", identifier, err)
```

### 8.4 Log Schema (slog JSON)

```json
{
  "time": "2026-02-23T14:30:00.000Z",
  "level": "INFO",
  "msg": "event description",
  "service": "svc-radar-ingestion",
  "trace_id": "abc123...",
  "span_id": "def456...",
  "component": "handler",
  "key": "value"
}
```

**Never log**: classified data, raw sensor payloads, PII, credentials, certificate contents.

### 8.5 Metric Naming

```
rtsa_<service>_<metric_name>_<unit>
```

Examples: `rtsa_radar_ingestion_events_total`, `rtsa_fusion_processing_seconds`, `rtsa_track_cache_size`

### 8.6 Config Environment Variables

All prefixed with `RTSA_`:

| Variable                | Default                                        | Description                          |
| ----------------------- | ---------------------------------------------- | ------------------------------------ |
| `RTSA_GRPC_PORT`        | `50051`                                        | gRPC server port                     |
| `RTSA_HEALTH_PORT`      | `8081`                                         | Health check HTTP port               |
| `RTSA_METRICS_PORT`     | `9090`                                         | Prometheus metrics port              |
| `RTSA_REDPANDA_BROKERS` | `localhost:19092`                              | Redpanda broker addresses            |
| `RTSA_CLICKHOUSE_DSN`   | `clickhouse://default:dev@localhost:9000/rtsa` | ClickHouse connection string         |
| `RTSA_TLS_CA_CERT`      | `./certs/dev/ca.crt`                           | TLS CA certificate path              |
| `RTSA_TLS_SERVER_CERT`  | `./certs/dev/server.crt`                       | TLS server certificate path          |
| `RTSA_TLS_SERVER_KEY`   | `./certs/dev/server.key`                       | TLS server key path                  |
| `RTSA_LOG_LEVEL`        | `info`                                         | Log level (debug, info, warn, error) |
| `RTSA_LOG_FORMAT`       | `json`                                         | Log format (json, text)              |
| `RTSA_OTEL_ENDPOINT`    | `localhost:4317`                               | OpenTelemetry collector endpoint     |
| `RTSA_SERVICE_NAME`     | (per service)                                  | Service name for telemetry           |

---

## 9. Agent Invocation Template

To invoke the Greatest Ever Developer agent for any module:

```
@greatest-ever-developer Implement Module XX from docs/implementation/XX-module-name.md

Context:
- Read docs/implementation/00-implementation-overview.md for global conventions and interface contracts
- Read docs/implementation/XX-module-name.md for complete implementation specification
- Follow all SDLC guidelines in docs/sdlc_guidelines/
- Reference architecture docs in docs/architecture/ for design decisions
- Target: development environment (Docker Compose, not K8s)

Deliverables:
1. All source files as specified in the module document
2. Unit tests with ≥80% line coverage
3. Integration test stubs (//go:build integration tag)
4. Dockerfile (multi-stage, distroless runtime)
5. README.md for the module/service
6. All tests passing: go test ./... -race -count=1
```

---

## 10. Definition of Done — Per Module

A module is **complete** when:

- [ ] All source files created as specified in the module document
- [ ] Classification header present on every file
- [ ] Code compiles without errors (`go build ./...` or `pnpm build`)
- [ ] Linter passes (`golangci-lint run ./...` or `eslint .`)
- [ ] Unit tests pass (`go test ./... -race -count=1`)
- [ ] Unit test coverage ≥ 80% line coverage
- [ ] Interface contracts match the definitions in this document
- [ ] Mock implementations provided for all external dependencies
- [ ] Dockerfile builds successfully
- [ ] Health check endpoint responds on `:8081/healthz`
- [ ] README.md documents the service, its configuration, and its dependencies
- [ ] No hardcoded secrets, credentials, or classified data

---

## 11. Project Directory Structure

```
RTSA_VS_Opus/
├── proto/                           # Protobuf definitions (Module 02)
│   └── rtsa/
│       ├── common/v1/
│       ├── ingestion/v1/
│       ├── entity/v1/
│       ├── inference/v1/
│       ├── feedback/v1/
│       ├── query/v1/
│       └── audit/v1/
├── gen/                             # Generated code (Module 02)
│   ├── go/
│   └── ts/
├── pkg/                             # Shared Go libraries (Module 03)
│   ├── config/
│   ├── redpanda/
│   ├── health/
│   ├── shutdown/
│   ├── classification/
│   ├── audit/
│   ├── telemetry/
│   ├── interceptors/
│   └── testutil/
├── svc-radar-ingestion/             # Radar ingestion (Module 04)
├── svc-ew-ingestion/                # EW/SIGINT ingestion (Module 05)
├── svc-elint-ingestion/             # ELINT/COMINT ingestion (Module 05)
├── svc-isr-ingestion/               # ISR ingestion (Module 05)
├── svc-ais-ingestion/               # AIS/BFT ingestion (Module 05)
├── svc-cyber-ingestion/             # Cyber ingestion (Module 05)
├── svc-fusion-engine/               # Fusion engine (Module 07)
├── svc-anomaly-detection/           # Anomaly detection (Module 08)
├── svc-feedback/                    # Feedback service (Module 09)
├── svc-track/                       # Track service (Module 10)
├── svc-alert/                       # Alert service (Module 11)
├── svc-query/                       # Query service (Module 12)
├── svc-audit/                       # Audit service (Module 13)
├── svc-nato-adapter/                # NATO adapter — stub (Module 07 stub)
├── svc-training/                    # Training pipeline — stub (Module 09 stub)
├── transforms/                      # Wasm transforms (Module 06)
│   ├── sensor-validator/
│   └── classification-guard/
├── tools/                           # Development tools (Module 16)
│   └── simulator/
├── ui/                              # COP Web Application (Module 15)
├── deploy/                          # Deployment configs (Module 01)
│   ├── docker-compose.yml
│   ├── docker-compose.services.yml
│   ├── envoy/
│   ├── prometheus/
│   ├── grafana/
│   ├── loki/
│   ├── otel/
│   └── redpanda-connect/
├── tests/                           # Cross-module tests (Module 17)
│   ├── integration/
│   └── e2e/
├── buf.yaml
├── buf.gen.yaml
├── Makefile
├── go.work                          # Go workspace for multi-module
├── go.work.sum
└── docs/
    └── implementation/              # These instruction files
```

---

## 12. Technology Versions (Dev Environment)

| Technology       | Version | Image                                         |
| ---------------- | ------- | --------------------------------------------- |
| Go               | 1.22+   | `golang:1.22-alpine` (build)                  |
| Redpanda         | 24.x    | `redpandadata/redpanda:v24.1.1`               |
| ClickHouse       | 24.x    | `clickhouse/clickhouse-server:24.1`           |
| Prometheus       | 2.x     | `prom/prometheus:v2.51.0`                     |
| Grafana          | 10.x    | `grafana/grafana:10.4.0`                      |
| Loki             | 2.x     | `grafana/loki:2.9.0`                          |
| Tempo            | 2.x     | `grafana/tempo:2.4.0`                         |
| OTel Collector   | 0.96+   | `otel/opentelemetry-collector-contrib:0.96.0` |
| Envoy            | 1.29+   | `envoyproxy/envoy:v1.29.0`                    |
| Redpanda Console | 2.x     | `redpandadata/console:v2.4.0`                 |
| Redpanda Connect | 4.x     | `redpandadata/connect:4.27.0`                 |
| Node.js          | 20 LTS  | —                                             |
| React            | 18+     | —                                             |
| TypeScript       | 5+      | —                                             |
| buf              | 1.32+   | —                                             |
| Runtime image    | —       | `gcr.io/distroless/static-debian12:nonroot`   |
