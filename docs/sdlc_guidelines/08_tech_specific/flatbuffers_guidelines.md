<!-- CLASSIFICATION: UNCLASSIFIED -->

# FlatBuffers Guidelines

> **Document**: RTSA FlatBuffers Development Guidelines
> **Version**: 1.0
> **Classification**: UNCLASSIFIED
> **Last Updated**: 2026-03-05
> **Prerequisite**: Load `general_coding.md` and `secure_coding.md` first

---

## 1. Overview

FlatBuffers is the **hot-path wire format** in the RTSA WebGPU architecture. Track position updates flow from Go services through Redpanda, are serialized into 128-byte FlatBuffer records, sent via WebTransport datagrams, and decoded in the browser by a Rust Wasm module directly into SharedArrayBuffer — with **zero copies and zero allocations** on the read path.

### Why FlatBuffers (Not Protobuf)

| Concern                  | Protobuf                          | FlatBuffers                        |
| ------------------------ | --------------------------------- | ---------------------------------- |
| **Hot-path decode cost** | Full deserialization + allocation | Zero-copy field access             |
| **GPU alignment**        | Varint encoding, not aligned      | Fixed-size records, GPU-aligned    |
| **Record size**          | Variable (compact)                | Fixed 128 bytes (predictable)      |
| **Schema evolution**     | Full backward compat              | Append-only (fields never removed) |
| **Use in RTSA**          | Cold path (gRPC, commands)        | Hot path (track updates)           |

**Protobuf remains** the schema definition language for gRPC services (cold path). FlatBuffers is used exclusively for the hot-path data pipeline.

---

## 2. Schema Conventions

### 2.1 File Organization

```
proto/
└── rtsa/
    └── flatbuf/
        └── v1/
            ├── track_update.fbs     # Track position update record
            ├── alert_event.fbs      # Alert event for hot-path push
            └── common.fbs           # Shared enums (ThreatLevel, etc.)
```

### 2.2 Naming Rules

| Element    | Convention        | Example                |
| ---------- | ----------------- | ---------------------- |
| File name  | `snake_case.fbs`  | `track_update.fbs`     |
| Namespace  | `rtsa.flatbuf.v1` | Matches directory path |
| Table name | `PascalCase`      | `TrackUpdate`          |
| Field name | `snake_case`      | `track_id_hash`        |
| Enum name  | `PascalCase`      | `ThreatLevel`          |
| Enum value | `PascalCase`      | `Hostile`, `Friendly`  |

### 2.3 Primary Schema — Track Update

```flatbuffers
// CLASSIFICATION: UNCLASSIFIED
// proto/rtsa/flatbuf/v1/track_update.fbs

namespace rtsa.flatbuf.v1;

enum ThreatLevel : uint32 {
  Unknown = 0,
  Pending = 1,
  Friendly = 2,
  Neutral = 3,
  Suspect = 4,
  Hostile = 5,
}

// 128-byte GPU-aligned track record
// IMPORTANT: Field order matches SharedArrayBuffer binary layout
// and WGSL TrackRecord struct — do NOT reorder
table TrackUpdate {
  // Position (20 bytes)
  longitude: float32;           // offset 0x00 — radians
  latitude: float32;            // offset 0x04 — radians
  course: float32;              // offset 0x08 — radians
  speed: float32;               // offset 0x0C — m/s
  altitude: float32;            // offset 0x10 — meters

  // Identity (16 bytes)
  track_id_hash: uint32;        // offset 0x14 — FNV-1a hash of track ID
  source_bitmap: uint32;        // offset 0x18 — sensor source bitmask
  classification_level: uint32; // offset 0x1C — security classification
  threat_level: ThreatLevel;    // offset 0x20 — enum

  // Display (12 bytes)
  icon_index: uint32;           // offset 0x24 — atlas lookup index
  alert_flags: uint32;          // offset 0x28 — alert bitmask
  update_epoch_ms: uint32;      // offset 0x2C — server timestamp

  // Trail ring buffer (80 bytes = 5 × 16 bytes)
  trail_0: [float32:4];         // offset 0x30 — (lon, lat, lon, lat)
  trail_1: [float32:4];         // offset 0x40
  trail_2: [float32:4];         // offset 0x50
  trail_3: [float32:4];         // offset 0x60
  trail_4: [float32:4];         // offset 0x70

  // Total: 128 bytes (aligned to GPU requirements)
}

root_type TrackUpdate;
```

---

## 3. Go Serializer Implementation

### 3.1 Location

The FlatBuffer serializer lives in the Go backend, converting Protobuf track events from Redpanda into FlatBuffer records for WebTransport:

```
pkg/flatbuf/
├── serializer.go           # Protobuf → FlatBuffer conversion
├── serializer_test.go      # Unit tests
└── layout.go               # Constants for field offsets
```

### 3.2 Serialization Pattern

