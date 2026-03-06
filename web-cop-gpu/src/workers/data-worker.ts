// CLASSIFICATION: UNCLASSIFIED
// src/workers/data-worker.ts — Data Worker shell
//
// Responsibilities:
//   1. Accept SharedArrayBuffer via postMessage on init
//   2. Stub WebTransport connection (writes mock 128-byte records to SAB for testing)
//   3. Report connection status back to main thread
//
// Reference: docs/sdlc_guidelines/08_tech_specific/webtransport_guidelines.md §6

const RECORD_SIZE = 128;
const MAX_BACKOFF_MS = 30_000;
const BASE_BACKOFF_MS = 1_000;
const MOCK_INTERVAL_MS = 16; // ~60 Hz mock update rate

interface InitMessage {
  type: "init";
  sab: SharedArrayBuffer;
  url?: string;
}

interface StatusMessage {
  type: "connection_status";
  connected: boolean;
  latency: number;
}

let sab: SharedArrayBuffer | null = null;
let trackData: Uint8Array | null = null;
let maxSlots = 0;
let slotIndex = 0;
let mockIntervalId: ReturnType<typeof setInterval> | null = null;

/**
 * Write a mock 128-byte track record into the SharedArrayBuffer.
 * Used for testing the data flow before WebTransport is available.
 */
function writeMockRecord(slot: number): void {
  if (!trackData || slot >= maxSlots) return;

  const offset = slot * RECORD_SIZE;
  const view = new DataView(trackData.buffer, trackData.byteOffset + offset, RECORD_SIZE);

  // Write mock field values at their documented offsets
  const lon = (Math.random() * 2 - 1) * Math.PI; // random longitude in radians
  const lat = (Math.random() * 2 - 1) * (Math.PI / 2); // random latitude in radians
  view.setFloat32(0x00, lon, true); // longitude
  view.setFloat32(0x04, lat, true); // latitude
  view.setFloat32(0x08, 0, true);   // course
  view.setFloat32(0x0c, 10, true);  // speed m/s
  view.setFloat32(0x10, 1000, true); // altitude meters
  view.setUint32(0x14, slot, true);  // track_id_hash (use slot as mock ID)
  view.setUint32(0x18, 1, true);     // source_bitmap
  view.setUint32(0x1c, 1, true);     // classification_level
  view.setUint32(0x20, 0, true);     // threat_level = Unknown
  view.setUint32(0x24, 0, true);     // icon_index
  view.setUint32(0x28, 0, true);     // alert_flags
  view.setUint32(0x2c, Date.now() & 0xffffffff, true); // update_epoch_ms
}

function startMockUpdates(): void {
  if (mockIntervalId !== null) return;

  mockIntervalId = setInterval(() => {
    if (!trackData) return;
    writeMockRecord(slotIndex % maxSlots);
    slotIndex = (slotIndex + 1) % maxSlots;
  }, MOCK_INTERVAL_MS);

  const status: StatusMessage = {
    type: "connection_status",
    connected: true,
    latency: 0,
  };
  postMessage(status);
}

function stopMockUpdates(): void {
  if (mockIntervalId !== null) {
    clearInterval(mockIntervalId);
    mockIntervalId = null;
  }
}

/**
 * Connect to a WebTransport server with exponential backoff retry.
 * In stub mode (no url), falls back to mock data generation.
 *
 * Reference: docs/sdlc_guidelines/08_tech_specific/webtransport_guidelines.md §6.3
 */
async function connectWithRetry(url: string): Promise<void> {
  let attempt = 0;

  while (true) {
    try {
      await connectWebTransport(url);
    } catch (_err) {
      attempt++;
      const delay = Math.min(BASE_BACKOFF_MS * Math.pow(2, attempt), MAX_BACKOFF_MS);
      const status: StatusMessage = {
        type: "connection_status",
        connected: false,
        latency: -1,
      };
      postMessage(status);
      await sleep(delay);
    }
  }
}

async function connectWebTransport(url: string): Promise<void> {
  const transport = new WebTransport(url);
  await transport.ready;

  const status: StatusMessage = {
    type: "connection_status",
    connected: true,
    latency: 0,
  };
  postMessage(status);

  const reader = transport.datagrams.readable.getReader();

  while (true) {
    const { value: datagram, done } = await reader.read();
    if (done) break;
    if (!trackData || !datagram) continue;

    // TODO(Phase 2): Replace with Rust Wasm decoder
    // For now, copy raw bytes directly to SAB at current slot
    if (datagram.byteLength === RECORD_SIZE) {
      const offset = (slotIndex % maxSlots) * RECORD_SIZE;
      trackData.set(datagram, offset);
      slotIndex++;
    }
  }
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

self.addEventListener("message", (event: MessageEvent<InitMessage>) => {
  const msg = event.data;

  if (msg.type === "init") {
    sab = msg.sab;
    // Track data starts after header (4096 bytes) + dirty bitfield (8192 bytes)
    const TRACK_DATA_OFFSET = 4096 + 8192;
    maxSlots = new Uint32Array(sab, 8, 1)[0]; // read max_slots from header offset 8
    trackData = new Uint8Array(sab, TRACK_DATA_OFFSET);

    if (msg.url) {
      // Real WebTransport connection
      connectWithRetry(msg.url).catch((_err) => {
        const status: StatusMessage = {
          type: "connection_status",
          connected: false,
          latency: -1,
        };
        postMessage(status);
      });
    } else {
      // Stub mode: generate mock data for testing
      startMockUpdates();
    }
  }
});

// Clean up on termination
self.addEventListener("close", () => {
  stopMockUpdates();
});
