<!-- CLASSIFICATION: UNCLASSIFIED -->

# V1 — Deferred Capability Stubs

> **Version**: 1.0
> **Category**: Deferred Capabilities — NATO Interoperability & Training Pipeline
> **Priority**: P3 — Required for v1 completeness; noop stubs only
> **Depends On**: `01-infrastructure-fixes.md` (topics must exist)
> **Agent**: `@greatest-ever-developer`

---

## Purpose

Several requirements and use cases reference capabilities that are not yet stubbed in the
codebase. For the v1 milestone these must exist as **noop services** — structurally
complete, buildable, testable, wired into Docker Compose, but performing no real work
beyond logging and returning success. This satisfies requirement traceability and lets
integration / E2E tests exercise the full topic graph without blocking on algorithm
implementation.

### Traceability

| Requirement               | Feature | Use Case     | What This File Covers       |
| ------------------------- | ------- | ------------ | --------------------------- |
| CR-NATO-001 … CR-NATO-005 | FEAT-15 | UC011        | NATO adapter noop stub      |
| CR-FB-003, CR-FB-004      | FEAT-12 | UC014, UC015 | Training pipeline noop stub |

---

## Task STUB-01 — NATO Adapter Noop Service (`svc-nato-adapter`)

### Context

Requirements CR-NATO-001 through CR-NATO-005 mandate STANAG 5516 / NFFI / MIP
interoperability. For v1, create a structurally complete service that accepts
`ExportTracks` and `ImportTracks` RPCs but returns a hard-coded success with an empty
payload. This lets E2E tests confirm the service is reachable and the gRPC reflection
endpoint works.

### Proto Definition

Create `proto/rtsa/nato/v1/nato_service.proto`:

```protobuf
// CLASSIFICATION: UNCLASSIFIED
syntax = "proto3";

package rtsa.nato.v1;

option go_package = "github.com/vertexcover-io/rtsa/gen/go/rtsa/nato/v1;natov1";

import "google/protobuf/timestamp.proto";

// NatoAdapterService — noop stub for STANAG 5516 / NFFI / MIP interoperability.
// All RPCs log the request and return success with empty payloads.
service NatoAdapterService {
  // ExportTracks sends fused tracks to a NATO partner in STANAG 5516 format.
  rpc ExportTracks(ExportTracksRequest) returns (ExportTracksResponse);

  // ImportTracks receives tracks from a NATO partner feed.
  rpc ImportTracks(ImportTracksRequest) returns (ImportTracksResponse);
}

message ExportTracksRequest {
  repeated string track_ids = 1;
  string destination_partner = 2;  // e.g. "NATO_PARTNER_ALPHA"
  google.protobuf.Timestamp as_of = 3;
}

message ExportTracksResponse {
  bool accepted = 1;
  string message = 2;  // human-readable status
}

message ImportTracksRequest {
  bytes raw_payload = 1;           // opaque STANAG / NFFI blob — ignored in noop
  string source_partner = 2;
  google.protobuf.Timestamp received_at = 3;
}

message ImportTracksResponse {
  int32 tracks_imported = 1;       // always 0 in noop
  string message = 2;
}
```

After creating the proto file, run `buf generate` to produce Go and TypeScript stubs.

### Service Scaffold

Create the service under `svc-nato-adapter/` following the existing service layout:

```
svc-nato-adapter/
├── Dockerfile
├── go.mod
├── README.md
├── cmd/
│   └── server/
│       └── main.go
└── internal/
    ├── config/
    │   └── config.go
    ├── handler/
    │   ├── handler.go
    │   └── handler_test.go
    └── server/
        └── server.go
```

#### `go.mod`

```
module github.com/vertexcover-io/rtsa/svc-nato-adapter

go 1.24

require (
    github.com/vertexcover-io/rtsa/gen/go   v0.0.0
    github.com/vertexcover-io/rtsa/pkg       v0.0.0
    google.golang.org/grpc                   latest
    go.uber.org/zap                          latest
)

replace (
    github.com/vertexcover-io/rtsa/gen/go => ../gen/go
    github.com/vertexcover-io/rtsa/pkg    => ../pkg
)
```

Add `./svc-nato-adapter` to `go.work`.

