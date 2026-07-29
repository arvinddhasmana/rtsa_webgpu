<!-- CLASSIFICATION: UNCLASSIFIED -->

# v1 Module 02 — Ingestion Service Completion

> **Module**: 02-ingestion-service-completion
> **Phase**: P1 (depends on Module 01 — Infrastructure Fixes)
> **Dependencies**: Module 01 (fixed topic names), Module 04 (radar reference), Module 03 (shared libraries)
> **Agent**: `@greatest-ever-developer`
> **Estimated Effort**: 3 days
> **Traceability**: UC003–UC007, FEAT-05–FEAT-09, CR-ING-002–CR-ING-006

---

## 1. Objective

Elevate 5 partial ingestion services to production parity with `svc-radar-ingestion`. Each service currently has **complete domain logic** (validators + normalizers with tests) but uses a minimal `main.go` with `ingestion.LogProducer` (stdout-only), no telemetry, no interceptor chain, no classification guard, no audit emitter, and no health endpoints.

### Current State vs Target State

| Capability           | Current (Partial)                       | Target (Production)                                          |
| -------------------- | --------------------------------------- | ------------------------------------------------------------ |
| Domain validator     | ✅ Complete + tested                    | ✅ No change needed                                          |
| Domain normalizer    | ✅ Complete + tested                    | ✅ No change needed                                          |
| `main.go` wiring     | `LogProducer`, basic `grpc.NewServer()` | Real `redpanda.Producer`, full interceptor chain             |
| `internal/config/`   | Uses `pkg/ingestion.MustLoad()`         | Sensor-specific `config.MustLoad()` with thresholds          |
| `internal/handler/`  | Uses `pkg/ingestion.NewHandler()`       | Sensor-specific handler (or shared `pkg/ingestion.Handler`)  |
| `internal/producer/` | `LogProducer` (stdout)                  | `producer.ObservationProducer` backed by `redpanda.Producer` |
| `internal/mapper/`   | Not present                             | `mapper.Enricher` with classification guard                  |
| Telemetry (OTel)     | ❌ Missing                              | ✅ `pkg/telemetry.Init()`                                    |
| Interceptors         | ❌ Missing                              | ✅ `pkg/interceptors.BuildUnaryServerInterceptors()`         |
| Classification guard | ❌ Missing                              | ✅ `pkg/classification.NewGuard()`                           |
| Audit emitter        | ❌ Missing                              | ✅ `pkg/audit.NewEmitter()`                                  |
| Health endpoints     | ❌ Missing                              | ✅ `pkg/health.NewServer()`                                  |
| Graceful shutdown    | Basic signal handler                    | ✅ `pkg/shutdown.NewManager()`                               |
| Unit tests (handler) | ❌ Missing                              | ✅ Handler + producer + enricher tests                       |
| Integration test     | ❌ Missing (AIS has one)                | ✅ Per-service integration test                              |

---

## 2. Services to Elevate

| #   | Service                | Module Path            | Output Topic               | DLQ Topic           | gRPC Port | Use Case | Feature |
| --- | ---------------------- | ---------------------- | -------------------------- | ------------------- | --------- | -------- | ------- |
| 1   | EW/SIGINT Ingestion    | `svc-ew-ingestion/`    | `sensors.ew.intercepts`    | `dlq.sensors.ew`    | 50052     | UC003    | FEAT-05 |
| 2   | ELINT/COMINT Ingestion | `svc-elint-ingestion/` | `sensors.elint.detections` | `dlq.sensors.elint` | 50053     | UC004    | FEAT-06 |
| 3   | ISR Metadata Ingestion | `svc-isr-ingestion/`   | `sensors.isr.observations` | `dlq.sensors.isr`   | 50054     | UC005    | FEAT-07 |
| 4   | AIS/BFT Ingestion      | `svc-ais-ingestion/`   | `sensors.ais.positions`    | `dlq.sensors.ais`   | 50055     | UC006    | FEAT-08 |
| 5   | Cyber Threat Ingestion | `svc-cyber-ingestion/` | `sensors.cyber.iocs`       | `dlq.sensors.cyber` | 50056     | UC007    | FEAT-09 |

