<!-- CLASSIFICATION: UNCLASSIFIED -->

# Sensor Health Monitoring

> **CLASSIFICATION: UNCLASSIFIED**
> **Audience**: Sensor Operators
> **Version**: 2.0
> **Last Updated**: 2026-02-28

---

## Overview

The **Sensor Health Dashboard** gives you a real-time view of the operational status of every sensor connected to RTSA. Access it by selecting the **Sensor Operator** role — the Sensor Health dashboard opens automatically as your default view.

> **v2.0**: The Sensor Health Dashboard is now a dedicated Level-2 dashboard view available exclusively to the Sensor Operator role. It displays per-sensor status cards with geographic coverage overlays on the map.

---

## The Sensor Health Dashboard *(v2.0)*

The dashboard shows each individual sensor and its current health status as a **status card**. Cards are colour-coded for immediate situational awareness. Below the cards, a geographic **Coverage Map** shows each sensor's effective detection footprint.

### Sensor Status Cards

Each card displays:

```
┌──────────────────────────────────────┐
│ 📡 RADAR-01                  🟢 CONNECTED │
│ Sensor ID: sensor-radar-northwest    │
│ Rate: 12,450 obs/s    DLQ: 23 (1h)  │
│ Last seen: 2s ago                    │
│ Quality: 99.8%                       │
└──────────────────────────────────────┘
```

**Card colour indicators:**
| Colour | Status | Condition |
|---|---|---|
| 🟢 Green | CONNECTED | Last observation received < 30 seconds ago |
| 🟡 Amber | STALE | Last observation received 30s – 2 minutes ago |
| 🔴 Red | OFFLINE | No data received > 2 minutes, or `connected = false` |

**Legacy summary view** (sensor type aggregate):

```
┌──────────────────────────────────────────────────────────────────────┐
│  SENSOR HEALTH OVERVIEW                          Last updated: 14:32  │
├────────────────┬────────────┬──────────────┬────────────┬────────────┤
│  Sensor Type   │  Status    │  Rate (msg/s)│  DLQ (1h)  │  Latency   │
├────────────────┼────────────┼──────────────┼────────────┼────────────┤
│  RADAR         │ ● Active   │  12,450      │  23        │  45ms      │
│  EW/SIGINT     │ ● Active   │   3,210      │   5        │  38ms      │
│  ELINT/COMINT  │ ⚠ Degraded │     450      │  142       │  520ms     │
│  ISR           │ ● Active   │     120      │   2        │  62ms      │
│  AIS           │ ● Active   │   8,340      │   8        │  30ms      │
│  BFT           │ ● Active   │   1,205      │   0        │  28ms      │
│  CYBER         │ ● Active   │     320      │  11        │  75ms      │
└────────────────┴────────────┴──────────────┴────────────┴────────────┘
```

### Coverage Map

Below the sensor cards, a geographic map shows each sensor's effective coverage area:

| Sensor Type | Map Rendering |
|---|---|
| Radar | Fan sector arc from sensor position (range + bearing sector) |
| EW/SIGINT | Circular range ring (omnidirectional) |
| ELINT/COMINT | Directional arc or range ring |
| ISR | Coverage polygon (swath footprint) |
| AIS | Reception range ring |

Coverage overlays are colour-coded to match sensor status (green/amber/red fill, 30% opacity). Offline sensors dim to 10% opacity.

---

## Status Indicators

| Indicator | Colour | Meaning |
|---|---|---|
| ● **Active** | Green | Sensor is reporting data at expected rate; ingestion healthy |
| ⚠ **Degraded** | Yellow | Sensor is reporting but at lower than expected rate, or with elevated errors |
| ✖ **Inactive** | Red | No data received from this sensor in the last 5 minutes |
| ⏸ **Paused** | Grey | Sensor ingestion manually paused by system administrator |
| ? **Unknown** | Blue | Ingestion service is running but sensor connectivity is uncertain |

---

## Metric Columns Explained

### Rate (msg/s)
Messages per second being received and successfully processed by the ingestion service.

**Expected ranges** (normal operations, data centre):

