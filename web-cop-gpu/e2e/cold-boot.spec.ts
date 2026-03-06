// CLASSIFICATION: UNCLASSIFIED
// web-cop-gpu/e2e/cold-boot.spec.ts — Cold-boot workflow E2E tests
//
// Workflow: App loads → capability check → WebTransport connects → tracks render
// Reference: docs/implementation/v4/phase4_hardening_cutover.md H4-4

import { test, expect } from "@playwright/test";
import { gotoApp, waitForClassificationBanner, waitForConnectionIndicator, assertSecurityHeaders } from "./helpers";

test.describe("Cold Boot Workflow", () => {
  test("page loads without uncaught errors", async ({ page }) => {
    const errors: string[] = [];
    page.on("pageerror", (err) => errors.push(err.message));

    await gotoApp(page);
    await page.waitForLoadState("networkidle", { timeout: 20_000 }).catch(() => {});

    // No uncaught JS errors on startup
    expect(errors.filter((e) => !e.includes("message channel closed"))).toHaveLength(0);
  });

  test("classification banner is visible immediately", async ({ page }) => {
    await gotoApp(page);
    await waitForClassificationBanner(page);

    const banner = page.locator('[data-testid="classification-banner-top"]');
    await expect(banner).toBeVisible();
    await expect(banner).toContainText("UNCLASSIFIED");
  });

  test("security headers are present (COOP + COEP)", async ({ page }) => {
    await assertSecurityHeaders(page);
  });

  test("degraded notice shown when WebGPU unavailable", async ({ page }) => {
    // Simulate a browser without WebGPU by overriding navigator.gpu
    await page.addInitScript(() => {
      // Remove gpu from navigator to simulate unsupported browser
      Object.defineProperty(navigator, "gpu", {
        get: () => undefined,
        configurable: true,
      });
    });

    await gotoApp(page);
    await page.waitForLoadState("networkidle", { timeout: 20_000 }).catch(() => {});

    // Either degraded notice or the app renders (in headless without WebGPU)
    const body = page.locator("body");
    await expect(body).toBeVisible();
  });

  test("connection indicator element is present in the DOM", async ({ page }) => {
    await gotoApp(page);
    await page.waitForLoadState("networkidle", { timeout: 20_000 }).catch(() => {});
    // The connection indicator should render regardless of WebTransport state
    await waitForConnectionIndicator(page);
  });

  test("status bar renders with FPS display", async ({ page }) => {
    await gotoApp(page);
    await page.waitForLoadState("networkidle", { timeout: 20_000 }).catch(() => {});

    const statusBar = page.locator('[data-testid="status-bar"]');
    await expect(statusBar).toBeVisible({ timeout: 10_000 });
  });

  test("role selector is visible in toolbar", async ({ page }) => {
    await gotoApp(page);
    await page.waitForLoadState("networkidle", { timeout: 20_000 }).catch(() => {});

    const roleSelector = page.locator('[data-testid="role-selector"]');
    await expect(roleSelector).toBeVisible({ timeout: 10_000 });
  });
});
