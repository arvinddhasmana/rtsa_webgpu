<!-- CLASSIFICATION: UNCLASSIFIED -->

# RTSA Testing How-To Guide

> **Version**: 1.1 | **Date**: 2026-03-02
>
> Step-by-step instructions for running every type of test in the RTSA project.

---

## Prerequisites

| Tool           | Version | Check                    |
| -------------- | ------- | ------------------------ |
| Go             | 1.22+   | `go version`             |
| Docker         | 24+     | `docker --version`       |
| Docker Compose | v2+     | `docker compose version` |
| Node.js        | 20+     | `node --version`         |
| pnpm           | 9+      | `pnpm --version`         |

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
- Run `go test -race -coverprofile=... ./...` in each module
- Enforce ≥ 80% coverage threshold
- Generate HTML coverage reports in `.coverage/`
- Save per-module `go test -v` output to `.test-results/<timestamp>/unit/<module>.log`
- Append failed test names to `.test-results/<timestamp>/unit/failures.txt`
- Print a detailed failure list (first 10 failures) at the end of the run

When called by `test-all.sh`, logs go into the master run directory. When run standalone, a new timestamped directory is created automatically.

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
ok      github.com/arvinddhasmana/rtsa_webgpu/svc-anomaly-detection/internal/domain/detectors   0.004s

─────────────────────────────────────────────────────
  Go unit test summary:
  Tests  : 47 run | 46 passed | 1 failed
  Modules: 15 total | 14 passed | 1 failed
  Coverage reports : .coverage/
  Per-module logs  : .test-results/2026-03-02T14-30-00/unit/*.log
  Failed test list : .test-results/2026-03-02T14-30-00/unit/failures.txt
─────────────────────────────────────────────────────
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

| Problem                                | Solution                                                                                               |
| -------------------------------------- | ------------------------------------------------------------------------------------------------------ |
| Port conflict (port already allocated) | `docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.services.yml down -v` then retry |
| Container unhealthy                    | Check logs: `docker logs rtsa-<service>`                                                               |
| `UNKNOWN_TOPIC_OR_PARTITION`           | Run `./scripts/dev/init-topics.sh` before tests                                                        |

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
cd web-cop-gpu
pnpm test
```

### Run with coverage

```bash
cd web-cop-gpu
pnpm test:coverage
```

### Run in watch mode

```bash
cd web-cop-gpu
pnpm test:watch
```

---

## 6. Frontend E2E Tests (Playwright)

### Prerequisites

```bash
cd web-cop-gpu
pnpm exec playwright install --with-deps
```

### Run all browser E2E tests

```bash
cd web-cop-gpu
pnpm test:e2e
```

### Run headed (visible browser)

```bash
cd web-cop-gpu
pnpm test:e2e:headed
```

### View HTML report

```bash
cd web-cop-gpu
pnpm test:e2e:report
```

---

## 7. Coverage Report Generation

### Go coverage — aggregate across all modules

```bash
# Run the script (generates per-service .out and .html files and per-module log files)
./scripts/dev/test-go.sh

# Coverage HTML reports:
ls .coverage/
# svc-radar-ingestion.out  svc-radar-ingestion.html
# svc-alert.out            svc-alert.html
# …

# Raw test logs (go test -v output per module):
ls .test-results/<timestamp>/unit/
# svc-radar-ingestion.log  svc-alert.log  …  failures.txt  counts.txt
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
cd web-cop-gpu
pnpm test:coverage
# Coverage report in: web-cop-gpu/coverage/
```

---

## 8. Run All Tests

### Master script

```bash
./scripts/dev/test-all.sh [--skip-e2e] [--skip-bench]
```

Runs in sequence:

1. Go unit tests (all modules)
2. Frontend unit tests (Vitest)
3. Integration tests (testcontainers)
4. E2E tests (Docker Compose lifecycle) — skipped with `--skip-e2e`
5. Benchmarks — skipped with `--skip-bench`
6. Stage table + failure details + results directory location

### Summary output

After all stages complete, a table is printed followed by per-stage failure details:

```
  Stage                                   Tests  Passed  Failed    Status
  ──────────────────────────────────────  ─────  ───────  ───────  ────────
  Unit — Go                                 142      138        4    FAILED
  Unit — Frontend                            23       23        0    PASSED
  Integration                                18       16        2    FAILED
  E2E                                         8        7        1    FAILED
  Benchmarks                                  4        4        0    SKIPPED

  Stages: 4 run | 2 passed | 2 failed | 1 skipped
  Elapsed: 183s

════════════════════════════════════════════════════════════════
  FAILURE DETAILS BY STAGE
════════════════════════════════════════════════════════════════
── FAILED TESTS: Unit — Go [4 failure(s), showing first 10] ────────
   1. svc-radar-ingestion/TestParseNMEA_Invalid_ReturnsError
   2. pkg/TestClassificationPropagation_Downgrade_Blocked
   3. svc-fusion-engine/TestFuseTrack_DuplicateID_Deduped
   4. svc-anomaly-detection/TestScore_BelowThreshold_NoAlert

── FAILED TESTS: Integration [2 failure(s), showing first 10] ──────
   1. integration/centralized/TestIngestionPipeline_RadarTrack
   2. integration/centralized/TestFusionPipeline_AISPosition

── Test Results Location ────────────────────────────────────────────
  /home/user/workspace/rtsa_webgpu/.test-results/2026-03-02T14-30-00/

  Subdirectories:
    unit/         — per-module Go unit test logs (*.log, failures.txt)
    frontend/     — Vitest output (run.log, failures.txt)
    integration/  — integration test logs (*.log, run.log, failures.txt)
    e2e/          — E2E test log (run.log, failures.txt)
    benchmark/    — benchmark log (run.log, failures.txt)
    SUMMARY.txt   — machine-readable aggregate for AI agents
```

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

| Gate              | What it checks                                            |
| ----------------- | --------------------------------------------------------- |
| SG-1: Pre-Build   | Secrets (gitleaks), classification headers, formatting    |
| SG-2: Build       | Go build, TypeScript compile, proto lint                  |
| SG-3: Test        | Unit tests + 80% coverage                                 |
| SG-4: Security    | gosec, govulncheck, golangci-lint, semgrep, npm audit     |
| SG-5: Integration | Integration tests (optional — skips if stack not running) |

---

## 10. Test Results & Logs

Every test script persists structured output under `.test-results/` so failures can be inspected offline and AI agents can consume a machine-readable aggregate without re-parsing terminal output.

### Directory structure

```
.test-results/<YYYY-MM-DDTHH-MM-SS>/        ← created once per test-all.sh run
  unit/
    <module-name>.log       ← full go test -v output per Go module
    failures.txt            ← one "module/TestName" entry per failed test
    counts.txt              ← TOTAL=N PASSED=N FAILED=N MODULES_TOTAL=N …
  frontend/
    run.log  failures.txt  counts.txt
  integration/
    centralized.log         ← tests/integration/ output
    <svc-name>.log          ← per-service integration test output
    run.log                 ← all integration logs concatenated
    failures.txt            ← one "integration/<suite>/TestName" per failure
    counts.txt
  e2e/
    run.log  failures.txt  counts.txt
  benchmark/
    run.log  failures.txt  counts.txt
  SUMMARY.txt               ← machine-readable aggregate (AI agent entry point)
```

> **`.test-results/` is in `.gitignore`** — log artifacts are never committed.

### Standalone vs. orchestrated runs

| How script is invoked                            | Where logs go                                           |
| ------------------------------------------------ | ------------------------------------------------------- |
| `./scripts/dev/test-all.sh`                      | `.test-results/<ts>/` — all stages in one directory     |
| `./scripts/dev/test-go.sh` (standalone)          | `.test-results/<ts>/unit/` — own timestamped dir        |
| `./scripts/dev/test-integration.sh` (standalone) | `.test-results/<ts>/integration/` — own timestamped dir |
| `./scripts/dev/test-e2e.sh` (standalone)         | `.test-results/<ts>/e2e/` — own timestamped dir         |

This is controlled by the `RTSA_TEST_RESULTS_DIR` environment variable, which `test-all.sh` exports before calling sub-scripts.

### Reading results programmatically or with an AI agent

```bash
# Latest run directory
LATEST=$(ls -td .test-results/*/ | head -1)

