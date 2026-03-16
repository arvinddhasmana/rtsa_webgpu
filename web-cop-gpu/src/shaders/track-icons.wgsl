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
  view_proj:       mat4x4<f32>,
  viewport_size:   vec2<f32>,
  current_time_ms: u32,
  track_count:     u32,
  dashboard_mode:  u32,
  _pad0:           u32,
  _pad1:           u32,
  _pad2:           u32,
}

struct VertexOutput {
  @builtin(position)                    position:     vec4<f32>,
  @location(0)                          uv:           vec2<f32>,
  @location(1) @interpolate(flat)       icon_index:   u32,
  @location(2) @interpolate(flat)       threat_level: u32,
}

@group(0) @binding(0) var<uniform>          uniforms:        Uniforms;
@group(1) @binding(0) var<storage, read>    tracks:          array<TrackRecord>;
@group(1) @binding(1) var<storage, read>    positions:       array<vec4<f32>>;
@group(1) @binding(2) var<storage, read>    visible_indices: array<u32>;
@group(2) @binding(0) var                   icon_atlas:      texture_2d<f32>;
@group(2) @binding(1) var                   icon_sampler:    sampler;

// Atlas layout: 2048×2048 px, each icon is 64×64 px → 32 columns × 32 rows
const ATLAS_COLS: u32  = 32u;
const ATLAS_ROWS: u32  = 32u;
const ICON_SIZE_PX: f32 = 32.0; // rendered size in pixels (screen space)

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

  // Project world position to clip space
  let clip = uniforms.view_proj * vec4<f32>(pos.x, pos.y, 0.0, 1.0);
  let w    = select(clip.w, 0.00001, abs(clip.w) < 0.00001);
  let ndc  = clip.xy / w;

  // Billboard offset: pixel-space quad, converted to NDC delta
  let uv     = QUAD_UVS[vid];
  let offset = (uv - vec2<f32>(0.5, 0.5)) * ICON_SIZE_PX / uniforms.viewport_size * 2.0;

  var out: VertexOutput;
  out.position     = vec4<f32>(ndc + offset, clip.z / w, 1.0);
  out.uv           = uv;
  out.icon_index   = track.icon_index;
  out.threat_level = track.threat_level;
  return out;
}

@fragment
fn fs_main(in: VertexOutput) -> @location(0) vec4<f32> {
  // Map icon_index to atlas row/column
  let col      = in.icon_index % ATLAS_COLS;
  let row      = in.icon_index / ATLAS_COLS;
  let atlas_uv = (vec2<f32>(f32(col), f32(row)) + in.uv)
                 / vec2<f32>(f32(ATLAS_COLS), f32(ATLAS_ROWS));

  let color = textureSample(icon_atlas, icon_sampler, atlas_uv);

  // Discard fully transparent pixels (alpha cutout)
  if (color.a < 0.1) { discard; }

  return color;
}
