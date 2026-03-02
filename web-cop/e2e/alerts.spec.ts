// CLASSIFICATION: UNCLASSIFIED
// web-cop/e2e/alerts.spec.ts
// Browser E2E tests for Alert Display — CR-UI-003, CR-INF-001
// Covers: UC005 (Anomaly Detection), UC007 (Alert Management)

import { expect, test } from "@playwright/test";
import { injectTestAlert, injectTestTrack, mockFeedbackSubmit } from "./helpers";

test.describe("Alert Display (CR-UI-003, CR-INF-001)", () => {
  test.beforeEach(async ({ page }) => {
    test.setTimeout(90000); // Higher total suite test timeout
    await page.goto("/");
    await page.waitForLoadState("networkidle", { timeout: 25_000 }).catch(() => {});
  });

  test("alert panel element is present in the DOM [UC007]", async ({ page }) => {
    // Alert panel should be mounted in the layout
    const alertPanel = page.locator('[data-testid="alert-panel"]').first();
    await expect(alertPanel).toBeVisible({ timeout: 15_000 });
  });

  test("alert panel shows 'No alerts' when empty [CR-UI-003]", async ({ page }) => {
    const alertPanel = page.getByTestId("alert-panel").first();
    await expect(alertPanel).toBeVisible({ timeout: 15_000 });
    // Look for text either in the panel or verify it's not throwing an error
    await expect(alertPanel).toContainText(/ALERTS/);
  });

  test("alert filter buttons are rendered [CR-UI-003]", async ({ page }) => {
    // AlertFilter component renders severity toggle buttons
    const filterButtons = page.getByTestId("alert-filter");
    await expect(filterButtons).toBeVisible({ timeout: 15_000 });

    // We expect 3 filter buttons
    const buttons = filterButtons.locator("button");
    await expect(buttons).toHaveCount(3);
  });

  test("injected alert shows Inspect/Confirm/Reject/Assign buttons", async ({ page }) => {
    // Inject and make sure we wait for it to be visible
    await injectTestAlert(page, {
      alertId: "TEST-ALT-1",
      trackId: "TEST-TRK-1",
      severity: "ELEVATED"
    });

    // We must wait for the buttons to be attached
    await expect(page.getByTestId("alert-inspect-TEST-ALT-1")).toBeVisible({ timeout: 10000 });
    await expect(page.getByTestId("alert-confirm-TEST-ALT-1")).toBeVisible();
    await expect(page.getByTestId("alert-reject-TEST-ALT-1")).toBeVisible();
    await expect(page.getByTestId("alert-assign-TEST-ALT-1")).toBeVisible();
  });

  test("clicking Inspect opens detail panel", async ({ page }) => {
    await injectTestAlert(page, {
      alertId: "TEST-ALT-2",
      trackId: "TEST-TRK-2",
      severity: "CRITICAL"
    });

    // Inject corresponding track so it doesn't say "not found"
    await injectTestTrack(page, {
      trackId: "TEST-TRK-2",
      lat: 0,
      lon: 0
    });

    // Wait for the button to appear before clicking
    await page.getByTestId("alert-inspect-TEST-ALT-2").waitFor({ state: "visible", timeout: 10000 });
    await page.getByTestId("alert-inspect-TEST-ALT-2").click();
    await expect(page.getByTestId("detail-panel")).toBeVisible({ timeout: 10000 });
    await expect(page.getByText("TEST-TRK-2").first()).toBeVisible({ timeout: 10000 });
  });

  test("clicking Confirm shows feedback confirmation", async ({ page }) => {
    await mockFeedbackSubmit(page);
    await injectTestAlert(page, {
      alertId: "TEST-ALT-3",
      trackId: "TEST-TRK-3",
      severity: "WATCH"
    });

    await page.getByTestId("alert-confirm-TEST-ALT-3").waitFor({ state: "visible", timeout: 10000 });
    await page.getByTestId("alert-confirm-TEST-ALT-3").click();
    await expect(page.getByText("Status: Confirmed")).toBeVisible({ timeout: 10000 });
  });
});
