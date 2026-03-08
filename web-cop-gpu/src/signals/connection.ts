// CLASSIFICATION: UNCLASSIFIED
// src/signals/connection.ts — WebTransport and gRPC connection state signals
//
// Driven by messages from the Data Worker.
// Reference: docs/sdlc_guidelines/04_coding_standards/solidjs_standards.md §3.3

import { createSignal } from "solid-js";

/** WebTransport connection state driven by the Data Worker. */
export const [wtConnected, setWtConnected] = createSignal(false);

/** Whether the gRPC cold-path is reachable (set by service calls). */
export const [grpcConnected, setGrpcConnected] = createSignal(true);

/** True while an initial connection attempt is in progress. */
export const [connecting, setConnecting] = createSignal(true);

/**
 * Whether the gRPC alert stream is healthy.
 * Set to false when a non-AbortError occurs in startAlertStream; the UI can
 * observe this signal to show an "alerts stream error" indicator without
 * relying on console logging in production.
 */
export const [alertStreamHealthy, setAlertStreamHealthy] = createSignal(true);
