// CLASSIFICATION: UNCLASSIFIED
// src/services/sab.ts — SharedArrayBuffer ring buffer allocation
//
// Layout (per docs/architecture/data_architecture.md §12.3):
//
//  ┌───────────────────────────────────────────────────────┐
//  │ Header (4096 bytes)                                   │
//  │   [0..3]    uint32  active_track_count                │
//  │   [4..7]    uint32  write_generation                  │
//  │   [8..11]   uint32  max_slots                         │
//  │   [12..4095] reserved                                 │
//  ├───────────────────────────────────────────────────────┤
//  │ Dirty Bitfield (8192 bytes)                           │
//  │   1 bit per slot × 50,000 ≈ 6,250 bytes padded       │
//  ├───────────────────────────────────────────────────────┤
//  │ Track Data (MAX_SLOTS × 128 bytes)                    │
//  └───────────────────────────────────────────────────────┘

export const RECORD_SIZE = 128;
export const MAX_SLOTS = 65_536; // default 65,536 slots ≈ 8 MB track data
export const HEADER_SIZE = 4096;
export const DIRTY_BITFIELD_SIZE = 8192;

export const HEADER_OFFSET_ACTIVE_TRACK_COUNT = 0;
export const HEADER_OFFSET_WRITE_GENERATION = 1;
export const HEADER_OFFSET_MAX_SLOTS = 2;

export const TRACK_DATA_OFFSET = HEADER_SIZE + DIRTY_BITFIELD_SIZE; // 12288

export interface RingBuffer {
  sab: SharedArrayBuffer;
  header: Uint32Array;
  trackData: Uint8Array;
  maxSlots: number;
}

/**
 * Allocate and initialise the SharedArrayBuffer ring buffer.
 * Returns a RingBuffer handle containing typed array views.
 *
 * Total size: 4096 (header) + 8192 (dirty bitfield) + MAX_SLOTS * 128 (tracks)
 */
export function allocateSAB(maxSlots: number = MAX_SLOTS): RingBuffer {
  const totalBytes = HEADER_SIZE + DIRTY_BITFIELD_SIZE + maxSlots * RECORD_SIZE;
  const sab = new SharedArrayBuffer(totalBytes);

  const header = new Uint32Array(sab, 0, HEADER_SIZE / 4);
  header[HEADER_OFFSET_ACTIVE_TRACK_COUNT] = 0;
  header[HEADER_OFFSET_WRITE_GENERATION] = 0;
  header[HEADER_OFFSET_MAX_SLOTS] = maxSlots;

  const trackData = new Uint8Array(sab, TRACK_DATA_OFFSET, maxSlots * RECORD_SIZE);

  return { sab, header, trackData, maxSlots };
}

/**
 * Write a 128-byte record into the ring buffer at the given slot index.
 * Returns false if the slot index is out of bounds.
 */
export function writeSlot(
  buf: RingBuffer,
  slotIndex: number,
  record: Uint8Array,
): boolean {
  if (slotIndex >= buf.maxSlots) return false;
  if (record.byteLength !== RECORD_SIZE) return false;

  const offset = slotIndex * RECORD_SIZE;
  buf.trackData.set(record, offset);
  return true;
}

/**
 * Read a 128-byte record from the ring buffer at the given slot index.
 * Returns null if the slot index is out of bounds.
 */
export function readSlot(buf: RingBuffer, slotIndex: number): Uint8Array | null {
  if (slotIndex >= buf.maxSlots) return null;

  const offset = slotIndex * RECORD_SIZE;
  return buf.trackData.subarray(offset, offset + RECORD_SIZE);
}
