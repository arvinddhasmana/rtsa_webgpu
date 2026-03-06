<!-- CLASSIFICATION: UNCLASSIFIED -->

# RTSA v1 Implementation — Master Orchestration Document

> **Document**: RTSA v1 Pending Functionality — Module Orchestration
> **Version**: 1.0
> **Classification**: UNCLASSIFIED
> **Last Updated**: 2026-02-25
> **Target Agent**: Greatest Ever Developer (`@greatest-ever-developer`)

---

## 1. Purpose

This document is the master orchestration guide for completing RTSA v1. It tracks all **pending functionality** identified during the full-system audit performed against the 72 business requirements (CR-ING, CR-FUS, CR-INF, CR-FB, CR-UI, CR-HIS, CR-NATO, CR-SEC, NFR), 15 features (FEAT-01–FEAT-15), and 15 use cases (UC001–UC015).

**Scope of v1 pending work:**

| Category                  | Description                                                                             | Est. Effort |
| ------------------------- | --------------------------------------------------------------------------------------- | ----------- |
| Infrastructure Fixes      | 3 critical config bugs blocking end-to-end data flow + TS codegen                       | 2 days      |
| Ingestion Completion      | 5 partial ingestion services elevated to production parity                              | 3 days      |
| Service Hardening         | mTLS wiring, audit emitter integration, classification headers, NVG mode, sensor health | 3 days      |
| Deferred Capability Stubs | NATO adapter noop (UC014/UC015) + training pipeline noop (UC011)                        | 2 days      |
| Testing & Demo            | Browser E2E tests (Playwright), demo scenarios, launch scripts, negative tests          | 4 days      |

**Total estimated effort**: ~14 days

**What is already complete** (from the initial 18-module implementation):

- All 13 proto schemas (7 gRPC services, ~45 messages, ~12 enums)
- 9 shared Go libraries (`pkg/`) — all fully tested
- `svc-radar-ingestion` — full reference implementation with real Redpanda, telemetry, interceptors
- `svc-fusion-engine` — full Kalman filter, multi-sensor correlation, Prometheus metrics
- `svc-anomaly-detection` — 6 real detectors (speed, route, AIS, behavioral, temporal, proximity)
- `svc-feedback` — anti-poisoning guard with chi-squared test, trust scoring
- `svc-track` — in-memory cache, gRPC server-streaming, 3 handler types
- `svc-alert` — priority queue, severity ordering, streaming, acknowledgment
- `svc-query` — ClickHouse queries, classification filtering, guardrails, pagination
- `svc-audit` — immutable append-only repo, batched consumer, classification filtering
- 4 Wasm data transforms (sensor, track, feedback, alert validators)
- Full COP web application (SolidJS + WebGPU COP — see `implementation/v4/` for current plan; legacy React/MapLibre COP archived)
- Comprehensive simulator (6 sensor types, 4 movement patterns, 4 anomaly injectors)
- 120+ Go test files, 23 TypeScript test files
- 14 integration tests, 4 E2E tests, 4 benchmarks
- Docker Compose (infrastructure + services), Envoy proxy, 5 ETL pipelines, full observability stack
- 5 lightweight ingestion services with completed domain validation logic (validators + normalizers + tests)

---

## 2. Module Inventory

| #   | File                                    | Description                                                      | Phase | Est. Effort |
| --- | --------------------------------------- | ---------------------------------------------------------------- | ----- | ----------- |
| 00  | `v1/00-v1-overview.md`                  | This file — orchestration, traceability, progress tracking       | —     | —           |
| 01  | `v1/01-infrastructure-fixes.md`         | Envoy ports, topic names, table names, TS codegen                | P0    | 2d          |
| 02  | `v1/02-ingestion-service-completion.md` | Elevate 5 ingestion services to production parity                | P1    | 3d          |
| 03  | `v1/03-service-hardening.md`            | mTLS, audit emitters, classification headers, NVG, sensor health | P2    | 3d          |
| 04  | `v1/04-deferred-capability-stubs.md`    | NATO adapter noop + training pipeline noop                       | P3    | 2d          |
| 05  | `v1/05-testing-and-demo.md`             | Browser E2E, demo scenarios, launch scripts, negative tests      | P4    | 4d          |

