// CLASSIFICATION: UNCLASSIFIED
// src/components/fusion/FusionSidePanel.tsx
// Fusion Engine telemetry panel — confidence histogram + track list.

import React, { useMemo, useState } from "react";
import { useSensorStore } from "../../stores/sensorStore";
import { useTrackStore } from "../../stores/trackStore";
import { useUIStore } from "../../stores/uiStore";
import { FusedTrack } from "../../types/track";

/** Confidence histogram bucket */
const BUCKETS = [
  { label: "High", min: 0.8, max: 1.01, color: "#10B981" },
  { label: "Medium", min: 0.6, max: 0.8, color: "#3B82F6" },
  { label: "Low", min: 0.4, max: 0.6, color: "#F59E0B" },
  { label: "Tentative", min: 0, max: 0.4, color: "#EF4444" },
];

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

function getHostileBadgeStyle(cls: string): React.CSSProperties {
  if (cls === "HOSTILE") return { color: "#EF4444", fontWeight: "bold" };
  if (cls === "FRIENDLY") return { color: "#10B981" };
  if (cls === "NEUTRAL") return { color: "#94A3B8" };
  return { color: "#6B7280" };
}

type SortKey = "confidence" | "hostile" | "domain" | "updated";

export const FusionSidePanel: React.FC = () => {
  const currentTracksMap = useTrackStore((s) => s.tracks);
  const selectedTrackId = useTrackStore((s) => s.selectedTrackId);
  const selectTrack = useTrackStore((s) => s.selectTrack);
  const currentObservations = useSensorStore((s) => s.rawObservations);
  const toggleDetailPanel = useUIStore((s) => s.toggleDetailPanel);
  const detailPanelOpen = useUIStore((s) => s.detailPanelOpen);
  const [sortKey, setSortKey] = useState<SortKey>("confidence");
  const [searchFilter, setSearchFilter] = useState("");

  const tracks = useMemo(() => Array.from(currentTracksMap.values()), [currentTracksMap]);

  // Domain breakdown with hostile counts
  const domainStats = useMemo(() => {
    const map: Record<string, { count: number; hostile: number }> = {};
    tracks.forEach((t) => {
      const domain = getEntityDomain(t.entityType);
      if (!map[domain]) map[domain] = { count: 0, hostile: 0 };
      map[domain].count++;
      if (t.hostileClass === "HOSTILE") map[domain].hostile++;
    });
    return Object.entries(map).map(([domain, stats]) => ({ domain, ...stats }));
  }, [tracks]);

  // Confidence histogram
  const histogram = useMemo(() => {
    const total = tracks.length || 1;
    return BUCKETS.map((b) => {
      const count = tracks.filter(
        (t) => t.confidenceScore >= b.min && t.confidenceScore < b.max
      ).length;
      return { ...b, count, pct: (count / total) * 100 };
    });
  }, [tracks]);

  const avgConfidence = useMemo(() => {
    if (!tracks.length) return 0;
    return tracks.reduce((sum, t) => sum + t.confidenceScore, 0) / tracks.length;
  }, [tracks]);

  // Sorted + filtered track list
  const sortedTracks = useMemo(() => {
    let list = tracks.filter(
      (t) =>
        !searchFilter ||
        t.trackId.toLowerCase().includes(searchFilter.toLowerCase()) ||
        t.entityType.toLowerCase().includes(searchFilter.toLowerCase())
    );
    switch (sortKey) {
      case "confidence":
        list = list.sort((a, b) => b.confidenceScore - a.confidenceScore);
        break;
      case "hostile":
        list = list.sort((a, b) => {
          const rank = { HOSTILE: 0, UNKNOWN: 1, NEUTRAL: 2, FRIENDLY: 3 };
          return (rank[a.hostileClass as keyof typeof rank] ?? 4) -
            (rank[b.hostileClass as keyof typeof rank] ?? 4);
        });
        break;
      case "domain":
        list = list.sort((a, b) =>
          getEntityDomain(a.entityType).localeCompare(getEntityDomain(b.entityType))
        );
        break;
      case "updated":
        list = list.sort(
          (a, b) => b.updatedAt.getTime() - a.updatedAt.getTime()
        );
        break;
    }
    return list.slice(0, 80);
  }, [tracks, sortKey, searchFilter]);

  const totalTracks = tracks.length;
  const hostileTracks = tracks.filter((t) => t.hostileClass === "HOSTILE").length;
  const rawSensors = currentObservations.size;
  const correlatedSensors = Array.from(currentObservations.values()).filter(
    (o) => o.correlatedTrackId
  ).length;

  const handleTrackClick = (t: FusedTrack) => {
    selectTrack(t.trackId);
    if (!detailPanelOpen) toggleDetailPanel();
  };

  return (
    <div
      data-testid="fusion-side-panel"
      style={{
        display: "flex",
        flexDirection: "column",
        height: "100%",
        color: "#F1F5F9",
        overflowY: "auto",
        fontSize: "0.75rem",
      }}
    >
      {/* ── Header KPI Grid ─────────────────────────────── */}
      <div
        style={{
          padding: "12px 14px 8px",
          borderBottom: "1px solid #1E293B",
        }}
      >
        <div
          style={{
            fontSize: "0.7rem",
            color: "#F59E0B",
            fontWeight: "bold",
            letterSpacing: "0.08em",
            marginBottom: "10px",
          }}
        >
          ⚡ FUSION ENGINE
        </div>
        <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "8px" }}>
          <MetricBox label="Active Tracks" value={totalTracks} />
          <MetricBox label="Hostile" value={hostileTracks} accent="#EF4444" />
          <MetricBox label="Raw Obs/10s" value={rawSensors} />
          <MetricBox label="Correlated" value={correlatedSensors} accent="#3B82F6" />
        </div>

        {/* Avg confidence badge */}
        <div
          style={{
            marginTop: "10px",
            display: "flex",
            alignItems: "center",
            gap: "8px",
          }}
        >
          <span style={{ color: "#94A3B8", fontSize: "0.65rem" }}>AVG CONFIDENCE</span>
          <div
            style={{
              flex: 1,
              height: "6px",
              backgroundColor: "#1E293B",
              borderRadius: "3px",
              overflow: "hidden",
            }}
          >
            <div
              style={{
                width: `${avgConfidence * 100}%`,
                height: "100%",
                backgroundColor:
                  avgConfidence > 0.8
                    ? "#10B981"
                    : avgConfidence > 0.6
                    ? "#3B82F6"
                    : "#F59E0B",
                transition: "width 0.3s ease",
              }}
            />
          </div>
          <span
            style={{
              fontSize: "0.7rem",
              fontWeight: "bold",
              color: "#F1F5F9",
              minWidth: "36px",
              textAlign: "right",
            }}
          >
            {(avgConfidence * 100).toFixed(0)}%
          </span>
        </div>
      </div>

      {/* ── Confidence Histogram ─────────────────────────── */}
      <div
        style={{
          padding: "10px 14px 8px",
          borderBottom: "1px solid #1E293B",
        }}
      >
        <div
          style={{
            fontSize: "0.65rem",
            color: "#64748B",
            textTransform: "uppercase",
            letterSpacing: "0.06em",
            marginBottom: "8px",
          }}
        >
          Confidence Distribution
        </div>
        <div
          data-testid="confidence-histogram"
          style={{ display: "flex", flexDirection: "column", gap: "5px" }}
        >
          {histogram.map((b) => (
            <div key={b.label} style={{ display: "flex", alignItems: "center", gap: "6px" }}>
              <span
                style={{ width: "58px", fontSize: "0.65rem", color: b.color, flexShrink: 0 }}
              >
                {b.label}
              </span>
              <div
                style={{
                  flex: 1,
                  height: "8px",
                  backgroundColor: "#1E293B",
                  borderRadius: "2px",
                  overflow: "hidden",
                }}
              >
                <div
                  style={{
                    width: `${b.pct}%`,
                    height: "100%",
                    backgroundColor: b.color,
                    transition: "width 0.3s ease",
                    opacity: 0.8,
                  }}
                />
              </div>
              <span
                style={{ width: "24px", fontSize: "0.65rem", color: "#64748B", textAlign: "right" }}
              >
                {b.count}
              </span>
            </div>
          ))}
        </div>
      </div>

      {/* ── Domain Breakdown ─────────────────────────────── */}
      {domainStats.length > 0 && (
        <div style={{ padding: "10px 14px 8px", borderBottom: "1px solid #1E293B" }}>
          <div
            style={{
              fontSize: "0.65rem",
              color: "#64748B",
              textTransform: "uppercase",
              letterSpacing: "0.06em",
              marginBottom: "8px",
            }}
          >
            By Domain
          </div>
          <div style={{ display: "flex", flexWrap: "wrap", gap: "6px" }}>
            {domainStats.map((d) => (
              <div
                key={d.domain}
                style={{
                  backgroundColor: "rgba(255,255,255,0.04)",
                  border: `1px solid ${DOMAIN_COLORS[d.domain] ?? "#334155"}`,
                  borderRadius: "4px",
                  padding: "4px 8px",
                  display: "flex",
                  alignItems: "center",
                  gap: "6px",
                }}
              >
                <span
                  style={{
                    fontSize: "0.6rem",
                    color: DOMAIN_COLORS[d.domain] ?? "#94A3B8",
                    fontWeight: "bold",
                  }}
                >
                  {d.domain}
                </span>
                <span style={{ fontSize: "0.7rem", color: "#F1F5F9" }}>{d.count}</span>
                {d.hostile > 0 && (
                  <span style={{ fontSize: "0.6rem", color: "#EF4444" }}>
                    ⬆{d.hostile}
                  </span>
                )}
              </div>
            ))}
          </div>
        </div>
      )}

      {/* ── Track List ───────────────────────────────────── */}
      <div style={{ flex: 1, display: "flex", flexDirection: "column", overflow: "hidden" }}>
        {/* Search + Sort bar */}
        <div
          style={{
            padding: "8px 14px",
            borderBottom: "1px solid #1E293B",
            display: "flex",
            gap: "6px",
            alignItems: "center",
          }}
        >
          <input
            placeholder="Filter tracks…"
            value={searchFilter}
            onChange={(e) => setSearchFilter(e.target.value)}
            style={{
              flex: 1,
              padding: "3px 8px",
              backgroundColor: "#0F172A",
              color: "#F1F5F9",
              border: "1px solid #334155",
              borderRadius: "4px",
              fontSize: "0.7rem",
            }}
          />
          <select
            value={sortKey}
            onChange={(e) => setSortKey(e.target.value as SortKey)}
            style={{
              backgroundColor: "#1E293B",
              color: "#94A3B8",
              border: "1px solid #334155",
              borderRadius: "4px",
              padding: "3px 6px",
              fontSize: "0.65rem",
              cursor: "pointer",
            }}
          >
            <option value="confidence">Confidence ↓</option>
            <option value="hostile">Hostile First</option>
            <option value="domain">Domain</option>
            <option value="updated">Latest</option>
          </select>
        </div>

        {/* Track rows */}
        <div style={{ flex: 1, overflowY: "auto" }}>
          {sortedTracks.length === 0 ? (
            <div
              style={{
                padding: "24px",
                textAlign: "center",
                color: "#475569",
                fontSize: "0.75rem",
              }}
            >
              {tracks.length === 0
                ? "Awaiting track data…"
                : "No tracks match filter"}
            </div>
          ) : (
            sortedTracks.map((t) => (
              <TrackRow
                key={t.trackId}
                track={t}
                selected={t.trackId === selectedTrackId}
                onClick={() => handleTrackClick(t)}
              />
            ))
          )}
        </div>
      </div>
    </div>
  );
};

