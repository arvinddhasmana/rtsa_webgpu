<!-- CLASSIFICATION: UNCLASSIFIED -->

# Module 04 — Sensor Ingestion: Radar (Reference Implementation)

> **Module**: 04-sensor-ingestion-radar
> **Phase**: P1 (Ingestion)
> **Dependencies**: Module 02 (protos), Module 03 (shared libraries)
> **Agent**: `@greatest-ever-developer`
> **Estimated Effort**: 4 days

---

## 1. Objective

Implement the Radar Ingestion Service (`svc-radar-ingestion`) as the **reference implementation** for all sensor ingestion services. This service receives radar track reports via gRPC, validates, normalizes, enriches, and produces them to Redpanda. All other sensor ingestion services (Module 05) follow this exact pattern with sensor-specific validators and normalizers.

**Acceptance Criteria**:

- gRPC `IngestionService` accepts radar observations via `IngestSingleObservation` and `IngestSensorData` (client-streaming)
- All validation rules enforced; invalid messages routed to DLQ
- Valid messages produced to `sensors.radar.tracks` with all 5 standard headers
- Health check endpoint functional (liveness + readiness)
- Metrics exposed: `rtsa_ingestion_observations_total`, `rtsa_ingestion_validation_errors_total`, `rtsa_ingestion_produce_duration_seconds`
- ≥80% line coverage
- `GetSensorStatus` returns live sensor statistics

---

## 2. Service Structure

```
svc-radar-ingestion/
├── cmd/
│   └── radar-ingestion/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go           # Service-specific config
│   ├── domain/
│   │   ├── validator.go        # Radar-specific validation rules
│   │   ├── validator_test.go
│   │   ├── normalizer.go       # Raw → SensorObservation mapping
│   │   └── normalizer_test.go
│   ├── handler/
│   │   ├── ingestion.go        # gRPC handler implementation
│   │   └── ingestion_test.go
│   ├── producer/
│   │   ├── observation.go      # Produces to sensors.radar.tracks
│   │   └── observation_test.go
│   └── mapper/
│       ├── enricher.go         # Adds classification, trace, observation_id
│       └── enricher_test.go
├── go.mod
├── Dockerfile
├── Makefile
└── README.md
```

---

## 3. `go.mod`

```go
// CLASSIFICATION: UNCLASSIFIED
module github.com/arvinddhasmana/RTSA_VS_Opus/svc-radar-ingestion

go 1.22.0

require (
    github.com/arvinddhasmana/RTSA_VS_Opus/gen/go v0.0.0
    github.com/arvinddhasmana/RTSA_VS_Opus/pkg v0.0.0
    google.golang.org/grpc v1.62.0
    google.golang.org/protobuf v1.33.0
    go.uber.org/zap v1.27.0
    github.com/google/uuid v1.6.0
)

replace (
    github.com/arvinddhasmana/RTSA_VS_Opus/gen/go => ../gen/go
    github.com/arvinddhasmana/RTSA_VS_Opus/pkg => ../pkg
)
```

---

## 4. Component Specifications

### 4.1 `cmd/radar-ingestion/main.go` — Entry Point

