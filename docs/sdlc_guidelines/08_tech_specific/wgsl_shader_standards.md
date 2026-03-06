<!-- CLASSIFICATION: UNCLASSIFIED -->

# WGSL Shader Standards

> **Document**: RTSA WGSL Shader Coding Standards
> **Version**: 1.0
> **Classification**: UNCLASSIFIED
> **Last Updated**: 2026-03-05
> **Prerequisite**: Load `webgpu_guidelines.md` first

---

## 1. Overview

All GPU shaders in the RTSA COP are written in WGSL (WebGPU Shading Language). This document defines naming conventions, structural patterns, performance rules, and testing guidance for compute and render shaders.

### Shader Inventory

| Shader File          | Type              | Purpose                                      |
| -------------------- | ----------------- | -------------------------------------------- |
| `interpolation.wgsl` | Compute           | Extrapolate track positions between updates  |
| `culling.wgsl`       | Compute           | View-frustum culling → visible instance list |
| `track-icons.wgsl`   | Vertex + Fragment | Instanced track icon quads                   |
| `trail.wgsl`         | Vertex + Fragment | Track trail polylines                        |
| `halos.wgsl`         | Vertex + Fragment | Alert halo circles around tracks             |
| `labels.wgsl`        | Vertex + Fragment | SDF text labels                              |
| `pick.wgsl`          | Fragment          | Write `track_id_hash` to pick buffer         |

---

## 2. File & Naming Conventions

### 2.1 File Names

- `kebab-case.wgsl` — one shader module per file
- Colocated in `src/shaders/`

### 2.2 Entry Point Names

| Shader Type | Entry Point | Convention                            |
| ----------- | ----------- | ------------------------------------- |
| Compute     | `main`      | Single entry point per compute shader |
| Vertex      | `vs_main`   | Prefix `vs_`                          |
| Fragment    | `fs_main`   | Prefix `fs_`                          |

### 2.3 Struct Names

```wgsl
// PascalCase for all struct types
struct TrackRecord {
  lon: f32,
  lat: f32,
  course: f32,
  speed: f32,
  // ...
}

struct Uniforms {
  view_proj: mat4x4<f32>,
  viewport_size: vec2<f32>,
  time_ms: u32,
  track_count: u32,
}

struct VertexOutput {
  @builtin(position) position: vec4<f32>,
  @location(0) uv: vec2<f32>,
  @location(1) icon_index: u32,
}
```

### 2.4 Variable Names

| Scope                | Convention         | Example                                           |
| -------------------- | ------------------ | ------------------------------------------------- |
| Struct fields        | `snake_case`       | `track_id_hash`                                   |
| Local variables      | `snake_case`       | `screen_pos`                                      |
| Constants            | `UPPER_SNAKE_CASE` | `MAX_TRACKS`                                      |
| Binding group labels | `snake_case`       | `@group(0) @binding(0) var<storage, read> tracks` |

---

## 3. Binding Layout Conventions

### 3.1 Group Assignment

All shaders use a consistent bind group layout:

| Group       | Purpose             | Contents                                   |
| ----------- | ------------------- | ------------------------------------------ |
| `@group(0)` | Per-frame uniforms  | `Uniforms` buffer                          |
| `@group(1)` | Storage buffers     | Track data, instance output, indirect args |
| `@group(2)` | Textures + Samplers | Icon atlas, SDF atlas, pick texture        |

### 3.2 Binding Order Within Groups

Bindings within a group follow declaration order in the TypeScript pipeline definition. Do not reorder WGSL bindings without updating the corresponding TypeScript `GPUBindGroupLayout`.

### 3.3 Example

```wgsl
// Group 0: Uniforms (updated per-frame)
@group(0) @binding(0) var<uniform> uniforms: Uniforms;

// Group 1: Storage (track data)
@group(1) @binding(0) var<storage, read> tracks: array<TrackRecord>;
@group(1) @binding(1) var<storage, read_write> instances: array<InstanceData>;

// Group 2: Textures
@group(2) @binding(0) var icon_atlas: texture_2d<f32>;
@group(2) @binding(1) var icon_sampler: sampler;
```

