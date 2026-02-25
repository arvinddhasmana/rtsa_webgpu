// CLASSIFICATION: UNCLASSIFIED
// src/components/map/SensorCoverageLayer.tsx

import React from "react";

interface SensorCoverage {
  sensorId: string;
  coverageType: string;
}

interface SensorCoverageLayerProps {
  coverages: SensorCoverage[];
}

/**
 * SensorCoverageLayer — renders sensor coverage arc overlays on the map.
 */
export const SensorCoverageLayer: React.FC<SensorCoverageLayerProps> = ({
  coverages,
}) => {
  return (
    <div data-testid="sensor-coverage-layer">
      {coverages.map((coverage) => (
        <div
          key={coverage.sensorId}
          data-testid={`coverage-${coverage.sensorId}`}
          style={{
            position: "absolute",
            borderRadius: "50%",
            border: "1px dashed #2563EB",
            opacity: 0.3,
            pointerEvents: "none",
          }}
          title={`Sensor: ${coverage.sensorId} (${coverage.coverageType})`}
        />
      ))}
    </div>
  );
};
