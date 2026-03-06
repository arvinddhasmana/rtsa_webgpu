// CLASSIFICATION: UNCLASSIFIED
// src/components/layout/IntelSearchDashboard.tsx
//
// Intelligence Analyst — Intel Search Dashboard.
//
// A temporal + spatial forensics query interface for historical track
// and anomaly analysis.
//
// Layout:
//   LEFT:   Query builder (time range, domain, entity type, spatial box)
//   CENTER: Results (sortable track/alert table with MapReplay preview)
//   RIGHT:  ConfidenceHistogram + summary stats for query result set

import React, { useState } from "react";
import type { AnomalyAlert } from "../../types/alert";
import type { FusedTrack } from "../../types/track";
import { HistoricalQueryRequest, queryClient } from "../../api/query-client";
import { ConfidenceHistogram } from "../fusion/ConfidenceHistogram";
import { MapReplay } from "../forensics/MapReplay";
import { QueryBuilder } from "../forensics/QueryBuilder";
import { ResultsView } from "../forensics/ResultsView";
import { useAuthStore } from "../../stores/authStore";

/**
 * IntelSearchDashboard — Intelligence Analyst default view.
 *
 * Provides full forensic query capability:
 *   - Time range filter (absolute or relative)
 *   - Domain filter (AIR / SURFACE / SUBSURFACE / LAND / CYBER)
 *   - Bounding box filter (spatial)
 *   - Entity type and hostile classification filter
 *   - Results: track table + alert table + map replay
 *   - Confidence histogram for quality assessment
 */
