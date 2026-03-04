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
import { TrackLabelsOverlay } from "./TrackLabelsOverlay";

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
  strokeColor: string,
  size = 40,
): { width: number; height: number; data: Uint8Array } {
  const canvas = document.createElement("canvas");
  // Double pixel ratio for crisp render on high-DPI screens
  const scale = 2;
  canvas.width = size * scale;
  canvas.height = size * scale;
  const ctx = canvas.getContext("2d")!;
  ctx.scale(scale, scale);
  const cx = size / 2;
  const cy = size / 2;
  const r = size / 2 - 4; // leave room for thick stroke

  ctx.clearRect(0, 0, size, size);

  // Glass look: very subtle translucent fill + sharp colored outline
  // Parse hex color to rgba with 12% opacity for fill
  const hexToRgba = (hex: string, a: number) => {
    const n = parseInt(hex.slice(1), 16);
    return `rgba(${(n >> 16) & 255},${(n >> 8) & 255},${n & 255},${a})`;
  };

  ctx.strokeStyle = strokeColor;
  ctx.lineWidth = 0.5;   // Extra thin edge — barely visible, glassmorphism style
  ctx.lineJoin = "round";

  switch (domain) {
    case "SURFACE": {
      // Upward triangle ▲
      ctx.beginPath();
      ctx.moveTo(cx, cy - r);
      ctx.lineTo(cx + r, cy + r * 0.78);
      ctx.lineTo(cx - r, cy + r * 0.78);
      ctx.closePath();
      break;
    }
    case "AIR": {
      // Diamond ◇ (rotated square)
      ctx.beginPath();
      ctx.moveTo(cx, cy - r);
      ctx.lineTo(cx + r * 0.85, cy);
      ctx.lineTo(cx, cy + r);
      ctx.lineTo(cx - r * 0.85, cy);
      ctx.closePath();
      break;
    }
    case "SUBSURFACE": {
      // Inverted triangle ▽
      ctx.beginPath();
      ctx.moveTo(cx, cy + r);
      ctx.lineTo(cx + r, cy - r * 0.78);
      ctx.lineTo(cx - r, cy - r * 0.78);
      ctx.closePath();
      break;
    }
    case "LAND": {
      // Square ■
      const half = r * 0.78;
      ctx.beginPath();
      ctx.rect(cx - half, cy - half, half * 2, half * 2);
      break;
    }
    default: {
      // Circle ● (UNKNOWN/CYBER/SPACE)
      ctx.beginPath();
      ctx.arc(cx, cy, r * 0.82, 0, Math.PI * 2);
      break;
    }
  }

  // Draw near-invisible translucent fill (1.5% opacity), then extra-thin outline on top.
  // Net result: icon barely tints the map whilst remaining precisely locatable.
  ctx.fillStyle = hexToRgba(strokeColor, 0.015);
  ctx.fill();
  ctx.stroke();

  // MapLibre addImage requires {width, height, data: Uint8Array} — use
  // actual canvas pixel size (scaled), then pass logical size to MapLibre.
  const imgData = ctx.getImageData(0, 0, size * scale, size * scale);
  return { width: size * scale, height: size * scale, data: new Uint8Array(imgData.data.buffer) };
}

