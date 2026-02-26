<!-- CLASSIFICATION: UNCLASSIFIED -->

# Forensic Investigation

> **CLASSIFICATION: UNCLASSIFIED**
> **Audience**: Intelligence Analysts, Forensic Investigators
> **Version**: 1.0
> **Last Updated**: 2026-02-26

---

## Overview

Forensic investigation in RTSA means conducting a **structured, evidence-based reconstruction** of an incident using historical data. This guide walks through a typical forensic investigation workflow, from defining the incident scope to generating a report.

---

## When Do You Conduct a Forensic Investigation?

- A reported incident needs after-action reconstruction
- An operator escalated a suspicious entity that requires deeper analysis
- A pattern of anomalies suggests a coordinated threat
- A scheduled post-operation intelligence review is due
- A security audit requires investigation of a specific event or time period

---

## Investigation Workflow

### Step 1 — Define the Investigation Scope

Before querying, define:

| Element | Decision |
|---|---|
| **Incident time window** | When did the event(s) occur? Give yourself ± 2 hours buffer. |
| **Geographic area of interest** | Where did the event(s) occur? Define a bounding box. |
| **Entities of interest** | Are there known track IDs? Or do you need to discover them? |
| **Anomaly types** | Are specific behaviours suspected? |
| **Classification ceiling** | What is the highest classification you will work with? |

Document your scope — it becomes part of the investigation record.

---

### Step 2 — Broad Time-Range Query

Start with a **broad time-range track query** to get the full picture of activity during your investigation window.

1. Open the Historical Query panel (🔍 History)
2. Select **Time-Range Track Query**
3. Set your time range (remember: max 30 days per query)
4. Set **Entity Type** and **Hostile Status** filters broadly at first
5. Click **▶ Run Query**
6. Switch to **Map View** to see the geographic distribution
7. Use the **Timeline Scrubber** to animate the period

**What to look for:**
- Entities that appear and disappear unusually
- Unexpected clustering of entities in specific areas
- Tracks with unusually short or fragmented histories (indicator of identity deception)

---

### Step 3 — Anomaly Pattern Review

Run an **Anomaly Pattern Query** for the same time window to see which anomaly types were most active.

1. Select **Anomaly Pattern Query**
2. Use the same time range
3. Group by **Anomaly Type** and **Time (hourly)**
4. Look for:
   - Spikes in anomaly frequency (sudden increase suggests a triggering event)
   - Correlated anomaly types (LOCATION_ANOMALY + BEHAVIORAL_ANOMALY together suggests deliberate behaviour)
   - Multiple entities showing the same anomaly type in the same area

---

### Step 4 — Entity Deep Dive

For each entity of interest, run an **Entity Historical Deep Dive** to get their complete record.

1. Select **Entity Historical Deep Dive**
2. Enter the Track ID
3. Set your time range

**Reviewing the entity timeline:**

```
Timeline for TRK-4501 · SURFACE · SUSPECT
────────────────────────────────────────────────────────
2026-02-20 12:45  Track created (RADAR-001, single source)
2026-02-20 13:02  AIS correlation added (MMSI: 123456789)
                  → Confidence improved: 0.54 → 0.82
2026-02-20 14:22  ELEVATED: LOCATION_ANOMALY (score: 0.64)
                  → 12 km from normal shipping lane
2026-02-20 14:28  ELEVATED: BEHAVIORAL_ANOMALY (score: 0.71)
                  → AIS transponder toggled OFF
2026-02-20 14:31  CRITICAL: SPEED_ANOMALY (score: 0.87)
                  → 45 kts (expected max: 20 kts)
                  → AIS transponder toggled ON
2026-02-20 14:33  Operator ACK: Watch Officer Smith (Confirmed hostile)
2026-02-20 14:35  Track lost (no sensor returns)
────────────────────────────────────────────────────────
```

**Key investigative questions to answer from the timeline:**
- When did anomalous behaviour begin relative to the overall incident?
- Were there precursor events (minor anomalies before escalation)?
- Did the entity interact with or correlate to other entities?
- Was any operator feedback submitted, and was it timely?
- Were there sensor gaps that could indicate deliberate evasion?

---

### Step 5 — Cross-Domain Correlation

If the incident may involve cyber elements alongside physical entities:

1. Select **Cross-Domain Correlation Query**
2. Use the same time range and geographic area
3. Check for: AIS/Cyber overlap, Radar/EW overlap

**Example finding**:
> "TRK-4501 (surface vessel, CRITICAL SPEED_ANOMALY at 14:31) is geolocated within 50 km of a geolocated cyber IOC (TAXII feed ID: CY-8821, detected 14:28). Both anomalies share the same 3-minute window. This correlation elevates the threat assessment."

---

### Step 6 — Review Operator Feedback

Check whether operators submitted feedback for the entities during the incident period:

1. Use the **Entity Deep Dive** and scroll to the **Operator Feedback History** section
2. Review: feedback type, timing, trust score, and whether it was accepted

**What to assess:**
- Was anomaly feedback submitted promptly? (Delay may indicate missed detection)
- Were there REJECT_ANOMALY submissions that now appear incorrect in retrospect?
- Did trust scoring flag any feedback for review?

---

### Step 7 — Generate a Forensic Report

When your analysis is complete:

1. Click 📄 **Generate Report** (top right of the Historical Query panel)
2. Fill in:
   - **Investigation title**
   - **Classification level of the report**
   - **Summary** (free text — your key findings)
   - **Queries to include** (checkboxes for each query you ran)
3. Click **Generate**

The system automatically:
- Executes the selected queries and includes results
- Applies classification markings to the report
- Stores the report in the audit trail
- Makes the report available for download (classification-appropriate channel)

> All forensic reports are logged in the audit trail — report title, generating analyst, generation time, queries included, and result counts.

---

## Evidence Chain of Custody

For investigations that may result in disciplinary or legal proceedings:

- Every query you ran is logged with full parameters and result counts
- The audit trail entry for each query is tamper-proof (stored in immutable Redpanda audit topics)
- Your analyst identity is captured from your authentication certificate
- Reports generated include a system-signed provenance block

If you need to formally certify the chain of custody, contact your Security Officer to initiate the formal evidence preservation process.

---

> **Next**: [Anomaly Pattern Analysis →](03_pattern_analysis.md)

---

> **CLASSIFICATION: UNCLASSIFIED** — This document contains no classified information.
