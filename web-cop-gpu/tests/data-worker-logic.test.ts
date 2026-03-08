// CLASSIFICATION: UNCLASSIFIED
// tests/data-worker-logic.test.ts — Unit tests for data-worker pure logic
//
// Verifies:
//   - buildTransportUrl correctly appends the JWT token query parameter
//   - getJwtExpiryMs parses valid JWT exp claims and returns -1 for invalid input
//   - writeMockRecord writes correct field values to a Uint8Array slot
//   - writeRecordToSlot copies record bytes with and without a Wasm decoder
//   - processDatagram correctly decodes batched records and tracks slot index

import { describe, it, expect, vi } from "vitest";
import {
  buildTransportUrl,
  getJwtExpiryMs,
  writeMockRecord,
  writeRecordToSlot,
  processDatagram,
  RECORD_SIZE,
  type WasmDecoder,
} from "../src/workers/data-worker-logic";

// ── buildTransportUrl ─────────────────────────────────────────────────────────

describe("buildTransportUrl", () => {
  it("appends token as query parameter when token is provided", () => {
    expect(buildTransportUrl("https://rtsa.mil.ca:4443/wt", "my.jwt.token")).toBe(
      "https://rtsa.mil.ca:4443/wt?token=my.jwt.token",
    );
  });

  it("returns the bare URL when token is undefined", () => {
    expect(buildTransportUrl("https://rtsa.mil.ca:4443/wt", undefined)).toBe(
      "https://rtsa.mil.ca:4443/wt",
    );
  });

  it("appends token with & when URL already has a query string", () => {
    expect(buildTransportUrl("https://rtsa.mil.ca:4443/wt?op=1", "tok")).toBe(
      "https://rtsa.mil.ca:4443/wt?op=1&token=tok",
    );
  });
});

// ── getJwtExpiryMs ────────────────────────────────────────────────────────────

function makeJwt(payload: Record<string, unknown>): string {
  const header = btoa(JSON.stringify({ alg: "HS256", typ: "JWT" }));
  const body = btoa(JSON.stringify(payload));
  return `${header}.${body}.fake-signature`;
}

describe("getJwtExpiryMs", () => {
  it("returns exp × 1000 for a valid JWT", () => {
    const exp = 1_800_000_000; // seconds
    const token = makeJwt({ sub: "op1", exp });
    expect(getJwtExpiryMs(token)).toBe(exp * 1000);
  });

  it("returns -1 for a token with no exp claim", () => {
    const token = makeJwt({ sub: "op1" });
    expect(getJwtExpiryMs(token)).toBe(-1);
  });

  it("returns -1 for a malformed token (not 3 parts)", () => {
    expect(getJwtExpiryMs("not.a.valid.jwt.token")).toBe(-1);
    expect(getJwtExpiryMs("twosections")).toBe(-1);
  });

  it("returns -1 for a token with non-numeric exp", () => {
    const token = makeJwt({ sub: "op1", exp: "not-a-number" });
    expect(getJwtExpiryMs(token)).toBe(-1);
  });

  it("returns -1 for invalid base64 payload", () => {
    expect(getJwtExpiryMs("header.!!!invalid!!!.sig")).toBe(-1);
  });
});

// ── writeMockRecord ───────────────────────────────────────────────────────────

describe("writeMockRecord", () => {
  it("writes a non-zero track_id_hash (= slot index) at offset 0x14", () => {
    const maxSlots = 4;
    const trackData = new Uint8Array(maxSlots * RECORD_SIZE);
    const slot = 2;
    writeMockRecord(trackData, slot, maxSlots);

    const view = new DataView(trackData.buffer, slot * RECORD_SIZE, RECORD_SIZE);
    const trackIdHash = view.getUint32(0x14, true);
    expect(trackIdHash).toBe(slot); // track_id_hash is set to slot index
  });

  it("writes speed = 10 m/s at offset 0x0c", () => {
    const maxSlots = 2;
    const trackData = new Uint8Array(maxSlots * RECORD_SIZE);
    writeMockRecord(trackData, 0, maxSlots);

    const view = new DataView(trackData.buffer, 0, RECORD_SIZE);
    expect(view.getFloat32(0x0c, true)).toBe(10);
  });

  it("writes altitude = 1000m at offset 0x10", () => {
    const maxSlots = 2;
    const trackData = new Uint8Array(maxSlots * RECORD_SIZE);
    writeMockRecord(trackData, 0, maxSlots);

    const view = new DataView(trackData.buffer, 0, RECORD_SIZE);
    expect(view.getFloat32(0x10, true)).toBe(1000);
  });

  it("returns false for out-of-bounds slot", () => {
    const maxSlots = 2;
    const trackData = new Uint8Array(maxSlots * RECORD_SIZE);
    expect(writeMockRecord(trackData, 2, maxSlots)).toBe(false);
    expect(writeMockRecord(trackData, 99, maxSlots)).toBe(false);
  });

  it("does not write to neighbouring slots", () => {
    const maxSlots = 4;
    const trackData = new Uint8Array(maxSlots * RECORD_SIZE);
    writeMockRecord(trackData, 1, maxSlots);

    // Slot 0 should be untouched
    const slot0Sum = trackData.slice(0, RECORD_SIZE).reduce((a, b) => a + b, 0);
    expect(slot0Sum).toBe(0);
  });
});

