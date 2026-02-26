// CLASSIFICATION: UNCLASSIFIED
// src/api/grpc-client.ts

/**
 * Shared gRPC-Web transport for all RTSA service clients.
 * Uses @connectrpc/connect-web which handles proper binary protobuf framing,
 * gRPC-Web envelope encoding, and trailer decoding.
 *
 * URL resolution priority:
 *   1. VITE_GRPC_WEB_URL env var (if explicitly set and non-empty).
 *   2. window.location.origin — same-origin requests are proxied by Nginx
 *      to Envoy on the Docker internal network (no TLS/CORS issues).
 *   3. https://localhost:8443 — fallback for SSR / non-browser contexts.
 */
import { createGrpcWebTransport } from "@connectrpc/connect-web";

function resolveGrpcWebUrl(): string {
  const envUrl = (import.meta as { env?: Record<string, string> }).env?.[
    "VITE_GRPC_WEB_URL"
  ];
  if (envUrl) return envUrl;
  // Same-origin: Nginx proxies /rtsa.* to envoy:8443 inside Docker network.
  if (typeof window !== "undefined") return window.location.origin;
  return "https://localhost:8443";
}

export const grpcWebUrl = resolveGrpcWebUrl();

/**
 * Singleton transport — all service clients share this instance.
 */
export const transport = createGrpcWebTransport({
  baseUrl: grpcWebUrl,
});
