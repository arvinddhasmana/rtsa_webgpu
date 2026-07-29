<!-- CLASSIFICATION: UNCLASSIFIED -->

# Module 03 — Shared Go Libraries

> **Module**: 03-shared-go-libraries
> **Phase**: P0 (Foundation)
> **Dependencies**: Module 02 (proto definitions)
> **Agent**: `@greatest-ever-developer`
> **Estimated Effort**: 4 days

---

## 1. Objective

Implement the shared Go library packages used by every RTSA microservice. These packages provide cross-cutting concerns: configuration, messaging, health checks, graceful shutdown, classification enforcement, audit event emission, telemetry initialization, gRPC interceptors, and test utilities.

**Acceptance Criteria**:

- All packages under `pkg/` compile with zero errors
- Each package has ≥80% line coverage
- All packages use `github.com/arvinddhasmana/rtsa_webgpu/pkg` module path
- No external dependencies outside approved supply chain (see `docs/sdlc_guidelines/01_security_compliance/supply_chain_security.md`)
- Zero uses of `panic()` outside test files
- All errors wrapped with `fmt.Errorf("context: %w", err)` pattern

---

## 2. Module Structure

```
pkg/
├── go.mod
├── go.sum
├── config/
│   ├── config.go          # Environment config loader
│   └── config_test.go
├── redpanda/
│   ├── producer.go        # franz-go producer wrapper
│   ├── consumer.go        # franz-go consumer wrapper
│   ├── headers.go         # Standard message headers
│   ├── options.go         # Connection options & TLS
│   ├── producer_test.go
│   ├── consumer_test.go
│   └── headers_test.go
├── health/
│   ├── server.go          # gRPC health check server
│   ├── checker.go         # Component health registry
│   └── server_test.go
├── shutdown/
│   ├── manager.go         # Graceful shutdown orchestrator
│   └── manager_test.go
├── classification/
│   ├── guard.go           # Classification level enforcement
│   ├── propagation.go     # Classification MAX propagation
│   └── guard_test.go
├── audit/
│   ├── emitter.go         # Audit event builder & producer
│   ├── types.go           # Audit event constants
│   └── emitter_test.go
├── telemetry/
│   ├── provider.go        # OTel tracer/meter/logger init
│   ├── attributes.go      # Standard metric attributes
│   └── provider_test.go
├── interceptors/
│   ├── chain.go           # Interceptor chain builder
│   ├── logging.go         # Structured logging interceptor
│   ├── metrics.go         # gRPC metrics interceptor
│   ├── tracing.go         # Distributed tracing interceptor
│   ├── classification.go  # Classification check interceptor
│   ├── recovery.go        # Panic recovery interceptor
│   └── interceptors_test.go
└── testutil/
    ├── redpanda.go        # Test Redpanda container (testcontainers-go)
    ├── clickhouse.go       # Test ClickHouse container
    ├── grpc.go            # Test gRPC server/client setup
    ├── proto.go           # Proto message builders for tests
    └── assertions.go      # Custom test assertions
```

---

## 3. `pkg/go.mod`

```go
// CLASSIFICATION: UNCLASSIFIED
module github.com/arvinddhasmana/rtsa_webgpu/pkg

go 1.22.0

require (
    github.com/arvinddhasmana/rtsa_webgpu/gen/go v0.0.0
    github.com/twmb/franz-go v1.16.0
    github.com/twmb/franz-go/pkg/kadm v1.11.0
    go.opentelemetry.io/otel v1.24.0
    go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.24.0
    go.opentelemetry.io/otel/exporters/prometheus v0.46.0
    go.opentelemetry.io/otel/sdk v1.24.0
    go.opentelemetry.io/otel/sdk/metric v1.24.0
    go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.49.0
    google.golang.org/grpc v1.62.0
    google.golang.org/protobuf v1.33.0
    go.uber.org/zap v1.27.0
    github.com/google/uuid v1.6.0
    github.com/testcontainers/testcontainers-go v0.29.0
)

replace github.com/arvinddhasmana/rtsa_webgpu/gen/go => ../gen/go
```

---

## 4. Package Specifications

### 4.1 `pkg/config` — Environment Configuration Loader

**File: `config.go`**

