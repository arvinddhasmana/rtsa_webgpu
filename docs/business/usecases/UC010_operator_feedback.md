<!-- CLASSIFICATION: UNCLASSIFIED -->
# UC010 — Operator Feedback Submission

> **Use Case ID**: UC010
> **Feature**: FEAT-12 (Operator Feedback & Trust Scoring)
> **Priority**: MUST
> **Actors**: Operations Commander, Intelligence Analyst
> **Classification**: UNCLASSIFIED
> **Last Updated**: 2026-02-23

---

## 1. Description

An operator reviews an anomaly detection alert or entity track and submits feedback to confirm, reject, or reclassify the assessment. The feedback is trust-scored using the operator's clearance, historical accuracy, temporal consistency, and statistical deviation from model consensus.

## 2. Preconditions

- Operator is authenticated via certificate (mTLS)
- Anomaly detection is producing scored tracks (UC009)
- Situational awareness UI is displaying entity tracks and anomaly alerts

## 3. Triggers

- Operator views an anomaly alert and decides to provide feedback
- Operator observes an entity track needing reclassification

## 4. Main Flow

```mermaid
sequenceDiagram
    actor OP as Operator
    participant UI as Situational<br/>Awareness UI
    participant FB as Feedback Service
    participant TRUST as Trust Scoring<br/>Engine
    participant RP as Redpanda<br/>feedback.operator.submissions
    participant RPV as Redpanda<br/>feedback.operator.validated
    participant AUDIT as Audit Trail

    OP->>UI: Select entity track / anomaly alert
    OP->>UI: Submit feedback<br/>(confirm hostile, reject anomaly, etc.)
    UI->>FB: gRPC: SubmitFeedback(FeedbackRequest)

    FB->>FB: Validate feedback fields
    FB->>TRUST: Calculate trust score
    TRUST->>TRUST: Evaluate:<br/>1. Operator clearance (0.2)<br/>2. Historical accuracy (0.3)<br/>3. Temporal consistency (0.2)<br/>4. Statistical deviation (0.3)
    TRUST-->>FB: Trust score + breakdown

    FB->>RP: Produce to feedback.operator.submissions
    FB->>AUDIT: Audit: feedback_submitted

    alt Trust score >= 0.5
        FB->>RPV: Produce to feedback.operator.validated
        FB-->>UI: Feedback accepted
    else Trust score 0.2–0.49
        FB->>FB: Flag for human review
        FB-->>UI: Feedback submitted for review
    else Trust score < 0.2
        FB->>FB: Reject feedback; alert SecOps
        FB->>AUDIT: Audit: feedback_rejected_low_trust
        FB-->>UI: Feedback rejected (low trust)
    end
```

## 5. Feedback Types

| Type | Description | Example |
|---|---|---|
| CONFIRM_HOSTILE | Operator confirms entity is hostile | "Track TRK-4501 confirmed hostile" |
| CONFIRM_FRIENDLY | Operator confirms entity is friendly | "Track TRK-1200 is known friendly" |
| RECLASSIFY | Operator changes hostile status | "Reclassify TRK-3001 from UNKNOWN to NEUTRAL" |
| REJECT_ANOMALY | Operator rejects anomaly as false positive | "Speed anomaly on TRK-4501 is normal for this vessel" |
| CONFIRM_ANOMALY | Operator confirms anomaly is valid | "Behavioral anomaly on TRK-2200 is suspicious" |

## 6. Trust Score Calculation

| Factor | Weight | Input | Scoring |
|---|---|---|---|
| Clearance level | 0.2 | Operator's security clearance | UC=0.2, PA=0.4, PB=0.6, PC=0.8, SECRET=1.0 |
| Historical accuracy | 0.3 | Past feedback vs. ground truth | Rolling average of correctness |
| Temporal consistency | 0.2 | Time between observation and feedback | < 5 min = 1.0, < 1h = 0.7, > 1h = 0.3 |
| Statistical deviation | 0.3 | How far from model consensus | Low deviation = high score |

**Formula**: $\text{Trust} = 0.2 \cdot C + 0.3 \cdot A + 0.2 \cdot T + 0.3 \cdot (1 - D)$

## 7. Alternative Flows

### 7a. Edge Deployment (Queued Feedback)
- Operator submits feedback at tactical edge
- Feedback stored locally and applied to local model
- Queued for sync to data centre when connectivity available
- Data centre reprocesses feedback with full trust scoring

### 7b. Bulk Anomalous Feedback (Anti-Poisoning)
- If operator submits > 10 feedback items with trust score < 0.5 in 1 hour
- System flags operator for review
- All pending feedback from operator held for human review
- Alert sent to Security Operations

## 8. Security Considerations

- Operator identity extracted from mTLS certificate (no impersonation)
- Complete audit trail for every feedback operation
- Feedback cannot modify historical records — only influences future scoring
- Anti-poisoning safeguards prevent model degradation
- Low-trust feedback never enters training pipeline automatically

## 9. Requirements Traced

| Requirement | Description |
|---|---|
| CR-FB-001 | Operators can confirm or reject anomaly detections |
| CR-FB-002 | Operators can reclassify entity hostile status |
| CR-FB-003 | Trust scores assigned to all feedback |
| CR-FB-004 | Reject feedback that fails anti-poisoning validation |
| CR-FB-005 | Route validated feedback to model retraining |
| CR-FB-006 | Complete audit trail for all feedback |
| CR-FB-007 | Queue feedback at edge; sync when connected |
