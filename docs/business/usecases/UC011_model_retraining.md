<!-- CLASSIFICATION: UNCLASSIFIED -->
# UC011 — Feedback-Driven Model Retraining

> **Use Case ID**: UC011
> **Feature**: FEAT-13 (Feedback-Driven Model Retraining)
> **Priority**: MUST
> **Actors**: ML Pipeline (automated), Security Officer (oversight)
> **Classification**: UNCLASSIFIED
> **Last Updated**: 2026-02-23

---

## 1. Description

Validated operator feedback is consumed from Redpanda, aggregated into training batches, and used to retrain anomaly detection models. The retraining pipeline enforces anti-poisoning safeguards, validates model quality against holdout datasets, and promotes models only after human sign-off. Edge-deployed models receive updates via model distribution topics.

## 2. Preconditions

- Validated feedback exists in `feedback.operator.validated` (UC010)
- Minimum batch of 50 validated feedback items accumulated
- Existing baseline model deployed and producing predictions

## 3. Triggers

- Batch threshold reached (50+ validated feedback items)
- Scheduled retraining window (weekly)
- Manual trigger by authorized ML engineer

## 4. Main Flow

```mermaid
sequenceDiagram
    participant RPV as Redpanda<br/>feedback.operator.validated
    participant AGG as Feedback<br/>Aggregator
    participant VALID as Anti-Poisoning<br/>Validator
    participant TRAIN as Training<br/>Pipeline
    participant EVAL as Model<br/>Evaluator
    participant REG as Model<br/>Registry
    participant RPM as Redpanda<br/>models.anomaly.published
    participant AUDIT as Audit Trail

    RPV->>AGG: Consume validated feedback
    AGG->>AGG: Accumulate batch (≥ 50 items)
    AGG->>VALID: Submit training batch

    VALID->>VALID: Anti-poisoning checks:<br/>1. Distribution analysis<br/>2. Label flip detection<br/>3. Source diversity check
    alt Batch passes validation
        VALID->>TRAIN: Forward clean batch
    else Batch fails validation
        VALID->>AUDIT: Audit: batch_rejected_poisoning
        VALID-->>AGG: Quarantine batch
    end

    TRAIN->>TRAIN: Split: 70% train / 15% val / 15% holdout
    TRAIN->>TRAIN: Retrain model (incremental)
    TRAIN->>EVAL: Submit candidate model

    EVAL->>EVAL: Evaluate on holdout set
    EVAL->>EVAL: Compare to baseline:<br/>- Accuracy delta ≥ -2%<br/>- Precision ≥ baseline<br/>- False positive rate ≤ baseline + 5%

    alt Candidate passes evaluation
        EVAL->>REG: Register candidate model (STAGED)
        EVAL->>AUDIT: Audit: model_candidate_staged
        Note over REG: Awaits human sign-off
        REG->>REG: Human approval received
        REG->>RPM: Publish model artifact reference
        REG->>AUDIT: Audit: model_promoted_production
    else Candidate fails evaluation
        EVAL->>AUDIT: Audit: model_candidate_rejected
        EVAL->>TRAIN: Log failure metrics
    end
```

## 5. Anti-Poisoning Validation Rules

| Check | Description | Threshold |
|---|---|---|
| Distribution analysis | Compare label distribution to historical baseline | Chi-squared test, p > 0.05 |
| Label flip detection | Identify bulk label reversals | Max 20% of batch may flip a single label |
| Source diversity | Ensure feedback from multiple operators | Minimum 3 distinct operators per batch |
| Temporal clustering | Detect suspiciously timed submissions | No single 5-min window contributes > 40% |
| High-trust ratio | Minimum proportion of high-trust items | ≥ 60% of batch must have trust ≥ 0.7 |

## 6. Model Evaluation Criteria

| Metric | Requirement |
|---|---|
| Accuracy | Must not decrease by > 2% vs. baseline |
| Precision (hostile) | Must be ≥ baseline |
| Recall (hostile) | Must be ≥ 90% of baseline |
| False positive rate | Must not increase by > 5% |
| Inference latency | Must remain < 200ms (p99) |
| Model size | Must fit within edge constraints (< 500MB) |

## 7. Edge Model Distribution

```mermaid
sequenceDiagram
    participant REG as Model Registry
    participant RPM as Redpanda<br/>models.anomaly.published
    participant EDGE as Edge Node<br/>Model Loader
    participant INF as Edge Inference<br/>Engine
    participant AUDIT as Audit Trail

    RPM->>EDGE: Model artifact reference received
    EDGE->>REG: Download model artifact
    EDGE->>EDGE: Verify signature (cosign)
    EDGE->>EDGE: Validate size constraints
    EDGE->>INF: Hot-swap model (blue-green)
    INF->>INF: Run validation inference on test set
    alt Validation passes
        INF->>INF: Activate new model
        INF->>AUDIT: Audit: edge_model_updated
    else Validation fails
        INF->>INF: Rollback to previous model
        INF->>AUDIT: Audit: edge_model_rollback
    end
```

## 8. Alternative Flows

### 8a. Insufficient Feedback Volume
- Batch threshold not reached within retraining window
- System logs warning and extends window
- No retraining occurs; existing model continues unchanged

### 8b. Model Degradation Detected Post-Deployment
- Monitoring detects performance degradation (false positive spike)
- Automatic rollback to previous model version
- Degraded model quarantined; contributing feedback batch flagged for review
- Alert sent to ML engineer and Security Officer

### 8c. Edge-Only Retraining (Disconnected)
- Edge accumulates validated feedback locally
- Lightweight incremental update applied to local model
- When connectivity restored, feedback synced to central for full retraining

## 9. Security Considerations

- Model artifacts must be cryptographically signed (cosign)
- Only promoted models deployed to production/edge
- Complete lineage: feedback items → training batch → model version
- Anti-poisoning prevents adversarial manipulation of model behavior
- All model lifecycle events produce immutable audit records

## 10. Requirements Traced

| Requirement | Description |
|---|---|
| CR-FB-005 | Route validated feedback to model retraining |
| CR-INF-005 | Support model updates without service interruption |
| CR-INF-006 | Log all model versions, inputs, and outputs |
| CR-INF-007 | Detect model degradation and alert operators |
| CR-SEC-006 | Anti-poisoning controls on training data |
| CR-SEC-008 | Immutable audit log for model lifecycle |