# All failed unit tests
cat "${LATEST}unit/failures.txt"

# Stage-level summary
cat "${LATEST}SUMMARY.txt"

# Full integration run log
cat "${LATEST}integration/run.log"

# Count of failed E2E tests
grep -oP 'FAILED=\K[0-9]+' "${LATEST}e2e/counts.txt"
```

### File formats

| File           | Format                                | Example                                                                           |
| -------------- | ------------------------------------- | --------------------------------------------------------------------------------- |
| `failures.txt` | One `<stage/TestName>` per line       | `svc-radar-ingestion/TestParseNMEA_Invalid_ReturnsError`                          |
| `counts.txt`   | Space-separated `KEY=VALUE`           | `TOTAL=47 PASSED=46 FAILED=1 MODULES_TOTAL=15 MODULES_PASSED=14 MODULES_FAILED=1` |
| `SUMMARY.txt`  | Structured plaintext sections         | `STAGES`, `STAGES_RUN=N`, `RESULTS_DIR=`, `FAILURE DETAILS`                       |
| `*.log`        | Raw `go test -v` / `pnpm test` output | Standard Go test output, unmodified                                               |

---

## 11. Quick Reference

| What                  | Command                                                                  |
| --------------------- | ------------------------------------------------------------------------ |
| All Go unit tests     | `./scripts/dev/test-go.sh`                                               |
| Single service tests  | `cd svc-<name> && go test -race -v ./...`                                |
| Integration tests     | `cd tests && make test-integration`                                      |
| E2E tests             | `cd tests && make test-e2e`                                              |
| Benchmarks            | `cd tests && make test-bench`                                            |
| Frontend unit tests   | `cd web-cop-gpu && pnpm test`                                            |
| Frontend E2E          | `cd web-cop-gpu && pnpm test:e2e`                                        |
| Coverage (Go)         | `./scripts/dev/test-go.sh` → `.coverage/`                                |
| Coverage (frontend)   | `cd web-cop-gpu && pnpm test:coverage`                                   |
| Pre-PR gate           | `./scripts/dev/pre-pr-check.sh`                                          |
| Everything            | `./scripts/dev/test-all.sh`                                              |
| Skip E2E + benchmarks | `./scripts/dev/test-all.sh --skip-e2e --skip-bench`                      |
| View latest results   | `cat .test-results/$(ls -t .test-results/ \| head -1)/SUMMARY.txt`       |
| Latest unit failures  | `cat .test-results/$(ls -t .test-results/ \| head -1)/unit/failures.txt` |
