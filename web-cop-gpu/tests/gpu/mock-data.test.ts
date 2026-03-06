// CLASSIFICATION: UNCLASSIFIED
// tests/gpu/mock-data.test.ts — Unit tests for mock track data generator
//
// Verifies that mock track initialisation and SAB writing produce
// correctly structured data at the expected byte offsets.
//
// Reference: docs/implementation/v4/phase1_core_rendering.md §5

import { describe, it, expect, beforeEach } from "vitest";
import {
  initMockTracks,
  writeMockTracksToSAB,
  tickMockTracks,
  MOCK_TRACK_COUNT,
} from "../../src/gpu/mock-data";
import {
  allocateSAB,
  RECORD_SIZE,
  HEADER_SIZE,
  DIRTY_BITFIELD_SIZE,
  TRACK_DATA_OFFSET,
  HEADER_OFFSET_ACTIVE_TRACK_COUNT,
} from "../../src/services/sab";

const SMALL_COUNT = 10;

describe("initMockTracks", () => {
  it("default MOCK_TRACK_COUNT is 50,000", () => {
    expect(MOCK_TRACK_COUNT).toBe(50_000);
  });

  it("can be initialised with a smaller count for tests", () => {
    // This just verifies the function doesn't throw
    expect(() => initMockTracks(SMALL_COUNT)).not.toThrow();
  });
});

describe("writeMockTracksToSAB", () => {
  beforeEach(() => {
    initMockTracks(SMALL_COUNT);
  });

  it("writes active_track_count to header index 0", () => {
    const buf = allocateSAB(SMALL_COUNT);
    writeMockTracksToSAB(buf.sab, SMALL_COUNT);

    const header = new Uint32Array(buf.sab, 0, HEADER_SIZE / 4);
    expect(Atomics.load(header, HEADER_OFFSET_ACTIVE_TRACK_COUNT)).toBe(SMALL_COUNT);
  });

  it("writes non-zero longitude at track offset 0x00", () => {
    const buf = allocateSAB(SMALL_COUNT);
    writeMockTracksToSAB(buf.sab, SMALL_COUNT);

    // Not all tracks are guaranteed non-zero lon, but at least one should be
    const trackData = new DataView(buf.sab, TRACK_DATA_OFFSET);
    let hasNonZero = false;
    for (let i = 0; i < SMALL_COUNT; i++) {
      if (trackData.getFloat32(i * RECORD_SIZE + 0x00, true) !== 0) {
        hasNonZero = true;
        break;
      }
    }
    expect(hasNonZero).toBe(true);
  });

  it("writes track_id_hash (non-zero) at offset 0x14", () => {
    const buf = allocateSAB(SMALL_COUNT);
    writeMockTracksToSAB(buf.sab, SMALL_COUNT);

    const trackData = new DataView(buf.sab, TRACK_DATA_OFFSET);
    // First track: id hash = 1 (slot 0 + 1)
    const idHash = trackData.getUint32(0x14, true);
    expect(idHash).toBeGreaterThan(0);
  });

  it("writes threat_level in range 0–5 at offset 0x20", () => {
    const buf = allocateSAB(SMALL_COUNT);
    writeMockTracksToSAB(buf.sab, SMALL_COUNT);

    const trackData = new DataView(buf.sab, TRACK_DATA_OFFSET);
    for (let i = 0; i < SMALL_COUNT; i++) {
      const level = trackData.getUint32(i * RECORD_SIZE + 0x20, true);
      expect(level).toBeGreaterThanOrEqual(0);
      expect(level).toBeLessThanOrEqual(5);
    }
  });

  it("writes trail data at offset 0x30 (non-zero for non-zero tracks)", () => {
    const buf = allocateSAB(SMALL_COUNT);
    writeMockTracksToSAB(buf.sab, SMALL_COUNT);

    const trackData = new DataView(buf.sab, TRACK_DATA_OFFSET);
    // Trail segment 0, lon_a (at offset 0x30)
    let hasTrail = false;
    for (let i = 0; i < SMALL_COUNT; i++) {
      const trailLon = trackData.getFloat32(i * RECORD_SIZE + 0x30, true);
      if (trailLon !== 0) { hasTrail = true; break; }
    }
    expect(hasTrail).toBe(true);
  });

  it("total SAB has correct layout offsets", () => {
    // Verify constants match our understanding
    expect(HEADER_SIZE).toBe(4096);
    expect(DIRTY_BITFIELD_SIZE).toBe(8192);
    expect(TRACK_DATA_OFFSET).toBe(4096 + 8192);
    expect(RECORD_SIZE).toBe(128);
  });
});

describe("tickMockTracks", () => {
  beforeEach(() => {
    initMockTracks(SMALL_COUNT);
  });

  it("does not throw for a 16ms tick", () => {
    expect(() => tickMockTracks(16)).not.toThrow();
  });

  it("changes track positions after ticking", () => {
    const buf = allocateSAB(SMALL_COUNT);
    writeMockTracksToSAB(buf.sab, SMALL_COUNT);

    const before = new DataView(buf.sab, TRACK_DATA_OFFSET);
    const lon0Before = before.getFloat32(0, true);

    tickMockTracks(1000); // 1 second at track speed
    writeMockTracksToSAB(buf.sab, SMALL_COUNT);

    const after = new DataView(buf.sab, TRACK_DATA_OFFSET);
    const lon0After = after.getFloat32(0, true);

    // Position should have moved after 1 second at flight speed
    expect(lon0After).not.toBe(lon0Before);
  });

  it("keeps longitude within [-π, π] after many ticks", () => {
    const buf = allocateSAB(SMALL_COUNT);

    // Tick for a long simulated time
    for (let i = 0; i < 100; i++) {
      tickMockTracks(1000);
    }
    writeMockTracksToSAB(buf.sab, SMALL_COUNT);

    const trackData = new DataView(buf.sab, TRACK_DATA_OFFSET);
    for (let i = 0; i < SMALL_COUNT; i++) {
      const lon = trackData.getFloat32(i * RECORD_SIZE + 0x00, true);
      expect(Math.abs(lon)).toBeLessThanOrEqual(Math.PI + 0.01); // small epsilon for float precision
    }
  });
});
