// CLASSIFICATION: UNCLASSIFIED
// web-cop/e2e/map.spec.ts
// Browser E2E tests for Map Rendering — CR-UI-001, CR-UI-002
// Covers: UC001 (Sensor Ingestion), UC002 (Multi-Sensor Fusion), UC006 (COP Display)

import { test, expect } from "@playwright/test";
import { waitForMapLoad } from "./helpers";

test.describe("Map Rendering (CR-UI-001, CR-UI-002)", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/");
    // Allow the app to bootstrap; map may not load in headless without MapLibre token
    await page.waitForLoadState("networkidle", { timeout: 15_000 }).catch(() => {});
  });

  test("map renders and shows canvas element [UC006]", async ({ page }) => {
    // The map container or a canvas should be present in the DOM
    const mapContainer = page.locator('[data-testid="map-container"], canvas, .maplibregl-map, #map');
    await expect(mapContainer.first()).toBeVisible({ timeout: 20_000 });
  });

  test("map container has non-zero dimensions [CR-UI-001]", async ({ page }) => {
    const mapEl = page.locator('[data-testid="map-container"], canvas, .maplibregl-map, #map');
    const el = mapEl.first();
    await expect(el).toBeVisible({ timeout: 20_000 });
    const box = await el.boundingBox();
    expect(box).not.toBeNull();
    expect(box!.width).toBeGreaterThan(0);
    expect(box!.height).toBeGreaterThan(0);
  });

  test("page title or heading identifies RTSA COP [UC006]", async ({ page }) => {
    // The COP app should have a recognizable title or heading
    const title = await page.title();
    const bodyText = await page.locator("body").innerText().catch(() => "");
    expect(
      title.toLowerCase().includes("rtsa") ||
      title.toLowerCase().includes("cop") ||
      bodyText.toLowerCase().includes("unclassified") ||
      bodyText.toLowerCase().includes("rtsa")
    ).toBeTruthy();
  });
});
