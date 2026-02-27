// CLASSIFICATION: UNCLASSIFIED
// src/components/map/MapView.tsx

import "maplibre-gl/dist/maplibre-gl.css";
import React, { useEffect, useRef } from "react";
import { useTrackStore } from "../../stores/trackStore";
import { useUIStore } from "../../stores/uiStore";
import { LayerControls } from "./LayerControls";

/**
 * MapView — real-time track display using MapLibre GL JS.
 *
 * Performance design:
 *   - Tracks are rendered as a native GeoJSON circle layer (WebGL / GPU).
 *     A single setData() call replaces the entire track set in one GPU upload.
 *     No DOM elements are created per track — scales to thousands of tracks at 60fps.
 *   - Does NOT subscribe to `tracks` via React state.
 *     Uses useTrackStore.subscribe() + requestAnimationFrame throttling so that
 *     rapid stream bursts (large snapshots) collapse into ≤ 1 redraw per frame.
 *   - maplibre-gl is dynamically imported once and cached in maplibreglRef.
 *   - Only `trackCount` (tracks.size crossing 0 ↔ non-zero) causes a React re-render,
 *     purely to control the "No tracks" overlay.
 */
export const MapView: React.FC = () => {
  const mapContainerRef = useRef<HTMLDivElement>(null);
  const mapRef = useRef<any>(null);
  // Cache the maplibre-gl module after the first import.
  const maplibreglRef = useRef<any>(null);

  const mapCenter = useUIStore((s) => s.mapCenter);
  const mapZoom = useUIStore((s) => s.mapZoom);
  const layerVisibility = useUIStore((s) => s.layerVisibility);
  // Only re-renders when the count crosses zero — avoids per-track React cycles.
  const trackCount = useTrackStore((s) => s.tracks.size);
  const selectTrack = useTrackStore((s) => s.selectTrack);
  const toggleDetailPanel = useUIStore((s) => s.toggleDetailPanel);

  useEffect(() => {
    let isMounted = true;
    let unsubscribeStore: (() => void) | null = null;
    let rafId: number | null = null;
    let updateScheduled = false;

    // RAF-throttled trigger — coalesces burst updates into one GPU upload per frame.
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

    // Synchronous map update — reads freshest store state, uploads one GeoJSON blob.
    // No DOM element creation. All rendering happens on the GPU via MapLibre GL.
    const updateMapData = () => {
      if (!mapRef.current) return;
      const map = mapRef.current;
      const currentTracks = useTrackStore.getState().tracks;

      // ── Track circles ────────────────────────────────────────────────────────
      // Build GeoJSON features for all current tracks.
      const trackFeatures = Array.from(currentTracks.values()).map((t) => ({
        type: "Feature" as const,
        geometry: {
          type: "Point" as const,
          coordinates: [t.position.longitude, t.position.latitude],
        },
        properties: {
          trackId: t.trackId,
          hostileClass: t.hostileClass,
          entityType: t.entityType,
          confidence: t.confidenceScore,
          classification: t.classification,
        },
      }));

      const tracksSource = map.getSource("tracks");
      if (tracksSource) {
        tracksSource.setData({
          type: "FeatureCollection",
          features: trackFeatures,
        });
      }

      // ── Threat halos (50 km circle polygons for HOSTILE tracks) ───────────
      const haloFeatures = Array.from(currentTracks.values())
        .filter((t) => t.hostileClass === "HOSTILE")
        .map((t) => {
          const points = 32;
          const radiusInDeg = 50 / 111.32;
          const coords = [];
          for (let i = 0; i <= points; i++) {
            const angle = (i / points) * Math.PI * 2;
            const dx = Math.cos(angle) * radiusInDeg;
            const dy = Math.sin(angle) * radiusInDeg;
            const adjustedDx =
              dx / Math.cos((t.position.latitude * Math.PI) / 180);
            coords.push([
              t.position.longitude + adjustedDx,
              t.position.latitude + dy,
            ]);
          }
          return {
            type: "Feature" as const,
            geometry: { type: "Polygon" as const, coordinates: [coords] },
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
                glyphs: "https://demotiles.maplibre.org/font/{fontstack}/{range}.pbf",
                sources: {
                  tiles: {
                    type: "raster" as const,
                    tiles: [tileUrl],
                    tileSize: 256,
                  },
                },
                layers: [
                  { id: "base", type: "raster" as const, source: "tiles" },
                ],
              }
              : {
                version: 8 as const,
                glyphs: "https://demotiles.maplibre.org/font/{fontstack}/{range}.pbf",
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

          // ── Geo-fences ────────────────────────────────────────────────────
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

          // ── Threat halos ─────────────────────────────────────────────────
          map.addSource("threat-halos", {
            type: "geojson",
            data: { type: "FeatureCollection", features: [] },
          });

          map.addLayer({
            id: "threat-halos-layer",
            type: "fill",
            source: "threat-halos",
            paint: { "fill-color": "#DC2626", "fill-opacity": 0.2 },
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

          // ── Track circles (GPU rendered, zero DOM overhead) ──────────────
          map.addSource("tracks", {
            type: "geojson",
            data: { type: "FeatureCollection", features: [] },
          });

          // Data-driven color by hostileClass property — evaluated on the GPU.
          map.addLayer({
            id: "tracks-circle",
            type: "circle",
            source: "tracks",
            paint: {
              "circle-radius": 6,
              "circle-color": [
                "match",
                ["get", "hostileClass"],
                "HOSTILE",
                "#DC2626",
                "FRIENDLY",
                "#2563EB",
                "NEUTRAL",
                "#16A34A",
                "UNKNOWN",
                "#CA8A04",
                /* default */ "#6B7280",
              ],
              "circle-stroke-width": 1.5,
              "circle-stroke-color": "#FFFFFF",
              "circle-opacity": 0.9,
            },
          });

          // Classification label (Badge) — evaluated on GPU
          map.addLayer({
            id: "tracks-label",
            type: "symbol",
            source: "tracks",
            layout: {
              "text-field": ["get", "classification"],
              "text-font": ["Open Sans Semibold", "Arial Unicode MS Bold"],
              "text-size": 10,
              "text-offset": [0, 1.2],
              "text-anchor": "top",
              "visibility": "visible",
            },
            paint: {
              "text-color": [
                "match",
                ["get", "classification"],
                "UNCLASSIFIED",
                "#16A34A",
                "PROTECTED_A",
                "#2563EB",
                "PROTECTED_B",
                "#2563EB",
                "PROTECTED_C",
                "#2563EB",
                "SECRET",
                "#DC2626",
                /* default */ "#F1F5F9",
              ],
              "text-halo-color": "#0F172A",
              "text-halo-width": 1.5,
            },
          });

          // Click to select track and open detail panel
          map.on("click", "tracks-circle", (e: any) => {
            const feature = e.features?.[0];
            if (feature?.properties?.trackId) {
              selectTrack(feature.properties.trackId);
              toggleDetailPanel();
            }
          });

          // Pointer cursor on hover to indicate interactivity
          map.on("mouseenter", "tracks-circle", () => {
            map.getCanvas().style.cursor = "pointer";
          });
          map.on("mouseleave", "tracks-circle", () => {
            map.getCanvas().style.cursor = "";
          });

          // ── Forensics Replay layer (MapReplay writes here) ────────────────
          // Separate source from live tracks so replay and live data coexist.
          // Replay circles have a yellow stroke to visually distinguish them.
          map.addSource("replay-tracks", {
            type: "geojson",
            data: { type: "FeatureCollection", features: [] },
          });

          map.addLayer({
            id: "replay-tracks-circle",
            type: "circle",
            source: "replay-tracks",
            paint: {
              "circle-radius": 9,
              "circle-color": [
                "match",
                ["get", "hostileClass"],
                "HOSTILE",
                "#DC2626",
                "FRIENDLY",
                "#2563EB",
                "NEUTRAL",
                "#16A34A",
                "UNKNOWN",
                "#CA8A04",
                /* default */ "#6B7280",
              ],
              // Yellow stroke — visually distinguishes replay from live tracks
              "circle-stroke-width": 2.5,
              "circle-stroke-color": "#FACC15",
              "circle-opacity": 0.8,
            },
          });

          mapRef.current = map;
          maplibreglRef.current = maplibregl;

          // Expose for E2E testing and devtools
          (window as unknown as Record<string, unknown>)["__RTSA_MAP__"] = map;

          // Subscribe to store — RAF-throttled, one GPU upload per frame max.
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

  // ── Layer Visibility Toggles ──────────────────────────────────────────────
  useEffect(() => {
    if (!mapRef.current) return;
    const map = mapRef.current;

    const setVis = (layerId: string, visible: boolean) => {
      if (map.getLayer(layerId)) {
        map.setLayoutProperty(layerId, "visibility", visible ? "visible" : "none");
      }
    };

    setVis("geofences-fill", layerVisibility.geofences);
    setVis("geofences-layer", layerVisibility.geofences);
    setVis("tracks-label", layerVisibility.trackLabels);
    // Note: trackTrails, sensorCoverage, mgrsGrid placeholders
    // Set up their logic if those layers are added.
  }, [layerVisibility]);

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
      <LayerControls />
    </div>
  );
};
