<!-- CLASSIFICATION: UNCLASSIFIED -->

# Phase 1 — Core Rendering

> **Document**: v4 Implementation — Phase 1
> **Version**: 1.0
> **Classification**: UNCLASSIFIED
> **Status**: Not Started
> **Prerequisite Phases**: Phase 0 (Foundation)
> **Parallel With**: Phase 2 (Backend Integration)
> **Architecture Reference**: `docs/architecture/v1/RTSA_WebGPU_Architecture_v1.md` §4, §5

---

## 1. Objective

Build the complete WebGPU rendering pipeline in the Render Worker — from device initialization through compute passes (interpolation, culling) to all render layers (icons, trails, halos, labels) and the pick buffer for O(1) track selection.

---

## 2. Deliverables

| #     | Deliverable                  | Description                                           |
| ----- | ---------------------------- | ----------------------------------------------------- |
| R1-1  | WebGPU device + context      | Adapter/device acquisition on OffscreenCanvas         |
| R1-2  | Buffer allocation            | Track storage, instance, uniform, pick, atlas buffers |
| R1-3  | Interpolation compute shader | Dead-reckoning position extrapolation                 |
| R1-4  | Culling compute shader       | View-frustum culling → indirect draw args             |
| R1-5  | Track icon render pass       | Instanced quads with atlas sampling                   |
| R1-6  | Trail line render pass       | Polylines from trail ring buffer                      |
| R1-7  | Alert halo render pass       | Animated circles for alerted tracks                   |
| R1-8  | SDF label render pass        | GPU-rendered text labels (callsign, speed)            |
| R1-9  | Pick buffer render pass      | `track_id_hash` output to pick texture                |
| R1-10 | Map tile layer               | Raster or vector tile background rendering            |
| R1-11 | Icon + SDF atlas             | Baked atlas textures (APP-6 icons, font glyphs)       |
| R1-12 | Per-frame pipeline           | Orchestrated compute → render pass sequence           |
| R1-13 | Tests + visual baselines     | Compute output tests, Playwright visual regression    |

---

## 3. Detailed Tasks

### R1-1: WebGPU Device + Context

- Implement `initGPU()` per `docs/sdlc_guidelines/08_tech_specific/webgpu_guidelines.md` §3
- Request `high-performance` adapter
- Handle device loss with re-initialization
- Configure canvas context with `premultiplied` alpha

### R1-2: Buffer Allocation

Allocate all buffers at startup per `webgpu_guidelines.md` §4.3:

| Buffer          | Size (50k tracks) | Usage                           |
| --------------- | ----------------- | ------------------------------- |
| Track storage   | 6.1 MB            | `STORAGE \| COPY_DST`           |
| Instance output | 3.1 MB            | `STORAGE \| VERTEX`             |
| Uniform         | 80 B              | `UNIFORM \| COPY_DST`           |
| Pick texture    | ~8.3 MB           | `RENDER_ATTACHMENT \| COPY_SRC` |
| Icon atlas      | 16 MB             | `TEXTURE_BINDING \| COPY_DST`   |
| SDF atlas       | 2 MB              | `TEXTURE_BINDING \| COPY_DST`   |
| Trail vertices  | 3.8 MB            | `STORAGE \| VERTEX`             |

**Rule**: All buffers pre-allocated to max capacity (65,536 tracks). No per-frame allocation.

### R1-3: Interpolation Compute Shader

- Implement `interpolation.wgsl` per `docs/sdlc_guidelines/08_tech_specific/wgsl_shader_standards.md` §4.2
- Dead-reckoning extrapolation based on course, speed, time delta
- Read track storage buffer → write position output buffer
- Workgroup size 256, dispatch `ceil(trackCount / 256)`

### R1-4: Culling Compute Shader

- Implement `culling.wgsl` per `wgsl_shader_standards.md` §4.3
- View-frustum test with icon-size margin
- Write visible track indices to instance list
- Atomically increment indirect draw argument `instanceCount`

### R1-5: Track Icon Render Pass

- Implement `track-icons.wgsl` per `wgsl_shader_standards.md` §5.1
- Instanced triangle-strip quads (4 vertices × N instances)
- Sample icon atlas using `icon_index` from track record
- Billboard facing (screen-aligned quads regardless of zoom)

### R1-6: Trail Line Render Pass