#### `cmd/server/main.go` Pattern

Follow `svc-track/cmd/server/main.go` exactly — load config, init telemetry, create
gRPC server with interceptors (audit, classification, logging), register the noop
handler, start health server, wire graceful shutdown. **Do NOT use `panic()`.**

#### `internal/handler/handler.go`

```go
// CLASSIFICATION: UNCLASSIFIED
package handler

import (
    "context"
    natov1 "github.com/vertexcover-io/rtsa/gen/go/rtsa/nato/v1"
    "go.uber.org/zap"
)

// NatoAdapterHandler is a noop implementation of NatoAdapterServiceServer.
type NatoAdapterHandler struct {
    natov1.UnimplementedNatoAdapterServiceServer
    logger *zap.Logger
}

func New(logger *zap.Logger) *NatoAdapterHandler {
    return &NatoAdapterHandler{logger: logger}
}

func (h *NatoAdapterHandler) ExportTracks(
    ctx context.Context,
    req *natov1.ExportTracksRequest,
) (*natov1.ExportTracksResponse, error) {
    h.logger.Info("noop ExportTracks called",
        zap.Int("track_count", len(req.GetTrackIds())),
        zap.String("destination", req.GetDestinationPartner()),
    )
    return &natov1.ExportTracksResponse{
        Accepted: true,
        Message:  "noop — export acknowledged but not transmitted",
    }, nil
}

func (h *NatoAdapterHandler) ImportTracks(
    ctx context.Context,
    req *natov1.ImportTracksRequest,
) (*natov1.ImportTracksResponse, error) {
    h.logger.Info("noop ImportTracks called",
        zap.Int("payload_bytes", len(req.GetRawPayload())),
        zap.String("source", req.GetSourcePartner()),
    )
    return &natov1.ImportTracksResponse{
        TracksImported: 0,
        Message:        "noop — import acknowledged but not processed",
    }, nil
}
```

#### `internal/handler/handler_test.go`

Unit test both RPCs — call each with a representative request, assert `err == nil`,
assert the response message contains `"noop"`, assert `TracksImported == 0` for import,
assert `Accepted == true` for export. Use `zap.NewNop()` for the logger.

#### `internal/config/config.go`

Follow `svc-track/internal/config/config.go` pattern — `GRPCPort`, `HealthPort`,
`OTelEndpoint`, `Environment`. Default gRPC port: `50074`.

#### Dockerfile

Copy `svc-track/Dockerfile`, change the module path.

### Docker Compose Entry

Add to `deploy/docker-compose.services.yml`:

```yaml
svc-nato-adapter:
  build:
    context: ..
    dockerfile: svc-nato-adapter/Dockerfile
  environment:
    GRPC_PORT: "50051"
    HEALTH_PORT: "8080"
    OTEL_EXPORTER_OTLP_ENDPOINT: "otel-collector:4317"
    ENVIRONMENT: "development"
  ports:
    - "50074:50051"
  depends_on:
    redpanda:
      condition: service_healthy
  networks:
    - rtsa-net
  restart: unless-stopped
```

Add an Envoy cluster + route for `rtsa.nato.v1.NatoAdapterService` in
`deploy/envoy/envoy-dev.yaml` mapping to `svc-nato-adapter:50051`.

### Acceptance Criteria

- [ ] `buf lint` passes for the new proto
- [ ] `buf generate` succeeds and produces Go + TS outputs
- [ ] `go build ./svc-nato-adapter/...` succeeds
- [ ] Unit tests pass: `go test ./svc-nato-adapter/...`
- [ ] `docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.services.yml up svc-nato-adapter` starts and passes health check
- [ ] `grpcurl -plaintext localhost:50074 rtsa.nato.v1.NatoAdapterService/ExportTracks` returns `accepted: true`
- [ ] Classification header present in every generated file

---

## Task STUB-02 — Training Pipeline Noop Service (`svc-training`)

### Context

Requirements CR-FB-003 and CR-FB-004 describe a feedback loop where validated operator
feedback is used to retrain anomaly-detection models. For v1, create a noop consumer
that reads from `feedback.operator.validated`, logs each message, and produces a noop
acknowledgement to `models.anomaly.candidates`. No real ML training occurs.

