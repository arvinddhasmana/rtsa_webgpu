<!-- CLASSIFICATION: UNCLASSIFIED -->
# Phase 3 — Dashboard Implementations

> **Phase**: 3 of 4 | **Depends on**: Phase 1 (Backend APIs), Phase 2 (Design System) | **Blocks**: Phase 4
> **Scope**: Operator UI, Fusion Dashboard, Multi-Domain Dashboard, shared component enhancements

---

## Purpose

Build the three premium dashboard views that replace the Phase 2 placeholders. Each dashboard composes the shared design system (glassmorphism, collapsible panels, design tokens) with dashboard-specific data streams and visualisations.

---

## Step 1: Fusion Dashboard (Commander Default)

### 1.1 Create the hook `useSensorStream.ts`

Create `web-cop/src/hooks/useSensorStream.ts`. This hook should:

1. Import the gRPC-Web client for `TrackService.StreamSensorObservations` from the generated TypeScript stubs.
2. Accept optional parameters: `sensorTypes`, `boundingBox`, `clearanceLevel`.
3. On mount, open a server-streaming gRPC call to `StreamSensorObservations`.
4. On each `SensorObservationUpdate` message received:
   - Parse the `observation` payload to extract position, sensor_type, and sensor-specific metadata.
   - Store the observations in a Zustand store (or local `useState` / `useRef` map keyed by `observation_id`).
   - Observations older than 60 seconds should be automatically evicted from the map.
5. Return `{ sensorObservations: Map<string, SensorObservationUpdate>, isConnected: boolean }`.
6. On unmount, cancel the gRPC stream.

### 1.2 Create `FusionSidePanel.tsx`

Create `web-cop/src/components/fusion/FusionSidePanel.tsx`. This component displays real-time fusion metrics in a glassmorphism side panel. Wrap it in the `CollapsiblePanel` from Phase 2 with `direction="horizontal"` and `defaultSize="340px"`.

The panel should contain the following sections:

**Section A — Detection Metrics** (top)
- Total active tracks (from `useTrackStore`)
- Tracks by domain: Surface, Air, Subsurface, Land, Cyber — each with a count and a small coloured dot
- Average confidence score across all active tracks
- Active sensor count (from `useSensorStream`)

**Section B — Confidence Breakdown** (middle)
- A simple horizontal bar chart or percentage bars showing the confidence score distribution of active tracks in 4 buckets: High (≥0.8), Medium (0.6–0.79), Low (0.4–0.59), Tentative (<0.4)
- Use CSS custom properties: High → `var(--accent-green)`, Medium → `var(--accent-blue)`, Low → `var(--accent-amber)`, Tentative → `var(--accent-red)`

**Section C — Active Track List** (bottom, scrollable)
- A list of the top 50 active tracks sorted by confidence score descending.
- Each row shows: track_id (abbreviated), entity_type icon, hostile_class badge, confidence score, source_count, and age.
- Clicking a row should call `useUIStore.getState().openDetailPanel(trackId)` to show the detail panel.

Style all text using `var(--text-primary)` and `var(--text-secondary)`. Use `var(--glass-bg)` for section backgrounds.

### 1.3 Create `FusionDashboard.tsx`

Create `web-cop/src/components/layout/FusionDashboard.tsx`. This is the main layout component rendered when `activeDashboardView === "fusion"`.

Layout:
```
┌─────────────────────────────────────────────────┬──────────────────┐
│                                                 │                  │
│                   MapView                       │  FusionSidePanel │
│   (full-colour, raw sensor icons + fused icons) │  (collapsible)   │
│                                                 │                  │
├─────────────────────────────────────────────────┴──────────────────┤
│                    DetailPanel (collapsible, 280px height)         │
└───────────────────────────────────────────────────────────────────┘
```