```go
// CLASSIFICATION: UNCLASSIFIED
package main

// main.go wires all dependencies and starts the service.
// Initialization order:
// 1. Load config (MustLoad)
// 2. Initialize telemetry (tracing + metrics + logger)
// 3. Create classification guard (ceiling from config)
// 4. Create Redpanda producer (for sensors.radar.tracks)
// 5. Create DLQ producer (for dlq.sensors.radar)
// 6. Create audit emitter
// 7. Create health checker, register components
// 8. Create domain validator, normalizer, enricher
// 9. Create gRPC handler
// 10. Create gRPC server with interceptor chain
// 11. Register IngestionService and HealthService
// 12. Create shutdown manager, register hooks (LIFO):
//     - gRPC server GracefulStop
//     - Redpanda producer Close
//     - DLQ producer Close
//     - Telemetry Shutdown
// 13. Start gRPC server (non-blocking)
// 14. Set health to SERVING
// 15. shutdown.Wait()

func main() {
    // Load config
    cfg := config.MustLoad()

    // Init telemetry
    tp, err := telemetry.Init(ctx, telemetry.Config{
        ServiceName:    "svc-radar-ingestion",
        ServiceVersion: cfg.ServiceVersion,
        Environment:    cfg.Environment,
        OTelEndpoint:   cfg.OTelEndpoint,
    })
    if err != nil { /* log.Fatal */ }

    // Classification guard
    guard := classification.NewGuard(classification.StringToLevel(cfg.MaxClassification))

    // Producers
    producer, _ := redpanda.NewProducer(ctx, redpanda.ProducerConfig{
        Connection:    buildConnOpts(cfg),
        ServiceName:   "svc-radar-ingestion",
        SchemaVersion: "1.0.0",
    })
    dlqProducer, _ := redpanda.NewProducer(ctx, redpanda.ProducerConfig{/*same*/})

    // Audit
    auditEmitter := audit.NewEmitter(producer, "svc-radar-ingestion", tp.Logger)

    // Health
    healthChecker := health.NewChecker()
    healthChecker.Register("redpanda")
    healthChecker.Register("grpc")

    // Domain
    validator := domain.NewRadarValidator(tp.Logger)
    normalizer := domain.NewRadarNormalizer()
    enricher := mapper.NewEnricher("svc-radar-ingestion", guard)

    // Handler
    handler := handler.NewIngestionHandler(validator, normalizer, enricher,
        producer, dlqProducer, auditEmitter, tp.Logger)

    // gRPC server
    srv := grpc.NewServer(
        grpc.ChainUnaryInterceptor(interceptors.BuildUnaryServerInterceptors(/*...*/)),
        grpc.ChainStreamInterceptor(interceptors.BuildStreamServerInterceptors(/*...*/)),
    )
    ingestionv1.RegisterIngestionServiceServer(srv, handler)
    commonv1.RegisterHealthServiceServer(srv, health.NewServer(healthChecker))

    // Shutdown manager
    sm := shutdown.NewManager(tp.Logger, 30*time.Second)
    sm.Register("grpc-server", func(ctx context.Context) error { srv.GracefulStop(); return nil })
    sm.Register("producer", producer.Close)
    sm.Register("dlq-producer", dlqProducer.Close)
    sm.Register("telemetry", tp.Shutdown)

    // Start
    lis, _ := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
    go srv.Serve(lis)
    healthChecker.SetStatus("grpc", health.StatusServing)
    healthChecker.SetStatus("redpanda", health.StatusServing)

    sm.Wait()
}
```

### 4.2 `internal/config/config.go` — Service Configuration

```go
// CLASSIFICATION: UNCLASSIFIED
package config

import "github.com/arvinddhasmana/RTSA_VS_Opus/pkg/config"

// Config extends BaseConfig with radar-specific settings.
type Config struct {
    config.BaseConfig

    // ── Radar-specific ──
    // Topic to produce validated observations to
    OutputTopic string // RTSA_RADAR_OUTPUT_TOPIC (default: "sensors.radar.tracks")
    // DLQ topic for invalid messages
    DLQTopic string // RTSA_RADAR_DLQ_TOPIC (default: "dlq.sensors.radar")
    // Maximum speed in knots for surface tracks (for validation)
    MaxSurfaceSpeedKnots float64 // RTSA_RADAR_MAX_SURFACE_SPEED (default: 999)
    // Maximum speed in knots for air tracks
    MaxAirSpeedKnots float64 // RTSA_RADAR_MAX_AIR_SPEED (default: 2500)
    // Maximum accepted future time offset (seconds)
    MaxFutureOffsetSec int // RTSA_RADAR_MAX_FUTURE_OFFSET (default: 300)
    // Maximum accepted past time offset (seconds)
    MaxPastOffsetSec int // RTSA_RADAR_MAX_PAST_OFFSET (default: 86400)
}

// MustLoad loads and validates radar ingestion config.
func MustLoad() *Config { /* implementation */ }
```

### 4.3 `internal/domain/validator.go` — Radar Validation Rules

