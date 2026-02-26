<!-- CLASSIFICATION: UNCLASSIFIED -->

# Situational Awareness — The COP Map

> **CLASSIFICATION: UNCLASSIFIED**
> **Audience**: Operations Commanders, Watch Officers
> **Version**: 1.0
> **Last Updated**: 2026-02-26

---

## Overview

The **Common Operating Picture (COP) Map** is the heart of RTSA. It displays all entity tracks in your operational area in real time, updated as new sensor data arrives. This guide explains how to read the map, filter what you see, and inspect entities in detail.

---

## Reading the Map

### Entity Icons

Each entity on the map is shown as a **NATO APP-6 standard tactical symbol**. The icon shape and fill tell you the entity type and hostile status at a glance:

| Hostile Status | Colour | Border Shape | Example |
|---|---|---|---|
| HOSTILE | Red | Diamond | ◆ |
| SUSPECT | Orange | Diamond (dashed) | ◇ |
| UNKNOWN | Yellow | Square | ■ |
| FRIENDLY | Blue | Circle | ● |
| NEUTRAL | Green | Square | □ |
| PENDING | Blue (outline) | Circle (outline) | ○ |

The **icon symbol inside** the shape tells you the entity type:
- ✈ Air (aircraft)
- ⛵ Surface (vessel)
- 🚢 Subsurface (submarine)
- 🚗 Land (ground vehicle)
- 🛰 Space (satellite)
- 💻 Cyber (threat actor)

### Track Trails

A **fading trail** behind each icon shows the last several positions of the entity, giving you a visual sense of its movement direction and history:

- **Longer trail** = entity has been tracked for a while with consistent data
- **Short or no trail** = entity newly detected or track data interrupted
- **Dotted trail** = position extrapolated (sensor data gap; system is estimating position)

### Anomaly Halos

If an entity has an **ELEVATED or CRITICAL anomaly score**, a pulsing halo appears around its icon:

- 🟠 Orange halo = ELEVATED anomaly (score 0.6–0.8)
- 🔴 Red pulsing halo = CRITICAL anomaly (score 0.8–1.0)

The pulsing is intentional — it demands your attention. Click the entity to review the anomaly details.

### Track Confidence Indicator

The **size** and **opacity** of the track icon reflects confidence:

- **Large, solid icon** = high confidence (multiple sensors agree, 0.8–1.0)
- **Medium icon** = moderate confidence (0.5–0.8)
- **Small, faded icon** = low confidence (single sensor, unverified, < 0.5)

---

## Filtering What You See

The filter toolbar at the top of the screen lets you narrow down the display. This is essential when your operational area is busy.

### Recommended Filters by Scenario

**Focus on air threats only:**
- Entity Type: **Air**
- Hostile Status: **Hostile, Suspect, Unknown**

**Maritime situational picture:**
- Entity Type: **Surface, Subsurface**
- Sensor Source: **AIS, Radar**

**Cyber threat overlay:**
- Entity Type: **Cyber**
- Alert Severity: **Elevated, Critical**

**Blue Force picture only:**
- Hostile Status: **Friendly**
- Sensor Source: **BFT**

> **Tip**: Apply multiple filters together. All filters combine with AND logic — a track must match all active filters to appear on the map.

---

## Inspecting an Entity

Click any track on the map to open the **Entity Detail Panel** on the right side.

### What You See in the Detail Panel

**Identity Section**
```
Track ID:       TRK-4501
Entity Type:    SURFACE (vessel)
Hostile Status: SUSPECT
Confidence:     0.82 (HIGH)
Classification: PROTECTED B
Last Updated:   14:32:05 UTC (12 seconds ago)
```

**Position & Kinematics**
```
Position:    48.2142°N, 123.9568°W
Speed:       45 knots  ⚠️ (expected < 20 knots for type)
Heading:     087° (East)
Altitude:    N/A (surface)
```

**Source Attribution**
```
Sensor          Contribution    Confidence
RADAR-001       Primary         0.91
AIS-002         Secondary       0.74
```

**Anomaly History** (most recent first)
```
14:31:58  SPEED_ANOMALY   Score: 0.87  CRITICAL
14:28:12  LOCATION_ANOMALY Score: 0.64  ELEVATED
14:22:45  NORMAL           Score: 0.18  —
```

**Explanation from AI**:
> "Track TRK-4501 scored 0.87 CRITICAL: Speed of 45 knots exceeds expected maximum of 20 knots for a vessel of this type. Entity is 15 km outside the normal shipping lane for this area. AIS transponder has been toggled 3 times in the past hour."

---

## Track Timeline

The timeline at the bottom of the detail panel shows the lifecycle of the track:

```
12:45  Track created (single radar source)
13:02  AIS correlation added (confidence improved)
14:22  ELEVATED anomaly detected
14:31  CRITICAL anomaly detected
14:32  Now →
```

---

## Searching for a Specific Track

If you know the Track ID, MMSI, or callsign of an entity you want to find:

1. Press `Ctrl + F` to open the search box
2. Type the identifier
3. Press `Enter` — the map centres on the entity and opens the detail panel

---

## Understanding Stale Data

Any track that has not received an update in the last **2 minutes** is automatically marked **STALE**:

- The track icon becomes semi-transparent
- A "STALE" label appears next to the icon
- The detail panel shows the age of the last update

Stale tracks represent the **last known position** — the entity may have moved since. This is normal in sensor-limited environments. The system continues to **extrapolate** the track position based on last known speed and heading (shown as a dotted trail).

---

## Map Replay (Historical Playback)

You can replay the map at any previous point in time using the **Timeline Scrubber**:

1. Click the ⏱️ **Timeline** button in the toolbar
2. Set your **Start Time** and **End Time**
3. Press **▶ Play** — the map animates historical positions

> This feature queries the historical database. It may take a moment for the data to load, especially for long time ranges.

---

> **Next**: [Anomaly Alerts — Detection & Response →](02_anomaly_alerts.md)

---

> **CLASSIFICATION: UNCLASSIFIED** — This document contains no classified information.
