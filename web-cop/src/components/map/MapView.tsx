// CLASSIFICATION: UNCLASSIFIED
// src/components/map/MapView.tsx

import "maplibre-gl/dist/maplibre-gl.css";
import React, { useEffect, useRef } from "react";
import { useTrackStore } from "../../stores/trackStore";
import { useUIStore } from "../../stores/uiStore";

/**
 * MapView — main map display using MapLibre GL JS.
 *
 * Performance design:
 *   - Does NOT subscribe to `tracks` via React state to avoid re-renders on every stream message.
 *   - Instead uses useTrackStore.subscribe() inside the map-init useEffect and throttles
 *     calls to updateMapData() with requestAnimationFrame (≤ 60fps, typically 30fps).
 *   - maplibre-gl is dynamically imported once and cached in maplibreglRef.
 *   - Only `trackCount` (tracks.size) triggers a React re-render, purely for the
 *     "No tracks" overlay (re-renders only when count crosses zero ↔ non-zero).
 */
export const MapView: React.FC = () => {
  const mapContainerRef = useRef<HTMLDivElement>(null);
  const mapRef = useRef<any>(null);
  const markersRef = useRef<Map<string, any>>(new Map());
  // Cache the maplibre-gl module after the first import — avoids repeated dynamic imports.
  const maplibreglRef = useRef<any>(null);

  const mapCenter = useUIStore((s) => s.mapCenter);
  const mapZoom = useUIStore((s) => s.mapZoom);
  // Lightweight selector: only triggers re-render when count crosses 0 ↔ non-zero.
  const trackCount = useTrackStore((s) => s.tracks.size);
  const selectTrack = useTrackStore((s) => s.selectTrack);
  const toggleDetailPanel = useUIStore((s) => s.toggleDetailPanel);

  useEffect(() => {
    let isMounted = true;
    // unsubscribeStore and rafId are captured in the closure so the cleanup
    // function can cancel them even if assigned asynchronously inside map.on("load").
    let unsubscribeStore: (() => void) | null = null;
    let rafId: number | null = null;
    let updateScheduled = false;

    // Called by the Zustand subscriber; rate-limited to one update per animation frame.
    const scheduleMapUpdate = () => {
      if (!updateScheduled && mapRef.current && maplibreglRef.current) {
        updateScheduled = true;
        rafId = requestAnimationFrame(() => {
          updateScheduled = false;
          rafId = null;
          updateMapData();
        });
      }
    };

    const initMap = async () => {
      if (!mapContainerRef.current || mapRef.current) return;

      try {
        const maplibregl = await import("maplibre-gl");

        const tileUrl =
          (import.meta as { env?: Record<string, string> }).env?.[
            "VITE_MAP_TILE_URL"
          ] ?? "";

        const map = new maplibregl.Map({
          container: mapContainerRef.current,
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

        map.on("load", () => {
          if (!isMounted) return;

          // Add sources and layers for halos and geofences
          map.addSource("threat-halos", {
            type: "geojson",
            data: { type: "FeatureCollection", features: [] },
          });

          map.addLayer({
            id: "threat-halos-layer",
            type: "fill",
            source: "threat-halos",
            paint: {
              "fill-color": "#DC2626",
              "fill-opacity": 0.2,
            },
          });

          map.addLayer({
            id: "threat-halos-outline",
            type: "line",
            source: "threat-halos",
            paint: {
              "line-color": "#DC2626",
              "line-width": 2,
              "line-dasharray": [2, 2],
            },
          });

          // Add geofences source
          map.addSource("geofences", {
            type: "geojson",
            data: {
              type: "FeatureCollection",
              features: [
                {
                  type: "Feature",
                  properties: { type: "exclusion" },
                  geometry: {
                    type: "Polygon",
                    coordinates: [
                      [
                        [-62.0, 44.0],
                        [-60.0, 44.0],
                        [-60.0, 46.0],
                        [-62.0, 46.0],
                        [-62.0, 44.0],
                      ],
                    ],
                  },
                },
              ],
            },
          });

          map.addLayer({
            id: "geofences-layer",
            type: "line",
            source: "geofences",
            paint: {
              "line-color": [
                "match",
                ["get", "type"],
                "exclusion",
                "#DC2626",
                "inclusion",
                "#16A34A",
                "#6B7280",
              ],
              "line-width": 2,
            },
          });

          map.addLayer({
            id: "geofences-fill",
            type: "fill",
            source: "geofences",
            paint: {
              "fill-color": [
                "match",
                ["get", "type"],
                "exclusion",
                "#DC2626",
                "inclusion",
                "#16A34A",
                "#6B7280",
              ],
              "fill-opacity": 0.1,
            },
          });

          mapRef.current = map;

          // Cache maplibregl module — updateMapData uses this ref instead of
          // re-importing on every track update.
          maplibreglRef.current = maplibregl;

          // Expose map instance for E2E testing and devtools inspection
          (window as unknown as Record<string, unknown>)["__RTSA_MAP__"] = map;

          // Subscribe to track store — schedule a RAF-throttled map redraw
          // on any store change. Coalesces rapid burst updates into one draw.
          unsubscribeStore = useTrackStore.subscribe(scheduleMapUpdate);

          updateMapData();
        });
      } catch (e) {
        console.error("Failed to initialize map", e);
      }
    };

    void initMap();

    return () => {
      isMounted = false;
      // Cancel any pending RAF before unsubscribing to avoid a ghost update.
      if (rafId !== null) {
        cancelAnimationFrame(rafId);
        rafId = null;
      }
      unsubscribeStore?.();
      maplibreglRef.current = null;
      if (mapRef.current) {
        (window as unknown as Record<string, unknown>)["__RTSA_MAP__"] =
          undefined;
        mapRef.current.remove();
        mapRef.current = null;
      }
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Synchronous: uses pre-cached maplibreglRef — no dynamic import overhead per call.
  const updateMapData = () => {
    if (!mapRef.current || !maplibreglRef.current) return;
    const map = mapRef.current;
    const maplibregl = maplibreglRef.current;

    // Always read freshest state from store — avoids React closure staleness.
    const currentTracks = useTrackStore.getState().tracks;

    // Update markers
    const currentTrackIds = new Set<string>();

    for (const track of currentTracks.values()) {
      currentTrackIds.add(track.trackId);

      let marker = markersRef.current.get(track.trackId);
      if (!marker) {
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

        marker = new maplibregl.Marker({ element: el })
          .setLngLat([track.position.longitude, track.position.latitude])
          .addTo(map);

        markersRef.current.set(track.trackId, marker);
      } else {
        marker.setLngLat([track.position.longitude, track.position.latitude]);
        // Update color if changed
        marker.getElement().style.backgroundColor = getTrackColor(
          track.hostileClass,
        );
      }
    }

    // Remove stale markers
    for (const [id, marker] of markersRef.current.entries()) {
      if (!currentTrackIds.has(id)) {
        marker.remove();
        markersRef.current.delete(id);
      }
    }

    // Update threat halos (simple circle approximation for hostile tracks)
    const haloFeatures = Array.from(currentTracks.values())
      .filter((t) => t.hostileClass === "HOSTILE")
      .map((t) => {
        // Create a rough circle polygon around the point (approx 50km radius)
        const points = 32;
        const radiusInDeg = 50 / 111.32; // rough approx
        const coords = [];
        for (let i = 0; i <= points; i++) {
          const angle = (i / points) * Math.PI * 2;
          const dx = Math.cos(angle) * radiusInDeg;
          const dy = Math.sin(angle) * radiusInDeg;
          // Adjust dx for latitude
          const adjustedDx =
            dx / Math.cos((t.position.latitude * Math.PI) / 180);
          coords.push([
            t.position.longitude + adjustedDx,
            t.position.latitude + dy,
          ]);
        }

        return {
          type: "Feature",
          geometry: {
            type: "Polygon",
            coordinates: [coords],
          },
          properties: { trackId: t.trackId },
        };
      });

    const haloSource = map.getSource("threat-halos");
    if (haloSource) {
      haloSource.setData({
        type: "FeatureCollection",
        features: haloFeatures as any,
      });
    }
  };

  return (
    <div
      data-testid="map-container"
      ref={mapContainerRef}
      style={{
        width: "100%",
        height: "100%",
        backgroundColor: "#1E293B",
        position: "relative",
      }}
    >
      {trackCount === 0 && (
        <div
          style={{
            position: "absolute",
            top: "50%",
            left: "50%",
            transform: "translate(-50%, -50%)",
            color: "#9CA3AF",
            fontSize: "0.875rem",
            zIndex: 10,
            pointerEvents: "none",
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
