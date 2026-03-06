// CLASSIFICATION: UNCLASSIFIED
// src/components/nato/NominationQueue.tsx
//
// Track nomination queue for NATO COP export.
// Shows outbound tracks awaiting review with [Nominate] / [Revoke] buttons.
// Classification guard: blocks export when track classification exceeds ceiling.

import React, { useState } from "react";
import type { FusedTrack } from "../../types/track";

interface NominationEntry {
  track: FusedTrack;
  status: "PENDING" | "NOMINATED" | "REVOKED" | "BLOCKED";
  blockedReason?: string;
}

interface NominationQueueProps {
  tracks: FusedTrack[];
  /** Maximum classification level the liaison is authorised to share */
  maxShareableClassification?: string;
}

const SHAREABLE_CLASSIFICATIONS = ["UNCLASSIFIED", "PROTECTED_A"];

function canShare(track: FusedTrack, maxCls: string): boolean {
  const trackIdx = SHAREABLE_CLASSIFICATIONS.indexOf(track.classification);
  const maxIdx = SHAREABLE_CLASSIFICATIONS.indexOf(maxCls);
  // If track classification is not in the shareable list, it cannot be shared.
  if (trackIdx === -1) return false;
  return trackIdx <= maxIdx;
}

/**
 * NominationQueue — outbound track nomination panel for NATO Liaison.
 *
 * Features:
 *   - Lists high-confidence tracks (≥ 0.75) as candidates for export
 *   - Classification guard: BLOCKED badge when track clearance exceeds ceiling
 *   - [Nominate] adds to outbound N-COP queue
 *   - [Revoke] removes from queue with audit note
 */
export const NominationQueue: React.FC<NominationQueueProps> = ({
  tracks,
  maxShareableClassification = "PROTECTED_A",
}) => {
  const [entries, setEntries] = useState<NominationEntry[]>(() =>
    tracks
      .filter((t) => t.confidenceScore >= 0.75 && t.status === "ACTIVE")
      .slice(0, 20)
      .map((t) => ({
        track: t,
        status: canShare(t, maxShareableClassification) ? "PENDING" : "BLOCKED",
        blockedReason: canShare(t, maxShareableClassification)
          ? undefined
          : `Classification ${t.classification} exceeds sharing ceiling`,
      }))
  );

  const nominate = (trackId: string) => {
    setEntries((prev) =>
      prev.map((e) =>
        e.track.trackId === trackId && e.status === "PENDING"
          ? { ...e, status: "NOMINATED" }
          : e
      )
    );
  };

  const revoke = (trackId: string) => {
    setEntries((prev) =>
      prev.map((e) =>
        e.track.trackId === trackId && e.status === "NOMINATED"
          ? { ...e, status: "REVOKED" }
          : e
      )
    );
  };

  const nominated = entries.filter((e) => e.status === "NOMINATED").length;
  const blocked = entries.filter((e) => e.status === "BLOCKED").length;

  return (
    <div
      data-testid="nomination-queue"
      style={{
        display: "flex",
        flexDirection: "column",
        height: "100%",
        overflow: "hidden",
      }}
    >
      {/* Header */}
      <div
        style={{
          display: "flex",
          alignItems: "center",
          gap: "12px",
          padding: "8px 12px",
          borderBottom: "1px solid #334155",
          fontSize: "0.7rem",
          color: "#94A3B8",
        }}
      >
        <span style={{ fontWeight: "bold", color: "#F1F5F9" }}>
          NOMINATION QUEUE
        </span>
        <span style={{ color: "#10B981" }}>{nominated} nominated</span>
        {blocked > 0 && (
          <span style={{ color: "#EF4444" }}>{blocked} blocked</span>
        )}
      </div>

      {/* Entries */}
      <div style={{ flex: 1, overflowY: "auto" }}>
        {entries.length === 0 ? (
          <div
            style={{
              padding: "24px",
              textAlign: "center",
              color: "#64748B",
              fontSize: "0.75rem",
            }}
          >
            No tracks qualify for nomination (confidence &lt; 75% or no active tracks)
          </div>
        ) : (
          entries.map((entry) => (
            <div
              key={entry.track.trackId}
              data-testid={`nomination-entry-${entry.track.trackId}`}
              style={{
                display: "flex",
                alignItems: "center",
                gap: "10px",
                padding: "8px 12px",
                borderBottom: "1px solid rgba(255,255,255,0.04)",
                backgroundColor:
                  entry.status === "NOMINATED"
                    ? "rgba(59,130,246,0.06)"
                    : entry.status === "BLOCKED"
                    ? "rgba(239,68,68,0.06)"
                    : "transparent",
              }}
            >
              {/* Classification badge */}
              <span
                style={{
                  fontSize: "0.55rem",
                  fontWeight: "bold",
                  padding: "2px 5px",
                  borderRadius: "3px",
                  backgroundColor:
                    entry.status === "BLOCKED" ? "#EF4444" : "#334155",
                  color: "#F1F5F9",
                  whiteSpace: "nowrap",
                }}
              >
                {entry.status === "BLOCKED" ? "⛔ " : ""}{entry.track.classification}
              </span>

              {/* Track info */}
              <div style={{ flex: 1, minWidth: 0 }}>
                <div
                  style={{
                    fontSize: "0.7rem",
                    fontFamily: "monospace",
                    color: "#60A5FA",
                  }}
                >
                  {entry.track.trackId.slice(-12)}
                </div>
                <div style={{ fontSize: "0.6rem", color: "#64748B" }}>
                  {entry.track.entityType} · {(entry.track.confidenceScore * 100).toFixed(0)}% conf
                </div>
                {entry.blockedReason && (
                  <div style={{ fontSize: "0.55rem", color: "#EF4444" }}>
                    {entry.blockedReason}
                  </div>
                )}
              </div>

              {/* Status badge */}
              <span
                style={{
                  fontSize: "0.6rem",
                  fontWeight: "bold",
                  color:
                    entry.status === "NOMINATED"
                      ? "#10B981"
                      : entry.status === "REVOKED"
                      ? "#64748B"
                      : entry.status === "BLOCKED"
                      ? "#EF4444"
                      : "#94A3B8",
                }}
              >
                {entry.status}
              </span>

              {/* Action buttons */}
              {entry.status === "PENDING" && (
                <button
                  data-testid={`nominate-btn-${entry.track.trackId}`}
                  onClick={() => nominate(entry.track.trackId)}
                  style={{
                    padding: "3px 10px",
                    backgroundColor: "rgba(59,130,246,0.15)",
                    color: "#60A5FA",
                    border: "1px solid #3B82F6",
                    borderRadius: "4px",
                    cursor: "pointer",
                    fontSize: "0.65rem",
                    fontWeight: "bold",
                    whiteSpace: "nowrap",
                  }}
                >
                  Nominate
                </button>
              )}
              {entry.status === "NOMINATED" && (
                <button
                  data-testid={`revoke-btn-${entry.track.trackId}`}
                  onClick={() => revoke(entry.track.trackId)}
                  style={{
                    padding: "3px 10px",
                    backgroundColor: "rgba(239,68,68,0.12)",
                    color: "#EF4444",
                    border: "1px solid #EF4444",
                    borderRadius: "4px",
                    cursor: "pointer",
                    fontSize: "0.65rem",
                    fontWeight: "bold",
                    whiteSpace: "nowrap",
                  }}
                >
                  Revoke
                </button>
              )}
            </div>
          ))
        )}
      </div>
    </div>
  );
};
