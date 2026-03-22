const fs = require('fs');
const content = fs.readFileSync('src/gpu/renderer.ts', 'utf8');
const fixed = content.replace(
  /\/\/ @ts-expect-error GPUAllowSharedBufferSource missing SAB\\n    sabTrackView,/g,
  "// @ts-expect-error WebGPU fully supports SharedArrayBuffer views\\n    sabTrackView,"
);
fs.writeFileSync('src/gpu/renderer.ts', fixed.replace(/\\n/g, '\n'));