---

## 4. Compute Shader Patterns

### 4.1 Standard Workgroup Size

```wgsl
@compute @workgroup_size(256)
fn main(@builtin(global_invocation_id) gid: vec3<u32>) {
  let idx = gid.x;
  if (idx >= uniforms.track_count) { return; }

  let track = tracks[idx];
  // ... process track ...
  instances[idx] = result;
}
```

**Rules**:

- Use `@workgroup_size(256)` for track-parallel shaders (optimal for most GPUs)
- Always guard with `if (idx >= count) { return; }` to handle non-multiple dispatch sizes
- Dispatch as: `ceil(trackCount / 256)` workgroups on the x-axis

### 4.2 Interpolation Compute Shader Pattern

```wgsl
// CLASSIFICATION: UNCLASSIFIED
// interpolation.wgsl — extrapolate track positions between server updates

struct TrackRecord {
  lon: f32,
  lat: f32,
  course: f32,
  speed: f32,
  altitude: f32,
  track_id_hash: u32,
  source_bitmap: u32,
  classification_level: u32,
  threat_level: u32,
  icon_index: u32,
  alert_flags: u32,
  update_epoch_ms: u32,
  trail: array<vec4<f32>, 5>,
}

struct Uniforms {
  view_proj: mat4x4<f32>,
  viewport_size: vec2<f32>,
  current_time_ms: u32,
  track_count: u32,
}

@group(0) @binding(0) var<uniform> uniforms: Uniforms;
@group(1) @binding(0) var<storage, read> tracks: array<TrackRecord>;
@group(1) @binding(1) var<storage, read_write> positions: array<vec4<f32>>;

const DEG_TO_RAD: f32 = 0.017453292;
const EARTH_RADIUS_M: f32 = 6371000.0;

@compute @workgroup_size(256)
fn main(@builtin(global_invocation_id) gid: vec3<u32>) {
  let idx = gid.x;
  if (idx >= uniforms.track_count) { return; }

  let track = tracks[idx];
  let dt_s = f32(uniforms.current_time_ms - track.update_epoch_ms) / 1000.0;

  // Dead-reckoning extrapolation
  let dx = track.speed * sin(track.course) * dt_s / EARTH_RADIUS_M;
  let dy = track.speed * cos(track.course) * dt_s / EARTH_RADIUS_M;

  positions[idx] = vec4<f32>(
    track.lon + dx,
    track.lat + dy,
    track.altitude,
    f32(track.track_id_hash)
  );
}
```

### 4.3 Culling Compute Shader Pattern

```wgsl
// Writes to indirect draw buffer and visible instance list
@compute @workgroup_size(256)
fn main(@builtin(global_invocation_id) gid: vec3<u32>) {
  let idx = gid.x;
  if (idx >= uniforms.track_count) { return; }

  let pos = positions[idx];
  let clip = uniforms.view_proj * vec4<f32>(pos.xy, 0.0, 1.0);
  let ndc = clip.xy / clip.w;

  // Frustum test (with margin for icon size)
  if (abs(ndc.x) < 1.2 && abs(ndc.y) < 1.2) {
    let slot = atomicAdd(&draw_args.instance_count, 1u);
    visible_indices[slot] = idx;
  }
}
```

---

## 5. Vertex / Fragment Shader Patterns

### 5.1 Instanced Track Icons

