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
