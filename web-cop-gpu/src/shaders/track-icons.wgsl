// CLASSIFICATION: UNCLASSIFIED
// src/shaders/track-icons.wgsl — Instanced track icon quad render pass

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
  @location(5) @interpolate(flat)       course:       f32,
}

@group(0) @binding(0) var<uniform>          uniforms:        Uniforms;
@group(1) @binding(0) var<storage, read>    tracks:          array<TrackRecord>;
@group(1) @binding(1) var<storage, read>    positions:       array<vec4<f32>>;
@group(1) @binding(2) var<storage, read>    visible_indices: array<u32>;

const ICON_BASE_SIZE_PX: f32 = 14.0; // tiny and sharp

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
  let scale = select(1.0, 1.3, is_selected > 0u);
  let size = ICON_BASE_SIZE_PX * scale;

  let clip = uniforms.view_proj * vec4<f32>(pos.x, pos.y, 0.0, 1.0);
  let w    = select(clip.w, 0.00001, abs(clip.w) < 0.00001);
  let ndc  = clip.xy / w;

  let uv     = QUAD_UVS[vid];
  let offset = (uv - vec2<f32>(0.5, 0.5)) * size / uniforms.viewport_size * 2.0;

  var out: VertexOutput;
  out.position     = vec4<f32>(ndc + offset, clip.z / w, 1.0);
  out.uv           = uv;
  out.icon_index   = track.icon_index;
  out.threat_level = track.threat_level;
  out.is_selected  = is_selected;
  out.alert_flags  = track.alert_flags;
  out.course       = track.course;
  return out;
}

fn threat_color(level: u32) -> vec3<f32> {
  switch (level) {
    case 1u: { return vec3<f32>(0.2, 0.7, 1.0); } // Friendly (Blue)
    case 5u: { return vec3<f32>(1.0, 0.3, 0.3); } // Hostile (Red)
    default: { return vec3<f32>(0.9, 0.8, 0.2); } // Neutral/Unknown (Yellow)
  }
}

// Procedural high-fidelity silhouettes
fn get_silhouette(uv: vec2<f32>, type_idx: u32, rotated_uv: vec2<f32>) -> f32 {
  let p = rotated_uv * 2.0 - 1.0;
  let p_fixed = uv * 2.0 - 1.0;

  switch(type_idx % 3u) {
    case 0u: { // AIR: Swept-back silhouette
      let d = max(abs(p.x) * 1.5 + p.y * 0.5, -p.y);
      let tail = step(abs(p.x * 4.0) + (p.y + 0.8), 0.2);
      return max(step(d, 0.8), tail);
    }
    case 1u: { // SURFACE: Diamond hull
      let d = abs(p.x) * 2.0 + abs(p.y) * 0.8;
      return step(d, 1.0);
    }
    case 2u: { // SUBSURFACE: Capsule
      let d = (p_fixed.x * p_fixed.x * 3.0) + (p_fixed.y * p_fixed.y * 1.2);
      return step(d, 1.0);
    }
    default: { return 0.0; }
  }
}

@fragment
fn fs_main(in: VertexOutput) -> @location(0) vec4<f32> {
  // Rotation for orientation
  let angle = in.course * 3.14159 / 180.0;
  let c = cos(angle);
  let s = sin(angle);
  let pivot = vec2<f32>(0.5, 0.5);
  let rotated_uv = vec2<f32>(
    c * (in.uv.x - pivot.x) - s * (in.uv.y - pivot.y) + pivot.x,
    s * (in.uv.x - pivot.x) + c * (in.uv.y - pivot.y) + pivot.y
  );

  let silhouette = get_silhouette(in.uv, in.icon_index, rotated_uv);

  // Tactical Pulse / Anomaly Highlight
  var alpha = silhouette;
  var color = threat_color(in.threat_level);

  if (in.alert_flags > 0u) {
    let t = f32(uniforms.current_time_ms % 2000u) / 2000.0;
    let pulse = 0.5 + 0.5 * sin(t * 12.56); // faster double pulse

    // Outer glow for anomaly
    let dist_center = length(in.uv - vec2<f32>(0.5, 0.5));
    let glow = smoothstep(0.5, 0.45, dist_center) * pulse;

    let anomaly_color = select(vec3<f32>(1.0, 0.8, 0.0), vec3<f32>(1.0, 0.2, 0.0), in.threat_level == 5u);
    color = mix(color, anomaly_color, glow * 0.6);
    alpha = max(alpha, glow * 0.4);
  }

  if (alpha < 0.1) { discard; }

  // Selection Highlight
  if (in.is_selected > 0u) {
    color = mix(color, vec3<f32>(0.0, 1.0, 1.0), 0.3);
    // Subtle outer ring for selected
    let dist = length(in.uv - vec2<f32>(0.5, 0.5));
    if (dist > 0.45) { color = vec3<f32>(0.0, 1.0, 1.0); alpha = 1.0; }
  }

  return vec4<f32>(color, alpha);
}