```wgsl
// CLASSIFICATION: UNCLASSIFIED
// track-icons.wgsl

struct VertexOutput {
  @builtin(position) position: vec4<f32>,
  @location(0) uv: vec2<f32>,
  @location(1) @interpolate(flat) icon_index: u32,
  @location(2) @interpolate(flat) threat_level: u32,
}

// Triangle strip quad vertices (4 verts per instance)
const QUAD_UVS = array<vec2<f32>, 4>(
  vec2<f32>(0.0, 0.0),
  vec2<f32>(1.0, 0.0),
  vec2<f32>(0.0, 1.0),
  vec2<f32>(1.0, 1.0),
);

@vertex
fn vs_main(
  @builtin(vertex_index) vid: u32,
  @builtin(instance_index) iid: u32,
) -> VertexOutput {
  let track_idx = visible_indices[iid];
  let pos = positions[track_idx];
  let track = tracks[track_idx];

  let clip = uniforms.view_proj * vec4<f32>(pos.xy, 0.0, 1.0);
  let icon_size = vec2<f32>(32.0, 32.0); // pixels
  let offset = (QUAD_UVS[vid] - 0.5) * icon_size / uniforms.viewport_size * 2.0;

  var out: VertexOutput;
  out.position = vec4<f32>(clip.xy / clip.w + offset, clip.z / clip.w, 1.0);
  out.uv = QUAD_UVS[vid];
  out.icon_index = track.icon_index;
  out.threat_level = track.threat_level;
  return out;
}

@fragment
fn fs_main(in: VertexOutput) -> @location(0) vec4<f32> {
  // Calculate atlas UV from icon_index
  let atlas_cols = 32u; // 2048 / 64
  let col = in.icon_index % atlas_cols;
  let row = in.icon_index / atlas_cols;
  let atlas_uv = (vec2<f32>(f32(col), f32(row)) + in.uv) / vec2<f32>(f32(atlas_cols), f32(atlas_cols));

  let color = textureSample(icon_atlas, icon_sampler, atlas_uv);
  if (color.a < 0.1) { discard; }
  return color;
}
```

### 5.2 SDF Label Rendering

```wgsl
// Fragment shader for Signed Distance Field text
@fragment
fn fs_main(in: VertexOutput) -> @location(0) vec4<f32> {
  let dist = textureSample(sdf_atlas, sdf_sampler, in.uv).r;
  let edge = 0.5;
  let smoothing = fwidth(dist) * 0.5;
  let alpha = smoothstep(edge - smoothing, edge + smoothing, dist);

  if (alpha < 0.01) { discard; }
  return vec4<f32>(in.text_color.rgb, alpha);
}
```

---

## 6. Uniform Buffer Layout Rules

### 6.1 Alignment

WGSL requires specific alignment for struct members:

| Type          | Alignment | Size     |
| ------------- | --------- | -------- |
| `f32`         | 4 bytes   | 4 bytes  |
| `u32`         | 4 bytes   | 4 bytes  |
| `vec2<f32>`   | 8 bytes   | 8 bytes  |
| `vec3<f32>`   | 16 bytes  | 12 bytes |
| `vec4<f32>`   | 16 bytes  | 16 bytes |
| `mat4x4<f32>` | 16 bytes  | 64 bytes |

### 6.2 Uniform Struct Layout

Place `mat4x4` first, then `vec4`, `vec2`, then scalars. Pad to 16-byte boundary:

```wgsl
struct Uniforms {
  view_proj: mat4x4<f32>,    // offset 0, size 64
  viewport_size: vec2<f32>,  // offset 64, size 8
  current_time_ms: u32,      // offset 72, size 4
  track_count: u32,          // offset 76, size 4
  // Total: 80 bytes (16-byte aligned ✓)
}
```

### 6.3 TypeScript Mirror

The TypeScript code that writes uniform data **must** match the WGSL struct layout byte-for-byte:

