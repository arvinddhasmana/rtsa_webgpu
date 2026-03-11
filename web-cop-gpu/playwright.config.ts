// CLASSIFICATION: UNCLASSIFIED
// web-cop-gpu/playwright.config.ts — Playwright configuration for Phase 4 E2E suite
//
// Covers all workflows in docs/implementation/v4/phase4_hardening_cutover.md H4-4
// and visual regression (H4-5).
//
// Reference: docs/sdlc_guidelines/05_testing/testing_strategy.md

import { defineConfig, devices } from "@playwright/test";

export default defineConfig({
  testDir: "./e2e",
  timeout: 60_000,
  retries: process.env.CI ? 2 : 1,
  reporter: [
    ["html", { open: "never", outputFolder: "playwright-report" }],
    ["list"],
    ["json", { outputFile: "playwright-results.json" }],
  ],
  use: {
    baseURL: process.env.BASE_URL ?? "http://localhost:5174",
    trace: "on-first-retry",
    screenshot: "only-on-failure",
    video: "retain-on-failure",
  },
  projects: [
    {
      name: "chromium",
      use: {
        ...devices["Desktop Chrome"],
        // WebGPU requires the flag to be enabled in headless Chromium.
        // --disable-web-security is intentionally NOT included: removing it
        // ensures security-header and CSP assertions run in a realistic context.
        channel: "chromium",
        launchOptions: {
          args: [
            "--enable-unsafe-webgpu",
            "--enable-features=WebGPU",
          ],
        },
      },
    },
  ],
  // webServer block is commented out — run `npm run dev` separately before tests.
  webServer: {
    command: "npm run dev",
    url: "http://localhost:5174",
    reuseExistingServer: !process.env.CI,
    timeout: 90_000,
  },
  snapshotDir: "./e2e/snapshots",
  expect: {
    // Allow up to 2% pixel difference for visual regression (anti-aliasing, subpixel)
    toHaveScreenshot: { maxDiffPixelRatio: 0.02 },
  },
});
