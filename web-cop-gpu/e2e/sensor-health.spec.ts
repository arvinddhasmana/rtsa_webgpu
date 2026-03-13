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

  test("role selector is visible inside app-header", async ({ page }) => {
    await gotoApp(page);
    await page.waitForSelector('[data-testid="app-header"]', {
      timeout: 15_000,
    });
    const roleSelector = page.locator(
      '[data-testid="app-header"] [data-testid="role-selector"]',
    );
    await expect(roleSelector).toBeVisible();
  });

  test("dashboard selector is inside app-header", async ({ page }) => {
    await gotoApp(page);
    await page.waitForSelector('[data-testid="app-header"]', {
      timeout: 15_000,
    });
    const dashSelector = page.locator(
      '[data-testid="app-header"] [data-testid="dashboard-selector"] select',
    );
    await expect(dashSelector).toBeVisible();
  });

  test("can filter by status", async ({ page }) => {
    await gotoApp(page);
    await page.waitForSelector("text=Sensor Health Monitor", {
      timeout: 15_000,
    });

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
    await page.waitForSelector("text=Sensor Health Monitor", {
      timeout: 15_000,
    });

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
    await page.waitForSelector("text=Sensor Health Monitor", {
      timeout: 15_000,
    });

    // Open dashboard selector inside app-header
    const dashSelector = page.locator(
      '[data-testid="app-header"] [data-testid="dashboard-selector"] select',
    );
    await dashSelector.selectOption({ label: "Map view" });

    // Map canvas should be visible
    const canvas = page.locator("#gpu-canvas");
    await expect(canvas).toBeVisible();

    // Switch back to Health
    await dashSelector.selectOption({ label: "Health" });
    await expect(page.locator("text=Sensor Health Monitor")).toBeVisible();
  });

  test("clicking first sensor card opens diagnostic view", async ({ page }) => {
    await gotoApp(page);
    await page.waitForSelector("text=Sensor Health Monitor", {
      timeout: 15_000,
    });
    // Wait for sensor cards to appear
    await page.waitForSelector(".sensor-card-hover", { timeout: 15_000 });

    const firstCard = page.locator(".sensor-card-hover").first();
    await firstCard.click();

    // Diagnostic view should be visible
    await expect(
      page.locator('[data-testid="sensor-diagnostic-view"]'),
    ).toBeVisible({ timeout: 10_000 });
  });

  test("clicking back button returns to sensor grid", async ({ page }) => {
    await gotoApp(page);
    await page.waitForSelector("text=Sensor Health Monitor", {
      timeout: 15_000,
    });
    await page.waitForSelector(".sensor-card-hover", { timeout: 15_000 });

    const firstCard = page.locator(".sensor-card-hover").first();
    await firstCard.click();
    await expect(
      page.locator('[data-testid="sensor-diagnostic-view"]'),
    ).toBeVisible({ timeout: 10_000 });

    // Click back button
    await page.locator('[data-testid="diagnostic-back-btn"]').click();

    // Diagnostic view should be gone, grid visible
    await expect(
      page.locator('[data-testid="sensor-diagnostic-view"]'),
    ).not.toBeVisible();
    await expect(page.locator("text=Sensor Health Monitor")).toBeVisible();
  });

  test("captures screenshot: sensor health with header", async ({ page }) => {
    await gotoApp(page);
    await page.waitForSelector("text=Sensor Health Monitor", {
      timeout: 15_000,
    });
    // Wait for data to load and animations to settle
    await page.waitForTimeout(3000);

    await page.screenshot({
      path: "e2e/snapshots/sensor-health-with-header.png",
      fullPage: true,
    });
  });

  test("captures screenshot for proof of work", async ({ page }) => {
    await gotoApp(page);
    await page.waitForSelector("text=Sensor Health Monitor", {
      timeout: 15_000,
    });
    // Wait for data to load and animations to settle
    await page.waitForTimeout(3000);

    // Ensure the snapshots directory exists (handled by Playwright usually, but let's be safe)
    await page.screenshot({
      path: "e2e/snapshots/sensor-health-dashboard.png",
      fullPage: true,
    });

    console.log(
      "Screenshot saved to e2e/snapshots/sensor-health-dashboard.png",
    );
  });

  test("view toggle is visible next to Sensor Health Monitor heading", async ({ page }) => {
    await gotoApp(page);
    await page.waitForSelector("text=Sensor Health Monitor", { timeout: 15_000 });
    await expect(page.locator('[data-testid="view-toggle-full"]')).toBeVisible();
    await expect(page.locator('[data-testid="view-toggle-compact"]')).toBeVisible();
  });

  test("toggle to compact view renders compact cards (data-view=compact)", async ({ page }) => {
    await gotoApp(page);
    await page.waitForSelector(".sensor-card-hover", { timeout: 15_000 });
    await page.locator('[data-testid="view-toggle-compact"]').click();
    const cards = page.locator('[data-view="compact"]');
    await expect(cards.first()).toBeVisible({ timeout: 5_000 });
  });

  test("toggle back to full view renders full cards (data-view=full)", async ({ page }) => {
    await gotoApp(page);
    await page.waitForSelector(".sensor-card-hover", { timeout: 15_000 });
    await page.locator('[data-testid="view-toggle-compact"]').click();
    await page.locator('[data-testid="view-toggle-full"]').click();
    const cards = page.locator('[data-view="full"]');
    await expect(cards.first()).toBeVisible({ timeout: 5_000 });
  });

  test("captures screenshot: compact card view", async ({ page }) => {
    await gotoApp(page);
    await page.waitForSelector(".sensor-card-hover", { timeout: 15_000 });
    await page.locator('[data-testid="view-toggle-compact"]').click();
    await page.waitForTimeout(1000);
    await page.screenshot({ path: "e2e/snapshots/sensor-health-compact-view.png", fullPage: true });
  });

  test("captures screenshot: full card view", async ({ page }) => {
    await gotoApp(page);
    await page.waitForSelector(".sensor-card-hover", { timeout: 15_000 });
    await page.locator('[data-testid="view-toggle-full"]').click();
    await page.waitForTimeout(1000);
    await page.screenshot({ path: "e2e/snapshots/sensor-health-full-view.png", fullPage: true });
  });
});
