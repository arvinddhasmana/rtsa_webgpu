<!-- CLASSIFICATION: UNCLASSIFIED -->

# Module 15 — COP Web Application (React)

> **Module**: 15-cop-web-app
> **Phase**: P4 (UI)
> **Dependencies**: Module 02 (protos), Module 14 (API Gateway/Envoy)
> **Agent**: `@greatest-ever-developer`
> **Estimated Effort**: 8 days

---

## 1. Objective

Implement the Common Operating Picture (COP) Web Application — a React 18 + TypeScript 5 single-page application that provides real-time situational awareness to military operators. The app connects to backend services via gRPC-Web through the Envoy API Gateway.

**v2.0 Enhancement**: Implement the Two-Level RBAC shell supporting all five operator roles (Commander, Analyst, Security, Sensor Operator, NATO Liaison), three premium dashboard views (Fusion, Multi-Domain, Operator UI), and a modern glassmorphism design system using Inter typography and CSS custom properties.

**Acceptance Criteria**:

- Real-time map view with track plotting (MapLibre GL JS)
- Threat halos and geo-fence overlays
- Alert panel with severity-based filtering and acknowledgment
- Entity detail panel with source attribution and feedback form
- Forensics panel with historical query builder
- Zustand state management (4 stores)
- gRPC-Web streaming hooks (`useTrackStream`, `useAlertStream`)
- Offline mode with Service Worker caching
- Responsive layout for large-screen operator workstations
- Classification banner on ALL views
- ≥80% line coverage (Vitest)
- **v2.0**: Two-Level RBAC shell — `RoleSelector` (Level 1) + `DashboardSelector` (Level 2)
- **v2.0**: All 5 roles in `RoleSelector`; `uiStore` extended with `activeDashboardView`
- **v2.0**: `FusionDashboard`, `OperatorDashboard` layout components
- **v2.0**: `useSensorStream` hook for `StreamSensorObservations`
- **v2.0**: Alert quick-actions: `[Inspect]`, `[Confirm]`, `[Reject]`, `[Assign]`
- **v2.0**: `EntityTimeline` sub-component using `GetEventTimeline` RPC
- **v2.0**: Inter font + CSS design token system + NVG theme
- **v2.0**: Dev server on port **5173** (Vite default)

---

## 2. Project Structure

