<!-- CLASSIFICATION: UNCLASSIFIED -->

# RTSA Testing How-To Guide

> **Version**: 1.0 | **Date**: 2026-03-02
>
> Step-by-step instructions for running every type of test in the RTSA project.

---

## Prerequisites

| Tool | Version | Check |
|---|---|---|
| Go | 1.22+ | `go version` |
| Docker | 24+ | `docker --version` |
| Docker Compose | v2+ | `docker compose version` |
| Node.js | 20+ | `node --version` |
| pnpm | 9+ | `pnpm --version` |

---

## 1. Go Unit Tests (per-service)

Unit tests run **without any infrastructure**. They use mocks for all external dependencies.

### Run all unit tests across every Go module

```bash
# From repo root — automated script with coverage enforcement
./scripts/dev/test-go.sh
```

This will:
- Discover all Go modules (`svc-*`, `pkg/`, `wasm-transforms`)
- Run `go test -race -coverprofile=... ./...` in each
- Enforce ≥ 80% coverage threshold
- Generate HTML coverage reports in `.coverage/`

### Run tests for a single service

```bash
cd svc-radar-ingestion
go test -race -v -coverprofile=coverage.out ./...

# View coverage report
go tool cover -html=coverage.out -o coverage.html
```

### Run a specific test

```bash
cd svc-anomaly-detection
go test -race -v -run TestSpeedDetector_HighSpeed ./internal/domain/detectors/
```

### Expected output

```
=== RUN   TestSpeedDetector_HighSpeed_ReturnsAnomaly
--- PASS: TestSpeedDetector_HighSpeed_ReturnsAnomaly (0.00s)
PASS
ok      github.com/arvinddhasmana/RTSA_VS_Opus/svc-anomaly-detection/internal/domain/detectors   0.004s
```

---

## 2. Integration Tests (testcontainers)

Integration tests (`tests/integration/`) spin up **real Redpanda and ClickHouse containers** automatically via `testcontainers-go`. No manual Docker setup is needed.

### Run all integration tests

```bash
# Using Make (recommended)
cd tests && make test-integration

# Or manually
cd tests
RTSA_INTEGRATION_TESTS=true go test -v -tags integration -timeout 10m ./integration/...
```

### Run a specific integration test

```bash
cd tests
RTSA_INTEGRATION_TESTS=true go test -v -tags integration -run TestIT01_RadarIngestion ./integration/...
```

### Run only unit-level integration tests (fast, no full containers)

```bash
cd tests && make test-unit
```

### Expected output

```
=== RUN   TestIT01_RadarIngestion
🐳 Creating container for image redpandadata/redpanda:v24.1.1
✅ Container started: abc123
🔔 Container is ready: abc123
    ingestion_test.go:82: IT01 PASS: radar observation on sensors.radar.tracks
🐳 Stopping container: abc123
--- PASS: TestIT01_RadarIngestion (12.34s)
```

> **Note**: First run downloads container images (~500MB). Subsequent runs use cache.

---

## 3. End-to-End Tests (Docker Compose)

E2E tests (`tests/e2e/`) require the **full Docker Compose stack** — all 14+ services, Redpanda, ClickHouse, Envoy, etc.

### Run E2E tests (automated)

```bash
# Using Make (recommended) — handles full lifecycle
cd tests && make test-e2e

# Or use the helper script
./scripts/dev/test-e2e.sh
```

This will:
1. Build all Docker images
2. Start the full stack (`docker compose up --build --wait`)
3. Initialize Redpanda topics and ClickHouse schema
4. Run `go test -tags e2e ./e2e/...`
5. Tear down the stack and remove volumes

### Run E2E tests manually (stack already running)

```bash
# If you already have the stack up:
cd tests
RTSA_INTEGRATION_TESTS=true go test -v -tags e2e -timeout 15m ./e2e/...
```

### Run a single E2E test

```bash
cd tests
RTSA_INTEGRATION_TESTS=true go test -v -tags e2e -run TestE2E01_FullPipeline ./e2e/...
```

### Expected output

```
=== RUN   TestE2E01_FullPipeline
    full_pipeline_test.go:100: E2E01: 10 radar observations produced
    full_pipeline_test.go:122: E2E01: fused tracks received
--- PASS: TestE2E01_FullPipeline (45.23s)
```

