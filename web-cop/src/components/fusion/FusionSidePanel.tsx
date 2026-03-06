// CLASSIFICATION: UNCLASSIFIED
// src/components/fusion/FusionSidePanel.tsx
// Fusion Engine telemetry panel — Variant C tabbed layout with sortable track table.

import React, { useMemo, useState } from "react";
import { useAlertStore } from "../../stores/alertStore";
import { useSensorStore } from "../../stores/sensorStore";
import { useTrackStore } from "../../stores/trackStore";
import { useUIStore } from "../../stores/uiStore";
import type { AnomalyAlert } from "../../types/alert";
import type { FusedTrack } from "../../types/track";
import { ConfidenceHistogram } from "./ConfidenceHistogram";

/* ─── Domain & Confidence helpers ──────────────────── */

const DOMAIN_COLORS: Record<string, string> = {
  AIR: "#38BDF8",
  SURFACE: "#60A5FA",
  SUBSURFACE: "#0D9488",
  LAND: "#A3844B",
  SPACE: "#A78BFA",
  CYBER: "#34D399",
  UNKNOWN: "#6B7280",
};

function getEntityDomain(entityType: string): string {
  const t = entityType.toUpperCase();
  if (t.includes("AIR")) return "AIR";
  if (t.includes("SURFACE") || t.includes("SHIP") || t.includes("VESSEL")) return "SURFACE";
  if (t.includes("SUB")) return "SUBSURFACE";
  if (t.includes("LAND") || t.includes("VEHICLE")) return "LAND";
  if (t.includes("SPACE") || t.includes("SAT")) return "SPACE";
  if (t.includes("CYBER")) return "CYBER";
  return "UNKNOWN";
}

function formatTimeCompact(date: Date): string {
  const diff = Date.now() - date.getTime();
  if (diff < 60_000) return `${Math.floor(diff / 1000)}s ago`;
  if (diff < 3600_000) return `${Math.floor(diff / 60_000)}m ago`;
  return date.toISOString().substring(11, 19) + "Z";
}

function domainIcon(domain: string): string {
  switch (domain) {
    case "AIR": return "✈️";
    case "SURFACE": return "🚢";
    case "SUBSURFACE": return "🤿";
    case "LAND": return "🚗";
    case "SPACE": return "🛰️";
    case "CYBER": return "💻";
    default: return "❓";
  }
}

/* ─── Sort types ───────────────────────────────────── */

type SortField = "domain" | "trackId" | "confidence" | "hostileClass" | "updatedAt";

interface FusionSidePanelProps {
  onOpenScrubber: () => void;
  scrubberOpen: boolean;
}

