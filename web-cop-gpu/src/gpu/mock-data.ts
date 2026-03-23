// CLASSIFICATION: UNCLASSIFIED
// src/gpu/mock-data.ts — Mock track data generator for West Asia demo scenario
//
// Generates ~150 mock fused track records in a SharedArrayBuffer.
// Used by the Render Worker until Phase 2 delivers real WebTransport data.
//
// Geographic focus: Persian Gulf / Gulf of Oman / West Asia
//   lon: 46°E – 62°E  |  lat: 22°N – 32°N
//
// icon_index encoding: context * 36 + entity_type * 6 + threat_level
//   context     (0=MILITARY, 1=CIVILIAN)
//   entity_type proto values: 0=Unspecified, 1=Surface, 2=Air,
//                             3=Subsurface, 4=Land, 5=Cyber
//   threat_level: 0=Unknown, 1=Pending, 2=Friendly, 3=Neutral, 4=Suspect, 5=Hostile
//
// The threat_level field is ALSO written separately (SAB offset 0x20) for use
// by pick and halo shaders that don't need to decode icon_index.
//
// Reference: docs/implementation/v5/operations_commander/plan-mil2525SymbologyWestAsiaDemo.md §Phase 3

import { HEADER_SIZE, RECORD_SIZE, TRACK_DATA_OFFSET } from "../services/sab";
import {
    encodeIconIndex,
    ENTITY_TYPE_AIR,
    ENTITY_TYPE_CYBER,
    ENTITY_TYPE_LAND,
    ENTITY_TYPE_SUBSURFACE,
    ENTITY_TYPE_SURFACE,
    TrackRenderContext,
} from "../types/symbology";

/** Number of mock tracks for the West Asia demo scenario. */
export const MOCK_TRACK_COUNT = 150;

/**
 * Geographic bounding box — Persian Gulf / Gulf of Oman (West Asia).
 * lon: 46°E – 62°E  |  lat: 22°N – 32°N
 * Values stored as radians to match the WebMercator projection.
 */
const LON_MIN = 46 * Math.PI / 180; // 46°E
const LON_MAX = 62 * Math.PI / 180; // 62°E
const LAT_MIN = 22 * Math.PI / 180; // 22°N
const LAT_MAX = 32 * Math.PI / 180; // 32°N

const TWO_PI = Math.PI * 2;

// ── Named strategic sensor positions ─────────────────────────────────────────
// 7 sensors at named West Asia / Persian Gulf locations.
// source_bitmap bit assignments:
//   Bit 0: RADAR-HORMUZ-001    (Strait of Hormuz)
//   Bit 1: RADAR-BAHRAIN-002   (Bahrain)
//   Bit 2: AIS-DUBAI-001       (Dubai)
//   Bit 3: EW-MUSCAT-001       (Muscat, Oman)
//   Bit 4: ELINT-CHABAHAR-001  (Chabahar, Iran)
//   Bit 5: ISR-BANDARABBAS-001 (Bandar Abbas, Iran)
//   Bit 6: RADAR-QESHM-003     (Qeshm Island)
const SENSOR_HORMUZ    = 1 << 0; // 0x01
const SENSOR_BAHRAIN   = 1 << 1; // 0x02
const SENSOR_DUBAI     = 1 << 2; // 0x04
const SENSOR_MUSCAT    = 1 << 3; // 0x08
const SENSOR_CHABAHAR  = 1 << 4; // 0x10
const SENSOR_BANDAR    = 1 << 5; // 0x20
const SENSOR_QESHM     = 1 << 6; // 0x40

export interface MockTrackState {
  /** Track identifier hash (slot index + 1; 0 is reserved for "no selection"). */
  trackIdHash: number;
  lon: number;
  lat: number;
  course: number;   // radians
  speed: number;    // m/s
  altitude: number; // metres (0 for surface / land / subsurface)
  threatLevel: number;
  alertFlags: number;
  iconIndex: number; // encoded: context * 36 + entity_type * 6 + threat_level
  sourceBitmap: number;
  /**
   * Trail ring buffer — 6 positions (newest at index 0, oldest at index 5).
   * trail[0] = current position, trail[5] = oldest.
   */
  trail: Array<{ lon: number; lat: number }>;
}

/** In-memory state for each mock track (mutated by tickMockTracks). */
let mockTracks: MockTrackState[] = [];

// ── Random helpers ────────────────────────────────────────────────────────────