```
web-cop/
├── public/
│   ├── index.html
│   ├── favicon.ico
│   └── sw.js                        # Service Worker for offline
├── src/
│   ├── main.tsx                     # Entry point
│   ├── App.tsx                      # Root component (AuthProvider → MainLayout)
│   ├── vite-env.d.ts
│   ├── api/
│   │   ├── grpc-client.ts           # gRPC-Web client initialization
│   │   ├── track-client.ts          # TrackService client
│   │   ├── alert-client.ts          # AlertService client
│   │   ├── query-client.ts          # QueryService client
│   │   ├── feedback-client.ts       # FeedbackService client
│   │   └── audit-client.ts          # AuditService client
│   ├── hooks/
│   │   ├── useTrackStream.ts        # Real-time track streaming
│   │   ├── useAlertStream.ts        # Real-time alert streaming
│   │   ├── useConnectionStatus.ts   # Backend connectivity monitor
│   │   ├── useOfflineMode.ts        # Offline detection + queue
│   │   └── useClassification.ts     # Classification context hook
│   ├── stores/
│   │   ├── trackStore.ts            # Zustand: active tracks
│   │   ├── alertStore.ts            # Zustand: alert queue
│   │   ├── authStore.ts             # Zustand: operator identity
│   │   └── uiStore.ts              # Zustand: panel state, filters
│   ├── components/
│   │   ├── layout/
│   │   │   ├── MainLayout.tsx       # Grid layout with panels
│   │   │   ├── ClassificationBanner.tsx  # Top/bottom classification banner
│   │   │   └── ConnectionIndicator.tsx
│   │   ├── map/
│   │   │   ├── MapView.tsx          # MapLibre GL container
│   │   │   ├── TrackLayer.tsx       # Track marker layer
│   │   │   ├── ThreatHaloLayer.tsx  # Threat proximity circles
│   │   │   ├── GeoFenceLayer.tsx    # Geo-fence polygon overlay
│   │   │   └── SensorCoverageLayer.tsx  # Sensor coverage arcs
│   │   ├── alerts/
│   │   │   ├── AlertPanel.tsx       # Alert list panel
│   │   │   ├── AlertCard.tsx        # Individual alert card
│   │   │   └── AlertFilter.tsx      # Severity/type filter controls
│   │   ├── detail/
│   │   │   ├── DetailPanel.tsx      # Entity detail panel
│   │   │   ├── IdentitySection.tsx  # Track identity info
│   │   │   ├── PositionSection.tsx  # Position / velocity display
│   │   │   ├── SourceAttribution.tsx # Contributing sensor list
│   │   │   ├── EntityTimeline.tsx   # Track history timeline
│   │   │   └── FeedbackForm.tsx     # Operator feedback submission
│   │   ├── forensics/
│   │   │   ├── ForensicsPanel.tsx   # Historical analysis panel
│   │   │   ├── QueryBuilder.tsx     # Query parameter form
│   │   │   ├── ResultsView.tsx      # Tabular results display
│   │   │   └── MapReplay.tsx        # Historical track replay on map
│   │   └── auth/
│   │       └── AuthProvider.tsx     # Authentication context
│   ├── types/
│   │   ├── track.ts                 # TypeScript types for tracks
│   │   ├── alert.ts                 # TypeScript types for alerts
│   │   ├── feedback.ts              # TypeScript types for feedback
│   │   └── common.ts               # Shared enums, ClassificationLevel
│   ├── utils/
│   │   ├── classification.ts        # Classification helpers
│   │   ├── coordinates.ts           # WGS-84 formatting
│   │   ├── time.ts                  # Time formatting (Zulu)
│   │   └── mil-symbology.ts         # MIL-STD-2525 symbol mapping
│   └── __tests__/
│       ├── stores/
│       │   ├── trackStore.test.ts
│       │   ├── alertStore.test.ts
│       │   ├── authStore.test.ts
│       │   └── uiStore.test.ts
│       ├── hooks/
│       │   ├── useTrackStream.test.ts
│       │   └── useAlertStream.test.ts
│       └── components/
│           ├── AlertCard.test.tsx
│           ├── DetailPanel.test.tsx
│           ├── FeedbackForm.test.tsx
│           └── ClassificationBanner.test.tsx
├── package.json
├── tsconfig.json
├── vite.config.ts
├── vitest.config.ts
├── Dockerfile
└── README.md
```

---

## 3. Technology Stack

| Technology             | Version | Purpose                                        |
| ---------------------- | ------- | ---------------------------------------------- |
| React                  | 18.x    | UI framework                                   |
| TypeScript             | 5.x     | Type safety                                    |
| Vite                   | 5.x     | Build tool                                     |
| Vitest                 | 1.x     | Unit testing                                   |
| Zustand                | 4.x     | State management                               |
| MapLibre GL JS         | 4.x     | Map rendering (open-source, no vendor lock-in) |
| grpc-web               | 1.5.x   | gRPC-Web client                                |
| protobuf-ts            | 2.x     | TypeScript protobuf runtime                    |
| @tanstack/react-query  | 5.x     | Server state for unary RPCs                    |
| Tailwind CSS           | 3.x     | Utility-first styling                          |
| @testing-library/react | 14.x    | Component testing                              |

---

## 4. State Management (Zustand Stores)

### 4.1 TrackStore

```typescript
// CLASSIFICATION: UNCLASSIFIED
// src/stores/trackStore.ts

import { create } from "zustand";
import { FusedTrack } from "../types/track";

interface TrackState {
  // State
  tracks: Map<string, FusedTrack>; // Indexed by track_id
  selectedTrackId: string | null;
  lastUpdateTime: Date | null;

  // Actions
  upsertTrack: (track: FusedTrack) => void;
  removeTrack: (trackId: string) => void;
  selectTrack: (trackId: string | null) => void;
  clearAll: () => void;

  // Selectors
  getTrackById: (trackId: string) => FusedTrack | undefined;
  getTracksByType: (entityType: string) => FusedTrack[];
  getHostileTracks: () => FusedTrack[];
  getActiveTrackCount: () => number;
}

export const useTrackStore = create<TrackState>((set, get) => ({
  tracks: new Map(),
  selectedTrackId: null,
  lastUpdateTime: null,

  upsertTrack: (track) =>
    set((state) => {
      const newTracks = new Map(state.tracks);
      newTracks.set(track.trackId, track);
      return { tracks: newTracks, lastUpdateTime: new Date() };
    }),

  removeTrack: (trackId) =>
    set((state) => {
      const newTracks = new Map(state.tracks);
      newTracks.delete(trackId);
      return { tracks: newTracks };
    }),

  selectTrack: (trackId) => set({ selectedTrackId: trackId }),
  clearAll: () => set({ tracks: new Map(), selectedTrackId: null }),

  getTrackById: (trackId) => get().tracks.get(trackId),
  getTracksByType: (entityType) =>
    Array.from(get().tracks.values()).filter(
      (t) => t.entityType === entityType,
    ),
  getHostileTracks: () =>
    Array.from(get().tracks.values()).filter(
      (t) => t.hostileClass === "HOSTILE",
    ),
  getActiveTrackCount: () =>
    Array.from(get().tracks.values()).filter((t) => t.status === "ACTIVE")
      .length,
}));
```