---

## 3. Target Directory Structure (Per Service)

Each service must have this structure matching `svc-radar-ingestion/`:

```
svc-<sensor>-ingestion/
├── cmd/<sensor>-ingestion/
│   └── main.go                    # REWRITE — full production wiring
├── internal/
│   ├── config/
│   │   ├── config.go              # NEW — sensor-specific config
│   │   └── config_test.go         # NEW
│   ├── domain/
│   │   ├── validator.go           # EXISTS — do NOT modify
│   │   ├── validator_test.go      # EXISTS — do NOT modify
│   │   ├── normalizer.go          # EXISTS — do NOT modify
│   │   └── normalizer_test.go     # EXISTS — do NOT modify
│   ├── handler/
│   │   ├── ingestion.go           # NEW — gRPC handler wrapping validator/normalizer/producer
│   │   └── ingestion_test.go      # NEW
│   ├── producer/
│   │   ├── observation.go         # NEW — Redpanda producer wrapper
│   │   └── observation_test.go    # NEW
│   └── mapper/
│       ├── enricher.go            # NEW — classification guard + metadata enrichment
│       └── enricher_test.go       # NEW
├── go.mod                         # EXISTS
├── Dockerfile                     # EXISTS
└── README.md                      # EXISTS — update if needed
```

---

## 4. Implementation Pattern

### 4.1 Reference: `svc-radar-ingestion/cmd/radar-ingestion/main.go`

The reference `main.go` follows this exact sequence:

1. `config.MustLoad()` — sensor-specific config from env vars
2. `telemetry.Init()` — OTel tracing + metrics + structured logger
3. `classification.NewGuard()` — classification ceiling from config
4. `redpanda.NewProducer()` — main output producer
5. `redpanda.NewProducer()` — DLQ producer
6. `audit.NewEmitter()` — audit events to Redpanda
7. `health.NewChecker()` — register `redpanda` and `grpc` checks
8. Domain components — `NewValidator()`, `NewNormalizer()`, `mapper.NewEnricher()`
9. `producer.NewObservationProducer()` — output + DLQ producer wrappers
10. `handler.NewIngestionHandler()` — ties everything together
11. gRPC server with `interceptors.BuildUnaryServerInterceptors()` + `BuildStreamServerInterceptors()`
12. Register `IngestionServiceServer` + `HealthServiceServer`
13. `shutdown.NewManager()` — register all cleanup functions
14. Start listener, serve, set health to SERVING
15. `sm.Wait()` — block until shutdown signal

**Every elevated service MUST follow this exact sequence.** The only differences are:

- Config struct field names and env var prefixes (e.g., `RTSA_EW_` vs `RTSA_RADAR_`)
- Service name string (e.g., `"svc-ew-ingestion"`)
- Topic names (e.g., `sensors.ew.intercepts`, `dlq.sensors.ew`)
- Port number (e.g., `50052`)
- Validator/normalizer constructors (e.g., `domain.NewValidator()` — already implemented)

### 4.2 Sensor-Specific Config

Each service needs an `internal/config/config.go` with sensor-specific thresholds. Follow the pattern from `svc-radar-ingestion/internal/config/config.go`:

```go
// CLASSIFICATION: UNCLASSIFIED
package config

import "github.com/arvinddhasmana/rtsa_webgpu/pkg/config"

type Config struct {
    // Service
    ServiceVersion    string
    Environment       string
    GRPCPort          int
    HealthPort        int

    // Redpanda
    RedpandaBrokers     []string
    RedpandaTLSEnabled  bool
    OutputTopic         string
    DLQTopic            string

    // Telemetry
    OTelEndpoint string

    // Classification
    MaxClassification string

    // Sensor-specific thresholds (vary per service)
    // ... e.g., MaxFrequencyMHz, MaxCEPMeters, etc.
}
```

### 4.3 Handler Pattern (Can Use `pkg/ingestion.Handler` or Custom)