/** Register all domain×hostileClass icon images on the map */
function registerDomainIcons(map: any): void {
  const size = 40;
  for (const domain of DOMAINS) {
    for (const hostile of HOSTILE_CLASSES) {
      const key = `domain-${domain}-${hostile}`;
      if (map.hasImage(key)) continue;
      const color = HOSTILE_COLORS[hostile] ?? "#6B7280";
      const icon = createDomainIcon(domain, color, size);
      // pixelRatio:2 matches the 2x canvas scale in createDomainIcon
      map.addImage(key, icon, { sdf: false, pixelRatio: 2 });
    }
    // Also add a default-colored version
    const defKey = `domain-${domain}-DEFAULT`;
    if (!map.hasImage(defKey)) {
      const icon = createDomainIcon(domain, "#6B7280", size);
      map.addImage(defKey, icon, { sdf: false, pixelRatio: 2 });
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
    let lastUpdateTime = 0;
    // Hard cap: max 10 map updates/sec regardless of how fast the store fires.
    // At 15k tracks, a full GeoJSON rebuild is expensive; more than 10Hz adds no
    // visual benefit since humans can't perceive faster than ~15fps for slow-moving
    // tactical symbols.
    const UPDATE_INTERVAL_MS = 100; // 10 Hz

    const scheduleMapUpdate = () => {
      if (rafId !== null || !mapRef.current || !maplibreglRef.current) return;
      rafId = requestAnimationFrame((now) => {
        rafId = null;
        if (!isMounted) return;
        // Throttle: skip frames that arrive too fast after the last real update
        if (now - lastUpdateTime < UPDATE_INTERVAL_MS) {
          // Re-schedule after remaining interval
          const delay = UPDATE_INTERVAL_MS - (now - lastUpdateTime);
          setTimeout(scheduleMapUpdate, delay);
          return;
        }
        lastUpdateTime = now;
        updateMapData();
      });
    };

    // Synchronous map update — reads freshest store state, uploads one GeoJSON blob.
    // No DOM element creation. All rendering happens on the GPU via MapLibre GL.
    const updateMapData = () => {
      const map = mapRef.current;
      if (!map) return;

      const { tracks: currentTracks, trackHistory } = useTrackStore.getState();
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

      // ── C: Track Trail Breadcrumbs ─────────────────────────────────────────
      // Build LineString features from stored position history per track.
      // Trails with < 2 points are skipped (can't form a line).
      const HOSTILE_TRAIL_COLORS: Record<string, string> = {
        HOSTILE:  "#DC2626",
        FRIENDLY: "#2563EB",
        NEUTRAL:  "#16A34A",
        UNKNOWN:  "#CA8A04",
      };
      // Max trail points: older positions are discarded to prevent stale
      // artifacts building up over time (white smear on dense scenes).
      const MAX_TRAIL_POINTS = 20;
      const trailFeatures: any[] = [];
      for (const t of currentTracks.values()) {
        const rawHistory = trackHistory.get(t.trackId);
        if (!rawHistory || rawHistory.length < 2) continue;
        // Trim to last MAX_TRAIL_POINTS so old segments fade off the map
        const history = rawHistory.length > MAX_TRAIL_POINTS
          ? rawHistory.slice(-MAX_TRAIL_POINTS)
          : rawHistory;
        trailFeatures.push({
          type: "Feature" as const,
          geometry: { type: "LineString" as const, coordinates: history },
          properties: {
            trackId: t.trackId,
            trailColor: HOSTILE_TRAIL_COLORS[t.hostileClass] ?? "#6B7280",
          },
        });
      }
      const trailSource = map.getSource("track-trails");
      if (trailSource) {
        trailSource.setData({ type: "FeatureCollection", features: trailFeatures });
      }

      const tracksSource = map.getSource("tracks");
      if (tracksSource) {
        tracksSource.setData({
          type: "FeatureCollection",
          features: trackFeatures,
        });
      }

      // ── Threat halos ──────────────────────────────────────────────────────
      // NOTE: Halo polygon computation is intentionally skipped here.
      // With 1000+ hostile tracks, computing 16-point ring polygons per track
      // per frame (1000 × 16 trig ops × 10 Hz) pegs the main thread CPU.
      // Hostile status is already communicated by the red icon color.
      // The halo source/layer still exists for future use (e.g., on track click);
      // we just don't rebuild it every frame.
      //
      // To show halos only for selected/focused tracks, set halo data when
      // a track is selected (handled in click handler below).

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

        // Use VITE_MAP_TILE_URL if configured; fall back to OpenStreetMap tiles
        // so the map terrain is always visible even in offline/dev deployments.
        const tileUrl =
          (import.meta as { env?: Record<string, string> }).env?.[
            "VITE_MAP_TILE_URL"
          ] ?? "https://tile.openstreetmap.org/{z}/{x}/{y}.png";

        const map = new maplibregl.Map({
          container: mapContainerRef.current,
          style: {
            version: 8 as const,
            sources: {
              tiles: {
                type: "raster" as const,
                tiles: [tileUrl],
                tileSize: 256,
                attribution: tileUrl.includes("openstreetmap")
                  ? '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
                  : "",
              },
            },
            layers: [
              { id: "base", type: "raster" as const, source: "tiles", paint: { "raster-opacity": 0.85 } },
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
                  properties: {
                    type: "exclusion",
                    name: "EXCL-01",
                    description: "Maritime Exclusion Zone Alpha — No entry without authorization",
                  },
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

          // G: Geo-fence — semi-transparent fill + dashed stroke
          map.addLayer({
            id: "geofences-fill",
            type: "fill",
            source: "geofences",
            paint: {
              "fill-color": [
                "match", ["get", "type"],
                "exclusion", "#DC2626",
                "inclusion", "#16A34A",
                /* default */ "#6B7280",
              ],
              // Subtle fill — dashed outline is the primary visual indicator
              "fill-opacity": 0.05,
            },
          });

          map.addLayer({
            id: "geofences-layer",
            type: "line",
            source: "geofences",
            paint: {
              "line-color": [
                "match", ["get", "type"],
                "exclusion", "#F87171",   // red-400 — slightly lighter than fill
                "inclusion", "#4ADE80",   // green-400
                /* default */ "#94A3B8",  // slate-400
              ],
              "line-width": 2.5,
              // G: dashed stroke for a proper geo-fence appearance
              "line-dasharray": [5, 3],
            },
          });

          // Show geo-fence zone name popup on hover (click handler below)
          map.on("click", "geofences-fill", (e: any) => {
            const feature = e.features?.[0];
            if (!feature) return;
            const name = feature.properties?.name ?? "Exclusion Zone";
            const desc = feature.properties?.description ?? "Restricted area";
            // Show a temporary popup
            new (maplibregl as any).Popup({ closeOnClick: true, className: "rtsa-fence-popup" })
              .setLngLat(e.lngLat)
              .setHTML(`<div style="font:11px/1.4 monospace;color:#F87171;">
                <strong>${name}</strong><br/><span style="color:#94A3B8;">${desc}</span>
              </div>`)
              .addTo(map);
          });
          map.on("mouseenter", "geofences-fill", () => { map.getCanvas().style.cursor = "crosshair"; });
          map.on("mouseleave", "geofences-fill", () => { map.getCanvas().style.cursor = ""; });
          // Fill layer intentionally removed — solid red fill at 0.2 opacity
          // was covering the entire map with 700+ hostile tracks.
          // We keep only a thin dashed outline so hostiles remain identifiable.
          map.addSource("threat-halos", {
            type: "geojson",
            data: { type: "FeatureCollection", features: [] },
          });

          map.addLayer({
            id: "threat-halos-outline",
            type: "line",
            source: "threat-halos",
            paint: {
              "line-color": "#EF4444",
              "line-width": 0.8,
              "line-opacity": 0.6,
              "line-dasharray": [3, 3],
            },
          });

          // ── C: Track trails source (line layer, drawn under symbols) ────────
          map.addSource("track-trails", {
            type: "geojson",
            lineMetrics: true,
            data: { type: "FeatureCollection", features: [] },
          });

          // Fading trail line — gradient from transparent tail to opaque tip.
          // Uses line-gradient (not line-opacity) which requires lineMetrics:true on the source.
          // This produces a comet-tail effect and eliminates the white smear
          // that occurs when old opaque segments accumulate on screen.
          map.addLayer({
            id: "track-trails-line",
            type: "line",
            source: "track-trails",
            layout: {
              "line-join": "round",
              "line-cap":  "round",
            },
            paint: {
              "line-color": ["get", "trailColor"],
              "line-width": 1.2,
              // line-gradient fades tail (index 0) fully transparent → tip (index 1) opaque
              "line-gradient": [
                "interpolate", ["linear"],
                ["line-progress"],
                0.0, "rgba(255,255,255,0)",
                1.0, ["get", "trailColor"],
              ],
            },
          });

          // ── Track symbols (domain-specific geometric icons) ──────────────
          // Register all domain×hostile canvas icons on the map
          registerDomainIcons(map);

          map.addSource("tracks", {
            type: "geojson",
            data: { type: "FeatureCollection", features: [] },
          });

          // Symbol layer using data-driven icon-image keyed by domain+hostileClass.
          //
          // icon-size uses ZOOM-LEVEL interpolation so icons scale naturally as the
          // user zooms in/out — small dots at overview, readable shapes close in.
          // A secondary confidence factor nudges size ±10% for low/high confidence tracks.
          // Effective range: ~8-28px, keeping icons subtle and non-blocking at all zooms.
          map.addLayer({
            id: "tracks-symbol",
            type: "symbol",
            source: "tracks",
            layout: {
              "icon-image": ["get", "iconKey"],
              // Zoom-adaptive sizing: tiny at overview, grows as you zoom in.
              // Combined confidence tweak: ±10% based on track confidence score.
              "icon-size": [
                "interpolate", ["linear"], ["zoom"],
                3,  0.08,   // world overview — barely a dot
                5,  0.12,   // continental scale
                7,  0.16,   // regional scale
                9,  0.22,   // area scale
                11, 0.30,   // city scale
                13, 0.40,   // neighborhood scale — label-aligned detail
              ],
              "icon-allow-overlap": true,
              "icon-ignore-placement": true,
            },
            paint: {
              // Slight opacity fade so icons don't compete with the map tile
              "icon-opacity": 0.85,
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

    // G: geo-fences — fill + outline
    setVis("geofences-fill",  layerVisibility.geofences);
    setVis("geofences-layer", layerVisibility.geofences);

    // C: track trails
    setVis("track-trails-line", layerVisibility.trackTrails);

    // F: sensor coverage (fill + outline)
    // Note: sensor coverage layers are owned by SensorCoverageLayer.tsx which
    // also reads layerVisibility.sensorCoverage — no-op here to avoid double-set.
    // (SensorCoverageLayer syncs visibility on its own useEffect.)

    // B: track labels — TrackLabelsOverlay reads layerVisibility.trackLabels itself.
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
      {/* B: Track Labels HTML overlay (toggled via layerVisibility.trackLabels) */}
      <TrackLabelsOverlay />
      <SensorCoverageLayer />
      <LayerControls />
    </div>
  );
};
