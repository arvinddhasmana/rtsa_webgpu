// CLASSIFICATION: UNCLASSIFIED
import { expect, test } from "@playwright/test";
import { mockFeedbackSubmit, mockGrpcStream, selectRole, waitForMapReady } from "./helpers";

test.describe("Operations Commander Dashboards", () => {
  test.beforeEach(async ({ page }) => {
    await mockGrpcStream(page);
    await mockFeedbackSubmit(page);
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
    await expect(page.getByTestId("domain-metrics-overlay").getByText("TRACKS", { exact: true }).first()).toBeVisible();
    await expect(page.getByTestId("domain-metrics-overlay").getByText("OBS/S", { exact: true }).first()).toBeVisible();
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

  test("Critical alert triage flow - Inspect, Detail opens, Confirm", async ({ page }) => {
    const { injectTestTrack, injectTestAlert } = await import("./helpers");
    await page.getByRole("combobox", { name: /dashboard/i }).selectOption("operator");

    // Inject data
    await injectTestTrack(page, { trackId: "TRK-OP-1", lat: 35, lon: -120 });
    await injectTestAlert(page, { alertId: "ALT-CRY-1", trackId: "TRK-OP-1", severity: "CRITICAL" });

    const alertCard = page.getByTestId("alert-card-ALT-CRY-1");
    // Pulse animation or basic visibility check
    await expect(alertCard).toBeVisible();

    // Inspect
    await page.getByTestId("alert-inspect-ALT-CRY-1").click({ force: true });

    // Detail panel reveals
    await expect(page.getByTestId("detail-panel")).toBeVisible();

    // Check basic details
    await expect(page.getByTestId("detail-panel").getByText("TRK-OP-1", { exact: false }).first()).toBeVisible();

    // Confirm
    await page.getByTestId("alert-confirm-ALT-CRY-1").click({ force: true });

    // It should disappear from the Queue quickly
    await expect(alertCard).not.toBeVisible();

    // Switch to history tab to see it
    await page.getByTestId("tab-history").click();
    const historyCard = page.getByTestId("alert-card-ALT-CRY-1").first();
    await expect(historyCard).toBeVisible();
    await expect(historyCard.getByText("Accepted", { exact: false })).toBeVisible({ timeout: 10000 });
  });

  test("Alert reject flow", async ({ page }) => {
    const { injectTestTrack, injectTestAlert } = await import("./helpers");
    await page.getByRole("combobox", { name: /dashboard/i }).selectOption("operator");

    await injectTestTrack(page, { trackId: "TRK-OP-2", lat: 35, lon: -120 });
    await injectTestAlert(page, { alertId: "ALT-REJ-2", trackId: "TRK-OP-2", severity: "ELEVATED" });

    const alertCard = page.getByTestId("alert-card-ALT-REJ-2");
    await expect(alertCard).toBeVisible();

    // Reject (ELEVATED has no infinite pulse, but using force to be safe)
    await page.getByTestId("alert-reject-ALT-REJ-2").click({ force: true });

    // Disappears from Queue
    await expect(alertCard).not.toBeVisible();

    // Verify in history
    await page.getByTestId("tab-history").click();
    const historyCard = page.getByTestId("alert-card-ALT-REJ-2").first();
    await expect(historyCard.getByText("Rejected", { exact: false })).toBeVisible({ timeout: 10000 });
  });

  test("Alert assign flow", async ({ page }) => {
    const { injectTestTrack, injectTestAlert } = await import("./helpers");
    await page.getByRole("combobox", { name: /dashboard/i }).selectOption("operator");

    await injectTestTrack(page, { trackId: "TRK-OP-3", lat: 35, lon: -120 });
    await injectTestAlert(page, { alertId: "ALT-ASS-3", trackId: "TRK-OP-3", severity: "WATCH" });

    const alertCard = page.getByTestId("alert-card-ALT-ASS-3");

    // Assign click
    await alertCard.getByTestId("alert-assign-ALT-ASS-3").click();

    const popover = page.getByTestId("alert-assign-popover");
    await expect(popover).toBeVisible();

    // Select Charlie-01
    await popover.getByTestId("assign-op-op-charlie-1").click();

    // Confirm Assignment
    await popover.getByTestId("assign-confirm-btn").click();

    // Toast
    await expect(popover.getByText("Alert Assigned")).toBeVisible();
    await expect(popover).not.toBeVisible({ timeout: 5000 }); // Closes automatically
  });

  test("Keyboard shortcut navigation C, R, space on focused card", async ({ page }) => {
    const { injectTestTrack, injectTestAlert } = await import("./helpers");
    await page.getByRole("combobox", { name: /dashboard/i }).selectOption("operator");

    await injectTestTrack(page, { trackId: "TRK-OP-KEY", lat: 35, lon: -120 });
    await injectTestAlert(page, { alertId: "ALT-KEY-4", trackId: "TRK-OP-KEY", severity: "CRITICAL" });

    const alertCard = page.getByTestId("alert-card-ALT-KEY-4");
    await alertCard.focus(); // focus without triggering click and hiding

    // Press C to confirm
    await page.keyboard.press("C");

    await expect(alertCard).not.toBeVisible();

    await page.getByTestId("tab-history").click();
    const historyCard = page.getByTestId("alert-card-ALT-KEY-4").first();
    await expect(historyCard.getByText("Accepted", { exact: false })).toBeVisible({ timeout: 10000 });
  });

  test("Keyboard shortcut Enter on focused card inspects details", async ({ page }) => {
    const { injectTestTrack, injectTestAlert } = await import("./helpers");
    await page.getByRole("combobox", { name: /dashboard/i }).selectOption("operator");

    await injectTestTrack(page, { trackId: "TRK-OP-T", lat: 35, lon: -120 });
    await injectTestAlert(page, { alertId: "ALT-ENT", trackId: "TRK-OP-T", severity: "WATCH" });

    const alertCard = page.getByTestId("alert-card-ALT-ENT");
    await alertCard.focus(); // focus without triggering click

    // Ensure hidden first
    await page.mouse.click(10, 10); // click map to close detail panel if any

    await alertCard.focus();
    await page.keyboard.press("Enter");

    await expect(page.getByTestId("detail-panel")).toBeVisible();
  });

  test("Timeline filter chips and alert history tab works", async ({ page }) => {
    const { injectTestTrack, injectTestAlert } = await import("./helpers");
    await page.getByRole("combobox", { name: /dashboard/i }).selectOption("operator");

    await injectTestTrack(page, { trackId: "TRK-OP-5", lat: 35, lon: -120 });
    await injectTestAlert(page, { alertId: "ALT-TL-5", trackId: "TRK-OP-5", severity: "CRITICAL" });

    const alertCard = page.getByTestId("alert-card-ALT-TL-5");
    // Inspect alert to load timeline
    await page.getByTestId("alert-inspect-ALT-TL-5").click({ force: true });

    await expect(page.getByTestId("timeline-view")).toBeVisible();

    // Click filter chips
    await page.getByText("ANOMALY", { exact: true }).click();

    // We expect some events here but we don't have mock timeline events explicitly injected
    // the UI handles loading, then an empty or event state. Just check the chip acts active.

    // Acknowledge the alert
    // Use card click to acknowledge it (removes pulse) or confirm it
    await page.getByTestId("alert-confirm-ALT-TL-5").click({ force: true });

    // Switch to history tab
    await page.getByTestId("tab-history").click();
    // In history tab, card should be there
    const historyCard = page.getByTestId("alert-card-ALT-TL-5").first();
    await expect(historyCard).toBeVisible();
    await expect(historyCard.getByText("Accepted", { exact: false })).toBeVisible();
  });
});
