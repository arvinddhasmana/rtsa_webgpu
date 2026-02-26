<!-- CLASSIFICATION: UNCLASSIFIED -->

# Data Quality & Dead-Letter Queue

> **CLASSIFICATION: UNCLASSIFIED**
> **Audience**: Sensor Operators
> **Version**: 1.0
> **Last Updated**: 2026-02-26

---

## Overview

When sensor data arrives at RTSA, the **ingestion service** validates every message before it enters the system. Messages that fail validation are sent to the **Dead-Letter Queue (DLQ)** — a holding area for invalid data that lets you diagnose what went wrong without losing the raw data.

Understanding the DLQ is essential for sensor health monitoring. A spike in DLQ events almost always indicates a problem worth investigating.

---

## Why Data Is Rejected

Every ingestion service validates the following fields before accepting data:

| Validation Check | What Is Verified |
|---|---|
| **Coordinates** | Latitude is between −90° and +90°; longitude between −180° and +180° |
| **Timestamp** | Message timestamp is not in the future; not older than 10 minutes |
| **Sensor ID** | Sensor ID is present and matches a known/authorized sensor |
| **Required fields** | All mandatory fields for the sensor type are present |
| **Data type / range** | Numeric fields are within physically plausible ranges (e.g., speed > 0) |
| **Protobuf schema** | Message conforms to the expected Protobuf schema version |

If any check fails, the message is rejected and sent to the DLQ with the reason for rejection.

---

## Accessing the Dead-Letter Queue

1. Open the **Sensor Health Panel** (📡 Sensors)
2. Click on any sensor row to open its detail view
3. Scroll to the **DLQ Rejection Reasons** section

Or access the full DLQ browser:
1. Click the **DLQ** tab at the top of the Sensor Health Panel
2. Filter by sensor type, time range, or rejection reason

---

## DLQ Event Layout

```
┌─────────────────────────────────────────────────────────────┐
│ DLQ EVENT                                                   │
│ Sensor Type:    AIS                                         │
│ Sensor ID:      AIS-002                                     │
│ Received at:    2026-02-20 14:15:32 UTC                     │
│ Rejection:      invalid_timestamp                           │
│ Details:        Message timestamp 2026-02-20 14:35:00 UTC   │
│                 is 19 minutes in the future                 │
│ Raw message ID: dlq-a9f23c41-b102                           │
│ Volume:         1 (isolated event)                          │
└─────────────────────────────────────────────────────────────┘
```

---

## Common Rejection Reasons and Their Causes

### `invalid_timestamp`

**What it means**: The timestamp on the message is in the future (clock skew) or too far in the past (stale data).

| Symptom | Likely Cause |
|---|---|
| Timestamp is future (< 5 minutes ahead) | Sensor clock is slightly out of sync with NTP |
| Timestamp is future (> 5 minutes ahead) | Significant NTP drift or sensor clock failure |
| Timestamp is > 10 minutes in the past | Sensor is buffering messages and flushing late; connectivity issue |

**Action**: Report to your system administrator with the sensor ID and timestamp difference. NTP sync issues are fixed at the sensor level.

---

### `missing_sensor_id`

**What it means**: The message does not include a sensor identifier, so the system cannot attribute the data.

| Likely Cause | Action |
|---|---|
| Sensor firmware update removed the ID field | Report to sensor technical team |
| New sensor added without configuration | System administrator needs to register the new sensor |

---

### `coordinates_out_of_range`

**What it means**: Latitude or longitude values are outside physically possible bounds, or are identically zero (0°, 0° — a common "null" GPS value).

| Symptom | Likely Cause |
|---|---|
| Coordinates are (0, 0) | GPS fix lost; sensor is sending null coordinates |
| Coordinates are > ±90° / ±180° | Data encoding error; contact sensor team |
| Coordinates jump thousands of km | GPS spoofing attempt or sensor malfunction |

**Action**: A sensor reporting (0, 0) has lost GPS lock. Report to sensor technical team. Do not accept these as valid position reports.

---

### `invalid_speed` / `invalid_heading`

**What it means**: Speed is negative or physically implausible (e.g., a surface vessel at 500 knots); heading is outside 0–360°.

**Likely Cause**: Sensor calculation error, unit conversion error, or data corruption.

**Action**: Report to sensor technical team with the affected sensor ID and example values.

---

### `schema_mismatch`

**What it means**: The message does not conform to the expected Protobuf schema version.

**Likely Cause**: Sensor firmware was updated to a newer message format that RTSA's ingestion service has not yet been updated to accept.

**Action**: Report to system administrator — this typically requires a coordinated update of both the sensor firmware and the RTSA ingestion service schema. Do not attempt to resolve independently.

---

## Identifying Patterns in DLQ Data

When DLQ events are numerous, look for patterns:

### Isolated Events (< 10 in an hour)
Normal occurrence — occasional malformed messages from any sensor. No action required unless the rejection type is `schema_mismatch` or `missing_sensor_id`.

### Burst Pattern (many events in a short window, then returns to normal)
Suggests a transient issue: brief sensor reboot, momentary GPS loss, network jitter. Document and monitor. If bursts recur on a schedule, investigate the sensor's maintenance cycle.

### Sustained Elevated Volume (ongoing elevated rejection rate)
Indicates a persistent problem requiring investigation. Escalate to system administrator. Notify Operations Commander that the affected sensor's data quality is degraded.

### All Same Rejection Type from Same Sensor
A configuration or firmware problem specific to that sensor. Report to the sensor technical team with the sensor ID, rejection type, and time window.

### Same Rejection Type Across Multiple Sensors
May indicate a systemic issue — an ingestion service configuration change, a schema version mismatch, or a network-level problem. Escalate to the system administrator.

---

## Escalation Reference

| Scenario | Escalate To |
|---|---|
| NTP / clock issues | System administrator → sensor technical team |
| GPS fix lost | Sensor technical team |
| Schema mismatch | System administrator |
| Sustained high DLQ volume (> 200/hr) | System administrator immediately |
| Sensor fully inactive (no data) | System administrator immediately |
| Suspected GPS spoofing (coordinated jump) | Security Officer + Operations Commander immediately |

---

> **Back to Role Overview**: [Sensor Operator Guide →](README.md)

---

> **CLASSIFICATION: UNCLASSIFIED** — This document contains no classified information.
