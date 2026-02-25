// CLASSIFICATION: UNCLASSIFIED
// src/hooks/useConnectionStatus.ts

import { useEffect, useState } from "react";

type ConnectionStatus = "connected" | "degraded" | "disconnected";

interface ConnectionState {
  status: ConnectionStatus;
  lastCheck: Date | null;
}

/**
 * useConnectionStatus — monitors connectivity to backend services.
 *
 * Uses gRPC Health Check protocol through Envoy.
 * Polls every 10 seconds.
 * Sets visual indicator: connected | degraded | disconnected.
 *
 * @returns { status, lastCheck }
 */
export function useConnectionStatus(): ConnectionState {
  const [state, setState] = useState<ConnectionState>({
    status: "disconnected",
    lastCheck: null,
  });

  useEffect(() => {
    let mounted = true;

    const check = async () => {
      const grpcWebUrl =
        (import.meta as { env?: Record<string, string> }).env?.["VITE_GRPC_WEB_URL"] ??
        "https://localhost:8443";

      try {
        const res = await fetch(`${grpcWebUrl}/healthz`, {
          method: "GET",
          signal: AbortSignal.timeout(5000),
        });
        if (!mounted) return;
        setState({
          status: res.ok ? "connected" : "degraded",
          lastCheck: new Date(),
        });
      } catch {
        if (!mounted) return;
        setState({ status: "disconnected", lastCheck: new Date() });
      }
    };

    void check();
    const interval = setInterval(() => void check(), 10_000);

    return () => {
      mounted = false;
      clearInterval(interval);
    };
  }, []);

  return state;
}