Two acceptable approaches:

**Option A — Use `pkg/ingestion.Handler` (simpler, less code duplication):**
The current lightweight services already use `ingestion.NewHandler()`. This is acceptable IF the handler supports the full interface (audit emitter, DLQ routing, stats). Review `pkg/ingestion/handler.go` to verify. If it does, keep using it and just upgrade the producer from `LogProducer` to real `redpanda.Producer`.

**Option B — Custom handler per service (matches radar exactly):**
Copy `svc-radar-ingestion/internal/handler/ingestion.go` and adapt. This gives per-service control over metrics, logging, and audit events.

**Recommendation**: Use Option A if `pkg/ingestion.Handler` already supports all features. Fall back to Option B otherwise.

### 4.4 Producer Pattern

```go
// CLASSIFICATION: UNCLASSIFIED
package producer

import (
    "context"
    "github.com/arvinddhasmana/rtsa_webgpu/pkg/redpanda"
)

type ObservationProducer struct {
    producer *redpanda.Producer
    topic    string
}

func NewObservationProducer(prod *redpanda.Producer, topic string) *ObservationProducer {
    return &ObservationProducer{producer: prod, topic: topic}
}

func (p *ObservationProducer) Produce(ctx context.Context, key string, value []byte) error {
    return p.producer.Produce(ctx, p.topic, key, value)
}

func (p *ObservationProducer) Topic() string { return p.topic }
func (p *ObservationProducer) Close() error  { return nil } // Producer lifecycle managed by main
```

### 4.5 Enricher Pattern

```go
// CLASSIFICATION: UNCLASSIFIED
package mapper

import "github.com/arvinddhasmana/rtsa_webgpu/pkg/classification"

type Enricher struct {
    serviceName string
    guard       *classification.Guard
}

func NewEnricher(serviceName string, guard *classification.Guard) *Enricher {
    return &Enricher{serviceName: serviceName, guard: guard}
}

// Enrich adds service metadata and validates classification ceiling.
func (e *Enricher) Enrich(obs *ingestionv1.SensorObservation) error {
    // Check classification ceiling
    if !e.guard.IsAllowed(obs.Classification) {
        return fmt.Errorf("classification %s exceeds ceiling", obs.Classification)
    }
    return nil
}
```

---

## 5. Per-Service Configuration Details

### 5.1 svc-ew-ingestion

| Config            | Env Var                         | Default                 |
| ----------------- | ------------------------------- | ----------------------- |
| Service name      | —                               | `svc-ew-ingestion`      |
| gRPC port         | `RTSA_EW_GRPC_PORT`             | `50052`                 |
| Output topic      | `RTSA_EW_OUTPUT_TOPIC`          | `sensors.ew.intercepts` |
| DLQ topic         | `RTSA_EW_DLQ_TOPIC`             | `dlq.sensors.ew`        |
| Max frequency MHz | `RTSA_EW_MAX_FREQUENCY_MHZ`     | `40000`                 |
| Max future offset | `RTSA_EW_MAX_FUTURE_OFFSET_SEC` | `300`                   |
| Max past offset   | `RTSA_EW_MAX_PAST_OFFSET_SEC`   | `86400`                 |

### 5.2 svc-elint-ingestion

| Config         | Env Var                     | Default                    |
| -------------- | --------------------------- | -------------------------- |
| Service name   | —                           | `svc-elint-ingestion`      |
| gRPC port      | `RTSA_ELINT_GRPC_PORT`      | `50053`                    |
| Output topic   | `RTSA_ELINT_OUTPUT_TOPIC`   | `sensors.elint.detections` |
| DLQ topic      | `RTSA_ELINT_DLQ_TOPIC`      | `dlq.sensors.elint`        |
| Max CEP meters | `RTSA_ELINT_MAX_CEP_METERS` | `50000`                    |

### 5.3 svc-isr-ingestion

