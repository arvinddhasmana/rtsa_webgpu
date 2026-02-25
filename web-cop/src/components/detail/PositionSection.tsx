// CLASSIFICATION: UNCLASSIFIED
// src/components/detail/PositionSection.tsx

import React from "react";
import { FusedTrack } from "../../types/track";
import { formatPosition } from "../../utils/coordinates";
import { relativeTime } from "../../utils/time";

interface PositionSectionProps {
  track: FusedTrack;
}

/**
 * PositionSection — displays position and velocity information in DMS format.
 */
export const PositionSection: React.FC<PositionSectionProps> = ({ track }) => {
  const { position } = track;

  return (
    <div data-testid="position-section" style={{ padding: "8px" }}>
      <table style={{ width: "100%", borderCollapse: "collapse", fontSize: "0.75rem" }}>
        <tbody>
          <tr>
            <td style={{ color: "#9CA3AF", paddingRight: "8px" }}>Position</td>
            <td style={{ fontFamily: "monospace" }}>
              {formatPosition(position.latitude, position.longitude)}
            </td>
          </tr>
          {position.altitudeMeters !== undefined && (
            <tr>
              <td style={{ color: "#9CA3AF" }}>Altitude</td>
              <td>{position.altitudeMeters.toFixed(0)} m</td>
            </tr>
          )}
          {position.speedKnots !== undefined && (
            <tr>
              <td style={{ color: "#9CA3AF" }}>Speed</td>
              <td>{position.speedKnots.toFixed(1)} kts</td>
            </tr>
          )}
          {position.headingDegrees !== undefined && (
            <tr>
              <td style={{ color: "#9CA3AF" }}>Heading</td>
              <td>{position.headingDegrees.toFixed(0)}°</td>
            </tr>
          )}
          <tr>
            <td style={{ color: "#9CA3AF" }}>Last Update</td>
            <td>{relativeTime(track.updatedAt)}</td>
          </tr>
        </tbody>
      </table>
    </div>
  );
};