```go
// CLASSIFICATION: UNCLASSIFIED
// pkg/flatbuf/serializer.go

package flatbuf

import (
    flatbuffers "github.com/google/flatbuffers/go"
    trackpb "rtsa/gen/go/rtsa/track/v1"
)

const RecordSize = 128

// SerializeTrackUpdate writes a 128-byte FlatBuffer record into buf.
// Pre-condition: len(buf) >= RecordSize
func SerializeTrackUpdate(builder *flatbuffers.Builder, track *trackpb.TrackUpdate) []byte {
    builder.Reset()

    // Build the FlatBuffer table matching the .fbs schema field order
    TrackUpdateStart(builder)
    TrackUpdateAddLongitude(builder, float32(track.Longitude))
    TrackUpdateAddLatitude(builder, float32(track.Latitude))
    TrackUpdateAddCourse(builder, float32(track.Course))
    TrackUpdateAddSpeed(builder, float32(track.Speed))
    TrackUpdateAddAltitude(builder, float32(track.Altitude))
    TrackUpdateAddTrackIdHash(builder, fnv1aHash(track.TrackId))
    TrackUpdateAddSourceBitmap(builder, track.SourceBitmap)
    TrackUpdateAddClassificationLevel(builder, track.ClassificationLevel)
    TrackUpdateAddThreatLevel(builder, ThreatLevel(track.ThreatLevel))
    TrackUpdateAddIconIndex(builder, track.IconIndex)
    TrackUpdateAddAlertFlags(builder, track.AlertFlags)
    TrackUpdateAddUpdateEpochMs(builder, uint32(track.UpdateEpochMs))
    // Trail points...
    offset := TrackUpdateEnd(builder)
    builder.Finish(offset)

    return builder.FinishedBytes()
}
```

### 3.3 Builder Reuse

```go
// Reuse builder across serializations to avoid allocations
builder := flatbuffers.NewBuilder(256)
for _, track := range batch {
    data := SerializeTrackUpdate(builder, track)
    // Send via WebTransport datagram
    conn.SendDatagram(data)
}
```

**Rule**: Always reuse `flatbuffers.Builder` — never allocate per-message.

---

## 4. Rust Wasm Decoder

### 4.1 Location

```
web-cop-gpu/wasm-decoder/
├── Cargo.toml
├── src/
│   └── lib.rs              # Zero-copy FlatBuffer decoder
└── tests/
    └── decode_test.rs
```

### 4.2 Decoder Pattern

```rust
// CLASSIFICATION: UNCLASSIFIED
// Zero-allocation FlatBuffer → SharedArrayBuffer decoder

use wasm_bindgen::prelude::*;

const RECORD_SIZE: usize = 128;

/// Decode a FlatBuffer datagram and write the 128-byte record
/// directly into the SharedArrayBuffer at the specified slot.
///
/// # Safety
/// - `sab_ptr` must point to a valid SharedArrayBuffer region
/// - `slot_index * RECORD_SIZE` must be within bounds
#[wasm_bindgen]
pub fn decode_track_update(
    datagram: &[u8],
    sab_ptr: *mut u8,
    slot_index: usize,
    max_slots: usize,
) -> bool {
    if slot_index >= max_slots {
        return false;
    }

    // FlatBuffer zero-copy access — no allocation
    let track = flatbuffers::root::<TrackUpdate>(datagram);
    let track = match track {
        Ok(t) => t,
        Err(_) => return false,
    };

    let offset = slot_index * RECORD_SIZE;
    // Safe: bounds checked above
    let dest = unsafe {
        std::slice::from_raw_parts_mut(sab_ptr.add(offset), RECORD_SIZE)
    };

    // Write fields in binary layout order
    write_f32(dest, 0x00, track.longitude());
    write_f32(dest, 0x04, track.latitude());
    write_f32(dest, 0x08, track.course());
    write_f32(dest, 0x0C, track.speed());
    write_f32(dest, 0x10, track.altitude());
    write_u32(dest, 0x14, track.track_id_hash());
    // ... remaining fields ...

    true
}
```

### 4.3 Build Rules

```toml
# Cargo.toml
[lib]
crate-type = ["cdylib"]

[dependencies]
wasm-bindgen = "0.2"
flatbuffers = "24.3"

[profile.release]
opt-level = "s"       # Optimize for size
lto = true            # Link-time optimization
strip = true          # Strip debug symbols
```

**Build command**: `wasm-pack build --target web --release`

---

## 5. Schema ↔ Protobuf Synchronization

### 5.1 Single Source of Truth

The **Protobuf schema** (`proto/rtsa/track/v1/track.proto`) is the canonical definition. The FlatBuffer schema (`proto/rtsa/flatbuf/v1/track_update.fbs`) is a **derived, GPU-optimized projection**.

### 5.2 Synchronization Rules

| Rule                                                       | Rationale                                                  |
| ---------------------------------------------------------- | ---------------------------------------------------------- |
| FlatBuffer schema is a **subset** of Protobuf fields       | Only hot-path fields are projected                         |
| Field semantics must match exactly                         | Same units, same encoding (radians, m/s, etc.)             |
| Adding a Protobuf field does NOT require FlatBuffer change | Only add to FlatBuffer if the GPU needs it                 |
| Adding a FlatBuffer field requires Protobuf equivalent     | Every FlatBuffer field must have a Protobuf source         |
| Hash functions must be identical in Go and Rust            | `track_id_hash` uses FNV-1a in both serializer and decoder |