1. Render `<MapView />` in the main area.
2. Render `<FusionSidePanel />` to the right, wrapped in `<CollapsiblePanel>`.
3. If `detailPanelOpen` is true, render `<DetailPanel />` below the map in a collapsible bottom sheet.
4. The MapView in this mode should render two layers:
   - The existing fused track layer (from `useTrackStream`)
   - A new raw sensor observation layer (from `useSensorStream`) rendered with distinct icons per sensor type:
     - Radar: diamond (◇) in light blue
     - EW/SIGINT: triangle (△) in amber
     - ELINT: square (◻) in purple
     - ISR: pentagon in white
     - AIS: circle (○) in green
     - Cyber: hexagon in red

For the raw sensor layer, you can add a new layer to MapView using MapLibre's `addSource` and `addLayer` APIs for GeoJSON point data, or you can render it as HTML markers using MapLibre's marker API. The positions come from `sensorObservation.observation.position`.

### 1.4 Wire into `MainLayout.tsx`

In `MainLayout.tsx`, import `FusionDashboard` and replace the Phase 2 placeholder:

```tsx
{activeDashboardView === "fusion" && <FusionDashboard />}
```

---

## Step 2: Operator UI (Commander Option)

### 2.1 Create the hook `useEventTimeline.ts`

Create `web-cop/src/hooks/useEventTimeline.ts`. This hook should:

1. Import the gRPC-Web client for `QueryService.GetEventTimeline`.
2. Accept parameters: `trackId: string`, `timeRange: { start: Date, end: Date }`.
3. Use `@tanstack/react-query` to fetch and cache the timeline data.
4. Return `{ events: TimelineEvent[], isLoading: boolean, error: Error | null }`.
5. Auto-refetch every 30 seconds while the component is mounted.

### 2.2 Create `TimelineView.tsx`

Create `web-cop/src/components/timeline/TimelineView.tsx`. This component renders a vertical chronological timeline of events. Wrap it in `<CollapsiblePanel>` with `direction="vertical"` and `defaultSize="320px"`.

The timeline should:

1. Accept a `trackId` prop. If no track is selected, show all recent events (from the alert stream).
2. Render events in a vertical list with timestamps on the left and event details on the right, connected by a vertical line.
3. Use colour coding for event types:
   - Track state changes → `var(--accent-blue)`
   - Anomaly detections → `var(--accent-red)` for CRITICAL, `var(--accent-amber)` for ELEVATED
   - Feedback submissions → `var(--accent-green)`
   - Audit events → `var(--text-muted)`
4. Each event node should show:
   - Timestamp (formatted as `HH:MM:SS UTC`)
   - Event type icon and label
   - Summary text from the `TimelineEvent.summary` field
   - For anomaly events: the confidence score as a small badge
5. Correlation markers: when two events from different sources reference the same entity within a 60-second window, draw a horizontal dashed line connecting them with a label "Correlated".

### 2.3 Enhance `AlertCard.tsx` with Quick Actions

Open `web-cop/src/components/alerts/AlertCard.tsx`. Add four action buttons below the existing alert content:

1. **[Inspect]** — calls `openDetailPanel(alert.track_id)` on the UI store
2. **[Confirm]** — calls `feedbackClient.submitFeedback({ trackId: alert.track_id, feedbackType: CONFIRM_ANOMALY, alertId: alert.alert_id, ... })` using the gRPC feedback client
3. **[Reject]** — calls `feedbackClient.submitFeedback({ ..., feedbackType: REJECT_ANOMALY })`
4. **[Assign]** — opens a small modal/popover asking for the Assignee Operator ID, then calls `alertClient.assignAlert({ alertId, assigneeOperatorId, ... })`

Style the buttons:
- `[Inspect]`: outline style, `var(--accent-blue)` border
- `[Confirm]`: filled, `var(--accent-green)` background
- `[Reject]`: filled, `var(--accent-red)` background
- `[Assign]`: outline style, `var(--accent-amber)` border

After a Confirm or Reject action, show a brief success toast and visually mark the alert as "actioned" (reduce opacity, show a checkmark/X).

### 2.4 Create `OperatorDashboard.tsx`

Create `web-cop/src/components/layout/OperatorDashboard.tsx`. Layout:

```
┌─────────────────────────────────────────────────────────────────┐
│                                                                 │
│           MapView (blurred with backdrop-filter)                │
│                                                                 │
├─────────────────────────────┬───────────────────────────────────┤
│                             │                                   │
│   AlertPanel (collapsible)  │   TimelineView (collapsible)      │
│   Left side, 30% width     │   Right side, 40% width           │
│   Enhanced with quick       │   Chronological events            │
│   action buttons            │                                   │
│                             │                                   │
└─────────────────────────────┴───────────────────────────────────┘
```

Key differences from Fusion Dashboard:
1. The MapView should have a CSS `filter: blur(3px) brightness(0.6)` applied over it, creating the blurred background effect.
2. The alert and timeline panels should be rendered ON TOP of the blurred map using absolute positioning or a z-index layer, with `glass-panel` styling.
3. The AlertPanel in this view should show only CRITICAL and ELEVATED alerts by default.

### 2.5 Wire into `MainLayout.tsx`

```tsx
{activeDashboardView === "operator" && <OperatorDashboard />}
```

---

## Step 3: Multi-Domain Dashboard (Commander Option)

### 3.1 Create `SensorCoverageLayer.tsx`

Create `web-cop/src/components/map/SensorCoverageLayer.tsx`. This component should:

1. Fetch sensor statuses using a new hook `useSensorStatuses` that calls `IngestionService.ListSensorStatuses`.
2. For each sensor with coverage geometry:
   - **Radar**: Render a fan/sector polygon using `sensor_position`, `range_nm`, `bearing_start_degrees`, `bearing_end_degrees`. Use a translucent blue fill.
   - **EW/SIGINT**: Render a circle polygon using `sensor_position` and `range_nm`. Use a translucent amber fill.
   - **ISR**: Render the `coverage_polygon` directly. Use a translucent green fill.
3. Add these as MapLibre `fill` layers with opacity 0.15 and `stroke` layers with opacity 0.5.
4. Add labels at each sensor position showing the `sensor_id`.

### 3.2 Create `DomainMetricsOverlay.tsx`

Create `web-cop/src/components/dashboard/DomainMetricsOverlay.tsx`. This is a floating overlay rendered on top of the map, positioned at the top-left corner. It should:

1. Show domain-split KPI cards in a horizontal row:
   - **Air** (✈): count of active air tracks, blue accent
   - **Surface** (🚢): count of surface tracks, cyan accent
   - **Subsurface** (🌊): count of subsurface tracks, teal accent
   - **Land** (🏔): count of land tracks, green accent
   - **Cyber** (💻): count of cyber tracks, red accent
2. Each card should be a small `glass-panel` with:
   - Domain icon + label
   - Track count (large number)
   - Observation rate (observations/sec from `useSensorStream`)
   - Hostile count (subset with `hostile_class = HOSTILE` or `SUSPECT`)
3. Below the domain cards, show a sensor throughput bar:
   - A horizontal bar for each sensor type showing observations per second
   - Use proportional width and color-code by sensor type

### 3.3 Create `MultiDomainDashboard.tsx`

Create `web-cop/src/components/layout/MultiDomainDashboard.tsx`. Layout:

```
┌─────────────────────────────────────────────────────────────────┐
│  DomainMetricsOverlay (floating, top-left)                      │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │                                                             ││
│  │         MapView (full width, sensor coverage layers ON)     ││
│  │         + Dense fused track distribution                    ││
│  │         + Raw sensor observation icons                      ││
│  │                                                             ││
│  └─────────────────────────────────────────────────────────────┘│
│  AlertPanel (collapsed by default, bottom strip)                │
└─────────────────────────────────────────────────────────────────┘
```

Key design points:
1. The map should be maximised — all panels collapsed by default.
2. `SensorCoverageLayer` should be enabled by default in this view.
3. The `DomainMetricsOverlay` floats over the map using `position: absolute; top: 12px; left: 12px; z-index: 10`.
4. The AlertPanel is rendered as a thin collapsed strip at the bottom (expand on click) showing only the total unacknowledged alert count.

