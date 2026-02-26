<!-- CLASSIFICATION: UNCLASSIFIED -->

# Feedback Integrity Review

> **CLASSIFICATION: UNCLASSIFIED**
> **Audience**: Security Officers, Security Authorities
> **Version**: 1.0
> **Last Updated**: 2026-02-26

---

## Overview

RTSA's **Human-in-the-Loop feedback system** allows operators to confirm, reject, and reclassify AI detections. This feedback drives model improvement — which makes the integrity of that feedback critically important. Malicious, erroneous, or coordinated bad feedback could degrade the AI model's accuracy over time.

The system defends against this with **anti-poisoning safeguards** that automatically flag suspicious feedback patterns and hold them for Security Officer review before they can enter the model training pipeline.

This guide explains how to review the feedback queue, investigate flagged submissions, and clear or escalate them.

---

## The Trust Scoring System

Every feedback submission is scored by the trust engine before it is accepted:

| Factor | Weight | Description |
|---|---|---|
| **Security Clearance** | 20% | Higher clearance operators have higher base trust |
| **Historical Accuracy** | 30% | Rolling accuracy of past feedback vs. ground truth |
| **Temporal Consistency** | 20% | How quickly after observation the feedback was submitted |
| **Statistical Deviation** | 30% | How far the feedback deviates from model/consensus assessment |

### Routing Based on Trust Score

| Score | Route | Your Role |
|---|---|---|
| ≥ 0.5 | Directly to training pipeline | No review needed |
| 0.2–0.49 | Held in review queue | You decide: accept or reject |
| < 0.2 | Rejected; you are alerted | Investigate and document |

---

## Accessing the Feedback Review Queue

1. Click 🔒 **Audit** in the toolbar
2. Select the **Feedback Integrity** tab
3. The queue shows all feedback items currently held for review

---

## Reading a Feedback Review Item

```
┌──────────────────────────────────────────────────────────────┐
│ FEEDBACK HELD FOR REVIEW                                     │
│                                                              │
│ Submitted by:  CN=Jones.R.K, O=CAF, OU=3 Svc Bn            │
│ Submitted at:  2026-02-20 14:38:02 UTC                       │
│ Entity:        TRK-7302 (SURFACE, SUSPECT)                   │
│ Feedback type: REJECT_ANOMALY                                │
│ Anomaly score: 0.86 (CRITICAL)                               │
│ Trust score:   0.34 ⚠️ (MEDIUM — held for review)           │
│                                                              │
│ Reason for hold: Statistical deviation = 0.72                │
│ (Feedback strongly disagrees with model consensus of 0.86)  │
│                                                              │
│ Operator comment: "This vessel is known to us and is         │
│  operating under special authority. Not hostile."            │
│                                                              │
│ Operator history:                                            │
│  - 47 past feedback submissions                              │
│  - 38 confirmed correct (81%)                               │
│  - 9 pending ground truth verification                       │
│                                                              │
│ [✅ Accept] [❌ Reject] [🔍 Investigate] [📋 Escalate]       │
└──────────────────────────────────────────────────────────────┘
```

---

## Evaluating a Held Feedback Item

When reviewing a held feedback item, consider:

**1. Does the operator's comment provide credible context?**
A strong, specific explanation ("This vessel is operating under special authority per [reference]") warrants more weight than vague or absent justification.

**2. What is the operator's historical accuracy?**
- 80%+ historical accuracy: Operator is generally reliable; higher weight on their explanation
- < 50% accuracy: Investigate more carefully; pattern of errors

**3. What is the anomaly score?**
- Feedback rejecting a CRITICAL anomaly (score > 0.8) requires more scrutiny than feedback on a WATCH-level alert

**4. Is this isolated or part of a pattern?**
- Single held item: Likely human judgment difference; lower risk
- Multiple held items from same operator: Investigate for systematic error or malicious pattern

**5. Has this entity been previously reviewed?**
- Click 🔍 **Investigate** to see the full entity history and prior feedback

---

## Your Decisions

### Accept ✅
The feedback appears legitimate based on your investigation. It will be:
- Routed to the model training pipeline
- Logged in the audit trail with your identity and decision rationale

Use when: The operator's explanation is credible and their history is good.

---

### Reject ❌
The feedback appears erroneous, suspicious, or without sufficient justification. It will be:
- Discarded from the training pipeline
- Logged with your rejection rationale
- The operator is notified that their feedback was not accepted

Use when: The explanation is insufficient, or the feedback strongly conflicts with corroborating evidence.

---

### Escalate 📋
You need more information or senior authority to decide. An escalation:
- Suspends the feedback pending further review
- Notifies your chain of command
- Creates a formal review record in the audit trail

Use when: The feedback may have significant operational implications, or you cannot determine legitimacy from available information.

---

## Anti-Poisoning Alerts — Operator Paused

If an operator's feedback is automatically paused (bulk anomalous pattern detected), you will receive an alert:

```
⚠️ ANTI-POISONING ALERT
Operator: CN=Brown.T.E, O=CAF
Reason: 12 feedback submissions in 45 minutes with avg trust score 0.31
Action: All feedback from this operator paused pending SO review
```

### Responding to an Anti-Poisoning Alert

1. Open the Feedback Integrity tab — the operator's queue items appear highlighted in orange
2. Review their feedback pattern:
   - Are they rejecting many high-score anomalies?
   - Are they all on the same entity or entity type?
   - Is there an operational explanation (e.g., they are monitoring a specific vessel and legitimately rejecting false positives)?
3. Contact the operator directly if possible to understand their intent
4. Check the audit trail for their session — look for any other unusual activity
5. Decide:
   - **Clear the hold**: If explained and legitimate — operator feedback resumes normally
   - **Maintain the hold**: If pattern is concerning — send items to manual review one by one
   - **Escalate**: If suspected malicious or systematic manipulation

---

## Model Retraining Oversight

The feedback you approve is eventually used to retrain the AI model. To review what has been incorporated into training:

1. Go to ⚙️ **Settings → Security → Model Retraining Oversight**
2. View the training pipeline feed: quantity of feedback items, trust score distribution, anomaly types covered
3. Model updates are logged with:
   - Model version number
   - Timestamp
   - Quantity and type of training data
   - Your organisation's representative who approved the training run

If a model update appears to have degraded detection quality (increase in false positives or missed critical alerts), contact the system administrator to initiate a **model rollback** — the system supports rolling back to any previous model version.

---

> **Back to Role Overview**: [Security Officer Guide →](README.md)

---

> **CLASSIFICATION: UNCLASSIFIED** — This document contains no classified information.
