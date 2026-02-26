<!-- CLASSIFICATION: UNCLASSIFIED -->

# Historical Queries

> **CLASSIFICATION: UNCLASSIFIED**
> **Audience**: Intelligence Analysts, Forensic Investigators
> **Version**: 1.0
> **Last Updated**: 2026-02-26

---

## Overview

The **Historical Query** panel provides access to the RTSA analytical database (ClickHouse), allowing you to retrieve historical entity tracks, anomaly detections, sensor events, and operator feedback for investigation and analysis. Queries are always classification-filtered based on your clearance level.

Access the panel via: 🔍 **History** button in the main toolbar.

---

## Opening the Historical Query Panel

1. Click the 🔍 **History** icon in the main toolbar
2. The panel opens as a side panel (or full-screen — toggle with the ⛶ expand button)
3. Select your **Query Type** from the dropdown (described below)
4. Fill in the parameters
5. Click **▶ Run Query**

Results appear in a table below the query form. Use the map toggle (🗺️) to visualize results on the historical map overlay.

---

## Query Types

### 1. Time-Range Track Query

Retrieve all entity tracks within a specified time period.

**Parameters:**

| Parameter | Description | Example |
|---|---|---|
| **Start Time** | Beginning of time range (UTC) | `2026-02-20 08:00:00` |
| **End Time** | End of time range (UTC) | `2026-02-20 16:00:00` |
| **Entity Type** | Filter by type (optional) | Air, Surface, Cyber |
| **Hostile Status** | Filter by status (optional) | Hostile, Unknown |
| **Result Limit** | Max rows to return (default: 1,000) | 5,000 |

**What You Get:**
- Track ID, entity type, hostile status, position, speed, heading
- Confidence score and number of contributing sensors
- Timestamp of each position update

**Use Case**: Getting the full operational picture for a specific shift or incident window.

---

### 2. Spatial Query (Geographic Bounds)

Retrieve all tracks that appeared within a defined geographic area during a time period.

**Parameters:**

| Parameter | Description | Example |
|---|---|---|
| **Start / End Time** | Time range | As above |
| **Bounding Box** | Draw on map, or enter lat/lon corners | 48.0°N–49.0°N, 123.0°W–124.5°W |
| **Entity Type** | Filter (optional) | |

> **Tip**: Use the **Draw on Map** tool to define the bounding box visually — click and drag a rectangle on the map panel.

**What You Get:**
- All tracks that appeared within the bounding box during the time range
- Full position data, entity type, and classification

**Use Case**: Investigating activity in a specific contested area or chokepoint.

---

### 3. Anomaly Pattern Query

Retrieve anomaly detection events grouped by type for statistical analysis.

**Parameters:**

| Parameter | Description |
|---|---|
| **Start / End Time** | Time range for the analysis |
| **Anomaly Types** | Select one or more types, or leave blank for all |
| **Minimum Severity** | WATCH / ELEVATED / CRITICAL |
| **Group By** | Anomaly type, time period (hour/day), entity type |

**What You Get:**
- Count of occurrences per anomaly type
- Average confidence score per type
- Number of distinct entities affected
- First and last detection timestamp per type

**Use Case**: Weekly anomaly trend review; identifying which anomaly types are most prevalent.

---

### 4. Entity Historical Deep Dive

Get the complete history of a single entity: every position update, every anomaly score, and every operator feedback submission.

**Parameters:**

| Parameter | Description |
|---|---|
| **Track ID** | The entity's track identifier (e.g., `TRK-4501`) |
| **Start / End Time** | Time range to investigate |

**What You Get:**
- Complete position history (every update)
- Anomaly scores at each time step
- Contributing sensors at each time step
- All operator feedback for this entity

**Use Case**: Full lifecycle investigation of a specific entity of interest.

---

### 5. Cross-Domain Correlation Query

Look for entities that triggered anomalies across multiple sensor domains simultaneously (e.g., a vessel showing both a location anomaly and a geolocated cyber IOC).

**Parameters:**

| Parameter | Description |
|---|---|
| **Start / End Time** | Time range |
| **Correlation Types** | Select domain pairs (e.g., AIS + Cyber, Radar + EW) |
| **Minimum Score** | Minimum anomaly score threshold |

**What You Get:**
- Track IDs showing correlated anomalies across domains
- Timing of each anomaly detection
- Geographic proximity analysis

**Use Case**: Investigating multi-domain threat actors where physical and cyber activity overlaps.

---

## Result Visualization

Query results can be visualized in three ways:

### Table View (Default)
Raw data in a sortable, filterable table. Export to CSV is available (subject to classification policy and your security officer's approval).

### Map View
Click 🗺️ **Show on Map** to overlay results on the historical map. Use the **Timeline Scrubber** to animate movement over time.

### Timeline View
A chronological chart showing track activity, anomaly events, and feedback submissions on a common timeline. Useful for seeing cause-and-effect relationships.

---

## Query Guardrails

The system enforces the following limits to protect performance:

| Guardrail | Limit | What To Do If You Hit It |
|---|---|---|
| Maximum time range | 30 days per query | Run multiple queries across shorter windows |
| Maximum result rows | 100,000 rows | Add more filters to narrow results |
| Query timeout | 30 seconds | Narrow your filters; complex queries time out |
| Concurrent queries | 5 per user | Wait for a query to complete before starting another |
| Classification filter | Always enforced server-side | Contact Security Officer if you believe data is missing |

---

## Saving and Reusing Queries

You can save query configurations for reuse:

1. Fill in your query parameters
2. Click 💾 **Save Query**
3. Give the query a name (e.g., "Straits weekly anomaly review")
4. Saved queries appear in the **Saved Queries** dropdown at the top of the panel

Saved queries are personal to your account. Share them with colleagues by exporting the query configuration (⚙️ → Export Query Config).

---

> **Next**: [Forensic Investigation →](02_forensic_analysis.md)

---

> **CLASSIFICATION: UNCLASSIFIED** — This document contains no classified information.
