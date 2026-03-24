// CLASSIFICATION: UNCLASSIFIED
// wasm-decoder/src/lib.rs — Zero-copy FlatBuffer decoder for RTSA track updates
//
// Decodes a 128-byte FlatBuffer record and writes it directly into a caller-
// supplied byte slice (representing a SharedArrayBuffer slot).
//
// Reference: docs/sdlc_guidelines/08_tech_specific/flatbuffers_guidelines.md §4

use wasm_bindgen::prelude::*;

pub const RECORD_SIZE: usize = 128;

// Field byte offsets within the 128-byte GPU-aligned record.
// These MUST match the WGSL TrackRecord struct and the .fbs schema.
// Reference: docs/sdlc_guidelines/08_tech_specific/webgpu_guidelines.md §4.1
pub mod offsets {
    pub const LONGITUDE: usize = 0x00;
    pub const LATITUDE: usize = 0x04;
    pub const COURSE: usize = 0x08;
    pub const SPEED: usize = 0x0c;
    pub const ALTITUDE: usize = 0x10;
    pub const TRACK_ID_HASH: usize = 0x14;
    pub const SOURCE_BITMAP: usize = 0x18;
    pub const CLASSIFICATION_LEVEL: usize = 0x1c;
    pub const THREAT_LEVEL: usize = 0x20;
    pub const ICON_INDEX: usize = 0x24;
    pub const ALERT_FLAGS: usize = 0x28;
    pub const UPDATE_EPOCH_MS: usize = 0x2c;
    pub const TRAIL_0: usize = 0x30;
    pub const TRAIL_1: usize = 0x40;
    pub const TRAIL_2: usize = 0x50;
    pub const TRAIL_3: usize = 0x60;
    pub const TRAIL_4: usize = 0x70;
}

// ── Inline helper writers ────────────────────────────────────────────────────

#[inline(always)]
fn write_f32(dest: &mut [u8], offset: usize, value: f32) {
    dest[offset..offset + 4].copy_from_slice(&value.to_le_bytes());
}

#[inline(always)]
fn write_u32(dest: &mut [u8], offset: usize, value: u32) {
    dest[offset..offset + 4].copy_from_slice(&value.to_le_bytes());
}

#[inline(always)]
fn read_f32(src: &[u8], offset: usize) -> f32 {
    f32::from_le_bytes(src[offset..offset + 4].try_into().unwrap_or([0u8; 4]))
}

#[inline(always)]
fn read_u32(src: &[u8], offset: usize) -> u32 {
    u32::from_le_bytes(src[offset..offset + 4].try_into().unwrap_or([0u8; 4]))
}

// ── Statistical Aggregation ──────────────────────────────────────────────────

/// Aggregated metrics for the Operations Commander dashboard.
#[wasm_bindgen]
pub struct FusionStats {
    pub high_confidence_count: u32,
    pub mid_confidence_count: u32,
    pub low_confidence_count: u32,
    pub radar_count: u32,
    pub sigint_count: u32,
    pub satellite_count: u32,
    pub ew_count: u32,
    pub others_count: u32,
    pub avg_latency_ms: f32,
    pub max_latency_ms: f32,
}

/// Compute statistical buckets for all tracks in the SharedArrayBuffer.
///
/// Iterates from slot 0 to `active_track_count` and aggregates confidence-score
/// distributions, sensor presence bits, and propagation latency.
///
/// # Returns
/// A `FusionStats` struct containing the computed metrics.
#[wasm_bindgen]
pub fn get_fusion_stats(
    sab_slice: &[u8],
    active_track_count: u32,
    current_time_ms: u32,
) -> FusionStats {
    let mut stats = FusionStats {
        high_confidence_count: 0,
        mid_confidence_count: 0,
        low_confidence_count: 0,
        radar_count: 0,
        sigint_count: 0,
        satellite_count: 0,
        ew_count: 0,
        others_count: 0,
        avg_latency_ms: 0.0,
        max_latency_ms: 0.0,
    };

    if active_track_count == 0 {
        return stats;
    }

    let mut total_latency = 0.0;
    let mut latency_count: u32 = 0;

    // Boundary check for the slice
    let max_len = sab_slice.len();

    for i in 0..active_track_count as usize {
        let offset = i * RECORD_SIZE;
        if offset + RECORD_SIZE > max_len {
            break;
        }

        let record = &sab_slice[offset..offset + RECORD_SIZE];

        // ── Data Extraction ──
        let source_bitmap = read_u32(record, offsets::SOURCE_BITMAP);
        let threat_level = read_u32(record, offsets::THREAT_LEVEL);
        let update_epoch_ms = read_u32(record, offsets::UPDATE_EPOCH_MS);

        // ── Confidence Buckets (Heuristic) ──
        // High: 3+ sensors OR Friendly(2)/Hostile(5)/Neutral(3)
        // Mid:  2 sensors OR Suspect(4)/Pending(1)
        // Low:  1 sensor  OR Unknown(0)
        let sensor_count = source_bitmap.count_ones();
        if sensor_count >= 3 || (threat_level >= 2 && threat_level <= 3) || threat_level == 5 {
            stats.high_confidence_count += 1;
        } else if sensor_count == 2 || threat_level == 4 || threat_level == 1 {
            stats.mid_confidence_count += 1;
        } else {
            stats.low_confidence_count += 1;
        }

        // ── Sensor Contribution (Bit assignments from mock-data.ts) ──
        // Radar: Bits 0, 1, 6 (Hormuz, Bahrain, Qeshm)
        if (source_bitmap & 0x43) != 0 {
            stats.radar_count += 1;
        }
        // SIGINT: Bit 4 (Chabahar ELINT)
        if (source_bitmap & 0x10) != 0 {
            stats.sigint_count += 1;
        }
        // Satellite: Bit 5 (Bandar Abbas ISR)
        if (source_bitmap & 0x20) != 0 {
            stats.satellite_count += 1;
        }
        // EW: Bit 3 (Muscat EW)
        if (source_bitmap & 0x08) != 0 {
            stats.ew_count += 1;
        }
        // Others: Bit 2 (Dubai AIS)
        if (source_bitmap & 0x04) != 0 {
            stats.others_count += 1;
        }

        // ── Latency (Time since server emission) ──
        if update_epoch_ms > 0 {
            let latency = if current_time_ms >= update_epoch_ms {
                current_time_ms - update_epoch_ms
            } else {
                0
            } as f32;

            if latency < 5000.0 {
                // Reject outliers > 5s
                total_latency += latency;
                latency_count += 1;
                if latency > stats.max_latency_ms {
                    stats.max_latency_ms = latency;
                }
            }
        }
    }

    if latency_count > 0 {
        stats.avg_latency_ms = total_latency / latency_count as f32;
    }

    stats
}

