// CLASSIFICATION: UNCLASSIFIED
// src/shaders/observations.wgsl — Raw sensor observation point render pass
//
// Renders small, semi-transparent circles at the locations of raw detections.
// Bypasses the complex culling/interpolation pipelines used for tracks.

struct Observation {
  lon:        f32,
  lat:        f32,
  sensor_type: f32,
  confidence:  f32,
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
  @builtin(position) position: vec4<f32>,
  @location(0)       color:    vec4<f32>,
  @location(1)       uv:       vec2<f32>,
}

@group(0) @binding(0) var<uniform> uniforms: Uniforms;
@group(1) @binding(0) var<storage, read> observations: array<Observation>;

const QUAD_UVS = array<vec2<f32>, 4>(
  vec2<f32>(0.0, 0.0),
  vec2<f32>(1.0, 0.0),
  vec2<f32>(0.0, 1.0),
  vec2<f32>(1.0, 1.0),
);

@vertex
fn vs_main(
  @builtin(vertex_index)   vid: u32,
  @builtin(instance_index) iid: u32,
) -> VertexOutput {
  let obs = observations[iid];

  // Project world position
  let clip = uniforms.view_proj * vec4<f32>(obs.lon, obs.lat, 0.0, 1.0);
  let w    = select(clip.w, 0.00001, abs(clip.w) < 0.00001);
  let ndc  = clip.xy / w;

  // Pixel-space billboard quad
  let uv     = QUAD_UVS[vid];
  let size   = 10.0; // 10px diameter
  let offset = (uv - vec2<f32>(0.5, 0.5)) * size / uniforms.viewport_size * 2.0;

  var out: VertexOutput;
  out.position = vec4<f32>(ndc + offset, clip.z / w, 1.0);
  out.uv       = uv;

  // Base color: Amber/Gold for raw observations
  // Alpha is tied to sensor confidence
  let alpha = max(0.2, obs.confidence);
  out.color = vec4<f32>(1.0, 0.7, 0.0, alpha);

  return out;
}

@fragment
fn fs_main(in: VertexOutput) -> @location(0) vec4<f32> {
  // Analytical circle with smooth edge
  let dist  = length(in.uv - vec2<f32>(0.5, 0.5));
  if (dist > 0.5) { discard; }

  let edge  = smoothstep(0.5, 0.4, dist);
  return vec4<f32>(in.color.rgb, in.color.a * edge);
}