export const FusionSidePanel: React.FC<FusionSidePanelProps> = ({
  onOpenScrubber,
  scrubberOpen,
}) => {
  const currentTracksMap = useTrackStore((s) => s.tracks);
  const selectedTrackId = useTrackStore((s) => s.selectedTrackId);
  const selectTrack = useTrackStore((s) => s.selectTrack);
  const currentObservations = useSensorStore((s) => s.rawObservations);
  const toggleDetailPanel = useUIStore((s) => s.toggleDetailPanel);
  const detailPanelOpen = useUIStore((s) => s.detailPanelOpen);
  const alertsMap = useAlertStore((s) => s.alerts);

  const [activeTab, setActiveTab] = useState<"tracks" | "alerts">("tracks");
  const [sortField, setSortField] = useState<SortField>("confidence");
  const [sortDirection, setSortDirection] = useState<"asc" | "desc">("desc");

  const tracks = useMemo(() => Array.from(currentTracksMap.values()), [currentTracksMap]);

  // KPI values
  const totalTracks = tracks.length;
  const hostileTracks = tracks.filter((t) => t.hostileClass === "HOSTILE").length;
  const avgConfidence = useMemo(() => {
    if (!tracks.length) return 0;
    return tracks.reduce((sum, t) => sum + t.confidenceScore, 0) / tracks.length;
  }, [tracks]);
  const rawObs = currentObservations.size;

  // Sort handler
  const handleSort = (field: SortField) => {
    if (field === sortField) {
      setSortDirection((d) => (d === "asc" ? "desc" : "asc"));
    } else {
      setSortField(field);
      setSortDirection(field === "confidence" ? "desc" : "asc");
    }
  };

  // Sorted tracks
  const sortedTracks = useMemo(() => {
    const list = [...tracks];
    const dir = sortDirection === "asc" ? 1 : -1;
    switch (sortField) {
      case "confidence":
        list.sort((a, b) => dir * (a.confidenceScore - b.confidenceScore));
        break;
      case "hostileClass": {
        const rank: Record<string, number> = { HOSTILE: 0, UNKNOWN: 1, NEUTRAL: 2, FRIENDLY: 3 };
        list.sort((a, b) => dir * ((rank[a.hostileClass] ?? 4) - (rank[b.hostileClass] ?? 4)));
        break;
      }
      case "domain":
        list.sort((a, b) => dir * getEntityDomain(a.entityType).localeCompare(getEntityDomain(b.entityType)));
        break;
      case "trackId":
        list.sort((a, b) => dir * a.trackId.localeCompare(b.trackId));
        break;
      case "updatedAt":
        list.sort((a, b) => dir * (a.updatedAt.getTime() - b.updatedAt.getTime()));
        break;
    }
    return list.slice(0, 100);
  }, [tracks, sortField, sortDirection]);

  const handleTrackClick = (t: FusedTrack) => {
    selectTrack(t.trackId);
    if (!detailPanelOpen) toggleDetailPanel();
  };

  const criticalAlerts = Array.from(alertsMap.values()).filter((a: AnomalyAlert) => a.severity === "CRITICAL").length;

  return (
    <div
      data-testid="fusion-side-panel"
      style={{
        display: "flex",
        flexDirection: "column",
        height: "100%",
        color: "var(--ds-text-primary, #F1F5F9)",
        overflow: "hidden",
        fontSize: "var(--ds-text-sm, 0.8125rem)",
      }}
    >
      {/* ── Tabs ──────────────────────────────────────── */}
      <div className="ds-tabs">
        <button
          className={`ds-tab ${activeTab === "tracks" ? "ds-tab--active" : ""}`}
          onClick={() => setActiveTab("tracks")}
          data-testid="tab-track-grid"
        >
          Track Grid
        </button>
        <button
          className={`ds-tab ${activeTab === "alerts" ? "ds-tab--active" : ""}`}
          onClick={() => setActiveTab("alerts")}
          data-testid="tab-alert-queue"
        >
          Alert Queue
          {criticalAlerts > 0 && (
            <span
              style={{
                marginLeft: "6px",
                padding: "1px 6px",
                borderRadius: "9999px",
                backgroundColor: "var(--ds-status-error, #EF4444)",
                color: "#fff",
                fontSize: "var(--ds-text-2xs, 0.625rem)",
                fontWeight: 700,
              }}
            >
              {criticalAlerts}
            </span>
          )}
        </button>
        {/* Replay button inside tabs bar */}
        <div style={{ flex: 1 }} />
        <button
          className={`ds-tab ${scrubberOpen ? "ds-tab--active" : ""}`}
          onClick={onOpenScrubber}
          style={{ fontSize: "var(--ds-text-xs, 0.75rem)" }}
          aria-label="Replay Mode"
        >
          {scrubberOpen ? "⏹ Live" : "⏮ Replay"}
        </button>
      </div>

      {/* ── Tab Content ───────────────────────────────── */}
      {activeTab === "tracks" ? (
        <TrackGridTab
          tracks={sortedTracks}
          totalTracks={totalTracks}
          hostileTracks={hostileTracks}
          avgConfidence={avgConfidence}
          rawObs={rawObs}
          sortField={sortField}
          sortDirection={sortDirection}
          onSort={handleSort}
          selectedTrackId={selectedTrackId}
          onTrackClick={handleTrackClick}
        />
      ) : (
        <AlertQueueTab />
      )}
    </div>
  );
};

/* ─── TrackGridTab ─────────────────────────────────── */

interface TrackGridTabProps {
  tracks: FusedTrack[];
  totalTracks: number;
  hostileTracks: number;
  avgConfidence: number;
  rawObs: number;
  sortField: SortField;
  sortDirection: "asc" | "desc";
  onSort: (field: SortField) => void;
  selectedTrackId: string | null;
  onTrackClick: (t: FusedTrack) => void;
}

