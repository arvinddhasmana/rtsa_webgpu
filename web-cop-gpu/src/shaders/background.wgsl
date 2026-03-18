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
  @location(0) world_pos: vec2<f32>,
}

@vertex
fn vs_main(@builtin(vertex_index) vid: u32) -> VertexOutput {
  var pos = array<vec2<f32>, 4>(
    vec2<f32>(-1.0, -1.0),
    vec2<f32>( 1.0, -1.0),
    vec2<f32>(-1.0,  1.0),
    vec2<f32>( 1.0,  1.0)
  );

  let ndc = pos[vid];

  // Transform NDC to World coordinates (Lon/Lat)
  let tx = uniforms.view_proj[3].x;
  let ty = uniforms.view_proj[3].y;
  let sx = uniforms.view_proj[0].x;
  let sy = uniforms.view_proj[1].y;

  let world_x = (ndc.x - tx) / sx;
  let world_y = (ndc.y - ty) / sy;

  var out: VertexOutput;
  out.position = vec4<f32>(ndc, 0.0, 1.0);
  out.world_pos = vec2<f32>(world_x, world_y);
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

// Fractal Brownian Motion for rich geographic detail
fn fbm(p: vec2<f32>) -> f32 {
    var v = 0.0;
    var a = 0.5;
    var shift = vec2<f32>(100.0);
    var pos = p;
    for (var i = 0; i < 5; i = i + 1) {
        v = v + a * noise(pos);
        pos = pos * 2.0 + shift;
        a = a * 0.5;
    }
    return v;
}

@fragment
fn fs_main(in: VertexOutput) -> @location(0) vec4<f32> {
    let style = uniforms.map_style;

    // Theme matched to mockup
    let standard_base = vec3<f32>(0.07, 0.12, 0.22); // Deep slate ocean
    let hd_base = vec3<f32>(0.05, 0.08, 0.15);       // Darker HD ocean
    var color = select(standard_base, hd_base, style == 1u);

    // High-Detail Fractal Terrain (Zoom-Aware)
    let zoom_scale = select(1.5, 2.2, style == 1u);
    let n = fbm(in.world_pos * zoom_scale);

    // Razor-sharp shorelines (Near zero threshold for crispness)
    let terrain = smoothstep(0.55, 0.555, n);

    let land_color = select(vec3<f32>(0.40, 0.38, 0.35), vec3<f32>(0.32, 0.30, 0.25), style == 1u);

    // Removed Coastline Glow to eliminate any "blur" perception
    color = mix(color, land_color, terrain);

    // Tactical Cyan Grid (Subtle and sharp)
    let grid_size = 64.0;
    let grid_uv = in.position.xy / grid_size;
    let grid_line = smoothstep(0.98, 1.0, fract(grid_uv.x)) + smoothstep(0.98, 1.0, fract(grid_uv.y));

    let grid_color = vec3<f32>(0.0, 1.0, 1.0); // Bright Cyan
    color += grid_line * grid_color * 0.1;

    return vec4<f32>(color, 1.0);
}