### 4.2 AlertStore

```typescript
// CLASSIFICATION: UNCLASSIFIED
// src/stores/alertStore.ts

import { create } from "zustand";
import { AnomalyAlert } from "../types/alert";

interface AlertState {
  // State
  alerts: Map<string, AnomalyAlert>; // Indexed by alert_id
  acknowledgedIds: Set<string>;
  minSeverityFilter: string; // 'WATCH' | 'ELEVATED' | 'CRITICAL'

  // Actions
  addAlert: (alert: AnomalyAlert) => void;
  acknowledgeAlert: (alertId: string) => void;
  setMinSeverityFilter: (severity: string) => void;
  clearAll: () => void;

  // Selectors
  getUnacknowledgedAlerts: () => AnomalyAlert[];
  getCriticalCount: () => number;
  getFilteredAlerts: () => AnomalyAlert[];
}

export const useAlertStore = create<AlertState>((set, get) => ({
  alerts: new Map(),
  acknowledgedIds: new Set(),
  minSeverityFilter: "WATCH",

  addAlert: (alert) =>
    set((state) => {
      const newAlerts = new Map(state.alerts);
      newAlerts.set(alert.alertId, alert);
      return { alerts: newAlerts };
    }),

  acknowledgeAlert: (alertId) =>
    set((state) => {
      const newAcked = new Set(state.acknowledgedIds);
      newAcked.add(alertId);
      return { acknowledgedIds: newAcked };
    }),

  setMinSeverityFilter: (severity) => set({ minSeverityFilter: severity }),
  clearAll: () => set({ alerts: new Map(), acknowledgedIds: new Set() }),

  getUnacknowledgedAlerts: () =>
    Array.from(get().alerts.values())
      .filter((a) => !get().acknowledgedIds.has(a.alertId))
      .sort((a, b) => severityRank(b.severity) - severityRank(a.severity)),

  getCriticalCount: () =>
    Array.from(get().alerts.values()).filter(
      (a) => a.severity === "CRITICAL" && !get().acknowledgedIds.has(a.alertId),
    ).length,

  getFilteredAlerts: () => {
    const minRank = severityRank(get().minSeverityFilter);
    return Array.from(get().alerts.values())
      .filter((a) => severityRank(a.severity) >= minRank)
      .sort((a, b) => severityRank(b.severity) - severityRank(a.severity));
  },
}));

function severityRank(s: string): number {
  switch (s) {
    case "CRITICAL":
      return 3;
    case "ELEVATED":
      return 2;
    case "WATCH":
      return 1;
    default:
      return 0;
  }
}
```

### 4.3 AuthStore

```typescript
// CLASSIFICATION: UNCLASSIFIED
// src/stores/authStore.ts

import { create } from "zustand";
import { ClassificationLevel } from "../types/common";

interface AuthState {
  operatorId: string | null;
  operatorName: string | null;
  unit: string | null;
  clearanceLevel: ClassificationLevel;
  roles: string[];
  isAuthenticated: boolean;

  // Actions
  setOperator: (operator: {
    id: string;
    name: string;
    unit: string;
    clearance: ClassificationLevel;
    roles: string[];
  }) => void;
  logout: () => void;

  // Selectors
  canAccess: (dataClassification: ClassificationLevel) => boolean;
  hasRole: (role: string) => boolean;
}

export const useAuthStore = create<AuthState>((set, get) => ({
  operatorId: null,
  operatorName: null,
  unit: null,
  clearanceLevel: "UNCLASSIFIED",
  roles: [],
  isAuthenticated: false,

  setOperator: (op) =>
    set({
      operatorId: op.id,
      operatorName: op.name,
      unit: op.unit,
      clearanceLevel: op.clearance,
      roles: op.roles,
      isAuthenticated: true,
    }),

  logout: () =>
    set({
      operatorId: null,
      operatorName: null,
      unit: null,
      clearanceLevel: "UNCLASSIFIED",
      roles: [],
      isAuthenticated: false,
    }),

  canAccess: (dataClassification) => {
    const levels: ClassificationLevel[] = [
      "UNCLASSIFIED",
      "PROTECTED_A",
      "PROTECTED_B",
      "PROTECTED_C",
      "SECRET",
    ];
    return (
      levels.indexOf(get().clearanceLevel) >= levels.indexOf(dataClassification)
    );
  },

  hasRole: (role) => get().roles.includes(role),
}));
```

