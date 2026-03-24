// CLASSIFICATION: UNCLASSIFIED
// src/signals/stats.ts — Render and data throughput signals
//
// Updated from Render Worker ("stats") and Data Worker ("stats") messages.
// Reference: docs/sdlc_guidelines/04_coding_standards/solidjs_standards.md §3.3

import { createSignal } from "solid-js";

// ── Render Worker stats ────────────────────────────────────────────────────────

/** Frames per second reported by the Render Worker (0 until first frame). */
export const [fps, setFps] = createSignal(0);

/** Total track count in the SharedArrayBuffer. */
export const [trackCount, setTrackCount] = createSignal(0);

/** Tracks passing the view-frustum cull (rendered this frame). */
export const [visibleCount, setVisibleCount] = createSignal(0);

// ── Data Worker stats ──────────────────────────────────────────────────────────

/** Datagrams received from the WebTransport server in the last second. */
export const [datagramsPerSec, setDatagramsPerSec] = createSignal(0);

/** Records decoded in the last second. */
export const [recordsPerSec, setRecordsPerSec] = createSignal(0);

/** Decode errors in the last second. */
export const [decodeErrors, setDecodeErrors] = createSignal(0);

/** WebTransport round-trip latency in ms (-1 = disconnected). */
export const [latencyMs, setLatencyMs] = createSignal(-1);

// ── Fusion Metrics (O(N) aggregation from Rust Data Worker) ───────────────────

import { FusionStats } from "../workers/shared-protocol";

const initialFusionStats: FusionStats = {
  high_confidence_count: 0,
  mid_confidence_count: 0,
  low_confidence_count: 0,
  radar_count: 0,
  sigint_count: 0,
  satellite_count: 0,
  ew_count: 0,
  others_count: 0,
  avg_latency_ms: 0,
  max_latency_ms: 0,
};

/** Aggregated Fusion metrics reported by the Data Worker. */
export const [fusionStats, setFusionStats] = createSignal<FusionStats>(initialFusionStats);

/** History of average latency for sparkline visualization (last 20 samples). */
export const [latencyHistory, setLatencyHistory] = createSignal<number[]>([]);
