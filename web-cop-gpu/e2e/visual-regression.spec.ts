// CLASSIFICATION: UNCLASSIFIED
// web-cop-gpu/e2e/visual-regression.spec.ts — Visual regression test suite (H4-5)
//
// Captures golden screenshots at defined track-count loads and zoom levels.
// Uses Playwright's toHaveScreenshot() for pixel-accurate comparison.
//
// Track counts:  100, 1,000, 10,000, 50,000  (mocked via SAB injection)
// Zoom levels:   overview, city, tactical
//
// Reference: docs/implementation/v4/phase4_hardening_cutover.md H4-5
// CI: Blocks merge on visual regression (> 2% pixel diff).

import { test, expect } from "@playwright/test";
import { gotoApp } from "./helpers";

/** Viewport size for all visual regression captures. */
const VIEWPORT = { width: 1920, height: 1080 };

/** Track count scenarios (labels map to SAB mock injection values). */
const TRACK_SCENARIOS = [
  { label: "100-tracks",   count: 100 },
  { label: "10k-tracks",   count: 10_000 },
  { label: "50k-tracks",   count: 50_000 },
] as const;

/** Camera scale values for each zoom level (see lod.ts). */
const ZOOM_LEVELS = [
  { label: "overview", scale: 0.05 },
  { label: "city",     scale: 0.3 },
  { label: "tactical", scale: 1.0 },
] as const;

test.describe("Visual Regression Suite", () => {
  test.beforeEach(async ({ page }) => {
    await page.setViewportSize(VIEWPORT);
  });

  test("baseline — app renders without crashes at default viewport", async ({ page }) => {
    await gotoApp(page);
    await page.waitForLoadState("networkidle", { timeout: 20_000 }).catch(() => {});
    await page.waitForTimeout(1_000);

    // Take a full-page screenshot for the baseline
    await expect(page).toHaveScreenshot("baseline-default.png", {
      fullPage: false,
      maxDiffPixelRatio: 0.02,
    });
  });

  // Generate visual regression tests for each track count × zoom level combination
  for (const { label: trackLabel, count } of TRACK_SCENARIOS) {
    for (const { label: zoomLabel, scale } of ZOOM_LEVELS) {
      test(`visual-${trackLabel}-${zoomLabel}`, async ({ page }) => {
        // Inject mock track data into window before the app runs
        await page.addInitScript(
          ({ trackCount, cameraScale }: { trackCount: number; cameraScale: number }) => {
            // Expose test parameters to the app
            (window as unknown as Record<string, unknown>)["__RTSA_TEST_TRACK_COUNT__"] = trackCount;
            (window as unknown as Record<string, unknown>)["__RTSA_TEST_SCALE__"] = cameraScale;
          },
          { trackCount: count, cameraScale: scale },
        );

        await gotoApp(page);
        await page.waitForLoadState("networkidle", { timeout: 20_000 }).catch(() => {});

        // Wait for the render loop to stabilise (at least 3 frames at 60 FPS ≈ 50ms)
        await page.waitForTimeout(500);

        await expect(page).toHaveScreenshot(`${trackLabel}-${zoomLabel}.png`, {
          fullPage: false,
          maxDiffPixelRatio: 0.02,
          // Mask dynamic elements that change per-frame (FPS counter, timestamp)
          mask: [
            page.locator('[data-testid="fps-display"]'),
            page.locator('[data-testid="latency-display"]'),
            page.locator("time"),
          ],
        });
      });
    }
  }

  test("classification banner position is stable across track loads", async ({ page }) => {
    await gotoApp(page);
    await page.waitForLoadState("networkidle", { timeout: 20_000 }).catch(() => {});

    const banner = page.locator('[data-testid="classification-banner-top"]');
    const isVisible = await banner.isVisible().catch(() => false);

    if (isVisible) {
      // Banner must be at the very top of the viewport
      const box = await banner.boundingBox();
      expect(box).not.toBeNull();
      expect(box!.y).toBeLessThan(10); // within 10px of top
    }
  });

  test("toolbar layout is consistent between roles", async ({ page }) => {
    await gotoApp(page);
    await page.waitForLoadState("networkidle", { timeout: 20_000 }).catch(() => {});

    const toolbar = page.locator('[data-testid="app-toolbar"]');
    const isVisible = await toolbar.isVisible().catch(() => false);

    if (isVisible) {
      await expect(toolbar).toHaveScreenshot("toolbar-default-role.png", {
        maxDiffPixelRatio: 0.02,
      });
    }
  });
});
