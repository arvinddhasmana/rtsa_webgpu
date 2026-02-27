// CLASSIFICATION: UNCLASSIFIED
// web-cop/e2e/toolbar.spec.ts

import { expect, test } from "@playwright/test";
import { waitForMapReady } from "./helpers";

test.describe("Toolbar Navigation", () => {
  test.beforeEach(async ({ page }) => {
    page.on("console", (msg) => console.log(`BROWSER CONSOLE: ${msg.text()}`));
    page.on("pageerror", (err) => console.log(`BROWSER ERROR: ${err.message}`));
    await page.goto("/");
    await waitForMapReady(page);
  });

  test("renders all toolbar buttons", async ({ page }) => {
    await expect(page.getByTestId("toolbar-map")).toBeVisible();
    await expect(page.getByTestId("toolbar-alerts")).toBeVisible();
    await expect(page.getByTestId("toolbar-history")).toBeVisible();
    await expect(page.getByTestId("toolbar-sensors")).toBeVisible();
    await expect(page.getByTestId("toolbar-nato")).toBeVisible();
    await expect(page.getByTestId("toolbar-audit")).toBeVisible();
  });

  test("role selector dropdown is present with 3 options", async ({ page }) => {
    const roleSelector = page.getByTestId("role-selector");
    await expect(roleSelector).toBeVisible();

    // Check initial selection
    await expect(roleSelector).toHaveValue("commander");

    // Check options
    const options = await roleSelector.locator("option").allTextContents();
    expect(options).toEqual([
      "Operations Commander",
      "Security Officer",
      "Intelligence Analyst"
    ]);
  });
});
