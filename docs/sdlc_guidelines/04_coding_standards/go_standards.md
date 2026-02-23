# Go Coding Standards

> **CLASSIFICATION: UNCLASSIFIED**
> **Document Type**: Coding Standard
> **Parent**: `04_coding_standards/general_coding.md`
> **Dependencies**: `04_coding_standards/secure_coding.md`
> **Last Updated**: 2026-02-23

---

## 1. Purpose

This document defines Go-specific coding standards for RTSA microservices. All Go code must follow these conventions in addition to the general coding standards.

## 2. Go Version and Module Setup

- **Go Version**: 1.22+ (latest stable)
- **Module Path**: `github.com/[org]/rtsa/[service-name]`
- **Module per service**: Each microservice is a separate Go module
- **Shared libraries**: Common code in `github.com/[org]/rtsa/pkg/[library]`

## 3. Project Structure

```
[service-name]/
├── cmd/
│   └── [service-name]/
│       └── main.go              # Entry point only: config load, DI wiring, server start
├── internal/
│   ├── config/
│   │   └── config.go            # Struct with env var tags; validation
│   ├── server/
│   │   └── grpc.go              # gRPC server setup, interceptor chain, graceful shutdown
│   ├── handler/
│   │   └── [domain]_handler.go  # gRPC handler implementations (thin layer)
│   ├── service/
│   │   └── [domain]_service.go  # Business logic; no infra deps; testable
│   ├── repository/
│   │   └── [store]_repo.go      # Redpanda producer/consumer, ClickHouse client
│   ├── model/
│   │   └── [domain].go          # Internal domain models (not Protobuf types)
│   ├── middleware/
│   │   ├── auth.go              # mTLS + RBAC interceptor
│   │   ├── audit.go             # Audit event producer interceptor
│   │   ├── logging.go           # Structured logging interceptor
│   │   ├── metrics.go           # Prometheus metrics interceptor
│   │   └── recovery.go         # Panic recovery interceptor (log + return INTERNAL)
│   └── mapper/
│       └── proto_mapper.go      # Protobuf <-> domain model conversions
├── proto/
│   └── [service]/v1/
│       └── [service].proto      # Service contract
├── generated/
│   └── [service]/v1/            # protoc-gen-go output (committed)
├── Dockerfile
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## 4. Error Handling

### 4.1 No Panics

```go
// PROHIBITED in production code:
panic("unexpected state")
log.Fatal("cannot start")  // Only acceptable in main.go init

// REQUIRED pattern:
if err != nil {
    return fmt.Errorf("[package].[function](%s): %w", key, err)
}
```

### 4.2 Error Wrapping

Always wrap errors with operation context. Use `%w` for wrappable errors:

```go
// GOOD — contextual, wrappable
func (s *IngestionService) ProcessEvent(ctx context.Context, event *model.SensorEvent) error {
    if err := s.validator.Validate(event); err != nil {
        return fmt.Errorf("ingestion.ProcessEvent(sensorID=%s): validate: %w", event.SensorID, err)
    }
    if err := s.producer.Publish(ctx, event); err != nil {
        return fmt.Errorf("ingestion.ProcessEvent(sensorID=%s): publish: %w", event.SensorID, err)
    }
    return nil
}

// BAD — no context, not wrappable
func (s *IngestionService) ProcessEvent(ctx context.Context, event *model.SensorEvent) error {
    err := s.producer.Publish(ctx, event)
    if err != nil {
        return err  // Lost context
    }
    return nil
}
```

### 4.3 Sentinel Errors

Define sentinel errors for domain-specific error types:

```go
var (
    ErrSensorNotRegistered = errors.New("sensor not registered")
    ErrInvalidCoordinates  = errors.New("coordinates out of valid range")
    ErrTrustScoreTooLow    = errors.New("feedback trust score below threshold")
)
```

### 4.4 gRPC Error Mapping

Map domain errors to gRPC status codes in handlers:

```go
func mapError(err error) error {
    switch {
    case errors.Is(err, service.ErrSensorNotRegistered):
        return status.Errorf(codes.NotFound, "sensor not registered: %v", err)
    case errors.Is(err, service.ErrInvalidCoordinates):
        return status.Errorf(codes.InvalidArgument, "invalid coordinates: %v", err)
    default:
        return status.Errorf(codes.Internal, "internal error")
    }
}
```

## 5. Context Propagation

### 5.1 Rules

- Every function that does I/O takes `context.Context` as the first parameter
- Never store `context.Context` in a struct
- Always propagate context from gRPC handlers to downstream calls
- Use context for cancellation, deadlines, and metadata propagation

```go
// GOOD
func (s *Service) Process(ctx context.Context, req *pb.Request) (*pb.Response, error) {
    result, err := s.repo.Fetch(ctx, req.GetId())
    // ...
}