### 4.4 UIStore

```typescript
// CLASSIFICATION: UNCLASSIFIED
// src/stores/uiStore.ts

import { create } from "zustand";

interface UIState {
  // Panel visibility
  alertPanelOpen: boolean;
  detailPanelOpen: boolean;
  forensicsPanelOpen: boolean;

  // Map state
  mapCenter: [number, number]; // [lng, lat]
  mapZoom: number;

  // Filters
  entityTypeFilter: string[]; // Empty = show all
  hostileClassFilter: string[]; // Empty = show all
  showStaleTracksEnabled: boolean;

  // Actions
  toggleAlertPanel: () => void;
  toggleDetailPanel: () => void;
  toggleForensicsPanel: () => void;
  setMapView: (center: [number, number], zoom: number) => void;
  setEntityTypeFilter: (types: string[]) => void;
  setHostileClassFilter: (classes: string[]) => void;
  toggleStaleTracks: () => void;
}

export const useUIStore = create<UIState>((set) => ({
  alertPanelOpen: true,
  detailPanelOpen: false,
  forensicsPanelOpen: false,

  // Default to Mid-Atlantic (43-47°N, 55-65°W)
  mapCenter: [-60.0, 45.0],
  mapZoom: 6,

  entityTypeFilter: [],
  hostileClassFilter: [],
  showStaleTracksEnabled: false,

  toggleAlertPanel: () => set((s) => ({ alertPanelOpen: !s.alertPanelOpen })),
  toggleDetailPanel: () =>
    set((s) => ({ detailPanelOpen: !s.detailPanelOpen })),
  toggleForensicsPanel: () =>
    set((s) => ({ forensicsPanelOpen: !s.forensicsPanelOpen })),
  setMapView: (center, zoom) => set({ mapCenter: center, mapZoom: zoom }),
  setEntityTypeFilter: (types) => set({ entityTypeFilter: types }),
  setHostileClassFilter: (classes) => set({ hostileClassFilter: classes }),
  toggleStaleTracks: () =>
    set((s) => ({ showStaleTracksEnabled: !s.showStaleTracksEnabled })),
}));
```

---

## 5. Real-Time Streaming Hooks

### 5.1 useTrackStream

```typescript
// CLASSIFICATION: UNCLASSIFIED
// src/hooks/useTrackStream.ts

import { useEffect, useRef, useCallback } from "react";
import { useTrackStore } from "../stores/trackStore";
import { useAuthStore } from "../stores/authStore";
import { useUIStore } from "../stores/uiStore";
import { trackClient } from "../api/track-client";

/**
 * useTrackStream — subscribes to real-time track updates via gRPC-Web server-streaming.
 *
 * Flow:
 *   1. Opens StreamTracks gRPC-Web stream to svc-track via Envoy
 *   2. Sends StreamTracksRequest with filters from UIStore + clearance from AuthStore
 *   3. Receives initial SNAPSHOT of all current tracks
 *   4. Then receives incremental updates (new, updated, removed tracks)
 *   5. Updates TrackStore on each message
 *   6. On disconnect: exponential backoff reconnect (1s, 2s, 4s, max 30s)
 *   7. Cleans up stream on unmount
 *
 * @returns { isConnected, error, reconnectAttempts }
 */
export function useTrackStream() {
  const upsertTrack = useTrackStore((s) => s.upsertTrack);
  const removeTrack = useTrackStore((s) => s.removeTrack);
  const clearance = useAuthStore((s) => s.clearanceLevel);
  const entityTypeFilter = useUIStore((s) => s.entityTypeFilter);
  // ... implementation
}
```

