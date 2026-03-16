// CLASSIFICATION: UNCLASSIFIED
// src/shaders/culling.wgsl — View-frustum culling for visible track list
//
// Tests each interpolated track position against the view frustum and writes
// visible track indices to the instance list. Atomically increments the
// indirect draw argument so the icon/trail/halo passes use the exact count.
//
// Reference: docs/sdlc_guidelines/08_tech_specific/wgsl_shader_standards.md §4.3

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

// Indirect draw arguments for a non-indexed instanced draw call:
//   vertex_count    (4 for a quad strip)
//   instance_count  (written by this shader)
//   first_vertex    (0)
//   first_instance  (0)
struct DrawIndirectArgs {
  vertex_count:    u32,
  instance_count:  atomic<u32>,
  first_vertex:    u32,
  first_instance:  u32,
}

@group(0) @binding(0) var<uniform>             uniforms:       Uniforms;
@group(1) @binding(0) var<storage, read>       positions:      array<vec4<f32>>;
@group(1) @binding(1) var<storage, read_write> visible_indices: array<u32>;
@group(1) @binding(2) var<storage, read_write> draw_args:      DrawIndirectArgs;

// NDC frustum margin — keeps tracks visible when their icon partially overlaps the edge
const FRUSTUM_MARGIN: f32 = 1.2;

@compute @workgroup_size(256)
fn main(@builtin(global_invocation_id) gid: vec3<u32>) {
  let idx = gid.x;
  if (idx >= uniforms.track_count) { return; }

  let pos = positions[idx];

  // Transform world position to clip space
  let clip = uniforms.view_proj * vec4<f32>(pos.x, pos.y, 0.0, 1.0);

  // Perspective divide to NDC
  let w = select(clip.w, 0.00001, abs(clip.w) < 0.00001); // guard against w = 0
  let ndc = clip.xy / w;

  // Frustum test: track is visible if NDC is within the margin boundary
  if (abs(ndc.x) < FRUSTUM_MARGIN && abs(ndc.y) < FRUSTUM_MARGIN) {
    let slot = atomicAdd(&draw_args.instance_count, 1u);
    visible_indices[slot] = idx;
  }
}
