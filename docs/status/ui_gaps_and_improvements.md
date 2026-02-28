# UI Gaps and Improvement Suggestions — Lead Architect Review

**Status:** COMPLETED | **Date:** 2026-02-27 | **Reviewer:** Lead Architect

---

## Overview

This document is a comprehensive analysis of the COP Web Application (`web-cop`) against the documented user-facing features (`docs/user_guide/shared/04_ui_navigation.md`), the component design hierarchy (`docs/architecture/component_design.md` §9), and the specific premium dashboard requirements (Operator UI, Fusion Dashboard, Multi-Domain Dashboard). Findings are verified against the React source code in `web-cop/src/`.

---

## 1. Critical Functional Gaps (FEAT-13: Situational Awareness UI)

### 1.1 Two-Level Role-Based Shell Architecture
| Attribute | Documented / Required | Current State |
|---|---|---|
| **Level 1 — Role Entry Point** | 5 roles (Commander, Analyst, Security, Sensor Operator, NATO Liaison) | 3 roles only (Commander, Analyst, Security) in `RoleSelector.tsx` |
| **Level 2 — Dashboard Views** | Commander → Fusion / Multi-Domain / Operator UI; Analyst → Forensics / Intelligence Search; Security → Audit & Feedback | No Level 2 switching; single layout per role |
| **Sensor Operator** | Sensor Health Monitoring, Data Quality views (`docs/user_guide/sensor_operator/`) | Role missing from `RoleSelector`; `SensorHealthPanel` exists but is always visible as a status bar |
| **NATO Liaison** | NATO Data Exchange, Manual Track Nomination (`docs/user_guide/nato_liaison/`) | Role entirely absent |

> [!CAUTION]
> The user guide documents 5 distinct roles. The UI only implements 3, and even those lack dedicated layouts.

### 1.2 Missing Premium Dashboard Views

| Dashboard | Required Features | Current State |
|---|---|---|
| **Operator UI** | Blurred map background; chronological event timeline with correlation markers; red/amber highlighted critical alerts; glassmorphism panels | Not implemented. Commander role shows generic map+alert grid |
| **Fusion Dashboard** | Map with distinct fused-track vs. raw-sensor icons (Radar, EW, SIGINT); side panel with real-time detection metrics, confidence scores, active track list | Not implemented. `MapView` renders only `FusedTrack` data via `useTrackStream` |
| **Multi-Domain Dashboard** | Wide-angle map (Land, Sea, Air); sensor coverage overlays; signal strengths; domain-specific metrics panels | Not implemented. No domain-filtered views or coverage rendering |

### 1.3 Alert Workflow Actions
| Feature | User Guide (`04_ui_navigation.md`) | Current State |
|---|---|---|
| `[Inspect]` | Opens entity detail panel | Missing — `AlertCard.tsx` only supports acknowledge-on-click |
| `[Confirm]` | Submits `CONFIRM_ANOMALY` feedback | Missing |
| `[Reject]` | Submits `REJECT_ANOMALY` feedback | Missing |
| `[Assign]` | Assigns alert to another operator | Missing |

### 1.4 Entity Detail Panel Completeness
The component design hierarchy (§9.1) defines 5 sub-components inside `DetailPanel`:

| Sub-component | Design Doc | Current State |
|---|---|---|
| `IdentitySection` | Track ID, entity type, hostile status, confidence | Partial — renders basic fields only |
| `PositionSection` | Lat/Lon, speed, heading, altitude, staleness | Partial |
| `SourceAttribution` | Per-sensor confidence breakdown | Missing |
| `EntityTimeline` | Track lifecycle events (created, merged, split) | Missing |
| `FeedbackForm` | Inline feedback submission | Missing |

### 1.5 Map Interactions & Layer Controls
| Feature | User Guide | Current State |
|---|---|---|
| Layers toggle button (bottom-right) | Track labels, trails, sensor coverage, geo-fences, MGRS grid | `uiStore` has `layerVisibility` state but no UI button to toggle; `SensorCoverageLayer` defined in design but not implemented |
| Full-screen map (`F` key) | Toggles fullscreen | Keyboard handler has a placeholder (`break`) — not implemented |
| Dismiss detail panel (`Escape`) | Closes detail | Implemented ✅ |

