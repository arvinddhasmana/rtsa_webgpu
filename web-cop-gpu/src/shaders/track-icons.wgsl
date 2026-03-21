// CLASSIFICATION: UNCLASSIFIED
// src/shaders/track-icons.wgsl — Instanced track icon quad render pass
//
// Renders MIL-STD-2525 symbology for each visible track:
//   • Outer affiliation frame  — shape keyed by threat_level
//       UNKNOWN (0)  → quatrefoil   FRIENDLY (2) → rounded rectangle
//       PENDING (1)  → quatrefoil (dashed outline)
//       NEUTRAL (3)  → circle       SUSPECT (4)  → diamond
//       HOSTILE (5)  → diamond
//   • Inner domain icon        — shape keyed by entity_type extracted from icon_index
//       Unspecified/Air (0/2) → swept-wing triangle
//       Surface (1)           → ship hull diamond
//       Subsurface (3)        → horizontal ellipse
//       Land (4)              → filled square
//       Cyber (5)             → hexagon
//   • Context (icon_index / 36): 0=MILITARY solid fill, 1=CIVILIAN outline-only
//
// icon_index encoding: context * 36 + entity_type * 6 + threat_level
//
// Reference: docs/implementation/v5/operations_commander/plan-mil2525SymbologyWestAsiaDemo.md §Phase 2
//            docs/sdlc_guidelines/04_coding_standards/solidjs_standards.md §3

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

// MIL-STD-2525 icons need more space for internal domain icons
const ICON_BASE_SIZE_PX: f32 = 40.0;

// Affiliation threat-level constants used in frame/rendering logic
const THREAT_LEVEL_UNKNOWN:  u32 = 0u;
const THREAT_LEVEL_PENDING:  u32 = 1u;
const THREAT_LEVEL_FRIENDLY: u32 = 2u;
const THREAT_LEVEL_NEUTRAL:  u32 = 3u;
const THREAT_LEVEL_SUSPECT:  u32 = 4u;
const THREAT_LEVEL_HOSTILE:  u32 = 5u;

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

// ── MIL-STD-2525 affiliation colour palette ───────────────────────────────────
// Aligned with AFFILIATION_FILL in src/types/track-symbol.ts and plan §Phase 2
fn affiliation_color(level: u32) -> vec3<f32> {
  switch (level) {
    case 0u: { return vec3<f32>(0.98, 0.75, 0.14); } // UNKNOWN  — Yellow   #FBBF24
    case 1u: { return vec3<f32>(0.50, 0.80, 1.00); } // PENDING  — Lt Blue  #80CCFF
    case 2u: { return vec3<f32>(0.22, 0.74, 1.00); } // FRIENDLY — Cyan-Bl  #38BDFF
    case 3u: { return vec3<f32>(0.34, 0.90, 0.53); } // NEUTRAL  — Green    #57E688
    case 4u: { return vec3<f32>(1.00, 0.60, 0.20); } // SUSPECT  — Orange   #FF9933
    case 5u: { return vec3<f32>(0.97, 0.44, 0.44); } // HOSTILE  — Red      #F87171
    default: { return vec3<f32>(0.98, 0.75, 0.14); } // fallback UNKNOWN
  }
}

// ── Outer affiliation frame SDF ────────────────────────────────────────────────
// p  — normalised coordinates in [-1, 1] range
// Returns SDF distance: negative = inside shape, positive = outside.
//
// Affiliation shapes (MIL-STD-2525D):
//   UNKNOWN (0)  → quatrefoil (4 overlapping circles)
//   PENDING (1)  → quatrefoil (same shape; dashed outline applied in fs_main)
//   FRIENDLY (2) → square/rectangle
//   NEUTRAL (3)  → circle
//   SUSPECT (4)  → diamond (45° rotated square)
//   HOSTILE (5)  → diamond
fn sdf_frame(p: vec2<f32>, affiliation: u32) -> f32 {
  switch (affiliation) {
    case 2u: { // FRIENDLY — square frame
      return max(abs(p.x) - 0.70, abs(p.y) - 0.70);
    }
    case 5u, 4u: { // HOSTILE / SUSPECT — diamond
      return abs(p.x) + abs(p.y) - 0.82;
    }
    case 3u: { // NEUTRAL — circle
      return length(p) - 0.76;
    }
    case 0u, 1u: { // UNKNOWN / PENDING — quatrefoil (union of 4 offset circles)
      let r  = 0.44;
      let cx = 0.36;
      let d0 = length(p - vec2<f32>( cx, 0.0)) - r;
      let d1 = length(p - vec2<f32>(-cx, 0.0)) - r;
      let d2 = length(p - vec2<f32>(0.0,  cx)) - r;
      let d3 = length(p - vec2<f32>(0.0, -cx)) - r;
      return min(min(d0, d1), min(d2, d3));
    }
    default: { return length(p) - 0.76; }
  }
}

