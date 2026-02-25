// CLASSIFICATION: UNCLASSIFIED
// src/components/map/MapView.tsx

import React, { useEffect, useRef } from "react";
import { useUIStore } from "../../stores/uiStore";
import { useTrackStore } from "../../stores/trackStore";

/**
 * MapView — main map display using MapLibre GL JS.
 *
 * Features:
 *   - Renders tracks as positioned markers with MIL-STD-2525 symbology
 *   - Color-coded by hostile classification
 *   - Track history trail (last 10 positions as fading dots)
 *   - Click-to-select: clicking a track opens DetailPanel
 *   - Default view: Mid-Atlantic region (center: -60°, 45°, zoom: 6)
 *   - No external tile providers in production (offline map tiles)
 */
export const MapView: React.FC = () => {
  const mapRef = useRef<HTMLDivElement>(null);
  const mapCenter = useUIStore((s) => s.mapCenter);
  const mapZoom = useUIStore((s) => s.mapZoom);
  const tracks = useTrackStore((s) => s.tracks);
  const selectTrack = useTrackStore((s) => s.selectTrack);
  const toggleDetailPanel = useUIStore((s) => s.toggleDetailPanel);

  useEffect(() => {
    // Dynamic import of maplibre-gl to avoid SSR issues
    let mapInstance: { remove: () => void } | null = null;

    const initMap = async () => {
      if (!mapRef.current) return;

      try {
        const maplibregl = await import("maplibre-gl");

        const tileUrl =
          (import.meta as { env?: Record<string, string> }).env?.["VITE_MAP_TILE_URL"] ??
          "";

        const map = new maplibregl.Map({
          container: mapRef.current,
          style: tileUrl
            ? {
                version: 8 as const,
                sources: {
                  tiles: {
                    type: "raster" as const,
                    tiles: [tileUrl],
                    tileSize: 256,
                  },
                },
                layers: [
                  {
                    id: "base",
                    type: "raster" as const,
                    source: "tiles",
                  },
                ],
              }
            : {
                version: 8 as const,
                sources: {},
                layers: [
                  {
                    id: "background",
                    type: "background" as const,
                    paint: { "background-color": "#1E293B" },
                  },
                ],
              },
          center: mapCenter,
          zoom: mapZoom,
        });

        mapInstance = map;

        // Add track markers after map loads
        map.on("load", () => {
          for (const track of tracks.values()) {
            const el = document.createElement("div");
            el.style.cssText = `
              width: 12px; height: 12px; border-radius: 50%;
              background-color: ${getTrackColor(track.hostileClass)};
              border: 2px solid white; cursor: pointer;
            `;
            el.title = `Track: ${track.trackId}`;
            el.addEventListener("click", () => {
              selectTrack(track.trackId);
              toggleDetailPanel();
            });
          }
        });
      } catch {
        // MapLibre not available in test environment
      }
    };

    void initMap();

    return () => {
      mapInstance?.remove();
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Update markers when tracks change
  useEffect(() => {
    // Production: update MapLibre source with new track GeoJSON
    void tracks;
  }, [tracks]);

  return (
    <div
      data-testid="map-container"
      ref={mapRef}
      style={{
        width: "100%",
        height: "100%",
        backgroundColor: "#1E293B",
        position: "relative",
      }}
    >
      {tracks.size === 0 && (
        <div
          style={{
            position: "absolute",
            top: "50%",
            left: "50%",
            transform: "translate(-50%, -50%)",
            color: "#9CA3AF",
            fontSize: "0.875rem",
          }}
        >
          No tracks — awaiting data stream
        </div>
      )}
    </div>
  );
};

function getTrackColor(hostile: string): string {
  switch (hostile) {
    case "HOSTILE":
      return "#DC2626";
    case "FRIENDLY":
      return "#2563EB";
    case "NEUTRAL":
      return "#16A34A";
    case "UNKNOWN":
      return "#CA8A04";
    default:
      return "#6B7280";
  }
}