const TrackGridTab: React.FC<TrackGridTabProps> = ({
  tracks,
  totalTracks,
  hostileTracks,
  avgConfidence,
  rawObs,
  sortField,
  sortDirection,
  onSort,
  selectedTrackId,
  onTrackClick,
}) => (
  <div style={{ display: "flex", flexDirection: "column", flex: 1, overflow: "hidden" }}>
    {/* KPI Tiles */}
    <div
      style={{
        display: "grid",
        gridTemplateColumns: "1fr 1fr 1fr 1fr",
        gap: "var(--ds-space-sm, 8px)",
        padding: "var(--ds-space-md, 12px) var(--ds-space-lg, 16px)",
      }}
    >
      <div className="ds-kpi-tile" data-testid="kpi-tracks">
        <span className="ds-kpi-value" style={{ color: "var(--ds-status-ok, #10B981)" }}>
          {totalTracks}
        </span>
        <span className="ds-kpi-label">Tracks</span>
      </div>
      <div className="ds-kpi-tile" data-testid="kpi-hostile">
        <span
          className="ds-kpi-value"
          style={{ color: hostileTracks > 0 ? "var(--ds-status-error, #EF4444)" : "var(--ds-text-muted, #64748B)" }}
        >
          {hostileTracks}
        </span>
        <span className="ds-kpi-label">Hostile</span>
      </div>
      <div className="ds-kpi-tile" data-testid="kpi-confidence">
        <span className="ds-kpi-value" style={{ color: "var(--ds-accent-blue, #3B82F6)" }}>
          {(avgConfidence * 100).toFixed(0)}
          <span style={{ fontSize: "var(--ds-text-xs, 0.75rem)" }}>%</span>
        </span>
        <span className="ds-kpi-label">Confidence</span>
      </div>
      <div className="ds-kpi-tile" data-testid="kpi-obs">
        <span className="ds-kpi-value" style={{ color: "var(--ds-accent-amber, #F59E0B)" }}>
          {rawObs}
        </span>
        <span className="ds-kpi-label">Obs/10s</span>
      </div>
    </div>

    {/* Confidence Distribution Histogram */}
    <div
      style={{
        padding: "4px 16px 8px",
        borderBottom: "1px solid var(--ds-border-subtle, rgba(255,255,255,0.06))",
      }}
    >
      <div style={{ fontSize: "var(--ds-text-2xs, 0.625rem)", color: "var(--ds-text-muted, #64748B)", marginBottom: "4px", letterSpacing: "0.05em" }}>
        CONFIDENCE DISTRIBUTION
      </div>
      <ConfidenceHistogram tracks={tracks} height={64} />
    </div>

    {/* Sortable Track Table */}
    <div className="ds-scroll" style={{ flex: 1, overflow: "auto" }}>
      <table className="ds-table" data-testid="fusion-track-table">
        <thead>
          <tr>
            <SortHeader field="domain" label="" current={sortField} direction={sortDirection} onSort={onSort} />
            <SortHeader field="trackId" label="Track ID" current={sortField} direction={sortDirection} onSort={onSort} />
            <SortHeader field="domain" label="Domain" current={sortField} direction={sortDirection} onSort={onSort} />
            <SortHeader field="confidence" label="Conf." current={sortField} direction={sortDirection} onSort={onSort} />
            <SortHeader field="hostileClass" label="Class" current={sortField} direction={sortDirection} onSort={onSort} />
            <SortHeader field="updatedAt" label="Updated" current={sortField} direction={sortDirection} onSort={onSort} />
          </tr>
        </thead>
        <tbody>
          {tracks.map((t) => {
            const domain = getEntityDomain(t.entityType);
            const confPct = Math.round(t.confidenceScore * 100);
            const isSelected = t.trackId === selectedTrackId;
            return (
              <React.Fragment key={t.trackId}>
                <tr
                  className={isSelected ? "ds-row-selected" : ""}
                  onClick={() => onTrackClick(t)}
                  data-testid={`fusion-track-row-${t.trackId}`}
                  style={{ cursor: "pointer" }}
                >
                  <td>
                    <span
                      style={{
                        display: "inline-block",
                        width: "8px",
                        height: "8px",
                        borderRadius: "50%",
                        backgroundColor: DOMAIN_COLORS[domain] ?? "#6B7280",
                      }}
                    />
                  </td>
                  <td>
                    <span style={{ fontWeight: 600, fontSize: "var(--ds-text-xs, 0.75rem)" }}>
                      {t.trackId.length > 12 ? t.trackId.slice(0, 12) + "…" : t.trackId}
                    </span>
                  </td>
                  <td>
                    <span style={{ color: "var(--ds-text-secondary, #94A3B8)", fontSize: "var(--ds-text-xs, 0.75rem)" }}>
                      {domainIcon(domain)} {domain}
                    </span>
                  </td>
                  <td>
                    <div style={{ display: "flex", alignItems: "center", gap: "4px" }}>
                      <div
                        style={{
                          width: "40px",
                          height: "4px",
                          backgroundColor: "var(--ds-bg-tertiary, #1E293B)",
                          borderRadius: "2px",
                          overflow: "hidden",
                        }}
                      >
                        <div
                          style={{
                            width: `${confPct}%`,
                            height: "100%",
                            backgroundColor:
                              confPct >= 80 ? "#10B981" : confPct >= 60 ? "#3B82F6" : "#F59E0B",
                            borderRadius: "2px",
                          }}
                        />
                      </div>
                      <span
                        style={{
                          fontFamily: "var(--ds-font-mono, monospace)",
                          fontSize: "var(--ds-text-2xs, 0.625rem)",
                          color: "var(--ds-text-muted, #64748B)",
                        }}
                      >
                        {confPct}%
                      </span>
                    </div>
                  </td>
                  <td>
                    <span
                      style={{
                        fontSize: "var(--ds-text-xs, 0.75rem)",
                        fontWeight: 600,
                        color:
                          t.hostileClass === "HOSTILE"
                            ? "var(--ds-status-error, #EF4444)"
                            : t.hostileClass === "FRIENDLY"
                            ? "var(--ds-status-ok, #10B981)"
                            : "var(--ds-text-muted, #6B7280)",
                      }}
                    >
                      {t.hostileClass.slice(0, 5)}
                    </span>
                  </td>
                  <td>
                    <span
                      style={{
                        fontFamily: "var(--ds-font-mono, monospace)",
                        fontSize: "var(--ds-text-2xs, 0.625rem)",
                        color: "var(--ds-text-muted, #64748B)",
                      }}
                    >
                      {formatTimeCompact(t.updatedAt)}
                    </span>
                  </td>
                </tr>

                {/* Inline expansion */}
                {isSelected && (
                  <tr className="ds-expand-row">
                    <td colSpan={6}>
                      <TrackDrillDown track={t} />
                    </td>
                  </tr>
                )}
              </React.Fragment>
            );
          })}

          {tracks.length === 0 && (
            <tr>
              <td
                colSpan={6}
                style={{
                  textAlign: "center",
                  color: "var(--ds-text-muted, #64748B)",
                  padding: "40px",
                }}
              >
                Awaiting track data…
              </td>
            </tr>
          )}
        </tbody>
      </table>
    </div>
  </div>
);