export const IntelSearchDashboard: React.FC = () => {
  const [results, setResults] = useState<{
    tracks: FusedTrack[];
    alerts: AnomalyAlert[];
    totalCount: number;
  } | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showReplay, setShowReplay] = useState(false);
  const clearanceLevel = useAuthStore((s) => s.clearanceLevel);

  const handleQuery = async (req: HistoricalQueryRequest) => {
    setLoading(true);
    setError(null);
    try {
      const res = await queryClient.queryHistory(req);
      setResults(res);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Query failed");
      setResults(null);
    } finally {
      setLoading(false);
    }
  };

  const handleTrackSelect = (_trackId: string) => {
    // Track selection will be used for detail panel in future
  };

  return (
    <div
      data-testid="intel-search-dashboard"
      style={{
        flex: 1,
        display: "flex",
        overflow: "hidden",
        backgroundColor: "var(--ds-bg-primary, #0F172A)",
      }}
    >
      {/* ── Left: Query Builder ────────────────────────────── */}
      <div
        style={{
          width: "280px",
          flexShrink: 0,
          borderRight: "1px solid #334155",
          display: "flex",
          flexDirection: "column",
          overflow: "hidden",
        }}
      >
        {/* Header */}
        <div
          style={{
            padding: "10px 14px",
            borderBottom: "1px solid #334155",
            fontSize: "0.8rem",
            fontWeight: "bold",
            color: "#10B981",
            letterSpacing: "0.05em",
            display: "flex",
            alignItems: "center",
            gap: "8px",
          }}
        >
          <span>🔍</span>
          <span>INTEL SEARCH</span>
        </div>

        {/* Query Builder */}
        <div style={{ flex: 1, overflowY: "auto" }}>
          <QueryBuilder onQuery={handleQuery} isLoading={loading} />
        </div>
      </div>

      {/* ── Centre: Results ────────────────────────────────── */}
      <div
        style={{
          flex: 1,
          display: "flex",
          flexDirection: "column",
          overflow: "hidden",
        }}
      >
        {/* Results header */}
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: "12px",
            padding: "8px 14px",
            borderBottom: "1px solid #334155",
            fontSize: "0.75rem",
            color: "#94A3B8",
            flexShrink: 0,
          }}
        >
          {loading && (
            <span style={{ color: "#F59E0B", animation: "pulse 1.5s infinite" }}>
              Searching…
            </span>
          )}
          {results && !loading && (
            <>
              <span style={{ color: "#10B981", fontWeight: "bold" }}>
                {results.totalCount} results
              </span>
              <span>·</span>
              <span>{results.tracks.length} tracks</span>
              <span>·</span>
              <span>{results.alerts.length} alerts</span>
              {results.tracks.length > 0 && (
                <button
                  onClick={() => setShowReplay((v) => !v)}
                  style={{
                    marginLeft: "auto",
                    padding: "3px 10px",
                    backgroundColor: showReplay
                      ? "rgba(59,130,246,0.15)"
                      : "transparent",
                    color: showReplay ? "#60A5FA" : "#64748B",
                    border: `1px solid ${showReplay ? "#3B82F6" : "#334155"}`,
                    borderRadius: "4px",
                    cursor: "pointer",
                    fontSize: "0.65rem",
                    fontWeight: "bold",
                  }}
                >
                  {showReplay ? "⏹ Hide Replay" : "▶ Map Replay"}
                </button>
              )}
            </>
          )}
          {error && (
            <span style={{ color: "#EF4444" }}>⚠ {error}</span>
          )}
          {!results && !loading && !error && (
            <span style={{ color: "#475569" }}>
              Use the query builder to search historical track data
            </span>
          )}
        </div>

        {/* Map Replay (collapsible) */}
        {showReplay && results && (
          <div
            style={{
              height: "220px",
              flexShrink: 0,
              borderBottom: "1px solid #334155",
            }}
          >
            <MapReplay tracks={results.tracks} />
          </div>
        )}

        {/* Results Table */}
        <div style={{ flex: 1, overflow: "hidden" }}>
          {results ? (
            <ResultsView
              tracks={results.tracks}
              alerts={results.alerts}
              totalCount={results.totalCount}
              classificationCeiling={clearanceLevel}
              onTrackSelect={handleTrackSelect}
            />
          ) : (
            <div
              style={{
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                height: "100%",
                color: "#475569",
                fontSize: "0.85rem",
                flexDirection: "column",
                gap: "12px",
              }}
            >
              <span style={{ fontSize: "2rem", opacity: 0.4 }}>🔍</span>
              <span>Build a query to search historical intelligence data</span>
            </div>
          )}
        </div>
      </div>

      {/* ── Right: Stats & Histogram ───────────────────────── */}
      {results && (
        <div
          style={{
            width: "220px",
            flexShrink: 0,
            borderLeft: "1px solid #334155",
            padding: "14px",
            display: "flex",
            flexDirection: "column",
            gap: "16px",
          }}
        >
          <div>
            <div
              style={{
                fontSize: "0.6rem",
                color: "#64748B",
                fontWeight: "bold",
                letterSpacing: "0.05em",
                marginBottom: "8px",
              }}
            >
              CONFIDENCE DISTRIBUTION
            </div>
            <ConfidenceHistogram tracks={results.tracks} height={90} />
          </div>

          {/* Domain breakdown */}
          <div>
            <div
              style={{
                fontSize: "0.6rem",
                color: "#64748B",
                fontWeight: "bold",
                letterSpacing: "0.05em",
                marginBottom: "8px",
              }}
            >
              DOMAIN BREAKDOWN
            </div>
            {["AIR", "SURFACE", "SUBSURFACE", "LAND", "CYBER", "UNKNOWN"].map(
              (domain) => {
                const count = results.tracks.filter((t) =>
                  t.entityType.toUpperCase().includes(domain.toLowerCase() === "unknown" ? "" : domain)
                ).length;
                if (count === 0) return null;
                return (
                  <div
                    key={domain}
                    style={{
                      display: "flex",
                      justifyContent: "space-between",
                      fontSize: "0.65rem",
                      color: "#94A3B8",
                      padding: "2px 0",
                    }}
                  >
                    <span>{domain}</span>
                    <span style={{ color: "#F1F5F9", fontWeight: "bold" }}>
                      {count}
                    </span>
                  </div>
                );
              }
            )}
          </div>

          {/* Hostile breakdown */}
          <div>
            <div
              style={{
                fontSize: "0.6rem",
                color: "#64748B",
                fontWeight: "bold",
                letterSpacing: "0.05em",
                marginBottom: "8px",
              }}
            >
              THREAT BREAKDOWN
            </div>
            {["HOSTILE", "NEUTRAL", "FRIENDLY", "UNKNOWN"].map((cls) => {
              const count = results.tracks.filter((t) => t.hostileClass === cls).length;
              if (count === 0) return null;
              const color =
                cls === "HOSTILE"
                  ? "#EF4444"
                  : cls === "FRIENDLY"
                  ? "#3B82F6"
                  : cls === "NEUTRAL"
                  ? "#10B981"
                  : "#F59E0B";
              return (
                <div
                  key={cls}
                  style={{
                    display: "flex",
                    justifyContent: "space-between",
                    fontSize: "0.65rem",
                    color: "#94A3B8",
                    padding: "2px 0",
                  }}
                >
                  <span style={{ color }}>{cls}</span>
                  <span style={{ color: "#F1F5F9", fontWeight: "bold" }}>
                    {count}
                  </span>
                </div>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
};
