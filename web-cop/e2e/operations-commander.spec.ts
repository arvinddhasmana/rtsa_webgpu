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

  test("Fusion Dashboard has split-pane layout with tabs and KPIs", async ({ page }) => {
    // Should default to fusion
    await expect(page.getByTestId("fusion-dashboard")).toBeVisible();
    await expect(page.getByTestId("fusion-side-panel")).toBeVisible();

    // Check resize handle exists (Variant C split-pane)
    await expect(page.getByTestId("fusion-resize-handle")).toBeVisible();

    // Check tabbed navigation
    await expect(page.getByTestId("tab-track-grid")).toBeVisible();
    await expect(page.getByTestId("tab-alert-queue")).toBeVisible();

    // Check KPI tiles (design system)
    await expect(page.getByTestId("kpi-tracks")).toBeVisible();
    await expect(page.getByTestId("kpi-hostile")).toBeVisible();
    await expect(page.getByTestId("kpi-confidence")).toBeVisible();
    await expect(page.getByTestId("kpi-obs")).toBeVisible();

    // Check sortable track table
    await expect(page.getByTestId("fusion-track-table")).toBeVisible();
  });

  test("Fusion Dashboard tabs switch and Replay works", async ({ page }) => {
    await expect(page.getByTestId("fusion-dashboard")).toBeVisible();

    // Switch to Alert Queue tab
    await page.getByTestId("tab-alert-queue").click();
    await expect(page.getByTestId("fusion-alert-table")).toBeVisible();

    // Switch back to Track Grid tab
    await page.getByTestId("tab-track-grid").click();
    await expect(page.getByTestId("fusion-track-table")).toBeVisible();

    // Open Replay / Scrubber via tab bar button
    await page.getByRole("button", { name: /Replay/i }).click();
    await expect(page.getByTestId("timeline-scrubber")).toBeVisible();
    await expect(page.getByTestId("scrubber-play-pause")).toBeVisible();
    await expect(page.getByTestId("scrubber-speed")).toBeVisible();
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
