// CLASSIFICATION: UNCLASSIFIED
// src/api/grpc-client.ts

/**
 * Shared gRPC-Web transport for all RTSA service clients.
 * Uses @connectrpc/connect-web which handles proper binary protobuf framing,
 * gRPC-Web envelope encoding, and trailer decoding.
 *
 * Base URL is configured via VITE_GRPC_WEB_URL (default: https://localhost:8443).
 */
import { createGrpcWebTransport } from "@connectrpc/connect-web";

export const grpcWebUrl =
  (import.meta as { env?: Record<string, string> }).env?.[
    "VITE_GRPC_WEB_URL"
  ] ?? "https://localhost:8443";

/**
 * Singleton transport — all service clients share this instance.
 */
export const transport = createGrpcWebTransport({
  baseUrl: grpcWebUrl,
});
