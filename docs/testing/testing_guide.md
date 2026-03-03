<!-- CLASSIFICATION: UNCLASSIFIED -->

# RTSA Testing Architecture Guide

> **Version**: 1.1
> **Date**: 2026-03-02
> **Compliance**: ITSG-33 SA-11, NIST 800-53 SA-11

---

## 1. Overview

The RTSA project follows a **test pyramid** strategy with four layers:

| Layer                 | Location                               | Build Tag     | Infrastructure            | Coverage Target        |
| --------------------- | -------------------------------------- | ------------- | ------------------------- | ---------------------- |
| **Unit Tests**        | `svc-*/…/*_test.go`, `pkg/…/*_test.go` | _(none)_      | None                      | 80%+ (90%+ for domain) |
| **Integration Tests** | `tests/integration/`                   | `integration` | testcontainers-go (auto)  | Scenario-complete      |
| **E2E Tests**         | `tests/e2e/`                           | `e2e`         | Full Docker Compose stack | Critical paths         |
| **Benchmarks**        | `tests/benchmark/`                     | `integration` | testcontainers-go (auto)  | NFR thresholds         |
| **Frontend Unit**     | `web-cop/src/__tests__/`               | —             | Vitest                    | 80%+                   |
| **Frontend E2E (T1)** | `web-cop/e2e/mocked/`                  | —             | Playwright (Mocked)       | Critical flows         |
| **Frontend E2E (T2)** | `web-cop/e2e/live/`                    | —             | Playwright (Live Backend) | Integration flows      |
| **Frontend QA (T3)**  | Ad-hoc / AI Agents                     | —             | Browser Subagent          | UX Verification        |

---

## 1.1 Frontend Hybrid Validation Strategy

The Web COP follows an industry-standard **Hybrid Validation** approach for UI testing to bridge the gap between fast deterministic feedback and real-world system behavior:

1. **Tier 1: Fully Mocked E2E (Pre-Merge)**: Deterministic, sub-second Playwright tests using mocked gRPC streams to validate state, routing, and component mounting. Used as CI/CD gates.
2. **Tier 2: Live Backend E2E (Nightly)**: Playwright tests executed against the real Docker cluster (`start-backend.sh`), processing live space-seed data to guarantee full-stack communication (Envoy → Handlers → ClickHouse).
3. **Tier 3: Visual Autonomous QA (Sprint-level)**: Dynamic exploration using visual AI browser subagents to organically navigate the UI, ensuring complex visual states, layout flows, and simulated sensor realities meet human-quality standards without brittle selectors.

## 2. Repository Layout

```
RTSA_VS_Opus/
├── pkg/                          # Shared Go library — unit tests inside
│   ├── audit/*_test.go
│   ├── classification/*_test.go
│   ├── config/*_test.go
│   ├── health/*_test.go
│   ├── ingestion/*_test.go
│   ├── interceptors/*_test.go
│   ├── redpanda/*_test.go
│   └── ...
├── svc-radar-ingestion/          # Each service has its own go.mod + tests
│   ├── internal/
│   │   ├── domain/*_test.go      # Domain logic unit tests
│   │   ├── handler/*_test.go     # gRPC handler tests
│   │   ├── integration/          # Service-level integration tests
│   │   └── ...
│   └── go.mod
├── svc-alert/                    # Same pattern for all 14 svc-* modules
├── svc-anomaly-detection/
├── svc-audit/
├── svc-feedback/
├── svc-fusion-engine/
├── svc-query/
├── svc-track/
├── svc-training/
├── svc-nato-adapter/
├── svc-ais-ingestion/
├── svc-cyber-ingestion/
├── svc-elint-ingestion/
├── svc-ew-ingestion/
├── svc-isr-ingestion/
├── wasm-transforms/              # WASM module — unit tests inside
│
├── tests/                        # Centralized cross-service test suites
│   ├── integration/              # IT01–IT14: testcontainers-go tests
│   │   ├── ingestion_test.go     # IT01–IT03
│   │   ├── fusion_test.go        # IT04–IT07
│   │   ├── anomaly_test.go       # IT08–IT10
│   │   ├── etl_test.go           # IT11–IT13
│   │   ├── classification_test.go # IT14 (MANDATORY)
│   │   ├── feedback_test.go
│   │   ├── query_test.go
│   │   ├── audit_test.go
│   │   └── testutil/             # Shared helpers, fixtures
│   ├── e2e/                      # E2E01–E2E03 + negative tests
│   │   ├── full_pipeline_test.go
│   │   ├── alert_workflow_test.go
│   │   ├── feedback_workflow_test.go
│   │   ├── forensics_query_test.go
│   │   └── negative_test.go      # Neg01–Neg04
│   ├── benchmark/                # B01–B04: performance thresholds
│   │   ├── ingestion_bench_test.go
│   │   └── query_bench_test.go
│   └── Makefile
│
├── web-cop/                      # React frontend
│   ├── src/__tests__/            # Vitest unit tests
│   ├── e2e/                      # Playwright browser E2E tests
│   ├── vitest.config.ts
│   └── playwright.config.ts
│
├── scripts/dev/                  # Helper scripts
│   ├── test-go.sh                # Go unit tests + coverage + per-module log files
│   ├── test-integration.sh       # Integration test runner + per-suite log files
│   ├── test-e2e.sh               # E2E test runner (Docker lifecycle) + run log
│   ├── test-all.sh               # Master runner — all stages, table summary, SUMMARY.txt
│   └── pre-pr-check.sh           # 5-gate local CI mirror
│
└── deploy/
    ├── docker-compose.yml        # Infrastructure (Redpanda, ClickHouse, ...)
    └── docker-compose.services.yml # All RTSA microservices
```

