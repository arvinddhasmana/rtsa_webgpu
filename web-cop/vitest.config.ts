// CLASSIFICATION: UNCLASSIFIED
import react from "@vitejs/plugin-react";
import path from "path";
import { defineConfig } from "vitest/config";

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
  test: {
    environment: "jsdom",
    globals: true,
    setupFiles: ["./src/__tests__/setup.ts"],
    exclude: ["e2e/**", "tests/e2e/**", "node_modules/**"],
    coverage: {
      provider: "v8",
      reporter: ["text", "json", "html"],
    },
  },
});
