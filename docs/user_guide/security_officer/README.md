<!-- CLASSIFICATION: UNCLASSIFIED -->

# Security Officer — Role Guide

> **CLASSIFICATION: UNCLASSIFIED**
> **Audience**: Security Officers, Security Authorities
> **Version**: 1.0
> **Last Updated**: 2026-02-26

---

## Your Role in RTSA

As a **Security Officer (SO)**, your primary responsibilities in RTSA are:

- Maintaining **classification compliance** — ensuring no data is exposed above its authorized level
- Reviewing the **immutable audit trail** — verifying all user actions are recorded correctly
- Investigating **potential classification spillage** — responding rapidly to classification incidents
- Overseeing **operator feedback integrity** — reviewing flagged feedback that may indicate manipulation or error
- Managing **user access and clearance levels** — ensuring operators only see what they are cleared for

RTSA is built with security as the foundation. Classification enforcement, audit logging, and anti-manipulation safeguards are all automatic and immutable. Your role is to verify, investigate, and respond — not to manually enforce what the system already enforces automatically.

---

## Your Quick-Start Checklist

When starting a security review session:

- [ ] Access the **Audit Trail Viewer** (🔒 Audit in toolbar)
- [ ] Review **recent unacknowledged security events** in the audit panel
- [ ] Check for any **feedback integrity alerts** (flagged low-trust feedback)
- [ ] Verify the **classification banner** is displaying correctly for the current data level
- [ ] Review any **pending feedback review queue** items

---

## Your Guide Contents

| Document | What It Covers |
|---|---|
| [Audit Trail Review](01_audit_trail.md) | Accessing the immutable audit log; searching and interpreting events |
| [Classification Management](02_classification_management.md) | Verifying classification enforcement; responding to potential spillage |
| [Feedback Integrity Review](03_feedback_review.md) | Reviewing flagged operator feedback; anti-poisoning investigation |

---

## Key Business Use Cases You Cover

| Use Case | Guide Section |
|---|---|
| UC010 — Operator Feedback (integrity) | [Feedback Integrity Review](03_feedback_review.md) |
| UC011 — Model Retraining (oversight) | [Feedback Integrity Review](03_feedback_review.md) |
| All UCs (audit) | [Audit Trail Review](01_audit_trail.md) |

---

## Security Contact Escalation

| Situation | Action |
|---|---|
| **Possible classification spillage** | Immediately review the audit trail for the involved session; quarantine the workstation if needed |
| **Feedback integrity alert received** | Review the feedback queue within 30 minutes of alert |
| **User reports unexpected data visible** | Check their clearance profile; review audit trail for their session |
| **Anti-poisoning pause active for an operator** | Review their recent feedback in the feedback queue; clear or escalate |
| **Audit trail gap detected** | Escalate immediately to system administrator and chain of command |

---

> **CLASSIFICATION: UNCLASSIFIED** — This document contains no classified information.