/* ─── Sort Header ──────────────────────────────────── */

const SortHeader: React.FC<{
  field: SortField;
  label: string;
  current: SortField;
  direction: "asc" | "desc";
  onSort: (f: SortField) => void;
}> = ({ field, label, current, direction, onSort }) => (
  <th
    className={current === field ? "ds-sort-active" : ""}
    onClick={() => onSort(field)}
    style={{ cursor: "pointer" }}
  >
    {label} {current === field ? (direction === "asc" ? "▲" : "▼") : ""}
  </th>
);

/* ─── Track Drill-Down ─────────────────────────────── */

const TrackDrillDown: React.FC<{ track: FusedTrack }> = ({ track }) => {
  const domain = getEntityDomain(track.entityType);
  const confPct = Math.round(track.confidenceScore * 100);

  return (
    <div
      data-testid={`track-detail-${track.trackId}`}
      style={{
        display: "grid",
        gridTemplateColumns: "1fr 1fr",
        gap: "var(--ds-space-lg, 16px)",
      }}
    >
      <div>
        <div className="ds-section-header">Identity</div>
        <DetailRow label="Track ID" value={track.trackId} />
        <DetailRow label="Domain" value={`${domainIcon(domain)} ${domain}`} />
        <DetailRow label="Entity Type" value={track.entityType} />
        <DetailRow
          label="Classification"
          value={track.hostileClass}
          valueColor={
            track.hostileClass === "HOSTILE"
              ? "var(--ds-status-error)"
              : track.hostileClass === "FRIENDLY"
              ? "var(--ds-status-ok)"
              : undefined
          }
        />
      </div>
      <div>
        <div className="ds-section-header">Metrics</div>
        <DetailRow label="Confidence" value={`${confPct}%`} />
        <DetailRow
          label="Position"
          value={`${track.position.latitude.toFixed(2)}°N, ${Math.abs(track.position.longitude).toFixed(2)}°${track.position.longitude < 0 ? "W" : "E"}`}
        />
        <DetailRow label="Sources" value={`${track.sourceCount} sensor(s)`} />
        <DetailRow label="Last Update" value={track.updatedAt.toISOString().substring(11, 19) + "Z"} />
      </div>
    </div>
  );
};

