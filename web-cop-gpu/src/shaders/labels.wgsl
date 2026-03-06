// CLASSIFICATION: UNCLASSIFIED
// src/shaders/labels.wgsl — SDF text label render pass
//
// Renders callsign labels as Signed Distance Field glyphs sampled from the
// pre-baked SDF font atlas. Labels are offset from the track icon position.
//
// Reference: docs/sdlc_guidelines/08_tech_specific/wgsl_shader_standards.md §5.2

struct GlyphInstance {
  // Screen position of glyph quad top-left (NDC)
  ndc_x:     f32,
  ndc_y:     f32,
  // Glyph size in NDC
  width:     f32,
  height:    f32,
  // Atlas UV rect (normalised 0..1 into the 2048×1024 R8 atlas)
  uv_x:      f32,
  uv_y:      f32,
  uv_w:      f32,
  uv_h:      f32,
  // Colour packed as RGBA8 (unused alpha = 255)
  color_rgb: u32,
  // Padding to 40 bytes
  _pad0:     u32,
}

struct Uniforms {
  view_proj:       mat4x4<f32>,
  viewport_size:   vec2<f32>,
  current_time_ms: u32,
  track_count:     u32,
}

struct VertexOutput {
  @builtin(position) position:   vec4<f32>,
  @location(0)       atlas_uv:   vec2<f32>,
  @location(1)       text_color: vec3<f32>,
}

@group(0) @binding(0) var<uniform>       uniforms:     Uniforms;
@group(1) @binding(0) var<storage, read> glyphs:       array<GlyphInstance>;
@group(2) @binding(0) var                sdf_atlas:    texture_2d<f32>;
@group(2) @binding(1) var                sdf_sampler:  sampler;

// Unit quad UVs for a triangle-strip (TL, TR, BL, BR)
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
  let g   = glyphs[iid];
  let uv  = QUAD_UVS[vid];

  // Glyph quad position in NDC
  let ndc_pos = vec2<f32>(g.ndc_x + uv.x * g.width,
                          g.ndc_y - uv.y * g.height); // y-flip for atlas coords

  // Decode colour from packed u32 (RGB, big-endian bytes)
  let r = f32((g.color_rgb >> 16u) & 0xFFu) / 255.0;
  let gv = f32((g.color_rgb >>  8u) & 0xFFu) / 255.0;
  let b = f32( g.color_rgb         & 0xFFu) / 255.0;

  // Atlas UV for this glyph
  let atlas_uv = vec2<f32>(g.uv_x + uv.x * g.uv_w,
                           g.uv_y + uv.y * g.uv_h);

  var out: VertexOutput;
  out.position   = vec4<f32>(ndc_pos, 0.0, 1.0);
  out.atlas_uv   = atlas_uv;
  out.text_color = vec3<f32>(r, gv, b);
  return out;
}

@fragment
fn fs_main(in: VertexOutput) -> @location(0) vec4<f32> {
  // Sample the SDF distance value from the R channel
  let dist      = textureSample(sdf_atlas, sdf_sampler, in.atlas_uv).r;
  let edge      = 0.5;
  let smoothing = fwidth(dist) * 0.5;
  let alpha     = smoothstep(edge - smoothing, edge + smoothing, dist);

  if (alpha < 0.01) { discard; }
  return vec4<f32>(in.text_color, alpha);
}
