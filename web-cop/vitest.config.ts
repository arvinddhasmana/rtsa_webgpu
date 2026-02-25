// CLASSIFICATION: UNCLASSIFIED
import react from "@vitejs/plugin-react";
import { defineConfig } from "vitest/config";

export default defineConfig({
  plugins: [react()],
  test: {
    environment: "jsdom",
    globals: true,
    setupFiles: ["./src/__tests__/setup.ts"],
    exclude: ["e2e/**", "node_modules/**"],
    coverage: {
      provider: "v8",
      reporter: ["text", "json", "html"],
      thresholds: {
        lines: 80,
        branches: 80,
        functions: 80,
        statements: 80,
      },
      include: ["src/**/*.{ts,tsx}"],
      exclude: [
        "src/__tests__/**",
        "src/main.tsx",
        "src/App.tsx",
        "src/vite-env.d.ts",
        "src/api/**",
        // Map components require maplibre-gl which is unavailable in jsdom
        "src/components/map/MapView.tsx",
        "src/components/map/ThreatHaloLayer.tsx",
        "src/components/map/GeoFenceLayer.tsx",
        "src/components/map/SensorCoverageLayer.tsx",
        // MainLayout orchestrates streaming + map (tested via integration)
        "src/components/layout/MainLayout.tsx",
        // Auth provider is a thin wrapper; tested via integration
        "src/components/auth/AuthProvider.tsx",
      ],
    },
  },
});