```go
// CLASSIFICATION: UNCLASSIFIED
package domain

import (
    "fmt"
    "time"
    ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
    commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
    "go.uber.org/zap"
)

// ValidationResult holds the outcome of validation.
type ValidationResult struct {
    Valid    bool
    Errors   []ValidationError
    Warnings []string
}

// ValidationError describes a single validation failure.
type ValidationError struct {
    Field   string // e.g., "position.latitude"
    Value   string // String representation of the invalid value
    Rule    string // e.g., "range:-90.0 to +90.0"
    Message string // Human-readable error
}

// RadarValidator validates radar-specific SensorObservation messages.
type RadarValidator struct {
    logger              *zap.Logger
    maxSurfaceSpeedKnots float64
    maxAirSpeedKnots    float64
    maxFutureOffset     time.Duration
    maxPastOffset       time.Duration
}

// NewRadarValidator creates a validator with configured thresholds.
func NewRadarValidator(logger *zap.Logger, opts ...ValidatorOption) *RadarValidator { /* implementation */ }

// Validate checks all validation rules for a SensorObservation.
// Rules (from data_architecture.md §8.1):
//
// REJECT TO DLQ (hard failures):
//   - sensor_id: non-empty, max 128 chars
//   - sensor_type: must be SENSOR_TYPE_RADAR
//   - observation_time: not > maxFutureOffset in future, not > maxPastOffset in past
//   - position.latitude: -90.0 to +90.0 (if position provided)
//   - position.longitude: -180.0 to +180.0 (if position provided)
//   - position.heading_degrees: 0.0 to 360.0 (if provided)
//   - classification: valid enum value (not UNSPECIFIED)
//   - radar.track_number: non-empty
//   - radar.track_quality: 0.0 to 1.0
//
// FLAG AS SUSPECT (soft warnings):
//   - position.speed_knots: 0 to maxSurfaceSpeedKnots (surface) or maxAirSpeedKnots (air)
//   - radar.range_nm: 0 to 500
//   - radar.bearing_degrees: 0.0 to 360.0
//
func (v *RadarValidator) Validate(obs *ingestionv1.SensorObservation) *ValidationResult { /* implementation */ }
```

**Test Cases** (`validator_test.go`):

| #   | Test                                | Input                          | Expected                                       |
| --- | ----------------------------------- | ------------------------------ | ---------------------------------------------- |
| T01 | Valid radar observation             | All fields valid               | Valid=true, no errors                          |
| T02 | Missing sensor_id                   | sensor_id=""                   | Valid=false, error on sensor_id                |
| T03 | Wrong sensor_type                   | sensor_type=EW_SIGINT          | Valid=false, error on sensor_type              |
| T04 | Latitude out of range               | lat=91.0                       | Valid=false, error on position.latitude        |
| T05 | Longitude out of range              | lon=-181.0                     | Valid=false, error on position.longitude       |
| T06 | Heading out of range                | heading=361.0                  | Valid=false, error on position.heading_degrees |
| T07 | Future timestamp                    | observation_time = now + 10min | Valid=false, error on observation_time         |
| T08 | Past timestamp                      | observation_time = now - 25h   | Valid=false, error on observation_time         |
| T09 | Speed warning                       | speed=1500 (surface)           | Valid=true, warning present                    |
| T10 | Missing classification              | classification=UNSPECIFIED     | Valid=false, error on classification           |
| T11 | Missing radar track_number          | track_number=""                | Valid=false, error on radar.track_number       |
| T12 | Track quality out of range          | quality=1.5                    | Valid=false, error on radar.track_quality      |
| T13 | No position (valid for some radars) | position=nil                   | Valid=true (position optional)                 |
| T14 | Multiple errors                     | lat=91, lon=181, speed=-1      | Valid=false, all 3 errors listed               |

### 4.4 `internal/domain/normalizer.go` — Proto Normalization

```go
// CLASSIFICATION: UNCLASSIFIED
package domain

import (
    ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
)

// RadarNormalizer ensures consistent field formatting.
type RadarNormalizer struct{}

// NewRadarNormalizer creates a normalizer.
func NewRadarNormalizer() *RadarNormalizer { /* implementation */ }

// Normalize standardizes the observation:
// - Trims whitespace from string fields
// - Normalizes heading to 0-360 range
// - Ensures position coordinates use correct precision (6 decimal places)
// - Copies sensor_type from radar context if not set
func (n *RadarNormalizer) Normalize(obs *ingestionv1.SensorObservation) *ingestionv1.SensorObservation { /* implementation */ }
```

### 4.5 `internal/mapper/enricher.go` — Metadata Enrichment

```go
// CLASSIFICATION: UNCLASSIFIED
package mapper

import (
    "context"
    "github.com/google/uuid"
    ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
    "github.com/arvinddhasmana/RTSA_VS_Opus/pkg/classification"
    "google.golang.org/protobuf/types/known/timestamppb"
)

// Enricher adds system-generated fields to observations.
type Enricher struct {
    serviceName string
    guard       *classification.Guard
}

// NewEnricher creates an enricher.
func NewEnricher(serviceName string, guard *classification.Guard) *Enricher { /* implementation */ }

// Enrich adds:
// - observation_id: UUID v7 (time-ordered) if not already set
// - Classification ceiling enforcement (returns error if above ceiling)
// - metadata["rtsa.source_service"] = serviceName
// - metadata["rtsa.ingestion_time"] = current UTC time
func (e *Enricher) Enrich(ctx context.Context, obs *ingestionv1.SensorObservation) error { /* implementation */ }
```