### 5.2 useAlertStream

```typescript
// CLASSIFICATION: UNCLASSIFIED
// src/hooks/useAlertStream.ts

/**
 * useAlertStream — subscribes to real-time alert updates via gRPC-Web server-streaming.
 *
 * Flow:
 *   1. Opens StreamAlerts gRPC-Web stream to svc-alert via Envoy
 *   2. Sends StreamAlertsRequest with min_severity from AlertStore
 *   3. Receives alerts in priority order (CRITICAL first)
 *   4. Updates AlertStore on each message
 *   5. Triggers browser notification for CRITICAL alerts
 *   6. On disconnect: exponential backoff reconnect
 *
 * @returns { isConnected, error, reconnectAttempts }
 */
export function useAlertStream() {
  const addAlert = useAlertStore((s) => s.addAlert);
  const minSeverity = useAlertStore((s) => s.minSeverityFilter);
  // ... implementation
}
```

### 5.3 useConnectionStatus

```typescript
// CLASSIFICATION: UNCLASSIFIED
// src/hooks/useConnectionStatus.ts

/**
 * useConnectionStatus — monitors connectivity to backend services.
 *
 * Uses gRPC Health Check protocol through Envoy.
 * Polls every 10 seconds.
 * Sets visual indicator: GREEN (connected), YELLOW (degraded), RED (disconnected).
 *
 * @returns { status: 'connected' | 'degraded' | 'disconnected', lastCheck: Date }
 */
export function useConnectionStatus() {
  // ... implementation
}
```

### 5.4 useOfflineMode

```typescript
// CLASSIFICATION: UNCLASSIFIED
// src/hooks/useOfflineMode.ts

/**
 * useOfflineMode — manages offline operation when backend is unreachable.
 *
 * When offline:
 *   1. Uses Service Worker to cache last known track state
 *   2. Displays stale tracks with "OFFLINE" indicator
 *   3. Queues operator feedback for later submission
 *   4. On reconnect: replays queued feedback, refreshes tracks
 *
 * SECURITY: Cached data is cleared when classification level changes.
 *
 * @returns { isOffline, queuedFeedbackCount, syncStatus }
 */
export function useOfflineMode() {
  // ... implementation
}
```

---

## 6. Key Components

### 6.1 ClassificationBanner

```tsx
// CLASSIFICATION: UNCLASSIFIED
// src/components/layout/ClassificationBanner.tsx

/**
 * ClassificationBanner — displayed at TOP and BOTTOM of viewport at all times.
 * Shows the highest classification level of data currently displayed.
 *
 * Colors follow Government of Canada classification marking:
 *   UNCLASSIFIED     → green (#008000) with white text
 *   PROTECTED A      → blue (#0000FF) with white text
 *   PROTECTED B      → blue (#0000FF) with white text
 *   PROTECTED C      → red (#FF0000) with white text
 *   SECRET           → red (#FF0000) with white text
 *
 * The text format is: "CLASSIFICATION: <LEVEL>"
 * Banner MUST be visible at all times — never hidden by scroll or overlay.
 */
export const ClassificationBanner: React.FC<{
  level: ClassificationLevel;
  position: "top" | "bottom";
}> = ({ level, position }) => {
  // ... implementation
};
```

### 6.2 MapView

```tsx
// CLASSIFICATION: UNCLASSIFIED
// src/components/map/MapView.tsx

/**
 * MapView — main map display using MapLibre GL JS.
 *
 * Features:
 *   - Renders tracks as positioned markers with MIL-STD-2525 symbology
 *   - Color-coded by hostile classification (Red=HOSTILE, Blue=FRIENDLY, Green=NEUTRAL, Yellow=UNKNOWN)
 *   - Track history trail (last 10 positions as fading dots)
 *   - Click-to-select: clicking a track opens DetailPanel
 *   - Default view: Mid-Atlantic region (center: -60°, 45°, zoom: 6)
 *   - No external tile providers in production (offline map tiles)
 *
 * Layers (bottom to top):
 *   1. Base map tiles (offline raster or vector)
 *   2. GeoFenceLayer — exclusion zone polygons
 *   3. SensorCoverageLayer — sensor coverage arcs
 *   4. ThreatHaloLayer — proximity circles around hostile tracks
 *   5. TrackLayer — track markers with labels
 */
export const MapView: React.FC = () => {
  // ... implementation using maplibre-gl
};
```

