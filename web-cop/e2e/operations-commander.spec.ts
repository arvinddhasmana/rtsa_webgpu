// CLASSIFICATION: UNCLASSIFIED
import { expect, test } from "@playwright/test";
import { mockGrpcStream, selectRole, waitForMapReady } from "./helpers";

test.describe("Operations Commander Dashboards", () => {
  test.beforeEach(async ({ page }) => {
    await mockGrpcStream(page);
    await page.goto("/");
    await waitForMapReady(page);

    // Role selection
    await selectRole(page, "commander");
  });

  test("Fusion Dashboard is default and has side panel", async ({ page }) => {
    // Should default to fusion
    await expect(page.getByTestId("fusion-dashboard")).toBeVisible();
    await expect(page.getByTestId("fusion-side-panel")).toBeVisible();

    // Check KPI boxes exist
    await expect(page.getByText("⚡ FUSION ENGINE")).toBeVisible();
    await expect(page.getByText("Active Tracks")).toBeVisible();
    await expect(page.getByText("Raw Obs/10s", { exact: false })).toBeVisible();

    // Check confidence histogram rendering
    await expect(page.getByTestId("confidence-histogram")).toBeVisible();
  });

  test("Multi-Domain Dashboard has metric overlay and alert strip", async ({ page }) => {
    // Switch to Multi-Domain
    await page.getByRole("combobox", { name: /dashboard/i }).selectOption("multi-domain");
    await expect(page.getByTestId("multi-domain-dashboard")).toBeVisible();

    // Check Domain Metrics overlay
    await expect(page.getByTestId("domain-metrics-overlay")).toBeVisible();
    await expect(page.getByTestId("domain-metrics-overlay").getByText("TRACKS", { exact: true })).toBeVisible();
    await expect(page.getByTestId("domain-metrics-overlay").getByText("MULTI-DOMAIN", { exact: true })).toBeVisible();

    // Check Alert Strip
    await expect(page.getByTestId("multi-domain-dashboard").getByText("🚨 ALERTS", { exact: true })).toBeVisible();

    // Expand alert strip
    await page.getByTestId("multi-domain-dashboard").getByText("🚨 ALERTS", { exact: true }).click();
    await expect(page.getByTestId("alert-panel")).toBeVisible();
  });

  test("Operator Dashboard has Timeline and Scrubber button", async ({ page }) => {
    // Switch to Operator
    await page.getByRole("combobox", { name: /dashboard/i }).selectOption("operator");
    await expect(page.getByTestId("operator-dashboard")).toBeVisible();

    // Check Entity Timeline left pane
    await expect(page.getByTestId("timeline-empty")).toBeVisible(); // Empty initially

    // Open Replay / Scrubber
    await page.getByRole("button", { name: /Replay Mode/i }).click();
    await expect(page.getByTestId("timeline-scrubber")).toBeVisible();

    // Play/Pause button and Speed
    await expect(page.getByTestId("scrubber-play-pause")).toBeVisible();
    await expect(page.getByTestId("scrubber-speed")).toBeVisible();
  });
});