// ── Inner domain icon silhouette ──────────────────────────────────────────────
// p  — normalised coordinates in [-1, 1] range (same space as frame SDF)
// entity_type — decoded from icon_index as (icon_index % 36) / 6
//   0=Unspecified(default Air), 1=Surface, 2=Air, 3=Subsurface, 4=Land, 5=Cyber
//
// Returns silhouette weight in [0, 1]. Icon is rendered at ~50% of frame size.
fn get_inner_icon(p: vec2<f32>, entity_type: u32) -> f32 {
  // Scale p to the inner icon region (icons use ±0.45 of the [-1,1] space)
  let q   = p * 2.2;
  let fw  = fwidth(length(q)) * 1.5;

  switch (entity_type % 6u) {
    case 0u, 2u: { // Unspecified / Air: swept-back triangle (nose up)
      let d = max(abs(q.x) * 1.5 + q.y * 0.5, -q.y - 0.2);
      return 1.0 - smoothstep(0.68 - fw, 0.68 + fw, d);
    }
    case 1u: { // Surface: diamond hull (naval vessel)
      let d = abs(q.x) + abs(q.y);
      return 1.0 - smoothstep(0.75 - fw, 0.75 + fw, d);
    }
    case 3u: { // Subsurface: wide horizontal ellipse (submarine)
      let d = q.x * q.x * 2.0 + q.y * q.y * 5.0;
      return 1.0 - smoothstep(0.72 - fw, 0.72 + fw, d);
    }
    case 4u: { // Land: square/rectangle (ground vehicle)
      let d = max(abs(q.x), abs(q.y));
      return 1.0 - smoothstep(0.60 - fw, 0.60 + fw, d);
    }
    case 5u: { // Cyber: hexagon (logical / non-physical domain)
      let hex1 = abs(q.x);
      let hex2 = abs(q.x * 0.5 + q.y * 0.866);
      let hex3 = abs(q.x * 0.5 - q.y * 0.866);
      let d = max(hex1, max(hex2, hex3));
      return 1.0 - smoothstep(0.72 - fw, 0.72 + fw, d);
    }
    default: {
      let d = max(abs(q.x) * 1.5 + q.y * 0.5, -q.y - 0.2);
      return 1.0 - smoothstep(0.68 - fw, 0.68 + fw, d);
    }
  }
}

@fragment
fn fs_main(in: VertexOutput) -> @location(0) vec4<f32> {
  // Decode icon_index components
  let entity_type = (in.icon_index % 36u) / 6u;
  let is_civilian = (in.icon_index / 36u) == 1u;

  // Rotate UV by course for oriented icons.
  // track.course is stored in radians (0 = north, clockwise) — use directly.
  let angle = in.course;
  let c     = cos(angle);
  let s     = sin(angle);
  let pivot = vec2<f32>(0.5, 0.5);
  let ruv   = vec2<f32>(
    c * (in.uv.x - pivot.x) - s * (in.uv.y - pivot.y) + pivot.x,
    s * (in.uv.x - pivot.x) + c * (in.uv.y - pivot.y) + pivot.y,
  );

  // Map to [-1, 1] normalised space
  let p = in.uv * 2.0 - 1.0;
  let p_rot = ruv * 2.0 - 1.0;

  // Affiliation frame SDF
  let frame_d = sdf_frame(p, in.threat_level);
  let fw      = max(fwidth(frame_d), 0.01);

  var color = affiliation_color(in.threat_level);
  var alpha: f32;

  if is_civilian {
    // ── Civilian: outline-only ──────────────────────────────────────────
    let outline_hw = 0.10; // half-width of the outline stroke
    alpha = 1.0 - smoothstep(outline_hw - fw, outline_hw + fw, abs(frame_d));
    // Faint inner icon overlay (only inside the frame boundary)
    // NOTE: edge0 must be < edge1 per WGSL spec; was previously swapped (fw,-fw)
    let icon_w = get_inner_icon(p_rot, entity_type);
    alpha = max(alpha, icon_w * 0.35 * (1.0 - smoothstep(-fw, fw, frame_d)));
  } else {
    // ── Military: filled frame ──────────────────────────────────────────
    let fill = 1.0 - smoothstep(-fw, fw, frame_d);
    alpha = fill * 0.85;

    // Subtle dark outline at frame boundary for contrast
    let outline = (1.0 - smoothstep(-0.04, 0.04, frame_d)) *
                   smoothstep(-0.18, -0.06, frame_d);
    color = mix(color, vec3<f32>(0.02, 0.05, 0.10), outline * 0.7);

    // Inner domain icon — rendered slightly darker inside the filled area
    let icon_w  = get_inner_icon(p_rot, entity_type);
    let in_fill = 1.0 - smoothstep(-fw, fw, frame_d); // same as fill
    color = mix(color,
                vec3<f32>(color.r * 0.5, color.g * 0.5, color.b * 0.5),
                icon_w * in_fill * 0.5);

    // Pending (affiliation=PENDING): dashed outline via stripe pattern
    if in.threat_level == THREAT_LEVEL_PENDING {
      // Use a world-space stripe along the icon UV to create dashes
      let stripe = step(0.5, fract((in.uv.x + in.uv.y) * 5.0));
      // Only apply to the outline band
      let at_outline = (1.0 - smoothstep(-0.04, 0.04, frame_d)) *
                        smoothstep(-0.22, -0.08, frame_d);
      alpha = mix(alpha, alpha * stripe, at_outline);
    }
  }

  // ── Alert pulsing (Suspect / Hostile) ────────────────────────────────────
  if (in.alert_flags > 0u) {
    let t     = f32(uniforms.current_time_ms % 8000u) / 8000.0;
    let pulse = 0.6 + 0.4 * sin(t * 6.28318);
    let dist  = length(p);
    let glow  = smoothstep(0.5, 0.2, dist) * pulse;
    let anomaly_color = select(
      vec3<f32>(1.0, 0.8, 0.0),
      vec3<f32>(1.0, 0.2, 0.0),
      in.threat_level == 5u,
    );
    color = mix(color, anomaly_color, glow * 0.5);
    alpha = max(alpha, glow * 0.3);
  }

  // ── Selection glow (cyan ring) ────────────────────────────────────────────
  if (in.is_selected > 0u) {
    let dist        = length(p);
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

