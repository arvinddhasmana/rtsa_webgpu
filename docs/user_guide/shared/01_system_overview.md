<!-- CLASSIFICATION: UNCLASSIFIED -->

# System Overview — What Is RTSA?

> **CLASSIFICATION: UNCLASSIFIED**
> **Audience**: All RTSA Users
> **Version**: 2.0
> **Last Updated**: 2026-02-28

---

## What Is RTSA?

The **Real-Time Situational Awareness & Risk Assessment (RTSA)** system is a command and control decision-support platform built for Canadian Armed Forces (CAF) operations. It brings together data from multiple sensor types, fuses them into a single coherent operational picture, and uses artificial intelligence to detect anomalous and potentially threatening behaviour — all in real time.

In plain terms: **RTSA shows you what is happening in your operational area, flags anything unusual, and helps you make better decisions faster.**

---

## The Problem RTSA Solves

Before RTSA, operators had to monitor multiple separate sensor feeds — radar terminals, maritime AIS displays, electronic warfare dashboards, cyber threat feeds — and mentally fuse that information themselves. This was slow, error-prone, and left gaps. Threats could hide in the gaps between systems.

RTSA eliminates those gaps by:

- **Ingesting** all sensor types simultaneously
- **Fusing** overlapping observations into a single entity track
- **Scoring** every entity for anomalous behaviour using AI
- **Presenting** a unified, real-time picture on one screen

---

## How Does Data Flow Through RTSA?

Understanding the data flow helps you interpret what you see in the interface.

