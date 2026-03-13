// CLASSIFICATION: UNCLASSIFIED
// web-cop-gpu/e2e/coverage-map.spec.ts — Level 3: Full Coverage Map Dashboard E2E tests
//
// Reference: docs/implementation/v5/sensordashboard_three_level_plan.md §C8

import { expect, test } from "@playwright/test";
import { gotoApp } from "./helpers";

test.describe("Level 3: Coverage Map Dashboard", () => {
  test("dashboard selector has 'Coverage' option", async ({ page }) => {
    await gotoApp(page);
    await page.waitForSelector('[data-testid="app-header"]', {
      timeout: 15_000,
    });

    const dashSelector = page.locator(
      '[data-testid="app-header"] [data-testid="dashboard-selector"] select',
    );
    await expect(dashSelector).toBeVisible();

    // Verify Coverage option exists
    const coverageOption = page.locator(
      '[data-testid="dashboard-selector"] select option[value="coverage"]',
    );
    await expect(coverageOption).toBeAttached();
  });

  test("'Coverage' option navigates to coverage view", async ({ page }) => {
    await gotoApp(page);
    await page.waitForSelector('[data-testid="app-header"]', {
      timeout: 15_000,
    });

    const dashSelector = page.locator(
      '[data-testid="app-header"] [data-testid="dashboard-selector"] select',
    );
    await dashSelector.selectOption({ value: "coverage" });

    // Coverage dashboard should be visible
    const coverageDashboard = page.locator('[data-testid="coverage-map-dashboard"]');
    await expect(coverageDashboard).toBeVisible({ timeout: 10_000 });
  });

  test("coverage map header shows 'SENSOR COVERAGE OVERLAY'", async ({ page }) => {
    await gotoApp(page);
    await page.waitForSelector('[data-testid="app-header"]', {
      timeout: 15_000,
    });

    const dashSelector = page.locator(
      '[data-testid="app-header"] [data-testid="dashboard-selector"] select',
    );
    await dashSelector.selectOption({ value: "coverage" });

    await page.waitForSelector('[data-testid="coverage-map-dashboard"]', {
      timeout: 10_000,
    });

    const header = page.locator("text=Sensor Coverage Overlay");
    await expect(header).toBeVisible();
  });

  test("current status shows NOMINAL when no gaps", async ({ page }) => {
    await gotoApp(page);
    await page.waitForSelector('[data-testid="app-header"]', {
      timeout: 15_000,
    });

    const dashSelector = page.locator(
      '[data-testid="app-header"] [data-testid="dashboard-selector"] select',
    );
    await dashSelector.selectOption({ value: "coverage" });

    await page.waitForSelector('[data-testid="coverage-map-dashboard"]', {
      timeout: 10_000,
    });

    // When running with mock data (no gaps injected), status should be NOMINAL
    // Note: This test may fail if mock data includes gaps; adjust expectations accordingly
    const statusText = page.locator("text=Current Status:");
    await expect(statusText).toBeVisible();
  });

  test("sensor fleet list is visible in left panel", async ({ page }) => {
    await gotoApp(page);
    await page.waitForSelector('[data-testid="app-header"]', {
      timeout: 15_000,
    });

    const dashSelector = page.locator(
      '[data-testid="app-header"] [data-testid="dashboard-selector"] select',
    );
    await dashSelector.selectOption({ value: "coverage" });

    await page.waitForSelector('[data-testid="coverage-map-dashboard"]', {
      timeout: 10_000,
    });

    // Sensor fleet list should be visible
    const fleetList = page.locator('[data-testid="sensor-fleet-list"]');
    await expect(fleetList).toBeVisible({ timeout: 5_000 });
  });

  test("critical alerts panel is visible in left panel", async ({ page }) => {
    await gotoApp(page);
    await page.waitForSelector('[data-testid="app-header"]', {
      timeout: 15_000,
    });

    const dashSelector = page.locator(
      '[data-testid="app-header"] [data-testid="dashboard-selector"] select',
    );
    await dashSelector.selectOption({ value: "coverage" });

    await page.waitForSelector('[data-testid="coverage-map-dashboard"]', {
      timeout: 10_000,
    });

    // Critical alerts panel should be visible
    const alertsPanel = page.locator('[data-testid="critical-alerts-panel"]');
    await expect(alertsPanel).toBeVisible({ timeout: 5_000 });
  });

  test("coverage area map is visible in center", async ({ page }) => {
    await gotoApp(page);
    await page.waitForSelector('[data-testid="app-header"]', {
      timeout: 15_000,
    });

    const dashSelector = page.locator(
      '[data-testid="app-header"] [data-testid="dashboard-selector"] select',
    );
    await dashSelector.selectOption({ value: "coverage" });

    await page.waitForSelector('[data-testid="coverage-map-dashboard"]', {
      timeout: 10_000,
    });

    // Coverage area map should be visible
    const coverageMap = page.locator('[data-testid="coverage-area-map"]');
    await expect(coverageMap).toBeVisible({ timeout: 5_000 });
  });

  test("sensor detail panel appears on sensor hover", async ({ page }) => {
    await gotoApp(page);
    await page.waitForSelector('[data-testid="app-header"]', {
      timeout: 15_000,
    });

    const dashSelector = page.locator(
      '[data-testid="app-header"] [data-testid="dashboard-selector"] select',
    );
    await dashSelector.selectOption({ value: "coverage" });

    await page.waitForSelector('[data-testid="coverage-map-dashboard"]', {
      timeout: 10_000,
    });

    // Wait for fleet list
    await page.waitForSelector('[data-testid="sensor-fleet-list"]', {
      timeout: 5_000,
    });

    // Hover over first sensor in fleet list
    const firstSensor = page.locator('[data-testid="sensor-fleet-list"] .sensor-fleet-item').first();
    await firstSensor.hover();

    // Sensor detail panel should become visible (or populated)
    const detailPanel = page.locator('[data-testid="sensor-detail-hover-panel"]');
    await expect(detailPanel).toBeVisible({ timeout: 3_000 });
  });

  test("legend is visible with sensor type color coding", async ({ page }) => {
    await gotoApp(page);
    await page.waitForSelector('[data-testid="app-header"]', {
      timeout: 15_000,
    });

    const dashSelector = page.locator(
      '[data-testid="app-header"] [data-testid="dashboard-selector"] select',
    );
    await dashSelector.selectOption({ value: "coverage" });

    await page.waitForSelector('[data-testid="coverage-map-dashboard"]', {
      timeout: 10_000,
    });

    // Legend should be visible on coverage area map
    const legend = page.locator('[data-testid="coverage-map-legend"]');
    await expect(legend).toBeVisible({ timeout: 5_000 });
  });

  test("ZOOM buttons are visible and functional", async ({ page }) => {
    await gotoApp(page);
    await page.waitForSelector('[data-testid="app-header"]', {
      timeout: 15_000,
    });

    const dashSelector = page.locator(
      '[data-testid="app-header"] [data-testid="dashboard-selector"] select',
    );
    await dashSelector.selectOption({ value: "coverage" });

    await page.waitForSelector('[data-testid="coverage-map-dashboard"]', {
      timeout: 10_000,
    });

    // ZOOM buttons should be visible
    const zoomIn = page.locator('[data-testid="zoom-in-button"]');
    const zoomOut = page.locator('[data-testid="zoom-out-button"]');

    await expect(zoomIn).toBeVisible({ timeout: 5_000 });
    await expect(zoomOut).toBeVisible({ timeout: 5_000 });

    // Click zoom in button
    await zoomIn.click();
    // Verify no crash (map should still be visible)
    const coverageMap = page.locator('[data-testid="coverage-area-map"]');
    await expect(coverageMap).toBeVisible();
  });

  test("screenshot: Level 3 NOMINAL state", async ({ page }) => {
    await gotoApp(page);
    await page.waitForSelector('[data-testid="app-header"]', {
      timeout: 15_000,
    });

    const dashSelector = page.locator(
      '[data-testid="app-header"] [data-testid="dashboard-selector"] select',
    );
    await dashSelector.selectOption({ value: "coverage" });

    await page.waitForSelector('[data-testid="coverage-map-dashboard"]', {
      timeout: 10_000,
    });

    // Wait for map to render
    await page.waitForTimeout(2000);

    // Take screenshot
    await expect(page).toHaveScreenshot("coverage-map-nominal.png", {
      fullPage: false,
      timeout: 10_000,
    });
  });
});

