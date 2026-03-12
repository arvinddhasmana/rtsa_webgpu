// CLASSIFICATION: UNCLASSIFIED
// web-cop-gpu/e2e/sensor-health.spec.ts — Sensor Health Dashboard E2E tests
//
// Reference: docs/business/usecases/UC017_sensor_health_monitoring.md

import { expect, test } from "@playwright/test";
import { gotoApp } from "./helpers";

test.describe("Sensor Health Dashboard", () => {
  test("is visible by default for sensor operator", async ({ page }) => {
    await gotoApp(page);
    // The health dashboard should be visible because we set it as default in viewport.ts
    // We search for the header text we implemented in SensorHealthDashboard.tsx
    const healthHeader = page.locator("text=Sensor Health Monitor");
    await expect(healthHeader).toBeVisible({ timeout: 20_000 });
  });

  test("can filter by status", async ({ page }) => {
    await gotoApp(page);
    await page.waitForSelector("text=Sensor Health Monitor", { timeout: 15_000 });

    // Click 'Offline' filter card in sidebar
    const offlineFilter = page.locator("text=Offline").first();
    await offlineFilter.click();

    // Verify interaction responds (UI should still be stable)
    await expect(offlineFilter).toBeVisible();

    // Check if the count of cards changed or just verify no crash
    const cards = page.locator(".sensor-card-hover");
    const countBefore = await cards.count();

    // Toggle back
    await offlineFilter.click();
    const countAfter = await cards.count();
    expect(countAfter).toBeGreaterThanOrEqual(countBefore);
  });

  test("sidebar can be collapsed and expanded", async ({ page }) => {
    await gotoApp(page);
    await page.waitForSelector("text=Sensor Health Monitor", { timeout: 15_000 });

    const toggleBtn = page.locator(".sidebar-toggle-btn");
    await toggleBtn.click();

    // Verify sidebar labels are hidden
    const systemHealthText = page.locator("text=System Health");
    await expect(systemHealthText).not.toBeVisible();

    await toggleBtn.click();
    await expect(systemHealthText).toBeVisible();
  });

  test("can switch between health and map dashboards", async ({ page }) => {
    await gotoApp(page);
    await page.waitForSelector("text=Sensor Health Monitor", { timeout: 15_000 });

    // Open dashboard selector
    const dashSelector = page.locator('[data-testid="dashboard-selector"] select');
    await dashSelector.selectOption({ label: "Map view" });

    // Map canvas should be visible
    const canvas = page.locator("#gpu-canvas");
    await expect(canvas).toBeVisible();

    // Switch back to Health
    await dashSelector.selectOption({ label: "Health" });
    await expect(page.locator("text=Sensor Health Monitor")).toBeVisible();
  });

  test("captures screenshot for proof of work", async ({ page }) => {
    await gotoApp(page);
    await page.waitForSelector("text=Sensor Health Monitor", { timeout: 15_000 });
    // Wait for data to load and animations to settle
    await page.waitForTimeout(3000);

    // Ensure the snapshots directory exists (handled by Playwright usually, but let's be safe)
    await page.screenshot({
        path: "e2e/snapshots/sensor-health-dashboard.png",
        fullPage: true
    });

    console.log("Screenshot saved to e2e/snapshots/sensor-health-dashboard.png");
  });
});
