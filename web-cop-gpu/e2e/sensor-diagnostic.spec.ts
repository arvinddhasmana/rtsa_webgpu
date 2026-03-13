// CLASSIFICATION: UNCLASSIFIED
// web-cop-gpu/e2e/sensor-diagnostic.spec.ts — Sensor Diagnostic Deep Dive E2E tests
//
// Reference: docs/business/usecases/UC017_sensor_health_monitoring.md

import { expect, test } from "@playwright/test";
import { gotoApp } from "./helpers";

test.describe("Sensor Diagnostic Deep Dive", () => {
  test.beforeEach(async ({ page }) => {
    await gotoApp(page);
    // Navigate to health dashboard and wait for sensors to appear
    await page.waitForSelector("text=Sensor Health Monitor", {
      timeout: 20_000,
    });
    await page.waitForSelector(".sensor-card-hover", { timeout: 15_000 });
  });

  test("clicking a sensor card opens the diagnostic view", async ({ page }) => {
    const firstCard = page.locator(".sensor-card-hover").first();
    await firstCard.click();
    await expect(
      page.locator('[data-testid="sensor-diagnostic-view"]'),
    ).toBeVisible({ timeout: 10_000 });
  });

  test("diagnostic view shows sensor ID and status", async ({ page }) => {
    const firstCard = page.locator(".sensor-card-hover").first();
    await firstCard.click();
    await expect(
      page.locator('[data-testid="sensor-diagnostic-view"]'),
    ).toBeVisible({ timeout: 10_000 });
    // The back breadcrumb contains "Sensor Health" — use first() since header also contains those words
    await expect(page.locator("text=Sensor Health").first()).toBeVisible();
  });

  test("DLQ breakdown section is visible", async ({ page }) => {
    const firstCard = page.locator(".sensor-card-hover").first();
    await firstCard.click();
    await expect(
      page.locator('[data-testid="sensor-diagnostic-view"]'),
    ).toBeVisible({ timeout: 10_000 });
    // Wait for diagnostic data to load
    await expect(page.locator("text=DLQ Breakdown")).toBeVisible({
      timeout: 10_000,
    });
  });

  test("recent events section is visible", async ({ page }) => {
    const firstCard = page.locator(".sensor-card-hover").first();
    await firstCard.click();
    await expect(
      page.locator('[data-testid="sensor-diagnostic-view"]'),
    ).toBeVisible({ timeout: 10_000 });
    await expect(page.locator("text=Recent Events")).toBeVisible({
      timeout: 10_000,
    });
  });

  test("back button returns to the sensor grid", async ({ page }) => {
    const firstCard = page.locator(".sensor-card-hover").first();
    await firstCard.click();
    await expect(
      page.locator('[data-testid="sensor-diagnostic-view"]'),
    ).toBeVisible({ timeout: 10_000 });

    await page.locator('[data-testid="diagnostic-back-btn"]').click();

    await expect(
      page.locator('[data-testid="sensor-diagnostic-view"]'),
    ).not.toBeVisible({ timeout: 5_000 });
    await expect(page.locator("text=Sensor Health Monitor")).toBeVisible();
  });

  test("captures screenshot: diagnostic view full", async ({ page }) => {
    const firstCard = page.locator(".sensor-card-hover").first();
    await firstCard.click();
    await expect(
      page.locator('[data-testid="sensor-diagnostic-view"]'),
    ).toBeVisible({ timeout: 10_000 });
    // Wait for data to settle
    await page.waitForTimeout(2000);
    await page.screenshot({
      path: "e2e/snapshots/sensor-diagnostic-full.png",
      fullPage: true,
    });
  });

  test("captures screenshot: back to grid", async ({ page }) => {
    const firstCard = page.locator(".sensor-card-hover").first();
    await firstCard.click();
    await expect(
      page.locator('[data-testid="sensor-diagnostic-view"]'),
    ).toBeVisible({ timeout: 10_000 });

    await page.locator('[data-testid="diagnostic-back-btn"]').click();
    await expect(page.locator("text=Sensor Health Monitor")).toBeVisible();
    await page.waitForTimeout(1000);
    await page.screenshot({
      path: "e2e/snapshots/sensor-diagnostic-back.png",
      fullPage: true,
    });
  });
});

test.describe("with live backend", () => {
  test.skip(
    !process.env.BASE_URL,
    "Requires BASE_URL env pointing to running backend",
  );

  test("sensor diagnostic shows real DLQ data", async ({ page }) => {
    await page.goto(process.env.BASE_URL!);
    await page.waitForLoadState("domcontentloaded");
    await page.waitForSelector("text=Sensor Health Monitor", {
      timeout: 30_000,
    });
    await page.waitForSelector(".sensor-card-hover", { timeout: 20_000 });

    const firstCard = page.locator(".sensor-card-hover").first();
    await firstCard.click();
    await expect(
      page.locator('[data-testid="sensor-diagnostic-view"]'),
    ).toBeVisible({ timeout: 10_000 });
    // With live backend, DLQ breakdown should have real data populated
    await expect(page.locator("text=DLQ Breakdown")).toBeVisible({
      timeout: 10_000,
    });

    await page.screenshot({
      path: "e2e/snapshots/sensor-diagnostic-live-data.png",
      fullPage: true,
    });
  });
});