```go
// CLASSIFICATION: UNCLASSIFIED
package config

// BaseConfig contains configuration fields shared by all services.
// All fields are loaded from environment variables with the RTSA_ prefix.
type BaseConfig struct {
    // ── Service Identity ──
    ServiceName    string // RTSA_SERVICE_NAME (required)
    ServiceVersion string // RTSA_SERVICE_VERSION (default: "dev")
    Environment    string // RTSA_ENVIRONMENT (default: "development")

    // ── gRPC Server ──
    GRPCPort int    // RTSA_GRPC_PORT (default: 50051)
    TLSCert  string // RTSA_TLS_CERT (default: "/certs/server.crt")
    TLSKey   string // RTSA_TLS_KEY (default: "/certs/server.key")
    TLSCA    string // RTSA_TLS_CA (default: "/certs/ca.crt")

    // ── Redpanda ──
    RedpandaBrokers  []string // RTSA_REDPANDA_BROKERS (comma-separated, default: "localhost:9092")
    RedpandaTLSEnabled bool   // RTSA_REDPANDA_TLS_ENABLED (default: true)

    // ── ClickHouse ──
    ClickHouseDSN string // RTSA_CLICKHOUSE_DSN (default: "clickhouse://default:@localhost:9440/rtsa?secure=true")

    // ── Observability ──
    OTelEndpoint string // RTSA_OTEL_ENDPOINT (default: "localhost:4317")
    LogLevel     string // RTSA_LOG_LEVEL (default: "info")
    LogFormat    string // RTSA_LOG_FORMAT (default: "json")

    // ── Classification ──
    MaxClassification string // RTSA_MAX_CLASSIFICATION (default: "UNCLASSIFIED")
}

// Load reads configuration from environment variables.
// Required variables cause an error if missing.
// All env vars use RTSA_ prefix.
// Returns a validated BaseConfig or error with context for each missing required var.
func Load() (*BaseConfig, error) { /* implementation */ }

// MustLoad calls Load and panics on error. Use only in main.go.
func MustLoad() *BaseConfig { /* implementation */ }

// LoadInto loads environment variables into a user-provided struct.
// Struct fields must be tagged with `env:"RTSA_VAR_NAME"` and optionally `envDefault:"value"`.
// Use `envRequired:"true"` for mandatory fields.
func LoadInto(cfg interface{}) error { /* implementation */ }
```

**Implementation Notes**:

- Use `os.LookupEnv` for each field
- Parse comma-separated values for slices
- Validate required fields; return `fmt.Errorf("config: RTSA_%s is required", name)` for missing required vars
- Log loaded config (redact TLS key paths, passwords) at DEBUG level
- Do NOT use third-party config libraries. Use stdlib only.

**Test Cases** (`config_test.go`):

| #   | Test                           | Setup                                 | Expected                                         |
| --- | ------------------------------ | ------------------------------------- | ------------------------------------------------ |
| T01 | Load with all env vars set     | Set all RTSA\_\* vars                 | Success, all fields populated                    |
| T02 | Load with missing required var | Unset RTSA_SERVICE_NAME               | Error containing "RTSA_SERVICE_NAME is required" |
| T03 | Load with defaults             | Set only required vars                | Default values applied                           |
| T04 | Comma-separated brokers        | RTSA_REDPANDA_BROKERS="a:9092,b:9092" | Slice with 2 elements                            |
| T05 | Invalid port                   | RTSA_GRPC_PORT="abc"                  | Error containing "invalid port"                  |
| T06 | LoadInto custom struct         | Tagged struct                         | Fields populated                                 |

---

### 4.2 `pkg/redpanda` — Kafka Producer/Consumer Wrapper

**File: `options.go`**

```go
// CLASSIFICATION: UNCLASSIFIED
package redpanda

import (
    "crypto/tls"
    "github.com/twmb/franz-go/pkg/kgo"
)

// ConnectionOptions configures the Redpanda client connection.
type ConnectionOptions struct {
    Brokers      []string
    TLSEnabled   bool
    TLSCertFile  string
    TLSKeyFile   string
    TLSCAFile    string
    ClientID     string // Set to service name
    SASL         *SASLConfig // nil for mTLS-only auth
}

// SASLConfig holds SASL authentication credentials.
type SASLConfig struct {
    Mechanism string // "SCRAM-SHA-256" or "SCRAM-SHA-512"
    Username  string
    Password  string
}

// BuildKgoOpts converts ConnectionOptions into franz-go kgo.Opt slice.
// Configures TLS, SASL, client ID, retry, and compression.
func (o *ConnectionOptions) BuildKgoOpts() ([]kgo.Opt, error) { /* implementation */ }
```

**File: `headers.go`**

```go
// CLASSIFICATION: UNCLASSIFIED
package redpanda

import (
    "time"
    "github.com/twmb/franz-go/pkg/kgo"
)

// Standard RTSA message header keys.
const (
    HeaderClassification = "rtsa-classification"
    HeaderSourceService  = "rtsa-source-service"
    HeaderTraceID        = "rtsa-trace-id"
    HeaderTimestamp       = "rtsa-timestamp"
    HeaderSchemaVersion  = "rtsa-schema-version"
)

// StandardHeaders returns the required headers for every RTSA message.
// classification: e.g., "UNCLASSIFIED"
// sourceService: e.g., "svc-radar-ingestion"
// traceID: OpenTelemetry trace ID (hex string)
// schemaVersion: e.g., "1.0.0"
func StandardHeaders(classification, sourceService, traceID, schemaVersion string) []kgo.RecordHeader {
    return []kgo.RecordHeader{
        {Key: HeaderClassification, Value: []byte(classification)},
        {Key: HeaderSourceService, Value: []byte(sourceService)},
        {Key: HeaderTraceID, Value: []byte(traceID)},
        {Key: HeaderTimestamp, Value: []byte(time.Now().UTC().Format(time.RFC3339Nano))},
        {Key: HeaderSchemaVersion, Value: []byte(schemaVersion)},
    }
}

// GetHeader extracts a header value from a record. Returns "" if not found.
func GetHeader(record *kgo.Record, key string) string { /* implementation */ }
```

