// CLASSIFICATION: UNCLASSIFIED
import path from "path";
import { defineConfig } from "vite";
import solid from "vite-plugin-solid";

// Content-Security-Policy for the WebGPU COP.
// Validated against docs/sdlc_guidelines/04_coding_standards/solidjs_standards.md §7.3
//
// 'wasm-unsafe-eval' is required by the Rust Wasm decoder module.
// 'connect-src' in dev also allows http: and ws: so that Vite HMR (ws://localhost)
// and the gRPC-Web gateway (http://localhost:8080) are not blocked during development.
// In production, connect-src is tightened to https: and wss: only.
// No 'unsafe-inline' scripts are permitted in either environment.
const isDev = process.env.NODE_ENV !== "production";

const CSP = [
  "default-src 'self'",
  "script-src 'self' 'wasm-unsafe-eval'",
  "style-src 'self' 'unsafe-inline'", // SolidJS injects minimal inline styles
  "img-src 'self' data: blob: https://*.tile.openstreetmap.org https://*.cartocdn.com https://server.arcgisonline.com",
  // In dev: allow http: (gRPC-Web at localhost:8080) and ws: (Vite HMR).
  // In prod: restrict to https: and wss: only.
  isDev
    ? "connect-src 'self' https: wss: http://localhost:* ws://localhost:*"
    : "connect-src 'self' https: wss:",
  "worker-src 'self' blob:",
  "font-src 'self' data:",
  "object-src 'none'",
  "base-uri 'self'",
  "frame-ancestors 'none'",
].join("; ");

export default defineConfig({
  plugins: [solid()],
  resolve: {
    alias: {
      "@src": path.resolve(__dirname, "src"),
      "@gen": path.resolve(__dirname, "../gen/ts"),
      "@bufbuild/protobuf": path.resolve(
        __dirname,
        "node_modules/@bufbuild/protobuf",
      ),
      "@connectrpc/connect": path.resolve(
        __dirname,
        "node_modules/@connectrpc/connect",
      ),
      "@connectrpc/connect-web": path.resolve(
        __dirname,
        "node_modules/@connectrpc/connect-web",
      ),
    },
  },
  server: {
    port: 5174,
    strictPort: true,
    headers: {
      // Required for SharedArrayBuffer and WebTransport (H4-3 Security Audit)
      "Cross-Origin-Opener-Policy": "same-origin",
      "Cross-Origin-Embedder-Policy": "require-corp",
      // Content-Security-Policy — blocks XSS and injection attacks
      "Content-Security-Policy": CSP,
      // Defence-in-depth headers
      "X-Content-Type-Options": "nosniff",
      "X-Frame-Options": "DENY",
      "Referrer-Policy": "strict-origin-when-cross-origin",
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
