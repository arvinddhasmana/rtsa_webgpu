// CLASSIFICATION: UNCLASSIFIED
// src/shaders/background.wgsl — Tactical map background renderer

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

@group(0) @binding(0) var<uniform> uniforms: Uniforms;

struct VertexOutput {
  @builtin(position) position: vec4<f32>,
  @location(0) uv: vec2<f32>,
}

@vertex
fn vs_main(@builtin(vertex_index) vid: u32) -> VertexOutput {
  var pos = array<vec2<f32>, 4>(
    vec2<f32>(-1.0, -1.0),
    vec2<f32>( 1.0, -1.0),
    vec2<f32>(-1.0,  1.0),
    vec2<f32>( 1.0,  1.0)
  );

  var out: VertexOutput;
  out.position = vec4<f32>(pos[vid], 0.0, 1.0);
  out.uv = pos[vid] * 0.5 + 0.5;
  return out;
}

fn hash(p: vec2<f32>) -> f32 {
    return fract(sin(dot(p, vec2<f32>(127.1, 311.7))) * 43758.5453123);
}

fn noise(p: vec2<f32>) -> f32 {
    let i = floor(p);
    let f = fract(p);
    let u = f * f * (3.0 - 2.0 * f);
    return mix(mix(hash(i + vec2<f32>(0.0, 0.0)),
                   hash(i + vec2<f32>(1.0, 0.0)), u.x),
               mix(hash(i + vec2<f32>(0.0, 1.0)),
                   hash(i + vec2<f32>(1.0, 1.0)), u.x), u.y);
}

@fragment
fn fs_main(in: VertexOutput) -> @location(0) vec4<f32> {
    let style = uniforms.map_style;

    // Brightened colors for verification
    let standard_base = vec3<f32>(0.05, 0.07, 0.12);
    let hd_base = vec3<f32>(0.08, 0.1, 0.15);

    var color = select(standard_base, hd_base, style == 1u);

    // Strong Grid
    let grid_size = 100.0;
    let grid_uv = in.position.xy / grid_size;
    let grid = smoothstep(0.95, 1.0, fract(grid_uv.x)) + smoothstep(0.95, 1.0, fract(grid_uv.y));
    color += grid * 0.1;

    if (style == 1u) {
        let uv_scale = in.uv * 5.0;
        let n = noise(uv_scale) * 0.5 + noise(uv_scale * 2.0) * 0.25;
        let terrain = smoothstep(0.3, 0.7, n);

        let ocean = vec3<f32>(0.05, 0.08, 0.2) * (1.0 - terrain);
        let land = vec3<f32>(0.15, 0.14, 0.12) * terrain;

        color = mix(ocean, land, terrain);
        color += grid * 0.05;
    }

    return vec4<f32>(color, 1.0);
}