**File: `producer.go`**

```go
// CLASSIFICATION: UNCLASSIFIED
package redpanda

import (
    "context"
    "github.com/twmb/franz-go/pkg/kgo"
)

// Producer wraps franz-go client for producing messages to Redpanda.
type Producer struct {
    client      *kgo.Client
    serviceName string
    schemaVer   string
}

// ProducerConfig configures a new Producer.
type ProducerConfig struct {
    Connection    ConnectionOptions
    ServiceName   string // Used in rtsa-source-service header
    SchemaVersion string // Used in rtsa-schema-version header
    // MaxRetries for transient failures (default: 3)
    MaxRetries int
    // Compression algorithm (default: zstd)
    Compression string
    // RequiredAcks: "all" (default), "leader", "none"
    RequiredAcks string
}

// NewProducer creates a Redpanda producer with standard configuration.
// Returns error if TLS setup fails or broker connection fails.
func NewProducer(ctx context.Context, cfg ProducerConfig) (*Producer, error) { /* implementation */ }

// Produce sends a single message to the specified topic.
// key: message key for partitioning
// value: serialized protobuf bytes
// classification: classification level of the data
// traceID: OpenTelemetry trace ID
// Additional headers can be appended.
// Returns error if produce fails after retries.
func (p *Producer) Produce(ctx context.Context, topic string, key []byte, value []byte,
    classification string, traceID string, extraHeaders ...kgo.RecordHeader) error { /* implementation */ }

// ProduceBatch sends multiple messages atomically (same partition).
// Returns error if any message fails.
func (p *Producer) ProduceBatch(ctx context.Context, records []*kgo.Record) error { /* implementation */ }

// Close flushes pending messages and closes the producer.
// Waits up to 10s for flush.
func (p *Producer) Close() error { /* implementation */ }

// Healthy returns true if the producer can reach brokers.
func (p *Producer) Healthy(ctx context.Context) bool { /* implementation */ }
```

**File: `consumer.go`**

```go
// CLASSIFICATION: UNCLASSIFIED
package redpanda

import (
    "context"
    "github.com/twmb/franz-go/pkg/kgo"
)

// MessageHandler processes a consumed message.
// Return nil to commit, return error to retry (with backoff).
type MessageHandler func(ctx context.Context, record *kgo.Record) error

// Consumer wraps franz-go client for consuming messages from Redpanda.
type Consumer struct {
    client  *kgo.Client
    handler MessageHandler
    topics  []string
    group   string
}

// ConsumerConfig configures a new Consumer.
type ConsumerConfig struct {
    Connection    ConnectionOptions
    Topics        []string
    ConsumerGroup string
    Handler       MessageHandler
    // StartOffset: "earliest" (default) or "latest"
    StartOffset string
    // MaxPollRecords per batch (default: 100)
    MaxPollRecords int
    // SessionTimeout in milliseconds (default: 30000)
    SessionTimeout int
}

// NewConsumer creates a Redpanda consumer.
// Returns error if config is invalid or connection fails.
func NewConsumer(ctx context.Context, cfg ConsumerConfig) (*Consumer, error) { /* implementation */ }

// Start begins consuming in a loop. Blocks until context is cancelled.
// Calls handler for each message. Commits offsets after successful handling.
// On handler error: logs error, produces to DLQ if configured, commits to skip.
func (c *Consumer) Start(ctx context.Context) error { /* implementation */ }

// Close stops consuming and commits final offsets.
func (c *Consumer) Close() error { /* implementation */ }

// Healthy returns true if the consumer is connected and consuming.
func (c *Consumer) Healthy(ctx context.Context) bool { /* implementation */ }
```

**Test Cases** (`producer_test.go`, `consumer_test.go`, `headers_test.go`):

| #   | Test                                     | Scope                        | Expected                               |
| --- | ---------------------------------------- | ---------------------------- | -------------------------------------- |
| T01 | StandardHeaders sets all 5 headers       | Unit                         | All headers present with correct keys  |
| T02 | GetHeader returns value for existing key | Unit                         | Correct value                          |
| T03 | GetHeader returns "" for missing key     | Unit                         | Empty string                           |
| T04 | Producer.Produce with valid message      | Integration (testcontainers) | Message appears in topic               |
| T05 | Producer.Produce with headers            | Integration                  | Headers readable by consumer           |
| T06 | Consumer.Start processes messages        | Integration                  | Handler called for each message        |
| T07 | Consumer offset commit                   | Integration                  | Messages not redelivered after restart |
| T08 | Producer.Close flushes pending           | Integration                  | No message loss                        |
| T09 | Consumer handler error → DLQ             | Integration                  | Failed message in DLQ topic            |
| T10 | TLS connection with mTLS certs           | Integration                  | Successful TLS handshake               |

