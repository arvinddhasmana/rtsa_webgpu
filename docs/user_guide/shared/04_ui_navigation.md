<!-- CLASSIFICATION: UNCLASSIFIED -->

# UI Navigation — Common Interface Patterns

> **CLASSIFICATION: UNCLASSIFIED**
> **Audience**: All RTSA Users
> **Version**: 1.0
> **Last Updated**: 2026-02-26

---

## Overview

This section covers the common interface patterns, controls, and keyboard shortcuts that apply to all users of the RTSA COP dashboard. Role-specific features are described in your role's guide.

---

## The Main Toolbar

The toolbar at the top of the screen provides quick access to the primary functions:

| Button / Control | Function |
|---|---|
| 🗺️ **Map** | Switch to / focus the main COP map view |
| 🚨 **Alerts** | Open / expand the alert panel |
| 🔍 **History** | Open the historical query panel (analysts) |
| 📡 **Sensors** | Open the sensor health panel (sensor operators) |
| 🌐 **NATO** | Open the NATO exchange status panel (liaison) |
| 🔒 **Audit** | Open the audit trail viewer (security officers) |
| ⚙️ **Settings** | Personal display preferences |
| 👤 **Profile** | Your identity, clearance level, and session info |

> Not all buttons are visible for all roles. The interface adapts based on your role and clearance level.

---

## Map Controls

### Navigating the Map

| Action | Mouse | Keyboard |
|---|---|---|
| Pan | Click and drag | Arrow keys |
| Zoom in | Scroll wheel up | `+` or `=` |
| Zoom out | Scroll wheel down | `-` |
| Reset view | Double-click (background) | `Home` |
| Full screen | — | `F` |

### Selecting Entities

| Action | How |
|---|---|
| View entity details | Single-click on track icon |
| Centre map on entity | Double-click on track icon |
| Dismiss detail panel | Press `Escape` or click elsewhere |
| Select multiple entities | `Shift` + click each entity |

### Map Layer Controls

Use the **Layers** button (bottom-right of map) to toggle:

- **Track labels**: Show / hide track ID labels on the map
- **Track trails**: Show the last N position history as a fading trail
- **Sensor coverage**: Show effective sensor coverage overlays
- **Geo-fences**: Show area-of-interest boundaries
- **Grid overlay**: Show military grid reference (MGRS)

---

## Filter Toolbar

The filter toolbar appears below the main toolbar. Use it to narrow down what you see on the map and in the alert panel.

### Filter Options

| Filter | Options |
|---|---|
| **Entity Type** | All / Air / Surface / Subsurface / Land / Space / Cyber |
| **Hostile Status** | All / Hostile / Suspect / Unknown / Friendly / Neutral / Pending |
| **Sensor Source** | All / Radar / EW / ELINT / ISR / AIS / BFT / Cyber |
| **Alert Severity** | All / Critical / Elevated / Watch |
| **Time Window** | Last 5 min / Last 15 min / Last 1 hour / Custom |

> Filters apply to both the map and the alert panel simultaneously.

---

## Alert Panel

The **alert panel** on the left side shows active anomaly alerts sorted by severity.

### Alert Card Layout

```
┌─────────────────────────────────┐
│ 🔴 CRITICAL                     │
│ Track: TRK-4501                 │
│ SPEED_ANOMALY · AIS/RADAR       │
│ Score: 0.87 | Confidence: 91%   │
│ Detected: 14:32:05 UTC          │
│ ─────────────────────────────── │
│ [Inspect] [Confirm] [Reject]    │
└─────────────────────────────────┘
```

### Alert Quick Actions

| Button | Action |
|---|---|
| **Inspect** | Open entity detail panel for this track |
| **Confirm** | Confirm the anomaly is valid (submits operator feedback) |
| **Reject** | Reject the anomaly as a false positive (submits operator feedback) |
| **Assign** | Assign this alert to another operator for follow-up |

---

## Entity Detail Panel

Clicking any track on the map or any alert card opens the **entity detail panel** on the right side.

### Detail Panel Sections

**Identity**
- Track ID, entity type, hostile status, classification, confidence score

**Position & Kinematics**
- Latitude / longitude, speed (knots), heading (degrees), altitude (if applicable)
- Last update timestamp and "STALE" indicator if data is old

**Source Attribution**
- List of sensors contributing to this track
- Individual confidence score per sensor

**Anomaly History**
- Chronological list of anomaly events for this entity
- Each entry shows anomaly type, score, and explanation

**Operator Feedback History**
- Past feedback submitted by operators for this entity
- Shows feedback type and trust score (anonymized by default)

**Track Timeline**
- Key lifecycle events: track created, updated, merged, split

---

## Keyboard Shortcuts Reference

| Shortcut | Action |
|---|---|
| `M` | Focus map view |
| `A` | Focus alert panel |
| `H` | Open historical query panel |
| `Escape` | Close open panel / dismiss detail |
| `F` | Toggle full-screen map |
| `+` / `-` | Zoom in / out |
| `Home` | Reset map to default view |
| `Tab` | Cycle focus between panels |
| `Enter` | Confirm selected action |
| `Ctrl + F` | Open entity search |
| `Ctrl + Z` | Undo last filter change |

> RTSA is designed for **full keyboard-only operation** to support tactical environments where mouse use is impractical.

---

## Display Modes

### Normal Mode
Default display for data centre operations. Full colour, 10 Hz track update rate.

### NVG-Compatible Dark Mode
For use with night-vision goggles or in low-light tactical environments.
- Activate via: ⚙️ **Settings → Display → NVG Dark Mode**
- Reduces screen brightness and adjusts colours for NVG compatibility

### High-Contrast Mode
For high-ambient-light environments or accessibility needs.
- Activate via: ⚙️ **Settings → Display → High Contrast**

### Reduced Bandwidth Mode
Automatically activated in edge / tactical environments. Reduces update frequency and simplifies map rendering to conserve bandwidth.

---

## Search and Entity Lookup

Press `Ctrl + F` to open the entity search box. You can search by:

- **Track ID** (e.g., `TRK-4501`)
- **MMSI** (maritime identifier)
- **Callsign**
- **Geographic area** (type a place name or MGRS grid)

---

> **Now go to your role-specific guide:**
> - [Operations Commander →](../operations_commander/README.md)
> - [Intelligence Analyst →](../intelligence_analyst/README.md)
> - [Security Officer →](../security_officer/README.md)
> - [Sensor Operator →](../sensor_operator/README.md)
> - [NATO Liaison Officer →](../nato_liaison/README.md)

---

> **CLASSIFICATION: UNCLASSIFIED** — This document contains no classified information.
