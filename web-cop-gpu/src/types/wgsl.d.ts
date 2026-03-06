// CLASSIFICATION: UNCLASSIFIED
// src/types/wgsl.d.ts — TypeScript declaration for WGSL shader imports
//
// Vite supports importing files as raw strings via the `?raw` suffix.
// This declaration tells TypeScript the module type for WGSL shader files.

declare module "*.wgsl?raw" {
  const content: string;
  export default content;
}
