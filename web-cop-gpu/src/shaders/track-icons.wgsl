// CLASSIFICATION: UNCLASSIFIED
// src/shaders/track-icons.wgsl — Instanced track icon quad render pass

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
  @location(0)                          uv:           vec2<f32>,
  @location(1) @interpolate(flat)       icon_index:   u32,
  @location(2) @interpolate(flat)       threat_level: u32,
  @location(3) @interpolate(flat)       is_selected:  u32,
  @location(4) @interpolate(flat)       alert_flags:  u32,
  @location(5) @interpolate(flat)       course:       f32,
}

@group(0) @binding(0) var<uniform>          uniforms:        Uniforms;
@group(1) @binding(0) var<storage, read>    tracks:          array<TrackRecord>;
@group(1) @binding(1) var<storage, read>    positions:       array<vec4<f32>>;
@group(1) @binding(2) var<storage, read>    visible_indices: array<u32>;

const ICON_BASE_SIZE_PX: f32 = 14.0; // tiny and sharp

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

  let is_selected = u32(track.track_id_hash == uniforms.selected_track_id_hash);
  let scale = select(1.0, 1.3, is_selected > 0u);
  let size = ICON_BASE_SIZE_PX * scale;

  let clip = uniforms.view_proj * vec4<f32>(pos.x, pos.y, 0.0, 1.0);
  let w    = select(clip.w, 0.00001, abs(clip.w) < 0.00001);
  let ndc  = clip.xy / w;

  let uv     = QUAD_UVS[vid];
  let offset = (uv - vec2<f32>(0.5, 0.5)) * size / uniforms.viewport_size * 2.0;

  var out: VertexOutput;
  out.position     = vec4<f32>(ndc + offset, clip.z / w, 1.0);
  out.uv           = uv;
  out.icon_index   = track.icon_index;
  out.threat_level = track.threat_level;
  out.is_selected  = is_selected;
  out.alert_flags  = track.alert_flags;
  out.course       = track.course;
  return out;
}

fn threat_color(level: u32) -> vec3<f32> {
  // Aligned with TrackAffiliation enum in src/types/track-symbol.ts
  switch (level) {
    case 0u: { return vec3<f32>(0.96, 0.62, 0.04); } // UNKNOWN  — Yellow  (#f59e0b)
    case 1u: { return vec3<f32>(0.02, 0.71, 0.83); } // PENDING  — Cyan    (#06b6d4)
    case 2u: { return vec3<f32>(0.23, 0.51, 0.96); } // FRIENDLY — Blue    (#3b82f6)
    case 3u: { return vec3<f32>(0.13, 0.77, 0.37); } // NEUTRAL  — Green   (#22c55e)
    case 4u: { return vec3<f32>(0.98, 0.45, 0.09); } // SUSPECT  — Orange  (#f97316)
    case 5u: { return vec3<f32>(0.94, 0.27, 0.27); } // HOSTILE  — Red     (#ef4444)
    default: { return vec3<f32>(0.96, 0.62, 0.04); } // fallback UNKNOWN
  }
}