// ── writeRecordToSlot ─────────────────────────────────────────────────────────

describe("writeRecordToSlot", () => {
  it("copies record bytes to the correct slot offset (no wasm decoder)", () => {
    const maxSlots = 4;
    const trackData = new Uint8Array(maxSlots * RECORD_SIZE);
    const record = new Uint8Array(RECORD_SIZE);
    for (let i = 0; i < RECORD_SIZE; i++) record[i] = i & 0xff;

    expect(writeRecordToSlot(trackData, 2, maxSlots, record)).toBe(true);

    const written = trackData.slice(2 * RECORD_SIZE, 3 * RECORD_SIZE);
    expect(Array.from(written)).toEqual(Array.from(record));
  });

  it("returns false for out-of-bounds slot index", () => {
    const maxSlots = 2;
    const trackData = new Uint8Array(maxSlots * RECORD_SIZE);
    const record = new Uint8Array(RECORD_SIZE);
    expect(writeRecordToSlot(trackData, 2, maxSlots, record)).toBe(false);
    expect(writeRecordToSlot(trackData, 100, maxSlots, record)).toBe(false);
  });

  it("returns false when record is shorter than RECORD_SIZE", () => {
    const maxSlots = 2;
    const trackData = new Uint8Array(maxSlots * RECORD_SIZE);
    const shortRecord = new Uint8Array(64);
    expect(writeRecordToSlot(trackData, 0, maxSlots, shortRecord)).toBe(false);
  });

  it("delegates to the Wasm decoder when provided", () => {
    const maxSlots = 4;
    const trackData = new Uint8Array(maxSlots * RECORD_SIZE);
    const record = new Uint8Array(RECORD_SIZE).fill(0xab);

    const mockDecoder: WasmDecoder = {
      write_slot_from_datagram: vi.fn(() => true),
    };

    const result = writeRecordToSlot(trackData, 0, maxSlots, record, mockDecoder);
    expect(result).toBe(true);
    expect(mockDecoder.write_slot_from_datagram).toHaveBeenCalledWith(
      trackData,
      0,
      maxSlots,
      record,
    );
  });

  it("falls back to direct copy when wasmDecoder is null", () => {
    const maxSlots = 2;
    const trackData = new Uint8Array(maxSlots * RECORD_SIZE);
    const record = new Uint8Array(RECORD_SIZE).fill(0xcd);

    expect(writeRecordToSlot(trackData, 0, maxSlots, record, null)).toBe(true);
    expect(trackData[0]).toBe(0xcd);
  });
});

// ── processDatagram ───────────────────────────────────────────────────────────

describe("processDatagram", () => {
  it("decodes a single-record datagram and advances slotIndex by 1", () => {
    const maxSlots = 10;
    const trackData = new Uint8Array(maxSlots * RECORD_SIZE);
    const datagram = new Uint8Array(RECORD_SIZE).fill(0x42);

    const result = processDatagram(datagram, trackData, 0, maxSlots, 9);

    expect(result.decoded).toBe(1);
    expect(result.errors).toBe(0);
    expect(result.nextSlotIndex).toBe(1);
    expect(trackData[0]).toBe(0x42);
  });

  it("decodes a multi-record datagram (batch of 3)", () => {
    const maxSlots = 10;
    const trackData = new Uint8Array(maxSlots * RECORD_SIZE);
    const datagram = new Uint8Array(3 * RECORD_SIZE);
    for (let i = 0; i < 3; i++) {
      datagram.fill(i + 1, i * RECORD_SIZE, (i + 1) * RECORD_SIZE);
    }

    const result = processDatagram(datagram, trackData, 0, maxSlots, 9);

    expect(result.decoded).toBe(3);
    expect(result.errors).toBe(0);
    expect(result.nextSlotIndex).toBe(3);
    expect(trackData[0 * RECORD_SIZE]).toBe(1);
    expect(trackData[1 * RECORD_SIZE]).toBe(2);
    expect(trackData[2 * RECORD_SIZE]).toBe(3);
  });

  it("caps batch size at maxBatchSize even for a larger datagram", () => {
    const maxSlots = 20;
    const trackData = new Uint8Array(maxSlots * RECORD_SIZE);
    const datagram = new Uint8Array(10 * RECORD_SIZE); // 10 records

    const result = processDatagram(datagram, trackData, 0, maxSlots, 9); // max 9

    expect(result.decoded).toBe(9);
  });

  it("wraps slot index around maxSlots", () => {
    const maxSlots = 3;
    const trackData = new Uint8Array(maxSlots * RECORD_SIZE);
    const datagram = new Uint8Array(2 * RECORD_SIZE);

    const result = processDatagram(datagram, trackData, 2, maxSlots, 9);

    expect(result.decoded).toBe(2);
    // slotIndex starts at 2, writes to slot 2 (2%3), then wraps to slot 0 (3%3)
    expect(result.nextSlotIndex).toBe(1);
  });

  it("returns decoded=0 for an empty datagram", () => {
    const maxSlots = 4;
    const trackData = new Uint8Array(maxSlots * RECORD_SIZE);
    const datagram = new Uint8Array(0);

    const result = processDatagram(datagram, trackData, 0, maxSlots, 9);

    expect(result.decoded).toBe(0);
    expect(result.errors).toBe(0);
    expect(result.nextSlotIndex).toBe(0);
  });
});