### Troubleshooting

| Problem | Solution |
|---|---|
| Port conflict (port already allocated) | `docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.services.yml down -v` then retry |
| Container unhealthy | Check logs: `docker logs rtsa-<service>` |
| `UNKNOWN_TOPIC_OR_PARTITION` | Run `./scripts/dev/init-topics.sh` before tests |

---

## 4. Performance Benchmarks

Benchmarks (`tests/benchmark/`) validate NFR performance thresholds.

### Run all benchmarks

```bash
cd tests && make test-bench

# Or manually
cd tests
RTSA_INTEGRATION_TESTS=true go test -v -tags integration \
  -bench=. -benchtime=10s -timeout 10m ./benchmark/...
```

### Threshold violations

Benchmarks call `b.Errorf()` if thresholds are not met:
- **B01**: Ingestion ≥ 1,000 obs/sec
- **B02**: Fusion p95 ≤ 100ms
- **B03**: Anomaly p95 ≤ 150ms
- **B04**: Query 100-row p95 ≤ 500ms

---

## 5. Frontend Unit Tests (Vitest)

### Run all frontend tests

```bash
cd web-cop
pnpm test
```

### Run with coverage

```bash
cd web-cop
pnpm test:coverage
```

### Run in watch mode

```bash
cd web-cop
pnpm test:watch
```

---

## 6. Frontend E2E Tests (Playwright)

### Prerequisites

```bash
cd web-cop
pnpm exec playwright install --with-deps
```

### Run all browser E2E tests

```bash
cd web-cop
pnpm test:e2e
```

### Run headed (visible browser)

```bash
cd web-cop
pnpm test:e2e:headed
```

### View HTML report

```bash
cd web-cop
pnpm test:e2e:report
```

---

## 7. Coverage Report Generation

### Go coverage — aggregate across all modules

```bash
# Run the script (generates per-service .out and .html files)
./scripts/dev/test-go.sh

# Reports are in:
ls .coverage/
# svc-radar-ingestion.out  svc-radar-ingestion.html
# svc-alert.out            svc-alert.html
# ...
```

### Go coverage — single module with HTML

```bash
cd svc-fusion-engine
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
go tool cover -func=coverage.out | tail -1   # Summary line
```

### Frontend coverage

```bash
cd web-cop
pnpm test:coverage
# Coverage report in: web-cop/coverage/
```

---

## 8. Run All Tests

### Master script

```bash
./scripts/dev/test-all.sh
```

Runs in sequence:
1. Go unit tests (all modules)
2. Frontend unit tests (Vitest)
3. Integration tests (testcontainers)
4. E2E tests (Docker Compose lifecycle)
5. Benchmarks
6. Summary report

### Make targets

```bash
cd tests
make test-all        # integration + e2e + benchmarks
make test-integration
make test-e2e
make test-bench
make test-unit       # Fast subset of integration tests
```

---

## 9. Pre-PR Gate (Local CI Mirror)

Before opening a Pull Request, run the full 5-gate check:

```bash
./scripts/dev/pre-pr-check.sh
```

| Gate | What it checks |
|---|---|
| SG-1: Pre-Build | Secrets (gitleaks), classification headers, formatting |
| SG-2: Build | Go build, TypeScript compile, proto lint |
| SG-3: Test | Unit tests + 80% coverage |
| SG-4: Security | gosec, govulncheck, golangci-lint, semgrep, npm audit |
| SG-5: Integration | Integration tests (optional — skips if stack not running) |

---

## 10. Quick Reference

| What | Command |
|---|---|
| All Go unit tests | `./scripts/dev/test-go.sh` |
| Single service tests | `cd svc-<name> && go test -race -v ./...` |
| Integration tests | `cd tests && make test-integration` |
| E2E tests | `cd tests && make test-e2e` |
| Benchmarks | `cd tests && make test-bench` |
| Frontend unit tests | `cd web-cop && pnpm test` |
| Frontend E2E | `cd web-cop && pnpm test:e2e` |
| Coverage (Go) | `./scripts/dev/test-go.sh` → `.coverage/` |
| Coverage (frontend) | `cd web-cop && pnpm test:coverage` |
| Pre-PR gate | `./scripts/dev/pre-pr-check.sh` |
| Everything | `./scripts/dev/test-all.sh` |