```typescript
// CLASSIFICATION: UNCLASSIFIED

const uniformData = new ArrayBuffer(80);
const f32View = new Float32Array(uniformData);
const u32View = new Uint32Array(uniformData);

// mat4x4 at offset 0 (16 floats)
f32View.set(viewProjMatrix, 0);
// vec2 at offset 64 bytes = 16 floats
f32View[16] = canvasWidth;
f32View[17] = canvasHeight;
// u32 at offset 72 bytes = 18 floats
u32View[18] = performance.now() | 0;
u32View[19] = trackCount;

device.queue.writeBuffer(uniformBuffer, 0, uniformData);
```

---

## 7. Performance Rules

### 7.1 Compute Shaders

| Rule                               | Rationale                                         |
| ---------------------------------- | ------------------------------------------------- |
| Workgroup size 256                 | Matches common GPU warp/wavefront sizes           |
| No barriers unless strictly needed | `workgroupBarrier()` serializes all threads       |
| Avoid divergent branches           | GPUs execute in lockstep; divergence wastes lanes |
| Read storage buffers sequentially  | Coalesced reads are faster than random access     |

### 7.2 Render Shaders

| Rule                                      | Rationale                                          |
| ----------------------------------------- | -------------------------------------------------- |
| Use `discard` sparingly                   | Disables early-Z; only use for alpha cutout        |
| `@interpolate(flat)` for integer varyings | Avoids undefined behavior on integer interpolation |
| Minimize texture samples per fragment     | Each sample has latency; max 2 per fragment shader |
| Use `smoothstep` for SDF, not branching   | GPU-friendly smooth alpha                          |

### 7.3 General

| Rule                                    | Rationale                                           |
| --------------------------------------- | --------------------------------------------------- |
| No dynamic indexing into uniform arrays | Uniform buffer access must be statically resolvable |
| Avoid `textureLoad` in loops            | Use `textureSample` with appropriate LOD            |
| Keep shader modules small               | Smaller modules compile faster at init              |
| Use `const` for compile-time constants  | Enables compiler optimization                       |

---

## 8. Testing Shaders

### 8.1 Strategy

WGSL shaders cannot be directly unit-tested in Vitest. Testing strategy:

1. **Visual regression tests**: Playwright screenshots comparing rendered output against golden images
2. **Compute shader output tests**: Run compute pass, `mapAsync` output buffer, verify values in TypeScript
3. **Pick buffer tests**: Render known track positions, read pick buffer, verify correct `track_id_hash`

### 8.2 Compute Shader Test Pattern

```typescript
// CLASSIFICATION: UNCLASSIFIED

// In Playwright or headless WebGPU test harness:
const output = await readGPUBuffer(device, instanceBuffer, trackCount * 16);
const instances = new Float32Array(output);
expect(instances[0]).toBeCloseTo(expectedLon, 5);
expect(instances[1]).toBeCloseTo(expectedLat, 5);
```

---

## 9. Code Review Checklist

When reviewing WGSL changes:

- [ ] All bindings match TypeScript `GPUBindGroupLayout` definitions
- [ ] Uniform struct alignment matches TypeScript `ArrayBuffer` layout
- [ ] Workgroup size is 256 for track-parallel compute shaders
- [ ] Out-of-bounds guard `if (idx >= count) { return; }` present
- [ ] No per-frame allocations in TypeScript pipeline code
- [ ] `@interpolate(flat)` on all integer/uint varyings
- [ ] `discard` usage justified and documented
- [ ] Shader file uses `kebab-case.wgsl` naming

---

## 10. Cross-References

| Document                           | Path                                                              |
| ---------------------------------- | ----------------------------------------------------------------- |
| WebGPU Guidelines                  | `docs/sdlc_guidelines/08_tech_specific/webgpu_guidelines.md`      |
| FlatBuffers Guidelines             | `docs/sdlc_guidelines/08_tech_specific/flatbuffers_guidelines.md` |
| v1 Architecture — GPU Section      | `docs/architecture/v1/RTSA_WebGPU_Architecture_v1.md` §4–5        |
| Component Design — Render Pipeline | `docs/architecture/component_design.md` §5                        |
