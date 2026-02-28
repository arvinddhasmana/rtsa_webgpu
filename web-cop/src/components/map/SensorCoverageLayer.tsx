// CLASSIFICATION: UNCLASSIFIED
// src/components/map/SensorCoverageLayer.tsx

import React, { useEffect } from "react";
import { useSensorCoverage } from "../../hooks/useSensorCoverage";
import { useUIStore } from "../../stores/uiStore";

// The MapView component exposes the map instance on window.__RTSA_MAP__
// We use a separate functional component here to strictly manage the coverage layer.
export const SensorCoverageLayer: React.FC = () => {
  const sensors = useSensorCoverage();
  const isVisible = useUIStore((s) => s.layerVisibility.sensorCoverage);

  useEffect(() => {
    const map = (window as any).__RTSA_MAP__;
    if (!map) return; // Wait for map to initialize

    const sourceId = "sensor-coverage";

    // Set up source and layers if they don't exist
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
          "fill-color": "#3B82F6", // Blue tint for coverage areas
          "fill-opacity": 0.15,
        },
      });

      map.addLayer({
        id: "sensor-coverage-outline",
        type: "line",
        source: sourceId,
        paint: {
          "line-color": "#60A5FA",
          "line-width": 1,
          "line-dasharray": [2, 2],
        },
      });
    }

    // Generate GeoJSON features from coverage responses
    const features = sensors.map((sensor) => {
      const cov = sensor.coverage;
      if (!cov) return null;

      // Type 1: Explicit polygon (ISR, Geo-fence)
      if (cov.coveragePolygon && cov.coveragePolygon.length >= 3) {
        return {
          type: "Feature",
          geometry: {
            type: "Polygon",
            coordinates: [cov.coveragePolygon.map(p => [p.longitude, p.latitude])]
          },
          properties: { sensorId: sensor.sensorId, type: "polygon" }
        };
      }

      // Type 2: Radar fan sector (requires position, range, start/end angles)
      if (cov.sensorPosition && cov.rangeNm !== undefined) {
        const radiusDeg = cov.rangeNm / 60; // Approximate degrees
        const startRad = ((cov.bearingStartDegrees || 0) * Math.PI) / 180;
        const endRad = ((cov.bearingEndDegrees || 360) * Math.PI) / 180;
        const points = 32;

        const centerLon = cov.sensorPosition.longitude;
        const centerLat = cov.sensorPosition.latitude;

        let coords: [number, number][] = [[centerLon, centerLat]];

        // Generate arc points
        const angleDiff = endRad < startRad ? (endRad + 2*Math.PI) - startRad : endRad - startRad;

        for (let i = 0; i <= points; i++) {
          const angle = startRad + (i / points) * angleDiff;
          const dx = Math.sin(angle) * radiusDeg; // Bearing is clockwise from North (Y-axis)
          const dy = Math.cos(angle) * radiusDeg;

          const adjustedDx = dx / Math.cos(centerLat * Math.PI / 180);
          coords.push([centerLon + adjustedDx, centerLat + dy]);
        }

        coords.push([centerLon, centerLat]); // Close polygon by returning to center

        return {
          type: "Feature",
          geometry: {
            type: "Polygon",
            coordinates: [coords]
          },
          properties: { sensorId: sensor.sensorId, type: "sector" }
        };
      }

      return null;
    }).filter(Boolean);

    // Update the source data
    const source = map.getSource(sourceId);
    if (source) {
      source.setData({
        type: "FeatureCollection",
        features: features as any,
      });

      // Toggle visibility based on UI state
      const visVal = isVisible ? "visible" : "none";
      map.setLayoutProperty("sensor-coverage-fill", "visibility", visVal);
      map.setLayoutProperty("sensor-coverage-outline", "visibility", visVal);
    }
  }, [sensors, isVisible]);

  return null; // This is a headless component that updates the maplibre instance
};
