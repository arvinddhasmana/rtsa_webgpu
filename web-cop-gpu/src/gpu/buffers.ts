// CLASSIFICATION: UNCLASSIFIED
// src/gpu/buffers.ts — Pre-allocated GPU buffer management
//
// All buffers are created at startup to max capacity (65,536 tracks).
// No per-frame allocation. Buffer lifecycle is tied to the GPUDevice.
//
// Reference: docs/sdlc_guidelines/08_tech_specific/webgpu_guidelines.md §4

/** Maximum number of tracks supported in a single render cycle. */
export const MAX_TRACKS = 65_536;

/** Size of one track record in bytes (mirrors the SAB layout). */
export const TRACK_RECORD_BYTES = 128;

/** Size of one interpolated position entry: vec4<f32> = 16 bytes. */
export const POSITION_ENTRY_BYTES = 16;

/** Size of one visible-index entry: u32 = 4 bytes. */
export const INDEX_ENTRY_BYTES = 4;

/** Indirect draw args: vertex_count, instance_count, first_vertex, first_instance (4 × u32). */
export const DRAW_ARGS_BYTES = 16;

/** Uniform buffer size in bytes (matches Uniforms struct in all shaders). */
export const UNIFORM_BYTES = 96;

/** Number of glyph instances pre-allocated for SDF labels. */
export const MAX_GLYPH_INSTANCES = MAX_TRACKS * 8; // up to 8 chars per callsign

/** Bytes per GlyphInstance struct (40 bytes, 8 fields × 4 bytes + 1 pad). */
export const GLYPH_INSTANCE_BYTES = 40;

/** Maximum number of coverage records (sectors + gaps). */
export const MAX_COVERAGE_RECORDS = 1024;

/** Bytes per CoverageRecord struct (32 bytes). */
export const COVERAGE_RECORD_BYTES = 32;

/** Maximum number of raw observations to buffer for rendering. */
export const MAX_OBSERVATIONS = 10_000;

/** Size of one observation record (lon, lat, type, confidence) = 4 * f32 = 16 bytes. */
export const OBSERVATION_RECORD_BYTES = 16;

export interface GPUBuffers {
  /** Track records uploaded from the SAB each frame — STORAGE | COPY_DST */
  trackStorage: GPUBuffer;
  /** Interpolated positions from compute pass — STORAGE | COPY_DST */
  positions: GPUBuffer;
  /** Visible track indices written by culling pass — STORAGE | COPY_DST */
  visibleIndices: GPUBuffer;
  /** Indirect draw arguments (vertex_count, instance_count, ...) — INDIRECT | STORAGE */
  drawArgs: GPUBuffer;
  /** Uniform buffer — UNIFORM | COPY_DST */
  uniform: GPUBuffer;
  /** SDF glyph instances written per-frame (CPU) — STORAGE | COPY_DST */
  glyphInstances: GPUBuffer;
  /** Coverage records (sectors/gaps) written per-frame (CPU) — STORAGE | COPY_DST */
  coverageStorage: GPUBuffer;
  /** Raw observations uploaded per-frame (CPU) — STORAGE | COPY_DST */
  observationStorage: GPUBuffer;
}

/**
 * Allocate all GPU buffers at startup.
 * None of these are re-allocated during normal operation.
 *
 * Rule: No per-frame buffer creation (webgpu_guidelines.md §4.4, rule 2).
 */
export function allocateBuffers(device: GPUDevice): GPUBuffers {
  const trackStorage = device.createBuffer({
    label: "track-storage",
    size: MAX_TRACKS * TRACK_RECORD_BYTES,
    usage: GPUBufferUsage.STORAGE | GPUBufferUsage.COPY_DST,
  });

  const positions = device.createBuffer({
    label: "interpolated-positions",
    size: MAX_TRACKS * POSITION_ENTRY_BYTES,
    usage: GPUBufferUsage.STORAGE | GPUBufferUsage.COPY_DST,
  });

  const visibleIndices = device.createBuffer({
    label: "visible-indices",
    size: MAX_TRACKS * INDEX_ENTRY_BYTES,
    usage: GPUBufferUsage.STORAGE | GPUBufferUsage.COPY_DST,
  });

  const drawArgs = device.createBuffer({
    label: "draw-indirect-args",
    size: DRAW_ARGS_BYTES,
    usage: GPUBufferUsage.STORAGE | GPUBufferUsage.INDIRECT | GPUBufferUsage.COPY_DST,
  });

  const uniform = device.createBuffer({
    label: "uniforms",
    size: UNIFORM_BYTES,
    usage: GPUBufferUsage.UNIFORM | GPUBufferUsage.COPY_DST,
  });

  const glyphInstances = device.createBuffer({
    label: "glyph-instances",
    size: MAX_GLYPH_INSTANCES * GLYPH_INSTANCE_BYTES,
    usage: GPUBufferUsage.STORAGE | GPUBufferUsage.COPY_DST,
  });

  const coverageStorage = device.createBuffer({
    label: "coverage-storage",
    size: MAX_COVERAGE_RECORDS * COVERAGE_RECORD_BYTES,
    usage: GPUBufferUsage.STORAGE | GPUBufferUsage.COPY_DST,
  });

  const observationStorage = device.createBuffer({
    label: "observation-storage",
    size: MAX_OBSERVATIONS * OBSERVATION_RECORD_BYTES,
    usage: GPUBufferUsage.STORAGE | GPUBufferUsage.COPY_DST,
  });

  return {
    trackStorage,
    positions,
    visibleIndices,
    drawArgs,
    uniform,
    glyphInstances,
    coverageStorage,
    observationStorage,
  };
}

/**
 * Destroy all GPU buffers. Call this on device loss before re-initialising.
 */
export function destroyBuffers(buffers: GPUBuffers): void {
  buffers.trackStorage.destroy();
  buffers.positions.destroy();
  buffers.visibleIndices.destroy();
  buffers.drawArgs.destroy();
  buffers.uniform.destroy();
  buffers.glyphInstances.destroy();
  buffers.coverageStorage.destroy();
  buffers.observationStorage.destroy();
}
