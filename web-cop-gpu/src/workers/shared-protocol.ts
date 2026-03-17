// CLASSIFICATION: UNCLASSIFIED
// src/workers/shared-protocol.ts — Typed Worker ↔ Main Thread message protocol
//
// All postMessage payloads between the Render Worker / Data Worker and the
// main SolidJS thread use these discriminated union types.
//
// Reference: docs/sdlc_guidelines/04_coding_standards/solidjs_standards.md §5.1

// ── Render Worker → Main Thread ────────────────────────────────────────────────

/** Stats emitted once per second by the Render Worker. */
export interface RenderStatsMessage {
  type: "stats";
  fps: number;
  trackCount: number;
  visibleCount: number;
}

/** Emitted when the pick buffer resolves after a canvas click. */
export interface PickedMessage {
  type: "picked";
  trackIdHash: number;
  x: number;
  y: number;
}

/** Emitted once the Render Worker has successfully initialised. */
export interface RenderReadyMessage {
  type: "status";
  ready: boolean;
  error?: string;
}

export type RenderToMainMessage =
  | RenderStatsMessage
  | PickedMessage
  | RenderReadyMessage;

// ── Main Thread → Render Worker ────────────────────────────────────────────────

export interface RenderInitMessage {
  type: "init";
  canvas: OffscreenCanvas;
  sab: SharedArrayBuffer;
  initialWidth: number;
  initialHeight: number;
  /**
   * When true, the Data Worker is the sole SAB writer for mock/live data.
   * The Render Worker must NOT write mock data to the SAB when this flag is set.
   */
  dataWorkerActive?: boolean;
}

export interface RenderResizeMessage {
  type: "resize";
  width: number;
  height: number;
}

export interface SelectTrackMessage {
  type: "select_track";
  x: number;
  y: number;
}

export interface SetDashboardMessage {
  type: "set_dashboard";
  dashboard: "sensor" | "commander" | "analytics" | "health" | "coverage";
}

export interface SetCoverageMessage {
  type: "set_coverage";
  records: {
    centerLon:    number;
    centerLat:    number;
    rangeNm:      number;
    bearingStart: number;
    bearingEnd:   number;
    recordType:   number; // 0 = Sector, 1 = Gap Polygon
    alertLevel:   number; // 0 = Normal, 1 = Warning, 2 = Critical
  }[];
}

export interface ObservationRecord {
  id: string;
  lat: number;
  lon: number;
  type: number; // SensorType enum
  confidence: number;
}

export interface SetObservationsMessage {
  type: "set_observations";
  observations: ObservationRecord[];
}

export interface SetViewportMessage {
  type:      "set_viewport";
  centerLat: number;
  centerLon: number;
  zoom:      number;
}

export interface SetTrackSelectionMessage {
  type: "set_track_selection";
  trackIdHash: number;
}

export interface SetMapStyleMessage {
  type: "set_map_style";
  mapStyle: number;
}

export type MainToRenderMessage =
  | RenderInitMessage
  | RenderResizeMessage
  | SelectTrackMessage
  | SetDashboardMessage
  | SetCoverageMessage
  | SetViewportMessage
  | SetObservationsMessage
  | SetTrackSelectionMessage
  | SetMapStyleMessage;

// ── Data Worker → Main Thread ──────────────────────────────────────────────────

/** Alert payload carried from the Data Worker to the main thread. */
export interface AlertPayload {
  alertId: string;
  trackId: string;
  severity: "UNSPECIFIED" | "NORMAL" | "WATCH" | "ELEVATED" | "CRITICAL";
  description: string;
  detectedAtMs: number;
  acknowledged: boolean;
}

/** Pushed whenever the alert list changes. */
export interface AlertsUpdatedMessage {
  type: "alerts_updated";
  alerts: AlertPayload[];
}

/** Connection status / latency update from the Data Worker. */
export interface DataConnectionStatusMessage {
  type: "connection_status";
  connected: boolean;
  /** RTT estimate in ms; -1 when disconnected. */
  latency: number;
}

/** Decode statistics emitted once per second. */
export interface DataStatsMessage {
  type: "stats";
  datagramsReceived: number;
  recordsDecoded: number;
  decodeErrors: number;
}

export type DataToMainMessage =
  | AlertsUpdatedMessage
  | DataConnectionStatusMessage
  | DataStatsMessage
  | TokenExpiringMessage;

// ── Main Thread → Data Worker ──────────────────────────────────────────────────

export interface DataInitMessage {
  type: "init";
  sab: SharedArrayBuffer;
  /** WebTransport server URL. Undefined in dev → worker falls back to mock mode. */
  url?: string;
  /**
   * Short-lived JWT for WebTransport authentication.
   * Appended as `?token=<jwt>` to the WebTransport URL.
   * NEVER log this value. (SDLC Rule 5)
   */
  token?: string;
}

/**
 * Sent by the main thread when a refreshed JWT is available.
 * The Data Worker reconnects to WebTransport with the new token.
 * NEVER log the token value. (SDLC Rule 5)
 */
export interface TokenRefreshMessage {
  type: "token-refresh";
  token: string;
}

export type MainToDataMessage = DataInitMessage | TokenRefreshMessage;

// ── Data Worker → Main Thread (additional) ────────────────────────────────────

/**
 * Sent by the Data Worker when the current JWT is approaching expiry
 * (60 seconds before the token expires). The main thread should fetch a
 * new token and send a TokenRefreshMessage back.
 */
export interface TokenExpiringMessage {
  type: "token-expiring";
}
