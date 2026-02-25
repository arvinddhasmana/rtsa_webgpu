// CLASSIFICATION: UNCLASSIFIED
// src/components/alerts/AlertPanel.tsx

import React from "react";
import { useAlertStore } from "../../stores/alertStore";
import { AlertFilter } from "./AlertFilter";
import { AlertCard } from "./AlertCard";

/**
 * AlertPanel — displays the prioritized alert queue in the right panel.
 *
 * Layout:
 *   - Header: "ALERTS" + unacknowledged count badge
 *   - AlertFilter: severity toggle buttons
 *   - Scrollable list of AlertCard components
 *   - CRITICAL alerts pulse with red border animation
 */
export const AlertPanel: React.FC = () => {
  const getFilteredAlerts = useAlertStore((s) => s.getFilteredAlerts);
  const getCriticalCount = useAlertStore((s) => s.getCriticalCount);
  const getUnacknowledgedAlerts = useAlertStore((s) => s.getUnacknowledgedAlerts);

  const filteredAlerts = getFilteredAlerts();
  const criticalCount = getCriticalCount();
  const unacknowledgedCount = getUnacknowledgedAlerts().length;

  return (
    <div
      data-testid="alert-panel"
      style={{
        display: "flex",
        flexDirection: "column",
        height: "100%",
        backgroundColor: "#1E293B",
      }}
    >
      {/* Header */}
      <div
        style={{
          display: "flex",
          alignItems: "center",
          padding: "8px 12px",
          borderBottom: "1px solid #334155",
          gap: "8px",
        }}
      >
        <span style={{ fontWeight: "bold", fontSize: "0.875rem", letterSpacing: "0.1em" }}>
          ALERTS
        </span>
        {unacknowledgedCount > 0 && (
          <span
            data-testid="unacknowledged-count"
            style={{
              backgroundColor: criticalCount > 0 ? "#DC2626" : "#EA580C",
              color: "white",
              borderRadius: "9999px",
              padding: "1px 6px",
              fontSize: "0.7rem",
              fontWeight: "bold",
            }}
          >
            {unacknowledgedCount}
          </span>
        )}
      </div>

      {/* Filters */}
      <AlertFilter />

      {/* Alert list */}
      <div style={{ flex: 1, overflowY: "auto" }}>
        {filteredAlerts.length === 0 ? (
          <div
            style={{
              padding: "16px",
              textAlign: "center",
              color: "#9CA3AF",
              fontSize: "0.8rem",
            }}
          >
            No alerts
          </div>
        ) : (
          filteredAlerts.map((alert) => (
            <AlertCard key={alert.alertId} alert={alert} />
          ))
        )}
      </div>
    </div>
  );
};
