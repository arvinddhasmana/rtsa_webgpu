// CLASSIFICATION: UNCLASSIFIED
// web-cop-gpu/e2e/commander-dashboards.spec.ts

import { expect, test } from "@playwright/test";
import { gotoApp } from "./helpers";

test.describe("Operations Commander dashboards", () => {
  test("commander role can navigate Fusion, Multi-Domain, and Operator UI dashboards", async ({
    page,
  }) => {
    await gotoApp(page);
    await page
      .waitForLoadState("networkidle", { timeout: 20_000 })
      .catch(() => {});

    const roleSelect = page.locator("#role-selector");
    const dashboardSelect = page.locator("#dashboard-selector");

    await roleSelect.selectOption("operations_commander");
    await expect(dashboardSelect).toHaveValue("commander");

    const commanderOptions = await dashboardSelect
      .locator("option")
      .allTextContents();
    expect(commanderOptions).toEqual(["Fusion", "Multi-Domain", "Operator UI"]);

    await expect(
      page.locator('[data-testid="commander-fusion-dashboard"]'),
    ).toBeVisible();
    await expect(
      page.locator('[data-testid="commander-observation-layer-mount"]'),
    ).toBeVisible();
    await expect(
      page.locator('[data-testid="commander-fused-layer-mount"]'),
    ).toBeVisible();

    await dashboardSelect.selectOption("coverage");
    await expect(
      page.locator('[data-testid="commander-multi-domain-dashboard"]'),
    ).toBeVisible();
    await expect(
      page.locator('[data-testid="commander-multidomain-kpi-overlay"]'),
    ).toBeVisible();
    await expect(
      page.locator('[data-testid="commander-multidomain-layer-toggles"]'),
    ).toBeVisible();

    await dashboardSelect.selectOption("analytics");
    await expect(
      page.locator('[data-testid="commander-operator-ui-dashboard"]'),
    ).toBeVisible();
    await expect(
      page.locator('[data-testid="commander-operator-alert-column"]'),
    ).toBeVisible();
    await expect(
      page.locator('[data-testid="commander-operator-detail-pane"]'),
    ).toBeVisible();
    await expect(
      page.locator('[data-testid="commander-operator-timeline-pane"]'),
    ).toBeVisible();
  });

  test("sensor operator dashboards remain unchanged", async ({ page }) => {
    await gotoApp(page);
    await page
      .waitForLoadState("networkidle", { timeout: 20_000 })
      .catch(() => {});

    const roleSelect = page.locator("#role-selector");
    const dashboardSelect = page.locator("#dashboard-selector");

    await roleSelect.selectOption("sensor_operator");

    const sensorOptions = await dashboardSelect
      .locator("option")
      .allTextContents();
    expect(sensorOptions).toEqual(["Sensor Health", "Coverage"]);

    await dashboardSelect.selectOption("health");
    await expect(page.locator("#sensor-health-dashboard-root")).toBeVisible();

    await dashboardSelect.selectOption("coverage");
    await expect(
      page.locator('[data-testid="coverage-map-dashboard"]'),
    ).toBeVisible();
  });
});