/* ─── Alert Queue Tab ──────────────────────────────── */

const AlertQueueTab: React.FC = () => {
  const alertsMap = useAlertStore((s) => s.alerts);
  const recentAlerts = useMemo(
    () => Array.from(alertsMap.values()).sort((a, b) => b.detectedAt.getTime() - a.detectedAt.getTime()).slice(0, 50),
    [alertsMap],
  );

  const severityColor = (sev: string) => {
    switch (sev) {
      case "CRITICAL": return "var(--ds-status-error, #EF4444)";
      case "ELEVATED": return "var(--ds-status-warn, #F59E0B)";
      case "WATCH": return "var(--ds-accent-blue, #3B82F6)";
      default: return "var(--ds-text-muted, #64748B)";
    }
  };

  return (
    <div className="ds-scroll" style={{ flex: 1, overflow: "auto" }}>
      <table className="ds-table" data-testid="fusion-alert-table">
        <thead>
          <tr>
            <th>Severity</th>
            <th>Alert ID</th>
            <th>Type</th>
            <th>Time</th>
          </tr>
        </thead>
        <tbody>
          {recentAlerts.map((a) => (
            <tr key={a.alertId}>
              <td>
                <span
                  style={{
                    color: severityColor(a.severity),
                    fontWeight: 700,
                    fontSize: "var(--ds-text-xs, 0.75rem)",
                  }}
                >
                  {a.severity}
                </span>
              </td>
              <td>
                <span style={{ fontWeight: 600, fontSize: "var(--ds-text-xs, 0.75rem)" }}>
                  {a.alertId.length > 12 ? a.alertId.slice(0, 12) + "…" : a.alertId}
                </span>
              </td>
              <td>
                <span style={{ color: "var(--ds-text-secondary, #94A3B8)", fontSize: "var(--ds-text-xs, 0.75rem)" }}>
                  {a.anomalyType}
                </span>
              </td>
              <td>
                <span
                  style={{
                    fontFamily: "var(--ds-font-mono, monospace)",
                    fontSize: "var(--ds-text-2xs, 0.625rem)",
                    color: "var(--ds-text-muted, #64748B)",
                  }}
                >
                  {formatTimeCompact(a.detectedAt)}
                </span>
              </td>
            </tr>
          ))}
          {recentAlerts.length === 0 && (
            <tr>
              <td
                colSpan={4}
                style={{
                  textAlign: "center",
                  color: "var(--ds-text-muted, #64748B)",
                  padding: "40px",
                }}
              >
                No active alerts
              </td>
            </tr>
          )}
        </tbody>
      </table>
    </div>
  );
};

/* ─── Detail Row helper ────────────────────────────── */

const DetailRow: React.FC<{
  label: string;
  value: string;
  valueColor?: string;
}> = ({ label, value, valueColor }) => (
  <div
    style={{
      display: "flex",
      justifyContent: "space-between",
      padding: "3px 0",
      fontSize: "var(--ds-text-xs, 0.75rem)",
    }}
  >
    <span style={{ color: "var(--ds-text-muted, #64748B)" }}>{label}</span>
    <span
      style={{
        fontFamily: "var(--ds-font-mono, monospace)",
        color: valueColor || "var(--ds-text-primary, #F1F5F9)",
        fontWeight: 500,
      }}
    >
      {value}
    </span>
  </div>
);