---

## 3. Implementation Phases & Dependencies

```
Phase P0 — Infrastructure Fixes (01)
  ↓
Phase P1 — Ingestion Completion (02) — depends on fixed topics/ports
  ↓
Phase P2 — Service Hardening (03) — can run in parallel with P1
  ↓
Phase P3 — Deferred Stubs (04) — depends on proto generation from P0
  ↓
Phase P4 — Testing & Demo (05) — depends on all above
```

### Parallelism Rules

| Phase  | Module | Can Start After       | Max Parallel Agents                     |
| ------ | ------ | --------------------- | --------------------------------------- |
| **P0** | 01     | Immediately           | 1                                       |
| **P1** | 02     | P0 complete           | 1 (5 services sequentially or parallel) |
| **P2** | 03     | P0 complete           | 1 (can run alongside P1)                |
| **P3** | 04     | P0 complete           | 1 (can run alongside P1/P2)             |
| **P4** | 05     | P1 + P2 + P3 complete | 1                                       |

---

## 4. Requirement → v1 Task Traceability Matrix

### 4.1 Capability Requirements

| Requirement                             | Status Before v1                             | v1 Module | v1 Task                                       |
| --------------------------------------- | -------------------------------------------- | --------- | --------------------------------------------- |
| CR-ING-001 (Radar real-time)            | **Complete**                                 | —         | —                                             |
| CR-ING-002 (EW real-time)               | **Partial** — domain logic done, LogProducer | 02        | Elevate svc-ew-ingestion                      |
| CR-ING-003 (ELINT real-time)            | **Partial** — domain logic done, LogProducer | 02        | Elevate svc-elint-ingestion                   |
| CR-ING-004 (ISR real-time)              | **Partial** — domain logic done, LogProducer | 02        | Elevate svc-isr-ingestion                     |
| CR-ING-005 (AIS real-time)              | **Partial** — domain logic done, LogProducer | 02        | Elevate svc-ais-ingestion                     |
| CR-ING-006 (Cyber real-time)            | **Partial** — domain logic done, LogProducer | 02        | Elevate svc-cyber-ingestion                   |
| CR-ING-007 (Validate all sensor data)   | **Complete** — validators + Wasm transforms  | —         | —                                             |
| CR-ING-008 (Reject invalid → DLQ)       | **Complete** — DLQ routing in handler + Wasm | —         | —                                             |
| CR-ING-009 (50K events/sec DC)          | **Complete** — architecture supports         | —         | —                                             |
| CR-ING-010 (5K events/sec edge)         | **Complete** — architecture supports         | —         | —                                             |
| CR-FUS-001..007                         | **Complete**                                 | —         | —                                             |
| CR-INF-001..004                         | **Complete**                                 | —         | —                                             |
| CR-INF-005 (Multiple models)            | **Deferred** — single engine, future work    | —         | —                                             |
| CR-INF-006 (Pre-trained edge models)    | **Deferred** — rule-based approach used      | —         | —                                             |
| CR-INF-007 (Model versioning)           | **Deferred**                                 | —         | —                                             |
| CR-FB-001..007                          | **Complete**                                 | 03        | Wire audit emitter in svc-feedback            |
| CR-UI-001..003                          | **Complete**                                 | 05        | Browser E2E validates                         |
| CR-UI-004 (Sensor coverage display)     | **Missing**                                  | 03        | Sensor health endpoint + UI component         |
| CR-UI-005 (Classification markings)     | **Complete**                                 | 05        | Browser E2E validates                         |
| CR-UI-006 (Disconnected mode)           | **Complete**                                 | 05        | Browser E2E validates                         |
| CR-UI-007 (NVG dark mode)               | **Missing**                                  | 03        | NVG CSS theme                                 |
| CR-UI-008 (Keyboard nav)                | **Partial** — WCAG mentions                  | —         | —                                             |
| CR-HIS-001..007                         | **Complete**                                 | —         | —                                             |
| CR-NATO-001 (STANAG 5516)               | **Missing**                                  | 04        | Noop stub                                     |
| CR-NATO-002 (NFFI exchange)             | **Missing**                                  | 04        | Noop stub                                     |
| CR-NATO-003 (MIP exchange)              | **Missing**                                  | 04        | Noop stub                                     |
| CR-NATO-004 (Classification mapping)    | **Missing**                                  | 04        | Noop stub                                     |
| CR-NATO-005 (Cross-domain security)     | **Missing**                                  | 04        | Noop stub                                     |
| CR-SEC-001 (Classification enforcement) | **Complete**                                 | —         | —                                             |
| CR-SEC-002 (mTLS all services)          | **Partial** — svc-track has TODO             | 03        | Wire mTLS in svc-track                        |
| CR-SEC-003 (Immutable audit)            | **Complete**                                 | 03        | Wire audit emitters in svc-feedback/svc-query |
| CR-SEC-004..007                         | **Complete**                                 | —         | —                                             |
| CR-SEC-008 (Cert-based auth)            | **Partial** — dev uses insecure              | 03        | mTLS wiring                                   |