### 5.3 Mapping Table

| FlatBuffer Field | Protobuf Source                           | Transform                           |
| ---------------- | ----------------------------------------- | ----------------------------------- |
| `longitude`      | `TrackUpdate.longitude`                   | Cast to float32, convert to radians |
| `latitude`       | `TrackUpdate.latitude`                    | Cast to float32, convert to radians |
| `track_id_hash`  | `TrackUpdate.track_id`                    | FNV-1a hash (string → u32)          |
| `source_bitmap`  | `TrackUpdate.sources[]`                   | Bitwise OR of source enum values    |
| `icon_index`     | `TrackUpdate.track_type` + `threat_level` | Lookup table → atlas index          |
| `trail_*`        | Last 5 `TrackUpdate.position_history[]`   | Projected to (lon, lat) pairs       |

---

## 6. Schema Evolution

### 6.1 Allowed Changes

| Change               | Allowed?   | Notes                                                                             |
| -------------------- | ---------- | --------------------------------------------------------------------------------- |
| Add new field at end | ✅ Yes     | Default value used by old decoders                                                |
| Add new enum value   | ✅ Yes     | Old parsers treat as 0 (Unknown)                                                  |
| Remove a field       | ❌ No      | Breaks binary layout                                                              |
| Reorder fields       | ❌ No      | Breaks offset assumptions in Wasm decoder                                         |
| Change field type    | ❌ No      | Breaks binary read                                                                |
| Change record size   | ⚠️ Careful | Requires updating Go serializer, Wasm decoder, GPU buffer stride, and WGSL struct |

### 6.2 Version Bump Process

If the record layout must change (e.g., expanding beyond 128 bytes):

1. Create `v2/track_update.fbs` with new layout
2. Update Go serializer to produce v2
3. Update Rust Wasm decoder to read v2
4. Update WGSL `TrackRecord` struct
5. Update SharedArrayBuffer slot size
6. Run full integration tests
7. Deploy backend first, then browser (decoder handles both versions during rollout)

---

## 7. Testing

### 7.1 Go Serializer Tests

```go
// CLASSIFICATION: UNCLASSIFIED
func TestSerializeTrackUpdate_RoundTrip(t *testing.T) {
    builder := flatbuffers.NewBuilder(256)
    track := &trackpb.TrackUpdate{
        Longitude: -73.5,
        Latitude:  45.5,
        Speed:     12.5,
        TrackId:   "TRACK-001",
    }

    data := SerializeTrackUpdate(builder, track)
    assert.Equal(t, RecordSize, len(data))

    // Verify FlatBuffer can be read back
    fb := flatbuffers.GetRootAs(data, 0)
    assert.InDelta(t, -73.5, fb.Longitude(), 0.001)
}
```

### 7.2 Rust Decoder Tests

```rust
// CLASSIFICATION: UNCLASSIFIED
#[test]
fn test_decode_track_update_writes_correct_values() {
    let datagram = build_test_datagram(/* ... */);
    let mut sab = vec![0u8; 128 * 10]; // 10 slots
    let ok = decode_track_update(&datagram, sab.as_mut_ptr(), 0, 10);
    assert!(ok);

    let lon = f32::from_le_bytes(sab[0..4].try_into().unwrap());
    assert!((lon - (-73.5_f32)).abs() < 0.001);
}
```

### 7.3 Cross-Language Integration Test

A CI test that:

1. Go serializer produces a FlatBuffer record
2. Wasm decoder reads it into a mock SharedArrayBuffer
3. Values match the original Protobuf source

This lives in `tests/integration/flatbuf_roundtrip_test.go`.

---

## 8. Performance Rules

| Rule                                   | Rationale                                       |
| -------------------------------------- | ----------------------------------------------- |
| Reuse `flatbuffers.Builder`            | Avoid per-message heap allocation               |
| Never decode FlatBuffers in JavaScript | Wasm decoder is 10x faster                      |
| 128-byte fixed record size             | Predictable memory layout, GPU-aligned          |
| No strings in FlatBuffer hot path      | Strings require offset resolution; use hashes   |
| Batch serialization                    | Serialize up to 64 records per Redpanda message |

---

## 9. Cross-References

| Document                      | Path                                                               |
| ----------------------------- | ------------------------------------------------------------------ |
| WebTransport Guidelines       | `docs/sdlc_guidelines/08_tech_specific/webtransport_guidelines.md` |
| WebGPU Guidelines             | `docs/sdlc_guidelines/08_tech_specific/webgpu_guidelines.md`       |
| WGSL Shader Standards         | `docs/sdlc_guidelines/08_tech_specific/wgsl_shader_standards.md`   |
| Data Architecture — Hot Path  | `docs/architecture/data_architecture.md` §12                       |
| v1 Architecture — Wire Format | `docs/architecture/v1/RTSA_WebGPU_Architecture_v1.md` §3           |
| Go Standards                  | `docs/sdlc_guidelines/04_coding_standards/go_standards.md`         |
