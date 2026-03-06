// CLASSIFICATION: UNCLASSIFIED
// src/gpu/pipelines.ts — GPU compute and render pipeline creation
//
// All pipelines are created once at init. Never re-created per-frame.
//
// Reference: docs/sdlc_guidelines/08_tech_specific/webgpu_guidelines.md §5.4

// Import WGSL shader source — Vite bundles these as strings
import interpolationWGSL from "../shaders/interpolation.wgsl?raw";
import cullingWGSL       from "../shaders/culling.wgsl?raw";
import trackIconsWGSL    from "../shaders/track-icons.wgsl?raw";
import trailWGSL         from "../shaders/trail.wgsl?raw";
import halosWGSL         from "../shaders/halos.wgsl?raw";
import labelsWGSL        from "../shaders/labels.wgsl?raw";
import pickWGSL          from "../shaders/pick.wgsl?raw";

export interface ComputePipelines {
  interpolation: GPUComputePipeline;
  culling:       GPUComputePipeline;
}

export interface RenderPipelines {
  trackIcons: GPURenderPipeline;
  trail:      GPURenderPipeline;
  halos:      GPURenderPipeline;
  labels:     GPURenderPipeline;
  pick:       GPURenderPipeline;
}

export interface AllPipelines {
  compute: ComputePipelines;
  render:  RenderPipelines;
}

/**
 * Create all compute and render pipelines.
 * This function is called once at init. Pipeline compilation is expensive.
 *
 * Reference: webgpu_guidelines.md §5.4 (pipeline caching)
 */
export function createPipelines(
  device: GPUDevice,
  swapChainFormat: GPUTextureFormat,
): AllPipelines {
  // --- Compute pipelines ---

  const interpolationPipeline = device.createComputePipeline({
    label:   "interpolation",
    layout:  "auto",
    compute: {
      module:     device.createShaderModule({ label: "interpolation", code: interpolationWGSL }),
      entryPoint: "main",
    },
  });

  const cullingPipeline = device.createComputePipeline({
    label:   "culling",
    layout:  "auto",
    compute: {
      module:     device.createShaderModule({ label: "culling", code: cullingWGSL }),
      entryPoint: "main",
    },
  });

  // --- Render pipelines ---

  // Track icon pipeline: instanced triangle-strip quads, alpha cutout
  const trackIconsPipeline = device.createRenderPipeline({
    label:  "track-icons",
    layout: "auto",
    vertex: {
      module:     device.createShaderModule({ label: "track-icons-vs", code: trackIconsWGSL }),
      entryPoint: "vs_main",
    },
    fragment: {
      module:     device.createShaderModule({ label: "track-icons-fs", code: trackIconsWGSL }),
      entryPoint: "fs_main",
      targets: [
        {
          format: swapChainFormat,
          blend: {
            color: { srcFactor: "src-alpha", dstFactor: "one-minus-src-alpha", operation: "add" },
            alpha: { srcFactor: "one",       dstFactor: "one-minus-src-alpha", operation: "add" },
          },
        },
      ],
    },
    primitive: { topology: "triangle-strip" },
  });

  // Trail pipeline: emulated line quads (triangle list)
  const trailPipeline = device.createRenderPipeline({
    label:  "trail",
    layout: "auto",
    vertex: {
      module:     device.createShaderModule({ label: "trail-vs", code: trailWGSL }),
      entryPoint: "vs_main",
    },
    fragment: {
      module:     device.createShaderModule({ label: "trail-fs", code: trailWGSL }),
      entryPoint: "fs_main",
      targets: [
        {
          format: swapChainFormat,
          blend: {
            color: { srcFactor: "src-alpha", dstFactor: "one-minus-src-alpha", operation: "add" },
            alpha: { srcFactor: "one",       dstFactor: "one-minus-src-alpha", operation: "add" },
          },
        },
      ],
    },
    primitive: { topology: "triangle-list" },
  });

  // Halo pipeline: animated billboard circles
  const halosPipeline = device.createRenderPipeline({
    label:  "halos",
    layout: "auto",
    vertex: {
      module:     device.createShaderModule({ label: "halos-vs", code: halosWGSL }),
      entryPoint: "vs_main",
    },
    fragment: {
      module:     device.createShaderModule({ label: "halos-fs", code: halosWGSL }),
      entryPoint: "fs_main",
      targets: [
        {
          format: swapChainFormat,
          blend: {
            color: { srcFactor: "src-alpha", dstFactor: "one-minus-src-alpha", operation: "add" },
            alpha: { srcFactor: "one",       dstFactor: "one-minus-src-alpha", operation: "add" },
          },
        },
      ],
    },
    primitive: { topology: "triangle-strip" },
  });

  // Labels pipeline: SDF text quad strips
  const labelsPipeline = device.createRenderPipeline({
    label:  "labels",
    layout: "auto",
    vertex: {
      module:     device.createShaderModule({ label: "labels-vs", code: labelsWGSL }),
      entryPoint: "vs_main",
    },
    fragment: {
      module:     device.createShaderModule({ label: "labels-fs", code: labelsWGSL }),
      entryPoint: "fs_main",
      targets: [
        {
          format: swapChainFormat,
          blend: {
            color: { srcFactor: "src-alpha", dstFactor: "one-minus-src-alpha", operation: "add" },
            alpha: { srcFactor: "one",       dstFactor: "one-minus-src-alpha", operation: "add" },
          },
        },
      ],
    },
    primitive: { topology: "triangle-strip" },
  });

  // Pick pipeline: separate R32Uint render target — NO blending
  const pickPipeline = device.createRenderPipeline({
    label:  "pick",
    layout: "auto",
    vertex: {
      module:     device.createShaderModule({ label: "pick-vs", code: pickWGSL }),
      entryPoint: "vs_main",
    },
    fragment: {
      module:     device.createShaderModule({ label: "pick-fs", code: pickWGSL }),
      entryPoint: "fs_main",
      targets: [{ format: "r32uint" }], // No blending on pick target
    },
    primitive: { topology: "triangle-strip" },
  });

  return {
    compute: {
      interpolation: interpolationPipeline,
      culling:       cullingPipeline,
    },
    render: {
      trackIcons: trackIconsPipeline,
      trail:      trailPipeline,
      halos:      halosPipeline,
      labels:     labelsPipeline,
      pick:       pickPipeline,
    },
  };
}
