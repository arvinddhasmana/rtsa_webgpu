// CLASSIFICATION: UNCLASSIFIED
// src/gpu/bind-groups.ts — GPU bind group creation for all pipelines
//
// Bind groups are created once at init and reused each frame.
// They wire up the uniform buffer, storage buffers, and atlas textures
// to each pipeline stage.
//
// Reference: docs/sdlc_guidelines/08_tech_specific/wgsl_shader_standards.md §3

import { AtlasTextures } from "./atlas";
import { GPUBuffers } from "./buffers";
import { PickResources } from "./pick";
import { AllPipelines } from "./pipelines";

export interface BindGroups {
  /** Bind groups for the interpolation compute pass */
  interpolation: { g0: GPUBindGroup; g1: GPUBindGroup };
  /** Bind groups for the culling compute pass */
  culling:       { g0: GPUBindGroup; g1: GPUBindGroup };
  /** Bind groups for the track icon render pass */
  trackIcons:    { g0: GPUBindGroup; g1: GPUBindGroup; g2: GPUBindGroup };
  /** Bind groups for the trail render pass */
  trail:         { g0: GPUBindGroup; g1: GPUBindGroup };
  /** Bind groups for the halo render pass */
  halos:         { g0: GPUBindGroup; g1: GPUBindGroup };
  /** Bind groups for the labels render pass */
  labels:        { g0: GPUBindGroup; g1: GPUBindGroup; g2: GPUBindGroup };
  /** Bind groups for the pick render pass */
  pick:          { g0: GPUBindGroup; g1: GPUBindGroup };
  /** Bind groups for raw observations */
  observations:  { g0: GPUBindGroup; g1: GPUBindGroup };
}

/**
 * Create all bind groups wiring buffers and textures to pipeline layouts.
 * Called once at init after pipelines and buffers are created.
 */
