// CLASSIFICATION: UNCLASSIFIED
// pkg/flatbuf/layout.go — Binary record field offset constants
//
// These byte offsets define the 128-byte GPU-aligned binary record layout
// shared by:
//   - Go FlatBuffer serializer (this package)
//   - Rust Wasm decoder (web-cop-gpu/wasm-decoder/src/lib.rs)
//   - WGSL TrackRecord struct (web-cop-gpu/src/shaders/)
//   - SharedArrayBuffer slot (web-cop-gpu/src/services/sab.ts)
//
// WARNING: Do NOT change any offset — doing so breaks the GPU pipeline.
// Reference: docs/sdlc_guidelines/08_tech_specific/flatbuffers_guidelines.md §2.3

package flatbuf

// RecordSize is the fixed size in bytes of one track record.
// Records are aligned to GPU 128-byte cache lines.
const RecordSize = 128

// MaxBatchSize is the maximum number of records per WebTransport datagram.
// 9 × 128 = 1152 bytes, safely below the 1200-byte QUIC MTU.
const MaxBatchSize = 9

// Field byte offsets within the 128-byte GPU-aligned record.
// These MUST match the Rust offsets in wasm-decoder/src/lib.rs.
const (
// Position fields (20 bytes)
OffLongitude = 0x00 // float32 — radians, WGS-84
OffLatitude  = 0x04 // float32 — radians, WGS-84
OffCourse    = 0x08 // float32 — radians, true north
OffSpeed     = 0x0C // float32 — m/s
OffAltitude  = 0x10 // float32 — meters AMSL

// Identity fields (16 bytes)
OffTrackIDHash         = 0x14 // uint32 — FNV-1a hash of track UUID
OffSourceBitmap        = 0x18 // uint32 — sensor source bitmask
OffClassificationLevel = 0x1C // uint32 — ClassificationLevel enum value
OffThreatLevel         = 0x20 // uint32 — ThreatLevel enum value

// Display fields (12 bytes)
OffIconIndex      = 0x24 // uint32 — atlas icon lookup index
OffAlertFlags     = 0x28 // uint32 — alert bitmask
OffUpdateEpochMs  = 0x2C // uint32 — server timestamp (ms, lower 32 bits)

// Trail ring buffer — 5 × 16 bytes = 80 bytes
// Each slot: (lon_a f32, lat_a f32, lon_b f32, lat_b f32)
OffTrail0 = 0x30 // most recent segment
OffTrail1 = 0x40
OffTrail2 = 0x50
OffTrail3 = 0x60
OffTrail4 = 0x70 // oldest segment
)

// TrailSlotSize is the size in bytes of each trail ring-buffer entry.
const TrailSlotSize = 16 // 4 × float32

// NumTrailSlots is the number of trail history entries per record.
const NumTrailSlots = 5

// ThreatLevel colour mapping (matches WGSL threat_color() function in trail.wgsl
// and the halo colour switch in halos.wgsl). Conforms to STANAG APP-6.
//
//	0 = Unknown  → grey
//	1 = Pending  → blue
//	2 = Friendly → green
//	3 = Neutral  → green   ← APP-6: neutral is green, not amber
//	4 = Suspect  → amber/orange
//	5 = Hostile  → red