---

### 4.3 `pkg/health` — gRPC Health Check Server

**File: `checker.go`**

```go
// CLASSIFICATION: UNCLASSIFIED
package health

import "sync"

// Status represents the health status of a component.
type Status int

const (
    StatusUnknown    Status = 0
    StatusServing    Status = 1
    StatusNotServing Status = 2
)

// Checker maintains the health status of registered components.
type Checker struct {
    mu     sync.RWMutex
    checks map[string]Status
}

// NewChecker creates a new Checker with no registered components.
func NewChecker() *Checker { /* implementation */ }

// Register adds a component to health tracking with initial StatusUnknown.
func (c *Checker) Register(name string) { /* implementation */ }

// SetStatus updates the health status of a named component.
func (c *Checker) SetStatus(name string, status Status) { /* implementation */ }

// Overall returns StatusServing if ALL components are Serving,
// StatusNotServing if ANY component is NotServing,
// StatusUnknown if ANY component is Unknown and none are NotServing.
func (c *Checker) Overall() Status { /* implementation */ }

// ComponentStatus returns the status of a single component.
func (c *Checker) ComponentStatus(name string) Status { /* implementation */ }
```

**File: `server.go`**

```go
// CLASSIFICATION: UNCLASSIFIED
package health

import (
    "context"
    commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
)

// Server implements the gRPC HealthService from common.v1.
type Server struct {
    commonv1.UnimplementedHealthServiceServer
    checker *Checker
}

// NewServer creates a health gRPC server backed by the given Checker.
func NewServer(checker *Checker) *Server { /* implementation */ }

// Check implements HealthService.Check.
func (s *Server) Check(ctx context.Context, req *commonv1.HealthCheckRequest) (*commonv1.HealthCheckResponse, error) { /* implementation */ }

// Watch implements HealthService.Watch (server-streaming).
// Sends status updates every 5 seconds.
func (s *Server) Watch(req *commonv1.HealthCheckRequest, stream commonv1.HealthService_WatchServer) error { /* implementation */ }
```

**Test Cases**:

| #   | Test                                            | Expected                |
| --- | ----------------------------------------------- | ----------------------- |
| T01 | Register + SetStatus serving → Overall serving  | StatusServing           |
| T02 | One component not serving → Overall not serving | StatusNotServing        |
| T03 | All unknown → Overall unknown                   | StatusUnknown           |
| T04 | gRPC Check with serving checker                 | SERVING response        |
| T05 | gRPC Check with specific service name           | Component-level status  |
| T06 | gRPC Watch sends periodic updates               | Receives stream updates |

---

### 4.4 `pkg/shutdown` — Graceful Shutdown Orchestrator

**File: `manager.go`**

```go
// CLASSIFICATION: UNCLASSIFIED
package shutdown

import (
    "context"
    "os"
    "os/signal"
    "sync"
    "syscall"
    "time"
    "go.uber.org/zap"
)

// ShutdownFunc is a function called during shutdown.
// It receives a context with timeout and should return quickly.
type ShutdownFunc func(ctx context.Context) error

// Manager orchestrates graceful shutdown of service components.
// Shutdown order is LIFO (last registered = first shutdown).
type Manager struct {
    logger   *zap.Logger
    timeout  time.Duration // Total shutdown timeout (default: 30s)
    hooks    []namedHook
    mu       sync.Mutex
}

type namedHook struct {
    name string
    fn   ShutdownFunc
}

// NewManager creates a shutdown manager with the given timeout.
// Listens for SIGINT and SIGTERM.
func NewManager(logger *zap.Logger, timeout time.Duration) *Manager { /* implementation */ }

// Register adds a shutdown hook. Hooks run in reverse order (LIFO).
// name: human-readable name for logging
func (m *Manager) Register(name string, fn ShutdownFunc) { /* implementation */ }

// Wait blocks until a shutdown signal is received, then executes all hooks.
// Returns the first error encountered, or nil if all hooks succeed.
// Logs each hook's execution and duration.
func (m *Manager) Wait() error { /* implementation */ }

// Trigger programmatically initiates shutdown (e.g., on fatal error).
func (m *Manager) Trigger() { /* implementation */ }
```

**Test Cases**:

| #   | Test                               | Expected                         |
| --- | ---------------------------------- | -------------------------------- |
| T01 | Hooks execute in LIFO order        | 3rd registered runs first        |
| T02 | Hook timeout enforced              | Slow hook cancelled at deadline  |
| T03 | All hooks called even if one fails | All called, first error returned |
| T04 | Trigger programmatic shutdown      | Hooks execute without OS signal  |