// ── Public API ───────────────────────────────────────────────────────────────

/// Decode a raw datagram and write the 128-byte record into `dest`.
///
/// For Phase 0, the datagram IS already a raw 128-byte binary record
/// (not a full FlatBuffer table with vtable overhead). The Rust Wasm
/// decoder validates length and copies fields at documented offsets.
///
/// Phase 2 will replace this with zero-copy FlatBuffer table access
/// once the Go serializer emits proper FlatBuffer datagrams.
///
/// # Returns
/// `true` if decoding succeeded; `false` if the datagram is malformed.
#[wasm_bindgen]
pub fn decode_track_update(datagram: &[u8], dest: &mut [u8]) -> bool {
    if datagram.len() < RECORD_SIZE || dest.len() < RECORD_SIZE {
        return false;
    }

    // Copy all 128 bytes directly — field layout is identical in both
    // the FlatBuffer binary record and the SAB slot.
    dest[..RECORD_SIZE].copy_from_slice(&datagram[..RECORD_SIZE]);
    true
}

/// Write a 128-byte record directly into a SharedArrayBuffer slice at
/// the given slot index.
///
/// # Arguments
/// * `sab_slice` — mutable view of the entire track-data region of the SAB
/// * `slot_index` — zero-based index of the target slot
/// * `max_slots` — total number of slots allocated (bounds check)
/// * `datagram`  — raw 128-byte FlatBuffer record
///
/// # Returns
/// `true` on success; `false` if slot_index >= max_slots or datagram too short.
#[wasm_bindgen]
pub fn write_slot_from_datagram(
    sab_slice: &mut [u8],
    slot_index: usize,
    max_slots: usize,
    datagram: &[u8],
) -> bool {
    if slot_index >= max_slots {
        return false;
    }
    if datagram.len() < RECORD_SIZE {
        return false;
    }

    let offset = slot_index * RECORD_SIZE;
    let end = offset + RECORD_SIZE;

    if end > sab_slice.len() {
        return false;
    }

    sab_slice[offset..end].copy_from_slice(&datagram[..RECORD_SIZE]);
    true
}

// ── Field accessors for unit tests ──────────────────────────────────────────

/// Extract the longitude field (f32, little-endian) from a raw record.
#[wasm_bindgen]
pub fn field_longitude(record: &[u8]) -> f32 {
    if record.len() < RECORD_SIZE {
        return 0.0;
    }
    read_f32(record, offsets::LONGITUDE)
}

/// Extract the latitude field (f32, little-endian) from a raw record.
#[wasm_bindgen]
pub fn field_latitude(record: &[u8]) -> f32 {
    if record.len() < RECORD_SIZE {
        return 0.0;
    }
    read_f32(record, offsets::LATITUDE)
}

/// Extract the track_id_hash field (u32, little-endian) from a raw record.
#[wasm_bindgen]
pub fn field_track_id_hash(record: &[u8]) -> u32 {
    if record.len() < RECORD_SIZE {
        return 0;
    }
    read_u32(record, offsets::TRACK_ID_HASH)
}

// ── Internal helpers (not exported to Wasm) ──────────────────────────────────

/// Build a synthetic 128-byte record for unit testing.
pub fn make_test_record(
    longitude: f32,
    latitude: f32,
    course: f32,
    speed: f32,
    altitude: f32,
    track_id_hash: u32,
    source_bitmap: u32,
    classification_level: u32,
    threat_level: u32,
    icon_index: u32,
    alert_flags: u32,
    update_epoch_ms: u32,
) -> [u8; RECORD_SIZE] {
    let mut record = [0u8; RECORD_SIZE];
    write_f32(&mut record, offsets::LONGITUDE, longitude);
    write_f32(&mut record, offsets::LATITUDE, latitude);
    write_f32(&mut record, offsets::COURSE, course);
    write_f32(&mut record, offsets::SPEED, speed);
    write_f32(&mut record, offsets::ALTITUDE, altitude);
    write_u32(&mut record, offsets::TRACK_ID_HASH, track_id_hash);
    write_u32(&mut record, offsets::SOURCE_BITMAP, source_bitmap);
    write_u32(&mut record, offsets::CLASSIFICATION_LEVEL, classification_level);
    write_u32(&mut record, offsets::THREAT_LEVEL, threat_level);
    write_u32(&mut record, offsets::ICON_INDEX, icon_index);
    write_u32(&mut record, offsets::ALERT_FLAGS, alert_flags);
    write_u32(&mut record, offsets::UPDATE_EPOCH_MS, update_epoch_ms);
    record
}
