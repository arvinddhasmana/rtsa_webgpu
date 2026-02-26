<!-- CLASSIFICATION: UNCLASSIFIED -->

# Tactical Edge Operations

> **CLASSIFICATION: UNCLASSIFIED**
> **Audience**: Operations Commanders, Watch Officers
> **Version**: 1.0
> **Last Updated**: 2026-02-26

---

## Overview

RTSA is designed to operate **fully autonomously at the tactical edge** — in environments where there is no reliable connectivity to the data centre. This guide explains what changes when operating at the edge, what capabilities remain, and how to interpret the edge-specific indicators in the UI.

---

## When Does Edge Mode Activate?

Edge mode activates automatically when:

1. **RTSA is deployed on a tactical edge node** (K3s single-node deployment, vehicle-mounted or field-deployed hardware)
2. **Connectivity to the data centre is lost** on any RTSA deployment (both data centre and edge deployments can enter disconnected mode)

You will see a clear visual indicator when edge mode is active.

---

## Edge Mode Indicators

### Primary Indicator (Status Bar)
```
[ EDGE MODE — LOCAL DATA ONLY ]  Connection ● (grey)  UTC: 14:32:05
```

### Disconnected Mode (Data Centre RTSA, link lost)
```
[ DISCONNECTED — LOCAL DATA ONLY ]  Last sync: 14:28:52 UTC (3 min ago)
```

These indicators are displayed prominently and cannot be dismissed.

---

## What Works in Edge / Disconnected Mode

| Capability | Edge Mode | Disconnected (DC RTSA) |
|---|---|---|
| Live map with entity tracks | ✅ (local sensor data only) | ✅ (last known state, STALE) |
| Anomaly detection | ✅ (pre-trained edge model) | ✅ (last received model) |
| Alert generation | ✅ | ✅ (from local model) |
| Operator feedback | ✅ (queued locally) | ✅ (queued locally) |
| Historical query (local) | ✅ (7 days retention) | ✅ (full retention) |
| Historical query (full) | ❌ (no DC access) | ❌ (connection lost) |
| NATO data exchange | ❌ (no DC relay) | ❌ (connection lost) |
| Model updates | ❌ (receive on reconnect) | ❌ (receive on reconnect) |

---

## How the Display Changes at Edge

### Track Update Rate
- **Normal**: 10 updates per second per track
- **Edge / disconnected**: 1 update per second (bandwidth conservation)

Tracks will appear to move slightly less smoothly. This is normal.

### Track Trail Length
- **Normal**: Last 20 positions shown as trail
- **Edge**: Last 5 positions shown (reduced to save resources)

### Alert Filtering in Bandwidth-Constrained Mode
- Only **ELEVATED and CRITICAL** alerts are streamed to the UI
- WATCH-level alerts are still detected and logged but not shown in the panel until bandwidth improves

### Map Tiles
- Map tiles are served from the local cache
- If you pan to an area outside the cached tiles, the map shows a grey background
- Only the area surrounding your deployed location is guaranteed to be cached

### "STALE" Track Indicators
In disconnected mode on a data centre RTSA:
- Tracks not updated in the last **2 minutes** show as **STALE** with reduced opacity
- The track's last known position and a dotted extrapolated trail are shown
- A timestamp on the detail panel tells you exactly how old the data is

---

## The Edge AI Model

At the tactical edge, RTSA uses a **pre-trained AI model** shipped as part of the deployment bundle. This model:

- **Is not updated in real time** — it reflects training data as of the last deployment sync
- **Is still accurate** for common anomaly types covered in the training set
- **May be less sensitive** to very recent threat TTPs that emerged after the last model update

Model updates are applied automatically when the edge node reconnects to the data centre. The model version in use is shown at the bottom of the anomaly alert cards.

---

## Feedback at the Edge

You can submit all types of operator feedback while operating at the edge:

1. Feedback is stored **locally** in a secure queue
2. The local edge model incorporates your feedback **immediately** (for local adaptation)
3. When connectivity to the data centre is restored, feedback is **automatically synced**
4. The data centre applies full trust scoring to edge feedback upon sync

**Your edge feedback counts** — it will eventually contribute to model improvement once synced.

### Monitoring the Feedback Queue

To see how much feedback is queued for sync:
- Go to ⚙️ **Settings → Edge Sync Status**
- The panel shows: feedback items queued, last sync time, and estimated sync queue size

---

## Reconnection and Sync

When connectivity to the data centre is restored:

1. The UI displays: **"✅ Connection restored — syncing data..."**
2. Historical track data from the data centre loads automatically
3. The AI model updates if a newer version is available
4. Queued feedback is sent to the data centre for evaluation
5. The display switches from STALE to live data

The sync typically takes **30–120 seconds** depending on how long you were disconnected and how much data needs to be exchanged.

---

## Operational Recommendations for Edge Environments

**Set your alert filters before deploying.** In bandwidth-constrained environments, configure your alert panel to show ELEVATED and CRITICAL only before going disconnected — this minimizes display updates.

**Check your cached map coverage.** Before moving into a new area, verify that map tiles for that area are cached. Ask your system administrator to pre-cache tiles for your expected area of operations.

**Submit feedback even at the edge.** Every feedback submission improves future detection. The queue holds up to 10,000 feedback items.

**Note the model version on alerts.** If operating with an older model version (visible on alert cards), be more cautious about NORMAL scores — the model may not detect very recent threat patterns.

**Use the 7-day local history.** Even at the edge, you can query the last 7 days of local data for pattern analysis. Use the Historical Query panel for incident reconstruction.

---

> **Back to Role Overview**: [Operations Commander Guide →](README.md)

---

> **CLASSIFICATION: UNCLASSIFIED** — This document contains no classified information.
