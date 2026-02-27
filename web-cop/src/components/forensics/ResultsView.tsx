// CLASSIFICATION: UNCLASSIFIED
// src/components/forensics/ResultsView.tsx

import React, { useState } from "react";
import { AnomalyAlert } from "../../types/alert";
import { ClassificationLevel } from "../../types/common";
import { FusedTrack } from "../../types/track";
import { formatZuluTime } from "../../utils/time";

interface ResultsViewProps {
  tracks: FusedTrack[];
  alerts: AnomalyAlert[];
  totalCount: number;
  classificationCeiling: ClassificationLevel;
  onTrackSelect: (trackId: string) => void;
}

type ActiveView = "tracks" | "alerts";
type TrackSortBy = "time" | "entity" | "hostile" | "confidence";
type AlertSortBy = "time" | "severity" | "type";

const HOSTILE_COLOR: Record<string, string> = {
  HOSTILE: "#DC2626",
  FRIENDLY: "#2563EB",
  NEUTRAL: "#16A34A",
  UNKNOWN: "#CA8A04",
};

const SEVERITY_COLOR: Record<string, string> = {
  CRITICAL: "#DC2626",
  ELEVATED: "#EA580C",
  WATCH: "#CA8A04",
  NORMAL: "#6B7280",
};

/**
 * ResultsView — dual-view table of forensics query results.
 * TRACKS tab: all returned track snapshots (primary — always populated).
 * ALERTS tab: anomaly detections for the same time window.
 * Sortable columns, click row to select track.
 * Export to CSV (classification-marked, covers both tracks and alerts).
 */
