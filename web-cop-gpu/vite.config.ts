// CLASSIFICATION: UNCLASSIFIED
import { defineConfig } from "vite";
import solid from "vite-plugin-solid";
import path from "path";

export default defineConfig({
  plugins: [solid()],
  resolve: {
    alias: {
      "@src": path.resolve(__dirname, "src"),
      "@gen": path.resolve(__dirname, "../gen/ts"),
    },
  },
  server: {
    port: 5174,
    strictPort: true,
    headers: {
      // Required for SharedArrayBuffer and WebTransport
      "Cross-Origin-Opener-Policy": "same-origin",
      "Cross-Origin-Embedder-Policy": "require-corp",
    },
  },
  worker: {
    format: "es",
  },
  build: {
    outDir: "dist",
    sourcemap: false,
    target: "es2022",
  },
  assetsInclude: ["**/*.wasm"],
});
