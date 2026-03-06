// CLASSIFICATION: UNCLASSIFIED
// tests/gpu/buffers.test.ts — Unit tests for GPU buffer size constants and logic
//
// Tests the buffer sizing calculations without requiring an actual GPUDevice.
// All size values are verified against the architecture spec.
//
// Reference: docs/sdlc_guidelines/08_tech_specific/webgpu_guidelines.md §4.3

import { describe, it, expect } from "vitest";
import {
  MAX_TRACKS,
  TRACK_RECORD_BYTES,
  POSITION_ENTRY_BYTES,
  INDEX_ENTRY_BYTES,
  DRAW_ARGS_BYTES,
  UNIFORM_BYTES,
  GLYPH_INSTANCE_BYTES,
  MAX_GLYPH_INSTANCES,
} from "../../src/gpu/buffers";

describe("GPU buffer size constants", () => {
  it("MAX_TRACKS is 65,536", () => {
    expect(MAX_TRACKS).toBe(65_536);
  });

  it("TRACK_RECORD_BYTES is 128 (matches SAB layout)", () => {
    expect(TRACK_RECORD_BYTES).toBe(128);
  });

  it("POSITION_ENTRY_BYTES is 16 (vec4<f32>)", () => {
    expect(POSITION_ENTRY_BYTES).toBe(16);
  });

  it("INDEX_ENTRY_BYTES is 4 (u32)", () => {
    expect(INDEX_ENTRY_BYTES).toBe(4);
  });

  it("DRAW_ARGS_BYTES is 16 (4 × u32)", () => {
    expect(DRAW_ARGS_BYTES).toBe(16);
  });

  it("UNIFORM_BYTES is 80 (matches WGSL Uniforms struct)", () => {
    // mat4x4<f32>=64 + vec2<f32>=8 + u32=4 + u32=4 = 80
    expect(UNIFORM_BYTES).toBe(80);
  });

  it("GLYPH_INSTANCE_BYTES is 40", () => {
    expect(GLYPH_INSTANCE_BYTES).toBe(40);
  });

  it("MAX_GLYPH_INSTANCES equals MAX_TRACKS × 8", () => {
    expect(MAX_GLYPH_INSTANCES).toBe(MAX_TRACKS * 8);
  });

  it("track storage total does not exceed 512 MB VRAM budget", () => {
    const trackStorageBytes = MAX_TRACKS * TRACK_RECORD_BYTES;       // 8 MB
    const positionsBytes    = MAX_TRACKS * POSITION_ENTRY_BYTES;     // 1 MB
    const indicesBytes      = MAX_TRACKS * INDEX_ENTRY_BYTES;        // 256 KB
    const glyphsBytes       = MAX_GLYPH_INSTANCES * GLYPH_INSTANCE_BYTES; // ~20 MB

    // Icon atlas: 2048×2048×4 = 16 MB
    // SDF atlas:  2048×1024×1 = 2 MB
    const iconAtlasBytes    = 2048 * 2048 * 4;
    const sdfAtlasBytes     = 2048 * 1024 * 1;

    const totalBytes =
      trackStorageBytes + positionsBytes + indicesBytes +
      glyphsBytes + iconAtlasBytes + sdfAtlasBytes;

    const BUDGET_BYTES = 512 * 1024 * 1024;
    expect(totalBytes).toBeLessThan(BUDGET_BYTES);
  });
});
