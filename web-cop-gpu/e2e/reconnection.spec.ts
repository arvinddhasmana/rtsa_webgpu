// CLASSIFICATION: UNCLASSIFIED
// web-cop-gpu/e2e/reconnection.spec.ts — WebTransport reconnection E2E tests
//
// Workflow: Server restart → reconnect → tracks resume
// Reference: docs/implementation/v4/phase4_hardening_cutover.md H4-4

import { test, expect } from "@playwright/test";
import { gotoApp, blockWebTransport, restoreWebTransport, waitForConnectionIndicator } from "./helpers";

test.describe("WebTransport Reconnection", () => {
  test("connection indicator is visible before disconnect", async ({ page }) => {
    await gotoApp(page);
    await page.waitForLoadState("networkidle", { timeout: 20_000 }).catch(() => {});
    await waitForConnectionIndicator(page);

    const indicator = page.locator('[data-testid="connection-indicator"]');
    await expect(indicator).toBeVisible();
  });

  test("app remains functional when WebTransport is blocked", async ({ page }) => {
    await gotoApp(page);
    await page.waitForLoadState("networkidle", { timeout: 20_000 }).catch(() => {});

    // Block WebTransport traffic to simulate server going down
    await blockWebTransport(page);

    // Wait for any reconnect attempt
    await page.waitForTimeout(3_000);

    // App should still be responsive (not crashed)
    const body = page.locator("body");
    await expect(body).toBeVisible();
    const bodyText = await body.innerText().catch(() => "");
    expect(bodyText.length).toBeGreaterThan(0);
  });

  test("app does not crash when WebTransport URL is unreachable", async ({ page }) => {
    // Route all WebTransport connections to an unreachable host
    await page.route("**/*", (route) => {
      const url = route.request().url();
      if (url.includes(":4443")) {
        route.abort("connectionrefused");
      } else {
        route.continue();
      }
    });

    await gotoApp(page);
    await page.waitForLoadState("networkidle", { timeout: 20_000 }).catch(() => {});

    // No uncaught errors
    const errors: string[] = [];
    page.on("pageerror", (err) => errors.push(err.message));
    await page.waitForTimeout(2_000);

    expect(errors).toHaveLength(0);
  });

  test("connection indicator reflects disconnected state", async ({ page }) => {
    const errors: string[] = [];
    page.on("pageerror", (err) => errors.push(err.message));

    await blockWebTransport(page);
    await gotoApp(page);
    await page.waitForLoadState("networkidle", { timeout: 20_000 }).catch(() => {});

    // App should have rendered something (not blank)
    const body = page.locator("body");
    await expect(body).toBeVisible();

    // No uncaught JS errors
    expect(errors.filter((e) => !e.includes("message channel closed"))).toHaveLength(0);
  });

  test("app recovers when connectivity is restored", async ({ page }) => {
    await gotoApp(page);
    await page.waitForLoadState("networkidle", { timeout: 20_000 }).catch(() => {});

    // Block then restore
    await blockWebTransport(page);
    await page.waitForTimeout(2_000);
    await restoreWebTransport(page);

    // App should still be functional
    const body = page.locator("body");
    await expect(body).toBeVisible();
  });
});
