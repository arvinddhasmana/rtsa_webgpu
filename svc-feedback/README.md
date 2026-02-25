<!-- CLASSIFICATION: UNCLASSIFIED -->

# svc-feedback — Feedback & Trust Scoring Service

> **Module**: 09 — Feedback & Trust Scoring  
> **Phase**: P2 (Core Processing)  
> **Classification**: UNCLASSIFIED artifact for Protected C / Secret system

---

## Overview

`svc-feedback` is the RTSA Feedback & Trust Scoring service. It receives operator feedback on entity classifications and anomaly alerts via gRPC, computes a multi-component trust score, applies anti-poisoning guards, and publishes events to Redpanda topics for downstream consumption by the reinforcement learning pipeline.

---

## Trust Formula

$$Trust = 0.2 \times C + 0.3 \times A + 0.2 \times T + 0.3 \times (1-D)$$

| Component | Symbol | Description |
|---|---|---|
| Clearance | C | Operator's clearance level mapped to [0.3, 1.0] |
| Accuracy | A | Historical confirmed-correct ratio (default 0.5 for new operators) |
| Temporal | T | Decay based on time between event and feedback (1.0 → 0.1) |
| Deviation | (1-D) | Alignment with consensus feedback for the same track |

---

## Anti-Poisoning Guard (5 Checks)

| Check | Threshold | Min Submissions |
|---|---|---|
| Distribution | Chi-squared p > 0.05 | ≥ 10 |
| Label Flip Rate | ≤ 20% | ≥ 5 |
| Source Diversity | ≥ 3 unique sensors | ≥ 5 |
| Temporal Clustering | No 5-min window > 40% | Always |
| High-Trust Ratio | ≥ 60% validated | ≥ 10 |

Anti-poisoning failure **does not block** submission. It forces `trust_score < 0.5`, preventing routing to the validated topic.

---

## Topics

| Topic | Content |
|---|---|
| `feedback.operator.submissions` | All feedback, regardless of trust score |
| `feedback.operator.validated` | Only feedback with `trust_score ≥ 0.5` AND anti-poison passed |
| `audit.events` | Audit trail for every feedback action |

---

## Rate Limiting

Per-operator sliding window: **10 requests per 60 seconds**.  
Exceeding the limit returns `codes.ResourceExhausted`.

---

## Configuration

All configuration is sourced from environment variables (no hardcoded secrets):

| Variable | Default | Description |
|---|---|---|
| `RTSA_GRPC_PORT` | `50051` | gRPC listener port |
| `RTSA_HEALTH_PORT` | `8081` | HTTP health endpoint port |
| `RTSA_REDPANDA_BROKERS` | `localhost:19092` | Comma-separated broker list |
| `RTSA_RATE_LIMIT_PER_MINUTE` | `10` | Per-operator request limit |
| `RTSA_LOG_LEVEL` | `info` | Log level (`info` or `debug`) |
| `RTSA_SERVICE_NAME` | `svc-feedback` | Service identifier in audit events |

---

## Running

```bash
# Build
go build ./cmd/feedback

# Run (default config)
./feedback

# Run with custom config
RTSA_GRPC_PORT=50052 RTSA_REDPANDA_BROKERS=redpanda:19092 ./feedback
```

## Testing

```bash
# Unit tests with race detector
go test ./... -race -count=1

# Coverage report
go test ./... -cover

# Integration tests (requires Redpanda)
go test -tags=integration ./tests/integration/...
```

---

## Service Structure

```
svc-feedback/
├── cmd/feedback/main.go          — Service entrypoint and wiring
├── internal/
│   ├── config/config.go          — Environment-based configuration
│   ├── domain/
│   │   ├── trust_scorer.go       — 4-component trust formula
│   │   ├── anti_poison.go        — 5 statistical anti-poisoning checks
│   │   └── rate_limiter.go       — Per-operator sliding window rate limiter
│   ├── handler/feedback.go       — gRPC FeedbackService implementation
│   ├── producer/feedback_producer.go — Redpanda message producer + mock
│   └── state/operator_history.go — Thread-safe in-memory operator history
├── tests/integration/            — Integration tests (build tag: integration)
├── Dockerfile                    — Multi-stage distroless build
└── README.md
```

---

## Security Notes

- No credentials or secrets are hardcoded; all config via environment variables
- In production, the gRPC server must use mTLS (`grpc.Creds(credentials.NewTLS(...))`)
- The distroless runtime image has no shell, minimising attack surface
- All logs use structured logging; no raw sensor data or PII is logged
- Anti-poisoning checks protect the ML training pipeline from data poisoning

---

## SDLC References

- Coding Standards: `docs/sdlc_guidelines/04_coding_standards/`
- Security Classification: `docs/sdlc_guidelines/01_security_compliance/security_classification.md`
- Testing Strategy: `docs/sdlc_guidelines/05_testing/testing_strategy.md`
- gRPC Guidelines: `docs/sdlc_guidelines/08_tech_specific/grpc_service_guidelines.md`
- Redpanda Guidelines: `docs/sdlc_guidelines/08_tech_specific/redpanda_guidelines.md`