// BAD — missing context
func (s *Service) Process(req *pb.Request) (*pb.Response, error) {
    result, err := s.repo.Fetch(req.GetId())
    // ...
}
```

### 5.2 Correlation ID

Extract or create a correlation ID from gRPC metadata and propagate through the context:

```go
func correlationID(ctx context.Context) string {
    md, ok := metadata.FromIncomingContext(ctx)
    if ok {
        if vals := md.Get("x-correlation-id"); len(vals) > 0 {
            return vals[0]
        }
    }
    return uuid.New().String()
}
```

## 6. Structured Logging with slog

### 6.1 Logger Setup

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
}))
slog.SetDefault(logger)
```

### 6.2 Usage Pattern

```go
slog.InfoContext(ctx, "sensor event processed",
    slog.String("sensor_id", event.SensorID),
    slog.String("sensor_type", event.SensorType.String()),
    slog.String("correlation_id", correlationID(ctx)),
    slog.Duration("processing_time", elapsed),
)

slog.ErrorContext(ctx, "failed to publish event",
    slog.String("sensor_id", event.SensorID),
    slog.String("error", err.Error()),
    slog.String("correlation_id", correlationID(ctx)),
)
```

### 6.3 Prohibited

```go
// NEVER log classified data
slog.Info("event", "payload", event.RawPayload)      // BAD
slog.Info("event", "position", event.Position)        // BAD
slog.Info("event", "sensor_id", event.SensorID)       // GOOD — metadata only

// NEVER use fmt.Println for logging
fmt.Println("processing event")                       // BAD
```

## 7. Testing Conventions

### 7.1 Table-Driven Tests

```go
func TestValidateCoordinates(t *testing.T) {
    tests := []struct {
        name    string
        lat     float64
        lon     float64
        wantErr bool
    }{
        {name: "valid", lat: 45.4215, lon: -75.6972, wantErr: false},
        {name: "lat too high", lat: 91.0, lon: -75.6972, wantErr: true},
        {name: "lon too low", lat: 45.4215, lon: -181.0, wantErr: true},
        {name: "zero coords", lat: 0, lon: 0, wantErr: false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateCoordinates(tt.lat, tt.lon)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateCoordinates(%f, %f) error = %v, wantErr %v",
                    tt.lat, tt.lon, err, tt.wantErr)
            }
        })
    }
}
```

### 7.2 Interface-Based Mocking

Define interfaces for infrastructure dependencies:

```go
// Repository interface — used by service layer
type EventRepository interface {
    Publish(ctx context.Context, event *model.SensorEvent) error
    Subscribe(ctx context.Context, topic string) (<-chan *model.SensorEvent, error)
}

// Mock for testing
type mockEventRepo struct {
    publishFn func(ctx context.Context, event *model.SensorEvent) error
}

func (m *mockEventRepo) Publish(ctx context.Context, event *model.SensorEvent) error {
    return m.publishFn(ctx, event)
}
```

### 7.3 Test File Location

Tests live next to the code they test: `handler.go` → `handler_test.go` (same package). Integration tests in a `_test` package suffix.

## 8. Memory Management for Edge Deployment

- Use `sync.Pool` for frequently allocated objects (event buffers, Protobuf messages)
- Set `GOMEMLIMIT` for constrained environments
- Profile with `pprof` before optimizing — don't optimize prematurely
- Avoid creating goroutines in hot paths without bounded concurrency (use `semaphore.Weighted`)
- Use streaming gRPC instead of loading full datasets into memory

## 9. Graceful Shutdown

Every service must handle OS signals for graceful shutdown:

```go
func main() {
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    // ... setup services ...

    // Start gRPC server in goroutine
    go func() {
        if err := grpcServer.Serve(lis); err != nil {
            slog.Error("gRPC server error", "error", err)
        }
    }()

    // Block until signal
    <-ctx.Done()
    slog.Info("shutting down gracefully")

    // Drain in-flight requests
    grpcServer.GracefulStop()
    // Close Redpanda connections
    producer.Close()
    consumer.Close()

    slog.Info("shutdown complete")
}
```

## 10. Linting Configuration

Use `golangci-lint` with at minimum these linters:

```yaml
# .golangci.yml
linters:
  enable:
    - errcheck        # Check for unchecked errors
    - govet           # Go vet checks
    - staticcheck     # Advanced static analysis
    - unused          # Unused code detection
    - gosec           # Security checks
    - bodyclose       # HTTP body close checks
    - contextcheck    # Context propagation checks
    - nilerr          # Nil error return checks
    - errorlint       # Error wrapping checks
    - gocritic        # Code style and correctness
    - revive          # Configurable linter
    - misspell        # Spelling
    - goimports       # Import ordering
```

## 11. AI Agent Instructions

When generating Go code:

1. Start every file with `// CLASSIFICATION: UNCLASSIFIED`
2. Follow the project structure in Section 3
3. Never use `panic()` or `log.Fatal()` outside of `main.go` init
4. Always wrap errors with context using `fmt.Errorf` and `%w`
5. Every function doing I/O takes `context.Context` as first parameter
6. Use `slog` for structured JSON logging — never log classified data
7. Write table-driven tests for all business logic
8. Use interfaces for infrastructure dependencies to enable testing
9. Include package-level traceability comments (Feature, UC, Requirements)
10. Handle graceful shutdown with signal handling in `main.go`