| Sensor | Expected Range |
|---|---|
| Radar | 5,000–20,000 msg/s |
| EW/SIGINT | 1,000–5,000 msg/s |
| ELINT/COMINT | 200–2,000 msg/s |
| ISR | 50–500 msg/s |
| AIS | 3,000–15,000 msg/s |
| BFT | 500–3,000 msg/s |
| Cyber | 100–1,000 msg/s |

> **Edge operations**: Expected rates are 10× lower. Consult your system administrator for edge-specific expected ranges for your deployment.

Rates **below the expected minimum** may indicate:
- Sensor connectivity issue
- Sensor transmitting less data than expected (physical issue)
- Network congestion between sensor and RTSA

Rates **above the expected maximum** may indicate:
- Sensor misconfiguration (sending duplicate data)
- Replay attack (malformed data flood)
- Normal surge during high-activity periods (validate with Operations Commander)

---

### DLQ (1h)
Number of messages rejected and sent to the **Dead-Letter Queue** in the last hour. Some DLQ events are normal (e.g., occasional malformed messages). Large numbers indicate a problem.

**Alert thresholds:**

| DLQ Count (1h) | Status | Action |
|---|---|---|
| 0–10 | Normal | Monitor |
| 11–50 | Watch | Investigate; check DLQ details |
| 51–200 | Elevated | Investigate promptly; escalate if not easily explained |
| > 200 | Critical | Escalate immediately to system administrator |

---

### Latency
End-to-end processing latency from when data is received by the ingestion service to when it is published to the event stream.

| Latency | Status |
|---|---|
| < 100ms | Normal |
| 100–500ms | Watch — possible processing bottleneck |
| > 500ms | Elevated — ingestion service may be overwhelmed; escalate |
| > 1,000ms | Critical — data significantly delayed; immediate escalation |

High latency affects the freshness of tracks on the COP map and can delay anomaly detection. Operators should be informed if latency exceeds 500ms.

---

## Sensor-Level Drill-Down

Click any sensor row to open its detailed view:

```
ELINT/COMINT — Detailed View
─────────────────────────────────────────────────────
Status:     ⚠ DEGRADED
Rate:       450 msg/s  (expected: 200–2,000)
DLQ (1h):   142 messages rejected
Latency:    520ms

Connected Sensors:
  ELINT-001  ● Active  (Rio area)
  ELINT-002  ✖ Inactive — No data since 14:05 UTC (27 min ago)
  ELINT-003  ● Active  (Pacific)

DLQ Rejection Reasons (last hour):
  invalid_timestamp:  98 events
  missing_sensor_id:  31 events
  coordinates_out_of_range:  13 events

Recent Events:
  14:05  ELINT-002 connection lost
  14:06  Degraded status set
  14:06  Degradation alert raised (Operations Commander notified)
```

This view helps you diagnose *which* sensor is causing the problem and *what* type of data is being rejected.

---

## Responding to Sensor Degradation

### Sensor Status: DEGRADED

1. **Click the sensor row** to see which sub-sensor is causing the degradation
2. **Check the DLQ** — what type of rejections are occurring? (See [Data Quality guide](02_data_quality.md))
3. **Notify the Operations Commander** — inform them that this sensor type is providing reduced coverage; tracks in this sensor's area may be less reliable
4. **Escalate to your system administrator or sensor technical team** if the cause is not apparent from the DLQ data

### Sensor Status: INACTIVE

An inactive sensor is more serious — no data at all is being received.

1. **Immediately notify the Operations Commander** — tracks in the affected area may now rely on fewer sensors; confidence scores will drop
2. **Escalate to system administrator** — an inactive sensor requires technical investigation beyond the UI
3. **Document the time of loss** in your duty log
4. **Monitor for restoration** — when the sensor comes back online, the status will change to ACTIVE

---

## Sensor Coverage Overlay on Map

To visualize which areas of the map are covered by which sensors:

1. Go to the main COP Map
2. Click the **Layers** button (bottom right)
3. Enable **Sensor Coverage**

Each sensor type is shown as a coloured overlay on the map. Areas where sensors are INACTIVE or DEGRADED will show reduced or missing coverage. This helps you advise Operations Commanders about areas of uncertainty.

---

> **Next**: [Data Quality & Dead-Letter Queue →](02_data_quality.md)

---

> **CLASSIFICATION: UNCLASSIFIED** — This document contains no classified information.
