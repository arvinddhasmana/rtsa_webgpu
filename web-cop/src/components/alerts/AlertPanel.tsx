// CLASSIFICATION: UNCLASSIFIED
// src/components/alerts/AlertPanel.tsx
// Operations Commander — Enhanced Alert Panel with History Tab

import React, { useState } from "react";
import { severityRank, useAlertStore } from "../../stores/alertStore";
import { AlertCard } from "./AlertCard";
import { AlertFilter } from "./AlertFilter";

const MAX_RENDERED_ALERTS = 120;

/**
 * AlertPanel — displays the prioritized alert queue and history in the right panel.
 *
 * Layout:
 *   - Tabs: QUEUE | HISTORY
 *   - Header: Unacknowledged count badge
 *   - AlertFilter: severity + new anomaly/entity toggles
 *   - Scrollable list of AlertCard components
 */
export const AlertPanel: React.FC = () => {
  const alerts = useAlertStore((s) => s.alerts);
  const acknowledgedIds = useAlertStore((s) => s.acknowledgedIds);
  const minSeverityFilter = useAlertStore((s) => s.minSeverityFilter);

  const [activeTab, setActiveTab] = useState<"QUEUE" | "HISTORY">("QUEUE");

  // Local state for the new filters (could be moved to store if needed globally)
  const [anomalyTypeFilter, setAnomalyTypeFilter] = useState<string>("ALL");
  const [timeRangeFilter, setTimeRangeFilter] = useState<number>(24 * 60 * 60 * 1000); // 24h default

  const { filteredAlerts, criticalCount, unacknowledgedCount, historyCount } =
    React.useMemo(() => {
      const allAlerts = Array.from(alerts.values());
      const minRank = severityRank(minSeverityFilter);
      const now = Date.now();

      const unacknowledgedCountLocal = allAlerts.filter(
        (alert) => !acknowledgedIds.has(alert.alertId),
      ).length;

      const criticalCountLocal = allAlerts.filter(
        (alert) =>
          alert.severity === "CRITICAL" && !acknowledgedIds.has(alert.alertId),
      ).length;

      const historyCountLocal = allAlerts.filter(
        (alert) => acknowledgedIds.has(alert.alertId)
      ).length;

      let filtered = allAlerts.filter((alert) => {
        // Tab filter
        if (activeTab === "QUEUE" && acknowledgedIds.has(alert.alertId)) return false;
        if (activeTab === "HISTORY" && !acknowledgedIds.has(alert.alertId)) return false;

        // Severity filter
        if (severityRank(alert.severity) < minRank) return false;

        // Anomaly type filter
        if (anomalyTypeFilter !== "ALL" && !alert.anomalyType.includes(anomalyTypeFilter)) return false;

        // Time range filter
        if (now - alert.detectedAt.getTime() > timeRangeFilter) return false;

        return true;
      });

      // Sort
      filtered.sort((a, b) => {
        if (activeTab === "QUEUE") {
          // Priority sort for queue
          const bySeverity = severityRank(b.severity) - severityRank(a.severity);
          if (bySeverity !== 0) return bySeverity;
        }
        // Chronological sort (newest first) for history and as tie-breaker for queue
        return b.detectedAt.getTime() - a.detectedAt.getTime();
      });

      filtered = filtered.slice(0, MAX_RENDERED_ALERTS);

      return {
        filteredAlerts: filtered,
        criticalCount: criticalCountLocal,
        unacknowledgedCount: unacknowledgedCountLocal,
        historyCount: historyCountLocal,
      };
    }, [alerts, acknowledgedIds, minSeverityFilter, activeTab, anomalyTypeFilter, timeRangeFilter]);

  return (
    <div
      data-testid="alert-panel"
      style={{
        display: "flex",
        flexDirection: "column",
        height: "100%",
        backgroundColor: "transparent", // relying on parent glassmorphism
      }}
    >
      {/* ── Tabs ── */}
      <div
        style={{
          display: "flex",
          borderBottom: "1px solid rgba(255,255,255,0.1)",
        }}
      >
        <button
          data-testid="tab-queue"
          onClick={() => setActiveTab("QUEUE")}
          style={{
            flex: 1,
            padding: "12px 0",
            background: "transparent",
            border: "none",
            borderBottom: activeTab === "QUEUE" ? "2px solid #60A5FA" : "2px solid transparent",
            color: activeTab === "QUEUE" ? "#60A5FA" : "#94A3B8",
            fontSize: "0.75rem",
            fontWeight: "bold",
            cursor: "pointer",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            gap: "8px",
          }}
        >
          TRIAGE QUEUE
          {unacknowledgedCount > 0 && (
            <span
              data-testid="unacknowledged-count"
              style={{
                backgroundColor: criticalCount > 0 ? "#DC2626" : "#EA580C",
                color: "white",
                borderRadius: "9999px",
                padding: "2px 8px",
                fontSize: "0.65rem",
              }}
            >
              {unacknowledgedCount}
            </span>
          )}
        </button>
        <button
          data-testid="tab-history"
          onClick={() => setActiveTab("HISTORY")}
          style={{
            flex: 1,
            padding: "12px 0",
            background: "transparent",
            border: "none",
            borderBottom: activeTab === "HISTORY" ? "2px solid #60A5FA" : "2px solid transparent",
            color: activeTab === "HISTORY" ? "#60A5FA" : "#94A3B8",
            fontSize: "0.75rem",
            fontWeight: "bold",
            cursor: "pointer",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            gap: "8px",
          }}
        >
          HISTORY
          {historyCount > 0 && (
             <span
             style={{
               backgroundColor: "rgba(255,255,255,0.1)",
               color: "#CBD5E1",
               borderRadius: "9999px",
               padding: "2px 8px",
               fontSize: "0.65rem",
             }}
           >
             {historyCount}
           </span>
          )}
        </button>
      </div>

      {/* ── Filters ── */}
      <div style={{ padding: "12px 12px 4px 12px", display: "flex", flexDirection: "column", gap: "8px" }}>
        <AlertFilter />

        {/* Advanced Filters Row */}
        <div style={{ display: "flex", gap: "8px" }}>
          <select
            value={anomalyTypeFilter}
            onChange={(e) => setAnomalyTypeFilter(e.target.value)}
            style={selectStyle}
          >
            <option value="ALL">All Anomalies</option>
            <option value="KINEMATIC">Kinematic</option>
            <option value="EMISSION">Emission</option>
            <option value="CYBER">Cyber</option>
          </select>

          <select
            value={timeRangeFilter}
            onChange={(e) => setTimeRangeFilter(Number(e.target.value))}
            style={selectStyle}
          >
            <option value={1 * 60 * 60 * 1000}>Last 1 Hour</option>
            <option value={6 * 60 * 60 * 1000}>Last 6 Hours</option>
            <option value={24 * 60 * 60 * 1000}>Last 24 Hours</option>
          </select>
        </div>
      </div>

      {/* ── Alert list ── */}
      <div style={{ flex: 1, overflowY: "auto", padding: "4px" }}>
        {filteredAlerts.length === 0 ? (
          <div
            style={{
              padding: "32px 16px",
              textAlign: "center",
              color: "#64748B",
              fontSize: "0.8rem",
            }}
          >
            {activeTab === "QUEUE" ? "Queue is clear. No unacknowledged alerts match filters." : "No alert history matches filters."}
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

const selectStyle: React.CSSProperties = {
  flex: 1,
  backgroundColor: "rgba(15, 23, 42, 0.5)",
  color: "#CBD5E1",
  border: "1px solid rgba(255,255,255,0.1)",
  borderRadius: "4px",
  padding: "4px 8px",
  fontSize: "0.7rem",
  outline: "none",
  cursor: "pointer",
};
