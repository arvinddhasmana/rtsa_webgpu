// CLASSIFICATION: UNCLASSIFIED
// src/shaders/trail.wgsl — Track trail polyline render pass
//
// Renders the 5-point trail ring buffer for each visible track as line
// quads (emulated thick lines). Color is derived from the track's threat level.
//
// Reference: docs/implementation/v4/phase1_core_rendering.md R1-6

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
  view_proj:       mat4x4<f32>,
  viewport_size:   vec2<f32>,
  current_time_ms: u32,
  track_count:     u32,
}

struct VertexOutput {
  @builtin(position)              position:     vec4<f32>,
  @location(0) @interpolate(flat) threat_level: u32,
  @location(1)                    fade:         f32,
}

@group(0) @binding(0) var<uniform>       uniforms:        Uniforms;
@group(1) @binding(0) var<storage, read> tracks:          array<TrackRecord>;
@group(1) @binding(1) var<storage, read> visible_indices: array<u32>;

// Trail width in pixels (screen space)
const TRAIL_WIDTH_PX: f32 = 2.0;

// Number of trail segments = 4 (5 points → 4 segments)
// Number of vertices per segment = 6 (two triangles forming a quad)
const TRAIL_SEGMENTS:         u32 = 4u;
const VERTS_PER_SEGMENT:      u32 = 6u;
const VERTS_PER_INSTANCE:     u32 = 24u; // 4 segments × 6 vertices

// Threat-level colour palette (STANAG colour codes)
fn threat_color(level: u32) -> vec3<f32> {
  switch (level) {
    case 0u: { return vec3<f32>(0.5,  0.5,  0.5);  } // Unknown  — grey
    case 1u: { return vec3<f32>(0.1,  0.6,  1.0);  } // Friendly — blue
    case 2u: { return vec3<f32>(0.2,  0.8,  0.2);  } // Neutral  — green
    case 3u: { return vec3<f32>(1.0,  0.85, 0.0);  } // Suspect  — amber
    case 4u: { return vec3<f32>(1.0,  0.5,  0.0);  } // Hostile  — orange
    case 5u: { return vec3<f32>(1.0,  0.2,  0.2);  } // Joker    — red
    default: { return vec3<f32>(0.5,  0.5,  0.5);  }
  }
}

@vertex
fn vs_main(
  @builtin(vertex_index)   vid: u32,
  @builtin(instance_index) iid: u32,
) -> VertexOutput {
  let track_idx = visible_indices[iid];
  let track     = tracks[track_idx];

  // Decode segment and vertex within segment
  let seg      = vid / VERTS_PER_SEGMENT; // 0..3
  let vert_in  = vid % VERTS_PER_SEGMENT; // 0..5

  // Trail point A → B for this segment
  // trail[i] = vec4(lon_a, lat_a, lon_b, lat_b)
  let trail_seg = track.trail[seg];
  let p_a = vec2<f32>(trail_seg.x, trail_seg.y);
  let p_b = vec2<f32>(trail_seg.z, trail_seg.w);

  // Project both endpoints to NDC
  let clip_a = uniforms.view_proj * vec4<f32>(p_a.x, p_a.y, 0.0, 1.0);
  let clip_b = uniforms.view_proj * vec4<f32>(p_b.x, p_b.y, 0.0, 1.0);
  let w_a = select(clip_a.w, 0.00001, abs(clip_a.w) < 0.00001);
  let w_b = select(clip_b.w, 0.00001, abs(clip_b.w) < 0.00001);
  let ndc_a = clip_a.xy / w_a;
  let ndc_b = clip_b.xy / w_b;

  // Screen-space direction and normal for line quad
  let screen_a = ndc_a * uniforms.viewport_size * 0.5;
  let screen_b = ndc_b * uniforms.viewport_size * 0.5;
  let dir      = normalize(screen_b - screen_a);
  let normal   = vec2<f32>(-dir.y, dir.x) * TRAIL_WIDTH_PX * 0.5;
  let ndc_norm = normal / (uniforms.viewport_size * 0.5);

  // Quad vertex selection: two triangles forming a rectangle
  // verts 0,1,2 = first triangle; 3,4,5 = second triangle
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
  out.position     = vec4<f32>(ndc_pos, 0.0, 1.0);
  out.threat_level = track.threat_level;
  // Fade trail older segments: seg 0 is most recent (1.0), seg 3 is oldest (0.25)
  out.fade         = fade_out * (1.0 - f32(seg) * 0.2);
  return out;
}

@fragment
fn fs_main(in: VertexOutput) -> @location(0) vec4<f32> {
  let col = threat_color(in.threat_level);
  return vec4<f32>(col, in.fade * 0.7);
}
