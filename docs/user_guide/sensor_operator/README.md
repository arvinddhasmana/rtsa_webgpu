<!-- CLASSIFICATION: UNCLASSIFIED -->

# Sensor Operator — Role Guide

> **CLASSIFICATION: UNCLASSIFIED**
> **Audience**: Sensor Operators
> **Version**: 1.0
> **Last Updated**: 2026-02-26

---

## Your Role in RTSA

As a **Sensor Operator**, you are responsible for the health and quality of the raw data flowing into RTSA. If the sensor data is bad, the fused tracks are bad, and the anomaly detection is unreliable. You are the first line of defence against data quality issues.

Your primary responsibilities are:

- Monitoring **sensor health and availability** — are all sensors reporting data?
- Assessing **data quality** — is the data valid, timely, and within expected parameters?
- Responding to **data rejections** — investigating sensor data that has been rejected by the ingestion service
- Escalating **sensor faults** to the appropriate technical teams

---

## Your Quick-Start Checklist

When starting a monitoring shift:

- [ ] Open the **Sensor Health Panel** (📡 Sensors in toolbar)
- [ ] Verify all expected sensors are **Active** (green status)
- [ ] Check the **throughput rates** — are they within expected ranges?
- [ ] Review the **Dead-Letter Queue** (DLQ) — any rejected data in the last hour?
- [ ] Check for any **sensor degradation alerts** in the alert panel

---

## Your Guide Contents

| Document | What It Covers |
|---|---|
| [Sensor Health Monitoring](01_sensor_health.md) | Understanding the sensor health dashboard; status indicators |
| [Data Quality & Dead-Letter Queue](02_data_quality.md) | Understanding rejected data; interpreting DLQ contents; escalating faults |

---

## Key Business Use Cases You Cover

| Use Case | Guide Section |
|---|---|
| UC002 — Radar Ingestion | [Sensor Health](01_sensor_health.md) |
| UC003 — EW/SIGINT Ingestion | [Sensor Health](01_sensor_health.md) |
| UC004 — ELINT/COMINT Ingestion | [Sensor Health](01_sensor_health.md) |
| UC005 — ISR Metadata Ingestion | [Sensor Health](01_sensor_health.md) |
| UC006 — AIS/BFT Ingestion | [Sensor Health](01_sensor_health.md) |
| UC007 — Cyber Threat Ingestion | [Sensor Health](01_sensor_health.md) |
| All Ingestion UCs (data quality) | [Data Quality & DLQ](02_data_quality.md) |

---

## Important Reminders

> 📡 **Sensor data quality directly affects threat detection.** A degraded sensor that produces bad data may cause false positives, false negatives, or track gaps. Respond to sensor faults promptly.

> ⚠️ **Do not tune or silence sensors yourself.** Configuration changes to sensor ingestion services must go through the system administrator with appropriate change management approval. Your role is to identify and escalate.

> 📋 **All sensor events are logged.** DLQ events, degradation alerts, and your acknowledgments are recorded in the audit trail.

---

> **CLASSIFICATION: UNCLASSIFIED** — This document contains no classified information.
