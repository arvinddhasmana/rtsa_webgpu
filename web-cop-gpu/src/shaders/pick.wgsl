// CLASSIFICATION: UNCLASSIFIED
// src/shaders/pick.wgsl — Pick buffer render pass (O(1) track selection)
//
// Renders each visible track icon into a separate R32Uint render target,
// writing the track_id_hash value instead of a colour. A click event reads
// back the single pixel under the cursor via mapAsync to identify the track.
//
// Reference: docs/sdlc_guidelines/08_tech_specific/webgpu_guidelines.md §7

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
  @builtin(position)              position:      vec4<f32>,
  @location(0) @interpolate(flat) track_id_hash: u32,
}

@group(0) @binding(0) var<uniform>       uniforms:        Uniforms;
@group(1) @binding(0) var<storage, read> tracks:          array<TrackRecord>;
@group(1) @binding(1) var<storage, read> positions:       array<vec4<f32>>;
@group(1) @binding(2) var<storage, read> visible_indices: array<u32>;

// Pick icon radius in pixels — slightly larger than visual icon for easy clicking
const PICK_RADIUS_PX: f32 = 20.0;

const QUAD_OFFSETS = array<vec2<f32>, 4>(
  vec2<f32>(-1.0, -1.0),
  vec2<f32>( 1.0, -1.0),
  vec2<f32>(-1.0,  1.0),
  vec2<f32>( 1.0,  1.0),
);

@vertex
fn vs_main(
  @builtin(vertex_index)   vid: u32,
  @builtin(instance_index) iid: u32,
) -> VertexOutput {
  let track_idx = visible_indices[iid];
  let track     = tracks[track_idx];
  let pos       = positions[track_idx];

  let clip   = uniforms.view_proj * vec4<f32>(pos.x, pos.y, 0.0, 1.0);
  let w      = select(clip.w, 0.00001, abs(clip.w) < 0.00001);
  let ndc    = clip.xy / w;
  let local  = QUAD_OFFSETS[vid];
  let offset = local * PICK_RADIUS_PX / (uniforms.viewport_size * 0.5);

  var out: VertexOutput;
  out.position      = vec4<f32>(ndc + offset, clip.z / w, 1.0);
  out.track_id_hash = track.track_id_hash;
  return out;
}

// Output to a R32Uint render target; fragment value = track_id_hash
@fragment
fn fs_main(in: VertexOutput) -> @location(0) u32 {
  return in.track_id_hash;
}