| Config               | Env Var                      | Default                    |
| -------------------- | ---------------------------- | -------------------------- |
| Service name         | —                            | `svc-isr-ingestion`        |
| gRPC port            | `RTSA_ISR_GRPC_PORT`         | `50054`                    |
| Output topic         | `RTSA_ISR_OUTPUT_TOPIC`      | `sensors.isr.observations` |
| DLQ topic            | `RTSA_ISR_DLQ_TOPIC`         | `dlq.sensors.isr`          |
| Min polygon vertices | `RTSA_ISR_MIN_POLYGON_VERTS` | `3`                        |

### 5.4 svc-ais-ingestion

| Config               | Env Var                         | Default                 |
| -------------------- | ------------------------------- | ----------------------- |
| Service name         | —                               | `svc-ais-ingestion`     |
| gRPC port            | `RTSA_AIS_GRPC_PORT`            | `50055`                 |
| Output topic         | `RTSA_AIS_OUTPUT_TOPIC`         | `sensors.ais.positions` |
| DLQ topic            | `RTSA_AIS_DLQ_TOPIC`            | `dlq.sensors.ais`       |
| Max speed jump knots | `RTSA_AIS_MAX_SPEED_JUMP_KNOTS` | `50`                    |

### 5.5 svc-cyber-ingestion

| Config           | Env Var                       | Default               |
| ---------------- | ----------------------------- | --------------------- |
| Service name     | —                             | `svc-cyber-ingestion` |
| gRPC port        | `RTSA_CYBER_GRPC_PORT`        | `50056`               |
| Output topic     | `RTSA_CYBER_OUTPUT_TOPIC`     | `sensors.cyber.iocs`  |
| DLQ topic        | `RTSA_CYBER_DLQ_TOPIC`        | `dlq.sensors.cyber`   |
| Dedup cache size | `RTSA_CYBER_DEDUP_CACHE_SIZE` | `1000`                |

---

## 6. Test Requirements

### 6.1 Unit Tests per Service (New Files)

Each service needs these new test files:

#### `internal/config/config_test.go`

| #   | Test                        | Expected                        |
| --- | --------------------------- | ------------------------------- |
| T01 | Load defaults (no env vars) | Default topic, port, thresholds |
| T02 | Override via env vars       | Custom values applied           |
| T03 | Invalid port (negative)     | Error or panic                  |

#### `internal/handler/ingestion_test.go`

| #   | Test                               | Expected                                      |
| --- | ---------------------------------- | --------------------------------------------- |
| T04 | Valid observation → produce        | Producer called with correct topic            |
| T05 | Invalid observation → DLQ          | DLQ producer called, main producer NOT called |
| T06 | Classification ceiling violation   | Rejected, DLQ, audit event emitted            |
| T07 | Streaming ingestion (multiple obs) | Each processed independently                  |
| T08 | Producer error → error returned    | gRPC error code INTERNAL                      |

#### `internal/producer/observation_test.go`

| #   | Test                              | Expected             |
| --- | --------------------------------- | -------------------- |
| T09 | Topic() returns configured topic  | Correct topic string |
| T10 | Produce() delegates to underlying | Mock verify          |

#### `internal/mapper/enricher_test.go`

| #   | Test                                       | Expected              |
| --- | ------------------------------------------ | --------------------- |
| T11 | Observation within classification ceiling  | Enriched successfully |
| T12 | Observation exceeds classification ceiling | Error returned        |

### 6.2 Integration Tests (New — 1 per service)

Add to `tests/integration/` or `svc-<sensor>/internal/integration/`:

| #        | Test                              | What It Validates                                      |
| -------- | --------------------------------- | ------------------------------------------------------ |
| IT-EW    | EW obs → gRPC → Redpanda topic    | Real Redpanda producer, correct topic, correct headers |
| IT-ELINT | ELINT obs → gRPC → Redpanda topic | Same as above                                          |
| IT-ISR   | ISR obs → gRPC → Redpanda topic   | Same as above                                          |
| IT-AIS   | AIS obs → gRPC → Redpanda topic   | Same as above (extend existing)                        |
| IT-CYBER | Cyber obs → gRPC → Redpanda topic | Same as above                                          |