---

### 4.5 `pkg/classification` — Classification Level Enforcement

**File: `guard.go`**

```go
// CLASSIFICATION: UNCLASSIFIED
package classification

import (
    commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
)

// LevelOrder maps ClassificationLevel to numeric order for comparison.
// Higher number = higher classification.
var LevelOrder = map[commonv1.ClassificationLevel]int{
    commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNSPECIFIED:   0,
    commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED:  1,
    commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_A:   2,
    commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B:   3,
    commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_C:   4,
    commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET:        5,
}

// Guard enforces classification level ceiling for a service.
type Guard struct {
    maxLevel commonv1.ClassificationLevel
}

// NewGuard creates a Guard with the specified maximum classification ceiling.
func NewGuard(maxLevel commonv1.ClassificationLevel) *Guard { /* implementation */ }

// Check returns nil if dataLevel ≤ maxLevel, error otherwise.
// Error: "classification: data level %s exceeds service ceiling %s"
func (g *Guard) Check(dataLevel commonv1.ClassificationLevel) error { /* implementation */ }

// CanAccess returns true if callerClearance ≥ dataLevel.
func CanAccess(callerClearance, dataLevel commonv1.ClassificationLevel) bool { /* implementation */ }

// Max returns the higher of two classification levels.
func Max(a, b commonv1.ClassificationLevel) commonv1.ClassificationLevel { /* implementation */ }

// MaxAll returns the highest classification level among all provided.
func MaxAll(levels ...commonv1.ClassificationLevel) commonv1.ClassificationLevel { /* implementation */ }

// LevelToString converts a ClassificationLevel to its string representation.
// e.g., CLASSIFICATION_LEVEL_SECRET → "SECRET"
func LevelToString(level commonv1.ClassificationLevel) string { /* implementation */ }

// StringToLevel converts a string to ClassificationLevel.
// Returns UNSPECIFIED for unknown strings.
func StringToLevel(s string) commonv1.ClassificationLevel { /* implementation */ }
```

**Test Cases**:

| #   | Test                                  | Expected                       |
| --- | ------------------------------------- | ------------------------------ |
| T01 | Guard.Check within ceiling            | nil                            |
| T02 | Guard.Check above ceiling             | Error with descriptive message |
| T03 | CanAccess with sufficient clearance   | true                           |
| T04 | CanAccess with insufficient clearance | false                          |
| T05 | Max of PROTECTED_B and SECRET         | SECRET                         |
| T06 | MaxAll of mixed levels                | Highest returned               |
| T07 | LevelToString roundtrip               | String ↔ Level consistent      |
| T08 | StringToLevel with unknown string     | UNSPECIFIED                    |

---

### 4.6 `pkg/audit` — Audit Event Builder & Producer

**File: `types.go`**

```go
// CLASSIFICATION: UNCLASSIFIED
package audit

// Standard event type constants.
const (
    EventTrackCreated     = "track.created"
    EventTrackUpdated     = "track.updated"
    EventTrackMerged      = "track.merged"
    EventTrackDropped     = "track.dropped"
    EventAlertGenerated   = "alert.generated"
    EventAlertAcknowledged = "alert.acknowledged"
    EventFeedbackSubmitted = "feedback.submitted"
    EventFeedbackValidated = "feedback.validated"
    EventFeedbackRejected  = "feedback.rejected"
    EventQueryExecuted    = "query.executed"
    EventModelPublished   = "model.published"
    EventModelRolledBack  = "model.rolled_back"
    EventSensorConnected  = "sensor.connected"
    EventSensorDisconnected = "sensor.disconnected"
    EventNATOExport       = "nato.export"
    EventNATOImport       = "nato.import"
    EventClassificationViolation = "classification.violation"
)

// AuditTopic is the Redpanda topic for audit events.
const AuditTopic = "audit.events"
```

**File: `emitter.go`**

```go
// CLASSIFICATION: UNCLASSIFIED
package audit

import (
    "context"
    "encoding/json"
    "time"

    "github.com/google/uuid"
    auditv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/audit/v1"
    commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
    "github.com/arvinddhasmana/rtsa_webgpu/pkg/redpanda"
    "go.uber.org/zap"
    "google.golang.org/protobuf/proto"
    "google.golang.org/protobuf/types/known/timestamppb"
)

// Emitter produces audit events to Redpanda.
type Emitter struct {
    producer    *redpanda.Producer
    serviceID   string
    logger      *zap.Logger
}

// NewEmitter creates an audit event emitter for the given service.
func NewEmitter(producer *redpanda.Producer, serviceID string, logger *zap.Logger) *Emitter { /* implementation */ }

// Emit produces an audit event. Never returns an error to the caller —
// audit failures are logged but do not block business logic.
// This follows the principle that audit should not impede operations.
func (e *Emitter) Emit(ctx context.Context, event AuditParams) {
    // Build AuditEvent proto
    // Serialize to bytes
    // Produce to AuditTopic with key = serviceID
    // Log error if produce fails, but do NOT return it
}

// AuditParams contains the parameters for an audit event.
type AuditParams struct {
    EventType          string
    ActorID            string
    ActorType          auditv1.ActorType
    ResourceType       string
    ResourceID         string
    Action             auditv1.AuditAction
    Detail             map[string]interface{} // Serialized to JSON
    ClassificationLevel commonv1.ClassificationLevel
}
```

