// CLASSIFICATION: UNCLASSIFIED
// src/shaders/track-icons.wgsl — Instanced track icon quad render pass
//
// Renders one screen-aligned billboard quad per visible track, sampling
// the appropriate NATO APP-6 icon from the atlas texture.
//
// Reference: docs/sdlc_guidelines/08_tech_specific/wgsl_shader_standards.md §5.1

struct TrackRecord {
  lon:                  f32,
  lat:                  f32,
  course:               f32,
  speed:                f32,
  altitude:             f32,
  track_id_hash:        u32,
  source_bitmap:        u32,
  classification_level: u32,
  threat_level:         u32,
  icon_index:           u32,
  alert_flags:          u32,
  update_epoch_ms:      u32,
  trail:                array<vec4<f32>, 5>,
}

struct Uniforms {
  view_proj:              mat4x4<f32>,
  viewport_size:          vec2<f32>,
  current_time_ms:        u32,
  track_count:            u32,
  dashboard_mode:         u32,
  selected_track_id_hash: u32,
  map_style:              u32,
  _padding:               u32,
}

struct VertexOutput {
  @builtin(position)                    position:     vec4<f32>,
  @location(0)                          uv:           vec2<f32>,
  @location(1) @interpolate(flat)       icon_index:   u32,
  @location(2) @interpolate(flat)       threat_level: u32,
  @location(3) @interpolate(flat)       is_selected:  u32,
  @location(4) @interpolate(flat)       alert_flags:  u32,
}

@group(0) @binding(0) var<uniform>          uniforms:        Uniforms;
@group(1) @binding(0) var<storage, read>    tracks:          array<TrackRecord>;
@group(1) @binding(1) var<storage, read>    positions:       array<vec4<f32>>;
@group(1) @binding(2) var<storage, read>    visible_indices: array<u32>;

// Atlas layout: 2048×2048 px, each icon is 64×64 px → 32 columns × 32 rows
const ATLAS_COLS: u32  = 32u;
const ATLAS_ROWS: u32  = 32u;
const ICON_BASE_SIZE_PX: f32 = 24.0; // standard tiny sharp icons

// Unit quad UV coordinates for a triangle-strip (CCW winding)
// vertexIndex: 0=TL 1=TR 2=BL 3=BR
const QUAD_UVS = array<vec2<f32>, 4>(
  vec2<f32>(0.0, 0.0), // TL
  vec2<f32>(1.0, 0.0), // TR
  vec2<f32>(0.0, 1.0), // BL
  vec2<f32>(1.0, 1.0), // BR
);

@vertex
fn vs_main(
  @builtin(vertex_index)   vid: u32,
  @builtin(instance_index) iid: u32,
) -> VertexOutput {
  let track_idx = visible_indices[iid];
  let pos       = positions[track_idx];
  let track     = tracks[track_idx];

  let is_selected = u32(track.track_id_hash == uniforms.selected_track_id_hash);

  // Selection scaling: 1.5x larger when selected
  let scale = select(1.0, 1.5, is_selected > 0u);
  let size = ICON_BASE_SIZE_PX * scale;

  // Project world position to clip space
  let clip = uniforms.view_proj * vec4<f32>(pos.x, pos.y, 0.0, 1.0);
  let w    = select(clip.w, 0.00001, abs(clip.w) < 0.00001);
  let ndc  = clip.xy / w;

  // Billboard offset: pixel-space quad, converted to NDC delta
  let uv     = QUAD_UVS[vid];
  let offset = (uv - vec2<f32>(0.5, 0.5)) * size / uniforms.viewport_size * 2.0;

  var out: VertexOutput;
  out.position     = vec4<f32>(ndc + offset, clip.z / w, 1.0);
  out.uv           = uv;
  out.icon_index   = track.icon_index;
  out.threat_level = track.threat_level;
  out.is_selected  = is_selected;
  out.alert_flags  = track.alert_flags;
  return out;
}

fn threat_color(level: u32) -> vec3<f32> {
  switch (level) {
    case 0u: { return vec3<f32>(0.5, 0.5, 0.5); } // Unknown
    case 1u: { return vec3<f32>(0.1, 0.6, 1.0); } // Friend (Blue)
    case 2u: { return vec3<f32>(0.2, 0.8, 0.2); } // Green
    case 3u: { return vec3<f32>(0.8, 0.8, 0.2); } // Yellow/Neutral
    case 4u: { return vec3<f32>(1.0, 0.6, 0.0); } // Orange
    case 5u: { return vec3<f32>(1.0, 0.2, 0.2); } // Red
    default: { return vec3<f32>(1.0, 1.0, 1.0); }
  }
}

// Procedural Silhouette Drawing (SDF-like logic in pixels)
fn get_silhouette(uv: vec2<f32>, type_idx: u32) -> f32 {
  let p = uv * 2.0 - 1.0; // translate to [-1, 1]

  switch(type_idx % 3u) {
    case 0u: { // AIR: Delta wing triangle
      let dist = max(abs(p.x * 1.5) + p.y, -p.y * 2.0);
      return step(dist, 1.0);
    }
    case 1u: { // SURFACE: Long diamond/rectangle
      let dist = abs(p.x) * 2.5 + abs(p.y) * 1.2;
      return step(dist, 1.0);
    }
    case 2u: { // SUBSURFACE: Capsule/Oval
      let dist = (p.x * p.x * 2.5) + (p.y * p.y * 1.2);
      return step(dist, 1.0);
    }
    default: { return 0.0; }
  }
}

@fragment
fn fs_main(in: VertexOutput) -> @location(0) vec4<f32> {
  // Use procedural silhouette for generic icons
  let silhouette = get_silhouette(in.uv, in.icon_index);

  if (silhouette < 0.1) { discard; }

  var color = vec4<f32>(threat_color(in.threat_level), 1.0);

  // Pulse effect for anomalies (alert_flags > 0)
  if (in.alert_flags > 0u) {
    let t = f32(uniforms.current_time_ms % 1000u) / 1000.0;
    let pulse = 0.5 + 0.5 * sin(t * 6.28);
    let glow_color = select(vec3<f32>(1.0, 1.0, 0.0), vec3<f32>(1.0, 0.0, 0.0), in.threat_level == 5u);
    color = vec4<f32>(mix(color.rgb, glow_color, pulse * 0.5), 1.0);
  }

  // Highlight selected track
  if (in.is_selected > 0u) {
    color = vec4<f32>(color.rgb * 1.5, 1.0);
  }

  return color;
}
