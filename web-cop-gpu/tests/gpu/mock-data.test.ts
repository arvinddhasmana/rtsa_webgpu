// CLASSIFICATION: UNCLASSIFIED
// tests/gpu/mock-data.test.ts — Unit tests for mock track data generator
//
// Verifies that:
//   1. update_epoch_ms written to the SAB is based on Date.now() (Unix epoch ms).
//   2. tickMockTracks() shifts the trail ring buffer so trail entries change.
//
// Reference: docs/implementation/v4/B2-MockData-Timestamp-Interpolation.md

import { describe, it, expect } from "vitest";
import {
  initMockTracks,
  writeMockTracksToSAB,
  tickMockTracks,
} from "../../src/gpu/mock-data";
import {
  HEADER_SIZE,
  DIRTY_BITFIELD_SIZE,
  RECORD_SIZE,
} from "../../src/services/sab";

const TRACK_DATA_OFFSET = HEADER_SIZE + DIRTY_BITFIELD_SIZE; // 12288

/** Allocate a minimal SAB large enough for N tracks plus the required header. */
function makeTestSAB(trackCount: number): SharedArrayBuffer {
  const totalBytes = TRACK_DATA_OFFSET + trackCount * RECORD_SIZE;
  const sab = new SharedArrayBuffer(totalBytes);
  // Write max_slots into the header (uint32 index 2)
  const header = new Uint32Array(sab, 0, HEADER_SIZE / 4);
  header[2] = trackCount;
  return sab;
}

// ── R-005 timestamp accuracy ──────────────────────────────────────────────────

describe("writeMockTracksToSAB — update_epoch_ms accuracy", () => {
  it("writes update_epoch_ms within 2 seconds of Date.now()", () => {
    const COUNT = 10;
    const sab = makeTestSAB(COUNT);

    initMockTracks(COUNT);

    const before = Date.now();
    writeMockTracksToSAB(sab, COUNT);
    const after = Date.now();

    const trackData = new DataView(sab, TRACK_DATA_OFFSET);

    for (let i = 0; i < COUNT; i++) {
      const base = i * RECORD_SIZE;
      // update_epoch_ms is at byte offset 0x2c within each record
      const epochMs = trackData.getUint32(base + 0x2c, true);

      // The value is stored as (Date.now() >>> 0) — unsigned 32-bit truncation.
      // Use >>> 0 here too so we get unsigned 32-bit values for comparison.
      const maskedBefore = before >>> 0;
      const maskedAfter  = (after + 2000) >>> 0;

      expect(epochMs).toBeGreaterThanOrEqual(maskedBefore);
      expect(epochMs).toBeLessThanOrEqual(maskedAfter);
    }
  });

  it("update_epoch_ms is NOT close to performance.now() (time-base sanity)", () => {
    const COUNT = 5;
    const sab = makeTestSAB(COUNT);

    initMockTracks(COUNT);
    writeMockTracksToSAB(sab, COUNT);

    const trackData = new DataView(sab, TRACK_DATA_OFFSET);
    const epochMs = trackData.getUint32(0x2c, true); // first track

    // performance.now() returns ms since time origin — typically a few million.
    // Date.now() returns Unix epoch ms — currently ~1.7 trillion masked to u32.
    // The two should differ by a large amount (hundreds of millions of ms).
    // If they are within 60 seconds of each other, the time base is wrong.
    const perfNow = performance.now() | 0;
    expect(Math.abs(epochMs - perfNow)).toBeGreaterThan(60_000);
  });
});

// ── R-007 trail ring buffer update ───────────────────────────────────────────

describe("tickMockTracks — trail ring buffer", () => {
  it("shifts the trail ring buffer so the newest entry reflects the new position", () => {
    const COUNT = 3;
    const sab = makeTestSAB(COUNT);

    initMockTracks(COUNT);

    // Capture SAB state before tick
    writeMockTracksToSAB(sab, COUNT);
    const trackData = new DataView(sab, TRACK_DATA_OFFSET);

    // Read the trail segment[0] lonB / latB (newest endpoint = trail[0])
    // Segment 0 starts at base + 0x30; lonB is at +8, latB at +12
    const lonBefore = trackData.getFloat32(0x30 + 8, true);
    const latBefore = trackData.getFloat32(0x30 + 12, true);

    // Advance tracks by a substantial interval (1 second)
    tickMockTracks(1000);
    writeMockTracksToSAB(sab, COUNT);

    const lonAfter = trackData.getFloat32(0x30 + 8, true);
    const latAfter = trackData.getFloat32(0x30 + 12, true);

    // After 1 second of movement the position must have changed
    const moved = lonBefore !== lonAfter || latBefore !== latAfter;
    expect(moved).toBe(true);
  });

  it("accumulates trail history — segment[1] after two ticks equals segment[0] after one tick", () => {
    const COUNT = 2;
    const sab = makeTestSAB(COUNT);

    initMockTracks(COUNT);
    writeMockTracksToSAB(sab, COUNT);

    const trackData = new DataView(sab, TRACK_DATA_OFFSET);

    // After first tick, capture segment[0] lonB/latB (= trail[0] = newest point)
    tickMockTracks(500);
    writeMockTracksToSAB(sab, COUNT);
    const seg0LonAfterTick1 = trackData.getFloat32(0x30 + 8,  true); // segment[0] lonB
    const seg0LatAfterTick1 = trackData.getFloat32(0x30 + 12, true); // segment[0] latB

    // After second tick, segment[1] lonB/latB should equal the previous segment[0]
    // because the ring shifted: what was trail[0] is now trail[1].
    tickMockTracks(500);
    writeMockTracksToSAB(sab, COUNT);
    const seg1LonAfterTick2 = trackData.getFloat32(0x30 + 16 + 8,  true); // segment[1] lonB
    const seg1LatAfterTick2 = trackData.getFloat32(0x30 + 16 + 12, true); // segment[1] latB

    expect(seg1LonAfterTick2).toBeCloseTo(seg0LonAfterTick1, 5);
    expect(seg1LatAfterTick2).toBeCloseTo(seg0LatAfterTick1, 5);
  });
});
