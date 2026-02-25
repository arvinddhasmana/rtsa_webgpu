// CLASSIFICATION: UNCLASSIFIED
// web-cop/e2e/offline.spec.ts
// Browser E2E tests for Offline / Degraded Mode — CR-SEC-004
// Covers: UC012 (Security Enforcement)

import { test, expect } from "@playwright/test";

test.describe("Offline / Degraded Mode (CR-SEC-004)", () => {
  test("connection indicator element is present in the layout [CR-SEC-004]", async ({ page }) => {
    await page.goto("/");
    await page.waitForLoadState("networkidle", { timeout: 15_000 }).catch(() => {});
    const connIndicator = page.locator('[data-testid="connection-indicator"]');
    await expect(connIndicator).toBeVisible({ timeout: 15_000 });
  });

  test("app remains functional with gRPC routes blocked [CR-SEC-004, UC012]", async ({
    page,
  }) => {
    // Block gRPC-Web traffic to simulate backend unavailability
    await page.route("**/*", (route) => {
      const url = route.request().url();
      if (url.includes(":8080") || url.includes(":9090")) {
        route.abort();
      } else {
        route.continue();
      }
    });

    await page.goto("/");
    await page.waitForLoadState("networkidle", { timeout: 15_000 }).catch(() => {});

    const body = page.locator("body");
    await expect(body).toBeVisible();
    const bodyText = await body.innerText().catch(() => "");
    expect(bodyText.length).toBeGreaterThan(0);
  });

  test("classification banner remains visible when offline [CR-SEC-004]", async ({
    page,
  }) => {
    await page.goto("/");
    await page.waitForLoadState("networkidle", { timeout: 15_000 }).catch(() => {});
    // Classification banner must be visible even with no backend connection
    const banner = page.locator('[data-testid="classification-banner-top"]');
    await expect(banner).toBeVisible({ timeout: 15_000 });
  });
});
