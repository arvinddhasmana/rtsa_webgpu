// CLASSIFICATION: UNCLASSIFIED
// src/components/forensics/ResultsView.tsx

import React, { useState } from "react";
import { FusedTrack } from "../../types/track";
import { AnomalyAlert } from "../../types/alert";
import { formatZuluTime } from "../../utils/time";
import { ClassificationLevel } from "../../types/common";

interface ResultsViewProps {
  tracks: FusedTrack[];
  alerts: AnomalyAlert[];
  totalCount: number;
  classificationCeiling: ClassificationLevel;
  onTrackSelect: (trackId: string) => void;
}

/**
 * ResultsView — paginated table of query results.
 * Sortable columns, click row to show on map.
 * Export to CSV (classification-marked).
 */
export const ResultsView: React.FC<ResultsViewProps> = ({
  tracks,
  alerts,
  totalCount,
  classificationCeiling,
  onTrackSelect,
}) => {
  const [sortBy, setSortBy] = useState<"time" | "severity" | "type">("time");

  const sortedAlerts = [...alerts].sort((a, b) => {
    if (sortBy === "time")
      return b.detectedAt.getTime() - a.detectedAt.getTime();
    if (sortBy === "severity")
      return a.severity.localeCompare(b.severity);
    return a.anomalyType.localeCompare(b.anomalyType);
  });

  const handleExportCsv = () => {
    const header = `// CLASSIFICATION: ${classificationCeiling}\n`;
    const csvHeader = "track_id,anomaly_type,severity,confidence,detected_at\n";
    const rows = sortedAlerts
      .map(
        (a) =>
          `${a.trackId},${a.anomalyType},${a.severity},${a.confidenceScore},${a.detectedAt.toISOString()}`
      )
      .join("\n");
    const blob = new Blob([header + csvHeader + rows], { type: "text/csv" });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = `rtsa-forensics-${Date.now()}.csv`;
    link.click();
    URL.revokeObjectURL(url);
  };

  return (
    <div data-testid="results-view" style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          padding: "4px 8px",
          fontSize: "0.7rem",
          color: "#9CA3AF",
        }}
      >
        <span>
          {totalCount} results | {tracks.length} tracks, {alerts.length} alerts
        </span>
        <button
          data-testid="export-csv"
          onClick={handleExportCsv}
          style={{
            padding: "2px 8px",
            backgroundColor: "#374151",
            color: "#F1F5F9",
            border: "1px solid #475569",
            borderRadius: "4px",
            cursor: "pointer",
            fontSize: "0.65rem",
          }}
        >
          Export CSV
        </button>
      </div>

      <div style={{ flex: 1, overflowY: "auto" }}>
        <table
          style={{
            width: "100%",
            borderCollapse: "collapse",
            fontSize: "0.7rem",
          }}
        >
          <thead>
            <tr style={{ backgroundColor: "#0F172A", color: "#9CA3AF" }}>
              <th
                style={{ padding: "4px 8px", textAlign: "left", cursor: "pointer" }}
                onClick={() => setSortBy("time")}
              >
                TIME {sortBy === "time" ? "↓" : ""}
              </th>
              <th style={{ padding: "4px 8px", textAlign: "left" }}>TRACK</th>
              <th
                style={{ padding: "4px 8px", textAlign: "left", cursor: "pointer" }}
                onClick={() => setSortBy("type")}
              >
                TYPE {sortBy === "type" ? "↓" : ""}
              </th>
              <th
                style={{ padding: "4px 8px", textAlign: "left", cursor: "pointer" }}
                onClick={() => setSortBy("severity")}
              >
                SEVERITY {sortBy === "severity" ? "↓" : ""}
              </th>
              <th style={{ padding: "4px 8px", textAlign: "left" }}>CONF</th>
            </tr>
          </thead>
          <tbody>
            {sortedAlerts.map((alert) => (
              <tr
                key={alert.alertId}
                data-testid={`result-row-${alert.alertId}`}
                onClick={() => onTrackSelect(alert.trackId)}
                style={{
                  cursor: "pointer",
                  borderBottom: "1px solid #1E293B",
                }}
              >
                <td style={{ padding: "4px 8px", color: "#6B7280", fontFamily: "monospace" }}>
                  {formatZuluTime(alert.detectedAt)}
                </td>
                <td style={{ padding: "4px 8px", color: "#60A5FA" }}>
                  {alert.trackId}
                </td>
                <td style={{ padding: "4px 8px" }}>
                  {alert.anomalyType.replace("_", " ")}
                </td>
                <td
                  style={{
                    padding: "4px 8px",
                    color:
                      alert.severity === "CRITICAL"
                        ? "#DC2626"
                        : alert.severity === "ELEVATED"
                        ? "#EA580C"
                        : "#CA8A04",
                    fontWeight: "bold",
                  }}
                >
                  {alert.severity}
                </td>
                <td style={{ padding: "4px 8px" }}>
                  {Math.round(alert.confidenceScore * 100)}%
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
};