export const ResultsView: React.FC<ResultsViewProps> = ({
  tracks,
  alerts,
  totalCount,
  classificationCeiling,
  onTrackSelect,
}) => {
  const [activeView, setActiveView] = useState<ActiveView>("tracks");
  const [trackSort, setTrackSort] = useState<TrackSortBy>("time");
  const [alertSort, setAlertSort] = useState<AlertSortBy>("time");

  const sortedTracks = [...tracks].sort((a, b) => {
    if (trackSort === "time")
      return b.updatedAt.getTime() - a.updatedAt.getTime();
    if (trackSort === "entity") return a.entityType.localeCompare(b.entityType);
    if (trackSort === "hostile")
      return a.hostileClass.localeCompare(b.hostileClass);
    return b.confidenceScore - a.confidenceScore;
  });

  const sortedAlerts = [...alerts].sort((a, b) => {
    if (alertSort === "time")
      return b.detectedAt.getTime() - a.detectedAt.getTime();
    if (alertSort === "severity") return b.severity.localeCompare(a.severity);
    return a.anomalyType.localeCompare(b.anomalyType);
  });

  const handleExportCsv = () => {
    const classHeader = `// CLASSIFICATION: ${classificationCeiling}\n`;

    // Tracks section
    const trackCsvHeader =
      "track_id,entity_type,hostile_class,latitude,longitude,speed_kts,heading_deg,confidence,status,timestamp\n";
    const trackRows = sortedTracks
      .map(
        (t) =>
          `${t.trackId},${t.entityType},${t.hostileClass},` +
          `${t.position.latitude.toFixed(6)},${t.position.longitude.toFixed(6)},` +
          `${t.position.speedKnots ?? ""},${t.position.headingDegrees ?? ""},` +
          `${(t.confidenceScore * 100).toFixed(1)},${t.status},${t.updatedAt.toISOString()}`,
      )
      .join("\n");

    // Alerts section
    const alertCsvHeader =
      "\n// Anomaly Alerts\nalert_id,track_id,anomaly_type,severity,confidence,detected_at\n";
    const alertRows = sortedAlerts
      .map(
        (a) =>
          `${a.alertId},${a.trackId},${a.anomalyType},${a.severity},` +
          `${(a.confidenceScore * 100).toFixed(1)},${a.detectedAt.toISOString()}`,
      )
      .join("\n");

    const content =
      classHeader + trackCsvHeader + trackRows + alertCsvHeader + alertRows;
    const blob = new Blob([content], { type: "text/csv" });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = `rtsa-forensics-${Date.now()}.csv`;
    link.click();
    URL.revokeObjectURL(url);
  };

  const th = (
    label: string,
    active: boolean,
    onClick: () => void,
  ): React.ReactNode => (
    <th
      style={{
        padding: "4px 8px",
        textAlign: "left",
        cursor: "pointer",
        whiteSpace: "nowrap",
        userSelect: "none",
      }}
      onClick={onClick}
    >
      {label} {active ? "↓" : ""}
    </th>
  );

  return (
    <div
      data-testid="results-view"
      style={{ height: "100%", display: "flex", flexDirection: "column" }}
    >
      {/* Summary bar */}
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          padding: "4px 8px",
          fontSize: "0.7rem",
          color: "#9CA3AF",
          flexShrink: 0,
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

      {/* Tab switcher */}
      <div
        style={{
          display: "flex",
          borderBottom: "1px solid #334155",
          flexShrink: 0,
        }}
      >
        {(["tracks", "alerts"] as ActiveView[]).map((v) => {
          const count = v === "tracks" ? tracks.length : alerts.length;
          const isActive = activeView === v;
          return (
            <button
              key={v}
              data-testid={`results-tab-${v}`}
              onClick={() => setActiveView(v)}
              style={{
                padding: "4px 12px",
                fontSize: "0.7rem",
                fontWeight: isActive ? "bold" : "normal",
                backgroundColor: isActive ? "#1E293B" : "transparent",
                color: isActive ? "#F1F5F9" : "#9CA3AF",
                border: "none",
                borderBottom: isActive
                  ? "2px solid #3B82F6"
                  : "2px solid transparent",
                cursor: "pointer",
                letterSpacing: "0.05em",
              }}
            >
              {v.toUpperCase()} ({count})
            </button>
          );
        })}
      </div>

      {/* Table content */}
      <div style={{ flex: 1, overflowY: "auto" }}>
        {activeView === "tracks" && (
          <table
            style={{
              width: "100%",
              borderCollapse: "collapse",
              fontSize: "0.7rem",
            }}
          >
            <thead>
              <tr
                style={{
                  backgroundColor: "#0F172A",
                  color: "#9CA3AF",
                  position: "sticky",
                  top: 0,
                }}
              >
                {th("TIME", trackSort === "time", () => setTrackSort("time"))}
                <th style={{ padding: "4px 8px", textAlign: "left" }}>
                  TRACK ID
                </th>
                {th("ENTITY", trackSort === "entity", () =>
                  setTrackSort("entity"),
                )}
                {th("CLASS", trackSort === "hostile", () =>
                  setTrackSort("hostile"),
                )}
                {th("CONF", trackSort === "confidence", () =>
                  setTrackSort("confidence"),
                )}
                <th style={{ padding: "4px 8px", textAlign: "left" }}>
                  STATUS
                </th>
              </tr>
            </thead>
            <tbody>
              {sortedTracks.length === 0 ? (
                <tr>
                  <td
                    colSpan={6}
                    style={{
                      padding: "16px",
                      textAlign: "center",
                      color: "#9CA3AF",
                    }}
                  >
                    No tracks in result
                  </td>
                </tr>
              ) : (
                sortedTracks.map((t) => (
                  <tr
                    key={t.trackId}
                    data-testid={`result-row-${t.trackId}`}
                    onClick={() => onTrackSelect(t.trackId)}
                    style={{
                      cursor: "pointer",
                      borderBottom: "1px solid #1E293B",
                    }}
                  >
                    <td
                      style={{
                        padding: "4px 8px",
                        color: "#6B7280",
                        fontFamily: "monospace",
                        whiteSpace: "nowrap",
                      }}
                    >
                      {formatZuluTime(t.updatedAt)}
                    </td>
                    <td
                      style={{
                        padding: "4px 8px",
                        color: "#60A5FA",
                        fontFamily: "monospace",
                      }}
                    >
                      {t.trackId}
                    </td>
                    <td style={{ padding: "4px 8px" }}>{t.entityType}</td>
                    <td
                      style={{
                        padding: "4px 8px",
                        color: HOSTILE_COLOR[t.hostileClass] ?? "#9CA3AF",
                        fontWeight: "bold",
                      }}
                    >
                      {t.hostileClass}
                    </td>
                    <td style={{ padding: "4px 8px" }}>
                      {Math.round(t.confidenceScore * 100)}%
                    </td>
                    <td style={{ padding: "4px 8px", color: "#6B7280" }}>
                      {t.status}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        )}

        {activeView === "alerts" && (
          <table
            style={{
              width: "100%",
              borderCollapse: "collapse",
              fontSize: "0.7rem",
            }}
          >
            <thead>
              <tr
                style={{
                  backgroundColor: "#0F172A",
                  color: "#9CA3AF",
                  position: "sticky",
                  top: 0,
                }}
              >
                {th("TIME", alertSort === "time", () => setAlertSort("time"))}
                <th style={{ padding: "4px 8px", textAlign: "left" }}>TRACK</th>
                {th("ANOMALY", alertSort === "type", () =>
                  setAlertSort("type"),
                )}
                {th("SEVERITY", alertSort === "severity", () =>
                  setAlertSort("severity"),
                )}
                <th style={{ padding: "4px 8px", textAlign: "left" }}>CONF</th>
              </tr>
            </thead>
            <tbody>
              {sortedAlerts.length === 0 ? (
                <tr>
                  <td
                    colSpan={5}
                    style={{
                      padding: "16px",
                      textAlign: "center",
                      color: "#9CA3AF",
                    }}
                  >
                    No anomaly alerts in this time window
                  </td>
                </tr>
              ) : (
                sortedAlerts.map((alert) => (
                  <tr
                    key={alert.alertId}
                    data-testid={`result-row-${alert.alertId}`}
                    onClick={() => onTrackSelect(alert.trackId)}
                    style={{
                      cursor: "pointer",
                      borderBottom: "1px solid #1E293B",
                    }}
                  >
                    <td
                      style={{
                        padding: "4px 8px",
                        color: "#6B7280",
                        fontFamily: "monospace",
                        whiteSpace: "nowrap",
                      }}
                    >
                      {formatZuluTime(alert.detectedAt)}
                    </td>
                    <td style={{ padding: "4px 8px", color: "#60A5FA" }}>
                      {alert.trackId}
                    </td>
                    <td style={{ padding: "4px 8px" }}>
                      {alert.anomalyType.replace(/_/g, " ")}
                    </td>
                    <td
                      style={{
                        padding: "4px 8px",
                        color: SEVERITY_COLOR[alert.severity] ?? "#6B7280",
                        fontWeight: "bold",
                      }}
                    >
                      {alert.severity}
                    </td>
                    <td style={{ padding: "4px 8px" }}>
                      {Math.round(alert.confidenceScore * 100)}%
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
};
