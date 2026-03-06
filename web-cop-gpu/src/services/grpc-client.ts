// CLASSIFICATION: UNCLASSIFIED
// src/services/grpc-client.ts — ConnectRPC transport for gRPC cold path
//
// All operator commands (feedback, alert ack, queries) use gRPC-Web via
// ConnectRPC routed through the Envoy proxy.
// Reference: docs/sdlc_guidelines/04_coding_standards/solidjs_standards.md §6

import { createConnectTransport } from "@connectrpc/connect-web";

/**
 * Shared ConnectRPC transport instance.
 * Base URL is injected via Vite environment variable at build time.
 * Fallback to localhost for development.
 */
export const transport = createConnectTransport({
  baseUrl: import.meta.env.VITE_API_GATEWAY_URL ?? "http://localhost:8080",
});
