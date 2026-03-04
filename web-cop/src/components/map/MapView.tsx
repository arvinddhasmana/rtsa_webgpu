// CLASSIFICATION: UNCLASSIFIED
// src/components/map/MapView.tsx

import "maplibre-gl/dist/maplibre-gl.css";
import React, { useEffect, useRef } from "react";
import { useSensorStore } from "../../stores/sensorStore";
import { useTrackStore } from "../../stores/trackStore";
import { useUIStore } from "../../stores/uiStore";
import { LayerControls } from "./LayerControls";
import { MapTooltip } from "./MapTooltip";
import { SensorCoverageLayer } from "./SensorCoverageLayer";

/* ─── Domain Classification ─────────────────────────────────────────────────── */

function getEntityDomain(entityType: string): string {
  const t = entityType.toUpperCase();
  if (t.includes("AIR") || t.includes("AIRCRAFT")) return "AIR";
  if (t.includes("SUB")) return "SUBSURFACE";
  if (t.includes("LAND") || t.includes("VEHICLE")) return "LAND";
  if (t.includes("SURFACE") || t.includes("SHIP") || t.includes("VESSEL")) return "SURFACE";
  return "UNKNOWN";
}

/* ─── Programmatic Domain Icon Generation ───────────────────────────────────── */
// Creates 24×24 canvas icons for each domain+hostileClass combo.
// Style 2 Option A: filled geometric shapes with white outline.

const HOSTILE_COLORS: Record<string, string> = {
  HOSTILE: "#DC2626",
  FRIENDLY: "#2563EB",
  NEUTRAL: "#16A34A",
  UNKNOWN: "#CA8A04",
};

const DOMAINS = ["SURFACE", "AIR", "SUBSURFACE", "LAND", "UNKNOWN"] as const;
const HOSTILE_CLASSES = ["HOSTILE", "FRIENDLY", "NEUTRAL", "UNKNOWN"] as const;

function createDomainIcon(
  domain: string,
  fillColor: string,
  size = 32,
): { width: number; height: number; data: Uint8Array } {
  const canvas = document.createElement("canvas");
  canvas.width = size;
  canvas.height = size;
  const ctx = canvas.getContext("2d")!;
  const cx = size / 2;
  const cy = size / 2;
  const r = size / 2 - 3; // leave room for stroke

  ctx.clearRect(0, 0, size, size);
  ctx.fillStyle = fillColor;
  ctx.strokeStyle = "#FFFFFF";
  ctx.lineWidth = 2;
  ctx.lineJoin = "round";

  switch (domain) {
    case "SURFACE": {
      // Upward triangle ▲
      ctx.beginPath();
      ctx.moveTo(cx, cy - r);
      ctx.lineTo(cx + r, cy + r * 0.8);
      ctx.lineTo(cx - r, cy + r * 0.8);
      ctx.closePath();
      break;
    }
    case "AIR": {
      // Diamond ◇ (rotated square)
      ctx.beginPath();
      ctx.moveTo(cx, cy - r);
      ctx.lineTo(cx + r, cy);
      ctx.lineTo(cx, cy + r);
      ctx.lineTo(cx - r, cy);
      ctx.closePath();
      break;
    }
    case "SUBSURFACE": {
      // Inverted triangle ▽
      ctx.beginPath();
      ctx.moveTo(cx, cy + r);
      ctx.lineTo(cx + r, cy - r * 0.8);
      ctx.lineTo(cx - r, cy - r * 0.8);
      ctx.closePath();
      break;
    }
    case "LAND": {
      // Filled square ■
      const half = r * 0.8;
      ctx.beginPath();
      ctx.rect(cx - half, cy - half, half * 2, half * 2);
      break;
    }
    default: {
      // Circle ● (UNKNOWN)
      ctx.beginPath();
      ctx.arc(cx, cy, r * 0.8, 0, Math.PI * 2);
      break;
    }
  }

  ctx.fill();
  ctx.stroke();

  // MapLibre addImage requires {width, height, data: Uint8Array} format.
  // Using ImageData directly can fail silently in some MapLibre versions.
  const imgData = ctx.getImageData(0, 0, size, size);
  return { width: size, height: size, data: new Uint8Array(imgData.data.buffer) };
}

