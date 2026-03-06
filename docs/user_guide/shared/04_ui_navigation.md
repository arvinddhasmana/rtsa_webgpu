<!-- CLASSIFICATION: UNCLASSIFIED -->

# UI Navigation — Common Interface Patterns

> **CLASSIFICATION: UNCLASSIFIED**
> **Audience**: All RTSA Users
> **Version**: 2.0
> **Last Updated**: 2026-02-28

---

> **⚠ WebGPU COP Migration Notice** — UI navigation steps below reflect the *projected* SolidJS + WebGPU interface. Specific elements may change during implementation. See `docs/architecture/v1/RTSA_WebGPU_Architecture_v1.md`.

## Overview

This section covers the common interface patterns, controls, and keyboard shortcuts that apply to all users of the RTSA COP dashboard. Role-specific features are described in your role's guide.

---

## The Main Toolbar

The toolbar at the top of the screen provides quick access to the primary functions. The toolbar contains two role selectors — your **Role** (Level 1) and your **Dashboard View** (Level 2).

### Level 1 — Role Selector

Select your operator role from the dropdown. The available dashboard views will update automatically to match your role.

| Role | Description |
|---|---|
| **Operations Commander** | Full COP access — Fusion, Multi-Domain, and Operator UI dashboards |
| **Intelligence Analyst** | Historical analysis — Forensics and Intelligence Search views |
| **Security Officer** | Compliance monitoring — Audit & Feedback view |
| **Sensor Operator** | Infrastructure monitoring — Sensor Health dashboard |
| **NATO Liaison** | Partner data exchange — NATO Exchange dashboard |

### Level 2 — Dashboard View Selector

After selecting your role, choose the dashboard view from the tabs that appear. Each view is purpose-built for a specific workflow.

| Role | Available Dashboards |
|---|---|
| Operations Commander | 🔗 **Fusion** (default) · 🌐 **Multi-Domain** · 🎯 **Operator UI** |
| Intelligence Analyst | 🔍 **Forensics** (default) · 📊 **Intel Search** |
| Security Officer | 🔒 **Audit & Feedback** |
| Sensor Operator | 📡 **Sensor Health** |
| NATO Liaison | 🌐 **NATO Exchange** |

### Other Toolbar Controls

| Button / Control | Function |
|---|---|
| ⚙️ **Settings** | Personal display preferences (theme, NVG mode) |
| 👤 **Profile** | Your identity, clearance level, and session info |

> The dashboard view resets to the role's default each time you change your role.

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
│ [Inspect] [Confirm] [Reject] [Assign] │
└─────────────────────────────────┘
```

> If an alert has been assigned, the card shows: **Assigned to: [Operator Name]**

### Alert Quick Actions

| Button | Action | Backend RPC |
|---|---|---|
| **[Inspect]** | Open entity detail panel for this track | — (local navigation) |
| **[Confirm]** | Confirm the anomaly is valid — submits `CONFIRM_ANOMALY` feedback | `FeedbackService.SubmitFeedback` |
| **[Reject]** | Reject as false positive — submits `REJECT_ANOMALY` feedback | `FeedbackService.SubmitFeedback` |
| **[Assign]** | Assign to another operator — opens operator picker dialog | `AlertService.AssignAlert` *(v2.0)* |

> **Note**: `[Confirm]` and `[Reject]` contribute to your operator trust score and model retraining. Use them deliberately.

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
