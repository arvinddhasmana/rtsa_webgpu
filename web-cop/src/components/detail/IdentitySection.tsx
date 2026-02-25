// CLASSIFICATION: UNCLASSIFIED
// src/components/detail/IdentitySection.tsx

import React from "react";
import { FusedTrack } from "../../types/track";

interface IdentitySectionProps {
  track: FusedTrack;
}

/**
 * IdentitySection — displays track identity info (ID, type, hostile class, confidence, status).
 */
export const IdentitySection: React.FC<IdentitySectionProps> = ({ track }) => {
  const hostileColors: Record<string, string> = {
    HOSTILE: "#DC2626",
    FRIENDLY: "#2563EB",
    NEUTRAL: "#16A34A",
    UNKNOWN: "#CA8A04",
  };

  return (
    <div data-testid="identity-section" style={{ padding: "8px" }}>
      <table style={{ width: "100%", borderCollapse: "collapse", fontSize: "0.75rem" }}>
        <tbody>
          <tr>
            <td style={{ color: "#9CA3AF", paddingRight: "8px" }}>Track ID</td>
            <td style={{ fontFamily: "monospace" }}>{track.trackId}</td>
          </tr>
          <tr>
            <td style={{ color: "#9CA3AF" }}>Entity Type</td>
            <td>{track.entityType}</td>
          </tr>
          <tr>
            <td style={{ color: "#9CA3AF" }}>Classification</td>
            <td style={{ color: hostileColors[track.hostileClass] ?? "#6B7280", fontWeight: "bold" }}>
              {track.hostileClass}
            </td>
          </tr>
          <tr>
            <td style={{ color: "#9CA3AF" }}>Confidence</td>
            <td>{Math.round(track.confidenceScore * 100)}%</td>
          </tr>
          <tr>
            <td style={{ color: "#9CA3AF" }}>Status</td>
            <td
              style={{
                color: track.status === "ACTIVE" ? "#16A34A" : "#CA8A04",
              }}
            >
              {track.status}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  );
};
