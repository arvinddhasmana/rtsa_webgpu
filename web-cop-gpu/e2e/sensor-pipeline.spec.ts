// CLASSIFICATION: UNCLASSIFIED
// web-cop-gpu/e2e/sensor-pipeline.spec.ts — End-to-End Pipeline Verification
//
// Verifies that spatial alerts (gaps) and coverage metadata flow from the
// simulated backend all the way to the WebGPU rendered map.

import { expect, test } from "@playwright/test";
import { gotoApp } from "./helpers";

test.describe("Sensor Coverage Pipeline", () => {

  test("Level 3: Full map view shows coverage footprints and tactical gaps", async ({ page }) => {
    await gotoApp(page);

    // 1. Switch to Coverage Dashboard (Level 3 Strategic View)
    const dashSelector = page.locator('[data-testid="app-header"] [data-testid="dashboard-selector"] select');
    await dashSelector.selectOption({ value: "coverage" });

    // 2. Verify Coverage Map Dashboard is active
    const coverageDashboard = page.locator('[data-testid="coverage-map-dashboard"]');
    await expect(coverageDashboard).toBeVisible({ timeout: 15_000 });

    // 3. Verify header text (replaces "Coverage Optimization" check)
    const sensorHeader = page.locator("text=Sensor Coverage Overlay");
    await expect(sensorHeader).toBeVisible({ timeout: 10_000 });

    // 4. Verify fleet list visible
    const fleetList = page.locator('[data-testid="sensor-fleet-list"]');
    await expect(fleetList).toBeVisible({ timeout: 5_000 });

    // 5. Verify coverage area map visible
    const coverageMap = page.locator('[data-testid="coverage-area-map"]');
    await expect(coverageMap).toBeVisible({ timeout: 5_000 });

    await page.screenshot({ path: "e2e/snapshots/level3-full-coverage-map.png" });
  });

  test("Level 2: Diagnostic view shows sensor-specific mini-map", async ({ page }) => {
    await gotoApp(page);

    // Wait for cards and click one
    await page.waitForSelector(".sensor-card-hover", { timeout: 15_000 });
    const firstCard = page.locator(".sensor-card-hover").first();
    await firstCard.click();

    // Verify Mini-Map presence in Diagnostic view
    const miniMap = page.locator('[data-testid="mini-coverage-map"]');
    await expect(miniMap).toBeVisible({ timeout: 10_000 });

    await page.screenshot({ path: "e2e/snapshots/level2-diagnostic-minimap.png" });
  });

  test("Level 1: Dashboard overview map shows coverage patterns", async ({ page }) => {
    await gotoApp(page);

    // Wait for health dashboard to load
    await page.waitForSelector("text=Sensor Health Monitor", {
      timeout: 15_000,
    });

    // Verify that sensor overview map is visible in the sidebar
    const overviewMap = page.locator('[data-testid="sensor-overview-map"]');
    await expect(overviewMap).toBeVisible({ timeout: 10_000 });

    await page.screenshot({ path: "e2e/snapshots/level1-dashboard-overview-map.png" });
  });
});
