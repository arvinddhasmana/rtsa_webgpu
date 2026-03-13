// CLASSIFICATION: UNCLASSIFIED
// src/shaders/coverage.wgsl — Sensor Coverage Rendering Pass
//
// Renders sensor footprints (arcs/sectors) and coverage gaps (polygons).
// Uses instanced rendering where each instance is a sector or a gap.

struct CoverageRecord {
  center_lon:     f32,
  center_lat:     f32,
  range_nm:       f32,
  bearing_start:  f32,
  bearing_end:    f32,
  record_type:    u32, // 0 = Sector, 1 = Gap Polygon
  alert_level:    u32, // 0 = Normal, 1 = Warning, 2 = Critical
  padding:        u32,
}

struct Uniforms {
  view_proj:       mat4x4<f32>,
  viewport_size:   vec2<f32>,
  current_time_ms: u32,
  record_count:    u32,
  dashboard_mode:  u32, // 0=sensor, 1=commander, 2=analytics, 3=health
}

struct VertexOutput {
  @builtin(position) position: vec4<f32>,
  @location(0)       uv:       vec2<f32>,
  @location(1) @interpolate(flat) alert_level: u32,
  @location(2) @interpolate(flat) record_type: u32,
}

@group(0) @binding(0) var<uniform>       uniforms: Uniforms;
@group(1) @binding(0) var<storage, read> records:  array<CoverageRecord>;

const VERTS_PER_INSTANCE: u32 = 6u; // Quad for sectors/gaps

// Earth radius in Nautical Miles for approximate distance calculation
const EARTH_RADIUS_NM: f32 = 3440.0;

@vertex
fn vs_main(
  @builtin(vertex_index)   vid: u32,
  @builtin(instance_index) iid: u32,
) -> VertexOutput {
  let record = records[iid];

  // Quadrant vertex selection (0,0) to (1,1)
  var uv = vec2<f32>(0.0);
  switch (vid) {
    case 0u: { uv = vec2<f32>(-1.0, -1.0); }
    case 1u: { uv = vec2<f32>( 1.0, -1.0); }
    case 2u: { uv = vec2<f32>(-1.0,  1.0); }
    case 3u: { uv = vec2<f32>(-1.0,  1.0); }
    case 4u: { uv = vec2<f32>( 1.0, -1.0); }
    case 5u: { uv = vec2<f32>( 1.0,  1.0); }
    default: { uv = vec2<f32>( 0.0,  0.0); }
  }

  // Calculate bounding box in Lon/Lat
  // 1 degree lat approx 60 NM
  let delta_lat = record.range_nm / 60.0;
  let delta_lon = record.range_nm / (60.0 * cos(record.center_lat * 3.14159 / 180.0));

  let lon = record.center_lon + uv.x * delta_lon;
  let lat = record.center_lat + uv.y * delta_lat;

  var out: VertexOutput;
  out.position = uniforms.view_proj * vec4<f32>(lon, lat, 0.0, 1.0);
  out.uv = uv;
  out.alert_level = record.alert_level;
  out.record_type = record.record_type;
  return out;
}

@fragment
fn fs_main(in: VertexOutput) -> @location(0) vec4<f32> {
  let dist = length(in.uv);

  // Circle mask for sector
  if (dist > 1.0) {
    discard;
  }

  // Sector Angle Mask (Simplified)
  // In a real implementation, we'd calculate bearing from UV
  // For the mockup pulse effect:
  let pulse = 0.5 + 0.5 * sin(f32(uniforms.current_time_ms) * 0.005);

  var color: vec3<f32>;
  if (in.record_type == 1u) {
    // Gap - Neon Red Pulse
    color = vec3<f32>(1.0, 0.1, 0.1) * pulse;
  } else {
    // Footprint - Neon Blue / Cyan
    color = vec3<f32>(0.0, 0.8, 1.0);
  }

  // Edge glow
  let edge = smoothstep(0.9, 1.0, dist);
  color += vec3<f32>(1.0) * edge * 0.5;

  var alpha = (1.0 - dist) * 0.4;
  if (uniforms.dashboard_mode == 0u) {
    alpha = (1.0 - dist) * 0.7; // Boosted for Sensor Dashboard
  }
  return vec4<f32>(color, alpha);
}
