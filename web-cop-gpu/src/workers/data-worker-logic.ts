// CLASSIFICATION: UNCLASSIFIED
// src/workers/data-worker-logic.ts — Pure data-worker logic extracted for unit testing
//
// Contains stateless helper functions extracted from data-worker.ts.
// These functions have no dependency on Worker globals (self, WebTransport, etc.)
// and can be imported and tested in the jsdom Vitest environment.
//
// Reference: docs/sdlc_guidelines/08_tech_specific/webtransport_guidelines.md §6

/** Fixed size in bytes of one 128-byte GPU-aligned track record. */
export const RECORD_SIZE = 128;

/**
 * Build the authenticated WebTransport URL by appending the JWT as a query
 * parameter. The Go server validates the `token` query param per auth.go §7.1.
 *
 * IMPORTANT: Do not log the constructed URL — it contains the token. (SDLC Rule 5)
 */
export function buildTransportUrl(url: string, token: string | undefined): string {
  if (!token) return url;
  const separator = url.includes("?") ? "&" : "?";
  return `${url}${separator}token=${token}`;
}

/**
 * Decode the `exp` claim from a JWT payload without verifying the signature.
 * Returns the expiry as a Unix timestamp in milliseconds, or -1 if unparseable.
 */
export function getJwtExpiryMs(token: string): number {
  try {
    const parts = token.split(".");
    if (parts.length !== 3 || !parts[1]) return -1;
    // JWT payloads are Base64URL-encoded; normalize to standard Base64 for atob().
    let base64 = parts[1].replace(/-/g, "+").replace(/_/g, "/");
    const padding = base64.length % 4;
    if (padding !== 0) {
      base64 += "=".repeat(4 - padding);
    }
    const payload = JSON.parse(atob(base64)) as { exp?: number };
    if (typeof payload.exp !== "number") return -1;
    return payload.exp * 1000;
  } catch {
    return -1;
  }
}

/**
 * Write a mock 128-byte track record into the given Uint8Array at the given
 * slot index. Used for testing the data flow without a real WebTransport server.
 *
 * Returns false if the slot is out-of-bounds.
 *
 * Field offsets match wasm-decoder/src/lib.rs offsets module.
 */
export function writeMockRecord(
  trackData: Uint8Array,
  slot: number,
  maxSlots: number,
): boolean {
  if (slot >= maxSlots) return false;
  const offset = slot * RECORD_SIZE;
  if (offset + RECORD_SIZE > trackData.byteLength) return false;

  const view = new DataView(trackData.buffer, trackData.byteOffset + offset, RECORD_SIZE);

  const lon = (Math.random() * 2 - 1) * Math.PI;
  const lat = (Math.random() * 2 - 1) * (Math.PI / 2);
  view.setFloat32(0x00, lon, true);                           // longitude
  view.setFloat32(0x04, lat, true);                           // latitude
  view.setFloat32(0x08, 0, true);                             // course
  view.setFloat32(0x0c, 10, true);                            // speed m/s
  view.setFloat32(0x10, 1000, true);                          // altitude meters
  view.setUint32(0x14, slot, true);                           // track_id_hash
  // Source Bitmap: 1=Radar, 4=SIGINT (0x10), 5=Satellite (0x20), 3=EW (0x08)
  // We'll randomize bits 0-6
  const sourceBits = Math.floor(Math.random() * 128);
  const threatLvl = Math.floor(Math.random() * 6); // 0-5
  view.setUint32(0x18, sourceBits, true);                     // source_bitmap
  view.setUint32(0x1c, 1, true);                              // classification_level
  view.setUint32(0x20, threatLvl, true);                      // threat_level
  view.setUint32(0x24, 0, true);                              // icon_index
  view.setUint32(0x28, 0, true);                              // alert_flags
  view.setUint32(0x2c, Date.now() & 0xffffffff, true);        // update_epoch_ms
  return true;
}

/**
 * Write a single 128-byte record into the SAB track-data region at the given slot.
 * Uses the Rust Wasm decoder if available, otherwise falls back to direct copy.
 *
 * Returns false if the slot is out of range or the record is too short.
 */
export interface WasmDecoder {
  write_slot_from_datagram(
    sab_slice: Uint8Array,
    slot_index: number,
    max_slots: number,
    datagram: Uint8Array,
  ): boolean;
  get_fusion_stats(
    sab_slice: Uint8Array,
    active_track_count: number,
    current_time_ms: number,
  ): any; // FusionStats (wasm-bindgen class)
}

export function writeRecordToSlot(
  trackData: Uint8Array,
  slot: number,
  maxSlots: number,
  record: Uint8Array,
  wasmDecoder?: WasmDecoder | null,
): boolean {
  if (slot >= maxSlots) return false;
  if (record.byteLength < RECORD_SIZE) return false;

  if (wasmDecoder) {
    return wasmDecoder.write_slot_from_datagram(trackData, slot, maxSlots, record);
  }

  // Fallback: direct copy at documented SAB offset
  const offset = slot * RECORD_SIZE;
  if (offset + RECORD_SIZE > trackData.byteLength) return false;
  trackData.set(record.subarray(0, RECORD_SIZE), offset);
  return true;
}

/**
 * Process a batch datagram containing 1–MAX_BATCH_SIZE concatenated 128-byte
 * records. Each record is decoded and written to successive SAB slots.
 *
 * Returns the number of records successfully decoded.
 */
export function processDatagram(
  datagram: Uint8Array,
  trackData: Uint8Array,
  slotIndex: number,
  maxSlots: number,
  maxBatchSize: number,
  wasmDecoder?: WasmDecoder | null,
): { decoded: number; errors: number; nextSlotIndex: number } {
  const recordCount = Math.min(
    Math.floor(datagram.byteLength / RECORD_SIZE),
    maxBatchSize,
  );

  let decoded = 0;
  let errors = 0;
  let idx = slotIndex;

  for (let i = 0; i < recordCount; i++) {
    const recordView = new Uint8Array(
      datagram.buffer,
      datagram.byteOffset + i * RECORD_SIZE,
      RECORD_SIZE,
    );
    const slot = idx % maxSlots;
    if (writeRecordToSlot(trackData, slot, maxSlots, recordView, wasmDecoder)) {
      idx = (idx + 1) % maxSlots;
      decoded++;
    } else {
      errors++;
    }
  }

  return { decoded, errors, nextSlotIndex: idx };
}
