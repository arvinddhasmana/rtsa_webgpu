<!-- CLASSIFICATION: UNCLASSIFIED -->

# v5 Common Guidelines — Shared Across All Phases

> **Classification**: UNCLASSIFIED
> **Date**: 2026-03-10

---

## 1. Validation Cycle (Per Use Case)

Every use case follows a 3-step cycle:

1. **Review** — Walk through the implementation code and ALL test files against the use case requirements document in `docs/business/usecases/UC<NNN>_*.md`
2. **Run** — Execute ALL relevant tests: Unit → Integration → E2E → Browser E2E. Capture results.
3. **Record** — Log each issue found into the Issue Log (at the end of the phase plan). Do NOT fix immediately.

After ALL use cases in a phase are reviewed, issues are **fixed in batch**.

---

## 2. Test Execution Commands

### 2.1 Go Unit Tests (per service)

```bash
# Run unit tests with race detector (one service)
cd svc-<name> && go test -race -count=1 -v ./...

# Run with coverage (one service)
cd svc-<name> && go test -race -count=1 -coverprofile=coverage.out ./... \
  && go tool cover -func=coverage.out | tail -1

# Run ALL Go unit tests via script (all services + pkg)
bash scripts/dev/test-go.sh
```

### 2.2 Integration Tests (testcontainers — self-contained)

```bash
# Runs IT01–IT14; testcontainers manage Docker images automatically
bash scripts/dev/test-integration.sh
```

### 2.3 E2E Tests (requires full Docker stack)

```bash
# Start infrastructure + services
docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.services.yml up -d --wait

# Initialize Redpanda topics + ClickHouse schema
bash scripts/dev/init-topics.sh
bash scripts/dev/init-clickhouse.sh

# Health check — all services must be [✓]
bash scripts/dev/health-check.sh

# Run E2E tests
bash scripts/dev/test-e2e.sh

# Teardown
docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.services.yml down -v
```

### 2.4 Benchmarks

```bash
bash scripts/dev/test-bench.sh
```

### 2.5 Full Suite (all stages, SUMMARY.txt output)

```bash
bash scripts/dev/test-all.sh
```

### 2.6 Frontend Unit Tests (Vitest)

```bash
cd web-cop-gpu && pnpm test
```

### 2.7 Frontend Coverage

```bash
cd web-cop-gpu && pnpm test:coverage
```

### 2.8 Frontend E2E — Playwright

```bash
# Tier 1: Headless (mocked backend — fast, deterministic)
cd web-cop-gpu && pnpm test:e2e

# Tier 1: Headed (visual verification via browser subagent)
cd web-cop-gpu && pnpm test:e2e -- --headed
```

---

## 3. Coverage Targets (from SDLC Guidelines)

| Component Type                        | Line Coverage | Branch Coverage |
| ------------------------------------- | ------------- | --------------- |
| Core domain logic (fusion, inference) | 90%+          | 85%+            |
| gRPC service handlers                 | 85%+          | 80%+            |
| Sensor adapters / parsers             | 90%+          | 85%+            |
| Feedback trust scoring                | 95%+          | 90%+            |
| Anti-poisoning logic                  | 95%+          | 90%+            |
| Configuration / startup               | 80%+          | 75%+            |
| SolidJS components                    | 80%+          | 75%+            |
| **Global minimum**                    | **80%**       | —               |

---

## 4. Test Categories Reference

| ID    | Category       | What It Validates                     |
|-------|----------------|---------------------------------------|
| IT01  | Ingestion      | Radar → Redpanda topic                |
| IT02  | Ingestion      | Invalid → DLQ routing                 |
| IT03  | Ingestion      | All 6 sensor types                    |
| IT04  | Fusion         | 3 obs → fused track                   |
| IT05  | Fusion         | Track merging (score ≥ 0.85)          |
| IT06  | Fusion         | Stale timeout (60s → STALE)           |
| IT07  | Fusion         | Classification MAX propagation        |
| IT08  | Anomaly        | Speed anomaly (>3σ)                   |
| IT09  | Anomaly        | AIS manipulation (>0.5 NM)            |
| IT10  | Anomaly        | Severity routing (CRITICAL/ELEVATED/WATCH) |
| IT11  | ETL            | Tracks → ClickHouse (100 rows)        |
| IT12  | ETL            | Audit → ClickHouse (ITSG-33 no TTL)   |
| IT13  | ETL            | Materialized view aggregations         |
| **IT14** | **Classification** | **MANDATORY: Full propagation chain** |
| E2E01 | Pipeline       | Sim → Ingestion → Fusion → Alert → Query |
| E2E02 | Workflow       | Alert generate → stream → acknowledge |
| E2E03 | Workflow       | Feedback → Trust → Validate            |
| Neg01 | Negative       | Malformed sensor → DLQ                 |
| Neg02 | Negative       | Anti-poisoning rejects feedback        |
| Neg03 | Negative       | Classification violation rejected      |
| Neg04 | Negative       | Rate limiting enforcement              |
| B01   | Performance    | Ingestion ≥ 1,000 obs/sec             |
| B02   | Performance    | Fusion p95 ≤ 100ms                    |
| B03   | Performance    | Anomaly p95 ≤ 150ms                   |
| B04   | Performance    | Query 100-row p95 ≤ 500ms             |

---

## 5. Issue Log Format

Use this format in each phase plan's Issue Log section:

```markdown
| # | UC | Severity | File(s) | Issue Description | Requirement |
|---|----|----------|---------|-------------------|-------------|
| 1 | UC002 | BLOCKING | svc-radar-ingestion/handler.go | Missing DLQ routing on validation error | CR-ING-008 |
```

Severity levels: `BLOCKING` (prevents E2E demo flow), `WARNING` (incorrect but not breaking), `IMPROVEMENT` (enhancement, not spec violation).

---

## 6. Key Documentation References

| Document | Path |
|----------|------|
| Business Requirements | `docs/business/requirements.md` |
| Feature List | `docs/business/feature_list.md` |
| Use Cases | `docs/business/usecases/UC<NNN>_*.md` |
| Demo Guide | `docs/demo/demo_setup_run_showcase.md` |
| User Guide (by role) | `docs/user_guide/` |
| SDLC Testing Strategy | `docs/sdlc_guidelines/05_testing/testing_strategy.md` |
| Testing Guide | `docs/testing/testing_guide.md` |
| v4 Implementation Review | `docs/implementation/v4/v4_implementation_review.md` |

---

## 7. Docker Stack Management

```bash
# Start full stack
docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.services.yml up -d --wait

# Initialize (first time or after volumes teardown)
bash scripts/dev/init-topics.sh
bash scripts/dev/init-clickhouse.sh

# Health check
bash scripts/dev/health-check.sh

# Seed demo data (for UI verification)
bash scripts/demo/seed-demo-data.sh

# Teardown (preserves volumes)
docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.services.yml down

# Full teardown (wipes data)
docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.services.yml down -v
```

---

## 8. SDLC Compliance Reminders

- **Test naming**: `Test<Function>_<Scenario>_<ExpectedResult>` (Go)
- **Build tags**: `//go:build integration` for integration tests, `//go:build e2e` for E2E tests
- **Guard variable**: `RTSA_INTEGRATION_TESTS=true` required for integration/E2E
- **Synthetic data only**: mid-Atlantic coordinates (43–47°N, 55–65°W), deterministic seeds
- **Classification**: Every file must have `<!-- CLASSIFICATION: UNCLASSIFIED -->` or Go comment equivalent
- **Mocking**: Interface-based, mock at boundary only, hand-written preferred
