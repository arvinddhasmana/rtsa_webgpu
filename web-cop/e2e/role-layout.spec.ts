// CLASSIFICATION: UNCLASSIFIED
// web-cop/e2e/role-layout.spec.ts

import { expect, test } from "@playwright/test";
import { waitForMapReady } from "./helpers";

test.describe("Role-Based Dashboards & Navigation", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/");
    await waitForMapReady(page);
  });

  test("default layout shows map and alert panel (Commander view)", async ({ page }) => {
    // Role selector should be commander
    await expect(page.getByTestId("role-selector")).toHaveValue("commander");

    // Map should be visible
    await expect(page.locator("canvas").first()).toBeVisible();

    // Alert panel should be visible
    await expect(page.getByLabel("Alert Panel").first()).toBeVisible();

    // Audit view should not be visible
    await expect(page.getByTestId("audit-view")).toBeHidden();
  });

  test("selecting 'Security Officer' hides the map and shows audit/feedback queue", async ({ page }) => {
    await page.getByTestId("role-selector").selectOption("security");

    // Map should be hidden
    await expect(page.locator("canvas").first()).toBeHidden();

    // Alert panel and Audit view should be visible side by side
    await expect(page.getByTestId("alert-panel").first()).toBeVisible();
    await expect(page.getByTestId("audit-view").first()).toBeVisible();
  });

  test("selecting 'Intelligence Analyst' shows expanded forensics panel", async ({ page }) => {
    await page.getByTestId("role-selector").selectOption("analyst");

    // Map should be visible
    await expect(page.locator("canvas").first()).toBeVisible();

    // Forensics panel should be visible
    await expect(page.getByTestId("forensics-panel").first()).toBeVisible();
  });
});
