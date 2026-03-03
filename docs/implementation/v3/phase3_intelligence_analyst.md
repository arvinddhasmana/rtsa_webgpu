<!-- CLASSIFICATION: UNCLASSIFIED -->
# Phase 3: Intelligence Analyst — Detail Implementation Plan

**Status**: ⬜ Planned
**Role**: Intelligence Analyst
**Use Cases**: UC010, UC011, UC013
**Dashboards**: Forensics (default), Intel Search

---

## Scope

### 3.1 Forensics Panel (Default Level-2 View)
- Right-side panel alongside map
- **Filters**: Time range picker, entity type dropdown, domain filter, spatial bounding box (drawn on map)
- **Query execution**: `QueryService.QueryTracks` + `QueryService.QueryAnomalies`
- **Results table**: Sortable by confidence, time, entity type
- Click result → Entity Detail Panel (shared from Phase 2) with full event timeline

### 3.2 Intel Search
- Search bar: MMSI, callsign, track ID, vessel name
- Auto-complete from recent queries
- Results displayed as cards with track summary
- Click card → Entity Detail Panel with 72-hour event timeline
- Feedback submission from search results (reuses Phase 2 FeedbackForm)

---

## Components to Build

| Component | File | Notes |
|---|---|---|
| `ForensicsPanel.tsx` | `src/components/analyst/` | New — filter form + results table |
| `IntelSearchPanel.tsx` | `src/components/analyst/` | New — search bar + results cards |
| `TimeRangePicker.tsx` | `src/components/shared/` | Shared — reused by Timeline Scrubber |
| `SpatialFilter.tsx` | `src/components/map/` | New — draw bounding box on map |

## Reused from Phase 2
- `EntityDetailPanel.tsx`, `EventTimeline.tsx`, `FeedbackForm.tsx`, `SourceAttribution.tsx`, `ConfidenceHistogram.tsx`, `TimelineScrubber.tsx`

---

## Key APIs

| API | Usage |
|---|---|
| `QueryService.QueryTracks` | Forensics time-range and spatial queries |
| `QueryService.QueryAnomalies` | Anomaly history search |
| `QueryService.GetEventTimeline` | Full entity timeline |
| `FeedbackService.SubmitFeedback` | Analyst feedback on historical tracks |