/** Register all domain×hostileClass icon images on the map */
function registerDomainIcons(map: any): void {
  const size = 32;
  for (const domain of DOMAINS) {
    for (const hostile of HOSTILE_CLASSES) {
      const key = `domain-${domain}-${hostile}`;
      if (map.hasImage(key)) continue;
      const color = HOSTILE_COLORS[hostile] ?? "#6B7280";
      const icon = createDomainIcon(domain, color, size);
      map.addImage(key, icon, { sdf: false });
    }
    // Also add a default-colored version
    const defKey = `domain-${domain}-DEFAULT`;
    if (!map.hasImage(defKey)) {
      const icon = createDomainIcon(domain, "#6B7280", size);
      map.addImage(defKey, icon, { sdf: false });
    }
  }
}

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

  const [hoverInfo, setHoverInfo] = React.useState<{ trackId: string, x: number, y: number } | null>(null);

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
      const map = mapRef.current;
      if (!map) return;

      const currentTracks = useTrackStore.getState().tracks;
      const currentSensors = useSensorStore.getState().rawObservations;

      // ── Track features with domain for icon selection ─────────────────────
      // Build GeoJSON features for all current tracks.
      const trackFeatures = Array.from(currentTracks.values()).map((t) => {
        const domain = getEntityDomain(t.entityType);
        return {
          type: "Feature" as const,
          geometry: {
            type: "Point" as const,
            coordinates: [t.position.longitude, t.position.latitude],
          },
          properties: {
            trackId: t.trackId,
            hostileClass: t.hostileClass,
            entityType: t.entityType,
            domain,
            iconKey: `domain-${domain}-${t.hostileClass}`,
            confidence: t.confidenceScore,
            classification: t.classification,
          },
        };
      });

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

      // ── Raw Sensor Observations ───────────────────────────────────────────────
      const sensorFeatures = Array.from(currentSensors.values()).map((o) => ({
        type: "Feature" as const,
        geometry: {
          type: "Point" as const,
          coordinates: [o.longitude, o.latitude],
        },
        properties: {
          observationId: o.observationId,
          sensorType: o.sensorType,
          isCorrelated: !!o.correlatedTrackId,
        },
      }));

      const sensorsSource = map.getSource("raw-sensors");
      if (sensorsSource) {
        sensorsSource.setData({
          type: "FeatureCollection",
          features: sensorFeatures,
        });
      }
    };

    const initMap = async () => {
      if (!mapContainerRef.current || mapRef.current) return;

      try {
        // maplibre-gl ships as a UMD bundle. Vite's esbuild CJS interop places
        // the maplibregl object on .default — NOT as named exports (maplibregl.Map
        // would be undefined, making `new maplibregl.Map()` throw silently).
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const maplibregl: any = ((await import("maplibre-gl")) as any).default;

        const tileUrl =
          (import.meta as { env?: Record<string, string> }).env?.[
            "VITE_MAP_TILE_URL"
          ] ?? "";

        // Note: no glyphs URL — text symbol layers are omitted so that the
        // map works offline / without a font PBF tile server. Track circles
        // render entirely on the GPU without glyph lookups.
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
                  { id: "base", type: "raster" as const, source: "tiles" },
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

          // ── Track symbols (domain-specific geometric icons) ──────────────
          // Register all domain×hostile canvas icons on the map
          registerDomainIcons(map);

          map.addSource("tracks", {
            type: "geojson",
            data: { type: "FeatureCollection", features: [] },
          });

          // Symbol layer using data-driven icon-image keyed by domain+hostileClass
          map.addLayer({
            id: "tracks-symbol",
            type: "symbol",
            source: "tracks",
            layout: {
              "icon-image": ["get", "iconKey"],
              "icon-size": 1.0,
              "icon-allow-overlap": true,
              "icon-ignore-placement": true,
            },
            paint: {
              "icon-opacity": 0.9,
            },
          });

          // Hover highlight layer — slightly larger, glow stroke
          // We keep a circle layer purely for the hover glow ring.
          map.addLayer({
            id: "tracks-hover-ring",
            type: "circle",
            source: "tracks",
            filter: ["==", "trackId", ""],
            paint: {
              "circle-radius": 14,
              "circle-color": "transparent",
              "circle-stroke-width": 2.5,
              "circle-stroke-color": "#60A5FA",
              "circle-opacity": 0,
              "circle-stroke-opacity": 0.8,
            },
          });

          // Note: tracks-label symbol layer is intentionally omitted —
          // symbol layers require a glyphs/font PBF server which is not
          // available in offline/dev deployments.

          // Click to select track and open detail panel
          map.on("click", "tracks-symbol", (e: any) => {
            const feature = e.features?.[0];
            if (feature?.properties?.trackId) {
              selectTrack(feature.properties.trackId);
              toggleDetailPanel();
            }
          });

          // Pointer cursor and tooltip on hover
          map.on("mousemove", "tracks-symbol", (e: any) => {
            map.getCanvas().style.cursor = "pointer";
            const feature = e.features?.[0];
            if (feature?.properties?.trackId) {
              const trackId = feature.properties.trackId;

              // Only update state if hover changes
              setHoverInfo(prev => {
                if (prev?.trackId !== trackId) {
                  // Update hover ring filter
                  map.setFilter("tracks-hover-ring", ["==", "trackId", trackId]);
                }
                return { trackId, x: e.point.x, y: e.point.y };
              });
            }
          });

          map.on("mouseleave", "tracks-symbol", () => {
            map.getCanvas().style.cursor = "";
            map.setFilter("tracks-hover-ring", ["==", "trackId", ""]);
            setHoverInfo(null);
          });

          // ── Raw Sensor Observations (Uncorrelated vs Correlated) ──────────
          map.addSource("raw-sensors", {
            type: "geojson",
            data: { type: "FeatureCollection", features: [] },
          });

          map.addLayer({
            id: "raw-sensors-circle",
            type: "circle",
            source: "raw-sensors",
            paint: {
              "circle-radius": 3,
              // Faded blue for correlated, bright cyan for uncorrelated
              "circle-color": [
                "case",
                ["get", "isCorrelated"],
                "#3B82F6",
                "#06B6D4",
              ],
              "circle-opacity": ["case", ["get", "isCorrelated"], 0.3, 0.8],
              "circle-stroke-width": 0,
            },
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

          // Subscribe to stores — RAF-throttled, one GPU upload per frame max.
          unsubscribeStore = useTrackStore.subscribe(scheduleMapUpdate);
          useSensorStore.subscribe(scheduleMapUpdate); // Also trigger redraws when sensors update

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
        map.setLayoutProperty(
          layerId,
          "visibility",
          visible ? "visible" : "none",
        );
      }
    };

    setVis("geofences-fill", layerVisibility.geofences);
    setVis("geofences-layer", layerVisibility.geofences);
    // Track labels layer not yet implemented (requires glyph server)
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
      {hoverInfo && (
        <MapTooltip trackId={hoverInfo.trackId} x={hoverInfo.x} y={hoverInfo.y} />
      )}
      <SensorCoverageLayer />
      <LayerControls />
    </div>
  );
};
