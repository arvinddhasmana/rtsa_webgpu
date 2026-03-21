// CLASSIFICATION: UNCLASSIFIED
// src/shaders/interpolation.wgsl — Dead-reckoning track position extrapolation
//
// Reads track records from the storage buffer and extrapolates current positions
// using course, speed, and the elapsed time since the last server update.
//
// Reference: docs/sdlc_guidelines/08_tech_specific/wgsl_shader_standards.md §4.2

// TrackRecord layout (128 bytes, verified 2026-03-06):
// Offset 0x00: lon, lat, course, speed, altitude (5×f32 = 20 bytes)
// Offset 0x14: track_id_hash, source_bitmap, classification_level, threat_level,
//              icon_index, alert_flags, update_epoch_ms (7×u32 = 28 bytes)
// Offset 0x30: trail (array<vec4<f32>, 5> = 80 bytes, 16-byte aligned ✓)
// Total: 128 bytes
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
  view_proj:       mat4x4<f32>,
  viewport_size:   vec2<f32>,
  current_time_ms: u32,
  track_count:     u32,
  dashboard_mode:  u32,
  _pad0:           u32,
  _pad1:           u32,
  _pad2:           u32,
}

@group(0) @binding(0) var<uniform>          uniforms:  Uniforms;
@group(1) @binding(0) var<storage, read>    tracks:    array<TrackRecord>;
@group(1) @binding(1) var<storage, read_write> positions: array<vec4<f32>>;

// Earth radius in meters (WGS-84 mean)
const EARTH_RADIUS_M: f32 = 6371000.0;

// Maximum dead-reckoning window — discard extrapolation beyond 5 minutes
const MAX_DR_S: f32 = 300.0;

const PI: f32 = 3.14159265359;
// Track records store lon/lat in degrees; shaders need radians.
const DEG_TO_RAD: f32 = PI / 180.0;

// Expects lon_rad and lat_rad in RADIANS and returns normalised Web Mercator [0,1].
fn to_web_mercator(lon_rad: f32, lat_rad: f32) -> vec2<f32> {
    let mx = lon_rad / (2.0 * PI) + 0.5;
    let my = 0.5 - log(tan(PI / 4.0 + lat_rad / 2.0)) / (2.0 * PI);
    return vec2<f32>(mx, my);
}

@compute @workgroup_size(256)
fn main(@builtin(global_invocation_id) gid: vec3<u32>) {
  let idx = gid.x;
  if (idx >= uniforms.track_count) { return; }

  let track = tracks[idx];

  // Clamp elapsed time to prevent runaway extrapolation for stale tracks
  let raw_dt_ms = i32(uniforms.current_time_ms) - i32(track.update_epoch_ms);
  let clamped_dt_ms = clamp(f32(raw_dt_ms), 0.0, MAX_DR_S * 1000.0);
  let dt_s = clamped_dt_ms / 1000.0;

  // SAB stores lon/lat in degrees; convert to radians for all trigonometry.
  let lon_rad = track.lon * DEG_TO_RAD;
  let lat_rad = track.lat * DEG_TO_RAD;

  // Dead-reckoning: angular displacement on a sphere (result in radians)
  // dx (east component) = speed * sin(course) * dt / R / cos(lat)
  // dy (north component) = speed * cos(course) * dt / R
  let cos_lat = cos(lat_rad);
  let safe_cos_lat = select(cos_lat, 0.0001, abs(cos_lat) < 0.0001); // avoid div/0 at poles
  let dx = track.speed * sin(track.course) * dt_s / (EARTH_RADIUS_M * safe_cos_lat);
  let dy = track.speed * cos(track.course) * dt_s / EARTH_RADIUS_M;

  // Add dead-reckoning displacement (in radians) to radian coordinates
  let final_lon = lon_rad + dx;
  let final_lat = lat_rad + dy;
  let m = to_web_mercator(final_lon, final_lat);

  // Store extrapolated position; w component carries track_id_hash for identification
  positions[idx] = vec4<f32>(
    m.x,
    m.y,
    track.altitude,
    bitcast<f32>(track.track_id_hash),
  );
}
