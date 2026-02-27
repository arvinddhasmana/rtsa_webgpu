// CLASSIFICATION: UNCLASSIFIED
// web-cop/e2e/search.spec.ts

import { expect, test } from "@playwright/test";
import { injectTestTrackWithClassification, waitForMapReady } from "./helpers";

test.describe("Search & Entity Lookup", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/");
    await waitForMapReady(page);

    // Inject a test track to search for
    await injectTestTrackWithClassification(page, {
      trackId: "MMSI-987654321",
      lat: 40.0,
      lon: -70.0,
      entityType: "CARGO",
      hostileClass: "NEUTRAL",
      classification: "UNCLASSIFIED"
    });
  });

  test("Ctrl+F opens search overlay", async ({ page }) => {
    // Try both to be safe
    await page.keyboard.press("Control+f");
    await page.keyboard.press("Meta+f");
    await expect(page.getByTestId("search-overlay")).toBeVisible();
    await expect(page.getByTestId("search-input")).toBeFocused();
  });

  test("typing in search input filters injected tracks", async ({ page }) => {
    await page.keyboard.press("Control+f");
    await page.keyboard.press("Meta+f");
    const searchInput = page.getByTestId("search-input");

    await searchInput.fill("MMSI");
    await expect(page.getByTestId("search-result-MMSI-987654321")).toBeVisible();

    await searchInput.fill("UNKNOWN_ID");
    await expect(page.getByText("No results found.")).toBeVisible();
  });

  test("clicking a search result opens detail panel", async ({ page }) => {
    await page.keyboard.press("Control+f");
    await page.keyboard.press("Meta+f");
    await page.getByTestId("search-input").fill("MMSI");

    const result = page.getByTestId("search-result-MMSI-987654321");
    // Ensure the result is visible before clicking
    await expect(result).toBeVisible();
    await result.click();

    // Search overlay should close
    await expect(page.getByTestId("search-overlay")).toBeHidden();

    // Detail panel should open and show the track ID - check the explicit testId we added earlier or just text if testId isn't on it
    await expect(page.getByText("MMSI-987654321")).toBeVisible();
  });

  test("Escape closes the search overlay", async ({ page }) => {
    await page.keyboard.press("Control+f");
    await page.keyboard.press("Meta+f");
    await expect(page.getByTestId("search-overlay")).toBeVisible();

    await page.keyboard.press("Escape");
    await expect(page.getByTestId("search-overlay")).toBeHidden();
  });
});