function randFloat(min: number, max: number): number {
  return min + Math.random() * (max - min);
}

/** Clamp a lon/lat-in-radians value to the West Asia bounding box. */
function clampToBox(lon: number, lat: number): { lon: number; lat: number } {
  return {
    lon: Math.max(LON_MIN, Math.min(LON_MAX, lon)),
    lat: Math.max(LAT_MIN, Math.min(LAT_MAX, lat)),
  };
}

const WATER_BOXES = [
  { minLon: 48, maxLon: 55, minLat: 24.5, maxLat: 29 }, // Persian Gulf
  { minLon: 56.5, maxLon: 61, minLat: 22, maxLat: 25.5 }, // Gulf of Oman
];

const LAND_BOXES = [
  { minLon: 50, maxLon: 60, minLat: 29.5, maxLat: 31.5 }, // Iran Inland
  { minLon: 52, maxLon: 56, minLat: 22, maxLat: 23.5 }, // UAE / Oman Inland
];

function randWater(): { lon: number; lat: number } {
  const b = WATER_BOXES[Math.floor(Math.random() * WATER_BOXES.length)]!;
  return clampToBox(
    randFloat(b.minLon * Math.PI / 180, b.maxLon * Math.PI / 180),
    randFloat(b.minLat * Math.PI / 180, b.maxLat * Math.PI / 180)
  );
}

function randLand(): { lon: number; lat: number } {
  const b = LAND_BOXES[Math.floor(Math.random() * LAND_BOXES.length)]!;
  return clampToBox(
    randFloat(b.minLon * Math.PI / 180, b.maxLon * Math.PI / 180),
    randFloat(b.minLat * Math.PI / 180, b.maxLat * Math.PI / 180)
  );
}

// ── Track factory helpers ─────────────────────────────────────────────────────

function makeTrail(lon: number, lat: number): Array<{ lon: number; lat: number }> {
  const trail: Array<{ lon: number; lat: number }> = [];
  for (let i = 0; i < 6; i++) trail.push({ lon, lat });
  return trail;
}

function makeTrack(
  idx: number,
  lon: number,
  lat: number,
  entityType: number,
  threatLevel: number,
  context: TrackRenderContext,
  speed: number,
  altitude: number,
  sourceBitmap: number,
): MockTrackState {
  const iconIndex = encodeIconIndex(context, entityType, threatLevel);
  const alertFlags = (threatLevel >= 4 && Math.random() < 0.3) ? 1 : 0;
  return {
    trackIdHash: idx + 1,
    lon,
    lat,
    course:     Math.random() * TWO_PI,
    speed,
    altitude,
    threatLevel,
    alertFlags,
    iconIndex,
    sourceBitmap,
    trail:      makeTrail(lon, lat),
  };
}

/** AIS sensor bitmap helper (cycling through Dubai + Hormuz for civilian surface tracks). */
function SENSOR_AIS_BITMAP(i: number): number {
  return (i % 2 === 0) ? (SENSOR_DUBAI | SENSOR_HORMUZ) : (SENSOR_BAHRAIN | SENSOR_QESHM);
}

/**
 * Initialise the mock track state array for the West Asia demo scenario.
 *
 * Distribution (~150 fused tracks):
 *   55  Surface/Vessel  — tankers, warships, fishing vessels (Persian Gulf, Hormuz, Gulf of Oman)
 *   40  Air             — military jets, commercial aircraft, UAVs
 *   30  Land            — military vehicles, SAM sites, ground forces
 *   15  Subsurface      — submarines (Gulf of Oman approaches)
 *   10  Cyber           — positioned at major urban centres
 *
 * Affiliation distribution:
 *   ~45% Friendly (coalition naval/air, UAE/Oman), ~10% Hostile (Iranian assets),
 *   ~20% Neutral (commercial), ~12% Unknown, ~8% Suspect, ~5% Pending
 *
 * Context distribution: ~55% Military, ~45% Civilian
 */