### 6.3 AlertPanel

```tsx
// CLASSIFICATION: UNCLASSIFIED
// src/components/alerts/AlertPanel.tsx

/**
 * AlertPanel — displays the prioritized alert queue in the right panel.
 *
 * Layout:
 *   - Header: "ALERTS" + unacknowledged count badge
 *   - AlertFilter: severity toggle buttons (WATCH, ELEVATED, CRITICAL)
 *   - Scrollable list of AlertCard components
 *   - Each card shows: severity icon, anomaly type, track ID, time, confidence
 *   - CRITICAL alerts pulse with red border animation
 *   - Click to acknowledge (opens confirmation dialog)
 *   - Click track ID to center map and open DetailPanel
 */
export const AlertPanel: React.FC = () => {
  // ... implementation
};
```

### 6.4 DetailPanel

```tsx
// CLASSIFICATION: UNCLASSIFIED
// src/components/detail/DetailPanel.tsx

/**
 * DetailPanel — shows full details for the selected track entity.
 *
 * Sections:
 *   1. IdentitySection: track_id, entity_type, hostile_class, confidence, status
 *   2. PositionSection: lat/lon (DMS format), altitude, speed, heading, last update
 *   3. SourceAttribution: list of contributing sensors with confidence per source
 *   4. EntityTimeline: chronological history of track updates and alerts
 *   5. FeedbackForm: operator can submit feedback (confirm/reclassify/reject)
 *
 * Classification: shows track classification level, greys out if above operator clearance
 */
export const DetailPanel: React.FC = () => {
  // ... implementation
};
```

### 6.5 FeedbackForm

```tsx
// CLASSIFICATION: UNCLASSIFIED
// src/components/detail/FeedbackForm.tsx

/**
 * FeedbackForm — allows operator to submit feedback on a track or alert.
 *
 * Fields:
 *   - feedback_type: dropdown (CONFIRM_HOSTILE, CONFIRM_FRIENDLY, RECLASSIFY, REJECT_ANOMALY, CONFIRM_ANOMALY)
 *   - justification: textarea (required, min 10 chars)
 *   - Submit button
 *
 * On submit:
 *   1. Calls FeedbackService.SubmitFeedback via gRPC-Web
 *   2. Shows loading spinner
 *   3. On success: shows trust score returned, closes form
 *   4. On error: shows error message, keeps form open
 *   5. If offline: queues feedback for later submission
 *
 * Audit: feedback submission is audited server-side (Module 09)
 */
export const FeedbackForm: React.FC<{
  trackId: string;
  alertId?: string;
}> = ({ trackId, alertId }) => {
  // ... implementation
};
```

### 6.6 ForensicsPanel

```tsx
// CLASSIFICATION: UNCLASSIFIED
// src/components/forensics/ForensicsPanel.tsx

/**
 * ForensicsPanel — historical analysis and query interface.
 *
 * Sub-components:
 *   - QueryBuilder: form to build historical queries
 *       * Time range picker (max 30 days)
 *       * Entity type multi-select
 *       * Anomaly type multi-select
 *       * Severity filter
 *       * Bounding box draw-on-map
 *   - ResultsView: paginated table of query results
 *       * Sortable columns
 *       * Click row to show on map
 *       * Export to CSV (classification-marked)
 *   - MapReplay: animate historical track positions on map
 *       * Play/pause/speed controls
 *       * Time scrubber
 *
 * Uses QueryService gRPC-Web calls (unary RPCs, not streaming)
 */
export const ForensicsPanel: React.FC = () => {
  // ... implementation
};
```

---

## 7. gRPC-Web Client Setup

```typescript
// CLASSIFICATION: UNCLASSIFIED
// src/api/grpc-client.ts

import { GrpcWebFetchTransport } from "@protobuf-ts/grpc-web-transport";

/**
 * Creates the shared gRPC-Web transport.
 * All service clients use this transport.
 *
 * Configuration:
 *   - Base URL: configured via VITE_GRPC_WEB_URL env var (default: https://localhost:8443)
 *   - Format: binary (more efficient than text)
 *   - Metadata: classification header attached to all requests
 */
export function createTransport() {
  return new GrpcWebFetchTransport({
    baseUrl: import.meta.env.VITE_GRPC_WEB_URL || "https://localhost:8443",
    format: "binary",
  });
}

// Singleton transport instance
export const transport = createTransport();
```

---