**Test Cases**:

| #   | Test                                         | Expected                                   |
| --- | -------------------------------------------- | ------------------------------------------ |
| T01 | Emit produces valid protobuf to audit.events | Message in topic, deserializable           |
| T02 | Emit with nil detail                         | detail_json = "{}"                         |
| T03 | Emit with producer error                     | Error logged, no panic, caller not blocked |
| T04 | AuditEvent has UUID v7 format                | Valid UUID                                 |
| T05 | AuditEvent has correct service_id            | Matches constructor param                  |
| T06 | Emit sets trace_id from context              | Trace ID propagated                        |

---

### 4.7 `pkg/telemetry` — OpenTelemetry Initialization

**File: `provider.go`**

```go
// CLASSIFICATION: UNCLASSIFIED
package telemetry

import (
    "context"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/sdk/trace"
    sdkmetric "go.opentelemetry.io/otel/sdk/metric"
    "go.uber.org/zap"
)

// Config holds telemetry configuration.
type Config struct {
    ServiceName    string
    ServiceVersion string
    Environment    string
    OTelEndpoint   string // OTLP gRPC endpoint (default: localhost:4317)
    // MetricsPort for Prometheus scrape endpoint (default: 9090)
    MetricsPort int
}

// Provider holds initialized telemetry providers.
type Provider struct {
    TracerProvider *trace.TracerProvider
    MeterProvider  *sdkmetric.MeterProvider
    Logger         *zap.Logger
}

// Init initializes OpenTelemetry tracing, metrics, and structured logging.
// Sets global tracer/meter providers.
// Returns Provider with Shutdown method.
func Init(ctx context.Context, cfg Config) (*Provider, error) { /* implementation */ }

// Shutdown gracefully shuts down all telemetry providers.
// Flushes pending spans and metrics.
func (p *Provider) Shutdown(ctx context.Context) error { /* implementation */ }
```

**File: `attributes.go`**

```go
// CLASSIFICATION: UNCLASSIFIED
package telemetry

import "go.opentelemetry.io/otel/attribute"

// Standard metric attribute keys used across all services.
var (
    AttrServiceName     = attribute.Key("service.name")
    AttrSensorType      = attribute.Key("sensor.type")
    AttrEntityType      = attribute.Key("entity.type")
    AttrClassification  = attribute.Key("classification.level")
    AttrAnomalyType     = attribute.Key("anomaly.type")
    AttrAlertSeverity   = attribute.Key("alert.severity")
    AttrFeedbackType    = attribute.Key("feedback.type")
    AttrOperatorID      = attribute.Key("operator.id")
    AttrTrackStatus     = attribute.Key("track.status")
    AttrHostileClass    = attribute.Key("hostile.classification")
    AttrTopicName       = attribute.Key("redpanda.topic")
    AttrConsumerGroup   = attribute.Key("redpanda.consumer_group")
    AttrGRPCMethod      = attribute.Key("grpc.method")
    AttrGRPCStatusCode  = attribute.Key("grpc.status_code")
)
```

**Test Cases**:

| #   | Test                              | Expected                           |
| --- | --------------------------------- | ---------------------------------- |
| T01 | Init creates valid TracerProvider | Non-nil TracerProvider             |
| T02 | Init creates valid MeterProvider  | Non-nil MeterProvider              |
| T03 | Init sets global tracer           | `otel.GetTracerProvider()` matches |
| T04 | Shutdown flushes spans            | No pending spans after shutdown    |
| T05 | Logger is structured JSON         | Log output parseable as JSON       |

---

### 4.8 `pkg/interceptors` — gRPC Interceptor Chain

**File: `chain.go`**

```go
// CLASSIFICATION: UNCLASSIFIED
package interceptors

import (
    "google.golang.org/grpc"
    "go.uber.org/zap"
    "github.com/arvinddhasmana/rtsa_webgpu/pkg/classification"
    "github.com/arvinddhasmana/rtsa_webgpu/pkg/telemetry"
)

// ChainConfig configures the standard interceptor chain.
type ChainConfig struct {
    Logger             *zap.Logger
    TelemetryProvider  *telemetry.Provider
    ClassificationGuard *classification.Guard
    ServiceName        string
}

// BuildUnaryServerInterceptors returns the standard unary interceptor chain.
// Order: Recovery → Tracing → Metrics → Classification → Logging
func BuildUnaryServerInterceptors(cfg ChainConfig) []grpc.UnaryServerInterceptor { /* implementation */ }

// BuildStreamServerInterceptors returns the standard stream interceptor chain.
// Order: Recovery → Tracing → Metrics → Classification → Logging
func BuildStreamServerInterceptors(cfg ChainConfig) []grpc.StreamServerInterceptor { /* implementation */ }

// BuildDialOptions returns standard client dial options with interceptors.
func BuildDialOptions(cfg ChainConfig) []grpc.DialOption { /* implementation */ }
```