### 1.6 Search & Keyboard Shortcuts
| Feature | User Guide | Current State |
|---|---|---|
| `Ctrl+F` Search | Search by Track ID, MMSI, Callsign, MGRS | `SearchOverlay.tsx` exists and is wired to `Ctrl+F`, but search is not connected to `TrackStore` for actual entity lookup |
| `Ctrl+Z` Undo filter | Reverts last filter change | Not implemented |
| `Tab` panel cycling | Cycles focus between panels | Not implemented |
| `M` map focus | Focus map view | Placeholder `break` in handler |

### 1.7 Display Modes
| Mode | User Guide | Current State |
|---|---|---|
| Dark | Default | ✅ Implemented |
| NVG-Compatible | Green-on-black | `theme: "nvg"` exists in `uiStore` but no CSS rules apply the NVG palette |
| High-Contrast | Accessibility | Not defined as an option |
| Reduced Bandwidth | Auto-activated at edge | Not implemented |

### 1.8 Collapsible Panes
| Panel | Required | Current State |
|---|---|---|
| Alert Panel | User-collapsible | `toggleAlertPanel` exists in store but no collapse UI affordance (chevron/button) on the panel itself |
| Detail Panel | Collapsible / Escape dismissible | Toggle exists, Escape wired ✅ |
| Forensics Panel | Collapsible | Toggle exists but no UI affordance |
| All panels simultaneously | Collapse all to maximise map | No "focus mode" or collapse-all shortcut |

---

## 2. Premium Aesthetics Gap

| Aspect | Required | Current State |
|---|---|---|
| **Typography** | Modern sans-serif (Inter, Roboto) | `Courier New, monospace` in `index.css` |
| **Glassmorphism** | `backdrop-blur`, translucent panels, frosted card edges | Not implemented; all panels are opaque `#1E293B` |
| **Color Accents** | Vibrant blue (`#3B82F6`) highlights, amber (`#F59E0B`) warnings | Only dark slate palette used; no accent system |
| **Micro-Animations** | Hover effects, transition on collapse, track icon pulse | Only `pulse` keyframe for critical alerts exists |
| **Responsive Layout** | Panels resize gracefully | Fixed pixel widths (e.g., `maxWidth: 420px`) with no breakpoint logic |

---

## 3. Improvement Recommendations (Prioritised)

### P0 — Mission-Critical (Must Have)

1. **Implement full RBAC shell** with all 5 roles and Level-2 dashboard switching.
2. **Build Fusion Dashboard** — distinct fused vs. raw sensor track rendering, confidence metrics side-panel.
3. **Build Operator UI** — timeline view, highlighted alert workflow with `[Inspect]/[Confirm]/[Reject]` actions.
4. **Complete Alert Quick Actions** — wire `FeedbackService.SubmitFeedback` to Confirm/Reject buttons.
5. **Complete Entity Detail Panel** — implement `SourceAttribution`, `EntityTimeline`, `FeedbackForm` sub-components.

### P1 — Operational (Should Have)

6. **Build Multi-Domain Dashboard** — domain-split map with sensor coverage overlays and domain metrics.
7. **Map Layer Toggle UI** — expose `layerVisibility` state via a floating Layers button.
8. **Collapsible Pane UI** — add chevron/toggle affordances on every panel header; implement "focus mode".
9. **NVG and High-Contrast CSS** — implement actual theme variables for `nvg` and add `high-contrast` option.
10. **Implement entity search** — connect `SearchOverlay` to `TrackStore` for real lookups.

### P2 — Premium Polish (Nice to Have)

11. **Modern Typography** — replace Courier New with Inter/Roboto via Google Fonts.
12. **Glassmorphism design system** — define reusable CSS classes for translucent panel backgrounds.
13. **Micro-animations** — add transitions on panel collapse, hover effects on track icons, smooth map panning.
14. **Keyboard shortcuts** — complete all documented shortcuts (`M`, `F`, `Tab`, `Ctrl+Z`).