---

## 3. Build Tags & Guard Variables

All integration and E2E tests are gated by:

```go
//go:build integration   // or: //go:build e2e
```

Plus the environment variable:

```bash
export RTSA_INTEGRATION_TESTS=true
```

Unit tests have **no build tag** and run with a plain `go test ./...`.

---

## 4. Coverage Requirements

| Component Type                        | Line Coverage | Branch Coverage |
| ------------------------------------- | ------------- | --------------- |
| Core domain logic (fusion, inference) | 90%+          | 85%+            |
| gRPC service handlers                 | 85%+          | 80%+            |
| Sensor adapters / parsers             | 90%+          | 85%+            |
| Feedback trust scoring                | 95%+          | 90%+            |
| Anti-poisoning logic                  | 95%+          | 90%+            |
| Configuration / startup               | 80%+          | 75%+            |
| React components                      | 80%+          | 75%+            |
| **Global minimum**                    | **80%**       | —               |

Coverage below threshold causes CI failure.

---

## 5. Test Scenario Index

### Integration Tests (IT01–IT14)

| ID       | Category           | Validates                                  |
| -------- | ------------------ | ------------------------------------------ |
| IT01     | Ingestion          | Radar → Redpanda topic                     |
| IT02     | Ingestion          | Invalid → DLQ routing                      |
| IT03     | Ingestion          | All 6 sensor types                         |
| IT04     | Fusion             | 3 obs → fused track                        |
| IT05     | Fusion             | Track merging (score ≥ 0.85)               |
| IT06     | Fusion             | Stale timeout (60s → STALE)                |
| IT07     | Fusion             | Classification MAX propagation             |
| IT08     | Anomaly            | Speed anomaly (>3σ)                        |
| IT09     | Anomaly            | AIS manipulation (>0.5 NM)                 |
| IT10     | Anomaly            | Severity routing (CRITICAL/ELEVATED/WATCH) |
| IT11     | ETL                | Tracks → ClickHouse (100 rows)             |
| IT12     | ETL                | Audit → ClickHouse (ITSG-33 no TTL)        |
| IT13     | ETL                | Materialized view aggregations             |
| **IT14** | **Classification** | **MANDATORY: Full propagation chain**      |

### E2E Tests (E2E01–E2E03 + Negative)

| ID    | Validates                                |
| ----- | ---------------------------------------- |
| E2E01 | Sim → Ingestion → Fusion → Alert → Query |
| E2E02 | Alert generation → acknowledgment        |
| E2E03 | Feedback → Trust → Validated             |
| Neg01 | Malformed sensor → DLQ                   |
| Neg02 | Anti-poisoning rejects feedback          |
| Neg03 | Classification violation rejected        |
| Neg04 | Rate limiting enforcement                |

### Benchmarks (B01–B04)

| ID  | NFR          | Threshold                 |
| --- | ------------ | ------------------------- |
| B01 | NFR-PERF-002 | Ingestion ≥ 1,000 obs/sec |
| B02 | NFR-PERF-004 | Fusion p95 ≤ 100ms        |
| B03 | NFR-PERF-005 | Anomaly p95 ≤ 150ms       |
| B04 | NFR-PERF-001 | Query 100-row p95 ≤ 500ms |

---

## 6. Test Data Policy

- **NEVER** use real operational or classified data
- All fixtures use **synthetic data** with mid-Atlantic coordinates (43–47°N, 55–65°W)
- Deterministic: all fixtures use seed-based randomness (`NewSeededRand(seed)`)
- Golden files stored in `testdata/golden/` directories