## 8. TypeScript Types

```typescript
// CLASSIFICATION: UNCLASSIFIED
// src/types/common.ts

export type ClassificationLevel =
  | "UNCLASSIFIED"
  | "PROTECTED_A"
  | "PROTECTED_B"
  | "PROTECTED_C"
  | "SECRET";

export type EntityType = "SURFACE" | "AIR" | "SUBSURFACE" | "LAND" | "CYBER";

export type HostileClassification =
  | "HOSTILE"
  | "FRIENDLY"
  | "NEUTRAL"
  | "UNKNOWN";

export type TrackStatus = "ACTIVE" | "STALE" | "DROPPED" | "MERGED";

export type AlertSeverity = "NORMAL" | "WATCH" | "ELEVATED" | "CRITICAL";

export type AnomalyType =
  | "SPEED"
  | "ROUTE_DEVIATION"
  | "AIS_MANIPULATION"
  | "BEHAVIORAL"
  | "TEMPORAL"
  | "PROXIMITY";

export type FeedbackType =
  | "CONFIRM_HOSTILE"
  | "CONFIRM_FRIENDLY"
  | "RECLASSIFY"
  | "REJECT_ANOMALY"
  | "CONFIRM_ANOMALY";

// src/types/track.ts
export interface FusedTrack {
  trackId: string;
  entityType: EntityType;
  hostileClass: HostileClassification;
  position: Position;
  confidenceScore: number;
  sourceCount: number;
  sources: SourceAttribution[];
  status: TrackStatus;
  classification: ClassificationLevel;
  createdAt: Date;
  updatedAt: Date;
}

export interface Position {
  latitude: number;
  longitude: number;
  altitudeMeters?: number;
  speedKnots?: number;
  headingDegrees?: number;
}

export interface SourceAttribution {
  sensorId: string;
  sensorType: string;
  confidence: number;
  lastContribution: Date;
}

// src/types/alert.ts
export interface AnomalyAlert {
  alertId: string;
  trackId: string;
  anomalyType: AnomalyType;
  severity: AlertSeverity;
  confidenceScore: number;
  explanation: string;
  features: FeatureContribution[];
  classification: ClassificationLevel;
  detectedAt: Date;
}

export interface FeatureContribution {
  featureName: string;
  value: number;
  contributionWeight: number;
}
```

---

## 9. Visual Design Specifications

### 9.1 Color Scheme

| Element          | Color       | Hex                     |
| ---------------- | ----------- | ----------------------- |
| HOSTILE track    | Red         | `#DC2626`               |
| FRIENDLY track   | Blue        | `#2563EB`               |
| NEUTRAL track    | Green       | `#16A34A`               |
| UNKNOWN track    | Yellow      | `#CA8A04`               |
| CRITICAL alert   | Pulsing Red | `#DC2626`               |
| ELEVATED alert   | Orange      | `#EA580C`               |
| WATCH alert      | Yellow      | `#CA8A04`               |
| Stale track      | Grey        | `#6B7280` (50% opacity) |
| Map background   | Dark        | `#1E293B`               |
| Panel background | Dark grey   | `#0F172A`               |

### 9.2 Layout Grid

```
┌──────────────────────────────────────────────────────────┐
│ CLASSIFICATION: PROTECTED B                               │ ← Banner
├──────────────────────────────────────────┬─────────────────┤
│                                          │  ALERTS (12)    │
│                                          │  ┌─────────┐   │
│           MAP VIEW                       │  │ CRITICAL │   │
│           (70% width)                    │  │ speed... │   │
│                                          │  ├─────────┤   │
│                                          │  │ ELEVATED │   │
│                                          │  │ route... │   │
│                                          │  └─────────┘   │
│                                          │  (30% width)   │
├──────────────────────────────────────────┴─────────────────┤
│  DETAIL PANEL (collapsible, bottom or side)               │
│  ┌──────────┬──────────┬──────────┬──────────┬──────────┐ │
│  │ Identity │ Position │ Sources  │ Timeline │ Feedback │ │
│  └──────────┴──────────┴──────────┴──────────┴──────────┘ │
├──────────────────────────────────────────────────────────┤
│ CLASSIFICATION: PROTECTED B                               │ ← Banner
└──────────────────────────────────────────────────────────┘
```

---

## 10. Build Configuration

### 10.1 package.json Key Dependencies