export function initMockTracks(_count: number = MOCK_TRACK_COUNT): void {
  mockTracks = [];
  let idx = 0;

  // ── Surface / Vessel tracks (55) ────────────────────────────────────────────
  // 22 Military Friendly — coalition warships, patrol boats
  for (let i = 0; i < 22; i++) {
    const { lon, lat } = randWater();
    mockTracks.push(makeTrack(idx++, lon, lat, ENTITY_TYPE_SURFACE, 2, TrackRenderContext.MILITARY,
      randFloat(5, 18), 0, SENSOR_HORMUZ | SENSOR_BAHRAIN));
  }
  // 5 Military Hostile — Iranian warships
  for (let i = 0; i < 5; i++) {
    const { lon, lat } = randWater();
    mockTracks.push(makeTrack(idx++, lon, lat, ENTITY_TYPE_SURFACE, 5, TrackRenderContext.MILITARY,
      randFloat(8, 22), 0, SENSOR_HORMUZ | SENSOR_BANDAR));
  }
  // 5 Military Suspect — vessels with AIS anomalies
  for (let i = 0; i < 5; i++) {
    const { lon, lat } = randWater();
    mockTracks.push(makeTrack(idx++, lon, lat, ENTITY_TYPE_SURFACE, 4, TrackRenderContext.MILITARY,
      randFloat(5, 15), 0, SENSOR_DUBAI | SENSOR_HORMUZ));
  }
  // 8 Military Unknown — unidentified contacts
  for (let i = 0; i < 8; i++) {
    const { lon, lat } = randWater();
    mockTracks.push(makeTrack(idx++, lon, lat, ENTITY_TYPE_SURFACE, 0, TrackRenderContext.MILITARY,
      randFloat(3, 12), 0, SENSOR_CHABAHAR | SENSOR_MUSCAT));
  }
  // 15 Civilian Neutral — tankers, cargo ships
  for (let i = 0; i < 15; i++) {
    const { lon, lat } = randWater();
    mockTracks.push(makeTrack(idx++, lon, lat, ENTITY_TYPE_SURFACE, 3, TrackRenderContext.CIVILIAN,
      randFloat(6, 14), 0, SENSOR_AIS_BITMAP(i)));
  }

  // ── Air tracks (40) ────────────────────────────────────────────────────────
  // 15 Military Friendly — UAE/coalition aircraft
  for (let i = 0; i < 15; i++) {
    const { lon, lat } = clampToBox(
      randFloat(46 * Math.PI / 180, 58 * Math.PI / 180),
      randFloat(24 * Math.PI / 180, 31 * Math.PI / 180),
    );
    mockTracks.push(makeTrack(idx++, lon, lat, ENTITY_TYPE_AIR, 2, TrackRenderContext.MILITARY,
      randFloat(200, 500), randFloat(5000, 12000), SENSOR_BAHRAIN | SENSOR_DUBAI));
  }
  // 6 Military Hostile — Iranian military jets
  for (let i = 0; i < 6; i++) {
    const { lon, lat } = clampToBox(
      randFloat(56 * Math.PI / 180, 62 * Math.PI / 180),
      randFloat(26 * Math.PI / 180, 31 * Math.PI / 180),
    );
    mockTracks.push(makeTrack(idx++, lon, lat, ENTITY_TYPE_AIR, 5, TrackRenderContext.MILITARY,
      randFloat(300, 600), randFloat(3000, 10000), SENSOR_HORMUZ | SENSOR_BANDAR));
  }
  // 4 Military Suspect — unacknowledged military aircraft
  for (let i = 0; i < 4; i++) {
    const { lon, lat } = clampToBox(
      randFloat(54 * Math.PI / 180, 60 * Math.PI / 180),
      randFloat(25 * Math.PI / 180, 29 * Math.PI / 180),
    );
    mockTracks.push(makeTrack(idx++, lon, lat, ENTITY_TYPE_AIR, 4, TrackRenderContext.MILITARY,
      randFloat(150, 400), randFloat(1000, 8000), SENSOR_HORMUZ | SENSOR_QESHM));
  }
  // 5 Military Unknown — unidentified airborne contacts
  for (let i = 0; i < 5; i++) {
    const { lon, lat } = clampToBox(
      randFloat(55 * Math.PI / 180, 62 * Math.PI / 180),
      randFloat(23 * Math.PI / 180, 27 * Math.PI / 180),
    );
    mockTracks.push(makeTrack(idx++, lon, lat, ENTITY_TYPE_AIR, 0, TrackRenderContext.MILITARY,
      randFloat(100, 350), randFloat(500, 6000), SENSOR_MUSCAT | SENSOR_CHABAHAR));
  }
  // 10 Civilian Neutral — commercial flights
  for (let i = 0; i < 10; i++) {
    const { lon, lat } = clampToBox(
      randFloat(46 * Math.PI / 180, 62 * Math.PI / 180),
      randFloat(22 * Math.PI / 180, 32 * Math.PI / 180),
    );
    mockTracks.push(makeTrack(idx++, lon, lat, ENTITY_TYPE_AIR, 3, TrackRenderContext.CIVILIAN,
      randFloat(200, 280), randFloat(9000, 13000), SENSOR_DUBAI | SENSOR_BAHRAIN));
  }

  // ── Land tracks (30) ──────────────────────────────────────────────────────
  // 14 Military Friendly — UAE/Oman ground forces
  for (let i = 0; i < 14; i++) {
    const { lon, lat } = randLand();
    mockTracks.push(makeTrack(idx++, lon, lat, ENTITY_TYPE_LAND, 2, TrackRenderContext.MILITARY,
      randFloat(10, 60), 0, SENSOR_BAHRAIN | SENSOR_DUBAI));
  }
  // 8 Military Hostile — Iranian coastal SAM / IRGCN ground units
  for (let i = 0; i < 8; i++) {
    const { lon, lat } = randLand();
    mockTracks.push(makeTrack(idx++, lon, lat, ENTITY_TYPE_LAND, 5, TrackRenderContext.MILITARY,
      randFloat(0, 40), 0, SENSOR_BANDAR | SENSOR_CHABAHAR));
  }
  // 4 Military Pending — newly detected ground contacts
  for (let i = 0; i < 4; i++) {
    const { lon, lat } = randLand();
    mockTracks.push(makeTrack(idx++, lon, lat, ENTITY_TYPE_LAND, 1, TrackRenderContext.MILITARY,
      randFloat(0, 30), 0, SENSOR_QESHM | SENSOR_HORMUZ));
  }
  // 4 Military Unknown — unidentified ground targets
  for (let i = 0; i < 4; i++) {
    const { lon, lat } = randLand();
    mockTracks.push(makeTrack(idx++, lon, lat, ENTITY_TYPE_LAND, 0, TrackRenderContext.MILITARY,
      randFloat(0, 50), 0, SENSOR_MUSCAT | SENSOR_CHABAHAR));
  }

  // ── Subsurface tracks (15) ────────────────────────────────────────────────
  // 5 Military Friendly — allied submarines
  for (let i = 0; i < 5; i++) {
    const { lon, lat } = randWater();
    mockTracks.push(makeTrack(idx++, lon, lat, ENTITY_TYPE_SUBSURFACE, 2, TrackRenderContext.MILITARY,
      randFloat(8, 18), 0, SENSOR_MUSCAT | SENSOR_CHABAHAR));
  }
  // 4 Military Hostile — Iranian submarines
  for (let i = 0; i < 4; i++) {
    const { lon, lat } = randWater();
    mockTracks.push(makeTrack(idx++, lon, lat, ENTITY_TYPE_SUBSURFACE, 5, TrackRenderContext.MILITARY,
      randFloat(5, 15), 0, SENSOR_HORMUZ | SENSOR_BANDAR));
  }
  // 3 Military Suspect — unclassified subsurface contacts
  for (let i = 0; i < 3; i++) {
    const { lon, lat } = randWater();
    mockTracks.push(makeTrack(idx++, lon, lat, ENTITY_TYPE_SUBSURFACE, 4, TrackRenderContext.MILITARY,
      randFloat(3, 12), 0, SENSOR_MUSCAT | SENSOR_CHABAHAR));
  }
  // 3 Military Unknown — unidentified underwater contacts
  for (let i = 0; i < 3; i++) {
    const { lon, lat } = randWater();
    mockTracks.push(makeTrack(idx++, lon, lat, ENTITY_TYPE_SUBSURFACE, 0, TrackRenderContext.MILITARY,
      randFloat(2, 10), 0, SENSOR_HORMUZ | SENSOR_BANDAR));
  }

  // ── Cyber tracks (10) ────────────────────────────────────────────────────
  // Positioned at major urban centres: Tehran(51.4°E,35.7°N), Dubai(55.1°E,25.1°N),
  // Riyadh(46.7°E,24.7°N), Abu Dhabi(54.4°E,24.5°N), Muscat(58.6°E,23.6°N)
  const cyberPositions: [number, number][] = [
    [51.4, 35.7], [55.1, 25.1], [46.7, 24.7], [54.4, 24.5], [58.6, 23.6],
    [50.5, 26.2], // Manama, Bahrain
    [51.5, 25.3], // Doha, Qatar
    [56.3, 27.2], // Bandar Abbas area
    [53.7, 24.0], // UAE interior
    [48.7, 31.3], // Ahvaz, Iran
  ];
  const cyberAffils = [5, 4, 2, 3, 0, 2, 3, 5, 2, 4]; // Hostile/Suspect/Friendly/Neutral/Unknown
  for (let i = 0; i < 10; i++) {
    const [lonDeg, latDeg] = cyberPositions[i]!;
    const lon = lonDeg * Math.PI / 180;
    const lat = latDeg * Math.PI / 180;
    const threatLevel = cyberAffils[i]!;
    mockTracks.push(makeTrack(idx++, lon, lat, ENTITY_TYPE_CYBER, threatLevel, TrackRenderContext.MILITARY,
      0, 0, SENSOR_BAHRAIN | SENSOR_DUBAI));
  }
}

