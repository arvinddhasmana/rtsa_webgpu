// CLASSIFICATION: UNCLASSIFIED
// tests/sab.test.ts — Unit tests for SharedArrayBuffer ring buffer allocation
//
// Verifies:
//   - Correct total byte length
//   - Header layout (max_slots written correctly)
//   - writeSlot / readSlot round-trip
//   - Bounds checking on writeSlot / readSlot

import { describe, it, expect } from "vitest";
import {
  allocateSAB,
  writeSlot,
  readSlot,
  RECORD_SIZE,
  HEADER_SIZE,
  DIRTY_BITFIELD_SIZE,
  HEADER_OFFSET_ACTIVE_TRACK_COUNT,
  HEADER_OFFSET_MAX_SLOTS,
  TRACK_DATA_OFFSET,
} from "../src/services/sab";

describe("allocateSAB", () => {
  it("allocates the correct total byte length", () => {
    const maxSlots = 100;
    const expected = HEADER_SIZE + DIRTY_BITFIELD_SIZE + maxSlots * RECORD_SIZE;
    const buf = allocateSAB(maxSlots);
    expect(buf.sab.byteLength).toBe(expected);
  });

  it("writes max_slots into the header at the correct offset", () => {
    const maxSlots = 256;
    const buf = allocateSAB(maxSlots);
    expect(buf.header[HEADER_OFFSET_MAX_SLOTS]).toBe(maxSlots);
  });

  it("initialises active_track_count to 0", () => {
    const buf = allocateSAB(64);
    expect(buf.header[HEADER_OFFSET_ACTIVE_TRACK_COUNT]).toBe(0);
  });

  it("returns maxSlots matching the argument", () => {
    const buf = allocateSAB(512);
    expect(buf.maxSlots).toBe(512);
  });

  it("track data starts at TRACK_DATA_OFFSET", () => {
    const maxSlots = 10;
    const buf = allocateSAB(maxSlots);
    expect(buf.trackData.byteOffset).toBe(TRACK_DATA_OFFSET);
  });

  it("track data has correct byte length", () => {
    const maxSlots = 10;
    const buf = allocateSAB(maxSlots);
    expect(buf.trackData.byteLength).toBe(maxSlots * RECORD_SIZE);
  });
});

describe("writeSlot / readSlot", () => {
  it("round-trips a 128-byte record correctly", () => {
    const buf = allocateSAB(4);
    const record = new Uint8Array(RECORD_SIZE);
    for (let i = 0; i < RECORD_SIZE; i++) record[i] = i & 0xff;

    const wrote = writeSlot(buf, 0, record);
    expect(wrote).toBe(true);

    const read = readSlot(buf, 0);
    expect(read).not.toBeNull();
    expect(Array.from(read!)).toEqual(Array.from(record));
  });

  it("writes to correct offset for slot > 0", () => {
    const buf = allocateSAB(8);
    const record = new Uint8Array(RECORD_SIZE).fill(0xab);

    writeSlot(buf, 3, record);
    const read = readSlot(buf, 3);

    expect(read).not.toBeNull();
    expect(read![0]).toBe(0xab);
    expect(read![RECORD_SIZE - 1]).toBe(0xab);

    // Slot 0 should be untouched
    const slot0 = readSlot(buf, 0);
    expect(slot0![0]).toBe(0x00);
  });

  it("writeSlot returns false for out-of-bounds slot index", () => {
    const buf = allocateSAB(4);
    const record = new Uint8Array(RECORD_SIZE);
    expect(writeSlot(buf, 4, record)).toBe(false);
    expect(writeSlot(buf, 100, record)).toBe(false);
  });

  it("writeSlot returns false for wrong record size", () => {
    const buf = allocateSAB(4);
    const short = new Uint8Array(64);
    expect(writeSlot(buf, 0, short)).toBe(false);
  });

  it("readSlot returns null for out-of-bounds slot index", () => {
    const buf = allocateSAB(4);
    expect(readSlot(buf, 4)).toBeNull();
    expect(readSlot(buf, 999)).toBeNull();
  });
});