### Service Scaffold

```
svc-training/
├── Dockerfile
├── go.mod
├── README.md
├── cmd/
│   └── server/
│       └── main.go
└── internal/
    ├── config/
    │   └── config.go
    ├── consumer/
    │   ├── consumer.go
    │   └── consumer_test.go
    └── producer/
        ├── producer.go
        └── producer_test.go
```

#### Topic Configuration

| Direction | Topic                         | Format                                                       |
| --------- | ----------------------------- | ------------------------------------------------------------ |
| Consume   | `feedback.operator.validated` | Protobuf `FeedbackEvent`                                     |
| Produce   | `models.anomaly.candidates`   | JSON `{"model_id":"noop","status":"stub","timestamp":"..."}` |

Ensure both topics are created in `scripts/dev/init-topics.sh` (after BUG-02 fix in
`01-infrastructure-fixes.md`).

#### `go.mod`

```
module github.com/vertexcover-io/rtsa/svc-training

go 1.24

require (
    github.com/vertexcover-io/rtsa/gen/go   v0.0.0
    github.com/vertexcover-io/rtsa/pkg       v0.0.0
    github.com/twmb/franz-go                 latest
    go.uber.org/zap                          latest
)

replace (
    github.com/vertexcover-io/rtsa/gen/go => ../gen/go
    github.com/vertexcover-io/rtsa/pkg    => ../pkg
)
```

Add `./svc-training` to `go.work`.

#### `cmd/server/main.go` Pattern

Follow `svc-anomaly-detection/cmd/server/main.go` — it is also a consume→produce
service. Key steps:

1. Load config via `pkg/config`
2. Init telemetry via `pkg/telemetry`
3. Create consumer using `pkg/redpanda.NewConsumer` for `feedback.operator.validated`
4. Create producer using `pkg/redpanda.NewProducer` for `models.anomaly.candidates`
5. Start health server via `pkg/health`
6. In consume loop: deserialise `FeedbackEvent`, log it, produce noop JSON to output topic
7. Wire `pkg/shutdown.Wait()` for graceful shutdown

**No `panic()`. Wrap all errors with `fmt.Errorf("context: %w", err)`.**

#### `internal/consumer/consumer.go`

```go
// CLASSIFICATION: UNCLASSIFIED
package consumer

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/twmb/franz-go/pkg/kgo"
    "go.uber.org/zap"
)

// NoopModelCandidate is the JSON payload produced to models.anomaly.candidates.
type NoopModelCandidate struct {
    ModelID   string    `json:"model_id"`
    Status    string    `json:"status"`
    Timestamp time.Time `json:"timestamp"`
}

// TrainingConsumer reads validated feedback and produces noop model candidates.
type TrainingConsumer struct {
    consumer *kgo.Client
    producer *kgo.Client
    outTopic string
    logger   *zap.Logger
}

func New(consumer, producer *kgo.Client, outTopic string, logger *zap.Logger) *TrainingConsumer {
    return &TrainingConsumer{
        consumer: consumer,
        producer: producer,
        outTopic: outTopic,
        logger:   logger,
    }
}

func (tc *TrainingConsumer) Run(ctx context.Context) error {
    for {
        fetches := tc.consumer.PollFetches(ctx)
        if errs := fetches.Errors(); len(errs) > 0 {
            for _, e := range errs {
                if e.Err == context.Canceled {
                    return nil
                }
                tc.logger.Error("fetch error", zap.Error(e.Err))
            }
            continue
        }

        iter := fetches.RecordIter()
        for !iter.Done() {
            record := iter.Next()
            tc.logger.Info("noop training — received feedback",
                zap.String("topic", record.Topic),
                zap.Int64("offset", record.Offset),
                zap.Int("payload_bytes", len(record.Value)),
            )

            candidate := NoopModelCandidate{
                ModelID:   "noop-v0",
                Status:    "stub",
                Timestamp: time.Now().UTC(),
            }
            payload, err := json.Marshal(candidate)
            if err != nil {
                return fmt.Errorf("marshal noop candidate: %w", err)
            }

            tc.producer.Produce(ctx, &kgo.Record{
                Topic: tc.outTopic,
                Value: payload,
            }, func(_ *kgo.Record, err error) {
                if err != nil {
                    tc.logger.Error("produce noop candidate failed", zap.Error(err))
                }
            })
        }
    }
}
```