### 4.6 `internal/handler/ingestion.go` — gRPC Handler

```go
// CLASSIFICATION: UNCLASSIFIED
package handler

import (
    "context"
    "sync/atomic"
    "time"

    ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
    "github.com/arvinddhasmana/RTSA_VS_Opus/pkg/audit"
    "go.uber.org/zap"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

// IngestionHandler implements IngestionService for radar data.
type IngestionHandler struct {
    ingestionv1.UnimplementedIngestionServiceServer

    validator   *domain.RadarValidator
    normalizer  *domain.RadarNormalizer
    enricher    *mapper.Enricher
    producer    *producer.ObservationProducer
    dlqProducer *producer.ObservationProducer
    auditEmitter *audit.Emitter
    logger      *zap.Logger

    // Statistics (atomic for thread safety)
    totalReceived  atomic.Int64
    totalAccepted  atomic.Int64
    totalRejected  atomic.Int64
    lastObsTime    atomic.Value // time.Time
}

// IngestSingleObservation handles unary radar observation ingestion.
// Flow:
//   1. Increment totalReceived counter
//   2. Validate observation → if invalid: produce to DLQ, increment rejected, return IngestionAck{accepted: false}
//   3. Normalize observation
//   4. Enrich observation (add observation_id, check classification)
//   5. Produce to output topic
//   6. Emit audit event
//   7. Increment totalAccepted
//   8. Return IngestionAck{observation_id, accepted: true}
//
// Error handling:
//   - Validation failure: NOT a gRPC error — returns IngestionAck{accepted: false, rejection_reason: "..."}
//   - Classification violation: returns codes.PermissionDenied
//   - Producer failure: returns codes.Internal with wrapped error
func (h *IngestionHandler) IngestSingleObservation(ctx context.Context,
    obs *ingestionv1.SensorObservation) (*ingestionv1.IngestionAck, error) { /* implementation */ }

// IngestSensorData handles client-streaming radar observation ingestion.
// Processes each observation in the stream using the same validate→normalize→enrich→produce pipeline.
// Returns IngestSummary with total/accepted/rejected counts.
// Does NOT fail the entire stream on individual observation failures.
func (h *IngestionHandler) IngestSensorData(
    stream ingestionv1.IngestionService_IngestSensorDataServer) error { /* implementation */ }

// GetSensorStatus returns live statistics for the radar ingestion service.
func (h *IngestionHandler) GetSensorStatus(ctx context.Context,
    req *ingestionv1.GetSensorStatusRequest) (*ingestionv1.SensorStatusResponse, error) { /* implementation */ }
```

### 4.7 `internal/producer/observation.go` — Redpanda Producer

```go
// CLASSIFICATION: UNCLASSIFIED
package producer

import (
    "context"
    ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
    "github.com/arvinddhasmana/RTSA_VS_Opus/pkg/classification"
    "github.com/arvinddhasmana/RTSA_VS_Opus/pkg/redpanda"
    "google.golang.org/protobuf/proto"
)

// ObservationProducer produces SensorObservation messages to Redpanda.
type ObservationProducer struct {
    producer *redpanda.Producer
    topic    string
}

// NewObservationProducer creates a producer for a specific topic.
func NewObservationProducer(producer *redpanda.Producer, topic string) *ObservationProducer { /* implementation */ }

// Produce serializes and sends a SensorObservation.
// Key: sensor_id (for partition affinity — same sensor always same partition)
// Value: protobuf-serialized SensorObservation
// Headers: standard headers + sensor-specific metadata
func (p *ObservationProducer) Produce(ctx context.Context,
    obs *ingestionv1.SensorObservation) error { /* implementation */ }
```

---

## 5. Metrics Reference

| Metric Name                               | Type      | Labels                                      | Description                                 |
| ----------------------------------------- | --------- | ------------------------------------------- | ------------------------------------------- |
| `rtsa_ingestion_observations_total`       | Counter   | `sensor_type`, `status` (accepted/rejected) | Total observations processed                |
| `rtsa_ingestion_validation_errors_total`  | Counter   | `sensor_type`, `field`, `rule`              | Validation failures by field and rule       |
| `rtsa_ingestion_produce_duration_seconds` | Histogram | `sensor_type`, `topic`                      | Redpanda produce latency                    |
| `rtsa_ingestion_active_sensors`           | Gauge     | `sensor_type`                               | Number of sensors that reported in last 60s |
| `rtsa_ingestion_dlq_messages_total`       | Counter   | `sensor_type`, `reason`                     | Messages sent to DLQ                        |

---

## 6. Error Handling Patterns

