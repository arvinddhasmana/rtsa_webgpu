// CLASSIFICATION: UNCLASSIFIED
// src/components/forensics/ForensicsPanel.tsx

import React, { useState } from "react";
import { QueryBuilder } from "./QueryBuilder";
import { ResultsView } from "./ResultsView";
import { MapReplay } from "./MapReplay";
import { queryClient, HistoricalQueryRequest, HistoricalQueryResponse } from "../../api/query-client";
import { useTrackStore } from "../../stores/trackStore";
import { useAuthStore } from "../../stores/authStore";

/**
 * ForensicsPanel — historical analysis and query interface.
 *
 * Sub-components:
 *   - QueryBuilder: form to build historical queries
 *   - ResultsView: paginated table of query results
 *   - MapReplay: animate historical track positions on map
 */
export const ForensicsPanel: React.FC = () => {
  const [isLoading, setIsLoading] = useState(false);
  const [results, setResults] = useState<HistoricalQueryResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const selectTrack = useTrackStore((s) => s.selectTrack);
  const clearanceLevel = useAuthStore((s) => s.clearanceLevel);

  const handleQuery = async (req: HistoricalQueryRequest) => {
    setIsLoading(true);
    setError(null);
    try {
      const res = await queryClient.queryHistory(req);
      setResults(res);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Query failed");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div
      data-testid="forensics-panel"
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
          padding: "6px 12px",
          borderBottom: "1px solid #334155",
          fontSize: "0.875rem",
          fontWeight: "bold",
          letterSpacing: "0.1em",
        }}
      >
        FORENSICS
      </div>

      <div style={{ display: "flex", flex: 1, overflow: "hidden" }}>
        {/* Query builder (left 30%) */}
        <div
          style={{
            width: "30%",
            borderRight: "1px solid #334155",
            overflowY: "auto",
          }}
        >
          <QueryBuilder
            onQuery={(req) => void handleQuery(req)}
            isLoading={isLoading}
          />
        </div>

        {/* Results (right 70%) */}
        <div style={{ flex: 1, display: "flex", flexDirection: "column", overflow: "hidden" }}>
          {error && (
            <div
              style={{
                padding: "8px",
                color: "#DC2626",
                fontSize: "0.75rem",
                borderBottom: "1px solid #334155",
              }}
            >
              Error: {error}
            </div>
          )}

          {results ? (
            <>
              <div style={{ flex: 1, overflow: "hidden" }}>
                <ResultsView
                  tracks={results.tracks}
                  alerts={results.alerts}
                  totalCount={results.totalCount}
                  classificationCeiling={clearanceLevel}
                  onTrackSelect={selectTrack}
                />
              </div>
              {results.tracks.length > 0 && (
                <MapReplay tracks={results.tracks} />
              )}
            </>
          ) : (
            <div
              style={{
                flex: 1,
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                color: "#9CA3AF",
                fontSize: "0.8rem",
              }}
            >
              {isLoading ? "Running query..." : "Run a query to view results"}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