---

## 7. Mocking Standards

- Every external dependency is accessed through a Go interface
- Unit tests mock at the boundary (gRPC client, Redpanda producer, ClickHouse)
- Prefer hand-written mocks over code generation
- React tests use MSW (Mock Service Worker) for API mocking

---

## 8. CI Pipeline Stages

```
SG-1: Pre-Build  →  SG-2: Build  →  SG-3: Unit Tests  →  SG-4: Security  →  SG-5: Integration
 (secrets,lint,     (go build,       (go test + 80%     (gosec, govuln,    (testcontainers
  formatting)        tsc, buf)        coverage)          semgrep, audit)    + E2E optional)
```

| Stage                 | Blocks PR   | Time Budget |
| --------------------- | ----------- | ----------- |
| Lint & Formatting     | Yes         | < 2 min     |
| Build                 | Yes         | < 5 min     |
| Unit Tests + Coverage | Yes         | < 5 min     |
| Security (SAST/deps)  | Yes (High+) | < 15 min    |
| Integration           | Yes         | < 10 min    |
| E2E                   | Warn only   | < 30 min    |

---

## 9. Test Result Artifacts

Every test script saves structured output to a **timestamped results directory** so that failures can be inspected after the run, and so AI agents can consume a machine-readable summary without re-parsing terminal output.

### Directory layout

```
.test-results/<YYYY-MM-DDTHH-MM-SS>/     # created by test-all.sh or sub-scripts (standalone)
  unit/
    <module-name>.log       # raw go test -v output, one file per Go module
    failures.txt            # one "module/TestName" line per failed test
    counts.txt              # TOTAL=N PASSED=N FAILED=N MODULES_TOTAL=N …
  frontend/
    run.log                 # pnpm test stderr+stdout
    failures.txt
    counts.txt
  integration/
    centralized.log         # tests/integration/ output
    <svc-name>.log          # per-service integration test output
    run.log                 # concatenation of all integration logs
    failures.txt            # one "integration/<suite>/TestName" line per failure
    counts.txt
  e2e/
    run.log
    failures.txt
    counts.txt
  benchmark/
    run.log
    failures.txt
    counts.txt
  SUMMARY.txt               # machine-readable aggregate (AI agent entry point)
```

> **`.test-results/` is in `.gitignore`** — log artifacts are never committed.

### Environment variable coordination

`test-all.sh` exports `RTSA_TEST_RESULTS_DIR` pointing to the run directory. All sub-scripts (`test-go.sh`, `test-integration.sh`, `test-e2e.sh`) honour this variable so everything lands in **one directory per run**.

When a sub-script is run **standalone** (outside `test-all.sh`), it creates its own `.test-results/<timestamp>/` directory in the repo root.

### Key files for AI agents

| File                  | Contents                                                                   |
| --------------------- | -------------------------------------------------------------------------- |
| `SUMMARY.txt`         | Structured text: stage table, per-stage failure lists, `RESULTS_DIR=` path |
| `*/failures.txt`      | One `<stage/TestName>` per line — consumed by automated triage tools       |
| `*/counts.txt`        | `TOTAL=N PASSED=N FAILED=N` key=value — consumed by CI dashboards          |
| `unit/*.log`          | Full `go test -v` output per module — searched for diagnostic context      |
| `integration/run.log` | Full integration run — all suites concatenated                             |

### Summary output format (`test-all.sh`)

After all stages complete, `test-all.sh` prints:

1. A stage table showing Tests / Passed / Failed / Status for every stage
2. A **FAILURE DETAILS BY STAGE** block listing the first 10 failed tests per failing stage (full list in `failures.txt`)
3. The absolute path to the results directory with subdirectory descriptions

```
  Stage                                   Tests  Passed  Failed    Status
  ──────────────────────────────────────  ─────  ───────  ───────  ────────
  Unit — Go                                 142      138        4    FAILED
  Unit — Frontend                            23       23        0    PASSED
  Integration                                18       16        2    FAILED
  E2E                                         8        7        1    FAILED
  Benchmarks                                  4        4        0    PASSED

── FAILED TESTS: Unit — Go [4 failure(s), showing first 10] ──────────
   1. svc-radar-ingestion/TestParseNMEA_Invalid_ReturnsError
   2. pkg/TestClassificationPropagation_Downgrade_Blocked
   ...
```

---

## 10. Related Documents

- [Testing Strategy (SDLC)](../sdlc_guidelines/05_testing/testing_strategy.md)
- [Test Data Audit](test-data-audit.md)
- [Testing How-To (step-by-step)](testing_howto.md)
- [tests/README.md](../../tests/README.md)
