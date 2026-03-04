// CLASSIFICATION: UNCLASSIFIED
// src/components/map/SensorCoverageLayer.tsx
//
// Sensor coverage fan-sectors, colored by sensor type:
//   RADAR  → green  (#16A34A)
//   EW     → purple (#9333EA)
//   AIS    → blue   (#3B82F6)
//   ELINT  → amber  (#F59E0B)
//   ISR    → cyan   (#06B6D4)
//   other  → gray   (#6B7280)

import React, { useEffect } from "react";
import { useSensorCoverage } from "../../hooks/useSensorCoverage";
import { useUIStore } from "../../stores/uiStore";

// Sensor-type → fill color mapping (MIL-aligned color scheme)
const SENSOR_FILL_COLORS: Record<string, string> = {
  RADAR:  "#16A34A",
  EW:     "#9333EA",
  AIS:    "#3B82F6",
  ELINT:  "#F59E0B",
  ISR:    "#06B6D4",
};

function getSensorColor(sensorType: string): string {
  const key = sensorType.toUpperCase();
  return SENSOR_FILL_COLORS[key] ?? "#6B7280";
}

// The MapView component exposes the map instance on window.__RTSA_MAP__
// We use a separate functional component here to strictly manage the coverage layer.
export const SensorCoverageLayer: React.FC = () => {
  const sensors = useSensorCoverage();
  const isVisible = useUIStore((s) => s.layerVisibility.sensorCoverage);

  useEffect(() => {
    const map = (window as any).__RTSA_MAP__;
    if (!map) return; // Wait for map to initialize

    const sourceId = "sensor-coverage";

    // Set up source and layers if they don't exist.
    // We use separate layers per sensor type so each layer can have its own color.
    // However, MapLibre data-driven properties let us do this in a single layer
    // with a "match" expression or a "feature-state" approach.
    // We store sensorType in properties and use a match expression.
    if (!map.getSource(sourceId)) {
      map.addSource(sourceId, {
        type: "geojson",
        data: { type: "FeatureCollection", features: [] },
      });

      map.addLayer({
        id: "sensor-coverage-fill",
        type: "fill",
        source: sourceId,
        paint: {
          // Data-driven color by sensor type
          "fill-color": [
            "match", ["get", "sensorType"],
            "RADAR",  "#16A34A",
            "EW",     "#9333EA",
            "AIS",    "#3B82F6",
            "ELINT",  "#F59E0B",
            "ISR",    "#06B6D4",
            /* default */ "#6B7280",
          ],
          "fill-opacity": 0.15,
        },
      });

      map.addLayer({
        id: "sensor-coverage-outline",
        type: "line",
        source: sourceId,
        paint: {
          // Match outline color to fill
          "line-color": [
            "match", ["get", "sensorType"],
            "RADAR",  "#16A34A",
            "EW",     "#9333EA",
            "AIS",    "#3B82F6",
            "ELINT",  "#F59E0B",
            "ISR",    "#06B6D4",
            /* default */ "#6B7280",
          ],
          "line-width": 1.5,
          "line-dasharray": [3, 2],
          "line-opacity": 0.7,
        },
      });
    }

    // Generate GeoJSON features from coverage responses.
    // Each feature carries "sensorType" so the data-driven paint expression above works.
    const features = sensors.map((sensor) => {
      const cov = sensor.coverage;
      if (!cov) return null;

      const sensorType = (sensor as any).sensorType ?? "UNKNOWN";
      const fillColor = getSensorColor(sensorType);

      // Type 1: Explicit polygon (ISR, Geo-fence)
      if (cov.coveragePolygon && cov.coveragePolygon.length >= 3) {
        return {
          type: "Feature",
          geometry: {
            type: "Polygon",
            coordinates: [cov.coveragePolygon.map((p: any) => [p.longitude, p.latitude])],
          },
          properties: { sensorId: sensor.sensorId, type: "polygon", sensorType, fillColor },
        };
      }

      // Type 2: Radar/sensor fan sector (requires position, range, start/end angles)
      if (cov.sensorPosition && cov.rangeNm !== undefined) {
        const radiusDeg = cov.rangeNm / 60; // Approximate degrees
        const startRad = ((cov.bearingStartDegrees || 0) * Math.PI) / 180;
        const endRad   = ((cov.bearingEndDegrees   || 360) * Math.PI) / 180;
        const points   = 36;

        const centerLon = cov.sensorPosition.longitude;
        const centerLat = cov.sensorPosition.latitude;

        const coords: [number, number][] = [[centerLon, centerLat]];

        const angleDiff = endRad < startRad ? (endRad + 2 * Math.PI) - startRad : endRad - startRad;

        for (let i = 0; i <= points; i++) {
          const angle = startRad + (i / points) * angleDiff;
          const dx = Math.sin(angle) * radiusDeg;
          const dy = Math.cos(angle) * radiusDeg;
          const adjustedDx = dx / Math.cos((centerLat * Math.PI) / 180);
          coords.push([centerLon + adjustedDx, centerLat + dy]);
        }

        coords.push([centerLon, centerLat]); // Close polygon

        return {
          type: "Feature",
          geometry: { type: "Polygon", coordinates: [coords] },
          properties: { sensorId: sensor.sensorId, type: "sector", sensorType, fillColor },
        };
      }

      return null;
    }).filter(Boolean);

    // Update the source data
    const source = map.getSource(sourceId);
    if (source) {
      source.setData({ type: "FeatureCollection", features: features as any });

      // Sync visibility with layer toggle
      const visVal = isVisible ? "visible" : "none";
      if (map.getLayer("sensor-coverage-fill"))   map.setLayoutProperty("sensor-coverage-fill",    "visibility", visVal);
      if (map.getLayer("sensor-coverage-outline")) map.setLayoutProperty("sensor-coverage-outline", "visibility", visVal);
    }
  }, [sensors, isVisible]);

  return null; // Headless component — updates the maplibre instance directly
};
