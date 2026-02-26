<!-- CLASSIFICATION: UNCLASSIFIED -->

# Anomaly Pattern Analysis

> **CLASSIFICATION: UNCLASSIFIED**
> **Audience**: Intelligence Analysts
> **Version**: 1.0
> **Last Updated**: 2026-02-26

---

## Overview

**Pattern analysis** goes beyond individual incident investigation. It looks at anomaly data in aggregate — across time periods, geographic areas, and entity types — to identify trends, recurring threat behaviours, and systemic signals that no single alert would reveal. This is where intelligence is produced, not just incidents described.

---

## What Makes Pattern Analysis Valuable?

Individual anomaly alerts tell you **an entity is behaving unusually**. Pattern analysis tells you:

- **This type of behaviour is increasing** over the last 30 days
- **This area is seeing elevated anomaly rates** compared to baseline
- **These two anomaly types always occur together** — suggesting deliberate coordination
- **This entity class is disproportionately flagging** this anomaly type — possible sensor issue or real threat category

---

## Running a Pattern Analysis

### Step 1 — Define Your Analysis Window

Pattern analysis typically covers 7 to 30 days (the system's single-query maximum). For longer trends, run multiple queries and compare.

### Step 2 — Run the Anomaly Pattern Query

1. Open Historical Query panel (🔍 History)
2. Select **Anomaly Pattern Query**
3. Set your time range
4. Leave **Anomaly Types** blank for a comprehensive view, or select specific types
5. Set **Minimum Severity** to WATCH (include all levels for trends)
6. Group by **Anomaly Type** for the first run, then switch to **Time (hourly or daily)** for trend view
7. Click **▶ Run Query**

---

## Interpreting the Aggregate Results

### Frequency Table

The first view shows anomaly frequency by type:

```
Anomaly Type         | Count | Avg Confidence | Entities Affected | First Seen          | Last Seen
────────────────────────────────────────────────────────────────────────────────────────────────────
SPEED_ANOMALY        | 1,247 | 0.72           | 84                | 2026-02-01 04:12    | 2026-02-26 14:30
LOCATION_ANOMALY     |   832 | 0.68           | 63                | 2026-02-01 09:45    | 2026-02-26 12:18
BEHAVIORAL_ANOMALY   |   509 | 0.81           | 41                | 2026-02-03 17:22    | 2026-02-26 09:11
THREAT_INDICATOR     |   127 | 0.88           | 19                | 2026-02-14 22:05    | 2026-02-25 18:44
CORRELATION_ANOMALY  |    43 | 0.79           |  9                | 2026-02-18 11:30    | 2026-02-26 08:22
MOVEMENT_ANOMALY     |   291 | 0.65           | 55                | 2026-02-01 06:15    | 2026-02-26 13:55
```

**Key questions to ask:**
- Which anomaly type is most frequent? Is this expected?
- Which type has the highest average confidence? High-confidence alerts are most reliable.
- Are THREAT_INDICATOR counts low (normal) or spiking?
- Are CORRELATION_ANOMALY counts significant? These are high-value intelligence signals.

---

### Trend View (Time Series)

Switch to **Group by Time (Daily)** to see how anomaly rates evolve over time:

```
Date        | SPEED | LOCATION | BEHAVIORAL | THREAT | CORRELATION
────────────────────────────────────────────────────────────────────
2026-02-01  |  38   |   28     |    15      |   2    |    1
2026-02-05  |  42   |   31     |    18      |   4    |    2
2026-02-10  |  45   |   35     |    22      |   8    |    5
2026-02-15  |  58   |   44     |    31      |  18    |   12
2026-02-20  |  71   |   52     |    38      |  22    |   14
2026-02-25  |  80   |   58     |    42      |  25    |   15
```

In this example, **all anomaly types are trending upward** over 25 days — a potential indicator of increasing threat activity in the area, or a sensor quality issue. Investigate the cause.

---

## Identifying Significant Patterns

### Correlated Anomaly Types

Check for anomaly types that tend to **co-occur on the same entities** within a short time window. This suggests coordinated or intentional behaviour:

| Pattern | Possible Interpretation |
|---|---|
| LOCATION_ANOMALY + BEHAVIORAL_ANOMALY | Entity is out of normal area AND behaving strangely — deliberate action |
| SPEED_ANOMALY + BEHAVIORAL_ANOMALY | Entity using unusual speed AND has transponder anomalies — possible spoofing |
| THREAT_INDICATOR + CORRELATION_ANOMALY | Matches known threat signature AND has cyber-physical correlation — high-value target |
| MOVEMENT_ANOMALY + LOCATION_ANOMALY | Unusual route AND unusual area — possible surveillance run |

---

### Geographic Hotspot Analysis

Use the **Map View** on your pattern query results to overlay anomaly density on the map.

**Hotspot indicators:**
- **Dense clustering** of anomaly events in a small area — contested zone, chokepoint
- **Progressive movement** of anomaly clusters — possible entity moving through your area of operations
- **Recurring activity at the same coordinates** — surveillance pattern, regular meeting point

---

### Baseline vs. Anomaly Comparison

To understand whether current activity is elevated above normal:

1. Run a **Anomaly Pattern Query** for the most recent 7 days (current period)
2. Run the same query for a reference period (e.g., same 7 days last month)
3. Compare counts by anomaly type — any type more than 2× baseline deserves attention

---

## Pattern Analysis for Sensor Quality Assessment

Pattern analysis is also useful for **identifying sensor quality issues** rather than real threats:

| Pattern | Possible Sensor Issue |
|---|---|
| Sudden spike in SPEED_ANOMALY on a single sensor | Sensor calibration error; contact Sensor Operator |
| All anomalies involving one track disappear suddenly | Track lost / sensor coverage gap |
| Very high BEHAVIORAL_ANOMALY rate for AIS | AIS receiver interference or jamming |
| Low confidence scores across all anomalies | Model may be degraded; contact system administrator |

---

## Producing an Intelligence Product

After completing your pattern analysis:

1. Document your key findings in the **Investigation Notes** field (available in the History panel)
2. Save your query configurations for reproducibility (💾 Save Query)
3. Generate a **Forensic Report** including your pattern query results (see [Forensic Investigation guide](02_forensic_analysis.md))
4. Apply appropriate classification markings to your analysis product
5. Distribute through authorized channels per your unit's procedures

---

> **Back to Role Overview**: [Intelligence Analyst Guide →](README.md)

---

> **CLASSIFICATION: UNCLASSIFIED** — This document contains no classified information.
