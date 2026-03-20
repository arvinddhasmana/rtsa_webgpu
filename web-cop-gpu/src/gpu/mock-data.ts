// CLASSIFICATION: UNCLASSIFIED
// src/gpu/mock-data.ts — Mock track data generator for Phase 1 pipeline testing
//
// Generates up to 50,000 mock track records in a SharedArrayBuffer.
// Used by the Render Worker until Phase 2 delivers real WebTransport data.
//
// Geographic focus: West Asia (Iran / Persian Gulf region) — lon 44°E–63°E,
// lat 24°N–37°N — for realistic threat-scenario demonstration.
//
// Domain (iconIndex) and affiliation (threatLevel) values are aligned with the
// TrackDomain and TrackAffiliation enumerations in src/types/track-symbol.ts:
//
//   iconIndex:  0=AIR  1=SURFACE  2=SUBSURFACE  3=LAND  4=SPACE  5=CYBER
//   threatLevel: 0=UNKNOWN  1=PENDING  2=FRIENDLY  3=NEUTRAL  4=SUSPECT  5=HOSTILE
//
// Reference: docs/implementation/v4/phase1_core_rendering.md §5 (Mock Data Strategy)

import { HEADER_SIZE, RECORD_SIZE, TRACK_DATA_OFFSET } from "../services/sab";

/** Number of mock tracks to generate for tactical clarity. */
export const MOCK_TRACK_COUNT = 40;

/**
 * Geographic bounding box — West Asia (Iran / Persian Gulf region).
 * lon: 44°E – 63°E  |  lat: 24°N – 37°N
 * Values stored as radians to match the WebMercator projection used by the renderer.
 */
const LON_MIN =  44 * Math.PI / 180; //  44°E
const LON_MAX =  63 * Math.PI / 180; //  63°E
const LAT_MIN =  24 * Math.PI / 180; //  24°N
const LAT_MAX =  37 * Math.PI / 180; //  37°N

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
  /**
   * Trail ring buffer — 6 positions (newest first) forming 5 trail segments.
   * trail[0] = current position, trail[5] = oldest position.
   */
  trail: Array<{ lon: number; lat: number }>;
}

/** In-memory state for each mock track (used to animate them each frame). */
let mockTracks: MockTrackState[] = [];

/**
 * Initialise the mock track state array.
 * Call once at Render Worker startup.
 *
 * Domain distribution reflects a West Asia threat scenario:
 *   ~35% Air (fighters, drones)    iconIndex 0
 *   ~25% Surface (naval vessels)   iconIndex 1
 *   ~10% Subsurface (submarines)   iconIndex 2
 *   ~20% Land (ground vehicles)    iconIndex 3
 *   ~ 5% Space (satellites)        iconIndex 4
 *   ~ 5% Cyber (logical entities)  iconIndex 5
 *
 * Affiliation distribution (threat scenario bias):
 *   ~20% Friendly  (own-force tracks)   threatLevel 2
 *   ~15% Neutral                        threatLevel 3
 *   ~10% Unknown                        threatLevel 0
 *   ~ 5% Pending                        threatLevel 1
 *   ~20% Suspect                        threatLevel 4
 *   ~30% Hostile                        threatLevel 5
 */
export function initMockTracks(count: number = MOCK_TRACK_COUNT): void {
  mockTracks = [];

  // Weighted domain pool — 20 entries giving the distribution above
  const DOMAIN_POOL: number[] = [
    0, 0, 0, 0, 0, 0, 0,  // 35% Air
    1, 1, 1, 1, 1,         // 25% Surface
    2, 2,                  // 10% Subsurface
    3, 3, 3, 3,            // 20% Land
    4,                     // 5%  Space
    5,                     // 5%  Cyber
  ];

  // Weighted affiliation pool — 20 entries giving the distribution above
  const AFFIL_POOL: number[] = [
    2, 2, 2, 2,            // 20% Friendly
    3, 3, 3,               // 15% Neutral
    0, 0,                  // 10% Unknown
    1,                     //  5% Pending
    4, 4, 4, 4,            // 20% Suspect
    5, 5, 5, 5, 5, 5,      // 30% Hostile
  ];

  for (let i = 0; i < count; i++) {
    const lon = LON_MIN + Math.random() * (LON_MAX - LON_MIN);
    const lat = LAT_MIN + Math.random() * (LAT_MAX - LAT_MIN);
    // Initialise all trail positions to the starting point
    const trail: Array<{ lon: number; lat: number }> = [];
    for (let ti = 0; ti < 6; ti++) {
      trail.push({ lon, lat });
    }

    const iconIndex   = DOMAIN_POOL[Math.floor(Math.random() * DOMAIN_POOL.length)]!;
    const threatLevel = AFFIL_POOL[Math.floor(Math.random() * AFFIL_POOL.length)]!;

    mockTracks.push({
      trackIdHash: i + 1, // 0 reserved for "no track" in pick buffer
      lon,
      lat,
      course:     Math.random() * TWO_PI,
      speed:      50 + Math.random() * 550,  // 50–600 m/s (land to supersonic air)
      altitude:   iconIndex === 0 || iconIndex === 4
                    ? 1000 + Math.random() * 12_000 // Air / Space — higher altitude
                    : 0,                             // Surface / Land / Subsurface / Cyber
      threatLevel,
      alertFlags: threatLevel >= 4 && Math.random() < 0.3 ? 1 : 0, // Suspect/Hostile ~30% alert
      iconIndex,
      trail,
    });
  }
}

// Seeding logic moved to App.tsx for better module context sync

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
    // trail[0] = newest position, trail[5] = oldest position
    for (let s = 0; s < 5; s++) {
      const offset = base + 0x30 + s * 16;
      const pNewer = t.trail[s]!;
      const pOlder = t.trail[s + 1]!;
      trackData.setFloat32(offset +  0, pOlder.lon, true);
      trackData.setFloat32(offset +  4, pOlder.lat, true);
      trackData.setFloat32(offset +  8, pNewer.lon, true);
      trackData.setFloat32(offset + 12, pNewer.lat, true);
    }
  }
}

/**
 * Advance all mock tracks by one tick (animate position along course).
 * dt_ms is the elapsed time in milliseconds since the last tick.
 * After updating each track's position, shifts the trail ring buffer by one
 * entry and inserts the new position as the newest trail point.
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

    // Shift trail ring buffer: oldest entry drops off, current position is newest
    for (let ti = 5; ti > 0; ti--) {
      t.trail[ti] = t.trail[ti - 1]!;
    }
    t.trail[0] = { lon: t.lon, lat: t.lat };
  }
}