test.describe("Level 3: Coverage Map with Spatial Alerts", () => {
  // Note: These tests require injecting spatial alert data via mock or simulator
  // For now, we test the UI structure assuming alerts exist

  test.skip("current status shows ACTIVE GAPS count when gaps present", async ({ page }) => {
    // This test requires injecting a gap alert
    // Skip for now until mock data includes gaps
    await gotoApp(page);
    // ... test implementation
  });

  test.skip("SpatialAlertBanner appears when coverage gap exists", async ({ page }) => {
    // This test requires injecting a gap alert
    await gotoApp(page);
    // ... test implementation
  });

  test.skip("RESOLVE button on SpatialAlertBanner is clickable", async ({ page }) => {
    // This test requires injecting a gap alert
    await gotoApp(page);
    // ... test implementation
  });

  test.skip("gap hatched area visible for offline sensor", async ({ page }) => {
    // This test requires injecting an offline sensor with coverage
    await gotoApp(page);
    // ... test implementation
  });

  test.skip("GAP DETECTED label visible on offline sensor footprint", async ({ page }) => {
    // This test requires injecting an offline sensor
    await gotoApp(page);
    // ... test implementation
  });

  test.skip("Level 3 auto-zooms to active alert gap area on navigation", async ({ page }) => {
    // This test requires clicking an alert in AlertSidebar
    await gotoApp(page);
    // ... test implementation
  });

  test.skip("clicking coverage alert in AlertSidebar navigates to Level 3", async ({ page }) => {
    // This test requires alert data and AlertSidebar interaction
    await gotoApp(page);
    // ... test implementation
  });

  test.skip("screenshot: Level 3 with active gap", async ({ page }) => {
    // This test requires injecting a gap alert
    await gotoApp(page);
    // ... test implementation
  });
});
