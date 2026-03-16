// CLASSIFICATION: UNCLASSIFIED
// src/shaders/halos.wgsl — Alert halo animated circle render pass
//
// Renders animated pulsing circle outlines around tracks with non-zero alert_flags.
// Pulse frequency and scale driven by uniforms.current_time_ms.
//
// Reference: docs/implementation/v4/phase1_core_rendering.md R1-7

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
  @builtin(position)              position:     vec4<f32>,
  @location(0)                    uv:           vec2<f32>,    // -1..1 local quad
  @location(1) @interpolate(flat) alert_flags:  u32,
  @location(2) @interpolate(flat) threat_level: u32,
  @location(3)                    pulse:        f32,
}

@group(0) @binding(0) var<uniform>       uniforms:        Uniforms;
@group(1) @binding(0) var<storage, read> tracks:          array<TrackRecord>;
@group(1) @binding(1) var<storage, read> positions:       array<vec4<f32>>;
@group(1) @binding(2) var<storage, read> visible_indices: array<u32>;

// Halo base radius in screen pixels
const HALO_RADIUS_PX:  f32 = 24.0;
// Pulse amplitude (fraction of radius)
const PULSE_AMPLITUDE: f32 = 0.3;
// Pulse frequency in Hz
const PULSE_FREQ_HZ:   f32 = 1.5;
const TWO_PI:          f32 = 6.283185307;

// Quad UV offsets for a centred billboard quad
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

  // Skip rendering (degenerate quad) if no alert flags set
  if (track.alert_flags == 0u) {
    var out: VertexOutput;
    out.position    = vec4<f32>(0.0, 0.0, 0.0, 0.0); // degenerate — clipped
    out.uv          = vec2<f32>(0.0, 0.0);
    out.alert_flags = 0u;
    out.threat_level = track.threat_level;
    out.pulse       = 0.0;
    return out;
  }

  // Animated pulse: sin wave on the halo radius
  let time_s = f32(uniforms.current_time_ms) / 1000.0;
  let pulse  = 1.0 + PULSE_AMPLITUDE * sin(time_s * PULSE_FREQ_HZ * TWO_PI);

  // Compute billboard offset
  let clip    = uniforms.view_proj * vec4<f32>(pos.x, pos.y, 0.0, 1.0);
  let w       = select(clip.w, 0.00001, abs(clip.w) < 0.00001);
  let ndc     = clip.xy / w;
  let local   = QUAD_OFFSETS[vid];
  let radius  = HALO_RADIUS_PX * pulse;
  let offset  = local * radius / (uniforms.viewport_size * 0.5);

  var out: VertexOutput;
  out.position     = vec4<f32>(ndc + offset, clip.z / w, 1.0);
  out.uv           = local;
  out.alert_flags  = track.alert_flags;
  out.threat_level = track.threat_level;
  out.pulse        = pulse;
  return out;
}

@fragment
fn fs_main(in: VertexOutput) -> @location(0) vec4<f32> {
  // Discard immediately for non-alerted tracks (degenerate quads)
  if (in.alert_flags == 0u) { discard; }

  // SDF circle: distance from centre in local [-1..1] space
  let dist = length(in.uv);

  // Ring: render only the outer ring (annulus)
  let ring_outer = 1.0;
  let ring_inner = 0.7;
  if (dist > ring_outer || dist < ring_inner) { discard; }

  // Smooth edge fade
  let outer_alpha = smoothstep(ring_outer, ring_outer - 0.1, dist);
  let inner_alpha = smoothstep(ring_inner, ring_inner + 0.1, dist);
  let alpha = outer_alpha * inner_alpha * 0.85;

  // Colour based on threat level — STANAG APP-6 compliant.
  // Halos are only rendered for tracks with alert_flags > 0 (checked in vs_main),
  // so colour here applies to Suspect (4) and Hostile (5) tracks.
  var color: vec3<f32>;
  switch (in.threat_level) {
    case 5u:  { color = vec3<f32>(1.0, 0.2, 0.2); } // Hostile — red
    case 4u:  { color = vec3<f32>(1.0, 0.6, 0.0); } // Suspect — amber/orange
    default:  { color = vec3<f32>(0.2, 0.8, 1.0); } // lower threats — blue
  }

  return vec4<f32>(color, alpha);
}
