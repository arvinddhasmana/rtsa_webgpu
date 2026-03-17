// CLASSIFICATION: UNCLASSIFIED
// src/shaders/trail.wgsl — Track trail polyline render pass

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
  @location(0) @interpolate(flat)       threat_level: u32,
  @location(1)                          fade:         f32,
  @location(2) @interpolate(flat)       is_selected:  u32,
}

@group(0) @binding(0) var<uniform>       uniforms:        Uniforms;
@group(1) @binding(0) var<storage, read> tracks:          array<TrackRecord>;
@group(1) @binding(1) var<storage, read> visible_indices: array<u32>;

const TRAIL_WIDTH_BASE: f32 = 1.5;

const TRAIL_SEGMENTS:         u32 = 4u;
const VERTS_PER_SEGMENT:      u32 = 6u;

fn threat_color(level: u32) -> vec3<f32> {
  switch (level) {
    case 1u: { return vec3<f32>(0.2, 0.7, 1.0); }
    case 5u: { return vec3<f32>(1.0, 0.3, 0.3); }
    default: { return vec3<f32>(0.9, 0.8, 0.2); }
  }
}

@vertex
fn vs_main(
  @builtin(vertex_index)   vid: u32,
  @builtin(instance_index) iid: u32,
) -> VertexOutput {
  let track_idx = visible_indices[iid];
  let track     = tracks[track_idx];

  let is_selected = u32(track.track_id_hash == uniforms.selected_track_id_hash);

  // IN COMMANDER MODE, ONLY SHOW SELECTED OR HIGH-THREAT TRAILS
  let show_trail = select(true, is_selected > 0u || track.threat_level == 5u, uniforms.dashboard_mode == 1u);

  if (!show_trail) {
    return VertexOutput(vec4<f32>(0.0, 0.0, 0.0, 1.0), 0u, 0.0, 0u);
  }

  let seg      = vid / VERTS_PER_SEGMENT;
  let vert_in  = vid % VERTS_PER_SEGMENT;

  let trail_seg = track.trail[seg];
  let p_a = vec2<f32>(trail_seg.x, trail_seg.y);
  let p_b = vec2<f32>(trail_seg.z, trail_seg.w);

  let clip_a = uniforms.view_proj * vec4<f32>(p_a.x, p_a.y, 0.0, 1.0);
  let clip_b = uniforms.view_proj * vec4<f32>(p_b.x, p_b.y, 0.0, 1.0);
  let w_a = select(clip_a.w, 0.00001, abs(clip_a.w) < 0.00001);
  let w_b = select(clip_b.w, 0.00001, abs(clip_b.w) < 0.00001);
  let ndc_a = clip_a.xy / w_a;
  let ndc_b = clip_b.xy / w_b;

  let screen_a = ndc_a * uniforms.viewport_size * 0.5;
  let screen_b = ndc_b * uniforms.viewport_size * 0.5;
  let dir      = normalize(screen_b - screen_a);

  // Selection glow width
  var width = TRAIL_WIDTH_BASE;
  if (is_selected > 0u) {
      let t = f32(uniforms.current_time_ms % 1000u) / 1000.0;
      width = TRAIL_WIDTH_BASE * (2.0 + 0.5 * sin(t * 6.28 - f32(seg)));
  }

  let normal   = vec2<f32>(-dir.y, dir.x) * width * 0.5;
  let ndc_norm = normal / (uniforms.viewport_size * 0.5);

  var ndc_pos: vec2<f32>;
  var fade_out: f32;
  switch (vert_in) {
    case 0u: { ndc_pos = ndc_a + ndc_norm; fade_out = 1.0; }
    case 1u: { ndc_pos = ndc_a - ndc_norm; fade_out = 1.0; }
    case 2u: { ndc_pos = ndc_b + ndc_norm; fade_out = 0.0; }
    case 3u: { ndc_pos = ndc_b + ndc_norm; fade_out = 0.0; }
    case 4u: { ndc_pos = ndc_a - ndc_norm; fade_out = 1.0; }
    case 5u: { ndc_pos = ndc_b - ndc_norm; fade_out = 0.0; }
    default: { ndc_pos = ndc_a;            fade_out = 1.0; }
  }

  var out: VertexOutput;
  out.position     = vec4<f32>(ndc_pos, clip_a.z / w_a, 1.0);
  out.threat_level = track.threat_level;
  out.fade         = fade_out * (1.0 - f32(seg) * 0.2);
  out.is_selected  = is_selected;
  return out;
}

@fragment
fn fs_main(in: VertexOutput) -> @location(0) vec4<f32> {
  var col = threat_color(in.threat_level);
  var alpha = in.fade * 0.5;

  if (in.is_selected > 0u) {
      col = mix(col, vec3<f32>(0.0, 1.0, 1.0), 0.5);
      alpha = in.fade * 0.8;
  }

  return vec4<f32>(col, alpha);
}
