<!-- CLASSIFICATION: UNCLASSIFIED -->

# Phase 2 — Intelligence & Detection Validation

> **Use Cases**: UC009 (Anomaly Detection), UC010 (Operator Feedback), UC011 (Model Retraining)
> **Prerequisites**: Phase 1 complete — all ingestion and fusion tests passing
> **Common Guidelines**: See [00_common_guidelines.md](00_common_guidelines.md)
> **Classification**: UNCLASSIFIED
> **Date**: 2026-03-10

---

## 1. Objectives

- Verify anomaly detection engine scores fused tracks correctly with proper alert thresholds
- Verify operator feedback flow: submission → trust scoring → anti-poisoning → routing
- Verify the AssignAlert RPC (v2.0) produces audit events
- Verify model retraining pipeline consumes validated feedback
- Run all related unit, integration, E2E, and benchmark tests — zero silent failures
- Compile all issues into Issue Log, then fix in batch

---

## 2. UC009 — Anomaly Detection & Inference

**Spec**: `docs/business/usecases/UC009_anomaly_detection.md`
**Feature**: FEAT-11 (Anomaly Detection & Inference)
**Requirements**: CR-INF-001 through CR-INF-007

### 2.1 Code Review Checklist

| # | Review Item | File(s) |
|---|-------------|---------|
| 1 | Feature vector preparation from fused track (position, kinematics, type, history) | `svc-anomaly-detection/internal/domain/` |
| 2 | ML model inference call and response handling | domain logic |
| 3 | Anomaly types match spec: MOVEMENT_ANOMALY, LOCATION_ANOMALY, SPEED_ANOMALY, BEHAVIORAL_ANOMALY, THREAT_INDICATOR, CORRELATION_ANOMALY | domain + proto |
| 4 | Alert thresholds: NORMAL (0–0.3), WATCH (0.3–0.6), ELEVATED (0.6–0.8), CRITICAL (0.8–1.0) | domain logic |
| 5 | Severity routing to correct Redpanda topics (CRITICAL/ELEVATED/WATCH/NORMAL) per IT10 | producer code |
| 6 | Human-readable explanation generation (CR-INF-003) | domain logic |
| 7 | Model version included in output (CR-INF-007) | proto/handler |
| 8 | Degraded mode: model not available → score 0.0, alert to admin | fallback logic |
| 9 | Pre-trained model usage at edge (CR-INF-006) | config/deployment |
| 10 | Output AnomalyScore event contains all fields from UC009 §6 | proto definitions |
| 11 | Audit event produced for anomaly_detected | producer code |
| 12 | Unit tests cover: speed anomaly, AIS manipulation, severity routing, degraded mode | `*_test.go` files |

### 2.2 Tests to Run

```bash
# Unit tests
cd svc-anomaly-detection && go test -race -count=1 -v ./...

# Coverage (target: ≥90% for domain logic)
cd svc-anomaly-detection && go test -race -coverprofile=coverage.out ./... && \
  go tool cover -func=coverage.out | tail -1

# Integration tests: IT08, IT09, IT10
bash scripts/dev/test-integration.sh
# Verify IT08 (speed anomaly >3σ), IT09 (AIS manipulation >0.5NM), IT10 (severity routing)

# Benchmark: B03 (Anomaly p95 ≤ 150ms)
bash scripts/dev/test-bench.sh
```

### 2.3 Expected Outcomes

- Anomaly detection domain logic ≥90% line coverage
- IT08, IT09, IT10 all pass
- B03: Anomaly inference latency p95 ≤ 150ms
- All 6 anomaly types properly defined and tested

---

## 3. UC010 — Operator Feedback Submission

**Spec**: `docs/business/usecases/UC010_operator_feedback.md`
**Feature**: FEAT-12 (Operator Feedback & Trust Scoring)
**Requirements**: CR-FB-001 through CR-FB-008

### 3.1 Code Review Checklist

| # | Review Item | File(s) |
|---|-------------|---------|
| 1 | `SubmitFeedback` RPC handler validates all fields | `svc-feedback/internal/handler/` |
| 2 | All 5 feedback types supported: CONFIRM_HOSTILE, CONFIRM_FRIENDLY, RECLASSIFY, REJECT_ANOMALY, CONFIRM_ANOMALY | handler + proto |
| 3 | Trust score formula: `0.2·C + 0.3·A + 0.2·T + 0.3·(1-D)` matches spec exactly | `svc-feedback/internal/domain/trust.go` or equivalent |
| 4 | Clearance level weights: UC=0.2, PA=0.4, PB=0.6, PC=0.8, SECRET=1.0 | trust scoring |
| 5 | Temporal consistency: <5min=1.0, <1h=0.7, >1h=0.3 | trust scoring |
| 6 | Trust ≥0.5 → produce to `feedback.operator.validated` | routing logic |
| 7 | Trust 0.2–0.49 → flag for human review | routing logic |
| 8 | Trust <0.2 → reject, alert SecOps | routing logic |
| 9 | Anti-poisoning: >10 low-trust in 1h → flag operator, hold pending feedback (CR-FB-004) | anti-poisoning guard |
| 10 | Audit event for every feedback operation (CR-FB-006) | producer code |
| 11 | **AssignAlert RPC** in `svc-alert`: assigns alert to operator, produces audit event (CR-FB-008) | `svc-alert/internal/handler/` |
| 12 | AssignAlert response includes `assigned_at` timestamp | handler response |
| 13 | Trust scoring tests achieve ≥95% coverage | `*_test.go` files |
| 14 | Anti-poisoning tests achieve ≥95% coverage | `*_test.go` files |

