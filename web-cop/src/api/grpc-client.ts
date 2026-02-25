// CLASSIFICATION: UNCLASSIFIED
// src/api/grpc-client.ts

/**
 * Creates the shared gRPC-Web transport configuration.
 * All service clients use this base URL.
 *
 * Configuration:
 *   - Base URL: configured via VITE_GRPC_WEB_URL env var (default: https://localhost:8443)
 *   - Format: binary (more efficient than text)
 *   - Metadata: classification header attached to all requests
 */
export interface GrpcTransportOptions {
  baseUrl: string;
  format: "binary" | "text";
}

export function createTransport(): GrpcTransportOptions {
  return {
    baseUrl:
      (import.meta as { env?: Record<string, string> }).env?.[
        "VITE_GRPC_WEB_URL"
      ] ?? "https://localhost:8443",
    format: "binary",
  };
}

// Singleton transport instance
export const transport = createTransport();
