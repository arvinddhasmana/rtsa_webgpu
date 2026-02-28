<!-- CLASSIFICATION: UNCLASSIFIED -->

# NATO Data Exchange

> **CLASSIFICATION: UNCLASSIFIED**
> **Audience**: NATO Liaison Officers
> **Version**: 2.0
> **Last Updated**: 2026-02-28

---

## Overview

RTSA exchanges tactical data with NATO-allied systems using three standard formats:

| Format | Standard | Data Type |
|---|---|---|
| **Link 16 (STANAG 5516)** | J-Series messages | Tactical track positions, kinematics, identity |
| **NFFI XML** | NATO Friendly Force Information | Friendly force positions and status |
| **MIP** | Multilateral Interoperability Programme | Broader situational data and anomaly observations |

This guide explains how to monitor the exchange, interpret the status panel, and respond to link issues.

> **v2.0**: The NATO Data Exchange view is now a dedicated Level-2 dashboard accessible from the **NATO Liaison** role. Select the **NATO Liaison** role from the toolbar to see the NATO Exchange dashboard automatically.

---

## Accessing the NATO Exchange Panel

Click 🌐 **NATO** in the main toolbar to open the NATO Exchange Panel.

---

## NATO Exchange Status Panel

The status panel shows a real-time view of all NATO data exchange links:

```
┌────────────────────────────────────────────────────────────────────┐
│  NATO DATA EXCHANGE STATUS                      14:32:05 UTC       │
├──────────────────┬──────────────┬────────────────┬─────────────────┤
│  Channel         │  Status      │  Rate          │  Last Event     │
├──────────────────┼──────────────┼────────────────┼─────────────────┤
│  Link 16 (STANAG)│ ● Active     │  42 tracks/min │  14:31:58 UTC   │
│  NFFI/MIP Gateway│ ● Active     │  15 tracks/min │  14:31:45 UTC   │
│  Inbound (Link16)│ ● Receiving  │   8 tracks/min │  14:32:02 UTC   │
│  Inbound (NFFI)  │ ● Receiving  │   3 tracks/min │  14:31:59 UTC   │
├──────────────────┴──────────────┴────────────────┴─────────────────┤
│  Exercise Mode:  ● INACTIVE (Live Operations)                       │
│  Authorization:  Active since 2026-02-20 09:00 UTC                 │
│  Export rate:    57 tracks in last hour                             │
│  Blocked rate:   12 tracks blocked in last hour (NOFORN: 8, <0.6: 4)│
└────────────────────────────────────────────────────────────────────┘
```

---

## Link Status Indicators

| Indicator | Meaning |
|---|---|
| ● **Active** | Link is up and transmitting data |
| ● **Receiving** | Inbound link is active and receiving allied data |
| ⚠ **Degraded** | Link quality reduced; some messages may be lost |
| ✖ **Inactive** | Link is down; no data exchange |
| ⏸ **Paused** | Exchange manually paused by authorized personnel |

---

## Understanding the Export Log

Click **View Export Log** to see a chronological record of all outbound track exports.

```
┌────────────────────────────────────────────────────────────────────────┐
│ TIME (UTC)    TRACK ID    FORMAT    CLASSIFICATION  OUTCOME            │
├────────────────────────────────────────────────────────────────────────┤
│ 14:32:01      TRK-4201    Link 16   NATO RESTRICTED  Exported          │
│ 14:32:01      TRK-6030    NFFI      NATO RESTRICTED  Exported          │
│ 14:31:59      TRK-7302    Link 16   —               BLOCKED (NOFORN)   │
│ 14:31:58      TRK-9011    Link 16   —               BLOCKED (conf<0.6) │
│ 14:31:45      TRK-3140    NFFI      NATO RESTRICTED  Exported          │
└────────────────────────────────────────────────────────────────────────┘
```

### Export Outcomes