**File: `logging.go`**

```go
// CLASSIFICATION: UNCLASSIFIED
package interceptors

import (
    "context"
    "time"
    "go.uber.org/zap"
    "google.golang.org/grpc"
)

// UnaryLoggingInterceptor logs gRPC unary call details.
// Log fields: method, duration_ms, status_code, error (if any)
// Log level: INFO for success, ERROR for failures
// NEVER logs request/response payloads (may contain classified data)
func UnaryLoggingInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor { /* implementation */ }

// StreamLoggingInterceptor logs gRPC stream lifecycle.
// Log fields: method, stream_duration_ms, status_code
func StreamLoggingInterceptor(logger *zap.Logger) grpc.StreamServerInterceptor { /* implementation */ }
```

**File: `metrics.go`**

```go
// CLASSIFICATION: UNCLASSIFIED
package interceptors

import (
    "google.golang.org/grpc"
    "go.opentelemetry.io/otel/metric"
)

// Standard gRPC metric names
const (
    MetricGRPCRequestTotal     = "rtsa_grpc_requests_total"
    MetricGRPCRequestDuration  = "rtsa_grpc_request_duration_seconds"
    MetricGRPCActiveStreams     = "rtsa_grpc_active_streams"
)

// UnaryMetricsInterceptor records request count and duration histograms.
// Labels: method, status_code, service
func UnaryMetricsInterceptor(meter metric.Meter, serviceName string) grpc.UnaryServerInterceptor { /* implementation */ }

// StreamMetricsInterceptor records stream lifecycle metrics.
func StreamMetricsInterceptor(meter metric.Meter, serviceName string) grpc.StreamServerInterceptor { /* implementation */ }
```

**File: `tracing.go`**

```go
// CLASSIFICATION: UNCLASSIFIED
package interceptors

import (
    "google.golang.org/grpc"
    "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
)

// UnaryTracingInterceptor adds distributed tracing to unary calls.
// Uses OpenTelemetry gRPC instrumentation.
func UnaryTracingInterceptor() grpc.UnaryServerInterceptor { /* implementation */ }

// StreamTracingInterceptor adds distributed tracing to stream calls.
func StreamTracingInterceptor() grpc.StreamServerInterceptor { /* implementation */ }
```

**File: `classification.go`**

```go
// CLASSIFICATION: UNCLASSIFIED
package interceptors

import (
    "context"
    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/metadata"
    "google.golang.org/grpc/status"
    "github.com/arvinddhasmana/rtsa_webgpu/pkg/classification"
)

// UnaryClassificationInterceptor checks that the incoming request's
// classification level does not exceed the service's ceiling.
// Reads "rtsa-classification" from gRPC metadata.
// Returns PERMISSION_DENIED if classification exceeds ceiling.
func UnaryClassificationInterceptor(guard *classification.Guard) grpc.UnaryServerInterceptor { /* implementation */ }

// StreamClassificationInterceptor checks classification for streaming RPCs.
func StreamClassificationInterceptor(guard *classification.Guard) grpc.StreamServerInterceptor { /* implementation */ }
```

**File: `recovery.go`**

```go
// CLASSIFICATION: UNCLASSIFIED
package interceptors

import (
    "google.golang.org/grpc"
    "go.uber.org/zap"
)

// UnaryRecoveryInterceptor catches panics in handlers and converts them to
// gRPC INTERNAL errors. Logs the panic stack trace at ERROR level.
// This is the outermost interceptor.
func UnaryRecoveryInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor { /* implementation */ }

// StreamRecoveryInterceptor catches panics in stream handlers.
func StreamRecoveryInterceptor(logger *zap.Logger) grpc.StreamServerInterceptor { /* implementation */ }
```

**Test Cases** (`interceptors_test.go`):

| #   | Test                                    | Expected                          |
| --- | --------------------------------------- | --------------------------------- |
| T01 | Logging interceptor logs method name    | "method" field in log             |
| T02 | Logging interceptor never logs payload  | No "request" or "response" in log |
| T03 | Metrics interceptor increments counter  | Counter +1 after call             |
| T04 | Metrics interceptor records duration    | Histogram has observation         |
| T05 | Classification interceptor allows valid | Request succeeds                  |
| T06 | Classification interceptor rejects high | PERMISSION_DENIED status          |
| T07 | Recovery interceptor catches panic      | INTERNAL error, not process crash |
| T08 | Full chain processes request            | All interceptors invoked in order |

