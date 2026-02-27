// CLASSIFICATION: UNCLASSIFIED
// web-cop/e2e/keyboard.spec.ts

import { expect, test } from "@playwright/test";
import { injectTestTrackWithClassification, waitForMapReady } from "./helpers";

test.describe("Keyboard Navigation & Accessibility", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/");
    await waitForMapReady(page);

    await injectTestTrackWithClassification(page, {
      trackId: "KBD-12345",
      lat: 45.0,
      lon: -60.0,
      entityType: "SURFACE",
      classification: "UNCLASSIFIED"
    });
  });

  test("Escape closes detail panel when open", async ({ page }) => {
    // Search to open detail panel
    await page.keyboard.press("Control+f");
    await page.getByTestId("search-input").fill("KBD");
    await page.getByTestId("search-result-KBD-12345").click();

    await expect(page.getByTestId("detail-panel")).toBeVisible();

    // Press Escape
    await page.keyboard.press("Escape");

    // Detail panel should be gone
    await expect(page.getByTestId("detail-panel")).toBeHidden();
  });

  test("H key toggles forensics panel", async ({ page }) => {
    await expect(page.getByTestId("forensics-panel")).toBeHidden();

    // Press H to open
    await page.keyboard.press("h");
    await expect(page.getByTestId("forensics-panel")).toBeVisible();

    // Press H to close
    await page.keyboard.press("h");
    await expect(page.getByTestId("forensics-panel")).toBeHidden();
  });
});