export function createBindGroups(
  device: GPUDevice,
  pipelines: AllPipelines,
  buffers: GPUBuffers,
  atlas: AtlasTextures,
  _pick: PickResources,
): BindGroups {
  // --- Interpolation compute ---
  const interpG0 = device.createBindGroup({
    label:  "interp-g0",
    layout: pipelines.compute.interpolation.getBindGroupLayout(0),
    entries: [{ binding: 0, resource: { buffer: buffers.uniform } }],
  });
  const interpG1 = device.createBindGroup({
    label:  "interp-g1",
    layout: pipelines.compute.interpolation.getBindGroupLayout(1),
    entries: [
      { binding: 0, resource: { buffer: buffers.trackStorage } },
      { binding: 1, resource: { buffer: buffers.positions } },
    ],
  });

  // --- Culling compute ---
  const cullG0 = device.createBindGroup({
    label:  "cull-g0",
    layout: pipelines.compute.culling.getBindGroupLayout(0),
    entries: [{ binding: 0, resource: { buffer: buffers.uniform } }],
  });
  const cullG1 = device.createBindGroup({
    label:  "cull-g1",
    layout: pipelines.compute.culling.getBindGroupLayout(1),
    entries: [
      { binding: 0, resource: { buffer: buffers.positions } },
      { binding: 1, resource: { buffer: buffers.visibleIndices } },
      { binding: 2, resource: { buffer: buffers.drawArgs } },
    ],
  });

  // --- Track icons render ---
  const iconsG0 = device.createBindGroup({
    label:  "icons-g0",
    layout: pipelines.render.trackIcons.getBindGroupLayout(0),
    entries: [{ binding: 0, resource: { buffer: buffers.uniform } }],
  });
  const iconsG1 = device.createBindGroup({
    label:  "icons-g1",
    layout: pipelines.render.trackIcons.getBindGroupLayout(1),
    entries: [
      { binding: 0, resource: { buffer: buffers.trackStorage } },
      { binding: 1, resource: { buffer: buffers.positions } },
      { binding: 2, resource: { buffer: buffers.visibleIndices } },
    ],
  });
  const iconsG2 = device.createBindGroup({
    label:  "icons-g2",
    layout: pipelines.render.trackIcons.getBindGroupLayout(2),
    entries: [
      { binding: 0, resource: atlas.iconAtlas.createView() },
      { binding: 1, resource: atlas.iconSampler },
    ],
  });

  // --- Trail render ---
  const trailG0 = device.createBindGroup({
    label:  "trail-g0",
    layout: pipelines.render.trail.getBindGroupLayout(0),
    entries: [{ binding: 0, resource: { buffer: buffers.uniform } }],
  });
  const trailG1 = device.createBindGroup({
    label:  "trail-g1",
    layout: pipelines.render.trail.getBindGroupLayout(1),
    entries: [
      { binding: 0, resource: { buffer: buffers.trackStorage } },
      { binding: 1, resource: { buffer: buffers.visibleIndices } },
    ],
  });

  // --- Halos render ---
  const halosG0 = device.createBindGroup({
    label:  "halos-g0",
    layout: pipelines.render.halos.getBindGroupLayout(0),
    entries: [{ binding: 0, resource: { buffer: buffers.uniform } }],
  });
  const halosG1 = device.createBindGroup({
    label:  "halos-g1",
    layout: pipelines.render.halos.getBindGroupLayout(1),
    entries: [
      { binding: 0, resource: { buffer: buffers.trackStorage } },
      { binding: 1, resource: { buffer: buffers.positions } },
      { binding: 2, resource: { buffer: buffers.visibleIndices } },
    ],
  });

  // --- Labels render ---
  const labelsG0 = device.createBindGroup({
    label:  "labels-g0",
    layout: pipelines.render.labels.getBindGroupLayout(0),
    entries: [{ binding: 0, resource: { buffer: buffers.uniform } }],
  });
  const labelsG1 = device.createBindGroup({
    label:  "labels-g1",
    layout: pipelines.render.labels.getBindGroupLayout(1),
    entries: [{ binding: 0, resource: { buffer: buffers.glyphInstances } }],
  });
  const labelsG2 = device.createBindGroup({
    label:  "labels-g2",
    layout: pipelines.render.labels.getBindGroupLayout(2),
    entries: [
      { binding: 0, resource: atlas.sdfAtlas.createView() },
      { binding: 1, resource: atlas.sdfSampler },
    ],
  });

  // --- Pick render ---
  const pickG0 = device.createBindGroup({
    label:  "pick-g0",
    layout: pipelines.render.pick.getBindGroupLayout(0),
    entries: [{ binding: 0, resource: { buffer: buffers.uniform } }],
  });
  const pickG1 = device.createBindGroup({
    label:  "pick-g1",
    layout: pipelines.render.pick.getBindGroupLayout(1),
    entries: [
      { binding: 0, resource: { buffer: buffers.trackStorage } },
      { binding: 1, resource: { buffer: buffers.positions } },
      { binding: 2, resource: { buffer: buffers.visibleIndices } },
    ],
  });

  // --- Observations render ---
  const obsG0 = device.createBindGroup({
    label:  "obs-g0",
    layout: pipelines.render.observations.getBindGroupLayout(0),
    entries: [{ binding: 0, resource: { buffer: buffers.uniform } }],
  });
  const obsG1 = device.createBindGroup({
    label:  "obs-g1",
    layout: pipelines.render.observations.getBindGroupLayout(1),
    entries: [{ binding: 0, resource: { buffer: buffers.observationStorage } }],
  });

  return {
    interpolation: { g0: interpG0, g1: interpG1 },
    culling:       { g0: cullG0,   g1: cullG1   },
    trackIcons:    { g0: iconsG0,  g1: iconsG1,  g2: iconsG2  },
    trail:         { g0: trailG0,  g1: trailG1               },
    halos:         { g0: halosG0,  g1: halosG1               },
    labels:        { g0: labelsG0, g1: labelsG1, g2: labelsG2 },
    pick:          { g0: pickG0,   g1: pickG1                },
    observations:  { g0: obsG0,    g1: obsG1                 },
  };
}
