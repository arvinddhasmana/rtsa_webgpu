// CLASSIFICATION: UNCLASSIFIED
// wasm-decoder/tests/decode_test.rs — Unit tests for the FlatBuffer decoder
//
// Covers:
//   - Round-trip: field values survive encode → decode
//   - Bounds check: returns false for out-of-bounds slot index
//   - Short datagram: returns false when datagram.len() < RECORD_SIZE
//   - Zero record: all-zero datagram decodes to all-zero dest

use wasm_decoder::{
    decode_track_update, field_latitude, field_longitude, field_track_id_hash,
    make_test_record, write_slot_from_datagram, RECORD_SIZE,
};

// ── decode_track_update ──────────────────────────────────────────────────────

#[test]
fn test_decode_round_trip_field_values() {
    let record = make_test_record(
        1.234_f32,  // longitude
        0.567_f32,  // latitude
        0.1_f32,    // course
        15.0_f32,   // speed
        1500.0_f32, // altitude
        0xDEAD_BEEF, // track_id_hash
        0b0000_0001, // source_bitmap
        1,           // classification_level
        0,           // threat_level (Unknown)
        3,           // icon_index
        0,           // alert_flags
        1_700_000_000, // update_epoch_ms
    );

    let mut dest = [0u8; RECORD_SIZE];
    let ok = decode_track_update(&record, &mut dest);

    assert!(ok, "decode_track_update should return true for a valid record");
    assert_eq!(
        field_longitude(&dest),
        1.234_f32,
        "longitude must survive round-trip"
    );
    assert_eq!(
        field_latitude(&dest),
        0.567_f32,
        "latitude must survive round-trip"
    );
    assert_eq!(
        field_track_id_hash(&dest),
        0xDEAD_BEEF,
        "track_id_hash must survive round-trip"
    );
}

#[test]
fn test_decode_short_datagram_returns_false() {
    let short = [0u8; 64]; // only 64 bytes — less than RECORD_SIZE
    let mut dest = [0u8; RECORD_SIZE];
    let ok = decode_track_update(&short, &mut dest);
    assert!(!ok, "decode_track_update must return false for short datagram");
}

#[test]
fn test_decode_short_dest_returns_false() {
    let record = [0u8; RECORD_SIZE];
    let mut short_dest = [0u8; 64];
    let ok = decode_track_update(&record, &mut short_dest);
    assert!(!ok, "decode_track_update must return false when dest too small");
}

#[test]
fn test_decode_zero_record() {
    let record = [0u8; RECORD_SIZE];
    let mut dest = [0xff_u8; RECORD_SIZE]; // pre-fill with 0xff
    let ok = decode_track_update(&record, &mut dest);
    assert!(ok);
    assert!(
        dest.iter().all(|&b| b == 0),
        "zero datagram should produce zero dest"
    );
}

// ── write_slot_from_datagram ─────────────────────────────────────────────────

#[test]
fn test_write_slot_bounds_check_rejects_out_of_range() {
    let datagram = make_test_record(0.0, 0.0, 0.0, 0.0, 0.0, 0, 0, 0, 0, 0, 0, 0);
    let max_slots: usize = 4;
    let mut sab = vec![0u8; max_slots * RECORD_SIZE];

    // slot_index == max_slots should be rejected
    let ok = write_slot_from_datagram(&mut sab, max_slots, max_slots, &datagram);
    assert!(!ok, "slot_index == max_slots must be rejected");

    // slot_index > max_slots should also be rejected
    let ok2 = write_slot_from_datagram(&mut sab, max_slots + 10, max_slots, &datagram);
    assert!(!ok2, "slot_index > max_slots must be rejected");
}

#[test]
fn test_write_slot_valid_writes_to_correct_offset() {
    let datagram = make_test_record(
        2.718_f32, // longitude
        1.414_f32, // latitude
        0.0, 0.0, 0.0, 0xCAFE_BABE, 0, 0, 0, 0, 0, 0,
    );
    let max_slots: usize = 8;
    let mut sab = vec![0u8; max_slots * RECORD_SIZE];

    let ok = write_slot_from_datagram(&mut sab, 3, max_slots, &datagram);
    assert!(ok, "write_slot_from_datagram must return true for valid slot");

    // Verify the data landed at slot 3 (byte offset 3 * 128 = 384)
    let slot_start = 3 * RECORD_SIZE;
    let lon = f32::from_le_bytes(sab[slot_start..slot_start + 4].try_into().unwrap());
    let lat = f32::from_le_bytes(sab[slot_start + 4..slot_start + 8].try_into().unwrap());
    assert_eq!(lon, 2.718_f32, "longitude at slot 3 must match");
    assert_eq!(lat, 1.414_f32, "latitude at slot 3 must match");
}

#[test]
fn test_write_slot_short_datagram_returns_false() {
    let max_slots: usize = 4;
    let mut sab = vec![0u8; max_slots * RECORD_SIZE];
    let short = [0u8; 64];

    let ok = write_slot_from_datagram(&mut sab, 0, max_slots, &short);
    assert!(!ok, "short datagram must be rejected");
}