### 3.4 Wire into `MainLayout.tsx`

```tsx
{activeDashboardView === "multi-domain" && <MultiDomainDashboard />}
```

---

## Step 4: Enhance Shared Components

### 4.1 Complete the Entity Detail Panel

Open `web-cop/src/components/detail/DetailPanel.tsx`. Add or implement the following sub-components:

**SourceAttribution section**: For the selected track, iterate over `track.sources` and render a table showing `sensor_id`, `sensor_type` (with icon), `confidence`, `last_contribution`, `observation_count`. Sort by confidence descending.

**EntityTimeline section**: Embed a mini version of `TimelineView` filtered to the selected track's `track_id`. Show the last 20 events.

**FeedbackForm section**: Add a form with:
- Radio buttons for feedback type: Confirm Hostile, Confirm Friendly, Reclassify, Reject Anomaly, Confirm Anomaly
- A text area for justification (required, min 10 characters)
- A submit button that calls `feedbackClient.submitFeedback()`
- After submission, display the returned `trust_score` and show a success indicator

### 4.2 Add Map Layer Toggle Button

Create `web-cop/src/components/map/MapLayerToggle.tsx`. This component should:

1. Render a floating button in the bottom-right corner of the map: "Layers" with a stacked-layers icon.
2. On click, expand a small `glass-panel` popover with toggle switches for each layer from `uiStore.layerVisibility`:
   - Track Labels
   - Track Trails
   - Sensor Coverage
   - Geo-fences
   - MGRS Grid
3. Each toggle calls `toggleLayerVisibility(layerKey)` from the store.
4. Add the component inside `MapView.tsx`.

### 4.3 Connect Search to Track Store

Open `web-cop/src/components/layout/SearchOverlay.tsx`. Connect the search to real data:

1. When the user types a query, search the `trackStore.tracks` map for matches by:
   - `track_id` (contains match)
   - `label` (contains match)
   - `sources[].sensor_id` (exact match for MMSI, callsign embedded in sensor metadata)
2. Display results as a dropdown list below the search input.
3. Clicking a result should: centre the map on the track's position, open the detail panel, and close the search overlay.

---

## Verification

### Automated

1. Build the web-cop:
   ```bash
   cd web-cop && npm run build
   ```

2. Run unit tests:
   ```bash
   cd web-cop && npm run test
   ```

3. Run E2E tests:
   ```bash
   cd web-cop && npm run test:e2e
   ```

4. Add new Playwright E2E tests:
   - `fusion-dashboard.spec.ts`: Verify Commander role defaults to Fusion view, FusionSidePanel is visible, detection metrics render
   - `operator-dashboard.spec.ts`: Verify Commander can switch to Operator view, map is blurred, timeline renders, alert quick-actions are visible
   - `multi-domain-dashboard.spec.ts`: Verify Commander can switch to Multi-Domain view, domain metrics overlay is visible, all panels collapsed by default

### Visual Verification

1. Start dev server and the backend stack
2. Select **Operations Commander** role
3. Verify:
   - [ ] Fusion Dashboard loads by default with map + side panel
   - [ ] Side panel shows detection metrics with real data from gRPC streams
   - [ ] Raw sensor icons appear on the map with correct shapes/colours
   - [ ] Clicking a track in the list opens the detail panel
   - [ ] Side panel collapses and expands with animation
4. Switch to **Operator UI**:
   - [ ] Map is blurred in the background
   - [ ] Alert panel shows CRITICAL/ELEVATED alerts with [Inspect]/[Confirm]/[Reject]/[Assign] buttons
   - [ ] Timeline view shows chronological events
   - [ ] Clicking [Confirm] submits feedback and marks the alert
5. Switch to **Multi-Domain Dashboard**:
   - [ ] Map is maximised, all panels collapsed
   - [ ] Domain metrics overlay shows track counts per domain
   - [ ] Sensor coverage areas are visible on the map
   - [ ] Bottom alert strip shows count and expands on click