### 4.2 Use Case Coverage

| Use Case                    | Status Before v1 | v1 Module | v1 Task                     |
| --------------------------- | ---------------- | --------- | --------------------------- |
| UC001 (System Init)         | **Complete**     | 01        | Fix infra configs           |
| UC002 (Radar Ingestion)     | **Complete**     | —         | —                           |
| UC003 (EW Ingestion)        | **Partial**      | 02        | Elevate svc-ew-ingestion    |
| UC004 (ELINT Ingestion)     | **Partial**      | 02        | Elevate svc-elint-ingestion |
| UC005 (ISR Ingestion)       | **Partial**      | 02        | Elevate svc-isr-ingestion   |
| UC006 (AIS Ingestion)       | **Partial**      | 02        | Elevate svc-ais-ingestion   |
| UC007 (Cyber Ingestion)     | **Partial**      | 02        | Elevate svc-cyber-ingestion |
| UC008 (Multi-Source Fusion) | **Complete**     | —         | —                           |
| UC009 (Anomaly Detection)   | **Complete**     | —         | —                           |
| UC010 (Operator Feedback)   | **Complete**     | —         | —                           |
| UC011 (Model Retraining)    | **Missing**      | 04        | Training pipeline noop stub |
| UC012 (SA UI Display)       | **Complete**     | 05        | Browser E2E validates       |
| UC013 (Historical Query)    | **Complete**     | —         | —                           |
| UC014 (NATO Outbound)       | **Missing**      | 04        | NATO adapter noop stub      |
| UC015 (NATO Inbound)        | **Missing**      | 04        | NATO adapter noop stub      |

### 4.3 Feature Coverage

| Feature                           | Status Before v1 | v1 Module                         |
| --------------------------------- | ---------------- | --------------------------------- |
| FEAT-01 Platform Infrastructure   | **Complete**     | 01 (bug fixes)                    |
| FEAT-02 Security Framework        | **Partial**      | 03 (mTLS, audit)                  |
| FEAT-03 Event Streaming Backbone  | **Complete**     | 01 (topic name fixes)             |
| FEAT-04 Radar Sensor Ingestion    | **Complete**     | —                                 |
| FEAT-05 EW/SIGINT Ingestion       | **Partial**      | 02                                |
| FEAT-06 ELINT/COMINT Ingestion    | **Partial**      | 02                                |
| FEAT-07 ISR Metadata Ingestion    | **Partial**      | 02                                |
| FEAT-08 AIS/BFT Ingestion         | **Partial**      | 02                                |
| FEAT-09 Cyber Threat Ingestion    | **Partial**      | 02                                |
| FEAT-10 Multi-Source Fusion       | **Complete**     | —                                 |
| FEAT-11 Anomaly Detection         | **Complete**     | —                                 |
| FEAT-12 Operator Feedback & Trust | **Complete**     | 03 (audit emitter)                |
| FEAT-13 Situational Awareness UI  | **Partial**      | 03 (NVG, sensor health), 05 (E2E) |
| FEAT-14 Historical Analysis       | **Complete**     | —                                 |
| FEAT-15 NATO Data Exchange        | **Missing**      | 04 (noop stub)                    |