#### `internal/consumer/consumer_test.go`

Test `NoopModelCandidate` JSON marshalling. Test that `New` returns a non-nil
`TrainingConsumer`. For integration-level consumer test, use the testcontainers pattern
from `svc-radar-ingestion/internal/producer/producer_test.go`.

#### `internal/config/config.go`

Follow `svc-anomaly-detection/internal/config/config.go` pattern:

```go
// CLASSIFICATION: UNCLASSIFIED
package config

import "github.com/vertexcover-io/rtsa/pkg/config"

type Config struct {
    HealthPort   int      `env:"HEALTH_PORT"   default:"8080"`
    InputTopic   string   `env:"INPUT_TOPIC"   default:"feedback.operator.validated"`
    OutputTopic  string   `env:"OUTPUT_TOPIC"  default:"models.anomaly.candidates"`
    Brokers      []string `env:"REDPANDA_BROKERS" default:"localhost:9092"`
    ConsumerGroup string  `env:"CONSUMER_GROUP" default:"svc-training"`
    Environment  string   `env:"ENVIRONMENT" default:"development"`
    OTelEndpoint string   `env:"OTEL_EXPORTER_OTLP_ENDPOINT" default:"localhost:4317"`
}

func Load() (*Config, error) {
    var cfg Config
    if err := config.Load(&cfg); err != nil {
        return nil, err
    }
    return &cfg, nil
}
```

#### Dockerfile

Copy `svc-anomaly-detection/Dockerfile`, change the module path.

### Docker Compose Entry

Add to `deploy/docker-compose.services.yml`:

```yaml
svc-training:
  build:
    context: ..
    dockerfile: svc-training/Dockerfile
  environment:
    HEALTH_PORT: "8080"
    INPUT_TOPIC: "feedback.operator.validated"
    OUTPUT_TOPIC: "models.anomaly.candidates"
    REDPANDA_BROKERS: "redpanda:9092"
    CONSUMER_GROUP: "svc-training"
    OTEL_EXPORTER_OTLP_ENDPOINT: "otel-collector:4317"
    ENVIRONMENT: "development"
  depends_on:
    redpanda:
      condition: service_healthy
  networks:
    - rtsa-net
  restart: unless-stopped
```

### Init Script Updates

Add the following topics to `scripts/dev/init-topics.sh` (after applying BUG-02 fix):

```bash
rpk topic create models.anomaly.candidates \
    --partitions 1 \
    --replicas 1 \
    --brokers "$BROKERS"
```

> `feedback.operator.validated` should already exist after BUG-02 fix.

### Acceptance Criteria

- [ ] `go build ./svc-training/...` succeeds
- [ ] Unit tests pass: `go test ./svc-training/...`
- [ ] Service consumes from `feedback.operator.validated` and produces JSON to `models.anomaly.candidates`
- [ ] `docker compose … up svc-training` starts, health check passes
- [ ] Logs show `"noop training — received feedback"` when test messages are published
- [ ] Classification header present in every generated file
- [ ] No `panic()` in non-test code

---

## Agent Invocation

```
@greatest-ever-developer Implement all tasks in docs/implementation/v1/04-deferred-capability-stubs.md

Read the full task file first. For each task (STUB-01, STUB-02):
1. Read the referenced pattern files (svc-track/ and svc-anomaly-detection/ layouts).
2. Create the proto definition (STUB-01 only), run buf generate.
3. Scaffold the service directory with all files listed.
4. Implement the noop handler / consumer following the code patterns provided.
5. Write unit tests for every exported function.
6. Add Docker Compose entries and init-script topics.
7. Add the service module to go.work.
8. Run `go build`, `go vet`, `go test` for each new service.
9. Run `buf lint` to validate proto changes.
10. Verify classification headers exist in every new file.
11. Ensure zero uses of panic() outside _test.go files.

Execute tasks in order: STUB-01 then STUB-02.
Commit each task separately with message prefix "v1(stubs):".
```
