// CLASSIFICATION: UNCLASSIFIED
import react from "@vitejs/plugin-react";
import path from "path";
import { defineConfig } from "vite";

const nm = path.resolve(__dirname, "node_modules");

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: [
      // Resolve gen/ts files
      { find: "@gen", replacement: path.resolve(__dirname, "../gen/ts") },
      // Resolve @bufbuild subpaths from web-cop/node_modules (exact match first)
      {
        find: /^@bufbuild\/protobuf\/codegenv2$/,
        replacement: `${nm}/@bufbuild/protobuf/dist/esm/codegenv2/index.js`,
      },
      {
        find: /^@bufbuild\/protobuf\/wkt$/,
        replacement: `${nm}/@bufbuild/protobuf/dist/esm/wkt/index.js`,
      },
      {
        find: /^@bufbuild\/protobuf\/reflect$/,
        replacement: `${nm}/@bufbuild/protobuf/dist/esm/reflect/index.js`,
      },
      {
        find: /^@bufbuild\/protobuf\/wire$/,
        replacement: `${nm}/@bufbuild/protobuf/dist/esm/wire/index.js`,
      },
      {
        find: /^@bufbuild\/protobuf$/,
        replacement: `${nm}/@bufbuild/protobuf/dist/esm/index.js`,
      },
      // Resolve @connectrpc packages from web-cop/node_modules
      {
        find: /^@connectrpc\/connect-web$/,
        replacement: `${nm}/@connectrpc/connect-web/dist/esm/index.js`,
      },
      {
        find: /^@connectrpc\/connect$/,
        replacement: `${nm}/@connectrpc/connect/dist/esm/index.js`,
      },
    ],
  },
  server: {
    port: 5173,
    strictPort: true,
    proxy: {
      // Forward gRPC-Web requests to Envoy dev proxy (accepts self-signed cert)
      "/rtsa.": {
        target: "https://localhost:8443",
        changeOrigin: true,
        secure: false, // Envoy dev uses self-signed cert
        ws: false,
      },
    },
  },
  build: {
    outDir: "dist",
    sourcemap: false,
  },
});