---

## 5. Critical Infrastructure Bugs (Found During Audit)

These **block the demo** and must be fixed first (Module 01):

| #      | Bug                                                                                                                                                          | Severity     | Impact                                                                        |
| ------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------ | ----------------------------------------------------------------------------- |
| BUG-01 | Envoy upstream ports use host-mapped ports (50070-50073, 50062) instead of container port 50051                                                              | **Critical** | Envoy cannot reach any backend service in Docker networking                   |
| BUG-02 | Redpanda topic names in `init-topics.sh` don't match service code (`sensor.raw.radar` vs `sensors.radar.tracks`) — 8 of 11 topics mismatched                 | **Critical** | ETL pipelines read from empty topics; services produce to non-existent topics |
| BUG-03 | ClickHouse table names in `init-clickhouse.sh` don't match Redpanda Connect ETL targets (`sensor_events` vs `sensor_observations`) — all 5 tables mismatched | **Critical** | All ETL inserts fail; queries return empty results                            |
| BUG-04 | ClickHouse table column schemas don't match Redpanda Connect field mappings                                                                                  | **High**     | Even with fixed table names, column inserts would fail                        |
| BUG-05 | Envoy service declared in infra compose but `depends_on` references service overlay                                                                          | **Medium**   | `docker compose -f docker-compose.yml up` fails                               |
| BUG-06 | TypeScript audit service codegen missing (`audit_service_pb.ts`, `audit_service_connect.ts`)                                                                 | **Medium**   | COP web app cannot call AuditService                                          |

---

## 6. Implementation Progress Tracker

| Step | Module | Task                                             | Status    |
| ---- | ------ | ------------------------------------------------ | --------- |
| 1    | 01     | Fix Envoy upstream ports                         | ☐ Pending |
| 2    | 01     | Unify Redpanda topic names                       | ☐ Pending |
| 3    | 01     | Unify ClickHouse table/column names              | ☐ Pending |
| 4    | 01     | Move Envoy to services compose                   | ☐ Pending |
| 5    | 01     | Generate missing TS audit client                 | ☐ Pending |
| 6    | 01     | ETL integration validation test                  | ☐ Pending |
| 7    | 02     | Elevate svc-ew-ingestion                         | ☐ Pending |
| 8    | 02     | Elevate svc-elint-ingestion                      | ☐ Pending |
| 9    | 02     | Elevate svc-isr-ingestion                        | ☐ Pending |
| 10   | 02     | Elevate svc-ais-ingestion                        | ☐ Pending |
| 11   | 02     | Elevate svc-cyber-ingestion                      | ☐ Pending |
| 12   | 02     | Add handler/producer/enricher tests (5 services) | ☐ Pending |
| 13   | 02     | Add integration tests (5 services)               | ☐ Pending |
| 14   | 03     | Wire mTLS in svc-track                           | ☐ Pending |
| 15   | 03     | Wire audit emitter in svc-feedback               | ☐ Pending |
| 16   | 03     | Wire audit emitter in svc-query                  | ☐ Pending |
| 17   | 03     | Enforce classification headers (all files)       | ☐ Pending |
| 18   | 03     | Sensor coverage health endpoint (CR-UI-004)      | ☐ Pending |
| 19   | 03     | NVG dark mode CSS (CR-UI-007)                    | ☐ Pending |
| 20   | 04     | Create svc-nato-adapter noop stub                | ☐ Pending |
| 21   | 04     | Create svc-training noop stub                    | ☐ Pending |
| 22   | 04     | NATO stub unit tests                             | ☐ Pending |
| 23   | 04     | Training stub unit tests                         | ☐ Pending |
| 24   | 05     | Set up Playwright for web-cop                    | ☐ Pending |
| 25   | 05     | Write browser E2E tests (10+ cases)              | ☐ Pending |
| 26   | 05     | Maritime demo scenario YAML                      | ☐ Pending |
| 27   | 05     | Multi-domain demo scenario YAML                  | ☐ Pending |
| 28   | 05     | Demo launch/stop scripts                         | ☐ Pending |
| 29   | 05     | Negative E2E tests                               | ☐ Pending |
| 30   | 05     | Extended integration tests (simulator path)      | ☐ Pending |