```json
{
  "name": "@rtsa/web-cop",
  "version": "0.1.0",
  "private": true,
  "scripts": {
    "dev": "vite",
    "build": "tsc && vite build",
    "preview": "vite preview",
    "test": "vitest run",
    "test:watch": "vitest",
    "test:coverage": "vitest run --coverage",
    "lint": "eslint src/ --ext .ts,.tsx",
    "proto:gen": "buf generate --template buf.gen.yaml"
  },
  "dependencies": {
    "react": "^18.3.0",
    "react-dom": "^18.3.0",
    "zustand": "^4.5.0",
    "maplibre-gl": "^4.0.0",
    "@protobuf-ts/runtime": "^2.9.0",
    "@protobuf-ts/grpc-web-transport": "^2.9.0",
    "@tanstack/react-query": "^5.0.0",
    "tailwindcss": "^3.4.0"
  },
  "devDependencies": {
    "@types/react": "^18.3.0",
    "@types/react-dom": "^18.3.0",
    "@testing-library/react": "^14.0.0",
    "@testing-library/jest-dom": "^6.0.0",
    "vitest": "^1.6.0",
    "@vitest/coverage-v8": "^1.6.0",
    "typescript": "^5.4.0",
    "vite": "^5.2.0",
    "@vitejs/plugin-react": "^4.2.0",
    "eslint": "^8.57.0"
  }
}
```

### 10.2 Environment Variables

| Variable                      | Default                  | Description                   |
| ----------------------------- | ------------------------ | ----------------------------- |
| `VITE_GRPC_WEB_URL`           | `https://localhost:8443` | Envoy gateway URL             |
| `VITE_MAP_TILE_URL`           | (offline tiles)          | Map tile server URL           |
| `VITE_APP_TITLE`              | `RTSA COP`               | Application title             |
| `VITE_CLASSIFICATION_CEILING` | `PROTECTED_B`            | Default classification banner |

---

## 11. Test Scenarios

| #   | Test                                       | Expected                    |
| --- | ------------------------------------------ | --------------------------- |
| T01 | TrackStore.upsertTrack                     | Track added to map          |
| T02 | TrackStore.getHostileTracks                | Filters correctly           |
| T03 | AlertStore.addAlert + getFilteredAlerts    | Severity filtering works    |
| T04 | AlertStore.acknowledgeAlert                | Removed from unacknowledged |
| T05 | AuthStore.canAccess: clearance >= data     | Returns true                |
| T06 | AuthStore.canAccess: clearance < data      | Returns false               |
| T07 | UIStore.toggleAlertPanel                   | State toggles               |
| T08 | ClassificationBanner renders correct color | Color matches level         |
| T09 | AlertCard renders CRITICAL with pulse      | CSS animation present       |
| T10 | FeedbackForm validates justification       | Error on <10 chars          |
| T11 | FeedbackForm submit success                | Shows trust score, closes   |
| T12 | DetailPanel shows selected track           | All sections rendered       |
| T13 | useTrackStream reconnects on error         | Backoff + reconnect         |
| T14 | useAlertStream filters by severity         | Store updated correctly     |

---

## 12. Agent Invocation

```
@greatest-ever-developer Implement Module 15 from docs/implementation/15-cop-web-app.md

Context:
- Read docs/implementation/00-implementation-overview.md for global conventions
- Read docs/implementation/02-protobuf-schemas.md for all proto definitions
- Read docs/implementation/14-api-gateway.md for Envoy gRPC-Web routing
- Read docs/architecture/component_design.md §9 for COP component hierarchy
- Read docs/architecture/security_architecture.md §7 for classification enforcement

Deliverables:
1. Complete web-cop/ project with all files
2. Vite + React 18 + TypeScript 5 setup
3. 4 Zustand stores (TrackStore, AlertStore, AuthStore, UIStore)
4. gRPC-Web streaming hooks (useTrackStream, useAlertStream)
5. MapView with MapLibre GL JS
6. AlertPanel with severity filtering
7. DetailPanel with FeedbackForm
8. ForensicsPanel with QueryBuilder
9. ClassificationBanner (top + bottom, all views)
10. Unit tests with Vitest (≥80% coverage)

CRITICAL:
- Classification banner MUST be visible at ALL times
- Map default center: Mid-Atlantic (-60°, 45°)
- No external tile providers in production (offline capability)
- CRITICAL alerts must have red pulse animation
- All data access must respect operator clearance level
- No PII in browser console logs
```