```
┌─────────────────────────────────────────────────────────────────────┐
│  EXTERNAL SENSORS (Untrusted Zone)                                  │
│  Radar · EW/SIGINT · ELINT/COMINT · ISR · AIS/BFT · Cyber · NATO   │
└───────────────────────────┬─────────────────────────────────────────┘
                            │ Validated & normalized
                            ▼
┌─────────────────────────────────────────────────────────────────────┐
│  INGESTION LAYER                                                    │
│  Each sensor type has a dedicated ingestion service that validates  │
│  incoming data, rejects bad data, and normalizes to a common format │
└───────────────────────────┬─────────────────────────────────────────┘
                            │ Published to event stream
                            ▼
┌─────────────────────────────────────────────────────────────────────┐
│  EVENT STREAMING (Redpanda)                                         │
│  The central backbone. All sensor data flows through here. Every    │
│  event is immutable and can be replayed. Also stores audit events.  │
└──────────────┬────────────────────────────┬────────────────────────┘
               │                            │
               ▼                            ▼
┌──────────────────────┐      ┌─────────────────────────────────────┐
│  FUSION ENGINE       │      │  HISTORICAL STORAGE (ClickHouse)    │
│  Correlates sensor   │      │  All events stored for up to 90 days│
│  reports into single │      │  (sensor data) or 2 years (audits). │
│  entity tracks with  │      │  Analysts query this for forensics. │
│  position, type,     │      └─────────────────────────────────────┘
│  and confidence      │
└──────────┬───────────┘
           │ Fused tracks
           ▼
┌─────────────────────────────────────────────────────────────────────┐
│  ANOMALY DETECTION (AI Engine)                                      │
│  Scores every track for unusual behaviour using pre-trained ML      │
│  models. Produces anomaly scores from 0.0 (normal) to 1.0          │
│  (critical). Generates human-readable explanations.                 │
└──────────────────────────────┬──────────────────────────────────────┘
                               │ Alerts + tracks streamed
                               ▼
┌─────────────────────────────────────────────────────────────────────┐
│  PRESENTATION LAYER (What You See)                                  │
│  COP Web Application — live map, alert panel, entity detail panel,  │
│  historical query panel, feedback submission                         │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Key Concepts

### Entity Track

An **entity track** is the unified picture of a single real-world object (a vessel, an aircraft, a ground vehicle, a cyber threat actor). RTSA creates a track by correlating observations from multiple sensors. A track includes:

- **Position**: Latitude, longitude, altitude
- **Kinematics**: Speed and heading
- **Entity type**: Air, Surface, Subsurface, Land, Space, or Cyber
- **Hostile status**: Unknown, Pending, Friendly, Neutral, Hostile, or Suspect
- **Confidence score**: How certain the system is about this track (0.0–1.0)
- **Source attribution**: Which sensors contributed to this track

### Anomaly Score

The AI engine gives every track an **anomaly score** between 0.0 and 1.0:

| Score Range | Level | What It Means |
|---|---|---|
| 0.0–0.3 | **NORMAL** | No unusual behaviour detected |
| 0.3–0.6 | **WATCH** | Mild deviation from expected behaviour; monitor |
| 0.6–0.8 | **ELEVATED** | Clear behavioural anomaly; operator attention required |
| 0.8–1.0 | **CRITICAL** | Strong anomaly signal; immediate operator action required |

### Operator Feedback

Operators are a critical part of RTSA's accuracy. When you **confirm**, **reject**, or **reclassify** an AI detection, that feedback is trust-scored and can be used to improve the AI model over time. This keeps the system accurate and relevant to evolving threats.

### Classification Level

Every piece of data in RTSA carries a **classification marking**. The system enforces classification at every layer — you will only ever see data you are cleared to access. The UI always shows the current classification level of displayed content through coloured banners.

---

## Sensor Types at a Glance

| Sensor | What It Detects | Example |
|---|---|---|
| **Radar** | Air and surface targets by reflected radio waves | Aircraft at 200 km, vessel at 50 km |
| **EW / SIGINT** | Electronic emissions and signals intelligence | Radar lock-on, communications intercept |
| **ELINT / COMINT** | Emitter profiles and communications content metadata | Identifying a radar type by its emissions |
| **ISR** | Imagery and full-motion video metadata from ISR platforms | Drone sighting metadata, imagery collection reports |
| **AIS / BFT** | Maritime AIS transponders and Blue Force Tracking | Ship position reports, friendly unit positions |
| **Cyber** | Cyber threat indicators (IOCs, STIX/TAXII) | Malicious IP addresses, threat actor TTPs |

---

## Deployment Environments

RTSA operates in two distinct environments:

### Data Centre (Full Capability)
- Full sensor coverage at high throughput (up to 50,000 events/second)
- Complete AI model suite with continuous learning
- Full historical database (90-day sensor data, 2-year audit)
- NATO data exchange active
- Connectivity to all upstream systems

### Tactical Edge (Disconnected / Constrained)
- Subset of sensors depending on local deployment
- Pre-trained AI model (no live updates at edge)
- Limited local historical database (7 days)
- Feedback queued locally, synced when connectivity is restored
- Reduced UI update rate to conserve bandwidth
- Visual indicator shows: **"EDGE MODE — LOCAL DATA ONLY"**

---

## The Human-in-the-Loop Principle

RTSA is designed around one core philosophy: **AI assists, humans decide.**

The AI engine detects and scores anomalies automatically, but it never acts autonomously on critical decisions. Operators remain in control. Your feedback actively shapes the system's future accuracy. This design ensures that:

- False positives can be corrected
- The AI cannot be silently manipulated without audit records
- Operator judgment always takes precedence over automated scoring

---

## Stakeholders Using RTSA

| Role | Primary Use of RTSA | Default Dashboard (v2.0) |
|---|---|---|
| **Operations Commander** | Real-time COP, anomaly alert management, tactical decisions | Fusion Dashboard |
| **Watch Officer** | Continuous COP monitoring, alert acknowledgment, feedback | Operator UI |
| **Intelligence Analyst** | Historical investigation, pattern analysis, forensic reporting | Forensics Panel |
| **Security Officer** | Audit trail review, classification compliance, feedback integrity | Audit & Feedback |
| **Sensor Operator** | Sensor health monitoring, data quality management | Sensor Health Dashboard |
| **NATO Liaison Officer** | Outbound/inbound NATO data exchange management | NATO Exchange Dashboard |

> **v2.0 Change**: The COP application now enforces a Two-Level Role-Based shell. After selecting your role (Level 1), you choose from the dashboard views available to your role (Level 2). This provides a purpose-built interface optimized for each operator's workflow.

---

> **Next**: [Getting Started →](02_getting_started.md)

---

> **CLASSIFICATION: UNCLASSIFIED** — This document contains no classified information.
