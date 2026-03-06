<!-- CLASSIFICATION: UNCLASSIFIED -->
# Phase 2: Operations Commander — Detail Implementation Plan

**Status**: ⬜ Planned
**Role**: Operations Commander
**Use Cases**: UC008, UC009, UC010, UC012, UC016
**Dashboards**: Fusion (default), Multi-Domain, Operator UI

---

## Scope

### 2.1 Fusion Dashboard (Default Level-2 View)
- **Map layer**: Fused tracks (filled circles, color-coded by hostile classification) + raw sensor observations (diamonds, triangles, squares) via `StreamSensorObservations`
- **Fusion Side Panel** (collapsible right panel):
  - Active tracks by domain with color-coded dots
  - Average confidence score
  - Confidence histogram (High/Medium/Low/Tentative)
  - Scrollable track list sorted by confidence → click to open detail panel

### 2.2 Multi-Domain Dashboard
- Fullscreen map (all panels collapsed)
- **Domain Metrics Overlay** (floating, top-left): Air, Surface, Subsurface, Land, Cyber cards with track count, obs/sec, hostile count
- Sensor coverage overlays toggled via Layers button
- Collapsed alert strip at bottom (click to expand)

### 2.3 Operator UI Dashboard
- Blurred map background with glassmorphism overlay panels
- **Alert Panel** (left): CRITICAL (red-pulse) and ELEVATED (amber) alerts with `[Inspect]` `[Confirm]` `[Reject]` `[Assign]` buttons
- **Event Timeline** (right): Unified 4-source timeline via `GetEventTimeline` RPC
- **Entity Detail Panel**: Track identity, source attribution table, entity timeline, feedback form

### 2.4 Timeline Scrubber
- Horizontal slider control for historical map playback
- Scrubs `QueryService.QueryTracks` with time range filter
- Playback speed control (1x, 2x, 5x, 10x)
- Integration with Forensics panel (Phase 3)

---

## Key APIs

| API | Usage |
|---|---|
| `TrackService.StreamTracks` | Live fused track updates on map |
| `TrackService.StreamSensorObservations` | Raw pre-fusion sensor data on Fusion Dashboard |
| `QueryService.GetEventTimeline` | Unified timeline (tracks + anomalies + feedback + audit) |
| `AlertService.StreamAlerts` | Live alert updates |
| `AlertService.AssignAlert` | Operator assigns alert to another operator |
| `FeedbackService.SubmitFeedback` | Confirm/Reject anomaly, reclassify track |
| `QueryService.QueryTracks` | Historical track data for Timeline Scrubber |

---

## Components to Build

| Component | File | Shared? |
|---|---|---|
| `FusionDashboard.tsx` | `src/components/layout/` | Commander only |
| `FusionSidePanel.tsx` | `src/components/fusion/` | Commander only |
| `ConfidenceHistogram.tsx` | `src/components/fusion/` | Reused by Analyst |
| `MultiDomainDashboard.tsx` | `src/components/layout/` | Commander only |
| `DomainMetricsOverlay.tsx` | `src/components/map/` | Shared (Analyst, Commander) |
| `OperatorDashboard.tsx` | `src/components/layout/` | Commander only |
| `EntityDetailPanel.tsx` | `src/components/entity/` | Shared (all roles) |
| `EventTimeline.tsx` | `src/components/timeline/` | Shared (Commander, Analyst) |
| `SourceAttribution.tsx` | `src/components/entity/` | Shared |
| `FeedbackForm.tsx` | `src/components/feedback/` | Shared (Commander, Analyst, Security) |
| `TimelineScrubber.tsx` | `src/components/timeline/` | Shared (Commander, Analyst) |
| `AlertAssignPopover.tsx` | `src/components/alert/` | Commander only |

---

## Hooks to Build

| Hook | Purpose |
|---|---|
| `useTrackStream.ts` | Exists — may need enhancement for raw sensor overlay |
| `useAlertStream.ts` | Stream live alerts from `AlertService` |
| `useEventTimeline.ts` | Fetch unified timeline for a track via `GetEventTimeline` |
| `useFeedback.ts` | Submit feedback + receive trust score |

---

## Estimated Effort
- ~8-10 components, ~4 hooks
- Heavy map integration (MapLibre layers for raw sensors, domain overlays)
- Timeline Scrubber is the most complex new interaction pattern