```go
// Validation failure — NOT a gRPC error
return &ingestionv1.IngestionAck{
    ObservationId:   obs.GetObservationId(),
    Accepted:        false,
    RejectionReason: fmt.Sprintf("validation failed: %s", result.Errors[0].Message),
}, nil

// Classification violation — gRPC error
return nil, status.Errorf(codes.PermissionDenied,
    "classification: data level %s exceeds service ceiling %s",
    classification.LevelToString(obs.GetClassification()),
    classification.LevelToString(cfg.MaxClassification))

// Producer failure — gRPC error with context
return nil, status.Errorf(codes.Internal,
    "ingestion: failed to produce observation %s: %v",
    obs.GetObservationId(), err)
```

---

## 7. Structured Log Examples

```json
{"level":"info","ts":"2026-02-23T14:00:00.000Z","caller":"handler/ingestion.go:45","msg":"observation accepted","service":"svc-radar-ingestion","sensor_id":"RADAR-HALIFAX-01","observation_id":"01905b6e-...","sensor_type":"RADAR","classification":"UNCLASSIFIED","trace_id":"abc123"}

{"level":"warn","ts":"2026-02-23T14:00:00.100Z","caller":"handler/ingestion.go:52","msg":"observation rejected","service":"svc-radar-ingestion","sensor_id":"RADAR-HALIFAX-01","observation_id":"01905b6f-...","reason":"position.latitude out of range: 91.0","trace_id":"def456"}

{"level":"error","ts":"2026-02-23T14:00:01.000Z","caller":"producer/observation.go:35","msg":"produce failed","service":"svc-radar-ingestion","topic":"sensors.radar.tracks","error":"broker connection refused","trace_id":"ghi789"}
```

---

## 8. Test Scenarios

### 8.1 Unit Tests

| #   | Test                                       | Package | Expected                            |
| --- | ------------------------------------------ | ------- | ----------------------------------- |
| T01 | Valid radar observation accepted           | handler | IngestionAck.Accepted=true          |
| T02 | Invalid observation → DLQ                  | handler | DLQ producer called, Accepted=false |
| T03 | Client streaming with mix of valid/invalid | handler | Correct IngestSummary counts        |
| T04 | Classification violation                   | handler | PermissionDenied error              |
| T05 | Producer error propagated                  | handler | Internal error                      |
| T06 | GetSensorStatus returns stats              | handler | Correct counts                      |
| T07 | Enricher sets observation_id               | mapper  | UUID v7 format                      |
| T08 | Enricher checks classification ceiling     | mapper  | Error on exceeded ceiling           |
| T09 | All validator rules (see §4.3)             | domain  | Per test case table                 |
| T10 | Normalizer trims whitespace                | domain  | Fields trimmed                      |

### 8.2 Integration Tests

| #    | Test                                 | Setup                     | Expected                             |
| ---- | ------------------------------------ | ------------------------- | ------------------------------------ |
| IT01 | Full pipeline: gRPC → Redpanda       | testcontainers (Redpanda) | Message in sensors.radar.tracks      |
| IT02 | DLQ routing                          | testcontainers (Redpanda) | Invalid message in dlq.sensors.radar |
| IT03 | Headers present on produced messages | testcontainers (Redpanda) | All 5 standard headers               |
| IT04 | Client streaming 1000 messages       | testcontainers (Redpanda) | All 1000 in topic                    |
| IT05 | Health check reflects readiness      | -                         | SERVING after startup                |

---

## 9. Agent Invocation

```
@greatest-ever-developer Implement Module 04 from docs/implementation/04-sensor-ingestion-radar.md

Context:
- Read docs/implementation/00-implementation-overview.md for global conventions and project structure
- Read docs/implementation/02-protobuf-schemas.md for proto type definitions
- Read docs/implementation/03-shared-go-libraries.md for pkg/ interfaces
- This is the REFERENCE implementation for all ingestion services
- Follow the service structure in §2 exactly
- Implement ALL validation rules from the data_architecture.md validation table
- Use pkg/redpanda, pkg/health, pkg/shutdown, pkg/classification, pkg/audit, pkg/telemetry, pkg/interceptors
- All errors must use fmt.Errorf("context: %w", err) pattern
- No panic() in non-test code
- Structured JSON logging only; NEVER log classified data or sensor payloads

Deliverables:
1. go.mod with correct dependencies and replacements
2. All source files under svc-radar-ingestion/ as specified
3. Unit tests for every package (≥80% coverage)
4. Integration tests using testcontainers-go
5. Dockerfile (multi-stage, distroless base)
6. go vet ./... passes
7. go test ./... passes
```
