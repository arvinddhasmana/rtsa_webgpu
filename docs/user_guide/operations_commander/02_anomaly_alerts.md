<!-- CLASSIFICATION: UNCLASSIFIED -->

# Anomaly Alerts — Detection & Response

> **CLASSIFICATION: UNCLASSIFIED**
> **Audience**: Operations Commanders, Watch Officers
> **Version**: 1.0
> **Last Updated**: 2026-02-26

---

## Overview

RTSA's AI engine continuously analyses every tracked entity for signs of unusual behaviour. When it detects an anomaly, it generates an **alert** that appears in your alert panel. This guide explains how to read, triage, and respond to those alerts.

---

## How the AI Detects Anomalies

The anomaly detection engine analyses every fused entity track and scores it against several behavioural models:

| Anomaly Type | What It Detects | Example |
|---|---|---|
| **MOVEMENT_ANOMALY** | Unusual movement pattern | Vessel circling repeatedly in a shipping lane |
| **LOCATION_ANOMALY** | Entity in unexpected location | Military-type aircraft in civilian airspace corridor |
| **SPEED_ANOMALY** | Speed outside normal range for entity type | Surface vessel exceeding 40 knots |
| **BEHAVIORAL_ANOMALY** | Deviation from normal behavioural pattern | AIS transponder cycling on and off repeatedly |
| **THREAT_INDICATOR** | Matches a known threat signature | Radar emissions matching a hostile platform type |
| **CORRELATION_ANOMALY** | Cross-domain correlation between unrelated systems | Cyber IOC geolocated near a vessel showing location anomaly |

The AI scores each entity from **0.0** (no anomaly) to **1.0** (maximum anomaly), and assigns a **confidence level** to that score.

---

## The Alert Panel

The alert panel on the left side of the screen displays all active anomaly alerts, sorted from highest to lowest severity.

### Alert Card Layout

```
┌─────────────────────────────────────────┐
│ 🔴 CRITICAL                             │
│ TRK-4501 · SURFACE vessel               │
│ SPEED_ANOMALY + LOCATION_ANOMALY        │
│                                         │
│ Score: 0.87  Confidence: 91%            │
│ Detected: 14:32:05 UTC (2 min ago)      │
│ Sources: RADAR-001, AIS-002             │
│                                         │
│ "Speed 45 kts, 15 km outside shipping   │
│  lane. AIS toggled 3x in last hour."    │
│                                         │
│ [Inspect] [Confirm] [Reject] [Assign]   │
└─────────────────────────────────────────┘
```

| Field | Meaning |
|---|---|
| **Severity badge** | CRITICAL / ELEVATED / WATCH — colour-coded |
| **Track ID & type** | Which entity triggered the alert |
| **Anomaly types** | What kind of anomaly was detected (can be multiple) |
| **Score** | The AI's anomaly score (0.0–1.0) |
| **Confidence** | How certain the AI is about this score |
| **Detected** | When the anomaly was first detected |
| **Sources** | Which sensors contributed to the fused track |
| **Explanation** | Plain-language description of why the anomaly was flagged |

---

## Alert Severity Levels

| Severity | Score | Indicator | Required Action |
|---|---|---|---|
| **WATCH** | 0.3–0.6 | 🟡 Yellow | Monitor; no immediate action required |
| **ELEVATED** | 0.6–0.8 | 🟠 Orange + halo | Review and assess; consider response |
| **CRITICAL** | 0.8–1.0 | 🔴 Red + pulsing halo + audio cue | Acknowledge and respond immediately |

> **CRITICAL alerts require acknowledgment.** The system tracks unacknowledged CRITICAL alerts. Your Watch Officer may escalate if they are not addressed in a timely manner.

---

## Responding to an Alert

### Step 1: Triage the Alert

Read the alert card. Note:
- **What type of anomaly** is it?
- **How high is the confidence?** High confidence (> 80%) means the AI is very certain.
- **Which sensors detected it?** Multiple sensors agreeing increases reliability.
- **When was it detected?** Is this current or has it been active for a while?

### Step 2: Inspect the Entity

Click **[Inspect]** to open the entity detail panel and see:
- Full track history and trajectory
- All contributing sensors and their individual confidence levels
- Complete anomaly history for this entity
- Any previous operator feedback

Use the map to visually assess the entity's position and movement.

### Step 3: Make a Judgment

Based on your assessment, you have several options:

**If the anomaly appears valid and concerning:**
- Click **[Confirm]** to confirm the anomaly (submits your professional assessment)
- Escalate as appropriate through your chain of command
- Use the chat / task assignment features to coordinate response

**If the anomaly appears to be a false positive:**
- Click **[Reject]** to reject the anomaly, with an optional reason
- Your feedback helps the AI learn to avoid this type of false alarm in future

**If you are unsure:**
- Click **[Assign]** to hand the alert to another operator for assessment
- Or leave it as WATCH/ELEVATED for continued monitoring
- Do not reject an alert you are unsure about — leave it for further observation

### Step 4: Acknowledge (for CRITICAL only)

CRITICAL alerts require explicit acknowledgment. After reviewing, click **[Acknowledge]** in the detail panel. This records that you have reviewed the alert and taken appropriate action. The alert moves to "Acknowledged" state but remains visible until the anomaly score drops below threshold.

---

## Alert Filtering

Use the alert panel filter (top of the alert panel) to focus on specific alert types:

| Filter | Use Case |
|---|---|
| **By Severity** | Show only CRITICAL to reduce noise on busy watch |
| **By Anomaly Type** | Focus on a specific threat category (e.g., SPEED_ANOMALY only) |
| **By Entity Type** | Filter to air threats only, or maritime only |
| **By Time** | Show only alerts in the last 30 minutes |
| **Unacknowledged Only** | See what still needs your attention |

---

## Audio Alerts

CRITICAL anomaly alerts are accompanied by an **audio cue** — a short alert tone. You can configure audio alert behaviour in:
⚙️ **Settings → Notifications → Audio Alerts**

Options include:
- **All severities** (Watch, Elevated, Critical)
- **Elevated and Critical only** (recommended for busy watches)
- **Critical only** (high-activity environments)
- **Muted** (not recommended; requires supervisor override)

---

## Multiple Simultaneous Alerts

During high-activity periods, you may have many alerts in the queue. Recommended triage approach:

1. **Start with CRITICAL** — address highest-score alerts first
2. **Check for correlated alerts** — multiple alerts on nearby tracks or overlapping areas may indicate a coordinated event
3. **Use the ASSIGN function** — distribute alerts across your watch team
4. **Filter by type** — if you see many alerts of the same type (e.g., all SPEED_ANOMALY), this may indicate a data quality issue or sensor calibration problem; report to your Sensor Operator

---

## Alert History

Resolved or acknowledged alerts are moved to the **Alert History** tab in the alert panel. You can review past alerts by:
- Time range
- Anomaly type
- Entity
- Operator who acknowledged

Alert history is retained for the life of the historical database (up to 90 days for operational alerts).

---

> **Next**: [Operator Feedback →](03_operator_feedback.md)

---

> **CLASSIFICATION: UNCLASSIFIED** — This document contains no classified information.
