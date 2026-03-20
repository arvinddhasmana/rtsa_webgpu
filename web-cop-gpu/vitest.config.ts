// CLASSIFICATION: UNCLASSIFIED
import { defineConfig } from "vitest/config";
import solid from "vite-plugin-solid";
import path from "path";

export default defineConfig({
  plugins: [solid()],
  resolve: {
    alias: {
      "@src": path.resolve(__dirname, "src"),
      "@gen": path.resolve(__dirname, "../gen/ts"),
      // @bufbuild/protobuf is imported by generated files in ../gen/ts which live
      // outside this package directory.  Vite cannot resolve node_modules by
      // walking up from that external path, so we provide an explicit alias that
      // always points to the copy installed inside web-cop-gpu.
      "@bufbuild/protobuf": path.resolve(
        __dirname,
        "node_modules/@bufbuild/protobuf",
      ),
    },
  },
  test: {
    environment: "jsdom",
    globals: true,
    include: ["tests/**/*.test.ts", "tests/**/*.test.tsx", "src/**/*.test.ts", "src/**/*.test.tsx"],
    coverage: {
      provider: "v8",
      reporter: ["text", "json", "html"],
      include: ["src/**/*.ts", "src/**/*.tsx"],
      exclude: ["src/**/*.d.ts"],
    },
  },
  server: {
    fs: {
      allow: [__dirname, path.resolve(__dirname, "..")],
    },
  },
});