- Implement `trail.wgsl`
- Read trail ring buffer (5 × 2 points per track) from track storage
- Render as `LINE_STRIP` or emulated line quads
- Color by threat level

### R1-7: Alert Halo Render Pass

- Implement `halos.wgsl`
- Render animated circle outlines around tracks with `alert_flags != 0`
- Pulse animation driven by `uniforms.current_time_ms`
- Only render for visible (post-cull) tracks with active alerts

### R1-8: SDF Label Render Pass

- Implement `labels.wgsl` per `wgsl_shader_standards.md` §5.2
- Render callsign + speed text using SDF font atlas
- Position labels offset from track icon
- `smoothstep` alpha for smooth edges at all zoom levels

### R1-9: Pick Buffer Render Pass

- Implement `pick.wgsl` per `webgpu_guidelines.md` §7
- Separate render target (R32Uint or RGBA8 encoding track_id_hash)
- No multisampling on pick target
- `readBufferAsync` on click events only
- Half-resolution option for VRAM savings

### R1-10: Map Tile Layer

- Raster tile loading (pre-fetched tile pyramid)
- Or vector tile decoding → GPU rendering
- Rendered as first layer before track data
- Tile cache in Render Worker (not main thread)

### R1-11: Icon + SDF Atlas

- Bake NATO APP-6 icons into 2048×2048 RGBA atlas at build time
- Bake SDF font glyphs into 2048×1024 R8 atlas using `msdf-atlas-gen` or equivalent
- Load atlas textures during init, upload to GPU textures
- Reference: `webgpu_guidelines.md` §6

### R1-12: Per-Frame Pipeline

Orchestrate the full per-frame sequence per `webgpu_guidelines.md` §5.1:

```
1. Read SAB → writeBuffer (track data upload)
2. Compute: interpolation
3. Compute: culling
4. Render: map tiles (loadOp: clear)
5. Render: trail lines (loadOp: load)
6. Render: track icons (loadOp: load)
7. Render: alert halos (loadOp: load)
8. Render: SDF labels (loadOp: load)
9. Render: pick buffer (separate target)
10. Present frame
```

### R1-13: Tests + Visual Baselines

| Test                         | Type              | Description                                         |
| ---------------------------- | ----------------- | --------------------------------------------------- |
| Interpolation compute output | Unit              | Feed known positions, verify extrapolated results   |
| Culling compute output       | Unit              | Feed known viewport, verify visible indices         |
| Pick buffer read             | Unit              | Place known tracks, verify picked `track_id_hash`   |
| Full scene screenshot        | Visual regression | Playwright golden image at 100, 1k, 10k, 50k tracks |
| FPS at 50k tracks            | Performance       | Must sustain 60 FPS for 30 seconds                  |

---

## 4. Performance Budget

Per `docs/architecture/component_design.md` §6:

| Phase                      | Budget                                      |
| -------------------------- | ------------------------------------------- |
| SAB read + upload          | ≤ 2 ms                                      |
| Compute passes             | ≤ 1 ms                                      |
| Render passes (all layers) | ≤ 4 ms                                      |
| GPU present                | ≤ 1 ms                                      |
| **Total frame**            | **≤ 8 ms (leaves 8 ms headroom at 60 FPS)** |

---

## 5. Mock Data Strategy

Until Phase 2 delivers real WebTransport data, the Render Worker uses mock data:

- Generate 50,000 mock tracks in SAB at startup
- Random positions within a defined bounding box
- Animate with random course/speed values
- This allows full pipeline development and profiling independent of backend

---

## 6. Done Gate

| Criteria                                     | Verification                         |
| -------------------------------------------- | ------------------------------------ |
| All 7 shader files compile and run           | WebGPU validation passes             |
| 50k tracks render at ≥ 55 FPS (5% tolerance) | Playwright perf test                 |
| Pick buffer returns correct track on click   | Automated test                       |
| SDF labels readable at all zoom levels       | Visual inspection                    |
| Trail lines connect to track positions       | Visual inspection                    |
| Alert halos pulse on alerted tracks          | Visual inspection                    |
| GPU memory usage < 100 MB at 50k tracks      | `adapter.requestAdapterInfo()` check |
| All compute shader tests pass                | Vitest + GPU test harness            |
| Visual regression baselines committed        | CI golden images                     |
