<!-- CLASSIFICATION: UNCLASSIFIED -->

# Operator Feedback — Improving AI Accuracy

> **CLASSIFICATION: UNCLASSIFIED**
> **Audience**: Operations Commanders, Watch Officers
> **Version**: 1.0
> **Last Updated**: 2026-02-26

---

## Overview

RTSA uses **operator feedback** to continuously improve its AI detection models. When you confirm or reject an anomaly detection, or reclassify an entity's hostile status, that information is captured, evaluated, and — if trustworthy — fed into the model improvement pipeline.

This is not just housekeeping. **Your professional judgment directly shapes how the system detects threats in the future.** Accurate, timely feedback makes RTSA better for everyone on your team and future operators.

---

## When Should You Submit Feedback?

Submit feedback whenever you:

- **Agree with the AI's assessment** — the anomaly is real and the entity is behaving suspiciously
- **Disagree with the AI's assessment** — the anomaly is a false positive (normal behaviour flagged incorrectly)
- **Want to reclassify an entity** — the AI has assigned the wrong hostile status
- **Observe something the AI missed** — you notice unusual behaviour that has not been flagged

---

## Feedback Types

| Feedback Type | When to Use |
|---|---|
| **CONFIRM_HOSTILE** | You have assessed that this entity is hostile. The AI's assessment aligns with your professional judgment. |
| **CONFIRM_FRIENDLY** | You can positively identify this entity as a known friendly unit. |
| **RECLASSIFY** | The current hostile status is wrong. You are changing it to a different category. |
| **REJECT_ANOMALY** | The anomaly alert is a false positive. The flagged behaviour is normal for this entity, area, or time. |
| **CONFIRM_ANOMALY** | The anomaly alert is valid. The entity's behaviour is genuinely suspicious. |

---

## How to Submit Feedback

### From the Alert Panel

1. Find the alert card in the alert panel
2. Click **[Confirm]** to confirm the anomaly, or **[Reject]** to reject it
3. An optional **reason / comment** box appears — provide context when helpful
4. Click **Submit**

### From the Entity Detail Panel

1. Click the entity on the map to open the detail panel
2. Scroll to the **Feedback** section at the bottom
3. Select the feedback type from the dropdown
4. Optionally select a new hostile status (for RECLASSIFY)
5. Add a comment (recommended for reclassifications)
6. Click **Submit Feedback**

### Keyboard Shortcut (when alert card is focused)

| Key | Action |
|---|---|
| `C` | Confirm anomaly |
| `R` | Reject anomaly |
| `Enter` | Open entity detail for full feedback form |

---

## What Happens After You Submit

RTSA evaluates your feedback using a **trust-scoring algorithm** before incorporating it into the AI model. This protects against accidental or malicious model manipulation.

### Trust Score Calculation

Your feedback is scored based on four factors:

| Factor | Weight | Description |
|---|---|---|
| **Security Clearance** | 20% | Higher clearance = higher base trust |
| **Historical Accuracy** | 30% | Your past feedback vs. ground truth outcomes |
| **Temporal Consistency** | 20% | Feedback submitted soon after observation is more reliable |
| **Statistical Deviation** | 30% | Feedback that aligns with AI consensus is weighted higher |

### Possible Outcomes

| Trust Score | Result | What You See |
|---|---|---|
| ≥ 0.5 (High Trust) | Feedback accepted immediately | "✅ Feedback accepted — thank you" |
| 0.2–0.49 (Medium Trust) | Feedback held for human review | "⏳ Feedback submitted for review" |
| < 0.2 (Low Trust) | Feedback rejected; Security Officer alerted | "❌ Feedback rejected — contact Security Officer" |

> **If your feedback is rejected**, this is not necessarily a reflection of your judgment. It may indicate unusual patterns that the system's anti-manipulation safeguards need to verify. Contact your Security Officer if rejections persist.

---

## Tips for High-Quality Feedback

**Submit feedback promptly.** Feedback submitted within 5 minutes of your observation carries the highest temporal trust score. Waiting over an hour reduces trust weighting.

**Be specific.** If you are rejecting an anomaly, add a comment explaining why. For example: *"This vessel is a tanker operating under a known emergency deviation from the shipping lane. Coordinated with port authority."*

**Don't reject alerts you are unsure about.** If you are unsure whether an anomaly is real, leave it open or assign it to another operator. Rejecting uncertain alerts introduces noise into the model.

**Reclassify conservatively.** When reclassifying from UNKNOWN to HOSTILE, be certain. This feedback has significant downstream effect.

---

## Understanding Feedback Acknowledgment Statuses

After submitting feedback, the alert card shows a small indicator:

| Indicator | Meaning |
|---|---|
| ✅ Accepted | Feedback accepted and queued for model improvement |
| ⏳ Under Review | Feedback held for Security Officer review |
| ❌ Rejected | Feedback rejected (unusual; check with Security Officer) |
| 📡 Queued (Edge) | Feedback stored locally at edge; will sync when connected |

---

## Feedback at Tactical Edge

When operating at the **tactical edge** in disconnected mode:

- You can still submit all feedback types
- Feedback is stored locally and applied to the local edge model immediately
- When connectivity to the data centre is restored, your feedback is automatically synced
- The data centre re-evaluates the feedback with full trust scoring upon sync

> Feedback submitted at the edge still counts — it just takes longer to reach the full model retraining pipeline.

---

## Bulk Feedback and Anti-Poisoning Safeguards

If the system detects an unusual pattern in your feedback — such as many rejections in a short time or feedback that strongly deviates from the AI consensus — it will automatically pause your feedback submissions for security review.

This anti-poisoning safeguard exists to prevent malicious or accidental degradation of the AI model. If your feedback is paused:

1. A notification appears in the UI: "⚠️ Feedback submission temporarily paused — pending security review"
2. The Security Officer is automatically notified
3. Normal feedback submission resumes once the Security Officer clears the review

This is a system protection measure, not disciplinary action. If this happens, contact your Security Officer promptly.

---

> **Next**: [Tactical Edge Operations →](04_tactical_edge.md)

---

> **CLASSIFICATION: UNCLASSIFIED** — This document contains no classified information.