Use `testcontainers-go` for Redpanda. Follow pattern from `svc-radar-ingestion` or `tests/integration/ingestion_test.go`.

### 6.3 Coverage Target

| Component            | Target             |
| -------------------- | ------------------ |
| Each service overall | ≥80% line coverage |
| Validator (existing) | Already ≥90%       |
| Handler (new)        | ≥85%               |
| Config (new)         | ≥80%               |
| Producer (new)       | ≥80%               |
| Enricher (new)       | ≥85%               |

---

## 7. Build Verification

After completing all 5 services:

```bash
# All services compile
cd svc-ew-ingestion && go build ./... && cd ..
cd svc-elint-ingestion && go build ./... && cd ..
cd svc-isr-ingestion && go build ./... && cd ..
cd svc-ais-ingestion && go build ./... && cd ..
cd svc-cyber-ingestion && go build ./... && cd ..

# All tests pass with race detection
make test

# Coverage meets thresholds
make test-coverage

# Docker images build
docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.services.yml build \
  svc-ew-ingestion svc-elint-ingestion svc-isr-ingestion svc-ais-ingestion svc-cyber-ingestion

# Lint passes
make lint
```

---

## 8. IMPORTANT Constraints

1. **Do NOT modify existing `domain/validator.go` or `domain/normalizer.go` files** — they are complete and tested
2. **Do NOT modify existing `domain/validator_test.go` or `domain/normalizer_test.go`** — they must continue passing
3. **Do NOT modify `pkg/ingestion/`** interfaces unless absolutely necessary — existing code depends on them
4. **DO use `pkg/ingestion.Handler` if it supports the full feature set** (audit, DLQ, stats) — avoid unnecessary code duplication
5. **All new files must start with `// CLASSIFICATION: UNCLASSIFIED`**
6. **All existing tests must continue to pass** after elevation
7. **Use sensor-specific service names** in all telemetry, logging, and audit — never use generic names

---

## 9. Agent Invocation

```
@greatest-ever-developer Implement v1 Module 02 from docs/implementation/v1/02-ingestion-service-completion.md

Context:
- Read docs/implementation/v1/00-v1-overview.md for v1 scope and traceability
- Read docs/implementation/00-implementation-overview.md for global conventions
- READ the actual source code of svc-radar-ingestion/ — replicate its main.go wiring exactly
- READ pkg/ingestion/ to understand shared interfaces (Handler, Producer, Config)
- Each service already has working domain/validator.go and domain/normalizer.go — do NOT modify them
- The main.go in each service currently uses ingestion.LogProducer — replace with real redpanda.Producer
- Follow the exact 15-step wiring sequence from svc-radar-ingestion/cmd/radar-ingestion/main.go
- Add internal/config, internal/handler, internal/producer, internal/mapper packages
- Add unit tests for all new packages (handler, producer, config, enricher)
- Add integration test per service using testcontainers-go
- Topic names: sensors.ew.intercepts, sensors.elint.detections, sensors.isr.observations, sensors.ais.positions, sensors.cyber.iocs
- DLQ topics: dlq.sensors.ew, dlq.sensors.elint, dlq.sensors.isr, dlq.sensors.ais, dlq.sensors.cyber

Services to elevate (in order):
1. svc-ew-ingestion (port 50052)
2. svc-elint-ingestion (port 50053)
3. svc-isr-ingestion (port 50054)
4. svc-ais-ingestion (port 50055)
5. svc-cyber-ingestion (port 50056)

Deliverables:
1. 5 services with production-quality main.go (real Redpanda, telemetry, interceptors, health, audit)
2. internal/config/ with sensor-specific config + tests per service
3. internal/handler/ with ingestion handler + tests per service
4. internal/producer/ with observation producer + tests per service
5. internal/mapper/ with enricher + tests per service
6. Integration tests (1 per service) with testcontainers-go
7. All existing validator/normalizer tests continue to pass
8. ≥80% line coverage per service
9. go vet ./... passes for each service
10. golangci-lint run ./... passes for each service
```