---

### 4.9 `pkg/testutil` — Test Utilities

**File: `redpanda.go`**

```go
// CLASSIFICATION: UNCLASSIFIED
package testutil

import (
    "context"
    "testing"
    "github.com/testcontainers/testcontainers-go"
)

// StartRedpanda starts a Redpanda container for integration tests.
// Returns broker address (e.g., "localhost:19092") and cleanup function.
// Uses testcontainers-go with Redpanda 24.x image.
func StartRedpanda(t *testing.T) (brokers string, cleanup func()) { /* implementation */ }

// CreateTopics creates the specified topics in the test Redpanda instance.
func CreateTopics(t *testing.T, brokers string, topics ...string) { /* implementation */ }
```

**File: `clickhouse.go`**

```go
// CLASSIFICATION: UNCLASSIFIED
package testutil

import (
    "testing"
)

// StartClickHouse starts a ClickHouse container for integration tests.
// Returns DSN and cleanup function.
func StartClickHouse(t *testing.T) (dsn string, cleanup func()) { /* implementation */ }

// ApplySchema applies the RTSA ClickHouse schema from deploy/clickhouse/init/*.sql
func ApplySchema(t *testing.T, dsn string) { /* implementation */ }
```

**File: `grpc.go`**

```go
// CLASSIFICATION: UNCLASSIFIED
package testutil

import (
    "net"
    "testing"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

// StartTestGRPCServer starts an in-process gRPC server for testing.
// Returns *grpc.Server, listener address, and cleanup function.
// Uses insecure credentials for tests (no TLS).
func StartTestGRPCServer(t *testing.T, registerFn func(s *grpc.Server)) (addr string, cleanup func()) { /* implementation */ }

// DialTestGRPC creates a client connection to a test gRPC server.
func DialTestGRPC(t *testing.T, addr string) *grpc.ClientConn { /* implementation */ }
```

**File: `proto.go`**

```go
// CLASSIFICATION: UNCLASSIFIED
package testutil

import (
    commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
    ingestionv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/ingestion/v1"
    entityv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/entity/v1"
    inferencev1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/inference/v1"
    feedbackv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/feedback/v1"
    "google.golang.org/protobuf/types/known/timestamppb"
    "time"
)

// ── Test data coordinates: Mid-Atlantic 43-47°N, 55-65°W (UNCLASSIFIED synthetic only) ──

// NewTestRadarObservation builds a valid RadarTrack SensorObservation for testing.
// Position: ~45.0°N, -60.0°W (Mid-Atlantic, synthetic).
func NewTestRadarObservation(sensorID string) *ingestionv1.SensorObservation { /* returns fully populated observation */ }

// NewTestFusedTrack builds a valid FusedTrack for testing.
func NewTestFusedTrack(trackID string) *entityv1.FusedTrack { /* returns fully populated track */ }

// NewTestAnomalyAlert builds a valid AnomalyAlert for testing.
func NewTestAnomalyAlert(alertID, trackID string) *inferencev1.AnomalyAlert { /* returns fully populated alert */ }

// NewTestPosition builds a Position at the given coordinates.
func NewTestPosition(lat, lon float64) *commonv1.Position { /* implementation */ }
```

---

## 5. Integration Test Infrastructure

### 5.1 `testcontainers-go` Usage

All integration tests in `pkg/` that need Redpanda or ClickHouse must use `testcontainers-go` to start ephemeral containers. Tests must:

1. Check for `RTSA_INTEGRATION_TESTS=true` environment variable
2. Skip if not set: `if os.Getenv("RTSA_INTEGRATION_TESTS") != "true" { t.Skip("integration tests disabled") }`
3. Use `t.Cleanup()` for container teardown
4. Use unique consumer group names per test to avoid offset conflicts

### 5.2 Build Tags

```go
//go:build integration

package redpanda_test
```

Integration tests use the `integration` build tag. Run with:

```bash
go test -tags integration -count=1 ./pkg/...
```

---

## 6. Agent Invocation

```
@greatest-ever-developer Implement Module 03 from docs/implementation/03-shared-go-libraries.md

Context:
- Read docs/implementation/00-implementation-overview.md for global conventions
- Read docs/implementation/02-protobuf-schemas.md for generated proto types
- This module creates ALL shared Go packages under pkg/
- Each package must have ≥80% line coverage
- Use franz-go for Redpanda, OpenTelemetry for observability, zap for logging
- All errors must use fmt.Errorf("context: %w", err) pattern
- No panic() in non-test code
- Integration tests use testcontainers-go behind RTSA_INTEGRATION_TESTS=true

Deliverables:
1. pkg/go.mod with all dependencies
2. All package source files as specified
3. Unit tests for every package (≥80% coverage)
4. Integration tests for redpanda package (behind build tag)
5. go vet ./pkg/... passes
6. go test ./pkg/... passes (unit tests)
```