### 3.2 Tests to Run

```bash
# Feedback service unit tests
cd svc-feedback && go test -race -count=1 -v ./...

# Alert service unit tests (includes AssignAlert)
cd svc-alert && go test -race -count=1 -v ./...

# Coverage — CRITICAL: trust scoring + anti-poisoning must be ≥95%
cd svc-feedback && go test -race -coverprofile=coverage.out ./... && \
  go tool cover -func=coverage.out
cd svc-alert && go test -race -coverprofile=coverage.out ./... && \
  go tool cover -func=coverage.out

# Integration: feedback pipeline
bash scripts/dev/test-integration.sh
# Verify tests/integration/feedback_test.go passes

# E2E03: Feedback → Trust → Validated
bash scripts/dev/test-e2e.sh
# Verify E2E03 passes

# Negative: Neg02 (anti-poisoning rejects feedback)
# Part of E2E negative suite
```

### 3.3 Expected Outcomes

- Trust scoring logic ≥95% line coverage
- Anti-poisoning logic ≥95% line coverage
- E2E03 passes — complete feedback workflow
- Neg02 passes — anti-poisoning correctly rejects low-trust bulk feedback
- AssignAlert produces audit event

---

## 4. UC011 — Model Retraining

**Spec**: `docs/business/usecases/UC011_model_retraining.md`
**Feature**: FEAT-12 (validated feedback → retraining pipeline)
**Requirements**: CR-FB-005

### 4.1 Code Review Checklist

| # | Review Item | File(s) |
|---|-------------|---------|
| 1 | Training service consumes from `feedback.operator.validated` topic | `svc-training/` |
| 2 | Batch threshold for triggering retrain is configurable | config |
| 3 | Only high-trust (≥0.5) feedback enters training pipeline | consumer filter |
| 4 | Model version incremented after retrain | domain logic |
| 5 | Audit event for retrain operation | producer code |
| 6 | Edge mode: no live training, pre-trained only (CR-INF-006) | config/deployment |
| 7 | Unit tests for KafkaClient interface and training loop | `*_test.go` files |

### 4.2 Tests to Run

```bash
# Unit tests
cd svc-training && go test -race -count=1 -v ./...

# Coverage
cd svc-training && go test -race -coverprofile=coverage.out ./... && \
  go tool cover -func=coverage.out | tail -1

# Integration: training pipeline
bash scripts/dev/test-integration.sh
# Verify tests/integration/training_test.go passes
```

### 4.3 Expected Outcomes

- Training service unit tests pass with ≥80% coverage
- Integration test verifies validated feedback reaches training consumer
- Batch threshold triggers retrain correctly

---

## 5. E2E Flow Validation (Full Phase 2 Flow)

After all use cases above are individually reviewed:

```bash
# E2E02: Alert workflow — generate → stream → acknowledge
# E2E03: Feedback workflow — submit → trust score → validate
bash scripts/dev/test-e2e.sh

# Negative tests
# Neg02: Anti-poisoning rejects feedback
# (part of E2E negative suite in tests/e2e/negative_test.go)
```

### Expected Outcomes

- E2E02 and E2E03 pass
- Neg02 passes — anti-poisoning correctly blocks adversarial feedback
- No silent test failures (check `test-results/e2e/failures.txt`)

---

## 6. Issue Log

_Issues found during Phase 2 review are recorded here and fixed in batch._

| # | UC | Severity | File(s) | Issue Description | Requirement |
|---|----|----------|---------|-------------------|-------------|
| | | | | _(to be populated during execution)_ | |

---

## 7. Batch Fix Execution

After the Issue Log is populated:

1. Fix all BLOCKING issues first
2. Fix WARNING issues
3. Fix IMPROVEMENT items
4. Re-run ALL tests from sections 2–5 to confirm no regressions

---

## 8. Phase 2 Completion Criteria

- [ ] UC009: Anomaly detection passes unit tests with ≥90% domain coverage
- [ ] UC009: IT08, IT09, IT10 pass; B03 passes threshold
- [ ] UC010: Feedback + Alert services pass unit tests; trust scoring ≥95% coverage
- [ ] UC010: AssignAlert RPC produces audit events
- [ ] UC010: E2E03 passes; Neg02 passes
- [ ] UC011: Training service passes unit tests with ≥80% coverage
- [ ] E2E02 (alert workflow) passes
- [ ] All Issue Log items resolved
- [ ] No silent test failures in any stage