/* ─── Sub-components ─────────────────────────────────── */

const MetricBox: React.FC<{ label: string; value: number; accent?: string }> = ({
  label,
  value,
  accent,
}) => (
  <div
    style={{
      backgroundColor: "rgba(255,255,255,0.04)",
      border: "1px solid #1E293B",
      borderRadius: "6px",
      padding: "8px 10px",
      display: "flex",
      flexDirection: "column",
      alignItems: "center",
      gap: "2px",
    }}
  >
    <span
      style={{ fontSize: "1.35rem", fontWeight: "bold", color: accent ?? "#F1F5F9" }}
    >
      {value}
    </span>
    <span style={{ fontSize: "0.6rem", color: "#64748B", textAlign: "center" }}>
      {label}
    </span>
  </div>
);

const TrackRow: React.FC<{
  track: FusedTrack;
  selected: boolean;
  onClick: () => void;
}> = ({ track, selected, onClick }) => {
  const domain = getEntityDomain(track.entityType);
  const domainColor = DOMAIN_COLORS[domain] ?? "#6B7280";
  const confPct = Math.round(track.confidenceScore * 100);

  return (
    <div
      data-testid={`fusion-track-row-${track.trackId}`}
      onClick={onClick}
      style={{
        display: "flex",
        alignItems: "center",
        gap: "8px",
        padding: "7px 14px",
        cursor: "pointer",
        borderLeft: selected ? "3px solid #3B82F6" : "3px solid transparent",
        backgroundColor: selected
          ? "rgba(59, 130, 246, 0.1)"
          : "transparent",
        borderBottom: "1px solid rgba(255,255,255,0.04)",
        transition: "background-color 0.15s ease",
      }}
    >
      {/* Domain indicator */}
      <div
        style={{
          width: "8px",
          height: "8px",
          borderRadius: "50%",
          backgroundColor: domainColor,
          flexShrink: 0,
        }}
      />

      {/* Track info */}
      <div style={{ flex: 1, minWidth: 0 }}>
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
          }}
        >
          <span
            style={{
              fontSize: "0.7rem",
              fontWeight: "bold",
              color: "#E2E8F0",
              overflow: "hidden",
              textOverflow: "ellipsis",
              whiteSpace: "nowrap",
            }}
          >
            {track.trackId}
          </span>
          <span style={getHostileBadgeStyle(track.hostileClass)}>
            {track.hostileClass.slice(0, 4)}
          </span>
        </div>
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: "6px",
            marginTop: "3px",
          }}
        >
          <span style={{ fontSize: "0.6rem", color: "#64748B" }}>{domain}</span>
          {/* Confidence mini-bar */}
          <div
            style={{
              flex: 1,
              height: "3px",
              backgroundColor: "#1E293B",
              borderRadius: "2px",
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
          <span style={{ fontSize: "0.6rem", color: "#64748B" }}>{confPct}%</span>
        </div>
      </div>
    </div>
  );
};
