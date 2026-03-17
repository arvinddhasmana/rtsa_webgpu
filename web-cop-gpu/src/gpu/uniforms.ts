// CLASSIFICATION: UNCLASSIFIED
// src/gpu/uniforms.ts — Uniform buffer writer
//
// Encodes the 80-byte Uniforms struct that all shaders share.
// Layout must match the WGSL struct in every shader file.
//
// struct Uniforms {
//   view_proj:       mat4x4<f32>,  // offset 0,  size 64
//   viewport_size:   vec2<f32>,    // offset 64, size 8
//   current_time_ms: u32,          // offset 72, size 4
//   track_count:     u32,          // offset 76, size 4
//   dashboard_mode:  u32,          // offset 80, size 4
//   padding:         vec3<u32>,    // offset 84, size 12
//   // Total: 96 bytes
// }
//
// Reference: docs/sdlc_guidelines/08_tech_specific/wgsl_shader_standards.md §6

export const UNIFORM_BYTES = 96;

// Pre-allocated reusable ArrayBuffer — zero per-frame allocation
const _uniformData = new ArrayBuffer(UNIFORM_BYTES);
const _f32 = new Float32Array(_uniformData);
const _u32 = new Uint32Array(_uniformData);

/**
 * Build an identity perspective-like view-projection matrix.
 * In a production implementation this would be derived from the map camera.
 *
 * Returns a column-major mat4x4 as a 16-element Float32Array.
 */
export function makeViewProjection(
  canvasWidth: number,
  canvasHeight: number,
  centerLon: number,
  centerLat: number,
  scale: number,
): Float32Array {
  // Simple orthographic projection mapping lon/lat → NDC
  // NDC_x = (lon - centerLon) * scale * aspectRatio^-1
  // NDC_y = (lat - centerLat) * scale
  const aspect = canvasHeight > 0 ? canvasWidth / canvasHeight : 1;
  const sx = scale / aspect;
  const sy = scale;

  // Column-major mat4x4 (WebGPU convention)
  return new Float32Array([
    sx,           0,            0, 0,
    0,            sy,           0, 0,
    0,            0,            1, 0,
    -centerLon * sx, -centerLat * sy, 0, 1,
  ]);
}

/**
 * Write the Uniforms struct into the GPU uniform buffer.
 * Uses a pre-allocated ArrayBuffer to avoid per-frame heap allocation.
 *
 * Reference: webgpu_guidelines.md §10 (no per-frame allocation)
 */
export function writeUniforms(
  device: GPUDevice,
  uniformBuffer: GPUBuffer,
  viewProj: Float32Array,
  canvasWidth: number,
  canvasHeight: number,
  currentTimeMs: number,
  trackCount: number,
  dashboardMode: "sensor" | "commander" | "analytics" | "health" | "coverage",
  selectedTrackIdHash: number = 0,
  mapStyle: number = 0, // 0: Standard, 1: HD
): void {
  // mat4x4<f32> at offset 0 (16 floats × 4 bytes = 64 bytes)
  _f32.set(viewProj, 0);

  // vec2<f32> viewport_size at offset 64 bytes = float index 16
  _f32[16] = canvasWidth;
  _f32[17] = canvasHeight;

  // u32 current_time_ms at offset 72 bytes = uint32 index 18
  _u32[18] = currentTimeMs >>> 0;

  // u32 track_count at offset 76 bytes = uint32 index 19
  _u32[19] = trackCount >>> 0;

  // u32 dashboard_mode at offset 80 bytes = uint32 index 20
  const modeMap = { sensor: 0, commander: 1, analytics: 2, health: 3, coverage: 4 };
  _u32[20] = modeMap[dashboardMode] ?? 0;

  // u32 selected_track_id_hash at offset 84 bytes = uint32 index 21
  _u32[21] = selectedTrackIdHash >>> 0;

  // u32 map_style at offset 88 bytes = uint32 index 22
  _u32[22] = mapStyle >>> 0;

  // Padding filled by 0 implicitly via pre-allocated ArrayBuffer

  device.queue.writeBuffer(uniformBuffer, 0, _uniformData);
}
