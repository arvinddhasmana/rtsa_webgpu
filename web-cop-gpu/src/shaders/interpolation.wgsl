// CLASSIFICATION: UNCLASSIFIED
// src/shaders/interpolation.wgsl — Dead-reckoning track position extrapolation
//
// Reads track records from the storage buffer and extrapolates current positions
// using course, speed, and the elapsed time since the last server update.
//
// Reference: docs/sdlc_guidelines/08_tech_specific/wgsl_shader_standards.md §4.2

struct TrackRecord {
  lon:                  f32,   // offset 0x00
  lat:                  f32,   // offset 0x04
  course:               f32,   // offset 0x08 (radians, 0 = north, clockwise)
  speed:                f32,   // offset 0x0C (m/s)
  altitude:             f32,   // offset 0x10 (meters)
  track_id_hash:        u32,   // offset 0x14
  source_bitmap:        u32,   // offset 0x18
  classification_level: u32,   // offset 0x1C
  threat_level:         u32,   // offset 0x20 (0–5 enum)
  icon_index:           u32,   // offset 0x24 (atlas row)
  alert_flags:          u32,   // offset 0x28 (bitmask)
  update_epoch_ms:      u32,   // offset 0x2C
  trail:                array<vec4<f32>, 5>, // offset 0x30 (5 × 16 bytes)
}

struct Uniforms {
  view_proj:       mat4x4<f32>, // offset 0,  size 64
  viewport_size:   vec2<f32>,   // offset 64, size 8
  current_time_ms: u32,         // offset 72, size 4
  track_count:     u32,         // offset 76, size 4
  // Total: 80 bytes
}

@group(0) @binding(0) var<uniform>          uniforms:  Uniforms;
@group(1) @binding(0) var<storage, read>    tracks:    array<TrackRecord>;
@group(1) @binding(1) var<storage, read_write> positions: array<vec4<f32>>;

// Earth radius in meters (WGS-84 mean)
const EARTH_RADIUS_M: f32 = 6371000.0;

// Maximum dead-reckoning window — discard extrapolation beyond 5 minutes
const MAX_DR_S: f32 = 300.0;

@compute @workgroup_size(256)
fn main(@builtin(global_invocation_id) gid: vec3<u32>) {
  let idx = gid.x;
  if (idx >= uniforms.track_count) { return; }

  let track = tracks[idx];

  // Clamp elapsed time to prevent runaway extrapolation for stale tracks
  let raw_dt_ms = i32(uniforms.current_time_ms) - i32(track.update_epoch_ms);
  let clamped_dt_ms = clamp(f32(raw_dt_ms), 0.0, MAX_DR_S * 1000.0);
  let dt_s = clamped_dt_ms / 1000.0;

  // Dead-reckoning: angular displacement on a sphere
  // dx (east component) = speed * sin(course) * dt / R / cos(lat)
  // dy (north component) = speed * cos(course) * dt / R
  let cos_lat = cos(track.lat);
  let safe_cos_lat = select(cos_lat, 0.0001, abs(cos_lat) < 0.0001); // avoid div/0 at poles
  let dx = track.speed * sin(track.course) * dt_s / (EARTH_RADIUS_M * safe_cos_lat);
  let dy = track.speed * cos(track.course) * dt_s / EARTH_RADIUS_M;

  // Store extrapolated position; w component carries track_id_hash for identification
  positions[idx] = vec4<f32>(
    track.lon + dx,
    track.lat + dy,
    track.altitude,
    bitcast<f32>(track.track_id_hash),
  );
}