// Sharp, anti-aliased silhouettes using smoothstep and fwidth
// Domain mapping aligned with TrackDomain enum in src/types/track-symbol.ts:
//   0=AIR  1=SURFACE  2=SUBSURFACE  3=LAND  4=SPACE  5=CYBER
fn get_silhouette(uv: vec2<f32>, type_idx: u32, rotated_uv: vec2<f32>) -> f32 {
    let p = rotated_uv * 2.0 - 1.0;
    let p_fixed = uv * 2.0 - 1.0;
    let fw = fwidth(length(p)) * 1.5; // Edge softness based on screen-space derivative

    switch(type_idx % 6u) {
        case 0u: { // AIR: Pointed / Swept-back triangle (nose up)
            let d = max(abs(p.x) * 1.5 + p.y * 0.5, -p.y - 0.2);
            return 1.0 - smoothstep(0.7 - fw, 0.7 + fw, d);
        }
        case 1u: { // SURFACE: Diamond hull (naval vessel)
            let d = abs(p.x) + abs(p.y);
            return 1.0 - smoothstep(0.8 - fw, 0.8 + fw, d);
        }
        case 2u: { // SUBSURFACE: Wide horizontal ellipse (submarine)
            let d = (p_fixed.x * p_fixed.x * 2.5) + (p_fixed.y * p_fixed.y * 1.0);
            return 1.0 - smoothstep(0.8 - fw, 0.8 + fw, d);
        }
        case 3u: { // LAND: Square / rectangle (ground vehicle)
            let d = max(abs(p.x), abs(p.y));
            return 1.0 - smoothstep(0.75 - fw, 0.75 + fw, d);
        }
        case 4u: { // SPACE: Circle with pointed top fin (satellite)
            let body = length(vec2<f32>(p_fixed.x, p_fixed.y + 0.05));
            let fin  = max(abs(p.x) * 4.0 + (p.y + 0.6) * 2.0, -(p.y + 0.55));
            let body_s = 1.0 - smoothstep(0.72 - fw, 0.72 + fw, body);
            let fin_s  = 1.0 - smoothstep(0.55 - fw, 0.55 + fw, fin);
            return max(body_s, fin_s);
        }
        case 5u: { // CYBER: Hexagon (logical / non-physical domain)
            // Regular hexagon approximation via max of rotated coordinates
            let hex1 = abs(p.x);
            let hex2 = abs(p.x * 0.5 + p.y * 0.866);
            let hex3 = abs(p.x * 0.5 - p.y * 0.866);
            let d = max(hex1, max(hex2, hex3));
            return 1.0 - smoothstep(0.78 - fw, 0.78 + fw, d);
        }
        default: { return 0.0; }
    }
}

@fragment
fn fs_main(in: VertexOutput) -> @location(0) vec4<f32> {
  // Rotation for orientation
  let angle = in.course * 3.14159 / 180.0;
  let c = cos(angle);
  let s = sin(angle);
  let pivot = vec2<f32>(0.5, 0.5);
  let rotated_uv = vec2<f32>(
    c * (in.uv.x - pivot.x) - s * (in.uv.y - pivot.y) + pivot.x,
    s * (in.uv.x - pivot.x) + c * (in.uv.y - pivot.y) + pivot.y
  );

    let silhouette = get_silhouette(in.uv, in.icon_index, rotated_uv);
    let dist = length(in.uv - vec2<f32>(0.5, 0.5));

    // Smooth outline logic
    let outline_width = 0.05;
    let outer_silhouette = get_silhouette(in.uv, in.icon_index, rotated_uv);
    // We already have a smooth silhouette, so we can use it for the main shape.

    var color = threat_color(in.threat_level);
    var alpha = silhouette;

    // Add subtle dark outline for contrast
    let outline = smoothstep(0.45, 0.48, dist) * (1.0 - silhouette);
    color = mix(color, vec3<f32>(0.02, 0.05, 0.1), outline * 0.8);
    alpha = max(alpha, outline * 0.5);

    // Tactical Pulsing for Alerts (Slower, subtler glow)
    if (in.alert_flags > 0u) {
        let t = f32(uniforms.current_time_ms % 8000u) / 8000.0;
        let pulse = 0.6 + 0.4 * sin(t * 6.28318); // 0.2 to 1.0 range, 1 cycle per 8s
        let glow = smoothstep(0.5, 0.2, dist) * pulse;
        let anomaly_color = select(vec3<f32>(1.0, 0.8, 0.0), vec3<f32>(1.0, 0.2, 0.0), in.threat_level == 5u);
        color = mix(color, anomaly_color, glow * 0.5);
        alpha = max(alpha, glow * 0.3);
    }

    // Selection Glow (Cyan ring)
    if (in.is_selected > 0u) {
        let select_glow = smoothstep(0.5, 0.4, dist) * 0.4;
        color = mix(color, vec3<f32>(0.0, 1.0, 1.0), select_glow + 0.3);
        if (dist > 0.44 && dist < 0.5) {
            color = vec3<f32>(0.0, 1.0, 1.0);
            alpha = 1.0;
        }
    }

    if (alpha < 0.05) { discard; }
    return vec4<f32>(color, alpha);
}
