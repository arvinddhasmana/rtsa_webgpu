<!-- CLASSIFICATION: UNCLASSIFIED -->

# Audit Trail Review

> **CLASSIFICATION: UNCLASSIFIED**
> **Audience**: Security Officers, Security Authorities
> **Version**: 1.0
> **Last Updated**: 2026-02-26

---

## Overview

Every state-changing operation in RTSA — every user login, every feedback submission, every query, every NATO data export, every alert acknowledgment — produces an **immutable audit event** stored in the Redpanda audit topic. These events are append-only: they cannot be modified or deleted by any user, including system administrators.

The **Audit Trail Viewer** is your window into this record. Access it via: 🔒 **Audit** in the main toolbar.

---

## What Is Recorded in the Audit Trail?

| Event Category | Examples of Events Recorded |
|---|---|
| **Authentication** | Login, logout, session timeout, certificate validation |
| **Track Access** | Viewing an entity track above a threshold classification |
| **Alert Management** | Alert acknowledged, confirmed, rejected, assigned |
| **Operator Feedback** | Feedback submitted, trust score result, accepted/held/rejected |
| **Historical Queries** | Query executed, parameters, result count |
| **NATO Exchange** | Track exported, track blocked (not releasable), inbound track received |
| **Forensic Reports** | Report generated, analyst identity, queries included |
| **Feedback Review** | Security Officer cleared or escalated a feedback hold |
| **System Events** | Model updated, sensor degraded, dead-letter queue spike |
| **Classification Events** | Classification banner change, potential spillage alert |

---

## Accessing the Audit Trail Viewer

1. Click 🔒 **Audit** in the main toolbar
2. The Audit Trail Viewer opens in a full-width panel
3. By default, you see the **last 24 hours of events** sorted most-recent first

---

## Searching and Filtering Audit Events

### Filter Options

| Filter | Description |
|---|---|
| **Time Range** | Set start and end time (UTC) |
| **Event Category** | Authentication, Alert, Feedback, Query, NATO, etc. |
| **User / Operator** | Filter by operator identity (from their certificate) |
| **Entity / Track ID** | Show all audit events involving a specific entity |
| **Classification Level** | Show events involving data at or above a classification level |
| **Outcome** | Success, Rejected, Held, Blocked |

### Keyword Search

Type keywords in the search bar to search event descriptions. Examples:
- `TRK-4501` — all events involving this track
- `feedback_rejected` — all feedback rejections
- `nato_export_blocked` — all blocked NATO exports

---

## Reading an Audit Event

Each audit event has a structured format:

```
┌─────────────────────────────────────────────────────────────────┐
│ EVENT: feedback_submitted                                       │
│ Time:       2026-02-20 14:33:42 UTC                             │
│ User:       CN=Smith.J.A, O=CAF, OU=21 Ops                     │
│ Session:    sess-f0a234b8                                       │
│ Entity:     TRK-4501                                            │
│ Details:    feedback_type=CONFIRM_HOSTILE                       │
│             trust_score=0.74                                    │
│             outcome=accepted                                    │
│             routed_to=feedback.operator.validated               │
│ Immutable:  ✅ Signature verified                               │
└─────────────────────────────────────────────────────────────────┘
```

| Field | Meaning |
|---|---|
| **EVENT** | The event type identifier |
| **Time** | UTC timestamp of the event |
| **User** | The operator's certificate identity (CN = Common Name) |
| **Session** | The user's session identifier (links all actions in a session) |
| **Entity** | If applicable, the affected track or alert ID |
| **Details** | Event-specific key-value pairs |
| **Immutable** | Signature verification status — must be ✅ |

---

## Key Audit Events to Monitor Regularly

### Authentication Failures
```
EVENT: authentication_failed
Details: reason=invalid_certificate
```
Repeated failures may indicate an expired or compromised certificate, or an unauthorized access attempt.

---

### Classification Spillage Alerts
```
EVENT: classification_access_warning
Details: user_clearance=PROTECTED_B, data_level=PROTECTED_C
         action=access_denied
```
This event fires when a user attempts to access data above their clearance. Access is denied automatically. Investigate the session to determine whether the user received any indication of the classified data's existence.

---

### Feedback Rejected (Low Trust)
```
EVENT: feedback_rejected_low_trust
Details: trust_score=0.14, operator=Smith.J.A
         SecOps_alerted=true
```
A feedback rejection at trust < 0.2 is a significant security event. Review immediately. See [Feedback Integrity Review](03_feedback_review.md).

---

### NATO Export Blocked
```
EVENT: nato_export_blocked
Details: track_id=TRK-7302, reason=NOFORN_restriction
```
Normal and expected — tracks with NOFORN caveat are automatically blocked. If you see high volumes of unexpected blocks, investigate the classification marking pipeline.

---

### Historical Query — Unusual Parameters
```
EVENT: historical_query_executed
Details: user=Analyst.K.R, time_range=30d, result_count=98241
         entity_type=ALL, classification_filter=PROTECTED_C
```
Very large result sets or broad queries by users who do not normally run them deserve review.

---

## Audit Trail Integrity Verification

Each audit event is cryptographically signed. The **Immutable** field must always show ✅. If you ever see:

```
Immutable: ❌ Signature verification FAILED
```

**This is a critical security incident.** An audit event with a failed signature indicates possible tampering. Take the following steps immediately:

1. Note the event ID, timestamp, and event type
2. Do not attempt to modify or delete the event
3. Escalate immediately to the system administrator and your chain of command
4. Initiate an incident report per your unit's security procedures

---

## Audit Data Retention

| Data Type | Retention |
|---|---|
| Operational audit events | 2 years (data centre) |
| Classification events | 2 years |
| Authentication events | 2 years |
| Edge audit events (local) | 7 days (synced to DC when connected) |

After 2 years, events are archived off-system per GC data retention policy. Contact your system administrator for archival access.

---

> **Next**: [Classification Management →](02_classification_management.md)

---

> **CLASSIFICATION: UNCLASSIFIED** — This document contains no classified information.
