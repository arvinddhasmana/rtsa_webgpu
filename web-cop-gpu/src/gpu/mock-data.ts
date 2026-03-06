// CLASSIFICATION: UNCLASSIFIED
// src/gpu/mock-data.ts — Mock track data generator for Phase 1 pipeline testing
//
// Generates up to 50,000 mock track records in a SharedArrayBuffer.
// Used by the Render Worker until Phase 2 delivers real WebTransport data.
//
// Reference: docs/implementation/v4/phase1_core_rendering.md §5 (Mock Data Strategy)

import { TRACK_DATA_OFFSET, RECORD_SIZE, HEADER_SIZE } from "../services/sab";

/** Number of mock tracks to generate (Phase 1 target). */
export const MOCK_TRACK_COUNT = 50_000;

/** Geographic bounding box for track scatter (North Atlantic / Europe region). */
const LON_MIN = -0.5;      // ~-29°
const LON_MAX =  0.5;      // ~+29°
const LAT_MIN = -0.4;      // ~-23°
const LAT_MAX =  0.4;      // ~+23°

const TWO_PI = Math.PI * 2;

export interface MockTrackState {
  /** Track identifier hash (slot index used as a simple mock ID). */
  trackIdHash: number;
  lon: number;
  lat: number;
  course: number; // radians
  speed: number;  // m/s
  altitude: number;
  threatLevel: number;
  alertFlags: number;
  iconIndex: number;
}

/** In-memory state for each mock track (used to animate them each frame). */
let mockTracks: MockTrackState[] = [];

/**
 * Initialise the mock track state array.
 * Call once at Render Worker startup.
 */
export function initMockTracks(count: number = MOCK_TRACK_COUNT): void {
  mockTracks = [];
  for (let i = 0; i < count; i++) {
    mockTracks.push({
      trackIdHash: i + 1, // 0 reserved for "no track" in pick buffer
      lon:         LON_MIN + Math.random() * (LON_MAX - LON_MIN),
      lat:         LAT_MIN + Math.random() * (LAT_MAX - LAT_MIN),
      course:      Math.random() * TWO_PI,
      speed:       100 + Math.random() * 500, // 100–600 m/s
      altitude:    1000 + Math.random() * 10_000,
      threatLevel: Math.floor(Math.random() * 6),      // 0–5
      alertFlags:  Math.random() < 0.05 ? 1 : 0,      // ~5% have alerts
      iconIndex:   Math.floor(Math.random() * 16),     // 16 icon types
    });
  }
}

/**
 * Write all mock tracks into the SharedArrayBuffer track data region.
 * Also writes the active_track_count into the header.
 *
 * Call once per frame (or at init) to keep the mock data alive.
 *
 * Reference: webgpu_guidelines.md §4.2
 */
export function writeMockTracksToSAB(
  sab: SharedArrayBuffer,
  count: number = MOCK_TRACK_COUNT,
): void {
  const header = new Uint32Array(sab, 0, HEADER_SIZE / 4);
  // Update active track count (index 0 in header)
  Atomics.store(header, 0, count);

  const trackData = new DataView(sab, TRACK_DATA_OFFSET);
  const now = Date.now();

  const actualCount = Math.min(count, mockTracks.length);
  for (let i = 0; i < actualCount; i++) {
    const t = mockTracks[i];
    const base = i * RECORD_SIZE;

    trackData.setFloat32(base + 0x00, t.lon,         true); // longitude
    trackData.setFloat32(base + 0x04, t.lat,         true); // latitude
    trackData.setFloat32(base + 0x08, t.course,      true); // course
    trackData.setFloat32(base + 0x0c, t.speed,       true); // speed
    trackData.setFloat32(base + 0x10, t.altitude,    true); // altitude
    trackData.setUint32 (base + 0x14, t.trackIdHash, true); // track_id_hash
    trackData.setUint32 (base + 0x18, 1,             true); // source_bitmap
    trackData.setUint32 (base + 0x1c, 1,             true); // classification_level
    trackData.setUint32 (base + 0x20, t.threatLevel, true); // threat_level
    trackData.setUint32 (base + 0x24, t.iconIndex,   true); // icon_index
    trackData.setUint32 (base + 0x28, t.alertFlags,  true); // alert_flags
    trackData.setUint32 (base + 0x2c, now >>> 0,     true); // update_epoch_ms

    // Trail ring (5 × vec4<f32> — lon_a, lat_a, lon_b, lat_b per segment)
    for (let s = 0; s < 5; s++) {
      const offset = base + 0x30 + s * 16;
      // Mock trail: each segment slightly offset backwards
      const lonA = t.lon - Math.sin(t.course) * (s + 1) * 0.001;
      const latA = t.lat - Math.cos(t.course) * (s + 1) * 0.001;
      const lonB = t.lon - Math.sin(t.course) * s * 0.001;
      const latB = t.lat - Math.cos(t.course) * s * 0.001;
      trackData.setFloat32(offset +  0, lonA, true);
      trackData.setFloat32(offset +  4, latA, true);
      trackData.setFloat32(offset +  8, lonB, true);
      trackData.setFloat32(offset + 12, latB, true);
    }
  }
}

/**
 * Advance all mock tracks by one tick (animate position along course).
 * dt_ms is the elapsed time in milliseconds since the last tick.
 */
export function tickMockTracks(dt_ms: number): void {
  const dt_s = dt_ms / 1000;
  const EARTH_R = 6_371_000;

  for (const t of mockTracks) {
    const cos_lat = Math.cos(t.lat);
    const safe_cos = Math.abs(cos_lat) < 1e-4 ? 1e-4 : cos_lat;
    t.lon += (t.speed * Math.sin(t.course) * dt_s) / (EARTH_R * safe_cos);
    t.lat += (t.speed * Math.cos(t.course) * dt_s) / EARTH_R;

    // Wrap longitude to [-π, π]
    if (t.lon >  Math.PI) t.lon -= Math.PI * 2;
    if (t.lon < -Math.PI) t.lon += Math.PI * 2;

    // Bounce latitude at ±π/2
    if (t.lat >  Math.PI / 2) { t.lat =  Math.PI - t.lat; t.course = TWO_PI - t.course; }
    if (t.lat < -Math.PI / 2) { t.lat = -Math.PI - t.lat; t.course = TWO_PI - t.course; }
  }
}