/**
 * Write all mock tracks into the SharedArrayBuffer track data region.
 * Also writes the active_track_count into the header.
 * Call once per frame (or at init) to keep the mock data alive.
 *
 * Reference: webgpu_guidelines.md §4.2
 */
export function writeMockTracksToSAB(
  sab: SharedArrayBuffer,
  count: number = MOCK_TRACK_COUNT,
): void {
  const header = new Uint32Array(sab, 0, HEADER_SIZE / 4);
  Atomics.store(header, 0, count);

  const trackData = new DataView(sab, TRACK_DATA_OFFSET);
  const now = Date.now();

  const actualCount = Math.min(count, mockTracks.length);
  for (let i = 0; i < actualCount; i++) {
    const t = mockTracks[i]!;
    const base = i * RECORD_SIZE;

    trackData.setFloat32(base + 0x00, t.lon,          true); // longitude
    trackData.setFloat32(base + 0x04, t.lat,          true); // latitude
    trackData.setFloat32(base + 0x08, t.course,       true); // course
    trackData.setFloat32(base + 0x0c, t.speed,        true); // speed
    trackData.setFloat32(base + 0x10, t.altitude,     true); // altitude
    trackData.setUint32 (base + 0x14, t.trackIdHash,  true); // track_id_hash
    trackData.setUint32 (base + 0x18, t.sourceBitmap, true); // source_bitmap
    trackData.setUint32 (base + 0x1c, 1,              true); // classification_level
    trackData.setUint32 (base + 0x20, t.threatLevel,  true); // threat_level
    trackData.setUint32 (base + 0x24, t.iconIndex,    true); // icon_index (encoded)
    trackData.setUint32 (base + 0x28, t.alertFlags,   true); // alert_flags
    trackData.setUint32 (base + 0x2c, now >>> 0,      true); // update_epoch_ms

    // Trail ring (5 × vec4<f32> — lon_a, lat_a, lon_b, lat_b per segment)
    for (let s = 0; s < 5; s++) {
      const offset  = base + 0x30 + s * 16;
      const pNewer  = t.trail[s]!;
      const pOlder  = t.trail[s + 1]!;
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
 */
export function tickMockTracks(dt_ms: number): void {
  const dt_s   = dt_ms / 1000;
  const EARTH_R = 6_371_000;

  for (const t of mockTracks) {
    if (t.speed === 0) continue; // Cyber / static land tracks don't move
    const cos_lat  = Math.cos(t.lat);
    const safe_cos = Math.abs(cos_lat) < 1e-4 ? 1e-4 : cos_lat;
    t.lon += (t.speed * Math.sin(t.course) * dt_s) / (EARTH_R * safe_cos);
    t.lat += (t.speed * Math.cos(t.course) * dt_s) / EARTH_R;

    // Wrap longitude to [-π, π]
    if (t.lon >  Math.PI) t.lon -= Math.PI * 2;
    if (t.lon < -Math.PI) t.lon += Math.PI * 2;

    // Bounce latitude at ±π/2
    if (t.lat >  Math.PI / 2) { t.lat =  Math.PI - t.lat; t.course = TWO_PI - t.course; }
    if (t.lat < -Math.PI / 2) { t.lat = -Math.PI - t.lat; t.course = TWO_PI - t.course; }

    // Shift trail ring buffer: oldest entry drops off, current becomes newest
    for (let ti = 5; ti > 0; ti--) {
      t.trail[ti] = t.trail[ti - 1]!;
    }
    t.trail[0] = { lon: t.lon, lat: t.lat };
  }
}