---

## 7. Definition of Done — v1 Complete

All v1 work is **complete** when:

- [ ] `make docker-up-all` succeeds — all services healthy via `make health-check`
- [ ] Envoy proxies gRPC-Web requests to all 5 backend services (tracks, alerts, query, feedback, audit)
- [ ] All 6 ingestion services produce to real Redpanda topics (not LogProducer)
- [ ] All 5 Redpanda Connect ETL pipelines successfully populate ClickHouse tables
- [ ] `make test` passes for all 15 services (13 existing + 2 new stubs)
- [ ] `make test-coverage` shows ≥80% line coverage across all services
- [ ] `make integration-test` passes (14 existing + new tests)
- [ ] Browser E2E tests pass (`npx playwright test`)
- [ ] `scripts/demo/run-demo.sh maritime-demo` runs full pipeline end-to-end
- [ ] `grep -rL "// CLASSIFICATION:" svc-*/` returns zero results
- [ ] `grep -r "noopAuditEmitter" svc-*/` returns zero results
- [ ] `svc-nato-adapter/` and `svc-training/` exist with passing unit tests
- [ ] No hardcoded secrets, credentials, or classified data in any file

---

## 8. Agent Invocation Template

To invoke the Greatest Ever Developer agent for any v1 module:

```
@greatest-ever-developer Implement v1 Module XX from docs/implementation/v1/XX-module-name.md

Context:
- Read docs/implementation/v1/00-v1-overview.md for v1 scope, traceability, and progress
- Read docs/implementation/00-implementation-overview.md for global conventions and interface contracts
- Read docs/implementation/v1/XX-module-name.md for complete implementation specification
- Follow all SDLC guidelines in docs/sdlc_guidelines/
- Reference architecture docs in docs/architecture/ for design decisions
- Target: development environment (Docker Compose, not K8s)
- This is COMPLETION work — existing code is authoritative; do NOT rewrite working services

Deliverables:
1. All source files as specified in the module document
2. Unit tests with ≥80% line coverage
3. Integration tests where specified
4. All existing tests continue to pass
5. Classification header present on every new/modified file
6. All tests passing: go test ./... -race -count=1
```

---

## 9. Reference Architecture

For full details, see the original implementation orchestration and architecture docs:

| Document                         | Path                                                |
| -------------------------------- | --------------------------------------------------- |
| Original Implementation Overview | `docs/implementation/00-implementation-overview.md` |
| Module 04 — Radar Reference Impl | `docs/implementation/04-sensor-ingestion-radar.md`  |
| Module 05 — Batch Ingestion      | `docs/implementation/05-sensor-ingestion-batch.md`  |
| High-Level Architecture          | `docs/architecture/high_level_architecture.md`      |
| Component Design                 | `docs/architecture/component_design.md`             |
| Security Architecture            | `docs/architecture/security_architecture.md`        |
| Business Requirements            | `docs/business/requirements.md`                     |
| Feature List                     | `docs/business/feature_list.md`                     |
| Use Cases                        | `docs/business/usecases/UC*.md`                     |
