// CLASSIFICATION: UNCLASSIFIED
// src/components/layout/ConnectionIndicator.tsx

import React from "react";
import { useConnectionStatus } from "../../hooks/useConnectionStatus";

const STATUS_COLORS = {
  connected: "#16A34A",
  degraded: "#CA8A04",
  disconnected: "#DC2626",
};

/**
 * ConnectionIndicator — shows the current backend connectivity status.
 * GREEN = connected, YELLOW = degraded, RED = disconnected.
 */
export const ConnectionIndicator: React.FC = () => {
  const { status, lastCheck } = useConnectionStatus();

  return (
    <div
      data-testid="connection-indicator"
      style={{ display: "flex", alignItems: "center", gap: "6px" }}
      title={`Status: ${status}${lastCheck ? ` (checked: ${lastCheck.toISOString()})` : ""}`}
    >
      <div
        style={{
          width: "10px",
          height: "10px",
          borderRadius: "50%",
          backgroundColor: STATUS_COLORS[status],
        }}
      />
      <span style={{ fontSize: "0.75rem", color: "#9CA3AF" }}>
        {status.toUpperCase()}
      </span>
    </div>
  );
};
