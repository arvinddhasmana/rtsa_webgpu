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

export type MainToRenderMessage =
  | RenderInitMessage
  | RenderResizeMessage
  | SelectTrackMessage;

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
  | DataStatsMessage;

// ── Main Thread → Data Worker ──────────────────────────────────────────────────

export interface DataInitMessage {
  type: "init";
  sab: SharedArrayBuffer;
  url?: string;
}

export type MainToDataMessage = DataInitMessage;