| Outcome | Description |
|---|---|
| **Exported** | Track was successfully transmitted to NATO systems |
| **BLOCKED (NOFORN)** | Track has NOFORN caveat — not releasable to any foreign system. This is correct and expected. |
| **BLOCKED (conf<0.6)** | Track confidence below 0.6 threshold — automatic quality control. |
| **BLOCKED (classification)** | Track classification exceeds NATO SECRET — not exportable. |
| **BLOCKED (NOFORN policy)** | CAN EYES ONLY or other restrictive caveat applied. |

A healthy export log will have a mix of **Exported** and **BLOCKED** events. BLOCKED events are normal and show the cross-domain guard is working correctly.

**Investigate if you see:**
- Zero BLOCKED events over a long period (guard may not be functioning)
- Unusual spikes in export volume (verify operational context with Operations Commander)
- Exports of track types you do not expect to be shared

---

## Understanding Inbound NATO Data

Tracks received from allied systems are ingested into RTSA and appear on the COP map with a **NATO** source indicator in the detail panel.

Key differences from locally sensed tracks:

| Attribute | Locally Sensed Track | NATO Inbound Track |
|---|---|---|
| Source | CAF sensor systems | Allied NATO system |
| Classification | Up to Protected C | Typically NATO RESTRICTED |
| Confidence | Calculated by RTSA fusion | Provided by allied system |
| Anomaly scoring | Fully processed | May be limited (no raw sensor history) |
| MMSI/callsign | Often available | May be anonymized |

NATO inbound tracks are displayed with a ⬟ NATO alliance marker on the map icon. You can filter for NATO-sourced tracks using the **Sensor Source: NATO** filter in the filter toolbar.

---

## Responding to Link Degradation

### Link 16 Degraded

When the Link 16 terminal reports degraded link quality:

1. RTSA automatically **reduces export rate** to CRITICAL tracks only (HOSTILE + ELEVATED anomaly)
2. Non-critical tracks are **buffered** for retry when link quality improves
3. You receive an alert in the NATO Exchange Panel: "⚠️ Link 16 — Degraded Quality"

**Your actions:**
1. Note the time degradation started
2. Notify the Operations Commander that NATO track sharing is reduced
3. Monitor the link status panel — degradation often resolves within minutes
4. If degradation persists > 15 minutes, escalate to your communications team

### NFFI / MIP Gateway Inactive

If the NFFI/MIP gateway goes offline:

1. RTSA suspends NFFI and MIP exports automatically
2. Buffered exports await reconnection
3. You receive an alert: "✖ NFFI/MIP Gateway — Inactive"

**Your actions:**
1. Notify Operations Commander
2. Escalate to system administrator — gateway recovery is a technical issue
3. Contact the receiving allied system's liaison if there is an operational urgency

---

## Exercise Mode

RTSA supports a dedicated **Exercise Mode** for NATO exercises to keep exercise data strictly separated from live operations.

### Activating Exercise Mode

Exercise Mode must be activated by an authorized user before an exercise begins:

1. In the NATO Exchange Panel, click ⚙️ **Configure**
2. Select **Exercise Mode: Activate**
3. Enter the exercise name and your authorization code
4. Set the exercise time window (start and end)
5. Click **Confirm**

**When Exercise Mode is active:**
- All exported data is tagged with an **Exercise Indicator**
- Exercise-specific track numbering ranges are used (preventing confusion with live tracks)
- A prominent banner appears: "🏋 EXERCISE MODE ACTIVE — [Exercise Name]"

### Deactivating Exercise Mode

At exercise conclusion, deactivate by reversing the steps above. Do not leave Exercise Mode active after an exercise ends.

> ⚠️ **Never share live operational data as exercise data or vice versa.** Misuse of exercise mode is a significant classification incident.

---

> **Next**: [Manual Track Nomination →](02_track_nomination.md)

---

> **CLASSIFICATION: UNCLASSIFIED** — This document contains no classified information.
